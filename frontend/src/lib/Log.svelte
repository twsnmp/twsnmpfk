<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { onMount, onDestroy, tick } from "svelte";
  import { GetMapEventLogs } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getTableLang,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';
  import { currentLang } from '../i18n/i18n';
  let table: any = undefined;
  let data: any = [];
  let timer: any = undefined;
  let isMounted = false;
  let isUpdating = false;
  let lastLogSignature = "";

  const buildLogSignature = (rows: any[]): string => {
    if (!rows || rows.length === 0) {
      return "0";
    }
    return rows
      .map((r: any) => `${r.Time}|${r.Level}|${r.Type}|${r.NodeName}|${r.Event}`)
      .join("\n");
  };

  const showTable = (forceRebuild = false) => {
    const el = document.getElementById("logTable");
    if (!el) {
      return;
    }
    if (!forceRebuild && table && DataTable.isDataTable("#logTable")) {
      table.clear();
      table.rows.add(data || []);
      table.draw(false);
      return;
    }
    if (table && DataTable.isDataTable("#logTable")) {
      table.destroy();
      table = undefined;
    }
    el.innerHTML = "";
    table = new DataTable("#logTable", {
      destroy: true,
      stateSave: true,
      columns: getColumns(),
      paging: false,
      searching: false,
      info: false,
      scrollY: window.innerHeight / 5 + "px",
      data: data,
      language: getTableLang(),
      order: [[1, "desc"]],
    });
  };

  const getColumns = () => [
    {
      data: "Level",
      title: $_('Log.Level'),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_('Log.Time'),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Type",
      title: $_('Log.Type'),
      width: "10%",
    },
    {
      data: "NodeName",
      title: $_('Log.NodeName'),
      width: "15%",
    },
    {
      data: "Event",
      title: $_('Log.Event'),
      width: "50%",
    },
  ];

  const scheduleNextUpdate = () => {
    if (!isMounted) {
      return;
    }
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      updateLogs();
    }, 60 * 1000);
  };

  const updateLogs = async () => {
    if (!isMounted || isUpdating) {
      return;
    }
    if (typeof document !== "undefined" && document.hidden) {
      scheduleNextUpdate();
      return;
    }
    isUpdating = true;
    try {
      const nextData = (await GetMapEventLogs()) || [];
      if (!isMounted) {
        return;
      }
      const nextSig = buildLogSignature(nextData);
      if (nextSig !== lastLogSignature || !table) {
        lastLogSignature = nextSig;
        data = nextData;
        await tick();
        if (!isMounted) {
          return;
        }
        showTable(false);
      }
    } finally {
      isUpdating = false;
      scheduleNextUpdate();
    }
  };

  const handleVisibilityChange = () => {
    if (typeof document !== "undefined" && !document.hidden) {
      updateLogs();
    }
  };

  onMount(() => {
    isMounted = true;
    document.addEventListener("visibilitychange", handleVisibilityChange);
    updateLogs();
  });

  let prevLang = "";
  $: {
    const lang = $currentLang;
    if (prevLang && prevLang !== lang && table) {
      showTable(true);
    }
    prevLang = lang;
  }
  onDestroy(() => {
    isMounted = false;
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    if (typeof document !== "undefined") {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    }
    if (table && DataTable.isDataTable("#logTable")) {
      try {
        table.destroy();
      } catch (e) {
        // ignore
      }
      table = undefined;
    }
  });
</script>

<div>
  <table id="logTable" class="display compact" style="width:98%;"></table>
</div>

