// Package logger : ログ受信処理
package logger

/*
  syslog,tarpをログに記録する
*/

import (
	"context"
	"log"
	"sync"

	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

var logCh = make(chan *datastore.LogEnt, 5000)
var trapPort = 162
var syslogPort = 514

func Start(ctx context.Context, wg *sync.WaitGroup, _syslogPort, _trapPort int) error {
	syslogPort = _syslogPort
	trapPort = _trapPort
	logCh = make(chan *datastore.LogEnt, 100)
	wg.Add(1)
	go logger(ctx, wg)
	return nil
}

func logger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var syslogdRunning = false
	var trapdRunning = false
	var arpWatchRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopArpWatch chan bool
	log.Println("start logger")
	timer1 := time.NewTicker(time.Second * 10)
	timer2 := time.NewTicker(time.Second * 1)
	logBuffer := []*datastore.LogEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				timer1.Stop()
				timer2.Stop()
				if syslogdRunning {
					close(stopSyslogd)
				}
				if trapdRunning {
					close(stopTrapd)
				}
				if arpWatchRunning {
					close(stopArpWatch)
				}
				if len(logBuffer) > 0 {
					datastore.SaveLogBuffer(logBuffer)
				}
				log.Printf("stop logger")
				return
			}
		case l := <-logCh:
			logBuffer = append(logBuffer, l)
		case <-timer1.C:
			if len(logBuffer) > 0 {
				datastore.SaveLogBuffer(logBuffer)
				logBuffer = []*datastore.LogEnt{}
			}
		case <-timer2.C:
			if datastore.MapConf.EnableSyslogd && !syslogdRunning {
				stopSyslogd = make(chan bool)
				syslogdRunning = true
				go syslogd(stopSyslogd, syslogPort)
			} else if !datastore.MapConf.EnableSyslogd && syslogdRunning {
				close(stopSyslogd)
				syslogdRunning = false
			}
			if datastore.MapConf.EnableTrapd && !trapdRunning {
				stopTrapd = make(chan bool)
				trapdRunning = true
				go snmptrapd(stopTrapd, trapPort)
			} else if !datastore.MapConf.EnableTrapd && trapdRunning {
				close(stopTrapd)
				trapdRunning = false
			}
			if datastore.MapConf.EnableArpWatch && !arpWatchRunning {
				stopArpWatch = make(chan bool)
				arpWatchRunning = true
				go arpWatch(stopArpWatch)
			} else if !datastore.MapConf.EnableArpWatch && arpWatchRunning {
				close(stopArpWatch)
				arpWatchRunning = false
			}
		}
		if datastore.RestartSnmpTrapd && trapdRunning {
			close(stopTrapd)
			datastore.RestartSnmpTrapd = false
			trapdRunning = false
			log.Printf("resatrt trapd")
		}
	}
}
