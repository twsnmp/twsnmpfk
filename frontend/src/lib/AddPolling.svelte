<script lang="ts">
  import { Modal, Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy, tick, createEventDispatcher } from "svelte";
  import { GetPollingTemplates } from "../../wailsjs/go/main/App";
  import { renderState, getTableLang } from "./common";
  import Polling from "./Polling.svelte";

  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  export let nodeID = "";
  const dispatch = createEventDispatcher();

  let tmpTable = undefined;
  let selectedCount = 0;
  let show = false;
  let showEditPolling = false;
  let selectedTmplateID = "";

  const showTable = async () => {
    await tick();
    let order = [[2, "asc"]];
    if (tmpTable) {
      order = tmpTable.order();
      tmpTable.destroy();
      tmpTable = undefined;
    }
    selectedCount = 0;
    tmpTable = new DataTable("#tmpTable", {
      columns: columns,
      data: await GetPollingTemplates(),
      order: order,
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
    selectedTmplateID = selected[0];
    show = false;
    showEditPolling = true;
  };

  const columns = [
    {
      data: "Name",
      title: "名前",
      width: "30%",
    },
    {
      data: "Level",
      title: "レベル",
      width: "10%",
      render: renderState,
    },
    {
      data: "Type",
      title: "種別",
      width: "10%",
    },
    {
      data: "Descr",
      title: "説明",
      width: "50%",
    },
  ];

  onMount(() => {
    show = true;
    showTable();
  });

  onDestroy(() => {
    if (tmpTable) {
      tmpTable.destroy();
      tmpTable = undefined;
    }
  });
  const close = () => {
    show = false;
    showEditPolling = false;
    dispatch("close", {});
  };
</script>

<Modal bind:open={show} size="xl" permanent class="w-full" on:on:close={close}>
  <div class="flex flex-col">
    <div class="m-5 grow">
      <table id="tmpTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedCount == 1}
        <Button color="blue" type="button" on:click={add} size="xs">
          <Icon path={icons.mdiPlus} size={1} />
          追加
        </Button>
      {/if}
      <Button type="button" color="alternative" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </div>
</Modal>

{#if showEditPolling}
  <Polling
    {nodeID}
    pollingID={""}
    pollingTmpID={selectedTmplateID}
    on:close={close}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
