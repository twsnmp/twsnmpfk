import * as echarts from "echarts";
import "echarts-gl";
import { _, unwrapFunctionStore } from "svelte-i18n";
const $_ = unwrapFunctionStore(_);

let chart: any;

const disposeChart = () => {
  if (chart) {
    chart.dispose();
    chart = undefined;
  }
};

// 1. クライアントID別 (By Client ID)
export const showMqttClientIDChart = (div: string, stats: any) => {
  disposeChart();
  const map = new Map<string, { count: number; bytes: number; topics: Set<string> }>();
  if (stats) {
    stats.forEach((s: any) => {
      const cid = s.ClientID || "(Unknown)";
      const item = map.get(cid) || { count: 0, bytes: 0, topics: new Set<string>() };
      item.count += s.Count || 0;
      item.bytes += s.Bytes || 0;
      if (s.Topic) item.topics.add(s.Topic);
      map.set(cid, item);
    });
  }

  const list = Array.from(map.entries()).map(([cid, val]) => ({
    cid,
    count: val.count,
    bytes: val.bytes,
    topicCount: val.topics.size,
  }));
  list.sort((a, b) => b.count - a.count);

  const categories: string[] = [];
  const counts: number[] = [];
  const bytesMB: number[] = [];
  const topList = list.slice(0, 30).reverse();

  topList.forEach((item) => {
    categories.push(item.cid);
    counts.push(item.count);
    bytesMB.push(Number((item.bytes / (1024 * 1024)).toFixed(2)));
  });

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.CountByClientID"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    legend: {
      top: 30,
      textStyle: { color: "#ccc" },
      data: [$_("Mqtt.Count"), $_("Mqtt.Bytes") + " (MB)"],
    },
    grid: {
      left: "15%",
      right: "10%",
      top: 70,
      bottom: 40,
      containLabel: true,
    },
    xAxis: [
      {
        type: "value",
        name: $_("Mqtt.Count"),
        axisLabel: { color: "#ccc" },
      },
      {
        type: "value",
        name: "MB",
        axisLabel: { color: "#ccc" },
      },
    ],
    yAxis: {
      type: "category",
      data: categories,
      axisLabel: { color: "#ccc", fontSize: 10 },
    },
    series: [
      {
        name: $_("Mqtt.Count"),
        type: "bar",
        data: counts,
        itemStyle: { color: "#36a2eb" },
      },
      {
        name: $_("Mqtt.Bytes") + " (MB)",
        type: "bar",
        xAxisIndex: 1,
        data: bytesMB,
        itemStyle: { color: "#ff6384" },
      },
    ],
  });
  return chart;
};

// 2. 送信元別 (By Source / Remote IP)
export const showMqttRemoteChart = (div: string, stats: any) => {
  disposeChart();
  const map = new Map<string, { count: number; bytes: number }>();
  if (stats) {
    stats.forEach((s: any) => {
      const ip = s.Remote || "(Unknown)";
      const item = map.get(ip) || { count: 0, bytes: 0 };
      item.count += s.Count || 0;
      item.bytes += s.Bytes || 0;
      map.set(ip, item);
    });
  }

  const data = Array.from(map.entries()).map(([ip, val]) => ({
    name: ip,
    value: val.count,
    bytes: val.bytes,
  }));
  data.sort((a, b) => b.value - a.value);

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.CountByRemote"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      trigger: "item",
      formatter: (params: any) => {
        const bytesMB = (params.data.bytes / (1024 * 1024)).toFixed(2);
        return `${params.name}<br/>${$_("Mqtt.Count")}: ${params.value.toLocaleString()}<br/>${$_("Mqtt.Bytes")}: ${bytesMB} MB (${params.percent}%)`;
      },
    },
    legend: {
      type: "scroll",
      orient: "vertical",
      right: 10,
      top: 40,
      bottom: 20,
      textStyle: { color: "#ccc" },
    },
    series: [
      {
        name: $_("MqttReport.CountByRemote"),
        type: "pie",
        radius: ["30%", "70%"],
        center: ["40%", "55%"],
        data: data.slice(0, 30),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: "rgba(0, 0, 0, 0.5)",
          },
        },
      },
    ],
  });
  return chart;
};

