package logger

/*
  syslogをログに記録する
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Cistern/sflow"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/tehmaze/netflow/read"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func sflowd(stopCh chan bool) {
	log.Printf("start sflowd")
	sv, err := net.ListenPacket("udp", fmt.Sprintf(":%d", datastore.SFlowPort))
	if err != nil {
		log.Printf("sflowd err=%v", err)
		<-stopCh
		return
	}
	defer sv.Close()
	data := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			log.Printf("stop sflowd")
			return
		default:
			sv.SetDeadline(time.Now().Add(time.Second))
			l, ra, err := sv.ReadFrom(data)
			if err != nil {
				continue
			}
			r := bytes.NewReader(data[:l])
			d := sflow.NewDecoder(r)
			dg, err := d.Decode()
			if err != nil {
				log.Printf("sflow decode err=%v", err)
				continue
			}
			raIP := ""
			switch a := ra.(type) {
			case *net.UDPAddr:
				raIP = a.IP.String()
			case *net.TCPAddr:
				raIP = a.IP.String()
			}
			for _, sample := range dg.Samples {
				switch s := sample.(type) {
				case *sflow.CounterSample:
					for _, record := range s.Records {
						switch csr := record.(type) {
						case sflow.HostDiskCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostDiskCounter", raIP, string(s))
						case sflow.HostCPUCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostCPUCounter", raIP, string(s))
						case sflow.HostMemoryCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostMemoryCounter", raIP, string(s))
						case sflow.HostNetCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("HostNetCounter", raIP, string(s))
						case sflow.GenericInterfaceCounters:
							s, err := json.Marshal(record)
							if err != nil {
								log.Println(err)
								continue
							}
							sFlowCounter("GenericInterfaceCounter", raIP, string(s))
						default:
							log.Printf("sflow unknown counter sample %v", csr)
							continue
						}
					}
				case *sflow.FlowSample:
					for _, record := range s.Records {
						switch fsr := record.(type) {
						case sflow.RawPacketFlow:
							rawPacketFlow(&fsr, 0)
						}
					}
				case *sflow.EventDiscardedPacket:
					for _, record := range s.Records {
						switch fsr := record.(type) {
						case sflow.RawPacketFlow:
							rawPacketFlow(&fsr, int(s.Reason))
						}
					}
				}
			}
		}
	}
}

func rawPacketFlow(r *sflow.RawPacketFlow, reason int) {
	var e datastore.SFlowEnt
	packet := gopacket.NewPacket(r.Header, layers.LayerTypeEthernet, gopacket.Default)
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		if eth, ok := ethernetLayer.(*layers.Ethernet); ok {
			e.SrcMAC = eth.SrcMAC.String()
			e.DstMAC = eth.SrcMAC.String()
		}
	}
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ip, ok := ipv4Layer.(*layers.IPv4)
		if !ok {
			return
		}
		e.SrcAddr = ip.SrcIP.String()
		e.DstAddr = ip.DstIP.String()
		e.Bytes = int(ip.Length)
	} else {
		ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
		if ipv6Layer != nil {
			ipv6, ok := ipv6Layer.(*layers.IPv6)
			if !ok {
				return
			}
			e.SrcAddr = ipv6.SrcIP.String()
			e.DstAddr = ipv6.DstIP.String()
			e.Bytes = int(ipv6.Length)
		}
	}
	// UDP
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, ok := udpLayer.(*layers.UDP)
		if !ok {
			return
		}
		e.SrcPort = int(udp.SrcPort)
		e.DstPort = int(udp.DstPort)
		e.Protocol = "udp"
	} else {
		// TCP
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, ok := tcpLayer.(*layers.TCP)
			if !ok {
				return
			}
			e.SrcPort = int(tcp.SrcPort)
			e.DstPort = int(tcp.DstPort)
			var flag uint8
			if tcp.FIN {
				flag |= 0x01
			}
			if tcp.SYN {
				flag |= 0x02
			}
			if tcp.RST {
				flag |= 0x04
			}
			if tcp.PSH {
				flag |= 0x08
			}
			if tcp.ACK {
				flag |= 0x10
			}
			if tcp.URG {
				flag |= 0x20
			}
			if tcp.ECE {
				flag |= 0x40
			}
			if tcp.CWR {
				flag |= 0x80
			}
			e.TCPFlags = read.TCPFlags(flag)
			e.Protocol = "tcp"
		} else {
			icmpV4Layer := packet.Layer(layers.LayerTypeICMPv4)
			if icmpV4Layer != nil {
				icmp, ok := icmpV4Layer.(*layers.ICMPv4)
				if !ok {
					return
				}
				e.Protocol = "icmp"
				e.DstPort = int(icmp.TypeCode)
			} else {
				icmpV6Layer := packet.Layer(layers.LayerTypeICMPv6)
				if icmpV6Layer == nil {
					return
				}
				icmp, ok := icmpV6Layer.(*layers.ICMPv6)
				if !ok {
					return
				}
				e.Protocol = "icmpv6"
				e.DstPort = int(icmp.TypeCode)
			}
		}
	}
	e.Reason = reason
	e.SrcLoc = datastore.GetLoc(e.SrcAddr)
	e.DstLoc = datastore.GetLoc(e.DstAddr)
	s, err := json.Marshal(&e)
	if err != nil {
		log.Println(err)
		return
	}
	logCh <- &datastore.LogEnt{
		Time: time.Now().UnixNano(),
		Type: "sflow",
		Log:  string(s),
	}
}

func sFlowCounter(t, r, d string) {
	record := datastore.SFlowCounterEnt{
		Type:   t,
		Remote: r,
		Data:   d,
	}
	s, err := json.Marshal(record)
	if err != nil {
		log.Println(err)
		return
	}
	logCh <- &datastore.LogEnt{
		Time: time.Now().UnixNano(),
		Type: "sflowCounter",
		Log:  string(s),
	}
}
