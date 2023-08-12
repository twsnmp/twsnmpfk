<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Button,
    Checkbox,
    Alert,
  } from "flowbite-svelte";

  import { onMount, createEventDispatcher } from "svelte";
  import {
    GetNotifyConf,
    UpdateNotifyConf,
    TestNotifyConf,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";

  let show: boolean = false;
  let conf: datastore.NotifyConfEnt | undefined = undefined;
  let showTestError: boolean = false;
  let showTestOk: boolean = false;

  const dispatch = createEventDispatcher();

  const notifyLevelList = [
    { name: "通知しない", value: "none" },
    { name: "注意", value: "warn" },
    { name: "軽度", value: "low" },
    { name: "重度", value: "high" },
  ];

  const save = async () => {
    conf.Interval *= 1;
    await UpdateNotifyConf(conf);
    show = false;
    close();
  };

  const test = async () => {
    showTestError = false;
    conf.Interval *= 1;
    const ok = await TestNotifyConf(conf);
    showTestError = !ok;
    showTestOk = ok;
  };

  onMount(async () => {
    conf = await GetNotifyConf();
    show = true;
  });

  const close = () => {
    show = false;
    dispatch("close", {});
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
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">通知設定</h3>
    {#if showTestError}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          通知メールの送信テストに失敗しました
        </div>
      </Alert>
    {/if}
    {#if showTestOk}
      <Alert class="flex" color="blue" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          通知メールの送信テストに成功しました
        </div>
      </Alert>
    {/if}
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>メールサーバー</span>
        <Input
          bind:value={conf.MailServer}
          placeholder="host|ip:port"
          required
          size="sm"
        />
      </Label>
      <Checkbox bind:checked={conf.InsecureSkipVerify}
        >サーバー証明書をチェックしない</Checkbox
      >
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>ユーザー</span>
        <Input bind:value={conf.User} placeholder="smtp user" size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>パスワード</span>
        <Input
          type="password"
          bind:value={conf.Password}
          placeholder="•••••"
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>送信元</span>
        <Input
          bind:value={conf.MailFrom}
          placeholder="送信元メールアドレス"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>宛先</span>
        <Input
          bind:value={conf.MailTo}
          placeholder="宛先メールアドレス"
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2">
      <span> 件名 </span>
      <Input bind:value={conf.Subject} size="sm" />
    </Label>
    <div class="grid gap-4 md:grid-cols-4">
      <Label class="space-y-2">
        <span> 通知レベル </span>
        <Select
          items={notifyLevelList}
          bind:value={conf.Level}
          placeholder="通知レベルを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> 通知間隔(秒) </span>
        <Input
          type="number"
          min={60}
          max={3600 * 24}
          step={10}
          bind:value={conf.Interval}
          size="sm"
        />
      </Label>
      <Checkbox bind:checked={conf.Report}>定期レポート</Checkbox>
      <Checkbox bind:checked={conf.NotifyRepair}>復帰通知</Checkbox>
    </div>
    <Label class="space-y-2">
      <span> コマンド実行 </span>
      <Input class="w-full" bind:value={conf.ExecCmd} size="sm" />
    </Label>
    <div class="flex space-x-3">
      <Button type="button" on:click={save} size="sm">
        <Icon path={icons.mdiContentSave} size={1} />
        保存
      </Button>
      <Button type="button" color="red" on:click={test} size="sm">
        <Icon path={icons.mdiEmail} size={1} />
        テスト
      </Button>
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>
