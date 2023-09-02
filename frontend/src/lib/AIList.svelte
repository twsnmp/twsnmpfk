<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, tick, onDestroy } from "svelte";
  import {
    GetAIList,
    DeeleteAIResult,
  } from "../../wailsjs/go/main/App";
  import AIReport from "./AIReport.svelte";
  import { renderTime, getScoreIcon, getScoreColor,getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  let table = undefined;
  let data = [];
  let selectedCount = 0;
  let selectedID = "";
  let showReport = false;


  const formatScore = (score: number,type:string): string => {
    if (type == "sort") {
      return score + "";
    }
    return (
      `<span class="mdi ` +
      getScoreIcon(score) +
      ` text-xl" style="color:` +
      getScoreColor(score) +
      `;"></span><span class="ml-2">` +
      score.toFixed(2) +
      `</span>`
    );
  };

  const clearAIResult = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected) {
      for(let i = 0; i < selected.length;i++) {
        const id = selected[i];
        await DeeleteAIResult(id);
      }
    }
    refresh();
  };

  const show = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length == 1) {
      selectedID = selected[0];
      showReport = true;
    }
  };

  const columns = [
    {
      data: "Score",
      title: "異常スコア",
      width: "15%",
      render: formatScore,
    },
    {
      data: "Node",
      title: "ノード名",
      width: "20%",
    },
    {
      data: "Polling",
      title: "ポーリング",
      width: "15%",
    },
    {
      data: "Count",
      title: "データ数",
      width: "10%",
    },
    {
      data: "LastTime",
      title: "最終確認",
      width: "15%",
      render: (data, type, row, meta) =>
        renderTime(data * 1000 * 1000 * 1000,type),
    },
  ];

  const refresh = async () => {
    data = await GetAIList();
    if (table) {
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#table", {
      columns: columns,
      data: data,
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  };

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    if (table) {
      table.destroy();
    }
  });
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
  {#if selectedCount == 1}
    <Button color="green" type="button" on:click={show} size="xs">
      <Icon path={icons.mdiChartBarStacked} size={1} />
      レポート
    </Button>
  {/if}
  {#if selectedCount > 0}
    <Button color="red" type="button" on:click={clearAIResult} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      クリア
    </Button>
  {/if}
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>

{#if showReport}
  <AIReport
   id={selectedID}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}


<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
