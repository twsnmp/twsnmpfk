package logger

/*
  syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"bytes"
	"encoding/json"
	"log"

	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/ipfix"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/netflow9"
	"github.com/tehmaze/netflow/read"
	"github.com/tehmaze/netflow/session"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func netflowd(stopCh chan bool) {
	var readSize = 2 << 16
	var addr *net.UDPAddr
	var err error
	log.Println("start netflowd ")
	port := fmt.Sprintf(":%d", netflowPort)
	if addr, err = net.ResolveUDPAddr("udp", port); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	defer server.Close()
	if err = server.SetReadBuffer(readSize); err != nil {
		log.Printf("netflowd err=%v", err)
		return
	}
	decoders := make(map[string]*netflow.Decoder)
	buf := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			{
				log.Printf("stop netflowd")
				return
			}
		default:
			{
				_ = server.SetReadDeadline(time.Now().Add(time.Second * 2))
				var remote *net.UDPAddr
				var octets int
				if octets, remote, err = server.ReadFromUDP(buf); err != nil {
					if !strings.Contains(err.Error(), "timeout") {
						log.Printf("netflowd err=%v", err)
					}
					continue
				}
				d, found := decoders[remote.String()]
				if !found {
					s := session.New()
					d = netflow.NewDecoder(s)
					decoders[remote.String()] = d
				}
				{
					defer func() {
						if r := recover(); r != nil {
							log.Printf("recover netflow err=%v", r)
						}
					}()
					m, err := d.Read(bytes.NewBuffer(buf[:octets]))
					if err != nil {
						log.Printf("netflowd err=%v", err)
						continue
					}
					switch p := m.(type) {
					case *netflow5.Packet:
						logNetflow(p)
					case *netflow9.Packet:
						logNetflow9(p)
					case *ipfix.Message:
						logIPFIX(p)
					}
				}
			}
		}
	}
}

func getStringFromIPFIXFieldValue(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case net.IPAddr:
		return v.String()
	case net.IP:
		return v.String()
	case net.HardwareAddr:
		return v.String()
	}
	return ""
}

func getInt64FromIPFIXFieldValue(i interface{}) int64 {
	switch v := i.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint32:
		return int64(v)
	case int16:
		return int64(v)
	case uint16:
		return int64(v)
	case int8:
		return int64(v)
	case uint8:
		return int64(v)
	case int64:
		return v
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	case time.Time:
		return v.UnixNano()
	}
	return 0
}

func getIntFromIPFIXFieldValue(i interface{}) int {
	switch v := i.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case uint32:
		return int(v)
	case int16:
		return int(v)
	case uint16:
		return int(v)
	case int8:
		return int(v)
	case uint8:
		return int(v)
	case int64:
		return int(v)
	case uint64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

func logIPFIX(p *ipfix.Message) {
	for _, ds := range p.DataSets {
		if ds.Records == nil {
			continue
		}
		for _, dr := range ds.Records {
			var record = datastore.NetFlowEnt{}
			first := 0
			last := 0
			icmpType := 0
			for _, f := range dr.Fields {
				if f.Translated != nil {
					switch f.Translated.Name {
					case "sourceIPv4Address", "sourceIPv6Address":
						record.SrcAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "sourceMacAddress", "postSourceMacAddress":
						record.SrcMAC = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "sourceTransportPort":
						record.SrcPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "destinationIPv4Address", "destinationIPv6Address":
						record.DstAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "destinationMacAddress", "postDestinationMacAddress":
						record.DstMAC = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "destinationTransportPort":
						record.DstPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "octetDeltaCount":
						record.Bytes = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "packetDeltaCount":
						record.Packets = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowStartSysUpTime":
						first = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowEndSysUpTime":
						last = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowStartMilliseconds", "flowStartSeconds", "flowStartNanoSeconds":
						record.Start = getInt64FromIPFIXFieldValue(f.Translated.Value)
					case "flowEndMilliseconds", "flowEndSeconds", "flowEndNanoSeconds":
						record.End = getInt64FromIPFIXFieldValue(f.Translated.Value)
					case "tcpControlBits":
						record.TCPFlags = read.TCPFlags(uint8(getIntFromIPFIXFieldValue(f.Translated.Value)))
					case "protocolIdentifier":
						pi := uint8(getIntFromIPFIXFieldValue(f.Translated.Value))
						switch pi {
						case 1:
							record.Protocol = "icmp"
						case 2:
							record.Protocol = "igmp"
						case 6:
							record.Protocol = "tcp"
						case 8:
							record.Protocol = "egp"
						case 17:
							record.Protocol = "udp"
						case 58:
							record.Protocol = "ipv6-icmp"
						default:
							record.Protocol = read.Protocol(pi)
							if record.Protocol == "" {
								record.Protocol = fmt.Sprintf("%d", pi)
							}
						}
					case "ipClassOfService":
						record.ToS = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "icmpTypeCodeIPv6", "icmpTypeCodeIPv4":
						icmpType = getIntFromIPFIXFieldValue(f.Translated.Value)
					}
				}
			}
			if last > 0 {
				record.Dur = float64(last-first) / 100
			} else if record.Start > 0 {
				record.Dur = float64((record.End - record.Start)) / (1000 * 1000 * 1000)
			}
			record.SrcLoc = datastore.GetLoc(record.SrcAddr)
			record.DstLoc = datastore.GetLoc(record.DstAddr)
			if icmpType > 0 && strings.Contains(record.Protocol, "icmp") {
				record.SrcPort = icmpType / 256
				record.DstPort = icmpType % 256
			}
			s, err := json.Marshal(record)
			if err != nil {
				continue
			}
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "netflow",
				Log:  string(s),
			}
		}
	}
}

func logNetflow(p *netflow5.Packet) {
	for _, r := range p.Records {
		record := datastore.NetFlowEnt{
			SrcAddr:  r.SrcAddr.String(),
			SrcPort:  int(r.SrcPort),
			DstAddr:  r.DstAddr.String(),
			DstPort:  int(r.DstPort),
			Bytes:    int(r.Bytes),
			Packets:  int(r.Packets),
			TCPFlags: read.TCPFlags(r.TCPFlags),
			Protocol: read.Protocol(r.Protocol),
			ToS:      int(r.ToS),
			Dur:      float64(r.Last-r.First) / 100.0,
		}
		record.SrcLoc = datastore.GetLoc(record.SrcAddr)
		record.DstLoc = datastore.GetLoc(record.DstAddr)
		s, err := json.Marshal(record)
		if err != nil {
			fmt.Println(err)
			continue
		}
		logCh <- &datastore.LogEnt{
			Time: time.Now().UnixNano(),
			Type: "netflow",
			Log:  string(s),
		}
	}
}

func logNetflow9(p *netflow9.Packet) {
	for _, ds := range p.DataFlowSets {
		if ds.Records == nil {
			continue
		}
		for _, dr := range ds.Records {
			var record = datastore.NetFlowEnt{}
			first := 0
			last := 0
			icmpType := 0
			for _, f := range dr.Fields {
				if f.Translated != nil {
					switch f.Translated.Name {
					case "sourceIPv4Address", "sourceIPv6Address":
						record.SrcAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "sourceMacAddress", "postSourceMacAddress":
						record.SrcMAC = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "sourceTransportPort":
						record.SrcPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "destinationIPv4Address", "destinationIPv6Address":
						record.DstAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "destinationMacAddress", "postDestinationMacAddress":
						record.DstMAC = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "destinationTransportPort":
						record.DstPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "octetDeltaCount":
						record.Bytes = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "packetDeltaCount":
						record.Packets = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowStartSysUpTime":
						first = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowEndSysUpTime":
						last = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "flowStartMilliseconds", "flowStartSeconds", "flowStartNanoSeconds":
						record.Start = getInt64FromIPFIXFieldValue(f.Translated.Value)
					case "flowEndMilliseconds", "flowEndSeconds", "flowEndNanoSeconds":
						record.End = getInt64FromIPFIXFieldValue(f.Translated.Value)
					case "tcpControlBits":
						record.TCPFlags = read.TCPFlags(uint8(getIntFromIPFIXFieldValue(f.Translated.Value)))
					case "protocolIdentifier":
						pi := uint8(getIntFromIPFIXFieldValue(f.Translated.Value))
						switch pi {
						case 1:
							record.Protocol = "icmp"
						case 2:
							record.Protocol = "igmp"
						case 6:
							record.Protocol = "tcp"
						case 8:
							record.Protocol = "egp"
						case 17:
							record.Protocol = "udp"
						case 58:
							record.Protocol = "ipv6-icmp"
						default:
							record.Protocol = read.Protocol(pi)
							if record.Protocol == "" {
								record.Protocol = fmt.Sprintf("%d", pi)
							}
						}
					case "ipClassOfService":
						record.ToS = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "icmpTypeCodeIPv6", "icmpTypeCodeIPv4":
						icmpType = getIntFromIPFIXFieldValue(f.Translated.Value)
					}
				}
			}
			if last > 0 {
				record.Dur = float64(last-first) / 100
			} else if record.Start > 0 {
				record.Dur = float64((record.End - record.Start)) / (1000 * 1000 * 1000)
			}
			record.SrcLoc = datastore.GetLoc(record.SrcAddr)
			record.DstLoc = datastore.GetLoc(record.DstAddr)
			if icmpType > 0 && strings.Contains(record.Protocol, "icmp") {
				record.SrcPort = icmpType / 256
				record.DstPort = icmpType % 256
			}
			s, err := json.Marshal(record)
			if err != nil {
				continue
			}
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "netflow",
				Log:  string(s),
			}
		}
	}
}
