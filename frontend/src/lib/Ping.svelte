<script lang="ts">
  import ping_ok from "../assets/sound/ping_ok.mp3";
  import ping_ng from "../assets/sound/ping_ng.mp3";
  import { Modal, GradientButton, Tabs, TabItem, Input, Select,Toggle } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    getPingChartOption,
    showPing3DChart,
    showPingHistgram,
    showPingLinearChart,
    showPingMapChart,
  } from "./chart/ping";
  import { GetNode, Ping } from "../../wailsjs/go/main/App";
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

  export let nodeID = "";
  let show: boolean = false;
  const dispatch = createEventDispatcher();
  let pingTab = true;
  let wait = false;
  let table :any = undefined;
  let chart :any = undefined;
  let chartOption :any = undefined;
  let results :any = [];
  let ip = "";
  let ipColor: any = "base";
  let size = 64;
  let count = 10;
  let ttl = 64;
  const pingReq = {
    size: 0,
    count: 0,
    ttl: 64,
  };
  let timer :any = undefined;
  let canShowLinear = false;
  let canShowWorld = false;
  let canShowHistogram = false;
  let beep = false;
  let sound_ok :any;
  let sound_ng :any;
  let showHelp = false;

  onMount(async () => {
    const node = await GetNode(nodeID);
    if (node && node.IP) {
      ip = node.IP;
    }
    show = true;
    showPing();
  });

  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  });

  const showTable = () => {
    if (table && DataTable.isDataTable("#pingTable")) {
      table.clear();
      table.destroy();
      table = undefined;
    }
    table = new DataTable("#pingTable", {
      columns: columns,
      paging: false,
      searching:false,
      info:false,
      scrollY: "40vh",
      data: results,
      order: [[1, "asc"]],
      language: getTableLang(),
    });
  };

  const showPing = async () => {
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
  const countList = [
    { name: $_('Ping.Cont'), value: -1 },
    { name: $_('Ping.Coun1'), value: 1 },
    { name: $_('Ping.Count3'), value: 3 },
    { name: $_('Ping.Count5'), value: 5 },
    { name: $_('Ping.Count10'), value: 10 },
    { name: $_('Ping.Count20'), value: 20 },
    { name: $_('Ping.Count30'), value: 30 },
    { name: $_('Ping.Count50'), value: 50 },
    { name: $_('Ping.Count100'), value: 100 },
  ];

  const sizeList = [
    { name: $_('Ping.IncSize'), value: -1 },
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "256", value: 256 },
    { name: "512", value: 512 },
    { name: "1024", value: 1024 },
    { name: "1500", value: 1500 },
  ];

  const ttlList = [
    { name: $_('Ping.TraceRoute'), value: -1 },
    { name: "1", value: 1 },
    { name: "2", value: 2 },
    { name: "4", value: 4 },
    { name: "8", value: 8 },
    { name: "16", value: 16 },
    { name: "32", value: 32 },
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "254", value: 254 },
  ];

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
      ` text-xl" style="color:` +
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

  const showHistogram = async () => {
    await tick();
    reportChart = showPingHistgram("histogram", results);
  };

  const show3D = async () => {
    await tick();
    reportChart = showPing3DChart("chart3d", results);
  };

  const showLinear = async () => {
    await tick();
    reportChart = showPingLinearChart("linear", results);
  };

  const showWorld = async () => {
    await tick();
    reportChart = showPingMapChart("world", results);
  };

  const start = () => {
    if (!ip) {
      ipColor ="red";
      return;
    } else {
      ipColor = "base";
    }
    stopFlag = false;
    if (chart) {
      chartOption.series[0].data = [];
      chartOption.series[1].data = [];
      chartOption.series[2].data = [];
    }
    wait = true;
    pingReq.count = 0;
    pingReq.size = size < 0 ? 0 :size;
    if (ttl === -1) {
      pingReq.ttl = 1;
      count = -1;
      size = 64;
    } else {
      pingReq.ttl = ttl;
    }
    results = [];
    canShowWorld = false;
    _doPing();
  };

  let stopFlag = true;

  const stop = () => {
    stopFlag = true;
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
      if(beep && sound_ok) {
        sound_ok.play();
      }
      if(r.Loc && r.Loc.startsWith("LOCAL")) {
        canShowWorld = true;
      }
    } else {
      if(beep && sound_ng) {
        sound_ng.play();
      }
    }
    if ((count === -1 || pingReq.count < count) && !stopFlag) {
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
          return;
        }
      }
      timer = setTimeout(() => _doPing(), beep ? 2000 : 1000);
    } else {
      wait = false;
      canShowLinear = size == -1;
      canShowHistogram = !canShowLinear;
    }
  };

  const close = () => {
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

</script>

<svelte:window on:resize={resizeChart} />

<Modal bind:open={show} size="xl" dismissable={false} class="w-full" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem bind:open={pingTab} on:click={showPing}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiCheckNetwork} size={1} />
          { $_('Ping.DoPing') }
        </div>
        <div class="flex flex-row mb-2">
          <Input
            type="text"
            bind:value={ip}
            placeholder={ $_('Ping.IPOrHost') }
            color={ipColor}
            size="sm"
          />
          <Select
            class="ml-2"
            items={countList}
            bind:value={count}
            placeholder={ $_('Ping.Count') }
            size="sm"
          />
          <Select
            class="ml-2"
            items={sizeList}
            bind:value={size}
            placeholder={ $_('Ping.Size') }
            size="sm"
          />
          <Select
            class="ml-2"
            items={ttlList}
            bind:value={ttl}
            placeholder="TTL"
            size="sm"
          />
        </div>
        <div id="pingChart" class="mb-2" />
        <table id="pingTable" class="display compact" style="width:99%" />
      </TabItem>
      {#if !wait && results.length > 0}
        {#if canShowHistogram}
          <TabItem on:click={showHistogram}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiChartHistogram} size={1} />
              { $_('Ping.Histogram') }
            </div>
            <div id="histogram"/>
          </TabItem>
        {/if}
        <TabItem on:click={show3D}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiRotate3d} size={1} />
            { $_('Ping.Chart3D') }
          </div>
          <div id="chart3d" />
        </TabItem>
        {#if canShowLinear}
          <TabItem on:click={showLinear}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiChartScatterPlot} size={1} />
              { $_('Ping.LineSpeed') }
            </div>
            <div id="linear"/>
          </TabItem>
        {/if}
        {#if canShowWorld}
          <TabItem on:click={showWorld}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiMapMarker} size={1} />
              { $_('Ping.World') }
            </div>
            <div id="world" />
          </TabItem>
        {/if}
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if pingTab}
        <Toggle bind:checked={beep}>BEEP</Toggle>
        {#if wait}
          <GradientButton shadow type="button" color="red" on:click={stop} size="xs">
            <Icon path={icons.mdiStop} size={1} />
            { $_('Ping.Stop') }
          </GradientButton>
        {:else}
          <GradientButton shadow type="button" color="blue" on:click={start} size="xs">
            <Icon path={icons.mdiPlay} size={1} />
            { $_('Ping.Start') }
          </GradientButton>
          <GradientButton
            shadow
            type="button"
            size="xs"
            color="lime"
            class="ml-2"
            on:click={() => {
              showHelp = true;
            }}
          >
            <Icon path={icons.mdiHelp} size={1} />
            <span>
              {$_("Ping.Help")}
            </span>
          </GradientButton>
        {/if}
      {/if}
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('Ping.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

<audio src={ping_ok} bind:this={sound_ok}></audio>
<audio src={ping_ng} bind:this={sound_ng}></audio>

{#if showHelp}
  <Help
    page="ping"
    on:close={() => {
      showHelp = false;
    }}
  />
{/if}

<style>
  #pingChart {
    min-height: 200px;
    height: 25vh;
    width:  98%;
    margin: 0 auto;
  }
  #chart3d,
  #histogram,
  #linear,
  #world {
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
</style>