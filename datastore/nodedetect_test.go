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
			wantIcon:     "wifi",
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
			name: "Synology DiskStation with Net-SNMP linux SysObjectID (DiskStationYMI 192.168.1.203)",
			input: DetectInput{
				IP:          "192.168.1.203",
				Name:        "DiskStationYMI",
				SysObjectID: ".1.3.6.1.4.1.8072.3.2.10",
				SysDescr:    "Linux DiskStationYMI 4.4.180+ #42962 SMP Tue Oct 18 15:02:18 CST 2022 x86_64",
				Vendor:      "Synology Incorporated",
			},
			wantCategory: "appliance",
			wantRuleID:   "synology_nas",
			wantIcon:     "mdi-nas",
		},
		{
			name: "BUFFALO TeraStation with Net-SNMP",
			input: DetectInput{
				IP:          "192.168.1.204",
				SysObjectID: ".1.3.6.1.4.1.8072.3.2.10",
				SysDescr:    "Linux TeraStation 4.14.79 #1 SMP armv7l",
				Vendor:      "BUFFALO INC.",
			},
			wantCategory: "appliance",
			wantRuleID:   "buffalo_nas",
			wantIcon:     "mdi-nas",
		},
		{
			name: "Raspberry Pi with Net-SNMP linux SysObjectID",
			input: DetectInput{
				IP:          "192.168.1.50",
				SysObjectID: ".1.3.6.1.4.1.8072.3.2.10",
				SysDescr:    "Linux raspberrypi 6.1.21-v8+ #1642 SMP Debian",
				Vendor:      "Raspberry Pi Trading Ltd",
			},
			wantCategory: "pc_mobile",
			wantRuleID:   "raspberry_pi",
			wantIcon:     "mdi-raspberry-pi",
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
			name: "Windows 11 Client PC (YMIRYZ)",
			input: DetectInput{
				IP:          "192.168.1.220",
				SysObjectID: ".1.3.6.1.4.1.311.1.1.3.1.1",
				SysDescr:    "Hardware: Intel64 Family 6 - Software: Windows Version 10.0 (Build 22631 Multiprocessor Free)",
				Vendor:      "Microsoft Corporation",
			},
			wantCategory: "pc_mobile",
			wantRuleID:   "windows_client",
			wantIcon:     "mdi-microsoft-windows",
		},
		{
			name: "NTT Home Gateway (Oki Electric, 192.168.1.1 Default Gateway)",
			input: DetectInput{
				IP:        "192.168.1.1",
				Vendor:    "Oki Electric Industry Co., Ltd.",
				IsGateway: true,
			},
			wantCategory: "router",
			wantRuleID:   "ntt_hgw",
			wantIcon:     "router",
		},
		{
			name: "BUFFALO AirStation Router by HTTP Title and Vendor",
			input: DetectInput{
				IP:        "192.168.11.1",
				HTTPTitle: "AirStation WSR-3200AX4S",
				Vendor:    "BUFFALO.INC",
			},
			wantCategory: "wifi",
			wantRuleID:   "broadband_router_buffalo",
			wantIcon:     "wifi",
		},
		{
			name: "Generic Default Gateway Router without specific vendor",
			input: DetectInput{
				IP:        "192.168.1.254",
				IsGateway: true,
			},
			wantCategory: "router",
			wantRuleID:   "default_gateway_router",
			wantIcon:     "router",
		},
		{
			name: "I-O DATA TV Tuner HVTR-BCTZ2 by HTTP Body and Vendor",
			input: DetectInput{
				IP:       "192.168.1.3",
				HTTPBody: "<html><head><title>Network Device</title></head><body>Model: HVTR-BCTZ2 Version 1.00</body></html>",
				Vendor:   "I-O DATA DEVICE, INC.",
			},
			wantCategory: "appliance",
			wantRuleID:   "tv_tuner_iodata",
			wantIcon:     "mdi-television",
		},
		{
			name: "I-O DATA TV Tuner HVTR-BCTZ without title (HTTP body only)",
			input: DetectInput{
				IP:       "192.168.1.3",
				HTTPBody: "device_name=HVTR-BCTZ&status=ok",
				Vendor:   "I-O DATA DEVICE, INC.",
			},
			wantCategory: "appliance",
			wantRuleID:   "tv_tuner_iodata",
			wantIcon:     "mdi-television",
		},
		{
			name: "TP-Link TL-SG108E Easy Smart Switch",
			input: DetectInput{
				IP:       "192.168.1.5",
				HTTPBody: "<html><head><title>TL-SG108E</title></head><body>Easy Smart Switch</body></html>",
				Vendor:   "TP-LINK TECHNOLOGIES CO.,LTD.",
			},
			wantCategory: "switch",
			wantRuleID:   "tplink_switch",
			wantIcon:     "mdi-switch",
		},
		{
			name: "Amazon Fire TV Stick by MAC Vendor only (192.168.1.17 DC:91:BF)",
			input: DetectInput{
				IP:     "192.168.1.17",
				Vendor: "Amazon Technologies Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "amazon_firetv",
			wantIcon:     "mdi-amazon",
		},
		{
			name: "Amazon Fire TV Stick with HostName and Vendor",
			input: DetectInput{
				IP:       "192.168.1.17",
				HostName: "firetv-stick-living.local",
				Vendor:   "Amazon Technologies Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "amazon_firetv",
			wantIcon:     "mdi-amazon",
		},
		{
			name: "Amazon Echo Dot with HostName",
			input: DetectInput{
				IP:       "192.168.1.18",
				HostName: "echodot-kitchen.local",
				Vendor:   "Amazon Technologies Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "amazon_echo",
			wantIcon:     "mdi-amazon",
		},
		{
			name: "Google Chromecast with HostName and Vendor",
			input: DetectInput{
				IP:       "192.168.1.19",
				HostName: "Chromecast-Living",
				Vendor:   "Google, Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "google_device",
			wantIcon:     "mdi-google",
		},
		{
			name: "Apple TV with HostName and Apple Vendor",
			input: DetectInput{
				IP:       "192.168.1.20",
				HostName: "Apple-TV-BedRoom",
				Vendor:   "Apple, Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "apple_tv",
			wantIcon:     "mdi-apple",
		},
		{
			name: "ESP32 IoT Board by Vendor only (192.168.1.6 64:B7:08)",
			input: DetectInput{
				IP:     "192.168.1.6",
				Vendor: "Espressif Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "esp32_iot",
			wantIcon:     "mdi-developer-board",
		},
		{
			name: "ESP32 with HostName (m5stack)",
			input: DetectInput{
				IP:       "192.168.1.7",
				HostName: "m5stack-core2.local",
				Vendor:   "Espressif Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "esp32_iot",
			wantIcon:     "mdi-developer-board",
		},
		{
			name: "Arduino IoT Board",
			input: DetectInput{
				IP:     "192.168.1.8",
				Vendor: "Arduino SA",
			},
			wantCategory: "appliance",
			wantRuleID:   "arduino_iot",
			wantIcon:     "mdi-developer-board",
		},
		{
			name: "Network Receiver by D&M Holdings Vendor only (192.168.1.32 00:05:CD)",
			input: DetectInput{
				IP:     "192.168.1.32",
				Vendor: "D&M Holdings Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "network_receiver",
			wantIcon:     "mdi-amplifier",
		},
		{
			name: "Denon AV Receiver with HostName and Vendor",
			input: DetectInput{
				IP:       "192.168.1.33",
				HostName: "Denon-AVR-X2700H.local",
				Vendor:   "D&M Holdings Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "network_receiver",
			wantIcon:     "mdi-amplifier",
		},
		{
			name: "Marantz Network Receiver with HTTP Title",
			input: DetectInput{
				IP:        "192.168.1.34",
				HTTPTitle: "Marantz NR1200 Network Receiver",
				Vendor:    "D&M Holdings Inc.",
			},
			wantCategory: "appliance",
			wantRuleID:   "network_receiver",
			wantIcon:     "mdi-amplifier",
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
