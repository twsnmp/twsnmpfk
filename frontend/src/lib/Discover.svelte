<script lang="ts">
  import {
    Progressbar,
    Modal,
    Label,
    Input,
    Checkbox,
    Select,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import {
    GetDiscoverAddressRange,
    GetDiscoverConf,
    GetDiscoverStats,
    GetMapConf,
    StartDiscover,
    StopDiscover,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { snmpModeList } from "./common";

  export let show: boolean = false;
  export let posX = 0;
  export let posY = 0;

  let stats: any = undefined;
  let conf: any = undefined;
  let mapConf: any = undefined;
  let showStats = false;
  let showStop = true;
  let showHelp = false;
  let timer: any = undefined;
  const dispatch = createEventDispatcher();

  let newSnmp = {
    SnmpMode: "v2c",
    Community: "public",
    SnmpUser: "",
    SnmpPassword: "",
  };

  const updateDiscover = async () => {
    stats = await GetDiscoverStats();
    if (!stats.Running) {
      timer = undefined;
      showStop = false;
      return false;
    }
    timer = setTimeout(() => {
      updateDiscover();
    }, 2 * 1000);
    return true;
  };

  const onOpen = async () => {
    mapConf = await GetMapConf();
    conf = await GetDiscoverConf();
    if (!conf.SnmpConfigs) {
      conf.SnmpConfigs = [];
    }
    conf.X = posX;
    conf.Y = posY;
    if (await updateDiscover()) {
      showStats = true;
      show = false;
    } else {
      showStats = false;
      show = true;
    }
  };

  const close = () => {
    show = false;
    showStats = false;
    dispatch("close", {});
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  };

  const addSnmpConfig = () => {
    if (!newSnmp.SnmpMode) return;
    if (newSnmp.SnmpMode === "v1" || newSnmp.SnmpMode === "v2c") {
      if (!newSnmp.Community) return;
    } else {
      if (!newSnmp.SnmpUser) return;
    }
    if (!conf.SnmpConfigs) {
      conf.SnmpConfigs = [];
    }
    conf.SnmpConfigs = [
      ...conf.SnmpConfigs,
      {
        SnmpMode: newSnmp.SnmpMode,
        Community: newSnmp.Community,
        SnmpUser: newSnmp.SnmpUser,
        SnmpPassword: newSnmp.SnmpPassword,
      },
    ];
    newSnmp = {
      SnmpMode: "v2c",
      Community: "public",
      SnmpUser: "",
      SnmpPassword: "",
    };
  };

  const removeSnmpConfig = (index: number) => {
    if (!conf.SnmpConfigs) return;
    conf.SnmpConfigs = conf.SnmpConfigs.filter((_: any, i: number) => i !== index);
  };

  const moveSnmpConfig = (index: number, direction: number) => {
    if (!conf.SnmpConfigs) return;
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= conf.SnmpConfigs.length) return;
    const list = [...conf.SnmpConfigs];
    const temp = list[index];
    list[index] = list[targetIndex];
    list[targetIndex] = temp;
    conf.SnmpConfigs = list;
  };

  const start = async () => {
    conf.Retry *= 1;
    conf.Timeout *= 1;
    conf.AutoLine = Number(conf.AutoLine) || 0;
    if (!conf.SnmpConfigs) {
      conf.SnmpConfigs = [];
    }
    const r = await StartDiscover(conf);
    if (r) {
      showStats = true;
      show = false;
      showStop = true;
      updateDiscover();
    }
  };

  const stop = async () => {
    showStop = false;
    await StopDiscover();
  };

  let ipRanges: any = [];
  let selIPRange = 0;
  const getIPRange = async () => {
    if (ipRanges.length < 1) {
      ipRanges = await GetDiscoverAddressRange();
    }
    if (ipRanges.length < 2) {
      return;
    }
    conf.StartIP = ipRanges[selIPRange];
    conf.EndIP = ipRanges[selIPRange + 1];
    selIPRange += 2;
    if (selIPRange > ipRanges.length / 2) {
      selIPRange = 0;
    }
  };

  $: if (show) {
    onOpen();
  }
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
>
  {#if !conf}
    <div class="text-center mt-10"><Spinner size="16" /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Discover.Discover")}
      </h3>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Discover.StartIP")}</span>
          <Input class="h-8" bind:value={conf.StartIP} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Discover.EndIP")}</span>
          <Input class="h-8" bind:value={conf.EndIP} size="sm" />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span> {$_("Discover.Timeout")} </span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min={1}
            max={120}
            step={1}
            bind:value={conf.Timeout}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span> {$_("Discover.Retry")} </span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min={0}
            max={100}
            step={1}
            bind:value={conf.Retry}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 grid-cols-4">
        <Checkbox bind:checked={conf.PortScan}
          >{$_("Discover.PortScan")}</Checkbox
        >
        <Checkbox bind:checked={conf.AddPolling}
          >{$_("Discover.AutoAddPolling")}</Checkbox
        >
        <Checkbox bind:checked={conf.ReCheck}
          >{$_('Discover.ReChek')}</Checkbox
        >
        <Checkbox bind:checked={conf.AddNetwork}
          >{$_('Discover.AddNetwork')}</Checkbox
        >
        <Checkbox bind:checked={conf.AutoDetect}
          >{$_('Discover.AutoDetect')}</Checkbox
        >
        <Checkbox bind:checked={conf.AutoDetectAI}
          >{$_('Discover.AutoDetectAI')}</Checkbox
        >
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Discover.AutoLine")}</span>
          <Select
            class="h-8 text-xs"
            items={[
              { value: 0, name: $_("Discover.AutoLineNone") },
              { value: 1, name: $_("Discover.AutoLineStrict") },
              { value: 2, name: $_("Discover.AutoLineSpeculative") },
            ]}
            bind:value={conf.AutoLine}
            size="sm"
          />
        </Label>
        <div></div>
      </div>
      <!-- Additional SNMP Configurations -->
      <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-3 space-y-3 bg-gray-50/50 dark:bg-gray-800/50">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
            {$_("Discover.SnmpConfigs")}
          </span>
          {#if mapConf}
            <span class="text-xs text-gray-500 dark:text-gray-400">
              {$_("Discover.MapSnmpConfig")}: <span class="font-mono font-medium text-blue-600 dark:text-blue-400">{mapConf.SnmpMode} ({mapConf.SnmpMode === "v1" || mapConf.SnmpMode === "v2c" ? mapConf.Community : mapConf.SnmpUser})</span>
            </span>
          {/if}
        </div>

        <!-- Configured SNMP list -->
        {#if conf.SnmpConfigs && conf.SnmpConfigs.length > 0}
          <div class="space-y-1.5 max-h-36 overflow-y-auto pr-1">
            {#each conf.SnmpConfigs as snmp, index}
              <div class="flex items-center justify-between p-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded text-xs">
                <div class="flex items-center space-x-3 truncate">
                  <span class="font-semibold text-gray-700 dark:text-gray-300 w-24 truncate">{snmp.SnmpMode}</span>
                  <span class="text-gray-600 dark:text-gray-400 font-mono truncate">
                    {snmp.SnmpMode === "v1" || snmp.SnmpMode === "v2c" ? snmp.Community : snmp.SnmpUser}
                  </span>
                  {#if snmp.SnmpMode !== "v1" && snmp.SnmpMode !== "v2c" && snmp.SnmpPassword}
                    <span class="text-gray-400 font-mono">••••••••</span>
                  {/if}
                </div>
                <div class="flex items-center space-x-1 flex-shrink-0">
                  <button
                    type="button"
                    class="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 rounded disabled:opacity-30"
                    disabled={index === 0}
                    onclick={() => moveSnmpConfig(index, -1)}
                    title={$_("Discover.MoveUp")}
                  >
                    <Icon path={icons.mdiArrowUp} size={0.7} />
                  </button>
                  <button
                    type="button"
                    class="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 rounded disabled:opacity-30"
                    disabled={index === conf.SnmpConfigs.length - 1}
                    onclick={() => moveSnmpConfig(index, 1)}
                    title={$_("Discover.MoveDown")}
                  >
                    <Icon path={icons.mdiArrowDown} size={0.7} />
                  </button>
                  <button
                    type="button"
                    class="p-1 hover:bg-red-50 dark:hover:bg-red-900/30 text-red-500 rounded"
                    onclick={() => removeSnmpConfig(index)}
                    title={$_("Discover.Delete")}
                  >
                    <Icon path={icons.mdiDelete} size={0.7} />
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="text-xs text-gray-400 dark:text-gray-500 italic py-2 text-center border border-dashed border-gray-300 dark:border-gray-700 rounded">
            {$_("Discover.NoAdditionalSnmp")}
          </div>
        {/if}

        <!-- Add new SNMP config form -->
        <div class="pt-2 border-t border-gray-200 dark:border-gray-700">
          <div class="grid gap-2 grid-cols-1 sm:grid-cols-12 items-end">
            <div class="sm:col-span-4">
              <Label class="text-xxs space-y-1">
                <span>{$_("Discover.SnmpMode")}</span>
                <Select
                  class="h-7 text-xs"
                  items={snmpModeList}
                  bind:value={newSnmp.SnmpMode}
                  size="sm"
                />
              </Label>
            </div>
            {#if newSnmp.SnmpMode === "v1" || newSnmp.SnmpMode === "v2c"}
              <div class="sm:col-span-6">
                <Label class="text-xxs space-y-1">
                  <span>{$_("Discover.Community")}</span>
                  <Input class="h-7 text-xs" bind:value={newSnmp.Community} placeholder="public" size="sm" />
                </Label>
              </div>
            {:else}
              <div class="sm:col-span-3">
                <Label class="text-xxs space-y-1">
                  <span>{$_("Discover.SnmpUser")}</span>
                  <Input class="h-7 text-xs" bind:value={newSnmp.SnmpUser} placeholder="user" size="sm" />
                </Label>
              </div>
              <div class="sm:col-span-3">
                <Label class="text-xxs space-y-1">
                  <span>{$_("Discover.SnmpPassword")}</span>
                  <Input class="h-7 text-xs" type="password" bind:value={newSnmp.SnmpPassword} placeholder="•••••" size="sm" />
                </Label>
              </div>
            {/if}
            <div class="sm:col-span-2 flex justify-end">
              <button
                type="button"
                onclick={addSnmpConfig}
                class="w-full h-7 px-2 bg-blue-600 hover:bg-blue-700 text-white rounded text-xs font-medium flex items-center justify-center space-x-1 shadow-sm transition"
              >
                <Icon path={icons.mdiPlus} size={0.7} />
                <span>{$_("Discover.AddSnmp")}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        <GradientButton
          shadow
          color="blue"
          type="button"
          onclick={start}
          size="xs"
        >
          <Icon path={icons.mdiSearchWeb} size={1} />
          {$_("Discover.Start")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="red"
          onclick={getIPRange}
          size="xs"
        >
          <Icon path={icons.mdiMagicStaff} size={1} />
          {$_("Discover.AutoIPRange")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          size="xs"
          color="lime"
          class="ml-2"
          onclick={() => {
            showHelp = true;
          }}
        >
          <Icon path={icons.mdiHelp} size={1} />
          <span>
            {$_("Discover.Help")}
          </span>
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("Discover.Close")}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>
<Modal bind:open={showStats} size="lg" dismissable={false} class="w-full">
  <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
    {$_("Discover.Stats")} - {stats.Now - stats.StartTime}Sec
  </h3>
  <div class="flex flex-col space-y-4">
    <Progressbar
      progress={(stats.Total
        ? ((100 * stats.Sent) / stats.Total).toFixed(2)
        : 0) + ""}
      color="blue"
      size="h-5"
      labelOutside={
        $_("Discover.Total") + ' ' +
        stats.Wait + '/' + stats.Sent + '/' + stats.Total 
      }
    />
    <Progressbar
      progress={(stats.Total
        ? ((100 * stats.Found) / stats.Total).toFixed(2)
        : 0) + ""}
      color="indigo"
      size="h-5"
      labelOutside={
        $_("Discover.Found")
        + stats.Found + '/' + stats.Total
      }
    />
    <Progressbar
      progress={(stats.Found
        ? ((100 * stats.Snmp) / stats.Found).toFixed(2)
        : 0) + ""}
      color="red"
      size="h-5"
      labelOutside="SNMP:{stats.Snmp + '/' + stats.Found}"
    />
    {#if conf.PortScan}
      <div class="grid gap-2 grid-cols-2">
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.Web) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="Web:{stats.Web + '/' + stats.Found}"
          />
        </div>
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.Mail) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="Mail:{stats.Mail + '/' + stats.Found}"
          />
        </div>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.SSH) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="SSH:{stats.SSH + '/' + stats.Found}"
          />
        </div>
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.File) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="File:{stats.File + '/' + stats.Found}"
          />
        </div>
      </div>
      <div class="grid gap-2 grid-cols-2">
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.RDP) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="RDP/VNC:{stats.RDP + '/' + stats.Found}"
          />
        </div>
        <div>
          <Progressbar
            progress={(stats.Found
              ? ((100 * stats.LDAP) / stats.Found).toFixed(2)
              : 0) + ""}
            color="gray"
            size="h-5"
            labelOutside="LDAP/AD:{stats.SSH + '/' + stats.Found}"
          />
        </div>
      </div>
    {/if}
    <div class="flex justify-end space-x-2 mr-2">
      {#if showStop}
        <GradientButton
          shadow
          type="button"
          color="red"
          onclick={stop}
          size="xs"
        >
          <Icon path={icons.mdiStop} size={1} />
          {$_("Discover.Stop")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Discover.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Help bind:show={showHelp} page="discover" />
