<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { showArpLogIP, showArpLogIP3D } from "./chart/arp";
  import { getTableLang, renderCount, renderTime } from "./common";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainArpReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import DataTable from "datatables.net-dt";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "ip";
  let chart: any = undefined;
  let ipTable: any = undefined;

  let kpi = {
    total: 0,
    newCount: 0,
    changeCount: 0,
    uniqueIPs: 0,
    uniqueMACs: 0,
  };
  let ipRows: any[] = [];

  const calcStats = () => {
    const list = Array.isArray(logs) ? logs : [];
    const ipMap = new Map<
      string,
      {
        ip: string;
        total: number;
        newCount: number;
        changeCount: number;
        lastMAC: string;
        oldMAC: string;
        lastTime: number;
      }
    >();
    const macSet = new Set<string>();
    let newCount = 0;
    let changeCount = 0;

    for (const l of list) {
      const ip = l.IP || "-";
      const tm = l.Time || 0;
      const isNew = l.State === "New";
      const isChange = l.State === "Change";
      if (isNew) newCount++;
      if (isChange) changeCount++;
      if (l.NewMAC) macSet.add(l.NewMAC);
      if (l.OldMAC) macSet.add(l.OldMAC);

      let e = ipMap.get(ip);
      if (!e) {
        e = {
          ip,
          total: 0,
          newCount: 0,
          changeCount: 0,
          lastMAC: l.NewMAC || "-",
          oldMAC: l.OldMAC || "-",
          lastTime: 0,
        };
        ipMap.set(ip, e);
      }
      e.total++;
      if (isNew) e.newCount++;
      if (isChange) e.changeCount++;
      if (tm >= e.lastTime) {
        e.lastTime = tm;
        if (l.NewMAC) e.lastMAC = l.NewMAC;
        if (l.OldMAC) e.oldMAC = l.OldMAC;
      }
    }

    ipRows = Array.from(ipMap.values()).sort(
      (a, b) => b.changeCount - a.changeCount || b.total - a.total
    );

    kpi = {
      total: list.length,
      newCount,
      changeCount,
      uniqueIPs: ipMap.size,
      uniqueMACs: macSet.size,
    };
  };

  const destroyTable = () => {
    if (ipTable) {
      try {
        ipTable.destroy();
      } catch {
        // ignore
      }
      ipTable = undefined;
    }
  };

  const initIpTable = async () => {
    await tick();
    if (!document.getElementById("arpIpTable")) return;
    ipTable = new DataTable("#arpIpTable", {
      destroy: true,
      paging: ipRows.length > 10,
      pageLength: 10,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: ipRows,
      order: [[3, "desc"], [1, "desc"]],
      columns: [
        { data: "ip", title: $_("Arp.IPAddress") || "IP Address" },
        {
          data: "total",
          title: $_("ArpReport.Total") || "Total",
          render: renderCount,
        },
        {
          data: "newCount",
          title: $_("Ts.New") || "New",
          render: renderCount,
        },
        {
          data: "changeCount",
          title: $_("Ts.Change") || "Change",
          render: (v: number, type: string) => {
            if (type === "sort") return v;
            if (v > 0) {
              return `<span class="text-red-400 font-bold">${v.toLocaleString()}</span>`;
            }
            return "0";
          },
        },
        { data: "lastMAC", title: $_("Arp.NewMACAddress") || "New MAC" },
        { data: "oldMAC", title: $_("Arp.OldMACAddress") || "Old MAC" },
        {
          data: "lastTime",
          title: $_("Arp.DateTime") || "Last Time",
          render: renderTime,
        },
      ],
    });
  };

  const onOpen = async () => {
    chart = undefined;
    destroyTable();
    calcStats();
    activeTab = "ip";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("ip");
  };

  const showChart = async (t: string) => {
    activeTab = t;
    destroyTable();
    await tick();
    chart = undefined;
    switch (t) {
      case "ip":
        chart = showArpLogIP(t, logs);
        await initIpTable();
        break;
      case "ip3D":
        chart = showArpLogIP3D(t, logs);
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
          showChart("ip");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBarStacked} size={1} />
            {$_("ArpReport.CountByIP")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3">
          <!-- KPI Summary Cards -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiLan} size={0.9} class="text-cyan-500" />
                {$_("ArpReport.TotalEvents") || "総ARPイベント数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.total.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiPlusCircleOutline} size={0.9} class="text-blue-500" />
                {$_("ArpReport.NewCount") || "新規検出 (New)"}
              </div>
              <div class="mt-1 text-2xl font-bold text-blue-500 dark:text-blue-400">
                {kpi.newCount.toLocaleString()}
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiSwapHorizontalBold} size={0.9} class="text-red-500" />
                {$_("ArpReport.ChangeCount") || "MAC変更 (Change)"}
              </div>
              <div
                class="mt-1 text-2xl font-bold {kpi.changeCount > 0
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-emerald-500 dark:text-emerald-400'}"
              >
                {kpi.changeCount.toLocaleString()}
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiIpNetworkOutline} size={0.9} class="text-emerald-500" />
                {$_("ArpReport.UniqueIPs") || "ユニークIP / MAC"}
              </div>
              <div class="mt-1 text-2xl font-bold">
                {kpi.uniqueIPs.toLocaleString()}
                <span class="text-xs font-normal text-gray-400">
                  / {kpi.uniqueMACs.toLocaleString()} MAC
                </span>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
            <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
              <div id="ip" class="w-full h-[380px]"></div>
            </div>
            <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
              <div>
                <table id="arpIpTable" class="display compact w-full text-xs text-left text-gray-300"></table>
              </div>
            </div>
          </div>
        </div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("ip3D");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartScatterPlot} size={1} />
            {$_("ArpReport.Chart3DByIP")}
          </div>
        {/snippet}
        <div id="ip3D" class="w-full h-[460px]"></div>
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
      <GradientButton type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("ArpReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`arp_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainArpReport(logs || [], activeTab)}
/>
