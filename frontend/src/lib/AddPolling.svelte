<script lang="ts">
  import { Modal, Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy,tick,createEventDispatcher } from "svelte";
  import {
    GetPollingTemplates,
    AutoAddPolling,
  } from "../../wailsjs/go/main/App";
  import { renderState, getTableLang } from "./common";
  import Polling  from "./Polling.svelte";

  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  export let nodeID = "";
  const dispatch = createEventDispatcher();

  let tmpTable = undefined;
  let selectedCount = 0;
  let show = false;
  let showEditPolling = false;
  let canAuto = false;
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
      canAuto = false;
      selectedCount = tmpTable.rows({ selected: true }).count();
      if (selectedCount == 1) {
        const autoMode = tmpTable.rows({ selected: true }).data().pluck("AutoMode");
        canAuto = autoMode[0] != "disable";
      }
    });
    tmpTable.on("deselect", () => {
      canAuto = false;
      selectedCount = tmpTable.rows({ selected: true }).count();
    });
  };

  const add = async () => {
    const selected = tmpTable.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedTmplateID = selected[0];
    showEditPolling = true;
  };

  const auto = async () => {
    const selected = tmpTable.rows({ selected: true }).data().pluck("ID");
    const autoMode = tmpTable.rows({ selected: true }).data().pluck("AutoMode");
    if (selected.length != 1) {
      return;
    }
    if (autoMode[0] == "disable") {
      return;
    }
    const r = await AutoAddPolling(nodeID, selected[0]);
    if(r) {
      close();
    }
  };

  const columns = [
    {
      data: "Name",
      title: "名前",
      width: "25%",
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
      data: "AutoMode",
      title: "自動",
      width: "10%",
      render: (d) => d == "disable" ? "" : "Yes",
    },
    {
      data: "Descr",
      title: "説明",
      width: "45%",
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
      {#if selectedCount == 1 && canAuto}
        <Button color="red" type="button" on:click={auto} size="xs">
          <Icon path={icons.mdiBrain} size={1} />
          自動追加
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
    nodeID={nodeID}
    pollingID={""}
    pollingTmpID={selectedTmplateID}
    on:close={close}
  ></Polling>
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
