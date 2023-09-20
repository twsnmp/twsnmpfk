<script lang="ts">
  import { Modal, Button, Tabs, TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { showArpLogIP, showArpLogIP3D,showArpGraph } from "./chart/arp";

  export let logs: datastore.ArpLogEnt[] | undefined = undefined;
  export let arp : datastore.ArpEnt[] | undefined = undefined;
  export let changeMAC = undefined;
  export let changeIP = undefined;

  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("ip");
  });

  const showChart = async (t: string) => {
    await tick();
    switch (t) {
      case "ip":
        showArpLogIP(t, logs);
        break;
      case "ip3D":
        showArpLogIP3D(t, logs);
        break;
      case "graphForce":
        showArpGraph(t, arp,"force",changeIP,changeMAC);
        break;
      case "graphCircular":
        showArpGraph(t, arp,"circular",changeIP,changeMAC);
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
          showChart("ip");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          IPアドレス別
        </div>
        <div id="ip" style="height: 600px;" />
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("ip3D");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          IPアドレス別(3D)
        </div>
        <div id="ip3D" style="height: 600px;" />
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("graphForce");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiGraph} size={1} />
          IPとMACの関係(力学モデル)
        </div>
        <div id="graphForce" style="height: 600px;" />
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("graphCircular");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiCircle} size={1} />
          IPとMACの関係(円形)
        </div>
        <div id="graphCircular" style="height: 600px;" />
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>
