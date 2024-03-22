import * as echarts from "echarts";
import * as ecStat from "echarts-stat";
import { setZoomCallback } from "./utils.js";
import { _, unwrapFunctionStore } from "svelte-i18n";
const $_ = unwrapFunctionStore(_);

let chart: any;

const makeLogCountChart = (div: string) => {
  chart = echarts.init(document.getElementById(div), "dark");
  const option = {
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
      left: "5%",
      right: "5%",
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: "time",
      name: $_("Ts.DateTime"),
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter: (value: any, index: any) => {
          const date = new Date(value);
          return echarts.time.format(date, "{yyyy}/{MM}/{dd} {HH}:{mm}", false);
        },
      },
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
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
      name: $_("Ts.NumberOfLog"),
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        type: "bar",
        name: $_("Ts.NumberOfLog"),
        color: "#1f78b4",
        large: true,
        data: [],
      },
    ],
  };
  chart.setOption(option);
  chart.resize();
};

const addChartData = (
  data: any,
  count: number,
  ctm: number,
  newCtm: number
) => {
  let t = new Date(ctm * 60 * 1000);
  data.push({
    name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}", false),
    value: [t, count],
  });
  ctm++;
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000);
    data.push({
      name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}", false),
      value: [t, 0],
    });
  }
  return ctm;
};

export const showLogCountChart = (
  div: string,
  logs: any,
  zoomCallback: any
) => {
  if (chart) {
    chart.dispose();
  }
  makeLogCountChart(div);
  if (logs.length < 1) {
    return;
  }
  const data: any = [];
  let count = 0;
  let ctm: number = 0;
  let st = Infinity;
  let lt = 0;
  logs.sort((a: any, b: any) => a.Time - b.Time);
  logs.forEach((e: any) => {
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60));
    if (!ctm) {
      ctm = newCtm;
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, count, ctm, newCtm);
      count = 0;
    }
    count++;
    if (st > e.Time) {
      st = e.Time;
    }
    if (lt < e.Time) {
      lt = e.Time;
    }
  });
  addChartData(data, count, ctm, ctm + 1);
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  });
  chart.resize();
  setZoomCallback(chart, zoomCallback, st, lt);
};

const makeMagicTimeChart = (div: string, ent: string) => {
  chart = echarts.init(document.getElementById(div), "dark");
  const option = {
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
      left: "5%",
      right: "5%",
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: "time",
      name: $_("Ts.DateTime"),
      axisLabel: {
        color: "#ccc",
        fontSize: "8px",
        formatter: (value: any, index: any) => {
          const date = new Date(value);
          return echarts.time.format(date, "{yyyy}/{MM}/{dd} {HH}:{mm}", false);
        },
      },
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
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
      name: ent,
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        type: "line",
        name: ent,
        color: "#1f78b4",
        large: true,
        data: [],
      },
    ],
  };
  chart.setOption(option);
  chart.resize();
};

export const showMagicTimeChart = (div: string, logs: any, ent: string) => {
  if (chart) {
    chart.dispose();
  }
  makeMagicTimeChart(div, ent);
  if (ent == "" || logs.length < 1) {
    return;
  }
  const data: any = [];
  logs.sort((a: any, b: any) => a.Time - b.Time);
  logs.forEach((e: any) => {
    const t = new Date(e.Time / (1000 * 1000));
    data.push({
      name: t.toString(),
      value: [t, e[ent]],
    });
  });
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  });
  chart.resize();
};

