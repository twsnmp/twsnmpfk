package logreport

// twWifiScan: type=APInfo,ssid=%s,bssid=%s,rssi=%s,Channel=%s,info=%s,count=%d,change=%d,ft=%s,lt=%s
func (s *Session) processWifi(r Record, t string, m map[string]string) bool {
	if t != "APInfo" {
		return false
	}
	bssid := m["bssid"]
	if bssid == "" {
		return false
	}
	rssi := atoi(m["rssi"])
	id := r.Host + ":" + bssid
	now := r.Time
	if e := getEnt[WifiAPEnt](s, KindWifiAP, id); e != nil {
		e.Count++
		if e.SSID != m["ssid"] || e.Channel != m["Channel"] || e.Info != m["info"] {
			e.Change++
		}
		if e.Vendor == "" {
			e.Vendor = s.store.FindVendor(bssid)
		}
		e.SSID = m["ssid"]
		e.Channel = m["Channel"]
		e.Info = m["info"]
		if now > e.LastTime {
			e.LastTime = now
		}
		e.RSSI = appendLimit(e.RSSI, RSSIEnt{Value: rssi, Time: now}, MaxSeriesSize)
		putEnt(s, KindWifiAP, id, e)
		return true
	}
	putEnt(s, KindWifiAP, id, &WifiAPEnt{
		ID:        id,
		Host:      r.Host,
		BSSID:     bssid,
		SSID:      m["ssid"],
		Channel:   m["Channel"],
		Info:      m["info"],
		Vendor:    s.store.FindVendor(bssid),
		Count:     1,
		RSSI:      []RSSIEnt{{Value: rssi, Time: now}},
		FirstTime: now,
		LastTime:  now,
	})
	return true
}

func appendLimit[T any](list []T, v T, limit int) []T {
	list = append(list, v)
	if len(list) > limit {
		list = list[len(list)-limit:]
	}
	return list
}
