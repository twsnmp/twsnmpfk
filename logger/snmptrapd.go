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

func snmptrapd(stopCh chan bool) {
	log.Printf("start snmp trapd")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Start snmptrapd"),
	})
	tl := gosnmp.NewTrapListener()
	tl.Params = &gosnmp.GoSNMP{}
	tl.Params.Port = uint16(datastore.TrapPort)
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
	case "v3sha256aes128":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	case "v3sha512aes256":
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		tl.Params.MsgFlags = gosnmp.AuthPriv
		tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA512,
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
		if ReportEngine != nil {
			ReportEngine.ProcessTrap(record)
		}
		if u != nil {
			srcIP := u.IP.String()
			enterprise := fmt.Sprintf("%v", record["Enterprise"])
			variables := fmt.Sprintf("%v", record["Variables"])
			if strings.Contains(strings.ToLower(enterprise), "twagentneo") || strings.Contains(variables, "twagentNeo") {
				datastore.UpdateSensor(srcIP, "twagentneo", enterprise, 1)
			} else {
				datastore.UpdateSensor(srcIP, "snmptrap", enterprise, 1)
			}
		}
		js, err := json.Marshal(record)
		if err == nil {
			logCh <- &datastore.LogEnt{
				Time: time.Now().UnixNano(),
				Type: "trap",
				Log:  string(js),
			}
		}
	}
	testConn, err := net.ListenPacket("udp", fmt.Sprintf("0.0.0.0:%d", datastore.TrapPort))
	if err != nil {
		log.Printf("snmp trap bind error: %v", err)
		SetReceiverStatus("trap", ReceiverStatus{
			Status: "error",
			Port:   fmt.Sprintf("UDP :%d (v1/v2c/v3)", datastore.TrapPort),
			Error:  err.Error(),
		})
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "warn",
			Event: fmt.Sprintf(i18n.Trans("SNMP trap bind error: %v"), err),
		})
		<-stopCh
		log.Printf("stop snmp trapd")
		SetReceiverStatus("trap", ReceiverStatus{
			Status: "stopped",
			Port:   fmt.Sprintf("UDP :%d (v1/v2c/v3)", datastore.TrapPort),
		})
		return
	}
	_ = testConn.Close()

	SetReceiverStatus("trap", ReceiverStatus{
		Status: "running",
		Port:   fmt.Sprintf("UDP :%d (v1/v2c/v3)", datastore.TrapPort),
	})

	defer tl.Close()
	go func() {
		if err := tl.Listen(fmt.Sprintf("0.0.0.0:%d", datastore.TrapPort)); err != nil {
			log.Printf("snmp trap listen err=%v", err)
			SetReceiverStatus("trap", ReceiverStatus{
				Status: "error",
				Port:   fmt.Sprintf("UDP :%d (v1/v2c/v3)", datastore.TrapPort),
				Error:  err.Error(),
			})
		}
		log.Printf("close snmp trapd")
	}()
	<-stopCh
	log.Printf("stop snmp trapd")
	SetReceiverStatus("trap", ReceiverStatus{
		Status: "stopped",
		Port:   fmt.Sprintf("UDP :%d (v1/v2c/v3)", datastore.TrapPort),
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Stop snmptrapd"),
	})
}