const getMagicHourChartOption = (ent: string) => {
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
        formatter(value: any) {
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
        name: ent,
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

const setHourChartData = (series: any, t: any, values: any) => {
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

export const showMagicHourChart = (div: string, logs: any, ent: string) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  const option: any = getMagicHourChartOption(ent);
  chart.setOption(option);
  if (ent === "" || logs.length < 1) {
    chart.resize();
    return;
  }
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
  const values: any = [];
  const dt = 3600 * 1000;
  logs.forEach((l: any) => {
    const t = new Date(l.Time / (1000 * 1000));
    const tC = Math.floor(t.getTime() / dt);
    if (tS !== tC) {
      if (tS > 0) {
        if (values.length > 0) {
          tS++;
          setHourChartData(option.series, new Date(tS * dt), values);
          values.length = 0;
          while (tS < tC) {
            tS++;
            setHourChartData(option.series, new Date(tS * dt), [0, 0, 0, 0]);
          }
        }
      }
      tS = tC;
    }
    const numVal = l[ent] ? l[ent] * 1.0 : 0.0;
    values.push(numVal);
  });
  if (values.length > 0) {
    tS++;
    setHourChartData(option.series, new Date(tS * dt), values);
  }
  chart.setOption(option);
  chart.resize();
};

export const showMagicSumChart = (div: string, logs: any, ent: string) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        // Use axis to trigger tooltip
        type: "shadow", // 'shadow' as default; can also be 'line' or 'shadow'
      },
    },
    legend: {},
    grid: {
      left: "3%",
      right: "4%",
      bottom: "3%",
      containLabel: true,
    },
    xAxis: {
      type: "value",
    },
    yAxis: {
      type: "category",
      data: [ent],
    },
    series: [] as any,
  };
  if (ent === "" || logs.length < 1) {
    chart.setOption(option);
    chart.resize();
    return;
  }
  const countMap = new Map();

  logs.forEach((e: any) => {
    const c = e[ent];
    if (c) {
      if (countMap.has(c)) {
        countMap.set(c, countMap.get(c) + 1);
      } else {
        countMap.set(c, 1);
      }
    }
  });
  const list = [...countMap].sort((a, b) => a[1] - b[1]);
  for (let i = 0; i < 10 && i < list.length; i++) {
    option.series.push({
      name: list[i][0],
      type: "bar",
      stack: "total",
      data: [list[i][1]],
    });
  }
  console.log(option);
  chart.setOption(option);
  chart.resize();
};

export const showMagicGraphChart = (
  div: string,
  logs: any,
  ent1: string,
  ent2: string
) => {
  if (chart) {
    chart.dispose();
  }
  const nodeMap = new Map();
  const edgeMap = new Map();
  logs.forEach((l: any) => {
    let nk1 = l[ent1];
    let nk2 = l[ent2];
    if (!nk1 || !nk2) {
      return;
    }
    let ek = nk1 + "|" + nk2;
    let e = edgeMap.get(ek);
    if (!e) {
      edgeMap.set(ek, {
        source: nk1,
        target: nk2,
        lineStyle: {
          width: 2,
        },
      });
    }
    let n = nodeMap.get(nk1);
    if (!n) {
      nodeMap.set(nk1, {
        name: nk1,
        draggable: true,
        category: 0,
        symbolSize: 8,
        label: { show: true },
      });
    }
    n = nodeMap.get(nk2);
    if (!n) {
      nodeMap.set(nk2, {
        name: nk2,
        draggable: true,
        category: 1,
        symbolSize: 8,
        label: { show: true },
      });
    }
  });
  const nodes = Array.from(nodeMap.values());
  const edges = Array.from(edgeMap.values());
  const categories = [{ name: ent1 }, { name: ent2 }];
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  const options: any = {
    title: {
      show: false,
    },
    grid: {
      left: "7%",
      right: "4%",
      bottom: "3%",
      containLabel: true,
    },
    tooltip: {
      trigger: "item",
      textStyle: {
        fontSize: 8,
      },
      position: "bottom",
    },
    legend: [
      {
        orient: "vertical",
        top: 50,
        right: 20,
        textStyle: {
          fontSize: 10,
          color: "#ccc",
        },
        data: categories.map(function (a) {
          return a.name;
        }),
      },
    ],
    animationDurationUpdate: 1500,
    animationEasingUpdate: "quinticInOut",
    series: [
      {
        name: "IP Flows",
        type: "graph",
        layout: "force",
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: "right",
          formatter: "{b}",
          fontSize: 8,
          fontStyle: "normal",
          color: "#ccc",
        },
        color: ["#e31a1c", "#1f78b4"],
        lineStyle: {
          color: "#2122e9",
        },
      },
    ],
  };
  chart.setOption(options);
  chart.resize();
};

export const resizeLogCountChart = () => {
  if (chart) {
    chart.resize();
  }
};
