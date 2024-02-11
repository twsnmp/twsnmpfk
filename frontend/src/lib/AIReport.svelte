<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {showAIHeatMap,showAIPieChart,showAITimeChart } from "./chart/ai";
  import { GetAIResult } from "../../wailsjs/go/main/App";
  import { _ } from 'svelte-i18n';

  export let id = "";
  export let show: boolean = false;

  let results : any =undefined;

  const onOpen = async () => {
    results = await GetAIResult(id);
    showChart("heatmap");
  };

  let chart :any = undefined;
  const showChart = async (t:string) => {
    await tick();
    chart = undefined;
    switch(t) {
      case "heatmap":
        chart = showAIHeatMap(t,results.ScoreData);
        break;
      case "pie":
        chart = showAIPieChart(t,results.ScoreData);
        break;
      case "time":
        chart = showAITimeChart(t,results.ScoreData);
        break;
    }
  }

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
      <TabItem open on:click={()=>{showChart("heatmap")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          { $_('AIReport.Heatmap') }
        </div>
        <div id="heatmap"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("pie")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          { $_('AIReport.PieChart') }
        </div>
        <div id="pie"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("time")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          { $_('AIReport.TimeChart') }
        </div>
        <div id="time"></div>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_('AIReport.Close')}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #heatmap,
  #pie,
  #time {
    min-height: 500px;
    height: 70vh;
    width: 98%;
  }
</style>