<script lang="ts">
  import "../../assets/css/jquery.dataTables.css";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { onMount, tick } from "svelte";
  import { _ } from "svelte-i18n";
  import * as echarts from "echarts";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Modal, Button, ButtonGroup, GradientButton, Input, Label } from "flowbite-svelte";
  import { formatTimeStr, isDarkMode, exportReportData, echartAction } from "./utils";
  import { getTableLang, renderCount, renderTime } from "../common";
  import {
    GetLogReport,
    ResetLogReport,
    DeleteLogReport,
    SetLogReportName,
  } from "../../../wailsjs/go/main/App";
  import type {
    EnvMonitorEnt,
    PowerMonitorEnt,
    MotionSensorEnt,
  } from "./types";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  type TabType = "env" | "power" | "motion";
  let activeTab = $state<TabType>("env");

  let envs = $state<EnvMonitorEnt[]>([]);
  let powers = $state<PowerMonitorEnt[]>([]);
  let motions = $state<MotionSensorEnt[]>([]);
  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<any>(null);
  let selectedCount = $state(0);

  // Modals state
  let detailModalOpen = $state(false);
  let detailItem = $state<any>(null);

  let editNameModalOpen = $state(false);
  let editingItem = $state<{ ID?: string; Address: string; Name: string; kind: string } | null>(null);
  let editingNameValue = $state("");

  let deleteModalOpen = $state(false);
  let deletingItem = $state<{ ID?: string; Address: string; Name: string; kind: string } | null>(null);

  let showResetModal = $state(false);

  const getDataList = (item: any): any[] => item?.EnvData || item?.Data || [];
  const getLastData = (item: any) => {
    const list = getDataList(item);
    return list.length > 0 ? list[list.length - 1] : undefined;
  };

  const getColumns = (tab: TabType) => {
    const commonCols = [
      {
        data: "Address",
        title: $_("report.sensorAddress") || "アドレス",
        width: "15%",
      },
      {
        data: "Name",
        title: $_("report.sensorName") || "名前",
        width: "18%",
        render: (n: string) => n || '<span class="text-gray-400">-</span>',
      },
    ];

    let specificCols: any[] = [];
    if (tab === "env") {
      specificCols = [
        {
          data: null,
          title: $_("report.sensorTemp") || "気温",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Temp != null ? `<span class="text-orange-500 font-mono font-medium">${l.Temp.toFixed(1)} °C</span>` : "-";
          },
        },
        {
          data: null,
          title: $_("report.sensorHumidity") || "湿度",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Humidity != null ? `<span class="text-cyan-500 font-mono font-medium">${l.Humidity.toFixed(1)} %</span>` : "-";
          },
        },
        {
          data: null,
          title: "CO2 / TVOC",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            if (!l) return "-";
            const parts: string[] = [];
            if (l.ECo2) parts.push(`${l.ECo2} ppm`);
            if (l.ETVOC) parts.push(`${l.ETVOC} ppb`);
            return parts.length > 0 ? `<span class="font-mono">${parts.join(" / ")}</span>` : "-";
          },
        },
      ];
    } else if (tab === "power") {
      specificCols = [
        {
          data: null,
          title: $_("report.sensorPower") || "電力 (W)",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            const val = l?.Load ?? l?.Power;
            return val != null ? `<span class="text-amber-500 font-mono font-medium">${Number(val).toFixed(1)} W</span>` : "-";
          },
        },
        {
          data: null,
          title: $_("report.sensorSwitch") || "スイッチ",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Switch ? '<span class="text-emerald-500 font-bold">ON</span>' : '<span class="text-gray-400">OFF</span>';
          },
        },
        {
          data: null,
          title: $_("report.sensorOverload") || "過負荷",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            const isOver = l?.Over ?? l?.Overload;
            return isOver ? '<span class="text-red-500 font-bold">OVERLOAD</span>' : '<span class="text-gray-400">OK</span>';
          },
        },
      ];
    } else {
      specificCols = [
        {
          data: null,
          title: $_("report.sensorMoving") || "人感",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Moving ? '<span class="text-emerald-500 font-bold">MOVING</span>' : '<span class="text-gray-400">STILL</span>';
          },
        },
        {
          data: null,
          title: $_("report.sensorLight") || "明暗",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Light ? '<span class="text-amber-400 font-bold">BRIGHT</span>' : '<span class="text-gray-400">DARK</span>';
          },
        },
        {
          data: null,
          title: $_("report.sensorBattery") || "電池",
          width: "12%",
          render: (_: any, __: string, row: any) => {
            const l = getLastData(row);
            return l?.Battery != null ? `<span class="font-mono">${l.Battery} %</span>` : "-";
          },
        },
      ];
    }

    const trailingCols = [
      {
        data: "Count",
        title: $_("report.sensorDataCount") || "データ数",
        width: "10%",
        render: renderCount,
      },
      {
        data: "Host",
        title: $_("report.scannerHost") || "ホスト",
        width: "10%",
      },
      {
        data: "LastTime",
        title: $_("report.lastSeen") || "最終確認",
        width: "15%",
        render: renderTime,
      },
    ];

    return [...commonCols, ...specificCols, ...trailingCols];
  };

  const currentDataset = $derived.by(() => {
    switch (activeTab) {
      case "env":
        return envs;
      case "power":
        return powers;
      case "motion":
        return motions;
    }
  });

  const showTable = async () => {
    await tick();
    if (table) {
      try {
        table.destroy(true);
      } catch (e) {
        console.error(e);
      }
      table = undefined;
    }
    const base = document.getElementById("envSensorTableBase");
    if (!base) return;
    base.innerHTML = `<table id="envSensorTable" class="display compact" style="width:99%"></table>`;

    selectedCount = 0;
    selected = null;
    table = new DataTable("#envSensorTable", {
      columns: getColumns(activeTab),
      data: currentDataset,
      stateSave: false,
      order: [[getColumns(activeTab).length - 1, "desc"]],
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
        openDetail(selected);
      }
    });
  };

  const switchTab = (tab: TabType) => {
    activeTab = tab;
    showTable();
  };

  const loadAll = async () => {
    internalLoading = true;
    try {
      const [e, p, m] = await Promise.all([
        GetLogReport("envMonitor"),
        GetLogReport("powerMonitor"),
        GetLogReport("motionSensor"),
      ]);
      envs = (e || []) as EnvMonitorEnt[];
      powers = (p || []) as PowerMonitorEnt[];
      motions = (m || []) as MotionSensorEnt[];
      await showTable();
    } catch (err) {
      console.error("Failed to load environment sensors:", err);
    } finally {
      internalLoading = false;
    }
  };

  export const refresh = () => {
    loadAll();
  };

  onMount(() => {
    loadAll();
    return () => {
      if (table) {
        try {
          table.destroy(true);
        } catch (e) {}
        table = undefined;
      }
    };
  });

  const stats = $derived.by(() => {
    const totalSensors = envs.length + powers.length + motions.length;
    let sumTemp = 0;
    let sumHum = 0;
    let envCount = 0;
    for (const e of envs) {
      const last = getLastData(e);
      if (last?.Temp != null) {
        sumTemp += last.Temp;
        sumHum += last.Humidity || 0;
        envCount++;
      }
    }
    const avgTemp = envCount > 0 ? (sumTemp / envCount).toFixed(1) : "-";
    const avgHum = envCount > 0 ? (sumHum / envCount).toFixed(1) : "-";

    let powerOnCount = 0;
    let powerOverCount = 0;
    for (const p of powers) {
      const last = getLastData(p);
      if (last?.Switch) powerOnCount++;
      if (last?.Over ?? last?.Overload) powerOverCount++;
    }

    let movingActive = 0;
    let lightActive = 0;
    for (const m of motions) {
      const last = getLastData(m);
      if (last?.Moving) movingActive++;
      if (last?.Light) lightActive++;
    }

    return { totalSensors, avgTemp, avgHum, powerOnCount, powerOverCount, movingActive, lightActive };
  });

  const detailChartOption = $derived.by((): echarts.EChartsOption | null => {
    const data = getDataList(detailItem);
    if (!detailItem || data.length === 0) return null;
    const dark = isDarkMode();

    if (activeTab === "env") {
      const times = data.map((d: any) => formatTimeStr(d.Time));
      const temps = data.map((d: any) => d.Temp);
      const hums = data.map((d: any) => d.Humidity);
      return {
        backgroundColor: "transparent",
        tooltip: { trigger: "axis" },
        legend: { top: 0, data: ["Temperature (°C)", "Humidity (%)"], textStyle: { color: dark ? "#cbd5e1" : "#475569", fontSize: 11 } },
        grid: { top: 34, left: "3%", right: "3%", bottom: 8, containLabel: true },
        xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        yAxis: [
          { type: "value", name: "°C", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
          { type: "value", name: "%", max: 100, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 }, splitLine: { show: false } },
        ],
        series: [
          { name: "Temperature (°C)", type: "line", smooth: true, data: temps, lineStyle: { color: "#f97316", width: 2 } },
          { name: "Humidity (%)", type: "line", yAxisIndex: 1, smooth: true, data: hums, lineStyle: { color: "#06b6d4", width: 2 } },
        ],
      };
    } else if (activeTab === "power") {
      const times = data.map((d: any) => formatTimeStr(d.Time));
      const pData = data.map((d: any) => d.Load ?? d.Power ?? 0);
      return {
        backgroundColor: "transparent",
        tooltip: { trigger: "axis" },
        grid: { top: 30, left: "3%", right: "3%", bottom: 8, containLabel: true },
        xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        yAxis: { type: "value", name: "Watts", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        series: [
          { name: "Power (W)", type: "line", smooth: true, data: pData, lineStyle: { color: "#f59e0b", width: 2 }, areaStyle: { color: "rgba(245, 158, 11, 0.2)" } },
        ],
      };
    } else {
      const times = data.map((d: any) => formatTimeStr(d.Time));
      const movings = data.map((d: any) => (d.Moving ? 1 : 0));
      return {
        backgroundColor: "transparent",
        tooltip: { trigger: "axis" },
        grid: { top: 20, left: "3%", right: "3%", bottom: 8, containLabel: true },
        xAxis: { type: "category", data: times, axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        yAxis: { type: "value", min: 0, max: 1, interval: 1, axisLabel: { formatter: (v: number) => (v === 1 ? "MOVE" : "STILL"), color: dark ? "#94a3b8" : "#64748b", fontSize: 10 } },
        series: [
          { name: "Motion", type: "bar", data: movings, itemStyle: { color: "#10b981" } },
        ],
      };
    }
  });

  const openDetail = (item: any) => {
    detailItem = item;
    detailModalOpen = true;
  };

  const openEditName = (item: any) => {
    editingItem = { ID: item.ID, Address: item.Address, Name: item.Name || "", kind: activeTab };
    editingNameValue = item.Name || "";
    editNameModalOpen = true;
  };

  const saveEditedName = async () => {
    if (!editingItem) return;
    try {
      await SetLogReportName(editingItem.kind, editingItem.ID || editingItem.Address, editingNameValue);
      editNameModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to update env sensor name:", err);
    }
  };

  const openDelete = (item: any) => {
    deletingItem = { ID: item.ID, Address: item.Address, Name: item.Name || "", kind: activeTab };
    deleteModalOpen = true;
  };

  const confirmDelete = async () => {
    if (!deletingItem) return;
    const reportKind =
      deletingItem.kind === "env"
        ? "envMonitor"
        : deletingItem.kind === "power"
        ? "powerMonitor"
        : "motionSensor";
    try {
      await DeleteLogReport(reportKind, [deletingItem.ID || deletingItem.Address]);
      deleteModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to delete sensor:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    if (activeTab === "env") {
      const headers = ["Address", "Name", "Host", "Latest Temp", "Latest Hum", "Data Count", "Last Seen"];
      const rows = envs.map((e) => {
        const last = getLastData(e);
        return [
          e.Address,
          e.Name,
          e.Host,
          last?.Temp ?? "-",
          last?.Humidity ?? "-",
          e.Count,
          formatTimeStr(e.LastTime),
        ];
      });
      exportReportData(format, "TWSNMP_Env_Monitors", headers, rows);
    } else if (activeTab === "power") {
      const headers = ["Address", "Name", "Host", "Power (W)", "Switch", "Overload", "Data Count", "Last Seen"];
      const rows = powers.map((p) => {
        const last = getLastData(p);
        return [
          p.Address,
          p.Name,
          p.Host,
          last?.Load ?? last?.Power ?? "-",
          last?.Switch ? "ON" : "OFF",
          last?.Over ?? last?.Overload ? "OVERLOAD" : "OK",
          p.Count,
          formatTimeStr(p.LastTime),
        ];
      });
      exportReportData(format, "TWSNMP_Power_Monitors", headers, rows);
    } else {
      const headers = ["Address", "Name", "Host", "Moving", "Light", "Battery (%)", "Data Count", "Last Seen"];
      const rows = motions.map((m) => {
        const last = getLastData(m);
        return [
          m.Address,
          m.Name,
          m.Host,
          last?.Moving ? "MOVING" : "STILL",
          last?.Light ? "BRIGHT" : "DARK",
          last?.Battery ?? "-",
          m.Count,
          formatTimeStr(m.LastTime),
        ];
      });
      exportReportData(format, "TWSNMP_Motion_Sensors", headers, rows);
    }
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    const kind =
      activeTab === "env"
        ? "envMonitor"
        : activeTab === "power"
        ? "powerMonitor"
        : "motionSensor";
    try {
      await ResetLogReport(kind);
      await loadAll();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset sensors:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiThermometer} size={0.9} class="text-orange-500" />
        {$_("report.sensorKpiTotal") || "検出センサー総数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.totalSensors}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Env: {envs.length} / Power: {powers.length} / Motion: {motions.length}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiWaterPercent} size={0.9} class="text-cyan-500" />
        {$_("report.sensorTemp") || "平均気温 / 湿度"}
      </div>
      <div class="mt-1 text-xl font-bold text-cyan-600 dark:text-cyan-400">
        {stats.avgTemp} °C / {stats.avgHum} %
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiFlash} size={0.9} class="text-amber-500" />
        {$_("report.tabPower") || "電力プラグ (ON / 過負荷)"}
      </div>
      <div class="mt-1 text-xl font-bold text-amber-600 dark:text-amber-400">
        ON: {stats.powerOnCount} / Over: {stats.powerOverCount}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiWalk} size={0.9} class="text-emerald-500" />
        {$_("report.tabMotion") || "人感検知 (Moving / Light)"}
      </div>
      <div class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">
        Mov: {stats.movingActive} / Light: {stats.lightActive}
      </div>
    </div>
  </div>

  <!-- Sub-Tabs Switcher -->
  <div class="flex items-center gap-2 pb-1">
    <ButtonGroup>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'env' ? 'blue' : 'alternative'}
        onclick={() => switchTab('env')}
      >
        <Icon path={icons.mdiThermometer} size={0.8} />
        {$_("report.tabEnv") || "環境センサー"} ({envs.length})
      </Button>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'power' ? 'blue' : 'alternative'}
        onclick={() => switchTab('power')}
      >
        <Icon path={icons.mdiFlash} size={0.8} />
        {$_("report.tabPower") || "電力センサー"} ({powers.length})
      </Button>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'motion' ? 'blue' : 'alternative'}
        onclick={() => switchTab('motion')}
      >
        <Icon path={icons.mdiWalk} size={0.8} />
        {$_("report.tabMotion") || "人感センサー"} ({motions.length})
      </Button>
    </ButtonGroup>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div id="envSensorTableBase">
      <table id="envSensorTable" class="display compact" style="width:99%"></table>
    </div>
  </div>

  <!-- Unified Bottom Action Buttons Bar -->
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount === 1 && selected}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={() => selected && openDetail(selected)}
        size="xs"
      >
        <Icon path={icons.mdiChartLine} size={1} />
        {$_("report.sensorDetail") || "センサー履歴グラフ"}
      </GradientButton>

      <GradientButton
        shadow
        color="cyan"
        type="button"
        onclick={() => selected && openEditName(selected)}
        size="xs"
      >
        <Icon path={icons.mdiPencilOutline} size={1} />
        {$_("report.sensorEditName") || "名前変更"}
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
      onclick={loadAll}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("report.btnReload") || "更新"}
    </GradientButton>
  </div>
