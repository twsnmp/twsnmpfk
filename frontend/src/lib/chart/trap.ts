import * as echarts from 'echarts';
import 'echarts-gl';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart;

export const showTrapFromAddr = (div, logs) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option = {
    title: {
      show: false,
    },
    legend: {
      top: 15,
      textStyle: {
        fontSize: 12,
      },
      data: [],
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
      data: [],
      nameTextStyle: {
        color: '#ccc',
        fontSize: 12,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
    },
    series: [],
  };
  const fam = new Map();
  const tm = new Map();
  
  logs.forEach((l) => {
    if (!tm.get(l.TrapType)) {
      tm.set(l.TrapType,true);
    }
    const fa = fam.get(l.FromAddress);
    if (!fa) {
      const m = new Map();
      m.set(l.TrapType,1);
      fam.set(l.FromAddress,{
        Name:l.FromAddress,
        Total: 1,
        TypeMap: m, 
      })
    } else {
      fa.Total++;
      const c = fa.TypeMap.get(l.TrapType);
      fa.TypeMap.set(l.TrapType, c ? c+1 :1);
    }
  });
  const list = Array.from(fam.values());
  const types = Array.from(tm.keys());
  for(let t of types) {
    option.legend.data.push(t);
    option.series.push({
        name: t,
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: [],
    });
  }
  list.sort((a, b) => b.Total - a.Total);
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    option.yAxis.data.push(list[i].Name);
    for(let j = 0; j < types.length;j++) {
      option.series[j].data.push(list[i].TypeMap.get(types[j]) || 0);
    }
  }
  chart.setOption(option);
  chart.resize()
}


export const showTrapLog3D = (div, logs) => {
  const m = new Map();
  const tm = new Map();
  logs.forEach((l) => {
    if(!tm.has(l.TrapType)) {
      tm.set(l.TrapType,true);
    }
    const level = getTrapTypeLevel(l.TrapType)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(l.FromAddress)
    if (!e) {
      m.set(l.FromAddress, {
        Name: l.FromAddress,
        Time: [t],
        Type: [l.TrapType],
        Level: [level],
      })
    } else {
      e.Time.push(t)
      e.Type.push(l.TrapType)
      e.Level.push(level)
    }
  })
  const froms = Array.from(m.keys());
  const types = Array.from(tm.keys());
  const l = Array.from(m.values());
  const data = []
  l.forEach((e) => {
    for (let i = 0; i < e.Time.length && i < 15000; i++) {
      data.push([e.Name, e.Time[i], e.Type[i], e.Level[i]])
    }
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
      show: false,
      min: 0,
      max: 3,
      dimension: 3,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: $_("Ts.Sender"),
      data: froms,
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
        formatter(value, index) {
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
      type: 'category',
      name: 'TRAP Type',
      data: types,
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
        name: $_("Ts.LogCount"),
        type: 'scatter3D',
        symbolSize: 8,
        dimensions: ['From', 'Time', 'Type', 'Level'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const getTrapTypeLevel = (l:string) :number => {
  switch (l) {
    case 'linkDown':
    case 'nsNotifyShutdown':
      return 0;
    case 'coldStart':
      return 2;
  }
  return 3
}



export const showTrapTypeChart = (div:string, logs) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: [],
      textStyle: {
        color: '#ccc',
      },
    },
    series: [
      {
        name: 'TRAP Type',
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          color: '#ccc',
          fontSize: 12,
        },
        data: [],
      },
    ],
  }
  if (logs) {
    let i = 0;
    const typeMap = new Map();
    logs.forEach((l) => {
      const t = typeMap.get(l.TrapType);
      typeMap.set(l.TrapType,t ? t+1: 1);
    });
    typeMap.forEach((v,k)=>{
      option.legend.data.push(k);
      option.series[0].data.push({
        name:k,
        value: v,
      });
    });
  }
  chart.setOption(option)
  chart.resize()
}

