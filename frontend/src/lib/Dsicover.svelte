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
  export let posX = 0;
  export let posY = 0;

  let stats = undefined;
  let conf = undefined;
  let showConf = false;
  let showStats = false;
  let timer: number | undefined = undefined;
  const dispatch = createEventDispatcher();

  const updateDiscover = async () => {
    stats = await GetDiscoverStats();
    showStats = stats.Running;
    showConf = !stats.Running;
    if (!stats.Running) {
      timer = undefined;
      return;
    }
    timer = setTimeout(() => {
      updateDiscover();
    }, 2 * 1000);
  };

  onMount(async () => {
    conf = await GetDiscoverConf();
    conf.X = posX;
    conf.Y = posY;
    updateDiscover();
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
      close();
    }
  };

  const stop = async () => {
    await StopDiscover();
    close();
  };
</script>

<Modal bind:open={showConf} size="lg" permanent class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">自動発見</h3>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>開始IP</span>
        <Input
          bind:value={conf.StartIP}
          placeholder="開始IP"
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>終了IP</span>
        <Input
          bind:value={conf.EndIP}
          placeholder="終了IP"
          required
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> タイムアウト(秒) </span>
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
        <span> リトライ(回) </span>
        <Input
          type="number"
          min={0}
          max={100}
          step={1}
          bind:value={conf.Retry}
          size="sm"
        />
      </Label>
      <Checkbox bind:checked={conf.AddPolling}>ポーリング自動設定</Checkbox>
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <Button type="button" on:click={start} size="sm">
        <Icon path={icons.mdiRun} size={1} />
        開始
      </Button>
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </form>
</Modal>
<Modal bind:open={showStats} size="lg" permanent class="w-full">
  <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
    自動発見の進行状況
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
    <div class="flex justify-end space-x-2 mr-2">
      <Button type="button" on:click={stop} size="sm">
        <Icon path={icons.mdiRun} size={1} />
        停止
      </Button>
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>
