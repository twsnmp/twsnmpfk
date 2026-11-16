<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showMqttClientIDChart,
    showMqttRemoteChart,
    showMqttTopicChart,
    showMqttHeatmap,
    showMqttStateChart,
    showMqttTopicTreemap,
    showMqtt3DChart,
  } from "./chart/mqtt";
  import {
    getTableLang,
    renderBytes,
    renderCount,
    renderState,
    renderTime,
  } from "./common";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainMqttReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import DataTable from "datatables.net-dt";

  export let show: boolean = false;
  export let stats: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let chart: any = undefined;
  let activeTable: any = undefined;
  let heatmapMode: "time" | "client_topic" = "time";
  let activeTab: string = "client";

  let kpi = {
    totalCount: 0,
    totalBytes: 0,
    uniqueClients: 0,
    uniqueTopics: 0,
    alertCount: 0,
  };
  let clientRows: any[] = [];
  let remoteRows: any[] = [];
  let topicRows: any[] = [];

  const statePriority = (s: string): number => {
    switch (s) {
      case "high":
        return 3;
      case "low":
        return 2;
      case "warn":
        return 1;
      default:
        return 0;
    }
  };

  const calcStats = () => {
    const list = Array.isArray(stats) ? stats : [];
    let totalCount = 0;
    let totalBytes = 0;
    let alertCount = 0;

    const clientMap = new Map<string, any>();
    const remoteMap = new Map<string, any>();
    const topicMap = new Map<string, any>();

    for (const s of list) {
      const cnt = s.Count || 0;
      const bytes = s.Bytes || 0;
      const st = s.State || "normal";
      const last = s.Last || 0;
      const cid = s.ClientID || "-";
      const rem = s.Remote || "-";
      const top = s.Topic || "-";

      totalCount += cnt;
      totalBytes += bytes;
      if (st !== "normal") alertCount++;

      // Client aggregation
      let ce = clientMap.get(cid);
      if (!ce) {
        ce = {
          clientID: cid,
          count: 0,
          bytes: 0,
          topics: new Set<string>(),
          state: st,
          last: 0,
        };
        clientMap.set(cid, ce);
      }
      ce.count += cnt;
      ce.bytes += bytes;
      ce.topics.add(top);
      if (statePriority(st) > statePriority(ce.state)) ce.state = st;
      if (last > ce.last) ce.last = last;

      // Remote aggregation
      let re = remoteMap.get(rem);
      if (!re) {
        re = {
          remote: rem,
          count: 0,
          bytes: 0,
          clients: new Set<string>(),
          topics: new Set<string>(),
          last: 0,
        };
        remoteMap.set(rem, re);
      }
      re.count += cnt;
      re.bytes += bytes;
      re.clients.add(cid);
      re.topics.add(top);
      if (last > re.last) re.last = last;

      // Topic aggregation
      let te = topicMap.get(top);
      if (!te) {
        te = {
          topic: top,
          count: 0,
          bytes: 0,
          clients: new Set<string>(),
          state: st,
          value: s.Value || "",
          last: 0,
        };
        topicMap.set(top, te);
      }
      te.count += cnt;
      te.bytes += bytes;
      te.clients.add(cid);
      if (statePriority(st) > statePriority(te.state)) te.state = st;
      if (last >= te.last) {
        te.last = last;
        if (s.Value) te.value = s.Value;
      }
    }

    clientRows = Array.from(clientMap.values())
      .map((e) => ({ ...e, topicCount: e.topics.size }))
      .sort((a, b) => b.count - a.count);

    remoteRows = Array.from(remoteMap.values())
      .map((e) => ({
        ...e,
        clientCount: e.clients.size,
        topicCount: e.topics.size,
      }))
      .sort((a, b) => b.count - a.count);

    topicRows = Array.from(topicMap.values())
      .map((e) => ({ ...e, clientCount: e.clients.size }))
      .sort((a, b) => b.count - a.count);

    kpi = {
      totalCount,
      totalBytes,
      uniqueClients: clientMap.size,
      uniqueTopics: topicMap.size,
      alertCount,
    };
  };

  const destroyTable = () => {
    if (activeTable) {
      try {
        activeTable.destroy();
      } catch {
        // ignore
      }
      activeTable = undefined;
    }
  };

  const initTable = async (t: string) => {
    await tick();
    if (t === "client" && document.getElementById("mqttClientTable")) {
      activeTable = new DataTable("#mqttClientTable", {
        destroy: true,
        paging: clientRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: clientRows,
        order: [[1, "desc"]],
        columns: [
          { data: "clientID", title: $_("MqttReport.ClientID") || "Client ID" },
          {
            data: "count",
            title: $_("MqttReport.Messages") || "Messages",
            render: renderCount,
          },
          {
            data: "bytes",
            title: $_("MqttReport.Bytes") || "Bytes",
            render: renderBytes,
          },
          {
            data: "topicCount",
            title: $_("MqttReport.Topics") || "Topics",
            render: renderCount,
          },
          {
            data: "state",
            title: $_("MqttReport.State") || "State",
            render: renderState,
          },
          {
            data: "last",
            title: $_("MqttReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
    } else if (t === "remote" && document.getElementById("mqttRemoteTable")) {
      activeTable = new DataTable("#mqttRemoteTable", {
        destroy: true,
        paging: remoteRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: remoteRows,
        order: [[1, "desc"]],
        columns: [
          { data: "remote", title: $_("MqttReport.RemoteIP") || "Source IP" },
          {
            data: "count",
            title: $_("MqttReport.Messages") || "Messages",
            render: renderCount,
          },
          {
            data: "bytes",
            title: $_("MqttReport.Bytes") || "Bytes",
            render: renderBytes,
          },
          {
            data: "clientCount",
            title: $_("MqttReport.Clients") || "Clients",
            render: renderCount,
          },
          {
            data: "topicCount",
            title: $_("MqttReport.Topics") || "Topics",
            render: renderCount,
          },
          {
            data: "last",
            title: $_("MqttReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
    } else if (t === "topic" && document.getElementById("mqttTopicTable")) {
      activeTable = new DataTable("#mqttTopicTable", {
        destroy: true,
        paging: topicRows.length > 10,
        pageLength: 10,
        searching: true,
        info: false,
        scrollY: "300px",
        scrollCollapse: true,
        language: getTableLang(),
        data: topicRows,
        order: [[1, "desc"]],
        columns: [
          { data: "topic", title: $_("MqttReport.Topic") || "Topic" },
          {
            data: "count",
            title: $_("MqttReport.Messages") || "Messages",
            render: renderCount,
          },
          {
            data: "bytes",
            title: $_("MqttReport.Bytes") || "Bytes",
            render: renderBytes,
          },
          {
            data: "clientCount",
            title: $_("MqttReport.Clients") || "Clients",
            render: renderCount,
          },
          {
            data: "state",
            title: $_("MqttReport.State") || "State",
            render: renderState,
          },
          {
            data: "last",
            title: $_("MqttReport.LastTime") || "Last Time",
            render: renderTime,
          },
        ],
      });
    }
  };

  const onOpen = async () => {
    chart = undefined;
    destroyTable();
    calcStats();
    activeTab = "client";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("client");
  };

  const showChart = async (t: string) => {
    activeTab = t;
    destroyTable();
    await tick();
    switch (t) {
      case "client":
        chart = showMqttClientIDChart("mqttClientID", stats);
        await initTable("client");
        break;
      case "remote":
        chart = showMqttRemoteChart("mqttRemote", stats);
        await initTable("remote");
        break;
      case "topic":
        chart = showMqttTopicChart("mqttTopic", stats);
        await initTable("topic");
        break;
      case "heatmap":
        chart = showMqttHeatmap("mqttHeatmap", stats, heatmapMode);
        break;
      case "state":
        chart = showMqttStateChart("mqttState", stats);
        break;
      case "treemap":
        chart = showMqttTopicTreemap("mqttTreemap", stats);
        break;
      case "chart3D":
        chart = showMqtt3DChart("mqtt3D", stats);
        break;
      default:
        chart = undefined;
        break;
    }
    if (chart) {
      chart.resize();
    }
  };

  const toggleHeatmapMode = (mode: "time" | "client_topic") => {
    heatmapMode = mode;
    showChart("heatmap");
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
      <!-- (1) クライアントID別 -->
      <TabItem
        open
        onclick={() => {
          showChart("client");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiAccountMultiple} size={1} />
            {$_("MqttReport.CountByClientID")}
          </div>
        {/snippet}
        <div class="flex flex-col space-y-3">
          <!-- KPI Summary Cards -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiMessageTextOutline} size={0.9} class="text-cyan-500" />
                {$_("MqttReport.TotalMessages") || "総メッセージ数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.totalCount.toLocaleString()}</div>
              <div class="text-[11px] text-gray-400">
                {$_("MqttReport.TotalTraffic") || "通信量"}: {renderBytes(kpi.totalBytes, "")}
              </div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiAccountMultiple} size={0.9} class="text-blue-500" />
                {$_("MqttReport.UniqueClients") || "クライアント数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.uniqueClients.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiTagMultiple} size={0.9} class="text-emerald-500" />
                {$_("MqttReport.UniqueTopics") || "トピック数"}
              </div>
              <div class="mt-1 text-2xl font-bold">{kpi.uniqueTopics.toLocaleString()}</div>
            </div>

            <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/80 p-3 shadow-sm">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon path={icons.mdiAlertCircleOutline} size={0.9} class="text-red-500" />
                {$_("MqttReport.AlertCount") || "異常/警告状態"}
              </div>
              <div
                class="mt-1 text-2xl font-bold {kpi.alertCount > 0
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-emerald-500 dark:text-emerald-400'}"
              >
                {kpi.alertCount.toLocaleString()}
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
            <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
              <div id="mqttClientID" class="w-full h-[380px]"></div>
            </div>
            <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
              <div>
                <table id="mqttClientTable" class="display compact w-full text-xs text-left text-gray-300"></table>
              </div>
            </div>
          </div>
        </div>
      </TabItem>

      <!-- (2) 送信元別 -->
      <TabItem
        onclick={() => {
          showChart("remote");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiIpNetwork} size={1} />
            {$_("MqttReport.CountByRemote")}
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="mqttRemote" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="mqttRemoteTable" class="display compact w-full text-xs text-left text-gray-300"></table>
            </div>
          </div>
        </div>
      </TabItem>

      <!-- (3) トピック別 -->
      <TabItem
        onclick={() => {
          showChart("topic");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiTagMultiple} size={1} />
            {$_("MqttReport.CountByTopic")}
          </div>
        {/snippet}
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
          <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800">
            <div id="mqttTopic" class="w-full h-[380px]"></div>
          </div>
          <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto">
            <div>
              <table id="mqttTopicTable" class="display compact w-full text-xs text-left text-gray-300"></table>
            </div>
          </div>
        </div>
      </TabItem>

      <!-- (4) ヒートマップ -->
      <TabItem
        onclick={() => {
          showChart("heatmap");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartBox} size={1} />
            {$_("MqttReport.Heatmap")}
          </div>
        {/snippet}
        <div class="flex justify-center space-x-2 my-2">
          <GradientButton
            size="xs"
            color={heatmapMode === "time" ? "purple" : "gray"}
            onclick={() => toggleHeatmapMode("time")}
          >
            {$_("MqttReport.HeatmapTime")}
          </GradientButton>
          <GradientButton
            size="xs"
            color={heatmapMode === "client_topic" ? "purple" : "gray"}
            onclick={() => toggleHeatmapMode("client_topic")}
          >
            {$_("MqttReport.HeatmapClientTopic")}
          </GradientButton>
        </div>
        <div id="mqttHeatmap" class="w-full h-[420px]"></div>
      </TabItem>

      <!-- (5) 状態別 -->
      <TabItem
        onclick={() => {
          showChart("state");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("MqttReport.CountByState")}
          </div>
        {/snippet}
        <div id="mqttState" class="w-full h-[460px]"></div>
      </TabItem>

      <!-- (6) トピック階層ツリーマップ -->
      <TabItem
        onclick={() => {
          showChart("treemap");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiSitemap} size={1} />
            {$_("MqttReport.TopicTreemap")}
          </div>
        {/snippet}
        <div id="mqttTreemap" class="w-full h-[460px]"></div>
      </TabItem>

      <!-- (7) 3D散布図 -->
      <TabItem
        onclick={() => {
          showChart("chart3D");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartScatterPlot} size={1} />
            {$_("MqttReport.Chart3D")}
          </div>
        {/snippet}
        <div id="mqtt3D" class="w-full h-[460px]"></div>
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
        {$_("MqttReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`mqtt_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainMqttReport(stats || [], activeTab)}
/>
