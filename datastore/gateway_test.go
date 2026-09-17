package datastore

import (
	"testing"
)

func TestGetDefaultGateway(t *testing.T) {
	gw := GetDefaultGateway()
	t.Logf("Detected default gateway: %s", gw)
}
