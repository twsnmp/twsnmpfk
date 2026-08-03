<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { main } from "wailsjs/go/models";
  import { showArpLogIP, showArpLogIP3D } from "./chart/arp";
  import { _ } from 'svelte-i18n';
  import { GetMapConf, LLMExplainArpReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let logs: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "ip";

  const onOpen = async () => {
    activeTab = "ip";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("ip");
  };

  let chart :any = undefined;
  const showChart = async (t: string) => {
    activeTab = t;
    await tick();
    chart = undefined;
    switch (t) {
      case "ip":
        chart = showArpLogIP(t, logs);
        break;
      case "ip3D":
        chart= showArpLogIP3D(t, logs);
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
          showChart("ip");
        }}
      >
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          { $_('ArpReport.CountByIP') }
        </div>
      {/snippet}
        <div id="ip"></div>
      </TabItem>
      <TabItem
        onclick={() => {
          showChart("ip3D");
        }}
      >
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          { $_('ArpReport.Chart3DByIP') }
        </div>
      {/snippet}
        <div id="ip3D"></div>
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
        { $_('ArpReport.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`arp_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainArpReport(activeTab)}
/>

<style>
 #ip,
 #ip3D {
  min-height: 500px;
  height: 70vh;
  width: 98%;
 } 
</style>