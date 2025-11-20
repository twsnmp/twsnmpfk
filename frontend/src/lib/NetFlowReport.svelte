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
    getNetFlowFlowList,
    getNetFlowSenderList,
    getNetFlowServiceList,
    getNetFlowFumbleList,
    showNetFlowHistogram,
    showNetFlowTop,
    showNetFlowTraffic,
    showNetFlowGraph,
    showNetFlowSender3D,
    showNetFlowService3D,
    showNetFlowFlow3D,
    getNetFlowFFTMap,
    showNetFlowFFT,
    showNetFlowFFT3D,
  } from "./chart/netflow";
  import { showLogHeatmap } from "./chart/eventlog";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { getTableLang,renderBytes,renderCount,renderSpeed } from "./common";
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
  let locConf :any = undefined;

  const onOpen = async () => {
    locConf = await GetLocConf();
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

  let histogramType: string = "size";
  const histogramTypes = [
    { value: "size", name: $_('NetFlowReport.Size') },
    { value: "dur", name: $_('NetFlowReport.Dur') },
    { value: "speed", name: $_('NetFlowReport.Speed') },
  ];

  const showHistogram = async () => {
    await tick();
    tab = "histogram";
    chart = showNetFlowHistogram("histogram", logs, histogramType);
  };

  let trafficType: string = "bytes";
  const trafficTypes = [
    { value: "bytes", name: $_('NetFlowReport.Bytes') },
    { value: "packets", name: $_('NetFlowReport.Packets') },
    { value: "bps", name: $_('NetFlowReport.BPS') },
    { value: "pps", name: $_('NetFlowReport.PPS') },
  ];

  const showTraffic = async () => {
    await tick();
    tab = "traffic";
    chart = showNetFlowTraffic("traffic", logs, trafficType);
  };

  let topListType: string = "sender";
  const topListTypes = [
    { value: "sender", name: $_('NetFlowReport.Sender') },
    { value: "sender_mac", name: $_('NetFlowReport.SenderMAC') },
    { value: "service", name: $_('NetFlowReport.Service') },
    { value: "flow", name: $_('NetFlowReport.Flow') },
    { value: "flow_mac", name: $_('NetFlowReport.MACFlow') },
    { value: "fumble_src", name: $_('NetFlowReport.FumbleSrc')},
    { value: "fumble_flow", name: $_('NetFlowReport.FumbleFlow')},
  ];

  let topListDataType: string = "bytes";
  const topListDataTypes = [
    { value: "bytes", name: $_('NetFlowReport.Bytes') },
    { value: "packets", name: $_('NetFlowReport.Packets') },
    { value: "dur", name: $_('NetFlowReport.Dur') },
    { value: "bps", name: $_('NetFlowReport.BPS') },
    { value: "pps", name: $_('NetFlowReport.PPS') },
  ];

  const columnsTop = [
    {
      data: "Name",
      title: $_('NetFlowReport.Name'),
      width: "30%",
    },
    {
      data: "Bytes",
      title: $_('NetFlowReport.Bytes'),
      render: renderBytes,
      "className": "dt-right",
    },
    {
      data: "Packets",
      title: $_('NetFlowReport.Packets'),
      render: renderCount,
      "className": "dt-right",
    },
    {
      data: "Dur",
      title: $_('NetFlowReport.Dur'),
      "className": "dt-right",
    },
    {
      data: "bps",
      title: $_('NetFlowReport.BPS'),
      render: renderSpeed,
      "className": "dt-right",
    },
    {
      data: "pps",
      title: $_('NetFlowReport.PPS'),
      render: renderCount,
      "className": "dt-right",
    },
  ];

  const showTopTable = () => {
    let order = [[1, "desc"]];
    selectedCountTop = 0;
    tableTop = new DataTable("#topTable", {
      destroy: true,
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
        topList = getNetFlowSenderList(logs,false);
        break;
      case "sender_mac":
        topList = getNetFlowSenderList(logs,true);
        break;
      case "service":
        topList = getNetFlowServiceList(logs);
        break;
      case "flow":
        topList = getNetFlowFlowList(logs,false);
        break;
      case "flow_mac":
        topList = getNetFlowFlowList(logs,true);
        break;
      case "fumble_src":
        topList = getNetFlowFumbleList(logs,false);
        break;
      case "fumble_flow":
        topList = getNetFlowFumbleList(logs,true);
        break;
    }
    await tick();
    showTopTable();
    chart = showNetFlowTop("topList", topList, topListDataType);
  };

  const updateTopList = async () => {
    await tick();
    chart = showNetFlowTop("topList", topList, topListDataType);
  };

  const topList3DDataTypes = [
    { value: "bytes", name: $_('NetFlowReport.Bytes') },
    { value: "packets", name: $_('NetFlowReport.Packets') },
    { value: "dur", name: $_('NetFlowReport.Dur') },
  ];

  const showTopList3D = async () => {
    tab = "topList3D";
    await tick();
    switch (topListType) {
      case "sender":
        chart = showNetFlowSender3D("topList3D", logs, topListDataType,false,false);
        break;
      case "sender_mac":
        chart = showNetFlowSender3D("topList3D", logs, topListDataType,true,false);
        break;
      case "fumble_src":
        chart = showNetFlowSender3D("topList3D", logs, topListDataType,false,true);
        break;
      case "service":
        chart = showNetFlowService3D("topList3D", logs, topListDataType);
        break;
      case "flow":
        chart = showNetFlowFlow3D("topList3D", logs, topListDataType,false,false);
        break;
      case "flow_mac":
        chart = showNetFlowFlow3D("topList3D", logs, topListDataType,true,false);
        break;
      case "fumble_flow":
        chart = showNetFlowFlow3D("topList3D", logs, topListDataType,false,true);
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
    const r = showNetFlowGraph("flow", logs, flowMode, flowType);
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
      render: renderBytes,
      "className": "dt-right",
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
    fftMap = getNetFlowFFTMap(logs);
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
    chart = showNetFlowFFT("fft", fftMap, fftSrc, fftType);
  };

  const updateFFT = async () => {
    await tick();
    tab = "fft";
    chart = showNetFlowFFT("fft", fftMap, fftSrc, fftType);
  };

  const showFFT3D = async () => {
    fftMap = getNetFlowFFTMap(logs);
    await tick();
    tab = "fft3d";
    chart = showNetFlowFFT3D("fft3d", fftMap, fftType);
  };

  const updateFFT3D = async () => {
    await tick();
    tab = "fft3d";
    chart = showNetFlowFFT3D("fft3d", fftMap, fftType);
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
          showHistogram();
        }}
      >
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartHistogram} size={1} />
          {$_('PollingReport.Histogram')}
        </div>
        <div id="histogram" />
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
      {#if locConf && locConf.Style}
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
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if tab == "histogram"}
        <Select
          placeholder={$_('NetFlowReport.DataType')}
          class="ml-10 w-48"
          items={histogramTypes}
          bind:value={histogramType}
          size="sm"
          on:change={() => {
            showHistogram();
          }}
        />
      {/if}
      {#if tab == "traffic"}
        <Select
          placeholder={$_('NetFlowReport.DataType')}
          class="ml-10 w-48"
          items={trafficTypes}
          bind:value={trafficType}
          size="sm"
          on:change={() => {
            showTraffic();
          }}
        />
      {/if}
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
        <Select
          placeholder={$_('NetFlowReport.DataType')}
          class="ml-10 w-48"
          items={topListDataTypes}
          bind:value={topListDataType}
          size="sm"
          on:change={() => {
            updateTopList();
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
        <Select
          placeholder={$_('NetFlowReport.DataType')}
          class="ml-10 w-48"
          items={topList3DDataTypes}
          bind:value={topListDataType}
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
  #histogram,
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
