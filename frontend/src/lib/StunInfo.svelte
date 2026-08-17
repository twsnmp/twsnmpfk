<script lang="ts">
  import {
    Modal,
    GradientButton,
    Label,
    Spinner,
    Input,
    Select,
    Alert,
  } from "flowbite-svelte";
  import {
    getTableLang,
    renderState,
  } from "./common";
  import { GetStunInfo } from "../../wailsjs/go/main/App";
  import { createEventDispatcher, tick } from "svelte";
  import DataTable from "datatables.net-dt";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import { copyText } from "svelte-copy";

  export let show: boolean = false;

  const dispatch = createEventDispatcher();

  let stunServer = "stun.cloudflare.com:3478";
  let protocol = "udp4";
  let timeout = 5;

  const serverPresets = [
    { name: "Cloudflare (stun.cloudflare.com:3478)", value: "stun.cloudflare.com:3478" },
    { name: "Google 1 (stun.l.google.com:19302)", value: "stun.l.google.com:19302" },
    { name: "Google 2 (stun1.l.google.com:19302)", value: "stun1.l.google.com:19302" },
    { name: "Tencent (stun.qq.com:3478)", value: "stun.qq.com:3478" },
  ];

  const protoList = [
    { name: "IPv4 (UDP)", value: "udp4" },
    { name: "IPv6 (UDP)", value: "udp6" },
  ];

  let stunResult: any = undefined;
  let stunEntries: any = [];
  let wait = false;
  let errorMessage = "";
  let latLong = "";
  let isGlobalIP = false;
  let table: any = undefined;
  let copied = false;

  const columns = [
    {
      data: "Level",
      title: $_('Address.State') || 'State',
      width: "15%",
      render: renderState,
    },
    {
      data: "Title",
      title: $_('Address.Name') || 'Name',
      width: "35%",
    },
    {
      data: "Value",
      title: $_('Address.Value') || 'Value',
      width: "50%",
    },
  ];

  const showStunTable = () => {
    table = new DataTable("#stunInfoTable", {
      destroy: true,
      columns: columns,
      pageLength: 10,
      stateSave: false,
      data: stunEntries,
      language: getTableLang(),
    });
  };

  const fetchStunInfo = async () => {
    wait = true;
    errorMessage = "";
    latLong = "";
    isGlobalIP = false;
    stunResult = undefined;
    stunEntries = [];

    try {
      const res = await GetStunInfo(stunServer, protocol, timeout);
      stunResult = res;
      if (res.Error) {
        errorMessage = res.Error;
      }
      stunEntries = res.Entries || [];

      if (res.IP) {
        if (res.Location && !res.Location.includes("LOCAL")) {
          const a = res.Location.split(",");
          if (a.length > 2) {
            latLong = a[1] + "," + a[2];
          }
          isGlobalIP = true;
        }
      }
    } catch (e: any) {
      errorMessage = String(e);
    } finally {
      wait = false;
    }

    if (stunEntries && stunEntries.length > 0) {
      await tick();
      showStunTable();
    }
  };

  const onOpen = () => {
    if (!stunResult) {
      fetchStunInfo();
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const copy = () => {
    let s: string[] = [];
    const h = columns.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < stunEntries.length; i++) {
      const row: any = [];
      row.push(stunEntries[i].Level);
      row.push(stunEntries[i].Title);
      row.push(stunEntries[i].Value);
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };

  $: if (show) {
    onOpen();
  }
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[85vh]"
>
  <div class="flex flex-col space-y-4">
    <div class="flex items-center space-x-2">
      <Icon path={icons.mdiWan} size={1.2} class="text-blue-500" />
      <h3 class="font-medium text-gray-900 dark:text-white text-lg">
        {$_('Stun.Title') || 'STUN Global IP Information'}
      </h3>
    </div>

    <!-- Query Parameters Form -->
    <div class="grid gap-3 grid-cols-1 md:grid-cols-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
      <Label class="space-y-1 text-xs md:col-span-2">
        <span>{$_('Stun.Server') || 'STUN Server'}</span>
        <Input class="h-8" bind:value={stunServer} placeholder="stun.cloudflare.com:3478" size="sm" />
      </Label>
      <Label class="space-y-1 text-xs">
        <span>{$_('Stun.Protocol') || 'Protocol'}</span>
        <Select items={protoList} bind:value={protocol} size="sm" />
      </Label>
      <Label class="space-y-1 text-xs">
        <span>{$_('Stun.Timeout') || 'Timeout (sec)'}</span>
        <Input class="h-8" type="number" min="1" max="30" bind:value={timeout} size="sm" />
      </Label>
    </div>

    {#if wait}
      <div class="text-center my-12"><Spinner size="16" /></div>
    {:else}
      {#if errorMessage}
        <Alert color="red" class="my-2">
          <div class="flex items-center space-x-2">
            <Icon path={icons.mdiAlertCircle} size={1} />
            <span>{errorMessage}</span>
          </div>
        </Alert>
      {/if}

      {#if stunResult && stunResult.IP}
        <!-- Key Metrics Cards -->
        <div class="grid gap-3 grid-cols-2 sm:grid-cols-4 my-2">
          <div class="p-3 bg-blue-50 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-800 rounded-lg">
            <div class="text-xs text-blue-600 dark:text-blue-400 font-medium">{$_('Stun.GlobalIP') || 'Global IP'}</div>
            <div class="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate" title={stunResult.IP}>
              {stunResult.IP}
            </div>
          </div>
          <div class="p-3 bg-green-50 dark:bg-green-950/40 border border-green-200 dark:border-green-800 rounded-lg">
            <div class="text-xs text-green-600 dark:text-green-400 font-medium">{$_('Stun.ReverseDNS') || 'Reverse DNS'}</div>
            <div class="text-sm sm:text-base font-semibold text-gray-900 dark:text-white truncate" title={stunResult.Hostname || '-'}>
              {stunResult.Hostname || '-'}
            </div>
          </div>
          <div class="p-3 bg-purple-50 dark:bg-purple-950/40 border border-purple-200 dark:border-purple-800 rounded-lg">
            <div class="text-xs text-purple-600 dark:text-purple-400 font-medium">{$_('Stun.MappedPort') || 'Mapped Port'}</div>
            <div class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">
              {stunResult.Port}
            </div>
          </div>
          <div class="p-3 bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800 rounded-lg">
            <div class="text-xs text-amber-600 dark:text-amber-400 font-medium">{$_('Stun.RTT') || 'Response Time'}</div>
            <div class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">
              {stunResult.RTT}
            </div>
          </div>
        </div>
      {/if}

      <!-- Wrapped DataTable -->
      <div class="overflow-x-auto my-2">
        <table id="stunInfoTable" class="display compact w-full"></table>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-wrap justify-end items-center gap-2 pt-2 border-t dark:border-gray-700">
        <GradientButton
          type="button"
          color="blue"
          onclick={fetchStunInfo}
          size="xs"
        >
          <Icon path={icons.mdiRecycle} size={1} />
          {$_('Stun.Fetch') || 'Fetch'}
        </GradientButton>

        {#if stunEntries.length > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            onclick={copy}
            size="xs"
          >
            {#if copied}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            {$_('Stun.Copy') || 'Copy'}
          </GradientButton>
        {/if}

        {#if isGlobalIP && stunResult}
          {#if latLong}
            <GradientButton
              shadow
              type="button"
              color="lime"
              size="xs"
              onclick={() => {
                BrowserOpenURL(
                  `https://www.google.com/maps/search/?api=1&query=` + latLong
                );
              }}
            >
              <Icon path={icons.mdiGoogleMaps} size={1} />
              Google MAP
            </GradientButton>
          {/if}
          <GradientButton
            shadow
            type="button"
            color="lime"
            size="xs"
            onclick={() => {
              BrowserOpenURL(
                `https://www.virustotal.com/gui/ip-address/` + stunResult.IP
              );
            }}
          >
            <Icon path={icons.mdiVirus} size={1} />
            VirusTotal
          </GradientButton>
        {/if}

        <GradientButton type="button" color="teal" onclick={close} size="xs">
          <Icon path={icons.mdiCancel} size={1} />
          {$_('ArpReport.Close') || 'Close'}
        </GradientButton>
      </div>
    {/if}
  </div>
</Modal>
