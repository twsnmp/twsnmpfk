<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton, Modal, Spinner } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import { GetMonitorDatas, GetSystemInfo, Backup } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, renderBytes, renderSpeed } from "./common";
  import {
    showMonitorResChart,
    showMonitorNetChart,
    showMonitorForecastChart,
    resizeMonitorChart,
  } from "./chart/system";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { backend, main } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";
  import { currentLang } from "../i18n/i18n";

  type TabMode = "status" | "monitor";

  let activeTab = $state<TabMode>("status");
  let sysInfo = $state<main.SystemInfo | null>(null);
  let logs = $state<backend.MonitorDataEnt[]>([]);
  let table: any = undefined;
  let showLoading = $state(false);
  let showForecast = $state(false);

  const latestSnapshot = $derived.by<backend.MonitorDataEnt | null>(() => {
    if (logs.length === 0) return null;
    return logs[0];
  });

  const formatReceiverPort = (port?: string) => {
    if (!port || port === "-") return $_("System.NotConfigured");
    return port;
  };

  const getTimezoneLabel = (timeStr?: string) => {
    if (!timeStr) return "";
    const match = timeStr.match(/([+-]\d{2}:\d{2}|Z)$/);
    const offset = match ? (match[1] === "Z" ? "UTC" : match[1]) : "";
    const tzName = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (tzName && offset) {
      return `${$_("System.TZ")}: ${tzName} (${offset})`;
    } else if (offset) {
      return `${$_("System.TZ")}: ${offset}`;
    }
    return tzName ? `${$_("System.TZ")}: ${tzName}` : "";
  };

  const renderPercent = (val: number | undefined) => {
    if (val === undefined || val === null || isNaN(val)) return "0.00%";
    return `${val.toFixed(2)}%`;
  };

  const getReceiverCardClass = (recv?: main.ReceiverStatusInfo) => {
    if (recv?.status === "error") {
      return "border-rose-300 dark:border-rose-800/80 bg-rose-50/50 dark:bg-rose-950/20 shadow-sm shadow-rose-500/10";
    }
    if (recv?.status === "running") {
      return "border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-slate-950/60";
    }
    return "border-slate-200/80 dark:border-slate-800/60 bg-slate-50/60 dark:bg-slate-950/40 opacity-80";
  };

  const showTable = async () => {
    await tick();
    const el = document.getElementById("systemTable");
    if (!el) {
      return;
    }
    if (table && DataTable.isDataTable("#systemTable")) {
      table.destroy();
      table = undefined;
    }
    el.innerHTML = "";
    table = new DataTable("#systemTable", {
      destroy: true,
      stateSave: true,
      columns: getColumns(),
      paging: false,
      searching: false,
      info: false,
      scrollY: "19vh",
      data: logs,
      order: [[0, "desc"]],
      language: getTableLang(),
    });
  };

  const showChart = async () => {
    await tick();
    if (!document.getElementById("resChart") || !document.getElementById("netChart")) {
      return;
    }
    showMonitorResChart("resChart", logs);
    showMonitorNetChart("netChart", logs);
  };

  const refresh = async () => {
    showLoading = true;
    try {
      const [info, mon] = await Promise.all([
        GetSystemInfo().catch(() => null),
        GetMonitorDatas().catch(() => []),
      ]);
      if (info) sysInfo = info;
      const rawLogs = mon || [];
      logs = [...rawLogs].reverse();

      if (activeTab === "monitor") {
        await tick();
        await showTable();
        await showChart();
      }
    } catch (e) {
      console.error(e);
    } finally {
      showLoading = false;
    }
  };

  const setTab = async (tab: TabMode) => {
    activeTab = tab;
    if (tab === "monitor") {
      await tick();
      await showTable();
      await showChart();
    }
  };

  const renderPer = (v: any, t: any) => {
    if (t == "sort") {
      return v;
    }
    return v.toFixed(2) + "%";
  };

  const getColumns = () => [
    {
      data: "Time",
      title: $_("System.Time"),
      width: "10%",
      render: renderTime,
    },
    {
      data: "CPU",
      title: "CPU",
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "Mem",
      title: $_("System.Memory"),
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "MyCPU",
      title: "My CPU",
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "MyMem",
      title: "My" + $_("System.Memory"),
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "Swap",
      title: "Swap",
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "Disk",
      title: $_("System.Disk"),
      width: "6%",
      render: renderPer,
      className: "dt-right",
    },
    {
      data: "Load",
      title: $_("System.Load"),
      width: "6%",
      render: (v: any) => v.toFixed(2),
      className: "dt-right",
    },
    {
      data: "Net",
      title: $_("System.Net"),
      width: "7%",
      render: renderSpeed,
      className: "dt-right",
    },
    {
      data: "Conn",
      title: $_("System.Conn"),
      width: "6%",
      className: "dt-right",
    },
    {
      data: "Proc",
      title: $_("System.Proc"),
      width: "6%",
      className: "dt-right",
    },
    {
      data: "NumGoroutine",
      title: $_("System.Goroutine"),
      width: "6%",
      className: "dt-right",
    },
    {
      data: "HeapAlloc",
      title: "Heap",
      width: "7%",
      render: renderBytes,
      className: "dt-right",
    },
    {
      data: "Sys",
      title: "Sys",
      width: "7%",
      render: renderBytes,
      className: "dt-right",
    },
    {
      data: "DBSize",
      title: $_("System.DBSize"),
      width: "7%",
      render: renderBytes,
      className: "dt-right",
    },
  ];

  onMount(() => {
    refresh();
  });

  let prevLang = "";
  $effect(() => {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && activeTab === "monitor") {
      showTable();
      showChart();
    }
    prevLang = lang;
  });

  const backup = async () => {
    Backup();
  };

  const forecast = async () => {
    showForecast = true;
    await tick();
    showMonitorForecastChart("forecast", logs);
  };

  const resizeChart = () => {
    resizeMonitorChart(showForecast);
  };
