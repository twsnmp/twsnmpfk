<script lang="ts">
  import { Button, Select } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import { onMount } from "svelte";
  import jaJP from "./gridjsJaJP";
  import {
    GetNodes,
    GetPollings,
    DeletePollings,
    ExportPollings,
  } from "../../wailsjs/go/main/App";
  import {
    cmpState,
    getStateIcon,
    getStateColor,
    getStateName,
    formatTimeFromNano,
    getLogModeName,
  } from "./common";
  import Polling from "./Polling.svelte";

  let data = [];
  let nodes = {};
  let showEditPolling = false;
  let selectedPolling = "";

  const refreshPollings = async () => {
    data = [];
    nodes = await GetNodes();
    data = await GetPollings("");
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

  const editPolling = (id: string) => {
    if (!id) {
      return;
    }
    selectedPolling = id;
    showEditPolling = true;
  };

  const deletePolling = async (id: string) => {
    await DeletePollings([id]);
    refreshPollings();
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
      id: "NodeID",
      name: "ノード名",
      width: "15%",
      formatter: (id) => nodes[id].Name,
    },
    {
      id: "Name",
      name: "名前",
      width: "25%",
    },
    {
      id: "Level",
      name: "レベル",
      width: "10%",
      formatter: formatState,
    },
    {
      id: "Type",
      name: "種別",
      width: "8%",
    },
    {
      id: "LogMode",
      name: "ログ",
      width: "7%",
      formatter: getLogModeName,
    },
    {
      id: "LastTime",
      name: "最終確認",
      width: "15%",
      formatter: formatTimeFromNano,
    },
    {
      id: "ID",
      name: "編集",
      sort: false,
      width: "5%",
      formatter: (id: string) => {
        return h(
          "button",
          {
            className: "",
            onClick: () => {
              editPolling(id);
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
              deletePolling(id);
            },
          },
          html(`<span class="mdi mdi-delete text-red-600 text-lg" />`)
        );
      },
    },
  ];

  onMount(() => {
    refreshPollings();
  });

  const saveCSV = () => {
    ExportPollings("csv");
  };

  const saveExcel = () => {
    ExportPollings("excel");
  };

  let pagination : any= {
    limit: 10,
  };
  let pp = 10;
  const ppList = [
    { name: "10", value: 10 },
    { name: "20", value: 20 },
    { name: "100", value: 100 },
  ];
</script>

<div class="flex flex-col">
  <div class="m-5 twsnmpfk grow">
    <Grid {data} {columns} {pagination} sort search language={jaJP} />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <Select
      class="w-20"
      items={ppList}
      bind:value={pp}
      on:change={() => {
        pagination = {
          limit: pp,
        };
      }}
    />
    <Button color="blue" type="button" on:click={saveCSV} size="xs">
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </Button>
    <Button color="blue" type="button" on:click={saveExcel} size="xs">
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </Button>
    <Button
      type="button"
      color="alternative"
      on:click={refreshPollings}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      更新
    </Button>
  </div>
</div>

{#if showEditPolling}
  <Polling
    nodeID=""
    pollingID={selectedPolling}
    on:close={(e) => {
      showEditPolling = false;
      refreshPollings();
    }}
  />
{/if}

<style>
  @import "../assets/css/gridjs.css";
</style>
