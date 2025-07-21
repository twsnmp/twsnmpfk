<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showSyslogLevelChart,showSyslogHost,showSyslogHost3D,showSyslogFFT3D, getSyslogSummary,showSyslogSummary  } from "./chart/syslog";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";
  import {
    getTableLang,
    renderCount,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  export let show: boolean = false;
  export let logs : datastore.SyslogEnt[] | undefined =undefined;

  const onOpen = async () => {
    showChart("level");
  };

  let chart :any  = undefined;
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

  const showSummary = async () => {
    const list = getSyslogSummary(logs);    
    await tick();
    const table = new DataTable("#syslogSummaryTable", {
      destroy: true,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      stateSave: true,
      data: list,
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
         {
            className: 'dt-control',
            orderable: false,
            data: null,
            defaultContent: '',
            width:'5%'
        },
        {
          data: "Pattern",
          title: "パターン",
          width: "80%",
        },
        {
          data: "Count",
          title: "回数",
          width: "15%",
          render: renderCount,
          className: "dt-body-right",
        },
      ],
    });
    table.on('click', 'tbody td.dt-control', function (e:any) {
      let tr = e.target.closest('tr');
      let row = table.row(tr);
      if (row.child.isShown()) {
        row.child.hide();
      } else {
        const d = row.data()
        row.child(d.Sample).show();
      }
  });    
    chart = showSyslogSummary("syslogSummary",list)
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
      <TabItem on:click={showSummary}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiFilterCheck} size={1} />
          正規化分析
        </div>
        <div id="syslogSummary"></div>
        <table
          id="syslogSummaryTable"
          class="display compact mt-5"
          style="width:99%"
        />
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