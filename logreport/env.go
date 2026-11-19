package logreport

import (
	"encoding/json"
	"sync"

	"github.com/twsnmp/twsnmpfk/datastore"
)

var earthEnvMu sync.Mutex

// RecordEarthEnvSample persists a sample from an open environmental/disaster/space API polling
// into the KindEarthEnv report bucket, and optionally mirrors weather data into KindEnvMonitor.
func RecordEarthEnvSample(
	pollingID, nodeID, host, mode, name, location, summary, state string,
	sample EarthEnvDataEnt,
	syncEnvMonitor bool,
) error {
	if pollingID == "" {
		return nil
	}
	earthEnvMu.Lock()
	defer earthEnvMu.Unlock()

	now := sample.Time
	if now <= 0 {
		return nil
	}
	id := makeID("earth:" + pollingID)
	address := mode + ":" + location
	if location == "" {
		address = mode + ":" + host
	}

	var ent EarthEnvMonitorEnt
	if b, err := datastore.GetLogReportData(KindEarthEnv, id); err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &ent)
	}
	if ent.ID == "" {
		ent = EarthEnvMonitorEnt{
			ID:        id,
			PollingID: pollingID,
			NodeID:    nodeID,
			Host:      host,
			Address:   address,
			Mode:      mode,
			Name:      name,
			Location:  location,
			Summary:   summary,
			State:     state,
			Count:     1,
			Data:      []EarthEnvDataEnt{sample},
			FirstTime: now,
			LastTime:  now,
		}
	} else {
		ent.PollingID = pollingID
		ent.NodeID = nodeID
		ent.Host = host
		ent.Address = address
		ent.Mode = mode
		if ent.Name == "" {
			ent.Name = name
		}
		ent.Location = location
		ent.Summary = summary
		ent.State = state
		ent.Count++
		if now > ent.LastTime {
			ent.LastTime = now
		}
		ent.Data = appendLimit(ent.Data, sample, MaxSeriesSize)
	}

	if b, err := json.Marshal(&ent); err == nil {
		_ = datastore.SaveLogReportData(KindEarthEnv, map[string][]byte{id: b})
	}

	// Mirror Open-Meteo temperature/humidity/pressure to standard KindEnvMonitor so it appears alongside BLE sensors
	if mode == "weather" && syncEnvMonitor {
		envAddr := "open-meteo:" + location
		if location == "" {
			envAddr = "open-meteo:" + host
		}
		envID := makeID(host + ":" + envAddr)
		envSample := EnvDataEnt{
			Time:               now,
			Temp:               sample.Temp,
			Humidity:           sample.Humidity,
			BarometricPressure: sample.Pressure,
		}
		var envEnt EnvMonitorEnt
		if b, err := datastore.GetLogReportData(KindEnvMonitor, envID); err == nil && len(b) > 0 {
			_ = json.Unmarshal(b, &envEnt)
		}
		if envEnt.ID == "" {
			envEnt = EnvMonitorEnt{
				ID:        envID,
				Host:      host,
				Address:   envAddr,
				Name:      name,
				Count:     1,
				EnvData:   []EnvDataEnt{envSample},
				FirstTime: now,
				LastTime:  now,
			}
		} else {
			envEnt.Host = host
			if envEnt.Name == "" {
				envEnt.Name = name
			}
			envEnt.Count++
			if now > envEnt.LastTime {
				envEnt.LastTime = now
			}
			envEnt.EnvData = appendLimit(envEnt.EnvData, envSample, MaxSeriesSize)
		}
		if b, err := json.Marshal(&envEnt); err == nil {
			_ = datastore.SaveLogReportData(KindEnvMonitor, map[string][]byte{envID: b})
		}
	}
	return nil
}
