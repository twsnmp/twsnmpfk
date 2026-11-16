<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    ExportArpTable,
    GetArpLogs,
    GetArpTable,
    GetNodes,
    ResetArpTable,
    DeleteArpEnt,
    GetIPAM,
    GetMapConf,
  } from "../../wailsjs/go/main/App";
  import {
    getTableLang,
    renderTime,
    renderTimeUnix,
    getStateColor,
    getStateIcon,
  } from "./common";
  import AddressReport from "./AddressReport.svelte";
  import AddressInfo from "./AddressInfo.svelte";
  import StunInfo from "./StunInfo.svelte";
  import AddressAIDialog from "./AddressAIDialog.svelte";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { currentLang } from "../i18n/i18n";
  import { copyText } from "svelte-copy";

  let arp: any = [];
  let ipam: any = [];
  let nodes :any= undefined;
  let showReport = false;
  let arpTable :any = undefined;
  let newIP = new Map();
  let changeIP:any = new Map();
  let changeMAC:any  = new Map();
  let selectedIP = "";
  let selectedNodeID = "";
  let selectedCount =0;

  let showEditNode = false;
  let showAddNode = false;
  let showNodeReport = false;
  let showAddressInfo = false;
  let showStunInfo = false;
  let showAddressAI = false;
  let hasAI = false;
  let stateFilter: "all" | "normal" | "alert" | "unregistered" = "all";

  let kpi = {
    total: 0,
    normal: 0,
    dup: 0,
    ipChange: 0,
    macChange: 0,
    alertCount: 0,
    unregistered: 0,
    registered: 0,
    uniqueVendors: 0,
    normalRate: "0.0",
  };

  const updateKPI = () => {
    let normal = 0;
    let dup = 0;
    let ipChange = 0;
    let macChange = 0;
    let unregistered = 0;
    let registered = 0;
    const vendors = new Set<string>();

    for (const e of arp) {
      switch (Number(e.State)) {
        case 0:
          dup++;
          break;
        case 1:
          ipChange++;
          break;
        case 2:
          macChange++;
          break;
        default:
          normal++;
          break;
      }
      if (e.NodeID && nodes && nodes[e.NodeID]) {
        registered++;
      } else {
        unregistered++;
      }
      if (e.Vendor) {
        vendors.add(e.Vendor);
      }
    }

    const total = arp.length;
    kpi = {
      total,
      normal,
      dup,
      ipChange,
      macChange,
      alertCount: dup + ipChange + macChange,
      unregistered,
      registered,
      uniqueVendors: vendors.size,
      normalRate: total > 0 ? ((normal / total) * 100).toFixed(1) : "0.0",
    };
  };

  const getFilteredData = () => {
    switch (stateFilter) {
      case "normal":
        return arp.filter((e: any) => Number(e.State) === 3);
      case "alert":
        return arp.filter((e: any) => Number(e.State) !== 3);
      case "unregistered":
        return arp.filter((e: any) => !e.NodeID || !nodes || !nodes[e.NodeID]);
      default:
        return arp;
    }
  };

  const setStateFilter = (
    next: "all" | "normal" | "alert" | "unregistered"
  ) => {
    stateFilter = stateFilter === next ? "all" : next;
    if (arpTable) {
      selectedIP = selectedNodeID = "";
      selectedCount = 0;
      arpTable.clear().rows.add(getFilteredData()).draw();
    }
  };

  const showArpTable = () => {
    selectedIP = selectedNodeID = "";
    selectedCount = 0;
    const el = document.getElementById("arpTable");
    if (!el) {
      return;
    }
    if (arpTable && DataTable.isDataTable("#arpTable")) {
      arpTable.off("select");
      arpTable.off("deselect");
      arpTable.destroy();
      arpTable = undefined;
    }
    el.innerHTML = "";
    arpTable = new DataTable("#arpTable", {
      destroy: true,
      columns: getArpColumns(),
      pageLength: window.innerHeight > 800 ? 25 : 10,
      stateSave: true,
      data: getFilteredData(),
      order: [[0,"asc"]],
      language: getTableLang(),
      select: {
        style: "multi+shift",
      },
    });
    const updateArpSelection = () => {
      if (!arpTable) return;
      selectedIP = selectedNodeID = "";
      const d = arpTable.rows({ selected: true }).data();
      selectedCount = d.length;
      if (!d || d.length != 1) {
        return;
      }
      if (d[0].NodeID) {
        selectedNodeID = d[0].NodeID;
      }
      if (d[0].IP) {
        selectedIP = d[0].IP;
      }
    };
    arpTable.on("select", updateArpSelection);
    arpTable.on("deselect", updateArpSelection);
    updateArpSelection();
  };

  const renderNode = (id:any) => {
    return nodes[id] ? nodes[id].Name : "";
  };

  const renderArpIP = (ip: string, type: string) => {
    if (type == "sort") {
      return ip
        .split(".")
        .reduce((int, v) => Number(int) * 256 + Number(v) + "");
    }
    if (changeIP.has(ip)) {
      return `<span class="text-yellow-500">${ip}</span>`;
    }
    if (ip.startsWith("169.254.")) {
      return `<span class="text-red-500">${ip}</span>`;
    }
    return ip;
  };

  const renderArpMAC = (mac: string, type: string) => {
    if (type == "sort") {
      return mac;
    }
    if (changeMAC.has(mac)) {
      return `<span class="text-red-600">${mac}</span>`;
    }
    return mac;
  };

  const renderArpState = (s: string, type: string) => {
    if (type == "sort") {
      return Number(s);
    }
    switch (Number(s)) {
      case 0:
        return (
          `<span class="mdi ` +
          getStateIcon("high") +
          ` text-xs" style="color:` +
          getStateColor("high") +
          `;"></span><span class="ml-2">` +
          $_('Address.Dup') +
          `</span>`
        );
      case 1:
        return (
          `<span class="mdi ` +
          getStateIcon("low") +
          ` text-xs" style="color:` +
          getStateColor("low") +
          `;"></span><span class="ml-2">` +
          $_('Address.IPChange') +
          `</span>`
        );
      case 2:
        return (
          `<span class="mdi ` +
          getStateIcon("warn") +
          ` text-xs" style="color:` +
          getStateColor("warn") +
          `;"></span><span class="ml-2">` +
          $_('Address.MACChange') +
          `</span>`
        );
    }
    return (
      `<span class="mdi ` +
      getStateIcon("normal") +
      ` text-xs" style="color:` +
      getStateColor("normal") +
      `;"></span><span class="ml-2">` +
      $_('Address.Normal') +
      `</span>`
    );
  };

  const renderArpStateString = (s: string) => {
    switch (Number(s)) {
      case 0:
        return $_('Address.Dup');
      case 1:
        return $_('Address.IPChange');
      case 2:
        return $_('Address.MACChange');
    }
    return $_('Address.Normal');
  };

  const getArpColumns = () => [
    {
      data: "State",
      title: $_('Address.State'),
      width: "10%",
      render: renderArpState,
    },
    {
      data: "IP",
      title: $_("Address.IPAddress"),
      width: "10%",
      render: renderArpIP,
    },
    {
      data: "MAC",
      title: $_("Address.MACAddress"),
      width: "10%",
      render: renderArpMAC,
    },
    {
      data: "NodeID",
      title: $_("Address.NodeName"),
      width: "15%",
      render: renderNode,
    },
    {
      data: "Vendor",
      title: $_("Address.Vendor"),
      width: "25%",
    },
    {
      data: "Last",
      title: $_('Address.Last'),
      width: "10%",
      render: renderTime,
    },
    {
      data: "FirstTime",
      title: $_('Address.FirstTime'),
      width: "10%",
      render: renderTimeUnix,
    },
    {
      data: "LastTime",
      title: $_('Address.LastTime'),
      width: "10%",
      render: renderTimeUnix,
    },
  ];

  const refresh = async () => {
    ipam = await GetIPAM()
    nodes = await GetNodes();
    const arpLogs = await GetArpLogs();
    arp = await GetArpTable();
    changeIP.clear();
    newIP.clear();
    changeMAC.clear();
    arpLogs.reverse();
    for (let i = 0; i < arpLogs.length; i++) {
      if (arpLogs[i].State == "Change") {
        changeIP.set(arpLogs[i].IP, arpLogs[i].Time);
        changeMAC.set(arpLogs[i].NewMAC, true);
        changeMAC.set(arpLogs[i].OldMAC, true);
      } else {
        newIP.set(arpLogs[i].IP, arpLogs[i].Time);
      }
    }
    arp.forEach((e:any) => {
      e.State = changeIP.has(e.IP)
        ? 1
        : e.IP.startsWith("169.254.")
        ? 0
        : changeMAC.has(e.MAC)
        ? 2
        : 3;
      e.Last = changeIP.has(e.IP)
        ? changeIP.get(e.IP)
        : newIP.has(e.IP)
        ? newIP.get(e.IP)
        : 0;
    });
    updateKPI();
    showArpTable();
  };

  onMount(async () => {
    const conf = await GetMapConf();
    hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    refresh();
  });

  let prevLang = "";
  $: {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && arpTable) {
      showArpTable();
    }
    prevLang = lang;
  }

  const saveCSV = () => {
    ExportArpTable("csv");
  };

  const saveExcel = () => {
    ExportArpTable("excel");
  };

  let copied = false;
  const copy = () => {
    const selected = arpTable.rows({ selected: true }).data();
    const cols = getArpColumns();
    let s: string[] = [];
    const h = cols.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of cols) {
        switch (c.data){
        case "Last":
          row.push(renderTime(selected[i][c.data] || "", ""));
          break;
        case "State":
          row.push(renderArpStateString(selected[i][c.data] || ""));
          break;
        case  "NodeID":
          row.push(renderNode(selected[i][c.data] || ""));
          break;
        default:
          row.push(selected[i][c.data] || "");
        }
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };
 
  const reset = async () => {
    await ResetArpTable();
    refresh();
  };

  const deleteAddress = async () => {
    const selected = arpTable.rows({ selected: true }).data().pluck("IP");
    if (selected.length < 1) {
      return;
    }
    const ips = selected.toArray();
    await DeleteArpEnt(ips);
    const ipSet = new Set(ips);
    arp = arp.filter((e: any) => !ipSet.has(e.IP));
    updateKPI();
    arpTable.rows({ selected: true }).remove().draw();
    selectedNodeID = selectedIP = "";
    selectedCount  = 0;
  }
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
        <Icon path={icons.mdiIpNetworkOutline} size={0.9} class="text-cyan-500" />
        {$_("Address.TotalAddresses") || "総アドレス数"}
      </div>
      <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-gray-100">{kpi.total.toLocaleString()}</div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Address.NormalRate") || "正常率"}: {kpi.normalRate}%
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
        {$_("Address.Normal")}
      </div>
      <div class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">
        {kpi.normal.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Address.Normal")}: {kpi.normal.toLocaleString()} / {kpi.total.toLocaleString()}
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
        {$_("Address.AlertAddresses") || "異常 (重複/変化)"}
      </div>
      <div
        class="mt-1 text-2xl font-bold {kpi.alertCount > 0
          ? 'text-rose-600 dark:text-rose-400'
          : 'text-emerald-600 dark:text-emerald-400'}"
      >
        {kpi.alertCount.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Address.Dup")}: {kpi.dup.toLocaleString()} / {$_("Address.IPChange")}: {kpi.ipChange.toLocaleString()} / {$_("Address.MACChange")}: {kpi.macChange.toLocaleString()}
      </div>
    </button>

    <button
      type="button"
      onclick={() => setStateFilter("unregistered")}
      class="text-left text-gray-800 dark:text-gray-100 cursor-pointer rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm transition hover:border-amber-400 dark:hover:border-amber-500 {stateFilter === 'unregistered'
        ? 'ring-2 ring-amber-500/80'
        : ''}"
    >
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiHelpCircleOutline} size={0.9} class="text-amber-500" />
        {$_("Address.UnregisteredAddresses") || "未登録アドレス"}
      </div>
      <div
        class="mt-1 text-2xl font-bold {kpi.unregistered > 0
          ? 'text-amber-500 dark:text-amber-400'
          : 'text-emerald-600 dark:text-emerald-400'}"
      >
        {kpi.unregistered.toLocaleString()}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        {$_("Address.Registered") || "登録済"}: {kpi.registered.toLocaleString()} / {$_("Address.Vendor")}: {kpi.uniqueVendors.toLocaleString()}
      </div>
    </button>
  </div>
  <div class="mx-2 mt-2">
    <div>
      <table id="arpTable" class="display compact" style="width:99%"></table>
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedNodeID}
      <GradientButton
        shadow
        color="green"
        type="button"
        onclick={() => (showNodeReport = true)}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("Address.NodeInfo")}
      </GradientButton>
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={() => (showEditNode = true)}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("Address.EditNode")}
      </GradientButton>
    {/if}

    {#if selectedIP && !selectedNodeID}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={() => {
          showAddNode = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiPlus} size={1} />
        {$_("Address.AddNode")}
      </GradientButton>
    {/if}

    {#if selectedCount > 0 }
      <GradientButton
        shadow
        color="cyan"
        type="button"
        onclick={copy}
        size="xs"
      >
        {#if copied}
          <Icon path={icons.mdiCheck} size={1} />
        {:else}
          <Icon path={icons.mdiContentCopy} size={1} />
        {/if}
        Copy
      </GradientButton>
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={deleteAddress}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_('Address.Delete')}
      </GradientButton>
    {/if}
    {#if arp.length > 0}
      <GradientButton
        type="button"
        color="green"
        onclick={() => {
          showReport = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiChartPie} size={1} />
        {$_("Address.Report")}
      </GradientButton>
      {/if}
      {#if hasAI}
      <GradientButton
        type="button"
        color="pink"
        shadow
        onclick={() => {
          showAddressAI = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiBrain} size={1} />
        {$_("Address.AIExplain")}
      </GradientButton>
      {/if}
      <GradientButton
        type="button"
        color="green"
        onclick={() => {
          showAddressInfo = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiInbox} size={1} />
        {$_('Address.AddressInfo')}
      </GradientButton>
      <GradientButton
        type="button"
        color="blue"
        onclick={() => {
          showStunInfo = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiWan} size={1} />
        {$_('Address.StunInfo') || 'STUN'}
      </GradientButton>
    <GradientButton shadow color="red" type="button" onclick={reset} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Address.Clear")}
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
      {$_("Address.Reload")}
    </GradientButton>
  </div>
</div>

<AddressReport
  bind:show ={showReport}
  {arp}
  {changeIP}
  {changeMAC}
  {ipam}
  {nodes}
/>

<AddressInfo
  bind:show = {showAddressInfo}
/>

<StunInfo
  bind:show = {showStunInfo}
/>

<Node
  bind:show={showAddNode}
  ip={selectedIP}
  posX={100}
  posY={120}
  on:close={(e) => {
    refresh();
  }}
/>

<Node
  bind:show={showEditNode}
  nodeID={selectedNodeID}
  on:close={(e) => {
    refresh();
  }}
/>

<NodeReport
  bind:show={showNodeReport}
  id={selectedNodeID}
/>

<AddressAIDialog
  bind:show={showAddressAI}
/>

