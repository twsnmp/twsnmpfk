<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {
    showTrapFromAddr,
    showTrapLog3D,
    showTrapTypeChart,
  } from "./chart/trap";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";

  export let logs: datastore.TrapEnt[] | undefined = undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("type");
  });

  const showChart = async (t: string) => {
    await tick();
    switch (t) {
      case "type":
        showTrapTypeChart(t, logs);
        break;
      case "heatmap":
        showLogHeatmap(t, logs);
        break;
      case "from":
        showTrapFromAddr(t, logs);
        break;
      case "trap3D":
        showTrapLog3D(t, logs);
        break;
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full min-h-[90vh]"
  on:on:close={close}
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
        <div id="type" style="height: 500px;" />
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
        <div id="heatmap" style="height: 500px;" />
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
        <div id="from" style="height: 500px;" />
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
        <div id="trap3D" style="height: 500px;" />
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
