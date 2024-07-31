<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Select,
    Spinner,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNode,
    GetPolling,
    GetPollingLogs,
    GetAIResult,
    ExportAny,
  } from "../../wailsjs/go/main/App";
  import { showLogStateChart } from "./chart/logstate";
  import {
    showPollingChart,
    showPollingHistogram,
    getChartParams,
  } from "./chart/polling";
  import {
    getStateIcon,
    getStateColor,
    getStateName,
    getTableLang,
    renderTime,
    renderState,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { showAIHeatMap } from "./chart/ai";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  export let id = "";

  let polling: any = undefined;
  let node: any = undefined;
  let logs: any = undefined;
  let dispLogs: any = [];
  let aiResult: any = undefined;
  let entList: any = [];
  let selectedEnt = "";
  let pollingLogTable: any = undefined;
  let resultTable: any = undefined;
  let resultData: any = [];

  let chart: any = undefined;
  let selectedTab = "";

  const close = () => {
    show = false;
  };

  const onOpen = async () => {
    polling = await GetPolling(id);
    node = await GetNode(polling.NodeID);
    if (polling.LogMode > 0) {
      loadLogs();
    } else {
      logs = undefined;
      dispLogs = [];
      aiResult = undefined;
    }
    resultData = [];
    entList.length = 0;
    selectedTab = "";
    if (polling && polling.Result) {
      for (const k of Object.keys(polling.Result)) {
        selectedEnt = k;
        const dp = getChartParams(k);
        entList.push({
          name: dp.axis,
          value: k,
        });
        resultData.push({
          name: k,
          value: polling.Result[k],
        });
      }
    }
    showResultTable();
  };

  const showResultTable = async () => {
    await tick();
    if (resultTable && DataTable.isDataTable("#resultTable")) {
      resultTable.clear();
      resultTable.destroy();
      resultTable = undefined;
    }
    resultTable = new DataTable("#resultTable", {
      columns: [
        {
          data: "name",
          title: $_("PollingReport.Item"),
        },
        {
          data: "value",
          title: $_("PollingReport.Content"),
        },
      ],
      paging: false,
      searching: false,
      info: false,
      scrollY: "60vh",
      data: resultData,
      language: getTableLang(),
    });
  };

  const loadLogs = async () => {
    dispLogs = [];
    logs = await GetPollingLogs(id);
    for (let i = 0; i < logs.length; i++) {
      dispLogs.push(logs[i]);
    }
    logs.reverse();
    aiResult = await GetAIResult(id);
  };

  const zoomCallBack = (st: number, et: number) => {
    dispLogs = [];
    for (let i = logs.length - 1; i >= 0; i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        dispLogs.push(logs[i]);
      }
    }
    showLogTable();
  };

  const showLog = async () => {
    selectedTab = "log";
    await tick();
    showLogTable();
    chart = showLogStateChart("log", logs, zoomCallBack);
  };

  const renderResult = (r: any) => {
    let l = [];
    for (const k of Object.keys(r)) {
      l.push(k + "=" + r[k]);
    }
    return l.join(" ");
  };

  const logsColumns = [
    {
      data: "State",
      title: $_("PollingReport.State"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("PollingReport.Time"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Result",
      title: $_("PollingReport.Result"),
      width: "75%",
      render: renderResult,
    },
  ];

  const showLogTable = () => {
    if (pollingLogTable && DataTable.isDataTable("#pollingLogTable")) {
      pollingLogTable.clear();
      pollingLogTable.destroy();
      pollingLogTable = undefined;
    }
    pollingLogTable = new DataTable("#pollingLogTable", {
      data: dispLogs,
      stateSave: true,
      order: [[0, "desc"]],
      language: getTableLang(),
      columns: logsColumns,
    });
  };

  const showTimeChart = async () => {
    selectedTab = "time";
    await tick();
    chart = showPollingChart("time", logs, selectedEnt);
  };

  const showHistogram = async () => {
    selectedTab = "histogram";
    await tick();
    chart = showPollingHistogram("histogram", logs, selectedEnt);
  };

  const showAI = async () => {
    selectedTab = "ai";
    await tick();
    chart = showAIHeatMap("ai", aiResult.ScoreData);
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };
  const exportLogs = (t: string) => {
    const ed: any = {
      Title: "TWSNMP_Polling_Log",
      Header: logsColumns.map((e: any) => e.title),
      Data: [],
      Image: t == "excel" && chart ? chart.getDataURL() : "",
    };
    for (const l of logs) {
      const row: any = [];
      for (const c of logsColumns) {
        switch (c.data) {
          case "Time":
            row.push(renderTime(l.Time, ""));
            break;
          case "State":
            row.push(l.State || "");
            break;
          case "Result":
            row.push(renderResult(l.Result));
            break;
        }
      }
      ed.Data.push(row);
    }
    ExportAny(t, ed);
  };
</script>

<svelte:window on:resize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  {#if !node}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <div class="flex flex-col space-y-4">
      <Tabs style="underline">
        <TabItem
          open
          on:click={() => {
            chart = undefined;
            selectedTab = "";
            showResultTable();
          }}
        >
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("PollingReport.BasicInfo")}
          </div>
          <div class="grid gap-2 grid-cols-2">
            <Table striped={true}>
              <TableHead>
                <TableHeadCell>{$_("PollingReport.Item")}</TableHeadCell>
                <TableHeadCell>{$_("PollingReport.Content")}</TableHeadCell>
              </TableHead>
              <TableBody tableBodyClass="divide-y">
                <TableBodyRow>
                  <TableBodyCell>{$_("PollingReport.NodeName")}</TableBodyCell>
                  <TableBodyCell>{node.Name}</TableBodyCell>
                </TableBodyRow>
                <TableBodyRow>
                  <TableBodyCell>{$_("PollingReport.Name")}</TableBodyCell>
                  <TableBodyCell>{polling.Name}</TableBodyCell>
                </TableBodyRow>
                <TableBodyRow>
                  <TableBodyCell>{$_("PollingReport.State")}</TableBodyCell>
                  <TableBodyCell>
                    <span
                      class="mdi {getStateIcon(polling.State)} text-xl"
                      style="color:{getStateColor(polling.State)};"
                    />
                    <span class="ml-2 text-xs text-black dark:text-white"
                      >{getStateName(polling.State)}</span
                    >
                  </TableBodyCell>
                </TableBodyRow>
                <TableBodyRow>
                  <TableBodyCell>{$_("PollingReport.LastTime")}</TableBodyCell>
                  <TableBodyCell
                    >{renderTime(polling.LastTime, "")}</TableBodyCell
                  >
                </TableBodyRow>
              </TableBody>
            </Table>
            <table id="resultTable" class="display compact" style="width:99%" />
          </div>
        </TabItem>
        {#if polling.LogMode > 0}
          <TabItem on:click={showLog}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiLanCheck} size={1} />
              {$_("PollingReport.PollingLog")}
            </div>
            <div id="log" />
            <table
              id="pollingLogTable"
              class="display compact"
              style="width:99%;"
            />
          </TabItem>
          <TabItem on:click={showTimeChart}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiCalendarCheck} size={1} />
              {$_("PollingReport.TimeChart")}
            </div>
            <div id="time" />
          </TabItem>
          <TabItem on:click={showHistogram}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiAppsBox} size={1} />
              {$_("PollingReport.Histogram")}
            </div>
            <div id="histogram" />
          </TabItem>
          {#if polling.LogMode == 3 && aiResult}
            <TabItem on:click={showAI}>
              <div slot="title" class="flex items-center gap-2">
                <Icon path={icons.mdiAppsBox} size={1} />
                {$_("PollingReport.AI")}
              </div>
              <div id="ai" />
            </TabItem>
          {/if}
        {/if}
      </Tabs>
      <div class="flex justify-end space-x-2 mr-2">
        {#if selectedTab == "time"}
          <Select
            class="w-64"
            size="sm"
            items={entList}
            bind:value={selectedEnt}
            on:change={showTimeChart}
            placeholder={$_("PollingReport.SelectVal")}
          />
        {/if}
        {#if selectedTab == "histogram"}
          <Select
            class="w-64"
            size="sm"
            items={entList}
            bind:value={selectedEnt}
            on:change={showHistogram}
            placeholder={$_("PollingReport.SelectVal")}
          />
        {/if}
        {#if selectedTab == "log" && logs.length > 0 }
          <GradientButton
            shadow
            color="lime"
            type="button"
            on:click={() => exportLogs("csv")}
            size="xs"
          >
            <Icon path={icons.mdiFileDelimited} size={1} />
            CSV
          </GradientButton>
          <GradientButton
            shadow
            color="lime"
            type="button"
            on:click={() => exportLogs("excel")}
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
          on:click={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("PollingReport.Close")}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<style>
  #log {
    min-height: 200px;
    height: 30vh;
    width: 98%;
    margin: 0 auto;
  }
  #time,
  #histogram,
  #ai {
    min-height: 500px;
    height: 70vh;
    widows: 98%;
    margin: 0 auto;
  }
</style>
