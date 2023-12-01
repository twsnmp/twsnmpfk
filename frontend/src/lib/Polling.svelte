<script context="module">
  import Prism from "prismjs";
  Prism.languages.grok = {
    number: /%\{.+?\}/,
    string: /\.\+/,
    regex: /\\s\+/,
  };
  const highlight = (code, syntax) =>
    Prism.highlight(code, Prism.languages[syntax], syntax);
</script>

<script lang="ts">
  import { CodeJar } from "@novacbn/svelte-codejar";

  import { Select, Modal, Label, Input, GradientButton } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher, tick } from "svelte";
  import {
    GetPolling,
    UpdatePolling,
    GetAutoPollings,
    GetNodes,
  } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { levelList, typeList, logModeList, getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";

  export let nodeID: string = "";
  export let pollingID: string = "";
  export let pollingTmpID: number = 0;
  export let pollingTmp = undefined;

  let polling: datastore.PollingEnt | undefined = undefined;
  let show: boolean = false;
  let list = [];
  let showList: boolean = false;
  let showHelp = false;

  const nodeList = [];
  const dispatch = createEventDispatcher();

  onMount(async () => {
    const nodes = await GetNodes();
    for (const k in nodes) {
      nodeList.push({
        name: nodes[k].Name,
        value: k,
      });
    }
    if (pollingID) {
      polling = await GetPolling(pollingID);
      nodeID = polling.NodeID;
      show = true;
    } else if (pollingTmp) {
      polling = pollingTmp;
      show = true;
    } else if (pollingTmpID) {
      list = await GetAutoPollings(nodeID, pollingTmpID);
      if (list.length == 1) {
        polling = list[0];
        show = true;
      } else {
        showPollingList();
        showList = true;
      }
    } else {
      close();
      return;
    }
  });

  let pollingTable = undefined;
  let selectedCount = 0;

  const showPollingList = async () => {
    if (pollingTable) {
      pollingTable.destroy();
      pollingTable = undefined;
    }
    await tick();
    selectedCount = 0;
    pollingTable = new DataTable("#pollingTable", {
      data: list,
      language: getTableLang(),
      order: [[1, "desc"]],
      select: {
        style: "multi",
      },
      columns: [
        {
          data: "Name",
          title: $_('Polling.Name'),
          width: "35%",
        },
        {
          data: "Type",
          title: $_('Polling.Type'),
          width: "10%",
        },
        {
          data: "Mode",
          title: $_('Polling.Mode'),
          width: "10%",
        },
        {
          data: "Params",
          title: $_('Polling.Params'),
          width: "10%",
        },
        {
          data: "Filter",
          title: $_('Polling.Filter'),
          width: "10%",
        },
      ],
    });
    pollingTable.on("select", () => {
      selectedCount = pollingTable.rows({ selected: true }).count();
    });
    pollingTable.on("deselect", () => {
      selectedCount = pollingTable.rows({ selected: true }).count();
    });
  };

  onDestroy(() => {
    if (pollingTable) {
      pollingTable.destroy();
    }
  });

  const select = () => {
    const p = pollingTable.rows({ selected: true }).data();
    if (!p || p.length != 1) {
      return;
    }
    polling = p[0];
    selectedCount = 0;
    showList = false;
    show = true;
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };
  let paramsColor: any = "base";
  let filterColor: any = "base";
  const save = async () => {
    filterColor = "base";
    paramsColor = "base";
    if (polling.Filter.startsWith("TODO:")) {
      filterColor = "red";
      return;
    }
    if (polling.Params.startsWith("TODO:")) {
      paramsColor = "red";
      return;
    }
    polling.Extractor.replaceAll("\n", "");
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
      { $_('Polling.EditPolling') }
    </h3>
    {#if !nodeID}
      <Label class="space-y-2">
        <span> { $_('Polling.Node') } </span>
        <Select
          items={nodeList}
          bind:value={polling.NodeID}
          placeholder="{ $_('Polling.SelectNode') }"
          size="sm"
        />
      </Label>
    {/if}
    <Label class="space-y-2">
      <span>{ $_('Polling.Name') }</span>
      <Input
        bind:value={polling.Name}
        required
        size="sm"
      />
    </Label>
    <div class="grid gap-4 mb-4 md:grid-cols-4">
      <Label class="space-y-2">
        <span> { $_('Polling.Level') } </span>
        <Select
          items={levelList}
          bind:value={polling.Level}
          placeholder={ $_('Polling.SelectLevel') }
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> { $_('Polling.Type') } </span>
        <Select
          items={typeList}
          bind:value={polling.Type}
          placeholder={ $_('Polling.SelectPollingType') }
          size="sm"
          disabled={pollingID != ""}
        />
      </Label>
      <Label class="space-y-2">
        <span>{ $_('Polling.Mode') }</span>
        <Input bind:value={polling.Mode} size="sm" />
      </Label>
      <Label class="space-y-2">
        <span> { $_('Polling.LogMode') } </span>
        <Select
          items={logModeList}
          bind:value={polling.LogMode}
          placeholder={ $_('Polling.SelectLogMode') }
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2">
      <span>{ $_('Polling.Params') }</span>
      <Input
        bind:value={polling.Params}
        placeholder={ $_('Polling.Params') }
        color={paramsColor}
        size="sm"
      />
    </Label>
    <Label class="space-y-2">
      <span>{ $_('Polling.Filter') }</span>
      <Input
        bind:value={polling.Filter}
        placeholder={ $_('Polling.Filter') }
        color={filterColor}
        size="sm"
      />
    </Label>
    <Label class="space-y-2">
      <span>{ $_('Polling.GrokPat') }</span>
      <CodeJar syntax="grok" {highlight} bind:value={polling.Extractor} />
    </Label>
    <Label class="space-y-2">
      <span>{ $_('Polling.Script') }</span>
      <CodeJar syntax="javascript" {highlight} bind:value={polling.Script} />
    </Label>
    <div class="grid gap-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span>{ $_('Polling.IntSec') }</span>
        <Input
          type="number"
          min="5"
          max="3600"
          bind:value={polling.PollInt}
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>{ $_('Polling.TimeoutSec') }</span>
        <Input
          type="number"
          min="0"
          max="3600"
          bind:value={polling.Timeout}
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>{ $_('Polling.Retry') }</span>
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
      <GradientButton shadow color="blue" type="button" on:click={save} size="xs">
        <Icon path={icons.mdiContentSave} size={1} />
        { $_('Polling.Save') }
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
          {$_("Polling.Help")}
        </span>
      </GradientButton>
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('Polling.Cancel') }
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal
  bind:open={showList}
  size="xl"
  permanent
  class="w-full"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <table id="pollingTable" class="display compact mt-2" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedCount == 1}
        <GradientButton shadow type="button" color="blue" on:click={select} size="xs">
          <Icon path={icons.mdiCheck} size={1} />
          { $_('Polling.Select') }
        </GradientButton>
      {/if}
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('Polling.Cancel') }
      </GradientButton>
    </div>
  </div>
</Modal>

{#if showHelp}
  <Help
    page="editpolling"
    on:close={() => {
      showHelp = false;
    }}
  />
{/if}

<style>
  @import "prismjs/themes/prism.css";
</style>
