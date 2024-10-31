<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton, Modal, Spinner } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import { GetMonitorDatas, Backup } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, renderBytes, renderSpeed } from "./common";
  import {
    showMonitorResChart,
    showMonitorNetChart,
    showMonitorForecastChart,
    resizeMonitorChart,
  } from "./chart/system";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { backend, datastore } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";

  let logs: backend.MonitorDataEnt[] = [];
  let table : any = undefined;
  let showLoading = false;

  const showTable = () => {
    table = new DataTable("#systemTable", {
      destroy: true,
      stateSave: true,
      columns: columns,
      paging: false,
      searching: false,
      info: false,
      scrollY: "25vh",
      data: logs,
      order: [[0, "desc"]],
      language: getTableLang(),
    });
  };

  const refresh = async () => {
    showLoading = true;
    logs = await GetMonitorDatas();
    logs.reverse();
    showTable();
    showChart();
    showLoading = false;
  };

  const showChart = async () => {
    await tick();
    showMonitorResChart("resChart", logs);
    showMonitorNetChart("netChart", logs);
  };

  const renderPer = (v:any, t:any) => {
    if (t == "sort") {
      return v;
    }
    return v.toFixed(2) + "%";
  };

  const columns = [
    {
      data: "Time",
      title: $_("System.Time"),
      width: "10%",
      render: renderTime,
    },
    {
      data: "CPU",
      title: "CPU",
      width: "6%",
      render: renderPer,
    },
    {
      data: "Mem",
      title: $_("System.Memory"),
      width: "6%",
      render: renderPer,
    },
    {
      data: "MyCPU",
      title: "My CPU",
      width: "6%",
      render: renderPer,
    },
    {
      data: "MyMem",
      title: "My" + $_("System.Memory"),
      width: "6%",
      render: renderPer,
    },
    {
      data: "Swap",
      title: "Swap",
      width: "6%",
      render: renderPer,
    },
    {
      data: "Disk",
      title: $_("System.Disk"),
      width: "6%",
      render: renderPer,
    },
    {
      data: "Load",
      title: $_("System.Load"),
      width: "6%",
      render: (v:any) => v.toFixed(2),
    },
    {
      data: "Net",
      title: $_("System.Net"),
      width: "7%",
      render: renderSpeed,
    },
    {
      data: "Conn",
      title: $_("System.Conn"),
      width: "6%",
    },
    {
      data: "Proc",
      title: $_("System.Proc"),
      width: "6%",
    },
    {
      data: "NumGoroutine",
      title: "GOルーチン",
      width: "6%",
    },
    {
      data: "HeapAlloc",
      title: "Heap",
      width: "7%",
      render: renderBytes,
    },
    {
      data: "Sys",
      title: "Sys",
      width: "7%",
      render: renderBytes,
    },
    {
      data: "DBSize",
      title: $_("System.DBSize"),
      width: "7%",
      render: renderBytes,
    },
  ];

  onMount(() => {
    refresh();
  });

  const backup = async () => {
    Backup();
  };

  let showForecast = false;

  const forecast = async () => {
    showForecast = true;
    await tick();
    showMonitorForecastChart("forecast", logs);
  };

  const resizeChart = () => {
    resizeMonitorChart(showForecast);
  };
</script>

<svelte:window on:resize={resizeChart} />

<div class="flex flex-col">
  <div id="resChart"/>
  <div id="netChart"/>
  <div class="m-5 grow">
    <table id="systemTable" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="green"
      on:click={forecast}
      size="xs"
    >
      <Icon path={icons.mdiChartLine} size={1} />
      {$_("System.SizeForecast")}
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={backup}
      size="xs"
    >
      <Icon path={icons.mdiDatabaseArrowDown} size={1} />
      {$_("System.Backup")}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("System.Reload")}
    </GradientButton>
  </div>
</div>

<Modal bind:open={showForecast} size="xl" dismissable={false} class="w-full">
  <div id="forecast" />
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={() => {
        showForecast = false;
      }}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("System.Close")}
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showLoading} size="sm" dismissable={false} class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> $_('System.Loading') </span>
  </div>
</Modal>

<style>
  #resChart,
  #netChart{
    min-height: 200px;
    height: 25vh;
    width: 98%;
    margin: 0 auto;
  }
  #forecast {
    min-height: 500px;
    height: 65vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
