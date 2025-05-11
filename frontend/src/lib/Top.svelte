<script lang="ts">
  import logo from "../assets/images/appicon.png";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    Button,
    Badge,
    Tooltip,
  } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetMapName,
    IsDark,
    IsLatest,
    SetDark,
    GetSettings,
    GetLocConf,
    GetIcons,
    GetVersion,
  } from "../../wailsjs/go/main/App";
  import Map from "./Map.svelte";
  import Log from "./Log.svelte";
  import NodeList from "./NodeList.svelte";
  import PollingList from "./PollingList.svelte";
  import EventLog from "./EventLog.svelte";
  import Syslog from "./Syslog.svelte";
  import Trap from "./Trap.svelte";
  import NetFlow from "./NetFlow.svelte";
  import SFlow from "./SFlow.svelte";
  import Arp from "./Arp.svelte";
  import Address from "./Address.svelte";
  import AIList from "./AIList.svelte";
  import Config from "./Config.svelte";
  import System from "./System.svelte";
  import PKI from "./PKI.svelte";
  import Help from "./Help.svelte";
  import { _ } from "svelte-i18n";
  import Location from "./Location.svelte";
  import type { datastore } from "wailsjs/go/models";
  import { setIconToList } from "./common";
  import { Style } from "maplibre-gl";

  let dark: boolean = false;
  let mainHeight = 0;
  let mapName = "";
  let page = "map";
  let oldPage = "";
  let showConfig = false;
  let showHelp = false;
  let latest = true;
  let lock = "";
  let version = "";
  let locConf: datastore.LocConfEnt = {
    Style: "",
    IconSize: 24,
    Zoom: 2,
    Center: "",
  };

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
    if (e) {
      if (await IsDark()) {
        e.classList.add("dark");
        dark = true;
      } else {
        e.classList.remove("dark");
        dark = false;
      }
    }
    const settings = await GetSettings();
    lock = settings.Lock;
    locConf = await GetLocConf();
    if (lock == "loc" && locConf.Style != "") {
      page = "loc";
    }
    const l = await GetIcons();
    if (l) {
      for (const icon of l) {
        setIconToList(icon);
      }
    }
    version = await GetVersion();
    await tick();
    mainHeight = window.innerHeight - 96;
    updateMapName();
    checkLatest();
  });

  const toggleDark = () => {
    const e = document.querySelector("html");
    if(e) {
      e.classList.toggle("dark");
      dark = e.classList.contains("dark");
      SetDark(dark);
    };
    }
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
  <NavUl ulClass="flex flex-col p-3 mt-3 md:flex-row md:space-x-5 rtl:space-x-reverse md:mt-0 md:text-xs md:font-medium">
    {#if !lock}
      <NavLi
        active={page == "map"}
        on:click={() => {
          page = "map";
        }}
      >
        <Icon path={icons.mdiLan} size={1.8} />
        {$_("Top.Map")}
      </NavLi>
      {#if locConf.Style}
        <NavLi
          active={page == "loc"}
          on:click={() => {
            page = "loc";
          }}
        >
          <Icon path={icons.mdiMap} size={1.8} />
          {$_("Top.Loc")}
        </NavLi>
      {/if}
      <NavLi
        active={page == "node"}
        on:click={() => {
          page = "node";
        }}
      >
        <Icon path={icons.mdiLaptop} size={1.8} />
        {$_("Top.Node")}
      </NavLi>
      <NavLi
        active={page == "polling"}
        on:click={() => {
          page = "polling";
        }}
      >
        <Icon path={icons.mdiLanCheck} size={1.8} />
        {$_("Top.Polling")}
      </NavLi>
      <NavLi
        active={page == "address"}
        on:click={() => {
          page = "address";
        }}
      >
        <Icon path={icons.mdiListStatus} size={1.8} />
        {$_("Top.Address")}
      </NavLi>
      <NavLi
        active={page == "pki"}
        on:click={() => {
          page = "pki";
        }}
      >
        <Icon path={icons.mdiCertificate} size={1.8} />
        PKI
      </NavLi>
      <NavLi
        active={page == "eventlog"}
        on:click={() => {
          page = "eventlog";
        }}
      >
        <Icon path={icons.mdiCalendarCheck} size={1.8} />
        {$_("Top.Log")}
      </NavLi>
      <NavLi
        active={page == "syslog"}
        on:click={() => {
          page = "syslog";
        }}
      >
        <Icon path={icons.mdiCalendarText} size={1.8} />
        syslog
      </NavLi>
      <NavLi
        active={page == "trap"}
        on:click={() => {
          page = "trap";
        }}
      >
        <Icon path={icons.mdiAlert} size={1.8} />
        TRAP
      </NavLi>
      <NavLi
        active={page == "netflow"}
        on:click={() => {
          page = "netflow";
        }}
      >
        <Icon path={icons.mdiCompareHorizontal} size={1.8} />
        NetFlow
      </NavLi>
      <NavLi
        active={page == "sflow"}
        on:click={() => {
          page = "sflow";
        }}
      >
        <Icon path={icons.mdiClockCheckOutline} size={1.8} />
        sFlow
      </NavLi>
      <NavLi
        active={page == "arp"}
        on:click={() => {
          page = "arp";
        }}
      >
        <Icon path={icons.mdiCheckNetwork} size={1.8} />
        ARP
      </NavLi>
      <NavLi
        active={page == "ai"}
        on:click={() => {
          page = "ai";
        }}
      >
        <Icon path={icons.mdiBrain} size={1.8} />
        {$_("Top.AI")}
      </NavLi>
      <NavLi
        active={page == "system"}
        on:click={() => {
          page = "system";
        }}
      >
        <Icon path={icons.mdiChartLine} size={1.8} />
        {$_("Top.System")}
      </NavLi>
      <NavLi
        active={showConfig}
        on:click={() => {
          oldPage = page;
          page = "";
          showConfig = true;
        }}
      >
        <Icon path={icons.mdiCog} size={1.8} />
        {$_("Top.Config")}
      </NavLi>
    {/if}
  </NavUl>
  <div class="flex justify-right">
    {#if !latest}
      <Badge class="mr-2 h-8" border color="red">{$_("Top.HasUpdate")}</Badge>
    {/if}
    <Button class="!p-2" color="alternative" on:click={toggleDark}>
      {#if dark}
        <Icon path={icons.mdiWeatherSunny} size={1} />
      {:else}
        <Icon path={icons.mdiMoonWaxingCrescent} size={1} />
      {/if}
    </Button>
    <Button
      id="help"
      class="!p-2 ml-2"
      color="alternative"
      on:click={() => {
        oldPage = page;
        page = "";
        showHelp = true;
      }}
    >
      <Icon path={icons.mdiHelp} size={1} />
    </Button>
    <Tooltip triggeredBy="#help" placement="left">{version}</Tooltip>
  </div>
</Navbar>

{#if page == "map"}
  <div class="fex fex-col w-full" style="height:{mainHeight}px;">
    <div class="relative" style="height: {mainHeight - window.innerHeight / 5}px">
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
{:else if page == "netflow"}
  <NetFlow />
{:else if page == "sflow"}
  <SFlow />
{:else if page == "arp"}
  <Arp />
{:else if page == "address"}
  <Address />
{:else if page == "ai"}
  <AIList />
{:else if page == "system"}
  <System />
{:else if page == "loc"}
  <Location />
{:else if page == "pki"}
  <PKI />
{/if}

<Config
  bind:show={showConfig}
  on:close={async () => {
    page = oldPage;
    updateMapName();
    locConf = await GetLocConf();
    if (page == "loc" && !locConf.Style) {
      page = "map";
    }
  }}
/>

<Help 
  bind:show={showHelp} 
  page = {oldPage}
  on:close={()=> {
    page = oldPage;
  }}
/>
