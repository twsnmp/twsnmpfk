package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/logreport"
)

// GetLogReport returns parsed JSON entities for a given report kind.
func (a *App) GetLogReport(kind string) ([]any, error) {
	kind = logreport.NormalizeKind(kind)
	if !logreport.IsKind(kind) {
		return nil, fmt.Errorf("unknown report kind: %s", kind)
	}
	logger.FlushReportEngine()
	items, err := datastore.ListLogReportData(kind)
	if err != nil {
		return nil, err
	}
	if kind == logreport.KindTwAgentInfo {
		items = deduplicateTwAgentInfo(items)
	}
	list := make([]any, 0, len(items))
	for _, b := range items {
		var obj any
		if err := json.Unmarshal(b, &obj); err == nil {
			list = append(list, obj)
		}
	}
	return list, nil
}

func deduplicateTwAgentInfo(items map[string][]byte) map[string][]byte {
	if len(items) <= 1 {
		return items
	}
	ents := make([]*logreport.TwAgentInfoEnt, 0, len(items))
	for _, b := range items {
		var e logreport.TwAgentInfoEnt
		if err := json.Unmarshal(b, &e); err == nil {
			if e.Host == "" && e.Hostname != "" {
				e.Host = e.Hostname
			}
			if e.Hostname == "" && e.Host != "" {
				e.Hostname = e.Host
			}
			ents = append(ents, &e)
		}
	}

	merged := make([]*logreport.TwAgentInfoEnt, 0)
	toDelete := make([]string, 0)

	for _, e := range ents {
		var match *logreport.TwAgentInfoEnt
		for _, m := range merged {
			if e.AgentID != "" && m.AgentID != "" && strings.EqualFold(e.AgentID, m.AgentID) {
				match = m
				break
			}
			if e.IP != "" && m.IP != "" && net.ParseIP(e.IP) != nil && net.ParseIP(m.IP) != nil && e.IP == m.IP {
				match = m
				break
			}
			if e.IP != "" && m.Hostname != "" && strings.EqualFold(e.IP, m.Hostname) {
				match = m
				break
			}
			if m.IP != "" && e.Hostname != "" && strings.EqualFold(m.IP, e.Hostname) {
				match = m
				break
			}
			if e.Hostname != "" && m.Hostname != "" && strings.EqualFold(e.Hostname, m.Hostname) {
				match = m
				break
			}
			eIsIP := net.ParseIP(e.IP) != nil
			mIsIP := net.ParseIP(m.IP) != nil
			if (mIsIP && m.Hostname == "" && !eIsIP && e.Hostname != "") ||
				(eIsIP && e.Hostname == "" && !mIsIP && m.Hostname != "") {
				match = m
				break
			}
		}

		if match == nil {
			merged = append(merged, e)
		} else {
			if (match.Hostname == "" || match.Hostname == match.IP) && e.Hostname != "" {
				match.Hostname = e.Hostname
				match.Host = e.Hostname
			}
			if net.ParseIP(match.IP) == nil && net.ParseIP(e.IP) != nil {
				match.IP = e.IP
			}
			if match.AgentID == "" && e.AgentID != "" {
				match.AgentID = e.AgentID
			}
			if match.MAC == "" && e.MAC != "" {
				match.MAC = e.MAC
			}
			if match.OS == "" && e.OS != "" {
				match.OS = e.OS
			}
			if match.Arch == "" && e.Arch != "" {
				match.Arch = e.Arch
			}
			if match.Version == "" && e.Version != "" {
				match.Version = e.Version
			}
			if match.CPUPercent == 0 && e.CPUPercent > 0 {
				match.CPUPercent = e.CPUPercent
			}
			if match.MemPercent == 0 && e.MemPercent > 0 {
				match.MemPercent = e.MemPercent
			}
			if match.DiskPercent == 0 && e.DiskPercent > 0 {
				match.DiskPercent = e.DiskPercent
			}
			if len(match.Disks) == 0 && len(e.Disks) > 0 {
				match.Disks = e.Disks
			}
			if len(match.Interfaces) == 0 && len(e.Interfaces) > 0 {
				match.Interfaces = e.Interfaces
			}
			if len(match.History) == 0 && len(e.History) > 0 {
				match.History = e.History
			}
			if e.FirstTime > 0 && (match.FirstTime == 0 || e.FirstTime < match.FirstTime) {
				match.FirstTime = e.FirstTime
			}
			if e.LastTime > match.LastTime {
				match.LastTime = e.LastTime
			}
			for _, p := range e.Protocols {
				found := false
				for _, mp := range match.Protocols {
					if mp == p {
						found = true
						break
					}
				}
				if !found && p != "" {
					match.Protocols = append(match.Protocols, p)
				}
			}
			toDelete = append(toDelete, e.ID)
		}
	}

	if len(toDelete) > 0 {
		_ = datastore.DeleteLogReportData(logreport.KindTwAgentInfo, toDelete)
		toSave := make(map[string][]byte)
		for _, m := range merged {
			if b, err := json.Marshal(m); err == nil {
				toSave[m.ID] = b
			}
		}
		if len(toSave) > 0 {
			_ = datastore.SaveLogReportData(logreport.KindTwAgentInfo, toSave)
		}
		return toSave
	}
	return items
}

