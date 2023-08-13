<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Button,
    Checkbox,
  } from "flowbite-svelte";

  import { onMount, createEventDispatcher,onDestroy } from "svelte";
  import { GetMapConf, UpdateMapConf } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { snmpModeList } from "./common";

  let show: boolean = false;
  let conf: datastore.MapConfEnt | undefined = undefined;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    conf = await GetMapConf();
    show = true;
  });
  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    conf.PollInt *= 1;
    conf.Timeout *= 1;
    conf.Retry *= 1;
    conf.LogDays *= 1;
    const r = await UpdateMapConf(conf);
    close();
  };
</script>

<Modal
  bind:open={show}
  size="lg"
  permanent
  class="w-full"
  on:on:close={close}
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">マップ設定</h3>
    <Label class="space-y-2">
      <span>マップ名</span>
      <Input
        bind:value={conf.MapName}
        placeholder="マップ名"
        required
        size="sm"
      />
    </Label>
    <div class="grid gap-4 mb-4 md:grid-cols-4">
      <Label class="space-y-2">
        <span> ポーリング間隔(秒) </span>
        <Input
          type="number"
          min={5}
          max={3600 * 24}
          step={1}
          bind:value={conf.PollInt}
          size="sm"
        />
      </Label>
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
      <Label class="space-y-2">
        <span> ログ保存日数(日) </span>
        <Input
          type="number"
          min={1}
          max={365 * 5}
          step={1}
          bind:value={conf.LogDays}
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> SNMPモード </span>
        <Select
          items={snmpModeList}
          bind:value={conf.SnmpMode}
          placeholder="SNMPのモードを選択"
          size="sm"
        />
      </Label>
      {#if conf.SnmpMode == "v1" || conf.SnmpMode == "v2c"}
        <Label class="space-y-2">
          <span>SNMP Community</span>
          <Input bind:value={conf.Community} placeholder="public" size="sm" />
        </Label>
      {:else}
        <Label class="space-y-2">
          <span>SNMPユーザー</span>
          <Input bind:value={conf.SnmpUser} placeholder="snmp user" size="sm" />
        </Label>
        <Label class="space-y-2">
          <span>SNMPパスワード</span>
          <Input
            type="password"
            bind:value={conf.SnmpPassword}
            placeholder="•••••"
            size="sm"
          />
        </Label>
      {/if}
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Checkbox bind:checked={conf.EnableSyslogd}>Syslog受信</Checkbox>
      <Checkbox bind:checked={conf.EnableTrapd}>SNMP TRAP受信</Checkbox>
      <Checkbox bind:checked={conf.EnableArpWatch}>ARP監視</Checkbox>
    </div>
    <div class="flex space-x-2">
      <Button type="button" on:click={save} size="sm">
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
