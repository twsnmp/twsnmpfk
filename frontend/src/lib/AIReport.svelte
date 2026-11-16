<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { showAIHeatMap, showAIPieChart, showAITimeChart } from "./chart/ai";
  import { getTableLang, renderTimeUnix } from "./common";
  import { GetAIResult } from "../../wailsjs/go/main/App";
  import { _ } from "svelte-i18n";
  import DataTable from "datatables.net-dt";

  export let id = "";
  export let show: boolean = false;

  let results: any = undefined;
  let chart: any = undefined;
  let anomalyTable: any = undefined;

  let kpi = {
    total: 0,
    anomalyCount: 0,
    warnCount: 0,
    anomalyRate: "0.0",
    warnRate: "0.0",
    maxScore: 0,
    avgScore: 0,
    lastTime: 0,
  };
  let alertRows: any[] = [];

  const calcStats = () => {
    const scores: any[] = results?.ScoreData || [];
    let anomalyCount = 0;
    let warnCount = 0;
    let maxScore = 0;
    let sumScore = 0;
    let lastTime = 0;
    const alerts: any[] = [];

    for (const e of scores) {
      const ts = Number(e[0]) || 0;
      const score = Number(e[1]) || 0;
      sumScore += score;
      if (score > maxScore) maxScore = score;
      if (ts > lastTime) lastTime = ts;

      if (score > 66.0) {
        anomalyCount++;
        alerts.push({ time: ts, score, level: "Anomaly" });
      } else if (score > 50.0) {
        warnCount++;
        alerts.push({ time: ts, score, level: "Warn" });
      }
    }

    const total = scores.length;
    alertRows = alerts.sort((a, b) => b.score - a.score);
    kpi = {
      total,
      anomalyCount,
      warnCount,
      anomalyRate: total > 0 ? ((anomalyCount / total) * 100).toFixed(1) : "0.0",
      warnRate: total > 0 ? ((warnCount / total) * 100).toFixed(1) : "0.0",
      maxScore,
      avgScore: total > 0 ? sumScore / total : 0,
      lastTime,
    };
  };

  const destroyTable = () => {
    if (anomalyTable) {
      try {
        anomalyTable.destroy();
      } catch {
        // ignore
      }
      anomalyTable = undefined;
    }
  };

  const initAnomalyTable = async () => {
    await tick();
    if (!document.getElementById("aiAnomalyTable")) return;
    const anomalyLabel = $_("Ts.Anamary") || "Anomaly";
    const warnLabel = $_("Ts.Warn") || "Warn";
    anomalyTable = new DataTable("#aiAnomalyTable", {
      destroy: true,
      paging: alertRows.length > 10,
      pageLength: 10,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: alertRows,
      order: [[1, "desc"]],
      columns: [
        {
          data: "time",
          title: $_("Ts.DateTime") || "Date/Time",
          render: renderTimeUnix,
        },
        {
          data: "score",
          title: $_("Ts.AnmaryScore") || "Score",
          render: (v: number, type: string) => {
            if (type === "sort") return v;
            return v.toFixed(2);
          },
        },
        {
          data: "level",
          title: $_("AIReport.Level") || "Level",
          render: (lvl: string) => {
            if (lvl === "Anomaly") {
              return `<span class="px-2 py-0.5 rounded text-xs font-semibold bg-red-900 text-red-200">${anomalyLabel}</span>`;
            }
            return `<span class="px-2 py-0.5 rounded text-xs font-semibold bg-yellow-900 text-yellow-200">${warnLabel}</span>`;
          },
        },
      ],
    });
  };

  const onOpen = async () => {
    results = undefined;
    chart = undefined;
    destroyTable();
    results = await GetAIResult(id);
    calcStats();
    showChart("heatmap");
  };

  const showChart = async (t: string) => {
    destroyTable();
    await tick();
    chart = undefined;
    const scoreData = results?.ScoreData || [];
    switch (t) {
      case "heatmap":
        chart = showAIHeatMap(t, scoreData);
        break;
      case "pie":
        chart = showAIPieChart(t, scoreData);
        break;
      case "time":
        chart = showAITimeChart(t, scoreData);
        await initAnomalyTable();
        break;
    }
    if (chart) {
      chart.resize();
    }
  };

  const close = () => {
    destroyTable();
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
  outsideclose={false}
  class="w-full min-h-[90vh]"
>
  <div class="flex flex-col space-y-3">
    <Tabs style="underline">
      <TabItem
        open
        onclick={() => {
          showChart("heatmap");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBox} size={1} />
            {$_("AIReport.Heatmap")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3">
          <!-- KPI Summary Cards -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiDatabaseOutline} size={0.9} class="text-cyan-500" />
                {$_("AIReport.TotalSamples") || "分析サンプル数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.total.toLocaleString()}</div>
              {#if kpi.lastTime > 0}
                <div class="text-[11px] text-gray-400 truncate">
                  {$_("AIReport.Latest") || "最新"}: {renderTimeUnix(kpi.lastTime, "")}
                </div>
              {/if}
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiAlertCircleOutline} size={0.9} class="text-red-500" />
                {$_("AIReport.AnomalyCount") || "異常検出 (>66)"}
              </div>
              <div
                class="mt-1 text-2xl font-bold {kpi.anomalyCount > 0
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-emerald-500 dark:text-emerald-400'}"
              >
                {kpi.anomalyCount.toLocaleString()}
                <span class="text-xs font-normal text-gray-400">({kpi.anomalyRate}%)</span>
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiAlertOutline} size={0.9} class="text-amber-500" />
                {$_("AIReport.WarnCount") || "警告検出 (>50)"}
              </div>
              <div class="mt-1 text-2xl font-bold text-amber-500 dark:text-amber-400">
                {kpi.warnCount.toLocaleString()}
                <span class="text-xs font-normal text-gray-400">({kpi.warnRate}%)</span>
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiSpeedometer} size={0.9} class="text-purple-500" />
                {$_("AIReport.MaxAvgScore") || "最大 / 平均スコア"}
              </div>
              <div class="mt-1 text-2xl font-bold">
                {kpi.maxScore.toFixed(1)}
                <span class="text-xs font-normal text-gray-400">
                  / {kpi.avgScore.toFixed(1)}
                </span>
              </div>
            </div>
          </div>

          <div id="heatmap" class="w-full h-[400px]"></div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("pie");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("AIReport.PieChart")}
          </div>
        {/snippet}
        <div id="pie" class="w-full h-[460px]"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("time");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartLine} size={1} />
            {$_("AIReport.TimeChart")}
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="time" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div class="text-xs font-semibold text-gray-300 mb-1.5">
              {$_("AIReport.AlertEventsTitle") || "警告・異常イベント一覧 (Score > 50)"}
            </div>
            <div>
              <table id="aiAnomalyTable" class="display compact w-full text-xs text-left text-gray-300"></table>
            </div>
          </div>
        </div>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("AIReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>
