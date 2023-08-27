<script lang="ts">
  import { Button,Select } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import { onMount } from "svelte";
  import jaJP from "./gridjsJaJP";
  import { GetAIList, DeeleteAIResult,GetAIResult } from "../../wailsjs/go/main/App";
  import {
    formatTimeFromNano,
    getScoreIcon,
    getScoreColor,
  } from "./common";

  let data = [];

  const refresh = async () => {
    data = await GetAIList();
  };

  const formatScore = (score:number) => {
    return html(
      `<span class="mdi ` +
        getScoreIcon(score) +
        ` text-xl" style="color:` +
        getScoreColor(score) +
        `;" /><span class="ml-2 text-xs text-black dark:text-white">` +
        score +
        `</span>`
    );
  };

  const deleteAI = async (id: string) => {
    await DeeleteAIResult(id);
    refresh();
  };

  const show = async (id: string) => {
    console.log(id);
  };

  const columns = [
    {
      id: "Score",
      name: "異常スコア",
      width: "15%",
      formatter: formatScore,
    },
    {
      id: "Node",
      name: "ノード名",
      width: "20%",
    },
    {
      id: "Polling",
      name: "ポーリング",
      width: "15%",
    },
    {
      id: "Count",
      name: "データ数",
      width: "10%",
    },
    {
      id: "LastTime",
      name: "最終確認",
      width: "15%",
      formatter: formatTimeFromNano,
    },
    {
      id: "ID",
      name: "確認",
      sort: false,
      width: "5%",
      formatter: (id) => {
        return h(
          "button",
          {
            className: "",
            onClick: () => {
              show(id);
            },
          },
          html(`<span class="mdi mdi-eye text-lg" />`)
        );
      },
    },
    {
      name: "削除",
      width: "5%",
      formatter: (_, row) => {
        const id = row._cells[row._cells.length - 2].data;
        return h(
          "button",
          {
            className: "",
            onClick: () => {
              deleteAI(id);
            },
          },
          html(`<span class="mdi mdi-delete text-red-600 text-lg" />`)
        );
      },
    },
  ];

  onMount(() => {
    refresh();
  });

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

<div class="flex flex-col">
  <div class="m-5 twsnmpfk grow">
    <Grid {data} {columns} {pagination} sort search language={jaJP} />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
      <Select class="w-20" items={ppList} bind:value={pp} on:change={()=>{
        pagination = {
          limit:pp,
        }
      }}/>
    <Button type="button" color="alternative" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>


<style>
  @import "../assets/css/gridjs.css";
</style>
