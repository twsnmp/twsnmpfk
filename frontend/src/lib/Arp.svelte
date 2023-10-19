<script lang="ts">
  import { GradientButton } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import { ExportArpLogs, GetArpLogs } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, renderState } from "./common";
  import { showLogCountChart, resizeLogCountChart } from "./chart/logcount";
  import ArpReport from "./ArpReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let arpLogs = [];
  let arpLogData = [];
  let showReport = false;
  let arpLogTable = undefined;

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

  const renderArpLogIP = (ip: string, type: string) => {
    if (type == "sort") {
      return ip
        .split(".")
        .reduce((int, v) => Number(int) * 256 + Number(v) + "");
    }
    if (ip.startsWith("169.254.")) {
      return `<span class="text-red-500">${ip}</span>`;
    }
    return ip;
  };


  const arpLogColumns = [
    {
      data: "State",
      title: $_("Arp.State"),
      width: "5%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("Arp.DateTime"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "IP",
      title: $_("Arp.IPAddress"),
      width: "10%",
      render: renderArpLogIP,
    },
    {
      data: "Node",
      title: "ノード",
      width: "15%",
    },
    {
      data: "NewMAC",
      title: $_("Arp.NewMACAddress"),
      width: "10%",
    },
    {
      data: "NewVendor",
      title: "新ベンダー",
      width: "15%",
    },
    {
      data: "OldMAC",
      title: $_("Arp.OldMACAddress"),
      width: "10%",
    },
    {
      data: "OldVendor",
      title: "旧ベンダー",
      width: "15%",
    },
  ];

  const refresh = async () => {
    arpLogs = await GetArpLogs();
    arpLogData = [];
    for (let i = 0; i < arpLogs.length; i++) {
      arpLogData.push(arpLogs[i]);
    }
    arpLogs.reverse();
    showArpLogTable();
    showChart();
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

  const saveCSV = () => {
    ExportArpLogs("csv");
  };

  const saveExcel = () => {
    ExportArpLogs("excel");
  };
</script>

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;" />
  <table id="arpLogTable" class="display compact" style="width:99%" />
  <div class="flex justify-end space-x-2 mr-2 mt-2">
    <GradientButton
      type="button"
      color="green"
      on:click={() => {
        showReport = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiChartPie} size={1} />
      {$_("Arp.Report")}
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
      {$_("Arp.Reload")}
    </GradientButton>
  </div>
</div>

{#if showReport}
  <ArpReport
    logs={arpLogs}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
