package polling

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"golang.org/x/sync/singleflight"
)

const (
	envMaxBodySize         = 10 * 1024 * 1024 // 10 MB limit
	envBackoffWindow       = 30 * time.Second
	envRateLimitBaseWait   = 5 * time.Minute
	envRateLimitMaxWait    = 60 * time.Minute
	envDefaultHostInterval = 2 * time.Second
	envUserAgent           = "TWSNMP-FK/1.0 (+https://github.com/twsnmp/twsnmpfk)"
)

type envCacheEntry struct {
	data         []byte
	fetchedAt    time.Time
	expiresAt    time.Time
	staleUntil   time.Time
	backoffUntil time.Time
	lastErr      error
}

type envHostLimiter struct {
	mu           sync.Mutex
	lastReqAt    time.Time
	backoffUntil time.Time
	consec429    int
	lastErr      error
}

type envAPICache struct {
	mu           sync.RWMutex
	entries      map[string]*envCacheEntry
	hosts        map[string]*envHostLimiter
	sf           singleflight.Group
	client       *http.Client
	hostInterval time.Duration
}

var globalEnvCache = newEnvAPICache()

func newEnvAPICache() *envAPICache {
	return &envAPICache{
		entries:      make(map[string]*envCacheEntry),
		hosts:        make(map[string]*envHostLimiter),
		client:       &http.Client{},
		hostInterval: envDefaultHostInterval,
	}
}

// ClearEnvAPICache clears all cached API entries and host rate-limit states (useful for unit tests and manual refresh).
func ClearEnvAPICache() {
	globalEnvCache.mu.Lock()
	defer globalEnvCache.mu.Unlock()
	globalEnvCache.entries = make(map[string]*envCacheEntry)
	globalEnvCache.hosts = make(map[string]*envHostLimiter)
}

func (c *envAPICache) getHostLimiter(rawURL string) *envHostLimiter {
	host := "default"
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		host = strings.ToLower(u.Host)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	hl, ok := c.hosts[host]
	if !ok {
		hl = &envHostLimiter{}
		c.hosts[host] = hl
	}
	return hl
}

func parseRetryAfter(h string) time.Duration {
	h = strings.TrimSpace(h)
	if h == "" {
		return 0
	}
	if sec, err := strconv.Atoi(h); err == nil && sec > 0 {
		d := time.Duration(sec) * time.Second
		if d > envRateLimitMaxWait {
			return envRateLimitMaxWait
		}
		return d
	}
	if t, err := http.ParseTime(h); err == nil {
		d := time.Until(t)
		if d > 0 {
			if d > envRateLimitMaxWait {
				return envRateLimitMaxWait
			}
			return d
		}
	}
	return 0
}

