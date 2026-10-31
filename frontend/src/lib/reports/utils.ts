import { ExportAny } from "../../../wailsjs/go/main/App";

export function isDarkMode(): boolean {
  return typeof document !== "undefined" && document.documentElement.classList.contains("dark");
}

export function formatTimeStr(ns: number | string | undefined | null): string {
  if (!ns) return "-";
  const num = typeof ns === "string" ? parseInt(ns, 10) : ns;
  if (!num || isNaN(num) || num <= 0) return "-";
  // Check if nanoseconds or milliseconds
  const ms = num > 1e14 ? Math.floor(num / 1e6) : num > 1e11 ? num : num * 1000;
  const d = new Date(ms);
  if (isNaN(d.getTime())) return "-";
  const Y = d.getFullYear();
  const M = String(d.getMonth() + 1).padStart(2, "0");
  const D = String(d.getDate()).padStart(2, "0");
  const h = String(d.getHours()).padStart(2, "0");
  const m = String(d.getMinutes()).padStart(2, "0");
  const s = String(d.getSeconds()).padStart(2, "0");
  return `${Y}/${M}/${D} ${h}:${m}:${s}`;
}

export function ipToNum(ip: string): number {
  if (!ip) return -1;
  const parts = ip.trim().split(".");
  if (parts.length === 4) {
    let num = 0;
    for (let i = 0; i < 4; i++) {
      const octet = parseInt(parts[i], 10);
      if (isNaN(octet) || octet < 0 || octet > 255) return -1;
      num = num * 256 + octet;
    }
    return num;
  }
  return -1;
}

export function getVendor(mac: string): string {
  if (!mac) return "Unknown / Generic";
  const clean = mac.replace(/[:-]/g, "").toUpperCase();
  if (clean.startsWith("525400") || clean.startsWith("00163E") || clean.startsWith("080027")) return "QEMU / KVM / Virtual";
  if (clean.startsWith("000C29") || clean.startsWith("005056") || clean.startsWith("000569")) return "VMware";
  if (clean.startsWith("001A2B") || clean.startsWith("00000C")) return "Cisco Systems";
  if (clean.startsWith("00A0DE") || clean.startsWith("AC44F2")) return "Yamaha Network";
  if (clean.startsWith("F01898") || clean.startsWith("ACDE48")) return "Apple";
  if (clean.startsWith("B827EB") || clean.startsWith("DCA632")) return "Raspberry Pi";
  return "Network Equipment";
}

export function isVirtualMachine(d: any): boolean {
  const v = (d.vendor || "").toLowerCase();
  const n = (d.name || "").toLowerCase();
  const m = (d.mac || "").toUpperCase().replace(/[:-]/g, "");
  return (
    v.includes("vmware") ||
    v.includes("qemu") ||
    v.includes("kvm") ||
    v.includes("virtual") ||
    v.includes("virtualbox") ||
    v.includes("hyper-v") ||
    v.includes("xen") ||
    v.includes("parallels") ||
    n.includes("vmware") ||
    n.includes("qemu") ||
    n.includes("kvm") ||
    n.includes("vbox") ||
    m.startsWith("525400") ||
    m.startsWith("00163E") ||
    m.startsWith("080027") ||
    m.startsWith("000569") ||
    m.startsWith("000C29") ||
    m.startsWith("005056") ||
    m.startsWith("00155D")
  );
}

export function getServiceName(port: number, proto: string): string {
  if (port === 443) return "TLS / HTTPS (TCP 443)";
  if (port === 80 || port === 8080) return "HTTP (TCP 80/8080)";
  if (port === 53) return "DNS (UDP/TCP 53)";
  if (port === 1812 || port === 1813) return "RADIUS (UDP 1812/1813)";
  if (port === 161 || port === 162) return "SNMP (UDP 161/162)";
  if (port === 123) return "NTP (UDP 123)";
  if (port === 22) return "SSH (TCP 22)";
  if (port === 514) return "Syslog (UDP 514)";
  if (port === 389 || port === 636) return "LDAP / LDAPS";
  if (port === 445 || port === 139) return "SMB / CIFS";
  if (port === 3389) return "RDP (TCP 3389)";
  if (port === 1883 || port === 8883) return "MQTT (TCP 1883/8883)";
  return `${(proto || "TCP").toUpperCase()}/${port}`;
}

