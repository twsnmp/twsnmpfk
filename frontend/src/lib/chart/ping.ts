import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import * as numeral from 'numeral';
import { isDark } from './utils';
import WorldData from 'world-map-geojson';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart;

export const getPingChartOption = () => {
  const dark = isDark();
  return {
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
      left: '5%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: [$_("Ts.RespTimeSec"), $_("Ts.SendTTL"),$_("Ts.RecvTTL")],
      textStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: $_("Ts.DateTime"),
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: '8px',
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: dark ? "#ccc" : "#222",
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: $_("Ts.RespTimeSec"),
        nameTextStyle: {
          color: dark ? "#ccc" : "#222",
          fontSize: 10,
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
      {
        type: 'value',
        name: 'TTL',
        nameTextStyle: {
          color: dark ? "#ccc" : "#222",
          fontSize: 10,
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
    ],
    series: [
      {
        name: $_("Ts.RespTimeSec"),
        color: '#1f78b4',
        type: 'line',
        showSymbol: false,
        data: [],
      },
      {
        name: $_("Ts.SendTTL"),
        color: '#dfdf22',
        type: 'line',
        showSymbol: false,
        yAxisIndex: 1,
        data: [],
      },
      {
        name: $_("Ts.RecvTTL"),
        color: '#e31a1c',
        type: 'line',
        showSymbol: false,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
}

export const showPing3DChart = (div, results) => {
  if (chart) {
    chart.dispose();
  }
  let maxRtt = 0.0
  const data = []
  results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    const t = new Date(r.TimeStamp * 1000)
    const rtt = r.Time / (1000 * 1000 * 1000)
    if (rtt > maxRtt) {
      maxRtt = rtt
    }
    data.push([r.Size, t, rtt])
  })
  const dark = isDark();
  chart = echarts.init(document.getElementById(div),dark ? "dark" : "");
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
      max: maxRtt,
      dimension: 2,
      inRange: {
        color: [
          '#1710c0',
          '#0b9df0',
          '#00fea8',
          '#00ff0d',
          '#f5f811',
          '#f09a09',
          '#fe0300',
        ],
      },
    },
    xAxis3D: {
      type: 'value',
      name: $_("Ts.Size"),
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
      name: $_("Ts.DateTime"),
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
      type: 'value',
      name: $_("Ts.RespTimeSec"),
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
        lineStyle: { 
          color: dark ? "#ccc" : "#222",
        },
      },
      axisPointer: {
        lineStyle: { 
          color: dark ? "#ccc" : "#222",
        },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: $_("Ts.PingAn3D"),
        type: 'scatter3D',
        symbolSize: 10,
        dimensions: [$_("Ts.Size"),$_("Ts.DateTime"),$_("Ts.RespTimeSec")],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export const showPingMapChart = (div, results) => {
  const dark = isDark();
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  echarts.registerMap('world', WorldData)
  const option = {
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    geo: {
      map: 'world',
      silent: true,
      emphasis: {
        label: {
          show: false,
          areaColor: dark ? '#eee' : '#222',
        },
      },
      itemStyle: {
        borderWidth: 0.2,
        borderColor: '#404a59',
      },
      roam: true,
    },
    series: [
      {
        type: 'scatter',
        coordinateSystem: 'geo',
        label: {
          formatter: '{b}',
          position: 'right',
          color: '#0ef',
          show: true,
          fontSize: 12,
        },
        emphasis: {
          label: {
            show: true,
          },
        },
        symbolSize: 10,
        itemStyle: {
          color: (params) => {
            const t = params.data.value[2]
            if (t < 0.005) {
              return '#1f78b4'
            } else if (t < 0.05) {
              return '#a6cee3'
            } else if (t < 0.2) {
              return '#dfdf22'
            } else if (t < 0.8) {
              return '#fb9a99'
            }
            return '#e31a1c'
          },
        },
        data: [],
      },
    ],
  }
  if (!results) {
    return
  }
  const locMap = {}
  results.forEach((e) => {
    const loc = e.Loc
    if (!loc || loc.indexOf('LOCAL') === 0) {
      return
    }
    if (!locMap[loc] || locMap[loc].time > e.Time) {
      locMap[loc] = {
        time: e.Time,
        ip: e.RecvSrc,
      }
    }
  })
  for (const k in locMap) {
    const a = k.split(',')
    if (a.length < 4 || !a[1]) {
      continue
    }
    option.series[0].data.push({
      name: locMap[k].ip + ':' + a[3] + '/' + a[0],
      value: [
        Number(a[2]) * 1.0,
        Number(a[1]) * 1.0,
        (locMap[k].time / (1000 * 1000 * 100)).toFixed(6),
      ],
    })
  }
  chart.setOption(option)
  chart.resize()
  chart.on('dblclick', (p) => {
    const url =
      'https://www.google.com/maps/search/?api=1&zoom=10&query=' +
      p.value[1] +
      ',' +
      p.value[0]
    window.open(url, '_blank')
  })
}

export const showPingHistgram = (div, results) => {
  const dark = isDark();
  if (chart) {
    chart.dispose()
  }
  const data = []
  results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    data.push(r.Time / (1000 * 1000 * 1000))
  })
  const bins = ecStat.histogram(data,"squareRoot")
  chart = echarts.init(document.getElementById(div))
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      formatter(params) {
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
      name: $_("Ts.RespTimeSec"),
      min: 0,
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: 8,
        margin: 2,
      },
    },
    yAxis: {
      name: $_("Ts.Count"),
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: dark ? "#ccc" : "#222",
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: bins.data,
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

export const showPingLinearChart = (div, results) => {
  const dark = isDark();
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const data = []
  results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    data.push([r.Size * 8, r.Time / (1000 * 1000 * 1000)])
  })
  const reg :any = ecStat.regression('linear', data,0)
  const speed =
    numeral(reg.parameter.gradient ? 1.0 / reg.parameter.gradient : 0.0).format(
      '0.00a'
    ) + 'bps';
  const delay = reg.parameter.intercept.toFixed(6) + `sec`
  const option = {
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
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
      name: $_("Ts.Size"),
      nameTextStyle: {
        color: dark ? "#ccc" : "#222",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: dark ? "#ccc" : "#222",
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: $_("Ts.RespTimeSec"),
        nameTextStyle: {
          color: dark ? "#ccc" : "#222",
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: dark ? "#ccc" : "#222",
          },
        },
        axisLabel: {
          color: dark ? "#ccc" : "#222",
          fontSize: 8,
          margin: 2,
        },
      },
    ],
    series: [
      {
        name: 'scatter',
        type: 'scatter',
        label: {
          emphasis: {
            show: true,
          },
        },
        data,
      },
      {
        name: 'line',
        type: 'line',
        showSymbol: false,
        data: reg.points,
        markPoint: {
          itemStyle: {
            normal: {
              color: 'transparent',
            },
          },
          label: {
            normal: {
              show: true,
              formatter: `回線速度=${speed} 遅延=${delay}`,
              textStyle: {
                color: dark ? "#ccc" : "#222",
                fontSize: 12,
              },
            },
          },
          data: [
            {
              coord: reg.points[reg.points.length - 1],
            },
          ],
        },
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}
