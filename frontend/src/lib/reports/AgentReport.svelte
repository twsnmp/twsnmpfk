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
  import { getTableLang, renderTime, renderCount } from "../common";
  import {
    GetLogReport,
    ResetLogReport,
    DeleteLogReport,
    GetNodes,
    UpdateNode,
  } from "../../../wailsjs/go/main/App";
  import type {
    TwAgentInfoEnt,
    TwAgentAlertEnt,
    TwAgentFIMEnt,
    TwAgentSCAEnt,
    TwAgentTCPEnt,
  } from "./types";
  import type { datastore } from "../../../wailsjs/go/models";

  let {
    onRefresh = () => {},
  }: {
    onRefresh?: () => void;
  } = $props();

  type TabType = "agents" | "alerts" | "fim" | "sca" | "tcp";
  type AlertFilterType = "all" | "threat" | "compliance" | "benign";

  let activeTab = $state<TabType>("agents");
  let alertFilter = $state<AlertFilterType>("all");

  let agents = $state<TwAgentInfoEnt[]>([]);
  let alerts = $state<TwAgentAlertEnt[]>([]);
  let fimEvents = $state<TwAgentFIMEnt[]>([]);
  let scaFindings = $state<TwAgentSCAEnt[]>([]);
  let tcpConns = $state<TwAgentTCPEnt[]>([]);
  let nodesMap = $state<Record<string, datastore.NodeEnt>>({});

  let internalLoading = $state(false);

  let table: any = undefined;
  let selected = $state<any>(null);
  let selectedCount = $state(0);

  // Modals (RULE 1: Flowbite Modal, NEVER window.confirm)
  let deleteModalOpen = $state(false);
  let deletingItem = $state<{ id: string; name: string; kind: string } | null>(null);
  let showResetModal = $state(false);
  let nodeAddedMsg = $state<string | null>(null);

  const findMatchedNode = (agent: TwAgentInfoEnt): datastore.NodeEnt | undefined => {
    if (!agent) return undefined;
    const list = Object.values(nodesMap);
    return list.find(
      (n) =>
        (n.IP && n.IP === agent.IP) ||
        (n.MAC && agent.MAC && n.MAC.toLowerCase() === agent.MAC.toLowerCase()) ||
        (n.Name && (n.Name === agent.Host || n.Name === agent.ID))
    );
  };

  const handleAddNode = async (agent: TwAgentInfoEnt) => {
    const existing = findMatchedNode(agent);
    if (existing) {
      nodeAddedMsg = `ノード「${existing.Name} (${existing.IP})」は既に登録されています。`;
      return;
    }
    const newNode: datastore.NodeEnt = {
      ID: "",
      Name: agent.Host || agent.ID || agent.IP || "twagentneo",
      IP: agent.IP || "127.0.0.1",
      MAC: agent.MAC || "",
      Icon: agent.OS?.toLowerCase().includes("win") ? "windows" : "server",
      State: "normal",
      AddrMode: agent.IP ? "ip" : "host",
      Descr: `twagentneo ${agent.Version || ""} (${agent.OS || "agent"})`,
      X: 200,
      Y: 200,
      Community: "",
      SnmpMode: "",
      User: "",
      Password: "",
      Loc: "",
    };
    const ok = await UpdateNode(newNode);
    if (ok) {
      nodeAddedMsg = `ノード「${newNode.Name} (${newNode.IP})」をマップに追加しました。`;
      await loadAll();
      onRefresh();
    }
  };

  const getColumns = (tab: TabType) => {
    if (tab === "agents") {
      return [
        {
          data: "Host",
          title: $_("report.twagentColHostname") || "ホスト名 / ID",
          width: "20%",
          render: (h: string, type: any, row: any) => {
            const dispName = row.Hostname || row.Host || row.AgentID || row.ID || "-";
            const subName =
              row.AgentID && row.AgentID !== dispName
                ? row.AgentID
                : row.ID && row.ID !== dispName
                ? row.ID
                : "";
            return `<div class="flex flex-col min-w-0"><span class="font-bold truncate" title="${dispName}">${dispName}</span>${
              subName
                ? `<span class="text-[10px] text-gray-400 font-mono truncate" title="${subName}">${subName}</span>`
                : ""
            }</div>`;
          },
        },
        {
          data: "IP",
          title: "IP / MAC",
          width: "14%",
          render: (ip: string, type: any, row: any) => {
            const isIP =
              ip && /^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$/.test(ip.trim());
            const dispIP = isIP
              ? ip
              : row.IP &&
                /^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$/.test(row.IP.trim())
              ? row.IP
              : "-";
            return `<div class="font-mono text-xs"><div>${dispIP}</div><div class="text-[10px] text-gray-400">${
              row.MAC || ""
            }</div></div>`;
          },
        },
        {
          data: "OS",
          title: "OS / Arch",
          width: "14%",
          render: (os: string, type: any, row: any) =>
            `<div><div>${os || "-"}</div><div class="text-[10px] text-gray-400">${
              row.Arch || ""
            } ${row.Version ? "v" + row.Version : ""}</div></div>`,
        },
        {
          data: "CPUPercent",
          title: "CPU",
          width: "8%",
          render: (cpu: number) => {
            const val = cpu != null ? `${cpu.toFixed(1)}%` : "-";
            const cls =
              cpu > 80
                ? "text-rose-500 font-bold"
                : "text-cyan-600 dark:text-cyan-400 font-mono";
            return `<span class="${cls}">${val}</span>`;
          },
        },
        {
          data: "MemPercent",
          title: "Memory",
          width: "8%",
          render: (mem: number) => {
            const val = mem != null ? `${mem.toFixed(1)}%` : "-";
            const cls =
              mem > 80
                ? "text-rose-500 font-bold"
                : "text-purple-600 dark:text-purple-400 font-mono";
            return `<span class="${cls}">${val}</span>`;
          },
        },
        {
          data: "DiskPercent",
          title: "Disk",
          width: "8%",
          render: (disk: number) => {
            const val = disk != null ? `${disk.toFixed(1)}%` : "-";
            const cls =
              disk > 85
                ? "text-rose-500 font-bold"
                : "text-gray-600 dark:text-gray-400 font-mono";
            return `<span class="${cls}">${val}</span>`;
          },
        },
        {
          data: "Score",
          title: "スコア",
          width: "7%",
          render: (score: number) => {
            const s = score != null ? score : 100;
            const cls =
              s >= 67
                ? "text-emerald-500"
                : s >= 42
                ? "text-amber-500"
                : "text-rose-500";
            return `<span class="font-bold font-mono ${cls}">${s}</span>`;
          },
        },
        {
          data: "ID",
          title: "マップ登録",
          width: "9%",
          render: (id: string, type: any, row: any) => {
            const matched = findMatchedNode(row);
            return matched
              ? `<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">登録済</span>`
              : `<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300">未登録</span>`;
          },
        },
        {
          data: "LastTime",
          title: $_("report.lastSeen") || "最終確認",
          width: "12%",
          render: renderTime,
        },
      ];
    } else if (tab === "alerts") {
      return [
        {
          data: "Level",
          title: $_("report.twagentColLevel") || "レベル",
          width: "10%",
          render: (lvl: string) => {
            const isHigh =
              lvl === "high" ||
              lvl === "emergency" ||
              lvl === "alert" ||
              lvl === "critical";
            const cls = isHigh
              ? "bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300"
              : "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300";
            return `<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${cls}">${
              lvl || "-"
            }</span>`;
          },
        },
        {
          data: "Host",
          title: $_("report.twagentColHost") || "ホスト",
          width: "15%",
          render: (h: string, type: any, row: any) => h || row.Hostname || row.IP || "-",
        },
        {
          data: "Category",
          title: $_("report.twagentColCategory") || "カテゴリ",
          width: "12%",
          render: (c: string) => c || "-",
        },
        {
          data: "RuleTitle",
          title: $_("report.twagentColRuleTitle") || "検知ルール / AI要約",
          width: "25%",
          render: (title: string, type: any, row: any) => {
            const summary = row.SummaryJa || row.AiSummary;
            return `<div><div class="font-medium">${title || "-"}</div>${
              summary ? `<div class="text-[10px] text-gray-400 truncate">${summary}</div>` : ""
            }</div>`;
          },
        },
        {
          data: "ActionJa",
          title: "推奨アクション",
          width: "18%",
          render: (action: string, type: any, row: any) => row.ActionJa || row.AiAction || "-",
        },
        {
          data: "Count",
          title: "件数",
          width: "8%",
          render: renderCount,
        },
        {
          data: "LastTime",
          title: $_("report.lastSeen") || "最終確認",
          width: "12%",
          render: renderTime,
        },
      ];
    } else if (tab === "fim") {
      return [
        {
          data: "Host",
          title: $_("report.twagentColHost") || "ホスト",
          width: "15%",
          render: (h: string, type: any, row: any) => h || row.Hostname || row.IP || "-",
        },
        {
          data: "Op",
          title: $_("report.twagentColFimOp") || "操作",
          width: "10%",
          render: (op: string) => {
            const cls =
              op === "delete" || op === "REMOVE"
                ? "bg-rose-100 text-rose-800 dark:bg-rose-950"
                : "bg-cyan-100 text-cyan-800 dark:bg-cyan-950";
            return `<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${cls}">${
              op || "-"
            }</span>`;
          },
        },
        {
          data: "Path",
          title: $_("report.twagentColFimPath") || "監視ファイルパス",
          width: "40%",
          render: (p: string) => `<span class="font-mono text-xs">${p || "-"}</span>`,
        },
        {
          data: "Hash",
          title: "SHA256 ハッシュ",
          width: "15%",
          render: (h: string) =>
            h
              ? `<span class="font-mono text-[10px] text-gray-400" title="${h}">${h.substring(
                  0,
                  16
                )}...</span>`
              : "-",
        },
        {
          data: "Count",
          title: "件数",
          width: "8%",
          render: renderCount,
        },
        {
          data: "LastTime",
          title: $_("report.lastSeen") || "最終確認",
          width: "12%",
          render: renderTime,
        },
      ];
    } else if (tab === "sca") {
      return [
        {
          data: "Host",
          title: $_("report.twagentColHost") || "ホスト",
          width: "15%",
          render: (h: string, type: any, row: any) => h || row.Hostname || row.IP || "-",
        },
        {
          data: "Rule",
          title: $_("report.twagentColScaRule") || "監査ルール",
          width: "25%",
          render: (r: string, type: any, row: any) => {
            const title = row.Title || row.RuleID || row.Rule || "-";
            return `<div class="font-medium">${title}</div>`;
          },
        },
        {
          data: "Status",
          title: "状態 / 深刻度",
          width: "12%",
          render: (st: string, type: any, row: any) => {
            const sev = row.Severity || (st === "pass" ? "info" : "warn");
            const isCrit = sev === "critical" || sev === "high";
            const cls = isCrit
              ? "bg-rose-100 text-rose-800 dark:bg-rose-950"
              : "bg-emerald-100 text-emerald-800 dark:bg-emerald-950";
            return `<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium ${cls}">${
              row.Severity || st || "-"
            }</span>`;
          },
        },
        {
          data: "Remediation",
          title: "是正推奨手順 / 要約",
          width: "33%",
          render: (rem: string, type: any, row: any) => {
            const txt = rem || row.ActionJa || row.SummaryJa || row.Description || "-";
            return `<span class="text-xs text-gray-600 dark:text-gray-300">${txt}</span>`;
          },
        },
        {
          data: "Count",
          title: "件数",
          width: "7%",
          render: renderCount,
        },
        {
          data: "LastTime",
          title: $_("report.lastSeen") || "最終確認",
          width: "8%",
          render: renderTime,
        },
      ];
    } else {
      return [
        {
          data: "Host",
          title: $_("report.twagentColHost") || "ホスト",
          width: "15%",
          render: (h: string, type: any, row: any) => h || row.Hostname || row.IP || "-",
        },
        {
          data: "RemoteIP",
          title: "通信先 IP:Port",
          width: "20%",
          render: (rip: string, type: any, row: any) =>
            `<span class="font-mono">${rip || "-"}:${row.RemotePort || "-"}</span>`,
        },
        {
          data: "RemoteLoc",
          title: "国 / 地域",
          width: "15%",
          render: (loc: string, type: any, row: any) => loc || row.Location || "-",
        },
        {
          data: "Process",
          title: "プロセス",
          width: "20%",
          render: (proc: string) =>
            proc
              ? `<span class="font-mono text-cyan-600 dark:text-cyan-400">${proc}</span>`
              : "-",
        },
        {
          data: "Count",
          title: "累計接続数",
          width: "12%",
          render: renderCount,
        },
        {
          data: "LastTime",
          title: $_("report.lastSeen") || "最終確認",
          width: "18%",
          render: renderTime,
        },
      ];
    }
  };

  const currentDataset = $derived.by(() => {
    switch (activeTab) {
      case "agents":
        return agents;
      case "alerts": {
        let res = alerts;
        if (alertFilter === "threat") {
          res = res.filter((a) => a.AiVerdict === "threat" || a.Level === "high");
        } else if (alertFilter === "compliance") {
          res = res.filter((a) => a.Category === "compliance");
        } else if (alertFilter === "benign") {
          res = res.filter((a) => a.AiVerdict === "benign");
        }
        return res;
      }
      case "fim":
        return fimEvents;
      case "sca":
        return scaFindings;
      case "tcp":
        return tcpConns;
    }
  });

  const showTable = async () => {
    await tick();
    if (table) {
      try {
        table.destroy(true);
      } catch (e) {
        console.error("agent table destroy error:", e);
      }
      table = undefined;
    }
    const base = document.getElementById("agentTableBase");
    if (!base) return;
    base.innerHTML = `<table id="agentReportTable" class="display compact" style="width:99%"></table>`;

    selectedCount = 0;
    selected = null;
    table = new DataTable("#agentReportTable", {
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

  const switchAlertFilter = (f: AlertFilterType) => {
    alertFilter = f;
    showTable();
  };

  const loadAll = async () => {
    internalLoading = true;
    try {
      const [infoList, alertList, fimList, scaList, tcpList, nodesData] =
        await Promise.all([
          GetLogReport("twAgentInfo"),
          GetLogReport("twAgentAlert"),
          GetLogReport("twAgentFIM"),
          GetLogReport("twAgentSCA"),
          GetLogReport("twAgentTCP"),
          GetNodes(),
        ]);
      agents = (infoList || []) as TwAgentInfoEnt[];
      alerts = (alertList || []) as TwAgentAlertEnt[];
      fimEvents = (fimList || []) as TwAgentFIMEnt[];
      scaFindings = (scaList || []) as TwAgentSCAEnt[];
      tcpConns = (tcpList || []) as TwAgentTCPEnt[];
      nodesMap = nodesData || {};
      await showTable();
    } catch (err) {
      console.error("Failed to load twagent reports:", err);
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
    const totalAgents = agents.length;
    let threatCount = 0;
    let complianceCount = 0;
    for (const a of alerts) {
      if (a.AiVerdict === "threat" || a.Level === "high") threatCount++;
      if (a.Category === "compliance") complianceCount++;
    }
    const fimCount = fimEvents.length;
    const tcpCount = tcpConns.length;

    return {
      totalAgents,
      threatCount,
      complianceCount,
      fimCount,
      tcpCount,
    };
  });

  const openDelete = (item: any) => {
    const kind =
      activeTab === "agents"
        ? "twAgentInfo"
        : activeTab === "alerts"
        ? "twAgentAlert"
        : activeTab === "fim"
        ? "twAgentFIM"
        : activeTab === "sca"
        ? "twAgentSCA"
        : "twAgentTCP";
    deletingItem = { id: item.ID, name: item.Host || item.ID, kind };
    deleteModalOpen = true;
  };

  const confirmDelete = async () => {
    if (!deletingItem) return;
    try {
      await DeleteLogReport(deletingItem.kind, [deletingItem.id]);
      deleteModalOpen = false;
      await loadAll();
    } catch (err) {
      console.error("Failed to delete agent report item:", err);
    }
  };

  export function handleExport(format: string = "csv"): void {
    const kind =
      activeTab === "agents"
        ? "TWSNMP_Agent_Hosts"
        : activeTab === "alerts"
        ? "TWSNMP_Agent_Alerts"
        : activeTab === "fim"
        ? "TWSNMP_Agent_FIM"
        : activeTab === "sca"
        ? "TWSNMP_Agent_SCA"
        : "TWSNMP_Agent_TCP";

    if (activeTab === "agents") {
      const headers = [
        "Host",
        "Agent ID",
        "IP",
        "OS",
        "Arch",
        "Version",
        "CPU %",
        "Mem %",
        "Disk %",
        "Score",
        "Last Seen",
      ];
      const rows = agents.map((a) => [
        a.Host,
        a.ID,
        a.IP,
        a.OS,
        a.Arch,
        a.Version,
        a.CPUPercent,
        a.MemPercent,
        a.DiskPercent,
        a.Score,
        formatTimeStr(a.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else if (activeTab === "alerts") {
      const headers = [
        "Host",
        "Level",
        "Verdict",
        "Category",
        "Rule",
        "Summary",
        "Action",
        "Last Seen",
      ];
      const rows = alerts.map((a) => [
        a.Host,
        a.Level,
        a.AiVerdict,
        a.Category,
        a.RuleTitle,
        a.AiSummary,
        a.AiAction,
        formatTimeStr(a.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else if (activeTab === "fim") {
      const headers = [
        "Host",
        "Path",
        "Operation",
        "Hash",
        "Size (Bytes)",
        "Count",
        "Last Seen",
      ];
      const rows = fimEvents.map((f) => [
        f.Host,
        f.Path,
        f.Op,
        f.Hash,
        f.Size,
        f.Count,
        formatTimeStr(f.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else if (activeTab === "sca") {
      const headers = [
        "Host",
        "Rule",
        "Path",
        "Status",
        "Proof",
        "Remediation",
        "Last Seen",
      ];
      const rows = scaFindings.map((s) => [
        s.Host,
        s.Rule,
        s.Path,
        s.Status,
        s.Proof,
        s.Remediation,
        formatTimeStr(s.LastTime),
      ]);
      exportReportData(format, kind, headers, rows);
    } else {
      const headers = [
        "Host",
        "Remote IP",
        "Remote Port",
        "Process",
        "Location",
        "Count",
        "Last Seen",
      ];
      const rows = tcpConns.map((t) => [
        t.Host,
        t.RemoteIP,
        t.RemotePort,
        t.Process,
        t.Location,
        t.Count,
        formatTimeStr(t.LastTime),
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
      activeTab === "agents"
        ? "twAgentInfo"
        : activeTab === "alerts"
        ? "twAgentAlert"
        : activeTab === "fim"
        ? "twAgentFIM"
        : activeTab === "sca"
        ? "twAgentSCA"
        : "twAgentTCP";
    try {
      await ResetLogReport(kind);
      await loadAll();
      onRefresh();
    } catch (err) {
      console.error("Failed to reset agent reports:", err);
    }
  };
</script>

<div class="flex flex-col gap-3 p-2 text-gray-800 dark:text-gray-100">
  <!-- Feedback Alert -->
  {#if nodeAddedMsg}
    <div
      class="p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950/50 border border-emerald-300 dark:border-emerald-800 text-xs text-emerald-800 dark:text-emerald-200 flex items-center justify-between"
    >
      <span>{nodeAddedMsg}</span>
      <button
        type="button"
        class="cursor-pointer"
        onclick={() => (nodeAddedMsg = null)}
      >
        <Icon path={icons.mdiClose} size={0.8} />
      </button>
    </div>
  {/if}

  <!-- KPI Summary Cards -->
  <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
    <div
      class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm"
    >
      <div
        class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400"
      >
        <Icon path={icons.mdiRobot} size={0.9} class="text-cyan-500" />
        {$_("report.twagentKpiAgents") || "稼働エージェント数"}
      </div>
      <div class="mt-1 text-2xl font-bold">{stats.totalAgents}</div>
    </div>

    <div
      class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm"
    >
      <div
        class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400"
      >
        <Icon path={icons.mdiShieldAlert} size={0.9} class="text-rose-500" />
        {$_("report.twagentKpiThreats") || "AI 脅威検知数"}
      </div>
      <div class="mt-1 text-2xl font-bold text-rose-600 dark:text-rose-400">
        {stats.threatCount}
      </div>
    </div>

    <div
      class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm"
    >
      <div
        class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400"
      >
        <Icon path={icons.mdiFileDocumentAlert} size={0.9} class="text-amber-500" />
        {$_("report.twagentTabFim") || "FIM / SCA 検知"}
      </div>
      <div class="mt-1 text-2xl font-bold text-amber-600 dark:text-amber-400">
        {stats.fimCount}
      </div>
    </div>

    <div
      class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-3 shadow-sm"
    >
      <div
        class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400"
      >
        <Icon path={icons.mdiLanConnect} size={0.9} class="text-blue-500" />
        {$_("report.twagentTabTcp") || "プロセス通信フロー"}
      </div>
      <div class="mt-1 text-2xl font-bold text-blue-600 dark:text-blue-400">
        {stats.tcpCount}
      </div>
    </div>
  </div>

  <!-- Sub-Tabs Switcher & Filter -->
  <div class="flex flex-wrap items-center justify-between gap-2 pb-1">
    <ButtonGroup class="flex-wrap">
      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === "agents" ? "blue" : "alternative"}
        onclick={() => switchTab("agents")}
      >
        <Icon path={icons.mdiRobot} size={0.8} />
        {$_("report.twagentTabAgents") || "エージェント一覧"} ({agents.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === "alerts" ? "blue" : "alternative"}
        onclick={() => switchTab("alerts")}
      >
        <Icon path={icons.mdiShieldAlert} size={0.8} />
        {$_("report.twagentTabAlerts") || "AIセキュリティアラート"} ({alerts.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === "fim" ? "blue" : "alternative"}
        onclick={() => switchTab("fim")}
      >
        <Icon path={icons.mdiFileDocumentAlert} size={0.8} />
        {$_("report.twagentTabFim") || "FIM 整合性"} ({fimEvents.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === "sca" ? "blue" : "alternative"}
        onclick={() => switchTab("sca")}
      >
        <Icon path={icons.mdiCheckDecagram} size={0.8} />
        {$_("report.twagentTabSca") || "SCA 監査"} ({scaFindings.length})
      </Button>

      <Button
        size="xs"
        class="!py-1 !px-2.5 text-xs flex items-center gap-1"
        color={activeTab === "tcp" ? "blue" : "alternative"}
        onclick={() => switchTab("tcp")}
      >
        <Icon path={icons.mdiLanConnect} size={0.8} />
        {$_("report.twagentTabTcp") || "TCP 通信"} ({tcpConns.length})
      </Button>
    </ButtonGroup>

    {#if activeTab === "alerts"}
      <div class="flex items-center gap-1 text-xs">
        <button
          type="button"
          class="px-2 py-1 rounded text-[11px] cursor-pointer {alertFilter ===
          'all'
            ? 'bg-gray-200 dark:bg-gray-700 font-semibold'
            : 'text-gray-500 hover:bg-gray-100'}"
          onclick={() => switchAlertFilter("all")}
        >
          {$_("report.twagentFilterAll") || "すべて"}
        </button>
        <button
          type="button"
          class="px-2 py-1 rounded text-[11px] cursor-pointer {alertFilter ===
          'threat'
            ? 'bg-rose-100 text-rose-800 dark:bg-rose-950 font-semibold'
            : 'text-gray-500 hover:bg-gray-100'}"
          onclick={() => switchAlertFilter("threat")}
        >
          {$_("report.twagentFilterThreat") || "脅威のみ"}
        </button>
        <button
          type="button"
          class="px-2 py-1 rounded text-[11px] cursor-pointer {alertFilter ===
          'compliance'
            ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 font-semibold'
            : 'text-gray-500 hover:bg-gray-100'}"
          onclick={() => switchAlertFilter("compliance")}
        >
          {$_("report.twagentFilterCompliance") || "コンプライアンス"}
        </button>
        <button
          type="button"
          class="px-2 py-1 rounded text-[11px] cursor-pointer {alertFilter ===
          'benign'
            ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 font-semibold'
            : 'text-gray-500 hover:bg-gray-100'}"
          onclick={() => switchAlertFilter("benign")}
        >
          {$_("report.twagentFilterBenign") || "良性除外"}
        </button>
      </div>
    {/if}
  </div>

  <!-- DataTables Container (RULE 2: wrapped in a div) -->
  <div class="mx-1 mt-1">
    <div id="agentTableBase">
      <table id="agentReportTable" class="display compact" style="width:99%"></table>
    </div>
  </div>

  <!-- Unified Bottom Action Buttons Bar -->
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedCount === 1 && selected}
      {#if activeTab === "agents"}
        {#if !findMatchedNode(selected)}
          <GradientButton
            shadow
            color="green"
            type="button"
            onclick={() => selected && handleAddNode(selected)}
            size="xs"
          >
            <Icon path={icons.mdiPlus} size={1} />
            {$_("report.btnAddNode") || "+ マップ追加"}
          </GradientButton>
        {/if}

      {/if}

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
      onclick={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("report.btnRefresh") || "更新"}
    </GradientButton>
  </div>
</div>

<!-- Delete Item Modal (RULE 1: Flowbite Modal, NEVER window.confirm) -->
<Modal bind:open={deleteModalOpen} size="xs" dismissable={false} outsideclose={false}>
  <div class="text-center p-4">
    <Icon path={icons.mdiAlertCircleOutline} size={2.5} class="mx-auto mb-4 text-rose-500" />
    <h3 class="mb-5 text-sm font-normal text-gray-600 dark:text-gray-300">
      {$_("report.confirmClearItem", { values: { name: deletingItem?.name } }) ||
        "選択したエントリーを削除しますか？"}
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
      {$_("report.confirmClearItem", { values: { name: activeTab } }) ||
        `${activeTab} のレポートデータを全消去しますか？`}
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
