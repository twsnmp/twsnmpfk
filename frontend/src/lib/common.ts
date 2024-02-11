import * as echarts from 'echarts';
import numeral from 'numeral';
import ja from "datatables.net-plugins/i18n/ja.json";
import { _,unwrapFunctionStore } from 'svelte-i18n';
import {lang} from '../i18n/i18n';
const $_ = unwrapFunctionStore(_);


export const stateList = [
  { text: $_("Ts.High"), color: '#e31a1c', icon: "mdi-alert-circle", value: 'high' },
  { text: $_("Ts.Low"), color: '#fb9a99', icon: "mdi-alert-circle", value: 'low' },
  { text: $_("Ts.Warn"), color: '#dfdf22', icon: "mdi-alert", value: 'warn' },
  { text: $_("Ts.Normal"), color: '#33a02c', icon: "mdi-check-circle", value: 'normal' },
  { text: $_("Ts.Repair"), color: '#1f78b4', icon: "mdi-autorenew", value: 'repair' },
  { text: $_("Ts.Info"), color: '#1f78b4', icon: "mdi-information", value: 'info' },
  { text: $_("Ts.New"), color: '#1f78b4', icon: "mdi-information", value: 'New' },
  { text: $_("Ts.Change"), color: '#e31a1c', icon: "mdi-autorenew", value: 'Change' },
  { text: $_("Ts.Error"), color: '#e31a1c', icon: "mdi-alert-circle", value: 'error' },
  { text: $_("Ts.Stop"), color: '#777', icon: "mdi-stop", value: 'off' },
  { text: 'Down', color: '#e31a1c', icon: "mdi-alert-circle", value: 'down' },
  { text: 'Up', color: '#33a02c', icon:  "mdi-check-circle", value: 'up' },
  { text: 'Debug', color: '#777', icon: "mdi-bug", value: 'debug' },
]

export const stateMap:any = {}

stateList.forEach((e:any) => {
  stateMap[e.value] = e
})

export const getStateColor = (state:string) :string => {
  return stateMap[state] ? stateMap[state].color : 'gray'
}

export const getStateName = (state:string) : string => {
  return stateMap[state] ? stateMap[state].text : $_("Ts.Unknown")
}

export const getStateIcon = (state:string) => {
  return stateMap[state] ? stateMap[state].icon :  "mdi-comment-question-outline";
}

export const levelList = [
  { name: $_("Ts.High"), value: 'high' },
  { name: $_("Ts.Low"), value: 'low' },
  { name: $_("Ts.Warn"), value: 'warn' },
  { name: $_("Ts.Info"), value: 'info' },
  { name: $_("Ts.Stop"), value: 'off' },
]

export const filterEventLevelList = [
  { name: $_("Ts.None"), value: '' },
  { name: $_("Ts.OverHigh"), value: 'high' },
  { name: $_("Ts.OverLow"), value: 'low' },
  { name: $_("Ts.OverWarn"), value: 'warn' },
]

export const filterEventTypeList = [
  { name: $_("Ts.None"), value: '' },
  { name: $_("Ts.System"), value: 'system' },
  { name: $_("Ts.Porlling"), value: 'polling' },
  { name: $_("Ts.AI"), value: 'ai' },
  { name: $_("Ts.Oprate"), value: 'oprate' },
  { name: $_("Ts.ArpWatch"), value: 'arpwatch' },
]

export const typeList = [
  { name: 'PING', value: 'ping' },
  { name: 'SNMP', value: 'snmp' },
  { name: 'TCP', value: 'tcp' },
  { name: 'HTTP', value: 'http' },
  { name: 'TLS', value: 'tls' },
  { name: 'DNS', value: 'dns' },
  { name: 'NTP', value: 'ntp' },
  { name: 'SYSLOG', value: 'syslog' },
  { name: 'SNMP TRAP', value: 'trap' },
  { name: 'ARP Log', value: 'arplog' },
  { name: 'Command', value: 'cmd' },
  { name: 'SSH', value: 'ssh' },
  { name: 'Report', value: 'report' },
  { name: 'TWSNMP', value: 'twsnmp' },
  { name: 'LXI', value: 'lxi' },
]

export const logModeList = [
  { name: $_("Ts.NoLog"), value: 0 },
  { name: $_("Ts.LogAll"), value: 1 },
  { name: $_("Ts.LogChnage"), value: 2 },
  { name: $_("Ts.AI"), value: 3 },
]

export const addrModeList = [
  { name: $_("Ts.FixedIP"), value: 'ip' },
  { name: $_("Ts.FixedMAC"), value: 'mac' },
  { name: $_("Ts.FixeHost"), value: 'host' },
]

export const snmpModeList = [
  { name: 'SNMPv1', value: 'v1' },
  { name: 'SNMPv2c', value: 'v2c' },
  { name: $_("Ts.SNMPv3Auth"), value: 'v3auth' },
  { name: $_("Ts.SNMPv3AuthPriv"), value: 'v3authpriv' },
  { name: $_("Ts.SNMPv3AuthPrivEx"), value: 'v3authprivex' },
]


