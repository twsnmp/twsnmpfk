package logreport

// twBlueScan record types: Device, OMRONEnv, SwitchBotEnv, InkbirdEnv,
// SwitchBotPlugMini, SwitchBotMotionSensor.
func (s *Session) processBlue(r Record, t string, m map[string]string) bool {
	addr := m["address"]
	if addr == "" {
		return false
	}
	switch t {
	case "Device":
		return s.blueDevice(r, addr, m)
	case "OMRONEnv", "SwitchBotEnv", "InkbirdEnv":
		return s.blueEnv(r, t, addr, m)
	case "SwitchBotPlugMini":
		return s.bluePlug(r, addr, m)
	case "SwitchBotMotionSensor":
		return s.blueMotion(r, addr, m)
	}
	return false
}

// type=Device,address=%s,name=%s,rssi=%d,addrType=%s,vendor=%s,info=%s,ft=%s,lt=%s
func (s *Session) blueDevice(r Record, addr string, m map[string]string) bool {
	lt := parseTime(m["lt"], r.Time)
	rssi := atoi(m["rssi"])
	id := makeID(r.Host + ":" + addr)
	info := m["info"]
	if u, ok := m["uuid"]; ok {
		info += " UUID:" + u
	}
	if e := getEnt[BlueDeviceEnt](s, KindBlueDevice, id); e != nil {
		e.Count++
		if lt > e.LastTime {
			e.LastTime = lt
		}
		if info != "" && info != e.Info {
			e.Info = info
		}
		if v := m["vendor"]; v != "" && v != e.Vendor {
			e.Vendor = v
		}
		e.RSSI = appendLimit(e.RSSI, RSSIEnt{Value: rssi, Time: lt}, MaxSeriesSize)
		putEnt(s, KindBlueDevice, id, e)
		return true
	}
	putEnt(s, KindBlueDevice, id, &BlueDeviceEnt{
		ID:          id,
		Host:        r.Host,
		Address:     addr,
		AddressType: m["addrType"],
		Name:        m["name"],
		Count:       1,
		RSSI:        []RSSIEnt{{Value: rssi, Time: lt}},
		Vendor:      m["vendor"],
		Info:        info,
		FirstTime:   parseTime(m["ft"], lt),
		LastTime:    lt,
	})
	return true
}

// type=OMRONEnv,address=%s,name=%s,rssi=%d,seq=%d,temp=%.02f,hum=%.02f,lx=%d,press=%.02f,sound=%.02f,eTVOC=%d,eCO2=%d
// type=SwitchBotEnv|InkbirdEnv,address=%s,name=%s,rssi=%d,temp=%.02f,hum=%.02f,bat=%d,co2=%d
func (s *Session) blueEnv(r Record, t, addr string, m map[string]string) bool {
	d := EnvDataEnt{
		Time:     r.Time,
		RSSI:     atoi(m["rssi"]),
		Temp:     atof(m["temp"]),
		Humidity: atof(m["hum"]),
	}
	if t == "OMRONEnv" {
		d.Illuminance = atof(m["lx"])
		d.BarometricPressure = atof(m["press"])
		d.Sound = atof(m["sound"])
		d.ETVOC = atof(m["eTVOC"])
		d.ECo2 = atof(m["eCO2"])
	} else {
		d.Battery = atoi(m["bat"])
		d.ECo2 = atof(m["co2"])
	}
	id := makeID(r.Host + ":" + addr)
	if e := getEnt[EnvMonitorEnt](s, KindEnvMonitor, id); e != nil {
		e.Count++
		if r.Time > e.LastTime {
			e.LastTime = r.Time
		}
		e.EnvData = appendLimit(e.EnvData, d, MaxSeriesSize)
		putEnt(s, KindEnvMonitor, id, e)
		return true
	}
	putEnt(s, KindEnvMonitor, id, &EnvMonitorEnt{
		ID:        id,
		Host:      r.Host,
		Address:   addr,
		Name:      m["name"],
		Count:     1,
		EnvData:   []EnvDataEnt{d},
		FirstTime: r.Time,
		LastTime:  r.Time,
	})
	return true
}

// type=SwitchBotPlugMini,address=%s,name=%s,rssi=%d,sw=%v,over=%v,load=%d
func (s *Session) bluePlug(r Record, addr string, m map[string]string) bool {
	if _, ok := m["load"]; !ok {
		return false
	}
	d := PowerMonitorDataEnt{
		Time:   r.Time,
		Load:   atof(m["load"]) / 10.0,
		Switch: m["sw"] == "true",
		Over:   m["over"] == "true",
		RSSI:   atoi(m["rssi"]),
	}
	id := makeID(r.Host + ":" + addr)
	if e := getEnt[PowerMonitorEnt](s, KindPowerMonitor, id); e != nil {
		e.Count++
		if r.Time > e.LastTime {
			e.LastTime = r.Time
		}
		e.Data = appendLimit(e.Data, d, MaxSeriesSize)
		putEnt(s, KindPowerMonitor, id, e)
		return true
	}
	putEnt(s, KindPowerMonitor, id, &PowerMonitorEnt{
		ID:        id,
		Host:      r.Host,
		Address:   addr,
		Name:      m["name"],
		Count:     1,
		Data:      []PowerMonitorDataEnt{d},
		FirstTime: r.Time,
		LastTime:  r.Time,
	})
	return true
}

// type=SwitchBotMotionSensor,address=%s,name=%s,rssi=%d,moving=%v,event=%s,lastMoveDiff=%d,lastMove=%s,battery=%d,light=%v
func (s *Session) blueMotion(r Record, addr string, m map[string]string) bool {
	if _, ok := m["moving"]; !ok {
		return false
	}
	d := MotionSensorDataEnt{
		Time:         r.Time,
		Event:        m["event"],
		Moving:       m["moving"] == "true",
		Light:        m["light"] == "true",
		Battery:      atoi(m["battery"]),
		LastMove:     parseTime(m["lastMove"], 0),
		LastMoveDiff: atoi(m["lastMoveDiff"]),
		RSSI:         atoi(m["rssi"]),
	}
	id := makeID(r.Host + ":" + addr)
	if e := getEnt[MotionSensorEnt](s, KindMotionSensor, id); e != nil {
		e.Count++
		if r.Time > e.LastTime {
			e.LastTime = r.Time
		}
		e.Data = appendLimit(e.Data, d, MaxSeriesSize*2)
		putEnt(s, KindMotionSensor, id, e)
		return true
	}
	putEnt(s, KindMotionSensor, id, &MotionSensorEnt{
		ID:        id,
		Host:      r.Host,
		Address:   addr,
		Name:      m["name"],
		Count:     1,
		Data:      []MotionSensorDataEnt{d},
		FirstTime: r.Time,
		LastTime:  r.Time,
	})
	return true
}
