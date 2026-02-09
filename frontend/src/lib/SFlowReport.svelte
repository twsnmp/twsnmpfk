<script lang="ts">
  import "maplibre-gl/dist/maplibre-gl.css";
  import { Map as MapGl, NavigationControl, Marker,Popup } from "maplibre-gl";
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Select,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {GetLocConf} from "../../wailsjs/go/main/App";
  import {
    getSFlowFlowList,
    getSFlowSenderList,
    getSFlowServiceList,
    getSFlowReasonList,
    showSFlowTop,
    showSFlowTraffic,
    showSFlowGraph,
    showSFlowSender3D,
    showSFlowService3D,
    showSFlowReason3D,
    showSFlowFlow3D,
    getSFlowFFTMap,
    showSFlowFFT,
    showSFlowFFT3D,
  } from "./chart/sflow";
  import { showLogHeatmap } from "./chart/eventlog";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { getTableLang } from "./common";
  import { _ } from "svelte-i18n";
  import { copyText } from "svelte-copy";

  export let show: boolean = false;
  export let logs: any = undefined;

  let chart: any = undefined;
  let topList: any = [];
  let fftMap: any = undefined;
  let tab: string = "heatmap";
  let tableTop: any = undefined;
  let selectedCountTop = 0;
  let tableFlow: any = undefined;
  let selectedCountFlow = 0;

  const onOpen = () => {
    chart = undefined;
    topList = [];
    fftMap = undefined;
    tab = "heatmap";
    tableTop = undefined;
    selectedCountTop = 0;
    tableFlow = undefined;
    selectedCountFlow = 0;
    flowList = [];
    showHeatmap();
  };

  const close = () => {
    show = false;
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  const showHeatmap = async () => {
    await tick();
    tab = "heatmap";
    chart = showLogHeatmap("heatmap", logs);
  };


  const showTraffic = async () => {
    await tick();
    tab = "traffic";
    chart = showSFlowTraffic("traffic", logs);
  };

  let topListType: string = "sender";
  const topListTypes = [
    { value: "sender", name: $_('NetFlowReport.Sender') },
    { value: "sender_mac", name: $_('NetFlowReport.SenderMAC') },
    { value: "service", name: $_('NetFlowReport.Service') },
    { value: "flow", name: $_('NetFlowReport.Flow') },
    { value: "flow_mac", name: $_('NetFlowReport.MACFlow') },
    { value: "reason", name: $_("SFlow.Reason")},
  ];

  const columnsTop = [
    {
      data: "Name",
      title: $_('NetFlowReport.Name'),
      width: "70%",
    },
    {
      data: "Count",
      with: "15%",
      title: $_("Ts.Count"),
    },
    {
      data: "Bytes",
      with: "15%",
      title: $_('NetFlowReport.Bytes'),
    },
  ];

  const showTopTable = () => {
    let order = [[1, "desc"]];
    selectedCountTop = 0;
    tableTop = new DataTable("#topTable", {
      destroy: true,
      stateSave: true,
      columns: columnsTop,
      data: topList,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order,
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    tableTop.on("select", () => {
      selectedCountTop = tableTop.rows({ selected: true }).count();
    });
    tableTop.on("deselect", () => {
      selectedCountTop = tableTop.rows({ selected: true }).count();
    });
  };

  const showTopList = async () => {
    tab = "topList";
    switch (topListType) {
      case "sender":
        topList = getSFlowSenderList(logs,false);
        break;
      case "sender_mac":
        topList = getSFlowSenderList(logs,true);
        break;
      case "service":
        topList = getSFlowServiceList(logs);
        break;
      case "flow":
        topList = getSFlowFlowList(logs,false);
        break;
      case "flow_mac":
        topList = getSFlowFlowList(logs,true);
        break;
      case "reason":
        topList = getSFlowReasonList(logs);
        break;
    }
    await tick();
    showTopTable();
    chart = showSFlowTop("topList", topList);
  };

  const showTopList3D = async () => {
    tab = "topList3D";
    await tick();
    switch (topListType) {
      case "sender":
        chart = showSFlowSender3D("topList3D", logs, false);
        break;
      case "sender_mac":
        chart = showSFlowSender3D("topList3D", logs,true);
        break;
      case "service":
        chart = showSFlowService3D("topList3D", logs);
        break;
      case "flow":
        chart = showSFlowFlow3D("topList3D", logs,false);
        break;
      case "flow_mac":
        chart = showSFlowFlow3D("topList3D", logs, true);
      case "reason":
        chart = showSFlowReason3D("topList3D", logs);
        break;
    }
  };

  let flowType: string = "circular";
  const flowTypes = [
    { value: "force", name: $_('NetFlowReport.Force') },
    { value: "circular", name: $_('NetFlowReport.Circular') },
    { value: "gl", name: $_('NetFlowReport.GL') },
  ];

  let flowMode: number = 0;
  const flowModes = [
    { value: 0, name: $_('NetFlowReport.SrcDstIP')},
    { value: 1, name: $_('NetFlowReport.SrcDstMac')},
    { value: 2, name: $_('NetFlowReport.SrcMacIP') },
    { value: 3, name: $_('NetFlowReport.DstMacIP') },
  ];

  let flowList :any = [];

  const showFlow = async () => {
    await tick();
    tab = "flow";
    const r = showSFlowGraph("flow", logs, flowMode, flowType);
    chart = r.chart;
    flowList = r.edges;
    showFlowTable();
  };

  const columnsFlow = [
    {
      data: "source",
      title: "Source",
      width: "40%",
    },
    {
      data: "target",
      title: "Target",
      width: "40%",
    },
    {
      data: "value",
      title: $_('NetFlowReport.Bytes'),
      width: "20%",
    },
  ];
 
  const showFlowTable = async () => {
    let order = [[1, "desc"]];
    selectedCountFlow = 0;
    tableFlow = new DataTable("#flowTable", {
      destroy: true,
      columns: columnsFlow,
      data: flowList,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order,
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    tableFlow.on("select", () => {
      selectedCountFlow = tableFlow.rows({ selected: true }).count();
    });
    tableFlow.on("deselect", () => {
      selectedCountFlow = tableFlow.rows({ selected: true }).count();
    });
  }

  let fftSrc: string = "Total";
  const fftSrcs: any = [];

  let fftType: string = "hz";
  const fftTypes = [
    { value: "hz", name: $_('NetFlowReport.HZ') },
    { value: "sec", name: $_('NetFlowReport.Sec') },
  ];

  const showFFT = async () => {
    fftMap = getSFlowFFTMap(logs);
    fftSrcs.length = 0;
    fftSrcs.push({
      value: "Total",
      name: $_('NetFlowReport.Total'),
    });
    fftMap.forEach((v: any, k: any) => {
      fftSrcs.push({
        value: k,
        name: k,
      });
    });
    await tick();
    tab = "fft";
    chart = showSFlowFFT("fft", fftMap, fftSrc, fftType);
  };

  const updateFFT = async () => {
    await tick();
    chart = showSFlowFFT("fft", fftMap, fftSrc, fftType);
  };

  const showFFT3D = async () => {
    fftMap = getSFlowFFTMap(logs);
    await tick();
    tab = "fft3d";
    chart = showSFlowFFT3D("fft3d", fftMap, fftType);
  };

  const updateFFT3D = async () => {
    await tick();
    chart = showSFlowFFT3D("fft3d", fftMap, fftType);
  };

  const getLngLat = (loc: string): [number, number] => {
    const a = loc.split(",");
    if (a.length < 2) {
      return [0, 0];
    }
    if (a.length == 2) {
      return [Number(a[0]), Number(a[1])];
    }
    return [Number(a[2]), Number(a[1])];
  };

  const showMap = async () => {
    await tick();
    tab = "map";
    const locConf = await GetLocConf();
    const s = locConf.Style.startsWith("{") ? JSON.parse(locConf.Style) : locConf.Style;
    const map = new MapGl({
      container: "map",
      style: s,
      center: getLngLat(locConf.Center),
      zoom: locConf.Zoom,
    });
    const srcLocs = new Map();
    const dstLocs = new Map();
    logs.forEach((l:any) => {
      if(l.SrcLoc && !l.SrcLoc.startsWith("LOCAL") && !l.SrcLoc.startsWith(",0")) {
        srcLocs.set(l.SrcLoc,l.SrcAddr + "<br/>" +l.SrcLoc);
      }
      if(l.DstLoc && !l.DstLoc.startsWith("LOCAL") && !l.DstLoc.startsWith(",0")) {
        dstLocs.set(l.DstLoc,l.DstAddr + "<br/>" +l.DstLoc);
      }
    });
    srcLocs.forEach((v,k)=> {
      if (dstLocs.has(k)) {
        const marker = new Marker({
          color: "#00c",
        }).setLngLat(getLngLat(k))
        .setPopup(new Popup().setHTML("Both<br/>" + v))
        .addTo(map);
      } else {
        const marker = new Marker({
          color: "#cc0",
        }).setLngLat(getLngLat(k))
        .setPopup(new Popup().setHTML("Only Src<br/>" + v))
        .addTo(map);
      }
    });
    dstLocs.forEach((v,k)=> {
      if (!srcLocs.has(k)) {
        const marker = new Marker({
          color: "#c00",
        }).setLngLat(getLngLat(k))
        .setPopup(new Popup().setHTML("Only Dst<br/>" + v))
        .addTo(map);
      }
    });
    map.addControl(
      new NavigationControl({
        visualizePitch: true,
      })
    );
  }

  let copiedTop = false;
  const copyTop = () => {
    const selected = tableTop.rows({ selected: true }).data();
    let s: string[] = [];
    const h = columnsTop.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of columnsTop) {
        const d = (selected[i][c.data] || "") + "";
        row.push(d.replaceAll("\n", " "));
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copiedTop = true;
    setTimeout(() => (copiedTop = false), 2000);
  };

  let copiedFlow = false;
  const copyFlow = () => {
    const selected = tableFlow.rows({ selected: true }).data();
    let s: string[] = [];
    const h = columnsFlow.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of columnsFlow) {
        const d = (selected[i][c.data] || "") + "";
        row.push(d.replaceAll("\n", " "));
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copiedFlow = true;
    setTimeout(() => (copiedFlow = false), 2000);
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
    <Tabs style="underline">
      <TabItem
        open
        on:click={() => {
          showHeatmap();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          {$_("TrapReport.Heatmap")}
        </div>
        <div id="heatmap" />
      </TabItem>
      <TabItem
        on:click={() => {
          showTraffic();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiTrafficCone} size={1} />
          {$_('NetFlowReport.Traffic')}
        </div>
        <div id="traffic" />
      </TabItem>
      <TabItem
        on:click={() => {
          showTopList();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiFormatListNumbered} size={1} />
          {$_('NetFlowReport.TopLIst')}
        </div>
        <div class="grid gap-2 grid-cols-2">
          <div id="topList" />
          <div>
            <table id="topTable" class="display compact" style="width:99%" />
          </div>
        </div>
      </TabItem>
      <TabItem
        on:click={() => {
          showTopList3D();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiFormatListNumbered} size={1} />
          {$_('NetFlowReport.TopList3D')}
        </div>
        <div id="topList3D" />
      </TabItem>
      <TabItem
        on:click={() => {
          showFlow();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiFormatListNumbered} size={1} />
          {$_('NetFlowReport.FlowGraph')}
        </div>
        <div class="grid gap-2 grid-cols-5">
          <div id="flow" class="col-span-3" />
          <div class="col-span-2">
            <table id="flowTable" class="display compact" style="width:99%" />
          </div>
        </div>
      </TabItem>
      <TabItem
        on:click={() => {
          showFFT();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiWaveform} size={1} />
          {$_('NetFlowReport.FFT')}
        </div>
        <div id="fft" />
      </TabItem>
      <TabItem
        on:click={() => {
          showFFT3D();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiWaveform} size={1} />
          {$_('NetFlowReport.FFT3D')}
        </div>
        <div id="fft3d" />
      </TabItem>
      <TabItem
        on:click={() => {
          showMap();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiMapMarker} size={1} />
          {$_('NetFlowReport.Map')}
        </div>
        <div id="map" />
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if tab == "topList"}
        <Select
          placeholder={$_('NetFlowReport.SumType')}
          class="ml-10 w-48"
          items={topListTypes}
          bind:value={topListType}
          size="sm"
          on:change={() => {
            showTopList();
          }}
        />
        {#if selectedCountTop > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={copyTop}
            size="xs"
          >
            {#if copiedTop}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            Copy
          </GradientButton>
        {/if}
      {/if}
      {#if tab == "topList3D"}
        <Select
          placeholder={$_('NetFlowReport.SumType')}
          class="ml-10 w-48"
          items={topListTypes}
          bind:value={topListType}
          size="sm"
          on:change={() => {
            showTopList3D();
          }}
        />
      {/if}
      {#if tab == "flow"}
        <Select
          placeholder={$_('NetFlowReport.FlowMode')}
          class="ml-10 w-48"
          items={flowModes}
          bind:value={flowMode}
          size="sm"
          on:change={() => {
            showFlow();
          }}
        />
        <Select
          placeholder={$_('NetFlowReport.FlowType')}
          class="ml-10 w-48"
          items={flowTypes}
          bind:value={flowType}
          size="sm"
          on:change={() => {
            showFlow();
          }}
        />
        {#if selectedCountFlow > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={copyFlow}
            size="xs"
          >
            {#if copiedFlow}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            Copy
          </GradientButton>
        {/if}
      {/if}
      {#if tab == "fft"}
        <Select
          placeholder={$_('NetFlow.SrcAddr')}
          class="ml-10 w-48"
          items={fftSrcs}
          bind:value={fftSrc}
          size="sm"
          on:change={() => {
            updateFFT();
          }}
        />
        <Select
          placeholder={$_('NetFlowReport.DIspType')}
          class="ml-10 w-48"
          items={fftTypes}
          bind:value={fftType}
          size="sm"
          on:change={() => {
            updateFFT();
          }}
        />
      {/if}
      {#if tab == "fft3d"}
        <Select
          placeholder={$_('NetFlowReport.DIspType')}
          class="ml-10 w-48"
          items={fftTypes}
          bind:value={fftType}
          size="sm"
          on:change={() => {
            updateFFT3D();
          }}
        />
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("TrapReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #heatmap,
  #traffic,
  #topList3D,
  #fft,
  #fft3d,
  #flow {
    min-height: 500px;
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
  #topList {
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }
  #map {
    height: 70vh;
    width: 98%;
    margin: 0 auto;
  }

</style>
