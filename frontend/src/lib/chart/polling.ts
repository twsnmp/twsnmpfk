import * as echarts from "echarts";
import * as ecStat from "echarts-stat";

const vmapUsage = [
  {
    gt: 0,
    lte: 90,
    color: "#0062f7",
  },
  {
    gt: 90,
    lte: 100,
    color: "#FBFB0F",
  },
  {
    gt: 100,
    color: "#FF2000",
  },
];

const chartParams = {
  rtt: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: "応答時間(秒)",
  },
  rtt_cv: {
    mul: 1.0,
    axis: "応答時間変動係数",
  },
  successRate: {
    mul: 100.0,
    axis: "成功率(%)",
  },
  speed: {
    mul: 1.0,
    axis: "回線速度(Mbps)",
  },
  speed_cv: {
    mul: 1.0,
    axis: "回線速度変動係数",
  },
  feels_like: {
    mul: 1.0,
    axis: "体感温度(℃）",
  },
  humidity: {
    mul: 1.0,
    axis: "湿度(%)",
  },
  pressure: {
    mul: 1.0,
    axis: "気圧(hPa)",
  },
  temp: {
    mul: 1.0,
    axis: "温度(℃）",
  },
  temp_max: {
    mul: 1.0,
    axis: "最高温度(℃）",
  },
  temp_min: {
    mul: 1.0,
    axis: "最低温度(℃）",
  },
  wind: {
    mul: 1.0,
    axis: "風速(m/sec)",
  },
  offset: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: "時刻差(秒)",
  },
  stratum: {
    mul: 1,
    axis: "階層",
  },
  load1m: {
    mul: 1.0,
    axis: "１分間負荷",
  },
  load5m: {
    mul: 1.0,
    axis: "５分間負荷",
  },
  load15m: {
    mul: 1.0,
    axis: "１５分間負荷",
  },
  up: {
    mul: 1.0,
    axis: "稼働数",
  },
  total: {
    mul: 1.0,
    axis: "総数",
  },
  rate: {
    mul: 1.0,
    axis: "稼働率",
  },
  capacity: {
    mul: 0.000000001,
    axis: "総容量(GB)",
  },
  freeSpace: {
    mul: 0.000000001,
    axis: "空き容量(GB)",
  },
  usage: {
    mul: 1.0,
    axis: "使用率(%)",
    vmap: vmapUsage,
  },
  totalCPU: {
    mul: 0.001,
    axis: "総CPUクロック(GHz)",
  },
  usedCPU: {
    mul: 0.001,
    axis: "使用中のCPUクロック(GHz)",
  },
  usageCPU: {
    mul: 1.0,
    axis: "CPU使用率(%)",
  },
  totalMEM: {
    mul: 0.000001,
    axis: "総メモリー(MB)",
  },
  usedMEM: {
    mul: 0.000001,
    axis: "使用中メモリー(MB)",
  },
  usageMEM: {
    mul: 1.0,
    axis: "メモリー使用率(%)",
  },
  totalHost: {
    mul: 1.0,
    axis: "ホスト数",
  },
  fail: {
    mul: 1.0,
    axis: "失敗回数",
  },
  count: {
    mul: 1.0,
    axis: "回数",
  },
};

export const getChartParams = (ent) => {
  const r = chartParams[ent];
  if (r) {
    return r;
  }
  return {
    mul: 1.0,
    axis: ent,
  };
};

let chart;

const getPollingChartOption = (div) => {
  return {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: "#4b5769",
      },
      {
        offset: 1,
        color: "#404a59",
      },
    ]),
    toolbox: {
      iconStyle: {
        color: "#ccc",
      },
      feature: {
        dataZoom: {},
        saveAsImage: { name: "twsnmp_" + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow",
      },
    },
    grid: {
      left: "10%",
      right: "10%",
      top: 60,
      buttom: 0,
    },
    legend: {
      data: [""],
      textStyle: {
        color: "#ccc",
        fontSize: 10,
      },
    },
    xAxis: {
      type: "time",
      name: "日時",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter(value, index) {
          const date = new Date(value);
          return echarts.time.format(date, "{yyyy}/{MM}/{dd} {HH}:{mm}", false);
        },
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: "value",
        nameTextStyle: {
          color: "#ccc",
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          color: "#ccc",
          fontSize: 8,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: "#ccc",
          },
        },
      },
    ],
    series: [
      {
        color: "#1f78b4",
        type: "line",
        showSymbol: false,
        data: [],
      },
    ],
  };
};

const setChartData = (series, t, values) => {
  const data = [t.getTime() * 1000 * 1000];
  const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}", false);
  const mean = ecStat.statistics.mean(values);
  series[0].data.push({
    name,
    value: [t, mean],
  });
  data.push(mean);
  const max = ecStat.statistics.max(values);
  series[1].data.push({
    name,
    value: [t, max],
  });
  data.push(max);
  const min = ecStat.statistics.min(values);
  series[2].data.push({
    name,
    value: [t, min],
  });
  data.push(min);
  const median = ecStat.statistics.median(values);
  series[3].data.push({
    name,
    value: [t, median],
  });
  data.push(median);
  const variance = ecStat.statistics.sampleVariance(values);
  series[4].data.push({
    name,
    value: [t, variance],
  });
  data.push(variance);
};

