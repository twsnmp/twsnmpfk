<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick, onDestroy } from "svelte";
  import { GetSyslogs, ExportSyslogs,GetDefaultPolling,AutoGrok } from "../../wailsjs/go/main/App";
  import { renderState, renderTime, getTableLang } from "./common";
  import { showLogLevelChart, resizeLogLevelChart } from "./chart/loglevel";
  import SyslogReport from "./SyslogReport.svelte";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { datastore } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";

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
      order: [1, "desc"],
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
  };

  const refresh = async () => {
    logs = await GetSyslogs();
    data = [];
    for (let i = 0; i < logs.length; i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
  };

  const showChart = async () => {
    await tick();
    showLogLevelChart("chart", logs, zoomCallBack);
  };

  const zoomCallBack = (st: number, et: number) => {
    data = [];
    for (let i = logs.length - 1; i >= 0; i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
    showTable();
  };

  const columns = [
    {
      data: "Level",
      title: $_('Syslog.Level'),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_('Syslog.Time'),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Host",
      title: $_('Syslog.Host'),
      width: "15%",
    },
    {
      data: "Type",
      title: $_('Syslog.Type'),
      width: "10%",
    },
    {
      data: "Tag",
      title: $_('Syslog.Tag'),
      width: "10%",
    },
    {
      data: "Message",
      title: $_('Syslog.Message'),
      width: "40%",
    },
  ];

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    if (table) {
      table.destroy();
      table = undefined;
    }
  });

  const saveCSV = () => {
    ExportSyslogs("csv");
  };

  const saveExcel = () => {
    ExportSyslogs("excel");
  };

  let polling : datastore.PollingEnt | undefined = undefined;

  const watch = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length !=1 ) {
      return;
    }
    polling = await GetDefaultPolling(d[0].Host);
    polling.Extractor = await AutoGrok(d[0].Message);
    if (polling.Extractor == "") {
      polling.Mode = "count";
      polling.Script = "count < 1";
    }
    polling.Name = `syslog`; 
    polling.Type = "syslog";
    polling.Filter = d[0].Type + " " + d[0].Tag;
    polling.Params = d[0].Host;
    showPolling = true;
  }

</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;" />
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <Button color="green" type="button" on:click={watch} size="xs">
        <Icon path={icons.mdiEye} size={1} />
        { $_('Syslog.Polling') }
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
    <Button
      type="button"
      color="green"
      on:click={() => {
        showReport = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiChartPie} size={1} />
      { $_('Syslog.Report') }
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('Syslog.Reload') }
    </Button>
  </div>
</div>

{#if showReport}
  <SyslogReport
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
