<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetNodes,
    DeleteNodes,
    ExportNodes,
    CheckPolling,
  } from "../../wailsjs/go/main/App";
  import { getTableLang, renderNodeState, renderIP } from "./common";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let data :any = [];
  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let selectedNode = "";
  let table :any = undefined;
  let selectedCount = 0;

  const showTable = () => {
    let order = [
      [0, "asc"],
      [2, "asc"],
    ];
    if (table && DataTable.isDataTable("#table")) {
      order = table.order();
      table.clear();
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
    selected.array.forEach((n:any) => {
      CheckPolling(n);
    });
    refresh();
  };

  const columns = [
    {
      data: "State",
      title: $_('NodeList.State'),
      width: "10%",
      render: renderNodeState,
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
      width: "10%",
    },
    {
      data: "Vendor",
      title: $_('NodeList.Vendor'),
      width: "15%",
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
      <GradientButton shadow color="blue" type="button" on:click={edit} size="xs">
        <Icon path={icons.mdiPencil} size={1} />
        { $_('NodeList.Edit') }
      </GradientButton>
      <GradientButton shadow color="blue" type="button" on:click={polling} size="xs">
        <Icon path={icons.mdiLanCheck} size={1} />
        { $_('NodeList.Polling') }
      </GradientButton>
      <GradientButton shadow color="green" type="button" on:click={report} size="xs">
        <Icon path={icons.mdiChartBar} size={1} />
        { $_('NodeList.Report') }
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton shadow color="red" type="button" on:click={deleteNodes} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        { $_('NodeList.Delete') }
      </GradientButton>
      <GradientButton shadow color="teal" type="button" on:click={check} size="xs">
        <Icon path={icons.mdiCheck} size={1} />
        { $_('NodeList.ReCheck') }
      </GradientButton>
    {/if}
    <GradientButton shadow color="teal" type="button" on:click={checkAll} size="xs">
      <Icon path={icons.mdiCheckAll} size={1} />
      { $_('NodeList.CheckAll') }
    </GradientButton>
    <GradientButton shadow color="lime" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton shadow color="lime" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton shadow type="button" color="teal" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('NodeList.Reload') }
    </GradientButton>
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
