<script lang="ts">
  import "../../assets/css/jquery.dataTables.css";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { onMount, tick } from "svelte";
  import { _ } from "svelte-i18n";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Modal, Button, ButtonGroup, GradientButton } from "flowbite-svelte";
  import { formatTimeStr, exportReportData } from "./utils";
  import { getTableLang, renderCount, renderTime } from "../common";
  import {
    GetLogReport,
    ResetLogReport,
    DeleteLogReport,
  } from "../../../wailsjs/go/main/App";
  import type {
    WinEventIDEnt,
    WinLogonEnt,
    WinAccountEnt,
    WinKerberosEnt,
    WinPrivilegeEnt,
    WinProcessEnt,
    WinTaskEnt,
  } from "./types";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  type TabType =
    | "eventid"
    | "logon"
    | "account"
    | "kerberos"
    | "privilege"
    | "process"
    | "task";

  let activeTab = $state<TabType>("eventid");

  let eventIDs = $state<WinEventIDEnt[]>([]);
  let logons = $state<WinLogonEnt[]>([]);
  let accounts = $state<WinAccountEnt[]>([]);
  let kerberosList = $state<WinKerberosEnt[]>([]);
  let privileges = $state<WinPrivilegeEnt[]>([]);
  let processes = $state<WinProcessEnt[]>([]);
  let tasks = $state<WinTaskEnt[]>([]);

  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<any>(null);
  let selectedCount = $state(0);

  // Modals
  let deleteModalOpen = $state(false);
  let deletingItem = $state<{ id: string; name: string; kind: string } | null>(null);
  let showResetModal = $state(false);

  const getColumns = (tab: TabType) => {
    const commonFront = [
      {
        data: "Computer",
        title: $_("report.winColComputer") || "コンピューター",
        width: "15%",
      },
    ];

    let specificCols: any[] = [];
    if (tab === "eventid") {
      specificCols = [
        {
          data: "EventID",
          title: "Event ID",
          width: "12%",
          render: (id: number) => `<span class="font-mono font-bold text-cyan-500">${id}</span>`,
        },
        {
          data: "Provider",
          title: $_("report.winColProvider") || "プロバイダー",
          width: "20%",
        },
        {
          data: "Channel",
          title: $_("report.winColChannel") || "チャネル",
          width: "15%",
        },
        {
          data: "Level",
          title: $_("report.winColLevel") || "重要度",
          width: "10%",
        },
      ];
    } else if (tab === "logon") {
      specificCols = [
        {
          data: "Target",
          title: $_("report.winColTargetAccount") || "アカウント",
          width: "15%",
        },
        {
          data: "SrcIP",
          title: $_("report.winColSrcIp") || "接続元IP",
          width: "15%",
        },
        {
          data: "LogonCount",
          title: "Logon",
          width: "8%",
          render: renderCount,
        },
        {
          data: "FailedCount",
          title: "Failed",
          width: "8%",
          render: (f: number) =>
            f > 0
              ? `<span class="text-rose-500 font-bold">${renderCount(f, "")}</span>`
              : "0",
        },
        {
          data: "Score",
          title: $_("report.winColTrustScore") || "スコア",
          width: "8%",
        },
      ];
    } else if (tab === "account") {
      specificCols = [
        {
          data: "Subject",
          title: $_("report.winColSubject") || "実行者",
          width: "18%",
        },
        {
          data: "Target",
          title: $_("report.winColTarget") || "対象",
          width: "18%",
        },
        {
          data: "Type",
          title: $_("report.colType") || "操作種別",
          width: "18%",
        },
      ];
    } else if (tab === "kerberos") {
      specificCols = [
        {
          data: "User",
          title: $_("report.winColUser") || "ユーザー",
          width: "18%",
        },
        {
          data: "Service",
          title: $_("report.winColService") || "サービス",
          width: "18%",
        },
        {
          data: "TicketType",
          title: $_("report.winColTicketType") || "チケット種別",
          width: "18%",
        },
      ];
    } else if (tab === "privilege") {
      specificCols = [
        {
          data: "User",
          title: $_("report.winColPrivUser") || "特権実行者",
          width: "25%",
        },
        {
          data: "Privilege",
          title: "特権名",
          width: "25%",
          render: (p: string) => p || "-",
        },
      ];
    } else if (tab === "process") {
      specificCols = [
        {
          data: "Process",
          title: $_("report.winColProcessName") || "プロセス",
          width: "20%",
        },
        {
          data: "Executor",
          title: $_("report.winColExecutor") || "実行者",
          width: "15%",
        },
        {
          data: "Parent",
          title: $_("report.winColParentProcess") || "親プロセス",
          width: "20%",
        },
      ];
    } else {
      specificCols = [
        {
          data: "Task",
          title: $_("report.winColTaskName") || "タスク名",
          width: "25%",
        },
        {
          data: "ExecAccount",
          title: $_("report.winColExecAccount") || "実行アカウント",
          width: "20%",
        },
        {
          data: "Action",
          title: "アクション",
          width: "15%",
          render: (a: string) => a || "-",
        },
      ];
    }

    const commonEnd = [
      {
        data: "Count",
        title: $_("report.winColCount") || "件数",
        width: "10%",
        render: renderCount,
      },
      {
        data: "LastTime",
        title: $_("report.lastSeen") || "最終確認",
        width: "15%",
        render: renderTime,
      },
    ];

    return [...commonFront, ...specificCols, ...commonEnd];
  };

  const currentDataset = $derived.by(() => {
    switch (activeTab) {
      case "eventid":
        return eventIDs;
      case "logon":
        return logons;
      case "account":
        return accounts;
      case "kerberos":
        return kerberosList;
      case "privilege":
        return privileges;
      case "process":
        return processes;
      case "task":
        return tasks;
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
    const base = document.getElementById("winLogTableBase");
    if (!base) return;
    base.innerHTML = `<table id="winLogTable" class="display compact" style="width:99%"></table>`;

    selectedCount = 0;
    selected = null;
    table = new DataTable("#winLogTable", {
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

  const switchTab = (t: TabType) => {
    activeTab = t;
    showTable();
  };

  const loadAll = async () => {
    internalLoading = true;
    try {
      const [e, l, a, k, p, pr, t] = await Promise.all([
        GetLogReport("winEventID"),
        GetLogReport("winLogon"),
        GetLogReport("winAccount"),
        GetLogReport("winKerberos"),
        GetLogReport("winPrivilege"),
        GetLogReport("winProcess"),
        GetLogReport("winTask"),
      ]);
      eventIDs = (e || []) as WinEventIDEnt[];
      logons = (l || []) as WinLogonEnt[];
      accounts = (a || []) as WinAccountEnt[];
      kerberosList = (k || []) as WinKerberosEnt[];
      privileges = (p || []) as WinPrivilegeEnt[];
      processes = (pr || []) as WinProcessEnt[];
      tasks = (t || []) as WinTaskEnt[];
      await showTable();
    } catch (err) {
      console.error("Failed to load Windows Log report:", err);
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
    let totalEvents = 0;
    for (const e of eventIDs) totalEvents += e.Count || 0;

    let successLogons = 0;
    let failedLogons = 0;
    for (const l of logons) {
      successLogons += l.LogonCount || 0;
      failedLogons += l.FailedCount || 0;
    }

    let privUses = 0;
    for (const p of privileges) privUses += p.Count || 0;

    return {
      totalEvents,
      successLogons,
      failedLogons,
      privUses,
    };
  });

  const openDelete = (item: any) => {
    const kind =
      activeTab === "eventid"
        ? "winEventID"
        : activeTab === "logon"
        ? "winLogon"
        : activeTab === "account"
        ? "winAccount"
        : activeTab === "kerberos"
        ? "winKerberos"
        : activeTab === "privilege"
        ? "winPrivilege"
        : activeTab === "process"
        ? "winProcess"
        : "winTask";
    const id = item.ID || `${item.Computer}_${item.EventID || item.Target || item.User || item.Process || item.Task || ""}`;
    const name = `${item.Computer} - ${item.EventID || item.Target || item.User || item.Process || item.Task || ""}`;
    deletingItem = { id, name, kind };
    deleteModalOpen = true;
  };

  const confirmDelete = async () => {
    if (!deletingItem) return;
    try {
      await DeleteLogReport(deletingItem.kind, [deletingItem.id]);
      deleteModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to delete windows log item:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    const kind = `TWSNMP_Win_${activeTab}_Report`;
    if (activeTab === "eventid") {
      const headers = ["Computer", "Event ID", "Provider", "Channel", "Level", "Count", "First Seen", "Last Seen"];
      const rows = eventIDs.map((e) => [
        e.Computer,
        e.EventID,
        e.Provider,
        e.Channel,
        e.Level,
        e.Count,
        formatTimeStr(e.FirstTime),
        formatTimeStr(e.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else if (activeTab === "logon") {
      const headers = ["Computer", "Target Account", "Source IP", "Success Logon", "Failed Logon", "Trust Score", "First Seen", "Last Seen"];
      const rows = logons.map((l) => [
        l.Computer,
        l.Target,
        l.SrcIP,
        l.LogonCount,
        l.FailedCount,
        l.Score,
        formatTimeStr(l.FirstTime),
        formatTimeStr(l.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else {
      const headers = ["Computer", "Key 1", "Key 2", "Count", "First Seen", "Last Seen"];
      const rows = (currentDataset as any[]).map((it) => [
        it.Computer,
        it.Subject || it.User || it.Process || it.Task || "-",
        it.Target || it.Service || it.Action || "-",
        it.Count,
        formatTimeStr(it.FirstTime),
        formatTimeStr(it.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    }
  }

  export function handleClear(): void {
    showResetModal = true;
  }

  const confirmReset = async () => {
    showResetModal = false;
    const kind =
      activeTab === "eventid"
        ? "winEventID"
        : activeTab === "logon"
        ? "winLogon"
        : activeTab === "account"
        ? "winAccount"
        : activeTab === "kerberos"
        ? "winKerberos"
        : activeTab === "privilege"
        ? "winPrivilege"
        : activeTab === "process"
        ? "winProcess"
        : "winTask";
    try {
      await ResetLogReport(kind);
      await loadAll();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset windows log report:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiMicrosoftWindows} size={0.9} class="text-blue-500" />
        {$_("report.winKpiTotalEvents") || "総Windowsイベント数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.totalEvents.toLocaleString()}</div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiShieldKey} size={0.9} class="text-emerald-500" />
        {$_("report.winKpiSuccessLogon") || "ログオン成功"}
      </div>
      <div class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">
        {stats.successLogons.toLocaleString()}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiShieldAlert} size={0.9} class="text-rose-500" />
        {$_("report.winKpiFailedLogon") || "ログオン失敗"}
      </div>
      <div class="mt-1 text-2xl font-bold text-rose-600 dark:text-rose-400">
        {stats.failedLogons.toLocaleString()}
      </div>
    </div>

    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiShieldCrown} size={0.9} class="text-amber-500" />
        {$_("report.winTabPrivilege") || "特権操作検知"}
      </div>
      <div class="mt-1 text-2xl font-bold text-amber-600 dark:text-amber-400">
        {stats.privUses.toLocaleString()}
      </div>
    </div>
  </div>

  <!-- Sub-Tabs Switcher -->
  <div class="flex flex-wrap items-center gap-1.5 pb-1">
    <ButtonGroup class="flex-wrap">
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'eventid' ? 'blue' : 'alternative'}
        onclick={() => switchTab('eventid')}
      >
        <Icon path={icons.mdiCalendarCheck} size={0.8} />
        {$_("report.winTabEventID") || "Event ID"} ({eventIDs.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'logon' ? 'blue' : 'alternative'}
        onclick={() => switchTab('logon')}
      >
        <Icon path={icons.mdiLogin} size={0.8} />
        {$_("report.winTabLogon") || "ログオン"} ({logons.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'account' ? 'blue' : 'alternative'}
        onclick={() => switchTab('account')}
      >
        <Icon path={icons.mdiAccountCog} size={0.8} />
        {$_("report.winTabAccount") || "アカウント"} ({accounts.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'kerberos' ? 'blue' : 'alternative'}
        onclick={() => switchTab('kerberos')}
      >
        <Icon path={icons.mdiKeyChain} size={0.8} />
        {$_("report.winTabKerberos") || "Kerberos"} ({kerberosList.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'privilege' ? 'blue' : 'alternative'}
        onclick={() => switchTab('privilege')}
      >
        <Icon path={icons.mdiShieldCrown} size={0.8} />
        {$_("report.winTabPrivilege") || "特権利用"} ({privileges.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'process' ? 'blue' : 'alternative'}
        onclick={() => switchTab('process')}
      >
        <Icon path={icons.mdiCogSync} size={0.8} />
        {$_("report.winTabProcess") || "プロセス"} ({processes.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === 'task' ? 'blue' : 'alternative'}
        onclick={() => switchTab('task')}
      >
        <Icon path={icons.mdiCalendarClock} size={0.8} />
        {$_("report.winTabTask") || "タスク"} ({tasks.length})
      </Button>
    </ButtonGroup>
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div id="winLogTableBase">
      <table id="winLogTable" class="display compact" style="width:99%"></table>
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
      {$_("report.confirmClearItem", { values: { name: deletingItem?.name } }) || "選択したエントリーを削除しますか？"}
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
