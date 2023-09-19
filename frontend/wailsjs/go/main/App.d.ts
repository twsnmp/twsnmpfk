// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {datastore} from '../models';
import {discover} from '../models';
import {backend} from '../models';

export function CheckPolling(arg1:string):Promise<boolean>;

export function CopyDrawItem(arg1:string):Promise<boolean>;

export function CopyNode(arg1:string):Promise<boolean>;

export function DeeleteAIResult(arg1:string):Promise<boolean>;

export function DeleteDrawItems(arg1:Array<string>):Promise<void>;

export function DeleteLine(arg1:string):Promise<boolean>;

export function DeleteNodes(arg1:Array<string>):Promise<void>;

export function DeletePollings(arg1:Array<string>):Promise<void>;

export function ExportAny(arg1:string,arg2:main.ExportData):Promise<string>;

export function ExportEventLogs(arg1:string):Promise<string>;

export function ExportNodes(arg1:string):Promise<string>;

export function ExportPollings(arg1:string):Promise<string>;

export function ExportSyslogs(arg1:string):Promise<string>;

export function ExportTraps(arg1:string):Promise<string>;

export function GetAIConf():Promise<datastore.AIConfEnt>;

export function GetAIList():Promise<Array<main.AIList>>;

export function GetAIResult(arg1:string):Promise<datastore.AIResult>;

export function GetAlertEventLogs():Promise<Array<datastore.EventLogEnt>>;

export function GetArpLogs():Promise<Array<datastore.LogEnt>>;

export function GetAutoPollings(arg1:string,arg2:number):Promise<Array<datastore.PollingEnt>>;

export function GetBackImage():Promise<datastore.BackImageEnt>;

export function GetDiscoverConf():Promise<datastore.DiscoverConfEnt>;

export function GetDiscoverStats():Promise<discover.DiscoverStat>;

export function GetDrawItem(arg1:string):Promise<datastore.DrawItemEnt>;

export function GetDrawItems():Promise<{[key: string]: datastore.DrawItemEnt}>;

export function GetEventLogs(arg1:string):Promise<Array<datastore.EventLogEnt>>;

export function GetHostResource(arg1:string):Promise<backend.HostResourceEnt>;

export function GetImage(arg1:string):Promise<string>;

export function GetLine(arg1:string,arg2:string):Promise<datastore.LineEnt>;

export function GetLines():Promise<Array<datastore.LineEnt>>;

export function GetMIBTree():Promise<Array<datastore.MIBTreeEnt>>;

export function GetMapConf():Promise<datastore.MapConfEnt>;

export function GetMapName():Promise<string>;

export function GetNode(arg1:string):Promise<datastore.NodeEnt>;

export function GetNodes():Promise<{[key: string]: datastore.NodeEnt}>;

export function GetNotifyConf():Promise<datastore.NotifyConfEnt>;

export function GetPolling(arg1:string):Promise<datastore.PollingEnt>;

export function GetPollingLogs(arg1:string):Promise<Array<datastore.PollingLogEnt>>;

export function GetPollingTemplate(arg1:number):Promise<datastore.PollingTemplateEnt>;

export function GetPollingTemplates():Promise<Array<datastore.PollingTemplateEnt>>;

export function GetPollings(arg1:string):Promise<Array<datastore.PollingEnt>>;

export function GetSettings():Promise<main.Settings>;

export function GetSyslogs():Promise<Array<datastore.SyslogEnt>>;

export function GetTraps():Promise<Array<datastore.TrapEnt>>;

export function GetVPanelPorts(arg1:string):Promise<Array<backend.VPanelPortEnt>>;

export function GetVPanelPowerInfo(arg1:string):Promise<boolean>;

export function GetVersion():Promise<string>;

export function Ping(arg1:main.PingReq):Promise<main.PingRes>;

export function SelectFile(arg1:string,arg2:boolean):Promise<string>;

export function SnmpWalk(arg1:string,arg2:string,arg3:boolean):Promise<Array<main.MibEnt>>;

export function StartDiscover(arg1:datastore.DiscoverConfEnt):Promise<boolean>;

export function StopDiscover():Promise<void>;

export function TestNotifyConf(arg1:datastore.NotifyConfEnt):Promise<boolean>;

export function UpdateAIConf(arg1:datastore.AIConfEnt):Promise<boolean>;

export function UpdateDrawItem(arg1:datastore.DrawItemEnt):Promise<boolean>;

export function UpdateDrawItemPos(arg1:Array<main.UpdatePosEnt>):Promise<void>;

export function UpdateLine(arg1:datastore.LineEnt):Promise<boolean>;

export function UpdateMapConf(arg1:datastore.MapConfEnt):Promise<boolean>;

export function UpdateNode(arg1:datastore.NodeEnt):Promise<boolean>;

export function UpdateNodePos(arg1:Array<main.UpdatePosEnt>):Promise<void>;

export function UpdateNotifyConf(arg1:datastore.NotifyConfEnt):Promise<boolean>;

export function UpdatePolling(arg1:datastore.PollingEnt):Promise<boolean>;

export function WakeOnLan(arg1:string):Promise<boolean>;
