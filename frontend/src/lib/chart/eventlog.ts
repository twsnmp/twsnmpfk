import * as echarts from 'echarts';
import { _,unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart:any;

export const showLogHeatmap = (div:string, logs:any) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const hours = [
    '0',
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    '12',
    '13',
    '14',
    '15',
    '16',
    '17',
    '18',
    '19',
    '20',
    '21',
    '22',
    '23',
  ]
  const option :any = {
    title: {
      show: false,
    },
    grid: {
      left: '5%',
      right: '5%',
      top: 30,
      bottom: 70,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [
      {
        bottom: 15,
        height: 15,
      }
    ],
    tooltip: {
      trigger: 'item',
      formatter(params:any) {
        return (
          params.name +
          ' ' +
          params.data[1] +
          ': ' +
          params.data[2].toFixed(1)
        )
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    xAxis: {
      type: 'category',
      name: $_("Ts.Date"),
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
      data: [],
    },
    yAxis: {
      type: 'category',
      name: $_("Ts.TimeRange"),
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
      data: hours,
    },
    visualMap: {
      min: Infinity,
      max: -Infinity,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    series: [
      {
        name: 'Score',
        type: 'heatmap',
        data: [],
        emphasis: {
          itemStyle: {
            borderColor: '#ccc',
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  }
  if (logs) {
    let nD = 0
    let nH = 0
    let x = -1
    let sum = 0
    logs.sort((a:any, b:any) => a.Time - b.Time)
    logs.forEach((l:any) => {
      const t = new Date(l.Time / (1000 * 1000))
      if (nD === 0) {
        nH = t.getHours()
        nD = t.getDate()
        option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}',false))
        x++
        sum++
        return
      }
      if (t.getHours() !== nH) {
        if (nD !== t.getDate()) {
          option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}',false))
          nD = t.getDate()
          x++
        }
        option.series[0].data.push([x, t.getHours(), sum])
        if (option.visualMap.min > sum) {
          option.visualMap.min = sum
        }
        if (option.visualMap.max < sum) {
          option.visualMap.max = sum
        }
        sum = 0
        nH = t.getHours()
      }
      sum++
    })
    if (option.series[0].data.length < 1) {
      option.series[0].data.push([x, nH, sum])
      option.visualMap.min = 0
      option.visualMap.max = sum
    } else {
      option.series[0].data.push([x, nH, sum])
    }
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showEventLogStateChart = (div:string, logs:any) => {
  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div),"dark");
  const option = {
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      top: 15,
      data: [ $_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"), $_("Ts.Normal"),$_("Ts.Repair"),$_("Ts.Other")],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: $_("Ts.CountByState"),
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [
          { name: $_("Ts.High"), value: 0 },
          { name: $_("Ts.Low"), value: 0 },
          { name: $_("Ts.Warn"), value: 0 },
          { name: $_("Ts.Normal"), value: 0 },
          { name: $_("Ts.Repair"), value: 0 },
          { name: $_("Ts.Other"), value: 0 },
        ],
      },
    ],
  }
  if (logs) {
    logs.forEach((l:any) => {
      switch (l.Level) {
        case 'high':
          option.series[0].data[0].value++
          break
        case 'low':
          option.series[0].data[1].value++
          break
        case 'warn':
          option.series[0].data[2].value++
          break
        case 'normal':
          option.series[0].data[3].value++
          break
        case 'repair':
          option.series[0].data[4].value++
          break
        default:
          option.series[0].data[5].value++
      }
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

export const showEventLogTimeChart = (div:string, type:any, logs:any) => {
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
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [
      {
        bottom: 15,
        height: 15,
      }
    ],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 30,
      bottom: 70,
    },
    xAxis: {
      type: 'time',
      name: $_("Ts.DateTime"),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter(value:any) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}',false)
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: type === 'oprate' ? $_("Ts.Oprate")+ "%" : $_("Ts.Usage") +"%",
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
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
    series: [
      {
        color: '#1f78b4',
        type: 'line',
        name: type === 'oprate' ? $_("Ts.Oprate") : $_("Ts.Usage"),
        showSymbol: false,
        data: [],
      },
    ],
  }
  if (logs) {
    logs.forEach((l:any) => {
      if (l.Type !== type) {
        return
      }
      const t = new Date(l.Time / (1000 * 1000))
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',false)
      const m = l.Event.match(/[0-9.]+%/)
      if (!m || m.length < 1) {
        return
      }
      const val = m[0].replace('%', '') * 1.0
      option.series[0].data.push({
        name: ts,
        value: [t, val],
      })
    })
  }
  chart.setOption(option);
  chart.resize();
  return chart;
}

const getEventLogNodeList = (logs:any) => {
  const m = new Map()
  logs.forEach((l:any) => {
    if (!l.NodeID) {
      return
    }
    let e = m.get(l.NodeID)
    if (!e) {
      m.set(l.NodeID, {
        Name: l.NodeName,
        total: 0,
        high: 0,
        low: 0,
        warn: 0,
        normal: 0,
        repair: 0,
        other: 0,
      })
      e = m.get(l.NodeID)
      if (!e) {
        return
      }
    }
    e.total++
    switch (l.Level) {
      case 'high':
        e.high++
        break
      case 'low':
        e.low++
        break
      case 'warn':
        e.warn++
        break
      case 'normal':
        e.normal++
        break
      case 'repair':
        e.repair++
        break
      default:
        e.other++
    }
  })
  const r = Array.from(m.values())
  return r
}

export const showEventLogNodeChart = (div:any, logs:any) => {
  const list = getEventLogNodeList(logs)
  const high = []
  const low = []
  const warn = []
  const normal = []
  const repair = []
  const other = []
  const category = []
  list.sort((a, b) => b.total - a.total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    high.push(list[i].high)
    low.push(list[i].low)
    warn.push(list[i].warn)
    normal.push(list[i].normal)
    repair.push(list[i].repair)
    other.push(list[i].other)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div),"dark");
  chart.setOption({
    title: {
      show: false,
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
    legend: {
      top: 15,
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [ $_("Ts.High"),$_("Ts.Low"),$_("Ts.Warn"), $_("Ts.Normal"),$_("Ts.Repair"),$_("Ts.Other")],
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
      data: category,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: $_("Ts.High"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: high,
      },
      {
        name: $_("Ts.Low"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: low,
      },
      {
        name: $_("Ts.Warn"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: warn,
      },
      {
        name: $_("Ts.Normal"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: normal,
      },
      {
        name: $_("Ts.Repair"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: repair,
      },
      {
        name: $_("Ts.Other"),
        type: 'bar',
        stack: $_("Ts.NumberOfLog"),
        data: other,
      },
    ],
  })
  chart.resize();
  return chart;
}

export interface DowntimeIncident {
  nodeID: string;
  nodeName: string;
  event: string;
  level: string;
  startTime: number;
  endTime: number;
  durationSec: number;
  ongoing: boolean;
  estimatedStart?: boolean;
}

export interface NodeDowntimeStat {
  nodeID: string;
  nodeName: string;
  totalDowntimeSec: number;
  maxDowntimeSec: number;
  count: number;
  ongoing: boolean;
  currentLevel: string;
  sla: number;
  incidents: DowntimeIncident[];
}

interface TimeInterval {
  start: number;
  end: number;
  ongoing: boolean;
}

// 時間区間のマージアルゴリズム（重複なしの総ダウンタイム計算）
const mergeIntervals = (intervals: TimeInterval[]): { totalSec: number; maxSec: number } => {
  if (intervals.length === 0) return { totalSec: 0, maxSec: 0 };

  const sorted = [...intervals].sort((a, b) => a.start - b.start);
  const merged: TimeInterval[] = [];

  let current = { ...sorted[0] };

  for (let i = 1; i < sorted.length; i++) {
    const next = sorted[i];
    if (next.start <= current.end) {
      current.end = Math.max(current.end, next.end);
      if (next.ongoing) current.ongoing = true;
    } else {
      merged.push(current);
      current = { ...next };
    }
  }
  merged.push(current);

  let totalSec = 0;
  let maxSec = 0;

  merged.forEach((item) => {
    const dur = Math.max(0, (item.end - item.start) / (1000 * 1000 * 1000));
    totalSec += dur;
    if (dur > maxSec) {
      maxSec = dur;
    }
  });

  return { totalSec, maxSec };
};

export const getPollingName = (event: string): string => {
  if (!event) return '';
  let name = event;
  const colonIdx = name.indexOf(':');
  if (colonIdx >= 0) {
    name = name.substring(colonIdx + 1).trim();
  }
  const parenIdx = name.lastIndexOf('(');
  if (parenIdx > 0 && name.endsWith(')')) {
    name = name.substring(0, parenIdx).trim();
  }
  return name || event;
};

export const calcEventLogDowntimeAndSLA = (logs: any) => {
  if (!logs || logs.length === 0) {
    return {
      totalIncidents: 0,
      ongoingIncidents: 0,
      totalDowntimeSec: 0,
      maxDowntimeSec: 0,
      mttrSec: 0,
      overallSLA: 100,
      nodeStats: [],
      incidents: [],
    };
  }

  // polling タイプのログのみに限定 (unknown レベルのログは除外)
  const pollingLogs = logs.filter((l: any) => l.Type === 'polling' && l.Level !== 'unknown');
  if (pollingLogs.length === 0) {
    return {
      totalIncidents: 0,
      ongoingIncidents: 0,
      totalDowntimeSec: 0,
      maxDowntimeSec: 0,
      mttrSec: 0,
      overallSLA: 100,
      nodeStats: [],
      incidents: [],
    };
  }

  const sortedLogs = [...pollingLogs].sort((a: any, b: any) => a.Time - b.Time);
  const minTime = sortedLogs[0].Time;
  const maxTime = sortedLogs[sortedLogs.length - 1].Time;
  let totalSpanSec = (maxTime - minTime) / (1000 * 1000 * 1000);
  if (totalSpanSec <= 0) {
    totalSpanSec = 1;
  }

  const isFailure = (lvl: string) => lvl === 'high' || lvl === 'low' || lvl === 'warn';
  // 復旧は repair のみ (normal は通常ログ・初回ログ等のため含めない)
  const isRepair = (lvl: string) => lvl === 'repair';

  // ポーリング単位 (NodeID + Event または Event) でログをグループ化
  const logsByPolling = new Map<string, { nodeID: string; nodeName: string; event: string; logs: any[] }>();

  sortedLogs.forEach((l: any) => {
    const pName = getPollingName(l.Event);
    const pKey = (l.NodeID || l.NodeName || '') + '_' + pName;
    let group = logsByPolling.get(pKey);
    if (!group) {
      group = {
        nodeID: l.NodeID || '',
        nodeName: l.NodeName || l.NodeID || 'unknown',
        event: pName,
        logs: [],
      };
      logsByPolling.set(pKey, group);
    }
    group.logs.push(l);
  });

  // ポーリング単位で障害・復旧ペアリングを処理
  const nodeIntervalsMap = new Map<
    string,
    { nodeID: string; nodeName: string; lastLevel: string; intervals: TimeInterval[]; incidents: DowntimeIncident[] }
  >();

  let totalRecoveredSec = 0;
  let recoveredCount = 0;
  let allIncidentsCount = 0;

  logsByPolling.forEach((group) => {
    let isDown = false;
    let downStart = 0;
    let lastLevel = 'normal';

    const nKey = group.nodeID || group.nodeName;
    let nData = nodeIntervalsMap.get(nKey);
    if (!nData) {
      nData = {
        nodeID: group.nodeID,
        nodeName: group.nodeName,
        lastLevel: 'normal',
        intervals: [],
        incidents: [],
      };
      nodeIntervalsMap.set(nKey, nData);
    }

    group.logs.forEach((l: any) => {
      lastLevel = l.Level;
      nData!.lastLevel = l.Level;

      if (isFailure(l.Level)) {
        if (!isDown) {
          isDown = true;
          downStart = l.Time;
        }
      } else if (l.Level === 'repair' || l.Level === 'normal') {
        if (isDown) {
          const durSec = Math.max(0, (l.Time - downStart) / (1000 * 1000 * 1000));
          nData!.intervals.push({ start: downStart, end: l.Time, ongoing: false });
          nData!.incidents.push({
            nodeID: group.nodeID,
            nodeName: group.nodeName,
            event: group.event,
            level: l.Level,
            startTime: downStart,
            endTime: l.Time,
            durationSec: durSec,
            ongoing: false,
          });
          totalRecoveredSec += durSec;
          recoveredCount++;
          allIncidentsCount++;
          isDown = false;
        } else if (l.Level === 'repair') {
          // 復旧イベント(repair)のみ存在する場合 -> ログ開始時刻 minTime を仮の障害発生時刻とする
          const durSec = Math.max(0, (l.Time - minTime) / (1000 * 1000 * 1000));
          if (durSec > 0) {
            nData!.intervals.push({ start: minTime, end: l.Time, ongoing: false });
            nData!.incidents.push({
              nodeID: group.nodeID,
              nodeName: group.nodeName,
              event: group.event,
              level: l.Level,
              startTime: minTime,
              endTime: l.Time,
              durationSec: durSec,
              ongoing: false,
              estimatedStart: true,
            });
            totalRecoveredSec += durSec;
            recoveredCount++;
            allIncidentsCount++;
          }
        }
      }
    });

    // ログ終了時点で障害が継続しているポーリング -> min(maxTime) まで継続
    if (isDown) {
      const durSec = Math.max(0, (maxTime - downStart) / (1000 * 1000 * 1000));
      nData!.intervals.push({ start: downStart, end: maxTime, ongoing: true });
      nData!.incidents.push({
        nodeID: group.nodeID,
        nodeName: group.nodeName,
        event: group.event,
        level: lastLevel,
        startTime: downStart,
        endTime: maxTime,
        durationSec: durSec,
        ongoing: true,
      });
      allIncidentsCount++;
    }
  });

  // ノード単位での統合（区間マージによる重複除去とノードSLA算出）
  const nodeStats: NodeDowntimeStat[] = [];
  let globalTotalDowntimeSec = 0;
  let globalMaxDowntimeSec = 0;
  let ongoingNodeCount = 0;
  const allIncidentsList: DowntimeIncident[] = [];

  nodeIntervalsMap.forEach((nData) => {
    const { totalSec, maxSec } = mergeIntervals(nData.intervals);
    const hasOngoing = nData.intervals.some((i) => i.ongoing);
    if (hasOngoing) ongoingNodeCount++;

    if (maxSec > globalMaxDowntimeSec) {
      globalMaxDowntimeSec = maxSec;
    }
    globalTotalDowntimeSec += totalSec;

    const sla = Math.max(0, Math.min(100, (1 - totalSec / totalSpanSec) * 100));

    nodeStats.push({
      nodeID: nData.nodeID,
      nodeName: nData.nodeName,
      totalDowntimeSec: totalSec,
      maxDowntimeSec: maxSec,
      count: nData.incidents.length,
      ongoing: hasOngoing,
      currentLevel: nData.lastLevel,
      sla,
      incidents: nData.incidents,
    });

    allIncidentsList.push(...nData.incidents);
  });

  const nodeCount = Math.max(1, nodeStats.length);
  const overallSLA = nodeStats.reduce((acc, n) => acc + n.sla, 0) / nodeCount;
  const mttrSec = recoveredCount > 0 ? totalRecoveredSec / recoveredCount : 0;

  return {
    totalIncidents: allIncidentsCount,
    ongoingIncidents: ongoingNodeCount,
    totalDowntimeSec: globalTotalDowntimeSec,
    maxDowntimeSec: globalMaxDowntimeSec,
    mttrSec,
    overallSLA,
    nodeStats,
    incidents: allIncidentsList,
  };
};

export const showEventLogDowntimeChart = (div: string, logs: any) => {
  const stats = calcEventLogDowntimeAndSLA(logs);
  const nodeStats = stats.nodeStats.sort((a, b) => b.totalDowntimeSec - a.totalDowntimeSec);

  const categories: string[] = [];
  const downtimeData: number[] = [];
  const slaData: number[] = [];

  // 上位 15 ノード
  const topNodes = nodeStats.slice(0, 15).reverse();
  topNodes.forEach((n) => {
    let name = n.nodeName;
    if (name.length > 15) {
      name = name.substring(0, 13) + '...';
    }
    categories.push(name);
    downtimeData.push(Number((n.totalDowntimeSec / 60).toFixed(1)));
    slaData.push(Number(n.sla.toFixed(3)));
  });

  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), 'dark');
  chart.setOption({
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      formatter: (params: any) => {
        let res = params[0].name + '<br/>';
        params.forEach((p: any) => {
          if (p.seriesName === 'SLA (%)') {
            res += `${p.marker} ${p.seriesName}: ${p.value}%<br/>`;
          } else {
            res += `${p.marker} Downtime: ${p.value} m (${(p.value * 60).toFixed(0)} s)<br/>`;
          }
        });
        return res;
      },
    },
    legend: {
      top: 5,
      data: ['Downtime (min)', 'SLA (%)'],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    grid: {
      left: '3%',
      right: '12%',
      top: '15%',
      bottom: '12%',
      containLabel: true,
    },
    xAxis: [
      {
        type: 'value',
        name: 'Downtime (min)',
        position: 'bottom',
        axisLabel: {
          color: '#ccc',
          fontSize: 9,
        },
      },
      {
        type: 'value',
        name: 'SLA (%)',
        min: 0,
        max: 100,
        position: 'top',
        axisLabel: {
          color: '#ccc',
          fontSize: 9,
          formatter: '{value}%',
        },
        splitLine: {
          show: false,
        },
      },
    ],
    yAxis: {
      type: 'category',
      data: categories,
      axisLabel: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    series: [
      {
        name: 'Downtime (min)',
        type: 'bar',
        color: '#e31a1c',
        data: downtimeData,
      },
      {
        name: 'SLA (%)',
        type: 'line',
        xAxisIndex: 1,
        color: '#33a02c',
        data: slaData,
        markLine: {
          data: [{ xAxis: 99.9, name: 'Target 99.9%' }],
          lineStyle: {
            color: '#dfdf22',
            type: 'dashed',
          },
          label: {
            formatter: 'Target 99.9%',
            color: '#dfdf22',
            fontSize: 9,
          },
        },
      },
    ],
  });
  chart.resize();
  return chart;
};

export interface PollingDowntimeStat {
  event: string;
  totalDowntimeSec: number;
  maxDowntimeSec: number;
  count: number;
  ongoing: boolean;
  currentLevel: string;
  sla: number;
  incidents: DowntimeIncident[];
}

export const calcNodeDowntimeAndSLA = (logs: any) => {
  if (!logs || logs.length === 0) {
    return {
      totalIncidents: 0,
      ongoingIncidents: 0,
      totalDowntimeSec: 0,
      maxDowntimeSec: 0,
      mttrSec: 0,
      overallSLA: 100,
      pollingStats: [] as PollingDowntimeStat[],
      incidents: [] as DowntimeIncident[],
    };
  }

  // polling タイプのログのみに限定 (unknown レベルのログは除外)
  const pollingLogs = logs.filter((l: any) => l.Type === 'polling' && l.Level !== 'unknown');
  if (pollingLogs.length === 0) {
    return {
      totalIncidents: 0,
      ongoingIncidents: 0,
      totalDowntimeSec: 0,
      maxDowntimeSec: 0,
      mttrSec: 0,
      overallSLA: 100,
      pollingStats: [] as PollingDowntimeStat[],
      incidents: [] as DowntimeIncident[],
    };
  }

  const sortedLogs = [...pollingLogs].sort((a: any, b: any) => a.Time - b.Time);
  const minTime = sortedLogs[0].Time;
  const maxTime = sortedLogs[sortedLogs.length - 1].Time;
  let totalSpanSec = (maxTime - minTime) / (1000 * 1000 * 1000);
  if (totalSpanSec <= 0) {
    totalSpanSec = 1;
  }

  const isFailure = (lvl: string) => lvl === 'high' || lvl === 'low' || lvl === 'warn';
  const isRepair = (lvl: string) => lvl === 'repair';

  // ポーリング項目 (Event) でログをグループ化
  const logsByPolling = new Map<string, { event: string; logs: any[] }>();

  sortedLogs.forEach((l: any) => {
    const pName = getPollingName(l.Event) || 'default';
    const pKey = pName;
    let group = logsByPolling.get(pKey);
    if (!group) {
      group = {
        event: pName,
        logs: [],
      };
      logsByPolling.set(pKey, group);
    }
    group.logs.push(l);
  });

  const pollingStats: PollingDowntimeStat[] = [];
  const nodeAllIntervals: TimeInterval[] = [];
  let totalRecoveredSec = 0;
  let recoveredCount = 0;
  let allIncidentsCount = 0;
  let ongoingPollingCount = 0;
  const allIncidentsList: DowntimeIncident[] = [];

  logsByPolling.forEach((group) => {
    let isDown = false;
    let downStart = 0;
    let lastLevel = 'normal';
    const intervals: TimeInterval[] = [];
    const incidents: DowntimeIncident[] = [];

    group.logs.forEach((l: any) => {
      lastLevel = l.Level;

      if (isFailure(l.Level)) {
        if (!isDown) {
          isDown = true;
          downStart = l.Time;
        }
      } else if (l.Level === 'repair' || l.Level === 'normal') {
        if (isDown) {
          const durSec = Math.max(0, (l.Time - downStart) / (1000 * 1000 * 1000));
          intervals.push({ start: downStart, end: l.Time, ongoing: false });
          incidents.push({
            nodeID: l.NodeID || '',
            nodeName: l.NodeName || '',
            event: group.event,
            level: l.Level,
            startTime: downStart,
            endTime: l.Time,
            durationSec: durSec,
            ongoing: false,
          });
          totalRecoveredSec += durSec;
          recoveredCount++;
          allIncidentsCount++;
          isDown = false;
        } else if (l.Level === 'repair') {
          const durSec = Math.max(0, (l.Time - minTime) / (1000 * 1000 * 1000));
          if (durSec > 0) {
            intervals.push({ start: minTime, end: l.Time, ongoing: false });
            incidents.push({
              nodeID: l.NodeID || '',
              nodeName: l.NodeName || '',
              event: group.event,
              level: l.Level,
              startTime: minTime,
              endTime: l.Time,
              durationSec: durSec,
              ongoing: false,
              estimatedStart: true,
            });
            totalRecoveredSec += durSec;
            recoveredCount++;
            allIncidentsCount++;
          }
        }
      }
    });

    if (isDown) {
      const durSec = Math.max(0, (maxTime - downStart) / (1000 * 1000 * 1000));
      intervals.push({ start: downStart, end: maxTime, ongoing: true });
      incidents.push({
        nodeID: group.logs[0]?.NodeID || '',
        nodeName: group.logs[0]?.NodeName || '',
        event: group.event,
        level: lastLevel,
        startTime: downStart,
        endTime: maxTime,
        durationSec: durSec,
        ongoing: true,
      });
      allIncidentsCount++;
    }

    const { totalSec, maxSec } = mergeIntervals(intervals);
    const hasOngoing = intervals.some((i) => i.ongoing);
    if (hasOngoing) ongoingPollingCount++;

    const sla = Math.max(0, Math.min(100, (1 - totalSec / totalSpanSec) * 100));

    pollingStats.push({
      event: group.event,
      totalDowntimeSec: totalSec,
      maxDowntimeSec: maxSec,
      count: incidents.length,
      ongoing: hasOngoing,
      currentLevel: lastLevel,
      sla,
      incidents,
    });

    nodeAllIntervals.push(...intervals);
    allIncidentsList.push(...incidents);
  });

  const { totalSec: nodeTotalDowntimeSec, maxSec: nodeMaxDowntimeSec } = mergeIntervals(nodeAllIntervals);
  const overallSLA = Math.max(0, Math.min(100, (1 - nodeTotalDowntimeSec / totalSpanSec) * 100));
  const mttrSec = recoveredCount > 0 ? totalRecoveredSec / recoveredCount : 0;

  return {
    totalIncidents: allIncidentsCount,
    ongoingIncidents: ongoingPollingCount,
    totalDowntimeSec: nodeTotalDowntimeSec,
    maxDowntimeSec: nodeMaxDowntimeSec,
    mttrSec,
    overallSLA,
    pollingStats,
    incidents: allIncidentsList,
  };
};

export const showNodeDowntimeChart = (div: string, logs: any) => {
  const stats = calcNodeDowntimeAndSLA(logs);
  const pollingStats = stats.pollingStats.sort((a, b) => b.totalDowntimeSec - a.totalDowntimeSec);

  const categories: string[] = [];
  const downtimeData: number[] = [];
  const slaData: number[] = [];

  const topItems = pollingStats.slice(0, 15).reverse();
  topItems.forEach((p) => {
    let name = p.event;
    if (name.length > 20) {
      name = name.substring(0, 18) + '...';
    }
    categories.push(name);
    downtimeData.push(Number((p.totalDowntimeSec / 60).toFixed(1)));
    slaData.push(Number(p.sla.toFixed(3)));
  });

  if (chart) {
    chart.dispose();
  }
  chart = echarts.init(document.getElementById(div), 'dark');
  chart.setOption({
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      formatter: (params: any) => {
        let res = params[0].name + '<br/>';
        params.forEach((p: any) => {
          if (p.seriesName === 'SLA (%)') {
            res += `${p.marker} ${p.seriesName}: ${p.value}%<br/>`;
          } else {
            res += `${p.marker} Downtime: ${p.value} m (${(p.value * 60).toFixed(0)} s)<br/>`;
          }
        });
        return res;
      },
    },
    legend: {
      top: 5,
      data: ['Downtime (min)', 'SLA (%)'],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    grid: {
      left: '3%',
      right: '12%',
      top: '15%',
      bottom: '12%',
      containLabel: true,
    },
    xAxis: [
      {
        type: 'value',
        name: 'Downtime (min)',
        position: 'bottom',
        axisLabel: {
          color: '#ccc',
          fontSize: 9,
        },
      },
      {
        type: 'value',
        name: 'SLA (%)',
        min: 0,
        max: 100,
        position: 'top',
        axisLabel: {
          color: '#ccc',
          fontSize: 9,
          formatter: '{value}%',
        },
        splitLine: {
          show: false,
        },
      },
    ],
    yAxis: {
      type: 'category',
      data: categories,
      axisLabel: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    series: [
      {
        name: 'Downtime (min)',
        type: 'bar',
        color: '#e31a1c',
        data: downtimeData,
      },
      {
        name: 'SLA (%)',
        type: 'line',
        xAxisIndex: 1,
        color: '#33a02c',
        data: slaData,
        markLine: {
          data: [{ xAxis: 99.9, name: 'Target 99.9%' }],
          lineStyle: {
            color: '#dfdf22',
            type: 'dashed',
          },
          label: {
            formatter: 'Target 99.9%',
            color: '#dfdf22',
            fontSize: 9,
          },
        },
      },
    ],
  });
  chart.resize();
  return chart;
};