// 3. トピック別 (By Topic)
export const showMqttTopicChart = (div: string, stats: any) => {
  disposeChart();
  const map = new Map<string, { count: number; bytes: number }>();
  if (stats) {
    stats.forEach((s: any) => {
      const topic = s.Topic || "(Unknown)";
      const item = map.get(topic) || { count: 0, bytes: 0 };
      item.count += s.Count || 0;
      item.bytes += s.Bytes || 0;
      map.set(topic, item);
    });
  }

  const list = Array.from(map.entries()).map(([topic, val]) => ({
    topic,
    count: val.count,
    bytes: val.bytes,
  }));
  list.sort((a, b) => b.count - a.count);

  const categories: string[] = [];
  const counts: number[] = [];
  const topList = list.slice(0, 25).reverse();

  topList.forEach((item) => {
    categories.push(item.topic);
    counts.push(item.count);
  });

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.CountByTopic"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    grid: {
      left: "20%",
      right: "10%",
      top: 50,
      bottom: 30,
      containLabel: true,
    },
    xAxis: {
      type: "value",
      name: $_("Mqtt.Count"),
      axisLabel: { color: "#ccc" },
    },
    yAxis: {
      type: "category",
      data: categories,
      axisLabel: {
        color: "#ccc",
        fontSize: 10,
        formatter: (val: string) => {
          if (val.length > 40) return val.substring(0, 37) + "...";
          return val;
        },
      },
    },
    series: [
      {
        name: $_("Mqtt.Count"),
        type: "bar",
        data: counts,
        itemStyle: { color: "#4bc0c0" },
      },
    ],
  });
  return chart;
};

