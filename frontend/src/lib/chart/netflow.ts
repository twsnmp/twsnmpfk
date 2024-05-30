import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import 'echarts-gl';
import { isMACFormat, isPrivateIP, isV4Format } from './utils.js'
import { doFFT } from './fft.js'

import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart :any;


const makeNetFlowHistogram = (div:string) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const option = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      formatter(params:any) {
        const p = params[0]
        return p.value[0] + ':' + p.value[1]
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      scale: true,
      min: 0,
    },
    yAxis: {
      name: $_("Ts.Count"),
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
  return chart
}

export const showNetFlowHistogram = (div:string, logs:any,type:string) => {
  makeNetFlowHistogram(div)
  if (type === '') {
    type = 'size'
  }
  const data :any = []
  logs.forEach((l:any) => {
    if (type === 'size') {
      if (l.Packets > 0) {
        data.push(l.Bytes / l.Packets)
      }
    } else if (type === 'dur') {
      if (l.Dur >= 0.0) {
        data.push(l.Dur)
      }
    } else if (type === 'speed') {
      if (l.Dur > 0) {
        data.push(l.Bytes / l.Dur)
      }
    }
  })
  const bins = ecStat.histogram(data,"squareRoot")
  chart.setOption({
    xAxis: {
      name: type,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize();
  return chart;
}

const makeNetFlowTraffic = (div:string, type:string) => {
  let yAxis = ''
  switch (type) {
    case 'bytes':
      yAxis = $_("Ts.Bytes")
      break
    case 'packets':
      yAxis = $_("Ts.Packets")
      break
    case 'bps':
      yAxis = $_("Ts.BPS")
      break
    case 'pps':
      yAxis = $_("Ts.PPS")
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const option = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: $_("Ts.DateTime"),
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value:any) => {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: yAxis,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        type: 'bar',
        color: '#1f78b4',
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

const addChartData = (data :any, type :string, ent:any, ctm:number, newCtm:number) => {
  let t = new Date(ctm * 60 * 1000)
  let d = 0
  switch (type) {
    case 'bytes':
      d = ent.bytes
      break
    case 'packets':
      d = ent.packets
      break
    case 'bps':
      if (ent.dur > 0) {
        d = ent.bytes / ent.dur
      }
      break
    case 'pps':
      if (ent.dur > 0) {
        d = ent.packets / ent.dur
      }
      break
  }
  data.push({
    name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
    value: [t, d],
  })
  ctm++
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000)
    data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
      value: [t, 0],
    })
  }
  return ctm
}

export const showNetFlowTraffic = (div:string, logs:any, type:string) => {
  makeNetFlowTraffic(div, type)
  const data :any = []
  const ent = {
    bytes: 0,
    packets: 0,
    dur: 0,
  }
  let ctm:number = 0
  logs.forEach((l:any) => {
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, type, ent, ctm, newCtm)
      ent.bytes = 0
      ent.dur = 0
      ent.packets = 0
    }
    ent.bytes += l.Bytes
    ent.dur += l.Dur
    ent.packets += l.Packets
  })
  addChartData(data, type, ent, ctm, ctm + 1)
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  })
  chart.resize();
  return chart;
}

export const showNetFlowTop = (div : string, list:any, type:string) => {
  const data = []
  const category = []

  let xAxis = ''
  switch (type) {
    case 'bytes':
      xAxis = $_("Ts.Bytes");
      list.sort((a:any, b:any) => b.Bytes - a.Bytes)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Bytes)
        category.push(list[i].Name)
      }
      break
    case 'packets':
      xAxis = $_("Ts.Packets");
      list.sort((a:any, b:any) => b.Packets - a.Packets)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Packets)
        category.push(list[i].Name)
      }
      break
    case 'dur':
      xAxis = $_("Ts.Dur")
      list.sort((a:any, b:any) => b.Dur - a.Dur)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Dur)
        category.push(list[i].Name)
      }
      break
    case 'bps':
      xAxis =$_("Ts.BPS");
      list.sort((a:any, b:any) => b.bps - a.bps)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].bps)
        category.push(list[i].Name)
      }
      break
    case 'pps':
      xAxis =  $_("Ts.PPS")
      list.sort((a:any, b:any) => b.pps - a.pps)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].pps)
        category.push(list[i].Name)
      }
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  chart.setOption({
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: 10,
      buttom: 10,
    },
    xAxis: {
      type: 'value',
      name: xAxis,
      boundaryGap: [0, 0.01],
    },
    yAxis: {
      type: 'category',
      data: category,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: xAxis,
        type: 'bar',
        data,
      },
    ],
  })
  chart.resize();
  return chart;
}

