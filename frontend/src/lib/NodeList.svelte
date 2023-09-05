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
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  let data = [];
  let showEditNode = false;
  let showNodeReport = false;
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
      title: "状態",
      width: "10%",
      render: renderState,
    },
    {
      data: "Name",
      title: "名前",
      width: "15%",
    },
    {
      data: "IP",
      title: "IPアドレス",
      width: "10%",
      render: renderIP,
    },
    {
      data: "MAC",
      title: "MACアドレス",
      width: "30%",
    },
    {
      data: "Descr",
      title: "説明",
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
        編集
      </Button>
      <Button color="green" type="button" on:click={report} size="xs">
        <Icon path={icons.mdiChartBar} size={1} />
        レポート
      </Button>
    {/if}
    {#if selectedCount > 0}
      <Button color="red" type="button" on:click={deleteNodes} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        削除
      </Button>
      <Button color="blue" type="button" on:click={check} size="xs">
        <Icon path={icons.mdiCheck} size={1} />
        再確認
      </Button>
    {/if}
    <Button color="blue" type="button" on:click={checkAll} size="xs">
      <Icon path={icons.mdiCheckAll} size={1} />
      すべて再確認
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
      更新
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

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