export const showPollingChart = (div, logs, ent) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div));

  const option: any = getPollingChartOption(div);

  chart.setOption(option);
  if (ent === "") {
    chart.resize();
    return;
  }
  const dp = getChartParams(ent);
  option.series[0].name = "平均値";
  option.series.push({
    name: "最大値",
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: "最小値",
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: "中央値",
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: "分散",
    type: "line",
    large: true,
    yAxisIndex: 1,
    data: [],
  });
  option.yAxis.push({
    type: "value",
    name: "分散",
    nameTextStyle: {
      color: "#ccc",
      fontSize: 10,
      margin: 2,
    },
    axisLabel: {
      color: "#ccc",
      fontSize: 8,
      margin: 2,
    },
    axisLine: {
      lineStyle: {
        color: "#ccc",
      },
    },
  });
  option.legend.data[0] = "平均値";
  option.legend.data.push("最大値");
  option.legend.data.push("最小値");
  option.legend.data.push("中央値");
  option.legend.data.push("分散");
  let tS = -1;
  const values = [];
  const dt = 3600 * 1000;
  logs.forEach((l) => {
    const t = new Date(l.Time / (1000 * 1000));
    const tC = Math.floor(t.getTime() / dt);
    if (tS !== tC) {
      if (tS > 0) {
        if (values.length > 0) {
          tS++;
          setChartData(option.series, new Date(tS * dt), values);
          values.length = 0;
          while (tS < tC) {
            tS++;
            setChartData(option.series, new Date(tS * dt), [0, 0, 0, 0]);
          }
        }
      }
      tS = tC;
    }
    let numVal = getNumVal(ent, l.Result);
    numVal *= dp.mul;
    values.push(numVal || 0.0);
  });
  if (values.length > 0) {
    tS++;
    setChartData(option.series, new Date(tS * dt), values);
  }
  option.yAxis.name = dp.axis;
  chart.setOption(option);
  chart.resize();
};

const makePollingHistogram = (div) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div));
  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: "#4b5769",
      },
      {
        offset: 1,
        color: "#404a59",
      },
    ]),
    toolbox: {
      iconStyle: {
        color: "#ccc",
      },
      feature: {
        dataZoom: {},
        saveAsImage: { name: "twsnmp_" + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: "axis",
      formatter(params) {
        const p = params[0];
        return p.value[0] + "の回数:" + p.value[1];
      },
      axisPointer: {
        type: "shadow",
      },
    },
    grid: {
      left: "10%",
      right: "10%",
      top: 30,
      buttom: 0,
    },
    xAxis: {
      scale: true,
      min: 0,
    },
    yAxis: {
      name: "回数",
    },
    series: [
      {
        color: "#1f78b4",
        type: "bar",
        showSymbol: false,
        barWidth: "99.3%",
        data: [],
      },
    ],
  };
  chart.setOption(option);
  chart.resize();
};

export const showPollingHistogram = (div, logs, ent) => {
  makePollingHistogram(div);
  if (ent === "") {
    return;
  }
  const data = [];
  const dp = getChartParams(ent);
  logs.forEach((l) => {
    if (!l.Result.error) {
      let numVal = getNumVal(ent, l.Result);
      numVal *= dp.mul;
      data.push(numVal);
    }
  });
  const bins = ecStat.histogram(data, "squareRoot");
  chart.setOption({
    xAxis: {
      name: dp.axis,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  });
  chart.resize();
};

const getChartModeName = (mode) => {
  const r = getChartParams(mode);
  if (r && r.axis) {
    return r.axis;
  }
  return mode;
};


const getNumVal = (key, r) => {
  return r[key] || 0.0;
};

const makeSTLChart = (div) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div));
  const option = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: "#4b5769",
      },
      {
        offset: 1,
        color: "#404a59",
      },
    ]),
    toolbox: {
      iconStyle: {
        color: "#ccc",
      },
      feature: {
        dataZoom: {},
        saveAsImage: { name: "twsnmp_" + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "cross",
        label: {
          backgroundColor: "#6a7985",
        },
      },
    },
    legend: {
      data: ["Seasonal", "Trend", "Resid"],
      textStyle: {
        fontSize: 10,
        color: "#ccc",
      },
    },
    grid: {
      left: "10%",
      right: "10%",
      top: "10%",
      buttom: "10%",
    },
    xAxis: {
      type: "time",
      name: "日時",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter(value, index) {
          const date = new Date(value);
          return echarts.time.format(date, "{yyyy}/{MM}/{dd} {HH}:{mm}", false);
        },
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: "value",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
    },
    series: [
      {
        name: "Resid",
        type: "line",
        stack: "stl",
        color: "#fac858",
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
        showSymbol: false,
        data: [],
      },
      {
        name: "Trend",
        type: "line",
        stack: "stl",
        color: "#91cc75",
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
        showSymbol: false,
        data: [],
      },
      {
        name: "Seasonal",
        type: "line",
        color: "#5470c6",
        stack: "stl",
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
        showSymbol: false,
        data: [],
      },
    ],
  };
  chart.setOption(option);
  chart.resize();
};

