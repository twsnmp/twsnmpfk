<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy } from "svelte";
  import {
    GetNodes,
    DeleteNodes,
    ExportNodes,
    CheckPolling,
  } from "../../wailsjs/go/main/App";
  import { getTableLang, renderState, renderIP } from "./common";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let data = [];
  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let selectedNode = "";
  let table = undefined;
  let selectedCount = 0;

  const showTable = () => {
    let order = [
      [0, "asc"],
      [2, "asc"],
    ];
    if (table) {
      order = table.order();
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#table", {
      columns: columns,
      data: data,
      order: order,
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

  const deleteNodes = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeleteNodes(selected.toArray());
    refresh();
  };

  const check = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    selected.array.forEach((n) => {
      CheckPolling(n);
    });
    refresh();
  };

  const columns = [
    {
      data: "State",
      title: $_('NodeList.State'),
      width: "10%",
      render: renderState,
    },
    {
      data: "Name",
      title: $_('NodeList.Name'),
      width: "15%",
    },
    {
      data: "IP",
      title: $_('NodeList.IPAddress'),
      width: "10%",
      render: renderIP,
    },
    {
      data: "MAC",
      title: $_('NodeList.MACAddress'),
      width: "30%",
    },
    {
      data: "Descr",
      title: $_('NodeList.Descr'),
      width: "30%",
    },
  ];

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    if (table) {
      table.destroy();
      table = undefined;
    }
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
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <Button color="blue" type="button" on:click={edit} size="xs">
        <Icon path={icons.mdiPencil} size={1} />
        { $_('NodeList.Edit') }
      </Button>
      <Button color="blue" type="button" on:click={polling} size="xs">
        <Icon path={icons.mdiLanCheck} size={1} />
        { $_('NodeList.Polling') }
      </Button>
      <Button color="green" type="button" on:click={report} size="xs">
        <Icon path={icons.mdiChartBar} size={1} />
        { $_('NodeList.Report') }
      </Button>
    {/if}
    {#if selectedCount > 0}
      <Button color="red" type="button" on:click={deleteNodes} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        { $_('NodeList.Delete') }
      </Button>
      <Button color="blue" type="button" on:click={check} size="xs">
        <Icon path={icons.mdiCheck} size={1} />
        { $_('NodeList.ReCheck') }
      </Button>
    {/if}
    <Button color="blue" type="button" on:click={checkAll} size="xs">
      <Icon path={icons.mdiCheckAll} size={1} />
      { $_('NodeList.CheckAll') }
    </Button>
    <Button color="blue" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </Button>
    <Button color="blue" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('NodeList.Reload') }
    </Button>
  </div>
</div>

{#if showEditNode}
  <Node
    nodeID={selectedNode}
    on:close={(e) => {
      showEditNode = false;
      refresh();
    }}
  />
{/if}

{#if showNodeReport}
  <NodeReport
    id={selectedNode}
    on:close={(e) => {
      showNodeReport = false;
    }}
  />
{/if}

{#if showPolling}
  <NodePolling
    nodeID={selectedNode}
    on:close={(e) => {
      showPolling = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
