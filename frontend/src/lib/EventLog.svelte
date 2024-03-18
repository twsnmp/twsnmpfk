<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Spinner,
    Label,
    Select,
    Input,
    Button,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
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
  import type { main } from "wailsjs/go/models";
  import { CodeJar } from "@novacbn/svelte-codejar";
  import Prism from "prismjs";
  import "prismjs/components/prism-regex";

  let data: any = [];
  let logs: any = [];
  let table: any = undefined;
  let showReport = false;
  let showLoading = false;
  let showFilter = false;

  const filter: main.EventLogFilterEnt = {
    NodeID: "",
    Start: "",
    End: "",
    NodeName: "",
    EventType: "",
    Event: "",
    Level: 0,
  };

  const levelList = [
    { name: $_("EventLog.All"), value: 0 },
    { name: $_("EventLog.Warn"), value: 1 },
    { name: $_("EventLog.Low"), value: 2 },
    { name: $_("EventLog.High"), value: 3 },
  ];

  const showTable = () => {
    if (table && DataTable.isDataTable("#table")) {
      table.clear();
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
    logs = await GetEventLogs(filter);
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

  const saveCSV = () => {
    ExportEventLogs("csv", filter);
  };

  const saveExcel = () => {
    ExportEventLogs("excel", filter);
  };

  const deleteAll = async () => {
    if (await DeleteAllEventLogs()) {
      refresh();
    }
  };

  const highlight = (code: string, syntax: string | undefined) => {
    if (!syntax) {
      return "";
    }
    return Prism.highlight(code, Prism.languages[syntax], syntax);
  };
</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" />
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

<EventLogReport bind:show={showReport} {logs} />

<Modal bind:open={showLoading} size="sm" dismissable={false} class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("EventLog.Loading")} </span>
  </div>
</Modal>

<Modal bind:open={showFilter} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-2" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("EventLog.Filter")}
    </h3>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_('EventLog.Start')}</span>
        <Input type="datetime-local" bind:value={filter.Start} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_('EventLog.End')}</span>
        <Input type="datetime-local" bind:value={filter.End} size="sm" />
      </Label>
      <div class="flex">
        <Button
          class="!p-2 w-8 h-8 mt-6 ml-4"
          color="red"
          on:click={() => {
            filter.Start = "";
            filter.End = "";
          }}
        >
          <Icon path={icons.mdiCancel} size={1} />
        </Button>
      </div>
    </div>
    <div class="grid gap-2 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Level")}</span>
        <Select
          items={levelList}
          bind:value={filter.Level}
          placeholder={$_("EventLog.SelectLevel")}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Type")}</span>
        <CodeJar syntax="regex" {highlight} bind:value={filter.EventType}/>
      </Label>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("EventLog.NodeName")}</span>
      <CodeJar syntax="regex" {highlight} bind:value={filter.NodeName}/>
    </Label>
    <Label class="space-y-2 text-xs">
      <span>{$_("EventLog.Event")}</span>
      <CodeJar syntax="regex" {highlight} bind:value={filter.Event}/>
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
  #chart {
    min-height: 200px;
    height: 20vh;
    width: 95vw;
    margin: 0 auto;
  }
</style>
