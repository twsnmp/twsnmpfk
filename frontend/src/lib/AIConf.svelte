<script lang="ts">
  import { Select, Modal, Label, Button } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import { GetAIConf, UpdateAIConf } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";

  let show: boolean = false;
  let conf: datastore.AIConfEnt | undefined = undefined;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    conf = await GetAIConf();
    show = true;
  });

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

  const save = async () => {
    conf.HighThreshold *= 1;
    conf.LowThreshold *= 1;
    conf.WarnThreshold *= 1;
    await UpdateAIConf(conf);
    close();
  };

  const close = () => {
    show = false;
    dispatch("close", {});
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
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">AI分析設定</h3>
    <div class="grid gap-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> 重度と判定するレベル </span>
        <Select
          items={aiLevelList}
          bind:value={conf.HighThreshold}
          placeholder="レベルを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> 軽度と判定するレベル </span>
        <Select
          items={aiLevelList}
          bind:value={conf.LowThreshold}
          placeholder="レベルを選択"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> 注意と判定するレベル </span>
        <Select
          items={aiLevelList}
          bind:value={conf.WarnThreshold}
          placeholder="レベルを選択"
          size="sm"
        />
      </Label>
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
