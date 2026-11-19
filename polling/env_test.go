package polling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setMockEnvHTTPHandler(fn func(req *http.Request) (int, string)) func() {
	origClient := globalEnvCache.client
	origInterval := globalEnvCache.hostInterval
	globalEnvCache.hostInterval = 0
	globalEnvCache.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			code, body := fn(req)
			return &http.Response{
				StatusCode: code,
				Status:     fmt.Sprintf("%d %s", code, http.StatusText(code)),
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Request:    req,
			}, nil
		}),
	}
	return func() {
		globalEnvCache.client = origClient
		globalEnvCache.hostInterval = origInterval
	}
}

func TestCalcDistanceKmAndLocResolve(t *testing.T) {
	// Distance between Tokyo Station (35.6812, 139.7671) and Yokohama Station (35.4658, 139.6223) is ~27.3 km
	dist := calcDistanceKm(35.6812, 139.7671, 35.4658, 139.6223)
	if dist < 26.0 || dist > 29.0 {
		t.Fatalf("unexpected Haversine distance Tokyo->Yokohama: %.2f km", dist)
	}

	// Standard Location.svelte format "lng,lat"
	lat, lon, ok := parseLocLngLat("139.7671,35.6812")
	if !ok || math.Abs(lat-35.6812) > 0.001 || math.Abs(lon-139.7671) > 0.001 {
		t.Fatalf("parseLocLngLat failed for lng,lat: lat=%f, lon=%f, ok=%v", lat, lon, ok)
	}

	// Reversed "lat,lng" input where lon > 90
	lat2, lon2, ok2 := parseLocLngLat("35.6812,139.7671")
	if !ok2 || math.Abs(lat2-35.6812) > 0.001 || math.Abs(lon2-139.7671) > 0.001 {
		t.Fatalf("parseLocLngLat failed for reversed lat,lng: lat=%f, lon=%f, ok=%v", lat2, lon2, ok2)
	}
}

func TestPollingEnvQuake(t *testing.T) {
	ClearEnvAPICache()

	nowStr := time.Now().In(jstZone).Format("2006/01/02 15:04:05")
	mockP2P := fmt.Sprintf(`[
		{
			"id": "q1",
			"code": 551,
			"time": "%s",
			"earthquake": {
				"time": "%s",
				"maxScale": 45,
				"hypocenter": {
					"name": "千葉県北西部",
					"latitude": 35.65,
					"longitude": 140.05,
					"depth": 70,
					"magnitude": 5.4
				}
			},
			"points": [
				{"pref": "東京都", "addr": "東京千代田区", "scale": 30},
				{"pref": "千葉県", "addr": "千葉中央区", "scale": 45}
			]
		}
	]`, nowStr, nowStr)

	var reqCount int32
	restore := setMockEnvHTTPHandler(func(r *http.Request) (int, string) {
		atomic.AddInt32(&reqCount, 1)
		return http.StatusOK, mockP2P
	})
	defer restore()

	// 1. Tokyo node: scale=30 -> intensity=3.0 -> warn level via setLevel('warn')
	peTokyo := &datastore.PollingEnt{
		ID:      "test-env-quake-tokyo",
		Name:    "東京地震監視",
		Type:    "env",
		Mode:    "quake",
		State:   "normal",
		Params:  "url=https://mock.local/p2pquake,lat=35.68,lon=139.75,pref=東京都,max_age=3600",
		Script:  "if (intensity >= 5.0) { setLevel('high'); false; } else if (intensity >= 3.0) { setLevel('warn'); false; } else { true; }",
		Level:   "high",
		PollInt: 120,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peTokyo)

	if peTokyo.State != "warn" {
		t.Fatalf("expected Tokyo state warn for intensity 3.0, got %s (result=%+v)", peTokyo.State, peTokyo.Result)
	}
	if peTokyo.Result["intensity"] != 3.0 || peTokyo.Result["max_intensity"] != 5.0 || peTokyo.Result["magnitude"] != 5.4 {
		t.Fatalf("unexpected quake metrics for Tokyo: %+v", peTokyo.Result)
	}

	// 2. Chiba node polling same URL immediately -> should hit cache (reqCount remains 1) and evaluate scale=45 (5.0 -> high)
	peChiba := &datastore.PollingEnt{
		ID:      "test-env-quake-chiba",
		Name:    "千葉地震監視",
		Type:    "env",
		Mode:    "quake",
		State:   "normal",
		Params:  "url=https://mock.local/p2pquake,lat=35.60,lon=140.10,pref=千葉県,max_age=3600",
		Script:  "if (intensity >= 5.0) { setLevel('high'); false; } else if (intensity >= 3.0) { setLevel('warn'); false; } else { true; }",
		Level:   "high",
		PollInt: 120,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peChiba)

	if peChiba.State != "high" {
		t.Fatalf("expected Chiba state high for intensity 5.0 (5弱), got %s", peChiba.State)
	}
	if atomic.LoadInt32(&reqCount) != 1 {
		t.Fatalf("expected shared cache hit (reqCount=1), got %d", atomic.LoadInt32(&reqCount))
	}
	if peChiba.Result["cached"] != 1.0 {
		t.Fatalf("expected cached=1.0 on second request, got %v", peChiba.Result["cached"])
	}
}

