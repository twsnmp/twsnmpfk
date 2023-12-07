package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
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
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(datastore.MapConf.Timeout) * time.Second,
		Retries:            datastore.MapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
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
		value := ""
		switch variable.Type {
		case gosnmp.OctetString:
			mi := datastore.FindMIBInfo(name)
			if mi != nil {
				switch mi.Type {
				case "PhysAddress", "OctetString":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					mac := []string{}
					for _, m := range a {
						mac = append(mac, fmt.Sprintf("%02X", m&0x00ff))
					}
					value = strings.Join(mac, ":")
				case "BITS":
					a, ok := variable.Value.([]uint8)
					if !ok {
						a = []uint8(getMIBStringVal(variable.Value))
					}
					hex := []string{}
					ap := []string{}
					bit := 0
					for _, m := range a {
						hex = append(hex, fmt.Sprintf("%02X", m&0x00ff))
						if !raw && mi.Enum != "" {
							for i := 0; i < 8; i++ {
								if (m & 0x80) == 0x80 {
									if n, ok := mi.EnumMap[bit]; ok {
										ap = append(ap, fmt.Sprintf("%s(%d)", n, bit))
									}
								}
								m <<= 1
								bit++
							}
						}
					}
					value = strings.Join(hex, " ")
					if len(ap) > 0 {
						value += " " + strings.Join(ap, " ")
					}
				case "DisplayString":
					value = getMIBStringVal(variable.Value)
					value = logger.CheckCharCode(value)
				case "DateAndTime":
					value = getDateAndTime(variable.Value)
				default:
					value = getMIBStringVal(variable.Value)
				}
			} else {
				value = getMIBStringVal(variable.Value)
			}
		case gosnmp.ObjectIdentifier:
			value = datastore.MIBDB.OIDToName(getMIBStringVal(variable.Value))
		case gosnmp.TimeTicks:
			t := gosnmp.ToBigInt(variable.Value).Uint64()
			if raw {
				value = fmt.Sprintf("%d", t)
			} else {
				if t > (24 * 3600 * 100) {
					d := t / (24 * 3600 * 100)
					t -= d * (24 * 3600 * 100)
					value = fmt.Sprintf("%d(%d days, %v)", t, d, time.Duration(t*10*uint64(time.Millisecond)))
				} else {
					value = fmt.Sprintf("%d(%v)", t, time.Duration(t*10*uint64(time.Millisecond)))
				}
			}
		default:
			v := int(gosnmp.ToBigInt(variable.Value).Uint64())
			if raw {
				value = fmt.Sprintf("%d", v)
			} else {
				value = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
				mi := datastore.FindMIBInfo(name)
				if mi != nil {
					if mi.Enum != "" {
						if vn, ok := mi.EnumMap[v]; ok {
							value += "(" + vn + ")"
						}
					} else {
						if mi.Hint != "" {
							value = datastore.PrintHintedMIBIntVal(int32(v), mi.Hint, variable.Type != gosnmp.Integer)
						}
						if mi.Units != "" {
							value += " " + mi.Units
						}
					}
				}
			}
		}
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
	if oid == ".0.0" {
		if name == "iso" || name == "org" ||
			name == "dod" || name == "internet" ||
			name == ".1" || name == ".1.3" || name == ".1.3.6" {
			return ".1.3.6.1"
		}
		if matched, _ := regexp.MatchString(`\.[0-9.]+`, name); matched {
			return name
		}
	}
	return oid
}

func getMIBStringVal(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return ""
}

// DISPLAY-HINT "2d-1d-1d,1d:1d:1d.1d,1a1d:1d"
func getDateAndTime(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		if len(v) == 11 {
			return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%02d%c%02d%02d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10])
		} else if len(v) == 8 {
			return fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d.%02d",
				(int(v[0])*256 + int(v[1])), v[2], v[3], v[4], v[5], v[6], v[7])
		}
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("Invalid Date And Time %v", i)
}
