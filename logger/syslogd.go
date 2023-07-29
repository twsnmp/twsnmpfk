package logger

/*
  syslogをログに記録する
*/

import (
	"encoding/json"
	"log"

	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func syslogd(stopCh chan bool) {
	syslogCh := make(syslog.LogPartsChannel, 2000)
	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(syslog.NewChannelHandler(syslogCh))
	_ = server.ListenUDP("0.0.0.0:514")
	_ = server.ListenTCP("0.0.0.0:514")
	_ = server.Boot()
	log.Printf("start syslogd")
	for {
		select {
		case <-stopCh:
			{
				log.Printf("stop syslogd")
				_ = server.Kill()
				return
			}
		case sl := <-syslogCh:
			{
				if datastore.MapConf.AutoCharCode {
					if c, ok := sl["content"].(string); ok {
						sl["content"] = CheckCharCode(c)
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
