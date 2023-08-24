<script lang="ts">
  import { Modal, Button,Tabs,TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher,tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { showEventLogStateChart,showLogHeatmap,showEventLogTimeChart,showEventLogNodeChart } from "./chart/eventlog";

  export let logs : datastore.EventLogEnt[] | undefined =undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("state");
  });

  const showChart = async (t:string) => {
    await tick();
    switch(t) {
      case "state":
        showEventLogStateChart(t,logs);
        break;
      case "heatmap":
        showLogHeatmap(t,logs);
        break;
      case "oprate":
        showEventLogTimeChart(t,"oprate",logs);
        break;
      case "arpwatch":
        showEventLogTimeChart(t,"arpwatch",logs);
        break;
      case "node":
        showEventLogNodeChart(t,logs);
        break;
    }
  }

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={()=>{showChart("state")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          状態別
        </div>
        <div id="state" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("heatmap")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          ヒートマップ
        </div>
        <div id="heatmap" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("node")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          ノード別
        </div>
        <div id="node" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("oprate")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          稼働率
        </div>
        <div id="oprate" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("arpwatch")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          ARP監視
        </div>
        <div id="arpwatch" style="height: 500px;"></div>
      </TabItem>
    </Tabs>
    <div class="flex space-x-2">
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>
