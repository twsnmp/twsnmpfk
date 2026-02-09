<script lang="ts">
  import { 
    Modal, 
    GradientButton, 
    Tabs, 
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Progressbar,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { showArpGraph,showIPAMHeatmap } from "./chart/arp";
  import { _ } from 'svelte-i18n';

  export let show: boolean = false;
  export let arp : any = undefined;
  export let changeMAC:any = undefined;
  export let changeIP:any = undefined;
  export let ipam:any = undefined;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    chart = undefined;
    showIPAM();
  };

  let chart :any  = undefined;
  const showChart = async (t: string) => {
    await tick();
    chart = undefined;
    switch (t) {
      case "graphForce":
        chart= showArpGraph(t, arp,"force",changeIP,changeMAC);
        break;
      case "graphCircular":
        chart = showArpGraph(t, arp,"circular",changeIP,changeMAC);
        break;
    }
  };

  const showIPAM = async () => {
    await tick();
    chart = undefined;
    chart = showIPAMHeatmap("ipam",ipam);
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

  const getUsageColor = (u:number) => {
    if (u < 60.0) {
      return "blue";
    } else if (u < 90.0) {
      return "yellow";
    }
    return "red";
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
      <TabItem
        open
        on:click={() => {
          showIPAM();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          {$_('ArpReport.IPAM')}
        </div>
        <div id="ipam" />
        <Table striped={true}>
          <TableHead>
            <TableHeadCell>{$_('ArpReport.IPRange')}</TableHeadCell>
            <TableHeadCell>{$_('ArpREport.Size')}</TableHeadCell>
            <TableHeadCell>{$_('ArpReport.Used')}</TableHeadCell>
            <TableHeadCell>{$_('ArpReport.Usage')}</TableHeadCell>
          </TableHead>
          <TableBody tableBodyClass="divide-y">
            {#each ipam as i }
              <TableBodyRow>
                <TableBodyCell>{i.Range}</TableBodyCell>
                <TableBodyCell>{i.Size}</TableBodyCell>
                <TableBodyCell>{i.Used}</TableBodyCell>
                <TableBodyCell>
                  <Progressbar progress={i.Usage.toFixed(2)} size="h-5"  color={getUsageColor(i.Usage)} labelInside />
                </TableBodyCell>
              </TableBodyRow>
            {/each}
          </TableBody>
        </Table>
      </TabItem>
      <TabItem
        on:click={() => {
          showChart("graphForce");
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiGraph} size={1} />
          { $_('ArpReport.IPtoMACForceGraph') }
        </div>
        <div id="graphForce" />
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
        <div id="graphCircular"/>
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
  #graphForce,
  #graphCircular {
    min-height:  500px;
    width:  98%;
    height: 70vh;
  }
  #ipam {
    min-width: 300px;
    width: 98%;
    height: 35vh;
  }
</style>