package polling

// 外部コマンド実行で監視する。

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/cmd"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/vjeantet/grok"
)

func doPollingCmd(pe *datastore.PollingEnt) {
	c := getReplacedCmd(pe)
	extractor := pe.Extractor
	script := pe.Script
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result = make(map[string]interface{})
	cl := strings.Split(c, " ")
	if len(cl) < 1 {
		setPollingError("cmd", pe, fmt.Errorf("no cmd"))
		return
	}
	tio := &timeout.Timeout{
		Duration:  time.Duration(pe.Timeout) * time.Second,
		KillAfter: 5 * time.Second,
	}

	if filepath.Ext(cl[0]) == ".sh" {
		cl[0] = filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		tio.Cmd = cmd.GetCmd("/bin/sh", []string{"-c", strings.Join(cl, " ")})
	} else {
		exe := filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		if len(cl) == 1 {
			tio.Cmd = cmd.GetCmd(exe, nil)
		} else {
			tio.Cmd = cmd.GetCmd(exe, cl[1:])
		}
	}

	exitStatus, stdout, stderr, err := tio.Run()
	if err != nil {
		setPollingError("cmd", pe, err)
		return
	}
	pe.Result["lastTime"] = time.Now().Format("2006-01-02T15:04")
	pe.Result["stderr"] = stderr
	pe.Result["exitCode"] = exitStatus.Code
	if err := vm.Set("exitCode", exitStatus.Code); err != nil {
		log.Printf("cmd polling err=%v", err)
	}
	if err := vm.Set("interval", pe.PollInt); err != nil {
		log.Printf("cmd polling err=%v", err)
	}
	if extractor == "goquery" {
		setPollingError("cmd", pe, fmt.Errorf("goquery not supported"))
		return
	} else if extractor == "getBody" {
		vm.Set("getBody", func(call otto.FunctionCall) otto.Value {
			if r, err := otto.ToValue(string(stdout)); err == nil {
				return r
			}
			return otto.UndefinedValue()
		})
	} else if extractor != "" {
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err := g.AddPattern("TWSNMP", extractor); err != nil {
			log.Printf("cmd polling err=%v", err)
		}
		values, err := g.Parse("%{TWSNMP}", string(stdout))
		if err != nil {
			setPollingError("cmd", pe, err)
			return
		}
		for k, v := range values {
			if err := vm.Set(k, v); err != nil {
				log.Printf("cmd polling err=%v", err)
			}
			pe.Result[k] = v
		}
	}
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("cmd", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}

func getReplacedCmd(pe *datastore.PollingEnt) string {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return ""
	}
	cmd := pe.Params
	cmd = strings.ReplaceAll(cmd, "$IP", n.IP)
	cmd = strings.ReplaceAll(cmd, "$NODE", strings.ReplaceAll(n.Name, " ", "_"))
	return cmd
}
