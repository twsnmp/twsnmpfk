<script lang="ts">
  import { initMAP, updateMAP } from "./map";
  import { onMount,onDestroy } from "svelte";
  import {Modal,GradientButton} from "flowbite-svelte";
  import * as icons from '@mdi/js';
  import Icon from "mdi-svelte";
  import Discover from "./Dsicover.svelte";

  let map: any;
  let posX:number = 0;
  let posY:number = 0;
  let showMapMenu :boolean= false;
  let showNodeMenu :boolean= false;
  let showDrawItemMenu :boolean= false;
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

  const callBack = async (p) => {
    console.log(p);
    switch(p.Cmd){
    case "contextMenu":
      posX = p.x;
      posY = p.y;
      if (p.Node) {
        showNodeMenu = true;
      } else if (p.Item) {
        showDrawItemMenu = true;
      } else {
        showMapMenu = true;
      }
      break;
    }
  }
  let count = 0;
  let oldDark = false;
	const refreshMap = async() => {
    if (count < 3|| count % 5 == 0 || dark != oldDark) {
      updateMAP(dark);
      oldDark = dark;
    }
    count++;
    setTimeout(refreshMap,1000);
	}
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
    <GradientButton color="red" class="w-full">
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
    <GradientButton color="red" class="w-full">
      <Icon path={icons.mdiDelete}></Icon>
      削除
    </GradientButton>
  </div>
</Modal>

{#if showDiscover}

<Discover on:close={()=>{showDiscover = false}}></Discover>

{/if}