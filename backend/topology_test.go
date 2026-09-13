package backend

import (
	"encoding/json"
	"testing"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestMIBSymbolsLoaded(t *testing.T) {
	// Initialize or check MIBDB has loaded our new symbols
	symbols := []string{
		"dot1dBaseBridgeAddress",
		"dot1dBasePortIfIndex",
		"dot1dStpPortTable",
		"dot1dStpPortDesignatedBridge",
		"dot1dStpPortDesignatedPort",
		"dot1dTpFdbTable",
		"dot1dTpFdbPort",
		"dot1qTpFdbTable",
		"dot1qTpFdbPort",
		"cdpCacheEntry",
		"cdpCacheDeviceId",
		"cdpCacheAddress",
		"cdpCacheDevicePort",
		"lldpRemoteSystemsData",
		"lldpRemManAddrTable",
	}

	for _, s := range symbols {
		oid := datastore.MIBDB.NameToOID(s)
		if oid == "" || oid == s {
			t.Errorf("MIB symbol %s was not resolved to an OID", s)
		}
	}
}

func TestNeighborLineEntJSON(t *testing.T) {
	line := NeighborLineEnt{
		LineEnt: datastore.LineEnt{
			ID:         "L1",
			NodeID1:    "NET:N1",
			PollingID1: "P1",
			NodeID2:    "NODE:N2",
			PollingID2: "P2",
			Width:      2,
		},
		Confidence: "strict",
		Reason:     "LLDP",
	}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed NeighborLineEnt
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed.NodeID1 != "NET:N1" || parsed.Confidence != "strict" || parsed.Reason != "LLDP" {
		t.Errorf("unmarshaled line mismatch: %+v", parsed)
	}
}

func TestHasNeighborLine(t *testing.T) {
	lines := []NeighborLineEnt{
		{
			LineEnt: datastore.LineEnt{
				NodeID1:    "A",
				PollingID1: "1",
				NodeID2:    "B",
				PollingID2: "2",
			},
			Confidence: "strict",
		},
	}

	// Forward match
	match1 := &datastore.LineEnt{
		NodeID1:    "A",
		PollingID1: "1",
		NodeID2:    "B",
		PollingID2: "2",
	}
	if !hasNeighborLine(lines, match1) {
		t.Errorf("expected hasNeighborLine to return true for exact match")
	}

	// Reverse match
	match2 := &datastore.LineEnt{
		NodeID1:    "B",
		PollingID1: "2",
		NodeID2:    "A",
		PollingID2: "1",
	}
	if !hasNeighborLine(lines, match2) {
		t.Errorf("expected hasNeighborLine to return true for reversed match")
	}

	// No match
	noMatch := &datastore.LineEnt{
		NodeID1:    "A",
		PollingID1: "1",
		NodeID2:    "C",
		PollingID2: "3",
	}
	if hasNeighborLine(lines, noMatch) {
		t.Errorf("expected hasNeighborLine to return false for non-matching line")
	}
}

func TestAutoConnectLinesNone(t *testing.T) {
	strict, spec, err := AutoConnectLines(datastore.AutoLineNone)
	if err != nil {
		t.Errorf("unexpected error on AutoLineNone: %v", err)
	}
	if strict != 0 || spec != 0 {
		t.Errorf("expected 0 lines connected for AutoLineNone, got %d, %d", strict, spec)
	}
}
