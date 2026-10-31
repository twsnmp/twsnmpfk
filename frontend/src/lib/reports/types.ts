// Report entity types matching Go logreport models

export interface RSSIEnt {
  Time: number;
  Value: number;
}

export interface WifiAPEnt {
  ID?: string;
  BSSID: string;
  SSID: string;
  Channel: string;
  RSSI: RSSIEnt[];
  Vendor: string;
  Host: string;
  FirstTime: number;
  LastTime: number;
  Count: number;
  Change: number;
  Info: string;
}

export interface BlueDeviceEnt {
  ID?: string;
  Address: string;
  AddressType: string;
  Name: string;
  RSSI: RSSIEnt[];
  Vendor: string;
  Host: string;
  FirstTime: number;
  LastTime: number;
  Count: number;
  Change: number;
  Info: string;
}

export interface EnvDataEnt {
  Time: number;
  Temp: number;
  Humidity: number;
  Illuminance: number;
  BarometricPressure?: number;
  Pressure?: number;
  Sound?: number;
  Noise?: number;
  ETVOC: number;
  ECo2: number;
  Battery: number;
  RSSI: number;
}

export interface EnvMonitorEnt {
  ID?: string;
  Address: string;
  Host: string;
  Name: string;
  EnvData?: EnvDataEnt[];
  Data?: EnvDataEnt[];
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface PowerMonitorDataEnt {
  Time: number;
  Load?: number;
  Power?: number;
  Switch: boolean;
  Over?: boolean;
  Overload?: boolean;
  RSSI: number;
}

export interface PowerMonitorEnt {
  ID?: string;
  Address: string;
  Host: string;
  Name: string;
  Data: PowerMonitorDataEnt[];
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface MotionSensorDataEnt {
  Time: number;
  Event?: string;
  Moving: boolean;
  Light: boolean;
  LastMove?: number;
  LastMoveDiff?: number;
  LastMoveTime?: number;
  Battery: number;
  RSSI: number;
}

export interface MotionSensorEnt {
  ID?: string;
  Address: string;
  Host: string;
  Name: string;
  Data: MotionSensorDataEnt[];
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface EtherTypeEnt {
  ID?: string;
  Host: string;
  Type: string;
  Name: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface DNSQEnt {
  ID?: string;
  Host: string;
  Name: string;
  Type: string;
  Server: string;
  Client?: string;
  LastClient?: string;
  ClientMAC?: string;
  LastMAC?: string;
  Count: number;
  Change: number;
  FirstTime: number;
  LastTime: number;
}

export interface RADIUSFlowEnt {
  ID: string;
  Client: string;
  ClientName?: string;
  ClientNodeID?: string;
  Server: string;
  ServerName?: string;
  ServerNodeID?: string;
  Count?: number;
  Request: number;
  Accept: number;
  Reject: number;
  Challenge: number;
  FirstTime: number;
  LastTime: number;
  Score: number;
  Penalty: number;
}

export interface TLSFlowEnt {
  ID: string;
  Client: string;
  ClientName?: string;
  ClientNodeID?: string;
  ClientLoc?: string;
  Server: string;
  ServerName?: string;
  ServerNodeID?: string;
  ServerLoc?: string;
  Service: string;
  Version: string;
  Cipher: string;
  ServerCountry?: string;
  ClientCountry?: string;
  Count: number;
  Score?: number;
  Penalty?: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinEventIDEnt {
  ID: string;
  Computer: string;
  Provider: string;
  EventID: number;
  Channel: string;
  Level: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinLogonEnt {
  ID: string;
  Computer: string;
  Target: string;
  SrcIP: string;
  LogonType: number;
  LogonCount: number;
  FailedCount: number;
  LogoffCount: number;
  FirstTime: number;
  LastTime: number;
  Score: number;
}

export interface WinAccountEnt {
  ID: string;
  Computer: string;
  Subject: string;
  Target: string;
  Type: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinKerberosEnt {
  ID: string;
  Computer: string;
  User: string;
  Service: string;
  TicketType: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinPrivilegeEnt {
  ID: string;
  Computer: string;
  Subject: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinProcessEnt {
  ID: string;
  Computer: string;
  Process: string;
  Subject: string;
  Parent: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface WinTaskEnt {
  ID: string;
  Computer: string;
  Task: string;
  Account: string;
  Count: number;
  FirstTime: number;
  LastTime: number;
}

export interface TwAgentResourcePoint {
  Time: number;
  CPU: number;
  Mem: number;
  Disk: number;
  Sent: number;
  Recv: number;
}

export interface TwAgentDiskUsage {
  Path: string;
  UsedPercent: number;
  TotalBytes: number;
}

export interface TwAgentNetIF {
  Name: string;
  IP: string;
  MAC: string;
  BytesSent: number;
  BytesRecv: number;
}

export interface TwAgentInfoEnt {
  ID: string;
  AgentID?: string;
  Host?: string;
  Hostname?: string;
  OS?: string;
  Arch?: string;
  Version?: string;
  IP?: string;
  MAC?: string;
  State?: string;
  Status?: string;
  CPUPercent?: number;
  MemPercent?: number;
  DiskPercent?: number;
  CPUCores?: number;
  Disks?: TwAgentDiskUsage[];
  Interfaces?: TwAgentNetIF[];
  History?: TwAgentResourcePoint[];
  FirstTime?: number;
  LastTime?: number;
  Score?: number;
  AlertCount?: number;
  ThreatCount?: number;
  ThreatsCount?: number;
  ComplianceCount?: number;
  FIMCount?: number;
  FimCount?: number;
  SCACount?: number;
  Count?: number;
}

export interface TwAgentAlertEnt {
  ID: string;
  AgentID?: string;
  Host: string;
  IP?: string;
  Protocol?: string;
  Protocols?: string[];
  Category?: string;
  RuleID?: string;
  RuleTitle?: string;
  Level?: string;
  AiVerdict?: string;
  Compliance?: string;
  ComplianceViolated?: boolean;
  ComplianceTypes?: string[];
  IsThreat?: boolean;
  AIJudged?: boolean;
  Confidence?: number;
  SourceLog?: string;
  RawLog?: string;
  SummaryJa?: string;
  ActionJa?: string;
  AiSummary?: string;
  AiAction?: string;
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface TwAgentFIMEnt {
  ID: string;
  AgentID?: string;
  Host: string;
  IP?: string;
  Protocol?: string;
  Protocols?: string[];
  Path: string;
  Op: string;
  Hash?: string;
  Size?: number;
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface TwAgentSCAEnt {
  ID: string;
  AgentID?: string;
  Host: string;
  IP?: string;
  Protocol?: string;
  Protocols?: string[];
  RuleID?: string;
  Rule?: string;
  Title?: string;
  Description?: string;
  FilePath?: string;
  Path?: string;
  Severity?: string;
  Status?: string;
  Proof?: string;
  Remediation?: string;
  AIJudged?: boolean;
  IsThreat?: boolean;
  ComplianceViolated?: boolean;
  SummaryJa?: string;
  ActionJa?: string;
  FirstTime: number;
  LastTime: number;
  Count: number;
}

export interface TwAgentTCPEnt {
  ID: string;
  AgentID?: string;
  Host: string;
  IP?: string;
  RemoteIP: string;
  RemotePort?: number;
  RemoteLoc?: string;
  Location?: string;
  Process: string;
  Services?: string[];
  ConnCount?: number;
  Count: number;
  FirstTime: number;
  LastTime: number;
}
