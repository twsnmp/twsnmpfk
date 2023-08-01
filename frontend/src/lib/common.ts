import * as echarts from 'echarts'

export const stateList = [
  { text: '重度', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'high' },
  { text: '軽度', color: '#fb9a99', icon: 'mdi-alert-circle', value: 'low' },
  { text: '注意', color: '#dfdf22', icon: 'mdi-alert', value: 'warn' },
  { text: '正常', color: '#33a02c', icon: 'mdi-check-circle', value: 'normal' },
  { text: 'Up', color: '#33a02c', icon: 'mdi-check-circle', value: 'up' },
  { text: '復帰', color: '#1f78b4', icon: 'mdi-autorenew', value: 'repair' },
  { text: '情報', color: '#1f78b4', icon: 'mdi-information', value: 'info' },
  { text: '新規', color: '#1f78b4', icon: 'mdi-information', value: 'New' },
  { text: '変化', color: '#e31a1c', icon: 'mdi-autorenew', value: 'Change' },
  { text: 'エラー', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'error' },
  { text: 'Down', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'down' },
  { text: '停止', color: '#777', icon: 'mdi-stop', value: 'off' },
  { text: 'Debug', color: '#777', icon: 'mdi-bug', value: 'debug' },
]

export const stateMap = {}

stateList.forEach((e:any) => {
  stateMap[e.value] = e
})

export const getStateColor = (state:string) => {
  return stateMap[state] ? stateMap[state].color : 'gray'
}

export const getStateName = (state:string) => {
  return stateMap[state] ? stateMap[state].text : '不明'
}

export const getStateIconName = (state:string) => {
  return stateMap[state] ? stateMap[state].icon : 'mdi-comment-question-outline'
}

export const levelList = [
  { text: '重度', value: 'high' },
  { text: '軽度', value: 'low' },
  { text: '注意', value: 'warn' },
  { text: '情報', value: 'info' },
  { text: '停止', value: 'off' },
]

export const filterEventLevelList = [
  { text: '指定しない', value: '' },
  { text: '重度以上', value: 'high' },
  { text: '軽度以上', value: 'low' },
  { text: '注意以上', value: 'warn' },
]

export const filterEventTypeList = [
  { text: '指定しない', value: '' },
  { text: 'システム', value: 'system' },
  { text: 'ポーリング', value: 'polling' },
  { text: 'AI分析', value: 'ai' },
  { text: '稼働率', value: 'oprate' },
  { text: 'ARP監視', value: 'arpwatch' },
]

export const typeList = [
  { text: 'PING', value: 'ping' },
  { text: 'SNMP', value: 'snmp' },
  { text: 'TCP', value: 'tcp' },
  { text: 'HTTP', value: 'http' },
  { text: 'TLS', value: 'tls' },
  { text: 'DNS', value: 'dns' },
  { text: 'NTP', value: 'ntp' },
  { text: 'SYSLOG', value: 'syslog' },
  { text: 'SNMP TRAP', value: 'trap' },
  { text: 'ARP Log', value: 'arplog' },
  { text: 'Command', value: 'cmd' },
  { text: 'SSH', value: 'ssh' },
  { text: 'Report', value: 'report' },
  { text: 'TWSNMP', value: 'twsnmp' },
  { text: 'VMware', value: 'vmware' },
  { text: 'LXI', value: 'lxi' },
]

export const logModeList = [
  { text: '記録しない', value: 0 },
  { text: '常に記録', value: 1 },
  { text: '状態変化時のみ記録', value: 2 },
  { text: 'AI分析', value: 3 },
]

export const addrModeList = [
  { text: 'IP固定', value: '' },
  { text: 'MAC固定', value: 'mac' },
  { text: 'ホスト名固定', value: 'host' },
]

export const snmpModeList = [
  { name: 'SNMPv1', value: 'v1' },
  { name: 'SNMPv2c', value: 'v2c' },
  { name: 'SNMPv3認証', value: 'v3auth' },
  { name: 'SNMPv3認証暗号化(AES128)', value: 'v3authpriv' },
  { name: 'SNMPv3認証暗号化(SHA256/AES256)', value: 'v3authprivex' },
]

export const aiThList = [
  { text: '0.01%以下', value: 88 },
  { text: '0.1%以下', value: 81 },
  { text: '1%以下', value: 74 },
]

