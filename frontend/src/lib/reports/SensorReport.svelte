<script lang="ts">
  import "../../assets/css/jquery.dataTables.css";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { onMount, tick } from "svelte";
  import { _ } from "svelte-i18n";
  import * as echarts from "echarts";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Modal, Button, GradientButton } from "flowbite-svelte";
  import { formatTimeStr, isDarkMode, exportReportData, echartAction } from "./utils";
  import { getTableLang, renderState, renderTime, renderCount } from "../common";
  import {
    GetSensors,
    DeleteSensors,
    DeleteAllSensors,
    ToggleSensorIgnore,
  } from "../../../wailsjs/go/main/App";
  import type { datastore } from "../../../wailsjs/go/models";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  let sensors = $state<datastore.SensorEnt[]>([]);
  let internalLoading = $state(false);

  let filterHost = $state("");
  let filterType = $state("");
  let filterState = $state("all");

  let table: any = undefined;
  let selected = $state<datastore.SensorEnt | null>(null);
  let selectedCount = $state(0);

  let showInfoModal = $state(false);
  let showDeleteModal = $state(false);
  let showResetModal = $state(false);

  // Available data flags for telemetry charts
  const hasStats = $derived(
    Boolean(
      selected?.Stats &&
        selected.Stats.length > 0 &&
        selected.Stats.some((st) => st.PS > 0 || st.Count > 0)
    )
  );
  const hasCpuMem = $derived(
    Boolean(
      selected?.Monitors &&
        selected.Monitors.length > 0 &&
        selected.Monitors.some((m) => m.CPU > 0 || m.Mem > 0 || m.Load > 0)
    )
  );
  const hasNet = $derived(
    Boolean(
      selected?.Monitors &&
        selected.Monitors.length > 0 &&
        selected.Monitors.some((m) => m.TxSpeed > 0 || m.RxSpeed > 0)
    )
  );
  const hasProc = $derived(
    Boolean(
      selected?.Monitors &&
        selected.Monitors.length > 0 &&
        selected.Monitors.some((m) => m.Process > 0)
    )
  );
  const chartCount = $derived((hasStats ? 1 : 0) + (hasCpuMem ? 1 : 0) + (hasNet ? 1 : 0) + (hasProc ? 1 : 0));

  const columns = [
    {
      data: "State",
      title: $_("report.colState") || "ステート",
      width: "10%",
      render: renderState,
    },
    {
      data: "Host",
      title: $_("report.colHost") || "送信元",
      width: "15%",
    },
    {
      data: "Type",
      title: $_("report.colType") || "種別",
      width: "15%",
    },
    {
      data: "Param",
      title: $_("report.colParam") || "パラメータ",
      width: "20%",
    },
    {
      data: "Total",
      title: $_("report.colTotal") || "受信数",
      width: "10%",
      render: renderCount,
    },
    {
      data: "Send",
      title: $_("report.colSend") || "送信数",
      width: "10%",
      render: renderCount,
    },
    {
      data: "LastTime",
      title: $_("report.colLastTime") || "最終確認",
      width: "15%",
      render: renderTime,
    },
    {
      data: "Ignore",
      title: "監視",
      width: "5%",
      render: (v: boolean) =>
        v
          ? '<span class="text-gray-400">無視</span>'
          : '<span class="text-emerald-500 font-semibold">監視</span>',
    },
  ];

  const filteredSensors = $derived.by(() => {
    let res = sensors;
    if (filterHost) {
      res = res.filter((s) => s.Host === filterHost);
    }
    if (filterType) {
      res = res.filter((s) => s.Type === filterType);
    }
    if (filterState !== "all") {
      res = res.filter((s) => s.State === filterState);
    }
    return res;
  });

  const showTable = async () => {
    await tick();
    if (!document.getElementById("sensorTable")) return;

    if (table && DataTable.isDataTable("#sensorTable")) {
      selectedCount = 0;
      selected = null;
      table.clear();
      table.rows.add(filteredSensors);
      table.draw();
      return;
    }

    selectedCount = 0;
    selected = null;
    table = new DataTable("#sensorTable", {
      destroy: true,
      columns: columns,
      data: filteredSensors,
      stateSave: true,
      order: [[6, "desc"]],
      pageLength: window.innerHeight > 800 ? 25 : 10,
      language: getTableLang(),
      select: {
        style: "single",
      },
    });

    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
      const selData = table.rows({ selected: true }).data();
      selected = selData && selData.length > 0 ? selData[0] : null;
    });

    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
      const selData = table.rows({ selected: true }).data();
      selected = selData && selData.length > 0 ? selData[0] : null;
    });

    table.on("dblclick", "tbody tr", () => {
      if (selected) {
        openSensorInfo(selected);
      }
    });
  };

  $effect(() => {
    // Redraw table when filters change
    if (filteredSensors) {
      showTable();
    }
  });

  const loadSensors = async () => {
    internalLoading = true;
    try {
      const data = await GetSensors();
      sensors = (data || []) as datastore.SensorEnt[];
      await showTable();
    } catch (err) {
      console.error("Failed to load sensors:", err);
    } finally {
      internalLoading = false;
    }
  };

  export const refresh = () => {
    loadSensors();
  };

  onMount(() => {
    loadSensors();
    return () => {
      if (table && DataTable.isDataTable("#sensorTable")) {
        table.destroy();
        table = undefined;
      }
    };
  });

  const uniqueHosts = $derived.by(() => {
    return [...new Set(sensors.map((s) => s.Host).filter(Boolean))].sort();
  });

  const uniqueTypes = $derived.by(() => {
    return [...new Set(sensors.map((s) => s.Type).filter(Boolean))].sort();
  });

  const stats = $derived.by(() => {
    const total = sensors.length;
    let running = 0;
    let alerts = 0;
    let totalEvents = 0;

    for (const s of sensors) {
      if (s.State === "normal") running++;
      else if (s.State === "warn" || s.State === "low" || s.State === "high") alerts++;
      totalEvents += s.Total || 0;
    }
    return { total, running, alerts, totalEvents };
  });

  const statsChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!selected?.Stats || selected.Stats.length === 0) return null;
    const dark = isDarkMode();
    const times = selected.Stats.map((st) => formatTimeStr(st.Time));
    const pps = selected.Stats.map((st) => st.PS);
    const counts = selected.Stats.map((st) => st.Count);
    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      legend: { top: 0, data: ["Packets/sec (PS)", "Count"], textStyle: { color: dark ? "#cbd5e1" : "#475569", fontSize: 11 } },
      grid: { top: 34, left: "3%", right: "3%", bottom: 8, containLabel: true },
      xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      yAxis: [
        { type: "value", name: "PS", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        { type: "value", name: "Count", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 }, splitLine: { show: false } },
      ],
      series: [
        { name: "Packets/sec (PS)", type: "line", data: pps, smooth: true, lineStyle: { color: "#06b6d4", width: 2 } },
        { name: "Count", type: "bar", yAxisIndex: 1, data: counts, itemStyle: { color: "rgba(59, 130, 246, 0.4)" } },
      ],
    };
  });

  const cpuMemChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!selected?.Monitors || selected.Monitors.length === 0) return null;
    const dark = isDarkMode();
    const times = selected.Monitors.map((m) => formatTimeStr(m.Time));
    const cpus = selected.Monitors.map((m) => m.CPU);
    const mems = selected.Monitors.map((m) => m.Mem);
    const loads = selected.Monitors.map((m) => m.Load);
    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      legend: { top: 0, data: ["CPU (%)", "Memory (%)", "Load Avg"], textStyle: { color: dark ? "#cbd5e1" : "#475569", fontSize: 11 } },
      grid: { top: 34, left: "3%", right: "3%", bottom: 8, containLabel: true },
      xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      yAxis: { type: "value", max: 100, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      series: [
        { name: "CPU (%)", type: "line", data: cpus, smooth: true, lineStyle: { color: "#f97316", width: 2 } },
        { name: "Memory (%)", type: "line", data: mems, smooth: true, lineStyle: { color: "#8b5cf6", width: 2 } },
        { name: "Load Avg", type: "line", data: loads, smooth: true, lineStyle: { color: "#10b981", width: 1.5, type: "dashed" } },
      ],
    };
  });

  const netChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!selected?.Monitors || selected.Monitors.length === 0) return null;
    const dark = isDarkMode();
    const times = selected.Monitors.map((m) => formatTimeStr(m.Time));
    const tx = selected.Monitors.map((m) => (m.TxSpeed / (1000 * 1000)).toFixed(2));
    const rx = selected.Monitors.map((m) => (m.RxSpeed / (1000 * 1000)).toFixed(2));
    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      legend: { top: 0, data: ["Tx (Mbps)", "Rx (Mbps)"], textStyle: { color: dark ? "#cbd5e1" : "#475569", fontSize: 11 } },
      grid: { top: 34, left: "3%", right: "3%", bottom: 8, containLabel: true },
      xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      yAxis: { type: "value", name: "Mbps", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      series: [
        { name: "Tx (Mbps)", type: "line", data: tx, smooth: true, lineStyle: { color: "#3b82f6", width: 2 } },
        { name: "Rx (Mbps)", type: "line", data: rx, smooth: true, lineStyle: { color: "#10b981", width: 2 } },
      ],
    };
  });

  const procChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!selected?.Monitors || selected.Monitors.length === 0) return null;
    const dark = isDarkMode();
    const times = selected.Monitors.map((m) => formatTimeStr(m.Time));
    const procs = selected.Monitors.map((m) => m.Process);
    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      grid: { top: 30, left: "3%", right: "3%", bottom: 8, containLabel: true },
      xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      yAxis: { type: "value", name: "Procs", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
      series: [
        { name: "Process Count", type: "line", data: procs, smooth: true, lineStyle: { color: "#6366f1", width: 2 }, areaStyle: { color: "rgba(99, 102, 241, 0.2)" } },
      ],
    };
  });

  const openSensorInfo = (s: datastore.SensorEnt) => {
    selected = s;
    showInfoModal = true;
  };

  const handleToggleIgnore = async (s: datastore.SensorEnt) => {
    try {
      await ToggleSensorIgnore(s.ID);
      await loadSensors();
    } catch (err) {
      console.error("Failed to toggle sensor ignore:", err);
    }
  };

  const openDelete = (s: datastore.SensorEnt) => {
    selected = s;
    showDeleteModal = true;
  };

  const confirmDelete = async () => {
    if (!selected) return;
    try {
      await DeleteSensors([selected.ID]);
      showDeleteModal = false;
      await loadSensors();
    } catch (err) {
      console.error("Failed to delete sensor:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    const headers = ["State", "Host", "Type", "Param", "Total Events", "Send", "First Seen", "Last Seen", "Ignored"];
    const rows = sensors.map((s) => [
      s.State,
      s.Host,
      s.Type,
      s.Param,
      s.Total,
      s.Send,
      formatTimeStr(s.FirstTime),
      formatTimeStr(s.LastTime),
      s.Ignore ? "YES" : "NO",
    ]);
    exportReportData(format, "TWSNMP_Sensor_Report", headers, rows);
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    try {
      await DeleteAllSensors();
      await loadSensors();
      onRefresh();
    } catch (err) {
      console.error("Failed to delete all sensors:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiRadioTower} size={0.9} class="text-cyan-500" />
        {$_("report.totalSensors") || "総センサー数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.total}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiCheckCircle} size={0.9} class="text-emerald-500" />
        {$_("report.runningSensors") || "正常稼働"}
      </div>
      <div class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">{stats.running}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiAlertCircle} size={0.9} class="text-rose-500" />
        {$_("report.alertSensors") || "警告・異常"}
      </div>
      <div class="mt-1 text-2xl font-bold text-rose-600 dark:text-rose-400">{stats.alerts}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiCounter} size={0.9} class="text-blue-500" />
        {$_("report.totalSensorEvents") || "総受信メッセージ数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.totalEvents.toLocaleString()}</div>
    </div>
  </div>

  <!-- Filters Row -->
  <div class="flex flex-wrap items-center gap-2 text-xs">
    <!-- Host Filter -->
    <select
      bind:value={filterHost}
      class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 focus:outline-none cursor-pointer"
    >
      <option value="">{$_("report.allHosts") || "すべての送信元"}</option>
      {#each uniqueHosts as h}
        <option value={h}>{h}</option>
      {/each}
    </select>

    <!-- Type Filter -->
    <select
      bind:value={filterType}
      class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 focus:outline-none cursor-pointer"
    >
      <option value="">{$_("report.allTypes") || "すべての種別"}</option>
      {#each uniqueTypes as t}
        <option value={t}>{t}</option>
      {/each}
    </select>

    <!-- State Filter -->
    <select
      bind:value={filterState}
      class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-2 py-1 text-xs text-gray-700 dark:text-gray-200 focus:outline-none cursor-pointer"
    >
      <option value="all">{$_("report.subAllSensors") || "すべてのステート"}</option>
      <option value="normal">{$_("report.stateNormal") || "正常"}</option>
      <option value="warn">{$_("report.stateWarn") || "注意"}</option>
      <option value="low">{$_("report.stateLow") || "軽度"}</option>
      <option value="high">{$_("report.stateHigh") || "重度"}</option>
      <option value="off">{$_("report.stateOff") || "停止"}</option>
    </select>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div>
      <table id="sensorTable" class="display compact" style="width:99%"></table>
    </div>
  </div>

  <!-- Unified Bottom Action Buttons Bar -->
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount === 1 && selected}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={() => selected && openSensorInfo(selected)}
        size="xs"
      >
        <Icon path={icons.mdiChartLine} size={1} />
        {$_("report.sensorDetail") || "テレメトリ詳細"}
      </GradientButton>

      <GradientButton
        shadow
        color="cyan"
        type="button"
        onclick={() => selected && handleToggleIgnore(selected)}
        size="xs"
      >
        <Icon path={selected?.Ignore ? icons.mdiEye : icons.mdiEyeOff} size={1} />
        {selected?.Ignore ? ($_("report.sensorMonitor") || "監視再開") : ($_("report.sensorIgnore") || "監視除外")}
      </GradientButton>

      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={() => selected && openDelete(selected)}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.delete") || "削除"}
      </GradientButton>
    {/if}

    <GradientButton
      shadow
      color="red"
      type="button"
      onclick={handleClear}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("report.btnClear") || "全消去"}
    </GradientButton>

    <GradientButton
      shadow
      color="green"
      type="button"
      onclick={() => handleExport("csv")}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>

    <GradientButton
      shadow
      color="lime"
      type="button"
      onclick={() => handleExport("excel")}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>

    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={loadSensors}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("report.btnReload") || "更新"}
    </GradientButton>
  </div>
</div>

<!-- Sensor Details & Charts Modal -->
<Modal bind:open={showInfoModal} size="xl" dismissable={false} outsideclose={false}>
  {#if selected}
    <div class="space-y-4 p-2 text-gray-800 dark:text-gray-100">
      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiRadioTower} size={1.2} class="text-cyan-500" />
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">
              {selected.Host} ({selected.Type})
            </h2>
            <div class="text-xs text-gray-500 font-mono">ID: {selected.ID} / {selected.Param}</div>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {selected.State === 'normal' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300'}">
            {selected.State}
          </span>
        </div>
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400">{$_("report.colTotal") || "回数"}</div>
          <div class="font-bold font-mono text-sm text-gray-900 dark:text-gray-100 mt-0.5">{selected.Total?.toLocaleString()}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400">{$_("report.colSend") || "送信数"}</div>
          <div class="font-bold font-mono text-sm text-gray-900 dark:text-gray-100 mt-0.5">{selected.Send?.toLocaleString()}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400">{$_("report.firstSeen") || "初回検知日時"}</div>
          <div class="font-mono text-xs text-gray-900 dark:text-gray-100 mt-0.5">{formatTimeStr(selected.FirstTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400">{$_("report.colLastTime") || "最終日時"}</div>
          <div class="font-mono text-xs text-gray-900 dark:text-gray-100 mt-0.5">{formatTimeStr(selected.LastTime)}</div>
        </div>
      </div>

      <!-- Charts Section (only displays charts that actually have data) -->
      {#if chartCount > 0}
        <div class="grid grid-cols-1 {chartCount > 1 ? 'md:grid-cols-2' : ''} gap-4">
          <!-- PPS / Count Chart -->
          {#if hasStats && statsChartOption}
            <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
              <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
                <Icon path={icons.mdiChartLine} size={0.8} class="text-cyan-500" />
                {$_("report.chartPPS") || "受信レート (Packets/sec & Count)"}
              </div>
              <div use:echartAction={statsChartOption} class="h-44 w-full"></div>
            </div>
          {/if}

          <!-- CPU / Memory / Load Chart -->
          {#if hasCpuMem && cpuMemChartOption}
            <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
              <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
                <Icon path={icons.mdiCpu64Bit} size={0.8} class="text-orange-500" />
                {$_("report.chartCpuMem") || "CPU / メモリ使用率 / Load"}
              </div>
              <div use:echartAction={cpuMemChartOption} class="h-44 w-full"></div>
            </div>
          {/if}

          <!-- Network Tx/Rx Chart -->
          {#if hasNet && netChartOption}
            <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
              <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
                <Icon path={icons.mdiNetworkOutline} size={0.8} class="text-blue-500" />
                {$_("report.chartNetwork") || "ネットワーク転送速度 (Tx / Rx Mbps)"}
              </div>
              <div use:echartAction={netChartOption} class="h-44 w-full"></div>
            </div>
          {/if}

          <!-- Process Count Chart -->
          {#if hasProc && procChartOption}
            <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
              <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
                <Icon path={icons.mdiCogOutline} size={0.8} class="text-indigo-500" />
                {$_("report.chartProcess") || "プロセス総数推移"}
              </div>
              <div use:echartAction={procChartOption} class="h-44 w-full"></div>
            </div>
          {/if}
        </div>
      {:else}
        <div class="flex flex-col items-center justify-center p-8 bg-gray-50 dark:bg-gray-800/40 rounded-lg border border-gray-200 dark:border-gray-700 text-center my-2">
          <Icon path={icons.mdiChartTimelineVariant} size={2} class="text-gray-400 mb-2" />
          <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
            {$_("report.noTelemetryData") || "テレメトリ時系列グラフの記録データはまだありません"}
          </div>
        </div>
      {/if}

      <!-- Bottom Close Button (matches standard dialogs in TWSNMP FK) -->
      <div class="flex justify-end space-x-2 pt-2 border-t border-gray-200 dark:border-gray-700">
        <GradientButton shadow type="button" color="teal" onclick={() => (showInfoModal = false)} size="xs">
          <Icon path={icons.mdiCancel} size={1} />
          {$_("common.close") || "閉じる"}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<!-- Delete Sensor Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={showDeleteModal} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {selected ? `センサー ${selected.Host} (${selected.Type}) を削除しますか？` : ""}
    </h3>
    <div class="flex justify-center gap-3">
      <GradientButton shadow color="red" onclick={confirmDelete} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.delete") || "削除"}
      </GradientButton>
      <GradientButton shadow color="teal" onclick={() => (showDeleteModal = false)} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("common.cancel") || "キャンセル"}
      </GradientButton>
    </div>
  </div>
</Modal>

<!-- Reset All Sensors Modal -->
<Modal bind:open={showResetModal} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.sensorDeleteConfirm") || "すべてのセンサー情報を消去しますか？"}
    </h3>
    <div class="flex justify-center gap-3">
      <GradientButton shadow color="red" onclick={confirmReset} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.ok") || "全消去"}
      </GradientButton>
      <GradientButton shadow color="teal" onclick={() => (showResetModal = false)} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("common.cancel") || "キャンセル"}
      </GradientButton>
    </div>
  </div>
</Modal>
