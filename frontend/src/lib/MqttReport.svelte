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
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainMqttReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let stats: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let chart: any = undefined;
  let heatmapMode: "time" | "client_topic" = "time";
  let activeTab: string = "client";

  const onOpen = async () => {
    chart = undefined;
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
    await tick();
    switch (t) {
      case "client":
        chart = showMqttClientIDChart("mqttClientID", stats);
        break;
      case "remote":
        chart = showMqttRemoteChart("mqttRemote", stats);
        break;
      case "topic":
        chart = showMqttTopicChart("mqttTopic", stats);
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
  };

  const toggleHeatmapMode = (mode: "time" | "client_topic") => {
    heatmapMode = mode;
    showChart("heatmap");
  };

  const close = () => {
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
  class="w-full min-h-[90vh]"
>
  <div class="flex flex-col space-y-4">
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
        <div id="mqttClientID"></div>
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
        <div id="mqttRemote"></div>
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
        <div id="mqttTopic"></div>
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
        <div id="mqttHeatmap"></div>
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
        <div id="mqttState"></div>
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
        <div id="mqttTreemap"></div>
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
        <div id="mqtt3D"></div>
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
  analyzeFunc={() => LLMExplainMqttReport(activeTab)}
/>

<style>
  #mqttClientID,
  #mqttRemote,
  #mqttTopic,
  #mqttHeatmap,
  #mqttState,
  #mqttTreemap,
  #mqtt3D {
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
