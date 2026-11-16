<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Select,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showSFlowIFCounter,
    showSFlowCpuCounter,
    showSFlowMemCounter,
    showSFlowDiskCounter,
    showSFlowNetCounter,
  } from "./chart/sflowCounter";
  import { showLogHeatmap } from "./chart/eventlog";
  import {
    getTableLang,
    renderBytes,
    renderCount,
    renderSpeed,
    renderTime,
  } from "./common";
  import { _ } from "svelte-i18n";
  import {
    GetMapConf,
    LLMExplainSFlowCounterReport,
  } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import DataTable from "datatables.net-dt";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let chart: any = undefined;
  let activeTable: any = undefined;
  let tab: string = "heatmap";

  const ifCounters = new Map<string, any[]>();
  const cpuCounters = new Map<string, any[]>();
  const memCounters = new Map<string, any[]>();
  const diskCounters = new Map<string, any[]>();
  const netCounters = new Map<string, any[]>();

  let ifCounterSrc = "";
  let ifCounterSrcList: any[] = [];
  let cpuCounterSrc = "";
  let cpuCounterSrcList: any[] = [];
  let memCounterSrc = "";
  let memCounterSrcList: any[] = [];
  let diskCounterSrc = "";
  let diskCounterSrcList: any[] = [];
  let netCounterSrc = "";
  let netCounterSrcList: any[] = [];

  let kpi = {
    totalSamples: 0,
    uniqueHosts: 0,
    uniqueIfs: 0,
    maxIfBps: 0,
    maxIfSrc: "-",
    maxCpuPct: 0,
    maxCpuSrc: "-",
  };

  let summaryRows: any[] = [];
  let ifRows: any[] = [];
  let cpuRows: any[] = [];
  let memRows: any[] = [];
  let diskRows: any[] = [];

  const destroyTable = () => {
    if (activeTable) {
      try {
        activeTable.destroy();
      } catch {
        // ignore
      }
      activeTable = undefined;
    }
  };

  const buildSummaries = () => {
    let maxIfBps = 0;
    let maxIfSrc = "-";
    let maxCpuPct = 0;
    let maxCpuSrc = "-";

    // I/F summary
    const newIfRows: any[] = [];
    ifCounters.forEach((arr, key) => {
      const sorted = [...arr].sort((a, b) => a.Time - b.Time);
      let maxRx = 0;
      let maxTx = 0;
      for (let i = 1; i < sorted.length; i++) {
        const dt = (sorted[i].Time - sorted[i - 1].Time) / 1e9;
        if (dt <= 0) continue;
        const rx = (sorted[i].InOctets - sorted[i - 1].InOctets) / dt;
        const tx = (sorted[i].OutOctets - sorted[i - 1].OutOctets) / dt;
        if (rx >= 0 && rx > maxRx) maxRx = rx;
        if (tx >= 0 && tx > maxTx) maxTx = tx;
      }
      const peak = Math.max(maxRx, maxTx);
      if (peak > maxIfBps) {
        maxIfBps = peak;
        maxIfSrc = key;
      }
      newIfRows.push({
        src: key,
        samples: sorted.length,
        maxRx,
        maxTx,
        lastTime: sorted[sorted.length - 1]?.Time || 0,
      });
    });
    ifRows = newIfRows.sort((a, b) => Math.max(b.maxRx, b.maxTx) - Math.max(a.maxRx, a.maxTx));

    // CPU summary
    const newCpuRows: any[] = [];
    cpuCounters.forEach((arr, key) => {
      const sorted = [...arr].sort((a, b) => a.Time - b.Time);
      let sumPct = 0;
      let cnt = 0;
      let maxPct = 0;
      for (let i = 1; i < sorted.length; i++) {
        const sys = sorted[i].CPUSys - sorted[i - 1].CPUSys;
        const user = sorted[i].CPUUser - sorted[i - 1].CPUUser;
        const io = sorted[i].CPUWio - sorted[i - 1].CPUWio;
        const idle = sorted[i].CPUIdle - sorted[i - 1].CPUIdle;
        const other =
          sorted[i].CPUIntr -
          sorted[i - 1].CPUIntr +
          (sorted[i].CPUNice - sorted[i - 1].CPUNice) +
          (sorted[i].CPUSoftIntr - sorted[i - 1].CPUSoftIntr) +
          (sorted[i].CPUSteal - sorted[i - 1].CPUSteal);
        const total = sys + user + io + idle + other;
        if (total <= 0) continue;
        const usedPct = (100 * (total - idle)) / total;
        if (usedPct >= 0 && usedPct <= 100) {
          sumPct += usedPct;
          cnt++;
          if (usedPct > maxPct) maxPct = usedPct;
        }
      }
      if (maxPct > maxCpuPct) {
        maxCpuPct = maxPct;
        maxCpuSrc = key;
      }
      newCpuRows.push({
        src: key,
        samples: sorted.length,
        avgPct: cnt > 0 ? sumPct / cnt : 0,
        maxPct,
        lastTime: sorted[sorted.length - 1]?.Time || 0,
      });
    });
    cpuRows = newCpuRows.sort((a, b) => b.maxPct - a.maxPct);

    // Memory summary
    const newMemRows: any[] = [];
    memCounters.forEach((arr, key) => {
      const sorted = [...arr].sort((a, b) => a.Time - b.Time);
      let sumPct = 0;
      let maxPct = 0;
      let totalBytes = 0;
      let cnt = 0;
      for (const m of sorted) {
        if (m.Total > 0) {
          totalBytes = m.Total;
          const pct = (100 * (m.Total - m.Free - m.Buffers - m.Cached)) / m.Total;
          sumPct += pct;
          cnt++;
          if (pct > maxPct) maxPct = pct;
        }
      }
      newMemRows.push({
        src: key,
        samples: sorted.length,
        avgPct: cnt > 0 ? sumPct / cnt : 0,
        maxPct,
        totalBytes,
        lastTime: sorted[sorted.length - 1]?.Time || 0,
      });
    });
    memRows = newMemRows.sort((a, b) => b.maxPct - a.maxPct);

    // Disk summary
    const newDiskRows: any[] = [];
    diskCounters.forEach((arr, key) => {
      const sorted = [...arr].sort((a, b) => a.Time - b.Time);
      let sumPct = 0;
      let maxPct = 0;
      let totalBytes = 0;
      let cnt = 0;
      for (const d of sorted) {
        if (d.Total > 0) {
          totalBytes = d.Total;
          const pct = (100 * (d.Total - d.Free)) / d.Total;
          sumPct += pct;
          cnt++;
          if (pct > maxPct) maxPct = pct;
        }
      }
      newDiskRows.push({
        src: key,
        samples: sorted.length,
        avgPct: cnt > 0 ? sumPct / cnt : 0,
        maxPct,
        totalBytes,
        lastTime: sorted[sorted.length - 1]?.Time || 0,
      });
    });
    diskRows = newDiskRows.sort((a, b) => b.maxPct - a.maxPct);

    // Overall Host x Type summary
    const hostSet = new Set<string>();
    const sumMap = new Map<string, { remote: string; type: string; samples: number; lastTime: number }>();
    for (const e of logs || []) {
      const rem = e.Remote || "-";
      const typ = e.Type || "-";
      const tm = e.Time || 0;
      hostSet.add(rem);
      const k = `${rem}\t${typ}`;
      let se = sumMap.get(k);
      if (!se) {
        se = { remote: rem, type: typ, samples: 0, lastTime: 0 };
        sumMap.set(k, se);
      }
      se.samples++;
      if (tm > se.lastTime) se.lastTime = tm;
    }
    summaryRows = Array.from(sumMap.values()).sort((a, b) => b.samples - a.samples);

    kpi = {
      totalSamples: (logs || []).length,
      uniqueHosts: hostSet.size,
      uniqueIfs: ifCounters.size,
      maxIfBps,
      maxIfSrc,
      maxCpuPct,
      maxCpuSrc,
    };
  };

  const renderPct = (v: number, type: string) => {
    if (type === "sort") return v;
    return v.toFixed(1) + "%";
  };

  const initTable = async (t: string) => {
    await tick();
    if (t === "heatmap" && document.getElementById("sflowSummaryTable")) {
      activeTable = new DataTable("#sflowSummaryTable", {
        destroy: true,
        paging: summaryRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: summaryRows,
        order: [[2, "desc"]],
        columns: [
          { data: "remote", title: $_("SFlowCounterReport.Host") || "Host" },
          { data: "type", title: $_("SFlowCounterReport.CounterType") || "Type" },
          {
            data: "samples",
            title: $_("SFlowCounterReport.Samples") || "Samples",
            render: renderCount,
          },
          {
            data: "lastTime",
            title: $_("SFlowCounterReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
    } else if (t === "ifCounter_bps" || t === "ifCounter_pps") {
      const tableId = t === "ifCounter_bps" ? "sflowIfTable" : "sflowIfPpsTable";
      const el = document.getElementById(tableId);
      if (!el) return;
      activeTable = new DataTable(`#${tableId}`, {
        destroy: true,
        paging: ifRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: ifRows,
        order: [[2, "desc"]],
        columns: [
          { data: "src", title: $_("SFlowCounterReport.Interface") || "Host:Index" },
          {
            data: "samples",
            title: $_("SFlowCounterReport.Samples") || "Samples",
            render: renderCount,
          },
          {
            data: "maxRx",
            title: $_("SFlowCounterReport.MaxRx") || "Max Rx",
            render: renderSpeed,
          },
          {
            data: "maxTx",
            title: $_("SFlowCounterReport.MaxTx") || "Max Tx",
            render: renderSpeed,
          },
          {
            data: "lastTime",
            title: $_("SFlowCounterReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
      activeTable.on("click", "tbody tr", function (this: any) {
        const rowData = activeTable.row(this).data();
        if (rowData && rowData.src) {
          ifCounterSrc = rowData.src;
          const mode = t === "ifCounter_bps" ? "bps" : "pps";
          chart = showSFlowIFCounter(t, ifCounters.get(ifCounterSrc) || [], mode);
        }
      });
    } else if (t === "cpuCounter" && document.getElementById("sflowCpuTable")) {
      activeTable = new DataTable("#sflowCpuTable", {
        destroy: true,
        paging: cpuRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: cpuRows,
        order: [[3, "desc"]],
        columns: [
          { data: "src", title: $_("SFlowCounterReport.Host") || "Host" },
          {
            data: "samples",
            title: $_("SFlowCounterReport.Samples") || "Samples",
            render: renderCount,
          },
          {
            data: "avgPct",
            title: $_("SFlowCounterReport.AvgUsage") || "Avg %",
            render: renderPct,
          },
          {
            data: "maxPct",
            title: $_("SFlowCounterReport.MaxUsage") || "Max %",
            render: renderPct,
          },
          {
            data: "lastTime",
            title: $_("SFlowCounterReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
      activeTable.on("click", "tbody tr", function (this: any) {
        const rowData = activeTable.row(this).data();
        if (rowData && rowData.src) {
          cpuCounterSrc = rowData.src;
          chart = showSFlowCpuCounter("cpuCounter", cpuCounters.get(cpuCounterSrc) || []);
        }
      });
    } else if (t === "memCounter" && document.getElementById("sflowMemTable")) {
      activeTable = new DataTable("#sflowMemTable", {
        destroy: true,
        paging: memRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: memRows,
        order: [[3, "desc"]],
        columns: [
          { data: "src", title: $_("SFlowCounterReport.Host") || "Host" },
          {
            data: "samples",
            title: $_("SFlowCounterReport.Samples") || "Samples",
            render: renderCount,
          },
          {
            data: "avgPct",
            title: $_("SFlowCounterReport.AvgUsage") || "Avg %",
            render: renderPct,
          },
          {
            data: "maxPct",
            title: $_("SFlowCounterReport.MaxUsage") || "Max %",
            render: renderPct,
          },
          {
            data: "totalBytes",
            title: $_("SFlowCounterReport.TotalSize") || "Total",
            render: renderBytes,
          },
        ],
      });
      activeTable.on("click", "tbody tr", function (this: any) {
        const rowData = activeTable.row(this).data();
        if (rowData && rowData.src) {
          memCounterSrc = rowData.src;
          chart = showSFlowMemCounter("memCounter", memCounters.get(memCounterSrc) || []);
        }
      });
    } else if (t === "diskCounter" && document.getElementById("sflowDiskTable")) {
      activeTable = new DataTable("#sflowDiskTable", {
        destroy: true,
        paging: diskRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: diskRows,
        order: [[3, "desc"]],
        columns: [
          { data: "src", title: $_("SFlowCounterReport.Host") || "Host" },
          {
            data: "samples",
            title: $_("SFlowCounterReport.Samples") || "Samples",
            render: renderCount,
          },
          {
            data: "avgPct",
            title: $_("SFlowCounterReport.AvgUsage") || "Avg %",
            render: renderPct,
          },
          {
            data: "maxPct",
            title: $_("SFlowCounterReport.MaxUsage") || "Max %",
            render: renderPct,
          },
          {
            data: "totalBytes",
            title: $_("SFlowCounterReport.TotalSize") || "Total",
            render: renderBytes,
          },
        ],
      });
      activeTable.on("click", "tbody tr", function (this: any) {
        const rowData = activeTable.row(this).data();
        if (rowData && rowData.src) {
          diskCounterSrc = rowData.src;
          chart = showSFlowDiskCounter("diskCounter", diskCounters.get(diskCounterSrc) || []);
        }
      });
    }
  };

  const onOpen = async () => {
    destroyTable();
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    ifCounters.clear();
    cpuCounters.clear();
    memCounters.clear();
    diskCounters.clear();
    netCounters.clear();
    (logs || []).forEach((e: any) => {
      const d = JSON.parse(e.Data);
      if (!d) {
        return;
      }
      d.Time = e.Time;
      switch (e.Type) {
        case "GenericInterfaceCounter":
          if (d.Index) {
            const k = e.Remote + ":" + d.Index;
            if (ifCounters.has(k)) {
              ifCounters.get(k)!.push(d);
            } else {
              ifCounters.set(k, [d]);
            }
          }
          break;
        case "HostCPUCounter": {
          const k = e.Remote;
          if (cpuCounters.has(k)) {
            cpuCounters.get(k)!.push(d);
          } else {
            cpuCounters.set(k, [d]);
          }
          break;
        }
        case "HostMemoryCounter": {
          const k = e.Remote;
          if (memCounters.has(k)) {
            memCounters.get(k)!.push(d);
          } else {
            memCounters.set(k, [d]);
          }
          break;
        }
        case "HostDiskCounter": {
          const k = e.Remote;
          if (diskCounters.has(k)) {
            diskCounters.get(k)!.push(d);
          } else {
            diskCounters.set(k, [d]);
          }
          break;
        }
        case "HostNetCounter": {
          const k = e.Remote;
          if (netCounters.has(k)) {
            netCounters.get(k)!.push(d);
          } else {
            netCounters.set(k, [d]);
          }
          break;
        }
      }
    });

    const toSelectList = (m: Map<string, any[]>) => {
      const list: any[] = [];
      let first = "";
      m.forEach((_, k) => {
        list.push({ name: k, value: k });
        if (!first) first = k;
      });
      return { list, first };
    };

    const ifRes = toSelectList(ifCounters);
    ifCounterSrcList = ifRes.list;
    ifCounterSrc = ifRes.first;

    const cpuRes = toSelectList(cpuCounters);
    cpuCounterSrcList = cpuRes.list;
    cpuCounterSrc = cpuRes.first;

    const memRes = toSelectList(memCounters);
    memCounterSrcList = memRes.list;
    memCounterSrc = memRes.first;

    const diskRes = toSelectList(diskCounters);
    diskCounterSrcList = diskRes.list;
    diskCounterSrc = diskRes.first;

    const netRes = toSelectList(netCounters);
    netCounterSrcList = netRes.list;
    netCounterSrc = netRes.first;

    buildSummaries();
    showHeatmap();
  };

  const showHeatmap = async () => {
    tab = "heatmap";
    destroyTable();
    await tick();
    chart = showLogHeatmap("heatmap", logs);
    await initTable("heatmap");
    if (chart) chart.resize();
  };

  const showIFCounter = async (type: string) => {
    tab = "ifCounter_" + type;
    destroyTable();
    await tick();
    chart = showSFlowIFCounter(tab, ifCounters.get(ifCounterSrc) || [], type);
    await initTable(tab);
    if (chart) chart.resize();
  };

  const showCpuCounter = async () => {
    tab = "cpuCounter";
    destroyTable();
    await tick();
    chart = showSFlowCpuCounter("cpuCounter", cpuCounters.get(cpuCounterSrc) || []);
    await initTable("cpuCounter");
    if (chart) chart.resize();
  };

  const showMemCounter = async () => {
    tab = "memCounter";
    destroyTable();
    await tick();
    chart = showSFlowMemCounter("memCounter", memCounters.get(memCounterSrc) || []);
    await initTable("memCounter");
    if (chart) chart.resize();
  };

  const showDiskCounter = async () => {
    tab = "diskCounter";
    destroyTable();
    await tick();
    chart = showSFlowDiskCounter("diskCounter", diskCounters.get(diskCounterSrc) || []);
    await initTable("diskCounter");
    if (chart) chart.resize();
  };

  const showNetCounter = async () => {
    tab = "netCounter";
    destroyTable();
    await tick();
    chart = showSFlowNetCounter("netCounter", netCounters.get(netCounterSrc) || []);
    if (chart) chart.resize();
  };

  const close = () => {
    destroyTable();
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  $: if (show) {
    onOpen();
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[90vh]"
>
  <div class="flex flex-col space-y-3">
    <Tabs style="underline">
      <TabItem
        open
        onclick={() => {
          showHeatmap();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBox} size={1} />
            {$_("TrapReport.Heatmap")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3">
          <!-- KPI Summary Cards -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiCounter} size={0.9} class="text-cyan-500" />
                {$_("SFlowCounterReport.TotalSamples") || "総サンプル数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.totalSamples.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiServerNetwork} size={0.9} class="text-blue-500" />
                {$_("SFlowCounterReport.UniqueHosts") || "監視ホスト / I/F数"}
              </div>
              <div class="mt-1 text-2xl font-bold">
                {kpi.uniqueHosts.toLocaleString()}
                <span class="text-xs font-normal text-gray-400">
                  / {kpi.uniqueIfs.toLocaleString()} I/F
                </span>
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiTrafficCone} size={0.9} class="text-emerald-500" />
                {$_("SFlowCounterReport.PeakIfBps") || "最大I/F帯域"}
              </div>
              <div class="mt-1 text-xl font-bold text-emerald-500 dark:text-emerald-400">
                {renderSpeed(kpi.maxIfBps, "")}
              </div>
              <div class="text-[11px] text-gray-400 truncate" title={kpi.maxIfSrc}>
                {kpi.maxIfSrc}
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiGauge} size={0.9} class="text-amber-500" />
                {$_("SFlowCounterReport.MaxCpu") || "最大CPU使用率"}
              </div>
              <div
                class="mt-1 text-2xl font-bold {kpi.maxCpuPct >= 80
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-amber-500 dark:text-amber-400'}"
              >
                {kpi.maxCpuPct.toFixed(1)}%
              </div>
              <div class="text-[11px] text-gray-400 truncate" title={kpi.maxCpuSrc}>
                {kpi.maxCpuSrc}
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
            <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
              <div id="heatmap" class="w-full h-[380px]"></div>
            </div>
            <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
              <div>
                <table id="sflowSummaryTable" class="display compact w-full text-xs text-left text-gray-300"></table>
              </div>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showIFCounter("bps");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiTrafficCone} size={1} />
            I/F BPS
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="ifCounter_bps" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="sflowIfTable" class="display compact w-full text-xs text-left text-gray-300 cursor-pointer"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showIFCounter("pps");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiTrafficCone} size={1} />
            I/F PPS
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="ifCounter_pps" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="sflowIfPpsTable" class="display compact w-full text-xs text-left text-gray-300 cursor-pointer"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showCpuCounter();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiGauge} size={1} />
            CPU
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="cpuCounter" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="sflowCpuTable" class="display compact w-full text-xs text-left text-gray-300 cursor-pointer"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showMemCounter();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiMemory} size={1} />
            Memory
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="memCounter" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="sflowMemTable" class="display compact w-full text-xs text-left text-gray-300 cursor-pointer"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showDiskCounter();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiHarddisk} size={1} />
            Disk
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="diskCounter" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-6 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="sflowDiskTable" class="display compact w-full text-xs text-left text-gray-300 cursor-pointer"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showNetCounter();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiNetwork} size={1} />
            Network
          </div>
        {/snippet}
        <div id="netCounter" class="w-full h-[460px]"></div>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if tab == "ifCounter_bps"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={ifCounterSrcList}
          bind:value={ifCounterSrc}
          size="sm"
          onchange={() => {
            showIFCounter("bps");
          }}
        />
      {/if}
      {#if tab == "ifCounter_pps"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={ifCounterSrcList}
          bind:value={ifCounterSrc}
          size="sm"
          onchange={() => {
            showIFCounter("pps");
          }}
        />
      {/if}
      {#if tab == "cpuCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={cpuCounterSrcList}
          bind:value={cpuCounterSrc}
          size="sm"
          onchange={showCpuCounter}
        />
      {/if}
      {#if tab == "memCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={memCounterSrcList}
          bind:value={memCounterSrc}
          size="sm"
          onchange={showMemCounter}
        />
      {/if}
      {#if tab == "diskCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={diskCounterSrcList}
          bind:value={diskCounterSrc}
          size="sm"
          onchange={showDiskCounter}
        />
      {/if}
      {#if tab == "netCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={netCounterSrcList}
          bind:value={netCounterSrc}
          size="sm"
          onchange={showNetCounter}
        />
      {/if}
      {#if hasAI}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={() => (showAIReport = true)}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_("ReportAI.AIExplain")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("TrapReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`sflow_counter_${tab}_ai_explanation`}
  analyzeFunc={() => LLMExplainSFlowCounterReport(logs || [], tab)}
/>
