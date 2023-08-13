<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Checkbox,
    Button,
    P,
  } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import {
    GetDrawItem,
    UpdateDrawItem,
    GetPollings,
    GetImage,
    SelectFile,
    GetNodes,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";

  export let id: string = "";
  export let posX = 0;
  export let posY = 0;
  let drawItem: datastore.DrawItemEnt | undefined = undefined;
  let show: boolean = false;
  let image: string = "";
  let nodeID: string = "";
  let pollingID: string = "";
  let pollings : datastore.PollingEnt[] = [];
  let pollingList = [];
  const nodeList = [];
  
  const dispatch = createEventDispatcher();

  const drawItemList = [
    { name: "矩形", value: 0 },
    { name: "楕円", value: 1 },
    { name: "ラベル", value: 2 },
    { name: "イメージ", value: 3 },
    { name: "ポーリング結果(テキスト)", value: 4 },
    { name: "ポーリング結果(ゲージ)", value: 5 },
  ];

  onMount(async () => {
    pollings = await GetPollings("");
    const nodes = await GetNodes();
    for(const k in nodes) {
      nodeList.push({
        name: nodes[k].Name,
        value: k,
      });
    }
    drawItem = await GetDrawItem(id);
    if (id == "") {
      drawItem.X = posX;
      drawItem.Y = posY;
    } else {
      if (drawItem.PollingID) {
        nodeID = "";
        for(const p of pollings) {
          if(p.ID == drawItem.PollingID) {
            nodeID = p.NodeID;
            updatePollingList();
            break
          }
        }
      }
    }
    if (drawItem.Path) {
      image = await GetImage(drawItem.Path);
    }
    show = true;
  });

  const updatePollingList = async () => {
    pollingList = [];
    for (let p of pollings) {
      if(nodeID == p.NodeID) {
        pollingList.push({
          name: p.Name,
          value: p.ID,
        });
      }
    }
    pollingID = drawItem.PollingID;
  }
  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    drawItem.W *= 1;
    drawItem.H *= 1;
    drawItem.Size *= 1;
    drawItem.Scale *= 1;
    drawItem.X *= 1;
    drawItem.Y *= 1;
    const r = await UpdateDrawItem(drawItem);
    if (r) {
      close();
    } else {
    }
  };

  const selectImage = async () => {
    const p = await SelectFile("描画アイテム画像ファイル", true);
    if (p) {
      drawItem.Path = p;
      image = await GetImage(p);
    }
  };
</script>

<Modal bind:open={show} size="lg" permanent class="w-full" on:on:close={close}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      描画アイテムの編集
    </h3>
    <Label class="space-y-2">
      <span> 種類 </span>
      <Select
        items={drawItemList}
        bind:value={drawItem.Type}
        placeholder="描画アイテムを選択"
        size="sm"
      />
    </Label>
    {#if drawItem.Type < 2}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>幅</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.W}
            placeholder="幅"
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <span>高さ</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.H}
            placeholder="高さ"
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <div>色</div>
          <input type="color" bind:value={drawItem.Color} />
        </Label>
        <div />
      </div>
    {/if}
    {#if drawItem.Type == 2}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>文字サイズ</span>
          <Input
            type="number"
            min={8}
            max={128}
            bind:value={drawItem.Size}
            placeholder="文字サイズ"
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <div>色</div>
          <input type="color" bind:value={drawItem.Color} />
        </Label>
        <div />
        <div />
      </div>
      <Label class="space-y-2">
        <span>文字列</span>
        <Input
          bind:value={drawItem.Text}
          placeholder="表示する文字列"
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 3}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>幅</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.W}
            placeholder="幅"
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <span>高さ</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.H}
            placeholder="高さ"
            size="sm"
          />
        </Label>
        <Button
          class="h-10 mt-4 w-32"
          type="button"
          size="sm"
          color="blue"
          on:click={selectImage}
        >
          <Icon path={icons.mdiImage} size={1} />
          選択
        </Button>
        <div />
      </div>
      <Label class="space-y-2">
        <span>イメージ</span>
        {#if image}
          <img src={image} alt="" class="h-32" />
        {:else}
          <div />
        {/if}
      </Label>
    {/if}
    {#if drawItem.Type == 4 || drawItem.Type == 5}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>サイズ</span>
          <Input
            type="number"
            min={8}
            max={128}
            bind:value={drawItem.Size}
            placeholder="サイズ"
            size="sm"
          />
        </Label>
        <div />
        <div />
        <div />
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2">
          <span> ノード </span>
          <Select
            items={nodeList}
            bind:value={nodeID}
            placeholder="ノードを選択"
            size="sm"
            on:change={updatePollingList}
          />
        </Label>
        <Label class="space-y-2">
          <span> ポーリング </span>
          <Select
            items={pollingList}
            bind:value={drawItem.PollingID}
            placeholder="ポーリングを選択"
            size="sm"
          />
        </Label>
      </div>
      <Label class="space-y-2">
        <span>変数名</span>
        <Input
          bind:value={drawItem.VarName}
          placeholder="変数名(空欄は自動設定)"
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 4}
      <Label class="space-y-2">
        <span>表示フォーマット</span>
        <Input
          bind:value={drawItem.Format}
          placeholder="表示フォーマット(空欄は自動設定)"
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 5}
      <Label class="space-y-2">
        <span>ゲージのラベル</span>
        <Input
          bind:value={drawItem.Text}
          placeholder="ゲージのラベル"
          size="sm"
        />
      </Label>
    {/if}
    <Label class="space-y-2">
      <span>倍率</span>
      <Input
        type="number"
        min={0.1}
        max={5.0}
        step={0.1}
        bind:value={drawItem.Scale}
        placeholder="倍率"
        size="sm"
      />
    </Label>
    <div class="flex space-x-2">
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
