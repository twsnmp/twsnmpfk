<script lang="ts">
  import { Button,Select } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import Grid from "gridjs-svelte";
  import {  html } from "gridjs";
  import { onMount,tick } from "svelte";
  import jaJP from "./gridjsJaJP";
  import { GetEventLogs, ExportEventLogs } from "../../wailsjs/go/main/App";
  import {
    cmpState,
    getStateIcon,
    getStateColor,
    getStateName,
    formatTimeFromNano,
  } from "./common";
  import {showLogLevelChart,resizeLogLevelChart} from "./chart/loglevel";

  let data = [];
  let logs = [];

  const refresh = async () => {
    logs = await GetEventLogs(0);
    data = [];
    for (let i =0; i < logs.length;i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showChart();
  };

  const showChart = async () => {
    tick();
    showLogLevelChart("chart",logs,zoomCallBack);
  };
 
  const zoomCallBack = (st:number, et:number) => {
    data = [];
    for(let i = logs.length -1 ; i >= 0;i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
  };

  const formatState = (state) => {
    return html(
      `<span class="mdi ` +
        getStateIcon(state) +
        ` text-xl" style="color:` +
        getStateColor(state) +
        `;" /><span class="ml-2 text-xs text-black dark:text-white">` +
        getStateName(state) +
        `</span>`
    );
  };

  const columns = [
    {
      id: "Level",
      name: "レベル",
      width: "10%",
      formatter: formatState,
      sort: {
        compare: cmpState,
      },
    },
    {
      id: "Time",
      name: "発生日時",
      width: "15%",
      formatter: formatTimeFromNano,
    },
    {
      id: "Type",
      name: "種別",
      width: "10%",
    },
    {
      id: "NodeName",
      name: "関連ノード",
      width: "15%",
    },
    {
      id: "Event",
      name: "イベント",
      width: "50%",
    },
  ];

  onMount(() => {
    refresh();
  });


  const saveCSV = () => {
    ExportEventLogs("csv");
  }

  const saveExcel = () => {
    ExportEventLogs("excel");
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
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>

<style>
  @import "../assets/css/gridjs.css";
</style>