// 4. ヒートマップ (Heatmap - Time activity or Client x Topic matrix)
export const showMqttHeatmap = (div: string, stats: any, mode: string = "time") => {
  disposeChart();
  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");

  if (mode === "client_topic") {
    // Client ID x Topic Matrix
    const clientSet = new Set<string>();
    const topicSet = new Set<string>();
    const matrixMap = new Map<string, number>();

    if (stats) {
      stats.forEach((s: any) => {
        const cid = s.ClientID || "(Unknown)";
        const topic = s.Topic || "(Unknown)";
        clientSet.add(cid);
        topicSet.add(topic);
        const key = `${cid}\t${topic}`;
        matrixMap.set(key, (matrixMap.get(key) || 0) + (s.Count || 0));
      });
    }

    const clients = Array.from(clientSet).slice(0, 30);
    const topics = Array.from(topicSet).slice(0, 30);

    const heatData: [number, number, number][] = [];
    let maxVal = 0;

    clients.forEach((c, cIdx) => {
      topics.forEach((t, tIdx) => {
        const val = matrixMap.get(`${c}\t${t}`) || 0;
        if (val > 0) {
          heatData.push([tIdx, cIdx, val]);
          if (val > maxVal) maxVal = val;
        }
      });
    });

    chart.setOption({
      title: {
        text: $_("MqttReport.HeatmapClientTopic"),
        left: "center",
        textStyle: { fontSize: 14, color: "#ccc" },
      },
      tooltip: {
        position: "top",
        formatter: (params: any) => {
          const tName = topics[params.data[0]];
          const cName = clients[params.data[1]];
          return `Client: ${cName}<br/>Topic: ${tName}<br/>Count: ${params.data[2].toLocaleString()}`;
        },
      },
      grid: {
        left: "15%",
        right: "10%",
        top: 60,
        bottom: 80,
        containLabel: true,
      },
      xAxis: {
        type: "category",
        data: topics,
        axisLabel: {
          color: "#ccc",
          fontSize: 9,
          rotate: 30,
          formatter: (val: string) => (val.length > 20 ? val.substring(0, 17) + "..." : val),
        },
      },
      yAxis: {
        type: "category",
        data: clients,
        axisLabel: { color: "#ccc", fontSize: 9 },
      },
      visualMap: {
        min: 0,
        max: maxVal || 1,
        calculable: true,
        orient: "horizontal",
        left: "center",
        bottom: 10,
        textStyle: { color: "#ccc" },
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
          ],
        },
      },
      series: [
        {
          name: "Count",
          type: "heatmap",
          data: heatData,
          emphasis: {
            itemStyle: {
              borderColor: "#fff",
              borderWidth: 1,
            },
          },
        },
      ],
    });
    return chart;
  }

  // Default: Time Activity Heatmap (Date x Hour)
  const hours = Array.from({ length: 24 }, (_, i) => `${i}:00`);
  const dateMap = new Map<string, number[]>(); // key: yyyy/MM/dd, value: number[24]

  if (stats) {
    stats.forEach((s: any) => {
      const firstMs = s.First ? Math.floor(s.First / 1e6) : 0;
      const lastMs = s.Last ? Math.floor(s.Last / 1e6) : firstMs;
      const count = s.Count || 1;

      if (!firstMs) return;

      const spanHours = Math.max(1, Math.ceil((lastMs - firstMs) / (3600 * 1000)));
      const countPerHour = count / spanHours;

      // Distribute messages over active hours
      let currentMs = firstMs;
      while (currentMs <= lastMs || currentMs === firstMs) {
        const d = new Date(currentMs);
        const dateStr = `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, "0")}/${String(d.getDate()).padStart(2, "0")}`;
        const hour = d.getHours();

        if (!dateMap.has(dateStr)) {
          dateMap.set(dateStr, new Array(24).fill(0));
        }
        dateMap.get(dateStr)![hour] += countPerHour;

        if (spanHours <= 1 || currentMs + 3600 * 1000 > lastMs + 3600 * 1000) {
          break;
        }
        currentMs += 3600 * 1000;
      }
    });
  }

  const sortedDates = Array.from(dateMap.keys()).sort();
  const heatData: [number, number, number][] = [];
  let maxCount = 0;

  sortedDates.forEach((dateStr, dIdx) => {
    const hourCounts = dateMap.get(dateStr)!;
    hourCounts.forEach((val, hIdx) => {
      if (val > 0) {
        const rounded = Number(val.toFixed(1));
        heatData.push([dIdx, hIdx, rounded]);
        if (rounded > maxCount) maxCount = rounded;
      }
    });
  });

  chart.setOption({
    title: {
      text: $_("MqttReport.HeatmapTime"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      position: "top",
      formatter: (params: any) => {
        const dStr = sortedDates[params.data[0]];
        const hStr = hours[params.data[1]];
        return `${dStr} ${hStr}<br/>Est. Count: ${params.data[2]}`;
      },
    },
    grid: {
      left: "10%",
      right: "5%",
      top: 60,
      bottom: 80,
    },
    xAxis: {
      type: "category",
      data: sortedDates,
      axisLabel: { color: "#ccc", fontSize: 10 },
    },
    yAxis: {
      type: "category",
      data: hours,
      axisLabel: { color: "#ccc", fontSize: 10 },
    },
    visualMap: {
      min: 0,
      max: maxCount || 1,
      calculable: true,
      orient: "horizontal",
      left: "center",
      bottom: 10,
      textStyle: { color: "#ccc" },
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
        ],
      },
    },
    series: [
      {
        name: "Activity",
        type: "heatmap",
        data: heatData,
        emphasis: {
          itemStyle: {
            borderColor: "#fff",
            borderWidth: 1,
          },
        },
      },
    ],
  });
  return chart;
};

// 5. 状態別 (Count by State)
export const showMqttStateChart = (div: string, stats: any) => {
  disposeChart();
  const map = new Map<string, number>();
  if (stats) {
    stats.forEach((s: any) => {
      const state = s.State || "normal";
      map.set(state, (map.get(state) || 0) + 1);
    });
  }

  const data = Array.from(map.entries()).map(([state, count]) => ({
    name: state,
    value: count,
  }));

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.CountByState"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      trigger: "item",
      formatter: "{b}: {c} ({d}%)",
    },
    legend: {
      orient: "vertical",
      right: 20,
      top: "center",
      textStyle: { color: "#ccc" },
    },
    series: [
      {
        name: $_("MqttReport.CountByState"),
        type: "pie",
        radius: ["40%", "70%"],
        data: data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: "rgba(0, 0, 0, 0.5)",
          },
        },
      },
    ],
  });
  return chart;
};

