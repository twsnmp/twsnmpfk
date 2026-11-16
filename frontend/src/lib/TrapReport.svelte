<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showTrapFromAddr,
    showTrapLog3D,
    showTrapTypeChart,
  } from "./chart/trap";
  import { showLogHeatmap } from "./chart/eventlog";
  import { getTableLang, renderCount, renderTime } from "./common";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainTrapReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import DataTable from "datatables.net-dt";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "type";
  let chart: any = undefined;
  let typeTable: any = undefined;
  let fromTable: any = undefined;

  let kpi = {
    total: 0,
    uniqueSources: 0,
    uniqueTypes: 0,
    topType: "-",
    topTypeCount: 0,
    topSource: "-",
    topSourceCount: 0,
  };
  let typeRows: any[] = [];
  let fromRows: any[] = [];

  const calcStats = () => {
    const list = Array.isArray(logs) ? logs : [];
    const typeMap = new Map<
      string,
      { type: string; count: number; sources: Set<string>; lastTime: number }
    >();
    const fromMap = new Map<
      string,
      {
        from: string;
        count: number;
        types: Map<string, number>;
        lastTime: number;
      }
    >();

    for (const l of list) {
      const t = l.TrapType || "-";
      const f = l.FromAddress || "-";
      const tm = l.Time || 0;

      let te = typeMap.get(t);
      if (!te) {
        te = { type: t, count: 0, sources: new Set(), lastTime: 0 };
        typeMap.set(t, te);
      }
      te.count++;
      te.sources.add(f);
      if (tm > te.lastTime) te.lastTime = tm;

      let fe = fromMap.get(f);
      if (!fe) {
        fe = { from: f, count: 0, types: new Map(), lastTime: 0 };
        fromMap.set(f, fe);
      }
      fe.count++;
      fe.types.set(t, (fe.types.get(t) || 0) + 1);
      if (tm > fe.lastTime) fe.lastTime = tm;
    }

    const total = list.length;
    typeRows = Array.from(typeMap.values())
      .map((e) => ({
        type: e.type,
        count: e.count,
        ratio: total > 0 ? ((e.count / total) * 100).toFixed(1) + "%" : "0.0%",
        sources: e.sources.size,
        lastTime: e.lastTime,
      }))
      .sort((a, b) => b.count - a.count);

    fromRows = Array.from(fromMap.values())
      .map((e) => {
        let topT = "-";
        let topTC = 0;
        e.types.forEach((c, k) => {
          if (c > topTC) {
            topTC = c;
            topT = k;
          }
        });
        return {
          from: e.from,
          count: e.count,
          typeCount: e.types.size,
          topType: topT,
          lastTime: e.lastTime,
        };
      })
      .sort((a, b) => b.count - a.count);

    kpi = {
      total,
      uniqueSources: fromMap.size,
      uniqueTypes: typeMap.size,
      topType: typeRows[0]?.type || "-",
      topTypeCount: typeRows[0]?.count || 0,
      topSource: fromRows[0]?.from || "-",
      topSourceCount: fromRows[0]?.count || 0,
    };
  };

  const destroyTables = () => {
    if (typeTable) {
      try {
        typeTable.destroy();
      } catch {
        // ignore
      }
      typeTable = undefined;
    }
    if (fromTable) {
      try {
        fromTable.destroy();
      } catch {
        // ignore
      }
      fromTable = undefined;
    }
  };

  const initTypeTable = async () => {
    await tick();
    if (!document.getElementById("trapTypeTable")) return;
    typeTable = new DataTable("#trapTypeTable", {
      destroy: true,
      paging: typeRows.length > 10,
      pageLength: 10,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: typeRows,
      order: [[1, "desc"]],
      columns: [
        { data: "type", title: $_("TrapReport.TrapType") || "Trap Type" },
        {
          data: "count",
          title: $_("TrapReport.Count") || "Count",
          render: renderCount,
        },
        { data: "ratio", title: $_("TrapReport.Ratio") || "Ratio" },
        {
          data: "sources",
          title: $_("TrapReport.UniqueSources") || "Sources",
          render: renderCount,
        },
        {
          data: "lastTime",
          title: $_("TrapReport.LastTime") || "Last Time",
          render: renderTime,
        },
      ],
    });
  };

  const initFromTable = async () => {
    await tick();
    if (!document.getElementById("trapFromTable")) return;
    fromTable = new DataTable("#trapFromTable", {
      destroy: true,
      paging: fromRows.length > 10,
      pageLength: 10,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: fromRows,
      order: [[1, "desc"]],
      columns: [
        { data: "from", title: $_("TrapReport.FromAddress") || "From Address" },
        {
          data: "count",
          title: $_("TrapReport.Count") || "Count",
          render: renderCount,
        },
        {
          data: "typeCount",
          title: $_("TrapReport.UniqueTypes") || "Types",
          render: renderCount,
        },
        { data: "topType", title: $_("TrapReport.TopTrapType") || "Top Type" },
        {
          data: "lastTime",
          title: $_("TrapReport.LastTime") || "Last Time",
          render: renderTime,
        },
      ],
    });
  };

  const onOpen = async () => {
    chart = undefined;
    destroyTables();
    calcStats();
    activeTab = "type";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("type");
  };

  const showChart = async (t: string) => {
    activeTab = t;
    destroyTables();
    await tick();
    switch (t) {
      case "type":
        chart = showTrapTypeChart(t, logs);
        await initTypeTable();
        break;
      case "heatmap":
        chart = showLogHeatmap(t, logs);
        break;
      case "from":
        chart = showTrapFromAddr(t, logs);
        await initFromTable();
        break;
      case "trap3D":
        chart = showTrapLog3D(t, logs);
        break;
      default:
        chart = undefined;
        break;
    }
    if (chart) {
      chart.resize();
    }
  };

  const close = () => {
    destroyTables();
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
          showChart("type");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("TrapReport.CountByType")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3">
          <!-- KPI Summary Cards -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiAlertOctagonOutline} size={0.9} class="text-cyan-500" />
                {$_("TrapReport.TotalTraps") || "総Trap数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.total.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiServerNetwork} size={0.9} class="text-blue-500" />
                {$_("TrapReport.UniqueSources") || "送信元数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.uniqueSources.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiTagMultiple} size={0.9} class="text-emerald-500" />
                {$_("TrapReport.UniqueTypes") || "Trap種別数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.uniqueTypes.toLocaleString()}</div>
              <div class="text-[11px] text-gray-400 truncate" title={kpi.topType}>
                Top: {kpi.topType} ({kpi.topTypeCount.toLocaleString()})
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiCrownOutline} size={0.9} class="text-amber-500" />
                {$_("TrapReport.TopSource") || "最多送信元"}
              </div>
              <div class="mt-1 text-base font-bold truncate" title={kpi.topSource}>
                {kpi.topSource}
              </div>
              <div class="text-[11px] text-gray-400">
                {kpi.topSourceCount.toLocaleString()} {$_("TrapReport.EventsUnit") || "件"}
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
            <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
              <div id="type" class="w-full h-[380px]"></div>
            </div>
            <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
              <div>
                <table id="trapTypeTable" class="display compact w-full text-xs text-left text-gray-300"></table>
              </div>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("heatmap");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBox} size={1} />
            {$_("TrapReport.Heatmap")}
          </div>
        {/snippet}
        <div id="heatmap" class="w-full h-[460px]"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("from");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBarStacked} size={1} />
            {$_("TrapReport.CountByFromAddress")}
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="from" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="trapFromTable" class="display compact w-full text-xs text-left text-gray-300"></table>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("trap3D");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartScatterPlot} size={1} />
            {$_("TrapReport.Chart3D")}
          </div>
        {/snippet}
        <div id="trap3D" class="w-full h-[460px]"></div>
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
        {$_("TrapReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`trap_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainTrapReport(logs || [], activeTab)}
/>
