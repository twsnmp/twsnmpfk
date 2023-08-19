<script lang="ts">
  import logo from "./assets/images/appicon.png";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    Button,
    Dropdown,
    DropdownItem,
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import { GetSettings, GetVersion, GetMapName } from "../wailsjs/go/main/App";
  import Map from "./lib/Map.svelte";
  import Log from "./lib/Log.svelte";
  import MapConf from "./lib/MapConf.svelte";
  import NotifyConf from "./lib/NotifyConf.svelte";
  import AIConf from "./lib/AIConf.svelte";
  import NodeList from "./lib/NodeList.svelte";
  import PollingList from "./lib/PollingList.svelte";

  let version = "";
  let settings: any = undefined;
  let dark: boolean = false;
  let showMapConf: boolean = false;
  let showNotifyConf: boolean = false;
  let showAIConf: boolean = false;
  let mainHeight = 0;
  let mapName = "";
  let page = "map";

  const updateMapName = async () => {
    mapName = await GetMapName();
  };

  onMount(async () => {
    version = await GetVersion();
    settings = await GetSettings();
    await tick();
    mainHeight = window.innerHeight - 96;
    updateMapName();
  });

  const toggleDark = () => {
    const e = document.querySelector("html");
    e.classList.toggle("dark");
    dark = e.classList.contains("dark");
  };
</script>

<svelte:window on:resize={() => (mainHeight = window.innerHeight - 96)} />

<Navbar let:hidden let:toggle style="--wails-draggable:drag">
  <NavBrand href="/">
    <img src={logo} class="mr-3 h-12" alt="TWSNMP FK Logo" />
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
    <NavLi>
      <Icon path={icons.mdiCalendarCheck} size={1} />
      ログ
    </NavLi>
    <NavLi>
      <Icon path={icons.mdiBrain} size={1} />
      AI分析
    </NavLi>
    <NavLi id="nav-config">
      <Icon path={icons.mdiCog} size={1} />
      設定
    </NavLi>
    <Dropdown triggeredBy="#nav-config" class="w-44 z-20">
      <DropdownItem
        on:click={() => {
          showMapConf = true;
        }}
      >
        マップ
      </DropdownItem>
      <DropdownItem
        on:click={() => {
          showNotifyConf = true;
        }}
      >
        通知
      </DropdownItem>
      <DropdownItem
        on:click={() => {
          showAIConf = true;
        }}
      >
        AI分析
      </DropdownItem>
    </Dropdown>
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
    <div class="row-span-1">
      <Log />
    </div>
  </div>
{:else if page == "node"}
  <NodeList />
{:else if page == "polling"}
  <PollingList />
{/if}

{#if showMapConf}
  <MapConf
    on:close={() => {
      updateMapName();
      showMapConf = false;
    }}
  />
{/if}

{#if showNotifyConf}
  <NotifyConf
    on:close={() => {
      showNotifyConf = false;
    }}
  />
{/if}

{#if showAIConf}
  <AIConf
    on:close={() => {
      showAIConf = false;
    }}
  />
{/if}
