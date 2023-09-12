<script lang="ts">
  import {
    Modal,
    Button,
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
          title: "状態",
          width: "10%",
          render: renderState,
        },
        {
          data: "Time",
          title: "日時",
          width: "15%",
          render: renderTime,
        },
        {
          data: "Result",
          title: "結果",
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

  const showTimeChart = async () => {
    await tick();
    showPollingChart("time",logs,selectedEnt)
  };

  const showHistogram = async () => {
    await tick();
    showPollingHistogram("histogram",logs,selectedEnt);
  };

  const showAI = async () => {
    await tick();
    showAIHeatMap("ai", aiResult.ScoreData);
  };

</script>

<Modal bind:open={show} size="xl" permanent class="w-full min-h-[90vh]" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          基本情報
        </div>
        <Table striped={true}>
          <TableHead>
            <TableHeadCell>項目</TableHeadCell>
            <TableHeadCell>内容</TableHeadCell>
          </TableHead>
          <TableBody tableBodyClass="divide-y">
            <TableBodyRow>
              <TableBodyCell>ノード名</TableBodyCell>
              <TableBodyCell>{node.Name}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>ポーリング名</TableBodyCell>
              <TableBodyCell>{polling.Name}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>状態</TableBodyCell>
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
              <TableBodyCell>最終実施</TableBodyCell>
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
            ポーリングログ
          </div>
          <div id="log" style="height: 200px; margin-bottom:10px" />
          <table id="pollingLogTable" class="display compact" style="width:99%" />
        </TabItem>
        <TabItem on:click={showTimeChart}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiCalendarCheck} size={1} />
            時系列
          </div>
          <Select class="mb-2" size="sm" items={entList} bind:value={selectedEnt} on:change={showTimeChart} placeholder="変数を選択"/>
          <div id="time" style="height: 500px;" />
        </TabItem>
        <TabItem on:click={showHistogram}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiAppsBox} size={1} />
            ヒストグラム
          </div>
          <Select class="mb-2" size="sm" items={entList} bind:value={selectedEnt} on:change={showHistogram} placeholder="変数を選択"/>
          <div id="histogram" style="height: 500px;" />
        </TabItem>
        {#if polling.LogMode == 3 && aiResult}
          <TabItem on:click={showAI}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiAppsBox} size={1} />
              AI分析
            </div>
            <div id="ai" style="height: 500px;" />
          </TabItem>
        {/if}
      {/if}
    </Tabs>
    <div class="flex justify-end space-x-2 mr-2">
      <Button type="button" color="alternative" on:click={close} size="sm" class="w-18">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>