</script>

<svelte:window onresize={resizeChart} />

<div class="flex flex-col max-h-[calc(100vh-50px)] overflow-y-auto pb-2">
  <!-- Top Navigation & Action Header -->
  <header class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 px-4 py-2.5 shrink-0 shadow-xs">
    <!-- Title & Tabs -->
    <div class="flex items-center gap-6">
      <div>
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiServer} size={0.9} class="text-cyan-500 dark:text-cyan-400" />
          <h1 class="text-sm font-bold text-slate-900 dark:text-slate-100 tracking-wide">{$_('System.Title')}</h1>
        </div>
        <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">{$_('System.Subtitle')}</p>
      </div>

      <!-- Navigation Tabs -->
      <div class="flex items-center rounded-xl bg-slate-100 dark:bg-slate-950 p-1 border border-slate-200 dark:border-slate-800">
        <button
          type="button"
          onclick={() => setTab("status")}
          class="flex items-center gap-1.5 rounded-lg px-3 py-1 text-xs font-semibold transition-all cursor-pointer {activeTab === 'status' ? 'bg-gradient-to-r from-cyan-600 to-cyan-500 text-white shadow-md shadow-cyan-600/30' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'}"
        >
          <Icon path={icons.mdiHeartPulse} size={0.7} />
          <span>{$_('System.TabStatus')}</span>
        </button>

        <button
          type="button"
          onclick={() => setTab("monitor")}
          class="flex items-center gap-1.5 rounded-lg px-3 py-1 text-xs font-semibold transition-all cursor-pointer {activeTab === 'monitor' ? 'bg-gradient-to-r from-cyan-600 to-cyan-500 text-white shadow-md shadow-cyan-600/30' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'}"
        >
          <Icon path={icons.mdiChartTimelineVariant} size={0.7} />
          <span>{$_('System.TabMonitor')}</span>
          {#if logs.length > 0}
            <span class="ml-1 rounded-md px-1.5 py-0.2 text-[10px] font-mono {activeTab === 'monitor' ? 'bg-white/25 text-white' : 'bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300'}">
              {logs.length}
            </span>
          {/if}
        </button>
      </div>
    </div>
  </header>

  <!-- Tab 1: System Status -->
  {#if activeTab === "status"}
    <div class="p-4 space-y-4">
      <!-- 6-Card Overview Grid -->
      <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-3">
        <!-- Card 1: Daemon Health -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiServer} size={0.7} />
              <span>{$_('System.DaemonHealth')}</span>
            </span>
            {#if sysInfo?.status === 'ok'}
              <span class="inline-flex items-center gap-1 rounded-full bg-emerald-100 dark:bg-emerald-950/80 border border-emerald-300 dark:border-emerald-800/80 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:text-emerald-400">
                ● ACTIVE
              </span>
            {:else if sysInfo?.status === 'error'}
              <span class="inline-flex items-center gap-1 rounded-full bg-rose-100 dark:bg-rose-950/80 border border-rose-300 dark:border-rose-800/80 px-2 py-0.5 text-[10px] font-semibold text-rose-700 dark:text-rose-400">
                ● ERROR
              </span>
            {:else}
              <span class="inline-flex items-center gap-1 rounded-full bg-amber-100 dark:bg-amber-950/80 border border-amber-300 dark:border-amber-800/80 px-2 py-0.5 text-[10px] font-semibold text-amber-700 dark:text-amber-400">
                ● {sysInfo?.status?.toUpperCase() || 'IDLE'}
              </span>
            {/if}
          </div>
          <div class="my-2 flex items-center gap-2 text-2xl font-bold {sysInfo?.status === 'ok' ? 'text-emerald-600 dark:text-emerald-400' : sysInfo?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-amber-600 dark:text-amber-400'}">
            {#if sysInfo?.status === 'ok'}
              <Icon path={icons.mdiCheckCircle} size={1} />
              <span>HEALTHY</span>
            {:else if sysInfo?.status === 'error'}
              <Icon path={icons.mdiAlertOctagon} size={1} />
              <span>ERROR</span>
            {:else}
              <Icon path={icons.mdiAlertCircle} size={1} />
              <span>{sysInfo?.status?.toUpperCase() || "IDLE"}</span>
            {/if}
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono flex items-center justify-between">
            <span>{$_('System.Uptime')} <span class="text-cyan-600 dark:text-cyan-300 font-semibold">{sysInfo?.uptime || $_('System.Starting')}</span></span>
          </div>
        </div>

        <!-- Card 2: Version -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiHeartPulse} size={0.7} />
              <span>{$_('System.Version')}</span>
            </span>
            {#if sysInfo?.has_update}
              <span class="flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 animate-pulse">
                <Icon path={icons.mdiSparkles} size={0.5} />
                <span>{$_('System.UpdateAvailable')}</span>
              </span>
            {/if}
          </div>
          <div class="my-2 flex items-baseline gap-2">
            <span class="text-xl font-bold font-mono text-slate-900 dark:text-slate-100">
              {sysInfo?.version || "v2.7.0"}
            </span>
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate">
            <span>{sysInfo?.commit ? `(${sysInfo.commit})` : "(none)"}</span>
          </div>
        </div>

        <!-- Card 3: Server Clock -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiClockOutline} size={0.7} />
              <span>{$_('System.ServerTime')}</span>
            </span>
          </div>
          <div class="my-2 text-base font-bold font-mono text-slate-900 dark:text-slate-100 leading-tight">
            {renderTime(latestSnapshot?.Time || Date.now() * 1000 * 1000, "")}
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono">
            {getTimezoneLabel(sysInfo?.time)}
          </div>
        </div>

        <!-- Card 4: CPU & Load -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiCpu64Bit} size={0.7} />
              <span>{$_('System.CPULoad')}</span>
            </span>
            <span class="text-[10px] font-mono text-slate-500 dark:text-slate-400">{sysInfo?.num_cpu || 1} {$_('System.Cores')}</span>
          </div>
          <div class="my-2 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-sky-600 dark:text-sky-400">
              {renderPercent(latestSnapshot?.CPU)}
            </span>
            <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">
              Load {latestSnapshot?.Load?.toFixed(2) ?? "0.00"}
            </span>
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate">
            {$_('System.MyCPU')} <span class="text-cyan-600 dark:text-cyan-300 font-semibold">{renderPercent(latestSnapshot?.MyCPU)}</span>
          </div>
        </div>

        <!-- Card 5: Memory Usage -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiMemory} size={0.7} />
              <span>{$_('System.MemoryUsage')}</span>
            </span>
            <span class="text-[10px] font-mono text-slate-500 dark:text-slate-400">Swap {renderPercent(latestSnapshot?.Swap)}</span>
          </div>
          <div class="my-2 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-purple-600 dark:text-purple-400">
              {renderPercent(latestSnapshot?.Mem)}
            </span>
            <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">
              My {renderPercent(latestSnapshot?.MyMem)}
            </span>
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate">
            Heap: <span class="text-purple-600 dark:text-purple-300 font-semibold">{renderBytes(latestSnapshot?.HeapAlloc || 0, "")}</span>
          </div>
        </div>

        <!-- Card 6: Database & Storage -->
        <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg flex flex-col justify-between">
          <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-medium">
            <span class="flex items-center gap-1.5 text-cyan-600 dark:text-cyan-400">
              <Icon path={icons.mdiDatabase} size={0.7} />
              <span>{$_('System.StorageDB')}</span>
            </span>
            <span class="text-[10px] font-mono text-emerald-600 dark:text-emerald-400">bbolt</span>
          </div>
          <div class="my-2 flex items-baseline gap-2">
            <span class="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
              {renderBytes(latestSnapshot?.DBSize || 0, "")}
            </span>
          </div>
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate">
            Disk: <span class="text-amber-600 dark:text-amber-300 font-semibold">{renderPercent(latestSnapshot?.Disk)}</span> {$_('System.DiskUsed')}
          </div>
        </div>
      </div>

      <!-- Protocol Servers Overview Card -->
      <div class="rounded-2xl border border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/80 p-4 shadow-sm dark:shadow-lg space-y-3">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-2.5">
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiShieldCheck} size={0.8} class="text-cyan-600 dark:text-cyan-400" />
            <h3 class="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider">{$_('System.ServerHealth')}</h3>
          </div>
          <div class="flex items-center gap-4 text-xs font-mono text-slate-500 dark:text-slate-400">
            <span>{$_('System.MonitoredNodes')} <strong class="text-cyan-600 dark:text-cyan-400">{sysInfo?.node_count ?? 0}</strong></span>
            <span>{$_('System.ActivePolling')} <strong class="text-cyan-600 dark:text-cyan-400">{sysInfo?.poll_count ?? 0}</strong></span>
          </div>
        </div>

        {#snippet statusBadge(recv)}
          {#if recv?.status === 'running'}
            <span class="inline-flex items-center gap-1 rounded-full bg-emerald-100 dark:bg-emerald-950/80 border border-emerald-300 dark:border-emerald-800/60 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:text-emerald-400">
              {$_('System.Running')}
            </span>
          {:else if recv?.status === 'error'}
            <span class="inline-flex items-center gap-1 rounded-full bg-rose-100 dark:bg-rose-950/80 border border-rose-300 dark:border-rose-800/60 px-2 py-0.5 text-[10px] font-semibold text-rose-700 dark:text-rose-400" title={recv?.error || $_('System.ErrorDesc')}>
              {$_('System.Error')}
            </span>
          {:else if recv?.status === 'stopped'}
            <span class="inline-flex items-center gap-1 rounded-full bg-amber-100 dark:bg-amber-950/80 border border-amber-300 dark:border-amber-800/60 px-2 py-0.5 text-[10px] font-semibold text-amber-700 dark:text-amber-400">
              {$_('System.Stopped')}
            </span>
          {:else}
            <span class="inline-flex items-center gap-1 rounded-full bg-slate-100 dark:bg-slate-800/80 border border-slate-300 dark:border-slate-700/60 px-2 py-0.5 text-[10px] font-semibold text-slate-500 dark:text-slate-400">
              {$_('System.Disabled')}
            </span>
          {/if}
        {/snippet}

        {#snippet errorNotice(recv)}
          {#if recv?.error}
            <div class="text-[10px] text-rose-600 dark:text-rose-400 font-medium truncate flex items-center gap-1 mt-0.5 bg-rose-500/10 px-1.5 py-0.5 rounded border border-rose-500/20" title={recv.error}>
              <Icon path={icons.mdiAlertCircle} size={0.5} class="shrink-0" />
              <span class="truncate">{recv.error}</span>
            </div>
          {/if}
        {/snippet}

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-xs">
          <!-- Syslog -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.syslog)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">Syslog</span>
              {@render statusBadge(sysInfo?.receivers?.syslog)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.syslog?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.syslog?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.syslog?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.syslog)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">RFC 3164 / RFC 5424</div>
          </div>

          <!-- SNMP TRAP -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.trap)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">SNMP TRAP</span>
              {@render statusBadge(sysInfo?.receivers?.trap)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.trap?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.trap?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.trap?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.trap)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">SNMP TRAP / InformRequest</div>
          </div>

          <!-- NetFlow / IPFIX -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.netflow)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">NetFlow / IPFIX</span>
              {@render statusBadge(sysInfo?.receivers?.netflow)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.netflow?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.netflow?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.netflow?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.netflow)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">NetFlow v5 / v9 / IPFIX</div>
          </div>

          <!-- MCP Server -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.mcp)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.MCPServer')}</span>
              {@render statusBadge(sysInfo?.receivers?.mcp)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.mcp?.status === 'running' ? 'text-emerald-600 dark:text-emerald-400' : sysInfo?.receivers?.mcp?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.mcp?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.mcp)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.MCPDesc')}</div>
          </div>

          <!-- OpenTelemetry -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.otel)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">OpenTelemetry</span>
              {@render statusBadge(sysInfo?.receivers?.otel)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.otel?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.otel?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.otel?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.otel)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">OTLP Traces / Metrics / Logs</div>
          </div>

          <!-- MQTT Broker -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.mqtt)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.MQTTBroker')}</span>
              {@render statusBadge(sysInfo?.receivers?.mqtt)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.mqtt?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.mqtt?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.mqtt?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.mqtt)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.MQTTDesc')}</div>
          </div>

          <!-- sFlow -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.sflow)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">sFlow</span>
              {@render statusBadge(sysInfo?.receivers?.sflow)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.sflow?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.sflow?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.sflow?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.sflow)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.SFlowDesc')}</div>
          </div>

          <!-- ARP Watch -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.arp)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.ARPWatch')}</span>
              {@render statusBadge(sysInfo?.receivers?.arp)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.arp?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.arp?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {sysInfo?.receivers?.arp?.status === 'running' ? (sysInfo?.receivers?.arp?.range || "Active Scan") : $_('System.Disabled')}
            </div>
            {@render errorNotice(sysInfo?.receivers?.arp)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.ARPDesc')}</div>
          </div>

          <!-- PING (ICMP Engine) -->
          <div class="rounded-xl border border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-slate-950/60 p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.PingEngine')}</span>
              <span class="inline-flex items-center gap-1 rounded-full bg-emerald-100 dark:bg-emerald-950/80 border border-emerald-300 dark:border-emerald-800/60 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:text-emerald-400">{$_('System.Running')}</span>
            </div>
            <div class="font-mono font-semibold text-cyan-600 dark:text-cyan-400">
              Mode: {sysInfo?.ping_mode?.toUpperCase() || "UDP"}
            </div>
            <div class="text-[10px] text-slate-500 dark:text-slate-400">
              {sysInfo?.ping_mode === "icmp" ? $_('System.PingModeIcmp') : $_('System.PingModeUdp')}
            </div>
          </div>

          <!-- SSH Server -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.sshd)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.SSHServer')}</span>
              {@render statusBadge(sysInfo?.receivers?.sshd)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.sshd?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.sshd?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.sshd?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.sshd)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.SSHDesc')}</div>
          </div>

          <!-- TCP Server -->
          <div class="rounded-xl border p-3 space-y-1 hover:border-slate-300 dark:hover:border-slate-700 transition-colors {getReceiverCardClass(sysInfo?.receivers?.tcpd)}">
            <div class="flex items-center justify-between">
              <span class="font-semibold text-slate-800 dark:text-slate-200">{$_('System.TCPServer')}</span>
              {@render statusBadge(sysInfo?.receivers?.tcpd)}
            </div>
            <div class="font-mono font-semibold {sysInfo?.receivers?.tcpd?.status === 'running' ? 'text-cyan-600 dark:text-cyan-400' : sysInfo?.receivers?.tcpd?.status === 'error' ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400 dark:text-slate-500'}">
              {formatReceiverPort(sysInfo?.receivers?.tcpd?.port)}
            </div>
            {@render errorNotice(sysInfo?.receivers?.tcpd)}
            <div class="text-[10px] text-slate-500 dark:text-slate-400">{$_('System.TCPDesc')}</div>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Tab 2: Resource Monitor -->
  {#if activeTab === "monitor"}
    <div class="flex flex-col pb-2">
      <div id="resChart"></div>
      <div id="netChart"></div>
      <div class="mx-2 mt-2">
        <div>
          <table id="systemTable" class="display compact" style="width:99%"></table>
        </div>
      </div>
    </div>
  {/if}

  <div class="flex justify-end space-x-2 mr-2 py-2">
    <GradientButton
      shadow
      type="button"
      color="green"
      onclick={forecast}
      size="xs"
    >
      <Icon path={icons.mdiChartLine} size={1} />
      {$_("System.SizeForecast")}
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      onclick={backup}
      size="xs"
    >
      <Icon path={icons.mdiDatabaseArrowDown} size={1} />
      {$_("System.Backup")}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("System.Reload")}
    </GradientButton>
  </div>
</div>

<!-- Forecast Modal -->
<Modal bind:open={showForecast} size="xl" dismissable={false} outsideclose={false} class="w-full">
  <div id="forecast"></div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={() => {
        showForecast = false;
      }}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("System.Close")}
    </GradientButton>
  </div>
</Modal>

<!-- Loading Modal -->
<Modal bind:open={showLoading} size="sm" dismissable={false} outsideclose={false} class="w-full" transitionParams={{ duration: 0 }}>
  <div>
    <Spinner />
    <span class="ml-2">{$_('System.Loading')}</span>
  </div>
</Modal>

<style>
  #resChart,
  #netChart {
    min-height: 250px;
    height: 25vh;
    width: 98%;
    margin: 0 auto;
  }
  #forecast {
    min-height: 500px;
    height: 65vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
