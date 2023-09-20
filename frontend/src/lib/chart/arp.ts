import * as echarts from 'echarts';
import 'echarts-gl';
import { isDark } from './utils'

let chart;

export const showArpLogIP = (div, logs) => {
  const list = getArpLogIPList(logs);
  const newLog = [];
  const changeLog = [];
  const ips = [];
  list.sort((a, b) => b.total - a.total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    newLog.push(list[i].newLog)
    changeLog.push(list[i].changeLog)
    ips.push(list[i].ip)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: '#4b5769',
      },
      {
        offset: 1,
        color: '#404a59',
      },
    ]),
    color: ['#1f78b4','#e31a1c'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['新規', '変化'],
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
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
      top: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '件数',
    },
    yAxis: {
      type: 'category',
      data: ips,
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
        name: '新規',
        type: 'bar',
        stack: '件数',
        data: newLog,
      },
      {
        name: '変化',
        type: 'bar',
        stack: '件数',
        data: changeLog,
      },
    ],
  })
  chart.resize()
}

const getArpLogIPList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    const e = m.get(l.IP)
    if (!e) {
      m.set(l.IP, {
        ip: l.IP,
        total: 1,
        newLog: l.State === 'New' ? 1 : 0,
        changeLog: l.State === 'Change' ? 1 : 0,
      })
    } else {
      e.total += 1;
      e.newLog += l.State === 'New' ? 1 : 0;
      e.changeLog += l.State === 'Change' ? 1 : 0;
    }
  })
  const r = Array.from(m.values())
  return r
}

export const showArpLogIP3D = (div, logs) => {
  const m = new Map()
  logs.forEach((l) => {
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(l.IP)
    if (!e) {
      m.set(l.IP, {
        ip: l.IP,
        time: [t],
        state: [l.State],
        level: [l.State == "New" ? 0 : 1],
      })
    } else {
      e.time.push(t)
      e.state.push(l.State)
      e.level.push(l.State == "New" ? 0 : 1)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  l.forEach((e) => {
    for (let i = 0; i < e.time.length && i < 15000; i++) {
      data.push([e.ip, e.time[i],e.state[i],e.level[i]])
    }
  })
  if (chart) {
    chart.dispose()
  }
  const dark = isDark();
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "");
  const options = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: dark ? "#ccc" : "#222",
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 1,
      dimension: 3,
      inRange: {
        color: ['#1f78b4','#e31a1c'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'ip',
      data: cat,
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: dark ? "#ccc" : "#222",
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'time',
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: 8,
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: dark ? "#ccc" : "#222",
        },
      },
    },
    zAxis3D: {
      type: 'category',
      name: 'satte',
      data: ["New","Change"],
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: dark ? "#ccc" : "#222",
        },
      },
    },
    grid3D: {
      axisLine: {
        color: dark ? "#ccc" : "#222",
      },
      axisPointer: {
        color: dark ? "#ccc" : "#222",
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'IP別ログ件数',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: ['ip', 'time', 'type'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export const showArpGraph = (div, arp, type,changeIP,changeMAC) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  arp.forEach((a) => {
    let ek = a.IP + '|' + a.MAC;
    let e = edgeMap.get(ek)
    if (!e) {
      edgeMap.set(ek, {
        source: a.IP,
        target: a.MAC,
        lineStyle:{
          width: 2,
        },
      })
    }
    let n = nodeMap.get(a.IP)
    if (!n) {
      nodeMap.set(a.IP, {
        name: a.IP,
        draggable: true,
        category: changeIP.has(a.IP) ? 0 : a.IP.startsWith("169.254.") ? 1 : 2,
        symbolSize: 4,
        label: { show: true },
      })
    }
    n = nodeMap.get(a.MAC)
    if (!n) {
      nodeMap.set(a.MAC, {
        name: a.MAC,
        draggable: true,
        category: changeMAC.has(a.MAC) ? 3 : 4,
        symbolSize: 4,
        label: { show: true },
      })
    }
  })
  const nodes = Array.from(nodeMap.values())
  const edges = Array.from(edgeMap.values())
  const categories = [
    { name: 'IP Changed' },
    { name: 'IP Duplicate' },
    { name: 'IP Normal' },
    { name: 'MAC Changed' },
    { name: 'MAC Normal' },
  ]
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: '#4b5769',
      },
      {
        offset: 1,
        color: '#404a59',
      },
    ]),
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
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
    color: ['#ea0','#e31a1c','#1f78b4','#cc0', '#165ee3' ],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  if (type === 'circular') {
    options.series = [
      {
        name: 'IP Flows',
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
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
          color: '#ccc',
        },
        lineStyle: {
          color: 'source',
          curveness: 0.3,
        },
      },
    ]
  } else {
    options.series = [
      {
        name: 'IP Flows',
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
          color: '#ccc',
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
}