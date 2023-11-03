<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher,tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showSyslogLevelChart,showSyslogHost,showSyslogHost3D,showSyslogFFT3D  } from "./chart/syslog";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";

  export let logs : datastore.SyslogEnt[] | undefined =undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("level");
  });

  let chart = undefined;
  const showChart = async (t:string) => {
    await tick();
    switch(t) {
      case "level":
        chart = showSyslogLevelChart(t,logs);
        break;
      case "heatmap":
        chart = showLogHeatmap(t,logs);
        break;
      case "host":
        chart = showSyslogHost(t,logs);
        break;
      case "host3D":
        chart = showSyslogHost3D(t,logs);
        break;
      case "fft":
        chart = showSyslogFFT3D(t,logs);
        break;
      default:
        chart = undefined;
        break;
    }
  }

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
  permanent
  class="w-full min-h-[90vh]"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={()=>{showChart("level")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          {$_('SyslogReport.CountByLevel')}
        </div>
        <div id="level"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("heatmap")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          {$_('SyslogReport.Heatmap')}
        </div>
        <div id="heatmap"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("host")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          {$_('SyslogReport.CountByHost')}
        </div>
        <div id="host"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("host3D")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          {$_('SyslogReport.Chart3D')}
        </div>
        <div id="host3D"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("fft")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          {$_('SyslogReport.FFT')}
        </div>
        <div id="fft"></div>
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_('SyslogReport.Close')}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #level,
  #heatmap,
  #host,
  #host3D,
  #fft{
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>