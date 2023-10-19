<script lang="ts">
  import { Modal, GradientButton, Tabs, TabItem } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore} from "wailsjs/go/models";
  import { showArpGraph } from "./chart/arp";
  import { _ } from 'svelte-i18n';

  export let arp : datastore.ArpEnt[] | undefined = undefined;
  export let changeMAC = undefined;
  export let changeIP = undefined;

  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    show = true;
    showChart("graphForce");
  });

  const showChart = async (t: string) => {
    await tick();
    switch (t) {
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
          showChart("graphForce");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiGraph} size={1} />
          { $_('ArpReport.IPtoMACForceGraph') }
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
          { $_('ArpReport.IPtoMACCircelGraph') }
        </div>
        <div id="graphCircular" style="height: 600px;" />
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
