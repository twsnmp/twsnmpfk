<script lang="ts">
  import {
    initMAP,
    updateMAP,
    resetMap,
    deleteMap,
    grid,
    setShowAllItems,
    zoom,
    horizontal,
    vertical,
    circle,
  } from "./map";
  import { onMount, onDestroy } from "svelte";
  import { Modal, GradientButton, Label, Input, Button } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import { Icon } from "mdi-svelte-ts";
  import Discover from "./Dsicover.svelte";
  import Node from "./Node.svelte";
  import Network from "./Network.svelte";
  import Line from "./Line.svelte";
  import DrawItem from "./DrawItem.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import Ping from "./Ping.svelte";
  import NetworkLines from "./NetworkLines.svelte";
  import NeighborNetworksAndLines from "./NeighborNetworksAndLines.svelte";
  import {
    CheckPolling,
    DeleteDrawItems,
    DeleteNodes,
    CopyNode,
    CopyDrawItem,
    WakeOnLan,
    GetNode,
    SelectFile,
    GetImage,
    GetBackImage,
    SetBackImage,
    ImportV4Map,
    DeleteNetwork,
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import MIBBrowser from "./MIBBrowser.svelte";
  import { _ } from "svelte-i18n";
  import type { datastore } from "wailsjs/go/models";

  let map: any;
  let posX: number = 0;
  let posY: number = 0;
  let showMapMenu: boolean = false;
  let showNodeMenu: boolean = false;
  let showDrawItemMenu: boolean = false;
  let showNetworkMenu: boolean = false;
  let showFormatNodesMenu: boolean = false;
  let showEditNode: boolean = false;
  let selectedNode: string = "";
  let showEditLine: boolean = false;
  let selectedLineNode1: string = "";
  let selectedLineNode2: string = "";
  let selectedLineID: string = "";
  let showEditDrawItem: boolean = false;
  let selectedDrawItem: string = "";
  let showEditNetwork: boolean = false;
  let selectedNetwork: string = "";
  let networkTemplate: any = undefined;
  let showNetworkLines: boolean = false;
  let showNeighborNetworksAndLines: boolean = false;
  let showDiscover: boolean = false;
  let showGrid: boolean = false;
  let gridSize: number = 40;
  let showNodeReport: boolean = false;
  let showPolling: boolean = false;
  let showPing: boolean = false;
  let showMibBr: boolean = false;
  let showAllItems: boolean = false;

  let timer: any = undefined;
  let urls: any = [];

  onMount(async () => {
    initMAP(map, callBack);
    refreshMap();
  });

  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    deleteMap();
  });

  const showNodeMenuFunc = async (id: string) => {
    selectedNode = id;
    urls = [];
    const n = await GetNode(id);
    urls = n.URL.split(",");
    showNodeMenu = true;
  };

  let selectedNodes :any = [];
  let mapPosX = 0;
  let mapPosY = 0;

  const callBack = (p: any) => {
    switch (p.Cmd) {
      case "contextMenu":
        posX = p.x;
        posY = p.y - 96;
        if (p.Node) {
          showNodeMenuFunc(p.Node);
        } else if (p.DrawItem) {
          showDrawItemMenu = true;
          selectedDrawItem = p.DrawItem;
        } else if (p.Network) {
          selectedNetwork =  p.Network;
          networkTemplate = undefined;
          showNetworkMenu = true;
        } else {
          showMapMenu = true;
          mapPosX = p.x;
          mapPosY = p.y;
          if (map) {
            mapPosX = Math.trunc(mapPosX +map.scrollLeft);
            mapPosY = Math.trunc(mapPosY +map.scrollTop);
          }
        }
        break;
      case "editLine":
        if (p.Param) {
          selectedLineNode1 = p.Param[0];
          selectedLineNode2 = p.Param[1];
          selectedLineID = "";
          showEditLine = true;
        }
        break;
      case "nodeDoubleClicked":
        selectedNode = p.Param;
        showNodeReport = true;
        break;
      case "itemDoubleClicked":
        selectedDrawItem = p.Param;
        showEditDrawItem = true;
        break;
      case "networkDoubleClicked":
        selectedNetwork = p.Param;
        showEditNetwork = true;
        break;
      case "deleteNodes":
        deleteNodes(p.Param);
        break;
      case "formatNodes":
        posX = p.x;
        posY = p.y - 96;
        selectedNodes = [];
        for(const id of p.Nodes) {
          selectedNodes.push(id);
        }
        showFormatNodesMenu = true;
        break;
      case "deleteDrawItems":
        deleteDrawItems(p.Param);
        break;
      case "deleteNetwork":
        deleteNetwork(p.Param);
        break;
    }
  };

  const refreshMap = async () => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    updateMAP();
    timer = setTimeout(refreshMap, 1000 * 10);
  };

  const deleteNodes = async (ids: string[]) => {
    await DeleteNodes(ids);
    showNodeMenu = false;
    refreshMap();
  };

  const deleteDrawItems = async (ids: string[]) => {
    await DeleteDrawItems(ids);
    showDrawItemMenu = false;
    refreshMap();
  };

  const deleteNetwork = async (id:string) => {
    await DeleteNetwork(id);
    showNetworkMenu = false;
    refreshMap();
  };

  let showEditBackImage = false;
  let backImage: datastore.BackImageEnt;
  let image: any = undefined;
  const showEditBackImageDlg = async () => {
    backImage = await GetBackImage();
    if (backImage.Path) {
      image = await GetImage(backImage.Path);
    }
    if (backImage.Height < 1) {
      backImage.Height = 100;
    }
    if (backImage.Width < 1) {
      backImage.Width = 100;
    }
    showMapMenu = false;
    showEditBackImage = true;
  };

  const selectImage = async () => {
    const p = await SelectFile($_("Map.BackImage"), true);
    if (p) {
      backImage.Path = p;
      image = await GetImage(p);
    }
  };

  const saveBackImage = async () => {
    showEditBackImage = false;
    backImage.Height *=1;
    backImage.Width *=1;
    backImage.X *=1;
    backImage.Y *=1;
    await SetBackImage(backImage);
    refreshMap();
  };

  const clearBackImage = async () => {
    showEditBackImage = false;
    backImage.Path = "";
    image = undefined;
    backImage.Height *=1;
    backImage.Width *=1;
    backImage.X *=1;
    backImage.Y *=1;
    await SetBackImage(backImage);
    refreshMap();
  };
