<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Spinner,
    Tabs,
    TabItem,
    MultiSelect,
    Table,
    TableBody,
    TableHead,
    TableHeadCell,
    TableBodyCell,
    TableBodyRow,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetOTelMetrics,
    GetOTelMetric,
    GetOTelTraceBucketList,
    GetLastOTelLogs,
    GetOTelTraces,
    DeleteAllOTelData,
    GetOTelTraceDAG,
  } from "../../wailsjs/go/main/App";
  import {
    renderTime,
    getTableLang,
    renderTimeMili,
    renderState,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import OTelMetric from "./OTelMetric.svelte";
  import OTelTrace from "./OTelTrace.svelte";
  import { showOTelTrace, showOTelDAG } from "./chart/otel";
  import { showLogLevelChart } from "./chart/loglevel";

  let metrics: any = [];
  let traces: any = [];
  let logs: any = [];
  let tab = "metric";
  let traceBuckets: any = [];
  let selectedTraceBuckets: any = [];
  let showMetricReport = false;
  let showTraceReport = false;
  let table: any = undefined;
  let selectedCount = 0;
  let showLoading = false;

  const showTable = (div: string, columns: any, data: any, scol: number) => {
    selectedCount = 0;
    table = new DataTable(div, {
      destroy: true,
      columns: columns,
      pageLength: 10,
      stateSave: true,
      scrollX: true,
      data: data,
      order: [[scol, "desc"]],
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

  const refresh = async () => {
    showLoading = true;
    switch (tab) {
      case "metric":
        metrics = await GetOTelMetrics();
        showTable("#otelMetricTable", columnsMetric, metrics, 0);
        break;
      case "trace":
        const bks = await GetOTelTraceBucketList();
        const sel: any = [];
        selectedTraceBuckets.forEach((t: any) => {
          if (bks.includes(t)) {
            sel.push(t);
          }
        });
        selectedTraceBuckets = sel;
        traceBuckets = [];
        bks.forEach((b: string) => {
          traceBuckets.push({ name: b, value: b });
        });
        if (selectedTraceBuckets.length < 1 && bks.length > 0) {
          selectedTraceBuckets.push(bks[bks.length - 1]);
        }
        if (selectedTraceBuckets.length > 0) {
          traces = await GetOTelTraces(selectedTraceBuckets);
        } else {
          traces = [];
        }
        showTable("#otelTraceTable", columnsTrace, traces, 0);
        showTraceChart();
        break;
      case "log":
        logs = await GetLastOTelLogs();
        showTable("#otelLogTable", columnsLog, logs, 1);
        showSyslogChart();
        break;
    }
    showLoading = false;
  };

  let chart: any = undefined;

  const showTraceChart = async () => {
    await tick();
    chart = showOTelTrace("otelTraceChart", traces);
  };

  const showSyslogChart = async () => {
    await tick();
    chart = showLogLevelChart("otelSyslogChart", logs, undefined);
  };

  let showDAG = false;
  let dagChart: any = undefined;

  const showDAGFunc = async () => {
    if (selectedTraceBuckets.length < 1) {
      return;
    }
    showLoading = true;
    const dag = await GetOTelTraceDAG(selectedTraceBuckets);
    showDAG = dag && dag.Nodes.length > 0;
    await tick();
    dagChart = showOTelDAG("dagChart", dag);
    showLoading = false;
  };

  const columnsMetric = [
    {
      data: "Host",
      title: $_("OTel.Host"),
      width: "10%",
    },
    {
      data: "Service",
      title: $_("OTel.Service"),
      width: "15%",
    },
    {
      data: "Scope",
      title: $_("OTel.Scope"),
      width: "25%",
    },
    {
      data: "Name",
      title: $_("OTel.Name"),
      width: "15%",
    },
    {
      data: "Type",
      title: $_("OTel.Type"),
      width: "10%",
    },
    {
      data: "Count",
      title: $_("OTel.Count"),
      width: "5%",
    },
    {
      data: "First",
      title: $_("OTel.First"),
      width: "10%",
      render: renderTime,
    },
    {
      data: "Last",
      title: $_("OTel.Last"),
      width: "10%",
      render: renderTime,
    },
  ];

  const columnsTrace = [
    {
      data: "Start",
      title: $_("OTel.Start"),
      width: "15%",
      render: renderTimeMili,
    },
    {
      data: "End",
      title: $_("OTel.End"),
      width: "15%",
      render: renderTimeMili,
    },
    {
      data: "Dur",
      title: $_("OTel.Dur") + "(mSec)",
      width: "10%",
      render: (v: number) => (v * 1000).toFixed(3),
    },
    {
      data: "TraceID",
      title: $_("OTel.TraceID"),
      width: "15%",
    },
    {
      data: "Hosts",
      title: $_("OTel.Host"),
      width: "10%",
    },
    {
      data: "Services",
      title: $_("OTel.Service"),
      width: "15%",
    },
    {
      data: "NumSpan",
      title: "Span",
      width: "5%",
    },
    {
      data: "Scopes",
      title: $_("OTel.Scope"),
      width: "20%",
    },
  ];

  const columnsLog = [
    {
      data: "Level",
      title: $_("Syslog.Level"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("Syslog.Time"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Host",
      title: $_("Syslog.Host"),
      width: "15%",
    },
    {
      data: "Type",
      title: $_("Syslog.Type"),
      width: "10%",
    },
    {
      data: "Tag",
      title: $_("Syslog.Tag"),
      width: "10%",
    },
    {
      data: "Message",
      title: $_("Syslog.Message"),
      width: "40%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const deleteAll = async () => {
    await DeleteAllOTelData();
    refresh();
  };

  let selectedMetric: any = undefined;
  let selectedTrace: any = undefined;

  const showReport = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
      return;
    }
    switch (tab) {
      case "metric":
        selectedMetric = d[0];
        showMetricReport = true;
        break;
      case "trace":
        selectedTrace = d[0];
        showTraceReport = true;
        break;
      case "log":
        break;
    }
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  let showMetricInfoDialog = false;
  let metric: any = undefined;

  const showMetricInfo = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
      return;
    }
    metric = await GetOTelMetric(d[0]);
    if (!metric) {
      return;
    }
    showMetricInfoDialog = true;
  };
</script>

<svelte:window on:resize={resizeChart} />

<div class="flex flex-col">
  <Tabs style="underline">
    <TabItem
      open
      on:click={() => {
        tab = "metric";
        refresh();
      }}
    >
      <div slot="title" class="flex items-center gap-2">
        <Icon path={icons.mdiChartHistogram} size={1} />
        {$_("OTel.Metric")}
      </div>
      <table id="otelMetricTable" class="display compact" style="width:99%" />
    </TabItem>
    <TabItem
      on:click={() => {
        tab = "trace";
        refresh();
      }}
    >
      <div slot="title" class="flex items-center gap-2">
        <Icon path={icons.mdiEye} size={1} />
        {$_("OTel.Trace")}
      </div>
      <div id="otelTraceChart" />
      <div class="m-5 grow">
        <table id="otelTraceTable" class="display compact" style="width:99%" />
      </div>
    </TabItem>
    <TabItem
      on:click={() => {
        tab = "log";
        refresh();
      }}
    >
      <div slot="title" class="flex items-center gap-2">
        <Icon path={icons.mdiTextBox} size={1} />
        {$_("OTel.Log")}
      </div>
      <div id="otelSyslogChart" />
      <div class="m-5 grow">
        <table id="otelLogTable" class="display compact" style="width:99%" />
      </div>
    </TabItem>
  </Tabs>

  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1 && tab != "log"}
      {#if tab == "metric"}
        <GradientButton
          shadow
          color="green"
          type="button"
          on:click={showMetricInfo}
          size="xs"
        >
          <Icon path={icons.mdiInformation} size={1} />
          {$_('OTel.MetricInfo')}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={showReport}
        size="xs"
      >
        <Icon path={icons.mdiEye} size={1} />
        {$_("OTerl.Report")}
      </GradientButton>
    {/if}
    {#if tab == "trace"}
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={showDAGFunc}
        size="xs"
      >
        <Icon path={icons.mdiGraph} size={1} />
        DAG
      </GradientButton>
      <MultiSelect
        items={traceBuckets}
        bind:value={selectedTraceBuckets}
        placeholder="Date and Time"
        class="h-10 mb-2 w-96"
        size="sm"
      />
    {/if}
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Syslog.DeleteAllLogs")}
    </GradientButton>

    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("Trap.Reload")}
    </GradientButton>
  </div>
</div>

<Modal bind:open={showLoading} size="sm" dismissable={false} class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("Syslog.Loading")} </span>
  </div>
</Modal>

<Modal bind:open={showDAG} size="xl" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <div id="dagChart"></div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showDAG = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("OTel.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showMetricInfoDialog}
  size="xl"
  dismissable={false}
  class="w-full"
>
  <div class="flex flex-col space-y-4">
    <Table>
      <TableHead>
        <TableHeadCell>{$_('OTel.Item')}</TableHeadCell>
        <TableHeadCell>{$_('OTel.Content')}</TableHeadCell>
      </TableHead>
      <TableBody>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Host')}</TableBodyCell>
          <TableBodyCell>{metric.Host}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Service')}</TableBodyCell>
          <TableBodyCell>{metric.Service}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Scope')}</TableBodyCell>
          <TableBodyCell>{metric.Scope}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Name')}</TableBodyCell>
          <TableBodyCell>{metric.Name}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Type')}</TableBodyCell>
          <TableBodyCell>{metric.Type}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Unit')}</TableBodyCell>
          <TableBodyCell>{metric.Unit}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Descr')}</TableBodyCell>
          <TableBodyCell>{metric.Description}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Count')}</TableBodyCell>
          <TableBodyCell>{metric.Count}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.First')}</TableBodyCell>
          <TableBodyCell>{renderTime(metric.First,"")}</TableBodyCell>
        </TableBodyRow>
        <TableBodyRow>
          <TableBodyCell>{$_('OTel.Last')}</TableBodyCell>
          <TableBodyCell>{renderTime(metric.Last,"")}</TableBodyCell>
        </TableBodyRow>
      </TableBody>
    </Table>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showMetricInfoDialog = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("OTel.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<OTelMetric bind:show={showMetricReport} metric={selectedMetric} />
<OTelTrace bind:show={showTraceReport} trace={selectedTrace} />

<style>
  #otelTraceChart,
  #otelSyslogChart {
    min-height: 200px;
    height: 20vh;
    width: 98%;
    margin: 0 auto;
  }
  #dagChart {
    min-height: 300px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>
