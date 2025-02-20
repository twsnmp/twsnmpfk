<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Spinner,
    Button,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
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
  import type { datastore, main } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";
  import { CodeJar } from "@novacbn/svelte-codejar";
  import Prism from "prismjs";
  import "prismjs/components/prism-regex";
  import { copyText } from "svelte-copy";

  let data: any = [];
  let logs: any = [];
  let showReport = false;
  let table: any = undefined;
  let selectedCount = 0;
  let showPolling = false;
  let showFilter = false;
  const filter: main.TrapFilterEnt = {
    Start: "",
    End: "",
    From: "",
    Type: "",
  };
  let showLoading = false;

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#trapTable", {
      destroy: true,
      columns: columns,
      stateSave: true,
      data: data,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order:[[0,"desc"]],
      language: getTableLang(),
      select: {
        style: "multi",
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
    logs = await GetTraps(filter);
    data = [];
    for (let i = 0; i < logs.length; i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
    showLoading = false;
  };

  let chart : any = undefined;
  const showChart = async () => {
    await tick();
    chart = showLogCountChart("chart", data, zoomCallBack);
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
      width: "15%",
      render: renderTime,
    },
    {
      data: "FromAddress",
      title: $_("Trap.FromAddress"),
      width: "20%",
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
    ExportTraps("csv", filter,"");
  };

  const saveExcel = () => {
    ExportTraps("excel", filter,chart ? chart.getDataURL() : "");
  };

  let copied = false;
  const copy = () => {
    const selected = table.rows({ selected: true }).data();
    let s: string[] = [];
    const h = columns.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of columns) {
        if (c.data == "Time") {
          row.push(renderTime(selected[i][c.data] || "", ""));
        } else {
          const d = selected[i][c.data] || "" 
          row.push(d.replaceAll("\n"," "));
        }
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };

  let polling: datastore.PollingEnt;

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

  const highlight = (code: string, syntax: string | undefined) => {
    if (!syntax) {
      return "";
    }
    return Prism.highlight(code, Prism.languages[syntax], syntax);
  };
</script>

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart" />
  <div class="m-5 grow">
    <table id="trapTable" class="display compact" style="width:99%" />
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
      {#if selectedCount > 0}
        <GradientButton
          shadow
          color="cyan"
          type="button"
          on:click={copy}
          size="xs"
        >
          {#if copied}
            <Icon path={icons.mdiCheck} size={1} />
          {:else}
            <Icon path={icons.mdiContentCopy} size={1} />
          {/if}
          Copy
        </GradientButton>
      {/if}
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

<TrapReport bind:show={showReport} {logs} />

<Polling bind:show={showPolling} pollingTmp={polling} />

<Modal bind:open={showFilter} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Trap.Filter")}
    </h3>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Start")}</span>
        <Input class="h-8" type="datetime-local" bind:value={filter.Start} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.End")}</span>
        <Input class="h-8" type="datetime-local" bind:value={filter.End} size="sm" />
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
    <Label class="space-y-2 text-xs">
      <span>{$_("Trap.FromAddress")} </span>
      <CodeJar style="padding: 6px;" syntax="regex" {highlight} bind:value={filter.From} />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>{$_("Trap.TrapType")}</span>
      <CodeJar style="padding: 6px;" syntax="regex" {highlight} bind:value={filter.Type} />
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
