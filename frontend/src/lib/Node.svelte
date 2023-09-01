<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Checkbox,
    Button,
  } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import { GetNode, UpdateNode } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { addrModeList, getIcon, iconList, snmpModeList } from "./common";

  export let nodeID: string = "";
  export let posX = 0;
  export let posY = 0;
  let node: datastore.NodeEnt | undefined = undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    if (nodeID) {
      node = await GetNode(nodeID);
    } else {
      node = {
        ID: "",
        Name: "新規ノード",
        Descr: "",
        Icon: "",
        State: "",
        X: posX,
        Y: posY,
        IP: "",
        IPv6: "",
        MAC: "",
        SnmpMode: "v2c",
        Community: "public",
        User: "",
        Password: "",
        PublicKey: "",
        URL: "",
        AddrMode: "ip",
        AutoAck: false,
      };
    }
    show = true;
  });

  onDestroy(() => {});

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    const r = await UpdateNode(node);
    if (r) {
      close();
    } else {
    }
  };
</script>

<Modal bind:open={show} size="lg" permanent class="w-full" on:on:close={close}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">ノードの編集</h3>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span>名前</span>
        <Input
          bind:value={node.Name}
          placeholder="ノード名"
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>IPアドレス</span>
        <Input
          bind:value={node.IP}
          placeholder="IPアドレス"
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> アドレスモード </span>
        <Select
          items={addrModeList}
          bind:value={node.AddrMode}
          placeholder="アドレスモードを選択"
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> アイコン </span>
        <Select
          items={iconList}
          bind:value={node.Icon}
          placeholder="アイコンを選択"
          size="sm"
        />
      </Label>
      <div class="mt-5 ml-5">
        <span class="mdi {getIcon(node.Icon)} text-4xl" />
      </div>
      <Checkbox bind:checked={node.AutoAck}>復帰時に自動確認</Checkbox>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span> SNMPモード </span>
        <Select
          items={snmpModeList}
          bind:value={node.SnmpMode}
          placeholder="SNMPのモードを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>SNMP Community</span>
        <Input bind:value={node.Community} placeholder="public" size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>ユーザー</span>
        <Input bind:value={node.User} placeholder="ユーザー名" size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>SNMPパスワード</span>
        <Input
          type="password"
          bind:value={node.Password}
          placeholder="•••••"
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2">
      <span>公開鍵</span>
      <Input bind:value={node.PublicKey} placeholder="公開鍵" size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>URL</span>
      <Input bind:value={node.URL} placeholder="URL" size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>説明</span>
      <Input bind:value={node.Descr} placeholder="説明" size="sm" />
    </Label>
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
