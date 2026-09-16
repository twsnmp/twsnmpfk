<script lang="ts">
  import { Modal, GradientButton,Tabs,TabItem } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {showSyslogLevelChart,showSyslogHost,showSyslogHost3D,showSyslogFFT3D, getSyslogSummary,showSyslogSummary, showSyslogAnomalyChart } from "./chart/syslog";
  import { showSigmaSeverityChart, showSigmaTagsChart, showSigmaTimelineChart } from "./chart/sigma";
  import { showLogHeatmap } from "./chart/eventlog";
  import { _ } from "svelte-i18n";
  import {
    getTableLang,
    renderCount,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  import { GetMapConf, LLMExplainSyslogReport, CalculateSyslogAnomaly, AnalyzeSigmaLogs, GetSigmaPacks } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let logs : datastore.SyslogEnt[] | undefined =undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "level";

  const onOpen = async () => {
    chart = undefined;
    activeTab = "level";
    sigmaResult = null;
    anomalyData = [];
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

  let sigmaMode: "threats" | "compliance" | "all" | "rules" | "tags" = "threats";
  let sigmaChartType: "severity" | "tags" | "timeline" = "severity";
  let sigmaPack: string = "all";
  let sigmaPacksList: any[] = [];
  let sigmaBusy = false;
  let sigmaDuration = "";
  let sigmaResult: any = null;
  let sigmaTable: any = null;

  const loadSigmaPacks = async () => {
    if (sigmaPacksList.length === 0) {
      try {
        const packs = await GetSigmaPacks();
        sigmaPacksList = packs || [];
      } catch (e) {
        console.warn("GetSigmaPacks err:", e);
      }
    }
  };

  const updateSigmaChart = () => {
    if (!sigmaResult) return;
    const isDark = document.documentElement.classList.contains("dark");
    if (sigmaChartType === "severity") {
      chart = showSigmaSeverityChart("syslogSigmaChart", sigmaResult.Stats, isDark);
    } else if (sigmaChartType === "tags") {
      chart = showSigmaTagsChart("syslogSigmaChart", sigmaResult.Stats.TopTags, isDark, $_('SyslogReport.SigmaTags'));
    } else if (sigmaChartType === "timeline") {
      chart = showSigmaTimelineChart("syslogSigmaChart", sigmaResult.Stats.Timeline, isDark);
    }
  };

  const renderSigmaTable = () => {
    const tableEl = document.querySelector("#syslogSigmaTable");
    if (!tableEl || !sigmaResult) return;

    if (sigmaTable) {
      try {
        sigmaTable.off('click', 'tbody td.dt-control');
        sigmaTable.destroy();
      } catch (e) {
        console.warn("destroy sigmaTable err:", e);
      }
      sigmaTable = null;
    }

    let columns: any[] = [];
    let tableData: any[] = [];

    if (sigmaMode === "threats" || sigmaMode === "compliance" || sigmaMode === "all") {
      let items = sigmaResult.Items || [];
      if (sigmaMode === "threats") {
        items = items.filter((i: any) => !i.IsCompliance);
      } else if (sigmaMode === "compliance") {
        items = items.filter((i: any) => i.IsCompliance);
      }
      tableData = items;

      columns = [
        {
          className: 'dt-control',
          orderable: false,
          data: null,
          defaultContent: '',
          width: '4%',
        },
        {
          data: "Level",
          title: $_('SyslogReport.SigmaLevel'),
          width: "9%",
          render: (lvl: string) => {
            const l = (lvl || "medium").toLowerCase();
            let bg = "#6e7681";
            if (l === "critical") bg = "#cf222e";
            else if (l === "high") bg = "#f85149";
            else if (l === "medium") bg = "#d29922";
            else if (l === "low") bg = "#58a6ff";
            return `<span style="display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;background-color:${bg};color:#fff;text-transform:uppercase;">${l}</span>`;
          },
        },
        {
          data: "Time",
          title: $_('SyslogReport.Time'),
          width: "14%",
          render: (t: number) => {
            const ms = t > 1e12 ? Math.floor(t / 1e6) : t * 1000;
            return new Date(ms).toLocaleString();
          },
        },
        {
          data: "Title",
          title: $_('SyslogReport.SigmaRuleTitle'),
          width: "23%",
        },
        {
          data: "LogSource",
          title: $_('SyslogReport.SigmaLogSource'),
          width: "12%",
        },
        {
          data: "Tags",
          title: $_('SyslogReport.SigmaTag'),
          width: "15%",
          render: (tags: string[]) => {
            if (!tags || tags.length === 0) return "";
            return tags.slice(0, 3).map((t) => {
              let color = "#38bdf8";
              const tl = t.toLowerCase();
              if (tl.startsWith("attack.")) color = "#f87171";
              else if (tl.startsWith("pci") || tl.startsWith("nist") || tl.startsWith("cis") || tl.startsWith("gdpr") || tl.startsWith("compliance")) color = "#4ade80";
              return `<span class="inline-block px-1.5 py-0.5 rounded text-[10px] mr-1 bg-gray-700 text-gray-200 border border-gray-600 font-mono" style="color:${color}">${t}</span>`;
            }).join("") + (tags.length > 3 ? `<span class="text-xs text-gray-400">+${tags.length - 3}</span>` : "");
          },
        },
        {
          data: "Message",
          title: $_('SyslogReport.Message'),
          width: "23%",
        },
      ];
    } else if (sigmaMode === "rules") {
      tableData = sigmaResult.Stats.TopRules || [];
      columns = [
        {
          data: "Count",
          title: $_('SyslogReport.SigmaCount'),
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "Level",
          title: $_('SyslogReport.SigmaLevel'),
          width: "12%",
          render: (lvl: string) => {
            const l = (lvl || "medium").toLowerCase();
            let bg = "#6e7681";
            if (l === "critical") bg = "#cf222e";
            else if (l === "high") bg = "#f85149";
            else if (l === "medium") bg = "#d29922";
            else if (l === "low") bg = "#58a6ff";
            return `<span style="display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;background-color:${bg};color:#fff;text-transform:uppercase;">${l}</span>`;
          },
        },
        {
          data: "Title",
          title: $_('SyslogReport.SigmaRuleTitle'),
          width: "38%",
        },
        {
          data: "Source",
          title: $_('SyslogReport.SigmaLogSource'),
          width: "15%",
          render: (s: string) => s.replace("pack:", ""),
        },
        {
          data: "Tags",
          title: $_('SyslogReport.SigmaTag'),
          width: "25%",
          render: (tags: string[]) => {
            if (!tags || tags.length === 0) return "";
            return tags.slice(0, 3).map((t) => `<span class="inline-block px-1.5 py-0.5 rounded text-[10px] mr-1 bg-gray-700 text-gray-300 font-mono">${t}</span>`).join("");
          },
        },
      ];
    } else if (sigmaMode === "tags") {
      tableData = sigmaResult.Stats.TopTags || [];
      columns = [
        {
          data: "Count",
          title: $_('SyslogReport.SigmaCount'),
          width: "15%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "Category",
          title: $_('SyslogReport.SigmaCategory'),
          width: "20%",
          render: (c: string) => {
            let label = c;
            let bg = "#0969da";
            if (c === "mitre") {
              label = "MITRE ATT&CK";
              bg = "#cf222e";
            } else if (c === "compliance") {
              label = "Compliance";
              bg = "#2ea44f";
            }
            return `<span style="padding:2px 8px;border-radius:10px;font-size:11px;color:#fff;background:${bg}">${label}</span>`;
          },
        },
        {
          data: "Tag",
          title: $_('SyslogReport.SigmaTag'),
          width: "65%",
        },
      ];
    }

    sigmaTable = new DataTable("#syslogSigmaTable", {
      destroy: true,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      stateSave: true,
      data: tableData,
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: columns,
    });

    if (sigmaMode === "threats" || sigmaMode === "compliance" || sigmaMode === "all") {
      sigmaTable.on('click', 'tbody td.dt-control', function (this: any, e: any) {
        let tr = (this as HTMLElement).closest('tr');
        if (!tr) return;
        let row = sigmaTable.row(tr);
        if (row.child.isShown()) {
          row.child.hide();
        } else {
          const d = row.data();
          row.child(`
            <div class="p-3 bg-gray-900/90 text-xs rounded border border-gray-700 space-y-1">
              <div><b class="text-gray-400">Rule ID:</b> <span class="font-mono text-gray-200">${d.RuleID}</span></div>
              <div><b class="text-gray-400">Source:</b> <span class="font-mono text-gray-200">${d.Source}</span></div>
              <div><b class="text-gray-400">Tags:</b> <span class="font-mono text-gray-200">${(d.Tags || []).join(", ")}</span></div>
              <div><b class="text-gray-400">Log Message:</b></div>
              <pre class="p-2 bg-black/60 rounded text-green-400 overflow-x-auto whitespace-pre-wrap font-mono">${d.Log || d.Message}</pre>
            </div>
          `).show();
        }
      });
    }
  };

  const calcSigma = async () => {
    if (!logs || logs.length === 0 || sigmaBusy) return;
    sigmaBusy = true;
    sigmaDuration = "";
    const st = Date.now();
    try {
      const packs = sigmaPack === "all" ? [] : [sigmaPack];
      const res = await AnalyzeSigmaLogs(logs, packs, []);
      if (res) {
        sigmaDuration = `${Date.now() - st} ms`;
        sigmaResult = res;
        await tick();
        updateSigmaChart();
        renderSigmaTable();
      }
    } catch (e) {
      console.error("AnalyzeSigmaLogs err:", e);
    } finally {
      sigmaBusy = false;
    }
  };

  const showSigma = async () => {
    activeTab = "sigma";
    await loadSigmaPacks();
    await tick();
    if (sigmaResult) {
      updateSigmaChart();
      renderSigmaTable();
    }
  };

  const handleSigmaModeChange = async () => {
    await tick();
    renderSigmaTable();
  };

  const handleSigmaChartTypeChange = async () => {
    await tick();
    updateSigmaChart();
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
  class="w-full max-h-[92vh] p-2"
>
  <div class="flex flex-col h-[84vh] max-h-[84vh] overflow-hidden">
    <div class="flex-1 overflow-y-auto pr-1">
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
      <TabItem onclick={showSigma}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiShieldAlert} size={1} />
          {$_('SyslogReport.Sigma')}
        </div>
      {/snippet}
        <div class="flex items-center gap-3 mt-1 mb-2.5 flex-wrap text-sm">
          <label class="flex items-center gap-1">
            <span class="text-xs text-gray-300">{$_('SyslogReport.SigmaMode')}:</span>
            <select class="bg-gray-700 text-white rounded px-2 py-1 text-xs" bind:value={sigmaMode} onchange={handleSigmaModeChange}>
              <option value="threats">{$_('SyslogReport.SigmaThreats')}</option>
              <option value="compliance">{$_('SyslogReport.SigmaCompliance')}</option>
              <option value="all">{$_('SyslogReport.SigmaAll')}</option>
              <option value="rules">{$_('SyslogReport.SigmaRules')}</option>
              <option value="tags">{$_('SyslogReport.SigmaTags')}</option>
            </select>
          </label>
          <label class="flex items-center gap-1">
            <span class="text-xs text-gray-300">{$_('SyslogReport.SigmaChartType')}:</span>
            <select class="bg-gray-700 text-white rounded px-2 py-1 text-xs" bind:value={sigmaChartType} onchange={handleSigmaChartTypeChange}>
              <option value="severity">{$_('SyslogReport.SigmaChartSeverity')}</option>
              <option value="tags">{$_('SyslogReport.SigmaChartTags')}</option>
              <option value="timeline">{$_('SyslogReport.SigmaChartTimeline')}</option>
            </select>
          </label>
          <label class="flex items-center gap-1">
            <span class="text-xs text-gray-300">{$_('SyslogReport.SigmaPacks')}:</span>
            <select class="bg-gray-700 text-white rounded px-2 py-1 text-xs" bind:value={sigmaPack}>
              <option value="all">{$_('SyslogReport.SigmaAllPacks')}</option>
              {#each sigmaPacksList as p}
                <option value={p.name}>{p.name} ({p.rule_count})</option>
              {/each}
            </select>
          </label>
          <GradientButton shadow type="button" color="blue" onclick={calcSigma} disabled={sigmaBusy} size="xs">
            <Icon path={icons.mdiShieldSearch} size={0.8} />
            {sigmaBusy ? $_('SyslogReport.SigmaDetecting') : $_('SyslogReport.SigmaDetect')}
          </GradientButton>
          {#if sigmaDuration}
            <span class="text-xs text-gray-400">{$_('SyslogReport.Duration')}: {sigmaDuration}</span>
          {/if}
        </div>

        {#if sigmaResult}
          <!-- Overview summary metrics cards -->
          <div class="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-8 gap-1.5 my-2 text-center">
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-gray-400">{$_('SyslogReport.SigmaTotalScanned')}</div>
              <div class="text-sm font-bold text-gray-200">{sigmaResult.Stats.TotalLogs.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-red-400">{$_('SyslogReport.SigmaTotalHits')}</div>
              <div class="text-sm font-bold text-red-400">{sigmaResult.Stats.TotalDetections.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-red-500">Critical</div>
              <div class="text-sm font-bold text-red-500">{sigmaResult.Stats.Critical.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-orange-400">High</div>
              <div class="text-sm font-bold text-orange-400">{sigmaResult.Stats.High.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-yellow-400">Medium</div>
              <div class="text-sm font-bold text-yellow-400">{sigmaResult.Stats.Medium.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-blue-400">Low / Info</div>
              <div class="text-sm font-bold text-blue-400">{(sigmaResult.Stats.Low + sigmaResult.Stats.Informational).toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-green-400">{$_('SyslogReport.SigmaComplianceHits')}</div>
              <div class="text-sm font-bold text-green-400">{sigmaResult.Stats.ComplianceHits.toLocaleString()}</div>
            </div>
            <div class="bg-gray-800/80 p-1.5 rounded border border-gray-700">
              <div class="text-[10px] text-gray-400">{$_('SyslogReport.SigmaActiveRules')}</div>
              <div class="text-sm font-bold text-gray-200">{sigmaResult.Stats.ActiveRules.toLocaleString()}</div>
            </div>
          </div>

          <div id="syslogSigmaChart"></div>

          <div class="mt-4">
            <table
              id="syslogSigmaTable"
              class="display compact"
              style="width:99%"
            ></table>
          </div>
        {:else}
          <div class="flex flex-col items-center justify-center p-8 bg-gray-800/40 rounded-lg border border-gray-700 text-center my-6">
            <Icon path={icons.mdiShieldAlert} size={2} class="text-blue-400 mb-3" />
            <h4 class="text-base font-semibold text-gray-200 mb-2">{$_('SyslogReport.SigmaTitle')}</h4>
            <p class="text-sm text-gray-400 max-w-lg mb-4">
              {$_('SyslogReport.SigmaGuideDesc')}
            </p>
            <div class="text-xs text-gray-400 space-y-1 text-left bg-gray-900/60 p-4 rounded border border-gray-700/60">
              <div><b>・{$_('SyslogReport.SigmaPacks')}:</b> {$_('SyslogReport.SigmaAllPacks')} / windows-essential / linux-auth / network-threats / web-attacks / wazuh-compliance ...</div>
              <div><b>・{$_('SyslogReport.SigmaMode')}:</b> {$_('SyslogReport.SigmaThreats')} / {$_('SyslogReport.SigmaCompliance')} / {$_('SyslogReport.SigmaAll')} / {$_('SyslogReport.SigmaRules')} / {$_('SyslogReport.SigmaTags')}</div>
              <div><b>・{$_('SyslogReport.SigmaChartType')}:</b> {$_('SyslogReport.SigmaChartSeverity')} / {$_('SyslogReport.SigmaChartTags')} / {$_('SyslogReport.SigmaChartTimeline')}</div>
            </div>
          </div>
        {/if}
      </TabItem>
    </Tabs>
    </div>
    <div class="flex justify-end space-x-2 mr-2 pt-2.5 mt-auto border-t border-gray-700 bg-gray-800 shrink-0">
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
  #syslogAnomalyChart,
  #syslogSigmaChart {
    min-height: 300px;
    height: 30vh;
    width: 98%;
    margin: 0 auto 1.5rem auto;
  }
</style>