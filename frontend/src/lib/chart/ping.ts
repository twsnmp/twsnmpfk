import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import numeral from 'numeral';
//@ts-ignore
import WorldData from 'world-map-geojson';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart :any;

export const getPingChartOption = () => {
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
      top: 45,
      bottom: 60,
    },
    legend: {
      top: 25,
      data: [$_("Ts.RespTimeSec"), $_("Ts.SendTTL"),$_("Ts.RecvTTL")],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
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
    yAxis: [
      {
        type: 'value',
        name: $_("Ts.RespTimeSec"),
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
      {
        type: 'value',
        name: 'TTL',
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

export const showPing3DChart = (div:string, results:any) => {
  if (chart) {
    chart.dispose();
  }
  let maxRtt = 0.0
  const data :any= []
  results.forEach((r:any) => {
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
      name: $_("Ts.DateTime"),
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
      name: $_("Ts.RespTimeSec"),
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
        name: $_("Ts.PingAn3D"),
        type: 'scatter3D',
        symbolSize: 10,
        dimensions: [$_("Ts.Size"),$_("Ts.DateTime"),$_("Ts.RespTimeSec")],
        data,
      },
    ],
  }
  chart.setOption(options);
  chart.resize();
  return chart;
}

export const showPingMapChart = (div :string, results:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  echarts.registerMap('world', WorldData)
  const option :any = {
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
          areaColor: '#ccc',
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
          color: (params:any) => {
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
  const locMap :any = {}
  results.forEach((e:any) => {
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
  chart.setOption(option);
  chart.resize();
  chart.on('dblclick', (p:any) => {
    const url =
      'https://www.google.com/maps/search/?api=1&zoom=10&query=' +
      p.value[1] +
      ',' +
      p.value[0]
    window.open(url, '_blank')
  });
  return chart;
}

export const showPingHistgram = (div:string, results:any) => {
  if (chart) {
    chart.dispose()
  }
  const data :any = []
  results.forEach((r:any) => {
    if (r.Stat !== 1) {
      return
    }
    data.push(r.Time / (1000 * 1000 * 1000))
  })
  const bins = ecStat.histogram(data,"squareRoot")
  chart = echarts.init(document.getElementById(div),"dark")
  const option = {
    title: {
      show: false,
    },
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
      bottom: 60,
    },
    xAxis: {
      scale: true,
      name: $_("Ts.RespTimeSec"),
      min: 0,
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
    },
    yAxis: {
      name: $_("Ts.Count"),
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
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showPingLinearChart = (div:string, results:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark")
  const data :any = []
  results.forEach((r:any) => {
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
      top: 35,
      bottom: 60,
    },
    xAxis: {
      type: 'value',
      name: $_("Ts.Size"),
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
    yAxis: [
      {
        type: 'value',
        name: $_("Ts.RespTimeSec"),
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
              formatter: $_('Ping.LineSpeedValue') + `=${speed} ` + $_('Ping.DelayValue') + `=${delay}`,
              textStyle: {
                color: '#ccc',
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
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showPingSmokeChart = (div: string, results: any) => {
  if (chart) {
    chart.dispose();
  }
  if (!results || results.length === 0) {
    return;
  }

  let windowSize = 5;
  if (results.length >= 1000) {
    windowSize = 25;
  } else if (results.length >= 300) {
    windowSize = 10;
  } else if (results.length < 10) {
    windowSize = Math.max(1, Math.floor(results.length / 2));
  }

  const times: string[] = [];
  const minData: (number | null)[] = [];
  const diffData: (number | null)[] = [];
  const medianData: (number | null)[] = [];
  const maxData: (number | null)[] = [];
  const jitterData: (number | null)[] = [];
  const lossData: number[] = [];
  const lossDetails: string[] = [];

  for (let i = 0; i < results.length; i += windowSize) {
    const chunk = results.slice(i, i + windowSize);
    const lastResult = chunk[chunk.length - 1];
    const t = new Date(lastResult.TimeStamp * 1000);
    const timeStr = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}', false);
    times.push(timeStr);

    const valids: number[] = [];
    let lossCount = 0;

    chunk.forEach((r: any) => {
      if (r.Stat === 1 || r.Stat === 4) {
        valids.push(r.Time / (1000 * 1000 * 1000));
      } else {
        lossCount++;
      }
    });

    const lossRate = (lossCount / chunk.length) * 100;
    lossData.push(Number(lossRate.toFixed(1)));
    lossDetails.push(`${lossCount} / ${chunk.length} pkts (${lossRate.toFixed(1)}%)`);

    if (valids.length > 0) {
      valids.sort((a, b) => a - b);
      const min = valids[0];
      const max = valids[valids.length - 1];
      const jitter = max - min;
      const midIdx = Math.floor(valids.length / 2);
      const median = valids.length % 2 === 0 ? (valids[midIdx - 1] + valids[midIdx]) / 2 : valids[midIdx];

      minData.push(Number(min.toFixed(6)));
      diffData.push(Number(jitter.toFixed(6)));
      medianData.push(Number(median.toFixed(6)));
      maxData.push(Number(max.toFixed(6)));
      jitterData.push(Number(jitter.toFixed(6)));
    } else {
      minData.push(null);
      diffData.push(null);
      medianData.push(null);
      maxData.push(null);
      jitterData.push(null);
    }
  }

  chart = echarts.init(document.getElementById(div), "dark");
  const option: any = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      formatter(params: any) {
        if (!params || params.length === 0) return '';
        const idx = params[0].dataIndex;
        const time = times[idx];
        const minVal = minData[idx] !== null ? minData[idx] + ' s' : 'N/A';
        const medVal = medianData[idx] !== null ? medianData[idx] + ' s' : 'N/A';
        const maxVal = maxData[idx] !== null ? maxData[idx] + ' s' : 'N/A';
        const jitVal = jitterData[idx] !== null ? jitterData[idx] + ' s' : 'N/A';
        const lossVal = lossDetails[idx];

        return `<div class="text-xs font-sans space-y-1">` +
               `<div class="font-bold border-b border-gray-600 pb-1 mb-1">${time}</div>` +
               `<div><span class="inline-block w-3 h-3 mr-1 bg-[#00fea8] rounded-full"></span>${$_("Ping.MedianRespTime") || "Median RTT"}: <strong>${medVal}</strong></div>` +
               `<div><span class="inline-block w-3 h-3 mr-1 bg-[#38bdf8] rounded-sm"></span>Smoke (Min-Max): ${minVal} ~ ${maxVal}</div>` +
               `<div>${$_("Ping.Jitter") || "Jitter"}: ${jitVal}</div>` +
               `<div>${$_("Ping.LossRate") || "Loss Rate"}: <strong>${lossVal}</strong></div>` +
               `</div>`;
      },
    },
    grid: {
      left: '7%',
      right: '7%',
      top: 45,
      bottom: 60,
    },
    legend: {
      top: 15,
      data: [
        $_("Ping.MedianRespTime") || "Median RTT",
        "Smoke Range (Min-Max)",
        $_("Ping.LossRate") || "Loss Rate (%)"
      ],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'category',
      data: times,
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
    yAxis: [
      {
        type: 'value',
        name: $_("Ts.RespTimeSec"),
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
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
        splitLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.1)',
          },
        },
      },
      {
        type: 'value',
        name: $_("Ping.LossRate") || 'Loss (%)',
        min: 0,
        max: 100,
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          formatter: '{value}%',
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
    ],
    series: [
      {
        name: 'Smoke Base',
        type: 'line',
        stack: 'smoke',
        symbol: 'none',
        lineStyle: { opacity: 0 },
        data: minData,
      },
      {
        name: 'Smoke Range (Min-Max)',
        type: 'line',
        stack: 'smoke',
        symbol: 'none',
        lineStyle: { opacity: 0 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(56, 189, 248, 0.65)' },
            { offset: 1, color: 'rgba(56, 189, 248, 0.12)' }
          ])
        },
        data: diffData,
      },
      {
        name: $_("Ping.MedianRespTime") || "Median RTT",
        type: 'line',
        showSymbol: true,
        symbolSize: 6,
        itemStyle: { color: '#00fea8' },
        lineStyle: { width: 2, color: '#00fea8' },
        data: medianData,
      },
      {
        name: $_("Ping.LossRate") || "Loss Rate (%)",
        type: 'bar',
        yAxisIndex: 1,
        barWidth: '40%',
        itemStyle: {
          color: (params: any) => {
            const loss = params.value;
            if (loss > 50) return '#ef4444';
            if (loss > 0) return '#f59e0b';
            return 'rgba(59, 130, 246, 0.3)';
          }
        },
        data: lossData,
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};

