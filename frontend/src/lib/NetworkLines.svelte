<script lang="ts">
  import {
    Modal,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import { GetLinesByNode, DeleteLine,GetNode,GetNetwork, GetPolling } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import type { datastore } from "wailsjs/go/models";

  export let show: boolean = false;
  export let id: string = ""; // Network ID

  let data: any = [];
  let table: any = undefined;
  let selectedCount = 0;
  let wait = true;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    wait = true;
    const lines = await GetLinesByNode("NET:"+id);
    data = [];
    for(const l of lines) {
      const n1 = l.NodeID1.startsWith("NET:") ? await GetNetwork(l.NodeID1) : await GetNode(l.NodeID1);
      const n2 = l.NodeID2.startsWith("NET:") ? await GetNetwork(l.NodeID2) : await GetNode(l.NodeID2);
      let p1 = l.PollingID1;
      let p2 = l.PollingID2;
      if (!l.NodeID1.startsWith("NET:")) {
        const p = await GetPolling(p1)
        if (p) {
          p1 = p.Name;
        }
      } else {
        const n = n1 as datastore.NetworkEnt
        const port = n.Ports.find((e) => e.ID == p1);
        if (port) {
          p1 = port.Name
        }
      }
      if (!l.NodeID2.startsWith("NET:")) {
        const p = await GetPolling(p2)
        if (p) {
          p2 = p.Name;
        }
      } else {
        const n = n2 as datastore.NetworkEnt
        const port = n.Ports.find((e) => e.ID == p2);
        if (port) {
          p2 = port.Name
        }
      }
      data.push({
        ID: l.ID,
        Node1: n1 ? n1.Name : "",
        Node2: n2 ? n2.Name : "",
        Polling1: p1,
        Polling2: p2,
      })
    }
    wait = false;
    showTable();
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const columns = [
    {
      data: "Node1",
      title: $_('Line.Node1'),
      width: "20%",
    },
    {
      data: "Polling1",
      title: $_('Line.Polling1'),
      width: "30%",
    },
    {
      data: "Node2",
      title: $_('Line.Node2'),
      width: "20%",
    },
    {
      data: "Polling2",
      title: $_('Line.Polling2'),
      width: "30%",
    },
  ];

  const showTable = async () => {
    await tick();
    selectedCount = 0;
    table = new DataTable("#lineTable", {
      destroy: true,
      stateSave: true,
      columns: columns,
      data: data,
      paging: false,
      searching: false,
      ordering: false,
      info: false,
      scrollY: "40vh",
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  };
  //
  const lineDelete =  async () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = data.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    await DeleteLine(data[i].ID);
    data.splice(i, 1);
    showTable();
  };
  const lineEdit = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = data.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    show = false;
    dispatch("editLine", data[i].ID);
  }
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if wait}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_('NetworkLine.EditLine')}
      </h3>
      <div class="m-5 grow">
        <table id="lineTable" class="display compact" style="width:99%" />
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if selectedCount > 0}
           <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={lineEdit}
            size="xs"
          >
            <Icon path={icons.mdiPencil} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={lineDelete}
            size="xs"
          >
            <Icon path={icons.mdiTrashCan} size={1} />
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          type="button"
          color="teal"
          on:click={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_('Config.Close')}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>
