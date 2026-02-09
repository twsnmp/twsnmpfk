import * as echarts from "echarts";
import "echarts-gl";

import { _, unwrapFunctionStore } from "svelte-i18n";
const $_ = unwrapFunctionStore(_);

let chart: any;

export const showFDBTableGraph = (div: string, fdbTable: any) => {
  const nodeMap = new Map();
  const edgeMap = new Map();
  fdbTable.forEach((f: any) => {
    let n1 = f.Port + "";
    let n2 = f.Node ? f.Node + "(" + f.MAC + ")" : f.MAC;
    let ek = n1 + "|" + n2;
    let e = edgeMap.get(ek);
    if (!e) {
      edgeMap.set(ek, {
        source: n1,
        target: n2,
        value: 1,
        lineStyle: {
          width: 1,
        },
      });
    } else {
      e.value += 1;
    }
    let n = nodeMap.get(n1);
    if (!n) {
      nodeMap.set(n1, {
        name: n1,
        value: 1,
        draggable: true,
        category: 0,
        symbolSize: 4,
      });
    } else {
      n.value += 1;
      n.category = 1;
    }
    n = nodeMap.get(n2);
    if (!n) {
      nodeMap.set(n2, {
        name: n2,
        value: 0,
        draggable: true,
        category: 2,
        symbolSize: 2,
      });
    }
  });
  const nodes = Array.from(nodeMap.values());
  const edges = Array.from(edgeMap.values());
  const categories = [
    { name: "Port", itemStyle: { color: "#1f78b4" } },
    { name: "Dup Port", itemStyle: { color: "#e31a1c" } },
    { name: "Node", itemStyle: { color: "#fbca00" } },
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
  options.series = [
    {
      name: "Flows",
      type: "graph",
      layout: "circular",
      circular: {
        rotateLabel: true,
      },
      data: nodes,
      links: edges,
      categories,
      roam: true,
      label: {
        show: true,
        position: "right",
        formatter: "{b}",
        fontSize: 8,
        fontStyle: "normal",
        color: "#eee",
      },
      lineStyle: {
        color: "source",
        curveness: 0.3,
      },
    },
  ];
  chart.setOption(options);
  chart.resize();
  return chart;
};
