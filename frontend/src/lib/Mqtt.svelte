<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetMqttStatList,
    DeleteMqttStats,
    DeleteAllMqttStat,
  } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getTableLang,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let data: any = [];
  let table: any = undefined;
  let selectedCount = 0;

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#mqttStatTable", {
      destroy: true,
      columns: columns,
      data: data,
      stateSave: true,
      order: [
        [0, "asc"],
        [1, "asc"],
      ],
      pageLength: window.innerHeight > 800 ? 25 : 10,
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

  const refresh = async () => {
    data = [];
    data = await GetMqttStatList();
    showTable();
  };

  const deleteMqttStats = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length < 1) {
      return;
    }
    await DeleteMqttStats(selected.toArray());
    table.rows({ selected: true }).remove().draw();
    selectedCount = 0;
  };

  const deleteAll = async () => {
    await DeleteAllMqttStat();
    refresh();
  }

  const columns = [
    {
      data: "State",
      title: $_("PollingList.State"),
      width: "10%",
      render: renderState,
    },
    {
      data: "ClientID",
      title: $_('Mqtt.ClientID'),
      width: "15%",
    },
    {
      data: "Remote",
      title: $_('Mqtt.Remote'),
      width: "15%",
    },
    {
      data: "Topic",
      title: $_('Mqtt.Topic'),
      width: "20%",
    },
    {
      data: "Count",
      title: $_('Mqtt.Count'),
      width: "8%",
      searchable: false,
     },
    {
      data: "Bytes",
      title: $_('Mqtt.Bytes'),
      width: "8%",
      searchable: false,
     },
    {
      data: "First",
      title: $_('Mqtt.First'),
      width: "12%",
      render: renderTime,
      searchable: false,
    },
    {
      data: "Last",
      title: $_("PollingList.LastTime"),
      width: "12%",
      render: renderTime,
      searchable: false,
    },
  ];
  onMount(() => {
    refresh();
  });
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="mqttStatTable" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={deleteMqttStats}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("PollingList.Delete")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_('Mqtt.DeleteAll')}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("PollingList.Reload")}
    </GradientButton>
  </div>
</div>
