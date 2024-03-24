<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
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
  import AddPolling from "./AddPolling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import PollingReport from "./PollingReport.svelte";
  import { _ } from "svelte-i18n";

  let data: any = [];
  let nodes: any = {};
  let showEditPolling = false;
  let showPollingReport = false;
  let selectedPolling = "";
  let table: any = undefined;
  let selectedCount = 0;

  const showTable = () => {
    let order = [
      [0, "asc"],
      [1, "asc"],
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
    data = [];
    nodes = await GetNodes();
    data = await GetPollings("");
    showTable();
  };

  let showCopyPolling = false;
  let showAddPolling = false;
  let pollingTmp: any = undefined;

  const add = async () => {
    pollingTmp = await GetDefaultPolling("");
    showAddPolling = true;
  };

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
    showCopyPolling = true;
  };

  const report = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showPollingReport = true;
  };

  const deletePollings = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeletePollings(selected.toArray());
    table.rows({ selected: true }).remove().draw();
  };

  const columns = [
    {
      data: "State",
      title: $_("PollingList.State"),
      width: "8%",
      render: renderState,
    },
    {
      data: "NodeID",
      title: $_("PollingList.Node"),
      width: "14%",
      render: (id: any) => nodes[id] ? nodes[id].Name : "???",
    },
    {
      data: "Name",
      title: $_("PollingList.Name"),
      width: "23%",
    },
    {
      data: "Level",
      title: $_("PollingList.Level"),
      width: "8%",
      render: renderState,
    },
    {
      data: "Type",
      title: $_("PollingList.Type"),
      width: "8%",
    },
    {
      data: "LogMode",
      title: $_("PollingList.LogMode"),
      width: "7%",
      render: getLogModeName,
    },
    {
      data: "FailAction",
      title: $_('PollingList.FailAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
    },
    {
      data: "RepairAction",
      title: $_('PollingList.RepairAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
    },
    {
      data: "LastTime",
      title: $_("PollingList.LastTime"),
      width: "12%",
      render: renderTime,
    },
  ];

  onMount(() => {
    refresh();
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
    <GradientButton shadow color="blue" type="button" on:click={add} size="xs">
      <Icon path={icons.mdiPlus} size={1} />
      {$_("PollingList.Add")}
    </GradientButton>
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={edit}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("PollingList.Edit")}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        on:click={copy}
        size="xs"
      >
        <Icon path={icons.mdiContentCopy} size={1} />
        {$_("PollingList.Copy")}
      </GradientButton>
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={report}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("PollingList.Report")}
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={deletePollings}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("PollingList.Delete")}
      </GradientButton>
    {/if}
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
      {$_("PollingList.Reload")}
    </GradientButton>
  </div>
</div>

<Polling
  bind:show={showEditPolling}
  nodeID=""
  pollingID={selectedPolling}
  on:close={(e) => {
    refresh();
  }}
/>

<Polling
  bind:show={showCopyPolling}
  nodeID=""
  pollingID=""
  {pollingTmp}
  on:close={(e) => {
    refresh();
  }}
/>

<AddPolling
  bind:show={showAddPolling}
  nodeID=""
  on:close={(e) => {
    refresh();
  }}
/>

<PollingReport bind:show={showPollingReport} id={selectedPolling} />
