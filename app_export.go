package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/twsnmp/twsnmpfk/backend"
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

// ExportTraps exports traps
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

// ExportMap exports the map or data to a file in various formats.
func (a *App) ExportMap(format string, pngBase64 string) (string, error) {
	d := time.Now().Format("20060102150405")
	var title string
	var defaultFilename string
	var filters []wails.FileFilter

	switch format {
	case "png":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".png"
		filters = []wails.FileFilter{{DisplayName: "PNG Image (*.png)", Pattern: "*.png"}}
	case "svg":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".svg"
		filters = []wails.FileFilter{{DisplayName: "SVG Image (*.svg)", Pattern: "*.svg"}}
	case "pdf":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".pdf"
		filters = []wails.FileFilter{{DisplayName: "PDF Document (*.pdf)", Pattern: "*.pdf"}}
	case "drawio":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".drawio"
		filters = []wails.FileFilter{{DisplayName: "Draw.io Diagram (*.drawio)", Pattern: "*.drawio"}}
	case "json_map":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".json"
		filters = []wails.FileFilter{{DisplayName: "JSON Map Data (*.json)", Pattern: "*.json"}}
	case "csv":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".csv"
		filters = []wails.FileFilter{{DisplayName: "CSV Node List (*.csv)", Pattern: "*.csv"}}
	case "excel":
		title = i18n.Trans("Export MAP")
		defaultFilename = "TWSNMPFK_MAP_" + d + ".xlsx"
		filters = []wails.FileFilter{{DisplayName: "Excel Document (*.xlsx)", Pattern: "*.xlsx"}}
	default:
		err := fmt.Errorf("unsupported format: %s", format)
		wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
			Type:    wails.ErrorDialog,
			Title:   i18n.Trans("Export MAP"),
			Message: err.Error(),
		})
		return "", err
	}

	selectedFile, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters:         filters,
	})
	if err != nil {
		wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
			Type:    wails.ErrorDialog,
			Title:   i18n.Trans("Export MAP"),
			Message: err.Error(),
		})
		return "", err
	}
	if selectedFile == "" {
		return "", nil // user cancelled
	}

	var pngBytes []byte
	if pngBase64 != "" {
		parts := strings.SplitN(pngBase64, ",", 2)
		base64Data := parts[0]
		if len(parts) > 1 {
			base64Data = parts[1]
		}
		decoded, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			errStr := fmt.Sprintf("failed to decode base64 PNG: %v", err)
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: errStr,
			})
			return "", fmt.Errorf(errStr)
		}
		pngBytes = decoded
	}

	switch format {
	case "png":
		if len(pngBytes) == 0 {
			errStr := "no PNG data provided"
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: errStr,
			})
			return "", fmt.Errorf(errStr)
		}
		if err := os.WriteFile(selectedFile, pngBytes, 0600); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "svg":
		var svgBuf bytes.Buffer
		minX, maxX := 0.0, 1000.0
		minY, maxY := 0.0, 800.0
		first := true

		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			x, y := float64(n.X), float64(n.Y)
			if first {
				minX, maxX = x, x
				minY, maxY = y, y
				first = false
			} else {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
			return true
		})

		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			x, y := float64(n.X), float64(n.Y)
			w, h := float64(n.W), float64(n.H)
			if first {
				minX, maxX = x, x+w
				minY, maxY = y, y+h
				first = false
			} else {
				if x < minX {
					minX = x
				}
				if x+w > maxX {
					maxX = x + w
				}
				if y < minY {
					minY = y
				}
				if y+h > maxY {
					maxY = y + h
				}
			}
			return true
		})

		margin := 100.0
		minX -= margin
		minY -= margin
		maxX += margin
		maxY += margin
		width := maxX - minX
		height := maxY - minY
		if width < 100 {
			width = 800
		}
		if height < 100 {
			height = 600
		}

		svgBuf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="%f %f %f %f" width="100%%" height="100%%" style="background-color: #f8fafc;">`, minX, minY, width, height))

		datastore.ForEachLines(func(l *datastore.LineEnt) bool {
			var x1, y1, x2, y2 float64
			found1, found2 := false, false

			if n := datastore.GetNode(l.NodeID1); n != nil {
				x1, y1 = float64(n.X), float64(n.Y)
				found1 = true
			} else if net := datastore.GetNetwork(l.NodeID1); net != nil {
				x1, y1 = float64(net.X+net.W/2), float64(net.Y+net.H/2)
				found1 = true
			}

			if n := datastore.GetNode(l.NodeID2); n != nil {
				x2, y2 = float64(n.X), float64(n.Y)
				found2 = true
			} else if net := datastore.GetNetwork(l.NodeID2); net != nil {
				x2, y2 = float64(net.X+net.W/2), float64(net.Y+net.H/2)
				found2 = true
			}

			if found1 && found2 {
				strokeWidth := fmt.Sprintf("%d", l.Width)
				if l.Width <= 0 {
					strokeWidth = "2"
				}
				color := "#94a3b8"
				switch l.State {
				case "warn":
					color = "#f59e0b"
				case "error":
					color = "#ef4444"
				case "normal":
					color = "#10b981"
				}

				svgBuf.WriteString(fmt.Sprintf(`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%s" />`,
					x1, y1, x2, y2, color, strokeWidth))

				if l.Info != "" {
					midX := (x1 + x2) / 2
					midY := (y1 + y2) / 2
					svgBuf.WriteString(fmt.Sprintf(`<text x="%f" y="%f" text-anchor="middle" font-size="10" fill="#64748b">%s</text>`, midX, midY-5, l.Info))
				}
			}
			return true
		})

		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			color := "#3b82f6"
			if n.Error != "" {
				color = "#ef4444"
			}
			svgBuf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="8" fill="%s" fill-opacity="0.1" stroke="%s" stroke-width="2" />`,
				n.X, n.Y, n.W, n.H, color, color))
			svgBuf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" font-weight="bold" fill="#1e293b">%s</text>`,
				n.X+10, n.Y+20, n.Name))
			if n.IP != "" {
				svgBuf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" fill="#64748b">%s</text>`,
					n.X+10, n.Y+35, n.IP))
			}
			return true
		})

		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			color := "#64748b"
			switch n.State {
			case "normal":
				color = "#10b981"
			case "warn":
				color = "#f59e0b"
			case "error":
				color = "#ef4444"
			}

			svgBuf.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="24" fill="%s" stroke="#ffffff" stroke-width="3" />`, n.X, n.Y, color))

			initial := "N"
			if len(n.Icon) > 0 {
				initial = strings.ToUpper(n.Icon[:1])
			} else if len(n.Name) > 0 {
				initial = strings.ToUpper(n.Name[:1])
			}
			svgBuf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" dy="6" text-anchor="middle" font-size="16" font-weight="bold" fill="#ffffff">%s</text>`, n.X, n.Y, initial))

			svgBuf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="11" font-weight="600" fill="#1e293b">%s</text>`, n.X, n.Y+38, n.Name))
			if n.IP != "" && n.IP != n.Name {
				svgBuf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-size="9" fill="#64748b">%s</text>`, n.X, n.Y+50, n.IP))
			}
			return true
		})

		svgBuf.WriteString(`</svg>`)
		if err := os.WriteFile(selectedFile, svgBuf.Bytes(), 0600); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "pdf":
		if len(pngBytes) == 0 {
			errStr := "no PNG data provided"
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: errStr,
			})
			return "", fmt.Errorf(errStr)
		}
		tmpFile, err := os.CreateTemp("", "twsnmpfk-*.png")
		if err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.Write(pngBytes); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}
		tmpFile.Close()

		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		pdf.AddPage()

		err = pdf.Image(tmpFile.Name(), 20, 20, &gopdf.Rect{W: 555, H: 416})
		if err != nil {
			errStr := fmt.Sprintf("failed to add image to PDF: %v", err)
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: errStr,
			})
			return "", fmt.Errorf(errStr)
		}

		if err := pdf.WritePdf(selectedFile); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "drawio":
		minX, minY := 0.0, 0.0
		first := true
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			x, y := float64(n.X), float64(n.Y)
			if first {
				minX, minY = x, y
				first = false
			} else {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
			}
			return true
		})
		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			x, y := float64(n.X), float64(n.Y)
			if first {
				minX, minY = x, y
				first = false
			} else {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
			}
			return true
		})

		var drawioBuf bytes.Buffer
		drawioBuf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<mxfile host="Electron" modified="2026-07-14T00:00:00.000Z" agent="5.0" version="20.0.0" type="device">
  <diagram id="netmap" name="Network Map">
    <mxGraphModel dx="1000" dy="1000" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />`)

		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			x := float64(n.X) - minX + 50.0
			y := float64(n.Y) - minY + 50.0
			fillColor := "#D5E8D4"
			strokeColor := "#82B366"
			switch n.State {
			case "warn":
				fillColor = "#FFF2CC"
				strokeColor = "#D6B656"
			case "error":
				fillColor = "#F8CECC"
				strokeColor = "#B85450"
			case "unknown":
				fillColor = "#F5F5F5"
				strokeColor = "#666666"
			}
			value := fmt.Sprintf("%s\n%s", n.Name, n.IP)
			value = strings.ReplaceAll(value, "\n", "&lt;br/&gt;")
			drawioBuf.WriteString(fmt.Sprintf(`
        <mxCell id="%s" value="%s" style="rounded=1;whiteSpace=wrap;html=1;fillColor=%s;strokeColor=%s;fontStyle=1" vertex="1" parent="1">
          <mxGeometry x="%f" y="%f" width="100" height="50" as="geometry" />
        </mxCell>`, n.ID, value, fillColor, strokeColor, x, y))
			return true
		})

		datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
			x := float64(net.X) - minX + 50.0
			y := float64(net.Y) - minY + 50.0
			fillColor := "#DAE8FC"
			strokeColor := "#6C8EBF"
			if net.Error != "" {
				fillColor = "#F8CECC"
				strokeColor = "#B85450"
			}
			value := fmt.Sprintf("%s\n%s", net.Name, net.IP)
			value = strings.ReplaceAll(value, "\n", "&lt;br/&gt;")
			drawioBuf.WriteString(fmt.Sprintf(`
        <mxCell id="%s" value="%s" style="rounded=0;whiteSpace=wrap;html=1;fillColor=%s;strokeColor=%s;fontStyle=1" vertex="1" parent="1">
          <mxGeometry x="%f" y="%f" width="%d" height="%d" as="geometry" />
        </mxCell>`, net.ID, value, fillColor, strokeColor, x, y, net.W, net.H))
			return true
		})

		datastore.ForEachLines(func(l *datastore.LineEnt) bool {
			strokeColor := "#94a3b8"
			switch l.State {
			case "warn":
				strokeColor = "#f59e0b"
			case "error":
				strokeColor = "#ef4444"
			case "normal":
				strokeColor = "#10b981"
			}
			strokeWidth := l.Width
			if strokeWidth <= 0 {
				strokeWidth = 2
			}
			style := fmt.Sprintf("endArrow=none;html=1;rounded=0;strokeColor=%s;strokeWidth=%d;", strokeColor, strokeWidth)
			drawioBuf.WriteString(fmt.Sprintf(`
        <mxCell id="%s" value="%s" style="%s" edge="1" parent="1" source="%s" target="%s">
          <mxGeometry relative="1" as="geometry" />
        </mxCell>`, l.ID, l.Info, style, l.NodeID1, l.NodeID2))
			return true
		})

		drawioBuf.WriteString(`
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`)
		if err := os.WriteFile(selectedFile, drawioBuf.Bytes(), 0600); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "json_map":
		type MapExportData struct {
			Nodes    []*datastore.NodeEnt    `json:"nodes"`
			Networks []*datastore.NetworkEnt `json:"networks"`
			Lines    []*datastore.LineEnt    `json:"lines"`
		}
		var expData MapExportData
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			expData.Nodes = append(expData.Nodes, n)
			return true
		})
		datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
			expData.Networks = append(expData.Networks, net)
			return true
		})
		datastore.ForEachLines(func(l *datastore.LineEnt) bool {
			expData.Lines = append(expData.Lines, l)
			return true
		})

		indent, err := json.MarshalIndent(expData, "", "  ")
		if err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}
		if err := os.WriteFile(selectedFile, indent, 0600); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "csv":
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		writer.Write([]string{"Type", "ID", "Name", "IP", "MAC", "State", "X", "Y", "Descr"})
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			writer.Write([]string{
				"node",
				n.ID,
				n.Name,
				n.IP,
				n.MAC,
				n.State,
				fmt.Sprintf("%d", n.X),
				fmt.Sprintf("%d", n.Y),
				n.Descr,
			})
			return true
		})
		datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
			writer.Write([]string{
				"network",
				net.ID,
				net.Name,
				net.IP,
				"",
				net.Error,
				fmt.Sprintf("%d", net.X),
				fmt.Sprintf("%d", net.Y),
				net.Descr,
			})
			return true
		})
		writer.Flush()
		if err := os.WriteFile(selectedFile, buf.Bytes(), 0600); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}

	case "excel":
		f := excelize.NewFile()
		f.SetCellValue("Sheet1", "A1", "TWSNMP FK MAP"+"-"+d)
		if len(pngBytes) > 0 {
			err = f.AddPictureFromBytes("Sheet1", "A3", &excelize.Picture{
				Extension: ".png",
				File:      pngBytes,
				Format:    &excelize.GraphicOptions{AltText: "MAP"},
			})
			if err != nil {
				log.Println(err)
			}
		}
		f.SetSheetName("Sheet1", "MAP")
		f.NewSheet("Sheet2")
		f.SetCellValue("Sheet2", "A1", "TWSNMP FK Node List"+"-"+d)
		row := 3
		col := 'A'
		for _, h := range []string{"Type", "Name", "IP", "MAC", "Descr"} {
			f.SetCellValue("Sheet2", fmt.Sprintf("%c%d", col, row), h)
			col++
		}
		row++
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			f.SetCellValue("Sheet2", fmt.Sprintf("A%d", row), "node")
			f.SetCellValue("Sheet2", fmt.Sprintf("B%d", row), n.Name)
			f.SetCellValue("Sheet2", fmt.Sprintf("C%d", row), n.IP)
			f.SetCellValue("Sheet2", fmt.Sprintf("D%d", row), n.MAC)
			f.SetCellValue("Sheet2", fmt.Sprintf("E%d", row), n.Descr)
			row++
			return true
		})
		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			f.SetCellValue("Sheet2", fmt.Sprintf("A%d", row), "network")
			f.SetCellValue("Sheet2", fmt.Sprintf("B%d", row), n.Name)
			f.SetCellValue("Sheet2", fmt.Sprintf("C%d", row), n.IP)
			f.SetCellValue("Sheet2", fmt.Sprintf("D%d", row), "")
			f.SetCellValue("Sheet2", fmt.Sprintf("E%d", row), n.Descr)
			row++
			return true
		})
		f.SetSheetName("Sheet2", "Node List")
		if err := f.SaveAs(selectedFile); err != nil {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.ErrorDialog,
				Title:   i18n.Trans("Export MAP"),
				Message: err.Error(),
			})
			return "", err
		}
	}

	wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:    wails.InfoDialog,
		Title:   i18n.Trans("Export MAP"),
		Message: i18n.Trans("Export MAP") + " " + i18n.Trans("Success"),
	})
	return selectedFile, nil
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
	return os.WriteFile(file, []byte(d), 0600)
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
	return os.WriteFile(file, j, 0600)
}

// ImportPollingTemplate : ポーリングテンプレートファイルを読み込む
func (a *App) ImportPollingTemplate() datastore.PollingTemplateEnt {
	var r datastore.PollingTemplateEnt
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
		log.Printf("err=%v", err)
	}
	return r
}

func (a *App) ExportAIData(id string) error {
	req := &backend.AIReq{PollingID: id}
	if err := backend.MakeAIData(req); err != nil {
		return err
	}
	ts := time.Now().Format("20060102150405")
	file, err := wails.SaveFileDialog(a.ctx, wails.SaveDialogOptions{
		DefaultFilename:      "twsnmpfk_ai_data_" + id + "_" + ts + ".csv",
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
	var b bytes.Buffer
	if err := req.Df.ToCSV(&b); err != nil {
		return err
	}
	return os.WriteFile(file, b.Bytes(), 0600)
}
