<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { GetEventLogs } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getTableLang,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  let table = undefined;
  let data = [];
  let timer: number | undefined = undefined;

  const showTable = () => {
    if (table) {
      table.destroy();
      table = undefined;
    }
    table = new DataTable("#table", {
      columns: columns,
      paging: false,
      searching:false,
      info:false,
      scrollY: "180px",
      data: data,
      language: getTableLang(),
      order: [[1,"desc"]],
    });
  }

  const columns = [
    {
      data: "Level",
      title: "レベル",
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: "発生日時",
      width: "15%",
      render: renderTime,
    },
    {
      data: "Type",
      title: "種別",
      width: "10%",
    },
    {
      data: "NodeName",
      title: "関連ノード",
      width: "15%",
    },
    {
      data: "Event",
      title: "イベント",
      width: "50%",
    },
  ];

  const updateLogs = async () => {
    data = await GetEventLogs("");
    showTable();
    timer = setTimeout(() => {
      updateLogs();
    }, 60 * 1000);
  };
  onMount(() => {
    updateLogs();
  });
  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    if(table) {
      table.destroy();
    }
  });
</script>

<table id="table" class="display compact" style="width:98%;" />

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