export const iconList = [
  {
    name: $_("Ts.Desktop"),
    icon: 'mdi-desktop-mac',
    value: 'desktop',
    code: 0xF01C4,
  },
  {
    name: $_("Ts.DesktopClassic"),
    icon: 'mdi-desktop-classic',
    value: 'desktop-classic',
    code: 0xF07C0,
  },
  { name: $_("Ts.Laptop"), icon: 'mdi-laptop', value: 'laptop' ,code: 0xF0322},
  { name: $_("Ts.Tablet"), icon: 'mdi-tablet-ipad', value: 'tablet' ,code:0xF04F8},
  { name: $_("Ts.Server"), icon: 'mdi-server', value: 'server' ,code: 0xF048B},
  { name: $_("Ts.NetworkDev"), icon: 'mdi-ip-network', value: 'hdd' ,code: 0xF0A60},
  { name: $_("Ts.IPDev"), icon: 'mdi-ip-network', value: 'ip' ,code: 0xF0A60},
  { name: $_("Ts.Network"), icon: 'mdi-lan', value: 'network' ,code: 0xF0317},
  { name: $_("Ts.Wifi"), icon: 'mdi-wifi', value: 'wifi' ,code: 0xF05A9},
  { name: $_("Ts.Cloud"), icon: 'mdi-cloud', value: 'cloud' ,code: 0xF015F },
  { name: $_("Ts.Printer"), icon: 'mdi-printer', value: 'printer' ,code: 0xF042A},
  { name: $_("Ts.Mobile"), icon: 'mdi-cellphone', value: 'cellphone' ,code: 0xF011C},
  { name: $_("Ts.Router1"), icon: 'mdi-router', value: 'router' ,code: 0xF11E2},
  { name: 'WEB', icon: 'mdi-web', value: 'web' ,code: 0xF059F},
  { name: $_("Ts.Database"), icon: 'mdi-database', value: 'db' ,code: 0xF01BC},
  { name: $_("Ts.WifiAP"), icon: 'mdi-router-wireless', value: 'mdi-router-wireless' ,code: 0xF0469},
  { name: $_("Ts.Router2"), icon: 'mdi-router-network', value: 'mdi-router-network' ,code: 0xF1087},
  { name: $_("Ts.Security"), icon: 'mdi-security', value: 'mdi-security' ,code: 0xF0483},
  { name: $_("Ts.TowerPC"), icon: 'mdi-desktop-tower', value: 'mdi-desktop-tower' ,code: 0xF01C5},
  { name: 'Windows', icon: 'mdi-microsoft-windows', value: 'mdi-microsoft-windows' ,code: 0xF05B3},
  { name: 'Linux', icon: 'mdi-linux', value: 'mdi-linux' ,code: 0xF033D},
  { name: 'Raspberry PI', icon: 'mdi-raspberry-pi', value: 'mdi-raspberry-pi' ,code: 0xF043F},
  { name: $_("Ts.MailServer"), icon: 'mdi-mailbox', value: 'mdi-mailbox' ,code: 0xF06EE},
  { name: $_("Ts.NTPServer"), icon: 'mdi-clock', value: 'mdi-clock' ,code: 0xF0954},
  { name: 'Android', icon: 'mdi-android', value: 'mdi-android' ,code: 0xF0032},
  { name: 'Azure', icon: 'mdi-microsoft-azure', value: 'mdi-microsoft-azure' ,code: 0xF0805},
  { name: 'Amazon', icon: 'mdi-amazon', value: 'mdi-amazon' ,code: 0xF002D},
  { name: 'Apple', icon: 'mdi-apple', value: 'mdi-apple' ,code: 0xF0035},
  { name: 'Google', icon: 'mdi-google', value: 'mdi-google' ,code: 0xF02AD},
  { name: 'CD Player', icon: 'mdi-disc-player', value: 'mdi-disc-player' ,code: 0xF0960},
  { name: 'TWSNMP', icon: 'mdi-layers-search', value: 'mdi-layers-search' ,code: 0xF1206},
]

const iconCodeMap = new Map();

const iconMap = new Map();

iconList.forEach((e) => {
  iconMap.set(e.value, e.icon);
  iconCodeMap.set(e.value,String.fromCodePoint(e.code));
})

export const getIcon = (icon:string) : string => {
  return iconMap.get(icon) || 'mdi-comment-question-outline';
}

export const getIconCode = (icon:string) : number => {
  return iconCodeMap.get(icon) || String.fromCodePoint(0xF0A39);
}

export const setIconToList = (e:any) => {
  for( let i = 0; i < iconList.length; i++) {
    if(iconList[i].value === e.Icon) {
      // update icon
      iconList[i].name = e.Name;
      iconList[i].code = e.Code;
      iconCodeMap.set(e.Icon,String.fromCodePoint(e.Code));
      return;
    }
  }
  // add icon
  iconList.push({
    name: e.Name,
    icon: e.Icon,
    value: e.Icon,
    code: e.Code,
  })
  iconMap.set(e.Icon,e.Icon);
  iconCodeMap.set(e.Icon,String.fromCodePoint(e.Code));
}

