<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { main } from "wailsjs/go/models";
  import { showArpLogIP, showArpLogIP3D } from "./chart/arp";
  import { _ } from 'svelte-i18n';

  export let logs: main.ArpLogEnt[] | undefined = undefined;

  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("ip");
  });

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
    dispatch("close", {});
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
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem
        open
        on:click={() => {
          showChart("ip");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          { $_('ArpReport.CountByIP') }
        </div>
        <div id="ip"/>
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("ip3D");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          { $_('ArpReport.Chart3DByIP') }
        </div>
        <div id="ip3D"/>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton type="button" color="teal" on:click={close} size="xs">
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