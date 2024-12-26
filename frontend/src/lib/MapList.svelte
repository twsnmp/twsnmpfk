<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Spinner,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNodes,
    GetNetworks,
    GetDrawItems,
    GetLines,
    DeleteNetwork,
    DeleteLine,
    DeleteDrawItems,
  } from "../../wailsjs/go/main/App";
  import { getTableLang } from "./common";

  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  let wait = false;

  let networkTable: any = undefined;
  let selectedCount = 0;
  let tab = "";
  const networks: any = [];

  const showNetworks = async () => {
    selectedCount = 0;
    tab= "network";
    wait = true;
    const n = await GetNetworks();
    networks.length = 0;
    if (n) {
      for (const k in n) {
        networks.push(n[k]);
      }
    }
    networkTable = new DataTable("#networkTable", {
      destroy: true,
      data: networks,
      language: getTableLang(),
      select: {
        style: "multi",
      },
      columns: [
        {
          data: "Name",
          title: $_('MapList.Name'),
          width: "20%",
        },
        {
          data: "IP",
          title: $_('MapList.IPAddress'),
          width: "20%",
        },
        {
          data: "Descr",
          title: $_('MapList.Descr'),
          width: "60%",
        },
      ],
    });
    networkTable.on("select", () => {
      selectedCount = networkTable.rows({ selected: true }).count();
    });
    networkTable.on("deselect", () => {
      selectedCount = networkTable.rows({ selected: true }).count();
    });
    wait = false;
  };

  let drawItemTable: any = undefined;
  const drawItems: any = [];

  const drawItemList = [
    $_("DrawItem.Rect"),
    $_("DrawItem.Ellipse"),
    $_("DrawItem.Label"),
    $_("DrawItem.Image"),
    $_("DrawItem.PollingText"),
    $_("DrawItem.PollingGauge"),
    $_("DrawItem.NewGauge"),
    $_("DrawItem.Bar"),
    $_("DrawItem.Line"),
  ];

  const getDrawItemType = (t: number): string => {
    if (t >= 0 && t < drawItemList.length) {
      return drawItemList[t];
    }
    return "Unknown";
  };

  const showDrawItems = async () => {
    tab = "drawItem";
    selectedCount = 0;
    wait = true;
    const d = await GetDrawItems();
    drawItems.length = 0;
    if (d) {
      for (const k in d) {
        const e = d[k];
        drawItems.push({
          ID: e.ID,
          Type: getDrawItemType(e.Type),
          Text: e.Text || e.Path || e.PollingID,
          X: e.X,
          Y: e.Y,
        });
      }
    }
    drawItemTable = new DataTable("#drawItemTable", {
      destroy: true,
      data: drawItems,
      language: getTableLang(),
      select: {
        style: "multi",
      },
      columns: [
        {
          data: "Type",
          title: $_('MapList.Type'),
          width: "20%",
        },
        {
          data: "X",
          title: "X",
          width: "10%",
        },
        {
          data: "Y",
          title: "Y",
          width: "10%",
        },
        {
          data: "Text",
          title: $_('MapList.Text'),
          width: "60%",
        },
      ],
    });
    drawItemTable.on("select", () => {
      selectedCount = drawItemTable.rows({ selected: true }).count();
    });
    drawItemTable.on("deselect", () => {
      selectedCount = drawItemTable.rows({ selected: true }).count();
    });
    wait = false;
  };

  let lineTable: any = undefined;
  const lines: any = [];

  const getNodeName = (id: string, nodes: any, networks: any): string => {
    if (id.startsWith("NET:")) {
      const a = id.split(":");
      if (a.length == 2) {
        const n = networks[a[1]];
        if (n) {
          return n.Name;
        }
      }
    } else {
      const n = nodes[id];
      if (n) {
        return n.Name;
      }
    }
    return "Unknown";
  };

  const showLines = async () => {
    tab = "line";
    selectedCount = 0;
    wait = true;
    const l = await GetLines();
    const nodes = await GetNodes();
    const nets = await GetNetworks();
    lines.length = 0;
    if (l) {
      for (const e of l) {
        lines.push({
          ID: e.ID,
          Src: getNodeName(e.NodeID1, nodes, nets),
          SrcPID: e.PollingID1,
          Dst: getNodeName(e.NodeID2, nodes, nets),
          DstPID: e.PollingID2,
        });
      }
    }
    lineTable = new DataTable("#lineTable", {
      destroy: true,
      data: lines,
      language: getTableLang(),
      select: {
        style: "multi",
      },
      columns: [
        {
          data: "Src",
          title: $_('MapList.Node1'),
          width: "25%",
        },
        {
          data: "SrcPID",
          title: $_('MapList.Polling1'),
          width: "25%",
        },
        {
          data: "Dst",
          title: $_('MapList.Node2'),
          width: "25%",
        },
        {
          data: "DstPID",
          title: $_('MapList.Polling2'),
          width: "25%",
        },
      ],
    });
    lineTable.on("select", () => {
      selectedCount = lineTable.rows({ selected: true }).count();
    });
    lineTable.on("deselect", () => {
      selectedCount = lineTable.rows({ selected: true }).count();
    });
    wait = false;
  };

  const deleteSelected = async () =>{
    selectedCount = 0;
    switch (tab) {
    case "network":{
        const selected = networkTable.rows({ selected: true }).data().pluck("ID");
        if (selected.length < 1) {
          return;
        }
        for(const id of selected.toArray()) {
          await DeleteNetwork(id);
        }
        networkTable.rows({ selected: true }).remove().draw();
        return;
      }
    case "drawItem": {
      const selected = drawItemTable.rows({ selected: true }).data().pluck("ID");
        if (selected.length < 1) {
          return;
        }
        await DeleteDrawItems(selected.toArray());
        drawItemTable.rows({ selected: true }).remove().draw();
        return;
      }
    case "line": {
      const selected = lineTable.rows({ selected: true }).data().pluck("ID");
        if (selected.length < 1) {
          return;
        }
        for(const id of selected.toArray()) {
          await DeleteLine(id);
        }
        lineTable.rows({ selected: true }).remove().draw();
        return;
      }
    }
  }

  const close = () => {
    show = false;
  };

  const onOpen = () => {
    showNetworks();
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  <div class="flex flex-col space-y-4">
    <Tabs tabStyle="underline">
      <TabItem open on:click={showNetworks}>
        <div slot="title" class="flex items-center gap-2">
          {#if wait}
            <Spinner color="red" size="6" />
          {:else}
            <Icon path={icons.mdiLan} size={1} />
          {/if}
          {$_('MapList.Network')}
        </div>
        <table id="networkTable" class="display compact" style="width:99%" />
      </TabItem>
      <TabItem on:click={showDrawItems}>
        <div slot="title" class="flex items-center gap-2">
          {#if wait}
            <Spinner color="red" size="6" />
          {:else}
            <Icon path={icons.mdiDrawing} size={1} />
          {/if}
          {$_('MapList.DrawItem')}
        </div>
        <table id="drawItemTable" class="display compact" style="width:99%" />
      </TabItem>
      <TabItem on:click={showLines}>
        <div slot="title" class="flex items-center gap-2">
          {#if wait}
            <Spinner color="red" size="6" />
          {:else}
            <Icon path={icons.mdiConnection} size={1} />
          {/if}
          {$_('MapList.Line')}
        </div>
        <table id="lineTable" class="display compact" style="width:99%" />
      </TabItem>
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedCount > 0}
        <GradientButton
          shadow
          color="red"
          type="button"
          on:click={deleteSelected}
          size="xs"
        >
          <Icon path={icons.mdiTrashCan} size={1} />
          {$_("NodeList.Delete")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("NodeReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>
