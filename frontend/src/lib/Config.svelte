<script lang="ts">
  import { Modal,Tabs,TabItem,Checkbox,Label,Input,Select,Button,Alert } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { snmpModeList } from "./common";
  import { GetMapConf, UpdateMapConf,GetNotifyConf,UpdateNotifyConf,TestNotifyConf,GetAIConf,UpdateAIConf } from "../../wailsjs/go/main/App";

  let show: boolean = false;
  let mapConf: datastore.MapConfEnt | undefined = undefined;

  let notifyConf: datastore.NotifyConfEnt | undefined = undefined;
  let showTestError: boolean = false;
  let showTestOk: boolean = false;


  const dispatch = createEventDispatcher();

  onMount(async () => {
    mapConf = await GetMapConf();
    notifyConf = await GetNotifyConf();
    aiConf = await GetAIConf();
    show = true;
  });

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const saveMapConf = async () => {
    mapConf.PollInt *= 1;
    mapConf.Timeout *= 1;
    mapConf.Retry *= 1;
    mapConf.LogDays *= 1;
    const r = await UpdateMapConf(mapConf);
    close();
  };

  const notifyLevelList = [
    { name: "通知しない", value: "none" },
    { name: "注意", value: "warn" },
    { name: "軽度", value: "low" },
    { name: "重度", value: "high" },
  ];

  const saveNotifyConf = async () => {
    notifyConf.Interval *= 1;
    await UpdateNotifyConf(notifyConf);
    close();
  };

  const testNotifyConf = async () => {
    showTestError = false;
    notifyConf.Interval *= 1;
    const ok = await TestNotifyConf(notifyConf);
    showTestError = !ok;
    showTestOk = ok;
  };

  let aiConf: datastore.AIConfEnt | undefined = undefined;

  const aiLevelList = [
    { name: "判定しない", value: 0 },
    { name: "10億回に1回の確率", value: 110.8 },
    { name: "1億回に1回の確率", value: 106.1 },
    { name: "1000万回に1回の確率", value: 101.9 },
    { name: "100万回に1回の確率", value: 97.5 },
    { name: "10万回に1回の確率", value: 92.6 },
    { name: "1万回に1回の確率", value: 86.8 },
    { name: "1000回に1回の確率", value: 80.8 },
    { name: "100回に1回の確率", value: 73.2 },
    { name: "10回に1回の確率", value: 62.8 },
  ];

  const saveAIConf = async () => {
    aiConf.HighThreshold *= 1;
    aiConf.LowThreshold *= 1;
    aiConf.WarnThreshold *= 1;
    await UpdateAIConf(aiConf);
    close();
  };

</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full min-h-[90vh]"
  on:on:close={close}
