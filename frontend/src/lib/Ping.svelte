<script lang="ts">
  import { Modal, Button, Tabs, TabItem, Input, Select } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import Icon from "mdi-svelte";
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

  export let nodeID = "";
  let show: boolean = false;
  const dispatch = createEventDispatcher();
  let pingTab = true;
  let wait = false;
  let table = undefined;
  let chart = undefined;
  let chartOption = undefined;
  let results = [];
  let ip = "";
  let size = 64;
  let count = 10;
  let ttl = 64;
  const pingReq = {
    size: 0,
    count: 0,
    ttl: 64,
  };
  let timer = undefined;
  let canShowLinear = false;
  let canShowWorld = false;
  let canShowHistogram = false;

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
    if (table) {
      table.destroy();
      table = undefined;
    }
    table = new DataTable("#pingTable", {
      columns: columns,
      data: results,
      order: [[1, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
  };
  const showPing = async () => {
    await tick();
    if (chart) {
      chart.dispose();
    }
    chartOption = getPingChartOption();
    chart = echarts.init(document.getElementById("pingChart"));
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
      showTable();
    }
    chart.setOption(chartOption);
    chart.resize();
  };
  const countList = [
    { name: "連続", value: -1 },
    { name: "1回", value: 1 },
    { name: "3回", value: 3 },
    { name: "5回", value: 5 },
    { name: "10回", value: 10 },
    { name: "20回", value: 20 },
    { name: "30回", value: 30 },
    { name: "50回", value: 50 },
    { name: "100回", value: 100 },
  ];

  const sizeList = [
    { name: "変化モード", value: -1 },
    { name: "64", value: 64 },
    { name: "128", value: 128 },
    { name: "256", value: 256 },
    { name: "512", value: 512 },
    { name: "1024", value: 1024 },
    { name: "1500", value: 1500 },
  ];

  const ttlList = [
    { name: "トレースルート", value: -1 },
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

  const renderPingStat = (s, type) => {
    if (type == "sort") {
      return s;
    }

    let state = "unknown";
    let name = "不明";
    switch (s) {
      case 1:
        state = "normal";
        name = "正常";
        break;
      case 2:
        state = "error";
        name = "Timeout";
        break;
      case 3:
        state = "warn";
        name = "エラー";
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

  const renderTimeStamp = (ts) => {
    return formatTime(new Date(ts * 1000), "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
  };

  const renderRespTime = (t) => {
    return (t / (1000 * 1000 * 1000)).toFixed(6);
  };

  const columns = [
    {
      data: "Stat",
      title: "結果",
      width: "10%",
      render: renderPingStat,
    },
    {
      data: "TimeStamp",
      title: "時刻",
      width: "15%",
      render: renderTimeStamp,
    },
    {
      data: "Time",
      title: "応答時間",
      width: "10%",
      render: renderRespTime,
    },
    {
      data: "Size",
      title: "サイズ",
      width: "10%",
    },
    {
      data: "SendTTL",
      title: "送信TTL",
      width: "10%",
    },
    {
      data: "RecvTTL",
      title: "受信TTL",
      width: "10%",
    },
    {
      data: "RecvSrc",
      title: "応答送信IP",
      width: "15%",
    },
    {
      data: "Loc",
      title: "位置",
      width: "20%",
    },
  ];

  const showHistogram = async () => {
    await tick();
    showPingHistgram("histogram", results);
  };

  const show3D = async () => {
    await tick();
    showPing3DChart("3d", results);
  };

  const showLinear = async () => {
    await tick();
    showPingLinearChart("linear", results);
  };

  const showWorld = async () => {
    await tick();
    showPingMapChart("world", results);
  };

  const start = () => {
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
      timer = setTimeout(() => _doPing(), 1000);
    } else {
      wait = false;
      canShowLinear = size == -1;
      canShowHistogram = !canShowLinear;
      canShowWorld = false;
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal bind:open={show} size="xl" permanent class="w-full" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem bind:open={pingTab} on:click={showPing}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          PING実行
        </div>
        <div class="flex flex-row mb-2">
          <Input
            type="text"
            bind:value={ip}
            placeholder="IPまたはホスト名"
            required
          />
          <Select
            class="ml-2"
            items={countList}
            bind:value={count}
            placeholder="回数"
          />
          <Select
            class="ml-2"
            items={sizeList}
            bind:value={size}
            placeholder="サイズ"
          />
          <Select
            class="ml-2"
            items={ttlList}
            bind:value={ttl}
            placeholder="TTL"
          />
        </div>
        <div id="pingChart" class="mb-2" style="height: 200px;" />
        <table id="pingTable" class="display compact" style="width:99%" />
      </TabItem>
      {#if !wait && results.length > 0}
        {#if canShowHistogram}
          <TabItem on:click={showHistogram}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiChartHistogram} size={1} />
              ヒストグラム
            </div>
            <div id="histogram" style="height: 500px;" />
          </TabItem>
        {/if}
        <TabItem on:click={show3D}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiRotate3d} size={1} />
            3D分析
          </div>
          <div id="3d" style="height: 500px;" />
        </TabItem>
        {#if canShowLinear}
          <TabItem on:click={showLinear}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiChartScatterPlot} size={1} />
              回線速度予測
            </div>
            <div id="linear" style="height: 500px;" />
          </TabItem>
        {/if}
        {#if canShowWorld}
          <TabItem on:click={showWorld}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiMapMarker} size={1} />
              経路分析
            </div>
            <div id="world" style="height: 500px;" />
          </TabItem>
        {/if}
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if pingTab}
        {#if wait}
          <Button type="button" color="red" on:click={stop} size="sm">
            <Icon path={icons.mdiStop} size={1} />
            停止
          </Button>
        {:else}
          <Button type="button" color="blue" on:click={start} size="sm">
            <Icon path={icons.mdiPlay} size={1} />
            開始
          </Button>
        {/if}
      {/if}
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>
