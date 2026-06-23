<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { main } from "wailsjs/go/models";
  import { showArpLogIP, showArpLogIP3D } from "./chart/arp";
  import { _ } from 'svelte-i18n';

  export let show: boolean = false;
  export let logs: any = undefined;

  const onOpen = async () => {
    showChart("ip");
  };

  let chart :any = undefined;
  const showChart = async (t: string) => {
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
      <GradientButton type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('ArpReport.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
 #ip,
 #ip3D {
  min-height: 500px;
  height: 70vh;
  width: 98%;
 } 
</style>