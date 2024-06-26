import * as echarts from 'echarts'

const chartMap = new Map();

export const showHrBarChart = (div:string, type:any, xAxis:any, list:any, max?:any) => {
  if (chartMap.has(div)) {
    chartMap.get(div).dispose();
  }
  const chart = echarts.init(document.getElementById(div),"dark");
  chartMap.set(div,chart);

  const yellow = max ? max * 0.8 : 80
  const red = max ? max * 0.9 : 90
  const option :any = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      top: '10%',
      left: '5%',
      right: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: xAxis,
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
    },
    yAxis: {
      type: 'category',
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      data: [],
    },
    series: [
      {
        name: type,
        type: 'bar',
        itemStyle: {
          color: (p:any) => {
            return p.value > red ? '#c00' : p.value > yellow ? '#cc0' : '#0cc'
          },
        },
        data: [],
      },
    ],
  }
  if (!list) {
    return;
  }
  list.forEach((e:any) => {
    option.yAxis.data.push(e.Name)
    option.series[0].data.push(e.Value)
  })
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showHrSummary = (div:string, data:any) => {
  if (chartMap.has(div)) {
    chartMap.get(div).dispose();
  }
  const chart = echarts.init(document.getElementById(div),"dark");
  chartMap.set(div,chart);

  const gaugeData = [
    {
      value: data.CPU.toFixed(1),
      name: 'CPU',
      title: {
        offsetCenter: ['-50%', '78%'],
      },
      detail: {
        offsetCenter: ['-50%', '95%'],
      },
    },
    {
      value: data.Mem.toFixed(1),
      name: 'Mem',
      title: {
        offsetCenter: ['0%', '78%'],
      },
      detail: {
        offsetCenter: ['0%', '95%'],
      },
    },
    {
      value: data.VM.toFixed(1),
      name: 'VM',
      title: {
        offsetCenter: ['50%', '78%'],
      },
      detail: {
        offsetCenter: ['50%', '95%'],
      },
    },
  ]
  const option = {
    color: ['#4575b4', '#abd9e9', '#FAC858'],
    series: [
      {
        type: 'gauge',
        anchor: {
          show: true,
          showAbove: true,
          size: 18,
          itemStyle: {
            color: '#FAC858',
          },
        },
        pointer: {
          icon: 'path://M2.9,0.7L2.9,0.7c1.4,0,2.6,1.2,2.6,2.6v115c0,1.4-1.2,2.6-2.6,2.6l0,0c-1.4,0-2.6-1.2-2.6-2.6V3.3C0.3,1.9,1.4,0.7,2.9,0.7z',
          width: 8,
          length: '80%',
          offsetCenter: [0, '8%'],
        },
        progress: {
          show: true,
          overlap: true,
          roundCap: true,
        },
        axisLine: {
          roundCap: true,
        },
        axisLabel: {
          color: '#ccc',
        },
        data: gaugeData,
        title: {
          fontSize: 12,
          color: '#ccc',
        },
        detail: {
          width: 40,
          height: 14,
          fontSize: 12,
          color: '#ccc',
          borderRadius: 3,
          formatter: '{value}%',
        },
      },
    ],
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