// DeleteLogReport removes specific report entities.
func (a *App) DeleteLogReport(kind string, ids []string) error {
	kind = logreport.NormalizeKind(kind)
	if !logreport.IsKind(kind) {
		return fmt.Errorf("unknown report kind: %s", kind)
	}
	logger.FlushReportEngine()
	resolved := make([]string, 0, len(ids))
	var allItems map[string][]byte
	for _, id := range ids {
		if b, _ := datastore.GetLogReportData(kind, id); b != nil {
			resolved = append(resolved, id)
			continue
		}
		if allItems == nil {
			allItems, _ = datastore.ListLogReportData(kind)
		}
		for k, v := range allItems {
			var m map[string]any
			if json.Unmarshal(v, &m) == nil {
				if addr, ok := m["Address"].(string); ok && addr == id {
					resolved = append(resolved, k)
				} else if entID, ok := m["ID"].(string); ok && entID == id {
					resolved = append(resolved, k)
				}
			}
		}
	}
	if len(resolved) == 0 {
		resolved = ids
	}
	return datastore.DeleteLogReportData(kind, resolved)
}

// ResetLogReport clears all entries for a report kind.
func (a *App) ResetLogReport(kind string) error {
	kind = logreport.NormalizeKind(kind)
	if !logreport.IsKind(kind) {
		return fmt.Errorf("unknown report kind: %s", kind)
	}
	logger.FlushReportEngine()
	return datastore.ResetLogReportData(kind)
}

// SetLogReportName updates the user-assigned name for an entity (BLE device, Env sensor).
func (a *App) SetLogReportName(kind, id, name string) error {
	kind = logreport.NormalizeKind(kind)
	switch kind {
	case "env":
		kind = logreport.KindEnvMonitor
	case "power":
		kind = logreport.KindPowerMonitor
	case "motion":
		kind = logreport.KindMotionSensor
	case "bluetooth":
		kind = logreport.KindBlueDevice
	}
	logger.FlushReportEngine()
	targetKey := id
	b, err := datastore.GetLogReportData(kind, id)
	if err != nil || b == nil {
		// Fallback: search by Address field
		if items, lerr := datastore.ListLogReportData(kind); lerr == nil {
			for k, v := range items {
				var m map[string]any
				if json.Unmarshal(v, &m) == nil {
					if addr, ok := m["Address"].(string); ok && addr == id {
						targetKey = k
						b = v
						break
					}
				}
			}
		}
	}
	if b == nil {
		return fmt.Errorf("entity not found")
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	m["Name"] = name
	nb, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return datastore.SaveLogReportData(kind, map[string][]byte{targetKey: nb})
}

// GetSensors returns all registered remote sensors.
func (a *App) GetSensors() []*datastore.SensorEnt {
	list := datastore.ListSensors()
	if list == nil {
		return []*datastore.SensorEnt{}
	}
	return list
}

// GetSensor returns a single sensor by ID.
func (a *App) GetSensor(id string) (*datastore.SensorEnt, error) {
	return datastore.GetSensor(id)
}

// DeleteSensors removes specific sensors by IDs.
func (a *App) DeleteSensors(ids []string) error {
	return datastore.DeleteSensors(ids)
}

// DeleteAllSensors purges all sensors.
func (a *App) DeleteAllSensors() error {
	return datastore.DeleteAllSensors()
}

// ToggleSensorIgnore toggles the ignore flag on a sensor.
func (a *App) ToggleSensorIgnore(id string) error {
	return datastore.ToggleSensorIgnore(id)
}
