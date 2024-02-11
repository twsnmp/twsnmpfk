import * as echarts from 'echarts';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart:any;

export const showLogHeatmap = (div:string, logs:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const hours = [
    '0',
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    '12',
    '13',
    '14',
    '15',
    '16',
    '17',
    '18',
    '19',
    '20',
    '21',
    '22',
    '23',
  ]
  const option :any = {
    title: {
      show: false,
    },
    grid: {
      left: '5%',
      right: '5%',
      top: 30,
      buttom: 0,
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
    tooltip: {
      trigger: 'item',
      formatter(params:any) {
        return (
          params.name +
          ' ' +
          params.data[1] +
          ': ' +
          params.data[2].toFixed(1)
        )
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    xAxis: {
      type: 'category',
      name: $_("Ts.Date"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
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
      data: [],
    },
    yAxis: {
      type: 'category',
      name: $_("Ts.TimeRange"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
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
      data: hours,
    },
    visualMap: {
      min: Infinity,
      max: -Infinity,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
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
    series: [
      {
        name: 'Score',
        type: 'heatmap',
        data: [],
        emphasis: {
          itemStyle: {
            borderColor: '#ccc',
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  }
  if (logs) {
    let nD = 0
    let nH = 0
    let x = -1
    let sum = 0
    logs.sort((a:any, b:any) => a.Time - b.Time)
    logs.forEach((l:any) => {
      const t = new Date(l.Time / (1000 * 1000))
      if (nD === 0) {
        nH = t.getHours()
        nD = t.getDate()
        option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}',false))
        x++
        sum++
        return
      }
      if (t.getHours() !== nH) {
        if (nD !== t.getDate()) {
          option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}',false))
          nD = t.getDate()
          x++
        }
        option.series[0].data.push([x, t.getHours(), sum])
        if (option.visualMap.min > sum) {
          option.visualMap.min = sum
        }
        if (option.visualMap.max < sum) {
          option.visualMap.max = sum
        }
        sum = 0
        nH = t.getHours()
      }
      sum++
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showEventLogStateChart = (div:string, logs:any) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option = {
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: [ $_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"), $_("Ts.Normal"),$_("Ts.Repair"),$_("Ts.Other")],
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
          { name: $_("Ts.Normal"), value: 0 },
          { name: $_("Ts.Repair"), value: 0 },
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
        case 'normal':
          option.series[0].data[3].value++
          break
        case 'repair':
          option.series[0].data[4].value++
          break
        default:
          option.series[0].data[5].value++
      }
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showEventLogTimeChart = (div:string, type:any, logs:any) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option :any = {
    title: {
      show: false,
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
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: $_("Ts.DateTime"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
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
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: type === 'oprate' ? $_("Ts.Oprate")+ "%" : $_("Ts.Usage") +"%",
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
        color: '#1f78b4',
        type: 'line',
        name: type === 'oprate' ? $_("Ts.Oprate") : $_("Ts.Usage"),
        showSymbol: false,
        data: [],
      },
    ],
  }
  if (logs) {
    logs.forEach((l:any) => {
      if (l.Type !== type) {
        return
      }
      const t = new Date(l.Time / (1000 * 1000))
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
      const m = l.Event.match(/[0-9.]+%/)
      if (!m || m.length < 1) {
        return
      }
      const val = m[0].replace('%', '') * 1.0
      option.series[0].data.push({
        name: ts,
        value: [t, val],
      })
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

const getEventLogNodeList = (logs:any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    if (!l.NodeID) {
      return
    }
    let e = m.get(l.NodeID)
    if (!e) {
      m.set(l.NodeID, {
        Name: l.NodeName,
        total: 0,
        high: 0,
        low: 0,
        warn: 0,
        normal: 0,
        repair: 0,
        other: 0,
      })
      e = m.get(l.NodeID)
      if (!e) {
        return
      }
    }
    e.total++
    switch (l.Level) {
      case 'high':
        e.high++
        break
      case 'low':
        e.low++
        break
      case 'warn':
        e.warn++
        break
      case 'normal':
        e.normal++
        break
      case 'repair':
        e.repair++
        break
      default:
        e.other++
    }
  })
  const r = Array.from(m.values())
  return r
}

export const showEventLogNodeChart = (div:any, logs:any) => {
  const list = getEventLogNodeList(logs)
  const high = []
  const low = []
  const warn = []
  const normal = []
  const repair = []
  const other = []
  const category = []
  list.sort((a, b) => b.total - a.total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    high.push(list[i].high)
    low.push(list[i].low)
    warn.push(list[i].warn)
    normal.push(list[i].normal)
    repair.push(list[i].repair)
    other.push(list[i].other)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  chart.setOption({
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [ $_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"), $_("Ts.Normal"),$_("Ts.Repair"),$_("Ts.Other")],
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
        name: $_("Ts.Normal"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: normal,
      },
      {
        name: $_("Ts.Repair"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: repair,
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
