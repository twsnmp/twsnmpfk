<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Spinner,
    Button,
    Checkbox,
    Dropdown,
    DropdownItem,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetNetFlow,
    ExportNetFlow,
    DeleteAllNetFlow,
  } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, renderTimeMili } from "./common";
  import { showLogCountChart, resizeLogCountChart } from "./chart/logcount";
  import NetFlowReport from "./NetFlowReport.svelte";
  import AddressInfo from "./AddressInfo.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { main } from "wailsjs/go/models";
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
  let showFilter = false;
  const filter: main.NetFlowFilterEnt = {
    Start: "",
    End: "",
    Single: true,
    SrcAddr: "",
    SrcPort: 0,
    SrcLoc: "",
    SrcMAC: "",
    DstAddr: "",
    DstPort: 0,
    DstLoc: "",
    DstMAC: "",
    TCPFlags: "",
    Protocol: "",
  };

  let showLoading = false;
  let addrInfoOpen = false;
  let showAddrInfo = false;
  let address = "";
  let addrList :any = [];

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#netFlowTable", {
      destroy: true,
      columns: columns,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      stateSave: true,
      data: data,
      order: [[0, "desc"]],
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
      updateAddrList();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
      updateAddrList();
    });
  };

  const updateAddrList = () => {
    const selected = table.rows({ selected: true }).data();
    addrList = [];
    const m = new Map();
    for (let i = 0; i < selected.length; i++) {
      for(const k of ["SrcAddr","SrcMAC","DstAddr","DstMAC"]) {
        const a = selected[i][k]
        if (a && !m.has(a)) {
          m.set(a,true)
          addrList.push(a);
        }
      }
      if (addrList.length > 10) {
        break;
      }
    }
  }

  const showAddrInfoFunc = (a:string) => {
    address = a;
    showAddrInfo = true;
  }

  const refresh = async () => {
    showLoading = true;
    filter.SrcPort *= 1;
    filter.DstPort *= 1;
    logs = await GetNetFlow(filter);
    data = [];
    for (let i = 0; i < logs.length; i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
    showLoading = false;
  };

  let chart: any = undefined;
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
      width: "13%",
      render: renderTimeMili,
    },
    {
      data: "SrcAddr",
      title: $_("NetFlow.SrcAddr"),
      width: "8%",
    },
    {
      data: "SrcPort",
      title: $_("NetFlow.Port"),
      width: "4%",
    },
    {
      data: "SrcLoc",
      title: $_("NetFlow.Loc"),
      width: "8%",
    },
    {
      data: "SrcMAC",
      title: "MAC",
      width: "6%",
    },
    {
      data: "DstAddr",
      title: $_("NetFlow.DstAddr"),
      width: "8%",
    },
    {
      data: "DstPort",
      title: $_("NetFlow.Port"),
      width: "4%",
    },
    {
      data: "DstLoc",
      title: $_("NetFlow.Loc"),
      width: "8%",
    },
    {
      data: "DstMAC",
      title: "MAC",
      width: "6%",
    },
    {
      data: "Protocol",
      title: $_("NetFlow.Protocol"),
      width: "7%",
    },
    {
      data: "TCPFlags",
      title: $_("NetFlow.TCPFlags"),
      width: "8%",
    },
    {
      data: "Packets",
      title: $_("NetFlow.Packets"),
      width: "5%",
    },
    {
      data: "Bytes",
      title: $_("NetFlow.Bytes"),
      width: "5%",
    },
    {
      data: "Dur",
      title: $_("NetFlow.Dur"),
      width: "5%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const saveCSV = () => {
    ExportNetFlow("csv", filter, "");
  };

  const saveExcel = () => {
    ExportNetFlow("excel", filter, chart ? chart.getDataURL() : "");
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
          const d = (selected[i][c.data] || "") + "";
          row.push(d.replaceAll("\n", " "));
        }
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };

  const deleteAll = async () => {
    if (await DeleteAllNetFlow()) {
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
    <table id="netFlowTable" class="display compact" style="width:99%" />
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
        <GradientButton>
          {$_('Address.AddressInfo')}
          <Icon path={icons.mdiChevronDown} size={1} />
        </GradientButton>
        <Dropdown bind:open={addrInfoOpen}>
          {#each addrList as a}
            <DropdownItem on:click={() => showAddrInfoFunc(a)}>{a}</DropdownItem>
          {/each}
        </Dropdown>
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

<NetFlowReport bind:show={showReport} {logs} />

<Modal bind:open={showFilter} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Trap.Filter")}
    </h3>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Start")}</span>
        <Input
          class="h-8"
          type="datetime-local"
          bind:value={filter.Start}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.End")}</span>
        <Input
          class="h-8"
          type="datetime-local"
          bind:value={filter.End}
          size="sm"
        />
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
    <Checkbox bind:checked={filter.Single}>{$_("NetFlow.Single")}</Checkbox>
    {#if filter.Single}
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>IP</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcAddr}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Port")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="0"
            max="65554"
            bind:value={filter.SrcPort}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Loc")}</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcLoc}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>MAC</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcMAC}
          />
        </Label>
      </div>
    {:else}
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.SrcAddr")}</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcAddr}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Port")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="0"
            max="65554"
            bind:value={filter.SrcPort}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Loc")}</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcLoc}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>MAC</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.SrcMAC}
          />
        </Label>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.DstAddr")}</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.DstAddr}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Port")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="0"
            max="65554"
            bind:value={filter.DstPort}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("NetFlow.Loc")}</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.DstLoc}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>MAC</span>
          <CodeJar
            style="padding: 6px;"
            syntax="regex"
            {highlight}
            bind:value={filter.DstMAC}
          />
        </Label>
      </div>
    {/if}
    <div class="grid gap-2 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>{$_("NetFlow.Protocol")}</span>
        <CodeJar
          style="padding: 6px;"
          syntax="regex"
          {highlight}
          bind:value={filter.Protocol}
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("NetFlow.TCPFlags")}</span>
        <CodeJar
          style="padding: 6px;"
          syntax="regex"
          {highlight}
          bind:value={filter.TCPFlags}
        />
      </Label>
    </div>
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

<AddressInfo
  bind:show = {showAddrInfo}
  {address}
/>

<style>
  #chart {
    min-height: 200px;
    height: 20vh;
    width: 95vw;
    margin: 0 auto;
  }
</style>
