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

var version string
var logCh = make(chan *datastore.LogEnt, 5000)
var flushReqCh = make(chan chan struct{})
var paused = false
var pauseMu sync.RWMutex

// Pause pauses log reception and flushes buffered logs.
func Pause() {
	pauseMu.Lock()
	paused = true
	pauseMu.Unlock()
	Flush()
	log.Println("logger paused")
}

// Resume resumes log reception.
func Resume() {
	pauseMu.Lock()
	defer pauseMu.Unlock()
	paused = false
	log.Println("logger resumed")
}

// IsPaused returns whether logger is paused.
func IsPaused() bool {
	pauseMu.RLock()
	defer pauseMu.RUnlock()
	return paused
}

// Flush flushes buffered logs to datastore.
func Flush() {
	req := make(chan struct{})
	select {
	case flushReqCh <- req:
		<-req
	case <-time.After(2 * time.Second):
	}
}

func Start(ctx context.Context, v string, wg *sync.WaitGroup) error {
	version = v
	wg.Add(1)
	go logger(ctx, wg)
	return nil
}

func logger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	var syslogdRunning = false
	var trapdRunning = false
	var arpWatchRunning = false
	var sshdRunning = false
	var netflowdRunning = false
	var sflowdRunning = false
	var tcpdRunning = false
	var oteldRunning = false
	var mqttdRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopArpWatch chan bool
	var stopSshd chan bool
	var stopNetflowd chan bool
	var stopSflowd chan bool
	var stopTcpd chan bool
	var stopOteld chan bool
	var stopMqttd chan bool
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
				if sshdRunning {
					close(stopSshd)
				}
				if netflowdRunning {
					close(stopNetflowd)
				}
				if sflowdRunning {
					close(stopSflowd)
				}
				if tcpdRunning {
					close(stopTcpd)
				}
				if oteldRunning {
					close(stopOteld)
				}
				if mqttdRunning {
					close(stopMqttd)
				}
				if len(logBuffer) > 0 {
					datastore.SaveLogBuffer(logBuffer)
				}
				log.Printf("stop logger")
				return
			}
		case req := <-flushReqCh:
			if len(logBuffer) > 0 {
				datastore.SaveLogBuffer(logBuffer)
				logBuffer = []*datastore.LogEnt{}
			}
			close(req)
		case l := <-logCh:
			if IsPaused() {
				continue
			}
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
				go syslogd(stopSyslogd)
			} else if !datastore.MapConf.EnableSyslogd && syslogdRunning {
				close(stopSyslogd)
				syslogdRunning = false
			}
			if datastore.MapConf.EnableTrapd && !trapdRunning {
				stopTrapd = make(chan bool)
				trapdRunning = true
				go snmptrapd(stopTrapd)
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
			if datastore.MapConf.EnableSshd && !sshdRunning {
				stopSshd = make(chan bool)
				sshdRunning = true
				go sshd(stopSshd)
			} else if !datastore.MapConf.EnableSshd && sshdRunning {
				close(stopSshd)
				sshdRunning = false
			}
			if datastore.MapConf.EnableNetflowd && !netflowdRunning {
				stopNetflowd = make(chan bool)
				netflowdRunning = true
				go netflowd(stopNetflowd)
			} else if !datastore.MapConf.EnableNetflowd && netflowdRunning {
				close(stopNetflowd)
				netflowdRunning = false
			}
			if datastore.MapConf.EnableSFlowd && !sflowdRunning {
				stopSflowd = make(chan bool)
				sflowdRunning = true
				go sflowd(stopSflowd)
			} else if !datastore.MapConf.EnableSFlowd && sflowdRunning {
				close(stopSflowd)
				sflowdRunning = false
			}
			if datastore.MapConf.EnableTcpd && !tcpdRunning {
				stopTcpd = make(chan bool)
				tcpdRunning = true
				go tcpd(stopTcpd)
			} else if !datastore.MapConf.EnableTcpd && tcpdRunning {
				close(stopTcpd)
				tcpdRunning = false
			}
			if datastore.MapConf.EnableOTel && !oteldRunning {
				stopOteld = make(chan bool)
				oteldRunning = true
				go oteld(stopOteld)
			} else if !datastore.MapConf.EnableOTel && oteldRunning {
				close(stopOteld)
				oteldRunning = false
			}
			if datastore.MapConf.EnableMqtt && !mqttdRunning {
				stopMqttd = make(chan bool)
				mqttdRunning = true
				go mqttd(stopMqttd)
			} else if !datastore.MapConf.EnableMqtt && mqttdRunning {
				close(stopMqttd)
				mqttdRunning = false
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
