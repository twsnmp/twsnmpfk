<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher,tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showAIHeatMap,showAIPieChart,showAITimeChart } from "./chart/ai";
  import { GetAIResult } from "../../wailsjs/go/main/App";
  import { _ } from 'svelte-i18n';

  export let id = "";
  export let results : datastore.AIResult | undefined =undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    results = await GetAIResult(id);
    show = true;
    showChart("heatmap");
  });

  const showChart = async (t:string) => {
    await tick();
    switch(t) {
      case "heatmap":
        showAIHeatMap(t,results.ScoreData);
        break;
      case "pie":
        showAIPieChart(t,results.ScoreData);
        break;
      case "time":
        showAITimeChart(t,results.ScoreData);
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
  class="w-full min-h-[90vh]"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={()=>{showChart("heatmap")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          { $_('AIReport.Heatmap') }
        </div>
        <div id="heatmap" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("pie")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          { $_('AIReport.PieChart') }
        </div>
        <div id="pie" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("time")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          { $_('AIReport.TimeChart') }
        </div>
        <div id="time" style="height: 500px;"></div>
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
