package logger

/*
	ssh server
*/

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/twsnmp/twsnmpfk/datastore"
	gossh "golang.org/x/crypto/ssh"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func sshd(stopCh chan bool) {
	log.Printf("start sshd")
	signer, err := gossh.ParsePrivateKey([]byte(datastore.GetPrivateKey(true)))
	if err != nil {
		log.Println("sshd", err)
		<-stopCh
		log.Printf("stop sshd err=%v", err)
		return
	}
	sv := ssh.Server{
		Addr:             fmt.Sprintf(":%d", sshdPort),
		Version:          "TWSNNMP FK v1.10.0",
		HostSigners:      []ssh.Signer{signer},
		IdleTimeout:      time.Second * 30,
		MaxTimeout:       time.Minute * 10,
		Handler:          sshdHandler,
		PublicKeyHandler: sshdPublicKeyHandler,
	}
	go func() {
		if err := sv.ListenAndServe(); err != nil {
			log.Printf("sshd err=%v", err)
		}
	}()
	<-stopCh
	log.Printf("stop sshd")
	sv.Shutdown(context.Background())
}

func sshdHandler(s ssh.Session) {
	cmds := s.Command()
	if len(cmds) < 2 {
		log.Println("sshd invalid command")
		io.WriteString(s.Stderr(), "invalid command\r\n")
		return
	}
	cmd := strings.ToLower(cmds[0])
	t := strings.ToLower(cmds[1])
	switch cmd {
	case "get":
		switch t {
		case "syslog", "trap", "arplog", "ipfix", "netflow", "eventlog":
			sshdGetlog(t, s)
			return
		case "node":
			sshdGetNodes(s)
			return
		case "polling":
			sshdGetPollings(s)
			return
		}
	case "put":
		switch t {
		case "syslog":
			sshdPutSyslog(s)
			return
		}
	}
	log.Printf("sshd invalid command %v\r\n", cmds)
}

func sshdPublicKeyHandler(ctx ssh.Context, keyCl ssh.PublicKey) bool {
	pubkeys := datastore.GetSshdPublicKeys()
	for _, pk := range strings.Split(pubkeys, "\n") {
		pk = strings.TrimSpace(pk)
		if keyReg, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pk)); err == nil && ssh.KeysEqual(keyCl, keyReg) {
			return true
		}
	}
	return false
}

// get log
func sshdGetlog(t string, s ssh.Session) {
	st := int64(0)
	et := time.Now().UnixNano()
	count := 1000
	cmds := s.Command()
	if len(cmds) > 3 {
		if v, err := strconv.ParseInt(cmds[2], 10, 64); err == nil {
			st = v
		} else if v, err := time.Parse(time.RFC3339, cmds[2]); err == nil {
			st = v.UnixNano()
		}
		if v, err := strconv.ParseInt(cmds[3], 10, 64); err == nil {
			count = int(v)
		}
	}
	if t == "eventlog" {
		datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
			io.WriteString(s, fmt.Sprintf("%d\t"+`{"level":"%s","type":"%s","node":"%s","event":"%s"}`+"\r\n",
				l.Time, l.Level, l.Type, l.NodeName, l.Event))
			count--
			return count > 0
		})
		return
	}
	datastore.ForEachLogs(st, et, cmds[1], func(l *datastore.LogEnt) bool {
		io.WriteString(s, fmt.Sprintf("%d\t%s\r\n", l.Time, l.Log))
		count--
		return count > 0
	})
}

// put syslog
func sshdPutSyslog(s ssh.Session) {
	client := s.RemoteAddr().String()
	if a := strings.Split(client, ":"); len(a) == 2 {
		client = a[0]
	} else if a := strings.Split(client, "]:"); len(a) == 2 && len(a[0]) > 1 {
		client = a[0][1:]
	}
	facility := 21
	severity := 6
	host := client
	if n := datastore.FindNodeFromIP(client); n != nil {
		host = n.Name
	}
	cmds := s.Command()
	if len(cmds) > 2 {
		if v, err := strconv.ParseInt(cmds[2], 10, 64); err == nil {
			facility = int(v)
		}
		if len(cmds) > 3 {
			if v, err := strconv.ParseInt(cmds[3], 10, 64); err == nil {
				severity = int(v)
			}
			if len(cmds) > 4 {
				host = cmds[4]
			}
		}
	}
	count := 0
	r := bufio.NewScanner(s)
	for r.Scan() {
		l := r.Text()
		a := strings.SplitN(l, "\t", 3)
		if len(a) != 3 {
			continue
		}
		ts, err := strconv.ParseInt(a[0], 10, 64)
		if err != nil {
			log.Printf("sshdPutSyslog err=%v", err)
			continue
		}
		sl := format.LogParts{
			"client":    client,
			"hostname":  host,
			"timestamp": time.Unix(0, ts).Format(time.RFC3339),
			"tag":       a[1],
			"facility":  facility,
			"severity":  severity,
			"priority":  facility*8 + severity,
			"content":   a[2],
		}
		j, err := json.Marshal(&sl)
		if err != nil {
			log.Printf("sshdPutSyslog err=%v", err)
			continue
		}
		logCh <- &datastore.LogEnt{
			Time: time.Unix(0, ts).UnixNano(),
			Type: "syslog",
			Log:  string(j),
		}
		count++
	}
	if r.Err() != nil {
		log.Printf("sshdPutSyslog err=%v", r.Err())
	}
}

// Node list
func sshdGetNodes(s ssh.Session) {
	cmds := s.Command()
	csv := len(cmds) > 2 && cmds[2] == "csv"
	if csv {
		io.WriteString(s, "Name,IP,MAC,State\r\n")
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if csv {
			io.WriteString(s, fmt.Sprintf("%s,%s,%s,%s\r\n", n.Name, n.IP, n.MAC, n.State))
		} else {
			nc := *n
			nc.Password = ""
			if j, err := json.Marshal(&nc); err == nil {
				io.WriteString(s, string(j)+"\r\n")
			}
		}
		return true
	})
}

// Polling list
func sshdGetPollings(s ssh.Session) {
	cmds := s.Command()
	csv := len(cmds) > 2 && cmds[2] == "csv"
	if csv {
		io.WriteString(s, "Node,Name,State\r\n")
	}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if csv {
			n := datastore.GetNode(p.NodeID)
			if n == nil {
				return true
			}
			io.WriteString(s, fmt.Sprintf("%s,%s,%s\r\n", n.Name, p.Name, p.State))
		} else {
			if j, err := json.Marshal(p); err == nil {
				io.WriteString(s, string(j)+"\r\n")
			}
		}
		return true
	})
}
