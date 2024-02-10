import * as echarts from 'echarts'
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart :any;

export const showAIHeatMap = (div:string, scores:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
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
      left: '10%',
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
          ' : ' +
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
      min: 40,
      max: 80,
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
  if (!scores) {
    chart.setOption(option)
    chart.resize()
    return
  }
  let nD = 0
  let x = -1
  scores.forEach((e:any) => {
    const t = new Date(e[0] * 1000)
    if (nD !== t.getDate()) {
      option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}',false))
      nD = t.getDate()
      x++
    }
    option.series[0].data.push([x, t.getHours(), e[1]])
  })
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showAIPieChart = (div:string, scores:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const option = {
    title: {
      show: false,
    },
    color: ['#1f78b4', '#dfdf22', '#e31a1c'],
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: [$_("Ts.Normal"), $_("Ts.Warn"), $_("Ts.Anamary")],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: $_("Ts.AnmaryScore"),
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [
          { name: $_("Ts.Normal"), value: 0 },
          { name: $_("Ts.Warn"), value: 0 },
          { name: $_("Ts.Anamary"), value: 0 },
        ],
      },
    ],
  }
  if (scores) {
    scores.forEach((e:any) => {
      if (e[1] > 66.0) {
        option.series[0].data[2].value++
      } else if (e[1] > 50.0) {
        option.series[0].data[1].value++
      } else {
        option.series[0].data[0].value++
      }
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showAITimeChart = (div:string, scores:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
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
        formatter(value:any, index:any) {
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
      name: $_("Ts.AnmaryScore"),
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
        name: $_("Ts.AnmaryScore"),
        showSymbol: false,
        data: [],
      },
    ],
  }
  if (scores) {
    scores.forEach((e:any) => {
      const t = new Date(e[0] * 1000)
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
      option.series[0].data.push({
        name: ts,
        value: [t, e[1]],
      })
    });
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}
