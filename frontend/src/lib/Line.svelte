<script lang="ts">
  import { Select, Modal, Label, Input, Button } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import {
    GetNode,
    GetLine,
    UpdateLine,
    DeleteLine,
    GetPollings,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";

  export let nodeID1: string = "";
  export let nodeID2: string = "";
  let node1: datastore.NodeEnt | undefined = undefined;
  let node2: datastore.NodeEnt | undefined = undefined;
  let line: datastore.LineEnt | undefined = undefined;

  let show: boolean = false;
  const dispatch = createEventDispatcher();
  const pollingList = [];
  const pollingList1 = [];
  const pollingList2 = [];

  onMount(async () => {
    node1 = await GetNode(nodeID1);
    node2 = await GetNode(nodeID2);
    const pollings1 = await GetPollings(nodeID1);
    const pollings2 = await GetPollings(nodeID2);
    for (let p of pollings1) {
      pollingList1.push({
        name: p.Name,
        value: p.ID,
      });
      pollingList.push({
        name: p.Name,
        value: p.ID,
      });
    }
    for (let p of pollings2) {
      pollingList2.push({
        name: p.Name,
        value: p.ID,
      });
      pollingList.push({
        name: p.Name,
        value: p.ID,
      });
    }
    line = await GetLine(nodeID1, nodeID2);
    show = true;
  });

  onDestroy(() => {});

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const connect = async () => {
    const r = await UpdateLine(line);
    if (r) {
      close();
    } else {
    }
  };

  const disconnect = async () => {
    const r = await DeleteLine(line.ID);
    if (r) {
      close();
    } else {
    }
  };
</script>

<Modal
  bind:open={show}
  size="lg"
  autoclose={false}
  class="w-full"
  on:on:close={close}
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">ラインの編集</h3>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>ノード1</span>
        <Input
          bind:value={node1.Name}
          placeholder="ノード1名"
          readonly
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>ノード2</span>
        <Input
          bind:value={node2.Name}
          placeholder="ノード2名"
          readonly
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span> ポーリング1 </span>
        <Select
          items={pollingList1}
          bind:value={line.PollingID1}
          placeholder="ノード１のポーリング"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> ポーリング2 </span>
        <Select
          items={pollingList2}
          bind:value={line.PollingID2}
          placeholder="ノード2のポーリング"
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span> 情報のためのポーリング </span>
        <Select
          items={pollingList}
          bind:value={line.PollingID}
          placeholder="情報のポーリング"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>情報</span>
        <Input bind:value={line.Info} placeholder="情報" size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>ラインの太さ</span>
        <Input
          bind:value={line.Width}
          placeholder="ラインの太さ"
          type="number"
          min="1"
          max="5"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>ポート</span>
        <Input bind:value={line.Port} placeholder="ポート1:ポート2" size="sm" />
      </Label>
    </div>
    <div class="flex space-x-2">
      {#if line.ID != ""}
        <Button color="red" type="button" on:click={disconnect} size="sm">
          <Icon path={icons.mdiLanDisconnect} size={1} />
          切断
        </Button>
      {/if}
      <Button color="blue" type="button" on:click={connect} size="sm">
        {#if line.ID != ""}
          <Icon path={icons.mdiContentSave} size={1} />
          更新
        {:else}
          <Icon path={icons.mdiLanConnect} size={1} />
          接続
        {/if}
      </Button>
      <Button color="alternative" type="button" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>
