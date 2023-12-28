<script lang="ts">
  import { Modal, GradientButton } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { GetPollings, DeletePollings } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getLogModeName,
    getTableLang,
  } from "./common";
  import Polling from "./Polling.svelte";
  import AddPolling from "./AddPolling.svelte";
  import PollingReport from "./PollingReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  export let nodeID = "";
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  let data = [];
  let showEditPolling = false;
  let showAddPolling = false;
  let showPollingReport = false;
  let selectedPolling = "";
  let table = undefined;
  let selectedCount = 0;

  const showTable = () => {
    let order = [
      [0, "asc"],
      [1, "asc"],
    ];
    if (table && DataTable.isDataTable("#nodePollingTable")) {
      order = table.order();
      table.clear();
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#nodePollingTable", {
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
    data = await GetPollings(nodeID);
    showTable();
  };

  const edit = () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedPolling = selected[0];
    showEditPolling = true;
  };

  let showCopyPolling = false;
  let pollingTmp = undefined;

  const copy = () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    pollingTmp = selected[0];
    pollingTmp.ID = "";
    showCopyPolling = true;
  };

  const add = () => {
    selectedPolling = "";
    showAddPolling = true;
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
    refresh();
  };

  const columns = [
    {
      data: "State",
      title: $_('NodePolling.State'),
      width: "15%",
      render: renderState,
    },
    {
      data: "Name",
      title: $_('NodePolling.Name'),
      width: "35%",
    },
    {
      data: "Level",
      title: $_('NodePolling.Level'),
      width: "15%",
      render: renderState,
    },
    {
      data: "Type",
      title: $_('NodePolling.Type'),
      width: "10%",
    },
    {
      data: "LogMode",
      title: $_('NodePolling.LogMode'),
      width: "10%",
      render: getLogModeName,
    },
    {
      data: "LastTime",
      title: $_('NodePolling.LastTime'),
      width: "15%",
      render: renderTime,
    },
  ];

  onMount(() => {
    show = true;
    refresh();
  });

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal bind:open={show} size="xl" permanent class="w-full" on:on:close={close}>
  <div class="flex flex-col">
    <div class="m-5 grow">
      <table id="nodePollingTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow color="blue" type="button" on:click={add} size="xs">
        <Icon path={icons.mdiPlus} size={1} />
        { $_('NodePolling.Add') }
      </GradientButton>
      {#if selectedCount == 1}
        <GradientButton shadow color="blue" type="button" on:click={edit} size="xs">
          <Icon path={icons.mdiPencil} size={1} />
          { $_('NodePolling.Edit') }
        </GradientButton>
        <GradientButton shadow color="lime" type="button" on:click={copy} size="xs">
          <Icon path={icons.mdiContentCopy} size={1} />
          { $_('NodePolling.Copy') }
        </GradientButton>
        <GradientButton shadow color="green" type="button" on:click={report} size="xs">
          <Icon path={icons.mdiChartBar} size={1} />
          { $_('NodePolling.Report') }
        </GradientButton>
      {/if}
      {#if selectedCount > 0}
        <GradientButton shadow color="red" type="button" on:click={deletePollings} size="xs">
          <Icon path={icons.mdiTrashCan} size={1} />
          { $_('NodePolling.Delete') }
        </GradientButton>
      {/if}
      <GradientButton shadow type="button" color="teal" on:click={refresh} size="xs">
        <Icon path={icons.mdiRecycle} size={1} />
        { $_('NodePolling.Reload') }
      </GradientButton>
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('NodePolling.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

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

{#if showPollingReport}
  <PollingReport
    id={selectedPolling}
    on:close={(e) => {
      showPollingReport = false;
    }}
  />
{/if}

{#if showAddPolling}
  <AddPolling
    {nodeID}
    on:close={(e) => {
      showAddPolling = false;
      refresh();
    }}
  />
{/if}

{#if showCopyPolling}
  <Polling
    nodeID=""
    pollingID= ""
    {pollingTmp}
    on:close={(e) => {
      showCopyPolling = false;
      refresh();
    }}
  />
{/if}


<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
