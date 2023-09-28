<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy } from "svelte";
  import {
    GetNodes,
    GetPollings,
    DeletePollings,
    ExportPollings,
    GetDefaultPolling,
  } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getLogModeName,
    getTableLang,
  } from "./common";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import PollingReport from "./PollingReport.svelte";
  import { _ } from "svelte-i18n";

  let data = [];
  let nodes = {};
  let showEditPolling = false;
  let showPollingReport = false;
  let selectedPolling = "";
  let table = undefined;
  let selectedCount = 0;

  const showTable = () => {
    let order = [
      [0, "asc"],
      [1, "asc"],
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
    data = [];
    nodes = await GetNodes();
    data = await GetPollings("");
    showTable();
  };

  let showAddPolling = false;
  let pollingTmp = undefined;

  const add = async () => {
    pollingTmp = await GetDefaultPolling("");
    showAddPolling = true;
  }

  const edit = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showEditPolling = true;
  };

  const copy = () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    pollingTmp = selected[0];
    pollingTmp.ID = "";
    showAddPolling = true;
  };

  const report = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showPollingReport = true;
  }

  const deletePollings = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeletePollings(selected.toArray());
    refresh();
  };

  const columns = [
    {
      data: "State",
      title: $_('PollingList.State'),
      width: "10%",
      render: renderState,
    },
    {
      data: "NodeID",
      title: $_('PollingList.Node'),
      width: "15%",
      render: (id) => nodes[id].Name,
    },
    {
      data: "Name",
      title: $_('PollingList.Name'),
      width: "25%",
    },
    {
      data: "Level",
      title: $_('PollingList.Level'),
      width: "10%",
      render: renderState,
    },
    {
      data: "Type",
      title: $_('PollingList.Type'),
      width: "8%",
    },
    {
      data: "LogMode",
      title: $_('PollingList.LogMode'),
      width: "7%",
      render: getLogModeName,
    },
    {
      data: "LastTime",
      title: $_('PollingList.LastTime'),
      width: "15%",
      render: renderTime,
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

  const saveCSV = () => {
    ExportPollings("csv");
  };

  const saveExcel = () => {
    ExportPollings("excel");
  };
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <Button color="blue" type="button" on:click={add} size="xs">
      <Icon path={icons.mdiPlus} size={1} />
      { $_('PollingList.Add') }
    </Button>
    {#if selectedCount == 1}
      <Button color="blue" type="button" on:click={edit} size="xs">
        <Icon path={icons.mdiPencil} size={1} />
        { $_('PollingList.Edit') }
      </Button>
      <Button color="blue" type="button" on:click={copy} size="xs">
        <Icon path={icons.mdiContentCopy} size={1} />
        { $_('PollingList.Copy') }
      </Button>
      <Button color="green" type="button" on:click={report} size="xs">
        <Icon path={icons.mdiChartBar} size={1} />
        { $_('PollingList.Report') }
      </Button>
    {/if}
    {#if selectedCount > 0}
      <Button color="red" type="button" on:click={deletePollings} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        { $_('PollingList.Delete') }
      </Button>
    {/if}
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
      { $_('PollingList.Reload') }
    </Button>
  </div>
</div>

{#if showEditPolling}
  <Polling
    nodeID=""
    pollingID={selectedPolling}
    on:close={(e) => {
      showEditPolling = false;
      refresh();
    }}
  />
{/if}

{#if showAddPolling}
  <Polling
    nodeID=""
    pollingID = ""
    {pollingTmp}
    on:close={(e) => {
      showAddPolling = false;
      refresh();
    }}
  />
{/if}

{#if showPollingReport}
  <PollingReport
    id={selectedPolling}
    on:close={(e) => {
      showPollingReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
