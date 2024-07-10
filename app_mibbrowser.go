package main

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/twsnmpfk/datastore"
)

// GetMIBTree は MIB Treeを返します。
func (a *App) GetMIBTree() []*datastore.MIBTreeEnt {
	return datastore.MIBTree
}

type MibEnt struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

func (a *App) SnmpWalk(nodeID, name string, raw bool) []*MibEnt {
	var ret []*MibEnt
	n := datastore.GetNode(nodeID)
	if n == nil {
		return ret
	}
	agent := &gosnmp.GoSNMP{
		Target:    n.IP,
		Port:      161,
		Transport: "udp",
		Community: n.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:   datastore.MapConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch n.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        n.Password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 n.User,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: n.Password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        n.Password,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Println(err)
		return ret
	}
	et := time.Now().Unix() + (3 * 60)
	defer agent.Conn.Close()
	err = agent.Walk(nameToOID(name), func(variable gosnmp.SnmpPDU) error {
		if et < time.Now().Unix() {
			return fmt.Errorf("timeout")
		}
		name := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(name, &variable, raw)
		ret = append(ret, &MibEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return ret
}

func nameToOID(name string) string {
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".1" {
		oid = ".1.3"
	}
	if oid == ".0.0" {
		if matched, _ := regexp.MatchString(`\.[0-9.]+`, name); matched {
			return name
		}
	}
	return oid
}
