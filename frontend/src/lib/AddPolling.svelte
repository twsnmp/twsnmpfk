<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { Modal, GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { tick, createEventDispatcher, onMount } from "svelte";
  import { GetDefaultPolling, GetPollingTemplates,ImportPollingTemplate, GetMapConf } from "../../wailsjs/go/main/App";
  import { getTableLang,renderPollingType } from "./common";
  import Polling from "./Polling.svelte";
  import AIPollingAssistDialog from "./AIPollingAssistDialog.svelte";

  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _, t } from 'svelte-i18n';

  export let nodeID = "";
  export let show = false;
  let hasAI = false;
  const checkAI = async () => {
    const conf = await GetMapConf();
    hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
  };

  onMount(async () => {
    await checkAI();
  });

  $: if (show) {
    checkAI();
  }
  const dispatch = createEventDispatcher();

  let tmpTable :any = undefined;
  let selectedCount = 0;
  let showEditPolling = false;
  let showAIAssist = false;
  let selectedTemplateID = 0;
  let pollingTmp : any  = undefined;

  const showTable = async () => {
    await tick();
    selectedCount = 0;
    tmpTable = new DataTable("#tmpTable", {
      columns: columns,
      data: await GetPollingTemplates(),
      stateSave: true,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    tmpTable.on("select", () => {
      selectedCount = tmpTable.rows({ selected: true }).count();
    });
    tmpTable.on("deselect", () => {
      selectedCount = tmpTable.rows({ selected: true }).count();
    });
  };

  const add = async () => {
    const selected = tmpTable.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    selectedTemplateID = Number(selected[0]);
    show = false;
    showEditPolling = true;
  };

  const addFromFile = async () => {
    selectedTemplateID = 0;
    const pt = await ImportPollingTemplate();
    if (!pt.Type) {
      return;
    }
    const p = await GetDefaultPolling(nodeID);
    p.Level = pt.Level;
    p.Type = pt.Type;
    p.Name = pt.Name;
    p.Mode = pt.Mode;
    p.Params = pt.Params;
    p.Filter = pt.Filter;
    p.Script = pt.Script;
    p.Extractor = pt.Extractor;
    pollingTmp = p
    showEditPolling = true;
  }

  const onApplyAIAssist = async (e: CustomEvent) => {
    const aiPolling = e.detail.polling;
    if (!aiPolling) return;
    const p = await GetDefaultPolling(nodeID);
    p.Level = aiPolling.Level || p.Level;
    p.Type = aiPolling.Type || p.Type;
    p.Name = aiPolling.Name || p.Name;
    p.Mode = aiPolling.Mode || p.Mode;
    p.Params = aiPolling.Params || p.Params;
    p.Filter = aiPolling.Filter || p.Filter;
    p.Script = aiPolling.Script || p.Script;
    p.Extractor = aiPolling.Extractor || p.Extractor;
    if (aiPolling.PollInt) p.PollInt = aiPolling.PollInt;
    if (aiPolling.Timeout) p.Timeout = aiPolling.Timeout;
    if (aiPolling.Retry) p.Retry = aiPolling.Retry;
    pollingTmp = p;
    selectedTemplateID = 0;
    show = false;
    showEditPolling = true;
  };

  const columns = [
    {
      data: "ID",
      title: "ID",
      width: "5%",
      searchable: false,
    },
    {
      data: "Name",
      title: $_('AppPolling.Name'),
      width: "30%",
    },
    {
      data: "Type",
      title: $_('AddPolling.Type'),
      width: "15%",
      render: renderPollingType,
    },
    {
      data: "Mode",
      title: $_('AddPolling.Mode'),
      width: "15%",
      searchable: false,
    },
    {
      data: "Descr",
      title: $_('AddPolling.Descr'),
      width: "40%",
    },
  ];

  const onOpen =() => {
    showTable();
  };

  const close = () => {
    showEditPolling = false;
    show = false;
    dispatch("close", {});
  };

  $: if (show) {
    onOpen();
  }
</script>

<Modal bind:open={show} size="xl" dismissable={false} class="w-full">
  <div class="flex flex-col">
    <div class="m-5 grow">
      <table id="tmpTable" class="display compact" style="width:99%"></table>
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if hasAI}
        <GradientButton shadow color="pink" type="button" onclick={() => (showAIAssist = true)} size="xs">
          <Icon path={icons.mdiAutoFix} size={1} />
          {$_('AIPollingAssist.AIAssist')}
        </GradientButton>
      {/if}
      {#if selectedCount == 1}
        <GradientButton shadow color="blue" type="button" onclick={add} size="xs">
          <Icon path={icons.mdiPlus} size={1} />
          { $_('AddPolling.Add') }
        </GradientButton>
      {:else}
        <GradientButton shadow color="blue" type="button" onclick={addFromFile} size="xs">
          <Icon path={icons.mdiFile} size={1} />
          {$_('AddPolling.AddFromTmpFile')}
        </GradientButton>
      {/if}
      <GradientButton shadow type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('AddPolling.Cancel') }
      </GradientButton>
    </div>
  </div>
</Modal>

<AIPollingAssistDialog
  bind:show={showAIAssist}
  {nodeID}
  on:apply={onApplyAIAssist}
/>

<Polling
  bind:show={showEditPolling}
  {nodeID}
  pollingID={""}
  pollingTmpID={selectedTemplateID}
  pollingTmp={pollingTmp}
  on:close={close}
/>
