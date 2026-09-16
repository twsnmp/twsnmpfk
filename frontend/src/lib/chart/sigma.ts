import * as echarts from "echarts";

let sigmaChart: echarts.ECharts | null = null;

const getChartElement = (div: string | HTMLElement): HTMLElement | null => {
  if (!div) return null;
  return typeof div === "string" ? document.getElementById(div) : div;
};

export interface SigmaChartStats {
  Critical?: number;
  High?: number;
  Medium?: number;
  Low?: number;
  Informational?: number;
}

export interface SigmaTagItem {
  Tag: string;
  Category: string;
  Count: number;
}

export interface SigmaTimelineItem {
  Time: number;
  Count: number;
  Critical: number;
  High: number;
  Medium: number;
  Low: number;
  Info: number;
}

export const showSigmaSeverityChart = (
  div: string | HTMLElement,
  stats: SigmaChartStats | undefined,
  dark: boolean = false,
  title?: string,
  noThreatText?: string
) => {
  const el = getChartElement(div);
  if (!el) return null;

  if (sigmaChart) {
    sigmaChart.dispose();
  }
  sigmaChart = echarts.init(el, dark ? "dark" : null);

  const data = [
    { name: "Critical", value: stats?.Critical || 0, itemStyle: { color: "#cf222e" } },
    { name: "High", value: stats?.High || 0, itemStyle: { color: "#f85149" } },
    { name: "Medium", value: stats?.Medium || 0, itemStyle: { color: "#d29922" } },
    { name: "Low", value: stats?.Low || 0, itemStyle: { color: "#58a6ff" } },
    { name: "Info", value: stats?.Informational || 0, itemStyle: { color: "#8b949e" } },
  ].filter((d) => d.value > 0);

  const option: echarts.EChartsOption = {
    backgroundColor: "transparent",
    title: {
      text: title || "重大度別検知割合 (Severity)",
      left: "center",
      top: 5,
      textStyle: {
        fontSize: 13,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "item",
      formatter: "{b}: {c} ({d}%)",
    },
    legend: {
      orient: "vertical",
      left: "5%",
      top: "center",
      textStyle: {
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    series: [
      {
        name: "Severity",
        type: "pie",
        radius: ["36%", "66%"],
        center: ["55%", "52%"],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 6,
          borderColor: dark ? "#161b22" : "#fff",
          borderWidth: 2,
        },
        label: {
          show: true,
          formatter: "{b}: {c} ({d}%)",
          color: dark ? "#c9d1d9" : "#24292f",
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: "bold",
          },
        },
        data: data.length > 0 ? data : [{ name: noThreatText || "検知なし (No Threat)", value: 0 }],
      },
    ],
  };

  sigmaChart.setOption(option, true);
  return sigmaChart;
};

export const showSigmaTagsChart = (
  div: string | HTMLElement,
  list: SigmaTagItem[] | undefined,
  dark: boolean = false,
  title?: string,
  noDataText?: string
) => {
  const el = getChartElement(div);
  if (!el) return null;

  if (sigmaChart) {
    sigmaChart.dispose();
  }
  sigmaChart = echarts.init(el, dark ? "dark" : null);

  const topItems = (list || []).slice(0, 15).reverse();
  const names = topItems.map((item) => item.Tag);
  const counts = topItems.map((item) => item.Count);

  const option: echarts.EChartsOption = {
    backgroundColor: "transparent",
    title: {
      text: title || "上位 MITRE ATT&CK & コンプライアンス タグ",
      left: "center",
      top: 5,
      textStyle: {
        fontSize: 13,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    grid: {
      left: "25%",
      right: "8%",
      bottom: "10%",
      top: "14%",
    },
    xAxis: {
      type: "value",
      axisLabel: { color: dark ? "#8b949e" : "#57606a" },
      splitLine: { lineStyle: { color: dark ? "#30363d" : "#d0d7de" } },
    },
    yAxis: {
      type: "category",
      data: names.length > 0 ? names : [noDataText || "データなし"],
      axisLabel: {
        color: dark ? "#c9d1d9" : "#24292f",
        fontSize: 11,
      },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: dark ? "#30363d" : "#d0d7de" } },
    },
    series: [
      {
        type: "bar",
        data: counts.length > 0 ? counts : [0],
        itemStyle: {
          color: (params: any) => {
            const tag = (names[params.dataIndex] || "").toLowerCase();
            if (tag.startsWith("attack.")) return "#f85149";
            if (
              tag.startsWith("pci") ||
              tag.startsWith("nist") ||
              tag.startsWith("gdpr") ||
              tag.startsWith("cis") ||
              tag.startsWith("compliance")
            ) {
              return "#2ea44f";
            }
            return "#0969da";
          },
          borderRadius: [0, 4, 4, 0],
        },
        label: {
          show: true,
          position: "right",
          color: dark ? "#c9d1d9" : "#24292f",
        },
      },
    ],
  };

  sigmaChart.setOption(option, true);
  return sigmaChart;
};

export const showSigmaTimelineChart = (
  div: string | HTMLElement,
  timeline: SigmaTimelineItem[] | undefined,
  dark: boolean = false,
  title?: string,
  noDataText?: string
) => {
  const el = getChartElement(div);
  if (!el) return null;

  if (sigmaChart) {
    sigmaChart.dispose();
  }
  sigmaChart = echarts.init(el, dark ? "dark" : null);

  const times = (timeline || []).map((tp) => {
    const ms = tp.Time > 1e12 ? Math.floor(tp.Time / 1e6) : tp.Time * 1000;
    const d = new Date(ms);
    return echarts.time.format(d, "{yyyy}/{MM}/{dd} {HH}:{mm}");
  });

  const critical = (timeline || []).map((tp) => tp.Critical);
  const high = (timeline || []).map((tp) => tp.High);
  const medium = (timeline || []).map((tp) => tp.Medium);
  const low = (timeline || []).map((tp) => tp.Low);

  const option: echarts.EChartsOption = {
    backgroundColor: "transparent",
    title: {
      text: title || "脅威検知推移 (Timeline)",
      left: "center",
      top: 5,
      textStyle: {
        fontSize: 13,
        color: dark ? "#c9d1d9" : "#24292f",
      },
    },
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    legend: {
      data: ["Critical", "High", "Medium", "Low"],
      top: "12%",
      textStyle: { color: dark ? "#c9d1d9" : "#24292f" },
    },
    grid: {
      left: "8%",
      right: "6%",
      bottom: "16%",
      top: "24%",
    },
    xAxis: {
      type: "category",
      data: times.length > 0 ? times : [noDataText || "データなし"],
      axisLabel: {
        color: dark ? "#8b949e" : "#57606a",
        rotate: 30,
      },
      axisLine: { lineStyle: { color: dark ? "#30363d" : "#d0d7de" } },
    },
    yAxis: {
      type: "value",
      axisLabel: { color: dark ? "#8b949e" : "#57606a" },
      splitLine: { lineStyle: { color: dark ? "#30363d" : "#d0d7de" } },
    },
    series: [
      {
        name: "Critical",
        type: "bar",
        stack: "total",
        itemStyle: { color: "#cf222e" },
        data: critical,
      },
      {
        name: "High",
        type: "bar",
        stack: "total",
        itemStyle: { color: "#f85149" },
        data: high,
      },
      {
        name: "Medium",
        type: "bar",
        stack: "total",
        itemStyle: { color: "#d29922" },
        data: medium,
      },
      {
        name: "Low",
        type: "bar",
        stack: "total",
        itemStyle: { color: "#58a6ff" },
        data: low,
      },
    ],
  };

  sigmaChart.setOption(option, true);
  return sigmaChart;
};

export const resizeSigmaChart = () => {
  if (sigmaChart) {
    sigmaChart.resize();
  }
};

export const disposeSigmaChart = () => {
  if (sigmaChart) {
    sigmaChart.dispose();
    sigmaChart = null;
  }
};
