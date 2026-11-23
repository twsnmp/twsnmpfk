<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton, Button, ButtonGroup } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy, tick } from "svelte";
  import {
    GetAIList,
    DeleteAIResult,
    ExportAIData,
  } from "../../wailsjs/go/main/App";
  import AIReport from "./AIReport.svelte";
  import Correlation from "./Correlation.svelte";
  import { renderTime, getScoreIcon, getScoreColor,getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';
  import { currentLang } from '../i18n/i18n';

  let activeTab: "ai" | "correlation" = "ai";
  let table :any = undefined;
  let data = [];
  let selectedCount = 0;
  let selectedID = "";
  let showReport = false;


  const formatScore = (score: number,type:string)=> {
    if (type == "sort") {
      return score;
    }
    return (
      `<span class="mdi ` +
      getScoreIcon(score) +
      ` text-xs" style="color:` +
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
      table.rows({ selected: true }).remove().draw();
    }
  };

  const exportAIData = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected && selected.length == 1) {
      const id = selected[0];
      ExportAIData(id);
    }
  };

  const show = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length == 1) {
      selectedID = selected[0];
      showReport = true;
    }
  };

  const getColumns = () => [
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

  const showTable = () => {
    const el = document.getElementById("aiListTable");
    if (!el) {
      return;
    }
    if (table && DataTable.isDataTable("#aiListTable")) {
      table.off("select");
      table.off("deselect");
      table.destroy();
      table = undefined;
    }
    el.innerHTML = "";
    selectedCount = 0;
    table = new DataTable("#aiListTable", {
      destroy: true,
      columns: getColumns(),
      pageLength: window.innerHeight > 800 ? 25 : 10,
      stateSave: true,
      data: data,
      order:[[0,"desc"]],
      language: getTableLang(),
      select: {
        style: "multi+shift",
      },
    });
    table.on("select", () => {
      if (!table) {
        return;
      }
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      if (!table) {
        return;
      }
      selectedCount = table.rows({ selected: true }).count();
    });
    selectedCount = table.rows({ selected: true }).count();
  };

  const refresh = async () => {
    data = await GetAIList();
    showTable();
  };

  const destroyTable = () => {
    if (table && DataTable.isDataTable("#aiListTable")) {
      table.off("select");
      table.off("deselect");
      table.destroy();
      table = undefined;
    }
  };

  const switchTab = async (next: "ai" | "correlation") => {
    if (activeTab === next) return;
    if (activeTab === "ai") {
      destroyTable();
    }
    activeTab = next;
    if (next === "ai") {
      await tick();
      showTable();
    }
  };

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    destroyTable();
  });

  let prevLang = "";
  $: {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && table && activeTab === "ai") {
      showTable();
    }
    prevLang = lang;
  }

</script>

<div class="flex flex-col max-h-[calc(100vh-50px)] overflow-y-auto pb-2">
  <div class="flex flex-wrap items-center gap-1 mx-2 mt-2 pb-2 border-b border-gray-200 dark:border-gray-700">
    <ButtonGroup>
      <Button
        size="xs"
        class="!py-1.5 !px-3 text-xs flex items-center gap-1.5"
        color={activeTab === "ai" ? "blue" : "alternative"}
        onclick={() => switchTab("ai")}
      >
        <Icon path={icons.mdiChartBellCurveCumulative} size={0.85} />
        <span>{$_("AIList.AnomalyTab") || "異常検知"}</span>
      </Button>
      <Button
        size="xs"
        class="!py-1.5 !px-3 text-xs flex items-center gap-1.5"
        color={activeTab === "correlation" ? "blue" : "alternative"}
        onclick={() => switchTab("correlation")}
      >
        <Icon path={icons.mdiChartScatterPlot} size={0.85} />
        <span>{$_("AIList.CorrelationTab") || "相関分析"}</span>
      </Button>
    </ButtonGroup>
  </div>

  {#if activeTab === "ai"}
    <div class="mx-2 mt-2">
      <div>
        <table id="aiListTable" class="display compact" style="width:99%"></table>
      </div>
    </div>
    <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount == 1}
      <GradientButton shadow color="green" type="button" onclick={show} size="xs">
        <Icon path={icons.mdiChartBarStacked} size={1} />
        {$_('AIList.Report')}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        onclick={exportAIData}
        size="xs"
      >
        <Icon path={icons.mdiFileDelimited} size={1} />
        CSV(Data)
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton shadow color="red" type="button" onclick={clearAIResult} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        { $_('AIList.Clear') }
      </GradientButton>
    {/if}
      <GradientButton shadow type="button" color="teal" onclick={refresh} size="xs">
        <Icon path={icons.mdiRecycle} size={1} />
        { $_('AIList.Reload') }
      </GradientButton>
    </div>
  {:else}
    <div class="mt-1">
      <Correlation />
    </div>
  {/if}
</div>

<AIReport bind:show={showReport} id={selectedID} />
