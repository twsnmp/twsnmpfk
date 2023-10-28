<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Select,
  } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {
    GetNode,
    GetPolling,
    GetPollingLogs,
    GetAIResult,
  } from "../../wailsjs/go/main/App";
  import {showLogStateChart} from "./chart/logstate";
  import {showPollingChart,showPollingHistogram,getChartParams} from "./chart/polling";
  import {
    getStateIcon,
    getStateColor,
    getStateName,
    getTableLang,
    renderTime,
    renderState,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { showAIHeatMap } from "./chart/ai";
  import { _ } from "svelte-i18n";

  export let id = "";

  let polling: datastore.PollingEnt | undefined = undefined;
  let node: datastore.NodeEnt | undefined = undefined;
  let logs: datastore.PollingLogEnt[] | undefined = undefined;
  let dispLogs: datastore.PollingLogEnt[] = [];
  let aiResult: datastore.AIResult | undefined = undefined;
  let entList = [];
  let selectedEnt = "";
  let pollingLogTable = undefined;
 
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  onMount(async () => {
    polling = await GetPolling(id);
    node = await GetNode(polling.NodeID);
    show = true;
    if (polling.LogMode > 0) {
      loadLogs();
    }
    if(polling && polling.Result) {
      for(const k of Object.keys(polling.Result)) {
        selectedEnt = k;
        const dp = getChartParams(k);
        entList.push({
          name:dp.axis,
          value:k,
        });
      }
    }
  });

  onDestroy(() => {
    if(pollingLogTable) {
      pollingLogTable.destroy();
    }
  });

  const loadLogs = async () => {
    logs = await GetPollingLogs(id);
    for(let i =0; i < logs.length;i++) {
      dispLogs.push(logs[i]);
    }
    logs.reverse();
    aiResult = await GetAIResult(id);
  };

  const zoomCallBack = (st:number, et:number) => {
    dispLogs = [];
    for(let i = logs.length -1 ; i >= 0;i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        dispLogs.push(logs[i]);
      }
    }
    showLogTable();
  };

  const showLog = async () => {
    await tick();
    showLogTable();
    showLogStateChart("log",logs,zoomCallBack);
  };

  const showLogTable = () => {
    if (pollingLogTable) {
      pollingLogTable.destroy();
    }
    pollingLogTable = new DataTable("#pollingLogTable", {
      data: dispLogs,
      order: [[0, "desc"]],
      language: getTableLang(),
      columns: [
        {
          data: "State",
          title: $_('PollingReport.State'),
          width: "10%",
          render: renderState,
        },
        {
          data: "Time",
          title: $_('PollingReport.Time'),
          width: "15%",
          render: renderTime,
        },
        {
          data: "Result",
          title: $_('PollingReport.Result'),
          width: "75%",
          render: renderResult,
        },
      ],
    });
  };

  const renderResult = (r) => {
    let l = [];
    for(const k of Object.keys(r)) {
      l.push(k +"=" + r[k]);
    }
    return l.join(" ");
  }
  let chart = undefined;

  const showTimeChart = async () => {
    await tick();
    chart = showPollingChart("time",logs,selectedEnt)
  };

  const showHistogram = async () => {
    await tick();
    chart = showPollingHistogram("histogram",logs,selectedEnt);
  };

  const showAI = async () => {
    await tick();
    chart = showAIHeatMap("ai", aiResult.ScoreData);
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  } 

</script>

<svelte:window on:resize={resizeChart} />

<Modal bind:open={show} size="xl" permanent class="w-full min-h-[90vh]" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={()=> {chart= undefined;}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          { $_('PollingReport.BasicInfo') }
        </div>
        <Table striped={true}>
          <TableHead>
            <TableHeadCell>{ $_('PollingReport.Item') }</TableHeadCell>
            <TableHeadCell>{ $_('PollingReport.Content') }</TableHeadCell>
          </TableHead>
          <TableBody tableBodyClass="divide-y">
            <TableBodyRow>
              <TableBodyCell>{ $_('PollingReport.NodeName') }</TableBodyCell>
              <TableBodyCell>{node.Name}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>{ $_('PollingReport.Name') }</TableBodyCell>
              <TableBodyCell>{polling.Name}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>{ $_('PollingReport.State') }</TableBodyCell>
              <TableBodyCell>
                <span
                  class="mdi {getStateIcon(polling.State)} text-xl"
                  style="color:{getStateColor(polling.State)};"
                />
                <span class="ml-2 text-xs text-black dark:text-white"
                  >{getStateName(polling.State)}</span
                >
              </TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>{ $_('PollingReport.LastTime') }</TableBodyCell>
              <TableBodyCell>{renderTime(polling.LastTime, "")}</TableBodyCell>
            </TableBodyRow>
            {#each Object.keys(polling.Result) as k}
              <TableBodyRow>
                <TableBodyCell>{k}</TableBodyCell>
                <TableBodyCell>{polling.Result[k]}</TableBodyCell>
              </TableBodyRow>
            {/each}
          </TableBody>
        </Table>
      </TabItem>
      {#if polling.LogMode > 0}
        <TabItem on:click={showLog}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiLanCheck} size={1} />
            { $_('PollingReport.PollingLog') }
          </div>
          <div id="log"/>
          <table id="pollingLogTable" class="display compact" style="width:99%;" />
        </TabItem>
        <TabItem on:click={showTimeChart}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiCalendarCheck} size={1} />
            { $_('PollingReport.TimeChart') }
          </div>
          <Select class="mb-2" size="sm" items={entList} bind:value={selectedEnt} on:change={showTimeChart} placeholder={ $_('PollingReport.SelectVal') }/>
          <div id="time" />
        </TabItem>
        <TabItem on:click={showHistogram}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiAppsBox} size={1} />
            { $_('PollingReport.Histogram') }
          </div>
          <Select class="mb-2" size="sm" items={entList} bind:value={selectedEnt} on:change={showHistogram} placeholder={ $_('PollingReport.SelectVal') }/>
          <div id="histogram"/>
        </TabItem>
        {#if polling.LogMode == 3 && aiResult}
          <TabItem on:click={showAI}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiAppsBox} size={1} />
              { $_('PollingReport.AI') }
            </div>
            <div id="ai"/>
          </TabItem>
        {/if}
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('PollingReport.Close') }
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #log {
    min-height: 200px;
    height:  30vh;
    width:  98%;
    margin: 0 auto;
  }
  #time,
  #histogram,
  #ai {
    min-height: 500px;
    height: 75vh;
    widows: 98%;
    margin:  0 auto;
  }
</style>