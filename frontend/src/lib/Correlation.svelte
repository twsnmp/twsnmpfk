<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    Button,
    ButtonGroup,
    GradientButton,
    Select,
    Label,
    Checkbox,
    Spinner,
    Badge,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy, tick } from "svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { currentLang } from "../i18n/i18n";
  import {
    GetCorrelationCandidates,
    AnalyzePollingCorrelation,
    DiscoverPollingCorrelations,
  } from "../../wailsjs/go/main/App";
  import { getTableLang } from "./common";
  import {
    showCorrelationTimeChart,
    showCorrelationScatterChart,
    showCorrelationLagChart,
    showCorrelationMatrixChart,
    showCorrelationNetworkChart,
    disposeCorrelationChart,
  } from "./chart/correlation";

  export let initialPollingID1 = "";
  export let initialPollingID2 = "";
  export let initialMode: "discover" | "pair" = "discover";
  export let modalMode = false;

  let activeTab: "discover" | "pair" = initialMode;
  let candidates: any[] = [];
  let loadingCandidates = false;

  // Discovery state
  let discoverScope: "all" | "base" = initialPollingID1 ? "base" : "all";
  let basePollingID = initialPollingID1;
  let discoverAlgo = "pearson";
  let discoverPeriodHours = 168;
  let discoverBucketSec = 0;
  let minAbsScore = 0.7;
  let minSamples = 10;
  let crossNodeOnly = true;
  let discoverIncludeState = false;
  let discoverMaxLag = 6;
  let discoverLimit = 100;
  let discovering = false;
  let discoverRes: any = null;
  let discoverTable: any = undefined;
  let selectedDiscoverItem: any = null;
  let networkChart: any = null;

  // Pair analysis state
  let nodeFilter1 = "";
  let nodeFilter2 = "";
  let pollingID1 = initialPollingID1;
  let pollingID2 = initialPollingID2;
  let selectedVars1: string[] = [];
  let selectedVars2: string[] = [];
  let pairAlgo = "pearson";
  let pairPeriodHours = 168;
  let pairBucketSec = 0;
  let pairMaxLag = 12;
  let pairIncludeState = false;
  let analyzing = false;
  let pairRes: any = null;
  let pairChartMode: "time" | "scatter" | "lag" | "matrix" = "time";
  let timeNormalized = false;
  let pairChart: any = null;

  $: scopeItems = ($_, $currentLang, [
    { value: "all", name: $_("Correlation.ScopeAll") || "全体 (N対N)" },
    { value: "base", name: $_("Correlation.ScopeBase") || "基準指定 (1対N)" },
  ]);

  $: algoItems = ($_, $currentLang, [
    { value: "pearson", name: $_("Correlation.AlgoPearson") || "Pearson (線形)" },
    { value: "spearman", name: $_("Correlation.AlgoSpearman") || "Spearman (順位)" },
    { value: "kendall", name: $_("Correlation.AlgoKendall") || "Kendall (順位τ)" },
    { value: "crosscorr", name: $_("Correlation.AlgoCrossCorr") || "Cross-Corr (ラグ)" },
    { value: "dcor", name: $_("Correlation.AlgoDCor") || "Distance Corr (非線形/多変量)" },
  ]);

  $: periodItems = ($_, $currentLang, [
    { value: 24, name: $_("Correlation.Period24h") || "直近 24時間" },
    { value: 168, name: $_("Correlation.Period7d") || "直近 7日間" },
    { value: 720, name: $_("Correlation.Period30d") || "直近 30日間" },
    { value: -1, name: $_("Correlation.PeriodAll") || "全期間" },
  ]);

  $: bucketItems = ($_, $currentLang, [
    { value: 0, name: $_("Correlation.BucketAuto") || "自動" },
    { value: 60, name: "1m (60s)" },
    { value: 300, name: "5m (300s)" },
    { value: 900, name: "15m (900s)" },
    { value: 3600, name: "1h (3600s)" },
  ]);

  const minScoreItems = [
    { value: 0.5, name: ">= 0.5" },
    { value: 0.6, name: ">= 0.6" },
    { value: 0.7, name: ">= 0.7" },
    { value: 0.8, name: ">= 0.8" },
    { value: 0.9, name: ">= 0.9" },
  ];

  $: nodeOptions = ($_, $currentLang, (() => {
    const map = new Map<string, string>();
    for (const c of candidates) {
      if (c.NodeID && !map.has(c.NodeID)) {
        map.set(c.NodeID, c.NodeName || c.NodeID);
      }
    }
    const opts = [{ value: "", name: $_("Correlation.AllNodes") || "全ノード" }];
    const sorted = Array.from(map.entries()).sort((a, b) => a[1].localeCompare(b[1]));
    for (const [id, name] of sorted) {
      opts.push({ value: id, name });
    }
    return opts;
  })());

  const buildPollingOptions = (nodeFilter: string) =>
    candidates
      .filter((c) => !nodeFilter || c.NodeID === nodeFilter)
      .map((c) => ({
        value: c.PollingID,
        name: `${c.NodeName || "?"} / ${c.PollingName} (${c.Type})`,
      }));

  $: pollingOptionsAll = buildPollingOptions("");
  $: pollingOptions1 = buildPollingOptions(nodeFilter1);
  $: pollingOptions2 = buildPollingOptions(nodeFilter2);

  $: cand1 = candidates.find((c) => c.PollingID === pollingID1);
  $: cand2 = candidates.find((c) => c.PollingID === pollingID2);
  $: availKeys1 = cand1?.Keys || [];
  $: availKeys2 = cand2?.Keys || [];

  const getTimeBounds = (hours: number): { startTime: number; endTime: number } => {
    if (hours < 0) {
      return { startTime: 0, endTime: 0 };
    }
    const endTime = Math.floor(Date.now() / 1000);
    const startTime = endTime - hours * 3600;
    return { startTime, endTime };
  };

  const destroyDiscoverTable = () => {
    if (discoverTable && DataTable.isDataTable("#corrDiscoverTable")) {
      try {
        discoverTable.off("select");
        discoverTable.off("deselect");
        discoverTable.destroy();
      } catch {
        // ignore
      }
      discoverTable = undefined;
    }
  };

  const formatScoreBadge = (score: number, type: string) => {
    if (type === "sort") return Math.abs(score);
    const abs = Math.abs(score);
    const cls =
      score >= 0.7
        ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
        : score <= -0.7
        ? "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
        : abs >= 0.5
        ? "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200"
        : "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";
    return `<span class="px-2 py-0.5 rounded text-xs font-semibold ${cls}">${score.toFixed(4)}</span>`;
  };

  const formatPValue = (p: number) => {
    if (p < 0.0001) return "< 0.0001";
    if (p < 0.001) return "< 0.001";
    return p.toFixed(4);
  };

  const openPairFromItem = async (item: any) => {
    if (!item) return;
    nodeFilter1 = "";
    nodeFilter2 = "";
    pollingID1 = item.PollingID1;
    pollingID2 = item.PollingID2;
    selectedVars1 = item.Var1 ? item.Var1.split(",").filter(Boolean) : [];
    selectedVars2 = item.Var2 ? item.Var2.split(",").filter(Boolean) : [];
    pairAlgo = item.Algorithm || discoverAlgo || "pearson";
    pairPeriodHours = discoverPeriodHours;
    pairBucketSec = discoverBucketSec;
    activeTab = "pair";
    await tick();
    await runPairAnalysis();
  };

  const renderDiscoverResults = async () => {
    await tick();
    destroyDiscoverTable();
    selectedDiscoverItem = null;
    const el = document.getElementById("corrDiscoverTable");
    const items = discoverRes?.Items || [];
    if (el) {
      el.innerHTML = "";
      discoverTable = new DataTable("#corrDiscoverTable", {
        destroy: true,
        data: items,
        pageLength: 10,
        order: [],
        language: getTableLang(),
        select: { style: "single" },
        columns: [
          {
            data: "NodeName1",
            title: $_("Correlation.NodeA") || "ノード A",
            width: "13%",
          },
          {
            data: "PollingName1",
            title: $_("Correlation.PollingA") || "ポーリング A",
            width: "17%",
            render: (v: string, _t: string, row: any) =>
              `${v} <span class="text-xs text-gray-400">(${row.Var1})</span>`,
          },
          {
            data: "NodeName2",
            title: $_("Correlation.NodeB") || "ノード B",
            width: "13%",
          },
          {
            data: "PollingName2",
            title: $_("Correlation.PollingB") || "ポーリング B",
            width: "17%",
            render: (v: string, _t: string, row: any) =>
              `${v} <span class="text-xs text-gray-400">(${row.Var2})</span>`,
          },
          {
            data: "Score",
            title: $_("Correlation.Score") || "相関係数",
            width: "12%",
            render: formatScoreBadge,
          },
          {
            data: "LagSec",
            title: $_("Correlation.Lag") || "ラグ",
            width: "10%",
            render: (_v: number, _t: string, row: any) => {
              if (!row.Lag) return "0";
              const sign = row.Lag > 0 ? "+" : "";
              return `${sign}${row.Lag} (${sign}${row.LagSec}s)`;
            },
          },
          {
            data: "SampleCount",
            title: $_("Correlation.Samples") || "サンプル数",
            width: "9%",
          },
          {
            data: "PValue",
            title: "p-value",
            width: "9%",
            render: (v: number, t: string) => (t === "sort" ? v : formatPValue(v)),
          },
        ],
      });
      discoverTable.on("select", (_e: any, _dt: any, _type: any, indexes: number[]) => {
        if (!discoverTable || !indexes || indexes.length === 0) return;
        selectedDiscoverItem = discoverTable.row(indexes[0]).data();
      });
      discoverTable.on("deselect", () => {
        selectedDiscoverItem = null;
      });
      el.ondblclick = (ev: MouseEvent) => {
        const tr = (ev.target as HTMLElement)?.closest("tbody tr");
        if (!tr || !discoverTable) return;
        const rowData = discoverTable.row(tr).data();
        if (rowData) {
          openPairFromItem(rowData);
        }
      };
    }
    disposeCorrelationChart("corrNetworkChart");
    if (items.length > 0) {
      networkChart = showCorrelationNetworkChart("corrNetworkChart", items, (it) => {
        openPairFromItem(it);
      });
    }
  };

  const runDiscovery = async () => {
    discovering = true;
    try {
      const { startTime, endTime } = getTimeBounds(discoverPeriodHours);
      discoverRes = await DiscoverPollingCorrelations({
        BasePollingID: discoverScope === "base" ? basePollingID : "",
        NodeIDs: [],
        Algorithm: discoverAlgo,
        BucketSec: Number(discoverBucketSec) || 0,
        StartTime: startTime,
        EndTime: endTime,
        MinAbsScore: Number(minAbsScore) || 0.7,
        MinSamples: Number(minSamples) || 10,
        CrossNodeOnly: crossNodeOnly,
        IncludeState: discoverIncludeState,
        MaxLag: Number(discoverMaxLag) || 6,
        Limit: Number(discoverLimit) || 100,
      } as any);
      await renderDiscoverResults();
    } finally {
      discovering = false;
    }
  };

  const drawPairChart = async () => {
    await tick();
    if (!pairRes || pairRes.Error) {
      disposeCorrelationChart("corrPairChart");
      pairChart = null;
      return;
    }
    switch (pairChartMode) {
      case "scatter":
        pairChart = showCorrelationScatterChart("corrPairChart", pairRes);
        break;
      case "lag":
        pairChart = showCorrelationLagChart("corrPairChart", pairRes);
        break;
      case "matrix":
        pairChart = showCorrelationMatrixChart("corrPairChart", pairRes);
        break;
      default:
        pairChart = showCorrelationTimeChart("corrPairChart", pairRes, timeNormalized);
        break;
    }
  };

  const runPairAnalysis = async () => {
    if (!pollingID1 || !pollingID2) return;
    analyzing = true;
    try {
      const { startTime, endTime } = getTimeBounds(pairPeriodHours);
      pairRes = await AnalyzePollingCorrelation({
        PollingID1: pollingID1,
        PollingID2: pollingID2,
        Vars1: selectedVars1,
        Vars2: selectedVars2,
        IncludeState: pairIncludeState,
        Algorithm: pairAlgo,
        BucketSec: Number(pairBucketSec) || 0,
        StartTime: startTime,
        EndTime: endTime,
        MaxLag: Number(pairMaxLag) || 12,
      } as any);
      if (pairAlgo === "crosscorr" && !pairRes?.Error) {
        pairChartMode = "lag";
      }
      await drawPairChart();
    } finally {
      analyzing = false;
    }
  };

  const onVarsOrPairUpdated = async () => {
    if (pairRes && pollingID1 && pollingID2 && !analyzing) {
      await tick();
      await runPairAnalysis();
    }
  };

  const swapPairPollings = async () => {
    const tmpP = pollingID1;
    pollingID1 = pollingID2;
    pollingID2 = tmpP;
    const tmpN = nodeFilter1;
    nodeFilter1 = nodeFilter2;
    nodeFilter2 = tmpN;
    const tmpV = selectedVars1;
    selectedVars1 = selectedVars2;
    selectedVars2 = tmpV;
    await onVarsOrPairUpdated();
  };

  const toggleVar = (list: string[], key: string): string[] => {
    if (list.includes(key)) {
      return list.filter((k) => k !== key);
    }
    return [...list, key];
  };

  const selectVarAuto1 = async () => {
    selectedVars1 = [];
    await onVarsOrPairUpdated();
  };

  const selectVarKey1 = async (k: string) => {
    selectedVars1 = toggleVar(selectedVars1, k);
    await onVarsOrPairUpdated();
  };

  const selectVarAuto2 = async () => {
    selectedVars2 = [];
    await onVarsOrPairUpdated();
  };

  const selectVarKey2 = async (k: string) => {
    selectedVars2 = toggleVar(selectedVars2, k);
    await onVarsOrPairUpdated();
  };

  const getScoreTextClass = (score: number, algo: string) => {
    const abs = algo === "dcor" ? score : Math.abs(score);
    if (algo !== "dcor" && score <= -0.7) return "text-red-600 dark:text-red-400";
    if (abs >= 0.7) return "text-blue-600 dark:text-blue-400";
    if (abs >= 0.4) return "text-sky-600 dark:text-sky-400";
    if (abs >= 0.2) return "text-amber-600 dark:text-amber-400";
    return "text-gray-600 dark:text-gray-300";
  };

  const getStrengthMeta = (score: number, algo: string, t: any) => {
    const abs = algo === "dcor" ? score : Math.abs(score);
    if (abs >= 0.7) {
      if (algo === "dcor") {
        return {
          level: "strong",
          label: t("Correlation.StrengthStrongDCor") || "強い関連性",
          badgeClass: "bg-blue-100 text-blue-800 dark:bg-blue-900/60 dark:text-blue-200",
          bannerClass: "border-blue-300 bg-blue-50 text-blue-900 dark:border-blue-600 dark:bg-blue-900/30 dark:text-blue-100",
          icon: icons.mdiCheckCircle,
        };
      }
      if (score >= 0) {
        return {
          level: "strong",
          label: t("Correlation.StrengthStrongPos") || "強い正の相関",
          badgeClass: "bg-blue-100 text-blue-800 dark:bg-blue-900/60 dark:text-blue-200",
          bannerClass: "border-blue-300 bg-blue-50 text-blue-900 dark:border-blue-600 dark:bg-blue-900/30 dark:text-blue-100",
          icon: icons.mdiTrendingUp,
        };
      }
      return {
        level: "strong",
        label: t("Correlation.StrengthStrongNeg") || "強い負の相関",
        badgeClass: "bg-red-100 text-red-800 dark:bg-red-900/60 dark:text-red-200",
        bannerClass: "border-red-300 bg-red-50 text-red-900 dark:border-red-600 dark:bg-red-900/30 dark:text-red-100",
        icon: icons.mdiTrendingDown,
      };
    }
    if (abs >= 0.4) {
      return {
        level: "moderate",
        label: t("Correlation.StrengthModerate") || "中程度の相関",
        badgeClass: "bg-sky-100 text-sky-800 dark:bg-sky-900/60 dark:text-sky-200",
        bannerClass: "border-sky-300 bg-sky-50 text-sky-900 dark:border-sky-700 dark:bg-sky-900/25 dark:text-sky-100",
        icon: icons.mdiTrendingNeutral,
      };
    }
    if (abs >= 0.2) {
      return {
        level: "weak",
        label: t("Correlation.StrengthWeak") || "弱い相関",
        badgeClass: "bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-200",
        bannerClass: "border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-700/60 dark:bg-amber-900/20 dark:text-amber-100",
        icon: icons.mdiInformationOutline,
      };
    }
    return {
      level: "none",
      label: t("Correlation.StrengthNone") || "ほぼ無相関",
      badgeClass: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300",
      bannerClass: "border-gray-300 bg-gray-50 text-gray-800 dark:border-gray-700 dark:bg-gray-800/60 dark:text-gray-200",
      icon: icons.mdiMinusCircleOutline,
    };
  };

  const getAlgoDesc = (algo: string, t: any) => {
    switch (algo) {
      case "pearson":
        return t("Correlation.AlgoDescPearson") || "直線的な比例・逆比例の関係を評価";
      case "spearman":
        return t("Correlation.AlgoDescSpearman") || "順位ベースで外れ値や曲線的な増減に強い";
      case "kendall":
        return t("Correlation.AlgoDescKendall") || "順位の一致度を評価（同値・少数に安定）";
      case "crosscorr":
        return t("Correlation.AlgoDescCrossCorr") || "時間を前後にずらして最大の連動を検出";
      case "dcor":
        return t("Correlation.AlgoDescDCor") || "非線形関係や複数変数まとめての関連を検出";
      default:
        return "";
    }
  };

  const getPairSummary = (res: any, algo: string, t: any) => {
    const scores: any[] = res?.Scores || [];
    if (scores.length === 0) return null;
    const active = scores.find((s) => s.Algorithm === algo) || scores[0];
    const meta = getStrengthMeta(active.Score, active.Algorithm, t);
    let summaryText = "";
    if (meta.level === "strong") {
      if (active.Algorithm === "dcor") {
        summaryText = t("Correlation.SummaryStrongDCor") || "";
      } else if (active.Score >= 0) {
        summaryText = t("Correlation.SummaryStrongPos") || "";
      } else {
        summaryText = t("Correlation.SummaryStrongNeg") || "";
      }
    } else if (meta.level === "moderate") {
      if (active.Algorithm === "dcor") {
        summaryText = t("Correlation.SummaryModerateDCor") || "";
      } else if (active.Score >= 0) {
        summaryText = t("Correlation.SummaryModeratePos") || "";
      } else {
        summaryText = t("Correlation.SummaryModerateNeg") || "";
      }
    } else if (meta.level === "weak") {
      summaryText = t("Correlation.SummaryWeak") || "";
    } else {
      summaryText = t("Correlation.SummaryNone") || "";
    }

    let lagHint = "";
    const cross = scores.find((s) => s.Algorithm === "crosscorr");
    if (cross && Math.abs(cross.Score) >= 0.4 && cross.Lag !== 0) {
      if (cross.Lag > 0) {
        lagHint = t("Correlation.LagHintForward", {
          values: { sec: cross.LagSec, lag: cross.Lag },
        });
      } else {
        lagHint = t("Correlation.LagHintBackward", {
          values: { sec: Math.abs(cross.LagSec), lag: cross.Lag },
        });
      }
    }

    let pNote = "";
    if (meta.level === "none" && active.PValue < 0.05) {
      pNote = t("Correlation.SmallPValueNote", {
        values: { count: res.SampleCount },
      });
    }

    return {
      meta,
      active,
      summaryText,
      lagHint,
      pNote,
    };
  };

  $: pairSummary = pairRes && !pairRes.Error ? getPairSummary(pairRes, pairAlgo, $_) : null;

  const algoLabel = (algo: string) => {
    switch (algo) {
      case "pearson":
        return "Pearson (r)";
      case "spearman":
        return "Spearman (ρ)";
      case "kendall":
        return "Kendall (τ)";
      case "crosscorr":
        return "Cross-Corr";
      case "dcor":
        return "Distance (dCor)";
      default:
        return algo;
    }
  };

  const switchMainTab = async (next: "discover" | "pair") => {
    activeTab = next;
    await tick();
    if (next === "discover" && discoverRes) {
      await renderDiscoverResults();
    } else if (next === "pair" && pairRes) {
      await drawPairChart();
    }
  };

  const handleResize = () => {
    pairChart?.resize();
    networkChart?.resize();
  };

  const loadCandidates = async () => {
    loadingCandidates = true;
    try {
      candidates = (await GetCorrelationCandidates()) || [];
      if (!pollingID1 && candidates.length >= 1) {
        pollingID1 = candidates[0].PollingID;
      }
      if (!pollingID2 && candidates.length >= 2) {
        pollingID2 = candidates[1].PollingID;
      }
      if (!basePollingID && candidates.length >= 1) {
        basePollingID = candidates[0].PollingID;
      }
    } finally {
      loadingCandidates = false;
    }
  };

  onMount(async () => {
    await loadCandidates();
    if (initialMode === "pair" && initialPollingID1 && initialPollingID2) {
      pollingID1 = initialPollingID1;
      pollingID2 = initialPollingID2;
      activeTab = "pair";
      await tick();
      await runPairAnalysis();
    } else if (initialMode === "discover" && initialPollingID1) {
      discoverScope = "base";
      basePollingID = initialPollingID1;
      activeTab = "discover";
      await tick();
      await runDiscovery();
    }
  });

  onDestroy(() => {
    destroyDiscoverTable();
    disposeCorrelationChart("corrNetworkChart");
    disposeCorrelationChart("corrPairChart");
  });

  let prevLang = "";
  $: {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && activeTab === "discover" && discoverRes) {
      renderDiscoverResults();
    }
    prevLang = lang;
  }
