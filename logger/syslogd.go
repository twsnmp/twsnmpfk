package logger

/*
  syslogをログに記録する
*/

import (
	"encoding/json"
	"fmt"
	"log"

	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func syslogd(stopCh chan bool, port int) {
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
	_ = server.ListenUDP(fmt.Sprintf("0.0.0.0:%d", port))
	_ = server.ListenTCP(fmt.Sprintf("0.0.0.0:%d", port))
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
