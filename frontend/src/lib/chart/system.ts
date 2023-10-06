import * as echarts from 'echarts';
import * as ecStat from 'echarts-stat';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let resChart;

export const showMonitorResChart = (div, monitor) => {
  if (resChart) {
    resChart.dispose()
  }
  resChart = echarts.init(document.getElementById(div),"dark")
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
      formatter: (params) => {
        return (
          params[0].name +
          '<br>' +
          params[0].seriesName +
          ':' +
          params[0].value[1].toFixed(2) +
          '<br>' +
          params[1].seriesName +
          ':' +
          params[1].value[1].toFixed(2) +
          '<br>' +
          params[2].seriesName +
          ':' +
          params[3].value[1].toFixed(2) +
          '<br>' +
          params[3].seriesName +
          ':' +
          params[3].value[1].toFixed(2)
        )
      },
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
      data: ['CPU', 'Mem', 'Disk','Load'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
    yAxis: [{
      type: 'value',
      name: '%',
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
    },{
      type: 'value',
      name: 'Load',
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
        name: 'CPU',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Mem',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Disk',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Load',
        type: 'bar',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.Time / (1000 * 1000) )
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
    option.series[0].data.push({
      name,
      value: [t, m.CPU],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Mem],
    })
    option.series[2].data.push({
      name,
      value: [t, m.Disk],
    })
    option.series[3].data.push({
      name,
      value: [t, m.Load],
    })
  })
  resChart.setOption(option)
  resChart.resize()
}

let netChart;

export const showMonitorNetChart = (div, monitor) => {
  if (netChart) {
    netChart.dispose()
  }
  netChart = echarts.init(document.getElementById(div),"dark")
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
      formatter: (params) => {
        return (
          params[0].name +
          '<br>' +
          params[0].seriesName +
          ':' +
          params[0].value[1].toFixed(2) +
          '<br>' +
          params[1].seriesName +
          ':' +
          params[1].value[1].toFixed(2)
        )
      },
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
      data: ['Speed', 'Connection'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
    yAxis: [
      {
        type: 'value',
        name: 'bps',
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
      {
        type: 'value',
        name: 'count',
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
        name: 'Speed',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Connection',
        type: 'bar',
        large: true,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.Time /(1000 *1000));
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false);
    option.series[0].data.push({
      name,
      value: [t, m.Net],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Conn],
    })
  })
  netChart.setOption(option)
  netChart.resize()
}

let forecastChart; 

export  const showMonitorForecastChart = (div, monitor) => {
  if (forecastChart) {
    forecastChart.dispose()
  }
  forecastChart = echarts.init(document.getElementById(div),"dark")
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
    legend: {
      data: ['Disk(%)', 'DB Size'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '5%',
      right: '15%',
      top: 60,
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
    yAxis: [
      {
        type: 'value',
        name: 'Disk(%)',
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
      {
        type: 'value',
        name: 'DB Size',
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
    visualMap: {
      top: 50,
      right: 10,
      seriesIndex: 0,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      pieces: [
        {
          gt: 0,
          lte: 90,
          color: '#0062f7',
        },
        {
          gt: 90,
          lte: 100,
          color: '#FBFB0F',
        },
        {
          gt: 100,
          color: '#FF2000',
        },
      ],
      outOfRange: {
        color: '#eee',
      },
    },
    series: [
      {
        name: 'Disk(%)',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'DB Size',
        large: true,
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  if (monitor) {
    const dataDisk = [];
    const dataDB = [];
    monitor.forEach((m) => {
      dataDisk.push([m.Time /(1000 * 1000), m.Disk]);
      dataDB.push([m.Time /(1000 * 1000), m.DBSize]);
    });
    const regDisk : any = ecStat.regression('linear', dataDisk,0);
    const regDB : any = ecStat.regression('linear', dataDB,0);
    const sd = Math.floor(Date.now() / (24 * 3600 * 1000))
    for (let d = sd; d < sd + 365; d++) {
      const x = d * 24 * 3600 * 1000
      const t = new Date(x);
      const yDisk = regDisk.parameter.intercept + regDisk.parameter.gradient * x;
      option.series[0].data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
        value: [t, Math.max(0,yDisk)],
      });
      const yDB = regDB.parameter.intercept + regDB.parameter.gradient * x;
      option.series[1].data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false),
        value: [t, Math.max(0,yDB)],
      })
    }
  }
  forecastChart.setOption(option);
  forecastChart.resize();
}


export const resizeMonitorChart = (f) => {
  if(f && forecastChart) {
    forecastChart.resize();
  } else {
    if(resChart) {
      resChart.resize();
    }
    if(netChart) {
      netChart.resize();
    }
  }
}