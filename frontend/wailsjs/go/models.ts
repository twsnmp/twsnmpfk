export namespace backend {
	
	export class VPanelPortEnt {
	    Index: number;
	    State: string;
	    Name: string;
	    Speed: number;
	    OutPacktes: number;
	    OutBytes: number;
	    OutError: number;
	    InPacktes: number;
	    InBytes: number;
	    InError: number;
	    Type: number;
	    Admin: number;
	    Oper: number;
	    MAC: string;
	
	    static createFrom(source: any = {}) {
	        return new VPanelPortEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.State = source["State"];
	        this.Name = source["Name"];
	        this.Speed = source["Speed"];
	        this.OutPacktes = source["OutPacktes"];
	        this.OutBytes = source["OutBytes"];
	        this.OutError = source["OutError"];
	        this.InPacktes = source["InPacktes"];
	        this.InBytes = source["InBytes"];
	        this.InError = source["InError"];
	        this.Type = source["Type"];
	        this.Admin = source["Admin"];
	        this.Oper = source["Oper"];
	        this.MAC = source["MAC"];
	    }
	}

}

export namespace datastore {
	
	export class AIConfEnt {
	    HighThreshold: number;
	    LowThreshold: number;
	    WarnThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.HighThreshold = source["HighThreshold"];
	        this.LowThreshold = source["LowThreshold"];
	        this.WarnThreshold = source["WarnThreshold"];
	    }
	}
	export class AIResult {
	    PollingID: string;
	    LastTime: number;
	    ScoreData: number[][];
	
	    static createFrom(source: any = {}) {
	        return new AIResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PollingID = source["PollingID"];
	        this.LastTime = source["LastTime"];
	        this.ScoreData = source["ScoreData"];
	    }
	}
	export class BackImageEnt {
	    X: number;
	    Y: number;
	    Width: number;
	    Height: number;
	    Data: string;
	
	    static createFrom(source: any = {}) {
	        return new BackImageEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	        this.Data = source["Data"];
	    }
	}
	export class DiscoverConfEnt {
	    StartIP: string;
	    EndIP: string;
	    Timeout: number;
	    Retry: number;
	    X: number;
	    Y: number;
	    AddPolling: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiscoverConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StartIP = source["StartIP"];
	        this.EndIP = source["EndIP"];
	        this.Timeout = source["Timeout"];
	        this.Retry = source["Retry"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.AddPolling = source["AddPolling"];
	    }
	}
	export class DrawItemEnt {
	    ID: string;
	    Type: number;
	    X: number;
	    Y: number;
	    W: number;
	    H: number;
	    Color: string;
	    Path: string;
	    Text: string;
	    Size: number;
	    PollingID: string;
	    VarName: string;
	    Format: string;
	    Value: number;
	    Scale: number;
	
	    static createFrom(source: any = {}) {
	        return new DrawItemEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Type = source["Type"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.W = source["W"];
	        this.H = source["H"];
	        this.Color = source["Color"];
	        this.Path = source["Path"];
	        this.Text = source["Text"];
	        this.Size = source["Size"];
	        this.PollingID = source["PollingID"];
	        this.VarName = source["VarName"];
	        this.Format = source["Format"];
	        this.Value = source["Value"];
	        this.Scale = source["Scale"];
	    }
	}
	export class EventLogEnt {
	    Time: number;
	    Type: string;
	    Level: string;
	    NodeName: string;
	    NodeID: string;
	    Event: string;
	    LastLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new EventLogEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.Type = source["Type"];
	        this.Level = source["Level"];
	        this.NodeName = source["NodeName"];
	        this.NodeID = source["NodeID"];
	        this.Event = source["Event"];
	        this.LastLevel = source["LastLevel"];
	    }
	}
	export class GrokEnt {
	    ID: string;
	    Name: string;
	    Decr: string;
	    Pat: string;
	    Ok: string;
	
	    static createFrom(source: any = {}) {
	        return new GrokEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Decr = source["Decr"];
	        this.Pat = source["Pat"];
	        this.Ok = source["Ok"];
	    }
	}
	export class LineEnt {
	    ID: string;
	    NodeID1: string;
	    PollingID1: string;
	    State1: string;
	    NodeID2: string;
	    PollingID2: string;
	    State2: string;
	    PollingID: string;
	    Width: number;
	    State: string;
	    Info: string;
	    Port: string;
	
	    static createFrom(source: any = {}) {
	        return new LineEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.NodeID1 = source["NodeID1"];
	        this.PollingID1 = source["PollingID1"];
	        this.State1 = source["State1"];
	        this.NodeID2 = source["NodeID2"];
	        this.PollingID2 = source["PollingID2"];
	        this.State2 = source["State2"];
	        this.PollingID = source["PollingID"];
	        this.Width = source["Width"];
	        this.State = source["State"];
	        this.Info = source["Info"];
	        this.Port = source["Port"];
	    }
	}
	export class MapConfEnt {
	    MapName: string;
	    PollInt: number;
	    Timeout: number;
	    Retry: number;
	    LogDays: number;
	    SnmpMode: string;
	    Community: string;
	    SnmpUser: string;
	    SnmpPassword: string;
	    EnableSyslogd: boolean;
	    EnableTrapd: boolean;
	    EnableArpWatch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MapConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MapName = source["MapName"];
	        this.PollInt = source["PollInt"];
	        this.Timeout = source["Timeout"];
	        this.Retry = source["Retry"];
	        this.LogDays = source["LogDays"];
	        this.SnmpMode = source["SnmpMode"];
	        this.Community = source["Community"];
	        this.SnmpUser = source["SnmpUser"];
	        this.SnmpPassword = source["SnmpPassword"];
	        this.EnableSyslogd = source["EnableSyslogd"];
	        this.EnableTrapd = source["EnableTrapd"];
	        this.EnableArpWatch = source["EnableArpWatch"];
	    }
	}
	export class NodeEnt {
	    ID: string;
	    Name: string;
	    Descr: string;
	    Icon: string;
	    State: string;
	    X: number;
	    Y: number;
	    IP: string;
	    IPv6: string;
	    MAC: string;
	    SnmpMode: string;
	    Community: string;
	    User: string;
	    Password: string;
	    PublicKey: string;
	    URL: string;
	    AddrMode: string;
	    AutoAck: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NodeEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Descr = source["Descr"];
	        this.Icon = source["Icon"];
	        this.State = source["State"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.IP = source["IP"];
	        this.IPv6 = source["IPv6"];
	        this.MAC = source["MAC"];
	        this.SnmpMode = source["SnmpMode"];
	        this.Community = source["Community"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.PublicKey = source["PublicKey"];
	        this.URL = source["URL"];
	        this.AddrMode = source["AddrMode"];
	        this.AutoAck = source["AutoAck"];
	    }
	}
	export class NotifyConfEnt {
	    MailServer: string;
	    InsecureSkipVerify: boolean;
	    User: string;
	    Password: string;
	    MailTo: string;
	    MailFrom: string;
	    Subject: string;
	    Interval: number;
	    Level: string;
	    Report: boolean;
	    NotifyRepair: boolean;
	    ExecCmd: string;
	
	    static createFrom(source: any = {}) {
	        return new NotifyConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MailServer = source["MailServer"];
	        this.InsecureSkipVerify = source["InsecureSkipVerify"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.MailTo = source["MailTo"];
	        this.MailFrom = source["MailFrom"];
	        this.Subject = source["Subject"];
	        this.Interval = source["Interval"];
	        this.Level = source["Level"];
	        this.Report = source["Report"];
	        this.NotifyRepair = source["NotifyRepair"];
	        this.ExecCmd = source["ExecCmd"];
	    }
	}
	export class PollingEnt {
	    ID: string;
	    Name: string;
	    NodeID: string;
	    Type: string;
	    Mode: string;
	    Params: string;
	    Filter: string;
	    Extractor: string;
	    Script: string;
	    Level: string;
	    PollInt: number;
	    Timeout: number;
	    Retry: number;
	    LogMode: number;
	    NextTime: number;
	    LastTime: number;
	    Result: {[key: string]: any};
	    State: string;
	
	    static createFrom(source: any = {}) {
	        return new PollingEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.NodeID = source["NodeID"];
	        this.Type = source["Type"];
	        this.Mode = source["Mode"];
	        this.Params = source["Params"];
	        this.Filter = source["Filter"];
	        this.Extractor = source["Extractor"];
	        this.Script = source["Script"];
	        this.Level = source["Level"];
	        this.PollInt = source["PollInt"];
	        this.Timeout = source["Timeout"];
	        this.Retry = source["Retry"];
	        this.LogMode = source["LogMode"];
	        this.NextTime = source["NextTime"];
	        this.LastTime = source["LastTime"];
	        this.Result = source["Result"];
	        this.State = source["State"];
	    }
	}
	export class PollingLogEnt {
	    Time: number;
	    PollingID: string;
	    State: string;
	    Result: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new PollingLogEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.PollingID = source["PollingID"];
	        this.State = source["State"];
	        this.Result = source["Result"];
	    }
	}
	export class PollingTemplateEnt {
	    ID: string;
	    Name: string;
	    Level: string;
	    Type: string;
	    Mode: string;
	    Params: string;
	    Filter: string;
	    Extractor: string;
	    Script: string;
	    Descr: string;
	    AutoMode: string;
	
	    static createFrom(source: any = {}) {
	        return new PollingTemplateEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Level = source["Level"];
	        this.Type = source["Type"];
	        this.Mode = source["Mode"];
	        this.Params = source["Params"];
	        this.Filter = source["Filter"];
	        this.Extractor = source["Extractor"];
	        this.Script = source["Script"];
	        this.Descr = source["Descr"];
	        this.AutoMode = source["AutoMode"];
	    }
	}
	export class SyslogEnt {
	    Time: number;
	    Level: string;
	    Host: string;
	    Type: string;
	    Tag: string;
	    Message: string;
	    Severity: number;
	    Facility: number;
	
	    static createFrom(source: any = {}) {
	        return new SyslogEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.Level = source["Level"];
	        this.Host = source["Host"];
	        this.Type = source["Type"];
	        this.Tag = source["Tag"];
	        this.Message = source["Message"];
	        this.Severity = source["Severity"];
	        this.Facility = source["Facility"];
	    }
	}
	export class TrapEnt {
	    Time: number;
	    FromAddress: string;
	    TrapType: string;
	    Variables: string;
	
	    static createFrom(source: any = {}) {
	        return new TrapEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.FromAddress = source["FromAddress"];
	        this.TrapType = source["TrapType"];
	        this.Variables = source["Variables"];
	    }
	}

}