</script>

<div bind:this={map} class="h-full w-full overflow-scroll" />

<Button
  color="alternative"
  class="!p-2 absolute end-20 bottom-6"
  on:click={() => zoom(true)}
>
  <Icon path={icons.mdiMagnifyPlus}></Icon>
</Button>

<Button
  color="alternative"
  class="!p-2 absolute end-6 bottom-6"
  on:click={() => zoom(false)}
>
  <Icon path={icons.mdiMagnifyMinus}></Icon>
</Button>

{#if showMapMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          selectedNode = "";
          showEditNode = true;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiPlus} size={0.7} />
        <div>
          {$_("Map.AddNode")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          selectedDrawItem = "";
          showEditDrawItem = true;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiDrawing} size={0.7} />
        <div>
          {$_("Map.AddDrawItem")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          selectedNetwork = "";
          showEditNetwork = true;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiDrawing} size={0.7} />
        <div>
          新規ネットワーク
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showMapMenu = false;
          CheckPolling("all");
          refreshMap();
        }}
      >
        <Icon path={icons.mdiCached} size={0.7} />
        <div>
          {$_("Map.CheckAll")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showMapMenu = false;
          showDiscover = true;
        }}
      >
        <Icon path={icons.mdiFileFind} size={0.7} />
        <div>
          {$_("Map.Discover")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 space-x-2 hover:bg-sky-500/[0.8]"
        on:click={async () => {
          showMapMenu = false;
          ImportV4Map();
          refreshMap();
        }}
      >
        <Icon path={icons.mdiImport} size={0.7} />
        <div>
          {$_('Map.Import')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showMapMenu = false;
          showGrid = true;
        }}
      >
        <Icon path={icons.mdiGrid} size={0.7} />
        <div>
          {$_("Map.Grid")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={showEditBackImageDlg}
      >
        <Icon path={icons.mdiImage} size={0.7} />
        <div>
          {$_("Map.BackImage")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showMapMenu = false;
          resetMap();
          refreshMap();
        }}
      >
        <Icon path={icons.mdiRecycle} size={0.7} />
        <div>
          {$_("Map.Reload")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showAllItems = !showAllItems;
          setShowAllItems(showAllItems);
          showMapMenu = false;
        }}
      >
        {#if showAllItems}
          <Icon path={icons.mdiEye} size={0.7} />
          <div>
            {$_("Map.showDrawItemNomal")}
          </div>
        {:else}
          <Icon path={icons.mdiDraw} size={0.7} />
          <div>
            {$_("Map.showDrawItemEdit")}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

{#if showNodeMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          showNodeReport = true;
        }}
      >
        <Icon path={icons.mdiChartBarStacked} size={0.7} />
        <div>
          {$_("Map.Report")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          showPing = true;
        }}
      >
        <Icon path={icons.mdiCheckNetwork} size={0.7} />
        <div>PING</div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          showMibBr = true;
        }}
      >
        <Icon path={icons.mdiEye} size={0.7} />
        <div>
          {$_("Map.MIBBrowser")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          WakeOnLan(selectedNode);
        }}
      >
        <Icon path={icons.mdiAlarm} size={0.7} />
        <div>Wake On Lan</div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          showEditNode = true;
        }}
      >
        <Icon path={icons.mdiPencil} size={0.7} />
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          showPolling = true;
        }}
      >
        <Icon path={icons.mdiLanCheck} size={0.7} />
        <div>
          {$_("Map.Polling")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNodeMenu = false;
          CheckPolling(selectedNode);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiCached} size={0.7} />
        <div>
          {$_("Map.ReCheck")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={async () => {
          showNodeMenu = false;
          await CopyNode(selectedNode);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiContentCopy} size={0.7} />
        <div>
          {$_("Map.Copy")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex text-red-500 space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          deleteNodes([selectedNode]);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiDelete} size={0.7} />
        <div>
          {$_("Map.Delete")}
        </div>
      </div>
      {#each urls as url}
        {#if url}
          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <div
            class="flex space-x-2 hover:bg-sky-500/[0.8]"
            on:click={() => {
              showNodeMenu = false;
              BrowserOpenURL(url);
            }}
          >
            <Icon path={icons.mdiLink} size={0.7} />
            <div>
              {url}
            </div>
          </div>
        {/if}
      {/each}
    </div>
  </div>
{/if}

{#if showDrawItemMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showDrawItemMenu = false;
          showEditDrawItem = true;
        }}
      >
        <Icon path={icons.mdiPencil} size={0.7} />
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={async () => {
          showDrawItemMenu = false;
          await CopyDrawItem(selectedDrawItem);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiContentCopy} size={0.7} />
        <div>
          {$_("Map.Copy")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex text-red-500 space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          deleteDrawItems([selectedDrawItem]);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiDelete} size={0.7} />
        <div>
          {$_("Map.Delete")}
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showNetworkMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNetworkMenu = false;
          showEditNetwork = true;
        }}
      >
        <Icon path={icons.mdiPencil} size={0.7} />
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNetworkMenu = false;
          showNetworkLines = true;
        }}
      >
        <Icon path={icons.mdiPlaylistEdit} size={0.7} />
        <div>
          ライン編集
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showNetworkMenu = false;
          showNeighborNetworksAndLines = true;
        }}
      >
        <Icon path={icons.mdiLanConnect} size={0.7} />
        <div>
          接続先を探す
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex text-red-500 space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          deleteNetwork(selectedNetwork);
          refreshMap();
        }}
      >
        <Icon path={icons.mdiDelete} size={0.7} />
        <div>
          {$_("Map.Delete")}
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showFormatNodesMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showFormatNodesMenu = false;
          horizontal(selectedNodes);
          selectedNodes = [];
        }}
      >
        <Icon path={icons.mdiFormatVerticalAlignCenter} size={0.7} />
        <div>
          {$_('Map.Horizontal')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showFormatNodesMenu = false;
          vertical(selectedNodes);
          selectedNodes = [];
        }}
      >
        <Icon path={icons.mdiFormatHorizontalAlignCenter} size={0.7} />
        <div>
          {$_('Map.Vertical')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex space-x-2 hover:bg-sky-500/[0.8]"
        on:click={() => {
          showFormatNodesMenu = false;
          circle(selectedNodes);
          selectedNodes = [];
        }}
      >
        <Icon path={icons.mdiCircleOutline} size={0.7} />
        <div>
          {$_('Map.Circle')}
        </div>
      </div>
    </div>
  </div>
{/if}

<Discover
  bind:show={showDiscover}
  on:close={() => {
    refreshMap();
  }}
/>

<Node
  bind:show={showEditNode}
  nodeID={selectedNode}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    refreshMap();
  }}
