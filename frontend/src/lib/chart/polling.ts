import * as echarts from "echarts";
import * as ecStat from "echarts-stat";
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

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

const chartParams :any = {
  rtt: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: $_("Ts.RespTimeSec"),
  },
  rtt_cv: {
    mul: 1.0,
    axis: $_("Ts.RespTimeCV"),
  },
  successRate: {
    mul: 100.0,
    axis: $_("Ts.SuccessRatePer"),
  },
  speed: {
    mul: 1.0,
    axis: $_("Ts.LineSpeedMbps"),
  },
  speed_cv: {
    mul: 1.0,
    axis: $_("Ts.LineSeedCV"),
  },
  feels_like: {
    mul: 1.0,
    axis: $_("Ts.FeelsLikeC"),
  },
  humidity: {
    mul: 1.0,
    axis: $_("Ts.Humidity"),
  },
  pressure: {
    mul: 1.0,
    axis: $_("Ts.PressurehPa"),
  },
  temp: {
    mul: 1.0,
    axis: $_("Ts.TempC"),
  },
  temp_max: {
    mul: 1.0,
    axis: $_("Ts.MaxTempC"),
  },
  temp_min: {
    mul: 1.0,
    axis: $_("Ts.MinTempC"),
  },
  wind: {
    mul: 1.0,
    axis: $_("Ts.WindSpeedMPS"),
  },
  offset: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: $_("Ts.OffsetSec"),
  },
  stratum: {
    mul: 1,
    axis: $_("Ts.Stratum"),
  },
  load1m: {
    mul: 1.0,
    axis: $_("Ts.Load1M"),
  },
  load5m: {
    mul: 1.0,
    axis: $_("Ts.Load5M"),
  },
  load15m: {
    mul: 1.0,
    axis: $_("Ts.Load15M"),
  },
  up: {
    mul: 1.0,
    axis: $_("Ts.UpCount"),
  },
  total: {
    mul: 1.0,
    axis: $_("Ts.Total"),
  },
  rate: {
    mul: 1.0,
    axis: $_("Ts.Oprate"),
  },
  usage: {
    mul: 1.0,
    axis: $_("Ts.UsagePer"),
    vmap: vmapUsage,
  },
  usageCPU: {
    mul: 1.0,
    axis: $_("Ts.CPUUsagePer"),
  },
  usageMEM: {
    mul: 1.0,
    axis: $_("Ts.MemUsagePer"),
  },
  totalHost: {
    mul: 1.0,
    axis: $_("Ts.TotalHost"),
  },
  fail: {
    mul: 1.0,
    axis: $_("Ts.FailCount"),
  },
  count: {
    mul: 1.0,
    axis: $_("Ts.Count"),
  },
};

export const getChartParams = (ent:any) => {
  const r = chartParams[ent];
  if (r) {
    return r;
  }
  return {
    mul: 1.0,
    axis: ent,
  };
};

let chart :any;

const getPollingChartOption = () => {
  return {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: "#ccc",
      },
      feature: {
        dataZoom: {},
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
      name: $_("Ts.DateTime"),
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter(value:any) {
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

const setChartData = (series:any, t:any, values:any) => {
  const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}", false);
  const mean = ecStat.statistics.mean(values);
  series[0].data.push({
    name,
    value: [t, mean],
  });
  const max = ecStat.statistics.max(values);
  series[1].data.push({
    name,
    value: [t, max],
  });
  const min = ecStat.statistics.min(values);
  series[2].data.push({
    name,
    value: [t, min],
  });
  const median = ecStat.statistics.median(values);
  series[3].data.push({
    name,
    value: [t, median],
  });
  const variance = ecStat.statistics.sampleVariance(values);
  series[4].data.push({
    name,
    value: [t, variance],
  });
};

export const showPollingChart = (div:string, logs:any, ent:any) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div),"dark");

  const option: any = getPollingChartOption();

  chart.setOption(option);
  if (ent === "") {
    chart.resize();
    return;
  }
  const dp = getChartParams(ent);
  option.series[0].name = $_("Ts.AvgVal");
  option.series.push({
    name: $_("Ts.MaxVal"),
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: $_("Ts.MinVal"),
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: $_("Ts.Median"),
    type: "line",
    large: true,
    data: [],
  });
  option.series.push({
    name: $_("Ts.Variance"),
    type: "line",
    large: true,
    yAxisIndex: 1,
    data: [],
  });
  option.yAxis.push({
    type: "value",
    name: $_("Ts.Variance"),
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
  option.legend.data[0] = $_("Ts.AvgVal");
  option.legend.data.push($_("Ts.MaxVal"));
  option.legend.data.push($_("Ts.MinVal"));
  option.legend.data.push($_("Ts.Median"));
  option.legend.data.push($_("Ts.Variance"));
  let tS = -1;
  const values :any = [];
  const dt = 3600 * 1000;
  logs.forEach((l:any) => {
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

const makePollingHistogram = (div:string) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option :any = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: "#ccc",
      },
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: "axis",
      formatter(params:any) {
        const p = params[0];
        return p.value[0] + ":" + p.value[1];
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
      name: $_("Ts.Count"),
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

export const showPollingHistogram = (div:string, logs:any, ent:any) => {
  makePollingHistogram(div);
  if (ent === "") {
    return;
  }
  const data :any = [];
  const dp = getChartParams(ent);
  logs.forEach((l:any) => {
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


const getNumVal = (key:any, r:any) => {
  return r[key] || 0.0;
};

