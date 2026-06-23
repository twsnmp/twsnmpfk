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
    GetDefaultPolling,
    GetNodes,
  } from "../../wailsjs/go/main/App";
  import {
    renderState,
    renderTime,
    getTableLang,
    renderCount,
    renderBytes,
  } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { copyText } from "svelte-copy";
  import Polling from "./Polling.svelte";

  let data: any = [];
  let table: any = undefined;
  let selectedCount = 0;
  let showPolling = false;
  let polling: any = undefined;
  let copied = false;

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

  const copyTopic = () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length < 1) {
      return;
    }
    const topics: string[] = [];
    for (let i = 0; i < selected.length; i++) {
      topics.push(selected[i].Topic);
    }
    copyText(topics.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };

  const makePolling = async () => {
    const selected = table.rows({ selected: true }).data();
    if (!selected || selected.length !== 1) {
      return;
    }
    const nodes = await GetNodes();
    const nodeList = Object.values(nodes);
    
    // (1) Remote IP address
    const remoteIP = selected[0].Remote;
    let node = nodeList.find(n => n.IP === remoteIP);
    
    // (2) MQTT broker (TWSNMP FK itself) node
    if (!node) {
      node = nodeList.find(n => n.IP === "127.0.0.1" || n.IP === "localhost" || n.IP === "::1");
      if (!node) {
        node = nodeList.find(n => n.Name.toLowerCase().includes("twsnmp"));
      }
    }
    
    // (3) First node in the node list
    if (!node && nodeList.length > 0) {
      node = nodeList[0];
    }
    
    const nodeID = node ? node.ID : "";
    polling = await GetDefaultPolling(nodeID);
    
    polling.Name = `mqtt ${selected[0].Topic}`;
    polling.Type = "mqtt";
    polling.Mode = "subscribe";
    polling.Params = "tcp://localhost:1883";
    polling.Filter = selected[0].Topic;
    polling.Extractor = "";
    polling.Script = "";
    showPolling = true;
  };

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
      render: renderCount,
      "className": "dt-right",
      searchable: false,
     },
    {
      data: "Bytes",
      title: $_('Mqtt.Bytes'),
      width: "8%",
      render: renderBytes,
      "className": "dt-right",
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
    <table id="mqttStatTable" class="display compact" style="width:99%"></table>
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount === 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={makePolling}
        size="xs"
      >
        <Icon path={icons.mdiEye} size={1} />
        {$_('Mqtt.CreatePolling')}
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
        shadow
        color="cyan"
        type="button"
        onclick={copyTopic}
        size="xs"
      >
        {#if copied}
          <Icon path={icons.mdiCheck} size={1} />
        {:else}
          <Icon path={icons.mdiContentCopy} size={1} />
        {/if}
        {$_('Mqtt.CopyTopic')}
      </GradientButton>
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={deleteMqttStats}
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
      onclick={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_('Mqtt.DeleteAll')}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("PollingList.Reload")}
    </GradientButton>
  </div>
</div>

<Polling bind:show={showPolling} pollingTmp={polling} />

