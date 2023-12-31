// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {datastore} from '../models';
import {discover} from '../models';
import {backend} from '../models';

export function AutoGrok(arg1:string):Promise<string>;

export function Backup():Promise<boolean>;

export function CheckPolling(arg1:string):Promise<boolean>;

export function CopyDrawItem(arg1:string):Promise<boolean>;

export function CopyNode(arg1:string):Promise<boolean>;

export function DeleteAIResult(arg1:string):Promise<boolean>;

export function DeleteAllEventLogs():Promise<boolean>;

export function DeleteAllPollingLogs():Promise<boolean>;

export function DeleteAllSyslog():Promise<boolean>;

export function DeleteAllTraps():Promise<boolean>;

export function DeleteArpEnt(arg1:string):Promise<boolean>;

export function DeleteDrawItems(arg1:Array<string>):Promise<void>;

export function DeleteIcon(arg1:string):Promise<boolean>;

export function DeleteLine(arg1:string):Promise<boolean>;

export function DeleteNodes(arg1:Array<string>):Promise<void>;

export function DeletePollings(arg1:Array<string>):Promise<void>;

export function ExportAny(arg1:string,arg2:main.ExportData):Promise<string>;

export function ExportArpLogs(arg1:string):Promise<string>;

export function ExportArpTable(arg1:string):Promise<string>;

export function ExportEventLogs(arg1:string,arg2:string,arg3:string,arg4:string,arg5:number):Promise<string>;

export function ExportNodes(arg1:string):Promise<string>;

export function ExportPollings(arg1:string):Promise<string>;

export function ExportSyslogs(arg1:string,arg2:string,arg3:string,arg4:string,arg5:number):Promise<string>;

export function ExportTraps(arg1:string,arg2:string,arg3:string):Promise<string>;

export function GetAIConf():Promise<datastore.AIConfEnt>;

export function GetAIList():Promise<Array<main.AIList>>;

export function GetAIResult(arg1:string):Promise<datastore.AIResult>;

export function GetArpLogs():Promise<Array<main.ArpLogEnt>>;

export function GetArpTable():Promise<Array<datastore.ArpEnt>>;

export function GetAudio(arg1:string):Promise<string>;

export function GetAutoPollings(arg1:string,arg2:number):Promise<Array<datastore.PollingEnt>>;

export function GetBackImage():Promise<datastore.BackImageEnt>;

export function GetDefaultPolling(arg1:string):Promise<datastore.PollingEnt>;

export function GetDiscoverAddressRange():Promise<Array<string>>;

export function GetDiscoverConf():Promise<datastore.DiscoverConfEnt>;

export function GetDiscoverStats():Promise<discover.DiscoverStat>;

export function GetDrawItem(arg1:string):Promise<datastore.DrawItemEnt>;

export function GetDrawItems():Promise<{[key: string]: datastore.DrawItemEnt}>;

export function GetEventLogs(arg1:string,arg2:string,arg3:string,arg4:string,arg5:number):Promise<Array<datastore.EventLogEnt>>;

export function GetHostResource(arg1:string):Promise<backend.HostResourceEnt>;

export function GetIcons():Promise<Array<datastore.IconEnt>>;

export function GetImage(arg1:string):Promise<string>;

export function GetLang():Promise<string>;

export function GetLine(arg1:string,arg2:string):Promise<datastore.LineEnt>;

export function GetLines():Promise<Array<datastore.LineEnt>>;

export function GetLocConf():Promise<datastore.LocConfEnt>;

export function GetMIBModules():Promise<Array<datastore.MIBModuleEnt>>;

export function GetMIBTree():Promise<Array<datastore.MIBTreeEnt>>;

export function GetMapConf():Promise<datastore.MapConfEnt>;

export function GetMapEventLogs():Promise<Array<datastore.EventLogEnt>>;

export function GetMapName():Promise<string>;

export function GetMonitorDatas():Promise<Array<backend.MonitorDataEnt>>;

export function GetNode(arg1:string):Promise<datastore.NodeEnt>;

export function GetNodes():Promise<{[key: string]: datastore.NodeEnt}>;

export function GetNotifyConf():Promise<datastore.NotifyConfEnt>;

export function GetPolling(arg1:string):Promise<datastore.PollingEnt>;

export function GetPollingLogs(arg1:string):Promise<Array<datastore.PollingLogEnt>>;

export function GetPollingTemplate(arg1:number):Promise<datastore.PollingTemplateEnt>;

export function GetPollingTemplates():Promise<Array<datastore.PollingTemplateEnt>>;

export function GetPollings(arg1:string):Promise<Array<datastore.PollingEnt>>;

export function GetSettings():Promise<main.Settings>;

export function GetSyslogs(arg1:number,arg2:string,arg3:string,arg4:string):Promise<Array<datastore.SyslogEnt>>;

export function GetTraps(arg1:string,arg2:string):Promise<Array<datastore.TrapEnt>>;

export function GetVPanelPorts(arg1:string):Promise<Array<backend.VPanelPortEnt>>;

export function GetVPanelPowerInfo(arg1:string):Promise<boolean>;

export function GetVersion():Promise<string>;

export function HasDatastore():Promise<boolean>;

export function IsDark():Promise<boolean>;

export function IsLatest():Promise<boolean>;

export function Ping(arg1:main.PingReq):Promise<main.PingRes>;

export function ResetArpTable():Promise<boolean>;

export function SelectAudioFile(arg1:string):Promise<string>;

export function SelectDatastore():Promise<boolean>;

export function SelectFile(arg1:string,arg2:boolean):Promise<string>;

export function SetDark(arg1:boolean):Promise<void>;

export function SnmpWalk(arg1:string,arg2:string,arg3:boolean):Promise<Array<main.MibEnt>>;

export function StartDiscover(arg1:datastore.DiscoverConfEnt):Promise<boolean>;

export function StopDiscover():Promise<void>;

export function TestNotifyConf(arg1:datastore.NotifyConfEnt):Promise<boolean>;

export function UpdateAIConf(arg1:datastore.AIConfEnt):Promise<boolean>;

export function UpdateDrawItem(arg1:datastore.DrawItemEnt):Promise<boolean>;

export function UpdateDrawItemPos(arg1:Array<main.UpdatePosEnt>):Promise<void>;

export function UpdateIcon(arg1:datastore.IconEnt):Promise<boolean>;

export function UpdateLine(arg1:datastore.LineEnt):Promise<boolean>;

export function UpdateLocConf(arg1:datastore.LocConfEnt):Promise<boolean>;

export function UpdateMapConf(arg1:datastore.MapConfEnt):Promise<boolean>;

export function UpdateNode(arg1:datastore.NodeEnt):Promise<boolean>;

export function UpdateNodeLoc(arg1:string,arg2:string):Promise<void>;

export function UpdateNodePos(arg1:Array<main.UpdatePosEnt>):Promise<void>;

export function UpdateNotifyConf(arg1:datastore.NotifyConfEnt):Promise<boolean>;

export function UpdatePolling(arg1:datastore.PollingEnt):Promise<boolean>;

export function WakeOnLan(arg1:string):Promise<boolean>;
