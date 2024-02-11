<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import {
    GetNode,
    GetLine,
    UpdateLine,
    DeleteLine,
    GetPollings,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { createEventDispatcher } from "svelte";

  export let show: boolean = false;
  export let nodeID1: string = "";
  export let nodeID2: string = "";
  let node1: any = undefined;
  let node2: any = undefined;
  let line: any = undefined;

  const dispatch = createEventDispatcher();

  let showHelp = false;

  const pollingList: any = [];
  const pollingList1: any = [];
  const pollingList2: any = [];

  const onOpen = async () => {
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
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const connect = async () => {
    const r = await UpdateLine(line);
    if (r) {
      close();
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
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if !line}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Line.EditLine")}
      </h3>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Node1")}</span>
          <Input bind:value={node1.Name} readonly size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Node2")}</span>
          <Input bind:value={node2.Name} readonly size="sm" />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span> {$_("Line.Polling1")} </span>
          <Select
            items={pollingList1}
            bind:value={line.PollingID1}
            placeholder={$_("Line.Node1Polling")}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span> {$_("Line.Polling2")} </span>
          <Select
            items={pollingList2}
            bind:value={line.PollingID2}
            placeholder={$_("Line.Node2Polling")}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-4 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span> {$_("Line.InfoPolling")} </span>
          <Select
            items={pollingList}
            bind:value={line.PollingID}
            placeholder={$_("Line.InfoPolling")}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Info")}</span>
          <Input bind:value={line.Info} size="sm" />
        </Label>
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.LineWidth")}</span>
          <Input
            bind:value={line.Width}
            type="number"
            min="1"
            max="5"
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Port")}</span>
          <Input bind:value={line.Port} size="sm" />
        </Label>
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if line.ID != ""}
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={disconnect}
            size="xs"
          >
            <Icon path={icons.mdiLanDisconnect} size={1} />
            {$_("LIne.Disconnect")}
          </GradientButton>
        {/if}
        {#if line.ID != ""}
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={connect}
            size="xs"
          >
            <Icon path={icons.mdiContentSave} size={1} />
            {$_("Line.Update")}
          </GradientButton>
        {:else}
          <GradientButton
            color="blue"
            type="button"
            on:click={connect}
            size="xs"
          >
            <Icon path={icons.mdiLanConnect} size={1} />
            {$_("Line.Connect")}
          </GradientButton>
        {/if}
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
            {$_("Line.Help")}
          </span>
        </GradientButton>
        <GradientButton
          shadow
          color="teal"
          type="button"
          on:click={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("Line.Cancel")}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>

<Help bind:show={showHelp} page="line" />
