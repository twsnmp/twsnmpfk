import * as echarts from "echarts";
import { _, unwrapFunctionStore } from "svelte-i18n";

const $_ = unwrapFunctionStore(_);

const isDark = (): boolean =>
  typeof document !== "undefined" &&
  document.documentElement.classList.contains("dark");

const initChart = (div: string): echarts.ECharts | null => {
  const el = document.getElementById(div);
  if (!el) {
    return null;
  }
  const existing = echarts.getInstanceByDom(el);
  if (existing && !existing.isDisposed()) {
    try {
      existing.dispose();
    } catch {
      // ignore
    }
  }
  return echarts.init(el, isDark() ? "dark" : undefined);
};

export const disposeCorrelationChart = (div: string) => {
  const el = document.getElementById(div);
  if (!el) return;
  const existing = echarts.getInstanceByDom(el);
  if (existing && !existing.isDisposed()) {
    try {
      existing.dispose();
    } catch {
      // ignore
    }
  }
};

const formatCompactAxis = (v: number): string => {
  if (!isFinite(v)) return "";
  const abs = Math.abs(v);
  if (abs === 0) return "0";
  if (abs >= 1e9) return `${Number((v / 1e9).toFixed(2))}G`;
  if (abs >= 1e6) return `${Number((v / 1e6).toFixed(2))}M`;
  if (abs >= 1e3) return `${Number((v / 1e3).toFixed(2))}k`;
  if (abs < 0.01) return v.toExponential(1);
  return `${Number(v.toFixed(2))}`;
};

