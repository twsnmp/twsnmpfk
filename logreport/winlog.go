package logreport

import "fmt"

func (s *Session) processWinLog(r Record, t string, m map[string]string) bool {
	switch t {
	case "EventID":
		return s.winEventID(r, m)
	case "Logon", "Logoff", "LogonFailed":
		return s.winLogon(r, t, m)
	case "Account":
		return s.winAccount(r, m)
	case "Kerberos":
		return s.winKerberos(r, m)
	case "Privilege":
		return s.winPrivilege(r, m)
	case "Process":
		return s.winProcess(r, m)
	case "Task":
		return s.winTask(r, m)
	}
	return false
}

// type=EventID,computer=%s,channel=%s,provider=%s,eventID=%d,total=%d,count=%d,ft=%s,lt=%s
func (s *Session) winEventID(r Record, m map[string]string) bool {
	eventID := atoi(m["eventID"])
	count := atoi(m["count"])
	if eventID < 1 || count < 1 {
		return false
	}
	level := "normal"
	if r.Severity < 4 {
		level = "error"
	} else if r.Severity == 4 {
		level = "warn"
	}
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s:%d", m["computer"], m["provider"], eventID))
	if e := getEnt[WinEventIDEnt](s, KindWinEventID, id); e != nil {
		if e.LastTime < lt {
			e.LastTime = lt
		}
		if level == "error" || (e.Level != "error" && level == "warn") {
			e.Level = level
		}
		e.Count += count
		putEnt(s, KindWinEventID, id, e)
		return true
	}
	if total := atoi(m["total"]); total > count {
		count = total
	}
	putEnt(s, KindWinEventID, id, &WinEventIDEnt{
		ID:        id,
		Level:     level,
		Provider:  m["provider"],
		EventID:   eventID,
		Computer:  m["computer"],
		Channel:   m["channel"],
		Count:     count,
		FirstTime: parseTime(m["ft"], lt),
		LastTime:  lt,
	})
	return true
}

// type=Logon|Logoff|LogonFailed,subject=%s,target=%s,computer=%s,ip=%s,logonType=%s,failedCode=%s,time=%s
func (s *Session) winLogon(r Record, tp string, m map[string]string) bool {
	target, ok := m["target"]
	if !ok {
		return false
	}
	computer := m["computer"]
	logonType := m["logonType"]
	failedCode := m["failedCode"]
	ts := parseTime(m["time"], r.Time)
	if tp == "Logoff" {
		for id, e := range allEnts[WinLogonEnt](s, KindWinLogon) {
			if e.Computer == computer && e.Target == target && e.LogonType[logonType] > 0 {
				e.Logoff++
				if e.LastTime < ts {
					e.LastTime = ts
				}
				putEnt(s, KindWinLogon, id, e)
				return true
			}
		}
		return false
	}
	id := makeID(fmt.Sprintf("%s:%s:%s", target, computer, m["ip"]))
	e := getEnt[WinLogonEnt](s, KindWinLogon, id)
	if e == nil {
		e = &WinLogonEnt{
			ID:         id,
			Target:     target,
			Computer:   computer,
			IP:         m["ip"],
			FirstTime:  ts,
			LastTime:   ts,
			LogonType:  map[string]int{},
			FailedCode: map[string]int{},
		}
	}
	if e.LogonType == nil {
		e.LogonType = map[string]int{}
	}
	if e.FailedCode == nil {
		e.FailedCode = map[string]int{}
	}
	if e.LastTime < ts {
		e.LastTime = ts
	}
	e.Count++
	if tp == "LogonFailed" {
		e.Failed++
		if failedCode != "" {
			e.FailedCode[failedCode]++
		}
	} else {
		e.Logon++
		if logonType != "" {
			e.LogonType[logonType]++
		}
	}
	setFailedPenalty(&e.ScoreInfo, e.Failed, e.Count)
	putEnt(s, KindWinLogon, id, e)
	return true
}

func setFailedPenalty(si *ScoreInfo, failed, count int) {
	si.Penalty = 0
	if failed > 0 {
		si.Penalty = 1
	}
	if count > 0 {
		si.Penalty += (10 * failed) / count
	}
}

// type=Account,subject=%s,target=%s,computer=%s,count=%d,edit=%d,password=%d,other=%d,ft=%s,lt=%s
func (s *Session) winAccount(r Record, m map[string]string) bool {
	target, ok := m["target"]
	count := atoi(m["count"])
	if !ok || count < 1 {
		return false
	}
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s:%s", target, m["computer"], m["subject"]))
	if e := getEnt[WinAccountEnt](s, KindWinAccount, id); e != nil {
		e.LastTime = lt
		e.Count += count
		e.Edit += atoi(m["edit"])
		e.Password += atoi(m["password"])
		e.Other += atoi(m["other"])
		putEnt(s, KindWinAccount, id, e)
		return true
	}
	putEnt(s, KindWinAccount, id, &WinAccountEnt{
		ID:        id,
		Subject:   m["subject"],
		Target:    target,
		Computer:  m["computer"],
		Count:     count,
		Edit:      atoi(m["edit"]),
		Password:  atoi(m["password"]),
		Other:     atoi(m["other"]),
		FirstTime: parseTime(m["ft"], lt),
		LastTime:  lt,
	})
	return true
}

