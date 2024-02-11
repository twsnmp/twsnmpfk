<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetAIList,
    DeleteAIResult,
  } from "../../wailsjs/go/main/App";
  import AIReport from "./AIReport.svelte";
  import { renderTime, getScoreIcon, getScoreColor,getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';

  let table :any = undefined;
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
      ` text-sm" style="color:` +
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
        await DeleteAIResult(id);
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
      title: $_('AIList.AnomaryScore'),
      width: "15%",
      render: formatScore,
    },
    {
      data: "Node",
      title: $_('AIList.Node'),
      width: "20%",
    },
    {
      data: "Polling",
      title: $_('AIList.Polling'),
      width: "15%",
    },
    {
      data: "Count",
      title: $_('AIList.Count'),
      width: "10%",
    },
    {
      data: "LastTime",
      title: $_('AIList.LastTime'),
      width: "15%",
      render: (data:any, type:any) =>
        renderTime(data * 1000 * 1000 * 1000,type),
    },
  ];

  const refresh = async () => {
    data = await GetAIList();
    if (table && DataTable.isDataTable("#table")) {
      table.clear();
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

</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
  {#if selectedCount == 1}
    <GradientButton shadow color="green" type="button" on:click={show} size="xs">
      <Icon path={icons.mdiChartBarStacked} size={1} />
      {$_('AIList.Report')}
    </GradientButton>
  {/if}
  {#if selectedCount > 0}
    <GradientButton shadow color="red" type="button" on:click={clearAIResult} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      { $_('AIList.Clear') }
    </GradientButton>
  {/if}
    <GradientButton shadow type="button" color="teal" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('AIList.Reload') }
    </GradientButton>
  </div>
</div>

<AIReport bind:show={showReport} id={selectedID} />
