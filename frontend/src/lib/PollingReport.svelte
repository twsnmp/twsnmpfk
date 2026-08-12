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
    showPollingQQPlot,
  } from "./chart/polling";
  import {
    getStateIcon,
    getStateColor,
    getStateName,
    getTableLang,
    renderTime,
    renderState,
    renderDuration,
  } from "./common";
  import {
    calcSinglePollingDowntimeAndSLA,
    showSinglePollingDowntimeChart,
  } from "./chart/eventlog";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { showAIHeatMap } from "./chart/ai";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainPollingReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let id = "";

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "state";

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
  let downtimeTable: any = undefined;
  let downtimeStats = {
    totalIncidents: 0,
    ongoingIncidents: 0,
    totalDowntimeSec: 0,
    maxDowntimeSec: 0,
    mttrSec: 0,
    overallSLA: 100,
    incidents: [] as any[],
  };

  let chart: any = undefined;
  let selectedTab = "";

  const clearDowntimeTable = () => {
    if (downtimeTable) {
      downtimeTable.destroy();
      downtimeTable = undefined;
    }
  };

  const close = () => {
    clearDowntimeTable();
    show = false;
  };

  const onOpen = async () => {
    clearDowntimeTable();
    polling = undefined;
    node = undefined;
    logs = undefined;
    dispLogs = [];
    aiResult = undefined;
    resultData = [];
    entList = [];
    selectedTab = "";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
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
    resultTable = new DataTable("#resultTable", {
      destroy: true,
      stateSave: true,
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
    pollingLogTable = new DataTable("#pollingLogTable", {
      destroy: true,
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

  const showQQPlot = async () => {
    selectedTab = "qqplot";
    await tick();
    chart = showPollingQQPlot("qqplot", logs, selectedEnt);
  };

  const showAI = async () => {
    selectedTab = "ai";
    await tick();
    chart = showAIHeatMap("ai", aiResult.ScoreData);
  };

  const showDowntime = async () => {
    selectedTab = "downtime";
    clearDowntimeTable();
    downtimeStats = calcSinglePollingDowntimeAndSLA(logs);
    await tick();
    chart = showSinglePollingDowntimeChart("pollingDowntimeChart", logs);
    await initDowntimeTable();
  };

  const initDowntimeTable = async () => {
    if (!document.getElementById("pollingDowntimeTable")) {
      return;
    }
    const data = downtimeStats.incidents.map((inc: any, idx: number) => {
      const stStr = renderTime(inc.startTime, "");
      const etStr = inc.ongoing
        ? $_("EventLogReport.Ongoing") || "障害発生中"
        : renderTime(inc.endTime, "");
      const durStr = renderDuration(inc.durationSec, "");
      return {
        no: `#${downtimeStats.incidents.length - idx}`,
        level: inc.level,
        start: stStr,
        end: etStr,
        duration: durStr,
      };
    });
    clearDowntimeTable();
    downtimeTable = new DataTable("#pollingDowntimeTable", {
      destroy: true,
      data: data,
      stateSave: true,
      order: [[0, "desc"]],
      language: getTableLang(),
      columns: [
        { data: "no", title: "#", width: "10%" },
        {
          data: "level",
          title: $_("PollingReport.State") || "State",
          width: "15%",
          render: renderState,
        },
        {
          data: "start",
          title: $_("EventLogReport.StartTime") || "開始日時",
          width: "25%",
        },
        {
          data: "end",
          title: $_("EventLogReport.EndTime") || "復旧日時",
          width: "25%",
        },
        {
          data: "duration",
          title: $_("EventLogReport.TotalDowntime") || "障害時間",
          width: "25%",
        },
      ],
    });
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
    const rkeys = [];
    if (logs.length > 1) {
      for (const k of  Object.keys(logs[0].Result)) {
        if (k != "error") {
          rkeys.push(k);
          ed.Header.push(k);
        }
      }
    }
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
            for (const k of rkeys) {
              row.push(l.Result[k] || "");
            }
            break;
        }
      }
      ed.Data.push(row);
    }
    ExportAny(t, ed);
  };
  const exportLogData = () => {
    if (!logs || logs.length < 1) {
      return
    }
    const ed: any = {
      Title: "TWSNMP_Polling_Log_Data",
      Header: [],
      Data: [],
      Image: "",
    };
    ed.Header.push("time");
    ed.Header.push("state");
    for (const k of  Object.keys(logs[0].Result)) {
        if (k != "error" && typeof logs[0].Result[k] === "number" ) {
          ed.Header.push(k);
        }
      }
    for (const l of logs) {
      const row: any = [];
      for (const k of ed.Header ) {
        switch (k) {
          case "time":
            row.push(renderTime(l.Time, ""));
            break;
            case "state":
            row.push(l.State || "");
            break;
          default:
            row.push(l.Result[k] || "");
            break;
        }
      }
      ed.Data.push(row);
    }
    ExportAny("csv", ed);
  };

  $: if (show) {
    onOpen();
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
>
  {#if !node}
    <div class="text-center mt-10"><Spinner size="16" /></div>
  {:else}
    <div class="flex flex-col space-y-4">
      <Tabs style="underline">
        <TabItem
          open
          onclick={() => {
            chart = undefined;
            selectedTab = "";
            showResultTable();
          }}
        >
          {#snippet titleSlot()}
        <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("PollingReport.BasicInfo")}
          </div>
      {/snippet}
          <div class="grid gap-2 grid-cols-2">
            <Table striped={true}>
              <TableHead>
                <TableHeadCell>{$_("PollingReport.Item")}</TableHeadCell>
                <TableHeadCell>{$_("PollingReport.Content")}</TableHeadCell>
              </TableHead>
              <TableBody class="divide-y">
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
                      style="color:{getStateColor(polling.State)};"></span>
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
            <div><table id="resultTable" class="display compact" style="width:99%"></table></div>
          </div>
        </TabItem>
        {#if polling.LogMode > 0}
          <TabItem onclick={showDowntime}>
            {#snippet titleSlot()}
              <div class="flex items-center gap-2">
                <Icon path={icons.mdiClockAlertOutline} size={1} />
                {$_("PollingReport.Downtime")}
              </div>
            {/snippet}
            <div class="flex flex-col space-y-3 w-full p-1">
              <!-- Compact Summary Cards -->
              <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-1">
                <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                  <div class="text-xs text-gray-400 font-semibold mb-0.5">
                    {$_("EventLogReport.OverallSLA")}
                  </div>
                  <div
                    class="text-lg font-bold {downtimeStats.overallSLA >= 99.9
                      ? 'text-green-400'
                      : downtimeStats.overallSLA >= 99.0
                      ? 'text-yellow-400'
                      : 'text-red-400'}"
                  >
                    {downtimeStats.overallSLA.toFixed(3)}%
                  </div>
                </div>
                <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                  <div class="text-xs text-gray-400 font-semibold mb-0.5">
                    {$_("EventLogReport.TotalIncidents")}
                  </div>
                  <div class="text-lg font-bold text-blue-400">
                    {downtimeStats.totalIncidents}
                    <span class="text-xs font-normal text-gray-400">
                      ({$_("EventLogReport.OngoingIncidents")}: {downtimeStats.ongoingIncidents})
                    </span>
                  </div>
                </div>
                <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                  <div class="text-xs text-gray-400 font-semibold mb-0.5">
                    {$_("EventLogReport.MTTR")}
                  </div>
                  <div class="text-lg font-bold text-teal-400">
                    {renderDuration(downtimeStats.mttrSec)}
                  </div>
                </div>
                <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                  <div class="text-xs text-gray-400 font-semibold mb-0.5">
                    {$_("EventLogReport.MaxDowntime")} / {$_("EventLogReport.TotalDowntime")}
                  </div>
                  <div class="text-xs font-bold text-red-400 mt-1">
                    {renderDuration(downtimeStats.maxDowntimeSec)}
                    <span class="text-xs font-normal text-gray-400">
                      / {renderDuration(downtimeStats.totalDowntimeSec)}
                    </span>
                  </div>
                </div>
              </div>

              <!-- 2 Columns Grid: Graph & Table -->
              <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
                <!-- Left Column: ECharts Graph (5 cols) -->
                <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 shadow-inner">
                  <div id="pollingDowntimeChart" class="w-full h-[360px]"></div>
                </div>

                <!-- Right Column: Detailed Table (7 cols) wrapped in div (AGENTS.md Rule) -->
                <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto shadow-inner">
                  <div>
                    <table
                      id="pollingDowntimeTable"
                      class="display compact w-full text-xs text-left text-gray-300"
                    ></table>
                  </div>
                </div>
              </div>
            </div>
          </TabItem>
          <TabItem onclick={showLog}>
            {#snippet titleSlot()}
        <div class="flex items-center gap-2">
              <Icon path={icons.mdiLanCheck} size={1} />
              {$_("PollingReport.PollingLog")}
            </div>
      {/snippet}
            <div id="log"></div>
            <div><table
              id="pollingLogTable"
              class="display compact"
              style="width:99%;"></table></div>
          </TabItem>
          <TabItem onclick={showTimeChart}>
            {#snippet titleSlot()}
        <div class="flex items-center gap-2">
              <Icon path={icons.mdiCalendarCheck} size={1} />
              {$_("PollingReport.TimeChart")}
            </div>
      {/snippet}
            <div id="time"></div>
          </TabItem>
          <TabItem onclick={showHistogram}>
            {#snippet titleSlot()}
        <div class="flex items-center gap-2">
              <Icon path={icons.mdiAppsBox} size={1} />
              {$_("PollingReport.Histogram")}
            </div>
      {/snippet}
            <div id="histogram"></div>
          </TabItem>
          <TabItem onclick={showQQPlot}>
            {#snippet titleSlot()}
        <div class="flex items-center gap-2">
              <Icon path={icons.mdiAppsBox} size={1} />
              {$_('PollingReport.QQPlot')}
            </div>
      {/snippet}
            <div id="qqplot"></div>
          </TabItem>
          {#if polling.LogMode == 3 && aiResult}
            <TabItem onclick={showAI}>
              {#snippet titleSlot()}
        <div class="flex items-center gap-2">
                <Icon path={icons.mdiAppsBox} size={1} />
                {$_("PollingReport.AI")}
              </div>
      {/snippet}
              <div id="ai"></div>
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
            onchange={showTimeChart}
            placeholder={$_("PollingReport.SelectVal")}
          />
        {/if}
        {#if selectedTab == "histogram"}
          <Select
            class="w-64"
            size="sm"
            items={entList}
            bind:value={selectedEnt}
            onchange={showHistogram}
            placeholder={$_("PollingReport.SelectVal")}
          />
        {/if}
        {#if selectedTab == "qqplot"}
          <Select
            class="w-64"
            size="sm"
            items={entList}
            bind:value={selectedEnt}
            onchange={showQQPlot}
            placeholder={$_("PollingReport.SelectVal")}
          />
        {/if}
        {#if selectedTab == "log" && logs.length > 0}
          <GradientButton
            shadow
            color="lime"
            type="button"
            onclick={() => exportLogs("csv")}
            size="xs"
          >
            <Icon path={icons.mdiFileDelimited} size={1} />
            CSV
          </GradientButton>
          <GradientButton
            shadow
            color="lime"
            type="button"
            onclick={() => exportLogData()}
            size="xs"
          >
            <Icon path={icons.mdiFileDelimited} size={1} />
            CSV(Data)
          </GradientButton>
          <GradientButton
            shadow
            color="lime"
            type="button"
            onclick={() => exportLogs("excel")}
            size="xs"
          >
            <Icon path={icons.mdiFileExcel} size={1} />
            Excel
          </GradientButton>
        {/if}
        {#if hasAI}
          <GradientButton
            shadow
            type="button"
            color="pink"
            onclick={() => (showAIReport = true)}
            size="xs"
          >
            <Icon path={icons.mdiBrain} size={1} />
            {$_("ReportAI.AIExplain")}
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("PollingReport.Close")}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`polling_${id}_ai_explanation`}
  analyzeFunc={() => LLMExplainPollingReport(id)}
/>

<style>
  #log {
    min-height: 200px;
    height: 25vh;
    width: 98%;
    margin: 0 auto;
  }
  #time,
  #histogram,
  #qqplot,
  #ai {
    min-height: 500px;
    height: 70vh;
    widows: 98%;
    margin: 0 auto;
  }
</style>
