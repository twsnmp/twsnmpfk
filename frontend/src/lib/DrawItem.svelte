<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    GradientButton,
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
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { _ } from 'svelte-i18n';
  import Help from "./Help.svelte";

  export let id: string = "";
  export let posX = 0;
  export let posY = 0;
  let drawItem: any = undefined;
  let show: boolean = false;
  let image: string = "";
  let nodeID: string = "";
  let pollingID: string = "";
  let pollings : any = [];
  let pollingList :any = [];
  const nodeList : any = [];
  let showHelp = false;
  
  const dispatch = createEventDispatcher();

  const drawItemList = [
    { name: $_('DrawItem.Rect'), value: 0 },
    { name: $_('DrawItem.Ellipse'), value: 1 },
    { name: $_('DrawItem.Label'), value: 2 },
    { name: $_('DrawItem.Image'), value: 3 },
    { name: $_('DrawItem.PollingText'), value: 4 },
    { name: $_('DrawItem.PollingGauge'), value: 5 },
  ];

  const condList = [
    {name:$_('DrawItem.showItemsAllways'), value:0},
    {name:$_('DrawItem.showItemsLow'), value:1},
    {name:$_('DrawItem.showItemsHigh'), value:2},
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
    const p = await SelectFile($_('DrawItem.ImageFile'), true);
    if (p) {
      drawItem.Path = p;
      image = await GetImage(p);
    }
  };
</script>

<Modal bind:open={show} size="lg" dismissable={false} class="w-full" on:on:close={close}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      { $_('DrawItem.EditDrawItem') }
    </h3>
    <Label class="space-y-2">
      <span> { $_('DrawItem.Type') } </span>
      <Select
        items={drawItemList}
        bind:value={drawItem.Type}
        placeholder={ $_('DrawItem.SelectType') }
        disabled={drawItem.ID != ""}
        size="sm"
      />
    </Label>
    {#if drawItem.Type < 2}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>{ $_('DrawItem.Width') }</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.W}
            placeholder={ $_('DrawItem.Width') }
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <span>{ $_('DrawItem.Height') }</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.H}
            placeholder={ $_('DrawItem.Height') }
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <div>{ $_('DrawItem.Color') }</div>
          <input type="color" bind:value={drawItem.Color} />
        </Label>
        <Label class="space-y-2">
          <span> {$_('DrawItem.showCond')} </span>
          <Select
            items={condList}
            bind:value={drawItem.Cond}
            placeholder={$_('DrawItem.selectShowCond')}
            size="sm"
          />
        </Label>
      </div>
    {/if}
    {#if drawItem.Type == 2}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>{ $_('DrawItem.FontSize') }</span>
          <Input
            type="number"
            min={8}
            max={128}
            bind:value={drawItem.Size}
            placeholder={ $_('DrawItem.FontSize') }
            size="sm"
          />
        </Label>
        <div />
        <Label class="space-y-2">
          <div>{ $_('DrawItem.Color') }</div>
          <input type="color" bind:value={drawItem.Color} />
        </Label>
        <Label class="space-y-2">
          <span> {$_('DrawItem.showCond')} </span>
          <Select
            items={condList}
            bind:value={drawItem.Cond}
            placeholder={$_('DrawItem.selectShowCond')}
            size="sm"
          />
        </Label>
      </div>
      <Label class="space-y-2">
        <span>{ $_('DrawItem.Text') }</span>
        <Input
          bind:value={drawItem.Text}
          placeholder={ $_('DrawItem.TextToDisplay') }
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 3}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>{ $_('DrawItem.Width') }</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.W}
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <span>{ $_('DrawItem.Height') }</span>
          <Input
            type="number"
            min={0}
            max={1000}
            bind:value={drawItem.H}
            size="sm"
          />
        </Label>
        <Label class="space-y-2">
          <span> {$_('DrawItem.showCond')} </span>
          <Select
            items={condList}
            bind:value={drawItem.Cond}
            placeholder={$_('DrawItem.selectShowCond')}
            size="sm"
          />
        </Label>
        <GradientButton
          shadow
          class="h-8 mt-6 w-28"
          type="button"
          size="xs"
          color="blue"
          on:click={selectImage}
        >
          <Icon path={icons.mdiImage} size={1} />
          { $_('DrawItem.Select') }
        </GradientButton>
      </div>
      <Label class="space-y-2">
        <span>{ $_('DrawItem.Image') }</span>
        {#if image}
          <img src={image} alt="" class="h-32" />
        {:else}
          <div />
        {/if}
      </Label>
    {/if}
    {#if drawItem.Type >= 4}
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2">
          <span>{ $_('DrawItem.Size') }</span>
          <Input
            type="number"
            min={8}
            max={128}
            bind:value={drawItem.Size}
            size="sm"
          />
        </Label>
        <div />
        <div />
        <div />
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2">
          <span> { $_('DrawItem.Node') } </span>
          <Select
            items={nodeList}
            bind:value={nodeID}
            placeholder={ $_('DrawItem.SelectNode') }
            size="sm"
            on:change={updatePollingList}
          />
        </Label>
        <Label class="space-y-2">
          <span> { $_('DrawItem.Polling') } </span>
          <Select
            items={pollingList}
            bind:value={drawItem.PollingID}
            placeholder={ $_('DrawItem.SelectPolling') }
            size="sm"
          />
        </Label>
      </div>
      <Label class="space-y-2">
        <span>{ $_('DrawItem.ValName') }</span>
        <Input
          bind:value={drawItem.VarName}
          placeholder={ $_('DrawItem.ValNamePH') }
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 4}
      <Label class="space-y-2">
        <span>{ $_('DrawItem.TextFormat') }</span>
        <Input
          bind:value={drawItem.Format}
          placeholder={ $_('DrawItem.TextFormatPH') }
          size="sm"
        />
      </Label>
    {/if}
    {#if drawItem.Type == 5}
      <Label class="space-y-2">
        <span>{ $_('DrawItem.GaugeLabel') }</span>
        <Input
          bind:value={drawItem.Text}
          size="sm"
        />
      </Label>
    {/if}
    <Label class="space-y-2">
      <span>{ $_('DrawItem.Zoom') }</span>
      <Input
        type="number"
        min={0.1}
        max={5.0}
        step={0.1}
        bind:value={drawItem.Scale}
        size="sm"
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow color="blue" type="button" on:click={save} size="xs">
        <Icon path={icons.mdiContentSave} size={1} />
        { $_('DrawItem.Save') }
      </GradientButton>
      <GradientButton
      shadow
      type="button"
      size="xs"
      color="lime"
      class="ml-2"
      on:click={() => {
        showHelp = true;
      }}
    >
      <Icon path={icons.mdiHelp} size={1} />
      <span>
        {$_("DrawItem.Help")}
      </span>
    </GradientButton>
    <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('DrawItem.Cancel') }
      </GradientButton>
    </div>
  </form>
</Modal>

<Help bind:show={showHelp} page="drawitem" />

