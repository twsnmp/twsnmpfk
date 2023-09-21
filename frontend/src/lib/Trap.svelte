<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount,tick,onDestroy } from "svelte";
  import { GetTraps, ExportTraps, GetDefaultPolling } from "../../wailsjs/go/main/App";
  import {
    renderTime,
    getTableLang,
  } from "./common";
  import {showLogCountChart,resizeLogCountChart} from "./chart/logcount";
  import TrapReport from "./TrapReport.svelte";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { datastore } from "wailsjs/go/models";

  let data = [];
  let logs = [];
  let showReport = false;
  let table = undefined;
  let selectedCount = 0;
  let showPolling = false;

  const showTable = () => {
    if (table) {
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#table", {
      columns: columns,
      data: data,
      order:[[0,"desc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  }

  const refresh = async () => {
    logs = await GetTraps();
    data = [];
    for (let i =0; i < logs.length;i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
  };

  const showChart = async () => {
    await tick();
    showLogCountChart("chart",data,zoomCallBack);
  }

  const zoomCallBack = (st:number, et:number) => {
    data = [];
    for(let i = logs.length -1 ; i >= 0;i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
    showTable();
  };

  const columns = [
    {
      data: "Time",
      title: "日時",
      width: "20%",
      render: renderTime,
    },
    {
      data: "FromAddress",
      title: "送信元",
      width: "15%",
    },
    {
      data: "TrapType",
      title: "タイプ",
      width: "15%",
    },
    {
      data: "Variables",
      title: "変数",
      width: "50%",
    },
  ];

  onMount(() => {
    refresh();
  });

  onDestroy(()=>{
    if(table) {
      table.destroy();
      table = undefined;
    }
  });

  const saveCSV = () => {
    ExportTraps("csv");
  }

  const saveExcel = () => {
    ExportTraps("excel");
  }

  let polling : datastore.PollingEnt | undefined = undefined;
  const watch = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length !=1 ) {
      return;
    }
    let ip = d[0].FromAddress;
    const a = ip.split("(");
    if (a.length > 1) {
      ip = a[0];
    }
    polling = await GetDefaultPolling(ip);
    polling.Name = `${d[0].TrapType} TRAP監視`; 
    polling.Type = "trap";
    polling.Mode = "count";
    polling.Script = "count < 1";
    polling.Params = d[0].FromAddress;
    polling.Filter = d[0].TrapType;
    showPolling = true;
  }

</script>

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;"></div>
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <Button color="green" type="button" on:click={watch} size="xs">
        <Icon path={icons.mdiEye} size={1} />
        ポーリング
      </Button>
    {/if}
    <Button color="blue" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </Button>
    <Button color="blue" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </Button>
    <Button type="button" color="green" on:click={() => {showReport=true}} size="xs">
      <Icon path={icons.mdiChartPie} size={1} />
      レポート
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>

{#if showReport}
  <TrapReport
   {logs}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

{#if showPolling}
  <Polling
   pollingTmp={polling}
    on:close={() => {
      showPolling = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
