<script lang="ts">
  import { setMAP, showMAP, setMapContextMenu } from "./map";
  import { onMount,onDestroy } from "svelte";
  import {
    GetMapConf, 
    GetSettings, 
    GetVersion,
    SetMapConf,
    GetNotifyConf,
    SetNotifyConf,
    TestNotifyConf,
    GetAIConf,
    SetAIConf,
  } from "../../wailsjs/go/main/App"
  import type { datastore} from "../../wailsjs/go/models";

  let map: any;
	export let dark: boolean = false;

  onMount(async () => {
    showMAP(map);
		maptest();
    setMapContextMenu(true);
    console.log("onMount map");
  });

  onDestroy(()=>{
    console.log("onDestroy map");
  });

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
