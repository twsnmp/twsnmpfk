<script lang="ts">
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Spinner,
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetTraps,
    ExportTraps,
    GetDefaultPolling,
    DeleteAllTraps,
  } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang } from "./common";
  import { showLogCountChart, resizeLogCountChart } from "./chart/logcount";
  import TrapReport from "./TrapReport.svelte";
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
  let showFilter = false;
  let from = "";
  let trapType = "";
  let showLoading = false;

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
      order: [[0, "desc"]],
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
    showLoading = true;
    logs = await GetTraps(from, trapType);
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
    showLogCountChart("chart", data, zoomCallBack);
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
      data: "Time",
      title: $_("Trap.Time"),
      width: "20%",
      render: renderTime,
    },
    {
      data: "FromAddress",
      title: $_("Trap.FromAddress"),
      width: "15%",
    },
    {
      data: "TrapType",
      title: $_("Trap.TrapType"),
      width: "15%",
    },
    {
      data: "Variables",
      title: $_("Trap.Variables"),
      width: "50%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const saveCSV = () => {
    ExportTraps("csv",from,trapType);
  };

  const saveExcel = () => {
    ExportTraps("excel",from,trapType);
  };

  let polling: datastore.PollingEnt | undefined = undefined;
  const watch = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
      return;
    }
    let ip = d[0].FromAddress;
    const a = ip.split("(");
    if (a.length > 1) {
      ip = a[0];
    }
    polling = await GetDefaultPolling(ip);
    polling.Name = `${d[0].TrapType}`;
    polling.Type = "trap";
    polling.Mode = "count";
    polling.Script = "count < 1";
    polling.Params = d[0].FromAddress;
    polling.Filter = d[0].TrapType;
    showPolling = true;
  };

  const deleteAll = async () => {
    if (await DeleteAllTraps()) {
      refresh();
    }
  };
</script>

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart"/>
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
        {$_("Trap.Polling")}
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
      {$_("Trap.Filter")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Traps.DeleteAllLogs")}
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
        {$_("Trap.Report")}
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
      {$_("Trap.Reload")}
    </GradientButton>
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

<Modal bind:open={showFilter} size="sm" permanent class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Trap.Filter")}
    </h3>
    <Label class="space-y-2">
      <span>{$_("Trap.FromAddress")} </span>
      <Input bind:value={from} size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>{$_("Trap.TrapType")}</span>
      <Input bind:value={trapType} size="sm" />
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
        {$_("Trap.Search")}
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
        {$_("Trap.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal bind:open={showLoading} size="sm" permanent class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("Syslog.Loading")} </span>
  </div>
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
