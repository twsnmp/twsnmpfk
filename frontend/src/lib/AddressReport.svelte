<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { showArpGraph, showIPAMHeatmap } from "./chart/arp";
  import { _ } from "svelte-i18n";
  import { GetMapConf, LLMExplainAddressReport } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";

  export let show: boolean = false;
  export let arp: any = undefined;
  export let changeMAC: any = undefined;
  export let changeIP: any = undefined;
  export let ipam: any = [];
  export let nodes: any = undefined;

  let hasAI = false;
  let showAIReport = false;
  let activeTab = "ipam";

  let selectedRangeIndex: number = 0;
  let selectedSubnetBlock: string | null = null;
  let selectedHostInfo: any | null = null;
  let searchQuery: string = "";

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    chart = undefined;
    activeTab = "ipam";
    selectedRangeIndex = 0;
    selectedSubnetBlock = null;
    selectedHostInfo = null;
    searchQuery = "";
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    showIPAM();
  };

  let chart: any = undefined;
  const showChart = async (t: string) => {
    activeTab = t;
    await tick();
    chart = undefined;
    switch (t) {
      case "graphForce":
        chart = showArpGraph(t, arp, "force", changeIP, changeMAC);
        break;
      case "graphCircular":
        chart = showArpGraph(t, arp, "circular", changeIP, changeMAC);
        break;
    }
  };

  const showIPAM = async () => {
    activeTab = "ipam";
    await tick();
    chart = undefined;
    chart = showIPAMHeatmap("ipam", ipam, (idx: number) => {
      selectedRangeIndex = idx;
      selectedSubnetBlock = null;
      selectedHostInfo = null;
    });
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  // Collect effective known devices from nodes and ARP table
  $: effectiveDevices = (() => {
    const list: any[] = [];
    if (nodes) {
      Object.values(nodes).forEach((n: any) => {
        if (n && n.IP) {
          list.push({
            ip: n.IP,
            name: n.Name,
            mac: n.MAC || "",
            vendor: n.Vendor || "",
            isManaged: true,
          });
        }
      });
    }
    if (arp && Array.isArray(arp)) {
      arp.forEach((a: any) => {
        if (a && a.IP && !list.some((d) => d.ip === a.IP)) {
          const nodeName =
            a.NodeID && nodes && nodes[a.NodeID] ? nodes[a.NodeID].Name : "";
          list.push({
            ip: a.IP,
            name: nodeName || a.Vendor || a.IP,
            mac: a.MAC || "",
            vendor: a.Vendor || "",
            isManaged: !!nodeName,
          });
        }
      });
    }
    return list;
  })();

  // Summary statistics across all IPAM ranges
  $: totalRanges = (ipam || []).length;
  $: totalSize = (ipam || []).reduce(
    (acc: number, r: any) => acc + (Number(r.Size) || 0),
    0
  );
  $: totalUsed = (ipam || []).reduce(
    (acc: number, r: any) => acc + (Number(r.Used) || 0),
    0
  );
  $: totalUsage = totalSize > 0 ? (totalUsed * 100.0) / totalSize : 0;

  // Selected range
  $: currentRange =
    ipam && ipam.length > 0
      ? ipam[Math.min(Math.max(0, selectedRangeIndex), ipam.length - 1)]
      : null;
  $: isLargeRange = currentRange ? currentRange.Size > 256 : false;

  // Active /24 subnet prefix
  $: activeSubnetPrefix = (() => {
    if (selectedSubnetBlock) {
      return selectedSubnetBlock.replace(/\.0\/24$/, "");
    }
    if (currentRange && currentRange.Size <= 256 && currentRange.StartIP) {
      const parts = currentRange.StartIP.split(".");
      if (parts.length === 4) {
        return `${parts[0]}.${parts[1]}.${parts[2]}`;
      }
    }
    return "";
  })();

  // Map of 1..254 host entries for the currently viewed /24 subnet
  $: currentSubnetHosts = (() => {
    const prefix = activeSubnetPrefix;
    const hostMap = new Map<number, any>();
    if (!prefix) return { prefix: "", hostMap };
    for (const d of effectiveDevices) {
      if (d.ip && d.ip.startsWith(prefix + ".")) {
        const last = parseInt(d.ip.substring(prefix.length + 1), 10);
        if (!isNaN(last) && last >= 1 && last <= 254) {
          hostMap.set(last, d);
        }
      }
    }
    return { prefix, hostMap };
  })();

  const getUsageBadgeColor = (usage: number) => {
    if (usage < 50) return "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30";
    if (usage < 80) return "bg-cyan-500/20 text-cyan-400 border border-cyan-500/30";
    if (usage < 95) return "bg-amber-500/20 text-amber-400 border border-amber-500/30";
    return "bg-rose-500/20 text-rose-400 border border-rose-500/30";
  };

  const getUsageProgressBarColor = (usage: number) => {
    if (usage < 50) return "bg-emerald-500";
    if (usage < 80) return "bg-cyan-500";
    if (usage < 95) return "bg-amber-500";
    return "bg-rose-500";
  };

  $: if (show) {
    onOpen();
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[90vh]"
>
  <div class="flex flex-col space-y-4 max-h-[85vh] overflow-y-auto pr-1">
    <Tabs style="underline">
      <!-- Tab 1: IPAM Hierarchical Report -->
      <TabItem
        open
        onclick={() => {
          showIPAM();
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("ArpReport.IPAM")}
          </div>
        {/snippet}

        <div class="space-y-5 mt-2">
          <!-- KPI Summary Cards across all ranges -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-3.5 space-y-1.5 shadow-sm">
              <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
                <span>{$_("ArpReport.TargetRangeCount")}</span>
                <Icon path={icons.mdiFolderTree} size={0.8} class="text-cyan-400" />
              </div>
              <div class="text-2xl font-bold font-mono text-cyan-400">
                {totalRanges} <span class="text-xs font-normal text-slate-400">{$_("ArpReport.UnitRanges")}</span>
              </div>
              <div class="text-[11px] text-slate-400 truncate">
                {currentRange ? $_("ArpReport.SelectedRange", { values: { range: currentRange.Range } }) : $_("ArpReport.NoSubnetDetected")}
              </div>
            </div>

            <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-3.5 space-y-1.5 shadow-sm">
              <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
                <span>{$_("ArpReport.TotalPoolAddresses")}</span>
                <Icon path={icons.mdiLayers || icons.mdiViewGridOutline} size={0.8} class="text-cyan-400" />
              </div>
              <div class="text-2xl font-bold font-mono text-slate-100">
                {totalSize.toLocaleString()} <span class="text-xs font-normal text-slate-400">{$_("ArpReport.UnitAddresses")}</span>
              </div>
              <div class="text-[11px] text-slate-400">{$_("ArpReport.PoolSub")}</div>
            </div>

            <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-3.5 space-y-1.5 shadow-sm">
              <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
                <span>{$_("ArpReport.UsedIpCount")}</span>
                <Icon path={icons.mdiCheckCircle} size={0.8} class="text-emerald-400" />
              </div>
              <div class="text-2xl font-bold font-mono text-emerald-400">
                {totalUsed.toLocaleString()} <span class="text-xs font-normal text-slate-400">/ {totalSize.toLocaleString()}</span>
              </div>
              <div class="text-[11px] text-emerald-400/80">{$_("ArpReport.UsedSub")}</div>
            </div>

            <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-3.5 space-y-1.5 shadow-sm">
              <div class="flex items-center justify-between text-xs font-semibold text-slate-400">
                <span>{$_("ArpReport.AvgUsageRate")}</span>
                <Icon path={icons.mdiPulse || icons.mdiChartLine} size={0.8} class="text-cyan-300" />
              </div>
              <div class="text-2xl font-bold font-mono text-cyan-300">
                {totalUsage.toFixed(1)} <span class="text-xs font-normal text-slate-400">%</span>
              </div>
              <div class="text-[11px] text-slate-400">
                {$_("ArpReport.FreeAddresses", { values: { count: (totalSize - totalUsed).toLocaleString() } })}
              </div>
            </div>
          </div>

          <!-- Section 1: Relative Density Heatmap -->
          <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-4 space-y-2 shadow-sm">
            <div class="flex flex-wrap items-center justify-between border-b border-slate-700/80 pb-2 gap-2">
              <div>
                <h3 class="text-xs font-bold text-slate-200 flex items-center gap-1.5">
                  <Icon path={icons.mdiLan || icons.mdiNetwork} size={0.8} class="text-cyan-400" />
                  {$_("ArpReport.HeatmapDensityTitle")}
                </h3>
                <p class="text-[11px] text-slate-400 mt-0.5">
                  {$_("ArpReport.HeatmapDensityDesc")}
                </p>
              </div>
              <div class="text-[10px] text-slate-400 font-mono">
                {$_("ArpReport.HeatmapDensityHint")}
              </div>
            </div>

            {#if !ipam || ipam.length === 0}
              <div class="p-6 text-center text-slate-400 text-xs">
                {$_("ArpReport.NoIpamRanges")}
              </div>
            {:else}
              <div id="ipam"></div>
            {/if}
          </div>

          <!-- Section 2: Subnets List Table -->
          <div class="rounded-xl border border-slate-700 bg-slate-900/80 shadow-sm overflow-hidden flex flex-col">
            <div class="border-b border-slate-700/80 bg-slate-950/60 px-4 py-2.5 flex items-center justify-between">
              <span class="text-xs font-bold text-slate-200 flex items-center gap-1.5">
                <Icon path={icons.mdiFolderTree} size={0.8} class="text-cyan-400" />
                {$_("ArpReport.SubnetListTitle", { values: { count: (ipam || []).length } })}
              </span>
              <span class="text-[11px] text-slate-400">{$_("ArpReport.SubnetListHint")}</span>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs border-collapse font-mono">
                <thead class="bg-slate-950 text-slate-400 uppercase text-[10px] font-semibold tracking-wider border-b border-slate-800 select-none">
                  <tr>
                    <th class="py-2 px-3 text-center w-12">{$_("ArpReport.ColSelect")}</th>
                    <th class="py-2 px-3">{$_("ArpReport.ColRange")}</th>
                    <th class="py-2 px-3">{$_("ArpReport.ColStartIp")}</th>
                    <th class="py-2 px-3">{$_("ArpReport.ColEndIp")}</th>
                    <th class="py-2 px-3 text-right">{$_("ArpReport.ColSize")}</th>
                    <th class="py-2 px-3 text-right">{$_("ArpReport.ColUsed")}</th>
                    <th class="py-2 px-3 w-48">{$_("ArpReport.ColUsage")}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-800/60 text-slate-300">
                  {#if !ipam || ipam.length === 0}
                    <tr>
                      <td colspan="7" class="p-6 text-center text-slate-400">
                        {$_("ArpReport.NoRangesRegistered")}
                      </td>
                    </tr>
                  {:else}
                    {#each ipam as r, idx}
                      {@const isSelected = selectedRangeIndex === idx}
                      <tr
                        onclick={() => {
                          selectedRangeIndex = idx;
                          selectedSubnetBlock = null;
                          selectedHostInfo = null;
                        }}
                        class="cursor-pointer transition-colors {isSelected ? 'bg-cyan-950/50 text-white font-semibold' : 'hover:bg-slate-800/50'}"
                      >
                        <td class="py-2 px-3 text-center">
                          {#if isSelected}
                            <span class="h-2.5 w-2.5 rounded-full bg-cyan-400 inline-block animate-pulse"></span>
                          {:else}
                            <span class="h-1.5 w-1.5 rounded-full bg-slate-600 inline-block"></span>
                          {/if}
                        </td>
                        <td class="py-2 px-3 text-cyan-400 font-bold flex items-center gap-1.5">
                          <span>{r.Range}</span>
                          {#if r.Size > 256}
                            <span class="rounded bg-indigo-950 border border-indigo-800/70 px-1 text-[9px] text-indigo-300 leading-none">
                              {$_("ArpReport.WideArea")}
                            </span>
                          {/if}
                        </td>
                        <td class="py-2 px-3 text-slate-300">{r.StartIP || "-"}</td>
                        <td class="py-2 px-3 text-slate-300">{r.EndIP || "-"}</td>
                        <td class="py-2 px-3 text-right text-slate-200">{Number(r.Size || 0).toLocaleString()}</td>
                        <td class="py-2 px-3 text-right text-emerald-400 font-bold">{Number(r.Used || 0).toLocaleString()}</td>
                        <td class="py-2 px-3">
                          <div class="flex items-center gap-2">
                            <div class="flex-1 h-2 rounded-full bg-slate-800 overflow-hidden">
                              <div
                                class="h-full rounded-full transition-all {getUsageProgressBarColor(r.Usage || 0)}"
                                style="width: {Math.min(100, Math.max(r.Used > 0 ? 3 : 0, r.Usage || 0))}%"
                              ></div>
                            </div>
                            <span class="text-[10px] w-12 text-right font-mono {r.Usage >= 90 ? 'text-rose-400 font-bold' : 'text-slate-300'}">
                              {(r.Usage || 0).toFixed(1)}%
                            </span>
                          </div>
                        </td>
                      </tr>
                    {/each}
                  {/if}
                </tbody>
              </table>
            </div>
          </div>

          <!-- Section 3: Adaptive Drill-down Visual View -->
          {#if currentRange}
            <div class="rounded-xl border border-slate-700 bg-slate-900/80 p-4 shadow-sm space-y-3">
              <!-- Navigation Header & Breadcrumb -->
              <div class="flex flex-wrap items-center justify-between border-b border-slate-700/80 pb-3 gap-3">
                <div class="flex items-center gap-2">
                  <Icon path={icons.mdiGrid} size={0.9} class="text-cyan-400" />
                  <div class="flex items-center gap-1.5 text-xs font-bold text-slate-200">
                    <button
                      type="button"
                      onclick={() => {
                        selectedSubnetBlock = null;
                        selectedHostInfo = null;
                      }}
                      class="text-slate-400 hover:text-cyan-400 transition-colors cursor-pointer"
                    >
                      {$_("ArpReport.RangeDetailHeader", {
                        values: {
                          range: currentRange.Range,
                          size: Number(currentRange.Size || 0).toLocaleString(),
                        },
                      })}
                    </button>
                    {#if selectedSubnetBlock}
                      <Icon path={icons.mdiChevronRight} size={0.7} class="text-slate-500" />
                      <span class="text-cyan-400 font-mono">
                        {selectedSubnetBlock} {$_("ArpReport.DetailTag")}
                      </span>
                    {/if}
                  </div>
                </div>

                <!-- Controls, Search & Legend -->
                <div class="flex flex-wrap items-center gap-3 text-[11px]">
                  {#if selectedSubnetBlock}
                    <button
                      type="button"
                      onclick={() => {
                        selectedSubnetBlock = null;
                        selectedHostInfo = null;
                      }}
                      class="flex items-center gap-1 text-xs text-cyan-400 hover:text-cyan-300 font-semibold cursor-pointer rounded-lg bg-cyan-950/60 border border-cyan-800/60 px-2.5 py-1"
                    >
                      <Icon path={icons.mdiUndo} size={0.7} />
                      <span>{$_("ArpReport.BtnBackToBlocks")}</span>
                    </button>
                  {/if}

                  <div class="flex items-center gap-2.5 text-[10px] text-slate-400">
                    <span class="flex items-center gap-1">
                      <span class="h-2.5 w-2.5 rounded bg-emerald-500"></span>
                      {$_("ArpReport.LegendUsed")}
                    </span>
                    <span class="flex items-center gap-1">
                      <span class="h-2.5 w-2.5 rounded bg-slate-800 border border-slate-700"></span>
                      {$_("ArpReport.LegendFree")}
                    </span>
                    {#if searchQuery}
                      <span class="flex items-center gap-1 text-amber-400">
                        <span class="h-2.5 w-2.5 rounded bg-amber-400 animate-pulse"></span>
                        {$_("ArpReport.LegendSearchMatch")}
                      </span>
                    {/if}
                  </div>

                  <input
                    type="text"
                    bind:value={searchQuery}
                    placeholder={$_("ArpReport.SearchPlaceholder")}
                    class="rounded-lg bg-slate-950 border border-slate-700 px-2.5 py-1 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-400 w-44 font-mono"
                  />
                </div>
              </div>

              <!-- CASE 1: Large Range Top-level /24 Block Heatmap -->
              {#if isLargeRange && !selectedSubnetBlock}
                <div class="space-y-2">
                  <div class="flex items-center justify-between text-[11px] text-slate-400">
                    <span>
                      {$_("ArpReport.WideAreaHint")}
                    </span>
                    <span class="font-mono text-cyan-400 font-bold shrink-0">
                      {currentRange.Subnets?.length || 0} {$_("ArpReport.UnitBlocks")}
                    </span>
                  </div>

                  <div
                    class="grid gap-1.5 max-h-96 overflow-y-auto p-2.5 rounded-xl bg-slate-950/80 border border-slate-800/80"
                    style="grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));"
                  >
                    {#each currentRange.Subnets || [] as block}
                      {@const hasUsed = block.Used > 0}
                      {@const isQueryMatch =
                        searchQuery &&
                        block.Subnet.toLowerCase().includes(searchQuery.toLowerCase())}
                      <button
                        type="button"
                        onclick={() => {
                          selectedSubnetBlock = block.Subnet;
                          selectedHostInfo = null;
                        }}
                        title={$_("ArpReport.BlockTooltip", {
                          values: {
                            subnet: block.Subnet,
                            used: block.Used,
                            size: block.Size,
                            usage: block.Usage.toFixed(1),
                          },
                        })}
                        class="p-2 rounded-lg border text-left flex flex-col justify-between transition-all cursor-pointer hover:scale-105 hover:z-10 {isQueryMatch
                          ? 'ring-2 ring-amber-400 border-amber-400 bg-amber-950/30'
                          : hasUsed
                          ? 'bg-slate-900 border-cyan-800/60 hover:border-cyan-400 shadow-xs'
                          : 'bg-slate-950/60 border-slate-800/80 hover:bg-slate-900 opacity-60'}"
                      >
                        <div class="text-[10px] font-mono font-bold truncate text-slate-200">
                          {block.Subnet.replace(/\.0\/24$/, "")}
                        </div>
                        <div class="flex items-center justify-between mt-1 text-[9px] font-mono">
                          <span class={hasUsed ? "text-emerald-400 font-bold" : "text-slate-500"}>
                            {block.Used}
                          </span>
                          <span class="rounded px-1 py-0 text-[8px] {getUsageBadgeColor(block.Usage)}">
                            {block.Usage.toFixed(0)}%
                          </span>
                        </div>
                      </button>
                    {/each}
                  </div>
                </div>

              <!-- CASE 2: Single /24 Subnet or Drilled-Down /24 Host Grid (1..254) -->
              {:else}
                <div class="space-y-3">
                  <div class="flex items-center justify-between text-[11px] text-slate-400">
                    <span>
                      {$_("ArpReport.SubnetHostHeader", {
                        values: { prefix: currentSubnetHosts.prefix },
                      })}
                    </span>
                    <span class="text-[10px] font-mono">
                      {$_("ArpReport.HostClickHint")}
                    </span>
                  </div>

                  <div
                    class="grid gap-1 max-h-72 overflow-y-auto p-2 rounded-xl bg-slate-950/80 border border-slate-800/80"
                    style="grid-template-columns: repeat(auto-fill, minmax(32px, 1fr));"
                  >
                    {#each Array.from({ length: 254 }, (_, i) => i + 1) as hostNum}
                      {@const host = currentSubnetHosts.hostMap.get(hostNum)}
                      {@const isUsed = !!host}
                      {@const hostIP = `${currentSubnetHosts.prefix}.${hostNum}`}
                      {@const isQueryMatch =
                        searchQuery &&
                        (hostIP.includes(searchQuery) ||
                          (host?.name &&
                            host.name
                              .toLowerCase()
                              .includes(searchQuery.toLowerCase())) ||
                          (host?.mac &&
                            host.mac
                              .toLowerCase()
                              .includes(searchQuery.toLowerCase())) ||
                          (host?.vendor &&
                            host.vendor
                              .toLowerCase()
                              .includes(searchQuery.toLowerCase())))}
                      <button
                        type="button"
                        onclick={() => {
                          selectedHostInfo = isUsed
                            ? {
                                ip: hostIP,
                                name: host.name,
                                mac: host.mac,
                                vendor: host.vendor,
                                isManaged: host.isManaged,
                              }
                            : { ip: hostIP, isFree: true };
                        }}
                        title="{hostIP} {isUsed ? `(${host.name || host.mac})` : `(${$_('ArpReport.LegendFree')})`}"
                        class="h-5 rounded text-[9px] flex items-center justify-center font-mono cursor-pointer transition-transform hover:scale-125 {isQueryMatch
                          ? 'ring-2 ring-amber-400 font-bold'
                          : ''} {isUsed
                          ? 'bg-emerald-500 text-slate-950 font-bold shadow-xs shadow-emerald-500/50'
                          : 'bg-slate-800/60 text-slate-400 hover:bg-slate-700'}"
                      >
                        {hostNum}
                      </button>
                    {/each}
                  </div>

                  <!-- Selected Host Detail Panel -->
                  {#if selectedHostInfo}
                    <div class="rounded-xl border border-cyan-800/60 bg-slate-950/90 p-3.5 shadow-md flex flex-wrap items-center justify-between gap-3 text-xs font-mono">
                      <div class="flex items-center gap-3">
                        <div
                          class="h-3 w-3 rounded-full shrink-0 {selectedHostInfo.isFree
                            ? 'bg-slate-700'
                            : 'bg-emerald-400'}"
                        ></div>
                        <div>
                          <span class="text-cyan-400 font-bold text-sm">
                            {selectedHostInfo.ip}
                          </span>
                          {#if selectedHostInfo.isFree}
                            <span class="ml-2 text-slate-400 font-sans text-xs">
                              {$_("ArpReport.UnassignedFreeIp")}
                            </span>
                          {:else}
                            <span class="ml-2 text-slate-100 font-bold font-sans">
                              {selectedHostInfo.name}
                            </span>
                            <span class="ml-2 text-slate-400 text-[11px]">
                              MAC: {selectedHostInfo.mac || "-"}
                            </span>
                            <span class="ml-2 rounded bg-slate-800 border border-slate-700 px-1.5 py-0.5 text-[10px] text-slate-300 font-sans">
                              {selectedHostInfo.vendor || "-"}
                            </span>
                          {/if}
                        </div>
                      </div>
                      <button
                        type="button"
                        onclick={() => {
                          selectedHostInfo = null;
                        }}
                        class="rounded px-2 py-0.5 text-slate-400 hover:text-white hover:bg-slate-800 text-[11px] font-sans cursor-pointer transition-colors"
                      >
                        {$_("ArpReport.Close")}
                      </button>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          {/if}
        </div>
      </TabItem>

      <!-- Tab 2: IP to MAC Force Graph -->
      <TabItem
        onclick={() => {
          showChart("graphForce");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiGraph} size={1} />
            {$_("ArpReport.IPtoMACForceGraph")}
          </div>
        {/snippet}
        <div id="graphForce"></div>
      </TabItem>

      <!-- Tab 3: IP to MAC Circular Graph -->
      <TabItem
        onclick={() => {
          showChart("graphCircular");
        }}
      >
        {#snippet titleSlot()}
          <div class="flex items-center gap-2">
            <Icon path={icons.mdiCircle} size={1} />
            {$_("ArpReport.IPtoMACCircelGraph")}
          </div>
        {/snippet}
        <div id="graphCircular"></div>
      </TabItem>
    </Tabs>

    <div class="flex justify-end space-x-2 mr-2">
      {#if hasAI}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={() => (showAIReport = true)}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_("ReportAI.AIExplain")}
        </GradientButton>
      {/if}
      <GradientButton type="button" color="teal" onclick={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("ArpReport.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`address_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainAddressReport(ipam || [], activeTab)}
/>

<style>
  #graphForce,
  #graphCircular {
    min-height: 500px;
    width: 98%;
    height: 70vh;
  }
  #ipam {
    min-width: 300px;
    width: 100%;
    height: 200px;
  }
</style>