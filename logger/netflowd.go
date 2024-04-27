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
	}
	return ""
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
			for _, f := range dr.Fields {
				if f.Translated != nil {
					switch f.Translated.Name {
					case "srcAddr":
						record.SrcAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "srcPort":
						record.SrcPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "dstAddr":
						record.DstAddr = getStringFromIPFIXFieldValue(f.Translated.Value)
					case "dstPort":
						record.DstPort = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "bytes":
						record.Bytes = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "packets":
						record.Packets = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "first":
						first = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "last":
						last = getIntFromIPFIXFieldValue(f.Translated.Value)
					case "tcpControlBits":
						record.TCPFlags = read.TCPFlags(uint8(getIntFromIPFIXFieldValue(f.Translated.Value)))
					case "protocolIdentifier":
						record.Protocol = read.Protocol(uint8(getIntFromIPFIXFieldValue(f.Translated.Value)))
					case "tos":
						record.ToS = getIntFromIPFIXFieldValue(f.Translated.Value)
					}
				}
			}
			record.Dur = float64(last-first) / 100
			record.SrcLoc = datastore.GetLoc(record.SrcAddr)
			record.DstLoc = datastore.GetLoc(record.DstAddr)
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