func TestPollingEnvLiden(t *testing.T) {
	ClearEnvAPICache()

	// Node at (35.6800, 139.7500).
	// Strike 1 at (35.7000, 139.7600) -> ~2.4 km away (within 10km warn_radius)
	// Strike 2 at (35.8000, 139.8000) -> ~14.1 km away (within 20km radius)
	// Strike 3 at (36.5000, 140.5000) -> ~113 km away (outside 20km radius)
	mockGeoJSON := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"geometry": {"type": "Point", "coordinates": [139.7600, 35.7000]},
				"properties": {"type": 2}
			},
			{
				"type": "Feature",
				"geometry": {"type": "Point", "coordinates": [139.8000, 35.8000]},
				"properties": {"type": 0}
			},
			{
				"type": "Feature",
				"geometry": {"type": "Point", "coordinates": [140.5000, 36.5000]},
				"properties": {"type": 2}
			}
		]
	}`

	restore := setMockEnvHTTPHandler(func(r *http.Request) (int, string) {
		if strings.HasSuffix(r.URL.Path, "/targetTimes_N3.json") {
			return http.StatusOK, `[
				{"basetime":"20261119070000","validtime":"20261119080000","elements":["thns","trns"]},
				{"basetime":"20261119070500","validtime":"20261119070500","elements":["liden"]}
			]`
		}
		if !strings.Contains(r.URL.Path, "20261119070500") {
			return http.StatusNotFound, ""
		}
		return http.StatusOK, mockGeoJSON
	})
	defer restore()

	pe := &datastore.PollingEnt{
		ID:      "test-env-liden",
		Name:    "東京DC落雷監視",
		Type:    "env",
		Mode:    "liden",
		State:   "normal",
		Params:  "url=https://mock.local/targetTimes_N3.json,data_url_fmt=https://mock.local/nowc/%s/%s/liden.geojson,lat=35.6800,lon=139.7500,radius=20,warn_radius=10",
		Script:  "if (close_count > 0 || nearest_km <= 10.0) { setLevel('high'); false; } else if (strike_count > 0) { setLevel('warn'); false; } else { true; }",
		Level:   "high",
		PollInt: 120,
		Timeout: 5,
		Result:  make(map[string]any),
	}

	doPollingEnv(pe)

	if pe.State != "high" {
		t.Fatalf("expected state high for close lightning strike, got %s (result=%+v)", pe.State, pe.Result)
	}
	if pe.Result["strike_count"] != 2.0 || pe.Result["close_count"] != 1.0 || pe.Result["national_count"] != 3.0 {
		t.Fatalf("unexpected liden strike counts: %+v", pe.Result)
	}
	nearest, _ := pe.Result["nearest_km"].(float64)
	if nearest < 2.0 || nearest > 3.0 {
		t.Fatalf("expected nearest_km ~2.4km, got %.1f", nearest)
	}
}

func TestPollingEnvSpace(t *testing.T) {
	ClearEnvAPICache()

	// M2.5 solar flare (2.5e-5 W/m^2) -> score ~ 3.40, and Kp = 6.33 (G2 storm)
	mockXray := `[
		{"time_tag": "2026-11-19T07:00:00Z", "flux": 1.2e-6, "energy": "0.05-0.4nm"},
		{"time_tag": "2026-11-19T07:00:00Z", "flux": 2.5e-5, "energy": "0.1-0.8nm"}
	]`
	mockKp := `[
		{"time_tag": "2026-11-19T06:59:00", "kp_index": 5, "estimated_kp": 5.33},
		{"time_tag": "2026-11-19T07:00:00", "kp_index": 6, "estimated_kp": 6.33}
	]`

	restore := setMockEnvHTTPHandler(func(r *http.Request) (int, string) {
		if strings.HasSuffix(r.URL.Path, "/xray.json") {
			return http.StatusOK, mockXray
		}
		return http.StatusOK, mockKp
	})
	defer restore()

	pe := &datastore.PollingEnt{
		ID:      "test-env-space",
		Name:    "宇宙天気監視",
		Type:    "env",
		Mode:    "space",
		State:   "normal",
		Params:  "xray_url=https://mock.local/xray.json,kp_url=https://mock.local/kp.json",
		Script:  "if (flare_score >= 4.0 || kp_index >= 7.0) { setLevel('high'); false; } else if (flare_score >= 3.0 || kp_index >= 5.0) { setLevel('warn'); false; } else { true; }",
		Level:   "warn",
		PollInt: 300,
		Timeout: 5,
		Result:  make(map[string]any),
	}

	doPollingEnv(pe)

	if pe.State != "warn" {
		t.Fatalf("expected state warn for M2.5 flare and Kp 6.33, got %s (result=%+v)", pe.State, pe.Result)
	}
	if pe.Result["flare_class"] != "M2.5" || pe.Result["kp_index"] != 6.33 || pe.Result["g_scale"] != 2.0 {
		t.Fatalf("unexpected space weather metrics: %+v", pe.Result)
	}
}

func TestPollingEnvWeatherAndStaleFallback(t *testing.T) {
	ClearEnvAPICache()

	// 3 hours ago pressure was 1012.0, current is 1004.5 -> diff = -7.5 hPa, CAPE = 1450
	mockWeather := `{
		"latitude": 35.68,
		"longitude": 139.75,
		"current": {
			"time": "2026-11-19T16:00",
			"temperature_2m": 22.4,
			"relative_humidity_2m": 78,
			"surface_pressure": 1004.5,
			"pressure_msl": 1008.2,
			"wind_speed_10m": 8.5,
			"wind_gusts_10m": 16.2,
			"precipitation": 2.5,
			"weather_code": 95,
			"cape": 1450
		},
		"hourly": {
			"time": ["t-6","t-5","t-4","t-3","t-2","t-1","t0"],
			"surface_pressure": [1014.0, 1013.5, 1013.0, 1012.0, 1009.5, 1007.0, 1004.5],
			"cape": [100, 200, 400, 600, 900, 1200, 1450]
		}
	}`

	var failMode int32
	restore := setMockEnvHTTPHandler(func(r *http.Request) (int, string) {
		if atomic.LoadInt32(&failMode) == 1 {
			return http.StatusServiceUnavailable, ""
		}
		return http.StatusOK, mockWeather
	})
	defer restore()

	pe := &datastore.PollingEnt{
		ID:      "test-env-weather",
		Name:    "拠点気象監視",
		Type:    "env",
		Mode:    "weather",
		State:   "normal",
		Params:  "url=https://mock.local/forecast,lat=35.68,lon=139.75",
		Script:  "if (wind_gust >= 25.0 || cape >= 2500 || pressure_diff_3h <= -10.0) { setLevel('high'); false; } else if (wind_gust >= 15.0 || cape >= 1000 || pressure_diff_3h <= -5.0 || weather_code >= 95) { setLevel('warn'); false; } else { true; }",
		Level:   "warn",
		PollInt: 600,
		Timeout: 5,
		Result:  make(map[string]any),
	}

	doPollingEnv(pe)

	if pe.State != "warn" {
		t.Fatalf("expected state warn for pressure_diff_3h=-7.5 & CAPE=1450, got %s (result=%+v)", pe.State, pe.Result)
	}
	if pe.Result["pressure"] != 1004.5 || pe.Result["pressure_diff_3h"] != -7.5 || pe.Result["cape"] != 1450.0 {
		t.Fatalf("unexpected weather metrics: %+v", pe.Result)
	}

	// Simulate cache expiration but server returning 503 -> Stale-if-Error should serve cached data without failing polling
	globalEnvCache.mu.Lock()
	for _, ent := range globalEnvCache.entries {
		ent.expiresAt = time.Now().Add(-1 * time.Second)
	}
	globalEnvCache.mu.Unlock()
	atomic.StoreInt32(&failMode, 1)

	doPollingEnv(pe)
	if pe.Result["error"] != nil {
		t.Fatalf("expected stale-if-error fallback without error, got error=%v", pe.Result["error"])
	}
	if pe.Result["cached"] != 1.0 {
		t.Fatalf("expected cached=1.0 during stale-if-error fallback, got %v", pe.Result["cached"])
	}
}

func TestEnvPollingTemplatesAndAutoLoc(t *testing.T) {
	// 1. Verify both polling_ja.json and polling_en.json parse cleanly and contain Type=="env" templates
	for _, path := range []string{"../datastore/conf/polling_ja.json", "../datastore/conf/polling_en.json"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		var list []datastore.PollingTemplateEnt
		if err := json.Unmarshal(b, &list); err != nil {
			t.Fatalf("failed to unmarshal %s: %v", path, err)
		}
		envCount := 0
		for _, pt := range list {
			if pt.Type == "env" {
				envCount++
			}
		}
		if envCount < 6 {
			t.Fatalf("expected at least 6 env templates in %s, found %d", path, envCount)
		}
	}

	// 2. Verify GetAutoPollings with AutoParam="loc" expands NodeEnt.Loc ("lng,lat") into Params ("lat=...,lon=...")
	node := &datastore.NodeEnt{
		ID:   "node-osaka",
		Name: "大阪拠点ルーター",
		Loc:  "135.5023,34.6937", // Osaka (lng, lat)
	}
	pt := &datastore.PollingTemplateEnt{
		Name:      "落雷監視",
		Type:      "env",
		Mode:      "liden",
		Params:    "radius=20,warn_radius=10",
		AutoParam: "loc",
		Level:     "high",
	}
	autoList := GetAutoPollings(node, pt)
	if len(autoList) != 1 {
		t.Fatalf("expected 1 auto polling, got %d", len(autoList))
	}
	gen := autoList[0]
	if gen.PollInt != 120 || gen.LogMode != datastore.LogModeAlways {
		t.Fatalf("expected PollInt=120 and LogMode=1, got PollInt=%d LogMode=%d", gen.PollInt, gen.LogMode)
	}
	pMap := parseEnvParams(gen.Params)
	if pMap["lat"] != "34.6937" || pMap["lon"] != "135.5023" || pMap["radius"] != "20" {
		t.Fatalf("unexpected auto-expanded Params: %s (parsed=%+v)", gen.Params, pMap)
	}
}

func TestPollingEnvGlobalUSGSAndLidenFallback(t *testing.T) {
	ClearEnvAPICache()

	// 1. Test USGS global earthquake feed for San Francisco (37.7749, -122.4194)
	nowMs := time.Now().Add(-5 * time.Minute).UnixMilli()
	mockUSGS := fmt.Sprintf(`{
		"type": "FeatureCollection",
		"features": [
			{
				"id": "nc123",
				"properties": {
					"mag": 5.2,
					"place": "12km SSW of San Jose, CA",
					"time": %d,
					"mmi": 5.8,
					"tsunami": 0
				},
				"geometry": {
					"coordinates": [-121.9, 37.3, 10.5]
				}
			},
			{
				"id": "far999",
				"properties": {
					"mag": 7.0,
					"place": "Far away",
					"time": %d,
					"mmi": 8.0,
					"tsunami": 1
				},
				"geometry": {
					"coordinates": [10.0, 10.0, 10.0]
				}
			}
		]
	}`, nowMs, nowMs)

	restore := setMockEnvHTTPHandler(func(r *http.Request) (int, string) {
		if strings.Contains(r.URL.Path, "usgs") {
			return http.StatusOK, mockUSGS
		}
		// Open-Meteo thunderstorm fallback response for San Francisco
		return http.StatusOK, `{
			"latitude": 37.77,
			"longitude": -122.42,
			"current": {
				"time": "2026-11-20T04:00",
				"weather_code": 96,
				"cape": 2200
			}
		}`
	})
	defer restore()

	peQuake := &datastore.PollingEnt{
		ID:      "test-env-quake-usgs",
		Name:    "SF地震監視",
		Type:    "env",
		Mode:    "quake",
		State:   "normal",
		Params:  "url=https://mock.local/usgs.geojson,source=usgs,lat=37.7749,lon=-122.4194,radius=300,max_age=3600",
		Script:  "if (intensity >= 5.0) { setLevel('high'); false; } else if (intensity >= 3.0) { setLevel('warn'); false; } else { true; }",
		Level:   "high",
		PollInt: 120,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peQuake)

	if peQuake.State != "high" {
		t.Fatalf("expected SF USGS quake state high for MMI 5.8 (intensity 5.0), got %s (result=%+v)", peQuake.State, peQuake.Result)
	}
	if peQuake.Result["mmi"] != 5.8 || peQuake.Result["magnitude"] != 5.2 || peQuake.Result["source"] != "usgs" {
		t.Fatalf("unexpected USGS quake metrics: %+v", peQuake.Result)
	}

	// 2. Test Global Liden fallback (source=auto with San Francisco coords -> automatically uses Open-Meteo thunderstorm check)
	peLidenGlobal := &datastore.PollingEnt{
		ID:      "test-env-liden-global",
		Name:    "SF落雷リスク監視",
		Type:    "env",
		Mode:    "liden",
		State:   "normal",
		Params:  "url=https://mock.local/openmeteo,lat=37.7749,lon=-122.4194",
		Script:  "if (close_count > 0) { setLevel('high'); false; } else if (strike_count > 0) { setLevel('warn'); false; } else { true; }",
		Level:   "high",
		PollInt: 120,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peLidenGlobal)

	if peLidenGlobal.State != "high" {
		t.Fatalf("expected SF global liden fallback state high for WMO 96 + CAPE 2200, got %s (result=%+v)", peLidenGlobal.State, peLidenGlobal.Result)
	}
	if peLidenGlobal.Result["source"] != "openmeteo" || peLidenGlobal.Result["close_count"] != 1.0 {
		t.Fatalf("unexpected global liden fallback result: %+v", peLidenGlobal.Result)
	}
}

func TestPollingEnv429RateLimitBackoff(t *testing.T) {
	ClearEnvAPICache()

	var reqCount int32
	origClient := globalEnvCache.client
	origInterval := globalEnvCache.hostInterval
	globalEnvCache.hostInterval = 0
	globalEnvCache.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&reqCount, 1)
			hdr := make(http.Header)
			hdr.Set("Retry-After", "120")
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Status:     "429 Too Many Requests",
				Header:     hdr,
				Body:       io.NopCloser(bytes.NewBufferString("")),
				Request:    req,
			}, nil
		}),
	}
	defer func() {
		globalEnvCache.client = origClient
		globalEnvCache.hostInterval = origInterval
	}()

	peTokyo := &datastore.PollingEnt{
		ID:      "test-429-tokyo",
		Name:    "東京429テスト",
		Type:    "env",
		Mode:    "weather",
		Params:  "url=https://api.open-meteo.com/v1/forecast,lat=35.68,lon=139.75",
		PollInt: 600,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peTokyo)
	if atomic.LoadInt32(&reqCount) != 1 {
		t.Fatalf("expected 1 HTTP call on first 429, got %d", atomic.LoadInt32(&reqCount))
	}

	// Second polling for a DIFFERENT coordinate on the SAME host (api.open-meteo.com) must be blocked by host-level backoff
	peSF := &datastore.PollingEnt{
		ID:      "test-429-sf",
		Name:    "SF429テスト",
		Type:    "env",
		Mode:    "weather",
		Params:  "url=https://api.open-meteo.com/v1/forecast,lat=37.77,lon=-122.42",
		PollInt: 600,
		Timeout: 5,
		Result:  make(map[string]any),
	}
	doPollingEnv(peSF)
	if atomic.LoadInt32(&reqCount) != 1 {
		t.Fatalf("expected host-level 429 backoff to suppress 2nd request to same host, got %d calls", atomic.LoadInt32(&reqCount))
	}
}
