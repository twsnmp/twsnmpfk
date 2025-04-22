<script lang="ts">
  import {
    Modal,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import { FindNeighborNetworksAndLines,GetNode,GetPolling,GetNetwork,UpdateLine } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import Network from "./Network.svelte";
  import type { datastore } from "wailsjs/go/models";

  export let show: boolean = false;
  export let id: string = "";

  let networkData: any = [];
  let lineData: any = [];
  let networkTable: any = undefined;
  let lineTable: any = undefined;
  let networkSelectedCount = 0;
  let lineSelectedCount = 0;
  let wait = false;
  let resp :any = undefined;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    wait = true;
    resp = await FindNeighborNetworksAndLines(id);
    wait = false;
    networkData = [];
    lineData = [];
    if (resp && resp.Networks) {
      for(let i =0; i < resp.Networks.length;i++) {
        const n = resp.Networks[i];
        networkData.push({
          Index: i,
          Name: n.Name,
          IP: n.IP,
          SystemID:n.SystemID,
          Descr: n.Descr,
        })
      }
    }
    if (resp && resp.Lines) {
      for(let i =0; i < resp.Lines.length;i++) {
        const l = resp.Lines[i];
        const n1 = l.NodeID1.startsWith("NET:") ? await GetNetwork(l.NodeID1) : await GetNode(l.NodeID1);
      const n2 = l.NodeID2.startsWith("NET:") ? await GetNetwork(l.NodeID2) : await GetNode(l.NodeID2);
      let p1 = l.PollingID1;
      let p2 = l.PollingID2;
      if (!l.NodeID1.startsWith("NET:")) {
        const p = await GetPolling(p1)
        if (p) {
          p1 = p.Name;
        }
      } else if (n1) {
        for (const p of (n1 as datastore.NetworkEnt).Ports) {
          if (p.ID == p1) {
            p1 = p.Name;
            break
          }
        }
      }
      if (!l.NodeID2.startsWith("NET:")) {
        const p = await GetPolling(p2)
        if (p) {
          p2 = p.Name;
        }
      } else if (n2) {
        for (const p of (n2 as datastore.NetworkEnt).Ports) {
          if (p.ID == p2) {
            p2 = p.Name;
            break
          }
        }
      }
      lineData.push({
        Index: i,
        Node1: n1 ? n1.Name : "",
        Node2: n2 ? n2.Name : "",
        Polling1: p1,
        Polling2: p2,
      })
      }
    }
    
    wait = false;
    showNetworkTable();
    showLineTable();
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };


  const networkColumns = [
    {
      data: "Name",
      title: $_("NodeList.Name"),
      width: "30%",
    },
    {
      data: "IP",
      title: "IP",
      width: "10%",
    },
    {
      data: "SystemID",
      title: "ID",
      width: "10%",
    },
    {
      data: "Descr",
      title: $_('NeighborNetworksAndLines.Descr'),
      width: "50%",
    },
  ];

  const showNetworkTable = async () => {
    await tick();
    networkSelectedCount = 0;
    networkTable = new DataTable("#networkTable", {
      destroy: true,
      columns: networkColumns,
      data: networkData,
      paging: false,
      searching: false,
      ordering: false,
      info: false,
      scrollY: "30vh",
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    networkTable.on("select", () => {
      networkSelectedCount = networkTable.rows({ selected: true }).count();
    });
    networkTable.on("deselect", () => {
      networkSelectedCount = networkTable.rows({ selected: true }).count();
    });
  };

  const lineColumns = [
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

  const showLineTable = async () => {
    await tick();
    lineSelectedCount = 0;
    lineTable = new DataTable("#lineTable", {
      destroy: true,
      columns: lineColumns,
      data: lineData,
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
    lineTable.on("select", () => {
      lineSelectedCount = lineTable.rows({ selected: true }).count();
    });
    lineTable.on("deselect", () => {
      lineSelectedCount = lineTable.rows({ selected: true }).count();
    });
  };

  const addNetwork = () => {
    if (networkSelectedCount != 1) {
      return;
    }
    const sels = networkTable.rows({ selected: true }).data();
    const i = networkData.indexOf(sels[0]);
    if (i < 0 || i >= resp.Networks.length) {
      return;
    }
    show = false;
    dispatch("addNetwork", resp.Networks[i]);
  };

  const connectLine = async () => {
    if (lineSelectedCount != 1) {
      return;
    }
    const sels = lineTable.rows({ selected: true }).data();
    if (sels.length != 1) {
      return;
    }
    const i = lineData.indexOf(sels[0]);
    const j = sels[0].Index;
    if (i < 0 || j >= resp.Lines.length) {
      return;
    }
    const l = resp.Lines[j];
    await UpdateLine(l);
    lineData.splice(i, 1);
    showLineTable();
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if wait}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_('NeighborNetworksAndLines.Title')}
      </h3>
      <div class="m-5 grow">
        <table id="networkTable" class="display compact" style="width:99%" />
      </div>
      <div class="m-5 grow">
        <table id="lineTable" class="display compact" style="width:99%" />
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if networkSelectedCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={addNetwork}
            size="xs"
          >
            <Icon path={icons.mdiPlus} size={1} />
            {$_('NeighborNetworksAndLines.AddNetwork')}
          </GradientButton>
        {/if}
        {#if lineSelectedCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={connectLine}
            size="xs"
          >
            <Icon path={icons.mdiLanConnect} size={1} />
            {$_('NeighborNetworksAndLines.ConnectLine')}
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
