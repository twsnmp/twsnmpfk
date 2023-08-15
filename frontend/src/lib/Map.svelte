<script lang="ts">
  import { initMAP, updateMAP, resetMap,deleteMap, grid } from "./map";
  import { onMount, onDestroy } from "svelte";
  import { Modal, GradientButton, Button, Label, Input } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import Icon from "mdi-svelte";
  import Discover from "./Dsicover.svelte";
  import Node from "./Node.svelte";
  import Line from "./Line.svelte";
  import DrawItem from "./DrawItem.svelte";
  import { DeleteDrawItems, DeleteNodes } from "../../wailsjs/go/main/App";

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

  export let dark: boolean = false;
  let timer = undefined;

  onMount(async () => {
    initMAP(map, callBack);
    refreshMap();
    console.log("onMount map");
  });

  onDestroy(() => {
    console.log("onDestroy map");
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    deleteMap();
  });

  const callBack = (p) => {
    switch (p.Cmd) {
      case "contextMenu":
        posX = p.x;
        posY = p.y;
        if (p.Node) {
          showNodeMenu = true;
          selectedNode = p.Node;
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
      case "itemDoubleClicked":
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

<Modal bind:open={showMapMenu} size="xs" outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton
      color="blue"
      class="w-full"
      on:click={() => {
        selectedNode = "";
        showEditNode = true;
        showMapMenu = false;
      }}
    >
      <Icon path={icons.mdiPlus} />
      新規ノード
    </GradientButton>
    <GradientButton
      color="blue"
      class="w-full"
      on:click={() => {
        selectedDrawItem = "";
        showEditDrawItem = true;
        showMapMenu = false;
      }}
    >
      <Icon path={icons.mdiDrawing} />
      描画アイテム
    </GradientButton>
    <GradientButton color="teal" class="w-full">
      <Icon path={icons.mdiCached} />
      全て再確認
    </GradientButton>
    <GradientButton
      color="cyan"
      class="w-full"
      on:click={() => {
        showMapMenu = false;
        showDiscover = true;
      }}
    >
      <Icon path={icons.mdiFileFind} />
      自動発見
    </GradientButton>
    <GradientButton color="red" class="w-full" on:click={()=>{
      showMapMenu = false;
      showGrid = true;
    }}>
      <Icon path={icons.mdiGrid} />
      グリッド整列
    </GradientButton>
    <GradientButton
      color="teal"
      class="w-full"
      on:click={() => {
        resetMap();
        count = 1;
        showMapMenu = false;
      }}
    >
      <Icon path={icons.mdiRecycle} />
      更新
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showNodeMenu} size="xs" outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton
      color="blue"
      class="w-full"
      on:click={() => {
        showNodeMenu = false;
        showEditNode = true;
      }}
    >
      <Icon path={icons.mdiPencil} />
      編集
    </GradientButton>
    <GradientButton color="teal" class="w-full">
      <Icon path={icons.mdiCached} />
      再確認
    </GradientButton>
    <GradientButton color="cyan" class="w-full">
      <Icon path={icons.mdiContentCopy} />
      コピー
    </GradientButton>
    <GradientButton
      color="red"
      class="w-full"
      on:click={() => {
        deleteNodes([selectedNode]);
      }}
    >
      <Icon path={icons.mdiDelete} />
      削除
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showDrawItemMenu} size="xs" outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton
      color="blue"
      class="w-full"
      on:click={() => {
        showDrawItemMenu = false;
        showEditDrawItem = true;
      }}
    >
      <Icon path={icons.mdiPencil} />
      編集
    </GradientButton>
    <GradientButton color="cyan" class="w-full">
      <Icon path={icons.mdiContentCopy} />
      コピー
    </GradientButton>
    <GradientButton
      color="red"
      class="w-full"
      on:click={() => {
        deleteDrawItems([selectedDrawItem]);
      }}
    >
      <Icon path={icons.mdiDelete} />
      削除
    </GradientButton>
  </div>
</Modal>

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
    <div class="flex space-x-2">
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