export const showCorrelationTimeChart = (
  div: string,
  res: any,
  normalized: boolean
): echarts.ECharts | null => {
  const chart = initChart(div);
  if (!chart || !res || !res.Timestamps) return chart;

  const dark = isDark();
  const textColor = dark ? "#cbd5e1" : "#334155";
  const name1 = `${res.NodeName1 ? res.NodeName1 + " / " : ""}${res.PollingName1} (${res.SelectedVar1})`;
  const name2 = `${res.NodeName2 ? res.NodeName2 + " / " : ""}${res.PollingName2} (${res.SelectedVar2})`;

  const s1Vals = normalized ? res.NormSeries1 || [] : res.Series1 || [];
  const s2Vals = normalized ? res.NormSeries2 || [] : res.Series2 || [];
  const d1 = res.Timestamps.map((ts: number, i: number) => [ts * 1000, s1Vals[i] ?? 0]);
  const d2 = res.Timestamps.map((ts: number, i: number) => [ts * 1000, s2Vals[i] ?? 0]);

  const yAxes: any[] = normalized
    ? [
        {
          type: "value",
          min: 0,
          max: 1,
          name: $_("Correlation.Normalized") || "Normalized (0-1)",
          nameTextStyle: { color: textColor, fontSize: 10 },
          axisLabel: { color: textColor, fontSize: 10 },
        },
      ]
    : [
        {
          type: "value",
          name: res.SelectedVar1 || "Var 1",
          nameTextStyle: { color: "#3b82f6", fontSize: 10 },
          axisLabel: {
            color: textColor,
            fontSize: 10,
            formatter: (v: number) => formatCompactAxis(v),
          },
        },
        {
          type: "value",
          name: res.SelectedVar2 || "Var 2",
          nameTextStyle: { color: "#10b981", fontSize: 10 },
          axisLabel: {
            color: textColor,
            fontSize: 10,
            formatter: (v: number) => formatCompactAxis(v),
          },
          splitLine: { show: false },
        },
      ];

  const option: any = {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "cross" },
    },
    legend: {
      type: "scroll",
      top: 4,
      itemGap: 16,
      data: [name1, name2],
      textStyle: { color: textColor, fontSize: 11 },
    },
    grid: {
      left: 16,
      right: 16,
      top: 48,
      bottom: 56,
      containLabel: true,
    },
    dataZoom: [
      { type: "inside" },
      { type: "slider", bottom: 10, height: 18 },
    ],
    xAxis: {
      type: "time",
      axisLabel: {
        color: textColor,
        fontSize: 10,
        formatter: (value: number) =>
          echarts.time.format(new Date(value), "{MM}/{dd} {HH}:{mm}", false),
      },
    },
    yAxis: yAxes,
    series: [
      {
        name: name1,
        type: "line",
        yAxisIndex: 0,
        showSymbol: false,
        smooth: false,
        lineStyle: { width: 2, color: "#3b82f6" },
        itemStyle: { color: "#3b82f6" },
        data: d1,
      },
      {
        name: name2,
        type: "line",
        yAxisIndex: normalized ? 0 : 1,
        showSymbol: false,
        smooth: false,
        lineStyle: { width: 2, color: "#10b981" },
        itemStyle: { color: "#10b981" },
        data: d2,
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};

export const showCorrelationScatterChart = (
  div: string,
  res: any
): echarts.ECharts | null => {
  const chart = initChart(div);
  if (!chart || !res || !res.Series1 || !res.Series2) return chart;

  const dark = isDark();
  const textColor = dark ? "#cbd5e1" : "#334155";
  const s1: number[] = res.Series1;
  const s2: number[] = res.Series2;
  const ts: number[] = res.Timestamps || [];

  let minX = Infinity;
  let maxX = -Infinity;
  const points = s1.map((x, i) => {
    const y = s2[i] ?? 0;
    if (x < minX) minX = x;
    if (x > maxX) maxX = x;
    return [x, y, ts[i] || 0];
  });

  const slope = Number(res.Slope) || 0;
  const intercept = Number(res.Intercept) || 0;
  const regLine =
    isFinite(minX) && isFinite(maxX) && s1.length >= 2
      ? [
          [minX, slope * minX + intercept],
          [maxX, slope * maxX + intercept],
        ]
      : [];

  const xName = `${res.PollingName1} (${res.SelectedVar1})`;
  const yName = `${res.PollingName2} (${res.SelectedVar2})`;

  const option: any = {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "item",
      formatter: (p: any) => {
        if (p.seriesType === "line") {
          return `y = ${slope.toFixed(4)}x + ${intercept.toFixed(4)}`;
        }
        const t = p.data[2]
          ? echarts.time.format(new Date(p.data[2] * 1000), "{yyyy}/{MM}/{dd} {HH}:{mm}", false)
          : "";
        return `${t ? t + "<br/>" : ""}${xName}: <b>${Number(p.data[0]).toFixed(4)}</b><br/>${yName}: <b>${Number(p.data[1]).toFixed(4)}</b>`;
      },
    },
    grid: {
      left: 16,
      right: 20,
      top: 32,
      bottom: 36,
      containLabel: true,
    },
    xAxis: {
      type: "value",
      name: xName,
      nameLocation: "middle",
      nameGap: 26,
      nameTextStyle: { color: textColor, fontSize: 11 },
      axisLabel: {
        color: textColor,
        fontSize: 10,
        formatter: (v: number) => formatCompactAxis(v),
      },
      scale: true,
    },
    yAxis: {
      type: "value",
      name: yName,
      nameTextStyle: { color: textColor, fontSize: 11 },
      axisLabel: {
        color: textColor,
        fontSize: 10,
        formatter: (v: number) => formatCompactAxis(v),
      },
      scale: true,
    },
    series: [
      {
        name: $_("Correlation.Samples") || "Samples",
        type: "scatter",
        symbolSize: 7,
        itemStyle: { color: "#3b82f6", opacity: 0.75 },
        data: points,
      },
      {
        name: $_("Correlation.RegressionLine") || "Regression",
        type: "line",
        showSymbol: false,
        lineStyle: { color: "#f59e0b", width: 2, type: "dashed" },
        data: regLine,
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};

export const showCorrelationLagChart = (
  div: string,
  res: any
): echarts.ECharts | null => {
  const chart = initChart(div);
  if (!chart || !res) return chart;

  const dark = isDark();
  const textColor = dark ? "#cbd5e1" : "#334155";
  const prof: any[] = res.LagProfile || [];
  const bestLag = Number(res.BestLag) || 0;

  const labels = prof.map((p) => {
    const sign = p.Lag > 0 ? "+" : "";
    return `${sign}${p.Lag} (${sign}${p.LagSec}s)`;
  });
  const values = prof.map((p) => ({
    value: Number(p.Score.toFixed(4)),
    itemStyle: {
      color:
        p.Lag === bestLag
          ? "#f59e0b"
          : p.Score >= 0
          ? "#3b82f6"
          : "#ef4444",
    },
  }));

  const option: any = {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
    },
    grid: {
      left: "8%",
      right: "5%",
      top: 32,
      bottom: 62,
    },
    xAxis: {
      type: "category",
      data: labels,
      name: $_("Correlation.LagAxis") || "Lag (buckets / seconds)",
      nameLocation: "middle",
      nameGap: 42,
      nameTextStyle: { color: textColor, fontSize: 11 },
      axisLabel: { color: textColor, fontSize: 9, rotate: 30 },
    },
    yAxis: {
      type: "value",
      min: -1,
      max: 1,
      name: $_("Correlation.Score") || "Correlation",
      nameTextStyle: { color: textColor, fontSize: 11 },
      axisLabel: { color: textColor, fontSize: 10 },
    },
    series: [
      {
        name: $_("Correlation.CrossCorr") || "Cross-Correlation",
        type: "bar",
        data: values,
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};

export const showCorrelationMatrixChart = (
  div: string,
  res: any
): echarts.ECharts | null => {
  const chart = initChart(div);
  if (!chart || !res) return chart;

  const dark = isDark();
  const textColor = dark ? "#cbd5e1" : "#334155";
  const vars1: string[] = res.Vars1 || [];
  const vars2: string[] = res.Vars2 || [];
  const cells: any[] = res.VarMatrix || [];

  const data: any[] = [];
  for (const c of cells) {
    const xi = vars1.indexOf(c.Var1);
    const yi = vars2.indexOf(c.Var2);
    if (xi >= 0 && yi >= 0) {
      data.push([xi, yi, Number((c.Score || 0).toFixed(3))]);
    }
  }

  const option: any = {
    backgroundColor: "transparent",
    tooltip: {
      position: "top",
      formatter: (p: any) => {
        const v1 = vars1[p.data[0]] || "";
        const v2 = vars2[p.data[1]] || "";
        return `${v1} × ${v2}: <b>${p.data[2]}</b>`;
      },
    },
    grid: {
      left: "14%",
      right: "10%",
      top: 24,
      bottom: 52,
    },
    xAxis: {
      type: "category",
      data: vars1,
      axisLabel: { color: textColor, fontSize: 10, rotate: vars1.length > 5 ? 25 : 0 },
      splitArea: { show: true },
    },
    yAxis: {
      type: "category",
      data: vars2,
      axisLabel: { color: textColor, fontSize: 10 },
      splitArea: { show: true },
    },
    visualMap: {
      min: -1,
      max: 1,
      calculable: true,
      orient: "vertical",
      right: 6,
      top: "center",
      textStyle: { color: textColor, fontSize: 10 },
      inRange: {
        color: ["#2563eb", "#93c5fd", dark ? "#334155" : "#f8fafc", "#fca5a5", "#dc2626"],
      },
    },
    series: [
      {
        name: $_("Correlation.Matrix") || "Matrix",
        type: "heatmap",
        data,
        label: {
          show: true,
          fontSize: 10,
          color: dark ? "#f8fafc" : "#0f172a",
        },
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};

export const showCorrelationNetworkChart = (
  div: string,
  items: any[],
  onEdgeClick?: (item: any) => void
): echarts.ECharts | null => {
  const chart = initChart(div);
  if (!chart) return chart;

  const dark = isDark();
  const textColor = dark ? "#e2e8f0" : "#1e293b";

  const nodeMap = new Map<string, any>();
  const categorySet = new Set<string>();

  for (const it of items || []) {
    const c1 = it.NodeName1 || it.NodeID1 || "Unknown";
    const c2 = it.NodeName2 || it.NodeID2 || "Unknown";
    categorySet.add(c1);
    categorySet.add(c2);
  }
  const categories = Array.from(categorySet).map((name) => ({ name }));
  const catIdxMap = new Map<string, number>();
  categories.forEach((c, idx) => catIdxMap.set(c.name, idx));

  const degreeMap = new Map<string, number>();
  for (const it of items || []) {
    degreeMap.set(it.PollingID1, (degreeMap.get(it.PollingID1) || 0) + 1);
    degreeMap.set(it.PollingID2, (degreeMap.get(it.PollingID2) || 0) + 1);
  }

  for (const it of items || []) {
    if (!nodeMap.has(it.PollingID1)) {
      const cat = it.NodeName1 || it.NodeID1 || "Unknown";
      const deg = degreeMap.get(it.PollingID1) || 1;
      nodeMap.set(it.PollingID1, {
        id: it.PollingID1,
        name: `${cat}\n${it.PollingName1}`,
        category: catIdxMap.get(cat) ?? 0,
        symbolSize: Math.min(40, 16 + deg * 4),
        value: deg,
      });
    }
    if (!nodeMap.has(it.PollingID2)) {
      const cat = it.NodeName2 || it.NodeID2 || "Unknown";
      const deg = degreeMap.get(it.PollingID2) || 1;
      nodeMap.set(it.PollingID2, {
        id: it.PollingID2,
        name: `${cat}\n${it.PollingName2}`,
        category: catIdxMap.get(cat) ?? 0,
        symbolSize: Math.min(40, 16 + deg * 4),
        value: deg,
      });
    }
  }

  const links = (items || []).map((it) => ({
    source: it.PollingID1,
    target: it.PollingID2,
    value: Number((it.Score || 0).toFixed(3)),
    rawItem: it,
    lineStyle: {
      width: 1 + (it.AbsScore || Math.abs(it.Score || 0)) * 4,
      color: (it.Score || 0) >= 0 ? "#3b82f6" : "#ef4444",
      curveness: 0.1,
      opacity: 0.8,
    },
    label: {
      show: (items || []).length <= 25,
      formatter: () => `${Number(it.Score || 0).toFixed(2)}`,
      fontSize: 9,
      color: textColor,
    },
  }));

  const option: any = {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "item",
      formatter: (p: any) => {
        if (p.dataType === "edge" && p.data?.rawItem) {
          const it = p.data.rawItem;
          return `${it.NodeName1}/${it.PollingName1} (${it.Var1})<br/>↔ ${it.NodeName2}/${it.PollingName2} (${it.Var2})<br/>Score: <b>${Number(it.Score).toFixed(4)}</b> (n=${it.SampleCount})`;
        }
        return `${String(p.name || "").replace("\n", " / ")} (${p.value || 0})`;
      },
    },
    legend:
      categories.length > 1 && categories.length <= 12
        ? [
            {
              data: categories.map((c) => c.name),
              textStyle: { color: textColor, fontSize: 10 },
              top: 4,
            },
          ]
        : [],
    series: [
      {
        type: "graph",
        layout: "force",
        roam: true,
        draggable: true,
        data: Array.from(nodeMap.values()),
        links,
        categories,
        label: {
          show: true,
          position: "right",
          fontSize: 10,
          color: textColor,
        },
        force: {
          repulsion: 220,
          edgeLength: [70, 160],
        },
        emphasis: {
          focus: "adjacency",
        },
      },
    ],
  };

  chart.setOption(option);
  if (onEdgeClick) {
    chart.off("click");
    chart.on("click", (params: any) => {
      if (params.dataType === "edge" && params.data?.rawItem) {
        onEdgeClick(params.data.rawItem);
      }
    });
  }
  chart.resize();
  return chart;
};
