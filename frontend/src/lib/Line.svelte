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
    GetLineByID,
    UpdateLine,
    DeleteLine,
    GetPollings,
    GetNetwork,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { createEventDispatcher } from "svelte";

  export let show: boolean = false;
  export let id: string = "";
  export let nodeID1: string = "";
  export let nodeID2: string = "";
  let node1: any = undefined;
  let node2: any = undefined;
  let line: any = undefined;
  let net1:boolean = false;
  let net2:boolean = false;
  let wait:boolean = false;

  const dispatch = createEventDispatcher();

  let showHelp = false;

  const pollingList: any = [];
  const pollingList1: any = [];
  const pollingList2: any = [];

  const onOpen = async () => {
    pollingList.length = 0;
    pollingList1.length = 0;
    pollingList2.length = 0;
    wait = true;
    if (id != "") {
      line = await GetLineByID(id);
    } else {
      line = await GetLine(nodeID1,nodeID2);
    }
    nodeID1 = line.NodeID1;
    nodeID2 = line.NodeID2;
    net1 = nodeID1.startsWith("NET:");
    net2 = nodeID2.startsWith("NET:");
    if (net1) {
      node1 = await GetNetwork(nodeID1);
      for (let p of node1.Ports) {
        pollingList1.push({
          name: p.Name,
          value: p.ID,
        });
      }
    } else {
      node1 = await GetNode(nodeID1);
      const pollings = await GetPollings(nodeID1);
      for (let p of pollings) {
        pollingList1.push({
          name: p.Name,
          value: p.ID,
        });
        pollingList.push({
          name: p.Name,
          value: p.ID,
        });
      }
    }
    if (net2) {
      node2 = await GetNetwork(nodeID2);
      for (let p of node2.Ports) {
        pollingList2.push({
          name: p.Name,
          value: p.ID,
        });
      }
    } else {
      node2 = await GetNode(nodeID2);
      const pollings = await GetPollings(nodeID2);
      for (let p of pollings) {
        pollingList2.push({
          name: p.Name,
          value: p.ID,
        });
        pollingList.push({
          name: p.Name,
          value: p.ID,
        });
      }
    }
    wait = false;
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
  {#if wait}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Line.EditLine")}
      </h3>
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Node1")}</span>
          <Input class="h-8" bind:value={node1.Name} readonly size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.Node2")}</span>
          <Input class="h-8" bind:value={node2.Name} readonly size="sm" />
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
      {#if !net1 || !net2 }
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
            <Input class="h-8" bind:value={line.Info} size="sm" />
          </Label>
        </div>
      {/if}
      <div class="grid gap-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Line.LineWidth")}</span>
          <Input
            class="h-8 w-24 text-right"
            bind:value={line.Width}
            type="number"
            min="1"
            max="5"
            size="sm"
          />
        </Label>
        {#if !net1 || !net2}
          <Label class="space-y-2 text-xs">
            <span>{$_("Line.Port")}</span>
            <Input class="h-8" bind:value={line.Port} size="sm" />
          </Label>
        {/if}
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
            {$_("Line.Disconnect")}
          </GradientButton>
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
