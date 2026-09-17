package datastore

import (
	"sync"
	"time"
)

var (
	cachedGateway   string
	cachedGatewayAt time.Time
	gatewayLock     sync.RWMutex
)

// GetDefaultGateway returns the IPv4 default gateway address of the host.
func GetDefaultGateway() string {
	gatewayLock.RLock()
	if cachedGateway != "" && time.Since(cachedGatewayAt) < time.Minute {
		gw := cachedGateway
		gatewayLock.RUnlock()
		return gw
	}
	gatewayLock.RUnlock()

	gatewayLock.Lock()
	defer gatewayLock.Unlock()

	// Double check
	if cachedGateway != "" && time.Since(cachedGatewayAt) < time.Minute {
		return cachedGateway
	}

	gw := discoverDefaultGateway()
	if gw != "" {
		cachedGateway = gw
		cachedGatewayAt = time.Now()
	}
	return gw
}

// SetDefaultGatewayForTest sets cached gateway for unit tests.
func SetDefaultGatewayForTest(gw string) {
	gatewayLock.Lock()
	defer gatewayLock.Unlock()
	cachedGateway = gw
	cachedGatewayAt = time.Now()
}
