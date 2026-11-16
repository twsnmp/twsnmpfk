<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetNodes,
    GetPollings,
    DeletePollings,
    DeletePollingLogs,
    ExportPollings,
    GetDefaultPolling,
    ExportPollingAsTemplate,
  } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getLogModeName,
    getTableLang,
    renderPollingType,
  } from "./common";
  import Polling from "./Polling.svelte";
  import AddPolling from "./AddPolling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import PollingReport from "./PollingReport.svelte";
  import { _ } from "svelte-i18n";
  import { currentLang } from "../i18n/i18n";

  let data: any = [];
  let nodes: any = {};
  let showEditPolling = false;
  let showPollingReport = false;
  let selectedPolling = "";
  let table: any = undefined;
  let selectedCount = 0;
  let stateFilter: "all" | "normal" | "alert" | "warn" = "all";

  let kpi = {
    total: 0,
    normal: 0,
    high: 0,
    low: 0,
    warn: 0,
    repair: 0,
    unknown: 0,
    off: 0,
    alertCount: 0,
    warnOtherCount: 0,
    normalRate: "0.0",
  };

  const updateKPI = () => {
    let normal = 0;
    let high = 0;
    let low = 0;
    let warn = 0;
    let repair = 0;
    let unknown = 0;
    let off = 0;

    for (const p of data) {
      if (p.Level === "off") {
        off++;
      }
      switch (p.State) {
        case "normal":
          normal++;
          break;
        case "high":
          high++;
          break;
        case "low":
          low++;
          break;
        case "warn":
          warn++;
          break;
        case "repair":
          repair++;
          break;
        default:
          unknown++;
          break;
      }
    }

    const total = data.length;
    kpi = {
      total,
      normal,
      high,
      low,
      warn,
      repair,
      unknown,
      off,
      alertCount: high + low,
      warnOtherCount: warn + repair + unknown,
      normalRate: total > 0 ? ((normal / total) * 100).toFixed(1) : "0.0",
    };
  };

  const getFilteredData = () => {
    switch (stateFilter) {
      case "normal":
        return data.filter((p: any) => p.State === "normal");
      case "alert":
        return data.filter((p: any) => p.State === "high" || p.State === "low");
      case "warn":
        return data.filter(
          (p: any) =>
            p.State !== "normal" && p.State !== "high" && p.State !== "low"
        );
      default:
        return data;
    }
  };

  const setStateFilter = (next: "all" | "normal" | "alert" | "warn") => {
    stateFilter = stateFilter === next ? "all" : next;
    if (table) {
      selectedCount = 0;
      table.clear().rows.add(getFilteredData()).draw();
    }
  };

  const showTable = () => {
    selectedCount = 0;
    const el = document.getElementById("pollingListTable");
    if (!el) {
      return;
    }
    if (table && DataTable.isDataTable("#pollingListTable")) {
      table.off("select");
      table.off("deselect");
      table.destroy();
      table = undefined;
    }
    el.innerHTML = "";
    table = new DataTable("#pollingListTable", {
      destroy: true,
      columns: getColumns(),
      data: getFilteredData(),
      stateSave: true,
      order: [
        [0, "asc"],
        [1, "asc"],
      ],
      pageLength: window.innerHeight > 800 ? 25 : 10,
      language: getTableLang(),
      select: {
        style: "multi+shift",
      },
    });
    table.on("select", () => {
      if (!table) return;
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      if (!table) return;
      selectedCount = table.rows({ selected: true }).count();
    });
    selectedCount = table.rows({ selected: true }).count();
  };

  const refresh = async () => {
    data = [];
    nodes = await GetNodes();
    data = await GetPollings("");
    updateKPI();
    showTable();
  };

  let showCopyPolling = false;
  let showAddPolling = false;
  let pollingTmp: any = undefined;

  const add = async () => {
    pollingTmp = await GetDefaultPolling("");
    showAddPolling = true;
  };

  const edit = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showEditPolling = true;
  };

  const copy = () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    pollingTmp = selected[0];
    pollingTmp.ID = "";
    showCopyPolling = true;
  };

  const exportTemp = async () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    await ExportPollingAsTemplate(selected[0].ID)
  };

  const report = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showPollingReport = true;
  };

  const deletePollings = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    const ids = selected.toArray();
    await DeletePollings(ids);
    const idSet = new Set(ids);
    data = data.filter((p: any) => !idSet.has(p.ID));
    updateKPI();
    selectedCount = 0;
    table.rows({ selected: true }).remove().draw();
  };

  const deletePollingLogs = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeletePollingLogs(selected.toArray());
  };

  const getColumns = () => [
    {
      data: "State",
      title: $_("PollingList.State"),
      width: "8%",
      render: renderState,
    },
    {
      data: "NodeID",
      title: $_("PollingList.Node"),
      width: "14%",
      render: (id: any) => nodes[id] ? nodes[id].Name : "???",
    },
    {
      data: "Name",
      title: $_("PollingList.Name"),
      width: "23%",
    },
    {
      data: "Level",
      title: $_("PollingList.Level"),
      width: "8%",
      render: renderState,
    },
    {
      data: "Type",
      title: $_("PollingList.Type"),
      width: "8%",
      render: renderPollingType,
    },
    {
      data: "LogMode",
      title: $_("PollingList.LogMode"),
      width: "7%",
      render: getLogModeName,
    },
    {
      data: "FailAction",
      title: $_('PollingList.FailAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
      searchable: false,
    },
    {
      data: "RepairAction",
      title: $_('PollingList.RepairAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
      searchable: false,
    },
    {
      data: "LastTime",
      title: $_("PollingList.LastTime"),
      width: "12%",
      render: renderTime,
      searchable: false,
    },
  ];

  onMount(() => {
    refresh();
  });

  let prevLang = "";
  $: {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && table) {
      showTable();
    }
    prevLang = lang;
  }

  const saveCSV = () => {
    ExportPollings("csv");
  };

  const saveExcel = () => {
    ExportPollings("excel");
  };
</script>

<div class="flex flex-col max-h-[calc(100vh-50px)] overflow-y-auto pb-2">
  <!-- KPI Summary Cards -->
  <div class="mx-2 mt-2 grid grid-cols-2 md:grid-cols-4 gap-3">
    <button
      type="button"
      onclick={() => setStateFilter("all")}
      class="text-left text-gray-800 dark:text-gray-100 cursor-pointer rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm transition hover:border-cyan-400 dark:hover:border-cyan-500"
    >
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiLanCheck} size={0.9} class="text-cyan-500" />
        {$_("PollingList.TotalPollings") || "総ポーリング数"}
      </div>
      <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-gray-100">{kpi.total.toLocaleString()}</div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("PollingList.NormalRate") || "正常率"}: {kpi.normalRate}%
        {#if kpi.off > 0}
          / {$_("Ts.Stop")}: {kpi.off.toLocaleString()}
        {/if}
      </div>
    </button>

    <button
      type="button"
      onclick={() => setStateFilter("normal")}
      class="text-left text-gray-800 dark:text-gray-100 cursor-pointer rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm transition hover:border-emerald-400 dark:hover:border-emerald-500 {stateFilter === 'normal'
        ? 'ring-2 ring-emerald-500/80'
        : ''}"
    >
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiCheckCircleOutline} size={0.9} class="text-emerald-500" />
        {$_("PollingList.NormalPollings") || "正常"}
      </div>
      <div class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">
        {kpi.normal.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Ts.Normal")}: {kpi.normal.toLocaleString()} / {kpi.total.toLocaleString()}
      </div>
    </button>

    <button
      type="button"
      onclick={() => setStateFilter("alert")}
      class="text-left text-gray-800 dark:text-gray-100 cursor-pointer rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm transition hover:border-rose-400 dark:hover:border-rose-500 {stateFilter === 'alert'
        ? 'ring-2 ring-rose-500/80'
        : ''}"
    >
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiAlertCircleOutline} size={0.9} class="text-rose-500" />
        {$_("PollingList.AlertPollings") || "異常 (重度/軽度)"}
      </div>
      <div
        class="mt-1 text-2xl font-bold {kpi.alertCount > 0
          ? 'text-rose-600 dark:text-rose-400'
          : 'text-emerald-600 dark:text-emerald-400'}"
      >
        {kpi.alertCount.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Ts.High")}: {kpi.high.toLocaleString()} / {$_("Ts.Low")}: {kpi.low.toLocaleString()}
      </div>
    </button>

    <button
      type="button"
      onclick={() => setStateFilter("warn")}
      class="text-left text-gray-800 dark:text-gray-100 cursor-pointer rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm transition hover:border-amber-400 dark:hover:border-amber-500 {stateFilter === 'warn'
        ? 'ring-2 ring-amber-500/80'
        : ''}"
    >
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiAlertOutline} size={0.9} class="text-amber-500" />
        {$_("PollingList.WarnPollings") || "注意・その他"}
      </div>
      <div
        class="mt-1 text-2xl font-bold {kpi.warnOtherCount > 0
          ? 'text-amber-500 dark:text-amber-400'
          : 'text-emerald-600 dark:text-emerald-400'}"
      >
        {kpi.warnOtherCount.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Ts.Warn")}: {kpi.warn.toLocaleString()} / {$_("Ts.Repair")}: {kpi.repair.toLocaleString()} / {$_("PollingList.Unknown") || "不明"}: {kpi.unknown.toLocaleString()}
      </div>
    </button>
  </div>
  <div class="mx-2 mt-2">
    <div>
      <table id="pollingListTable" class="display compact" style="width:99%"></table>
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 py-2">
    <GradientButton shadow color="blue" type="button" onclick={add} size="xs">
      <Icon path={icons.mdiPlus} size={1} />
      {$_("PollingList.Add")}
    </GradientButton>
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={edit}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("PollingList.Edit")}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        onclick={copy}
        size="xs"
      >
        <Icon path={icons.mdiContentCopy} size={1} />
        {$_("PollingList.Copy")}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        onclick={exportTemp}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_('PollingLis.Export')}
      </GradientButton>
      <GradientButton
        shadow
        color="green"
        type="button"
        onclick={report}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("PollingList.Report")}
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={deletePollingLogs}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_('PollingList.DeleteLogs')}
      </GradientButton>
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={deletePollings}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("PollingList.Delete")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="lime"
      type="button"
      onclick={saveCSV}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      onclick={saveExcel}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("PollingList.Reload")}
    </GradientButton>
  </div>
</div>

<Polling
  bind:show={showEditPolling}
  nodeID=""
  pollingID={selectedPolling}
  on:close={(e) => {
    refresh();
  }}
/>

<Polling
  bind:show={showCopyPolling}
  nodeID=""
  pollingID=""
  {pollingTmp}
  on:close={(e) => {
    refresh();
  }}
/>

<AddPolling
  bind:show={showAddPolling}
  nodeID=""
  on:close={(e) => {
    refresh();
  }}
/>

<PollingReport bind:show={showPollingReport} id={selectedPolling} />
