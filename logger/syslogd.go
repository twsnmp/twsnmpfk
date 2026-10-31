package logger

/*
  syslogをログに記録する
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func syslogd(stopCh chan bool) {
	log.Printf("start syslogd")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Start syslogd"),
	})
	syslogCh := make(syslog.LogPartsChannel, 2000)
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(syslog.NewChannelHandler(syslogCh))
	_ = server.ListenUDP(fmt.Sprintf("0.0.0.0:%d", datastore.SyslogPort))
	_ = server.ListenTCP(fmt.Sprintf("0.0.0.0:%d", datastore.SyslogPort))
	_ = server.Boot()
	for {
		select {
		case <-stopCh:
			{
				log.Printf("stop syslogd")
				datastore.AddEventLog(&datastore.EventLogEnt{
					Type:  "system",
					Level: "info",
					Event: i18n.Trans("Stop syslogd"),
				})
				_ = server.Kill()
				return
			}
		case sl := <-syslogCh:
			{
				if ReportEngine != nil {
					ReportEngine.ProcessSyslog(sl)
				}
				handleSyslogSensor(sl)
				if datastore.AutoCharCode {
					if c, ok := sl["content"].(string); ok {
						sl["content"] = datastore.CheckCharCode(c)
					}
				}
				s, err := json.Marshal(sl)
				if err == nil {
					logCh <- &datastore.LogEnt{
						Time: time.Now().UnixNano(),
						Type: "syslog",
						Log:  string(s),
					}
				} else {
					log.Printf("syslogd err=%v", err)
				}
			}
		}
	}
}

func handleSyslogSensor(sl map[string]interface{}) {
	host, _ := sl["hostname"].(string)
	if host == "" {
		if client, ok := sl["client"].(string); ok && client != "" {
			if h, _, err := net.SplitHostPort(client); err == nil {
				host = h
			} else {
				host = client
			}
		}
	}
	if host == "" {
		return
	}
	content, _ := sl["content"].(string)
	if content == "" {
		content, _ = sl["message"].(string)
	}
	tag, _ := sl["tag"].(string)
	if tag == "" {
		tag, _ = sl["app_name"].(string)
	}

	kvs := parseTWLogKVs(content)
	tType := kvs["type"]
	param := kvs["param"]

	sensorType := "syslog"
	cleanTag := strings.TrimSuffix(strings.TrimSpace(tag), ":")
	if idx := strings.Index(cleanTag, "["); idx >= 0 {
		cleanTag = strings.TrimSpace(cleanTag[:idx])
	}
	if cleanTag != "" {
		switch strings.ToLower(cleanTag) {
		case "twwifiscan":
			sensorType = "twWifiScan"
		case "twbluescan":
			sensorType = "twBlueScan"
		case "twpcap":
			sensorType = "twpcap"
		case "twwinlog":
			sensorType = "twwinlog"
		case "twsdrpower":
			sensorType = "twSdrPower"
		case "twagentneo", "twagentneo-fim", "twagentneo-sca", "twagentneo-resource":
			sensorType = "twagentneo"
			if param == "" {
				param = cleanTag
			}
		}
	} else if strings.Contains(strings.ToLower(content), "twpcap") {
		sensorType = "twpcap"
	}

	switch tType {
	case "Stats":
		count, _ := strconv.ParseInt(kvs["count"], 10, 64)
		send, _ := strconv.ParseInt(kvs["send"], 10, 64)
		total, _ := strconv.ParseInt(kvs["total"], 10, 64)
		ps, _ := strconv.ParseFloat(kvs["ps"], 64)
		datastore.CheckSensorStats(host, sensorType, param, count, send, total, ps)
	case "Monitor":
		cpu, _ := strconv.ParseFloat(kvs["cpu"], 64)
		mem, _ := strconv.ParseFloat(kvs["mem"], 64)
		load, _ := strconv.ParseFloat(kvs["load"], 64)
		txSpeed, _ := strconv.ParseFloat(kvs["txSpeed"], 64)
		rxSpeed, _ := strconv.ParseFloat(kvs["rxSpeed"], 64)
		sent, _ := strconv.ParseInt(kvs["sent"], 10, 64)
		recv, _ := strconv.ParseInt(kvs["recv"], 10, 64)
		proc, _ := strconv.ParseInt(kvs["process"], 10, 64)
		datastore.CheckSensorMonitor(host, sensorType, param, cpu, mem, load, txSpeed, rxSpeed, sent, recv, proc)
	default:
		datastore.UpdateSensor(host, sensorType, param, 1)
	}
}

func parseTWLogKVs(content string) map[string]string {
	m := make(map[string]string)
	if idx := strings.Index(content, "type="); idx >= 0 {
		content = content[idx:]
	}
	parts := strings.Split(content, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return m
}
