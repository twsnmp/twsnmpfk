<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { Modal, GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { tick, createEventDispatcher } from "svelte";
  import { GetPollingTemplates } from "../../wailsjs/go/main/App";
  import { getTableLang,renderPollingType } from "./common";
  import Polling from "./Polling.svelte";

  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';

  export let nodeID = "";
  export let show = false;
  const dispatch = createEventDispatcher();

  let tmpTable :any = undefined;
  let selectedCount = 0;
  let showEditPolling = false;
  let selectedTemplateID = 0;

  const showTable = async () => {
    await tick();
    selectedCount = 0;
    tmpTable = new DataTable("#tmpTable", {
      columns: columns,
      data: await GetPollingTemplates(),
      stateSave: true,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    tmpTable.on("select", () => {
      selectedCount = tmpTable.rows({ selected: true }).count();
    });
    tmpTable.on("deselect", () => {
      selectedCount = tmpTable.rows({ selected: true }).count();
    });
  };

  const add = async () => {
    const selected = tmpTable.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedTemplateID = Number(selected[0]);
    show = false;
    showEditPolling = true;
  };

  const columns = [
    {
      data: "ID",
      title: "ID",
      width: "5%",
      searchable: false,
    },
    {
      data: "Name",
      title: $_('AppPolling.Name'),
      width: "30%",
    },
    {
      data: "Type",
      title: $_('AddPolling.Type'),
      width: "15%",
      render: renderPollingType,
    },
    {
      data: "Mode",
      title: $_('AddPolling.Mode'),
      width: "15%",
      searchable: false,
    },
    {
      data: "Descr",
      title: $_('AddPolling.Descr'),
      width: "40%",
    },
  ];

  const onOpen =() => {
    showTable();
  };

  const close = () => {
    show = false;
    showEditPolling = false;
    dispatch("close", {});
  };
</script>

<Modal bind:open={show} size="xl" dismissable={false} class="w-full" on:open={onOpen}>
  <div class="flex flex-col">
    <div class="m-5 grow">
      <table id="tmpTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedCount == 1}
        <GradientButton shadow color="blue" type="button" on:click={add} size="xs">
          <Icon path={icons.mdiPlus} size={1} />
          { $_('AddPolling.Add') }
        </GradientButton>
      {/if}
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('AddPolling.Cancel') }
      </GradientButton>
    </div>
  </div>
</Modal>

<Polling
  bind:show={showEditPolling}
  {nodeID}
  pollingID={""}
  pollingTmpID={selectedTemplateID}
  on:close={close}
/>
