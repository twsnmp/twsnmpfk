<script lang="ts">
  import { initMAP, updateMAP } from "./map";
  import { onMount,onDestroy } from "svelte";
  import {Modal,GradientButton} from "flowbite-svelte";
  import * as icons from '@mdi/js';
  import Icon from "mdi-svelte";
  import Discover from "./Dsicover.svelte";
  import {DeleteDrawItems,DeleteNodes} from "../../wailsjs/go/main/App";

  let map: any;
  let posX:number = 0;
  let posY:number = 0;
  let showMapMenu :boolean= false;
  let showNodeMenu :boolean= false;
  let showDrawItemMenu :boolean= false;
  let showEditNode:boolean = false;
  let selectedNode: string = "";
  let showEditLine:boolean = false;
  let selectedLineNode1: string = "";
  let selectedLineNode2: string = "";
  let showEditDrawItem:boolean = false;
  let selectedDrawItem: string = "";
  let showDiscover :boolean= false;

	export let dark: boolean = false;

  onMount(async () => {
    initMAP(map,callBack);
		refreshMap();
    console.log("onMount map");
  });

  onDestroy(()=>{
    console.log("onDestroy map");
  });

  const callBack = (p) => {
    console.log(p);
    switch(p.Cmd){
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
    case "deletItems":
      deleteDrawItems(p.Param);
      break;
    }
  }
  let count = 0;
  let oldDark = false;
	const refreshMap = async() => {
    if (count < 2|| count % 5 == 0 || dark != oldDark) {
      updateMAP(dark);
      oldDark = dark;
    }
    count++;
    setTimeout(refreshMap,1000);
	};

  const deleteNodes = async (ids:string[]) => {
    await DeleteNodes(ids);
    count = 1;
    showNodeMenu = false;
  };

  const deleteDrawItems = async (ids:string[]) => {
    await DeleteDrawItems(ids);
    count = 1;
    showDrawItemMenu = false;
  };

</script>

<div bind:this={map} class="h-full w-full overflow-scroll"/>

<Modal bind:open={showMapMenu} size="xs"  outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton color="blue" class="w-full">
      <Icon path={icons.mdiPlus}></Icon>
      新規ノード
    </GradientButton>
    <GradientButton color="blue" class="w-full">
      <Icon path={icons.mdiDrawing}></Icon>
      描画アイテム
    </GradientButton>
    <GradientButton color="teal" class="w-full">
      <Icon path={icons.mdiCached}></Icon>
      全て再確認
    </GradientButton>
    <GradientButton color="cyan" class="w-full" on:click={() => {
        showMapMenu = false;
        showDiscover =true;
      }
      }>
      <Icon path={icons.mdiFileFind}></Icon>
      自動発見
    </GradientButton>
    <GradientButton color="red" class="w-full">
      <Icon path={icons.mdiGrid}></Icon>
      グリッド整列
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showNodeMenu} size="xs"  outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton color="blue" class="w-full">
      <Icon path={icons.mdiPencil}></Icon>
      編集
    </GradientButton>
    <GradientButton color="teal" class="w-full">
      <Icon path={icons.mdiCached}></Icon>
      再確認
    </GradientButton>
    <GradientButton color="cyan" class="w-full">
      <Icon path={icons.mdiContentCopy}></Icon>
      コピー
    </GradientButton>
    <GradientButton color="red" class="w-full" on:click={()=>{deleteNodes([selectedNode])}}>
      <Icon path={icons.mdiDelete}></Icon>
      削除
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showDrawItemMenu} size="xs"  outsideclose>
  <div class="flex flex-col space-y-2">
    <GradientButton color="blue" class="w-full">
      <Icon path={icons.mdiPencil}></Icon>
      編集
    </GradientButton>
    <GradientButton color="cyan" class="w-full">
      <Icon path={icons.mdiContentCopy}></Icon>
      コピー
    </GradientButton>
    <GradientButton color="red" class="w-full" on:click={()=>{deleteDrawItems([selectedDrawItem])}}>
      <Icon path={icons.mdiDelete}></Icon>
      削除
    </GradientButton>
  </div>
</Modal>

{#if showDiscover}
<Discover on:close={()=>{showDiscover = false}}></Discover>
{/if}