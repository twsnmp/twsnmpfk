<script lang="ts">
  import { Modal, GradientButton } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js"; 
  import { getTableLang, renderTimeMili } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { GetOTelTrace } from "../../wailsjs/go/main/App";
  import {showOTelTimeline  } from "./chart/otel";


  export let show: boolean = false;
  export let trace : any =undefined;

  let data : any = undefined;
  let spans : any = [];

  let name = "";

  const onOpen = async () => {
    data = await GetOTelTrace(trace.Bucket,trace.TraceID);
    name = trace.Hosts + " " + trace.Services + " " + trace.Scopes;
    if (data) {
      spans = data.Spans;
      showTimeline();
      showTable();
    }
  };

  let chart :any  = undefined;

  const showTimeline = async () => {
    await tick();
    chart = showOTelTimeline("traceTimeline", data);
  }

  const columns = [
    {
      data: "Name",
      title: $_('OTel.Name'),
      width: "15%",
    },
    {
      data: "Service",
      title: $_('OTel.Service'),
      width: "10%",
    },
    {
      data: "Start",
      title: $_('OTel.StartTime'),
      width: "13%",
      render: renderTimeMili,
    },
    {
      data: "End",
      title: $_('OTel.EndTime'),
      width: "13%",
      render: renderTimeMili,
    },
    {
      data: "Dur",
      title: $_('OTel.Dur'),
      width: "8%",
      render: (v:number) => (v * 1000).toFixed(3) + " mSec"
    },
    {
      data: "SpanID",
      title: "Span ID",
      width: "12%",
    },
    {
      data: "ParentSpanID",
      title: $_('OTel.PSpanID'),
      width: "12%",
    },
    {
      data: "Attributes",
      title: $_('OTel.Attributes'),
      width: "12%",
      render: (s: any) => s.join(" "),
    },
  ];

  let selectedCount = 0;
  let table: any = undefined;

  const showTable = async () => {
    selectedCount = 0;
    table = new DataTable("#traceTable", {
      destroy: true,
      paging: false,
      searching: false,
      info: false,
      scrollY: "20vh",
      columns: columns,
      data: spans,
      order: [[2, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  };



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
    <div class="mb-5">
      {name}
    </div>
    <div id="traceTimeline" />
    <div>
      <table id="traceTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_('SyslogReport.Close')}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #traceTimeline{
    min-height: 500px;
    height: 50vh;
    width: 98%;
    margin: 0 auto;
  }
</style>