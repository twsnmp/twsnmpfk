import * as echarts from 'echarts';
import 'echarts-gl';
import * as ecStat from 'echarts-stat';
import { doFFT } from './fft'
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart :any;

export const showSyslogHost = (div:string, logs:any) => {
  const list = getSyslogHostList(logs);
  const high = []
  const low = []
  const warn = []
  const other = []
  const category = []
  list.sort((a, b) => b.Total - a.Total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    high.push(list[i].High)
    low.push(list[i].Low)
    warn.push(list[i].Warn)
    other.push(list[i].Other)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  chart.setOption({
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [$_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"),$_("Ts.Other")],
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
      name: $_("Ts.NumberOfLog"),
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
        name: $_("Ts.High"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: high,
      },
      {
        name: $_("Ts.Low"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: low,
      },
      {
        name: $_("Ts.Warn"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: warn,
      },
      {
        name: $_("Ts.Other"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: other,
      },
    ],
  })
  chart.resize();
  return chart;
}

const getSyslogHostList = (logs:any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, {
        Name: l.Host,
        Total: 1,
        High: l.Level === 'high' ? 1 : 0,
        Low: l.Level === 'low' ? 1 : 0,
        Warn: l.Level === 'warn' ? 1 : 0,
        Other: l.Level === 'info' || l.Level === 'debug' ? 1 : 0,
      })
    } else {
      e.Total += 1;
      e.High += l.Level === 'high' ? 1 : 0;
      e.Low += l.Level === 'low' ? 1 : 0;
      e.Warn += l.Level === 'warn' ? 1 : 0;
      e.Other += l.Level === 'info' || l.Level === 'debug' ? 1 : 0;
    }
  })
  const r = Array.from(m.values())
  return r
}

export const showSyslogHost3D = (div:string, logs:any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    const level = getSyslogLevel(l.Level)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, {
        Name: l.Host,
        Total: 1,
        Time: [t],
        Priority: [l.Severity + l.Facility * 8],
        Level: [level],
      })
    } else {
      e.Total += 1
      e.Time.push(t)
      e.Priority.push(l.Severity + l.Facility * 8)
      e.Level.push(level)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data :any = []
  l.forEach((e) => {
    for (let i = 0; i < e.Time.length && i < 15000; i++) {
      data.push([e.Name, e.Time[i], e.Priority[i], e.Level[i]])
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const options :any = {
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
      dimension: 3,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host',
      data: cat,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
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
        color: '#ccc',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
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
      name: 'Priority',
      nameTextStyle: {
        color: '#ccc',
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
        color: '#ccc',
      },
      axisPointer: {
        color: '#ccc',
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: $_("Ts.LogCountByHost"),
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: ['Host', 'Time', 'Priority', 'Level'],
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

const getSyslogLevel = (l:any) => {
  switch (l) {
    case 'high':
      return 0
    case 'low':
      return 1
    case 'warn':
      return 2
    case 'debug':
      return 4
  }
  return 3
}


const getSyslogFFTMap = (logs:any) => {
  const m = new Map();
  if (logs.length < 50) {
    return m;
  }
  let st = Infinity;
  let lt = 0;
  m.set('Total', { Name: 'Total', Count: 0, Data: [] })
  logs.forEach((l:any) => {
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, { Name: l.Host, Count: 0, Data: [] })
    }
    if (st > l.Time) {
      st = l.Time
    }
    if (lt < l.Time) {
      lt = l.Time
    }
  })
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
  let cts : any = undefined;
  logs.forEach((l:any) => {
    if (!cts) {
      cts = Math.floor(l.Time / (1000 * 1000 * 1000 * sampleSec))
      m.get('Total').Count++
      m.get(l.Host).Count++
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
    m.get(l.Host).Count++
  })
  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1 / sampleSec)
  })
  return m
}

export const showSyslogFFT3D = (div :string, logs:any) => {
  const fftMap = getSyslogFFTMap(logs);
  const data :any = [];
  const colors :any = [];
  const cat :any = [];
  fftMap.forEach((e) => {
    if (e.Name === 'Total') {
      return
    }
    cat.push(e.Name)
    e.FFT.forEach((f:any) => {
      if (f.period === 0.0) {
        return
      }
      data.push([e.Name, f.period, f.magnitude, f.frequency])
      colors.push(f.magnitude)
    })
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
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
      name: 'Host',
      data: cat,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
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
      type: 'log',
      name: $_("Ts.CycleSec"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
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
      name: $_("Ts.Count"),
      nameTextStyle: {
        color: '#ccc',
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
        lineStyle: { 
          color: '#ccc',
        },
      },
      axisPointer: {
        lineStyle: { 
          color: '#ccc',
        },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'Syslog FFT',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: [
          'Host',
          $_("Ts.Cycle"),
          $_("Ts.Count"),
          $_("Ts.Frequency"),
        ],
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}


export const showSyslogLevelChart = (div:string, logs:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const option = {
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4'],
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: [$_("Ts.High"),$_("Ts.Low") ,$_("Ts.Warn"), $_("Ts.Other")],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: $_("Ts.CountByState"),
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [
          { name: $_("Ts.High"), value: 0 },
          { name: $_("Ts.Low"), value: 0 },
          { name: $_("Ts.Warn"), value: 0 },
          { name: $_("Ts.Other"), value: 0 },
        ],
      },
    ],
  }
  if (logs) {
    logs.forEach((l:any) => {
      switch (l.Level) {
        case 'high':
          option.series[0].data[0].value++
          break
        case 'low':
          option.series[0].data[1].value++
          break
        case 'warn':
          option.series[0].data[2].value++
          break
        default:
          option.series[0].data[3].value++
      }
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const getSyslogSummary = (logs :any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    const p = normalizeLog(
      l.Host + ' ' + l.Type + ' ' + l.Tag + ' ' + l.Message
    )
    if (!m.has(p)) {
      const d = new Date(l.Time /(1000*1000));
      const format = '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}';
      const sampleLog =
        echarts.time.format(d,format,false) + ' ' + l.Host + ' ' + l.Type + ' ' + l.Tag + ' ' + l.Message
      m.set(p, { Sample: sampleLog, Count: 1 })
    } else {
      m.get(p).Count++
    }
  })
  const list = []
  for (const [p, e] of m.entries()) {
    list.push({ Pattern: p, Sample: e.Sample, Count: e.Count })
  }
  return list
}

const regNum = /\b-?\d+(\.\d+)?\b/g
const regUUID = /[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}/gi
const regEmail = /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b/g
const regIP = /\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b/g
const regMAC = /\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b/g

function normalizeLog(log:string) {
  log = log.replace(regUUID, '#UUID#')
  log = log.replace(regEmail, '#EMAIL#')
  log = log.replace(regIP, '#IP#')
  log = log.replace(regMAC, '#MAC#')
  log = log.replace(regNum, '#NUM#')
  return log
}

export const showSyslogSummary = (div:string, list:any) => {
  const count = []
  const category = []
  list.sort((a:any, b:any) => b.Count - a.Count)
  for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
    count.push(list[i].Count)
    category.push(truncateString(list[i].Pattern, 100))
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div), 'dark')
  chart.setOption({
    title: {
      show: false,
    },
    toolbox: {
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
      name: 'Count',
    },
    yAxis: {
      type: 'category',
      data: category,
      nameTextStyle: {
        color: '#fff',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: 'Count',
        type: 'bar',
        data: count,
      },
    ],
  })
  chart.resize();
  return chart;
}

const truncateString = (str:string, maxLength:number) => {
  if (str.length > maxLength) {
    return str.slice(0, maxLength) + '...'
  }
  return str
}