/>

<Line
  bind:show={showEditLine}
  nodeID1={selectedLineNode1}
  nodeID2={selectedLineNode2}
  id={selectedLineID}
  on:close={(e) => {
    refreshMap();
  }}
/>

<DrawItem
  bind:show={showEditDrawItem}
  id={selectedDrawItem}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    refreshMap();
  }}
/>

<Network
  bind:show={showEditNetwork}
  id={selectedNetwork}
  template={networkTemplate}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    networkTemplate = undefined;
    refreshMap();
  }}
/>

<NetworkLines
  bind:show={showNetworkLines}
  id={selectedNetwork}
  on:close={(e) => {
    refreshMap();
  }}
  on:editLine={(e) => {
    selectedLineID = e.detail;
    selectedLineNode1= "";
    selectedLineNode2= "";
    showEditLine = true;
  }}
/>

<NeighborNetworksAndLines
  bind:show={showNeighborNetworksAndLines}
  id={selectedNetwork}
  on:close={(e) => {
    refreshMap();
  }}
  on:addNetwork={(e) => {
    networkTemplate = e.detail;
    showEditNetwork = true;
  }}
/>

<NodeReport bind:show={showNodeReport} id={selectedNode} />

<NodePolling
  bind:show={showPolling}
  nodeID={selectedNode}
  on:close={(e) => {
    refreshMap();
  }}
