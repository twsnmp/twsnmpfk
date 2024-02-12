<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Select,
    Spinner,
    Button,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick, onDestroy } from "svelte";
  import {
    GetSyslogs,
    ExportSyslogs,
    GetDefaultPolling,
    AutoGrok,
    DeleteAllSyslog,
  } from "../../wailsjs/go/main/App";
  import { renderState, renderTime, getTableLang } from "./common";
  import { showLogLevelChart, resizeLogLevelChart } from "./chart/loglevel";
  import SyslogReport from "./SyslogReport.svelte";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { datastore, main } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";

  let data: any = [];
  let logs: any = [];
  let showReport = false;
  let table: any = undefined;
  let selectedCount = 0;
  let showPolling = false;
  let showFilter = false;
  let showLoading = false;

  const filter: main.SyslogFilterEnt = {
    Start: "",
    End: "",
    Severity: 6,
    Host: "",
    Tag: "",
    Message: "",
  };

  const levelList = [
    { name: $_("Syslog.All"), value: 7 },
    { name: $_("Syslog.Info"), value: 6 },
    { name: $_("Syslog.Warn"), value: 4 },
    { name: $_("Syslog.Low"), value: 3 },
    { name: $_("Syslog.High"), value: 2 },
  ];

  const showTable = () => {
    if (table && DataTable.isDataTable("#table")) {
      table.clear();
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
    filter.Severity *= 1;
    showLoading = true;
    logs = await GetSyslogs(filter);
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
      title: $_("Syslog.Level"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("Syslog.Time"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Host",
      title: $_("Syslog.Host"),
      width: "15%",
    },
    {
      data: "Type",
      title: $_("Syslog.Type"),
      width: "10%",
    },
    {
      data: "Tag",
      title: $_("Syslog.Tag"),
      width: "10%",
    },
    {
      data: "Message",
      title: $_("Syslog.Message"),
      width: "40%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const saveCSV = () => {
    ExportSyslogs("csv", filter);
  };

  const saveExcel = () => {
    ExportSyslogs("excel", filter);
  };

  let polling: datastore.PollingEnt | undefined = undefined;

  const watch = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
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
  };

  const deleteAll = async () => {
    if (await DeleteAllSyslog()) {
      refresh();
    }
  };
</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" />
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={watch}
        size="xs"
      >
        <Icon path={icons.mdiEye} size={1} />
        {$_("Syslog.Polling")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="blue"
      type="button"
      on:click={() => (showFilter = true)}
      size="xs"
    >
      <Icon path={icons.mdiFilter} size={1} />
      {$_("Syslog.Filter")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Syslog.DeleteAllLogs")}
    </GradientButton>
    {#if logs.length > 0}
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
        {$_("Syslog.Report")}
      </GradientButton>
    {/if}
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
      {$_("Syslog.Reload")}
    </GradientButton>
  </div>
</div>

<SyslogReport bind:show={showReport} {logs} />

<Polling bind:show={showPolling} pollingTmp={polling} />

<Modal bind:open={showFilter} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Syslog.Filter")}
    </h3>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Start")}</span>
        <Input type="datetime-local" bind:value={filter.Start} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.End")}</span>
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
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Level")}</span>
        <Select
          items={levelList}
          bind:value={filter.Severity}
          placeholder={$_("Syslog.SelectLevel")}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Host")}</span>
        <Input bind:value={filter.Host} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Tag")}</span>
        <Input bind:value={filter.Tag} size="sm" />
      </Label>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("Syslog.Message")}</span>
      <Input bind:value={filter.Message} size="sm" />
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
        {$_("Syslog.Search")}
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
        {$_("Syslog.Calcel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal bind:open={showLoading} size="sm" dismissable={false} class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("Syslog.Loading")} </span>
  </div>
</Modal>

<style>
  #chart {
    min-height: 200px;
    height: 20vh;
    width: 95vw;
    margin: 0 auto;
  }
</style>
