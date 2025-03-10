export namespace backend {
	
	export class FindNeighborNetworksAndLinesResp {
	    Networks: datastore.NetworkEnt[];
	    Lines: datastore.LineEnt[];
	
	    static createFrom(source: any = {}) {
	        return new FindNeighborNetworksAndLinesResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Networks = this.convertValues(source["Networks"], datastore.NetworkEnt);
	        this.Lines = this.convertValues(source["Lines"], datastore.LineEnt);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HrProcess {
	    PID: string;
	    Name: string;
	    Type: string;
	    Status: string;
	    Path: string;
	    Param: string;
	    CPU: number;
	    Mem: number;
	
	    static createFrom(source: any = {}) {
	        return new HrProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PID = source["PID"];
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Status = source["Status"];
	        this.Path = source["Path"];
	        this.Param = source["Param"];
	        this.CPU = source["CPU"];
	        this.Mem = source["Mem"];
	    }
	}
	export class HrFileSystem {
	    Index: string;
	    Type: string;
	    Mount: string;
	    Remote: string;
	    Bootable: number;
	    Access: number;
	
	    static createFrom(source: any = {}) {
	        return new HrFileSystem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Type = source["Type"];
	        this.Mount = source["Mount"];
	        this.Remote = source["Remote"];
	        this.Bootable = source["Bootable"];
	        this.Access = source["Access"];
	    }
	}
	export class HrDevice {
	    Index: string;
	    Type: string;
	    Descr: string;
	    Status: string;
	    Errors: string;
	
	    static createFrom(source: any = {}) {
	        return new HrDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Type = source["Type"];
	        this.Descr = source["Descr"];
	        this.Status = source["Status"];
	        this.Errors = source["Errors"];
	    }
	}
	export class HrStorage {
	    Index: string;
	    Type: string;
	    Descr: string;
	    Size: number;
	    Used: number;
	    Unit: number;
	    Rate: number;
	
	    static createFrom(source: any = {}) {
	        return new HrStorage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Type = source["Type"];
	        this.Descr = source["Descr"];
	        this.Size = source["Size"];
	        this.Used = source["Used"];
	        this.Unit = source["Unit"];
	        this.Rate = source["Rate"];
	    }
	}
	export class HrSystem {
	    Index: number;
	    Key: string;
	    Value: string;
	
	    static createFrom(source: any = {}) {
	        return new HrSystem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Key = source["Key"];
	        this.Value = source["Value"];
	    }
	}
	export class HostResourceEnt {
	    System: HrSystem[];
	    Storage: HrStorage[];
	    Device: HrDevice[];
	    FileSystem: HrFileSystem[];
	    Process: HrProcess[];
	
	    static createFrom(source: any = {}) {
	        return new HostResourceEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.System = this.convertValues(source["System"], HrSystem);
	        this.Storage = this.convertValues(source["Storage"], HrStorage);
	        this.Device = this.convertValues(source["Device"], HrDevice);
	        this.FileSystem = this.convertValues(source["FileSystem"], HrFileSystem);
	        this.Process = this.convertValues(source["Process"], HrProcess);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MonitorDataEnt {
	    Time: number;
	    CPU: number;
	    Mem: number;
	    MyCPU: number;
	    MyMem: number;
	    Swap: number;
	    Disk: number;
	    Load: number;
	    Bytes: number;
	    Net: number;
	    Conn: number;
	    Proc: number;
	    DBSize: number;
	    HeapAlloc: number;
	    Sys: number;
	    NumGoroutine: number;
	
	    static createFrom(source: any = {}) {
	        return new MonitorDataEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.CPU = source["CPU"];
	        this.Mem = source["Mem"];
	        this.MyCPU = source["MyCPU"];
	        this.MyMem = source["MyMem"];
	        this.Swap = source["Swap"];
	        this.Disk = source["Disk"];
	        this.Load = source["Load"];
	        this.Bytes = source["Bytes"];
	        this.Net = source["Net"];
	        this.Conn = source["Conn"];
	        this.Proc = source["Proc"];
	        this.DBSize = source["DBSize"];
	        this.HeapAlloc = source["HeapAlloc"];
	        this.Sys = source["Sys"];
	        this.NumGoroutine = source["NumGoroutine"];
	    }
	}
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
	export class ArpEnt {
	    IP: string;
	    MAC: string;
	    NodeID: string;
	    Vendor: string;
	    FirstTime: number;
	    LastTime: number;
	
	    static createFrom(source: any = {}) {
	        return new ArpEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IP = source["IP"];
	        this.MAC = source["MAC"];
	        this.NodeID = source["NodeID"];
	        this.Vendor = source["Vendor"];
	        this.FirstTime = source["FirstTime"];
	        this.LastTime = source["LastTime"];
	    }
	}
	export class BackImageEnt {
	    X: number;
	    Y: number;
	    Width: number;
	    Height: number;
	    Path: string;
	
	    static createFrom(source: any = {}) {
	        return new BackImageEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.Width = source["Width"];
	        this.Height = source["Height"];
	        this.Path = source["Path"];
	    }
	}
	export class CreateCAReq {
	    RootCAKeyType: string;
	    Name: string;
	    SANs: string;
	    AcmePort: number;
	    HttpBaseURL: string;
	    AcmeBaseURL: string;
	    HttpPort: number;
	    RootCATerm: number;
	    CrlInterval: number;
	    CertTerm: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateCAReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RootCAKeyType = source["RootCAKeyType"];
	        this.Name = source["Name"];
	        this.SANs = source["SANs"];
	        this.AcmePort = source["AcmePort"];
	        this.HttpBaseURL = source["HttpBaseURL"];
	        this.AcmeBaseURL = source["AcmeBaseURL"];
	        this.HttpPort = source["HttpPort"];
	        this.RootCATerm = source["RootCATerm"];
	        this.CrlInterval = source["CrlInterval"];
	        this.CertTerm = source["CertTerm"];
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
	    PortScan: boolean;
	    ReCheck: boolean;
	    AddNetwork: boolean;
	
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
	        this.PortScan = source["PortScan"];
	        this.ReCheck = source["ReCheck"];
	        this.AddNetwork = source["AddNetwork"];
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
	    Cond: number;
	    Values: number[];
	
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
	        this.Cond = source["Cond"];
	        this.Values = source["Values"];
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
	export class IconEnt {
	    Name: string;
	    Icon: string;
	    Code: number;
	
	    static createFrom(source: any = {}) {
	        return new IconEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Icon = source["Icon"];
	        this.Code = source["Code"];
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
	export class LocConfEnt {
	    Style: string;
	    Center: string;
	    Zoom: number;
	    IconSize: number;
	
	    static createFrom(source: any = {}) {
	        return new LocConfEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Style = source["Style"];
	        this.Center = source["Center"];
	        this.Zoom = source["Zoom"];
	        this.IconSize = source["IconSize"];
	    }
	}
	export class MIBInfo {
	    OID: string;
	    Status: string;
	    Type: string;
	    Enum: string;
	    Defval: string;
	    Units: string;
	    Index: string;
	    Description: string;
	    EnumMap: {[key: number]: string};
	    Hint: string;
	
	    static createFrom(source: any = {}) {
	        return new MIBInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OID = source["OID"];
	        this.Status = source["Status"];
	        this.Type = source["Type"];
	        this.Enum = source["Enum"];
	        this.Defval = source["Defval"];
	        this.Units = source["Units"];
	        this.Index = source["Index"];
	        this.Description = source["Description"];
	        this.EnumMap = source["EnumMap"];
	        this.Hint = source["Hint"];
	    }
	}
	export class MIBModuleEnt {
	    Type: string;
	    File: string;
	    Name: string;
	    Error: string;
	
	    static createFrom(source: any = {}) {
	        return new MIBModuleEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.File = source["File"];
	        this.Name = source["Name"];
	        this.Error = source["Error"];
	    }
	}
	export class MIBTreeEnt {
	    oid: string;
	    name: string;
	    MIBInfo?: MIBInfo;
	    children: MIBTreeEnt[];
	
	    static createFrom(source: any = {}) {
	        return new MIBTreeEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oid = source["oid"];
	        this.name = source["name"];
	        this.MIBInfo = this.convertValues(source["MIBInfo"], MIBInfo);
	        this.children = this.convertValues(source["children"], MIBTreeEnt);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
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
	    EnableNetflowd: boolean;
	    EnableSshd: boolean;
	    EnableSFlowd: boolean;
	    EnableTcpd: boolean;
	    IconSize: number;
	    MapSize: number;
	    ArpWatchRange: string;
	    ArpTimeout: number;
	
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
	        this.EnableNetflowd = source["EnableNetflowd"];
	        this.EnableSshd = source["EnableSshd"];
	        this.EnableSFlowd = source["EnableSFlowd"];
	        this.EnableTcpd = source["EnableTcpd"];
	        this.IconSize = source["IconSize"];
	        this.MapSize = source["MapSize"];
	        this.ArpWatchRange = source["ArpWatchRange"];
	        this.ArpTimeout = source["ArpTimeout"];
	    }
	}
	export class NetFlowEnt {
	    Time: number;
	    SrcAddr: string;
	    SrcPort: number;
	    SrcLoc: string;
	    SrcMAC: string;
	    DstAddr: string;
	    DstPort: number;
	    DstLoc: string;
	    DstMAC: string;
	    Bytes: number;
	    Packets: number;
	    TCPFlags: string;
	    Protocol: string;
	    ToS: number;
	    Start: number;
	    End: number;
	    Dur: number;
	
	    static createFrom(source: any = {}) {
	        return new NetFlowEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.SrcAddr = source["SrcAddr"];
	        this.SrcPort = source["SrcPort"];
	        this.SrcLoc = source["SrcLoc"];
	        this.SrcMAC = source["SrcMAC"];
	        this.DstAddr = source["DstAddr"];
	        this.DstPort = source["DstPort"];
	        this.DstLoc = source["DstLoc"];
	        this.DstMAC = source["DstMAC"];
	        this.Bytes = source["Bytes"];
	        this.Packets = source["Packets"];
	        this.TCPFlags = source["TCPFlags"];
	        this.Protocol = source["Protocol"];
	        this.ToS = source["ToS"];
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.Dur = source["Dur"];
	    }
	}
	export class PortEnt {
	    ID: string;
	    Name: string;
	    Polling: string;
	    Index: string;
	    X: number;
	    Y: number;
	    State: string;
	
	    static createFrom(source: any = {}) {
	        return new PortEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Polling = source["Polling"];
	        this.Index = source["Index"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.State = source["State"];
	    }
	}
	export class NetworkEnt {
	    ID: string;
	    Name: string;
	    Descr: string;
	    IP: string;
	    SnmpMode: string;
	    Community: string;
	    User: string;
	    Password: string;
	    URL: string;
	    ArpWatch: boolean;
	    Unmanaged: boolean;
	    HPorts: number;
	    X: number;
	    Y: number;
	    W: number;
	    H: number;
	    SystemID: string;
	    Error: string;
	    LLDP: boolean;
	    Ports: PortEnt[];
	
	    static createFrom(source: any = {}) {
	        return new NetworkEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Descr = source["Descr"];
	        this.IP = source["IP"];
	        this.SnmpMode = source["SnmpMode"];
	        this.Community = source["Community"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.URL = source["URL"];
	        this.ArpWatch = source["ArpWatch"];
	        this.Unmanaged = source["Unmanaged"];
	        this.HPorts = source["HPorts"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.W = source["W"];
	        this.H = source["H"];
	        this.SystemID = source["SystemID"];
	        this.Error = source["Error"];
	        this.LLDP = source["LLDP"];
	        this.Ports = this.convertValues(source["Ports"], PortEnt);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NodeEnt {
	    ID: string;
	    Name: string;
	    Descr: string;
	    Icon: string;
	    Image: string;
	    State: string;
	    X: number;
	    Y: number;
	    IP: string;
	    MAC: string;
	    Vendor: string;
	    SnmpMode: string;
	    Community: string;
	    User: string;
	    SSHUser: string;
	    Password: string;
	    GNMIPort: string;
	    GNMIEncoding: string;
	    GNMIUser: string;
	    GNMIPassword: string;
	    PublicKey: string;
	    URL: string;
	    AddrMode: string;
	    AutoAck: boolean;
	    Loc: string;
	
	    static createFrom(source: any = {}) {
	        return new NodeEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Descr = source["Descr"];
	        this.Icon = source["Icon"];
	        this.Image = source["Image"];
	        this.State = source["State"];
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.IP = source["IP"];
	        this.MAC = source["MAC"];
	        this.Vendor = source["Vendor"];
	        this.SnmpMode = source["SnmpMode"];
	        this.Community = source["Community"];
	        this.User = source["User"];
	        this.SSHUser = source["SSHUser"];
	        this.Password = source["Password"];
	        this.GNMIPort = source["GNMIPort"];
	        this.GNMIEncoding = source["GNMIEncoding"];
	        this.GNMIUser = source["GNMIUser"];
	        this.GNMIPassword = source["GNMIPassword"];
	        this.PublicKey = source["PublicKey"];
	        this.URL = source["URL"];
	        this.AddrMode = source["AddrMode"];
	        this.AutoAck = source["AutoAck"];
	        this.Loc = source["Loc"];
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
	    BeepHigh: string;
	    BeepLow: string;
	
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
	        this.BeepHigh = source["BeepHigh"];
	        this.BeepLow = source["BeepLow"];
	    }
	}
	export class PKIControlEnt {
	    AcmeBaseURL: string;
	    EnableAcme: boolean;
	    EnableHttp: boolean;
	    AcmeStatus: string;
	    HttpStatus: string;
	    CrlInterval: number;
	    CertTerm: number;
	
	    static createFrom(source: any = {}) {
	        return new PKIControlEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AcmeBaseURL = source["AcmeBaseURL"];
	        this.EnableAcme = source["EnableAcme"];
	        this.EnableHttp = source["EnableHttp"];
	        this.AcmeStatus = source["AcmeStatus"];
	        this.HttpStatus = source["HttpStatus"];
	        this.CrlInterval = source["CrlInterval"];
	        this.CertTerm = source["CertTerm"];
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
	    FailAction: string;
	    RepairAction: string;
	
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
	        this.FailAction = source["FailAction"];
	        this.RepairAction = source["RepairAction"];
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
	    ID: number;
	    Name: string;
	    Level: string;
	    Type: string;
	    Mode: string;
	    Params: string;
	    Filter: string;
	    Extractor: string;
	    Script: string;
	    Descr: string;
	    AutoParam: string;
	
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
	        this.AutoParam = source["AutoParam"];
	    }
	}
	
	export class SFlowCounterEnt {
	    Time: number;
	    Remote: string;
	    Type: string;
	    Data: string;
	
	    static createFrom(source: any = {}) {
	        return new SFlowCounterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.Remote = source["Remote"];
	        this.Type = source["Type"];
	        this.Data = source["Data"];
	    }
	}
	export class SFlowEnt {
	    Time: number;
	    SrcAddr: string;
	    SrcPort: number;
	    SrcLoc: string;
	    SrcMAC: string;
	    DstAddr: string;
	    DstPort: number;
	    DstLoc: string;
	    DstMAC: string;
	    Bytes: number;
	    TCPFlags: string;
	    Protocol: string;
	    Reason: number;
	
	    static createFrom(source: any = {}) {
	        return new SFlowEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.SrcAddr = source["SrcAddr"];
	        this.SrcPort = source["SrcPort"];
	        this.SrcLoc = source["SrcLoc"];
	        this.SrcMAC = source["SrcMAC"];
	        this.DstAddr = source["DstAddr"];
	        this.DstPort = source["DstPort"];
	        this.DstLoc = source["DstLoc"];
	        this.DstMAC = source["DstMAC"];
	        this.Bytes = source["Bytes"];
	        this.TCPFlags = source["TCPFlags"];
	        this.Protocol = source["Protocol"];
	        this.Reason = source["Reason"];
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
	    Wait: number;
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
	        this.Wait = source["Wait"];
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
	export class ArpLogEnt {
	    Time: number;
	    State: string;
	    IP: string;
	    Node: string;
	    NewMAC: string;
	    NewVendor: string;
	    OldMAC: string;
	    OldVendor: string;
	
	    static createFrom(source: any = {}) {
	        return new ArpLogEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = source["Time"];
	        this.State = source["State"];
	        this.IP = source["IP"];
	        this.Node = source["Node"];
	        this.NewMAC = source["NewMAC"];
	        this.NewVendor = source["NewVendor"];
	        this.OldMAC = source["OldMAC"];
	        this.OldVendor = source["OldVendor"];
	    }
	}
	export class CertEnt {
	    Status: string;
	    ID: string;
	    Subject: string;
	    Node: string;
	    Created: number;
	    Revoked: number;
	    Expire: number;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new CertEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Status = source["Status"];
	        this.ID = source["ID"];
	        this.Subject = source["Subject"];
	        this.Node = source["Node"];
	        this.Created = source["Created"];
	        this.Revoked = source["Revoked"];
	        this.Expire = source["Expire"];
	        this.Type = source["Type"];
	    }
	}
	export class EventLogFilterEnt {
	    NodeID: string;
	    Start: string;
	    End: string;
	    EventType: string;
	    NodeName: string;
	    Event: string;
	    Level: number;
	
	    static createFrom(source: any = {}) {
	        return new EventLogFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.NodeID = source["NodeID"];
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.EventType = source["EventType"];
	        this.NodeName = source["NodeName"];
	        this.Event = source["Event"];
	        this.Level = source["Level"];
	    }
	}
	export class ExportData {
	    Title: string;
	    Header: string[];
	    Data: any[][];
	    Image: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.Header = source["Header"];
	        this.Data = source["Data"];
	        this.Image = source["Image"];
	    }
	}
	export class ModelData {
	    name?: string;
	    organization?: string;
	    version?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.organization = source["organization"];
	        this.version = source["version"];
	    }
	}
	export class GNMICapEnt {
	    Version: string;
	    Encodings: string;
	    Models: ModelData[];
	
	    static createFrom(source: any = {}) {
	        return new GNMICapEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Version = source["Version"];
	        this.Encodings = source["Encodings"];
	        this.Models = this.convertValues(source["Models"], ModelData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GNMIGetEnt {
	    Path: string;
	    Value: string;
	    Index: string;
	
	    static createFrom(source: any = {}) {
	        return new GNMIGetEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Value = source["Value"];
	        this.Index = source["Index"];
	    }
	}
	export class IPAMRangeEnt {
	    Range: string;
	    Size: number;
	    Used: number;
	    Usage: number;
	    UsedIP: number[];
	
	    static createFrom(source: any = {}) {
	        return new IPAMRangeEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Range = source["Range"];
	        this.Size = source["Size"];
	        this.Used = source["Used"];
	        this.Usage = source["Usage"];
	        this.UsedIP = source["UsedIP"];
	    }
	}
	export class MibEnt {
	    Name: string;
	    Value: string;
	
	    static createFrom(source: any = {}) {
	        return new MibEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Value = source["Value"];
	    }
	}
	export class NetFlowFilterEnt {
	    Start: string;
	    End: string;
	    Single: boolean;
	    SrcAddr: string;
	    SrcPort: number;
	    SrcLoc: string;
	    SrcMAC: string;
	    DstAddr: string;
	    DstPort: number;
	    DstLoc: string;
	    DstMAC: string;
	    Protocol: string;
	    TCPFlags: string;
	
	    static createFrom(source: any = {}) {
	        return new NetFlowFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.Single = source["Single"];
	        this.SrcAddr = source["SrcAddr"];
	        this.SrcPort = source["SrcPort"];
	        this.SrcLoc = source["SrcLoc"];
	        this.SrcMAC = source["SrcMAC"];
	        this.DstAddr = source["DstAddr"];
	        this.DstPort = source["DstPort"];
	        this.DstLoc = source["DstLoc"];
	        this.DstMAC = source["DstMAC"];
	        this.Protocol = source["Protocol"];
	        this.TCPFlags = source["TCPFlags"];
	    }
	}
	export class PingReq {
	    IP: string;
	    Size: number;
	    TTL: number;
	
	    static createFrom(source: any = {}) {
	        return new PingReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IP = source["IP"];
	        this.Size = source["Size"];
	        this.TTL = source["TTL"];
	    }
	}
	export class PingRes {
	    Stat: number;
	    TimeStamp: number;
	    Time: number;
	    Size: number;
	    SendTTL: number;
	    RecvTTL: number;
	    RecvSrc: string;
	    Loc: string;
	
	    static createFrom(source: any = {}) {
	        return new PingRes(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Stat = source["Stat"];
	        this.TimeStamp = source["TimeStamp"];
	        this.Time = source["Time"];
	        this.Size = source["Size"];
	        this.SendTTL = source["SendTTL"];
	        this.RecvTTL = source["RecvTTL"];
	        this.RecvSrc = source["RecvSrc"];
	        this.Loc = source["Loc"];
	    }
	}
	export class SFlowCounterFilterEnt {
	    Start: string;
	    End: string;
	    Type: string;
	    Remote: string;
	
	    static createFrom(source: any = {}) {
	        return new SFlowCounterFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.Type = source["Type"];
	        this.Remote = source["Remote"];
	    }
	}
	export class SFlowFilterEnt {
	    Start: string;
	    End: string;
	    Single: boolean;
	    SrcAddr: string;
	    SrcPort: number;
	    SrcLoc: string;
	    SrcMAC: string;
	    DstAddr: string;
	    DstPort: number;
	    DstLoc: string;
	    DstMAC: string;
	    Protocol: string;
	    TCPFlags: string;
	    Reason: number;
	
	    static createFrom(source: any = {}) {
	        return new SFlowFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.Single = source["Single"];
	        this.SrcAddr = source["SrcAddr"];
	        this.SrcPort = source["SrcPort"];
	        this.SrcLoc = source["SrcLoc"];
	        this.SrcMAC = source["SrcMAC"];
	        this.DstAddr = source["DstAddr"];
	        this.DstPort = source["DstPort"];
	        this.DstLoc = source["DstLoc"];
	        this.DstMAC = source["DstMAC"];
	        this.Protocol = source["Protocol"];
	        this.TCPFlags = source["TCPFlags"];
	        this.Reason = source["Reason"];
	    }
	}
	export class Settings {
	    Kiosk: boolean;
	    Lock: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Kiosk = source["Kiosk"];
	        this.Lock = source["Lock"];
	    }
	}
	export class SyslogFilterEnt {
	    Start: string;
	    End: string;
	    Host: string;
	    Tag: string;
	    Message: string;
	    Severity: number;
	
	    static createFrom(source: any = {}) {
	        return new SyslogFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.Host = source["Host"];
	        this.Tag = source["Tag"];
	        this.Message = source["Message"];
	        this.Severity = source["Severity"];
	    }
	}
	export class TrapFilterEnt {
	    Start: string;
	    End: string;
	    From: string;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new TrapFilterEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Start = source["Start"];
	        this.End = source["End"];
	        this.From = source["From"];
	        this.Type = source["Type"];
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

export namespace pki {
	
	export class CSRReqEnt {
	    KeyType: string;
	    CommonName: string;
	    OrganizationalUnit: string;
	    Organization: string;
	    Locality: string;
	    Province: string;
	    Country: string;
	    Sans: string;
	
	    static createFrom(source: any = {}) {
	        return new CSRReqEnt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.KeyType = source["KeyType"];
	        this.CommonName = source["CommonName"];
	        this.OrganizationalUnit = source["OrganizationalUnit"];
	        this.Organization = source["Organization"];
	        this.Locality = source["Locality"];
	        this.Province = source["Province"];
	        this.Country = source["Country"];
	        this.Sans = source["Sans"];
	    }
	}

}

