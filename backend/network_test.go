package backend

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestGetSNMPAgentForNetwork(t *testing.T) {
	// SNMPv2c with valid community
	v2Valid := &datastore.NetworkEnt{
		IP:        "192.168.1.1",
		SnmpMode:  "v2c",
		Community: "public",
	}
	agent := getSNMPAgentForNetwork(v2Valid)
	if agent == nil {
		t.Fatal("expected non-nil agent for valid v2c")
	}
	if agent.Community != "public" || agent.Version != gosnmp.Version2c {
		t.Errorf("unexpected agent settings: Community=%s, Version=%v", agent.Community, agent.Version)
	}

	// SNMPv2c with empty community should return nil
	v2EmptyComm := &datastore.NetworkEnt{
		IP:        "192.168.1.1",
		SnmpMode:  "v2c",
		Community: "",
	}
	if getSNMPAgentForNetwork(v2EmptyComm) != nil {
		t.Error("expected nil agent for empty community in v2c")
	}

	// SNMPv3 with valid user and empty community must return non-nil agent
	v3Modes := []string{"v3auth", "v3authpriv", "v3authprivex", "v3sha256aes128", "v3sha512aes256"}
	for _, mode := range v3Modes {
		v3Valid := &datastore.NetworkEnt{
			IP:        "192.168.1.2",
			SnmpMode:  mode,
			User:      "admin",
			Password:  "password123",
			Community: "",
		}
		a := getSNMPAgentForNetwork(v3Valid)
		if a == nil {
			t.Errorf("expected non-nil agent for %s with empty community", mode)
			continue
		}
		if a.Version != gosnmp.Version3 {
			t.Errorf("expected Version3 for %s, got %v", mode, a.Version)
		}
	}

	// SNMPv3 with empty user must return nil
	v3NoUser := &datastore.NetworkEnt{
		IP:       "192.168.1.2",
		SnmpMode: "v3auth",
		User:     "",
	}
	if getSNMPAgentForNetwork(v3NoUser) != nil {
		t.Error("expected nil agent for v3 with empty user")
	}
}
