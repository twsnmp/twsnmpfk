<script lang="ts">
  import { initMAP, updateMAP, resetMap, deleteMap, grid,setShowAllItems } from "./map";
  import { onMount, onDestroy } from "svelte";
  import { Modal, GradientButton, Label, Input } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import Icon from "mdi-svelte";
  import Discover from "./Dsicover.svelte";
  import Node from "./Node.svelte";
  import Line from "./Line.svelte";
  import DrawItem from "./DrawItem.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import Ping from "./Ping.svelte";
  import {
    CheckPolling,
    DeleteDrawItems,
    DeleteNodes,
    CopyNode,
    CopyDrawItem,
    WakeOnLan,
    GetNode,
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import MIBBrowser from "./MIBBrowser.svelte";
  import { _ } from "svelte-i18n";

  let map: any;
  let posX: number = 0;
  let posY: number = 0;
  let showMapMenu: boolean = false;
  let showNodeMenu: boolean = false;
  let showDrawItemMenu: boolean = false;
  let showEditNode: boolean = false;
  let selectedNode: string = "";
  let showEditLine: boolean = false;
  let selectedLineNode1: string = "";
  let selectedLineNode2: string = "";
  let showEditDrawItem: boolean = false;
  let selectedDrawItem: string = "";
  let showDiscover: boolean = false;
  let showGrid: boolean = false;
  let gridSize: number = 40;
  let showNodeReport: boolean = false;
  let showPolling: boolean = false;
  let showPing: boolean = false;
  let showMibBr: boolean = false;
  let showAllItems: boolean = false;

  let timer = undefined;
  let urls = [];

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
  const callBack = (p) => {
    switch (p.Cmd) {
      case "contextMenu":
        posX = p.x;
        posY = p.y;
        if (p.Node) {
          showNodeMenuFunc(p.Node);
        } else if (p.DrawItem) {
          showDrawItemMenu = true;
          selectedDrawItem = p.DrawItem;
        } else {
          showMapMenu = true;
        }
        break;
      case "editLine":
        if (p.Param) {
          showEditLine = true;
          selectedLineNode1 = p.Param[0];
          selectedLineNode2 = p.Param[1];
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
      case "deleteNodes":
        deleteNodes(p.Param);
        break;
      case "deleteDrawItems":
        deleteDrawItems(p.Param);
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
</script>

<div bind:this={map} class="h-full w-full overflow-scroll" />

{#if showMapMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-40 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-800 p-2"
    >
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
            {$_('Map.showDrawItemNomal')}
          </div>
        {:else}
          <Icon path={icons.mdiDraw} size={0.7} />
          <div>
            {$_('Map.showDrawItemEdit')}
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

{#if showDiscover}
  <Discover
    on:close={() => {
      showDiscover = false;
      refreshMap();
    }}
  />
{/if}

{#if showEditNode}
  <Node
    nodeID={selectedNode}
    {posX}
    {posY}
    on:close={(e) => {
      showEditNode = false;
      refreshMap();
    }}
  />
{/if}

{#if showEditLine}
  <Line
    nodeID1={selectedLineNode1}
    nodeID2={selectedLineNode2}
    on:close={(e) => {
      showEditLine = false;
      refreshMap();
    }}
  />
{/if}

{#if showEditDrawItem}
  <DrawItem
    id={selectedDrawItem}
    {posX}
    {posY}
    on:close={(e) => {
      showEditDrawItem = false;
      refreshMap();
    }}
  />
{/if}

{#if showNodeReport}
  <NodeReport
    id={selectedNode}
    on:close={(e) => {
      showNodeReport = false;
    }}
  />
{/if}

{#if showPolling}
  <NodePolling
    nodeID={selectedNode}
    on:close={(e) => {
      showPolling = false;
      refreshMap();
    }}
  />
{/if}

{#if showPing}
  <Ping
    nodeID={selectedNode}
    on:close={(e) => {
      showPing = false;
    }}
  />
{/if}

{#if showMibBr}
  <MIBBrowser
    nodeID={selectedNode}
    on:close={(e) => {
      showMibBr = false;
    }}
  />
{/if}

<Modal bind:open={showGrid} size="sm" permanent class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Map.Grid")}
    </h3>
    <Label class="space-y-2">
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

<svelte:window
  on:click={() => {
    showMapMenu = false;
    showNodeMenu = false;
    showDrawItemMenu = false;
  }}
/>