</div>

<!-- Sensor Time Series Detail Modal -->
<Modal bind:open={detailModalOpen} size="lg" dismissable={false} outsideclose={false}>
  {#if detailItem}
    <div class="space-y-4 p-2 text-gray-800 dark:text-gray-100">
      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiThermometer} size={1.2} class="text-orange-500" />
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">
              {detailItem.Name || "Sensor"} ({detailItem.Address})
            </h2>
            <div class="text-xs text-gray-500 font-mono">Host: {detailItem.Host}</div>
          </div>
        </div>
      </div>

      {#if getDataList(detailItem).length > 0}
        <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
          <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
            <Icon path={icons.mdiChartTimelineVariant} size={0.8} class="text-cyan-500" />
            {$_("report.sensorHistoryTrend") || "計測値時系列推移"}
          </div>
          <div use:echartAction={detailChartOption} class="h-44 w-full"></div>
        </div>
      {:else}
        <div class="flex flex-col items-center justify-center p-6 bg-gray-50 dark:bg-gray-800/40 rounded-lg border border-gray-200 dark:border-gray-700 text-center my-2">
          <Icon path={icons.mdiChartTimelineVariant} size={1.5} class="text-gray-400 mb-1" />
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {$_("report.noEnvData") || "環境センサー計測推移の記録データはありません"}
          </div>
        </div>
      {/if}

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs font-mono">
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">データポイント数</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold text-sm mt-0.5">{getDataList(detailItem).length}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">総検知数</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold text-sm mt-0.5">{detailItem.Count?.toLocaleString()}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">初回検知</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(detailItem.FirstTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">最終確認</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(detailItem.LastTime)}</div>
        </div>
      </div>

      <div class="flex justify-end space-x-2 pt-2 border-t border-gray-200 dark:border-gray-700">
        <GradientButton shadow type="button" color="teal" onclick={() => (detailModalOpen = false)} size="xs">
          <Icon path={icons.mdiCancel} size={1} />
          {$_("common.close") || "閉じる"}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<!-- Edit Sensor Name Modal -->
