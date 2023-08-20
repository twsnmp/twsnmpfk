<script lang="ts">
  import { Button,Select } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import { onMount } from "svelte";
  import jaJP from "./gridjsJaJP";
  import { GetNodes, DeleteNodes, ExportNodes,CheckPolling } from "../../wailsjs/go/main/App";
  import {
    cmpIP,
    cmpState,
    getIcon,
    getStateColor,
    getStateName,
  } from "./common";
  import Node from "./Node.svelte";

  let data = [];
  let showEditNode = false;
  let selectedNode = "";

  const refresh = async () => {
    const nodes = await GetNodes();
    data = [];
    for (const k in nodes) {
      data.push(nodes[k]);
    }
  };

  const formatState = (state, row) => {
    return html(
      `<span class="mdi ` +
        getIcon(row._cells[1].data) +
        ` text-xl" style="color:` +
        getStateColor(state) +
        `;" /><span class="ml-2 text-xs text-black dark:text-white">` +
        getStateName(state) +
        `</span>`
    );
  };

  const editNode = (id: string) => {
    if (!id) {
      return;
    }
    selectedNode = id;
    showEditNode = true;
  };

  const deleteNode = async (id: string) => {
    await DeleteNodes([id]);
    refresh();
  };

  const columns = [
    {
      id: "State",
      name: "状態",
      width: "10%",
      formatter: formatState,
      sort: {
        compare: cmpState,
      },
    },
    {
      id: "Icon",
      name: "",
      hidden: true,
    },
    {
      id: "Name",
      name: "名前",
      width: "20%",
    },
    {
      id: "IP",
      name: "IPアドレス",
      width: "15%",
      sort: {
        compare: cmpIP,
      },
    },
    {
      id: "MAC",
      name: "MACアドレス",
      width: "15%",
    },
    {
      id: "Descr",
      name: "説明",
      width: "30%",
    },
    {
      id: "ID",
      name: "編集",
      sort: false,
      width: "5%",
      formatter: (id) => {
        return h(
          "button",
          {
            className: "",
            onClick: () => {
              editNode(id);
            },
          },
          html(`<span class="mdi mdi-pencil text-lg" />`)
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
              deleteNode(id);
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

  const checkAll = () => {
    CheckPolling("all");
  }

  const saveCSV = () => {
    ExportNodes("csv");
  }

  const saveExcel = () => {
    ExportNodes("excel");
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
    <Button color="blue" type="button" on:click={checkAll} size="xs">
      <Icon path={icons.mdiCheckAll} size={1} />
      すべて再確認
    </Button>
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

{#if showEditNode}
  <Node
    nodeID={selectedNode}
    on:close={(e) => {
      showEditNode = false;
      refresh();
    }}
  />
{/if}

<style>
  @import "../assets/css/gridjs.css";
</style>