export function renderBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let i = 0;
  let val = bytes;
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024;
    i++;
  }
  return `${val.toFixed(val >= 10 || i === 0 ? 1 : 2)} ${units[i]}`;
}

export async function exportReportData(
  format: string,
  title: string,
  headers: string[],
  rows: any[][]
): Promise<void> {
  const ed: any = {
    Title: title,
    Header: headers,
    Data: rows,
    Image: "",
  };
  try {
    const err = await ExportAny(format, ed);
    if (err) {
      console.error("Export error:", err);
    }
  } catch (err) {
    console.error("ExportAny call failed:", err);
  }
}

export function exportCSV(filename: string, headers: string[], rows: any[][]): void {
  exportReportData("csv", filename, headers, rows);
}

import * as echarts from "echarts";

export type EchartActionParams = {
  option: echarts.EChartsOption | null | undefined;
  theme?: string | object;
  notMerge?: boolean;
};

export function echartAction(
  node: HTMLElement,
  param: echarts.EChartsOption | EchartActionParams | null | undefined
) {
  let currentParam = param;
  let chart: echarts.ECharts | null = null;
  let ro: ResizeObserver | null = null;

  const getOption = (): echarts.EChartsOption | null | undefined => {
    if (!currentParam) return null;
    if (typeof currentParam === "object" && "option" in currentParam) {
      return (currentParam as EchartActionParams).option;
    }
    return currentParam as echarts.EChartsOption;
  };

  const getTheme = (): string | object | undefined => {
    if (currentParam && typeof currentParam === "object" && "theme" in currentParam) {
      return (currentParam as EchartActionParams).theme;
    }
    return undefined;
  };

  const getNotMerge = (): boolean => {
    if (currentParam && typeof currentParam === "object" && "notMerge" in currentParam) {
      return (currentParam as EchartActionParams).notMerge ?? true;
    }
    return true;
  };

  const initAndSet = () => {
    const opt = getOption();
    if (!opt) return;

    if (!chart || chart.isDisposed()) {
      const existing = echarts.getInstanceByDom(node);
      if (existing && !existing.isDisposed()) {
        chart = existing;
      } else {
        chart = echarts.init(node, getTheme());
      }
    }

    if (chart && !chart.isDisposed()) {
      chart.setOption(opt, getNotMerge());
      if (node.clientWidth > 0 && node.clientHeight > 0) {
        chart.resize();
      }
    }
  };

  initAndSet();

  if (typeof ResizeObserver !== "undefined") {
    ro = new ResizeObserver(() => {
      if (chart && !chart.isDisposed() && node.clientWidth > 0 && node.clientHeight > 0) {
        chart.resize();
      }
    });
    ro.observe(node);
  }

  const t1 = setTimeout(() => {
    if (chart && !chart.isDisposed() && node.clientWidth > 0 && node.clientHeight > 0) {
      chart.resize();
    }
  }, 100);

  const t2 = setTimeout(() => {
    if (chart && !chart.isDisposed() && node.clientWidth > 0 && node.clientHeight > 0) {
      chart.resize();
    }
  }, 300);

  return {
    update(newParam: echarts.EChartsOption | EchartActionParams | null | undefined) {
      currentParam = newParam;
      initAndSet();
    },
    destroy() {
      clearTimeout(t1);
      clearTimeout(t2);
      if (ro) {
        ro.disconnect();
        ro = null;
      }
      if (chart && !chart.isDisposed()) {
        try {
          chart.dispose();
        } catch (e) {}
        chart = null;
      }
    },
  };
}

