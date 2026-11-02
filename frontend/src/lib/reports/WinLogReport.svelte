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
  let infoModalOpen = $state(false);
  let deleteModalOpen = $state(false);
  let deletingItem = $state<{ id: string; name: string; kind: string } | null>(null);
  let showResetModal = $state(false);

  // Render helpers
  const renderWinLevel = (level: string) => {
    const l = (level || "").toLowerCase();
    if (l === "error" || l === "high") {
      return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300">エラー</span>`;
    }
    if (l === "warn" || l === "warning" || l === "low") {
      return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300">注意</span>`;
    }
    return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">正常</span>`;
  };

  const renderScore = (score: number | undefined) => {
    const s = typeof score === "number" ? score : 0;
    const formatted = s.toFixed(1);
    if (s < 50) {
      return `<span class="font-bold text-rose-600 dark:text-rose-400">${formatted}</span>`;
    }
    if (s < 75) {
      return `<span class="font-semibold text-amber-500">${formatted}</span>`;
    }
    return `<span class="font-medium text-emerald-600 dark:text-emerald-400">${formatted}</span>`;
  };

  const renderFailed = (f: number | undefined) => {
    const v = f || 0;
    if (v > 0) {
      return `<span class="text-rose-600 dark:text-rose-400 font-bold">${renderCount(v, "")}</span>`;
    }
    return `<span class="text-gray-400 dark:text-gray-500">0</span>`;
  };

  const renderStatus = (st: string | undefined) => {
    if (!st) return `<span class="text-gray-400">-</span>`;
    if (st === "0" || st === "0x0" || st.toLowerCase() === "success") {
      return `<span class="text-emerald-600 dark:text-emerald-400 font-mono text-xs">${st}</span>`;
    }
    return `<span class="text-rose-500 font-mono text-xs font-semibold">${st}</span>`;
  };

  // Columns definition conforming to twsnmpfc
  const getColumns = (tab: TabType) => {
    switch (tab) {
      case "eventid":
        return [
          {
            data: "Level",
            title: $_("report.winColLevel") || "状態",
            width: "10%",
            render: renderWinLevel,
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "16%",
          },
          {
            data: "Channel",
            title: $_("report.winColChannel") || "チャネル",
            width: "14%",
          },
          {
            data: "Provider",
            title: $_("report.winColProvider") || "プロバイダー",
            width: "18%",
          },
          {
            data: "EventID",
            title: "Event ID",
            width: "12%",
            render: (id: number) => `<span class="font-mono font-bold text-cyan-600 dark:text-cyan-400">${id}</span>`,
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "12%",
            render: renderCount,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "18%",
            render: renderTime,
          },
        ];

      case "logon":
        return [
          {
            data: "Score",
            title: $_("report.winColTrustScore") || "信用スコア",
            width: "10%",
            render: renderScore,
          },
          {
            data: "Target",
            title: $_("report.winColTargetAccount") || "ログオン先",
            width: "16%",
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "14%",
          },
          {
            data: "IP",
            title: $_("report.winColSrcIp") || "接続元",
            width: "14%",
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Logon",
            title: $_("report.winColLogon") || "ログオン",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Logoff",
            title: $_("report.winColLogoff") || "ログオフ",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Failed",
            title: $_("report.winColFailed") || "失敗",
            width: "8%",
            render: renderFailed,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "14%",
            render: renderTime,
          },
        ];

      case "account":
        return [
          {
            data: "Subject",
            title: $_("report.winColSubject") || "操作アカウント",
            width: "18%",
          },
          {
            data: "Target",
            title: $_("report.winColTarget") || "対象アカウント",
            width: "18%",
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "16%",
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "10%",
            render: renderCount,
          },
          {
            data: "Edit",
            title: $_("report.winColEdit") || "編集",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Password",
            title: $_("report.winColPassword") || "パスワード",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Other",
            title: $_("report.winColOther") || "その他",
            width: "8%",
            render: renderCount,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "14%",
            render: renderTime,
          },
        ];

      case "kerberos":
        return [
          {
            data: "Score",
            title: $_("report.winColTrustScore") || "信用スコア",
            width: "10%",
            render: renderScore,
          },
          {
            data: "Target",
            title: $_("report.winColTarget") || "対象",
            width: "16%",
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "14%",
          },
          {
            data: "IP",
            title: $_("report.winColClientIp") || "要求元",
            width: "12%",
          },
          {
            data: "Service",
            title: $_("report.winColService") || "サービス",
            width: "14%",
          },
          {
            data: "TicketType",
            title: $_("report.winColTicketType") || "種別",
            width: "10%",
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Failed",
            title: $_("report.winColFailed") || "失敗",
            width: "8%",
            render: renderFailed,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "14%",
            render: renderTime,
          },
        ];

      case "privilege":
        return [
          {
            data: "Subject",
            title: $_("report.winColSubject") || "対象アカウント",
            width: "35%",
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "25%",
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "15%",
            render: renderCount,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "25%",
            render: renderTime,
          },
        ];

      case "process":
        return [
          {
            data: "LastStatus",
            title: $_("report.winColStatus") || "ステータス",
            width: "10%",
            render: renderStatus,
          },
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "16%",
          },
          {
            data: "Process",
            title: $_("report.winColProcessName") || "プロセス",
            width: "22%",
          },
          {
            data: "LastSubject",
            title: $_("report.winColSubject") || "操作アカウント",
            width: "16%",
          },
          {
            data: "Start",
            title: $_("report.winColStart") || "開始",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Exit",
            title: $_("report.winColExit") || "終了",
            width: "8%",
            render: renderCount,
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "8%",
            render: renderCount,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "12%",
            render: renderTime,
          },
        ];

      case "task":
        return [
          {
            data: "Computer",
            title: $_("report.winColComputer") || "コンピューター",
            width: "20%",
          },
          {
            data: "TaskName",
            title: $_("report.winColTaskName") || "タスク",
            width: "35%",
          },
          {
            data: "Subject",
            title: $_("report.winColSubject") || "操作アカウント",
            width: "20%",
          },
          {
            data: "Count",
            title: $_("report.winColCount") || "回数",
            width: "10%",
            render: renderCount,
          },
          {
            data: "LastTime",
            title: $_("report.lastSeen") || "最終",
            width: "15%",
            render: renderTime,
          },
        ];
    }
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
    const cols = getColumns(activeTab);
    table = new DataTable("#winLogTable", {
      columns: cols,
      data: currentDataset,
      stateSave: false,
      order: [[cols.length - 1, "desc"]],
      pageLength: window.innerHeight > 800 ? 25 : 10,
      language: getTableLang(),
      select: {
        style: "single",
      },
    });

    const updateSelection = () => {
      selectedCount = table.rows({ selected: true }).count();
      const selData = table.rows({ selected: true }).data();
      selected = selData && selData.length > 0 ? selData[0] : null;
    };

    table.on("select", updateSelection);
    table.on("deselect", updateSelection);

    // Double-click row to open detail modal
    const tbody = document.querySelector("#winLogTable tbody");
    if (tbody) {
      tbody.addEventListener("dblclick", () => {
        if (selected) {
          infoModalOpen = true;
        }
      });
    }
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

  // KPI Calculations (Hybrid Mode C: 2 Global Summary Cards + 2 Active Tab Context Cards)
  const stats = $derived.by(() => {
    // 1. Total events & unique hosts
    const hostSet = new Set<string>();
    let totalEvents = 0;
    let errorWarnCount = 0;
    for (const e of eventIDs) {
      totalEvents += e.Count || 0;
      if (e.Computer) hostSet.add(e.Computer);
      const l = (e.Level || "").toLowerCase();
      if (l === "error" || l === "warn" || l === "warning" || l === "high" || l === "low") {
        errorWarnCount += e.Count || 0;
      }
    }
    for (const l of logons) if (l.Computer) hostSet.add(l.Computer);
    for (const a of accounts) if (a.Computer) hostSet.add(a.Computer);
    for (const k of kerberosList) if (k.Computer) hostSet.add(k.Computer);
    for (const p of privileges) if (p.Computer) hostSet.add(p.Computer);
    for (const pr of processes) if (pr.Computer) hostSet.add(pr.Computer);
    for (const t of tasks) if (t.Computer) hostSet.add(t.Computer);

    // 2. Auth failures & Threats
    let successLogons = 0;
    let failedLogons = 0;
    let lowScoreLogons = 0;
    for (const l of logons) {
      successLogons += l.Logon || 0;
      failedLogons += l.Failed || 0;
      if ((l.Score != null && l.Score < 50) || (l.Penalty != null && l.Penalty > 0)) {
        lowScoreLogons++;
      }
    }

    let kerberosTotal = 0;
    let kerberosFailed = 0;
    let lowScoreKerberos = 0;
    for (const k of kerberosList) {
      kerberosTotal += k.Count || 0;
      kerberosFailed += k.Failed || 0;
      if ((k.Score != null && k.Score < 50) || (k.Penalty != null && k.Penalty > 0)) {
        lowScoreKerberos++;
      }
    }

    const totalAuthFailures = failedLogons + kerberosFailed;
    const totalThreatAlerts = errorWarnCount + totalAuthFailures + lowScoreLogons + lowScoreKerberos;

    // 3. Tab-specific metrics
    let tabMetric1 = { label: "", value: "", icon: icons.mdiInformation, color: "text-blue-500", detail: "" };
    let tabMetric2 = { label: "", value: "", icon: icons.mdiInformation, color: "text-blue-500", detail: "" };

    switch (activeTab) {
      case "eventid":
        tabMetric1 = {
          label: $_("report.winKpiErrorWarnEvents") || "エラー・警告イベント",
          value: errorWarnCount.toLocaleString(),
          icon: icons.mdiAlertCircle,
          color: errorWarnCount > 0 ? "text-rose-500" : "text-emerald-500",
          detail: `全 ${eventIDs.length} 種類のID`,
        };
        tabMetric2 = {
          label: $_("report.winColEventID") || "登録Event ID種別",
          value: `${eventIDs.length} 種`,
          icon: icons.mdiCalendarCheck,
          color: "text-cyan-500",
          detail: `総イベント数: ${totalEvents.toLocaleString()}`,
        };
        break;

      case "logon":
        tabMetric1 = {
          label: $_("report.winKpiSuccessLogon") || "ログオン成功",
          value: successLogons.toLocaleString(),
          icon: icons.mdiShieldKey,
          color: "text-emerald-500",
          detail: `エントリ数: ${logons.length}`,
        };
        tabMetric2 = {
          label: $_("report.winKpiFailedLogon") || "ログオン失敗",
          value: failedLogons.toLocaleString(),
          icon: icons.mdiShieldAlert,
          color: failedLogons > 0 ? "text-rose-500" : "text-emerald-500",
          detail: `要注意(低スコア): ${lowScoreLogons}`,
        };
        break;

      case "account": {
        let edits = 0;
        let passes = 0;
        let others = 0;
        for (const a of accounts) {
          edits += a.Edit || 0;
          passes += a.Password || 0;
          others += a.Other || 0;
        }
        tabMetric1 = {
          label: $_("report.winColPassword") || "パスワード変更操作",
          value: passes.toLocaleString(),
          icon: icons.mdiLockAlert,
          color: passes > 0 ? "text-amber-500" : "text-gray-400",
          detail: `アカウントエントリ: ${accounts.length}`,
        };
        tabMetric2 = {
          label: $_("report.winColEdit") || "アカウント編集 / その他",
          value: `${edits.toLocaleString()} / ${others.toLocaleString()}`,
          icon: icons.mdiAccountCog,
          color: "text-blue-500",
          detail: `総操作回数: ${(edits + passes + others).toLocaleString()}`,
        };
        break;
      }

      case "kerberos":
        tabMetric1 = {
          label: $_("report.winColTicketType") || "Kerberos チケット要求",
          value: kerberosTotal.toLocaleString(),
          icon: icons.mdiKeyChain,
          color: "text-cyan-500",
          detail: `エントリ数: ${kerberosList.length}`,
        };
        tabMetric2 = {
          label: $_("report.winColFailed") || "チケット発行失敗",
          value: kerberosFailed.toLocaleString(),
          icon: icons.mdiKeyAlert,
          color: kerberosFailed > 0 ? "text-rose-500" : "text-emerald-500",
          detail: `要注意: ${lowScoreKerberos}`,
        };
        break;

      case "privilege": {
        let privTotal = 0;
        const privUserSet = new Set<string>();
        for (const p of privileges) {
          privTotal += p.Count || 0;
          if (p.Subject) privUserSet.add(p.Subject);
        }
        tabMetric1 = {
          label: $_("report.winTabPrivilege") || "特権行使総数",
          value: privTotal.toLocaleString(),
          icon: icons.mdiShieldCrown,
          color: "text-amber-500",
          detail: `エントリ数: ${privileges.length}`,
        };
        tabMetric2 = {
          label: $_("report.winColPrivUser") || "特権実行ユーザー数",
          value: `${privUserSet.size} ユーザー`,
          icon: icons.mdiAccountLock,
          color: "text-purple-500",
          detail: `監視ホスト数: ${hostSet.size} 台`,
        };
        break;
      }

      case "process": {
        let pStart = 0;
        let pExit = 0;
        let errorStatus = 0;
        for (const pr of processes) {
          pStart += pr.Start || 0;
          pExit += pr.Exit || 0;
          const st = (pr.LastStatus || "").trim();
          if (st && st !== "0" && st !== "0x0" && st.toLowerCase() !== "success") {
            errorStatus++;
          }
        }
        tabMetric1 = {
          label: $_("report.winColStartExit") || "プロセス起動 / 終了",
          value: `${pStart.toLocaleString()} / ${pExit.toLocaleString()}`,
          icon: icons.mdiCogSync,
          color: "text-blue-500",
          detail: `登録プロセス種別: ${processes.length}`,
        };
        tabMetric2 = {
          label: $_("report.winColStatus") || "異常ステータス終了",
          value: errorStatus.toLocaleString(),
          icon: icons.mdiAlertOctagon,
          color: errorStatus > 0 ? "text-rose-500" : "text-emerald-500",
          detail: `プロセス監視数: ${processes.length}`,
        };
        break;
      }

      case "task": {
        let taskTotal = 0;
        const taskUserSet = new Set<string>();
        for (const t of tasks) {
          taskTotal += t.Count || 0;
          if (t.Subject) taskUserSet.add(t.Subject);
        }
        tabMetric1 = {
          label: $_("report.winTitleTask") || "タスク実行総数",
          value: taskTotal.toLocaleString(),
          icon: icons.mdiCalendarClock,
          color: "text-blue-500",
          detail: `登録タスク種別: ${tasks.length}`,
        };
        tabMetric2 = {
          label: $_("report.winColExecAccount") || "タスク実行アカウント",
          value: `${taskUserSet.size} アカウント`,
          icon: icons.mdiAccountClock,
          color: "text-cyan-500",
          detail: `ホスト: ${hostSet.size} 台`,
        };
        break;
      }
    }

    return {
      totalEvents,
      uniqueHosts: hostSet.size,
      totalThreatAlerts,
      totalAuthFailures,
      errorWarnCount,
      lowScoreTotal: lowScoreLogons + lowScoreKerberos,
      tabMetric1,
      tabMetric2,
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
    const name = item.ID || `${item.Computer || ""} - ${item.EventID || item.Target || item.Subject || item.Process || item.TaskName || ""}`;
    const id = item.ID || name;
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

    switch (activeTab) {
      case "eventid": {
        const headers = ["Level", "Computer", "Channel", "Provider", "EventID", "Count", "First Seen", "Last Seen"];
        const rows = eventIDs.map((e) => [
          e.Level,
          e.Computer,
          e.Channel,
          e.Provider,
          e.EventID,
          e.Count,
          formatTimeStr(e.FirstTime),
          formatTimeStr(e.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "logon": {
        const headers = ["Score", "Target", "Computer", "IP", "Count", "Logon", "Logoff", "Failed", "First Seen", "Last Seen"];
        const rows = logons.map((l) => [
          l.Score ?? 0,
          l.Target,
          l.Computer,
          l.IP,
          l.Count,
          l.Logon,
          l.Logoff,
          l.Failed,
          formatTimeStr(l.FirstTime),
          formatTimeStr(l.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "account": {
        const headers = ["Subject", "Target", "Computer", "Count", "Edit", "Password", "Other", "First Seen", "Last Seen"];
        const rows = accounts.map((a) => [
          a.Subject,
          a.Target,
          a.Computer,
          a.Count,
          a.Edit,
          a.Password,
          a.Other,
          formatTimeStr(a.FirstTime),
          formatTimeStr(a.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "kerberos": {
        const headers = ["Score", "Target", "Computer", "IP", "Service", "TicketType", "Count", "Failed", "First Seen", "Last Seen"];
        const rows = kerberosList.map((k) => [
          k.Score ?? 0,
          k.Target,
          k.Computer,
          k.IP,
          k.Service,
          k.TicketType,
          k.Count,
          k.Failed,
          formatTimeStr(k.FirstTime),
          formatTimeStr(k.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "privilege": {
        const headers = ["Subject", "Computer", "Count", "First Seen", "Last Seen"];
        const rows = privileges.map((p) => [
          p.Subject,
          p.Computer,
          p.Count,
          formatTimeStr(p.FirstTime),
          formatTimeStr(p.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "process": {
        const headers = ["Status", "Computer", "Process", "Subject", "Start", "Exit", "Count", "First Seen", "Last Seen"];
        const rows = processes.map((pr) => [
          pr.LastStatus || "-",
          pr.Computer,
          pr.Process,
          pr.LastSubject,
          pr.Start,
          pr.Exit,
          pr.Count,
          formatTimeStr(pr.FirstTime),
          formatTimeStr(pr.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
      case "task": {
        const headers = ["Computer", "TaskName", "Subject", "Count", "First Seen", "Last Seen"];
        const rows = tasks.map((t) => [
          t.Computer,
          t.TaskName,
          t.Subject,
          t.Count,
          formatTimeStr(t.FirstTime),
          formatTimeStr(t.LastTime),
        ]);
        exportReportData(format, kind, headers, rows);
        break;
      }
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
  <!-- KPI Summary Cards (Hybrid Option C: 2 Global Summary + 2 Dynamic Tab Context) -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <!-- Global Card 1: Total Events & Monitored Hosts -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiMicrosoftWindows} size={0.9} class="text-blue-500" />
        {$_("report.winKpiTotalAndHosts") || "総イベント / 監視ホスト"}
      </div>
      <div class="mt-1 text-2xl font-bold">
        {stats.totalEvents.toLocaleString()}
        <span class="text-xs font-normal text-gray-500 ml-1">件</span>
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
        監視ホスト: <span class="font-semibold text-gray-700 dark:text-gray-300">{stats.uniqueHosts}</span> 台
      </div>
    </div>

    <!-- Global Card 2: Security Threats & Warnings -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
        <Icon path={icons.mdiShieldAlert} size={0.9} class={stats.totalThreatAlerts > 0 ? "text-rose-500" : "text-emerald-500"} />
        {$_("report.winKpiSecurityAlert") || "脅威・警告イベント"}
      </div>
      <div class="mt-1 text-2xl font-bold {stats.totalThreatAlerts > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'}">
        {stats.totalThreatAlerts.toLocaleString()}
        <span class="text-xs font-normal text-gray-500 ml-1">件</span>
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate" title={`認証失敗: ${stats.totalAuthFailures} / 警告: ${stats.errorWarnCount} / 要注意: ${stats.lowScoreTotal}`}>
        失敗: <span class="font-semibold text-rose-500">{stats.totalAuthFailures}</span> / 警告: {stats.errorWarnCount} / 要注意: {stats.lowScoreTotal}
      </div>
    </div>

    <!-- Tab Dynamic Card 1 -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400 truncate">
        <Icon path={stats.tabMetric1.icon} size={0.9} class={stats.tabMetric1.color} />
        {stats.tabMetric1.label}
      </div>
      <div class="mt-1 text-2xl font-bold {stats.tabMetric1.color}">
        {stats.tabMetric1.value}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
        {stats.tabMetric1.detail}
      </div>
    </div>

    <!-- Tab Dynamic Card 2 -->
    <div class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm">
      <div class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400 truncate">
        <Icon path={stats.tabMetric2.icon} size={0.9} class={stats.tabMetric2.color} />
        {stats.tabMetric2.label}
      </div>
      <div class="mt-1 text-2xl font-bold {stats.tabMetric2.color}">
        {stats.tabMetric2.value}
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
        {stats.tabMetric2.detail}
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
        color="purple"
        type="button"
        onclick={() => (infoModalOpen = true)}
        size="xs"
      >
        <Icon path={icons.mdiEye} size={1} />
        {$_("common.info") || "詳細"}
      </GradientButton>

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

<!-- Detailed Info Modal (twsnmpfc equivalent) -->
<Modal bind:open={infoModalOpen} size="md" dismissable={true} outsideclose={true}>
  {#if selected}
    <div class="p-3">
      <div class="flex items-center gap-2 mb-4 pb-2 border-b border-gray-200 dark:border-gray-700">
        <Icon path={icons.mdiInformationOutline} size={1.2} class="text-blue-500" />
        <h3 class="text-base font-semibold text-gray-800 dark:text-gray-100">
          {$_("report.winDetailTitle") || "Windows イベント詳細"} ({activeTab.toUpperCase()})
        </h3>
      </div>

      <div class="max-h-[60vh] overflow-y-auto space-y-2 text-xs">
        <table class="w-full text-left border-collapse">
          <tbody>
            <tr class="border-b border-gray-100 dark:border-gray-800">
              <th class="py-1.5 px-2 text-gray-500 font-medium w-1/3">{$_("report.winColComputer") || "コンピューター"}</th>
              <td class="py-1.5 px-2 font-mono font-medium">{selected.Computer || "-"}</td>
            </tr>

            {#if activeTab === "eventid"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColLevel") || "重要度"}</th>
                <td class="py-1.5 px-2 font-semibold">{@html renderWinLevel(selected.Level)}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">Event ID</th>
                <td class="py-1.5 px-2 font-mono text-cyan-600 font-bold">{selected.EventID}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColChannel") || "チャネル"}</th>
                <td class="py-1.5 px-2">{selected.Channel || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColProvider") || "プロバイダー"}</th>
                <td class="py-1.5 px-2">{selected.Provider || "-"}</td>
              </tr>

            {:else if activeTab === "logon"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTrustScore") || "信用スコア"}</th>
                <td class="py-1.5 px-2">{@html renderScore(selected.Score)} (ペナルティ: {selected.Penalty ?? 0})</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTargetAccount") || "ログオン先"}</th>
                <td class="py-1.5 px-2 font-mono font-medium">{selected.Target || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColSrcIp") || "接続元IP"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.IP || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColLogon") || "ログオン / ログオフ / 失敗"}</th>
                <td class="py-1.5 px-2">
                  <span class="text-emerald-600 font-semibold">{selected.Logon ?? 0}</span> / 
                  <span>{selected.Logoff ?? 0}</span> / 
                  <span class="text-rose-500 font-semibold">{selected.Failed ?? 0}</span>
                </td>
              </tr>
              {#if selected.LogonType && Object.keys(selected.LogonType).length > 0}
                <tr class="border-b border-gray-100 dark:border-gray-800">
                  <th class="py-1.5 px-2 text-gray-500 font-medium align-top">{$_("report.winDetailLogonType") || "ログオン種別内訳"}</th>
                  <td class="py-1.5 px-2">
                    <div class="flex flex-wrap gap-1">
                      {#each Object.entries(selected.LogonType) as [k, v]}
                        <span class="px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 text-xs">
                          {k}: <strong>{v}</strong>
                        </span>
                      {/each}
                    </div>
                  </td>
                </tr>
              {/if}
              {#if selected.FailedCode && Object.keys(selected.FailedCode).length > 0}
                <tr class="border-b border-gray-100 dark:border-gray-800">
                  <th class="py-1.5 px-2 text-gray-500 font-medium align-top">{$_("report.winDetailFailedCode") || "失敗理由内訳"}</th>
                  <td class="py-1.5 px-2">
                    <div class="flex flex-wrap gap-1">
                      {#each Object.entries(selected.FailedCode) as [k, v]}
                        <span class="px-1.5 py-0.5 rounded bg-rose-100 dark:bg-rose-900/40 text-rose-700 dark:text-rose-300 text-xs font-semibold">
                          {k}: {v}
                        </span>
                      {/each}
                    </div>
                  </td>
                </tr>
              {/if}

            {:else if activeTab === "account"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColSubject") || "操作アカウント"}</th>
                <td class="py-1.5 px-2 font-mono font-medium">{selected.Subject || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTarget") || "対象アカウント"}</th>
                <td class="py-1.5 px-2 font-mono font-medium">{selected.Target || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColEdit") || "編集 / パスワード / その他"}</th>
                <td class="py-1.5 px-2">
                  <span>編集: <strong>{selected.Edit ?? 0}</strong></span> / 
                  <span class="text-amber-500">パスワード: <strong>{selected.Password ?? 0}</strong></span> / 
                  <span>その他: <strong>{selected.Other ?? 0}</strong></span>
                </td>
              </tr>

            {:else if activeTab === "kerberos"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTrustScore") || "信用スコア"}</th>
                <td class="py-1.5 px-2">{@html renderScore(selected.Score)} (ペナルティ: {selected.Penalty ?? 0})</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTarget") || "対象"}</th>
                <td class="py-1.5 px-2 font-mono font-medium">{selected.Target || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColClientIp") || "要求元IP"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.IP || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColService") || "サービス"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.Service || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTicketType") || "チケット種別"}</th>
                <td class="py-1.5 px-2">{selected.TicketType || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColFailed") || "失敗回数"}</th>
                <td class="py-1.5 px-2">{@html renderFailed(selected.Failed)}</td>
              </tr>

            {:else if activeTab === "privilege"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColSubject") || "対象アカウント"}</th>
                <td class="py-1.5 px-2 font-mono font-medium">{selected.Subject || "-"}</td>
              </tr>

            {:else if activeTab === "process"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColProcessName") || "プロセス"}</th>
                <td class="py-1.5 px-2 font-mono font-semibold text-blue-600 dark:text-blue-400">{selected.Process || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColStatus") || "最終ステータス"}</th>
                <td class="py-1.5 px-2">{@html renderStatus(selected.LastStatus)}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColSubject") || "操作アカウント"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.LastSubject || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColParentProcess") || "親プロセス"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.LastParent || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColStartExit") || "起動 / 終了"}</th>
                <td class="py-1.5 px-2">起動: {selected.Start ?? 0} / 終了: {selected.Exit ?? 0}</td>
              </tr>

            {:else if activeTab === "task"}
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColTaskName") || "タスク名"}</th>
                <td class="py-1.5 px-2 font-mono font-semibold text-blue-600 dark:text-blue-400">{selected.TaskName || "-"}</td>
              </tr>
              <tr class="border-b border-gray-100 dark:border-gray-800">
                <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColSubject") || "操作アカウント"}</th>
                <td class="py-1.5 px-2 font-mono">{selected.Subject || "-"}</td>
              </tr>
            {/if}

            <tr class="border-b border-gray-100 dark:border-gray-800">
              <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winColCount") || "回数"}</th>
              <td class="py-1.5 px-2 font-semibold">{(selected.Count ?? 0).toLocaleString()} 回</td>
            </tr>
            <tr class="border-b border-gray-100 dark:border-gray-800">
              <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.winDetailFirstSeen") || "初回日時"}</th>
              <td class="py-1.5 px-2 font-mono">{formatTimeStr(selected.FirstTime)}</td>
            </tr>
            <tr>
              <th class="py-1.5 px-2 text-gray-500 font-medium">{$_("report.lastSeen") || "最終日時"}</th>
              <td class="py-1.5 px-2 font-mono">{formatTimeStr(selected.LastTime)}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="flex justify-end mt-4 pt-2 border-t border-gray-200 dark:border-gray-700">
        <GradientButton shadow color="teal" size="xs" onclick={() => (infoModalOpen = false)}>
          <Icon path={icons.mdiClose} size={1} />
          {$_("common.close") || "閉じる"}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

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
