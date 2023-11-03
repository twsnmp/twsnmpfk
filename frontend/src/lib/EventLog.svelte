<script lang="ts">
  import { GradientButton, Modal, Spinner,Label,Select,Input } from "flowbite-svelte";
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
  let table = undefined;
  let showReport = false;
  let showLoading = false;
  let showFilter = false;
  let level = 0;
  let type = "";
  let node = "";
  let event = "";

  const levelList = [
    { name: $_("EventLog.All"), value: 0 },
    { name: $_("EventLog.Warn"), value: 1 },
    { name: $_("EventLog.Low"), value: 2 },
    { name: $_("EventLog.High"), value: 3 },
  ];

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
    logs = await GetEventLogs("",type,node,event,level);
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
    ExportEventLogs("csv",type,node,event,level);
  };

  const saveExcel = () => {
    ExportEventLogs("excel",type,node,event,level);
  };

  const deleteAll = async () => {
    if (await DeleteAllEventLogs()) {
      refresh();
    }
  };
</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart"/>
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      color="blue"
      type="button"
      on:click={() => (showFilter = true)}
      size="xs"
    >
      <Icon path={icons.mdiFilter} size={1} />
      {$_("Trap.Filter")}
    </GradientButton>
    {#if logs.length > 0}
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
    {/if}
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
    <span class="ml-2"> {$_("EventLog.Loading")} </span>
  </div>
</Modal>

<Modal bind:open={showFilter} size="sm" permanent class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("EventLog.Filter")}
    </h3>
    <Label class="space-y-2">
      <span>{$_("EventLog.Level")}</span>
      <Select
        items={levelList}
        bind:value={level}
        placeholder={$_("EventLog.SelectLevel")}
        size="sm"
      />
    </Label>
    <Label class="space-y-2">
      <span>{$_('EventLog.Type')}</span>
      <Input bind:value={type} size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>{$_('EventLog.NodeName')}</span>
      <Input bind:value={node} size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>{$_('EventLog.Event')}</span>
      <Input bind:value={event} size="sm" />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={() => {
          showFilter = false;
          refresh();
        }}
        size="xs"
      >
        <Icon path={icons.mdiSearchWeb} size={1} />
        {$_("EventLog.Search")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        on:click={() => {
          showFilter = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("EventLog.Calcel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<style>
  @import "../assets/css/jquery.dataTables.css";
  #chart {
    min-height: 200px;
    height: 20vh;
    width:  98vw;
    margin:  0 auto;
  }
</style>