export namespace discover {
	
	export class DiscoverStat {
	    Running: boolean;
	    Total: number;
	    Sent: number;
	    Found: number;
	    Snmp: number;
	    Web: number;
	    Mail: number;
	    SSH: number;
	    File: number;
	    RDP: number;
	    LDAP: number;
	    StartTime: number;
	    Now: number;
	
	    static createFrom(source: any = {}) {
	        return new DiscoverStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Running = source["Running"];
	        this.Total = source["Total"];
	        this.Sent = source["Sent"];
	        this.Found = source["Found"];
	        this.Snmp = source["Snmp"];
	        this.Web = source["Web"];
	        this.Mail = source["Mail"];
	        this.SSH = source["SSH"];
	        this.File = source["File"];
	        this.RDP = source["RDP"];
	        this.LDAP = source["LDAP"];
	        this.StartTime = source["StartTime"];
	        this.Now = source["Now"];
	    }
	}

}

export namespace main {
	
	export class AIList {
	    ID: string;
	    Node: string;
	    Polling: string;
	    Score: number;
	    Count: number;
	    LastTime: number;
	
	    static createFrom(source: any = {}) {
	        return new AIList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Node = source["Node"];
	        this.Polling = source["Polling"];
	        this.Score = source["Score"];
	        this.Count = source["Count"];
	        this.LastTime = source["LastTime"];
	    }
	}
	export class Settings {
	    Kiosk: boolean;
	    Lock: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Kiosk = source["Kiosk"];
	        this.Lock = source["Lock"];
	    }
	}
	export class UpdatePosEnt {
	    ID: string;
	    X: number;
	    Y: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdatePosEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	    }
	}

}

