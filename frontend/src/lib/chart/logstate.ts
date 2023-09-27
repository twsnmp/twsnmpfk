import * as echarts from 'echarts'
import { setZoomCallback } from './utils.js'
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart

const makeLogStateChart = (div) => {
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        name: $_("Ts.Normal"),
        type: 'bar',
        color: '#33a02c',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: $_("Ts.Unknown"),
        type: 'bar',
        color: 'gray',
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
      data: [$_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"), $_("Ts.Normal"), $_("Ts.Unknown")],
    },
  }
  chart.setOption(option)
  chart.resize()
}

export const showLogStateChart = (div:string, logs:any, zoomCallback:any) => {
  if (chart) {
    chart.dispose();
  }
  makeLogStateChart(div);
  const data = {
    high: [],
    low: [],
    warn: [],
    normal: [],
    unknown: [],
  };
  const count = {
    high: 0,
    low: 0,
    warn: 0,
    normal: 0,
    unknown: 0,
  };
  let cth :number | undefined = undefined;
  let st = Infinity;
  let lt = 0;
  logs.forEach((e:any) => {
    const lvl = data[e.State] ? e.State : 'normal';
    if (!cth) {
      cth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600));
      count[lvl]++;
      return;
    }
    const newCth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600));
    if (cth !== newCth) {
      let t = new Date(cth * 3600 * 1000);
      for (const k in count) {
        data[k].push({
          name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
          value: [t, count[k]],
        });
      }
      cth++
      for (; cth < newCth; cth++) {
        t = new Date(cth * 3600 * 1000);
        for (const k in count) {
          data[k].push({
            name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
            value: [t, 0],
          });
        }
      }
      for (const k in count) {
        count[k] = 0;
      }
    }
    count[lvl]++
    if (st > e.Time) {
      st = e.Time;
    }
    if (lt < e.Time) {
      lt = e.Time;
    }
  })
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
        data: data.normal,
      },
      {
        data: data.unknown,
      },
    ],
  });
  chart.resize();
  setZoomCallback(chart, zoomCallback, st, lt);
}
