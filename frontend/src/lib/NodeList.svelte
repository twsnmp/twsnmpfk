<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton, Dropdown, DropdownItem } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetNodes,
    DeleteNodes,
    ExportNodes,
    CheckPolling,
    WakeOnLan,
  } from "../../wailsjs/go/main/App";
  import { getTableLang, renderNodeState, renderIP } from "./common";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import Ping from "./Ping.svelte";
  import MIBBrowser from "./MIBBrowser.svelte";
  import GNMITool from "./GNMITool.svelte";
  import MapList from "./MapList.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let data: any = [];
  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let showMapList = false;
  let selectedNode = "";
  let table: any = undefined;
  let selectedCount = 0;

  let showPing: boolean = false;
  let showMibBr: boolean = false;
  let showGNMITool: boolean = false;
  let actionOpen: boolean = false;

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#nodeListTable", {
      destroy: true,
      columns: columns,
      data: data,
      stateSave: true,
      order: [
        [0, "asc"],
        [2, "asc"],
      ],
      pageLength: window.innerHeight > 800 ? 25 : 10,
      language: getTableLang(),
      select: {
        style: "multi",
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
    const nodes = await GetNodes();
    data = [];
    for (const k in nodes) {
      data.push(nodes[k]);
    }
    showTable();
  };

  const edit = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showEditNode = true;
  };

  const report = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showNodeReport = true;
  };

  const polling = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showPolling = true;
  };

  const ping = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showPing = true;
  };

  const MIBBr = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showMibBr = true;
  };

  const gNMITool = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    showGNMITool = true;
  };

  const doWakeOnLan = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedNode = selected[0];
    WakeOnLan(selectedNode);
    actionOpen = false;
  };

  const deleteNodes = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeleteNodes(selected.toArray());
    table.rows({ selected: true }).remove().draw();
  };

  const check = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    selected.array.forEach(async (n: any) => {
      await CheckPolling(n);
    });
    refresh();
  };

  const columns = [
    {
      data: "State",
      title: $_("NodeList.State"),
      width: "10%",
      render: renderNodeState,
    },
    {
      data: "Name",
      title: $_("NodeList.Name"),
      width: "15%",
    },
    {
      data: "IP",
      title: $_("NodeList.IPAddress"),
      width: "10%",
      render: renderIP,
    },
    {
      data: "MAC",
      title: $_("NodeList.MACAddress"),
      width: "10%",
    },
    {
      data: "Vendor",
      title: $_("NodeList.Vendor"),
      width: "15%",
    },
    {
      data: "Descr",
      title: $_("NodeList.Descr"),
      width: "30%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const checkAll = () => {
    CheckPolling("all");
  };

  const saveCSV = () => {
    ExportNodes("csv");
  };

  const saveExcel = () => {
    ExportNodes("excel");
  };
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="nodeListTable" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={edit}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("NodeList.Edit")}
      </GradientButton>
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={polling}
        size="xs"
      >
        <Icon path={icons.mdiLanCheck} size={1} />
        {$_("NodeList.Polling")}
      </GradientButton>
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={report}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("NodeList.Report")}
      </GradientButton>
      <GradientButton
        >{$_("NodeList.Action")}<Icon
          path={icons.mdiChevronDown}
          size={1}
        /></GradientButton
      >
      <Dropdown bind:open={actionOpen}>
        <DropdownItem on:click={ping}>PING</DropdownItem>
        <DropdownItem on:click={MIBBr}>{$_("Map.MIBBrowser")}</DropdownItem>
        <DropdownItem on:click={gNMITool}
          >{$_("GNMITool.gNMITool")}</DropdownItem
        >
        <DropdownItem on:click={doWakeOnLan}>Wake On Lan</DropdownItem>
      </Dropdown>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={deleteNodes}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("NodeList.Delete")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        on:click={check}
        size="xs"
      >
        <Icon path={icons.mdiCheck} size={1} />
        {$_("NodeList.ReCheck")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="green"
      type="button"
      on:click={()=> {
        showMapList = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiListBox} size={1} />
      {$_('NodeList.MapItems')}
    </GradientButton>
    <GradientButton
      shadow
      color="teal"
      type="button"
      on:click={checkAll}
      size="xs"
    >
      <Icon path={icons.mdiCheckAll} size={1} />
      {$_("NodeList.CheckAll")}
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveCSV}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveExcel}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("NodeList.Reload")}
    </GradientButton>
  </div>
</div>

<Node
  bind:show={showEditNode}
  nodeID={selectedNode}
  on:close={(e) => {
    refresh();
  }}
/>

<NodeReport bind:show={showNodeReport} id={selectedNode} />

<NodePolling bind:show={showPolling} nodeID={selectedNode} />

<Ping bind:show={showPing} nodeID={selectedNode} />

<MIBBrowser bind:show={showMibBr} nodeID={selectedNode} />

<GNMITool bind:show={showGNMITool} nodeID={selectedNode} />

<MapList bind:show={showMapList} />

