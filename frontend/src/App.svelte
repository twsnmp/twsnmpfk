<script lang="ts">
  import logo from "./assets/images/appicon.png";
  import { Navbar, NavBrand, NavLi, NavUl, Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import { GetMapName } from "../wailsjs/go/main/App";
  import Map from "./lib/Map.svelte";
  import Log from "./lib/Log.svelte";
  import NodeList from "./lib/NodeList.svelte";
  import PollingList from "./lib/PollingList.svelte";
  import EventLog from "./lib/EventLog.svelte";
  import Syslog from "./lib/Syslog.svelte";
  import Trap from "./lib/Trap.svelte";
  import Arp from "./lib/Arp.svelte";
  import AIList from "./lib/AIList.svelte";
  import Config from "./lib/Config.svelte";

  import {
    IsDark,
    SetDark,
  } from "../wailsjs/go/main/App";

  let dark: boolean = false;
  let mainHeight = 0;
  let mapName = "";
  let page = "map";
  let showConfig = false;

  const updateMapName = async () => {
    mapName = await GetMapName();
  };

  onMount(async () => {
    if (await IsDark()) {
      const e = document.querySelector("html");
       e.classList.add("dark"); 
    }
    await tick();
    mainHeight = window.innerHeight - 96;
    updateMapName();
  });

  const toggleDark = () => {
    const e = document.querySelector("html");
    e.classList.toggle("dark");
    dark = e.classList.contains("dark");
    SetDark(dark);
  };
</script>

<svelte:window on:resize={() => (mainHeight = window.innerHeight - 96)} />

<Navbar let:hidden let:toggle style="--wails-draggable:drag">
  <NavBrand href="/">
    <img src={logo} class="mr-3 h-12" alt="TWSNMP Logo" />
    <span
      class="self-center whitespace-nowrap text-xl font-semibold dark:text-white"
    >
      TWSNMP FK - {mapName}
    </span>
  </NavBrand>
  <NavUl>
    <NavLi
      active={page == "map"}
      on:click={() => {
        page = "map";
      }}
    >
      <Icon path={icons.mdiLan} size={1} />
      マップ
    </NavLi>
    <NavLi
      active={page == "node"}
      on:click={() => {
        page = "node";
      }}
    >
      <Icon path={icons.mdiLaptop} size={1} />
      ノード
    </NavLi>
    <NavLi
      active={page == "polling"}
      on:click={() => {
        page = "polling";
      }}
    >
      <Icon path={icons.mdiLanCheck} size={1} />
      ポーリング
    </NavLi>
    <NavLi
      active={page == "eventlog"}
      on:click={() => {
        page = "eventlog";
      }}
    >
      <Icon path={icons.mdiCalendarCheck} size={1} />
      ログ
    </NavLi>
    <NavLi
      active={page == "syslog"}
      on:click={() => {
        page = "syslog";
      }}
    >
      <Icon path={icons.mdiCalendarText} size={1} />
      syslog
    </NavLi>
    <NavLi
      active={page == "trap"}
      on:click={() => {
        page = "trap";
      }}
    >
      <Icon path={icons.mdiAlert} size={1} />
      TRAP
    </NavLi>
    <NavLi
      active={page == "arp"}
      on:click={() => {
        page = "arp";
      }}
    >
      <Icon path={icons.mdiCheckNetwork} size={1} />
      ARP
    </NavLi>
    <NavLi
      active={page == "ai"}
      on:click={() => {
        page = "ai";
      }}
    >
      <Icon path={icons.mdiBrain} size={1} />
      AI分析
    </NavLi>
    <NavLi
      active={showConfig}
      on:click={() => {
        showConfig = true;
      }}
    >
      <Icon path={icons.mdiCog} size={1} />
      設定
    </NavLi>
  </NavUl>
  <Button class="!p-2" color="alternative" on:click={toggleDark}>
    {#if dark}
      <Icon path={icons.mdiWeatherSunny} size={1} />
    {:else}
      <Icon path={icons.mdiMoonWaxingCrescent} size={1} />
    {/if}
  </Button>
</Navbar>

{#if page == "map"}
  <div
    class="grid grid-rows-4 grid-cols-1 gap-0 w-full"
    style="height:{mainHeight}px;"
  >
    <div class="row-span-3">
      <Map {dark} />
    </div>
    <div class="row-span-1 ml-2 mr-2">
      <Log />
    </div>
  </div>
{:else if page == "node"}
  <NodeList />
{:else if page == "polling"}
  <PollingList />
{:else if page == "eventlog"}
  <EventLog />
{:else if page == "syslog"}
  <Syslog />
{:else if page == "trap"}
  <Trap />
{:else if page == "arp"}
  <Arp />
{:else if page == "ai"}
  <AIList />
{/if}

{#if showConfig}
  <Config
    on:close={() => {
      updateMapName();
      showConfig = false;
    }}
  />
{/if}
