<script lang="ts">
  import ping_ok from "../assets/sound/ping_ok.mp3";
  import ping_ng from "../assets/sound/ping_ng.mp3";
  import { Modal, GradientButton, Tabs, TabItem, Input, Select, Toggle, ButtonGroup, Button } from "flowbite-svelte";
  import { tick } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { createEventDispatcher } from "svelte";
  import {
    getPingChartOption,
    showPing3DChart,
    showPingHistgram,
    showPingLinearChart,
    showPingMapChart,
    showPingSmokeChart,
  } from "./chart/ping";
  import { showMtrProfileChart, type MTRHopStat } from "./chart/mtr";
  import { GetNode, Ping, GetMapConf, LLMExplainPingReport } from "../../wailsjs/go/main/App";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import {
    getStateIcon,
    getStateColor,
    getTableLang,
    formatTime,
  } from "./common";
  import * as echarts from "echarts";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let nodeID = "";

  type PingMode = "normal" | "smoke" | "trace" | "mtr";
  let mode: PingMode = "normal";

  let pingTab = true;
  let wait = false;
  let hasAI = false;
  let showAIReport = false;
  let table :any = undefined;
  let chart :any = undefined;
  let chartOption :any = undefined;
  let results :any = [];
  let ip = "";
  let ipColor: any = undefined;
  let size = 64;
  let count = 10;
  let ttl = 64;
  const pingReq = {
    size: 0,
    count: 0,
    ttl: 64,
  };
  let timer :any = undefined;
  let startTime = 0;
  let canShowLinear = false;
  let canShowWorld = false;
  let canShowHistogram = false;
  let canShowSmoke = false;
  let canShowMtr = false;
  let mtrHops: (MTRHopStat & { history: number[]; lossCount: number })[] = [];
  let mtrTable: any = undefined;
  let beep = false;
  let sound_ok :any;
  let sound_ng :any;
  let showHelp = false;

  const changeMode = (m: PingMode) => {
    mode = m;
    if (mode === "normal") {
      ttl = 64;
      count = 10;
      size = 64;
    } else if (mode === "smoke") {
      ttl = 64;
      count = 2001;
      size = 64;
    } else if (mode === "trace") {
      ttl = -1;
      count = -1;
      size = 64;
    } else if (mode === "mtr") {
      ttl = -2;
      count = 2001;
      size = 64;
    }
  };

  const dispatch = createEventDispatcher();

  const resetState = () => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    stopFlag = true;
    wait = false;
    results = [];
    mtrHops = [];
    canShowLinear = false;
    canShowWorld = false;
    canShowHistogram = false;
    canShowSmoke = false;
    canShowMtr = false;
    pingTab = true;
    mode = "normal";
    ttl = 64;
    count = 10;
    size = 64;
  };

  const onOpen = async () => {
    resetState();
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    const node = await GetNode(nodeID);
    if (node && node.IP) {
      ip = node.IP;
    }
    showPing();
  };

  const showTable = () => {
    table = new DataTable("#pingTable", {
      destroy: true,
      columns: columns,
      paging: false,
      searching:false,
      info:false,
      scrollY: "25vh",
      scrollCollapse: true,
      data: results,
      order: [[1, "asc"]],
      language: getTableLang(),
    });
  };

  const showPing = async () => {
    activeTab = "ping";
    await tick();
    if (chart) {
      chart.dispose();
    }
    chartOption = getPingChartOption();
    chart = echarts.init(document.getElementById("pingChart"),"dark");
    if (results.length > 0) {
      for (const r of results) {
        if (r.Stat === 1 || r.Stat === 4) {
          const t = new Date(r.TimeStamp * 1000);
          const ts = echarts.time.format(
            t,
            "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}",
            false
          );
          chartOption.series[0].data.push({
            ts,
            value: [t, r.Time / (1000 * 1000 * 1000)],
          });
          chartOption.series[1].data.push({
            ts,
            value: [t, r.SendTTL],
          });
          chartOption.series[2].data.push({
            ts,
            value: [t, r.RecvTTL],
          });
        }
      }
    }
    showTable();
    chart.setOption(chartOption);
    chart.resize();
  };

  const normalCountList = [
    { name: $_('Ping.Count10'), value: 10 },
    { name: $_('Ping.Coun1'), value: 1 },
    { name: $_('Ping.Count3'), value: 3 },
    { name: $_('Ping.Count5'), value: 5 },
    { name: $_('Ping.Count20'), value: 20 },
    { name: $_('Ping.Count30'), value: 30 },
    { name: $_('Ping.Count50'), value: 50 },
    { name: $_('Ping.Count100'), value: 100 },
    { name: $_('Ping.Cont'), value: -1 },
  ];

  const smokeDurationList = [
    { name: $_('Ping.Duration1m') || "1分間", value: 2001 },
    { name: $_('Ping.Duration3m') || "3分間", value: 2003 },
    { name: $_('Ping.Duration5m') || "5分間", value: 2005 },
    { name: $_('Ping.Duration10m') || "10分間", value: 2010 },
    { name: $_('Ping.DurationCont') || "無制限 (連続)", value: -100 },
  ];

  const mtrDurationList = [
    { name: $_('Ping.Duration1m') || "1分間", value: 2001 },
    { name: $_('Ping.Duration3m') || "3分間", value: 2003 },
    { name: $_('Ping.Duration5m') || "5分間", value: 2005 },
    { name: $_('Ping.Duration10m') || "10分間", value: 2010 },
    { name: $_('Ping.DurationCont') || "無制限 (連続)", value: -1 },
  ];

  const sizeList = [
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "256", value: 256 },
    { name: "512", value: 512 },
    { name: "1024", value: 1024 },
    { name: "1500", value: 1500 },
    { name: $_('Ping.IncSize'), value: -1 },
  ];

  const fixedSizeList = [
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "256", value: 256 },
    { name: "512", value: 512 },
    { name: "1024", value: 1024 },
    { name: "1500", value: 1500 },
  ];

  const ttlList = [
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "254", value: 254 },
    { name: "1", value: 1 },
    { name: "2", value: 2 },
    { name: "4", value: 4 },
    { name: "8", value: 8 },
    { name: "16", value: 16 },
    { name: "32", value: 32 },
  ];

  const renderMtrLossRate = (val: any, type: string) => {
    if (type === "sort") return val;
    const loss = Number(val || 0);
    let colorClass = "text-emerald-400";
    let bgClass = "bg-emerald-500";
    if (loss > 20) {
      colorClass = "text-red-400 font-bold";
      bgClass = "bg-red-500";
    } else if (loss > 0) {
      colorClass = "text-amber-400";
      bgClass = "bg-amber-500";
    }
    return `<div class="flex items-center gap-2"><div class="w-12 bg-gray-700 h-2 rounded overflow-hidden"><div class="${bgClass} h-full" style="width:${Math.min(100, loss)}%"></div></div><span class="${colorClass}">${loss.toFixed(1)}%</span></div>`;
  };

  const renderMtrMs = (val: any, type: string) => {
    if (type === "sort") return val;
    if (val === undefined || val === null || val < 0) return "-";
    return val.toFixed(2);
  };

  const mtrColumns = [
    { data: "ttl", title: $_('Ping.Hop') || "Hop", width: "8%" },
    { data: "ip", title: $_('Ping.Host') || "Host / IP", width: "22%" },
    { data: "loc", title: $_('Ping.Loc') || "Location", width: "18%" },
    { data: "lossRate", title: $_('Ping.Loss') || "Loss %", width: "16%", render: renderMtrLossRate },
    { data: "snt", title: $_('Ping.Snt') || "Snt", width: "8%" },
    { data: "last", title: $_('Ping.Last') || "Last", width: "8%", render: renderMtrMs },
    { data: "avg", title: $_('Ping.Avg') || "Avg", width: "8%", render: renderMtrMs },
    { data: "best", title: $_('Ping.Best') || "Best", width: "8%", render: renderMtrMs },
    { data: "wrst", title: $_('Ping.Wrst') || "Wrst", width: "8%", render: renderMtrMs },
    { data: "stDev", title: $_('Ping.StDev') || "StDev", width: "8%", render: renderMtrMs },
  ];

  const showMtrTable = () => {
    mtrTable = new DataTable("#mtrTable", {
      destroy: true,
      columns: mtrColumns,
      paging: false,
      searching: false,
      info: false,
      scrollY: "90px",
      scrollCollapse: true,
      data: mtrHops,
      order: [[0, "asc"]],
      language: getTableLang(),
    });
  };

  const renderPingStat = (s:any, type:string) => {
    if (type == "sort") {
      return s;
    }
    let state = "unknown";
    let name = $_('Ping.Unknown');
    switch (s) {
      case 1:
        state = "normal";
        name = $_('Ping.Normal');
        break;
      case 2:
        state = "error";
        name = "Timeout";
        break;
      case 3:
        state = "warn";
        name = $_('Ping.Warn');
        break;
      case 4:
        state = "info";
        name = "GW";
        break;
    }
    return (
      `<span class="mdi ` +
      getStateIcon(state) +
      ` text-xs" style="color:` +
      getStateColor(state) +
      `;"></span><span class="ml-2">` +
      name +
      `</span>`
    );
  };

  const renderTimeStamp = (ts:any) => {
    return formatTime(new Date(ts * 1000), "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
  };

  const renderRespTime = (t:any) => {
    return (t / (1000 * 1000 * 1000)).toFixed(6);
  };

  const columns = [
    {
      data: "Stat",
      title: $_('Ping.Result'),
      width: "10%",
      render: renderPingStat,
    },
    {
      data: "TimeStamp",
      title: $_('Ping.TimeStamp'),
      width: "15%",
      render: renderTimeStamp,
    },
    {
      data: "Time",
      title: $_('Ping.RespTime'),
      width: "10%",
      render: renderRespTime,
    },
    {
      data: "Size",
      title: $_('Ping.Size'),
      width: "10%",
    },
    {
      data: "SendTTL",
      title: $_('Ping.SendTTL'),
      width: "10%",
    },
    {
      data: "RecvTTL",
      title: $_('Ping.RecvTTL'),
      width: "10%",
    },
    {
      data: "RecvSrc",
      title: $_('Ping.RecvSrc'),
      width: "15%",
    },
    {
      data: "Loc",
      title: $_('Ping.Loc'),
      width: "20%",
    },
  ];

  let reportChart :any  = undefined;
  let activeTab = "ping";

  const showHistogram = async () => {
    activeTab = "histogram";
    pingTab = false;
    await tick();
    reportChart = showPingHistgram("histogram", results);
  };

  const show3D = async () => {
    activeTab = "3d";
    pingTab = false;
    await tick();
    reportChart = showPing3DChart("chart3d", results);
  };

  const showLinear = async () => {
    activeTab = "linear";
    pingTab = false;
    await tick();
    reportChart = showPingLinearChart("linear", results);
  };

  const showWorld = async () => {
    activeTab = "world";
    pingTab = false;
    await tick();
    reportChart = showPingMapChart("world", results);
  };

  const showSmoke = async () => {
    activeTab = "smoke";
    pingTab = false;
    await tick();
    reportChart = showPingSmokeChart("smoke", results);
  };

  const showMTR = async () => {
    activeTab = "mtr";
    pingTab = false;
    await tick();
    showMtrTable();
    reportChart = showMtrProfileChart("mtrChart", mtrHops);
  };

  const updatePingChart = (r: any) => {
    if (chart && (r.Stat === 1 || r.Stat === 4)) {
      const t = new Date(r.TimeStamp * 1000);
      const ts = echarts.time.format(
        t,
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}",
        false
      );
      chartOption.series[0].data.push({
        ts,
        value: [t, r.Time / (1000 * 1000 * 1000)],
      });
      chartOption.series[1].data.push({
        ts,
        value: [t, r.SendTTL],
      });
      chartOption.series[2].data.push({
        ts,
        value: [t, r.RecvTTL],
      });
      chart.setOption(chartOption);
      chart.resize();
      if (beep && sound_ok) {
        sound_ok.play();
      }
      if (r.Loc && !r.Loc.startsWith("LOCAL")) {
        canShowWorld = true;
      }
    } else {
      if (beep && sound_ng) {
        sound_ng.play();
      }
    }
  };

  const start = () => {
    if (!ip) {
      ipColor ="red";
      return;
    } else {
      ipColor = undefined;
    }
    stopFlag = false;
    startTime = Date.now();
    if (chart) {
      chartOption.series[0].data = [];
      chartOption.series[1].data = [];
      chartOption.series[2].data = [];
    }
    wait = true;
    pingReq.count = 0;
    pingReq.size = size < 0 ? 0 : size;
    results = [];
    canShowWorld = false;
    canShowSmoke = false;

    if (mode === "trace") {
      ttl = -1;
      count = -1;
    } else if (mode === "mtr") {
      ttl = -2;
    } else if (mode === "normal") {
      if (ttl < 0) ttl = 64;
    } else if (mode === "smoke") {
      if (ttl < 0) ttl = 64;
    }

    if (ttl === -2) {
      canShowMtr = false;
      mtrHops = [];
      _doMtrProcess();
    } else {
      if (ttl === -1) {
        pingReq.ttl = 1;
        count = -1;
        if (size < 0) size = 64;
      } else {
        pingReq.ttl = ttl;
      }
      _doPing();
    }
  };

  let stopFlag = true;

  const stop = () => {
    stopFlag = true;
  };

  let mtrPhase: "discovery" | "sampling" = "discovery";
  let mtrDiscoveryTTL = 1;
  let mtrSampleRound = 0;

  const _doMtrProcess = async () => {
    if (stopFlag) {
      wait = false;
      if (mtrHops.length > 0) canShowMtr = true;
      return;
    }

    if (mtrHops.length === 0) {
      mtrPhase = "discovery";
      mtrDiscoveryTTL = 1;
    }

    if (mtrPhase === "discovery") {
      const r = await Ping({
        IP: ip,
        Size: pingReq.size > 0 ? pingReq.size : 64,
        TTL: mtrDiscoveryTTL,
      });
      results.push(r);
      showTable();
      updatePingChart(r);

      const recvIp = r.RecvSrc || "";
      const loc = r.Loc || "";
      const isTarget = r.Stat === 1 || recvIp === ip;

      mtrHops.push({
        ttl: mtrDiscoveryTTL,
        ip: recvIp,
        loc: loc,
        snt: 0,
        lossRate: 0,
        last: -1,
        avg: -1,
        best: -1,
        wrst: -1,
        stDev: -1,
        isTarget: isTarget,
        history: [],
        lossCount: 0,
      });
      mtrHops = [...mtrHops];

      if (isTarget || mtrDiscoveryTTL >= 32) {
        mtrPhase = "sampling";
        mtrSampleRound = 0;
      } else {
        mtrDiscoveryTTL++;
      }

      if (!stopFlag) {
        timer = setTimeout(() => _doMtrProcess(), 100);
      } else {
        wait = false;
        canShowMtr = true;
      }
      return;
    }

    // Sampling Phase (Continuous MTR Ping)
    for (let i = 0; i < mtrHops.length; i++) {
      if (stopFlag) break;
      const hop = mtrHops[i];
      const r = await Ping({
        IP: ip,
        Size: pingReq.size > 0 ? pingReq.size : 64,
        TTL: hop.ttl,
      });
      results.push(r);
      pingReq.count++;
      showTable();
      updatePingChart(r);

      hop.snt++;
      if (r.Stat === 1 || r.Stat === 4) {
        const rttMs = r.Time / (1000 * 1000);
        hop.last = rttMs;
        hop.history.push(rttMs);
        if (hop.best < 0 || rttMs < hop.best) hop.best = rttMs;
        if (rttMs > hop.wrst) hop.wrst = rttMs;

        const sum = hop.history.reduce((a, b) => a + b, 0);
        hop.avg = sum / hop.history.length;

        if (hop.history.length > 1) {
          const mean = hop.avg;
          const variance = hop.history.reduce((acc, val) => acc + Math.pow(val - mean, 2), 0) / hop.history.length;
          hop.stDev = Math.sqrt(variance);
        } else {
          hop.stDev = 0;
        }
      } else {
        hop.lossCount++;
      }
      hop.lossRate = (hop.lossCount / hop.snt) * 100;
      if (r.RecvSrc && !hop.ip) {
        hop.ip = r.RecvSrc;
      }
      if (r.Loc && !hop.loc) {
        hop.loc = r.Loc;
      }
    }

    mtrHops = [...mtrHops];
    mtrSampleRound++;
    if (activeTab === "mtr") {
      showMTR();
    }

    let elapsedSec = (Date.now() - startTime) / 1000;
    let maxDurationSec = 0;
    if (count === 2001) maxDurationSec = 60;
    if (count === 2003) maxDurationSec = 180;
    if (count === 2005) maxDurationSec = 300;

    let isTimeOver = maxDurationSec > 0 && elapsedSec >= maxDurationSec;
    let keepGoing = false;
    if (!stopFlag && !isTimeOver) {
      if (count === -1 || count === -100) {
        keepGoing = true;
      } else if (maxDurationSec > 0) {
        keepGoing = true;
      } else if (mtrSampleRound < count) {
        keepGoing = true;
      }
    }

    if (keepGoing) {
      timer = setTimeout(() => _doMtrProcess(), 500);
    } else {
      wait = false;
      canShowMtr = true;
    }
  };

  const _doPing = async () => {
    const r = await Ping({
      IP: ip,
      Size: pingReq.size,
      TTL: pingReq.ttl,
    });
    pingReq.count++;
    results.push(r);
    showTable();
    if (activeTab === "smoke") {
      showSmoke();
    }
    updatePingChart(r);
    let isSmokeMode = count >= 2001 || count === -100;
    let maxDurationSec = 0;
    if (count === 2001) maxDurationSec = 60;
    if (count === 2003) maxDurationSec = 180;
    if (count === 2005) maxDurationSec = 300;
    if (count === 2010) maxDurationSec = 600;

    let elapsedSec = (Date.now() - startTime) / 1000;
    let isTimeOver = maxDurationSec > 0 && elapsedSec >= maxDurationSec;
    let interval = isSmokeMode ? 200 : (beep ? 2000 : 1000);

    let keepGoing = false;
    if (!stopFlag && !isTimeOver) {
      if (count === -1 || count === -100) {
        keepGoing = true;
      } else if (maxDurationSec > 0) {
        keepGoing = true;
      } else if (pingReq.count < count) {
        keepGoing = true;
      }
    }

    if (keepGoing) {
      if (size === -1) {
        if (r.Stat !== 1) {
          pingReq.size = 0;
        }
        // サイズを変更するモード
        pingReq.size += 100;
      }
      if (ttl === -1) {
        pingReq.ttl++;
        if (r.Stat === 1 || pingReq.ttl > 254) {
          wait = false;
          canShowHistogram = false;
          canShowLinear = false;
          canShowSmoke = false;
          return;
        }
      }
      timer = setTimeout(() => _doPing(), interval);
    } else {
      wait = false;
      canShowLinear = size == -1;
      canShowHistogram = !canShowLinear;
      canShowSmoke = true;
    }
  };

  const close = () => {
    resetState();
    show = false;
    dispatch("close", {});
  };
  const resizeChart = () => {
    if(reportChart) {
      reportChart.resize();
    }
    if (chart) {
      chart.resize();
    }
  }


  $: if (show) {
    onOpen();
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal bind:open={show} size="xl" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <Tabs style="underline" contentClass="pt-2 bg-transparent">
      <TabItem bind:open={pingTab} onclick={showPing}>
        {#snippet titleSlot()}
        <div class="flex items-center gap-2">
          <Icon path={icons.mdiCheckNetwork} size={1} />
          { $_('Ping.DoPing') }
        </div>
      {/snippet}
        <div class="flex flex-col space-y-2 mb-2 bg-gray-800/40 p-2.5 rounded-lg border border-gray-700/60">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-gray-300">
                {$_('Ping.Mode') || "動作モード"}:
              </span>
              <ButtonGroup>
                <Button
                  size="xs"
                  class="!py-1 !px-2 text-[11px]"
                  color={mode === 'normal' ? 'blue' : 'alternative'}
                  onclick={() => changeMode('normal')}
                >
                  <Icon path={icons.mdiCheckNetwork} size={0.7} class="mr-1 inline" />
                  {$_('Ping.ModeNormal') || "通常 Ping"}
                </Button>
                <Button
                  size="xs"
                  class="!py-1 !px-2 text-[11px]"
                  color={mode === 'smoke' ? 'blue' : 'alternative'}
                  onclick={() => changeMode('smoke')}
                >
                  <Icon path={icons.mdiWeatherCloudy} size={0.7} class="mr-1 inline" />
                  {$_('Ping.ModeSmoke') || "Smoke"}
                </Button>
                <Button
                  size="xs"
                  class="!py-1 !px-2 text-[11px]"
                  color={mode === 'trace' ? 'blue' : 'alternative'}
                  onclick={() => changeMode('trace')}
                >
                  <Icon path={icons.mdiSourceBranch} size={0.7} class="mr-1 inline" />
                  {$_('Ping.ModeTrace') || "トレースルート"}
                </Button>
                <Button
                  size="xs"
                  class="!py-1 !px-2 text-[11px]"
                  color={mode === 'mtr' ? 'blue' : 'alternative'}
                  onclick={() => changeMode('mtr')}
                >
                  <Icon path={icons.mdiRoutes} size={0.7} class="mr-1 inline" />
                  {$_('Ping.ModeMTR') || "MTR"}
                </Button>
              </ButtonGroup>
            </div>
          </div>

          <div class="flex flex-row items-center gap-2">
            <Input
              class="h-[38px] min-w-[140px] flex-1"
              type="text"
              bind:value={ip}
              placeholder={ $_('Ping.IPOrHost') }
              color={ipColor}
              size="sm"
            />

            {#if mode === 'normal'}
              <Select
                class="h-[38px] w-28"
                items={normalCountList}
                bind:value={count}
                placeholder={ $_('Ping.Count') }
                size="sm"
              />
              <Select
                class="h-[38px] w-28"
                items={sizeList}
                bind:value={size}
                placeholder={ $_('Ping.Size') }
                size="sm"
              />
              <Select
                class="h-[38px] w-24"
                items={ttlList}
                bind:value={ttl}
                placeholder="TTL"
                size="sm"
              />
            {:else if mode === 'smoke'}
              <Select
                class="h-[38px] w-40"
                items={smokeDurationList}
                bind:value={count}
                placeholder={ $_('Ping.Duration') || "測定期間" }
                size="sm"
              />
              <Select
                class="h-[38px] w-28"
                items={fixedSizeList}
                bind:value={size}
                placeholder={ $_('Ping.Size') }
                size="sm"
              />
            {:else if mode === 'trace'}
              <Select
                class="h-[38px] w-28"
                items={fixedSizeList}
                bind:value={size}
                placeholder={ $_('Ping.Size') }
                size="sm"
              />
            {:else if mode === 'mtr'}
              <Select
                class="h-[38px] w-40"
                items={mtrDurationList}
                bind:value={count}
                placeholder={ $_('Ping.Duration') || "測定期間" }
                size="sm"
              />
              <Select
                class="h-[38px] w-28"
                items={fixedSizeList}
                bind:value={size}
                placeholder={ $_('Ping.Size') }
                size="sm"
              />
            {/if}
          </div>
        </div>
        <div id="pingChart" class="mb-2"></div>
        <div><table id="pingTable" class="display compact" style="width:99%"></table></div>
      </TabItem>
      {#if canShowMtr}
        <TabItem onclick={showMTR}>
          {#snippet titleSlot()}
            <div class="flex items-center gap-2">
              <Icon path={icons.mdiRoutes} size={1} />
              { $_('Ping.MTR') || "MTR" }
            </div>
          {/snippet}
          <div class="flex flex-col space-y-2.5 max-h-[60vh] overflow-y-auto pt-0 pb-1 pr-1">
            <!-- ① 経路トポロジー・フローマップ (Hop Flow Diagram) -->
            <div class="flex-shrink-0 bg-gray-800/80 p-2.5 rounded-lg border border-gray-700 overflow-x-auto">
              <div class="text-xs font-bold text-gray-300 mb-1.5 flex items-center gap-2">
                <Icon path={icons.mdiSourceBranch} size={0.8} />
                {$_('Ping.MTRProfile') || "Route Hop Flow"}
              </div>
              <div class="flex items-center gap-3 min-w-max pb-1">
                {#each mtrHops as hop, idx}
                  <div class={`flex flex-col p-2 rounded-lg border text-xs min-w-[135px] shadow-sm transition-all ${
                    hop.lossRate > 20
                      ? 'bg-red-950/60 border-red-500/80 text-red-200'
                      : hop.lossRate > 0
                      ? 'bg-amber-950/60 border-amber-500/80 text-amber-200'
                      : 'bg-gray-900/90 border-gray-700 text-gray-200'
                  }`}>
                    <div class="flex justify-between items-center mb-1 font-mono font-bold">
                      <span class="px-1.5 py-0.5 rounded bg-gray-800 border border-gray-600 text-[10px]">
                        Hop {hop.ttl}
                      </span>
                      {#if hop.isTarget}
                        <span class="px-1.5 py-0.5 rounded bg-blue-600 text-white text-[10px]">
                          Target
                        </span>
                      {/if}
                    </div>
                    <div class="font-bold truncate text-sm" title={hop.ip || 'No Response'}>
                      {hop.ip || '* No Response *'}
                    </div>
                    {#if hop.loc}
                      <div class="text-[10px] text-gray-400 truncate">{hop.loc}</div>
                    {/if}
                    <div class="mt-1.5 pt-1 border-t border-gray-700/60 flex justify-between items-center text-[11px]">
                      <span>Loss: <strong class={hop.lossRate > 0 ? (hop.lossRate > 20 ? 'text-red-400' : 'text-amber-400') : 'text-emerald-400'}>{hop.lossRate.toFixed(1)}%</strong></span>
                      <span>Avg: <strong>{hop.avg >= 0 ? hop.avg.toFixed(1) + 'ms' : 'N/A'}</strong></span>
                    </div>
                  </div>
                  {#if idx < mtrHops.length - 1}
                    <div class="flex flex-col items-center justify-center text-gray-500">
                      <Icon path={icons.mdiArrowRight} size={1} />
                    </div>
                  {/if}
                {/each}
              </div>
            </div>

            <!-- ② MTR 統計 DataTables (プロジェクトルール: divで囲む) -->
            <div class="flex-shrink-0 block w-full border border-gray-700/60 rounded-lg overflow-hidden p-0.5 bg-gray-900/40">
              <table id="mtrTable" class="display compact" style="width:100%"></table>
            </div>

            <!-- ③ MTR ホップ別プロファイルチャート -->
            <div id="mtrChart" class="flex-shrink-0 block w-full" style="height: 180px; min-height: 180px;"></div>
          </div>
        </TabItem>
      {/if}
      {#if !wait && results.length > 0}
        {#if canShowSmoke}
          <TabItem onclick={showSmoke}>
            {#snippet titleSlot()}
              <div class="flex items-center gap-2">
                <Icon path={icons.mdiWeatherCloudy} size={1} />
                { $_('Ping.Smokeping') }
              </div>
            {/snippet}
            <div id="smoke"></div>
          </TabItem>
        {/if}
        {#if canShowHistogram}
          <TabItem onclick={showHistogram}>
            {#snippet titleSlot()}
              <div class="flex items-center gap-2">
                <Icon path={icons.mdiChartHistogram} size={1} />
                { $_('Ping.Histogram') }
              </div>
            {/snippet}
            <div id="histogram"></div>
          </TabItem>
        {/if}
        <TabItem onclick={show3D}>
          {#snippet titleSlot()}
            <div class="flex items-center gap-2">
              <Icon path={icons.mdiRotate3d} size={1} />
              { $_('Ping.Chart3D') }
            </div>
          {/snippet}
          <div id="chart3d"></div>
        </TabItem>
        {#if canShowLinear}
          <TabItem onclick={showLinear}>
            {#snippet titleSlot()}
              <div class="flex items-center gap-2">
                <Icon path={icons.mdiChartScatterPlot} size={1} />
                { $_('Ping.LineSpeed') }
              </div>
            {/snippet}
            <div id="linear"></div>
          </TabItem>
        {/if}
        {#if canShowWorld}
          <TabItem onclick={showWorld}>
            {#snippet titleSlot()}
              <div class="flex items-center gap-2">
                <Icon path={icons.mdiMapMarker} size={1} />
                { $_('Ping.World') }
              </div>
            {/snippet}
            <div id="world"></div>
          </TabItem>
        {/if}
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if pingTab}
        <Toggle bind:checked={beep}>BEEP</Toggle>
        {#if wait}
          <GradientButton shadow type="button" color="red" onclick={stop} size="xs">
            <Icon path={icons.mdiStop} size={1} />
            { $_('Ping.Stop') }
          </GradientButton>
        {:else}
          <GradientButton shadow type="button" color="blue" onclick={start} size="xs">
            <Icon path={icons.mdiPlay} size={1} />
            { $_('Ping.Start') }
          </GradientButton>
          <GradientButton
            shadow
            type="button"
            size="xs"
            color="lime"
            class="ml-2"
            onclick={() => {
              showHelp = true;
            }}
          >
            <Icon path={icons.mdiHelp} size={1} />
            <span>
              {$_("Ping.Help")}
            </span>
          </GradientButton>
        {/if}
      {:else if hasAI}
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
        { $_('Ping.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

<audio src={ping_ok} bind:this={sound_ok}></audio>
<audio src={ping_ng} bind:this={sound_ng}></audio>

<Help bind:show={showHelp} page="ping" />

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`ping_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainPingReport(ip, results || [], activeTab)}
/>

<style>
  #pingChart {
    min-height: 200px;
    height: 20vh;
    width:  98%;
    margin: 0 auto;
  }
  #chart3d,
  #histogram,
  #linear,
  #world,
  #smoke {
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
  #mtrChart {
    min-height: 180px;
    height: 180px;
    width: 98%;
    margin: 0 auto;
  }
</style>