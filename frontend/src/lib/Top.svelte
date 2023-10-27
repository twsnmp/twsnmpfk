<script lang="ts">
  import logo from "../assets/images/appicon.png";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    Button,
    Badge,
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetMapName,
    IsDark,
    IsLatest,
    SetDark,
    GetSettings,
  } from "../../wailsjs/go/main/App";
  import Map from "./Map.svelte";
  import Log from "./Log.svelte";
  import NodeList from "./NodeList.svelte";
  import PollingList from "./PollingList.svelte";
  import EventLog from "./EventLog.svelte";
  import Syslog from "./Syslog.svelte";
  import Trap from "./Trap.svelte";
  import Arp from "./Arp.svelte";
  import Address from "./Address.svelte";
  import AIList from "./AIList.svelte";
  import Config from "./Config.svelte";
  import System from "./System.svelte";
  import { _ } from "svelte-i18n";

  let dark: boolean = false;
  let mainHeight = 0;
  let mapName = "";
  let page = "map";
  let oldPage = "";
  let showConfig = false;
  let latest = true;
  let lock = false;

  const updateMapName = async () => {
    mapName = await GetMapName();
  };

  const checkLatest = async () => {
    latest = await IsLatest();
    if (!latest) {
      return;
    }
    setTimeout(checkLatest, 1000 * 60);
  };

  onMount(async () => {
    const e = document.querySelector("html");
    if (await IsDark()) {
      e.classList.add("dark");
      dark = true;
    } else {
      e.classList.remove("dark");
      dark = false;
    }
    const settings = await GetSettings();
    lock = settings.Lock;
    await tick();
    mainHeight = window.innerHeight - 96;
    updateMapName();
    checkLatest();
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
    {#if !lock}
    <NavLi
      active={page == "map"}
      on:click={() => {
        page = "map";
      }}
    >
      <Icon path={icons.mdiLan} size={1} />
      {$_("Top.Map")}
    </NavLi>
      <NavLi
        active={page == "node"}
        on:click={() => {
          page = "node";
        }}
      >
        <Icon path={icons.mdiLaptop} size={1} />
        {$_("Top.Node")}
      </NavLi>
      <NavLi
        active={page == "polling"}
        on:click={() => {
          page = "polling";
        }}
      >
        <Icon path={icons.mdiLanCheck} size={1} />
        {$_("Top.Polling")}
      </NavLi>
      <NavLi
        active={page == "address"}
        on:click={() => {
          page = "address";
        }}
      >
        <Icon path={icons.mdiListStatus} size={1} />
        Address
      </NavLi>
      <NavLi
        active={page == "eventlog"}
        on:click={() => {
          page = "eventlog";
        }}
      >
        <Icon path={icons.mdiCalendarCheck} size={1} />
        {$_("Top.Log")}
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
        {$_("Top.AI")}
      </NavLi>
      <NavLi
        active={page == "system"}
        on:click={() => {
          page = "system";
        }}
      >
        <Icon path={icons.mdiChartLine} size={1} />
        System
      </NavLi>
      <NavLi
        active={showConfig}
        on:click={() => {
          oldPage = page;
          page = "";
          showConfig = true;
        }}
      >
        <Icon path={icons.mdiCog} size={1} />
        {$_("Top.Config")}
      </NavLi>
    {/if}
  </NavUl>
  {#if !latest}
    <Badge border color="red">{$_("Top.HasUpdate")}</Badge>
  {/if}
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
    class="fex fex-col w-full"
    style="height:{mainHeight}px;"
  >
    <div style="height: {mainHeight - window.innerHeight/5}px">
      <Map />
    </div>
    <div style="width: 99vw;margin: 0 auto;">
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
{:else if page == "address"}
  <Address />
{:else if page == "ai"}
  <AIList />
{:else if page == "system"}
  <System />
{/if}

{#if showConfig}
  <Config
    on:close={() => {
      page = oldPage;
      updateMapName();
      showConfig = false;
    }}
  />
{/if}
