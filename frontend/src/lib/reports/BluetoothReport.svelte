<script lang="ts">
  import "../../assets/css/jquery.dataTables.css";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { onMount, tick } from "svelte";
  import { _ } from "svelte-i18n";
  import * as echarts from "echarts";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Modal, Button, GradientButton, Input, Label } from "flowbite-svelte";
  import { formatTimeStr, isDarkMode, exportReportData, echartAction } from "./utils";
  import { getTableLang, renderCount, renderTime } from "../common";
  import {
    GetLogReport,
    ResetLogReport,
    DeleteLogReport,
    SetLogReportName,
  } from "../../../wailsjs/go/main/App";
  import type { BlueDeviceEnt } from "./types";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  let devices = $state<BlueDeviceEnt[]>([]);
  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<BlueDeviceEnt | null>(null);
  let selectedCount = $state(0);

  // Modals
  let detailModalOpen = $state(false);
  let detailDevice = $state<BlueDeviceEnt | null>(null);

  let editNameModalOpen = $state(false);
  let editingDevice = $state<BlueDeviceEnt | null>(null);
  let editingNameValue = $state("");

  let deleteModalOpen = $state(false);
  let deletingDevice = $state<BlueDeviceEnt | null>(null);

  let showResetModal = $state(false);

  const columns = [
    {
      data: "Address",
      title: $_("report.bluetoothAddress") || "アドレス (MAC)",
      width: "15%",
    },
    {
      data: "AddressType",
      title: "種別",
      width: "8%",
      render: (t: string) =>
        t === "public"
          ? '<span class="text-blue-500 font-semibold">Public</span>'
          : '<span class="text-gray-400">Random</span>',
    },
    {
      data: "Name",
      title: $_("report.bluetoothName") || "デバイス名",
      width: "18%",
      render: (n: string) => n || '<span class="text-gray-400">-</span>',
    },
    {
      data: "Vendor",
      title: $_("report.vendor") || "ベンダー",
      width: "14%",
      render: (v: string) => v || '<span class="text-gray-400">-</span>',
    },
    {
      data: "RSSI",
      title: "RSSI (dBm)",
      width: "10%",
      render: (r: any[]) => {
        const val = r && r.length > 0 ? r[r.length - 1].Value : -100;
        return `${val} dBm`;
      },
    },
    {
      data: "Count",
      title: $_("report.count") || "回数",
      width: "10%",
      render: renderCount,
    },
    {
      data: "Host",
      title: $_("report.scannerHost") || "スキャナー",
      width: "10%",
    },
    {
      data: "LastTime",
      title: $_("report.lastSeen") || "最終確認",
      width: "15%",
      render: renderTime,
    },
  ];

  const showTable = async () => {
    await tick();
    if (!document.getElementById("blueDeviceTable")) return;

    if (table && DataTable.isDataTable("#blueDeviceTable")) {
      selectedCount = 0;
      selected = null;
      table.clear();
      table.rows.add(devices);
      table.draw();
      return;
    }

    selectedCount = 0;
    selected = null;
    table = new DataTable("#blueDeviceTable", {
      destroy: true,
      columns: columns,
      data: devices,
      stateSave: true,
      order: [[7, "desc"]],
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

  const loadAll = async () => {
    internalLoading = true;
    try {
      const data = await GetLogReport("blueDevice");
      devices = (data || []) as BlueDeviceEnt[];
      await showTable();
    } catch (err) {
      console.error("Failed to load Bluetooth reports:", err);
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
      if (table && DataTable.isDataTable("#blueDeviceTable")) {
        table.destroy();
        table = undefined;
      }
    };
  });

  const getLatestRSSI = (dev: BlueDeviceEnt): number => {
    if (dev.RSSI && dev.RSSI.length > 0) {
      return dev.RSSI[dev.RSSI.length - 1].Value;
    }
    return -100;
  };

  const stats = $derived.by(() => {
    const total = devices.length;
    let publicCount = 0;
    let randomCount = 0;
    let goodSignalCount = 0;
    let sumRssi = 0;
    let rssiValidCount = 0;
    const vendorMap = new Map<string, number>();
    const hostSet = new Set<string>();

    for (const d of devices) {
      if (d.AddressType === "public") publicCount++;
      else randomCount++;

      const r = getLatestRSSI(d);
      if (r > -100) {
        sumRssi += r;
        rssiValidCount++;
        if (r >= -75) goodSignalCount++;
      }

      const v = d.Vendor || "Unknown";
      vendorMap.set(v, (vendorMap.get(v) || 0) + 1);

      if (d.Host) hostSet.add(d.Host);
    }

    let topVendor = "-";
    let maxV = 0;
    for (const [v, c] of vendorMap.entries()) {
      if (c > maxV) {
        maxV = c;
        topVendor = `${v} (${c})`;
      }
    }

    const avgRssi = rssiValidCount > 0 ? Math.round(sumRssi / rssiValidCount) : -100;
    return {
      total,
      publicCount,
      randomCount,
      goodSignalCount,
      avgRssi,
      vendorCount: vendorMap.size,
      topVendor,
      scannerHosts: hostSet.size,
    };
  });

  const vendorChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (devices.length === 0) return null;
    const dark = isDarkMode();
    const vendorMap = new Map<string, number>();
    for (const d of devices) {
      const v = d.Vendor || "Unknown";
      vendorMap.set(v, (vendorMap.get(v) || 0) + 1);
    }
    const data = [...vendorMap.entries()]
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 8);

    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "item" },
      legend: { orient: "vertical", right: "2%", top: "center", textStyle: { color: dark ? "#cbd5e1" : "#475569", fontSize: 10 } },
      series: [
        {
          name: "Vendor",
          type: "pie",
          radius: ["40%", "70%"],
          center: ["35%", "50%"],
          avoidLabelOverlap: false,
          itemStyle: { borderRadius: 4, borderColor: dark ? "#0f172a" : "#ffffff", borderWidth: 1 },
          label: { show: false },
          data,
        },
      ],
    };
  });

  const rssiChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (devices.length === 0) return null;
    const dark = isDarkMode();
    let strong = 0;
    let good = 0;
    let fair = 0;
    let weak = 0;

    for (const d of devices) {
      const r = getLatestRSSI(d);
      if (r >= -65) strong++;
      else if (r >= -75) good++;
      else if (r >= -85) fair++;
      else weak++;
    }

    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
      grid: { top: "10%", left: "4%", right: "4%", bottom: "10%", containLabel: true },
      xAxis: {
        type: "category",
        data: [">= -65dBm\n(強力)", "-66 ~ -75\n(良好)", "-76 ~ -85\n(普通)", "<= -86\n(微弱)"],
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 },
      },
      yAxis: { type: "value", axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 }, splitLine: { lineStyle: { color: dark ? "#1e293b" : "#f1f5f9" } } },
      series: [
        {
          name: "Devices",
          type: "bar",
          data: [
            { value: strong, itemStyle: { color: "#10b981" } },
            { value: good, itemStyle: { color: "#06b6d4" } },
            { value: fair, itemStyle: { color: "#f59e0b" } },
            { value: weak, itemStyle: { color: "#ef4444" } },
          ],
          barMaxWidth: 36,
          borderRadius: [4, 4, 0, 0],
        },
      ],
    };
  });

  const detailChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!detailDevice || !detailDevice.RSSI || detailDevice.RSSI.length === 0) return null;
    const dark = isDarkMode();
    const data = detailDevice.RSSI.map((p) => [formatTimeStr(p.Time), p.Value]);

    return {
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      grid: { top: "15%", left: "4%", right: "4%", bottom: "10%", containLabel: true },
      xAxis: {
        type: "category",
        data: data.map((d) => d[0]),
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 },
      },
      yAxis: {
        type: "value",
        name: "dBm",
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 },
        splitLine: { lineStyle: { color: dark ? "#1e293b" : "#f1f5f9" } },
      },
      series: [
        {
          name: "RSSI",
          type: "line",
          smooth: true,
          data: data.map((d) => d[1]),
          lineStyle: { color: "#3b82f6", width: 2 },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: "rgba(59, 130, 246, 0.3)" },
              { offset: 1, color: "rgba(59, 130, 246, 0.0)" },
            ]),
          },
        },
      ],
    };
  });

  const openDetail = (dev: BlueDeviceEnt) => {
    detailDevice = dev;
    detailModalOpen = true;
  };

  const openEditName = (dev: BlueDeviceEnt) => {
    editingDevice = dev;
    editingNameValue = dev.Name || "";
    editNameModalOpen = true;
  };

  const saveEditedName = async () => {
    if (!editingDevice) return;
    try {
      await SetLogReportName("blueDevice", editingDevice.ID || editingDevice.Address, editingNameValue);
      editNameModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to update device name:", err);
    }
  };

  const openDelete = (dev: BlueDeviceEnt) => {
    deletingDevice = dev;
    deleteModalOpen = true;
  };

  const confirmDelete = async () => {
    if (!deletingDevice) return;
    try {
      await DeleteLogReport("blueDevice", [deletingDevice.ID || deletingDevice.Address]);
      deleteModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to delete bluetooth device:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    const headers = [
      "Address",
      "Type",
      "Name",
      "Vendor",
      "Latest RSSI",
      "Count",
      "Scanner Host",
      "First Seen",
      "Last Seen",
      "Changes",
      "Info",
    ];
    const rows = devices.map((d) => [
      d.Address,
      d.AddressType,
      d.Name,
      d.Vendor,
      getLatestRSSI(d),
      d.Count,
      d.Host,
      formatTimeStr(d.FirstTime),
      formatTimeStr(d.LastTime),
      d.Change,
      d.Info,
    ]);
    exportReportData(format, "TWSNMP_Bluetooth_Report", headers, rows);
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    try {
      await ResetLogReport("blueDevice");
      await loadAll();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset bluetooth reports:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiBluetooth} size={0.9} class="text-blue-500" />
        {$_("report.bluetoothTotalDevices") || "検出BLEデバイス総数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.total.toLocaleString()}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Public: {stats.publicCount} / Random: {stats.randomCount}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiDomain} size={0.9} class="text-indigo-500" />
        {$_("report.bluetoothVendorCount") || "ベンダー種別数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.vendorCount}</div>
      <div class="mt-1 text-[11px] text-gray-400 truncate" title={stats.topVendor}>
        {stats.topVendor}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiSignalVariant} size={0.9} class="text-emerald-500" />
        {$_("report.bluetoothGoodSignal") || "良好信号 (>= -75dBm)"}
      </div>
      <div class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">{stats.goodSignalCount}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        {$_("report.bluetoothAvgRssi", { values: { rssi: stats.avgRssi } }) || `平均: ${stats.avgRssi} dBm`}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiLaptop} size={0.9} class="text-amber-500" />
        {$_("report.bluetoothScannerHosts") || "スキャナーホスト数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.scannerHosts.toLocaleString()}</div>
    </div>
  </div>

  <!-- Visual Analytics Charts Row -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
    <!-- Top Vendors Pie -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-1 flex items-center gap-1.5">
        <Icon path={icons.mdiChartPie} size={0.8} class="text-indigo-500" />
        {$_("report.bluetoothVendorShare") || "ベンダー構成比 (Top 8)"}
      </div>
      <div use:echartAction={vendorChartOption} class="h-36 w-full"></div>
    </div>

    <!-- RSSI Levels Distribution -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-1 flex items-center gap-1.5">
        <Icon path={icons.mdiSignalCellular3} size={0.8} class="text-emerald-500" />
        {$_("report.bluetoothSignalDistribution") || "電波強度 (RSSI) レベル分布"}
      </div>
      <div use:echartAction={rssiChartOption} class="h-36 w-full"></div>
    </div>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div>
      <table id="blueDeviceTable" class="display compact" style="width:99%"></table>
    </div>
  </div>

  <!-- Row Selection Action Buttons Bar -->
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
        <Icon path={icons.mdiChartTimelineVariant} size={1} />
        {$_("report.bluetoothRssiHistory") || "電波推移・詳細"}
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

<!-- BLE Device Detail & RSSI Chart Modal -->
<Modal bind:open={detailModalOpen} size="lg" dismissable={false} outsideclose={false}>
  {#if detailDevice}
    <div class="space-y-4 p-2 text-gray-800 dark:text-gray-100">
      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiBluetooth} size={1.2} class="text-blue-500" />
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">
              {detailDevice.Name || "Unknown BLE Device"} ({detailDevice.Address})
            </h2>
            <div class="text-xs text-gray-500 font-mono">
              Type: {detailDevice.AddressType} / Vendor: {detailDevice.Vendor || "Generic"}
            </div>
          </div>
        </div>
      </div>

      {#if detailDevice.RSSI && detailDevice.RSSI.length > 0}
        <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
          <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
            <Icon path={icons.mdiChartTimelineVariant} size={0.8} class="text-blue-500" />
            {$_("report.bluetoothRssiHistory") || "RSSI 電波強度推移 (直近検知)"}
          </div>
          <div use:echartAction={detailChartOption} class="h-44 w-full"></div>
        </div>
      {:else}
        <div class="flex flex-col items-center justify-center p-6 bg-gray-50 dark:bg-gray-800/40 rounded-lg border border-gray-200 dark:border-gray-700 text-center my-2">
          <Icon path={icons.mdiChartTimelineVariant} size={1.5} class="text-gray-400 mb-1" />
          <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
            {$_("report.noRssiData") || "電波強度推移の記録データはありません"}
          </div>
        </div>
      {/if}

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs font-mono">
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">初回検知</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(detailDevice.FirstTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">最終検知</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(detailDevice.LastTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">受信回数</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold text-sm mt-0.5">{detailDevice.Count?.toLocaleString()}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">スキャナー</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{detailDevice.Host}</div>
        </div>
      </div>

      {#if detailDevice.Info}
        <div class="p-2 rounded bg-gray-100 dark:bg-gray-800 text-xs font-mono border border-gray-200 dark:border-gray-700 break-all text-gray-800 dark:text-gray-200">
          {detailDevice.Info}
        </div>
      {/if}

      <div class="flex justify-end space-x-2 pt-2 border-t border-gray-200 dark:border-gray-700">
        <GradientButton shadow type="button" color="teal" onclick={() => (detailModalOpen = false)} size="xs">
          <Icon path={icons.mdiCancel} size={1} />
          {$_("common.close") || "閉じる"}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<!-- Edit Device Name Modal -->
<Modal bind:open={editNameModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="p-2 space-y-3">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-1.5">
      <Icon path={icons.mdiPencilOutline} size={0.8} />
      {$_("report.sensorEditName") || "デバイス名の変更"}
    </h3>
    <div class="text-xs text-gray-500 font-mono">
      {editingDevice?.Address}
    </div>
    <div>
      <Label for="devName" class="mb-1 text-xs">{$_("report.bluetoothDeviceNameLabel") || "デバイス名"}</Label>
      <Input id="devName" bind:value={editingNameValue} placeholder={$_("report.bluetoothNamePlaceholder") || "例: BLEビーコン A"} size="sm" />
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

<!-- Delete Device Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={deleteModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {deletingDevice ? `Bluetoothデバイス ${deletingDevice.Address} を削除しますか？` : ""}
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
      {$_("report.confirmClearBluetooth") || "Bluetoothデバイスのレポートデータを全消去しますか？"}
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