// type=Kerberos,target=%s,computer=%s,ip=%s,service=%s,ticketType=%s,count=%d,failed=%d,status=%s,cert=%s,ft=%s,lt=%s
func (s *Session) winKerberos(r Record, m map[string]string) bool {
	target, ok := m["target"]
	count := atoi(m["count"])
	if !ok || count < 1 {
		return false
	}
	failed := atoi(m["failed"])
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s:%s:%s:%s", target, m["computer"], m["ip"], m["service"], m["ticketType"]))
	e := getEnt[WinKerberosEnt](s, KindWinKerberos, id)
	if e != nil {
		e.LastTime = lt
		e.Count += count
		e.Failed += failed
	} else {
		e = &WinKerberosEnt{
			ID:         id,
			Target:     target,
			Computer:   m["computer"],
			Service:    m["service"],
			TicketType: m["ticketType"],
			IP:         m["ip"],
			Count:      count,
			Failed:     failed,
			FirstTime:  parseTime(m["ft"], lt),
			LastTime:   lt,
		}
	}
	setFailedPenalty(&e.ScoreInfo, e.Failed, e.Count)
	putEnt(s, KindWinKerberos, id, e)
	return true
}

// type=Privilege,subject=%s,computer=%s,count=%d,ft=%s,lt=%s
func (s *Session) winPrivilege(r Record, m map[string]string) bool {
	subject, ok := m["subject"]
	count := atoi(m["count"])
	if !ok || count < 1 {
		return false
	}
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s", subject, m["computer"]))
	if e := getEnt[WinPrivilegeEnt](s, KindWinPrivilege, id); e != nil {
		e.LastTime = lt
		e.Count += count
		putEnt(s, KindWinPrivilege, id, e)
		return true
	}
	putEnt(s, KindWinPrivilege, id, &WinPrivilegeEnt{
		ID:        id,
		Subject:   subject,
		Computer:  m["computer"],
		Count:     count,
		FirstTime: parseTime(m["ft"], lt),
		LastTime:  lt,
	})
	return true
}

// type=Process,computer=%s,process=%s,count=%d,start=%d,exit=%d,subject=%s,status=%s,parent=%s,ft=%s,lt=%s
func (s *Session) winProcess(r Record, m map[string]string) bool {
	process, ok := m["process"]
	count := atoi(m["count"])
	if !ok || count < 1 {
		return false
	}
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s", process, m["computer"]))
	if e := getEnt[WinProcessEnt](s, KindWinProcess, id); e != nil {
		e.LastTime = lt
		e.Count += count
		e.Start += atoi(m["start"])
		e.Exit += atoi(m["exit"])
		if v := m["subject"]; v != "" {
			e.LastSubject = v
		}
		if v := m["status"]; v != "" {
			e.LastStatus = v
		}
		if v := m["parent"]; v != "" {
			e.LastParent = v
		}
		putEnt(s, KindWinProcess, id, e)
		return true
	}
	putEnt(s, KindWinProcess, id, &WinProcessEnt{
		ID:          id,
		Process:     process,
		Computer:    m["computer"],
		Count:       count,
		Start:       atoi(m["start"]),
		Exit:        atoi(m["exit"]),
		LastSubject: m["subject"],
		LastParent:  m["parent"],
		LastStatus:  m["status"],
		FirstTime:   parseTime(m["ft"], lt),
		LastTime:    lt,
	})
	return true
}

// type=Task,subject=%s,taskname=%s,computer=%s,count=%d,ft=%s,lt=%s
func (s *Session) winTask(r Record, m map[string]string) bool {
	task, ok := m["taskname"]
	count := atoi(m["count"])
	if !ok || count < 1 {
		return false
	}
	lt := parseTime(m["lt"], r.Time)
	id := makeID(fmt.Sprintf("%s:%s:%s", task, m["computer"], m["subject"]))
	if e := getEnt[WinTaskEnt](s, KindWinTask, id); e != nil {
		e.LastTime = lt
		e.Count += count
		putEnt(s, KindWinTask, id, e)
		return true
	}
	putEnt(s, KindWinTask, id, &WinTaskEnt{
		ID:        id,
		TaskName:  task,
		Computer:  m["computer"],
		Subject:   m["subject"],
		Count:     count,
		FirstTime: parseTime(m["ft"], lt),
		LastTime:  lt,
	})
	return true
}
