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

  export let show: boolean = false;
  export let logs: any = undefined;

  const onOpen = async () => {
    showChart("type");
  };

  let chart :any = undefined;
  const showChart = async (t: string) => {
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

</script>

<svelte:window on:resize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem
        open
        on:click={() => {
          showChart("type");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          {$_("TrapReport.CountByType")}
        </div>
        <div id="type"/>
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("heatmap");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          {$_("TrapReport.Heatmap")}
        </div>
        <div id="heatmap"/>
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("from");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          {$_("TrapReport.CountByFromAddress")}
        </div>
        <div id="from"/>
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("trap3D");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          {$_("TrapReport.Chart3D")}
        </div>
        <div id="trap3D"/>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("TrapReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

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