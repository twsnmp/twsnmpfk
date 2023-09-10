<script lang="ts">
  import { initMAP, updateMAP, resetMap, deleteMap, grid } from "./map";
  import { onMount, onDestroy } from "svelte";
  import { Modal, GradientButton, Button, Label, Input } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import Icon from "mdi-svelte";
  import Discover from "./Dsicover.svelte";
  import Node from "./Node.svelte";
  import Line from "./Line.svelte";
  import DrawItem from "./DrawItem.svelte";
  import NodeReport from "./NodeReport.svelte";
  import AddPolling from "./AddPolling.svelte";
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

  let map: any;
  let posX: number = 0;
  let posY: number = 0;
  let showMapMenu: boolean = false;
  let showMapMenu2: boolean = false;
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
  let showAddPolling: boolean = false;
  let showPing: boolean = false;
  let showMibBr: boolean = false;

  export let dark: boolean = false;
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
  let count = 0;
  let oldDark = false;
  const refreshMap = async () => {
    if (count < 2 || count % 5 == 0 || dark != oldDark) {
      updateMAP(dark);
      oldDark = dark;
    }
    count++;
    timer = setTimeout(refreshMap, 1000);
  };

  const deleteNodes = async (ids: string[]) => {
    await DeleteNodes(ids);
    count = 1;
    showNodeMenu = false;
  };

  const deleteDrawItems = async (ids: string[]) => {
    await DeleteDrawItems(ids);
    count = 1;
    showDrawItemMenu = false;
  };
</script>

<div bind:this={map} class="h-full w-full overflow-scroll" />

{#if showMapMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-30 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-500 px-1"
    >
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          selectedNode = "";
          showEditNode = true;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiPlus} size={0.8} />
        新規ノード
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          selectedDrawItem = "";
          showEditDrawItem = true;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiDrawing} size={0.8} />
        描画アイテム
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showMapMenu = false;
          CheckPolling("");
        }}
      >
        <Icon path={icons.mdiCached} size={0.8} />
        全て再確認
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showMapMenu = false;
          showDiscover = true;
        }}
      >
        <Icon path={icons.mdiFileFind} size={0.8} />
        自動発見
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showMapMenu = false;
          showGrid = true;
        }}
      >
        <Icon path={icons.mdiGrid} size={0.8} />
        グリッド整列
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          resetMap();
          count = 1;
          showMapMenu = false;
        }}
      >
        <Icon path={icons.mdiRecycle} />
        更新
      </div>
    </div>
  </div>
{/if}

{#if showNodeMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-white w-30 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-500 px-1"
    >
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          showNodeReport = true;
        }}
      >
        <Icon path={icons.mdiChartBarStacked} size={0.8} />
        レポート
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          showPing = true;
        }}
      >
        <Icon path={icons.mdiShippingPallet}  size={0.8}/>
        PING
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          showMibBr = true;
        }}
      >
        <Icon path={icons.mdiShippingPallet} size={0.8} />
        MIBブラウザー
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          WakeOnLan(selectedNode);
        }}
      >
        <Icon path={icons.mdiAlarm} size={0.8} />
        Wake On Lan
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          showEditNode = true;
        }}
      >
        <Icon path={icons.mdiPencil} size={0.8} />
        編集
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          showAddPolling = true;
        }}
      >
        <Icon path={icons.mdiLanCheck} size={0.8} />
        ポーリング
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showNodeMenu = false;
          CheckPolling(selectedNode);
        }}
      >
        <Icon path={icons.mdiCached} size={0.8} />
        再確認
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={async () => {
          showNodeMenu = false;
          await CopyNode(selectedNode);
          count = 1;
        }}
      >
        <Icon path={icons.mdiContentCopy} size={0.8} />
        コピー
      </div>
      <div
        class="flex text-red-500 hover:bg-gray-100 "
        on:click={() => {
          deleteNodes([selectedNode]);
        }}
      >
        <Icon path={icons.mdiDelete} size={0.8} />
        削除
      </div>
      {#each urls as url}
        {#if url}
          <div
            class="flex hover:bg-gray-100"
            on:click={() => {
              showNodeMenu = false;
              BrowserOpenURL(url);
            }}
          >
            <Icon path={icons.mdiLink} size={0.8} />
            {url}
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
      class="bg-white w-30 border border-gray-300 flex flex-col text-xs space-y-1 text-gray-500 px-1"
    >
      <div
        class="flex hover:bg-gray-100"
        on:click={() => {
          showDrawItemMenu = false;
          showEditDrawItem = true;
        }}
      >
        <Icon path={icons.mdiPencil} size={0.8} />
        編集
      </div>
      <div
        class="flex hover:bg-gray-100"
        on:click={async () => {
          showDrawItemMenu = false;
          await CopyDrawItem(selectedDrawItem);
          count = 1;
        }}
      >
        <Icon path={icons.mdiContentCopy} size={0.8} />
        コピー
      </div>
      <div
        class="flex text-red-500 hover:bg-gray-100"
        on:click={() => {
          deleteDrawItems([selectedDrawItem]);
        }}
      >
        <Icon path={icons.mdiDelete} size={0.8} />
        削除
      </div>
    </div>
  </div>
{/if}

{#if showDiscover}
  <Discover
    on:close={() => {
      showDiscover = false;
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
      count = 1;
    }}
  />
{/if}

{#if showEditLine}
  <Line
    nodeID1={selectedLineNode1}
    nodeID2={selectedLineNode2}
    on:close={(e) => {
      showEditLine = false;
      count = 1;
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
      count = 1;
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

{#if showAddPolling}
  <AddPolling
    nodeID={selectedNode}
    on:close={(e) => {
      showAddPolling = false;
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
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">グリッド整列</h3>
    <Label class="space-y-2">
      <span>グリッドサイズ </span>
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
      <Button
        color="red"
        type="button"
        on:click={() => {
          showGrid = false;
          grid(gridSize, false);
        }}
        size="sm"
      >
        <Icon path={icons.mdiRun} size={1} />
        実行
      </Button>
      <Button
        color="blue"
        type="button"
        on:click={() => {
          showGrid = false;
          grid(gridSize, true);
        }}
        size="sm"
      >
        <Icon path={icons.mdiTestTube} size={1} />
        テスト
      </Button>
      <Button
        color="alternative"
        type="button"
        on:click={() => {
          showGrid = false;
        }}
        size="sm"
      >
        <Icon path={icons.mdiCancel} size={1} />
        キャンセル
      </Button>
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
