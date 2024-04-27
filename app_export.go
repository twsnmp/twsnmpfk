package main

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

type ExportData struct {
	Title  string          `json:"Title"`
	Header []string        `json:"Header"`
	Data   [][]interface{} `json:"Data"`
	Image  string          `json:"Image"`
}

func (a *App) ExportNodes(t string) string {
	data := ExportData{
		Title:  "TWSNMP Node List",
		Header: []string{"State", "Name", "IP", "MAC", "Descr"},
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		l := []any{}
		l = append(l, n.State)
		l = append(l, n.Name)
		l = append(l, n.IP)
		l = append(l, n.MAC)
		l = append(l, n.Descr)
		data.Data = append(data.Data, l)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("ExportTable err=%v", err)
		return fmt.Sprintf("export err=%v", err)
	}
	return ""
}

func (a *App) ExportPollings(t string) string {
	data := ExportData{
		Title:  "TWSNMP Polling List",
		Header: []string{"State", "Node Name", "Name", "Level", "Type", "Log Mode", "Last Time"},
	}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		l := []any{}
		l = append(l, p.State)
		l = append(l, n.Name)
		l = append(l, p.Name)
		l = append(l, p.Level)
		l = append(l, p.Type)
		l = append(l, p.LogMode)
		l = append(l, time.Unix(0, p.LastTime).Format("2006/01/02 15:04:05"))
		data.Data = append(data.Data, l)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("ExportTable err=%v", err)
		return fmt.Sprintf("export err=%v", err)
	}
	return ""
}

// ExportEventLogs  export event logs
func (a *App) ExportEventLogs(t string, filter EventLogFilterEnt) string {
	typeFilter := makeStringFilter(filter.EventType)
	nodeFilter := makeStringFilter(filter.NodeName)
	eventFilter := makeStringFilter(filter.Event)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP Event Log",
		Header: []string{"Level", "Time", "Type", "Node Name", "Event"},
	}
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if typeFilter != nil && !typeFilter.MatchString(l.Type) {
			return true
		}
		if nodeFilter != nil && !nodeFilter.MatchString(l.NodeName) {
			return true
		}
		if eventFilter != nil && !eventFilter.MatchString(l.Event) {
			return true
		}
		if filter.Level != 0 && filter.Level > getLevelNum(l.Level) {
			return true
		}
		e := []any{}
		e = append(e, l.Level)
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.Type)
		e = append(e, l.NodeName)
		e = append(e, l.Event)
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export eventlog err=%v", err)
		return fmt.Sprintf("export eventlog err=%v", err)
	}
	return ""
}

// ExportSyslogs  export syslogs
func (a *App) ExportSyslogs(t string, filter SyslogFilterEnt) string {
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	msgFilter := makeStringFilter(filter.Message)
	st := makeTimeFilter(filter.Start, 1)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP Syslog",
		Header: []string{"Level", "Time", "Host", "Type", "Tag", "Message"},
	}
	datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		if filter.Severity < l.Severity {
			return true
		}
		if hostFilter != nil && !hostFilter.MatchString(l.Host) {
			return true
		}
		if tagFilter != nil && !tagFilter.MatchString(l.Tag) {
			return true
		}
		if msgFilter != nil && !msgFilter.MatchString(l.Message) {
			return true
		}
		e := []any{}
		e = append(e, l.Level)
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.Host)
		e = append(e, l.Type)
		e = append(e, l.Tag)
		e = append(e, l.Message)
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export syslog err=%v", err)
		return fmt.Sprintf("export syslog err=%v", err)
	}
	return ""
}

// ExportTrap  export traps
func (a *App) ExportTraps(t string, filter TrapFilterEnt) string {
	fromFilter := makeStringFilter(filter.From)
	typeFilter := makeStringFilter(filter.Type)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP TRAP",
		Header: []string{"Time", "From", "Type", "Variables"},
	}
	datastore.ForEachTraps(st, et, func(l *datastore.TrapEnt) bool {
		if fromFilter != nil && !fromFilter.MatchString(l.FromAddress) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.TrapType) {
			return true
		}
		e := []any{}
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.FromAddress)
		e = append(e, l.TrapType)
		e = append(e, l.Variables)
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export TRAPs err=%v", err)
		return fmt.Sprintf("export traps err=%v", err)
	}
	return ""
}