export const getNetFlowSenderList = (logs:any,mac: boolean) => {
  const m = new Map()
  logs.forEach((l:any) => {
    const k = mac ? l.SrcMAC : l.SrcAddr;
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Dur: l.Dur,
      })
    } else {
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Dur += l.Dur
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Dur > 0) {
      e.bps = (e.Bytes / e.Dur).toFixed(3)
      e.pps = (e.Packets / e.Dur).toFixed(3)
      e.Dur = e.Dur.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

export const getNetFlowServiceList = (logs:any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    let k = getServiceName(l.SrcPort + '/' + l.Protocol)
    if (k === 'Other') {
      k = getServiceName(l.DstPort + '/' + l.Protocol)
    }
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Dur: l.Dur,
      })
    } else {
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Dur += l.Dur
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Dur > 0) {
      e.bps = (e.Bytes / e.Dur).toFixed(3)
      e.pps = (e.Packets / e.Dur).toFixed(3)
      e.Dur = e.Dur.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

// Service Name Map
const serviceNameArray :any = [
  ['80/tcp', 'HTTP'],
  ['443/tcp', 'HTTPS'],
  ['389/tcp', 'LDAP'],
  ['636/tcp', 'LDAP'],
  ['53/tcp', 'DNS'],
  ['53/udp', 'DNS'],
  ['161/udp', 'SNMP'],
  ['162/udp', 'SNMP'],
  ['123/udp', 'NTP'],
  ['25/tcp', 'SMTP'],
  ['587/tcp', 'SMTP'],
  ['110/tcp', 'POP3'],
  ['995/tcp', 'POP3'],
  ['143/tcp', 'IMAP'],
  ['943/tcp', 'IMAP'],
  ['22/tcp', 'SSH'],
  ['21/tcp', 'TELNET'],
  ['23/tcp', 'FTP'],
  ['67/udp', 'DHCP'],
  ['68/udp', 'DHCP'],
  ['514/udp', 'SYSLOG'],
  ['2049/tcp', 'NFS'],
  ['2049/udp', 'NFS'],
  ['445/tcp', 'CIFS'],
  ['3389/tcp', 'RDP'],
  ['5900/tcp', 'VNC'],
  ['137/udp', 'NETBIOS'],
  ['137/tcp', 'NETBIOS'],
  ['138/udp', 'NETBIOS'],
  ['138/tcp', 'NETBIOS'],
  ['139/udp', 'NETBIOS'],
  ['139/tcp', 'NETBIOS'],
  ['88/tcp', 'AD'],
  ['7680/tcp', 'WUDO'],
  ['1812/udp', 'RADIUS'],
  ['5223/tcp', 'APPLE'],
  ['5228/tcp', 'ANDROID'],
]

const serviceNameMap = new Map(serviceNameArray)

const getServiceName = (s:any) => {
  const ret = serviceNameMap.get(s)
  if (ret) {
    return ret
  }
  if (s.indexOf('/icmp') > 0) {
    return 'ICMP'
  }
  return 'Other'
}

export const getNetFlowFlowList = (logs:any,mac:boolean) => {
  const m = new Map()
  logs.forEach((l:any) => {
    let k = mac ? l.SrcMAC + '<->' + l.DstMAC : l.SrcAddr + '<->' + l.DstAddr
    let e = m.get(k)
    if (!e) {
      k = mac ? l.DstMAC + '<->' + l.SrcMAC : l.DstAddr + '<->' + l.SrcAddr
      e = m.get(k)
    }
    if (!e) {
      m.set(k, {
        Name: k,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Dur: l.Dur,
      })
    } else {
      const kalt = mac ? l.SrcMAC + '<->' + l.DstMAC : l.SrcAddr + '<->' + l.DstAddr;
      if (k !== kalt) {
        // 逆報告もある場合
        e.bidir = true
      }
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Dur += l.Dur
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Dur > 0) {
      if (e.bidir) {
        e.Dur /= 2.0
      }
      e.bps = (e.Bytes / e.Dur).toFixed(3)
      e.pps = (e.Packets / e.Dur).toFixed(3)
      e.Dur = e.Dur.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

export const showNetFlowGraph = (div:string, logs:any,mode :number, type:string) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  logs.forEach((l:any) => {
    let n1 = "";
    let n2 = "";
    switch (mode) {
      case 1:
        n1 = l.SrcMAC;
        n2 = l.DstMAC;
        break;
      case 2:
        n1 = l.SrcAddr;
        n2 = l.SrcMAC;
        break;
      case 3:
        n1 = l.DstAddr;
        n2 = l.DstMAC;
        break;
      default:
        n1 = l.SrcAddr;
        n2 = l.DstAddr;
        break;
    }
    let ek = n1 + '|' + n2
    let e = edgeMap.get(ek)
    if (!e) {
      ek = n2 + '|' + n1
      e = edgeMap.get(ek)
    }
    if (!e) {
      edgeMap.set(ek, {
        source: n1,
        target: n2,
        value: l.Bytes,
      })
    } else {
      e.value += l.Bytes
    }
    let n = nodeMap.get(n1)
    if (!n) {
      nodeMap.set(n1, {
        name: n1,
        value: l.Bytes,
        draggable: true,
        category: getNodeCategory(n1),
      })
    } else {
      n.value += l.Bytes
    }
    n = nodeMap.get(n2)
    if (!n) {
      nodeMap.set(n2, {
        name: n2,
        value: 0,
        draggable: true,
        category: getNodeCategory(n2),
      })
    }
  })
  const nodes = Array.from(nodeMap.values())
  const edges = Array.from(edgeMap.values())
  const nvs :any = []
  const evs :any = []
  nodes.forEach((e) => {
    nvs.push(e.value)
  })
  edges.forEach((e) => {
    evs.push(e.value)
  })
  const n95 = ecStat.statistics.quantile(nvs, 0.95)
  const n50 = ecStat.statistics.quantile(nvs, 0.5)
  const e95 = ecStat.statistics.quantile(evs, 0.95)
  const categories = [
    { name: 'IPv4 Private' },
    { name: 'IPv6 Private' },
    { name: 'IPv4 Global' },
    { name: 'IPV6 Global' },
    { name: 'MAC Address' },
  ]
  let mul = 1.0
  if (type === 'gl') {
    mul = 1.5
  }
  nodes.forEach((e) => {
    e.label = { show: e.value > n95 }
    e.symbolSize = e.value > n95 ? 5 : e.value > n50 ? 4 : 2
    e.symbolSize *= mul
  })
  edges.forEach((e) => {
    e.lineStyle = {
      width: e.value > e95 ? 2 : 1,
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const options :any = {
    title: {
      show: false,
    },
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    tooltip: {
      trigger: 'item',
      textStyle: {
        fontSize: 8,
      },
      position: 'bottom',
    },
    legend: [
      {
        orient: 'vertical',
        top: 50,
        right: 20,
        textStyle: {
          fontSize: 10,
          color: '#ccc',
        },
        data: categories.map(function (a) {
          return a.name
        }),
      },
    ],
    color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99', '#fbca00'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  if (type === 'circular') {
    options.series = [
      {
        name: 'Flows',
        type: 'graph',
        layout: 'circular',
        circular: {
          rotateLabel: true,
        },
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          show: true,
          position: 'right',
          formatter: '{b}',
          fontSize: 10,
          fontStyle: 'normal',
          color: '#eee',
        },
        lineStyle: {
          color: 'source',
          curveness: 0.3,
        },
      },
    ]
  } else if (type === 'gl') {
    options.series = [
      {
        name: 'Flows',
        type: 'graphGL',
        nodes,
        edges,
        modularity: {
          resolution: 2,
          sort: true,
        },
        lineStyle: {
          color: 'source',
          opacity: 0.5,
        },
        itemStyle: {
          opacity: 1,
        },
        focusNodeAdjacency: false,
        focusNodeAdjacencyOn: 'click',
        emphasis: {
          label: {
            show: false,
          },
          lineStyle: {
            opacity: 0.5,
            width: 4,
          },
        },
        forceAtlas2: {
          steps: 5,
          stopThreshold: 20,
          jitterTolerence: 10,
          edgeWeight: [0.2, 1],
          gravity: 5,
          edgeWeightInfluence: 0,
        },
        categories,
        label: {
          show: true,
          position: 'right',
          formatter: '{b}',
          fontSize: 10,
          fontStyle: 'normal',
          color: '#eee',
        },
      },
    ]
  } else {
    options.series = [
      {
        name: 'Flows',
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 10,
          fontStyle: 'normal',
          color: '#eee',
        },
        lineStyle: {
          color: 'source',
          curveness: 0,
        },
      },
    ]
  }
  chart.setOption(options)
  chart.resize()
  return {chart:chart,edges:edges}
}

const getNodeCategory = (ip:string) => {
  if (isMACFormat(ip)){
    return 4
  }
  if (isPrivateIP(ip)) {
    if (isV4Format(ip)) {
      return 0
    }
    return 1
  }
  if (isV4Format(ip)) {
    return 2
  }
  return 3
}

export const showNetFlowService3D = (div:string, logs:any, type:string) => {
  const m = new Map()
  logs.forEach((l:any) => {
    let k = getServiceName(l.SrcPort + '/' + l.Protocol)
    if (k === 'Other') {
      k = getServiceName(l.DstPort + '/' + l.Protocol)
    }
    const ipt = getNodeCategory(l.SrcAddr)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Dur: [l.Dur],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Dur.push(l.Dur)
      e.IPType.push(ipt)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data :any  = []
  let dim :any = []
  switch (type) {
    case 'bytes':
      dim = ['Service', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'packets':
      dim = ['Service', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'dur':
      dim = ['Service', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Dur[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    title: {
      show: false,
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 4,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99', '#fbca00'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Service',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value:any) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'Service',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

export const showNetFlowSender3D = (div:string, logs:any, type:string, mac:boolean) => {
  const m = new Map()
  logs.forEach((l:any) => {
    const k = mac ? l.SrcMAC : l.SrcAddr;
    const ipt = getNodeCategory(l.SrcAddr)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name:k,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Dur: [l.Dur],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Dur.push(l.Dur)
      e.IPType.push(ipt)
    }
  })
  // 上位500件に絞る
  const ks = Array.from(m.keys())
  if (ks.length > 500) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.TotalBytes - ea.TotalBytes
    })
    for (let i = 500; i < ks.length; i++) {
      m.delete(ks[i])
    }
  }
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data :any = []
  let dim :any = []
  switch (type) {
    case 'bytes':
      dim = ['Sender', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'packets':
      dim = ['Sender', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'dur':
      dim = ['Sender', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Dur[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const options = {
    title: {
      show: false,
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 4,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99','#fbca00'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Sender',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value:any) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'Sender',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

export const showNetFlowFlow3D = (div:string, logs:any, type:string,mac:boolean) => {
  const m = new Map()
  logs.forEach((l:any) => {
    let k = mac ? l.SrcMAC + '<->' + l.DstMAC : l.SrcAddr + '<->' + l.DstAddr
    let e = m.get(k)
    if (!e) {
      k = mac ? l.DstMAC + '<->' + l.SrcMAC : l.DstAddr + '<->' + l.SrcAddr;
      e = m.get(k)
    }
    const ipt = getNodeCategory(l.SrcAddr)
    const t = new Date(l.Time / (1000 * 1000))
    if (!e) {
      m.set(k, {
        Name: k,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Dur: [l.Dur],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Dur.push(l.Dur)
      e.IPType.push(ipt)
    }
  })
  // 上位500件に絞る
  const ks = Array.from(m.keys())
  if (ks.length > 500) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.TotalBytes - ea.TotalBytes
    })
    for (let i = 500; i < ks.length; i++) {
      m.delete(ks[i])
    }
  }
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data :any = []
  let dim :any = []
  switch (type) {
    case 'bytes':
      dim = [mac ? 'MAC':'IPs', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'packets':
      dim = [mac ? 'MAC':'IPs', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Dur[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'dur':
      dim = [mac ? 'MAC':'IPs', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Dur[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const options = {
    title: {
      show: false,
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 4,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99', '#fbca00'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Src <-> Dst',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value:any) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'flow',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

export const getNetFlowFFTMap = (logs:any) => {
  const m = new Map()
  m.set('Total', { Name: 'Total', Count: 0, Data: [], Total: 0 })
  let packets = 0
  let st = Infinity
  let lt = 0
  logs.forEach((l:any) => {
    const e = m.get(l.SrcAddr)
    if (!e) {
      m.set(l.SrcAddr, { Name: l.SrcAddr, Count: 0, Data: [], Total: l.Packets })
    } else {
      e.Total += l.Packets
    }
    packets += l.Packets
    if (st > l.Time) {
      st = l.Time
    }
    if (lt < l.Time) {
      lt = l.Time
    }
  })
  m.get('Total').Total = packets
  let sampleSec = 1
  const dur = (lt - st) / (1000 * 1000 * 1000)
  if (dur > 3600 * 24 * 365) {
    sampleSec = 3600
  } else if (dur > 3600 * 24 * 30) {
    sampleSec = 600
  } else if (dur > 3600 * 24 * 7) {
    sampleSec = 120
  } else if (dur > 3600 * 24) {
    sampleSec = 60
  }
  let cts :number = 0;
  logs.forEach((l:any) => {
    if (!cts) {
      cts = Math.floor(l.Time / (1000 * 1000 * 1000 * sampleSec))
      m.get('Total').Count++
      m.get(l.SrcAddr).Count++
      return
    }
    const newCts = Math.floor(l.Time / (1000 * 1000 * 1000 * sampleSec))
    if (cts !== newCts) {
      m.forEach((e) => {
        e.Data.push(e.Count)
        e.Count = 0
      })
      cts++
      for (; cts < newCts; cts++) {
        m.forEach((e) => {
          e.Data.push(0)
        })
      }
    }
    m.get('Total').Count++
    m.get(l.SrcAddr).Count++
  })
  const ks = Array.from(m.keys())
  if (ks.length > 50) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.Total - ea.Total
    })
    for (let i = 0; i < ks.length; i++) {
      const e = m.get(ks[i])
      if (i > 50 || e.Total < 10) {
        m.delete(ks[i])
      }
    }
  }
  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1 / sampleSec)
  })
  return m
}

export const showNetFlowFFT = (div:string, fftMap:any, src:string, type:string) => {
  if (chart) {
    chart.dispose()
  }
  if (!fftMap || !fftMap.get(src)) {
    return
  }
  const fftData = fftMap.get(src).FFT
  const freq = type === 'hz'
  const fft :any = []
  if (freq) {
    fftData.forEach((e:any) => {
      fft.push([e.frequency, e.magnitude])
    })
  } else {
    fftData.forEach((e:any) => {
      fft.push([e.period, e.magnitude])
    })
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const options = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '10%',
      buttom: '10%',
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [{}],
    xAxis: {
      type: 'value',
      name: freq ? $_("Ts.FrequencyHz") : $_("Ts.CycleSec"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: '回数',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    series: [
      {
        name: '回数',
        type: 'bar',
        color: '#5470c6',
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: fft,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

export const showNetFlowFFT3D = (div:string, fftMap:any, fftType:string) => {
  const data :any = []
  const freq = fftType === 'hz'
  const colors :any = []
  const cat :any = []
  fftMap.forEach((e:any) => {
    if (e.Name === 'Total') {
      return
    }
    cat.push(e.Name)
    if (freq) {
      e.FFT.forEach((f:any) => {
        if (f.frequency === 0.0) {
          return
        }
        data.push([e.Name, f.frequency, f.magnitude, f.period])
        colors.push(f.magnitude)
      })
    } else {
      e.FFT.forEach((f:any) => {
        if (f.period === 0.0) {
          return
        }
        data.push([e.Name, f.period, f.magnitude, f.frequency])
        colors.push(f.magnitude)
      })
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const options = {
    title: {
      show: false,
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
      dimension: 2,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Src IP',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: freq ? 'value' : 'log',
      name: freq ?  $_("Ts.FrequencyHz") : $_("Ts.CycleSec"),
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: '回数',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'NetFlow FFT',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: [
          'Host',
          freq ? '周波数' : '周期',
          '回数',
          freq ? '周期' : '周波数',
        ],
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}
