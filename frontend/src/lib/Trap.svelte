<script lang="ts">
  import { Button,Select } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import Grid from "gridjs-svelte";
  import { onMount,tick } from "svelte";
  import jaJP from "./gridjsJaJP";
  import { GetTraps, ExportTraps } from "../../wailsjs/go/main/App";
  import {
    formatTimeFromNano,
  } from "./common";
  import {showLogCountChart,resizeLogCountChart} from "./chart/logcount";
  import TrapReport from "./TrapReport.svelte";

  let data = [];
  let logs = [];
  let showReport = false;

  const refresh = async () => {
    logs = await GetTraps(0);
    data = [];
    for (let i =0; i < logs.length;i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showChart();
  };

  const showChart = async () => {
    tick();
    showLogCountChart("chart",data,zoomCallBack);
  }

  const zoomCallBack = (st:number, et:number) => {
    data = [];
    for(let i = logs.length -1 ; i >= 0;i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
  };

  const columns = [
    {
      id: "Time",
      name: "日時",
      width: "20%",
      formatter: formatTimeFromNano,
    },
    {
      id: "FromAddress",
      name: "送信元",
      width: "15%",
    },
    {
      id: "TrapType",
      name: "タイプ",
      width: "15%",
    },
    {
      id: "Variables",
      name: "変数",
      width: "50%",
    },
  ];

  onMount(() => {
    refresh();
  });


  const saveCSV = () => {
    ExportTraps("csv");
  }

  const saveExcel = () => {
    ExportTraps("excel");
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

<svelte:window on:resize={resizeLogCountChart} />

<div class="flex flex-col">
  <div id="chart" style="height: 200px;"></div>
  <div class="m-5 twsnmpfk grow">
    <Grid {data} {columns} {pagination} sort search language={jaJP} />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
      <Select class="w-20" items={ppList} bind:value={pp} on:change={()=>{
        pagination = {
          limit:pp,
        }
      }}/>
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
  <TrapReport
   {logs}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/gridjs.css";
</style>