/>

<Ping bind:show={showPing} nodeID={selectedNode} />

<MIBBrowser bind:show={showMibBr} nodeID={selectedNode} />

<Modal bind:open={showGrid} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Map.Grid")}
    </h3>
    <Label class="space-y-2 text-xs">
      <span>{$_("Map.GridSize")} </span>
      <Input
        type="number"
        min={20}
        max={120}
        step={1}
        bind:value={gridSize}
        size="sm"
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="red"
        type="button"
        on:click={() => {
          showGrid = false;
          grid(gridSize, false);
        }}
        size="xs"
      >
        <Icon path={icons.mdiRun} size={1} />
        {$_("Map.Exec")}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        on:click={() => {
          showGrid = false;
          grid(gridSize, true);
        }}
        size="xs"
      >
        <Icon path={icons.mdiTestTube} size={1} />
        {$_("Map.Test")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        on:click={() => {
          showGrid = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal
  bind:open={showEditBackImage}
  size="sm"
  dismissable={false}
  class="w-full"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Map.BackImage")}
    </h3>
    <div class="grid gap-4 mb-4 grid-cols-5">
      <Label class="space-y-2 text-xs">
        <span>X</span>
        <Input
          type="number"
          min={0}
          max={2000}
          bind:value={backImage.X}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>Y</span>
        <Input
          type="number"
          min={0}
          max={2000}
          bind:value={backImage.Y}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("DrawItem.Width")}</span>
        <Input
          type="number"
          min={10}
          max={1000}
          bind:value={backImage.Width}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("DrawItem.Height")}</span>
        <Input
          type="number"
          min={10}
          max={1000}
          bind:value={backImage.Height}
          size="sm"
        />
      </Label>
      <GradientButton
        shadow
        class="h-8 mt-6 w-20"
        type="button"
        size="xs"
        color="blue"
        on:click={selectImage}
      >
        <Icon path={icons.mdiImage} size={1} />
        {$_("DrawItem.Select")}
      </GradientButton>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("DrawItem.Image")}</span>
      {#if image}
        <img src={image} alt="" class="h-32" />
      {:else}
        <div />
      {/if}
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="blue"
        type="button"
        on:click={saveBackImage}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_("Map.Save")}
      </GradientButton>
      {#if backImage.Path}
        <GradientButton
          color="red"
          type="button"
          on:click={clearBackImage}
          size="xs"
        >
          <Icon path={icons.mdiDelete} size={1} />
          {$_("Map.Clear")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        color="teal"
        type="button"
        on:click={() => {
          showEditBackImage = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<svelte:window
  on:click={() => {
    showMapMenu = false;
    showNodeMenu = false;
    showDrawItemMenu = false;
    showFormatNodesMenu = false;
  }}
/>
