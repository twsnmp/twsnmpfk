<script lang="ts">
  import { GradientButton } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick, onDestroy } from "svelte";
  import {
    ExportArpTable,
    GetArpLogs,
    GetArpTable,
    GetNodes,
    ResetArpTable,
  } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, renderState } from "./common";
  import { showLogCountChart, resizeLogCountChart } from "./chart/logcount";
  import ArpReport from "./ArpReport.svelte";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';

  let arp = [];
  let nodes = undefined;
  let arpLogs = [];
  let arpLogData = [];
  let showReport = false;
  let arpTable = undefined;
  let arpLogTable = undefined;
  let changeMAC = new Map();
  let changeIP = new Map();
  let selectedIP = "";
  let selectedNodeID = "";

  let showEditNode = false;
  let showNodeReport = false;

  const showArpTable = () => {
    if (arpTable) {
      arpTable.destroy();
      arpTable = undefined;
    }
    selectedIP = selectedNodeID = "";
    arpTable = new DataTable("#arpTable", {
      columns: arpColumns,
      data: arp,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    arpTable.on("select", () => {
      selectedIP = selectedNodeID = "";
      const d  = arpTable.rows({ selected: true }).data();
      if (!d || d.length != 1) {
        return;
      }
      if (d[0].NodeID ) {
        selectedNodeID = d[0].NodeID;
      } else if (d[0].IP) {
        selectedIP = d[0].IP;
      }
    });
    arpTable.on("deselect", () => {
      const c  = arpTable.rows({ selected: true }).count();
      if(c != 1) {
        selectedIP = selectedNodeID = "";
      }
    });
  };

  const renderNode = (id) => {
    return nodes[id] ? nodes[id].Name : id;
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

  const arpColumns = [
    {
      data: "IP",
      title: $_('Arp.IPAddress'),
      width: "15%",
      render: renderArpIP,
    },
    {
      data: "MAC",
      title: $_('Arp.MACAddress'),
      width: "20%",
      render: renderArpMAC,
    },
    {
      data: "NodeID",
      title: $_('Arp.NodeName'),
      width: "20%",
      render: renderNode,
    },
    {
      data: "Vendor",
      title: $_('Arp.Vendor'),
      width: "45%",
      render: renderNode,
    },
  ];

  const showArpLogTable = () => {
    if (arpLogTable) {
      arpLogTable.destroy();
      arpLogTable = undefined;
    }
    arpLogTable = new DataTable("#arpLogTable", {
      columns: arpLogColumns,
      data: arpLogData,
      order: [[1, "desc"]],
      language: getTableLang(),
    });
  };

  const arpLogColumns = [
    {
      data: "State",
      title: $_('Arp.State'),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_('Arp.DateTime'),
      width: "30%",
      render: renderTime,
    },
    {
      data: "IP",
      title: $_('Arp.IPAddress'),
      width: "20%",
    },
    {
      data: "NewMAC",
      title: $_('Arp.NewMACAddress'),
      width: "20%",
    },
    {
      data: "OldMAC",
      title: $_('Arp.OldMACAddress'),
      width: "20%",
    },
  ];

  const refresh = async () => {
    nodes = await GetNodes();
    arpLogs = await GetArpLogs();
    arp = await GetArpTable();
    arpLogData = [];
    changeIP.clear();
    changeMAC.clear();

    for (let i = 0; i < arpLogs.length; i++) {
      arpLogData.push(arpLogs[i]);
      if (arpLogs[i].State == "Change") {
        changeIP.set(arpLogs[i].IP, true);
        changeMAC.set(arpLogs[i].NewMAC, true);
        changeMAC.set(arpLogs[i].oldMAC, true);
      }
    }
    arpLogs.reverse();
    showArpLogTable();
    showChart();
    showArpTable();
  };

  const showChart = async () => {
    await tick();
    showLogCountChart("chart", arpLogData, zoomCallBack);
  };

  const zoomCallBack = (st: number, et: number) => {
    arpLogData = [];
    for (let i = arpLogs.length - 1; i >= 0; i--) {
      if (arpLogs[i].Time >= st && arpLogs[i].Time <= et) {
        arpLogData.push(arpLogs[i]);
      }
    }
    showArpLogTable();
  };

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    if (arpTable) {
      arpTable.destroy();
      arpTable = undefined;
    }
    if (arpLogTable) {
      arpLogTable.destroy();
      arpLogTable = undefined;
    }
  });

  const saveCSV = () => {
    ExportArpTable("csv");
  };

  const saveExcel = () => {
    ExportArpTable("excel");
  };

  const reset = async() => {
    await ResetArpTable();
    refresh();
  }

</script>

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;" />
  <div class="mt-2 ml-2 mr-2 grid grid-cols-2 gap-2">
    <div>
      <table id="arpTable" class="display compact" style="width:99%" />
    </div>
    <div>
      <table id="arpLogTable" class="display compact" style="width:99%" />
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 mt-2">
    {#if selectedNodeID}
      <GradientButton shadow color="green" type="button" on:click={()=> showNodeReport= true} size="xs">
        <Icon path={icons.mdiChartBar} size={1} />
        { $_('Arp.NodeInfo') }
      </GradientButton>
    {/if}

    {#if selectedIP}
      <GradientButton shadow color="blue" type="button" on:click={()=>{showEditNode=true}} size="xs">
        <Icon path={icons.mdiPlus} size={1} />
        { $_('Arp.AddNode') }
      </GradientButton>
    {/if}

    <GradientButton
    type="button"
    color="green"
    on:click={() => {
      showReport = true;
    }}
      size="xs"
      >
      <Icon path={icons.mdiChartPie} size={1} />
      { $_('Arp.Report') }
    </GradientButton>
    <GradientButton shadow color="red" type="button" on:click={reset} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      { $_('Arp.Clear') }
    </GradientButton>
    <GradientButton shadow color="lime" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton shadow color="lime" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton shadow type="button" color="teal" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('Arp.Reload') }
    </GradientButton>
  </div>
</div>

{#if showReport}
  <ArpReport
   logs={arpLogs}
   {arp}
   {changeIP}
   {changeMAC}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

{#if showEditNode}
  <Node
    ip={selectedIP}
    posX={100}
    posY={120}
    on:close={(e) => {
      refresh();
      showEditNode = false;
    }}
  />
{/if}

{#if showNodeReport}
  <NodeReport
    id={selectedNodeID}
    on:close={(e) => {
      showNodeReport = false;
    }}
  />
{/if}


<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
