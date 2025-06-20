package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/twsnmp/twsnmpfk/datastore"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ImportV4Map import TWSNMP v4.x map data
func (a *App) ImportV4Map() bool {
	filter := []wails.FileFilter{}
	filter = append(filter, wails.FileFilter{
		DisplayName: "TWSNMP Map file(*.spm)",
		Pattern:     "*.spm;",
	})
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title:   "TWSNMP Map file",
		Filters: filter,
	})
	if err != nil {
		log.Printf("err=%v", err)
		return false
	}
	if file == "" {
		return true
	}
	return doImportV4Map(file)
}

// V4NodeEnt is the data structure for importing nodes from TWSNMP v4.
type V4NodeEnt struct {
	ID         string
	Name       string
	Descr      string
	Icon       string
	X          int
	Y          int
	IP         string
	MAC        string
	Vendor     string
	SnmpMode   string
	Community  string
	User       string
	Password   string
	URL        string
	AddrMode   string
	AutoAck    bool
	Interfaces []string
	Ping       string
}

type V4LineEnt struct {
	SrcNode string
	SrcIF   string
	DstNode string
	DstIF   string
	Width   int
}

func doImportV4Map(f string) bool {
	var lastNode *V4NodeEnt
	nodes := make(map[string]*V4NodeEnt)
	lines := []*V4LineEnt{}
	defComOrUser := "public"
	defPassword := ""
	defSnmpMode := "v2c"
	r, err := os.Open(f)
	if err != nil {
		log.Printf("doImportV4Map err=%v", err)
		return false
	}
	defer r.Close()
	reader := transform.NewReader(r, japanese.ShiftJIS.NewDecoder())
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		f := strings.Fields(l)
		if len(f) < 1 {
			continue
		}
		switch f[0] {
		case "DEFCOMORUSER":
			if len(f) > 1 {
				defComOrUser = f[1]
			}
		case "DEFPASSWD":
			if len(f) > 1 {
				defPassword = f[1]
			}
		case "DEFSNMPMODE":
			if len(f) > 1 {
				defSnmpMode = convertV4SnmpMode(f[1])
			}
		case "NODE":
			fmt.Printf("%v", f)
			if len(f) > 4 {
				x, _ := strconv.Atoi(f[2])
				y, _ := strconv.Atoi(f[3])
				lastNode = &V4NodeEnt{
					Name: f[1],
					X:    x,
					Y:    y,
					Icon: f[4],
				}
			}
		case "LINE":
			fmt.Printf("%v", f)
			if len(f) > 6 {
				w, _ := strconv.Atoi(f[5])
				lines = append(lines, &V4LineEnt{
					SrcNode: f[1],
					SrcIF:   f[2],
					DstNode: f[3],
					DstIF:   f[4],
					Width:   w,
				})
			}
		case "{":
		case "IPADDR":
			if len(f) > 1 && lastNode != nil {
				lastNode.IP = f[1]
			}
		case "MACADDR":
			if len(f) > 1 && lastNode != nil {
				lastNode.MAC = f[1]
			}
		case "URL":
			if len(f) > 1 && lastNode != nil {
				lastNode.URL = f[1]
			}
		case "SNMP_MODE":
			if len(f) > 1 && lastNode != nil {
				lastNode.SnmpMode = f[1]
			}
		case "COM_OR_USER":
			if len(f) > 1 && lastNode != nil {
				lastNode.User = f[1]
				if lastNode.Community == "" {
					lastNode.Community = f[1]
				}
			}
		case "RCOMMUNITY":
			if len(f) > 1 && lastNode != nil {
				if lastNode.Community == "" {
					lastNode.Community = f[1]
				}
			}
		case "WCOMMUNITY":
			if len(f) > 1 && lastNode != nil {
				if lastNode.Community == "" {
					lastNode.Community = f[1]
				}
			}
		case "PASSWD":
			if len(f) > 1 && lastNode != nil {
				lastNode.Password = f[1]
			}
		case "ADDRMODE":
			if len(f) > 1 && lastNode != nil {
				lastNode.AddrMode = f[1]
			}
		case "VENDOR":
			if len(f) > 1 && lastNode != nil {
				lastNode.Vendor = strings.Join(f[1:], " ")
			}
		case "AUTOACK":
			if lastNode != nil {
				lastNode.AutoAck = true
			}
		case "SYSCONTACT", "SYSLOCATION", "SYSOID":
			if len(f) > 1 && lastNode != nil {
				if lastNode.Descr != "" {
					lastNode.Descr += ","
				}
				lastNode.Descr += f[1]
			}
		case "INTERFACE":
			if len(f) > 6 && lastNode != nil && f[6] != "OFF" {
				lastNode.Interfaces = append(lastNode.Interfaces, f[1])
			}
		case "POLLING":
			if len(f) > 2 && lastNode != nil && f[2] != "OFF" && f[1] == "PING" {
				lastNode.Ping = f[2]
			}
		case "}":
			if lastNode != nil {
				nodes[lastNode.Name] = lastNode
				lastNode = nil
			}
		default:
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
	for _, n := range nodes {
		dn := datastore.FindNodeFromName(n.Name)
		if dn != nil {
			log.Printf("v4import skip dup name=%s", n.Name)
			continue
		}
		if n.Community == "" {
			n.Community = defComOrUser
		}
		if n.User == "" {
			n.User = defComOrUser
		}
		if n.Password == "" {
			n.Password = defPassword
		}
		if n.SnmpMode == "" {
			n.SnmpMode = defSnmpMode
		}
		if err := datastore.AddNode(&datastore.NodeEnt{
			Name:      n.Name,
			IP:        n.IP,
			MAC:       n.MAC,
			X:         n.X,
			Y:         n.Y,
			Icon:      convertV4Icon(n.Icon),
			Descr:     n.Descr,
			URL:       n.URL,
			User:      n.User,
			Password:  n.Password,
			Community: n.Community,
			SnmpMode:  convertV4SnmpMode(n.SnmpMode),
			AddrMode:  convertV4AddrMode(n.AddrMode),
			AutoAck:   n.AutoAck,
		}); err != nil {
			log.Printf("AddNode err=%v", err)
			continue
		}
		dn = datastore.FindNodeFromName(n.Name)
		if dn == nil {
			continue
		}
		n.ID = dn.ID
		if n.Ping != "" {
			if err := datastore.AddPolling(&datastore.PollingEnt{
				NodeID:  n.ID,
				Name:    "PING",
				Type:    "ping",
				Level:   convertV4Level(n.Ping),
				State:   "unknown",
				PollInt: datastore.MapConf.PollInt,
				Timeout: datastore.MapConf.Timeout,
				Retry:   datastore.MapConf.Retry,
			}); err != nil {
				log.Printf("AddPolling err=%v", err)
			}
		}
		for _, i := range n.Interfaces {
			a := strings.Split(i, ".")
			if len(a) != 2 {
				continue
			}
			if err := datastore.AddPolling(&datastore.PollingEnt{
				NodeID:  n.ID,
				Name:    "I/F " + a[1],
				Type:    "snmp",
				Mode:    "ifOperStatus",
				Params:  a[1],
				Level:   "low",
				State:   "unknown",
				PollInt: datastore.MapConf.PollInt,
				Timeout: datastore.MapConf.Timeout,
				Retry:   datastore.MapConf.Retry,
			}); err != nil {
				log.Printf("AddPolling err=%v", err)
			}
		}
	}
	for _, l := range lines {
		sn, ok := nodes[l.SrcNode]
		if !ok {
			continue
		}
		spid := getPollingID(sn.ID, l.SrcIF)
		if spid == "" {
			continue
		}
		dn, ok := nodes[l.DstNode]
		if !ok {
			continue
		}
		dpid := getPollingID(dn.ID, l.DstIF)
		if dpid == "" {
			continue
		}
		datastore.AddLine(&datastore.LineEnt{
			NodeID1:    sn.ID,
			PollingID1: spid,
			NodeID2:    dn.ID,
			PollingID2: dpid,
			Width:      l.Width,
		})
	}
	return true
}

func getPollingID(n, idx string) string {
	bPing := !strings.HasPrefix(idx, "ifOper")
	ret := ""
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == n {
			if bPing && p.Type == "ping" {
				ret = p.ID
				return false
			} else if !bPing && idx == p.Mode+"."+p.Params {
				ret = p.ID
				return false
			}
		}
		return true
	})
	return ret
}

func convertV4Icon(icon string) string {
	switch icon {
	case "PC1_PC":
		return "desktop"
	case "MS_CO":
		return "mdi-microsoft-windows"
	case "NOTE_PC":
		return "laptop"
	}
	if strings.Contains(icon, "_PC") {
		return "desktop"
	}
	if strings.Contains(icon, "_RT") {
		return "router"
	}
	return "mdi-comment-question-outline"
}

func convertV4AddrMode(m string) string {
	switch m {
	case "1":
		return "mac"
	case "2":
		return "host"
	}
	return "ip"
}

func convertV4SnmpMode(m string) string {
	switch m {
	case "1":
		return "v1"
	case "3", "5":
		return "v3auth"
	case "4", "6":
		return "v3authpriv"
	}
	return "v2c"
}

func convertV4Level(l string) string {
	switch l {
	case "HIGH":
		return "high"
	case "LOW":
		return "low"
	case "WARN":
		return "warn"
	}
	return "info"
}
