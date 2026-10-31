<script lang="ts">
  import "../../assets/css/jquery.dataTables.css";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { onMount, tick } from "svelte";
  import { _ } from "svelte-i18n";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Modal, Button, ButtonGroup, GradientButton } from "flowbite-svelte";
  import { formatTimeStr, isDarkMode, exportReportData } from "./utils";
  import { getTableLang, renderCount, renderTime } from "../common";
  import {
    GetLogReport,
    ResetLogReport,
    DeleteLogReport,
  } from "../../../wailsjs/go/main/App";
  import type {
    EtherTypeEnt,
    DNSQEnt,
    RADIUSFlowEnt,
    TLSFlowEnt,
  } from "./types";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  type TabType = "ether" | "dns" | "radius" | "tls";
  let activeTab = $state<TabType>("ether");

  let ethers = $state<EtherTypeEnt[]>([]);
  let dnsList = $state<DNSQEnt[]>([]);
  let radiusFlows = $state<RADIUSFlowEnt[]>([]);
  let tlsFlows = $state<TLSFlowEnt[]>([]);
  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<any>(null);
  let selectedCount = $state(0);

  // Modals
  let deleteModalOpen = $state(false);
  let deletingItem = $state<{ id: string; name: string; kind: string } | null>(null);
  let showResetModal = $state(false);

  const dnsTypeMap: Record<string, string> = {
    "1": "A",
    "28": "AAAA",
    "12": "PTR",
    "2": "NS",
    "15": "MX",
    "16": "TXT",
    "5": "CNAME",
    "6": "SOA",
    "65": "HTTPS",
    "255": "ANY",
  };

  const getDNSTypeLabel = (type: string): string => {
    return dnsTypeMap[type] || type || "A";
  };

  const formatHostWithIP = (name: string | undefined, ip: string | undefined): string => {
    const cleanIP = ip || "";
    const cleanName = name || "";
    if (cleanName && cleanName !== cleanIP) {
      return `${cleanName} <span class="text-xs text-gray-400">(${cleanIP})</span>`;
    }
    return cleanIP || cleanName || "-";
  };

  const getColumns = (tab: TabType) => {
    if (tab === "ether") {
      return [
        {
          data: "Name",
          title: "EtherType",
          width: "25%",
        },
        {
          data: "Type",
          title: "Code (Hex)",
          width: "15%",
          render: (t: string) => `<span class="font-mono text-cyan-500">${t}</span>`,
        },
        {
          data: "Count",
          title: "パケット数",
          width: "15%",
          render: renderCount,
        },
        {
          data: "Host",
          title: "ホスト",
          width: "20%",
        },
        {
          data: "LastTime",
          title: "最終確認",
          width: "25%",
          render: renderTime,
        },
      ];
    } else if (tab === "dns") {
      return [
        {
          data: "Name",
          title: "クエリ名 (FQDN)",
          width: "30%",
          render: (n: string) => `<span class="font-medium text-gray-900 dark:text-gray-100">${n}</span>`,
        },
        {
          data: "Type",
          title: "種別",
          width: "10%",
          render: (t: string) => `<span class="inline-flex px-1.5 py-0.5 rounded text-[11px] font-mono font-medium bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300">${getDNSTypeLabel(t)}</span>`,
        },
        {
          data: "LastClient",
          title: "クライアント",
          width: "15%",
          render: (c: string, _: any, row: any) => c || row?.Client || "-",
        },
        {
          data: "Server",
          title: "DNSサーバー",
          width: "15%",
        },
        {
          data: "Count",
          title: "回数",
          width: "10%",
          render: renderCount,
        },
        {
          data: "LastTime",
          title: "最終確認",
          width: "20%",
          render: renderTime,
        },
      ];
    } else if (tab === "radius") {
      return [
        {
          data: "Client",
          title: "クライアント",
          width: "20%",
          render: (c: string, __: string, row: any) => formatHostWithIP(row?.ClientName, c),
        },
        {
          data: "Server",
          title: "RADIUSサーバー",
          width: "20%",
          render: (s: string, __: string, row: any) => formatHostWithIP(row?.ServerName, s),
        },
        {
          data: "Count",
          title: "回数",
          width: "8%",
          render: renderCount,
        },
        {
          data: "Request",
          title: "Request",
          width: "8%",
          render: renderCount,
        },
        {
          data: "Challenge",
          title: "Challenge",
          width: "8%",
          render: renderCount,
        },
        {
          data: "Accept",
          title: "Accept",
          width: "8%",
          render: (a: number) => `<span class="text-emerald-500 font-semibold">${renderCount(a, "")}</span>`,
        },
        {
          data: "Reject",
          title: "Reject",
          width: "8%",
          render: (r: number) => r > 0 ? `<span class="text-rose-500 font-bold">${renderCount(r, "")}</span>` : "0",
        },
        {
          data: "Score",
          title: "スコア",
          width: "6%",
          render: (sc: number) => sc !== undefined ? sc.toFixed(1) : "-",
        },
        {
          data: "LastTime",
          title: "最終確認",
          width: "14%",
          render: renderTime,
        },
      ];
    } else {
      return [
        {
          data: "Client",
          title: "クライアント",
          width: "18%",
          render: (c: string, __: string, row: any) => formatHostWithIP(row?.ClientName, c),
        },
        {
          data: "Server",
          title: "サーバー",
          width: "18%",
          render: (s: string, __: string, row: any) => formatHostWithIP(row?.ServerName, s),
        },
        {
          data: "Service",
          title: "サービス",
          width: "8%",
          render: (sv: string) => sv || "-",
        },
        {
          data: "Version",
          title: "TLSバージョン",
          width: "10%",
          render: (v: string) => `<span class="font-mono text-cyan-500">${v || "-"}</span>`,
        },
        {
          data: "Cipher",
          title: "暗号スイート",
          width: "20%",
          render: (c: string) => c || '<span class="text-gray-400">-</span>',
        },
        {
          data: "Count",
          title: "回数",
          width: "7%",
          render: renderCount,
        },
        {
          data: "Score",
          title: "スコア",
          width: "6%",
          render: (sc: number) => sc !== undefined ? sc.toFixed(1) : "-",
        },
        {
          data: "LastTime",
          title: "最終確認",
          width: "13%",
          render: renderTime,
        },
      ];
    }
  };

  const currentDataset = $derived.by(() => {
    switch (activeTab) {
      case "ether":
        return ethers;
      case "dns":
        return dnsList;
      case "radius":
        return radiusFlows;
      case "tls":
        return tlsFlows;
    }
  });

  const showTable = async () => {
    await tick();
    if (table) {
      try {
        table.destroy(true);
      } catch (e) {
        console.error(e);
      }
      table = undefined;
    }
    const base = document.getElementById("pcapTableBase");
    if (!base) return;
    base.innerHTML = `<table id="pcapTable" class="display compact" style="width:99%"></table>`;

    selectedCount = 0;
    selected = null;
    table = new DataTable("#pcapTable", {
      columns: getColumns(activeTab),
      data: currentDataset,
      stateSave: false,
      order: [[getColumns(activeTab).length - 1, "desc"]],
      pageLength: window.innerHeight > 800 ? 25 : 10,
      language: getTableLang(),
      select: {
        style: "single",
      },
    });

    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
      const selData = table.rows({ selected: true }).data();
      selected = selData && selData.length > 0 ? selData[0] : null;
    });

    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
      const selData = table.rows({ selected: true }).data();
      selected = selData && selData.length > 0 ? selData[0] : null;
    });
  };

  const switchTab = (tab: TabType) => {
    activeTab = tab;
    showTable();
  };

  const loadAll = async () => {
    internalLoading = true;
    try {
      const [e, d, r, t] = await Promise.all([
        GetLogReport("etherType").catch((err) => { console.error("etherType error:", err); return []; }),
        GetLogReport("dnsq").catch((err) => { console.error("dnsq error:", err); return []; }),
        GetLogReport("radiusFlow").catch((err) => { console.error("radiusFlow error:", err); return []; }),
        GetLogReport("tlsFlow").catch((err) => { console.error("tlsFlow error:", err); return []; }),
      ]);
      ethers = (e || []) as EtherTypeEnt[];
      dnsList = (d || []) as DNSQEnt[];
      radiusFlows = (r || []) as RADIUSFlowEnt[];
      tlsFlows = (t || []) as TLSFlowEnt[];
      await showTable();
    } catch (err) {
      console.error("Failed to load Pcap report:", err);
    } finally {
      internalLoading = false;
    }
  };

  export const refresh = () => {
    loadAll();
  };

  onMount(() => {
    loadAll();
    return () => {
      if (table) {
        try {
          table.destroy(true);
        } catch (e) {}
        table = undefined;
      }
    };
  });

  const stats = $derived.by(() => {
    let etherPackets = 0;
    for (const e of ethers) etherPackets += e.Count || 0;
    let dnsQueries = 0;
    for (const d of dnsList) dnsQueries += d.Count || 0;
    let radiusTotal = 0;
    for (const r of radiusFlows) radiusTotal += (r.Request || 0) + (r.Accept || 0) + (r.Reject || 0);
    let tlsConns = 0;
    for (const t of tlsFlows) tlsConns += t.Count || 0;

    return {
      etherCount: ethers.length,
      etherPackets,
      dnsCount: dnsList.length,
      dnsQueries,
      radiusCount: radiusFlows.length,
      radiusTotal,
      tlsCount: tlsFlows.length,
      tlsConns,
    };
  });

  const openDelete = (item: any) => {
    const id = item.ID || item.Name || item.Client || "";
    const name = item.Name || item.Type || item.User || item.Client || "";
    deletingItem = { id, name, kind: activeTab };
    deleteModalOpen = true;
  };

  const confirmDelete = async () => {
    if (!deletingItem) return;
    const kind =
      deletingItem.kind === "ether"
        ? "etherType"
        : deletingItem.kind === "dns"
        ? "dnsq"
        : deletingItem.kind === "radius"
        ? "radiusFlow"
        : "tlsFlow";
    try {
      await DeleteLogReport(kind, [deletingItem.id]);
      deleteModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to delete pcap item:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    if (activeTab === "ether") {
      const headers = ["EtherType Name", "Code", "Count", "Host", "First Seen", "Last Seen"];
      const rows = ethers.map((e) => [
        e.Name,
        e.Type,
        e.Count,
        e.Host,
        formatTimeStr(e.FirstTime),
        formatTimeStr(e.LastTime),
      ]);
      exportReportData(format, "TWSNMP_Pcap_EtherTypes", headers, rows);
    } else if (activeTab === "dns") {
      const headers = ["Query Name", "Type", "Client", "Server", "Count", "First Seen", "Last Seen"];
      const rows = dnsList.map((d: any) => [
        d.Name,
        getDNSTypeLabel(d.Type),
        d.LastClient || d.Client || "",
        d.Server,
        d.Count,
        formatTimeStr(d.FirstTime),
        formatTimeStr(d.LastTime),
      ]);
      exportReportData(format, "TWSNMP_Pcap_DNS", headers, rows);
    } else if (activeTab === "radius") {
      const headers = ["Client", "Client Name", "Server", "Server Name", "Count", "Request", "Challenge", "Accept", "Reject", "Score", "First Seen", "Last Seen"];
      const rows = radiusFlows.map((r) => [
        r.Client,
        r.ClientName || r.Client,
        r.Server,
        r.ServerName || r.Server,
        r.Count || 0,
        r.Request,
        r.Challenge || 0,
        r.Accept,
        r.Reject,
        r.Score ?? "",
        formatTimeStr(r.FirstTime),
        formatTimeStr(r.LastTime),
      ]);
      exportReportData(format, "TWSNMP_Pcap_RADIUS", headers, rows);
    } else {
      const headers = ["Client", "Client Name", "Server", "Server Name", "Service", "Version", "Cipher", "Count", "Score", "First Seen", "Last Seen"];
      const rows = tlsFlows.map((t) => [
        t.Client,
        t.ClientName || t.Client,
        t.Server,
        t.ServerName || t.Server,
        t.Service,
        t.Version,
        t.Cipher,
        t.Count,
        t.Score ?? "",
        formatTimeStr(t.FirstTime),
        formatTimeStr(t.LastTime),
      ]);
      exportReportData(format, "TWSNMP_Pcap_TLS", headers, rows);
    }
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    const kind =
      activeTab === "ether"
        ? "etherType"
        : activeTab === "dns"
        ? "dnsq"
        : activeTab === "radius"
        ? "radiusFlow"
        : "tlsFlow";
    try {
      await ResetLogReport(kind);
      await loadAll();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset pcap report:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiEthernet} size={0.9} class="text-blue-500" />
        {$_("report.pcapEtherFrames") || "EtherType フレーム"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.etherCount}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Packets: {stats.etherPackets.toLocaleString()}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiDns} size={0.9} class="text-cyan-500" />
        {$_("report.pcapDnsQueries") || "DNS クエリ集計"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.dnsCount}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Queries: {stats.dnsQueries.toLocaleString()}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiShieldAccount} size={0.9} class="text-amber-500" />
        {$_("report.pcapRadiusFlows") || "RADIUS フロー"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.radiusCount}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Events: {stats.radiusTotal.toLocaleString()}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiLockCheck} size={0.9} class="text-emerald-500" />
        {$_("report.pcapTlsFlows") || "TLS フロー"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.tlsCount}</div>
      <div class="mt-1 text-[11px] text-gray-400">
        Sessions: {stats.tlsConns.toLocaleString()}
      </div>
    </div>
  </div>

  <!-- Sub-Tabs Switcher -->
  <div class="flex flex-wrap items-center gap-1.5 pb-1">
    <ButtonGroup class="flex-wrap">
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'ether' ? 'blue' : 'alternative'}
        onclick={() => switchTab('ether')}
      >
        <Icon path={icons.mdiEthernet} size={0.8} />
        {$_("report.pcapEtherTitle") || "EtherType"} ({ethers.length})
      </Button>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'dns' ? 'blue' : 'alternative'}
        onclick={() => switchTab('dns')}
      >
        <Icon path={icons.mdiDns} size={0.8} />
        {$_("report.pcapDnsTitle") || "DNS クエリ"} ({dnsList.length})
      </Button>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'radius' ? 'blue' : 'alternative'}
        onclick={() => switchTab('radius')}
      >
        <Icon path={icons.mdiShieldAccount} size={0.8} />
        {$_("report.pcapRadiusTitle") || "RADIUS 認証"} ({radiusFlows.length})
      </Button>
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'tls' ? 'blue' : 'alternative'}
        onclick={() => switchTab('tls')}
      >
        <Icon path={icons.mdiLockCheck} size={0.8} />
        {$_("report.pcapTlsTitle") || "TLS 暗号通信"} ({tlsFlows.length})
      </Button>
    </ButtonGroup>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div id="pcapTableBase">
      <table id="pcapTable" class="display compact" style="width:99%"></table>
    </div>
  </div>

  <!-- Unified Bottom Action Buttons Bar -->
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount === 1 && selected}
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={() => selected && openDelete(selected)}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.delete") || "削除"}
      </GradientButton>
    {/if}

    <GradientButton
      shadow
      color="red"
      type="button"
      onclick={handleClear}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("report.btnClear") || "全消去"}
    </GradientButton>

    <GradientButton
      shadow
      color="green"
      type="button"
      onclick={() => handleExport("csv")}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>

    <GradientButton
      shadow
      color="lime"
      type="button"
      onclick={() => handleExport("excel")}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>

    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={loadAll}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("report.btnReload") || "更新"}
    </GradientButton>
  </div>
</div>

<!-- Delete Item Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={deleteModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.pcapConfirmDelete") || "選択したエントリーを削除しますか？"}
    </h3>
    <div class="flex justify-center gap-3">
      <GradientButton shadow color="red" onclick={confirmDelete} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.delete") || "削除"}
      </GradientButton>
      <GradientButton shadow color="teal" onclick={() => (deleteModalOpen = false)} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("common.cancel") || "キャンセル"}
      </GradientButton>
    </div>
  </div>
</Modal>

<!-- Reset All Modal -->
<Modal bind:open={showResetModal} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.confirmClearItem", { values: { name: activeTab } }) || `${activeTab} のレポートデータを全消去しますか？`}
    </h3>
    <div class="flex justify-center gap-3">
      <GradientButton shadow color="red" onclick={confirmReset} size="xs">
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_("common.ok") || "全消去"}
      </GradientButton>
      <GradientButton shadow color="teal" onclick={() => (showResetModal = false)} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("common.cancel") || "キャンセル"}
      </GradientButton>
    </div>
  </div>
</Modal>
