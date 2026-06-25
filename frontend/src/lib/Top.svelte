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
    mainHeight = window.innerHeight - 96;
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

<svelte:window onresize={() => (mainHeight = window.innerHeight - 96)} />

<Navbar style="--wails-draggable:drag">
  {#snippet children({ hidden, toggle })}
  <NavBrand href="/">
    <img src={logo} class="mr-2 h-12" alt="TWSNMP Logo" />
    <span
      class="self-center whitespace-nowrap text-xl font-semibold dark:text-white"
    >
      TWSNMP FK <span class="text-xs font-normal text-gray-500 dark:text-gray-400 mx-1">{version}</span> - {mapConfig.MapName}
    </span>
  </NavBrand>
  <NavUl activeUrl={showConfig ? "config" : page} classes={{ ul: "flex flex-col p-2 mt-3 md:flex-row md:space-x-1 rtl:space-x-reverse md:mt-0 md:text-xs md:font-medium" }}>
    {#if !lock}
      <NavLi
  href="map"
  onclick={(e) => {
    e.preventDefault();
    page = "map";
  }}
>
        <Icon path={icons.mdiLan} size={1.8} />
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
          <Icon path={icons.mdiMap} size={1.8} />
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
        <Icon path={icons.mdiLaptop} size={1.8} />
        {$_("Top.Node")}
      </NavLi>
      <NavLi
  href="polling"
  onclick={(e) => {
    e.preventDefault();
    page = "polling";
  }}
>
        <Icon path={icons.mdiLanCheck} size={1.8} />
        {$_("Top.Polling")}
      </NavLi>
      <NavLi
  href="address"
  onclick={(e) => {
    e.preventDefault();
    page = "address";
  }}
>
        <Icon path={icons.mdiListStatus} size={1.8} />
        {$_("Top.Address")}
      </NavLi>
      <NavLi
  href="cert"
  onclick={(e) => {
    e.preventDefault();
    page = "cert";
  }}
>
        <Icon path={icons.mdiInvoiceList} size={1.8} />
        {$_('Top.Cert')}
      </NavLi>
      <NavLi
  href="pki"
  onclick={(e) => {
    e.preventDefault();
    page = "pki";
  }}
>
        <Icon path={icons.mdiCertificate} size={1.8} />
        PKI
      </NavLi>
      <NavLi
  href="eventlog"
  onclick={(e) => {
    e.preventDefault();
    page = "eventlog";
  }}
>
        <Icon path={icons.mdiCalendarCheck} size={1.8} />
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
        <Icon path={icons.mdiCalendarText} size={1.8} />
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
        <Icon path={icons.mdiAlert} size={1.8} />
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
        <Icon path={icons.mdiCompareHorizontal} size={1.8} />
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
        <Icon path={icons.mdiClockCheckOutline} size={1.8} />
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
        <Icon path={icons.mdiCheckNetwork} size={1.8} />
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
        <Icon path={icons.mdiTelescope} size={1.8} />
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
        <Icon path={icons.mdiQueueFirstInLastOut} size={1.8} />
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
        <Icon path={icons.mdiBrain} size={1.8} />
        {$_("Top.AI")}
      </NavLi>
      <NavLi
  href="system"
  onclick={(e) => {
    e.preventDefault();
    page = "system";
  }}
>
        <Icon path={icons.mdiChartLine} size={1.8} />
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
        <Icon path={icons.mdiCog} size={1.8} />
        {$_("Top.Config")}
      </NavLi>
    {/if}
  </NavUl>
  <div class="flex justify-end">
    {#if !latest}
      <Badge class="mr-1 h-8" border color="red">{$_("Top.HasUpdate")}</Badge>
    {/if}
    <Button class="!p-2" color="alternative" onclick={toggleDark}>
      {#if dark}
        <Icon path={icons.mdiWeatherSunny} size={1} />
      {:else}
        <Icon path={icons.mdiMoonWaxingCrescent} size={1} />
      {/if}
    </Button>
    <Button
      class="!p-2 ml-1"
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