<Modal bind:open={editNameModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="p-2 space-y-3">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-1.5">
      <Icon path={icons.mdiPencilOutline} size={0.8} />
      {$_("report.sensorEditName") || "センサー名の変更"}
    </h3>
    <div class="text-xs text-gray-500 font-mono">
      {editingItem?.Address}
    </div>
    <div>
      <Label for="sName" class="mb-1 text-xs">{$_("report.sensorNameLabel") || "表示名"}</Label>
      <Input id="sName" bind:value={editingNameValue} placeholder={$_("report.sensorNamePlaceholder") || "例: サーバールームA 温湿度計"} size="sm" />
    </div>
    <div class="flex justify-end gap-2 pt-2">
      <Button color="blue" size="xs" onclick={saveEditedName}>
        {$_("common.save") || "保存"}
      </Button>
      <Button color="alternative" size="xs" onclick={() => (editNameModalOpen = false)}>
        {$_("common.cancel") || "キャンセル"}
      </Button>
    </div>
  </div>
</Modal>

<!-- Delete Sensor Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={deleteModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {deletingItem ? `センサー ${deletingItem.Address} を削除しますか？` : ""}
    </h3>
    <div class="flex justify-center gap-3">
      <GradientButton shadow color="red" onclick={confirmDelete} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.delete") || "削除"}
      </GradientButton>
      <GradientButton shadow color="teal" onclick={() => (deleteModalOpen = false)} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("common.cancel") || "キャンセル"}
      </GradientButton>
    </div>
  </div>
</Modal>

<!-- Reset All Modal -->
<Modal bind:open={showResetModal} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.confirmClearItem", { values: { name: activeTab } }) || `${activeTab} のレポートデータを全消去しますか？`}
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
