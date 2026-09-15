<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showSyslogLevelChart,showSyslogHost,showSyslogHost3D,showSyslogFFT3D, getSyslogSummary,showSyslogSummary, showSyslogAnomalyChart } from "./chart/syslog";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";
  import {
    getTableLang,
    renderCount,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  import { GetMapConf, LLMExplainSyslogReport, CalculateSyslogAnomaly } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let logs : datastore.SyslogEnt[] | undefined =undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "level";

  const onOpen = async () => {
    chart = undefined;
    activeTab = "level";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showChart("level");
  };

  let chart :any  = undefined;
  const showChart = async (t:string) => {
    activeTab = t;
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
    activeTab = "syslogSummary";
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
          title: $_('SyslogReport.Pattern'),
          width: "80%",
        },
        {
          data: "Count",
          title: $_('SyslogReport.Count'),
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

  let anomalyAlgo = "iforest";
  let anomalyVMode = "tfidf";
  let anomalyBusy = false;
  let anomalyDuration = "";
  let anomalyData: any[] = [];
  let anomalyTable: any = null;

  const renderAnomalyTable = (data: any[]) => {
    const tableEl = document.querySelector("#syslogAnomalyTable");
    if (!tableEl) return;
    if (anomalyTable) {
      try {
        anomalyTable.off('click', 'tbody td.dt-control');
        anomalyTable.destroy();
      } catch (e) {
        console.warn("destroy table error:", e);
      }
      anomalyTable = null;
    }
    anomalyTable = new DataTable("#syslogAnomalyTable", {
      destroy: true,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      stateSave: true,
      data: data,
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
        {
          className: 'dt-control',
          orderable: false,
          data: null,
          defaultContent: '',
          width: '5%',
        },
        {
          data: "Score",
          title: $_('SyslogReport.Score'),
          width: "10%",
          render: (val: any) => {
            const n = Number(val);
            let color = '#3b82f6';
            if (n >= 70) color = '#ef4444';
            else if (n >= 60) color = '#f59e0b';
            return `<span style="font-weight:bold;color:${color}">${n.toFixed(1)}</span>`;
          },
          className: "dt-body-right",
        },
        {
          data: "Time",
          title: $_('SyslogReport.Time'),
          width: "15%",
          render: (t: any) => {
            const ms = t > 1e12 ? Math.floor(t / 1e6) : t * 1000;
            return new Date(ms).toLocaleString();
          },
        },
        {
          data: "Host",
          title: $_('SyslogReport.Host'),
          width: "15%",
        },
        {
          data: "Tag",
          title: $_('SyslogReport.Tag'),
          width: "10%",
        },
        {
          data: "Message",
          title: $_('SyslogReport.Message'),
          width: "45%",
        },
      ],
    });
    anomalyTable.on('click', 'tbody td.dt-control', function (this: any, e: any) {
      let tr = (this as HTMLElement).closest('tr');
      if (!tr) return;
      let row = anomalyTable.row(tr);
      if (row.child.isShown()) {
        row.child.hide();
      } else {
        const d = row.data();
        row.child(`<div class="p-2 bg-gray-800 text-xs font-mono rounded overflow-auto break-all">${d.Message}</div>`).show();
      }
    });
  };

  const calcAnomaly = async () => {
    if (!logs || logs.length === 0 || anomalyBusy) return;
    anomalyBusy = true;
    anomalyDuration = "";
    try {
      const res = await CalculateSyslogAnomaly(logs, anomalyAlgo, anomalyVMode);
      if (res) {
        anomalyDuration = `${res.DurationMs} ms`;
        anomalyData = res.Logs || [];
        await tick();
        chart = showSyslogAnomalyChart("syslogAnomalyChart", anomalyData);
        renderAnomalyTable(anomalyData);
      }
    } catch (e) {
      console.error("CalculateSyslogAnomaly err:", e);
    } finally {
      anomalyBusy = false;
    }
  };

  const showAnomaly = async () => {
    activeTab = "anomaly";
    await tick();
    if (anomalyData && anomalyData.length > 0) {
      chart = showSyslogAnomalyChart("syslogAnomalyChart", anomalyData);
      renderAnomalyTable(anomalyData);
    }
  };

  const close = () => {
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  }


  $: if (show) {
    onOpen();
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline" contentClass="p-1 pt-1">
      <TabItem onclick={()=>{showChart("level")}}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          {$_('SyslogReport.CountByLevel')}
        </div>
      {/snippet}
        <div id="level"></div>
      </TabItem>
      <TabItem onclick={()=>{showChart("heatmap")}}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          {$_('SyslogReport.Heatmap')}
        </div>
      {/snippet}
        <div id="heatmap"></div>
      </TabItem>
      <TabItem onclick={()=>{showChart("host")}}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          {$_('SyslogReport.CountByHost')}
        </div>
      {/snippet}
        <div id="host"></div>
      </TabItem>
      <TabItem onclick={showSummary}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiFilterCheck} size={1} />
          {$_('SyslogReport.Summary')}
        </div>
      {/snippet}
        <div id="syslogSummary"></div>
        <div><table
          id="syslogSummaryTable"
          class="display compact mt-5"
          style="width:99%"></table></div>
      </TabItem>
      <TabItem onclick={()=>{showChart("host3D")}}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          {$_('SyslogReport.Chart3D')}
        </div>
      {/snippet}
        <div id="host3D"></div>
      </TabItem>
      <TabItem onclick={()=>{showChart("fft")}}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiChartLine} size={1} />
          {$_('SyslogReport.FFT')}
        </div>
      {/snippet}
        <div id="fft"></div>
      </TabItem>
      <TabItem onclick={showAnomaly}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiAlertDecagram} size={1} />
          {$_('SyslogReport.Anomaly')}
        </div>
      {/snippet}
        <div class="flex items-center gap-3 mt-1 mb-2.5 flex-wrap text-sm">
          <label class="flex items-center gap-1">
            <span class="text-xs text-gray-300">{$_('SyslogReport.Algorithm')}:</span>
            <select class="bg-gray-700 text-white rounded px-2 py-1 text-xs" bind:value={anomalyAlgo}>
              <option value="iforest">Isolation Forest</option>
              <option value="zscore">Z-Score</option>
              <option value="lof">Local Outlier Factor</option>
              <option value="knn">k-NN</option>
              <option value="mahalanobis">Mahalanobis</option>
              <option value="autoencoder">Auto Encoder</option>
              <option value="lstm">LSTM</option>
            </select>
          </label>
          <label class="flex items-center gap-1">
            <span class="text-xs text-gray-300">{$_('SyslogReport.VectorMode')}:</span>
            <select class="bg-gray-700 text-white rounded px-2 py-1 text-xs" bind:value={anomalyVMode}>
              <option value="tfidf">{$_('SyslogReport.TFIDF')}</option>
              <option value="security">{$_('SyslogReport.Security')}</option>
              <option value="alltime">{$_('SyslogReport.AllTime')}</option>
              <option value="time">{$_('SyslogReport.TimeMode')}</option>
              <option value="num">{$_('SyslogReport.NumMode')}</option>
            </select>
          </label>
          <GradientButton shadow type="button" color="blue" onclick={calcAnomaly} disabled={anomalyBusy} size="xs">
            <Icon path={icons.mdiPlay} size={0.8} />
            {anomalyBusy ? $_('SyslogReport.Calculating') : $_('SyslogReport.Calculate')}
          </GradientButton>
          {#if anomalyDuration}
            <span class="text-xs text-gray-400">{$_('SyslogReport.Duration')}: {anomalyDuration}</span>
          {/if}
        </div>
        {#if anomalyData && anomalyData.length > 0}
          <div id="syslogAnomalyChart"></div>
          <div class="mt-4">
            <table
              id="syslogAnomalyTable"
              class="display compact"
              style="width:99%"
            ></table>
          </div>
        {:else}
          <div class="flex flex-col items-center justify-center p-8 bg-gray-800/40 rounded-lg border border-gray-700 text-center my-6">
            <Icon path={icons.mdiInformationOutline} size={2} class="text-blue-400 mb-3" />
            <h4 class="text-base font-semibold text-gray-200 mb-2">{$_('SyslogReport.AnomalyGuideTitle') || 'Syslog 異常検知'}</h4>
            <p class="text-sm text-gray-400 max-w-lg mb-4">
              {$_('SyslogReport.AnomalyGuideDesc') || 'アルゴリズムと特徴抽出方法を選択し、「計算実行」ボタンをクリックしてください。ログの異常スコアを分析してグラフと一覧テーブルを表示します。'}
            </p>
            <div class="text-xs text-gray-400 space-y-1 text-left bg-gray-900/60 p-4 rounded border border-gray-700/60">
              <div><b>・{$_('SyslogReport.Algorithm')}:</b> Isolation Forest / Z-Score / LOF / k-NN / Mahalanobis / Auto Encoder / LSTM</div>
              <div><b>・{$_('SyslogReport.VectorMode')}:</b> {$_('SyslogReport.TFIDF')} / {$_('SyslogReport.Security')} / {$_('SyslogReport.AllTime')} / {$_('SyslogReport.TimeMode')} / {$_('SyslogReport.NumMode')}</div>
            </div>
          </div>
        {/if}
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if hasAI}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={() => (showAIReport = true)}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_("ReportAI.AIExplain")}
        </GradientButton>
      {/if}
      <GradientButton shadow type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_('SyslogReport.Close')}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`syslog_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainSyslogReport(logs || [], activeTab)}
/>

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
  #syslogSummary {
    min-height: 300px;
    height: 30vh;
    width: 98%;
    margin: 0 auto;
  }
  #syslogAnomalyChart {
    min-height: 300px;
    height: 30vh;
    width: 98%;
    margin: 0 auto 1.5rem auto;
  }
</style>