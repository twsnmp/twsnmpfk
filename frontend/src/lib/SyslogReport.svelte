<script lang="ts">
  import { Modal, Button,Tabs,TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher,tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showSyslogLevelChart,showSyslogHost,showSyslogHost3D,showSyslogFFT3D  } from "./chart/syslog";
  import { showLogHeatmap } from "./chart/eventlog";

  export let logs : datastore.SyslogEnt[] | undefined =undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("level");
  });

  const showChart = async (t:string) => {
    await tick();
    switch(t) {
      case "level":
        showSyslogLevelChart(t,logs);
        break;
      case "heatmap":
        showLogHeatmap(t,logs);
        break;
      case "host":
        showSyslogHost(t,logs);
        break;
      case "host3D":
        showSyslogHost3D(t,logs);
        break;
      case "fft":
        showSyslogFFT3D(t,logs);
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
      <TabItem open on:click={()=>{showChart("level")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          状態別
        </div>
        <div id="level" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("heatmap")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          ヒートマップ
        </div>
        <div id="heatmap" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("host")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          ホスト別
        </div>
        <div id="host" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("host3D")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          ホスト別(3D)
        </div>
        <div id="host3D" style="height: 500px;"></div>
      </TabItem>
      <TabItem on:click={()=>{showChart("fft")}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          FFTによる周期分析
        </div>
        <div id="fft" style="height: 500px;"></div>
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