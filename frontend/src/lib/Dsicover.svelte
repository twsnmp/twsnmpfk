<script lang="ts">
  import {
    Progressbar,
    Modal,
    Label,
    Input,
    Checkbox,
    Button,
  } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import {
    GetDiscoverConf,
    GetDiscoverStats,
    StartDiscover,
    StopDiscover,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";

  export let posX = 0;
  export let posY = 0;

  let stats = undefined;
  let conf = undefined;
  let showConf = false;
  let showStats = false;
  let showStop = true;
  let timer: number | undefined = undefined;
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

  onMount(async () => {
    conf = await GetDiscoverConf();
    conf.X = posX;
    conf.Y = posY;
    if (await updateDiscover()) {
      showStats = true;
      showConf = false;
    } else {
      showStats = false;
      showConf = true;
    }
  });

  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  });

  const close = () => {
    dispatch("close", {});
  };

  const start = async () => {
    const r = await StartDiscover(conf);
    if (r) {
      showStats = true;
      showConf = false;
      showStop = true;
      updateDiscover();
    }
  };

  const stop = async () => {
    showStop = false;
    await StopDiscover();
  };
</script>

<Modal bind:open={showConf} size="lg" permanent class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Discover.Discover")}
    </h3>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>{$_("Discover.StartIP")}</span>
        <Input bind:value={conf.StartIP} size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>{$_("Discover.EndIP")}</span>
        <Input bind:value={conf.EndIP} size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
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
      <Label class="space-y-2">
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
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Checkbox bind:checked={conf.PortScan}>{$_("Discover.PortScan")}</Checkbox
      >
      <Checkbox bind:checked={conf.AddPolling}
        >{$_("Discover.AutoAddPolling")}</Checkbox
      >
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <Button type="button" on:click={start} size="xs">
        <Icon path={icons.mdiSearchWeb} size={1} />
        {$_("Discover.Start")}
      </Button>
      <Button type="button" color="alternative" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Discover.Close")}
      </Button>
    </div>
  </form>
</Modal>
<Modal bind:open={showStats} size="lg" permanent class="w-full">
  <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
    {$_("Discover.Stats")}
  </h3>
  <div class="flex flex-col space-y-4">
    <Progressbar
      progress={(stats.Total
        ? ((100 * stats.Sent) / stats.Total).toFixed(2)
        : 0) + ""}
      color="blue"
      size="h-5"
      labelOutside="Total:{stats.Sent + '/' + stats.Total}"
    />
    <Progressbar
      progress={(stats.Total
        ? ((100 * stats.Found) / stats.Total).toFixed(2)
        : 0) + ""}
      color="indigo"
      size="h-5"
      labelOutside="Found:{stats.Found + '/' + stats.Total}"
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
        <Button type="button" color="red" on:click={stop} size="xs">
          <Icon path={icons.mdiStop} size={1} />
          {$_("Discover.Stop")}
        </Button>
      {/if}
      <Button type="button" color="alternative" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Discover.Close")}
      </Button>
    </div>
  </div>
</Modal>
