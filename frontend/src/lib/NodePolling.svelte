<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { Modal, GradientButton } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { GetPollings, DeletePollings } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getLogModeName,
    getTableLang,
    renderPollingType,
  } from "./common";
  import Polling from "./Polling.svelte";
  import AddPolling from "./AddPolling.svelte";
  import PollingReport from "./PollingReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  export let nodeID = "";

  let data: any = [];
  let showEditPolling = false;
  let showAddPolling = false;
  let showPollingReport = false;
  let selectedPolling = "";
  let table: any = undefined;
  let selectedCount = 0;

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#nodePollingTable", {
      destroy: true,
      columns: columns,
      stateSave: true,
      data: data,
      order: [
        [0, "asc"],
        [1, "asc"],
      ],
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
  let pollingTmp: any = undefined;

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
      title: $_("NodePolling.State"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Name",
      title: $_("NodePolling.Name"),
      width: "35%",
    },
    {
      data: "Level",
      title: $_("NodePolling.Level"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Type",
      title: $_("NodePolling.Type"),
      width: "10%",
      render: renderPollingType,
    },
    {
      data: "LogMode",
      title: $_("NodePolling.LogMode"),
      width: "10%",
      render: getLogModeName,
      searchable: false,
    },
    {
      data: "FailAction",
      title: $_('NodePolling.FailAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
      searchable: false,
    },
    {
      data: "RepairAction",
      title: $_('NodePolling.RepairAction'),
      width: "5%",
      render: (a:any)=> a != "" ? "Action" : "",
      searchable: false,
    },
    {
      data: "LastTime",
      title: $_("NodePolling.LastTime"),
      width: "15%",
      render: renderTime,
      searchable: false,
    },
  ];

  const onOpen = () => {
    refresh();
  };

  const close = () => {
    show = false;
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  <div class="flex flex-col">
    <div class="m-5 grow">
      <table id="nodePollingTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={add}
        size="xs"
      >
        <Icon path={icons.mdiPlus} size={1} />
        {$_("NodePolling.Add")}
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
          {$_("NodePolling.Edit")}
        </GradientButton>
        <GradientButton
          shadow
          color="lime"
          type="button"
          on:click={copy}
          size="xs"
        >
          <Icon path={icons.mdiContentCopy} size={1} />
          {$_("NodePolling.Copy")}
        </GradientButton>
        <GradientButton
          shadow
          color="green"
          type="button"
          on:click={report}
          size="xs"
        >
          <Icon path={icons.mdiChartBar} size={1} />
          {$_("NodePolling.Report")}
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
          {$_("NodePolling.Delete")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={refresh}
        size="xs"
      >
        <Icon path={icons.mdiRecycle} size={1} />
        {$_("NodePolling.Reload")}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("NodePolling.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Polling
  bind:show={showEditPolling}
  nodeID=""
  pollingID={selectedPolling}
  on:close={(e) => {
    refresh();
  }}
/>

<PollingReport bind:show={showPollingReport} id={selectedPolling} />

<AddPolling
  bind:show={showAddPolling}
  {nodeID}
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
