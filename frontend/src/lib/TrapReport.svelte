<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showTrapFromAddr,
    showTrapLog3D,
    showTrapTypeChart,
  } from "./chart/trap";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainTrapReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "type";

  const onOpen = async () => {
    chart = undefined;
    activeTab = "type";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("type");
  };

  let chart :any = undefined;
  const showChart = async (t: string) => {
    activeTab = t;
    await tick();
    switch (t) {
      case "type":
        chart = showTrapTypeChart(t, logs);
        break;
      case "heatmap":
        chart = showLogHeatmap(t, logs);
        break;
      case "from":
        chart = showTrapFromAddr(t, logs);
        break;
      case "trap3D":
        chart = showTrapLog3D(t, logs);
        break;
      default:
        chart = undefined;
        break;
    }
  };

  const close = () => {
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  }


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
        <div id="type"></div>
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
        <div id="heatmap"></div>
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
        <div id="from"></div>
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
        <div id="trap3D"></div>
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
  analyzeFunc={() => LLMExplainTrapReport(activeTab)}
/>

<style>
  #heatmap,
  #type,
  #from,
  #trap3D{
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>