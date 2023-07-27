<script lang="ts">
	import logo from "./assets/images/appicon.png"; 
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
		Button,
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import { mdiMoonWaxingCrescent,mdiWeatherSunny } from "@mdi/js";
  import { setMAP, showMAP, setMapContextMenu } from "./lib/map";
  import { onMount, tick } from "svelte";
  import {GetSettings, GetVersion} from "../wailsjs/go/main/App"
  let version = "";
  let settings :any = undefined;
  let map: any;
	let dark: boolean = false;
  onMount(async () => {
    version = await GetVersion();
    settings = await GetSettings();
    await tick();
    showMAP(map);
		maptest();
    setMapContextMenu(true);
  });
	const maptest = async() => {
    await tick();
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
      settings.Lock,
    );
	}
	const toggleDark = () => {
		const e = document.querySelector('html');
		e.classList.toggle('dark');
		dark = e.classList.contains('dark');
		maptest();
	}
</script>

<Navbar let:hidden let:toggle>
  <NavBrand href="/">
    <img
      src="{logo}"
      class="mr-3 h-12"
      alt="TWSNMP FK Logo"
    />
    <span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">
      TWSNMP FK {version}
    </span>
  </NavBrand>
  <NavUl>
    <NavLi active={true}>Map</NavLi>
    <NavLi >Node</NavLi>
    <NavLi >Polling</NavLi>
    <NavLi >Log</NavLi>
    <NavLi >AI</NavLi>
  </NavUl>
	<Button class="!p-2" color="alternative" on:click={toggleDark} >
		{#if dark}
			<Icon path={mdiWeatherSunny} size={1} />
		{:else}
			<Icon path={mdiMoonWaxingCrescent} size={1} />
		{/if}
	</Button>
</Navbar>

<div bind:this={map} class="w-full h-screen" />
