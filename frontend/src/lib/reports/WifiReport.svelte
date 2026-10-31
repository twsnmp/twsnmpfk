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
  import { getTableLang, renderCount, renderTime } from "../common";
  import { GetLogReport, ResetLogReport } from "../../../wailsjs/go/main/App";
  import type { WifiAPEnt } from "./types";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  let apList = $state<WifiAPEnt[]>([]);
  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<WifiAPEnt | null>(null);
  let selectedCount = $state(0);

  let showDetailModal = $state(false);
  let showResetModal = $state(false);



  const columns = [
    {
      data: "SSID",
      title: "SSID",
      width: "20%",
      render: (d: string) => d || "&lt;Hidden&gt;",
    },
    {
      data: "BSSID",
      title: "BSSID (MAC)",
      width: "15%",
    },
    {
      data: "Vendor",
      title: $_("report.vendor") || "ベンダー",
      width: "15%",
    },
    {
      data: "Channel",
      title: $_("report.channel") || "チャネル",
      width: "10%",
      render: (c: string) => `Ch ${c || "?"}`,
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
      width: "10%",
      render: renderTime,
    },
  ];

  const showTable = async () => {
    await tick();
    if (!document.getElementById("wifiAPTable")) return;

    if (table && DataTable.isDataTable("#wifiAPTable")) {
      selectedCount = 0;
      selected = null;
      table.clear();
      table.rows.add(apList);
      table.draw();
      return;
    }

    selectedCount = 0;
    selected = null;
    table = new DataTable("#wifiAPTable", {
      destroy: true,
      columns: columns,
      data: apList,
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

  const loadAPData = async () => {
    internalLoading = true;
    try {
      const data = await GetLogReport("wifiAP");
      apList = (data || []) as WifiAPEnt[];
      await showTable();
      renderChart();
    } catch (err) {
      console.error("Failed to load WifiAP data:", err);
    } finally {
      internalLoading = false;
    }
  };

  export const refresh = () => {
    loadAPData();
  };

  onMount(() => {
    loadAPData();
    return () => {
      if (table && DataTable.isDataTable("#wifiAPTable")) {
        table.destroy();
        table = undefined;
      }
    };
  });

  const stats = $derived.by(() => {
    const total = apList.length;
    const uniqueSSIDs = new Set(apList.map((a) => a.SSID).filter(Boolean)).size;
    const uniqueHosts = new Set(apList.map((a) => a.Host).filter(Boolean)).size;
    const channels = new Map<string, number>();
    for (const a of apList) {
      if (a.Channel) {
        channels.set(a.Channel, (channels.get(a.Channel) || 0) + 1);
      }
    }
    let topChannel = "-";
    let maxC = 0;
    for (const [ch, count] of channels.entries()) {
      if (count > maxC) {
        maxC = count;
        topChannel = `Ch ${ch} (${count})`;
      }
    }
    return { total, uniqueSSIDs, uniqueHosts, topChannel };
  });

  const channelChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (apList.length === 0) return null;
    const channelMap = new Map<string, number>();
    for (const ap of apList) {
      const ch = ap.Channel ? `Ch ${ap.Channel}` : "Unknown";
      channelMap.set(ch, (channelMap.get(ch) || 0) + 1);
    }

    const sortedChannels = [...channelMap.entries()].sort((a, b) => b[1] - a[1]).slice(0, 10);
    const dark = isDarkMode();

    return {
      backgroundColor: "transparent",
      tooltip: {
        trigger: "axis",
        axisPointer: { type: "shadow" },
        backgroundColor: dark ? "rgba(15, 23, 42, 0.95)" : "rgba(255, 255, 255, 0.95)",
        borderColor: dark ? "#334155" : "#e2e8f0",
        textStyle: { color: dark ? "#f8fafc" : "#0f172a", fontSize: 12 },
      },
      grid: { top: "10%", left: "3%", right: "4%", bottom: "5%", containLabel: true },
      xAxis: {
        type: "category",
        data: sortedChannels.map((c) => c[0]),
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 11 },
        axisLine: { lineStyle: { color: dark ? "#334155" : "#cbd5e1" } },
      },
      yAxis: {
        type: "value",
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 11 },
        splitLine: { lineStyle: { color: dark ? "#1e293b" : "#f1f5f9" } },
      },
      series: [
        {
          name: $_("report.apCount") || "AP Count",
          type: "bar",
          data: sortedChannels.map((c) => c[1]),
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: "#06b6d4" },
              { offset: 1, color: "#0891b2" },
            ]),
            borderRadius: [4, 4, 0, 0],
          },
          barMaxWidth: 32,
        },
      ],
    };
  });

  const rssiChartOption = $derived.by((): echarts.EChartsOption | null => {
    if (!selected || !selected.RSSI || selected.RSSI.length === 0) return null;
    const dark = isDarkMode();
    const data = selected.RSSI.map((p) => [formatTimeStr(p.Time), p.Value]);

    return {
      backgroundColor: "transparent",
      tooltip: {
        trigger: "axis",
        backgroundColor: dark ? "rgba(15, 23, 42, 0.95)" : "rgba(255, 255, 255, 0.95)",
        borderColor: dark ? "#334155" : "#e2e8f0",
        textStyle: { color: dark ? "#f8fafc" : "#0f172a", fontSize: 12 },
      },
      grid: { top: "15%", left: "4%", right: "4%", bottom: "10%", containLabel: true },
      xAxis: {
        type: "category",
        data: data.map((d) => d[0]),
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 },
      },
      yAxis: {
        type: "value",
        name: "RSSI (dBm)",
        axisLabel: { color: dark ? "#94a3b8" : "#64748b", fontSize: 10 },
        splitLine: { lineStyle: { color: dark ? "#1e293b" : "#f1f5f9" } },
      },
      series: [
        {
          name: "RSSI",
          type: "line",
          smooth: true,
          data: data.map((d) => d[1]),
          lineStyle: { color: "#06b6d4", width: 2 },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: "rgba(6, 182, 212, 0.3)" },
              { offset: 1, color: "rgba(6, 182, 212, 0.0)" },
            ]),
          },
        },
      ],
    };
  });

  const openDetail = (ap: WifiAPEnt) => {
    selected = ap;
    showDetailModal = true;
  };

  export function handleExport(format: string = "csv"): void {
    const headers = [
      "SSID",
      "BSSID",
      "Vendor",
      "Channel",
      "Latest RSSI",
      "Count",
      "Scanner Host",
      "First Seen",
      "Last Seen",
      "Changes",
      "Info",
    ];
    const rows = apList.map((ap) => {
      const latestRSSI = ap.RSSI && ap.RSSI.length > 0 ? ap.RSSI[ap.RSSI.length - 1].Value : "-";
      return [
        ap.SSID,
        ap.BSSID,
        ap.Vendor,
        ap.Channel,
        latestRSSI,
        ap.Count,
        ap.Host,
        formatTimeStr(ap.FirstTime),
        formatTimeStr(ap.LastTime),
        ap.Change,
        ap.Info,
      ];
    });
    exportReportData(format, "TWSNMP_Wifi_Report", headers, rows);
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    try {
      await ResetLogReport("wifiAP");
      await loadAPData();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset wifi report:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiWifi} size={0.9} class="text-cyan-500" />
        {$_("report.wifiTotalAPs") || "Wi-Fi AP総数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.total.toLocaleString()}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiRadioTower} size={0.9} class="text-blue-500" />
        {$_("report.wifiUniqueSSIDs") || "ユニークSSID数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.uniqueSSIDs.toLocaleString()}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiChartBar} size={0.9} class="text-emerald-500" />
        {$_("report.wifiTopChannel") || "最多利用チャネル"}
      </div>
      <div class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">{stats.topChannel}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiLaptop} size={0.9} class="text-amber-500" />
        {$_("report.wifiScanners") || "スキャナーホスト数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.uniqueHosts.toLocaleString()}</div>
    </div>
  </div>

  <!-- Channel Distribution Chart Card -->
  <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
    <div class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-2 flex items-center gap-1.5">
      <Icon path={icons.mdiChartBar} size={0.9} class="text-cyan-500" />
      {$_("report.wifiChannelDistribution") || "チャネル別AP分布 (Top 10)"}
    </div>
    <div use:echartAction={channelChartOption} class="h-36 w-full"></div>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div>
      <table id="wifiAPTable" class="display compact" style="width:99%"></table>
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
        <Icon path={icons.mdiChartTimelineVariant} size={1} />
        {$_("report.wifiRssiTrend") || "RSSI 電波推移"}
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
      onclick={loadAPData}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("report.btnReload") || "更新"}
    </GradientButton>
  </div>
</div>

<!-- AP Detail & RSSI Chart Modal -->
<Modal bind:open={showDetailModal} size="lg" dismissable={false} outsideclose={false}>
  {#if selected}
    <div class="space-y-4 p-2 text-gray-800 dark:text-gray-100">
      <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiWifi} size={1.2} class="text-cyan-500" />
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-gray-100">
              {selected.SSID || "<Hidden>"} ({selected.BSSID})
            </h2>
            <div class="text-xs text-gray-500">Ch {selected.Channel} / {selected.Vendor || "Unknown Vendor"}</div>
          </div>
        </div>
      </div>

      {#if selected.RSSI && selected.RSSI.length > 0}
        <div class="rounded-lg border border-gray-200 dark:border-gray-800 p-3 bg-white dark:bg-gray-900">
          <div class="text-xs font-semibold mb-2 flex items-center gap-1.5 text-gray-700 dark:text-gray-300">
            <Icon path={icons.mdiChartTimelineVariant} size={0.8} class="text-cyan-500" />
            RSSI 電波強度推移 (直近)
          </div>
          <div use:echartAction={rssiChartOption} class="h-44 w-full"></div>
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
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(selected.FirstTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">最終検知</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{formatTimeStr(selected.LastTime)}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">検知回数</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold text-sm mt-0.5">{selected.Count?.toLocaleString()}</div>
        </div>
        <div class="rounded-lg bg-gray-100 dark:bg-gray-800 p-2.5 border border-gray-200 dark:border-gray-700">
          <div class="text-gray-500 dark:text-gray-400 font-sans">スキャナー</div>
          <div class="text-gray-900 dark:text-gray-100 font-bold mt-0.5">{selected.Host}</div>
        </div>
      </div>

      {#if selected.Info}
        <div class="p-2 rounded bg-gray-100 dark:bg-gray-800 text-xs font-mono border border-gray-200 dark:border-gray-700 break-all text-gray-800 dark:text-gray-200">
          {selected.Info}
        </div>
      {/if}

      <div class="flex justify-end space-x-2 pt-2 border-t border-gray-200 dark:border-gray-700">
        <GradientButton shadow type="button" color="teal" onclick={() => (showDetailModal = false)} size="xs">
          <Icon path={icons.mdiCancel} size={1} />
          {$_("common.close") || "閉じる"}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<!-- Clear/Reset Confirmation Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={showResetModal} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.confirmClearWifi") || "Wi-Fi APレポートデータを全消去しますか？"}
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
