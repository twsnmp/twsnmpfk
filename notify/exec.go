// Package notify : 通知処理
package notify

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/twsnmp/twsnmpfk/cmd"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func checkExecCmd() {
	if datastore.NotifyConf.ExecCmd == "" {
		return
	}
	execLevel := 3
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		ns := getLevelNum(n.State)
		if execLevel > ns {
			execLevel = ns
			if ns == 0 {
				return false
			}
		}
		return true
	})
	if execLevel != lastExecLevel {
		err := ExecNotifyCmd(datastore.NotifyConf.ExecCmd, execLevel)
		if err != nil {
			log.Printf("execNotifyCmd err=%v", err)
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "low",
				Event: fmt.Sprintf(i18n.Trans("Exec notify command err=%v"), err),
			})
		}
		lastExecLevel = execLevel
	}
}

func ExecNotifyCmd(c string, level int) error {
	cl := strings.Split(c, " ")
	if len(cl) < 1 {
		return fmt.Errorf("notify ExecCmd is empty")
	}
	strLevel := fmt.Sprintf("%d", level)
	for i, v := range cl {
		if v == "$level" {
			cl[i] = strLevel
		}
	}
	tio := &timeout.Timeout{
		Duration:  60 * time.Second,
		KillAfter: 5 * time.Second,
	}
	if len(cl) == 1 {
		tio.Cmd = cmd.GetCmd(cl[0], nil)
	} else {
		tio.Cmd = cmd.GetCmd(cl[0], cl[1:])
	}
	_, _, _, err := tio.Run()
	return err
}
