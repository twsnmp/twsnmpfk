package logger

/*
  syslog,tarpをログに記録する
*/

import (
	"encoding/json"
	"log"
	"strings"

	"fmt"
	"net"
	"time"

	gosnmp "github.com/gosnmp/gosnmp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func snmptrapd(stopCh chan bool, port int) {
	log.Printf("start snmp trapd")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Start snmptrapd"),
	})
	tl := gosnmp.NewTrapListener()
	tl.Params = &gosnmp.GoSNMP{}
	tl.Params.Port = uint16(port)
	switch datastore.MapConf.SnmpMode {
	case "v3auth":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthNoPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
		}
	case "v3authpriv":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	case "v3authprivex":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	default:
		// SNMPv2c
		tl.Params.Version = gosnmp.Version2c
		tl.Params.Community = datastore.MapConf.Community
	}
	tl.OnNewTrap = func(s *gosnmp.SnmpPacket, u *net.UDPAddr) {
		var record = make(map[string]interface{})
		record["FromAddress"] = u.String()
		record["Timestamp"] = s.Timestamp
		record["Enterprise"] = datastore.MIBDB.OIDToName(s.Enterprise)
		record["GenericTrap"] = s.GenericTrap
		record["SpecificTrap"] = s.SpecificTrap
		record["Variables"] = ""
		vbs := ""
		for _, vb := range s.Variables {
			key := datastore.MIBDB.OIDToName(vb.Name)
			val := datastore.GetMIBValueString(key, &vb, false)
			vbs += fmt.Sprintf("%s=%s\n", key, val)
			if strings.HasPrefix(key, "sysName") {
				n := datastore.FindNodeFromName(val)
				if n != nil {
					record["FromAddress"] = fmt.Sprintf("%s(%s) via %s", n.IP, n.Name, u.IP.String())
				}
			}
		}
		record["Variables"] = vbs
		js, err := json.Marshal(record)
		if err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "trap",
				Log:  string(js),
			}
		}
	}
	defer tl.Close()
	go func() {
		if err := tl.Listen(fmt.Sprintf("0.0.0.0:%d", port)); err != nil {
			log.Printf("snmp trap listen err=%v", err)
		}
		log.Printf("close snmp trapd")
	}()
	<-stopCh
	log.Printf("stop snmp trapd")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Stop snmptrapd"),
	})
}

func getSnmpString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	}
	return fmt.Sprintf("%v", i)
}

func getTimeTickStr(t int64) string {
	ft := float64(t) / 100
	if ft > 3600*24 {
		return fmt.Sprintf(i18n.Trans("%.2fDays(%d)"), ft/(3600*24), t)
	} else if ft > 3600 {
		return fmt.Sprintf(i18n.Trans("%.2fHours(%d)"), ft/(3600), t)
	}
	return fmt.Sprintf(i18n.Trans("%.2fSec(%d)"), ft, t)
}
