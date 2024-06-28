<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Select,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {GetLocConf} from "../../wailsjs/go/main/App";
  import {
    showSFlowIFCounter,
    showSFlowCpuCounter,
    showSFlowMemCounter,
    showSFlowDiskCounter,
    showSFlowNetCounter,
  } from "./chart/sflowCounter";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  export let logs: any = undefined;

  let chart: any = undefined;
  let tab: string = "heatmap";
  const ifCounters = new Map();
  const cpuCounters = new Map();
  const memCounters = new Map();
  const diskCounters = new Map();
  const netCounters = new Map();
  let ifCounterSrc = '';
  let ifCounterSrcList = [] as any;
  let cpuCounterSrc = '';
  let cpuCounterSrcList = [] as any;
  let memCounterSrc = '';
  let memCounterSrcList = [] as any;
  let diskCounterSrc = '';
  let diskCounterSrcList = [] as any;
  let netCounterSrc = '';
  let netCounterSrcList = [] as any;

  const onOpen = () => {
    ifCounters.clear();
    cpuCounters.clear();
    memCounters.clear();
    diskCounters.clear();
    netCounters.clear();
    logs.forEach((e:any) => {
      const t = new Date(e.Time / (1000 * 1000))
      const d = JSON.parse(e.Data)
      if (!d) {
        return
      }
      d.Time = e.Time
      switch (e.Type) {
        case 'GenericInterfaceCounter':
          if (d.Index) {
            const k = e.Remote + ':' + d.Index
            if (ifCounters.has(k)) {
              ifCounters.get(k).push(d)
            } else {
              ifCounters.set(k, [d])
            }
          }
          break
        case 'HostCPUCounter': {
          const k = e.Remote
          if (cpuCounters.has(k)) {
            cpuCounters.get(k).push(d)
          } else {
            cpuCounters.set(k, [d])
          }
          break
        }
        case 'HostMemoryCounter': {
          const k = e.Remote
          if (memCounters.has(k)) {
            memCounters.get(k).push(d)
          } else {
            memCounters.set(k, [d])
          }
          break
        }
        case 'HostDiskCounter': {
          const k = e.Remote
          if (diskCounters.has(k)) {
            diskCounters.get(k).push(d)
          } else {
            diskCounters.set(k, [d])
          }
          break
        }
        case 'HostNetCounter': {
          const k = e.Remote
          if (netCounters.has(k)) {
            netCounters.get(k).push(d)
          } else {
            netCounters.set(k, [d])
          }
          break
        }
      }
    })
    ifCounterSrcList = []
    ifCounterSrc = ''
    ifCounters.forEach((_, k) => {
      ifCounterSrcList.push({
        name: k,
        value: k,
      })
      if (!ifCounterSrc) {
        ifCounterSrc = k
      }
    })
    cpuCounterSrcList = []
    cpuCounterSrc = ''
    cpuCounters.forEach((_, k) => {
      cpuCounterSrcList.push({
        name: k,
        value: k,
      })
      if (!cpuCounterSrc) {
        cpuCounterSrc = k
      }
    })
    memCounterSrcList = []
    memCounterSrc = ''
    memCounters.forEach((_, k) => {
      memCounterSrcList.push({
        name: k,
        value: k,
      })
      if (!memCounterSrc) {
        memCounterSrc = k
      }
    })
    diskCounterSrcList = []
    diskCounterSrc = ''
    diskCounters.forEach((_, k) => {
      diskCounterSrcList.push({
        name: k,
        value: k,
      })
      if (!diskCounterSrc) {
        diskCounterSrc = k
      }
    })
    netCounterSrcList = []
    netCounterSrc = ''
    netCounters.forEach((_, k) => {
      netCounterSrcList.push({
        name: k,
        value: k,
      })
      if (!netCounterSrc) {
        netCounterSrc = k
      }
    })
    showHeatmap();
  };

  const close = () => {
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  const showHeatmap = async () => {
    await tick();
    tab = "heatmap";
    chart = showLogHeatmap("heatmap", logs);
  };

  const showIFCounter = async (t:string) => {
    if (!ifCounters.has(ifCounterSrc)) {
      return;
    }
    await tick();
    tab = "ifCounter_" + t;
    chart = showSFlowIFCounter("ifCounter_" + t, ifCounters.get(ifCounterSrc), t);
  };

  const showCpuCounter = async () => {
    if(!cpuCounters.has(cpuCounterSrc)) {
      return;
    }
    await tick();
    tab = "cpuCounter";
    chart = showSFlowCpuCounter("cpuCounter", cpuCounters.get(cpuCounterSrc));
  };

  const showMemCounter = async () => {
    if(!memCounters.has(memCounterSrc)) {
      return;
    }
    await tick();
    tab = "memCounter";
    chart = showSFlowMemCounter("memCounter", memCounters.get(memCounterSrc));
  };

  const showDiskCounter = async () => {
    if(!diskCounters.has(diskCounterSrc)) {
      return;
    }
    await tick();
    tab = "diskCounter";
    chart = showSFlowDiskCounter("diskCounter", diskCounters.get(diskCounterSrc));
  };

  const showNetCounter = async () => {
    if(!netCounters.has(netCounterSrc)) {
      return;
    }
    await tick();
    tab = "netCounter";
    chart = showSFlowNetCounter("netCounter", netCounters.get(netCounterSrc));
  };


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
          showHeatmap();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          {$_("TrapReport.Heatmap")}
        </div>
        <div id="heatmap" />
      </TabItem>
      <TabItem
        on:click={() => {
          showIFCounter("bps");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiTrafficCone} size={1} />
          I/F BPS
        </div>
        <div id="ifCounter_bps" />
      </TabItem>
      <TabItem
        on:click={() => {
          showIFCounter("pps");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiTrafficCone} size={1} />
          I/F PPS
        </div>
        <div id="ifCounter_pps" />
      </TabItem>
      <TabItem
        on:click={() => {
          showCpuCounter();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiGauge} size={1} />
          CPU
        </div>
        <div id="cpuCounter" />
      </TabItem>
      <TabItem
        on:click={() => {
          showMemCounter();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiMemory} size={1} />
          Memory
        </div>
        <div id="memCounter" />
      </TabItem>
      <TabItem
        on:click={() => {
          showDiskCounter();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiHarddisk} size={1} />
          Disk
        </div>
        <div id="diskCounter" />
      </TabItem>
      <TabItem
        on:click={() => {
          showNetCounter();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiNetwork} size={1} />
          Network
        </div>
        <div id="netCounter" />
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if tab == "ifCounter_bps"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={ifCounterSrcList}
          bind:value={ifCounterSrc}
          size="sm"
          on:change={() => {
            showIFCounter("bps")
          }}
        />
      {/if}
      {#if tab == "ifCounter_pps"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={ifCounterSrcList}
          bind:value={ifCounterSrc}
          size="sm"
          on:change={() => {
            showIFCounter("pps")
          }}
        />
      {/if}
      {#if tab == "cpuCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={cpuCounterSrcList}
          bind:value={cpuCounterSrc}
          size="sm"
          on:change={showCpuCounter}
        />
      {/if}
      {#if tab == "memCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={memCounterSrcList}
          bind:value={memCounterSrc}
          size="sm"
          on:change={showMemCounter}
        />
      {/if}
      {#if tab == "diskCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={diskCounterSrcList}
          bind:value={diskCounterSrc}
          size="sm"
          on:change={showDiskCounter}
        />
      {/if}
      {#if tab == "netCounter"}
        <Select
          placeholder="Src"
          class="ml-10 w-48"
          items={netCounterSrcList}
          bind:value={netCounterSrc}
          size="sm"
          on:change={showNetCounter}
        />
      {/if}
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
  #ifCounter_bps,
  #ifCounter_pps,
  #cpuCounter,
  #memCounter,
  #diskCounter,
  #netCounter {
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
