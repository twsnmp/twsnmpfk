import * as echarts from 'echarts'
import { setZoomCallback } from './utils.js'
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart:any;

const makeLogCountChart = (div:string) => {
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
      left: '5%',
      right: '5%',
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: $_("Ts.DateTime"),
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value:any, index:any) => {
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
      name: $_("Ts.NumberOfLog"),
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
        name: $_("Ts.NumberOfLog"),
        color: '#1f78b4',
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const addChartData = (data:any, count:number, ctm:number, newCtm:number) => {
  let t = new Date(ctm * 60 * 1000)
  data.push({
    name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
    value: [t, count],
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

export const showLogCountChart = (div:string, logs:any, zoomCallback:any) => {
  if (chart) {
    chart.dispose()
  }
  makeLogCountChart(div)
  const data:any = []
  let count = 0
  let ctm :number  = 0;
  let st = Infinity;
  let lt = 0;
  logs.sort((a:any, b:any) => a.Time - b.Time);
  logs.forEach((e:any) => {
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, count, ctm, newCtm)
      count = 0
    }
    count++
    if (st > e.Time) {
      st = e.Time
    }
    if (lt < e.Time) {
      lt = e.Time
    }
  })
  addChartData(data, count, ctm, ctm + 1)
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  })
  chart.resize()
  setZoomCallback(chart, zoomCallback, st, lt)
}

export const resizeLogCountChart = () => {
  if (chart) {
    chart.resize()
  }
}

