<script lang="ts">
  import { Button } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount,tick,onDestroy } from "svelte";
  import { GetSyslogs, ExportSyslogs } from "../../wailsjs/go/main/App";
  import {
    getStateIcon,
    getStateColor,
    getStateName,
    renderTime,
    getTableLang,
    levelNum,
  } from "./common";
  import {showLogLevelChart,resizeLogLevelChart} from "./chart/loglevel";
  import SyslogReport  from "./SyslogReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  let data = [];
  let logs = [];
  let showReport = false;
  let table = undefined;
  let selectedCount = 0;

  const showTable = () => {
    if (table) {
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#table", {
      columns: columns,
      data: data,
      order: [1,"desc"],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  }

  const refresh = async () => {
    logs = await GetSyslogs();
    data = [];
    for (let i =0; i < logs.length;i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
  };

  const showChart = async () => {
    await tick();
    showLogLevelChart("chart",logs,zoomCallBack);
  }

  const zoomCallBack = (st:number, et:number) => {
    data = [];
    for(let i = logs.length -1 ; i >= 0;i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
    showTable();
  };

  const formatState = (state:string,type:string) => {
    if(type =="sort") {
      return levelNum(state);
    }
    return `<span class="mdi ` +
        getStateIcon(state) +
        ` text-xl" style="color:` +
        getStateColor(state) +
        `;" ></span><span class="ml-2">` +
        getStateName(state) +
        `</span>`;
  };

  const columns = [
    {
      data: "Level",
      title: "レベル",
      width: "10%",
      render: formatState,
    },
    {
      data: "Time",
      title: "日時",
      width: "15%",
      render: renderTime,
    },
    {
      data: "Host",
      title: "ホスト",
      width: "15%",
    },
    {
      data: "Type",
      title: "タイプ",
      width: "10%",
    },
    {
      data: "Tag",
      title: "タグ",
      width: "10%",
    },
    {
      data: "Message",
      title: "メッセージ",
      width: "40%",
    },
  ];

  onMount(() => {
    refresh();
  });

  onDestroy(()=>{
    if(table) {
      table.destroy();
      table = undefined;
    }
  });


  const saveCSV = () => {
    ExportSyslogs("csv");
  }

  const saveExcel = () => {
    ExportSyslogs("excel");
  }

  let pagination: any = {
    limit: 10,
  };
  let pp = 10;
  const ppList = [
    { name:"10",value:10 },
    { name:"20",value:20 },
    { name:"100",value:100 },
  ]

</script>

<svelte:window on:resize={resizeLogLevelChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;"></div>
  <div class="m-5 grow">
    <table id="table" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <Button color="blue" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </Button>
    <Button color="blue" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </Button>
    <Button type="button" color="green" on:click={() => {showReport=true}} size="xs">
      <Icon path={icons.mdiChartPie} size={1} />
      レポート
    </Button>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>

{#if showReport}
  <SyslogReport
   {logs}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