export const iconList = [
  {
    text: 'デスクトップ',
    icon: 'mdi-desktop-mac',
    value: 'desktop',
    code: 0xF01C4,
  },
  {
    text: '古いデスクトップ',
    icon: 'mdi-desktop-classic',
    value: 'desktop-classic',
    code: 0xF07C0,
  },
  { text: 'ラップトップ', icon: 'mdi-laptop', value: 'laptop' ,code: 0xF0322},
  { text: 'タブレット', icon: 'mdi-tablet-ipad', value: 'tablet' ,code:0xF04F8},
  { text: 'サーバー', icon: 'mdi-server', value: 'server' ,code: 0xF048B},
  { text: 'ネットワーク機器', icon: 'mdi-ip-network', value: 'hdd' ,code: 0xF0A60},
  { text: 'IP機器', icon: 'mdi-ip-network', value: 'ip' ,code: 0xF0A60},
  { text: 'ネットワーク', icon: 'mdi-lan', value: 'network' ,code: 0xF0317},
  { text: '無線LAN', icon: 'mdi-wifi', value: 'wifi' ,code: 0xF05A9},
  { text: 'クラウド', icon: 'mdi-cloud', value: 'cloud' ,code: 0xF015F },
  { text: 'プリンター', icon: 'mdi-printer', value: 'printer' ,code: 0xF042A},
  { text: 'モバイル', icon: 'mdi-cellphone', value: 'cellphone' ,code: 0xF011C},
  { text: 'ルーター1', icon: 'mdi-router', value: 'router' ,code: 0xF11E2},
  { text: 'WEB', icon: 'mdi-web', value: 'web' ,code: 0xF059F},
  { text: 'データベース', icon: 'mdi-database', value: 'db' ,code: 0xF01BC},
  { text: '無線AP', icon: 'mdi-router-wireless', value: 'mdi-router-wireless' ,code: 0xF0469},
  { text: 'ルーター2', icon: 'mdi-router-network', value: 'mdi-router-network' ,code: 0xF1087},
  { text: 'セキュリティー', icon: 'mdi-security', value: 'mdi-security' ,code: 0xF0483},
  { text: 'タワーPC', icon: 'mdi-desktop-tower', value: 'mdi-desktop-tower' ,code: 0xF01C5},
  { text: 'Windows', icon: 'mdi-microsoft-windows', value: 'mdi-microsoft-windows' ,code: 0xF05B3},
  { text: 'Linux', icon: 'mdi-linux', value: 'mdi-linux' ,code: 0xF033D},
  { text: 'Raspberry PI', icon: 'mdi-raspberry-pi', value: 'mdi-raspberry-pi' ,code: 0xF043F},
  { text: 'メールサーバー', icon: 'mdi-mailbox', value: 'mdi-mailbox' ,code: 0xF06EE},
  { text: 'NTPサーバー', icon: 'mdi-clock', value: 'mdi-clock' ,code: 0xF0954},
  { text: 'Android', icon: 'mdi-android', value: 'mdi-android' ,code: 0xF0032},
  { text: 'Azure', icon: 'mdi-microsoft-azure', value: 'mdi-microsoft-azure' ,code: 0xF0805},
  { text: 'Amazon', icon: 'mdi-amazon', value: 'mdi-amazon' ,code: 0xF002D},
  { text: 'Apple', icon: 'mdi-apple', value: 'mdi-apple' ,code: 0xF0035},
  { text: 'Google', icon: 'mdi-google', value: 'mdi-google' ,code: 0xF02AD},
  { text: 'CDプレーヤー', icon: 'mdi-disc-player', value: 'mdi-disc-player' ,code: 0xF0960},
  { text: 'TWSNMP連携マップ', icon: 'mdi-layers-search', value: 'mdi-layers-search' ,code: 0xF1206},
]

const iconMap = {}

iconList.forEach((e) => {
  iconMap[e.value] = e.icon
})

export const getIconName = (icon:string) => {
  return iconMap[icon] ? iconMap[icon] : 'mdi-comment-question-outline'
}

export const setIcon = (e:any) => {
  for( let i = 0; i < iconList.length; i++) {
    if(iconList[i].value === e.Icon) {
      // updaxte icon
      iconList[i].text = e.Text
      iconList[i].code = e.Code
      return
    }
  }
  // add icon
  iconList.push({
    text: e.Text,
    icon: e.Icon,
    value: e.Icon,
    code: e.Code,
  })
  iconMap[e.Icon] = e.Icon
}

// delete icon
export const delIcon = (icon) => {
  for( let i = 0; i < iconList.length; i++) {
    if(iconList[i].value === icon) {
      iconList.splice(i+1,1)
      delete(iconMap[icon])
      return
    }
  }
}

export const timeFormat = (date:any, format:string) => {
  if (!format) {
      format = '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
  }
  return echarts.time.format(date,format,false)
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

export const getScoreIconName = (s) => {
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

export const cmpIP = (a :string,b:string) => {
  if (!a.includes(".") || !b.includes(".") ){
    return a < b  ? -1 : a > b ? 1 : 0 
  }
  const pa = a.split('.').map(function(s) {
    return parseInt(s); 
  });
  const pb = b.split('.').map(function(s) {
    return parseInt(s); 
  });
  for(let i =0;i < pa.length;i++){
    if (i >= pb.length){
      return -1;
    }
    if (pa[i] === pb[i]){
      continue;
    }
    if (pa[i] < pb[i]){
      return -1;
    }
    return 1;
  }
  return 0;
}

export const getLogModeName = (m) => {
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

