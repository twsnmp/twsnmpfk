import * as echarts from 'echarts'
import { setZoomCallback } from './utils.js'
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart;

const makeLogLevelChart = (div:string) => {
  chart = echarts.init(document.getElementById(div),"dark");
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
        formatter: (value, index) => {
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
        name: $_("Ts.High"),
        type: 'bar',
        color: '#e31a1c',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: $_("Ts.Low"),
        type: 'bar',
        color: '#fb9a99',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: $_("Ts.Warn"),
        type: 'bar',
        color: '#dfdf22',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: $_("Ts.Other"),
        type: 'bar',
        color: '#1f78b4',
        stack: 'count',
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [$_("Ts.High"), $_("Ts.Low"), $_("Ts.Warn"), $_("Ts.Other")],
    },
  }
  chart.setOption(option);
  chart.resize();
}

const addChartData = (data, count, ctm, newCtm) => {
  let t = new Date(ctm * 60 * 1000)
  for (const k in count) {
    data[k].push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
      value: [t, count[k]],
    })
  }
  ctm++
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000)
    for (const k in count) {
      data[k].push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
        value: [t, 0],
      })
    }
  }
  return ctm
}

export const showLogLevelChart = (div:string, logs, zoomCallback) => {
  if (chart) {
    chart.dispose();
  }
  makeLogLevelChart(div)
  const data = {
    high: [],
    low: [],
    warn: [],
    other: [],
  }
  const count = {
    high: 0,
    low: 0,
    warn: 0,
    other: 0,
  }
  let ctm : undefined | number = undefined;
  let st = Infinity
  let lt = 0
  logs.forEach((e) => {
    const lvl = data[e.Level] ? e.Level : 'other'
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm;
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, count, ctm, newCtm);
      for (const k in count) {
        count[k] = 0;
      }
    }
    count[lvl]++;
    if (st > e.Time) {
      st = e.Time;
    }
    if (lt < e.Time) {
      lt = e.Time;
    }
  })
  addChartData(data, count, ctm, ctm + 1);
  chart.setOption({
    series: [
      {
        data: data.high,
      },
      {
        data: data.low,
      },
      {
        data: data.warn,
      },
      {
        data: data.other,
      },
    ],
  });
  chart.resize();
  setZoomCallback(chart, zoomCallback, st, lt);
}

export const resizeLogLevelChart = () => {
  if (chart) {
    chart.resize();
  }
}

