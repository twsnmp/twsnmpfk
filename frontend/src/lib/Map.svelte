<script lang="ts">
  import { setMAP, showMAP } from "./map";
  import { onMount,onDestroy } from "svelte";
  import {Modal,GradientButton} from "flowbite-svelte";
  import * as icons from '@mdi/js';
  import Icon from "mdi-svelte";

  let map: any;
  let posX:number = 0;
  let posY:number = 0;
  let showMapMenu :boolean= false;
  let showNodeMenu :boolean= false;
  let showDrawItemMenu :boolean= false;

	export let dark: boolean = false;

  onMount(async () => {
    showMAP(map,callBack);
		maptest();
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

	const maptest = async() => {
    setMAP(
      {
        Nodes: {
          node1: {
            ID: "node1",
            X: 100,
            Y: 200,
            Icon: "mdi-microsoft-windows",
            State: "normal",
            Name: "Node1",
          },
          node2: {
            ID: "node2",
            X: 160,
            Y: 200,
            Icon: "mdi-linux",
            State: "low",
            Name: "Node2",
          },
        },
        Lines: {
          line1: {
            ID: "line1",
            NodeID1: "node1",
            NodeID2: "node2",
            State1: "normal",
            State2: "low",
          },
        },
        Items: {
          item1: {
            ID: "item1",
            Type: 2,
            Size: 24,
            X: 50,
            Y: 100,
            Text: "test",
            Color: "red",
          },
        },
        MapConf: {
          BackImage: {
             Color: dark ? "black" : "white",
          },
        },
      },
			dark,
      false,
    );
    console.log(dark);
    setTimeout(maptest,5000);
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
    <GradientButton color="cyan" class="w-full">
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
