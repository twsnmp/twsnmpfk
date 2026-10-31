<script lang="ts">
  import { _ } from "svelte-i18n";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Button, ButtonGroup } from "flowbite-svelte";

  import SensorReport from "./SensorReport.svelte";
  import WifiReport from "./WifiReport.svelte";
  import BluetoothReport from "./BluetoothReport.svelte";
  import EnvSensor from "./EnvSensor.svelte";
  import PcapReport from "./PcapReport.svelte";
  import WinLogReport from "./WinLogReport.svelte";
  import AgentReport from "./AgentReport.svelte";

  type ReportCategory =
    | "sensor"
    | "wifi"
    | "bluetooth"
    | "env"
    | "pcap"
    | "winlog"
    | "twagent";

  let activeReport = $state<ReportCategory>("sensor");

  const categories = [
    { id: "sensor" as const, name: "report.tabSensor", defaultName: "センサー", icon: icons.mdiRadioTower },
    { id: "wifi" as const, name: "report.tabWifi", defaultName: "Wi-Fi AP", icon: icons.mdiWifi },
    { id: "bluetooth" as const, name: "report.tabBluetooth", defaultName: "Bluetooth (BLE)", icon: icons.mdiBluetooth },
    { id: "env" as const, name: "report.tabEnv", defaultName: "環境センサー", icon: icons.mdiThermometer },
    { id: "pcap" as const, name: "report.tabPcap", defaultName: "Pcap パケット", icon: icons.mdiEthernet },
    { id: "winlog" as const, name: "report.tabWinlog", defaultName: "Windows ログ", icon: icons.mdiMicrosoftWindows },
    { id: "twagent" as const, name: "report.tabTwagent", defaultName: "twagentneo", icon: icons.mdiRobot },
  ];
</script>

<div class="flex flex-col max-h-[calc(100vh-50px)] overflow-y-auto pb-2">
  <!-- Top Report Switcher Button Tabs -->
  <div class="flex flex-wrap items-center gap-1 mx-2 mt-2 pb-2 border-b border-gray-200 dark:border-gray-700">
    <ButtonGroup class="flex-wrap">
      {#each categories as cat}
        <Button
          size="xs"
          class="!py-1.5 !px-2.5 text-xs flex items-center gap-1.5"
          color={activeReport === cat.id ? "blue" : "alternative"}
          onclick={() => {
            activeReport = cat.id;
          }}
        >
          <Icon path={cat.icon} size={0.8} />
          <span>{$_(cat.name) || cat.defaultName}</span>
        </Button>
      {/each}
    </ButtonGroup>
  </div>

  <!-- Component Content Panel -->
  <div class="mx-2 mt-1">
    {#if activeReport === "sensor"}
      <SensorReport />
    {:else if activeReport === "wifi"}
      <WifiReport />
    {:else if activeReport === "bluetooth"}
      <BluetoothReport />
    {:else if activeReport === "env"}
      <EnvSensor />
    {:else if activeReport === "pcap"}
      <PcapReport />
    {:else if activeReport === "winlog"}
      <WinLogReport />
    {:else if activeReport === "twagent"}
      <AgentReport />
    {/if}
  </div>
</div>
