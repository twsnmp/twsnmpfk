<script lang="ts">
  import {
    Modal,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import {
    FindNeighborNetworksAndLines,
    FindNeighborNetworksAndLinesWithAI,
    FindNodeConnection,
    ConnectLines,
    GetNode,
    GetPolling,
    GetNetwork,
    UpdateLine,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import type { datastore } from "wailsjs/go/models";

  export let show: boolean = false;
  export let id: string = "";

  let networkData: any = [];
  let lineData: any = [];
  let rawLines: any = [];
  let networkTable: any = undefined;
  let lineTable: any = undefined;
  let networkSelectedCount = 0;
  let lineSelectedCount = 0;
  let wait = false;
  let resp: any = undefined;
  let isNodeMode = false;

  const dispatch = createEventDispatcher();

  const parseLines = async (lines: any[]) => {
    rawLines = lines || [];
    const data = [];
    for (let i = 0; i < rawLines.length; i++) {
      const l = rawLines[i];
      const n1 = l.NodeID1.startsWith("NET:")
        ? await GetNetwork(l.NodeID1.replace("NET:", ""))
        : await GetNode(l.NodeID1);
      const n2 = l.NodeID2.startsWith("NET:")
        ? await GetNetwork(l.NodeID2.replace("NET:", ""))
        : await GetNode(l.NodeID2);
      let p1 = l.PollingID1;
      let p2 = l.PollingID2;

      if (!l.NodeID1.startsWith("NET:")) {
        const p = await GetPolling(p1);
        if (p) {
          p1 = p.Name;
        }
      } else if (n1) {
        for (const p of (n1 as datastore.NetworkEnt).Ports || []) {
          if (p.ID === p1) {
            p1 = p.Name;
            break;
          }
        }
      }

      if (!l.NodeID2.startsWith("NET:")) {
        const p = await GetPolling(p2);
        if (p) {
          p2 = p.Name;
        }
      } else if (n2) {
        for (const p of (n2 as datastore.NetworkEnt).Ports || []) {
          if (p.ID === p2) {
            p2 = p.Name;
            break;
          }
        }
      }

      data.push({
        Index: i,
        Node1: n1 ? n1.Name : "",
        Node2: n2 ? n2.Name : "",
        Polling1: p1,
        Polling2: p2,
        Confidence: l.Confidence || "strict",
        Reason: l.Reason || l.Info || "",
      });
    }
    return data;
  };

  const onOpen = async () => {
    wait = true;
    networkData = [];
    lineData = [];
    rawLines = [];
    if (id.startsWith("NODE:")) {
      isNodeMode = true;
    } else if (id.startsWith("NET:")) {
      isNodeMode = false;
    } else {
      const node = await GetNode(id);
      isNodeMode = !!node;
    }

    const cleanID = id.replace(/^(NODE:|NET:)/, "");

    if (isNodeMode) {
      // Find connections for a standard node across all switches
      const lines = await FindNodeConnection(cleanID);
      lineData = await parseLines(lines);
      wait = false;
      showLineTable();
      return;
    }

    resp = await FindNeighborNetworksAndLines(cleanID);
    if (resp && resp.Networks) {
      for (let i = 0; i < resp.Networks.length; i++) {
        const n = resp.Networks[i];
        networkData.push({
          Index: i,
          Name: n.Name,
          IP: n.IP,
          SystemID: n.SystemID,
          Descr: n.Descr,
        });
      }
    }
    if (resp && resp.Lines) {
      lineData = await parseLines(resp.Lines);
    }

    wait = false;
    showNetworkTable();
    showLineTable();
  };

  const runAI = async () => {
    wait = true;
    const cleanID = id.replace(/^(NODE:|NET:)/, "");
    resp = await FindNeighborNetworksAndLinesWithAI(isNodeMode ? ("NODE:" + cleanID) : ("NET:" + cleanID));
    if (resp && resp.Lines) {
      lineData = await parseLines(resp.Lines);
    }
    wait = false;
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
      title: $_("NeighborNetworksAndLines.Descr"),
      width: "50%",
    },
  ];

  const showNetworkTable = async () => {
    if (isNodeMode) return;
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
      scrollY: "25vh",
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
      data: "Confidence",
      title: $_("NeighborNetworksAndLines.Confidence"),
      width: "10%",
      render: (data: string) => {
        if (data === "strict") {
          return `<span class="bg-blue-100 text-blue-800 text-xs font-semibold px-2 py-0.5 rounded dark:bg-blue-900 dark:text-blue-200">${$_("NeighborNetworksAndLines.Strict")}</span>`;
        }
        return `<span class="bg-amber-100 text-amber-800 text-xs font-semibold px-2 py-0.5 rounded dark:bg-amber-900 dark:text-amber-200">${$_("NeighborNetworksAndLines.Speculative")}</span>`;
      },
    },
    {
      data: "Reason",
      title: $_("NeighborNetworksAndLines.Reason"),
      width: "15%",
    },
    {
      data: "Node1",
      title: $_("Line.Node1"),
      width: "20%",
    },
    {
      data: "Polling1",
      title: $_("Line.Polling1"),
      width: "18%",
    },
    {
      data: "Node2",
      title: $_("Line.Node2"),
      width: "20%",
    },
    {
      data: "Polling2",
      title: $_("Line.Polling2"),
      width: "17%",
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
      scrollY: isNodeMode ? "50vh" : "35vh",
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
    if (i < 0 || !resp || !resp.Networks || i >= resp.Networks.length) {
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
    if (i < 0 || j >= rawLines.length) {
      return;
    }
    const l = rawLines[j];
    await UpdateLine(l);
    lineData.splice(i, 1);
    rawLines.splice(j, 1);
    // update indices
    for (let k = 0; k < lineData.length; k++) {
      lineData[k].Index = k;
    }
    showLineTable();
  };

  const connectAllStrict = async () => {
    const toConnect: any[] = [];
    const remainingLineData: any[] = [];
    const remainingRawLines: any[] = [];

    for (let i = 0; i < lineData.length; i++) {
      if (lineData[i].Confidence === "strict") {
        toConnect.push(rawLines[lineData[i].Index]);
      } else {
        remainingLineData.push(lineData[i]);
        remainingRawLines.push(rawLines[lineData[i].Index]);
      }
    }

    if (toConnect.length > 0) {
      await ConnectLines(toConnect);
      lineData = remainingLineData;
      rawLines = remainingRawLines;
      for (let k = 0; k < lineData.length; k++) {
        lineData[k].Index = k;
      }
      showLineTable();
    }
  };

  const connectAll = async () => {
    if (rawLines.length > 0) {
      await ConnectLines(rawLines);
      lineData = [];
      rawLines = [];
      showLineTable();
    }
  };

  $: if (show) {
    onOpen();
  }
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full"
>
  {#if wait}
    <div class="text-center mt-10"><Spinner size="16" /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {isNodeMode ? $_("NeighborNetworksAndLines.NodeTitle") : $_("NeighborNetworksAndLines.Title")}
      </h3>
      {#if !isNodeMode}
        <div class="m-5 grow">
          <table id="networkTable" class="display compact" style="width:99%"></table>
        </div>
      {/if}
      <div class="m-5 grow">
        <table id="lineTable" class="display compact" style="width:99%"></table>
      </div>
      <div class="flex justify-between items-center mr-2">
        <div class="flex space-x-2">
          <GradientButton
            shadow
            color="purple"
            type="button"
            onclick={runAI}
            size="xs"
          >
            <Icon path={icons.mdiAutoFix} size={1} />
            {$_("NeighborNetworksAndLines.InferAI")}
          </GradientButton>
          {#if lineData.some((l) => l.Confidence === "strict")}
            <GradientButton
              shadow
              color="cyan"
              type="button"
              onclick={connectAllStrict}
              size="xs"
            >
              <Icon path={icons.mdiLanCheck} size={1} />
              {$_("NeighborNetworksAndLines.ConnectAllStrict")}
            </GradientButton>
          {/if}
          {#if lineData.length > 1}
            <GradientButton
              shadow
              color="blue"
              type="button"
              onclick={connectAll}
              size="xs"
            >
              <Icon path={icons.mdiLanConnect} size={1} />
              {$_("NeighborNetworksAndLines.ConnectAll")}
            </GradientButton>
          {/if}
        </div>
        <div class="flex space-x-2">
          {#if networkSelectedCount > 0}
            <GradientButton
              shadow
              color="blue"
              type="button"
              onclick={addNetwork}
              size="xs"
            >
              <Icon path={icons.mdiPlus} size={1} />
              {$_("NeighborNetworksAndLines.AddNetwork")}
            </GradientButton>
          {/if}
          {#if lineSelectedCount > 0}
            <GradientButton
              shadow
              color="blue"
              type="button"
              onclick={connectLine}
              size="xs"
            >
              <Icon path={icons.mdiLanConnect} size={1} />
              {$_("NeighborNetworksAndLines.ConnectLine")}
            </GradientButton>
          {/if}
          <GradientButton
            shadow
            type="button"
            color="teal"
            onclick={close}
            size="xs"
          >
            <Icon path={icons.mdiCancel} size={1} />
            {$_("Config.Close")}
          </GradientButton>
        </div>
      </div>
    </form>
  {/if}
</Modal>