// 6. トピック階層ツリーマップ (Topic Hierarchy Treemap)
export const showMqttTopicTreemap = (div: string, stats: any) => {
  disposeChart();
  interface TreeNode {
    name: string;
    value?: number;
    children?: TreeNode[];
    [key: string]: any;
  }

  const rootChildren: Map<string, TreeNode> = new Map();

  if (stats) {
    stats.forEach((s: any) => {
      const topic = s.Topic || "unknown";
      const parts = topic.split("/").filter((p: string) => p.length > 0);
      const count = s.Count || 1;

      let currentChildren = rootChildren;

      parts.forEach((part: string, idx: number) => {
        let node = currentChildren.get(part);
        if (!node) {
          node = { name: part };
          if (idx < parts.length - 1) {
            node.children = [];
          }
          currentChildren.set(part, node);
        }
        if (idx === parts.length - 1) {
          node.value = (node.value || 0) + count;
        } else {
          if (!node.childrenMap) {
            node.childrenMap = new Map<string, TreeNode>();
          }
          currentChildren = node.childrenMap;
        }
      });
    });
  }

  const buildTree = (map: Map<string, TreeNode>): TreeNode[] => {
    const res: TreeNode[] = [];
    map.forEach((node) => {
      if (node.childrenMap) {
        node.children = buildTree(node.childrenMap);
        delete node.childrenMap;
      }
      res.push(node);
    });
    return res;
  };

  const treeData = buildTree(rootChildren);

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.TopicTreemap"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {
      formatter: (info: any) => {
        return `${info.name}<br/>Count: ${info.value ? info.value.toLocaleString() : "-"}`;
      },
    },
    series: [
      {
        type: "treemap",
        data: treeData,
        levels: [
          {
            itemStyle: {
              borderColor: "#555",
              borderWidth: 2,
              gapWidth: 2,
            },
          },
          {
            colorSaturation: [0.3, 0.6],
            itemStyle: {
              borderColor: "#333",
              borderWidth: 1,
              gapWidth: 1,
            },
          },
        ],
      },
    ],
  });
  return chart;
};

// 7. 3D 散布図 (3D Scatter Chart)
export const showMqtt3DChart = (div: string, stats: any) => {
  disposeChart();
  const clientsSet = new Set<string>();
  const topicsSet = new Set<string>();
  const scatterData: [string, string, number][] = [];

  if (stats) {
    stats.forEach((s: any) => {
      const cid = s.ClientID || "(Unknown)";
      const topic = s.Topic || "(Unknown)";
      const count = s.Count || 0;
      clientsSet.add(cid);
      topicsSet.add(topic);
      scatterData.push([cid, topic, count]);
    });
  }

  const clients = Array.from(clientsSet).slice(0, 30);
  const topics = Array.from(topicsSet).slice(0, 30);

  const filteredData = scatterData
    .filter(([c, t]) => clients.includes(c) && topics.includes(t))
    .map(([c, t, val]) => [c, t, val]);

  const maxVal = Math.max(...filteredData.map((d) => d[2]), 1);

  const dom = document.getElementById(div);
  if (!dom) return;
  chart = echarts.init(dom, "dark");
  chart.setOption({
    title: {
      text: $_("MqttReport.Chart3D"),
      left: "center",
      textStyle: { fontSize: 14, color: "#ccc" },
    },
    tooltip: {},
    visualMap: {
      max: maxVal,
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
        ],
      },
    },
    xAxis3D: {
      type: "category",
      name: $_("Mqtt.ClientID"),
      data: clients,
      axisLabel: { color: "#ccc", fontSize: 9 },
    },
    yAxis3D: {
      type: "category",
      name: $_("Mqtt.Topic"),
      data: topics,
      axisLabel: {
        color: "#ccc",
        fontSize: 9,
        formatter: (v: string) => (v.length > 15 ? v.substring(0, 12) + "..." : v),
      },
    },
    zAxis3D: {
      type: "value",
      name: $_("Mqtt.Count"),
      axisLabel: { color: "#ccc", fontSize: 9 },
    },
    grid3D: {
      boxWidth: 200,
      boxDepth: 200,
      viewControl: {
        projection: "perspective",
      },
      light: {
        main: {
          intensity: 1.2,
        },
        ambient: {
          intensity: 0.3,
        },
      },
    },
    series: [
      {
        type: "bar3D",
        data: filteredData.map((item) => ({
          value: [item[0], item[1], item[2]],
        })),
        shading: "lambert",
        label: {
          show: false,
        },
        emphasis: {
          label: {
            show: true,
          },
        },
      },
    ],
  });
  return chart;
};
