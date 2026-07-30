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
    GetMapConf,
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
  import CertMonitor from "./CertMonitor.svelte";
  import PKI from "./PKI.svelte";
  import OTel from "./OTel.svelte";
  import Mqtt from "./Mqtt.svelte";
  import Help from "./Help.svelte";
  import { _ } from "svelte-i18n";
  import Location from "./Location.svelte";
  import type { datastore } from "wailsjs/go/models";
  import { setIconToList } from "./common";

  let dark: boolean = false;
  let mainHeight = 0;
  let mapConfig :any = {
    MapName: "",
  };
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
    mapConfig = await GetMapConf();
    const l = await GetIcons();
    if (l) {
      for (const icon of l) {
        setIconToList(icon);
      }
    }
    version = await GetVersion();
    await tick();
    mainHeight = window.innerHeight - 82;
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

<svelte:window onresize={() => (mainHeight = window.innerHeight - 82)} />

<Navbar fluid={true} class="!px-3 !py-1.5 w-full flex-nowrap shrink-0 overflow-hidden items-center" style="--wails-draggable:drag">
  {#snippet children({ hidden, toggle })}
  <NavBrand href="/" class="shrink-0 mr-2 md:mr-3.5 min-w-0 flex items-center">
    <img src={logo} class="mr-1.5 h-8 w-auto shrink-0" alt="TWSNMP Logo" />
    <span
      class="self-center font-semibold dark:text-white whitespace-nowrap text-xs sm:text-sm md:text-base lg:text-lg truncate max-w-[130px] sm:max-w-[180px] md:max-w-[240px] lg:max-w-[320px] shrink min-w-0"
    >
      TWSNMP FK <span class="text-[10px] font-normal text-gray-500 dark:text-gray-400 mx-0.5">{version}</span> - {mapConfig.MapName}
    </span>
  </NavBrand>
  <NavUl activeUrl={showConfig ? "config" : page} classes={{ ul: "flex flex-row items-center space-x-0 md:space-x-0.5 text-[10px] md:text-xs font-medium shrink-0 p-0 m-0" }}>
    {#if !lock}
      <NavLi
  href="map"
  onclick={(e) => {
    e.preventDefault();
    page = "map";
  }}
>
        <Icon path={icons.mdiLan} size={1.5} />
        {$_("Top.Map")}
      </NavLi>
      {#if locConf.Style}
        <NavLi
  href="loc"
  onclick={(e) => {
    e.preventDefault();
    page = "loc";
  }}
>
          <Icon path={icons.mdiMap} size={1.5} />
          {$_("Top.Loc")}
        </NavLi>
      {/if}
      <NavLi
  href="node"
  onclick={(e) => {
    e.preventDefault();
    page = "node";
  }}
>
        <Icon path={icons.mdiLaptop} size={1.5} />
        {$_("Top.Node")}
      </NavLi>
      <NavLi
  href="polling"
  onclick={(e) => {
    e.preventDefault();
    page = "polling";
  }}
>
        <Icon path={icons.mdiLanCheck} size={1.5} />
        {$_("Top.Polling")}
      </NavLi>
      <NavLi
  href="address"
  onclick={(e) => {
    e.preventDefault();
    page = "address";
  }}
>
        <Icon path={icons.mdiListStatus} size={1.5} />
        {$_("Top.Address")}
      </NavLi>
      <NavLi
  href="cert"
  onclick={(e) => {
    e.preventDefault();
    page = "cert";
  }}
>
        <Icon path={icons.mdiInvoiceList} size={1.5} />
        {$_('Top.Cert')}
      </NavLi>
      <NavLi
  href="pki"
  onclick={(e) => {
    e.preventDefault();
    page = "pki";
  }}
>
        <Icon path={icons.mdiCertificate} size={1.5} />
        PKI
      </NavLi>
      <NavLi
  href="eventlog"
  onclick={(e) => {
    e.preventDefault();
    page = "eventlog";
  }}
>
        <Icon path={icons.mdiCalendarCheck} size={1.5} />
        {$_("Top.Log")}
      </NavLi>
    {#if mapConfig.EnableSyslogd}
      <NavLi
  href="syslog"
  onclick={(e) => {
    e.preventDefault();
    page = "syslog";
  }}
>
        <Icon path={icons.mdiCalendarText} size={1.5} />
        syslog
      </NavLi>
    {/if}
    {#if mapConfig.EnableTrapd}
      <NavLi
  href="trap"
  onclick={(e) => {
    e.preventDefault();
    page = "trap";
  }}
>
        <Icon path={icons.mdiAlert} size={1.5} />
        TRAP
      </NavLi>
    {/if}
    {#if mapConfig.EnableNetflowd}
      <NavLi
  href="netflow"
  onclick={(e) => {
    e.preventDefault();
    page = "netflow";
  }}
>
        <Icon path={icons.mdiCompareHorizontal} size={1.5} />
        NetFlow
      </NavLi>
    {/if}
    {#if mapConfig.EnableSFlowd}
      <NavLi
  href="sflow"
  onclick={(e) => {
    e.preventDefault();
    page = "sflow";
  }}
>
        <Icon path={icons.mdiClockCheckOutline} size={1.5} />
        sFlow
      </NavLi>
    {/if}
      <NavLi
  href="arp"
  onclick={(e) => {
    e.preventDefault();
    page = "arp";
  }}
>
        <Icon path={icons.mdiCheckNetwork} size={1.5} />
        ARP
      </NavLi>
    {#if mapConfig.EnableOTel}
      <NavLi
  href="otel"
  onclick={(e) => {
    e.preventDefault();
    page = "otel";
  }}
>
        <Icon path={icons.mdiTelescope} size={1.5} />
        OTel
      </NavLi>
    {/if}
    {#if mapConfig.EnableMqtt}
      <NavLi
  href="mqtt"
  onclick={(e) => {
    e.preventDefault();
    page = "mqtt";
  }}
>
        <Icon path={icons.mdiQueueFirstInLastOut} size={1.5} />
        MQTT
      </NavLi>
    {/if}
      <NavLi
  href="ai"
  onclick={(e) => {
    e.preventDefault();
    page = "ai";
  }}
>
        <Icon path={icons.mdiBrain} size={1.5} />
        {$_("Top.AI")}
      </NavLi>
      <NavLi
  href="system"
  onclick={(e) => {
    e.preventDefault();
    page = "system";
  }}
>
        <Icon path={icons.mdiChartLine} size={1.5} />
        {$_("Top.System")}
      </NavLi>
      <NavLi
  href="config"
  onclick={(e) => {
    e.preventDefault();
    oldPage = page;
    page = "";
    showConfig = true;
  }}
>
        <Icon path={icons.mdiCog} size={1.5} />
        {$_("Top.Config")}
      </NavLi>
    {/if}
  </NavUl>
  <div class="flex items-center shrink-0 ml-auto gap-1 md:gap-1.5 pl-2">
    {#if !latest}
      <Badge class="mr-0.5 h-8 text-xs px-2.5 py-1 whitespace-nowrap flex items-center justify-center border" color="red">{$_("Top.HasUpdate")}</Badge>
    {/if}
    <Button class="!p-1.5 h-8 w-8 flex items-center justify-center shrink-0" color="alternative" onclick={toggleDark}>
      {#if dark}
        <Icon path={icons.mdiWeatherSunny} size={1} />
      {:else}
        <Icon path={icons.mdiMoonWaxingCrescent} size={1} />
      {/if}
    </Button>
    <Button
      class="!p-1.5 h-8 w-8 flex items-center justify-center shrink-0"
      color="alternative"
      onclick={() => {
        oldPage = page;
        page = "";
        showHelp = true;
      }}
    >
      <Icon path={icons.mdiHelp} size={1} />
    </Button>
  </div>
  {/snippet}
</Navbar>

{#if page == "map"}
  <div class="flex flex-col w-full" style="height:{mainHeight}px;">
    <div class="relative" style="height: {mainHeight - window.innerHeight / 5}px">
      <Map />
    </div>
    <div class="w-full px-2">
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
{:else if page == "otel"}
  <OTel />
{:else if page == "mqtt"}
  <Mqtt />
{:else if page == "cert"}
  <CertMonitor />
{/if}

<Config
  bind:show={showConfig}
  on:close={async () => {
    page = oldPage;
    mapConfig = await GetMapConf()
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

<style>
  :global(nav ul li a),
  :global(nav ul li button) {
    padding-top: 0.25rem !important;
    padding-bottom: 0.25rem !important;
    padding-left: 0.35rem !important;
    padding-right: 0.35rem !important;
    display: flex !important;
    flex-direction: column !important;
    align-items: center !important;
  }
</style>