// ExportNetFlow  export traps
func (a *App) ExportNetFlow(t string, filter NetFlowFilterEnt) string {
	srcFilter := makeStringFilter(filter.SrcAddr)
	srcLocFilter := makeStringFilter(filter.SrcLoc)
	dstFilter := makeStringFilter(filter.DstAddr)
	dstLocFilter := makeStringFilter(filter.DstLoc)
	tcpFlagsFilter := makeStringFilter(filter.TCPFlags)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP NetFlow",
		Header: []string{"Time", "Src IP", "Src Port", "Src Loc", "Dst IP", "Dst Port", "Dst Loc", "Protocol", "TCPFlags", "Packets", "Bytes", "Dur"},
	}
	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		if filter.Single {
			if srcFilter != nil && (!srcFilter.MatchString(l.SrcAddr) && !srcFilter.MatchString(l.DstAddr)) {
				return true
			}
			if srcLocFilter != nil && (!srcLocFilter.MatchString(l.SrcLoc) && !srcLocFilter.MatchString(l.DstLoc)) {
				return true
			}
			if filter.SrcPort > 0 && (filter.SrcPort != l.SrcPort && filter.SrcPort != l.DstPort) {
				return true
			}
		} else {
			if srcFilter != nil && !srcFilter.MatchString(l.SrcAddr) {
				return true
			}
			if srcLocFilter != nil && !srcLocFilter.MatchString(l.SrcLoc) {
				return true
			}
			if dstFilter != nil && !dstFilter.MatchString(l.DstAddr) {
				return true
			}
			if dstLocFilter != nil && !dstLocFilter.MatchString(l.DstLoc) {
				return true
			}
			if filter.SrcPort > 0 && filter.SrcPort != l.SrcPort {
				return true
			}
			if filter.DstPort > 0 && filter.DstPort != l.DstPort {
				return true
			}
		}
		if tcpFlagsFilter != nil && !tcpFlagsFilter.MatchString(l.TCPFlags) {
			return true
		}
		if protocolFilter != nil && !protocolFilter.MatchString(l.Protocol) {
			return true
		}
		e := []any{}
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.SrcAddr)
		e = append(e, l.SrcPort)
		e = append(e, l.SrcLoc)
		e = append(e, l.DstAddr)
		e = append(e, l.DstPort)
		e = append(e, l.DstLoc)
		e = append(e, l.Protocol)
		e = append(e, l.TCPFlags)
		e = append(e, l.Packets)
		e = append(e, l.Bytes)
		e = append(e, l.Dur)
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export NetFlow err=%v", err)
		return fmt.Sprintf("export NetFlow err=%v", err)
	}
	return ""
}

// ExportArpLogs  export arp watch logs
func (a *App) ExportArpLogs(t string) string {
	data := ExportData{
		Title:  "TWSNMP ARP Logs",
		Header: []string{"Time", "State", "IP", "Node", "New MAC", "New Vendor", "Old MAC", "Old Vendor"},
	}
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		e := []any{}
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.IP)
		if n := datastore.FindNodeFromIP(l.IP); n != nil {
			e = append(e, n.Name)
		} else {
			e = append(e, "")
		}
		e = append(e, l.NewMAC)
		e = append(e, datastore.FindVendor(l.NewMAC))
		e = append(e, l.OldMAC)
		e = append(e, datastore.FindVendor(l.OldMAC))
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export arp log err=%v", err)
		return fmt.Sprintf("export arp log err=%v", err)
	}
	return ""
}

// ExportArpTable  export arp Table
func (a *App) ExportArpTable(t string) string {
	data := ExportData{
		Title:  "TWSNMP ARP Table",
		Header: []string{"IP", "MAC", "Node", "Vendor"},
	}
	datastore.ForEachArp(func(l *datastore.ArpEnt) bool {
		e := []any{}
		e = append(e, l.IP)
		e = append(e, l.MAC)
		n := datastore.GetNode(l.NodeID)
		if n != nil {
			e = append(e, n.Name)
		} else {
			e = append(e, "")
		}
		e = append(e, l.Vendor)
		data.Data = append(data.Data, e)
		return true
	})
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export arp err=%v", err)
		return fmt.Sprintf("export arp err=%v", err)
	}
	return ""
}

func (a *App) ExportAny(t string, data ExportData) string {
	var err error
	switch t {
	case "excel":
		err = a.exportExcel(&data)
	case "csv":
		err = a.exportCSV(&data)
	default:
		return "not suppoerted"
	}
	if err != nil {
		log.Printf("Export any err=%v", err)
		return fmt.Sprintf("export any err=%v", err)
	}
	return ""
}

func (a *App) exportExcel(data *ExportData) error {
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      strings.ReplaceAll(data.Title, " ", "_") + "_" + d + ".xlsx",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "Excel",
			Pattern:     "*.xlsx",
		}},
	})
	if err != nil {
		return err
	}
	if file == "" {
		return nil
	}
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", data.Title+d)
	row := 3
	col := 'A'
	for _, h := range data.Header {
		f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", col, row), h)
		col++
	}
	imgCol := col + 2
	row++
	for _, l := range data.Data {
		col = 'A'
		for _, i := range l {
			f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", col, row), i)
			col++
		}
		row++
	}
	if data.Image != "" {
		b64data := data.Image[strings.IndexByte(data.Image, ',')+1:]
		img, err := base64.StdEncoding.DecodeString(b64data)
		if err != nil {
			return err
		}
		f.AddPictureFromBytes("Sheet1", fmt.Sprintf("%c2", imgCol), &excelize.Picture{
			Extension: ".png",
			File:      img,
		})
	}
	if err := f.SaveAs(file); err != nil {
		return err
	}
	return nil
}

func (a *App) exportCSV(data *ExportData) error {
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      strings.ReplaceAll(data.Title, " ", "_") + "_" + d + ".csv",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "CSV",
			Pattern:     "*.csv",
		}},
	})
	if err != nil {
		return err
	}
	if file == "" {
		return nil
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write(data.Header)
	for _, l := range data.Data {
		data := []string{}
		for _, i := range l {
			data = append(data, fmt.Sprintf("%v", i))
		}
		w.Write(data)
	}
	w.Flush()
	return w.Error()
}
