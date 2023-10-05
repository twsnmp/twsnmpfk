<script lang="ts">
  import { GradientButton, Modal, Spinner } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick, onDestroy } from "svelte";
  import {
    GetEventLogs,
    ExportEventLogs,
    DeleteAllEventLogs,
  } from "../../wailsjs/go/main/App";
  import { renderState, renderTime, getTableLang } from "./common";
  import { showLogLevelChart, resizeLogLevelChart } from "./chart/loglevel";
  import EventLogReport from "./EventLogReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let data = [];
  let logs = [];
  let showReport = false;
  let showLoading = false;
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
      order: [[1, "desc"]],
    });
  };

  const refresh = async () => {
    showLoading = true;
    logs = await GetEventLogs("");
    data = [];
    for (let i = 0; i < logs.length; i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
    showLoading = false;
  };

  const showChart = async () => {
    await tick();
    showLogLevelChart("chart", logs, zoomCallBack);
  };

  const zoomCallBack = (st: number, et: number) => {
    data = [];
    if (!st) {
      for (let i = 0; i < logs.length; i++) {
        data.push(logs[i]);
      }
      logs.reverse();
      return;
    }
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
      title: $_("EventLog.Level"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("EventLog.Time"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Type",
      title: $_("EventLog.Type"),
      width: "10%",
    },
    {
      data: "NodeName",
      title: $_("EventLog.NodeName"),
      width: "15%",
    },
    {
      data: "Event",
      title: $_("EventLog.Event"),
      width: "50%",
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
    ExportEventLogs("csv");
  };

  const saveExcel = () => {
    ExportEventLogs("excel");
  };

  const deleteAll = async () => {
    if (await DeleteAllEventLogs()) {
      refresh();
    }
  };
</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;" />
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="green"
      on:click={() => {
        showReport = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiChartPie} size={1} />
      {$_("EventLog.Report")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("EventLog.DeleteAllLogs")}
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveCSV}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveExcel}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("EventLog.Reload")}
    </GradientButton>
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

<Modal bind:open={showLoading} size="sm" permanent class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("Syslog.Loading")} </span>
  </div>
</Modal>

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
