<script context="module">
  import Prism from "prismjs";
  const highlight = (code, syntax) =>
    Prism.highlight(code, Prism.languages[syntax], syntax);
</script>

<script lang="ts">
  import { CodeJar } from "@novacbn/svelte-codejar";

  import { Select, Modal, Label, Input, Button } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import {
    GetNode,
    GetPolling,
    GetGroks,
    UpdatePolling,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { levelList, typeList, logModeList } from "./common";

  export let nodeID: string = "";
  export let pollingID: string = "";
  let node: datastore.NodeEnt | undefined = undefined;
  let polling: datastore.PollingEnt | undefined = undefined;
  let show: boolean = false;
  let extractorList = [
    {
      name: "goqueryによるデータ取得",
      value: "goquery",
    },
    {
      name: "getBodyによるデータ取得",
      value: "getBody",
    },
  ];
  const dispatch = createEventDispatcher();

  onMount(async () => {
    if (pollingID) {
      polling = await GetPolling(pollingID);
      node = await GetNode(polling.NodeID);
    } else if (nodeID) {
      node = await GetNode(nodeID);
      polling = {
        ID: "",
        Name: "新規ポーリング",
        NodeID: nodeID,
        Type: "ping",
        Mode: "",
        Params: "",
        Filter: "",
        Extractor: "",
        Script: "",
        Level: "",
        PollInt: 60,
        Timeout: 1,
        Retry: 1,
        LogMode: 0,
        NextTime: 0,
        LastTime: 0,
        Result: {},
        State: "unkown",
      };
    } else {
      close();
      return;
    }
    const groks = await GetGroks();
    for (const g of groks) {
      extractorList.push({
        name: g.Name,
        value: g.ID,
      });
    }
    show = true;
  });

  onDestroy(() => {});

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    polling.Timeout *= 1;
    polling.Retry *= 1;
    polling.PollInt *= 1;
    const r = await UpdatePolling(polling);
    if (r) {
      close();
    } else {
    }
  };
</script>

<Modal bind:open={show} size="lg" permanent class="w-full" on:on:close={close}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      ポーリングの編集
    </h3>
    <Label class="space-y-2">
      <span>名前</span>
      <Input
        bind:value={polling.Name}
        placeholder="ポーリングの名前"
        required
        size="sm"
      />
    </Label>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> レベル </span>
        <Select
          items={levelList}
          bind:value={polling.Level}
          placeholder="レベルを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> 種別 </span>
        <Select
          items={typeList}
          bind:value={polling.Type}
          placeholder="アドレスモードを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>モード</span>
        <Input
          bind:value={polling.Mode}
          placeholder="モード"
          required
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>パラメーター</span>
        <Input
          bind:value={polling.Params}
          placeholder="パラメーター"
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>フィルター</span>
        <Input
          bind:value={polling.Filter}
          placeholder="フィルター"
          required
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span> 抽出パターン </span>
        <Select
          items={extractorList}
          bind:value={polling.Extractor}
          placeholder="抽出パターンを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> ログモード </span>
        <Select
          items={logModeList}
          bind:value={polling.LogMode}
          placeholder="ログモードを選択"
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2">
      <span>スクリプト</span>
      <CodeJar syntax="javascript" {highlight} bind:value={polling.Script} />
    </Label>
    <div class="grid gap-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span>ポーリング間隔(秒)</span>
        <Input
          type="number"
          min="5"
          max="3600"
          bind:value={polling.PollInt}
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>タイムアウト(秒)</span>
        <Input
          type="number"
          min="0"
          max="3600"
          bind:value={polling.Timeout}
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>リトライ</span>
        <Input
          type="number"
          min="0"
          max="50"
          bind:value={polling.Retry}
          size="sm"
        />
      </Label>
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <Button color="blue" type="button" on:click={save} size="sm">
        <Icon path={icons.mdiContentSave} size={1} />
        保存
      </Button>
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>

<style>
  @import "prismjs/themes/prism.css";
</style>