// fetchWithCache fetches an HTTP JSON resource using singleflight deduplication,
// minimum TTL caching, per-host rate limiting, 429 exponential backoff, and stale-if-error fallback.
// It returns (bodyBytes, fromCache, rttNano, err).
func (c *envAPICache) fetchWithCache(key, rawURL string, minTTL time.Duration, timeoutSec int) ([]byte, bool, int64, error) {
	now := time.Now()

	c.mu.RLock()
	ent, ok := c.entries[key]
	if ok {
		if len(ent.data) > 0 && now.Before(ent.expiresAt) {
			dataCopy := append([]byte(nil), ent.data...)
			c.mu.RUnlock()
			return dataCopy, true, 0, nil
		}
		if now.Before(ent.backoffUntil) && ent.lastErr != nil {
			if len(ent.data) > 0 && now.Before(ent.staleUntil) {
				dataCopy := append([]byte(nil), ent.data...)
				c.mu.RUnlock()
				return dataCopy, true, 0, nil
			}
			err := ent.lastErr
			c.mu.RUnlock()
			return nil, false, 0, err
		}
	}
	c.mu.RUnlock()

	hl := c.getHostLimiter(rawURL)
	hl.mu.Lock()
	hostBackoffUntil := hl.backoffUntil
	hostLastErr := hl.lastErr
	hl.mu.Unlock()

	if now.Before(hostBackoffUntil) && hostLastErr != nil {
		c.mu.RLock()
		if ok && len(ent.data) > 0 && now.Before(ent.staleUntil) {
			dataCopy := append([]byte(nil), ent.data...)
			c.mu.RUnlock()
			return dataCopy, true, 0, nil
		}
		c.mu.RUnlock()
		return nil, false, 0, hostLastErr
	}

	if timeoutSec < 3 {
		timeoutSec = 10
	} else if timeoutSec > 15 {
		timeoutSec = 15
	}

	val, err, _ := c.sf.Do(key, func() (any, error) {
		nowInner := time.Now()
		c.mu.RLock()
		if existing, exists := c.entries[key]; exists && len(existing.data) > 0 && nowInner.Before(existing.expiresAt) {
			dataCopy := append([]byte(nil), existing.data...)
			c.mu.RUnlock()
			return envSFResult{data: dataCopy, fromCache: true, rtt: 0}, nil
		}
		c.mu.RUnlock()

		// Serialize & throttle requests per host to prevent burst 429 errors
		hl.mu.Lock()
		if time.Now().Before(hl.backoffUntil) && hl.lastErr != nil {
			err := hl.lastErr
			hl.mu.Unlock()
			return c.handleFetchError(key, envBackoffWindow, err)
		}
		if c.hostInterval > 0 && !hl.lastReqAt.IsZero() {
			wait := c.hostInterval - time.Since(hl.lastReqAt)
			if wait > 0 {
				time.Sleep(wait)
			}
		}
		hl.lastReqAt = time.Now()
		hl.mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
		defer cancel()

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("User-Agent", envUserAgent)
		req.Header.Set("Accept", "application/json, application/geo+json, */*")

		st := time.Now()
		resp, doErr := c.client.Do(req)
		rtt := time.Since(st).Nanoseconds()
		if doErr != nil {
			return c.handleFetchError(key, envBackoffWindow, doErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			statusErr := fmt.Errorf("http status %d (%s)", resp.StatusCode, resp.Status)
			retryWait := parseRetryAfter(resp.Header.Get("Retry-After"))
			hl.mu.Lock()
			if retryWait <= 0 {
				retryWait = envRateLimitBaseWait << hl.consec429
				if retryWait > envRateLimitMaxWait || retryWait <= 0 {
					retryWait = envRateLimitMaxWait
				}
			}
			hl.consec429++
			hl.backoffUntil = time.Now().Add(retryWait)
			hl.lastErr = statusErr
			hl.mu.Unlock()
			return c.handleFetchError(key, retryWait, statusErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			statusErr := fmt.Errorf("http status %d (%s)", resp.StatusCode, resp.Status)
			return c.handleFetchError(key, envBackoffWindow, statusErr)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, envMaxBodySize))
		if readErr != nil {
			return c.handleFetchError(key, envBackoffWindow, readErr)
		}
		if len(body) == 0 {
			return c.handleFetchError(key, envBackoffWindow, fmt.Errorf("empty response body"))
		}

		// Reset 429 streak on success
		hl.mu.Lock()
		hl.consec429 = 0
		hl.backoffUntil = time.Time{}
		hl.lastErr = nil
		hl.mu.Unlock()

		staleDur := minTTL * 4
		if staleDur < 30*time.Minute {
			staleDur = 30 * time.Minute
		}
		if staleDur > 2*time.Hour {
			staleDur = 2 * time.Hour
		}

		fetchedAt := time.Now()
		c.mu.Lock()
		c.entries[key] = &envCacheEntry{
			data:       body,
			fetchedAt:  fetchedAt,
			expiresAt:  fetchedAt.Add(minTTL),
			staleUntil: fetchedAt.Add(staleDur),
		}
		c.mu.Unlock()

		return envSFResult{data: body, fromCache: false, rtt: rtt}, nil
	})

	if err != nil {
		return nil, false, 0, err
	}
	res := val.(envSFResult)
	return res.data, res.fromCache, res.rtt, nil
}

type envSFResult struct {
	data      []byte
	fromCache bool
	rtt       int64
}

func (c *envAPICache) handleFetchError(key string, backoffDur time.Duration, fetchErr error) (any, error) {
	now := time.Now()
	if backoffDur <= 0 {
		backoffDur = envBackoffWindow
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[key]; ok {
		existing.backoffUntil = now.Add(backoffDur)
		existing.lastErr = fetchErr
		if len(existing.data) > 0 && now.Before(existing.staleUntil) {
			dataCopy := append([]byte(nil), existing.data...)
			return envSFResult{data: dataCopy, fromCache: true, rtt: 0}, nil
		}
	} else {
		c.entries[key] = &envCacheEntry{
			backoffUntil: now.Add(backoffDur),
			lastErr:      fetchErr,
		}
	}
	return nil, fetchErr
}

// calcDistanceKm calculates the great-circle distance in kilometers between two coordinates using the Haversine formula.
func calcDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	radLat1 := lat1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// parseEnvParams parses "k1=v1,k2=v2" parameter strings into a lowercase-key map.
func parseEnvParams(params string) map[string]string {
	ret := make(map[string]string)
	if strings.TrimSpace(params) == "" {
		return ret
	}
	for _, part := range strings.Split(params, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			k := strings.ToLower(strings.TrimSpace(kv[0]))
			v := strings.TrimSpace(kv[1])
			ret[k] = v
		}
	}
	return ret
}

// parseLocLngLat parses TWSNMP's Node.Loc / LocConf.Center string ("lng,lat") into (lat, lon, ok).
// If the first value is within [-90, 90] and the second value is outside [-90, 90] (e.g. "35.68,139.75"),
// it safely normalizes the order to (lat=35.68, lon=139.75).
func parseLocLngLat(loc string) (float64, float64, bool) {
	loc = strings.TrimSpace(loc)
	if loc == "" {
		return 0, 0, false
	}
	parts := strings.Split(loc, ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	v1, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	v2, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	if v1 == 0 && v2 == 0 {
		return 0, 0, false
	}
	// Standard Location.svelte format is "lng,lat" (e.g. "139.75,35.68")
	var lat, lon float64
	if math.Abs(v2) <= 90 && math.Abs(v1) <= 180 {
		lon = v1
		lat = v2
		// Check if user accidentally wrote "lat,lon" where lon > 90 (like Japan 35.68, 139.75)
		if math.Abs(v1) <= 90 && math.Abs(v2) > 90 {
			lat = v1
			lon = v2
		}
		return lat, lon, true
	}
	if math.Abs(v1) <= 90 && math.Abs(v2) <= 180 {
		return v1, v2, true
	}
	return 0, 0, false
}

// resolveNodeLatLon resolves target (lat, lon) using priority:
// 1. Explicit lat/lon in pMap
// 2. NodeEnt.Loc ("lng,lat")
// 3. GeoIP lookup from NodeEnt.IP ("CC,lat,lon,City")
// 4. datastore.LocConf.Center ("lng,lat", default Tokyo 35.68, 139.75)
func resolveTargetLatLon(pe *datastore.PollingEnt, pMap map[string]string) (float64, float64) {
	if pMap != nil {
		latStr := pMap["lat"]
		if latStr == "" {
			latStr = pMap["latitude"]
		}
		lonStr := pMap["lon"]
		if lonStr == "" {
			lonStr = pMap["lng"]
		}
		if lonStr == "" {
			lonStr = pMap["longitude"]
		}
		if latStr != "" && lonStr != "" {
			lat, err1 := strconv.ParseFloat(latStr, 64)
			lon, err2 := strconv.ParseFloat(lonStr, 64)
			if err1 == nil && err2 == nil && math.Abs(lat) <= 90 && math.Abs(lon) <= 180 {
				return lat, lon
			}
		}
	}

	if pe != nil && pe.NodeID != "" {
		if n := datastore.GetNode(pe.NodeID); n != nil {
			if lat, lon, ok := parseLocLngLat(n.Loc); ok {
				return lat, lon
			}
			if n.IP != "" {
				// datastore.GetLoc returns "ISO,lat,lon,City"
				geo := datastore.GetLoc(n.IP)
				if geo != "" && !strings.HasPrefix(geo, "LOCAL,") {
					gParts := strings.Split(geo, ",")
					if len(gParts) >= 3 {
						gLat, err1 := strconv.ParseFloat(gParts[1], 64)
						gLon, err2 := strconv.ParseFloat(gParts[2], 64)
						if err1 == nil && err2 == nil && (gLat != 0 || gLon != 0) {
							return gLat, gLon
						}
					}
				}
			}
		}
	}

	if lat, lon, ok := parseLocLngLat(datastore.LocConf.Center); ok {
		return lat, lon
	}
	// Default: Tokyo
	return 35.68, 139.75
}