>
    <Tabs style="underline">
      <TabItem open>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          マップ
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <h3 class="mb-1 font-medium text-gray-900 dark:text-white">マップ設定</h3>
          <Label class="space-y-2">
            <span>マップ名</span>
            <Input
              bind:value={mapConf.MapName}
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
                bind:value={mapConf.PollInt}
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
                bind:value={mapConf.Timeout}
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
                bind:value={mapConf.Retry}
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
                bind:value={mapConf.LogDays}
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 md:grid-cols-3">
            <Label class="space-y-2">
              <span> SNMPモード </span>
              <Select
                items={snmpModeList}
                bind:value={mapConf.SnmpMode}
                placeholder="SNMPのモードを選択"
                size="sm"
              />
            </Label>
            {#if mapConf.SnmpMode == "v1" || mapConf.SnmpMode == "v2c"}
              <Label class="space-y-2">
                <span>SNMP Community</span>
                <Input bind:value={mapConf.Community} placeholder="public" size="sm" />
              </Label>
            {:else}
              <Label class="space-y-2">
                <span>SNMPユーザー</span>
                <Input bind:value={mapConf.SnmpUser} placeholder="snmp user" size="sm" />
              </Label>
              <Label class="space-y-2">
                <span>SNMPパスワード</span>
                <Input
                  type="password"
                  bind:value={mapConf.SnmpPassword}
                  placeholder="•••••"
                  size="sm"
                />
              </Label>
            {/if}
          </div>
          <div class="grid gap-4 mb-4 md:grid-cols-3">
            <Checkbox bind:checked={mapConf.EnableSyslogd}>Syslog受信</Checkbox>
            <Checkbox bind:checked={mapConf.EnableTrapd}>SNMP TRAP受信</Checkbox>
            <Checkbox bind:checked={mapConf.EnableArpWatch}>ARP監視</Checkbox>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <Button type="button" on:click={saveMapConf} size="sm">
              <Icon path={icons.mdiContentSave} size={1} />
              保存
            </Button>
            <Button type="button" color="alternative" on:click={close} size="sm">
              <Icon path={icons.mdiCancel} size={1} />
              キャンセル
            </Button>
          </div>
                </form>
        <!-- <MapConf on:close={close}></MapConf> -->
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          通知
        </div>
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
                bind:value={notifyConf.MailServer}
                placeholder="host|ip:port"
                required
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.InsecureSkipVerify}
              >サーバー証明書をチェックしない</Checkbox
            >
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <Label class="space-y-2">
              <span>ユーザー</span>
              <Input bind:value={notifyConf.User} placeholder="smtp user" size="sm" />
            </Label>
            <Label class="space-y-2">
              <span>パスワード</span>
              <Input
                type="password"
                bind:value={notifyConf.Password}
                placeholder="•••••"
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <Label class="space-y-2">
              <span>送信元</span>
              <Input
                bind:value={notifyConf.MailFrom}
                placeholder="送信元メールアドレス"
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span>宛先</span>
              <Input
                bind:value={notifyConf.MailTo}
                placeholder="宛先メールアドレス"
                size="sm"
              />
            </Label>
          </div>
          <Label class="space-y-2">
            <span> 件名 </span>
            <Input bind:value={notifyConf.Subject} size="sm" />
          </Label>
          <div class="grid gap-4 md:grid-cols-4">
            <Label class="space-y-2">
              <span> 通知レベル </span>
              <Select
                items={notifyLevelList}
                bind:value={notifyConf.Level}
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
                bind:value={notifyConf.Interval}
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.Report}>定期レポート</Checkbox>
            <Checkbox bind:checked={notifyConf.NotifyRepair}>復帰通知</Checkbox>
          </div>
          <Label class="space-y-2">
            <span> コマンド実行 </span>
            <Input class="w-full" bind:value={notifyConf.ExecCmd} size="sm" />
          </Label>
          <div class="flex justify-end space-x-2 mr-2">
            <Button type="button" on:click={saveNotifyConf} size="sm">
              <Icon path={icons.mdiContentSave} size={1} />
              保存
            </Button>
            <Button type="button" color="red" on:click={testNotifyConf} size="sm">
              <Icon path={icons.mdiEmail} size={1} />
              テスト
            </Button>
            <Button type="button" color="alternative" on:click={close} size="sm">
              <Icon path={icons.mdiCancel} size={1} />
              キャンセル
            </Button>
          </div>
        </form>
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          AI分析
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <h3 class="mb-1 font-medium text-gray-900 dark:text-white">AI分析設定</h3>
          <div class="grid gap-4 md:grid-cols-3">
            <Label class="space-y-2">
              <span> 重度と判定するレベル </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.HighThreshold}
                placeholder="レベルを選択"
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> 軽度と判定するレベル </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.LowThreshold}
                placeholder="レベルを選択"
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> 注意と判定するレベル </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.WarnThreshold}
                placeholder="レベルを選択"
                size="sm"
              />
            </Label>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <Button type="button" on:click={saveAIConf} size="sm">
              <Icon path={icons.mdiContentSave} size={1} />
              保存
            </Button>
            <Button type="button" color="alternative" on:click={close} size="sm">
              <Icon path={icons.mdiCancel} size={1} />
              キャンセル
            </Button>
          </div>
        </form>
      </TabItem>
    </Tabs>
</Modal>
