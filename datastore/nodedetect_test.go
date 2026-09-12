package datastore

import (
	"context"
	"os"
	"sync"
	"testing"
)

func TestNodeDetect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	td, err := os.MkdirTemp("", "twsnmpfk_nodedetect_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	LoadNodeDetectRules()

	tests := []struct {
		name         string
		input        DetectInput
		wantCategory string
		wantRuleID   string
		wantIcon     string
	}{
		{
			name: "Yamaha RTX1210 by sysObjectID",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.1182.1.33",
				SysDescr:    "Yamaha Corporation RTX1210 Rev.14.01.40",
				Vendor:      "YAMAHA CORPORATION",
			},
			wantCategory: "router",
			wantRuleID:   "yamaha_rtx",
			wantIcon:     "router",
		},
		{
			name: "Yamaha SWX Switch by sysObjectID",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.1182.2.1",
				SysDescr:    "SWX2200-8G",
				Vendor:      "YAMAHA CORPORATION",
			},
			wantCategory: "switch",
			wantRuleID:   "yamaha_swx",
			wantIcon:     "mdi-switch",
		},
		{
			name: "Yamaha WLX AP by sysObjectID",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.1182.3.1",
				SysDescr:    "WLX202",
				Vendor:      "YAMAHA CORPORATION",
			},
			wantCategory: "wifi",
			wantRuleID:   "yamaha_wlx",
			wantIcon:     "mdi-router-wireless",
		},
		{
			name: "Cisco Catalyst 2960X",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.9.1.1208",
				SysDescr:    "Cisco IOS Software, C2960X Software (C2960X-UNIVERSALK9-M), Version 15.2(2)E7",
				Vendor:      "Cisco Systems, Inc",
			},
			wantCategory: "switch",
			wantRuleID:   "cisco_catalyst",
			wantIcon:     "mdi-switch",
		},
		{
			name: "FortiGate Firewall",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.12356.101.1.600",
				SysDescr:    "FortiGate-60F v6.4.5,build1828,210217",
				Vendor:      "Fortinet, Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "fortinet_fortigate",
			wantIcon:     "mdi-security",
		},
		{
			name: "Synology DiskStation",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.6574.1",
				SysDescr:    "Linux DS920+ 4.4.180+ #42962 SMP",
				HTTPTitle:   "Synology DiskStation - DSM 7.1",
				Vendor:      "Synology Incorporated",
			},
			wantCategory: "appliance",
			wantRuleID:   "synology_nas",
			wantIcon:     "mdi-nas",
		},
		{
			name: "APC Smart-UPS",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.318.1.3.27",
				SysDescr:    "APC Web/SNMP Management Card (Smart-UPS RT 3000 XL)",
				Vendor:      "American Power Conversion Corp",
			},
			wantCategory: "appliance",
			wantRuleID:   "apc_ups",
			wantIcon:     "mdi-battery-charging",
		},
		{
			name: "Canon Printer by HTTP and Vendor",
			input: DetectInput{
				HTTPTitle: "Canon LBP841C Remote UI",
				Vendor:    "CANON INC.",
			},
			wantCategory: "appliance",
			wantRuleID:   "printer_canon",
			wantIcon:     "printer",
		},
		{
			name: "Linux Server (Net-SNMP)",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.8072.3.2.10",
				SysDescr:    "Linux ubuntu22 5.15.0-76-generic #83-Ubuntu SMP",
			},
			wantCategory: "server_linux",
			wantRuleID:   "linux_server",
			wantIcon:     "mdi-linux",
		},
		{
			name: "Windows Server 2022",
			input: DetectInput{
				SysObjectID: ".1.3.6.1.4.1.311.1.1.3.1.2",
				SysDescr:    "Hardware: Intel64 Family 6 - Software: Windows Version 10.0 (Build 20348 Multiprocessor Free)",
				Vendor:      "Microsoft Corporation",
			},
			wantCategory: "server_windows",
			wantRuleID:   "windows_server",
			wantIcon:     "mdi-microsoft-windows",
		},
		{
			name: "Unknown node",
			input: DetectInput{
				SysObjectID: "",
				SysDescr:    "",
			},
			wantCategory: "unknown",
			wantRuleID:   "unknown",
			wantIcon:     "desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := DetectNode(&tt.input)
			if res.Category != tt.wantCategory {
				t.Errorf("DetectNode().Category = %q, want %q", res.Category, tt.wantCategory)
			}
			if res.RuleID != tt.wantRuleID {
				t.Errorf("DetectNode().RuleID = %q, want %q", res.RuleID, tt.wantRuleID)
			}
			if res.Icon != tt.wantIcon {
				t.Errorf("DetectNode().Icon = %q, want %q", res.Icon, tt.wantIcon)
			}
			if tt.wantRuleID == "yamaha_rtx" {
				if len(res.SensorPollings) < 3 {
					t.Errorf("Yamaha RTX should have sensor pollings, got %d", len(res.SensorPollings))
				}
				if res.SensorPollings[0].Params != "yamahaRTCPUUtil5sec.0" {
					t.Errorf("Yamaha CPU param should be symbol 'yamahaRTCPUUtil5sec.0', got %q", res.SensorPollings[0].Params)
				}
			}
		})
	}
}
