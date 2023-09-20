<script lang="ts">
  import { Button } from "flowbite-svelte";
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
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  let arp = [];
  let nodes = undefined;
  let arpLogs = [];
  let arpLogData = [];
  let showReport = false;
  let arpTable = undefined;
  let arpLogTable = undefined;
  let selectedArpCount = 0;
  let changeMAC = new Map();
  let changeIP = new Map();

  const showArpTable = () => {
    if (arpTable) {
      arpTable.destroy();
      arpTable = undefined;
    }
    selectedArpCount = 0;
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
      selectedArpCount = arpTable.rows({ selected: true }).count();
    });
    arpTable.on("deselect", () => {
      selectedArpCount = arpTable.rows({ selected: true }).count();
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
      title: "IPアドレス",
      width: "15%",
      render: renderArpIP,
    },
    {
      data: "MAC",
      title: "MACアドレス",
      width: "20%",
      render: renderArpMAC,
    },
    {
      data: "NodeID",
      title: "ノード名",
      width: "20%",
      render: renderNode,
    },
    {
      data: "Vendor",
      title: "ベンダー",
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
      title: "状態",
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: "日時",
      width: "30%",
      render: renderTime,
    },
    {
      data: "IP",
      title: "IPアドレス",
      width: "20%",
    },
    {
      data: "NewMAC",
      title: "新MACアドレス",
      width: "20%",
    },
    {
      data: "OldMAC",
      title: "旧MACアドレス",
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
    <Button color="blue" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </Button>
    <Button color="blue" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </Button>
    <Button
      type="button"
      color="green"
      on:click={() => {
        showReport = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiChartPie} size={1} />
      レポート
    </Button>
    <Button color="red" type="button" on:click={reset} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      クリア
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
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


<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
