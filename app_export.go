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
	Title  string
	Header []string
	Data   [][]interface{}
	Image  string
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
func (a *App) ExportEventLogs(t string) string {
	data := ExportData{
		Title:  "TWSNMP Event Log",
		Header: []string{"Level", "Time", "Type", "Node Name", "Event"},
	}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
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
		log.Printf("ExportTable err=%v", err)
		return fmt.Sprintf("export err=%v", err)
	}
	return ""
}

// ExportSyslogs  export syslogs
func (a *App) ExportSyslogs(t string) string {
	data := ExportData{
		Title:  "TWSNMP Syslog",
		Header: []string{"Level", "Time", "Host", "Type", "Tag", "Message"},
	}
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
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
		log.Printf("ExportTable err=%v", err)
		return fmt.Sprintf("export err=%v", err)
	}
	return ""
}

// ExportTrap  export traps
func (a *App) ExportTraps(t string) string {
	data := ExportData{
		Title:  "TWSNMP TRAP",
		Header: []string{"Time", "From", "Type", "Variables"},
	}
	datastore.ForEachLastTraps(func(l *datastore.TrapEnt) bool {
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
		log.Printf("ExportTable err=%v", err)
		return fmt.Sprintf("export err=%v", err)
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
