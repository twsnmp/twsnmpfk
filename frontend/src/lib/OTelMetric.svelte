<script lang="ts">
  import { Modal, GradientButton, Select } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { GetOTelMetric } from "../../wailsjs/go/main/App";
  import { showOTelTimeChart, showOTelHistogram } from "./chart/otel";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { getTableLang, renderTime } from "./common";

  export let show: boolean = false;
  export let metric: any = undefined;

  let data: any = [];
  let metricType = "";
  let metricName = "";
  let metricIndexes: any = [];
  let metricIndex = "";

  const onOpen = async () => {
    const d = await GetOTelMetric(metric);
    if (d && d.DataPoints) {
      data = d.DataPoints;
      metricType = d.Type;
      metricName = d.Host + " " + d.Service + " " + d.Scope + " " + d.Name;
      const m = new Map();
      metricIndexes = [];
      for (const d of data) {
        const k = d.Attributes.join(" ");
        if (!m.get(k)) {
          m.set(k, true);
          metricIndexes.push({
            name: k,
            value: k,
          });
        }
      }
      if (metricIndexes.length > 0) {
        metricIndex = metricIndexes[0].value;
      }
    }
    showTable();
    showTimeChart();
  };

  let chart: any = undefined;

  const showTimeChart = async () => {
    await tick();
    chart = showOTelTimeChart("metricChart", data, metricIndex, metricType);
  };

  const showHistogram = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1 || !d[0].BucketCounts) {
      return;
    }
    await tick();
    chart = showOTelHistogram("metricChart", d[0]);
  };

  const columns = [
    {
      data: "Time",
      title: $_('OTel.Time'),
      width: "10%",
      render: renderTime,
    },
    {
      data: "Start",
      title: $_('OTel.StartMetric'),
      width: "10%",
      render: renderTime,
    },
    {
      data: "Attributes",
      title: $_('OTel.Attributes'),
      width: "40%",
      render: (s: any) => s.join(" "),
    },
    {
      data: "Count",
      title: $_('OTel.Count'),
      width: "10%",
    },
    {
      data: "Sum",
      title: $_('OTel.Sum'),
      width: "10%",
    },
    {
      data: "Min",
      title: $_('OTel.Min'),
      width: "10%",
    },
    {
      data: "Max",
      title: $_('OTel.Max'),
      width: "10%",
    },
  ];

  let selectedCount = 0;
  let table: any = undefined;

  const showTable = async () => {
    selectedCount = 0;
    table = new DataTable("#metricTable", {
      destroy: true,
      paging: false,
      searching: false,
      info: false,
      scrollY: "20vh",
      columns: columns,
      data: data,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
      if (selectedCount == 1) {
        showHistogram();
      }
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
      if (selectedCount < 1)  {
        showTimeChart();
      }
    });
  };

  const close = () => {
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
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
    <div class="mb-2">{metricName}</div>
    <Select
      items={metricIndexes}
      bind:value={metricIndex}
      on:change={showTimeChart}
      placeholder=""
      class="h-10 mb-2"
      size="sm"
    />
    <div id="metricChart"></div>
    <div>
      <table id="metricTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("SyslogReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #metricChart {
    min-height: 400px;
    height: 50vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
