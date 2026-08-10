import * as echarts from 'echarts';
import { _, unwrapFunctionStore } from 'svelte-i18n';
const $_ = unwrapFunctionStore(_);

let chart: any;

export interface MTRHopStat {
  ttl: number;
  ip: string;
  loc: string;
  snt: number;
  lossRate: number;
  last: number; // ms
  avg: number;  // ms
  best: number; // ms
  wrst: number; // ms
  stDev: number;// ms
  isTarget: boolean;
}

export const showMtrProfileChart = (div: string, mtrList: MTRHopStat[]) => {
  const container = document.getElementById(div);
  if (!container || !mtrList || mtrList.length === 0) {
    return null;
  }

  let chart = echarts.getInstanceByDom(container);
  if (!chart) {
    chart = echarts.init(container, 'dark');
  }

  const hops: string[] = [];
  const minData: (number | null)[] = [];
  const rangeData: (number | null)[] = [];
  const avgData: (number | null)[] = [];
  const wrstData: (number | null)[] = [];
  const lossData: any[] = [];

  mtrList.forEach((h) => {
    const label = `${h.ttl}: ${h.ip || '???'}`;
    hops.push(label);

    if (h.snt > 0 && h.avg >= 0) {
      const best = Number(h.best.toFixed(2));
      const avg = Number(h.avg.toFixed(2));
      const wrst = Number(h.wrst.toFixed(2));
      const range = Number(Math.max(0, wrst - best).toFixed(2));

      minData.push(best);
      rangeData.push(range);
      avgData.push(avg);
      wrstData.push(wrst);
    } else {
      minData.push(null);
      rangeData.push(null);
      avgData.push(null);
      wrstData.push(null);
    }

    const lossVal = Number(h.lossRate.toFixed(1));
    let color = 'rgba(34, 197, 94, 0.3)'; // 0% loss: green
    if (lossVal > 20) {
      color = '#ef4444'; // >20%: red
    } else if (lossVal > 0) {
      color = '#f59e0b'; // >0%: yellow
    }
    lossData.push({
      value: lossVal,
      itemStyle: { color },
    });
  });

  chart = echarts.init(container, 'dark');

  const option: any = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
      formatter(params: any) {
        if (!params || params.length === 0) return '';
        const idx = params[0].dataIndex;
        const stat = mtrList[idx];
        if (!stat) return '';

        return `<div class="text-xs font-sans space-y-1 p-1">` +
               `<div class="font-bold border-b border-gray-600 pb-1 mb-1">Hop ${stat.ttl}: ${stat.ip || 'No Response'}</div>` +
               (stat.loc ? `<div class="text-gray-400">Location: ${stat.loc}</div>` : '') +
               `<div><span class="inline-block w-2.5 h-2.5 mr-1 bg-[#38bdf8] rounded-full"></span>Avg RTT: <strong>${stat.avg >= 0 ? stat.avg.toFixed(2) + ' ms' : 'N/A'}</strong></div>` +
               `<div><span class="inline-block w-2.5 h-2.5 mr-1 bg-[#00fea8] rounded-sm"></span>Min / Max: ${stat.best >= 0 ? stat.best.toFixed(2) : 'N/A'} / ${stat.wrst >= 0 ? stat.wrst.toFixed(2) : 'N/A'} ms</div>` +
               `<div>StDev: ${stat.stDev >= 0 ? stat.stDev.toFixed(2) + ' ms' : 'N/A'}</div>` +
               `<div><span class="inline-block w-2.5 h-2.5 mr-1 bg-[#ef4444] rounded-sm"></span>Loss Rate: <strong>${stat.lossRate.toFixed(1)}% (${stat.snt - Math.round(stat.snt * (1 - stat.lossRate / 100))}/${stat.snt})</strong></div>` +
               `</div>`;
      },
    },
    grid: {
      left: '6%',
      right: '6%',
      top: 45,
      bottom: 60,
    },
    legend: {
      top: 15,
      data: [
        $_('Ping.Avg') || 'Avg RTT (ms)',
        'RTT Range (Best-Wrst)',
        $_('Ping.Loss') || 'Loss (%)',
      ],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'category',
      data: hops,
      axisLabel: {
        color: '#ccc',
        fontSize: 9,
        rotate: hops.length > 10 ? 30 : 0,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'RTT (ms)',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        splitLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.1)',
          },
        },
      },
      {
        type: 'value',
        name: $_('Ping.Loss') || 'Loss (%)',
        min: 0,
        max: 100,
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          formatter: '{value}%',
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
    ],
    series: [
      {
        name: 'RTT Base',
        type: 'line',
        stack: 'rttRange',
        symbol: 'none',
        lineStyle: { opacity: 0 },
        data: minData,
      },
      {
        name: 'RTT Range (Best-Wrst)',
        type: 'line',
        stack: 'rttRange',
        symbol: 'none',
        color: '#38bdf8',
        lineStyle: { opacity: 0 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(56, 189, 248, 0.5)' },
            { offset: 1, color: 'rgba(56, 189, 248, 0.1)' },
          ]),
        },
        data: rangeData,
      },
      {
        name: $_('Ping.Avg') || 'Avg RTT (ms)',
        type: 'line',
        showSymbol: true,
        symbolSize: 6,
        color: '#00fea8',
        itemStyle: { color: '#00fea8' },
        lineStyle: { width: 2, color: '#00fea8' },
        data: avgData,
      },
      {
        name: $_('Ping.Loss') || 'Loss (%)',
        type: 'bar',
        yAxisIndex: 1,
        barWidth: '35%',
        data: lossData,
      },
    ],
  };

  chart.setOption(option);
  chart.resize();
  return chart;
};
