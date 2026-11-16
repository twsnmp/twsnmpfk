<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton, Dropdown, DropdownItem } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetNodes,
    DeleteNodes,
    ExportNodes,
    CheckPolling,
    WakeOnLan,
    GetMapConf,
  } from "../../wailsjs/go/main/App";
  import { getTableLang, renderNodeState, renderIP } from "./common";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import Ping from "./Ping.svelte";
  import MIBBrowser from "./MIBBrowser.svelte";
  import GNMITool from "./GNMITool.svelte";
  import NodeDiagnoseDialog from "./NodeDiagnoseDialog.svelte";
  import NodeAutoDetectDialog from "./NodeAutoDetectDialog.svelte";
  import MapList from "./MapList.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { currentLang } from "../i18n/i18n";

  let data: any = [];
  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let showMapList = false;
  let showAIDiagnose = false;
  let showAutoDetect = false;
  let selectedNode = "";
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

    for (const n of data) {
      switch (n.State) {
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
      alertCount: high + low,
      warnOtherCount: warn + repair + unknown,
      normalRate: total > 0 ? ((normal / total) * 100).toFixed(1) : "0.0",
    };
  };

  const getFilteredData = () => {
    switch (stateFilter) {
      case "normal":
        return data.filter((n: any) => n.State === "normal");
      case "alert":
        return data.filter((n: any) => n.State === "high" || n.State === "low");
      case "warn":
        return data.filter(
          (n: any) =>
            n.State !== "normal" && n.State !== "high" && n.State !== "low"
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

  let showPing: boolean = false;
  let showMibBr: boolean = false;
  let showGNMITool: boolean = false;
  let actionOpen: boolean = false;

  const showTable = () => {
    selectedCount = 0;
    const el = document.getElementById("nodeListTable");
    if (!el) {
      return;
    }
    if (table && DataTable.isDataTable("#nodeListTable")) {
      table.off("select");
      table.off("deselect");
      table.destroy();
      table = undefined;
    }
    el.innerHTML = "";
    table = new DataTable("#nodeListTable", {
      destroy: true,
      columns: getColumns(),
      data: getFilteredData(),
      stateSave: true,
      order: [
        [0, "asc"],
        [2, "asc"],
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
    const nodes = await GetNodes();
    data = [];
    for (const k in nodes) {
      data.push(nodes[k]);
    }
    updateKPI();
    showTable();
  };

  const edit = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showEditNode = true;
  };

  const report = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showNodeReport = true;
  };

  const aiDiagnose = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showAIDiagnose = true;
  };

  const polling = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showPolling = true;
  };

  const ping = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showPing = true;
  };

  const MIBBr = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showMibBr = true;
  };

  const gNMITool = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showGNMITool = true;
  };

  const doWakeOnLan = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    WakeOnLan(selectedNode);
    actionOpen = false;
  };

  const autoDetect = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showAutoDetect = true;
    actionOpen = false;
  };

  const deleteNodes = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    const ids = selected.toArray();
    await DeleteNodes(ids);
    const idSet = new Set(ids);
    data = data.filter((n: any) => !idSet.has(n.ID));
    updateKPI();
    selectedCount = 0;
    table.rows({ selected: true }).remove().draw();
  };

  const check = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID").toArray();
    if (selected.length < 1) {
      return;
    }
    await Promise.all(selected.map((n: any) => CheckPolling(n)));
    refresh();
  };

  const getColumns = () => [
    {
      data: "State",
      title: $_("NodeList.State"),
      width: "10%",
      render: renderNodeState,
    },
    {
      data: "Name",
      title: $_("NodeList.Name"),
      width: "15%",
    },
    {
      data: "IP",
      title: $_("NodeList.IPAddress"),
      width: "10%",
      render: renderIP,
    },
    {
      data: "MAC",
      title: $_("NodeList.MACAddress"),
      width: "10%",
    },
    {
      data: "Vendor",
      title: $_("NodeList.Vendor"),
      width: "15%",
    },
    {
      data: "Descr",
      title: $_("NodeList.Descr"),
      width: "30%",
    },
  ];

  let hasAI = false;

  onMount(async () => {
    const conf = await GetMapConf();
    hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
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

  const checkAll = async () => {
    await CheckPolling("all");
    refresh();
  };

  const saveCSV = () => {
    ExportNodes("csv");
  };

  const saveExcel = () => {
    ExportNodes("excel");
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
        <Icon path={icons.mdiServerNetwork} size={0.9} class="text-cyan-500" />
        {$_("NodeList.TotalNodes") || "総ノード数"}
      </div>
      <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-gray-100">{kpi.total.toLocaleString()}</div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("NodeList.NormalRate") || "正常率"}: {kpi.normalRate}%
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
        {$_("NodeList.NormalNodes") || "正常"}
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
        {$_("NodeList.AlertNodes") || "異常 (重度/軽度)"}
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
        {$_("NodeList.WarnNodes") || "注意・その他"}
      </div>
      <div
        class="mt-1 text-2xl font-bold {kpi.warnOtherCount > 0
          ? 'text-amber-500 dark:text-amber-400'
          : 'text-emerald-600 dark:text-emerald-400'}"
      >
        {kpi.warnOtherCount.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Ts.Warn")}: {kpi.warn.toLocaleString()} / {$_("Ts.Repair")}: {kpi.repair.toLocaleString()} / {$_("NodeList.Unknown") || "不明"}: {kpi.unknown.toLocaleString()}
      </div>
    </button>
  </div>
  <div class="mx-2 mt-2">
    <div>
      <table id="nodeListTable" class="display compact" style="width:99%"></table>
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={edit}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("NodeList.Edit")}
      </GradientButton>
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={polling}
        size="xs"
      >
        <Icon path={icons.mdiLanCheck} size={1} />
        {$_("NodeList.Polling")}
      </GradientButton>
      <GradientButton
        shadow
        color="green"
        type="button"
        onclick={report}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("NodeList.Report")}
      </GradientButton>
      {#if hasAI}
        <GradientButton
          shadow
          color="pink"
          type="button"
          onclick={aiDiagnose}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_("NodeList.AIDiagnose")}
        </GradientButton>
      {/if}
      <GradientButton
        >{$_("NodeList.Action")}<Icon
          path={icons.mdiChevronDown}
          size={1}
        /></GradientButton
      >
      <Dropdown bind:isOpen={actionOpen} class="w-44 bg-white dark:bg-gray-700 text-gray-900 dark:text-white list-none">
        <DropdownItem class="text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-600" onclick={ping}>PING</DropdownItem>
        <DropdownItem class="text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-600" onclick={MIBBr}>{$_("Map.MIBBrowser")}</DropdownItem>
        <DropdownItem class="text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-600" onclick={gNMITool}
          >{$_("GNMITool.gNMITool")}</DropdownItem
        >
        <DropdownItem class="text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-600" onclick={doWakeOnLan}>Wake On Lan</DropdownItem>
        <DropdownItem class="text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-600" onclick={autoDetect}>{$_("Node.AutoDetect") || "種別自動判定"}</DropdownItem>
      </Dropdown>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={deleteNodes}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("NodeList.Delete")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        onclick={check}
        size="xs"
      >
        <Icon path={icons.mdiCheck} size={1} />
        {$_("NodeList.ReCheck")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="green"
      type="button"
      onclick={()=> {
        showMapList = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiListBox} size={1} />
      {$_('NodeList.MapItems')}
    </GradientButton>
    <GradientButton
      shadow
      color="teal"
      type="button"
      onclick={checkAll}
      size="xs"
    >
      <Icon path={icons.mdiCheckAll} size={1} />
      {$_("NodeList.CheckAll")}
    </GradientButton>
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
      {$_("NodeList.Reload")}
    </GradientButton>
  </div>
</div>

<Node
  bind:show={showEditNode}
  nodeID={selectedNode}
  on:close={(e) => {
    refresh();
  }}
/>

<NodeReport bind:show={showNodeReport} id={selectedNode} />

<NodeDiagnoseDialog bind:show={showAIDiagnose} nodeID={selectedNode} />

<NodePolling bind:show={showPolling} nodeID={selectedNode} />

<Ping bind:show={showPing} nodeID={selectedNode} />

<MIBBrowser bind:show={showMibBr} nodeID={selectedNode} />

<GNMITool bind:show={showGNMITool} nodeID={selectedNode} />

<MapList bind:show={showMapList} />

<NodeAutoDetectDialog
  bind:show={showAutoDetect}
  nodeID={selectedNode}
  on:applied={() => {
    refresh();
  }}
/>

