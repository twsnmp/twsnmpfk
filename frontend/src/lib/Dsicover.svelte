<script lang="ts">
  import {
    Progressbar,
    Modal,
    Label,
    Input,
    Checkbox,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import {
    GetDiscoverAddressRange,
    GetDiscoverConf,
    GetDiscoverStats,
    StartDiscover,
    StopDiscover,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";

  export let show: boolean = false;
  export let posX = 0;
  export let posY = 0;

  let stats: any = undefined;
  let conf: any = undefined;
  let showStats = false;
  let showStop = true;
  let showHelp = false;
  let timer: any = undefined;
  const dispatch = createEventDispatcher();

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
    conf = await GetDiscoverConf();
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

  const start = async () => {
    conf.Retry *= 1;
    conf.Timeout *= 1;
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
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if !conf}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Discover.Discover")}
      </h3>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Discover.StartIP")}</span>
          <Input bind:value={conf.StartIP} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Discover.EndIP")}</span>
          <Input bind:value={conf.EndIP} size="sm" />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span> {$_("Discover.Timeout")} </span>
          <Input
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
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={start}
          size="xs"
        >
          <Icon path={icons.mdiSearchWeb} size={1} />
          {$_("Discover.Start")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="red"
          on:click={getIPRange}
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
          on:click={() => {
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
          on:click={close}
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
        $_("Discover.Total") + ' '
        + stats.Sent +'(' + stats.Wait + ')' + '/' + stats.Total 
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
          on:click={stop}
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
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Discover.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Help bind:show={showHelp} page="discover" />
