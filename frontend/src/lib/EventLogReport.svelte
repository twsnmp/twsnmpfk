<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showEventLogStateChart,
    showLogHeatmap,
    showEventLogTimeChart,
    showEventLogNodeChart,
    showEventLogDowntimeChart,
    calcEventLogDowntimeAndSLA,
  } from "./chart/eventlog";
  import { renderState, renderDuration, renderSLA, getTableLang } from "./common";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainEventLogReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import DataTable from "datatables.net-dt";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "downtime";

  let downtimeTable: any = undefined;
  let downtimeStats = {
    totalIncidents: 0,
    ongoingIncidents: 0,
    totalDowntimeSec: 0,
    maxDowntimeSec: 0,
    mttrSec: 0,
    overallSLA: 100,
    nodeStats: [] as any[],
    incidents: [] as any[],
  };

  const onOpen = async () => {
    chart = undefined;
    if (downtimeTable) {
      try {
        downtimeTable.destroy();
      } catch (e) {
        // ignore
      }
      downtimeTable = undefined;
    }
    activeTab = "downtime";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("downtime");
  };

  let chart: any = undefined;

  const showChart = async (t: string) => {
    activeTab = t;
    await tick();
    if (downtimeTable) {
      try {
        downtimeTable.destroy();
      } catch (e) {
        // ignore
      }
      downtimeTable = undefined;
    }

    switch (t) {
      case "downtime":
        downtimeStats = calcEventLogDowntimeAndSLA(logs);
        chart = showEventLogDowntimeChart("downtimeChart", logs);
        await initDowntimeTable();
        break;
      case "state":
        chart = showEventLogStateChart(t, logs);
        break;
      case "heatmap":
        chart = showLogHeatmap(t, logs);
        break;
      case "node":
        chart = showEventLogNodeChart(t, logs);
        break;
      case "oprate":
        chart = showEventLogTimeChart(t, "oprate", logs);
        break;
      case "arpwatch":
        chart = showEventLogTimeChart(t, "arpwatch", logs);
        break;
      default:
        chart = undefined;
        break;
    }
  };

  const initDowntimeTable = async () => {
    await tick();
    if (!document.getElementById("downtimeTable")) {
      return;
    }

    const sortedNodeStats = [...downtimeStats.nodeStats].sort(
      (a, b) => b.totalDowntimeSec - a.totalDowntimeSec
    );

    const ongoingLabel = $_("EventLogReport.Ongoing") || "障害発生中";

    const tableData = sortedNodeStats.map((n: any) => {
      const slaStr = renderSLA(n.sla, "");
      const totalDurationStr = renderDuration(n.totalDowntimeSec, "");
      const maxDurationStr = renderDuration(n.maxDowntimeSec, "");
      const statusHtml = n.ongoing
        ? `<span class="px-2 py-0.5 rounded text-xs font-semibold bg-red-900 text-red-200">${ongoingLabel}</span>`
        : renderState(n.currentLevel);

      return [
        n.nodeName,
        slaStr,
        n.count,
        totalDurationStr,
        maxDurationStr,
        statusHtml,
      ];
    });

    downtimeTable = new DataTable("#downtimeTable", {
      destroy: true,
      paging: false,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: tableData,
      order: [[3, "desc"]],
      columns: [
        { title: $_("EventLog.NodeName") || "Node name" },
        { title: $_("EventLogReport.OverallSLA") || "SLA (Availability)" },
        { title: $_("EventLogReport.TotalIncidents") || "Incidents" },
        { title: $_("EventLogReport.TotalDowntime") || "Total Downtime" },
        { title: $_("EventLogReport.MaxDowntime") || "Max Downtime" },
        { title: $_("EventLogReport.Status") || "Status" },
      ],
    });
  };

  const close = () => {
    if (downtimeTable) {
      try {
        downtimeTable.destroy();
      } catch (e) {
        // ignore
      }
      downtimeTable = undefined;
    }
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
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
  class="w-full max-h-[90vh] overflow-y-auto"
>
  <div class="flex flex-col space-y-2">
    <Tabs style="underline" contentClass="p-1 pt-2">
      <!-- 1st Tab: Downtime & SLA -->
      <TabItem
        open
        onclick={() => {
          showChart("downtime");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiTimerAlertOutline} size={1} />
            {$_("EventLogReport.Downtime")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3 w-full p-1">
          <!-- Summary Cards (Compact) -->
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

          <!-- 2 Columns Grid: Graph (Left 5 cols ~41.6%) & Table (Right 7 cols ~58.3%) -->
          <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
            <!-- Left Column: ECharts Graph (5 cols) -->
            <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 shadow-inner">
              <div id="downtimeChart" class="w-full h-[360px]"></div>
            </div>

            <!-- Right Column: Detailed Table (7 cols) wrapped in div (AGENTS.md Rule) -->
            <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto shadow-inner">
              <div>
                <table
                  id="downtimeTable"
                  class="display compact w-full text-xs text-left text-gray-300"
                ></table>
              </div>
            </div>
          </div>
        </div>
      </TabItem>

      <TabItem
        onclick={() => {
          showChart("state");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("EventLogReport.CountByState")}
          </div>
        {/snippet}
        <div id="state"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("heatmap");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBox} size={1} />
            {$_("EventLogReport.Heatmap")}
          </div>
        {/snippet}
        <div id="heatmap"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("node");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBarStacked} size={1} />
            {$_("EventLogReport.CountByNode")}
          </div>
        {/snippet}
        <div id="node"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("oprate");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartLine} size={1} />
            {$_("EventLogREport.Oprate")}
          </div>
        {/snippet}
        <div id="oprate"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("arpwatch");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartLine} size={1} />
            {$_("EventLogReport.ArpWatch")}
          </div>
        {/snippet}
        <div id="arpwatch"></div>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
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
        {$_("EventLogReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`eventlog_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainEventLogReport(logs || [], activeTab)}
/>

<style>
  #heatmap,
  #node,
  #state,
  #oprate,
  #arpwatch {
    width: 98%;
    height: 70vh;
    min-height: 500px;
    margin: 0 auto;
  }
</style>