// delete icon
export const deleteIconFromList = (icon:string) => {
  for( let i = 0; i < iconList.length; i++) {
    if(iconList[i].value === icon) {
      iconList.splice(i,1)
      iconCodeMap.delete(icon);
      iconMap.delete(icon);
      return
    }
  }
}

export const formatTime = (date:any, format:string) => {
  if (!format) {
      format = '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
  }
  return echarts.time.format(date,format,false)
}

export const renderTime = (t:number,type:string) => {
  if (t < 1) {
    return "";
  }
  const d = new Date(t /(1000*1000));
  return  formatTime(d,"");
}

export const getScoreColor = (s:number) => {
  if (s > 66) {
    return getStateColor('repair')
  } else if (s >= 50) {
    return getStateColor('info')
  } else if (s > 42) {
    return getStateColor('warn')
  } else if (s > 33) {
    return getStateColor('low')
  }
  return getStateColor('high')
}

export const getScoreIcon = (s:number) => {
  if (s > 66) {
    return 'mdi-emoticon-excited-outline'
  } else if (s >= 50) {
    return 'mdi-emoticon-outline'
  } else if (s > 42) {
    return 'mdi-emoticon-sad-outline'
  } else if (s > 33) {
    return 'mdi-emoticon-sick-outline'
  }
  return 'mdi-emoticon-dead-outline'
}


export const renderState = (state:string,type:string) => {
  if(type=="sort") {
    return levelNum(state);
  }
  return `<span class="mdi ` +
      getStateIcon(state) +
      ` text-sm" style="color:` +
      getStateColor(state) +
      `;"></span><span class="ml-2">` +
      getStateName(state) +
      `</span>`;
};

export const renderNodeState = (state:string,type:string,n:any) => {
  if(type=="sort") {
    return levelNum(state);
  }
  const icon = n.Icon ? getIcon(n.Icon) : getStateIcon(state);
  return `<span class="mdi ` +
      icon +
      ` text-sm" style="color:` +
      getStateColor(state) +
      `;"></span><span class="ml-2">` +
      getStateName(state) +
      `</span>`;
};

export const renderIP = (ip:string,type:string) => {
  if (type=="sort") {
    return ip.split(".").reduce((int, v) => (Number(int) * 256  +Number(v)) + "");
  }
  return ip;
}

export const  levelNum = (s :string) :number => {
	switch (s) {
	case "high":
		return 0;
	case "low":
		return 1;
	case "warn":
		return 2;
	case "normal":
		return 4
	case "repair":
		return 3
	}
	return 5
}


export const getLogModeName = (m:any) => {
  switch(m){
    case 0:
      return 'off'
    case 1:
      return 'all'
    case 2:
      return 'diff'
    case 3:
      return 'ai'
  }
  return ''
} 


export const getTableLang = () => {
  if (lang != "ja") {
    return undefined;
  }
  const r : any = ja;
  r.select = {
    cells: {
      "0": "",
      "1": "1 セル選択",
      _: "%d セル選択",
    },
    columns: {
      "0": "",
      "1": "1 カラム選択",
      _: "%d カラム選択",
    },
    rows: {
      "0": "",
      "1": " 1 行選択",
      _: " %d 行選択",
    },
  };
  return r;
}

export const renderSpeed = (n:number,type:string) => {
  if (type == "sort") {
    return n;
  }
  return numeral(n).format('0b') + 'PS';
}

export const renderCount = (n:number,type:string) => {
  if (type == "sort") {
    return n;
  }
  return numeral(n).format('0,0');
}

export const renderBytes = (n:number,type:string) => {
  if (type == "sort") {
    return n;
  }
  return numeral(n).format('0.000b');
}

export const renderHrSystemName = (k:string) => {
  switch (k) {
    case "hrSystemUptime":
      return $_("Ts.hrSystemUptime");
    case "hrSystemDate":
      return $_("Ts.hrSystemDate");
    case "hrSystemInitialLoadDevice":
      return $_("Ts.hrSystemInitialLoadDevice");
    case "hrSystemInitialLoadParameters":
      return $_("Ts.hrSystemInitialLoadParameters");
    case "hrSystemNumUsers":
      return $_("Ts.hrSystemNumUsers");
    case "hrSystemProcesses":
      return $_("Ts.hrSystemProcesses");
    case "hrSystemMaxProcesses":
      return $_("Ts.hrSystemMaxProcesses");
    case "hrMemorySize":
      return $_("Ts.hrMemorySize");
    case "hrProcessorLoad":
      return $_("Ts.hrProcessorLoad");
    case "hrProcessorCount":
      return $_("Ts.hrProcessorCount");
  }
  return k
}