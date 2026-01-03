<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import {
    GetPolling,
    UpdatePolling,
    GetAutoPollings,
    GetNodes,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { levelList, typeList, logModeList, getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { CodeJar } from "@novacbn/svelte-codejar";
  import Prism from "prismjs";
  import "prismjs/components/prism-regex";

  Prism.languages.grok = {
    number: /%\{.+?\}/,
    string: /\.\+/,
    regex: /\\s\+/,
  };

  Prism.languages.twaction = {
    regex: /[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}/,
    keyword: /(wol|mail|webhook|wait|cmd)/,
    number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
    string: /\b(?:false|true|up|down)\b/,
    url: /https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)/,
  };

  const highlight = (code: string, syntax: string | undefined) => {
    if (!syntax) {
      return "";
    }
    return Prism.highlight(code, Prism.languages[syntax], syntax);
  };

  export let show: boolean = false;
  export let nodeID: string = "";
  export let pollingID: string = "";
  export let pollingTmpID: number = 0;
  export let pollingTmp: any = undefined;

  let polling: any = undefined;
  let list: any = [];
  let showList: boolean = false;
  let showHelp = false;

  const nodeList: any = [];
  const dispatch = createEventDispatcher();

  const onOpen = async () => {
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
    } else if (pollingTmp) {
      polling = pollingTmp;
    } else if (pollingTmpID) {
      list = await GetAutoPollings(nodeID, pollingTmpID);
      if (!list) {
        close();
        return;
      }
      if (list.length == 1) {
        polling = list[0];
      } else {
        showPollingList();
        showList = true;
      }
    } else {
      close();
      return;
    }
  };

  let pollingTable: any = undefined;
  let selectedCount = 0;

  const showPollingList = async () => {
    await tick();
    selectedCount = 0;
    pollingTable = new DataTable("#pollingTable", {
      destroy: true,
      stateSave: true,
      data: list,
      language: getTableLang(),
      order: [[0, "desc"]],
      select: {
        style: "single",
      },
      columns: [
        {
          data: "Name",
          title: $_("Polling.Name"),
          width: "35%",
        },
        {
          data: "Type",
          title: $_("Polling.Type"),
          width: "10%",
        },
        {
          data: "Mode",
          title: $_("Polling.Mode"),
          width: "10%",
        },
        {
          data: "Params",
          title: $_("Polling.Params"),
          width: "10%",
        },
        {
          data: "Filter",
          title: $_("Polling.Filter"),
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

  const select = () => {
    const p = pollingTable.rows({ selected: true }).data();
    if (!p || p.length != 1) {
      return;
    }
    polling = p[0];
    selectedCount = 0;
    showList = false;
  };

  const close = () => {
    showList = false;
    show = false;
    dispatch("close", {});
  };

  let paramsColor: any = "base";
  let filterColor: any = "base";
  const save = async () => {
    filterColor = "base";
    paramsColor = "base";
    if (polling) {
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
      }
    }
  };
  
  const aiModeList = [
    { name: $_('Polling.Default'), value: "default" },
    { name: "Isolation Forest", value: "iforest" },
  ];

</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if !polling}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Polling.EditPolling")}
      </h3>
      <div class="grid gap-4 mb-4 grid-cols-2">
        {#if !nodeID}
          <Label class="space-y-2 text-xs">
            <span> {$_("Polling.Node")} </span>
            <Select
              items={nodeList}
              bind:value={polling.NodeID}
              placeholder={$_("Polling.SelectNode")}
              size="sm"
            />
          </Label>
        {/if}
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.Name")}</span>
          <Input class="h-8" bind:value={polling.Name} required size="sm" />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 grid-cols-4">
        <Label class="space-y-2 text-xs">
          <span> {$_("Polling.Level")} </span>
          <Select
            items={levelList}
            bind:value={polling.Level}
            placeholder={$_("Polling.SelectLevel")}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span> {$_("Polling.Type")} </span>
          <Select
            items={typeList}
            bind:value={polling.Type}
            placeholder={$_("Polling.SelectPollingType")}
            size="sm"
            disabled={pollingID != ""}
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.Mode")}</span>
          <Input class="h-8" bind:value={polling.Mode} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span> {$_("Polling.LogMode")} </span>
          <Select
            items={logModeList}
            bind:value={polling.LogMode}
            placeholder={$_("Polling.SelectLogMode")}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 grid-cols-3">
          <Label class="space-y-2 text-xs">
          <span>MQTTサーバーURL</span>
          <Input
            class="h-8"
            bind:value={polling.MqttURL}
            placeholder="tcp://localhost:1883"
            size="sm"
          />
        </Label>
          <Label class="space-y-2 text-xs">
          <span>トピック</span>
          <Input
            class="h-8"
            bind:value={polling.MqttTopic}
            placeholder="twsnmpfk/polling"
            size="sm"
          />
        </Label>
          <Label class="space-y-2 text-xs">
          <span>送信するデータ名</span>
          <Input
            class="h-8"
            bind:value={polling.MqttCols}
            size="sm"
          />
        </Label>
      </div>
      {#if polling.LogMode == 3}
      <div class="grid gap-4 mb-4 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_('Polling.AIMode')}</span>
          <Select
            items={aiModeList}
            bind:value={polling.AIMode}
            placeholder={$_('Polling.SelectAIMode')}
            size="sm"
          />
        </Label>

        <Label class="space-y-2 text-xs">
          <span>{$_('Polling.VectorCols')}</span>
          <Input
            class="h-8"
            bind:value={polling.VectorCols}
            placeholder={$_('Polling.VectorColsHelp')}
            size="sm"
          />
        </Label>
      </div>
      {/if}
      <div class="grid gap-4 mb-4 grid-cols-2">
          <Label class="space-y-2 text-xs">
          <span>{$_("Polling.Params")}</span>
          <Input
            class="h-8"
            bind:value={polling.Params}
            placeholder={$_("Polling.Params")}
            color={paramsColor}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs" color={filterColor}>
          <span>{$_("Polling.Filter")}</span>
          <CodeJar syntax="regex" {highlight} bind:value={polling.Filter}/>
        </Label>
      </div>
      <div class="grid gap-4 mb-4 grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.GrokPat")}</span>
          <CodeJar syntax="grok" {highlight} bind:value={polling.Extractor}/>
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.Script")}</span>
          <CodeJar syntax="javascript" {highlight} bind:value={polling.Script}/>
        </Label>
      </div>
      <div class="grid gap-4 grid-cols-3">
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.IntSec")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="5"
            max="86400"
            bind:value={polling.PollInt}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.TimeoutSec")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="0"
            max="3600"
            bind:value={polling.Timeout}
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Polling.Retry")}</span>
          <Input
            class="h-8 w-24 text-right"
            type="number"
            min="0"
            max="50"
            bind:value={polling.Retry}
            size="sm"
          />
        </Label>
      </div>
      <Label class="space-y-2 text-xs">
        <span>{$_('Polling.FailAction')}</span>
        <CodeJar syntax="twaction" {highlight} bind:value={polling.FailAction}/>
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_('Polling.RepairAction')}</span>
        <CodeJar syntax="twaction" {highlight} bind:value={polling.RepairAction}/>
      </Label>
      <div class="flex justify-end space-x-2 mr-2">
        <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={save}
          size="xs"
        >
          <Icon path={icons.mdiContentSave} size={1} />
          {$_("Polling.Save")}
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
        <GradientButton
          shadow
          type="button"
          color="teal"
          on:click={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("Polling.Cancel")}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>

<Modal bind:open={showList} size="xl" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <table id="pollingTable" class="display compact mt-2" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedCount == 1}
        <GradientButton
          shadow
          type="button"
          color="blue"
          on:click={select}
          size="xs"
        >
          <Icon path={icons.mdiCheck} size={1} />
          {$_("Polling.Select")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Polling.Cancel")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Help bind:show={showHelp} page="editpolling" />
