package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
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
func (a *App) ExportEventLogs(t string, filter EventLogFilterEnt, image string) string {
	typeFilter := makeStringFilter(filter.EventType)
	nodeFilter := makeStringFilter(filter.NodeName)
	eventFilter := makeStringFilter(filter.Event)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP Event Log",
		Header: []string{"Level", "Time", "Type", "Node Name", "Event"},
		Image:  image,
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
func (a *App) ExportSyslogs(t string, filter SyslogFilterEnt, image string) string {
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	msgFilter := makeStringFilter(filter.Message)
	st := makeTimeFilter(filter.Start, 1)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP Syslog",
		Header: []string{"Level", "Time", "Host", "Type", "Tag", "Message"},
		Image:  image,
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
func (a *App) ExportTraps(t string, filter TrapFilterEnt, image string) string {
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

// ExportNetFlow  export netflow
func (a *App) ExportNetFlow(t string, filter NetFlowFilterEnt, image string) string {
	srcFilter := makeStringFilter(filter.SrcAddr)
	srcLocFilter := makeStringFilter(filter.SrcLoc)
	srcMACFilter := makeStringFilter(filter.SrcMAC)
	dstFilter := makeStringFilter(filter.DstAddr)
	dstLocFilter := makeStringFilter(filter.DstLoc)
	dstMACFilter := makeStringFilter(filter.DstMAC)
	tcpFlagsFilter := makeStringFilter(filter.TCPFlags)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP NetFlow",
		Header: []string{"Time", "Src IP", "Src Port", "Src Loc", "Src MAC", "Dst IP", "Dst Port", "Dst Loc", "Dst MAC", "Protocol", "TCPFlags", "Packets", "Bytes", "Dur"},
	}
	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		if filter.Single {
			if srcFilter != nil && (!srcFilter.MatchString(l.SrcAddr) && !srcFilter.MatchString(l.DstAddr)) {
				return true
			}
			if srcLocFilter != nil && (!srcLocFilter.MatchString(l.SrcLoc) && !srcLocFilter.MatchString(l.DstLoc)) {
				return true
			}
			if srcMACFilter != nil && (!srcMACFilter.MatchString(l.SrcMAC) && !srcMACFilter.MatchString(l.DstMAC)) {
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
			if srcMACFilter != nil && !srcMACFilter.MatchString(l.SrcMAC) {
				return true
			}
			if dstFilter != nil && !dstFilter.MatchString(l.DstAddr) {
				return true
			}
			if dstLocFilter != nil && !dstLocFilter.MatchString(l.DstLoc) {
				return true
			}
			if dstMACFilter != nil && !dstMACFilter.MatchString(l.DstMAC) {
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
		e = append(e, l.SrcMAC)
		e = append(e, l.DstAddr)
		e = append(e, l.DstPort)
		e = append(e, l.DstLoc)
		e = append(e, l.DstMAC)
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

// ExportSFlow  export sFlow
func (a *App) ExportSFlow(t string, filter SFlowFilterEnt, image string) string {
	srcFilter := makeStringFilter(filter.SrcAddr)
	srcLocFilter := makeStringFilter(filter.SrcLoc)
	srcMACFilter := makeStringFilter(filter.SrcMAC)
	dstFilter := makeStringFilter(filter.DstAddr)
	dstLocFilter := makeStringFilter(filter.DstLoc)
	dstMACFilter := makeStringFilter(filter.DstMAC)
	tcpFlagsFilter := makeStringFilter(filter.TCPFlags)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP sFlow",
		Header: []string{"Time", "Src IP", "Src Port", "Src Loc", "Src MAC", "Dst IP", "Dst Port", "Dst Loc", "Dst MAC", "Protocol", "TCPFlags", "Bytes", "Reason"},
		Image:  image,
	}
	datastore.ForEachSFlow(st, et, func(l *datastore.SFlowEnt) bool {
		if filter.Single {
			if srcFilter != nil && (!srcFilter.MatchString(l.SrcAddr) && !srcFilter.MatchString(l.DstAddr)) {
				return true
			}
			if srcLocFilter != nil && (!srcLocFilter.MatchString(l.SrcLoc) && !srcLocFilter.MatchString(l.DstLoc)) {
				return true
			}
			if srcMACFilter != nil && (!srcMACFilter.MatchString(l.SrcMAC) && !srcMACFilter.MatchString(l.DstMAC)) {
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
			if srcMACFilter != nil && !srcMACFilter.MatchString(l.SrcMAC) {
				return true
			}
			if dstFilter != nil && !dstFilter.MatchString(l.DstAddr) {
				return true
			}
			if dstLocFilter != nil && !dstLocFilter.MatchString(l.DstLoc) {
				return true
			}
			if dstMACFilter != nil && !dstMACFilter.MatchString(l.DstMAC) {
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
		e = append(e, l.SrcMAC)
		e = append(e, l.DstAddr)
		e = append(e, l.DstPort)
		e = append(e, l.DstLoc)
		e = append(e, l.DstMAC)
		e = append(e, l.Protocol)
		e = append(e, l.TCPFlags)
		e = append(e, l.Bytes)
		e = append(e, l.Reason)
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
		log.Printf("Export sFlow err=%v", err)
		return fmt.Sprintf("export sFlow err=%v", err)
	}
	return ""
}

// ExportSFlowCounter  export sFlow Counter log
func (a *App) ExportSFlowCounter(t string, filter SFlowCounterFilterEnt, image string) string {
	remoteFilter := makeIPFilter(filter.Remote)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	data := ExportData{
		Title:  "TWSNMP sFlow Counter",
		Header: []string{"Time", "Type", "Remote", "Data"},
		Image:  image,
	}
	datastore.ForEachSFlowCounter(st, et, func(l *datastore.SFlowCounterEnt) bool {
		if remoteFilter != nil && !remoteFilter.MatchString(l.Remote) {
			return true
		}
		if filter.Type != "" && filter.Type != l.Type {
			return true
		}
		e := []any{}
		e = append(e, time.Unix(0, l.Time).Format("2006/01/02 15:04:05"))
		e = append(e, l.Type)
		e = append(e, l.Remote)
		e = append(e, makeJSONDataToString(l.Data))
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
		log.Printf("Export sFlow counter err=%v", err)
		return fmt.Sprintf("export sFlow  countererr=%v", err)
	}
	return ""
}

// CSVのためのデータ変換
func makeJSONDataToString(j string) string {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(j), &m); err == nil {
		r := []string{}
		for k, v := range m {
			r = append(r, fmt.Sprintf("%s=%v", k, v))
		}
		j = strings.Join(r, " ")
	}
	return strings.ReplaceAll(j, ",", " ")
}

// ExportArpLogs  export arp watch logs
func (a *App) ExportArpLogs(t, image string) string {
	data := ExportData{
		Title:  "TWSNMP ARP Logs",
		Header: []string{"Time", "State", "IP", "Node", "New MAC", "New Vendor", "Old MAC", "Old Vendor"},
		Image:  image,
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
	row++
	for _, l := range data.Data {
		col = 'A'
		for _, i := range l {
			f.SetCellValue("Sheet1", fmt.Sprintf("%c%d", col, row), i)
			col++
		}
		row++
	}
	f.SetSheetName("Sheet1", data.Title)
	if data.Image != "" {
		v := strings.SplitN(data.Image, ",", 2)
		if len(v) == 2 {
			if img, err := base64.StdEncoding.DecodeString(v[1]); err == nil {
				f.NewSheet("Chart")
				f.AddPictureFromBytes("Chart", "A1", &excelize.Picture{
					Extension: ".png",
					File:      img,
				})
			}
		}
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

// ExportMapは、マップのイメージを画像またはExcelに保存します。
func (a *App) ExportMap(data string) error {
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		Title:                i18n.Trans("Export MAP"),
		DefaultFilename:      "TWSNMPFK_MAP" + "_" + d + ".xlsx",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "PNG;Excel",
			Pattern:     "*.png;*.xlsx",
		}},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	if file == "" {
		return nil
	}
	v := strings.SplitN(data, ",", 2)
	if len(v) != 2 {
		return fmt.Errorf("invalid image data")
	}
	img, err := base64.StdEncoding.DecodeString(v[1])
	if err != nil {
		log.Println(err)
		return err
	}
	if filepath.Ext(file) == ".png" {
		return os.WriteFile(file, img, 0640)
	}
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "TWSNMP FK MAP"+"-"+d)
	err = f.AddPictureFromBytes("Sheet1", "A3", &excelize.Picture{
		Extension: ".png",
		File:      img,
		Format:    &excelize.GraphicOptions{AltText: "MAP"},
	})
	if err != nil {
		log.Println(err)
	}
	f.SetSheetName("Sheet1", "MAP")
	f.NewSheet("Sheet2")
	f.SetCellValue("Sheet2", "A1", "TWSNMP FK Node List"+"-"+d)
	row := 3
	col := 'A'
	for _, h := range []string{"Name", "IP", "MAC", "Descr"} {
		f.SetCellValue("Sheet2", fmt.Sprintf("%c%d", col, row), h)
		col++
	}
	row++
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		f.SetCellValue("Sheet2", fmt.Sprintf("A%d", row), n.Name)
		f.SetCellValue("Sheet2", fmt.Sprintf("B%d", row), n.IP)
		f.SetCellValue("Sheet2", fmt.Sprintf("C%d", row), n.MAC)
		f.SetCellValue("Sheet2", fmt.Sprintf("D%d", row), n.Descr)
		row++
		return true
	})
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		f.SetCellValue("Sheet2", fmt.Sprintf("A%d", row), n.Name)
		f.SetCellValue("Sheet2", fmt.Sprintf("B%d", row), n.IP)
		f.SetCellValue("Sheet2", fmt.Sprintf("C%d", row), "")
		f.SetCellValue("Sheet2", fmt.Sprintf("D%d", row), n.Descr)
		row++
		return true
	})
	f.SetSheetName("Sheet2", "Node List")
	if err := f.SaveAs(file); err != nil {
		return err
	}
	return nil
}

func (a *App) ExportPortDef(d string) error {
	ts := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "twsnmpfk_port_def_" + ts + ".csv",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "JSON",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return err
	}
	if file == "" {
		return nil
	}
	return os.WriteFile(file, []byte(d), 0644)
}

func (a *App) ImportPortDef() string {
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title: "TWSNMP FK Port Def file",
		Filters: []wails.FileFilter{{
			DisplayName: "TWSNMP FK Port Def file(*.json)",
			Pattern:     "*.json;",
		}},
	})
	if err != nil {
		log.Printf("err=%v", err)
		return ""
	}
	if file == "" {
		return ""
	}
	d, err := os.ReadFile(file)
	if err != nil {
		log.Printf("err=%v", err)
		return ""
	}
	if len(d) > 1024*1024 {
		return ""
	}
	return string(d)
}

func (a *App) ExportPollingAsTemplate(id string) error {
	p := datastore.GetPolling(id)
	if p == nil {
		return fmt.Errorf("polling not found")
	}
	d := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "TWSNMP_Polling_Template_" + d + ".json",
		CanCreateDirectories: true,
		Filters: []wails.FileFilter{{
			DisplayName: "TWSNMP Polling Template file(*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return err
	}
	if file == "" {
		return nil
	}
	pt := datastore.PollingTemplateEnt{
		Name:      p.Name,
		Type:      p.Type,
		Mode:      p.Mode,
		Filter:    p.Filter,
		Extractor: p.Extractor,
		Script:    p.Script,
		Descr:     p.Name,
		Params:    p.Params,
		Level:     "off",
	}
	j, err := json.MarshalIndent(pt, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, j, 0666)
}

// ImportPollingTemplate : ポーリングテンプレートファイルを読み込む
func (a *App) ImportPollingTemplate() datastore.PollingTemplateEnt {
	var r datastore.PollingTemplateEnt
	r.ID = -1
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title: "TWSNMP Polling Template",
		Filters: []wails.FileFilter{{
			DisplayName: "TWSNMP Polling Template file(*.json)",
			Pattern:     "*.json;",
		}},
	})
	if err != nil {
		log.Printf("err=%v", err)
		return r
	}
	if file == "" {
		return r
	}
	d, err := os.ReadFile(file)
	if err != nil {
		log.Printf("err=%v", err)
		return r
	}
	if err = json.Unmarshal(d, &r); err != nil {
		r.ID = -1
		log.Printf("err=%v", err)
	}
	return r
}