</script>

<svelte:window onresize={handleResize} />

<div class="flex flex-col space-y-3 {modalMode ? '' : 'mx-2 mt-1'}">
  <!-- Mode Switcher -->
  <div class="flex flex-wrap items-center justify-between gap-2">
    <ButtonGroup>
      <Button
        size="xs"
        color={activeTab === "discover" ? "blue" : "alternative"}
        onclick={() => switchMainTab("discover")}
      >
        <Icon path={icons.mdiMagnifyScan} size={0.85} class="mr-1" />
        {$_("Correlation.TabDiscover") || "自動探索"}
      </Button>
      <Button
        size="xs"
        color={activeTab === "pair" ? "blue" : "alternative"}
        onclick={() => switchMainTab("pair")}
      >
        <Icon path={icons.mdiCompare} size={0.85} class="mr-1" />
        {$_("Correlation.TabPair") || "2ポーリング比較"}
      </Button>
    </ButtonGroup>

    <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
      {#if loadingCandidates}
        <Spinner size="4" />
      {/if}
      <span>{$_("Correlation.CandidateCount") || "対象ポーリング"}: {candidates.length}</span>
      <Button size="xs" color="alternative" onclick={loadCandidates}>
        <Icon path={icons.mdiRefresh} size={0.8} />
      </Button>
    </div>
  </div>

  {#if activeTab === "discover"}
    <!-- Auto Discovery Controls -->
    <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-3 space-y-3">
      <div class="grid grid-cols-2 md:grid-cols-6 gap-3 items-end">
        <Label class="space-y-1 text-xs">
          <span>{$_("Correlation.Scope") || "探索範囲"}</span>
          <Select
            items={scopeItems}
            bind:value={discoverScope}
            classes={{ select: "h-9 py-1 text-xs" }}
          />
        </Label>

        {#if discoverScope === "base"}
          <Label class="space-y-1 text-xs md:col-span-2">
            <span>{$_("Correlation.BasePolling") || "基準ポーリング"}</span>
            <Select
              items={pollingOptionsAll}
              bind:value={basePollingID}
              classes={{ select: "h-9 py-1 text-xs" }}
            />
          </Label>
        {/if}

        <Label class="space-y-1 text-xs">
          <span>{$_("Correlation.Algorithm") || "アルゴリズム"}</span>
          <Select
            items={algoItems}
            bind:value={discoverAlgo}
            classes={{ select: "h-9 py-1 text-xs" }}
          />
        </Label>

        <Label class="space-y-1 text-xs">
          <span>{$_("Correlation.Period") || "対象期間"}</span>
          <Select
            items={periodItems}
            bind:value={discoverPeriodHours}
            classes={{ select: "h-9 py-1 text-xs" }}
          />
        </Label>

        <Label class="space-y-1 text-xs">
          <span>{$_("Correlation.Bucket") || "時間バケット"}</span>
          <Select
            items={bucketItems}
            bind:value={discoverBucketSec}
            classes={{ select: "h-9 py-1 text-xs" }}
          />
        </Label>

        <Label class="space-y-1 text-xs">
          <span>{$_("Correlation.MinScore") || "最小 |相関係数|"}</span>
          <Select
            items={minScoreItems}
            bind:value={minAbsScore}
            classes={{ select: "h-9 py-1 text-xs" }}
          />
        </Label>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-4 text-xs">
          <Checkbox bind:checked={crossNodeOnly}>
            {$_("Correlation.CrossNodeOnly") || "異なるノード間のみ"}
          </Checkbox>
          <Checkbox bind:checked={discoverIncludeState}>
            {$_("Correlation.IncludeState") || "状態値 (state) を含める"}
          </Checkbox>
        </div>

        <GradientButton
          shadow
          color="blue"
          size="xs"
          type="button"
          disabled={discovering || candidates.length < 2}
          onclick={runDiscovery}
        >
          {#if discovering}
            <Spinner size="4" class="mr-1.5" />
          {:else}
            <Icon path={icons.mdiPlay} size={0.9} class="mr-1" />
          {/if}
          {$_("Correlation.RunDiscover") || "相関を探索"}
        </GradientButton>
      </div>
    </div>

    {#if discoverRes}
      {#if discoverRes.Error}
        <div class="p-2 rounded bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200 text-xs">
          {discoverRes.Error}
        </div>
      {/if}

      <div class="flex flex-wrap items-center justify-between gap-2 text-xs text-gray-600 dark:text-gray-300 px-1">
        <div class="flex flex-wrap items-center gap-3">
          <span>{$_("Correlation.EvaluatedPollings") || "対象ポーリング"}: <b>{discoverRes.TotalPollings}</b></span>
          <span>{$_("Correlation.EvaluatedPairs") || "評価ペア数"}: <b>{discoverRes.EvaluatedPairs}</b></span>
          <span>{$_("Correlation.HitCount") || "検出件数"}: <b>{discoverRes.Items?.length || 0}</b></span>
          <span>{$_("Correlation.BucketSec") || "バケット"}: <b>{discoverRes.BucketSec}s</b></span>
          <span>{$_("Correlation.Duration") || "所要時間"}: <b>{discoverRes.DurationMs}ms</b></span>
        </div>

        {#if selectedDiscoverItem}
          <GradientButton
            shadow
            color="green"
            size="xs"
            type="button"
            onclick={() => openPairFromItem(selectedDiscoverItem)}
          >
            <Icon path={icons.mdiChartLine} size={0.9} class="mr-1" />
            {$_("Correlation.OpenDetail") || "詳細分析を開く"}
          </GradientButton>
        {/if}
      </div>

      <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-2">
        <div>
          <table id="corrDiscoverTable" class="display compact" style="width:99%"></table>
        </div>
      </div>

      {#if (discoverRes.Items?.length || 0) > 0}
        <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-2">
          <div class="text-xs font-semibold text-gray-700 dark:text-gray-200 mb-1 px-1">
            {$_("Correlation.NetworkTitle") || "相関関係ネットワーク (エッジをクリックして詳細分析)"}
          </div>
          <div id="corrNetworkChart" class="w-full h-[340px]"></div>
        </div>
      {/if}
    {/if}
  {:else}
    <!-- Pair Analysis Controls -->
    <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-3 space-y-3">
      <div class="grid grid-cols-1 md:grid-cols-11 gap-2 items-start">
        <!-- Polling 1 -->
        <div class="md:col-span-5 space-y-2 p-2 rounded border border-gray-100 dark:border-gray-800">
          <div class="text-xs font-semibold text-blue-600 dark:text-blue-400">
            {$_("Correlation.PollingA") || "ポーリング A"}
          </div>
          <div class="grid grid-cols-2 gap-2">
            <Select
              items={nodeOptions}
              bind:value={nodeFilter1}
              onchange={() => {
                const list = buildPollingOptions(nodeFilter1);
                if (list.length > 0 && !list.some((o) => o.value === pollingID1)) {
                  pollingID1 = list[0].value;
                  selectedVars1 = [];
                }
              }}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
            <Select
              items={pollingOptions1}
              bind:value={pollingID1}
              onchange={() => { selectedVars1 = []; }}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </div>
          {#if availKeys1.length > 0}
            <div class="flex flex-wrap items-center gap-1.5 text-[11px]">
              <span class="text-gray-500 dark:text-gray-400">{$_("Correlation.Vars") || "変数"}:</span>
              <button
                type="button"
                onclick={selectVarAuto1}
                class="px-2 py-0.5 rounded border text-[11px] {selectedVars1.length === 0
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300'}"
              >
                {$_("Correlation.VarAuto") || "自動"}
              </button>
              {#each availKeys1 as k}
                <button
                  type="button"
                  onclick={() => selectVarKey1(k)}
                  class="px-2 py-0.5 rounded border text-[11px] {selectedVars1.includes(k)
                    ? 'bg-blue-600 text-white border-blue-600'
                    : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300'}"
                >
                  {k}
                </button>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Swap button -->
        <div class="md:col-span-1 flex items-center justify-center md:pt-8">
          <Button size="xs" color="alternative" onclick={swapPairPollings} title={$_("Correlation.Swap") || "入替"}>
            <Icon path={icons.mdiSwapHorizontal} size={1} />
          </Button>
        </div>

        <!-- Polling 2 -->
        <div class="md:col-span-5 space-y-2 p-2 rounded border border-gray-100 dark:border-gray-800">
          <div class="text-xs font-semibold text-emerald-600 dark:text-emerald-400">
            {$_("Correlation.PollingB") || "ポーリング B"}
          </div>
          <div class="grid grid-cols-2 gap-2">
            <Select
              items={nodeOptions}
              bind:value={nodeFilter2}
              onchange={() => {
                const list = buildPollingOptions(nodeFilter2);
                if (list.length > 0 && !list.some((o) => o.value === pollingID2)) {
                  pollingID2 = list[0].value;
                  selectedVars2 = [];
                }
              }}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
            <Select
              items={pollingOptions2}
              bind:value={pollingID2}
              onchange={() => { selectedVars2 = []; }}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </div>
          {#if availKeys2.length > 0}
            <div class="flex flex-wrap items-center gap-1.5 text-[11px]">
              <span class="text-gray-500 dark:text-gray-400">{$_("Correlation.Vars") || "変数"}:</span>
              <button
                type="button"
                onclick={selectVarAuto2}
                class="px-2 py-0.5 rounded border text-[11px] {selectedVars2.length === 0
                  ? 'bg-emerald-600 text-white border-emerald-600'
                  : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300'}"
              >
                {$_("Correlation.VarAuto") || "自動"}
              </button>
              {#each availKeys2 as k}
                <button
                  type="button"
                  onclick={() => selectVarKey2(k)}
                  class="px-2 py-0.5 rounded border text-[11px] {selectedVars2.includes(k)
                    ? 'bg-emerald-600 text-white border-emerald-600'
                    : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300'}"
                >
                  {k}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="flex flex-wrap items-end justify-between gap-3">
        <div class="grid grid-cols-2 md:grid-cols-5 gap-2 items-end flex-1">
          <Label class="space-y-1 text-xs">
            <span>{$_("Correlation.Algorithm") || "優先アルゴリズム"}</span>
            <Select
              items={algoItems}
              bind:value={pairAlgo}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </Label>
          <Label class="space-y-1 text-xs">
            <span>{$_("Correlation.Period") || "対象期間"}</span>
            <Select
              items={periodItems}
              bind:value={pairPeriodHours}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </Label>
          <Label class="space-y-1 text-xs">
            <span>{$_("Correlation.Bucket") || "時間バケット"}</span>
            <Select
              items={bucketItems}
              bind:value={pairBucketSec}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </Label>
          <Label class="space-y-1 text-xs">
            <span>{$_("Correlation.MaxLag") || "最大ラグ (バケット)"}</span>
            <Select
              items={[
                { value: 6, name: "±6" },
                { value: 12, name: "±12" },
                { value: 24, name: "±24" },
                { value: 48, name: "±48" },
              ]}
              bind:value={pairMaxLag}
              classes={{ select: "h-8 py-1 text-xs" }}
            />
          </Label>
          <div class="pb-1.5">
            <Checkbox bind:checked={pairIncludeState} class="text-xs">
              {$_("Correlation.IncludeState") || "状態値を含む"}
            </Checkbox>
          </div>
        </div>

        <GradientButton
          shadow
          color="blue"
          size="xs"
          type="button"
          disabled={analyzing || !pollingID1 || !pollingID2}
          onclick={runPairAnalysis}
        >
          {#if analyzing}
            <Spinner size="4" class="mr-1.5" />
          {:else}
            <Icon path={icons.mdiChartScatterPlot} size={0.9} class="mr-1" />
          {/if}
          {$_("Correlation.RunAnalyze") || "相関を分析"}
        </GradientButton>
      </div>
    </div>

    {#if pairRes}
      {#if pairRes.Error}
        <div class="p-2.5 rounded bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200 text-xs">
          {pairRes.Error}
        </div>
      {:else}
        <!-- Summary Meta -->
        <div class="flex flex-wrap items-center gap-3 text-xs text-gray-600 dark:text-gray-300 px-1">
          <Badge color="blue">{pairRes.NodeName1} / {pairRes.PollingName1} ({pairRes.SelectedVar1})</Badge>
          <span>↔</span>
          <Badge color="green">{pairRes.NodeName2} / {pairRes.PollingName2} ({pairRes.SelectedVar2})</Badge>
          <span>{$_("Correlation.Samples") || "サンプル数"}: <b>{pairRes.SampleCount}</b></span>
          <span>{$_("Correlation.BucketSec") || "バケット"}: <b>{pairRes.BucketSec}s</b></span>
          <span>y = {Number(pairRes.Slope).toFixed(4)}x + {Number(pairRes.Intercept).toFixed(4)}</span>
        </div>

        <!-- Interpretation Summary Banner -->
        {#if pairSummary}
          <div class="rounded-lg border p-2.5 text-xs flex flex-col gap-1 {pairSummary.meta.bannerClass}">
            <div class="flex flex-wrap items-center gap-2">
              <Icon path={pairSummary.meta.icon} size={0.95} />
              <span class="font-bold">
                {algoLabel(pairSummary.active.Algorithm)}: {Number(pairSummary.active.Score).toFixed(4)}
              </span>
              <span class="px-2 py-0.5 rounded text-[11px] font-semibold {pairSummary.meta.badgeClass}">
                {pairSummary.meta.label}
              </span>
              <span>{pairSummary.summaryText}</span>
            </div>
            {#if pairSummary.lagHint}
              <div class="text-[11px] pl-5 opacity-90">{pairSummary.lagHint}</div>
            {/if}
            {#if pairSummary.pNote}
              <div class="text-[11px] pl-5 opacity-80">{pairSummary.pNote}</div>
            {/if}
          </div>
        {/if}

        <!-- 5 Algorithm Score Cards -->
        <div class="grid grid-cols-2 md:grid-cols-5 gap-2">
          {#each pairRes.Scores || [] as sc}
            {@const meta = getStrengthMeta(sc.Score, sc.Algorithm, $_)}
            <button
              type="button"
              onclick={() => {
                pairAlgo = sc.Algorithm;
                if (sc.Algorithm === "crosscorr") {
                  pairChartMode = "lag";
                  drawPairChart();
                }
              }}
              class="text-left rounded-lg border p-2.5 bg-white dark:bg-gray-900 transition {pairAlgo === sc.Algorithm
                ? 'border-blue-500 ring-1 ring-blue-500'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-400'}"
            >
              <div class="flex items-center justify-between gap-1">
                <span class="text-[11px] text-gray-500 dark:text-gray-400 font-medium">
                  {algoLabel(sc.Algorithm)}
                </span>
                <span class="px-1.5 py-0.5 rounded text-[10px] font-semibold {meta.badgeClass}">
                  {meta.label}
                </span>
              </div>
              <div class="text-lg font-bold mt-0.5 {getScoreTextClass(sc.Score, sc.Algorithm)}">
                {Number(sc.Score).toFixed(4)}
              </div>
              <div class="text-[10px] text-gray-500 dark:text-gray-400">
                p={formatPValue(sc.PValue)}
                {#if sc.Algorithm === "crosscorr"}
                  | lag={sc.Lag > 0 ? `+${sc.Lag}` : sc.Lag} ({sc.LagSec > 0 ? `+${sc.LagSec}` : sc.LagSec}s)
                {/if}
              </div>
              <div class="text-[10px] text-gray-400 dark:text-gray-500 mt-1 leading-tight">
                {getAlgoDesc(sc.Algorithm, $_)}
              </div>
            </button>
          {/each}
        </div>

        <!-- Chart Switcher & View -->
        <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-2 space-y-2">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <ButtonGroup>
              <Button
                size="xs"
                color={pairChartMode === "time" ? "blue" : "alternative"}
                onclick={() => { pairChartMode = "time"; drawPairChart(); }}
              >
                {$_("Correlation.ChartTime") || "時系列"}
              </Button>
              <Button
                size="xs"
                color={pairChartMode === "scatter" ? "blue" : "alternative"}
                onclick={() => { pairChartMode = "scatter"; drawPairChart(); }}
              >
                {$_("Correlation.ChartScatter") || "散布図"}
              </Button>
              <Button
                size="xs"
                color={pairChartMode === "lag" ? "blue" : "alternative"}
                onclick={() => { pairChartMode = "lag"; drawPairChart(); }}
              >
                {$_("Correlation.ChartLag") || "ラグ相関"}
              </Button>
              <Button
                size="xs"
                color={pairChartMode === "matrix" ? "blue" : "alternative"}
                onclick={() => { pairChartMode = "matrix"; drawPairChart(); }}
              >
                {$_("Correlation.ChartMatrix") || "変数行列"}
              </Button>
            </ButtonGroup>

            {#if pairChartMode === "time"}
              <Checkbox
                bind:checked={timeNormalized}
                onchange={() => drawPairChart()}
                class="text-xs"
              >
                {$_("Correlation.NormalizeOverlay") || "0-1 正規化して重ね合わせ"}
              </Checkbox>
            {/if}
          </div>

          <div id="corrPairChart" class="w-full {modalMode ? 'h-[340px]' : 'h-[380px]'}"></div>
        </div>
      {/if}
    {/if}
  {/if}
</div>
