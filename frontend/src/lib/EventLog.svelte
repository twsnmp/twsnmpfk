<script lang="ts">
  import { Button} from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount,tick,onDestroy } from "svelte";
  import { GetEventLogs, ExportEventLogs } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getTableLang,
  } from "./common";
  import {showLogLevelChart,resizeLogLevelChart} from "./chart/loglevel";
  import EventLogReport from "./EventLogReport.svelte";
  let data = [];
  let logs = [];
  let showReport = false;
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let table = undefined;

  const showTable = () => {
    if (table) {
      table.destroy();
      table = undefined;
    }
    table = new DataTable("#table", {
      columns: columns,
      data: data,
      language: getTableLang(),
      order: [[1,"desc"]],
    });
  }

  const refresh = async () => {
    logs = await GetEventLogs("");
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
    showLogLevelChart("chart",logs,zoomCallBack);
  };
 
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
      data: "Level",
      title:  $_('EventLog.Level') ,
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_('EventLog.Time'),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Type",
      title: $_('EventLog.Type'),
      width: "10%",
    },
    {
      data: "NodeName",
      title: $_('EventLog.NodeName'),
      width: "15%",
    },
    {
      data: "Event",
      title: $_('EventLog.Event'),
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
    ExportEventLogs("csv");
  }

  const saveExcel = () => {
    ExportEventLogs("excel");
  }

</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;"></div>
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
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
      { $_('EventLog.Report') }
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('EventLog.Reload') }
    </Button>
  </div>
</div>

{#if showReport}
  <EventLogReport
   {logs}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
