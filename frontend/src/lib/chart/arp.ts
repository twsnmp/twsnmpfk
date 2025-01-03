import * as echarts from "echarts";
import "echarts-gl";
import { _, unwrapFunctionStore } from "svelte-i18n";
const $_ = unwrapFunctionStore(_);

let chart: any;

export const showArpLogIP = (div: string, logs: any) => {
  const list = getArpLogIPList(logs);
  const newLog = [];
  const changeLog = [];
  const ips = [];
  list.sort((a, b) => b.total - a.total);
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    newLog.push(list[i].newLog);
    changeLog.push(list[i].changeLog);
    ips.push(list[i].ip);
  }
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  chart.setOption({
    title: {
      show: false,
    },
    color: ["#1f78b4", "#e31a1c"],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: "#ccc",
      },
      data: [$_("Ts.New"), $_("Ts.Change")],
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow",
      },
    },
    grid: {
      left: "20%",
      right: "10%",
      top: "10%",
      bottom: "10%",
      containLabel: true,
    },
    xAxis: {
      type: "value",
      name: $_("Ts.NumberOfLog"),
    },
    yAxis: {
      type: "category",
      data: ips,
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
        name: $_("Ts.New"),
        type: "bar",
        stack: $_("Ts.NumberOfLog"),
        data: newLog,
      },
      {
        name: $_("Ts.Change"),
        type: "bar",
        stack: $_("Ts.NumberOfLog"),
        data: changeLog,
      },
    ],
  });
  chart.resize();
  return chart;
};

const getArpLogIPList = (logs: any) => {
  const m = new Map();
  logs.forEach((l: any) => {
    const e = m.get(l.IP);
    if (!e) {
      m.set(l.IP, {
        ip: l.IP,
        total: 1,
        newLog: l.State === "New" ? 1 : 0,
        changeLog: l.State === "Change" ? 1 : 0,
      });
    } else {
      e.total += 1;
      e.newLog += l.State === "New" ? 1 : 0;
      e.changeLog += l.State === "Change" ? 1 : 0;
    }
  });
  const r = Array.from(m.values());
  return r;
};

