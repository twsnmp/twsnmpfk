<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    showEventLogStateChart,
    showLogHeatmap,
    showEventLogTimeChart,
    showEventLogNodeChart,
  } from "./chart/eventlog";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  export let logs: any = undefined;

  const onOpen = async () => {
    chart = undefined;
    showChart("state");
  };

  let chart :any = undefined;

  const showChart = async (t: string) => {
    await tick();
    switch (t) {
      case "state":
        chart = showEventLogStateChart(t, logs);
        break;
      case "heatmap":
        chart = showLogHeatmap(t, logs);
        break;
      case "oprate":
        chart = showEventLogTimeChart(t, "oprate", logs);
        break;
      case "arpwatch":
        chart = showEventLogTimeChart(t, "arpwatch", logs);
        break;
      case "node":
        chart = showEventLogNodeChart(t, logs);
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
      <TabItem
        open
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