export const showArpLogIP3D = (div: string, logs: any) => {
  const m = new Map();
  logs.forEach((l: any) => {
    const t = new Date(l.Time / (1000 * 1000));
    const e = m.get(l.IP);
    if (!e) {
      m.set(l.IP, {
        ip: l.IP,
        time: [t],
        state: [l.State],
        level: [l.State == "New" ? 0 : 1],
      });
    } else {
      e.time.push(t);
      e.state.push(l.State);
      e.level.push(l.State == "New" ? 0 : 1);
    }
  });
  const cat = Array.from(m.keys());
  const l = Array.from(m.values());
  const data: any = [];
  l.forEach((e) => {
    for (let i = 0; i < e.time.length && i < 15000; i++) {
      data.push([e.ip, e.time[i], e.state[i], e.level[i]]);
    }
  });
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  const options = {
    title: {
      show: false,
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: "quinticInOut",
    visualMap: {
      show: false,
      min: 0,
      max: 1,
      dimension: 3,
      inRange: {
        color: ["#1f78b4", "#e31a1c"],
      },
    },
    xAxis3D: {
      type: "category",
      name: "ip",
      data: cat,
      nameTextStyle: {
        color: "#ccc",
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
    },
    yAxis3D: {
      type: "time",
      name: "time",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
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
    },
    zAxis3D: {
      type: "category",
      name: "satte",
      data: ["New", "Change"],
      nameTextStyle: {
        color: "#ccc",
        fontSize: 12,
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
    grid3D: {
      axisLine: {
        color: "#ccc",
      },
      axisPointer: {
        color: "#ccc",
      },
      viewControl: {
        projection: "orthographic",
      },
    },
    series: [
      {
        name: $_("Ts.LogCountByIP"),
        type: "scatter3D",
        symbolSize: 4,
        dimensions: ["ip", "time", "type"],
        data,
      },
    ],
  };
  chart.setOption(options);
  chart.resize();
  return chart;
};

export const showArpGraph = (
  div: string,
  arp: any,
  type: any,
  changeIP: any,
  changeMAC: any
) => {
  const nodeMap = new Map();
  const edgeMap = new Map();
  arp.forEach((a: any) => {
    let ek = a.IP + "|" + a.MAC;
    let e = edgeMap.get(ek);
    if (!e) {
      edgeMap.set(ek, {
        source: a.IP,
        target: a.MAC,
        lineStyle: {
          width: 2,
        },
      });
    }
    let n = nodeMap.get(a.IP);
    if (!n) {
      nodeMap.set(a.IP, {
        name: a.IP,
        draggable: true,
        category: changeIP.has(a.IP) ? 0 : a.IP.startsWith("169.254.") ? 1 : 2,
        symbolSize: 4,
        label: { show: true },
      });
    }
    n = nodeMap.get(a.MAC);
    if (!n) {
      nodeMap.set(a.MAC, {
        name: a.MAC,
        draggable: true,
        category: changeMAC.has(a.MAC) ? 3 : 4,
        symbolSize: 4,
        label: { show: true },
      });
    }
  });
  const nodes = Array.from(nodeMap.values());
  const edges = Array.from(edgeMap.values());
  const categories = [
    { name: $_("Ts.IPChanged") },
    { name: $_("Ts.IPDup") },
    { name: $_("Ts.IPNormal") },
    { name: $_("Ts.MACChanged") },
    { name: $_("Ts.MACNormal") },
  ];
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
    series: [],
  };
  if (type === "circular") {
    options.series = [
      {
        name: "IP Flows",
        type: "graph",
        layout: "circular",
        circular: {
          rotateLabel: true,
        },
        data: nodes,
        links: edges,
        categories,
        roam: true,
        color: ["#ea0", "#e31a1c", "#1f78b4", "#cc0", "#165ee3"],
        label: {
          position: "right",
          formatter: "{b}",
          fontSize: 8,
          fontStyle: "normal",
          color: "#ccc",
        },
        lineStyle: {
          color: "source",
          curveness: 0.3,
        },
      },
    ];
  } else {
    options.series = [
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
        color: ["#ea0", "#e31a1c", "#1f78b4", "#cc0", "#165ee3"],
        lineStyle: {
          color: "source",
          curveness: 0,
        },
      },
    ];
  }
  chart.setOption(options);
  chart.resize();
  return chart;
};

export const showIPAMHeatmap = (div: string, ipam: any) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), "dark");
  const option: any = {
    title: {
      show: false,
    },
    grid: {
      left: "12%",
      right: "5%",
      top: 30,
      buttom: 0,
    },
    tooltip: {},
    xAxis: {
      type: "category",
      name: "%",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 12,
        margin: 3,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 8,
        margin: 3,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      data: [],
    },
    yAxis: {
      type: "category",
      name: "IP Range",
      nameTextStyle: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: "#ccc",
        },
      },
      data: [],
    },
    visualMap: {
      min: 0,
      max: 0,
      textStyle: {
        color: "#ccc",
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: [
          "#313695",
          "#4575b4",
          "#74add1",
          "#abd9e9",
          "#e0f3f8",
          "#ffffbf",
          "#fee090",
          "#fdae61",
          "#f46d43",
          "#d73027",
          "#a50026",
        ],
      },
    },
    series: [
      {
        name: "IP Usage",
        type: "heatmap",
        data: [],
        emphasis: {
          itemStyle: {
            borderColor: "#ccc",
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  };
  for (let x = 0; x < 100; x++) {
    option.xAxis.data.push(x);
  }
  for (let y = ipam.length - 1; y >= 0; y--) {
    const r = ipam[y];
    option.yAxis.data.push(r.Range);
    for (let x = 0; x < 100; x++) {
      const c = r.UsedIP[x];
      option.series[0].data.push([x, ipam.length - y - 1, c]);
      if (option.visualMap.max < c) {
        option.visualMap.max = c;
      }
    }
  }
  chart.setOption(option);
  chart.resize();
  return chart;
};
