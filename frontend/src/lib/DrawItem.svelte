<script lang="ts">
  import {
    Modal,
    Label,
    Input,
    Select,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import {
    GetDrawItem,
    UpdateDrawItem,
    GetPollings,
    GetImage,
    SelectFile,
    GetNodes,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { gauge, bar, line, kpi, classicGauge } from "./chart/drawitem";

  export let show: boolean = false;
  export let id: string = "";
  export let posX = 0;
  export let posY = 0;
  let drawItem: any = undefined;
  let image: string = "";
  let nodeID: string = "";
  let pollings: any = [];
  let pollingList: any = [];
  const nodeList: any = [];
  let showHelp = false;
  let alpha = 255;
  let previewDark = true;

  const dispatch = createEventDispatcher();

  const condList = [
    { name: $_("DrawItem.showItemsAllways") || "常に表示", value: 0 },
    { name: $_("DrawItem.showItemsLow") || "マップ状態が軽度以上の時", value: 1 },
    { name: $_("DrawItem.showItemsHigh") || "マップ状態が重度の時", value: 2 },
  ];

  // Recommendations for common variables
  interface VarRecommendation {
    format: string;
    scale: number;
    desc: string;
  }

  const getRecommendation = (varName: string, rawVal?: any): VarRecommendation | null => {
    const v = (varName || "").toLowerCase();
    if (v === "rtt") {
      if (typeof rawVal === "number" && rawVal > 1000) {
        return { format: "%.2f ms", scale: 0.000001, desc: "ms (ナノ秒変換)" };
      }
      return { format: "%.2f ms", scale: 0.001, desc: "ms (ミリ秒変換)" };
    }
    if (v.includes("bps") || v === "traffic" || v === "throughput") {
      return { format: "BPS", scale: 1.0, desc: "自動単位(bps/Mbps/Gbps)" };
    }
    if (v.includes("pps")) {
      return { format: "PPS", scale: 1.0, desc: "自動パケット数(PPS)" };
    }
    if (
      v.includes("cpu") ||
      v.includes("mem") ||
      v.includes("disk") ||
      v.includes("usage") ||
      v.includes("util")
    ) {
      if (typeof rawVal === "number" && rawVal <= 1.0 && rawVal > 0) {
        return { format: "%.1f%%", scale: 100.0, desc: "パーセント (0-1比率)" };
      }
      return { format: "%.1f%%", scale: 1.0, desc: "パーセント (%)" };
    }
    if (v.includes("temp")) {
      return { format: "%.1f ℃", scale: 1.0, desc: "温度 (℃)" };
    }
    if (v.includes("humid")) {
      return { format: "%.1f%%", scale: 1.0, desc: "湿度 (%)" };
    }
    if (v.includes("volt")) {
      return { format: "%.2f V", scale: 1.0, desc: "電圧 (V)" };
    }
    if (v === "loss") {
      return { format: "%.1f%%", scale: 1.0, desc: "パケット損失率 (%)" };
    }
    if (v === "load") {
      return { format: "LOAD=%.2f", scale: 1.0, desc: "システム負荷" };
    }
    if (v === "count") {
      return { format: "COUNT=%.0f", scale: 1.0, desc: "回数" };
    }
    return null;
  };

  const onOpen = async () => {
    pollings = await GetPollings("");
    const nodes = await GetNodes();
    nodeList.length = 0;
    for (const k in nodes) {
      nodeList.push({
        name: nodes[k].Name,
        value: k,
      });
    }
    drawItem = await GetDrawItem(id);
    if (id == "") {
      drawItem.X = posX;
      drawItem.Y = posY;
      alpha = 255;
      if (drawItem.Type === 11) {
        if (!drawItem.W) drawItem.W = 220;
        if (!drawItem.H) drawItem.H = 84;
      }
    } else {
      if (drawItem.Color.length === 9) {
        alpha = parseInt(drawItem.Color.substring(7), 16);
        drawItem.Color = drawItem.Color.substring(0, 7);
      } else {
        alpha = 255;
      }
      if (drawItem.Text && drawItem.Text.includes("\t")) {
        drawItem.Text = drawItem.Text.split("\t")[0];
      }
      if (drawItem.PollingID) {
        nodeID = "";
        for (const p of pollings) {
          if (p.ID == drawItem.PollingID) {
            nodeID = p.NodeID;
            updatePollingList();
            break;
          }
        }
      }
    }
    if (drawItem.Path) {
      image = await GetImage(drawItem.Path);
    }
  };

  const updatePollingList = () => {
    pollingList = [];
    for (let p of pollings) {
      if (nodeID == p.NodeID) {
        pollingList.push({
          name: p.Name,
          value: p.ID,
        });
      }
    }
  };

  const selectVariable = (key: string, val: any) => {
    if (!drawItem) return;
    drawItem.VarName = key;
    const rec = getRecommendation(key, val);
    if (rec) {
      drawItem.Format = rec.format;
      drawItem.Scale = rec.scale;
    }
  };

  interface SizePreset {
    label: string;
    w?: number;
    h?: number;
    size?: number;
  }

  $: currentPresets = ((): SizePreset[] => {
    if (!drawItem) return [];
    switch (drawItem.Type) {
      case 11: // KPI Card
        return [
          { label: `${$_("DrawItem.PresetCompact") || "コンパクト"} (180×70)`, w: 180, h: 70 },
          { label: `${$_("DrawItem.PresetNormal") || "標準"} (220×84)`, w: 220, h: 84 },
          { label: `${$_("DrawItem.PresetLarge") || "大"} (280×100)`, w: 280, h: 100 },
          { label: `${$_("DrawItem.PresetWide") || "ワイド"} (340×90)`, w: 340, h: 90 },
        ];
      case 6: // New Gauge
        return [
          { label: `${$_("DrawItem.PresetSmall") || "小"} (64px)`, h: 64, w: 64 },
          { label: `${$_("DrawItem.PresetNormal") || "標準"} (120px)`, h: 120, w: 120 },
          { label: `${$_("DrawItem.PresetLarge") || "大"} (180px)`, h: 180, w: 180 },
          { label: "特大 (240px)", h: 240, w: 240 },
        ];
      case 7: // Bar
      case 8: // Line
        return [
          { label: `${$_("DrawItem.PresetSmall") || "小"} (200×50)`, h: 50, w: 200 },
          { label: `${$_("DrawItem.PresetNormal") || "標準"} (320×80)`, h: 80, w: 320 },
          { label: `${$_("DrawItem.PresetLarge") || "大"} (440×110)`, h: 110, w: 440 },
        ];
      case 5: // Polling Gauge (Classic)
        return [
          { label: "12 (120px)", size: 12 },
          { label: "16 (160px)", size: 16 },
          { label: "20 (200px)", size: 20 },
          { label: "24 (240px)", size: 24 },
        ];
      case 2: // Label
      case 4: // Polling Text
        return [
          { label: "12px", size: 12 },
          { label: "16px", size: 16 },
          { label: "20px", size: 20 },
          { label: "24px", size: 24 },
          { label: "32px", size: 32 },
        ];
      default: // Rect, Ellipse, Image, GroupFrame, GroupFill
        return [
          { label: `${$_("DrawItem.PresetSmall") || "小"} (160×100)`, w: 160, h: 100 },
          { label: `${$_("DrawItem.PresetNormal") || "中"} (300×180)`, w: 300, h: 180 },
          { label: `${$_("DrawItem.PresetLarge") || "大"} (500×300)`, w: 500, h: 300 },
        ];
    }
  })();

  const isPresetActive = (preset: SizePreset): boolean => {
    if (!drawItem) return false;
    if (preset.size !== undefined) {
      return Number(drawItem.Size) === preset.size;
    }
    if (preset.w !== undefined && preset.h !== undefined) {
      return Number(drawItem.W) === preset.w && Number(drawItem.H) === preset.h;
    }
    if (preset.h !== undefined) {
      return Number(drawItem.H) === preset.h;
    }
    return false;
  };

  const applyPreset = (preset: SizePreset) => {
    if (!drawItem) return;
    if (preset.w !== undefined) drawItem.W = preset.w;
    if (preset.h !== undefined) drawItem.H = preset.h;
    if (preset.size !== undefined) drawItem.Size = preset.size;
  };

  const close = () => {
    show = false;
  };

  const save = async () => {
    drawItem.W *= 1;
    drawItem.H *= 1;
    drawItem.Size *= 1;
    drawItem.Scale *= 1;
    drawItem.X *= 1;
    drawItem.Y *= 1;
    const a = alpha.toString(16).padStart(2, "0");
    drawItem.Color = drawItem.Color.substring(0, 7) + a;
    const r = await UpdateDrawItem(drawItem);
    if (r) {
      close();
    }
  };

  const selectImage = async () => {
    const p = await SelectFile($_("DrawItem.ImageFile") || "画像ファイル選択", true);
    if (p) {
      drawItem.Path = p;
      image = await GetImage(p);
    }
  };

  // Selected polling and available variables
  $: selectedPolling = pollings?.find((p: any) => p.ID === drawItem?.PollingID);
  $: availableVars = selectedPolling?.Result
    ? Object.entries(selectedPolling.Result).map(([k, v]) => ({
        key: k,
        val: v,
        rec: getRecommendation(k, v),
      }))
    : [];

  // Default title based on polling/node
  $: defaultTitle = selectedPolling
    ? drawItem?.VarName
      ? `${selectedPolling.Name} (${drawItem.VarName})`
      : selectedPolling.Name
    : "METRIC";

  // Real-time preview calculation
  $: previewValue = (() => {
    if (!drawItem) return 0;
    if (selectedPolling && drawItem.VarName && selectedPolling.Result?.[drawItem.VarName] !== undefined) {
      const v = Number(selectedPolling.Result[drawItem.VarName]);
      if (!isNaN(v)) return v * (drawItem.Scale || 1.0);
    }
    // Fallback sample values
    if (drawItem.Type === 11) return 12.4;
    if (drawItem.Type === 6 || drawItem.Type === 5 || drawItem.Type === 7) return 68.5;
    return 25.0;
  })();

  $: previewFormattedText = (() => {
    if (!drawItem) return "";
    const val = previewValue;
    if (drawItem.Format) {
      if (drawItem.Format === "BPS") {
        if (val < 1000) return `${val.toFixed(1)} bps`;
        if (val < 1000000) return `${(val / 1000).toFixed(1)} Kbps`;
        if (val < 1000000000) return `${(val / 1000000).toFixed(1)} Mbps`;
        return `${(val / 1000000000).toFixed(1)} Gbps`;
      }
      if (drawItem.Format === "PPS") {
        return `${val.toLocaleString()} PPS`;
      }
      try {
        if (drawItem.Format.includes("%")) {
          // Simple format approximation for preview
          return drawItem.Format.replace(/%[0-9.]*f/, val.toFixed(1)).replace(/%s/, String(val));
        }
      } catch {}
    }
    if (drawItem.Type === 6 || drawItem.Type === 7 || drawItem.Type === 5) {
      return `${val.toFixed(1)}%`;
    }
    return `${val.toFixed(1)}`;
  })();

  $: previewDataUrl = (() => {
    if (!drawItem) return "";
    const color = drawItem.Color || "#00d2ff";
    const title = drawItem.Text || defaultTitle;
    const bg = previewDark ? "#171923" : "#f8fafc";
    const sampleHistory = [18, 24, 32, 28, 45, 40, 55, 62, 58, previewValue];

    try {
      switch (drawItem.Type) {
        case 5: // Classic Polling Gauge
          return classicGauge(title, color, previewValue, drawItem.Size || 16, previewDark);
        case 6: // New Gauge
          return gauge(title, previewValue, bg);
        case 7: // Bar
          return bar(title, color, previewValue, bg);
        case 8: // Line
          return line(title, color, sampleHistory, bg);
        case 11: // KPI Card
          return kpi(
            title,
            previewFormattedText,
            previewValue,
            color,
            sampleHistory,
            previewDark,
            drawItem.W || 220,
            drawItem.H || 84
          );
        default:
          return "";
      }
    } catch {
      return "";
    }
  })();

  let prevShow = false;
  $: {
    if (show && !prevShow) {
      onOpen();
    } else if (!show && prevShow) {
      dispatch("close", {});
    }
    prevShow = show;
  }
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  outsideclose={false}
  class="w-full"
>
  {#if !drawItem}
    <div class="text-center mt-10"><Spinner size="16" /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <div class="flex items-center justify-between border-b pb-2 dark:border-gray-700">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
          <Icon path={icons.mdiPaletteOutline} size={1} class="text-blue-500" />
          {$_("DrawItem.EditDrawItem") || "描画アイテムの編集"}
        </h3>
      </div>

      <!-- Main 2-Pane Layout -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        <!-- Left Pane: Settings Form (7 Cols) -->
        <div class="lg:col-span-7 space-y-4">
          <!-- Type Selection (Categorized & Compact) -->
          <div class="space-y-1">
            <Label class="text-xs font-semibold text-gray-700 dark:text-gray-300">
              {$_("DrawItem.Type") || "種類"}
              {#if id !== ""}
                <span class="text-xs text-gray-400 ml-1 font-normal">(変更不可)</span>
              {/if}
            </Label>
            <select
              class="w-full bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 p-2 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white disabled:opacity-60"
              bind:value={drawItem.Type}
              disabled={id !== ""}
            >
              <optgroup label="📊 {$_('DrawItem.CategoryPolling') || 'ポーリング測定値'}">
                <option value={11}>💎 {$_("DrawItem.KPI") || "ポーリング結果(KPIカード)"}</option>
                <option value={6}>🎯 {$_("DrawItem.NewGauge") || "ポーリング結果(新ゲージ)"}</option>
                <option value={7}>📊 {$_("DrawItem.Bar") || "ポーリング結果(バー)"}</option>
                <option value={8}>📈 {$_("DrawItem.Line") || "ポーリング結果(ライン)"}</option>
                <option value={4}>📝 {$_("DrawItem.PollingText") || "ポーリング結果(テキスト)"}</option>
                <option value={5}>⏱️ {$_("DrawItem.PollingGauge") || "ポーリング結果(ゲージ)"}</option>
              </optgroup>
              <optgroup label="📐 {$_('DrawItem.CategoryShape') || '基本図形・装飾'}">
                <option value={2}>🏷️ {$_("DrawItem.Label") || "ラベル"}</option>
                <option value={3}>🖼️ {$_("DrawItem.Image") || "イメージ"}</option>
                <option value={0}>⬜ {$_("DrawItem.Rect") || "矩形"}</option>
                <option value={1}>⚪ {$_("DrawItem.Ellipse") || "楕円"}</option>
              </optgroup>
              <optgroup label="🔲 {$_('DrawItem.CategoryGroup') || 'グループ化'}">
                <option value={9}>▢ {$_("DrawItem.GroupFrame") || "グループ(枠)"}</option>
                <option value={10}>⬛ {$_("DrawItem.GroupFill") || "グループ(背景)"}</option>
              </optgroup>
            </select>
          </div>

          <!-- Size Presets & Dimension Inputs -->
          <div class="p-3 bg-gray-50 dark:bg-gray-800/60 rounded-lg border border-gray-200 dark:border-gray-700 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
                {$_("DrawItem.Preset") || "推奨サイズ"}
              </span>
            </div>

            <!-- Size Preset Quick Buttons -->
            <div class="flex flex-wrap gap-1.5">
              {#each currentPresets as p}
                {@const active = isPresetActive(p)}
                <button
                  type="button"
                  class="px-2.5 py-1 text-xs rounded transition-all flex items-center gap-1.5 {active
                    ? 'bg-blue-600 text-white font-bold border border-blue-600 shadow-sm ring-2 ring-blue-300 dark:ring-blue-800'
                    : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:border-blue-400 hover:text-blue-600 dark:hover:text-blue-300 font-medium'}"
                  onclick={() => applyPreset(p)}
                >
                  {#if active}
                    <span class="text-white text-[11px] font-bold">✓</span>
                  {/if}
                  <span>{p.label}</span>
                </button>
              {/each}
            </div>

            <!-- Manual Fine-tuning Inputs -->
            <div class="grid grid-cols-2 gap-3 pt-1">
              {#if drawItem.Type === 11 || drawItem.Type < 2 || drawItem.Type === 3 || drawItem.Type === 9 || drawItem.Type === 10}
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Width") || "幅"} (px)</span>
                  <Input
                    class="h-8 text-right"
                    type="number"
                    min={0}
                    max={2000}
                    bind:value={drawItem.W}
                    size="sm"
                  />
                </Label>
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Height") || "高さ"} (px)</span>
                  <Input
                    class="h-8 text-right"
                    type="number"
                    min={0}
                    max={2000}
                    bind:value={drawItem.H}
                    size="sm"
                  />
                </Label>
              {:else if drawItem.Type === 6 || drawItem.Type === 7 || drawItem.Type === 8}
                <Label class="space-y-1 text-xs col-span-2">
                  <span>{$_("DrawItem.Height") || "サイズ/高さ"} (px)</span>
                  <Input
                    class="h-8 text-right"
                    type="number"
                    min={10}
                    max={1000}
                    bind:value={drawItem.H}
                    size="sm"
                  />
                </Label>
              {:else}
                <Label class="space-y-1 text-xs col-span-2">
                  <span>
                    {#if drawItem.Type === 5}
                      {$_("DrawItem.GaugeSize") || "サイズ (直径 = 設定値 × 10 px)"}
                    {:else}
                      {$_("DrawItem.FontSize") || "文字サイズ"} (px)
                    {/if}
                  </span>
                  <Input
                    class="h-8 text-right"
                    type="number"
                    min={8}
                    max={128}
                    bind:value={drawItem.Size}
                    size="sm"
                  />
                </Label>
              {/if}
            </div>
          </div>

          <!-- Polling Selection & Variable Assistant -->
          {#if (drawItem.Type >= 4 && drawItem.Type < 9) || drawItem.Type === 11}
            <div class="p-3 bg-gray-50 dark:bg-gray-800/60 rounded-lg border border-gray-200 dark:border-gray-700 space-y-3">
              <div class="grid grid-cols-2 gap-3">
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Node") || "ノード"}</span>
                  <Select
                    items={nodeList}
                    bind:value={nodeID}
                    placeholder={$_("DrawItem.SelectNode") || "ノードを選択"}
                    size="sm"
                    onchange={updatePollingList}
                  />
                </Label>
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Polling") || "ポーリング"}</span>
                  <Select
                    items={pollingList}
                    bind:value={drawItem.PollingID}
                    placeholder={$_("DrawItem.SelectPolling") || "ポーリングを選択"}
                    size="sm"
                  />
                </Label>
              </div>

              <!-- Available Variable Chips (Click to auto-apply format & scale) -->
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
                    {$_("DrawItem.AvailableVars") || "利用可能な変数（クリックで書式・倍率も自動設定）"}
                  </span>
                </div>
                {#if availableVars.length > 0}
                  <div class="flex flex-wrap gap-1.5 pt-1">
                    {#each availableVars as item}
                      <button
                        type="button"
                        class="px-2.5 py-1 text-xs rounded-full border transition-all flex items-center gap-1.5 {drawItem.VarName === item.key ? 'bg-blue-600 text-white border-blue-600 shadow-sm' : 'bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 border-gray-300 dark:border-gray-600 hover:border-blue-400'}"
                        onclick={() => selectVariable(item.key, item.val)}
                        title={item.rec ? `推奨: ${item.rec.format} (倍率: ${item.rec.scale})` : "クリックして選択"}
                      >
                        <span class="font-bold">{item.key}</span>
                        {#if item.val !== undefined}
                          <span class="opacity-75 text-[10px]">({item.val})</span>
                        {/if}
                        {#if item.rec}
                          <span class="text-[9px] bg-blue-100 dark:bg-blue-900/60 text-blue-700 dark:text-blue-300 px-1 rounded">
                            {item.rec.desc}
                          </span>
                        {/if}
                      </button>
                    {/each}
                  </div>
                {:else if drawItem.PollingID}
                  <p class="text-xs text-gray-400 italic">
                    ※ ポーリングの最新結果データを待機中、または変数が自動設定されます
                  </p>
                {:else}
                  <p class="text-xs text-gray-400 italic">
                    ※ ノードとポーリングを選択すると、利用可能な測定変数がここに表示されます
                  </p>
                {/if}
              </div>

              <!-- Variable Name, Format, Label, Scale -->
              <div class="grid grid-cols-2 gap-3 pt-1">
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.ValName") || "変数名"}</span>
                  <Input
                    class="h-8"
                    bind:value={drawItem.VarName}
                    placeholder={$_("DrawItem.ValNamePH") || "空欄時は自動設定"}
                    size="sm"
                  />
                </Label>
                {#if drawItem.Type === 4 || drawItem.Type === 11}
                  <Label class="space-y-1 text-xs">
                    <span>{$_("DrawItem.TextFormat") || "表示フォーマット"}</span>
                    <Input
                      class="h-8"
                      bind:value={drawItem.Format}
                      placeholder={$_("DrawItem.TextFormatPH") || "例: %.2f ms / BPS"}
                      size="sm"
                    />
                  </Label>
                {/if}
              </div>

              <div class="grid grid-cols-2 gap-3">
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.GaugeLabel") || "表示ラベル/タイトル"}</span>
                  <Input
                    class="h-8"
                    bind:value={drawItem.Text}
                    placeholder={defaultTitle}
                    size="sm"
                  />
                </Label>
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Zoom") || "倍率/スケール"}</span>
                  <Input
                    class="h-8 text-right"
                    type="number"
                    min={0.000000001}
                    max={1000}
                    step={0.1}
                    bind:value={drawItem.Scale}
                    size="sm"
                  />
                </Label>
              </div>
            </div>
          {/if}

          <!-- Colors & Image Selection for Shapes -->
          {#if drawItem.Type < 4 || drawItem.Type >= 9}
            <div class="p-3 bg-gray-50 dark:bg-gray-800/60 rounded-lg border border-gray-200 dark:border-gray-700 space-y-3">
              {#if drawItem.Type === 3}
                <!-- Image Selection -->
                <div class="flex items-center gap-3">
                  <GradientButton
                    shadow
                    class="h-8"
                    type="button"
                    size="xs"
                    color="blue"
                    onclick={selectImage}
                  >
                    <Icon path={icons.mdiImage} size={1} class="mr-1" />
                    {$_("DrawItem.Select") || "画像ファイルを選択"}
                  </GradientButton>
                  <span class="text-xs text-gray-500 truncate max-w-xs">{drawItem.Path || "未選択"}</span>
                </div>
              {:else}
                <div class="grid grid-cols-2 gap-3 items-center">
                  <Label class="space-y-1 text-xs">
                    <div>{$_("DrawItem.Color") || "色・不透明度"}</div>
                    <div class="flex items-center space-x-2">
                      <input type="color" class="h-8 w-12 rounded cursor-pointer border" bind:value={drawItem.Color} />
                      <input type="range" min={0} max={255} step={1} bind:value={alpha} class="w-24" />
                      <span class="text-xs text-gray-500">{Math.round((alpha / 255) * 100)}%</span>
                    </div>
                  </Label>
                  <Label class="space-y-1 text-xs">
                    <span>{$_("DrawItem.showCond") || "表示条件"}</span>
                    <Select
                      items={condList}
                      bind:value={drawItem.Cond}
                      placeholder={$_("DrawItem.selectShowCond") || "表示条件を選択"}
                      size="sm"
                    />
                  </Label>
                </div>
              {/if}

              {#if drawItem.Type === 2 || drawItem.Type === 9 || drawItem.Type === 10}
                <Label class="space-y-1 text-xs">
                  <span>{$_("DrawItem.Text") || "表示文字列"}</span>
                  <Input
                    class="h-8"
                    bind:value={drawItem.Text}
                    placeholder={$_("DrawItem.TextToDisplay") || "表示するテキスト"}
                    size="sm"
                  />
                </Label>
              {/if}
            </div>
          {/if}
        </div>

        <!-- Right Pane: Real-time Live Preview (5 Cols) -->
        <div class="lg:col-span-5 flex flex-col space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
              <Icon path={icons.mdiEyeOutline} size={0.9} class="text-emerald-500" />
              {$_("DrawItem.Preview") || "リアルタイムプレビュー"}
            </span>
            <!-- Theme Toggle for Preview -->
            <div class="inline-flex rounded-md shadow-sm" role="group">
              <button
                type="button"
                class="px-2 py-0.5 text-xs font-medium rounded-l-lg border {previewDark ? 'bg-gray-800 text-cyan-400 border-gray-600' : 'bg-white text-gray-600 border-gray-300'}"
                onclick={() => (previewDark = true)}
              >
                {$_("DrawItem.PreviewDark") || "Dark"}
              </button>
              <button
                type="button"
                class="px-2 py-0.5 text-xs font-medium rounded-r-lg border {previewDark ? 'bg-gray-700 text-gray-400 border-gray-600' : 'bg-blue-50 text-blue-600 border-blue-300'}"
                onclick={() => (previewDark = false)}
              >
                {$_("DrawItem.PreviewLight") || "Light"}
              </button>
            </div>
          </div>

          <!-- Preview Stage Container -->
          <div
            class="flex-1 min-h-[260px] rounded-xl border p-4 flex flex-col items-center justify-center relative overflow-hidden transition-colors {previewDark ? 'bg-[#12141c] border-gray-700' : 'bg-[#f1f5f9] border-gray-300'}"
          >
            <!-- Grid Background Pattern -->
            <div
              class="absolute inset-0 pointer-events-none opacity-20"
              style="background-image: radial-gradient({previewDark ? '#ffffff' : '#000000'} 1px, transparent 1px); background-size: 16px 16px;"
            ></div>

            <!-- Content Preview -->
            <div class="relative z-10 flex items-center justify-center max-w-full max-h-full overflow-hidden p-2">
              {#if previewDataUrl}
                <img
                  src={previewDataUrl}
                  alt="Preview"
                  class="max-w-full max-h-[220px] object-contain drop-shadow-md"
                />
              {:else if drawItem.Type === 0}
                <!-- Rect Preview -->
                <div
                  class="border"
                  style="width: {Math.min(260, Math.max(40, drawItem.W || 160))}px; height: {Math.min(180, Math.max(30, drawItem.H || 100))}px; background-color: {drawItem.Color}; border-radius: 6px; border-color: {previewDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.2)'};"
                ></div>
              {:else if drawItem.Type === 1}
                <!-- Ellipse Preview -->
                <div
                  class="border"
                  style="width: {Math.min(260, Math.max(40, drawItem.W || 160))}px; height: {Math.min(180, Math.max(30, drawItem.H || 100))}px; background-color: {drawItem.Color}; border-radius: 50%; border-color: {previewDark ? 'rgba(255,255,255,0.2)' : 'rgba(0,0,0,0.2)'};"
                ></div>
              {:else if drawItem.Type === 2}
                <!-- Label Preview -->
                <span
                  style="font-size: {Math.min(36, Math.max(10, drawItem.Size || 16))}px; color: {drawItem.Color || (previewDark ? '#f3f4f6' : '#1e293b')}; font-weight: 600;"
                >
                  {drawItem.Text || "Sample Label"}
                </span>
              {:else if drawItem.Type === 3}
                <!-- Image Preview -->
                {#if image}
                  <img
                    src={image}
                    alt="Preview"
                    class="max-w-full max-h-[180px] object-contain rounded"
                  />
                {:else}
                  <div class="text-xs text-gray-400 flex flex-col items-center gap-1">
                    <Icon path={icons.mdiImageOutline} size={2} />
                    <span>画像が未選択です</span>
                  </div>
                {/if}
              {:else if drawItem.Type === 4}
                <!-- Polling Text Preview -->
                <div class="px-3 py-1.5 rounded bg-black/40 backdrop-blur-sm border border-white/10 flex items-center gap-2">
                  <div class="w-2 h-2 rounded-full bg-emerald-400"></div>
                  <span
                    style="font-size: {Math.min(28, Math.max(10, drawItem.Size || 14))}px; color: {drawItem.Color || '#00d2ff'}; font-family: monospace;"
                  >
                    {previewFormattedText || "12.4 Mbps"}
                  </span>
                </div>
              {:else if drawItem.Type === 9 || drawItem.Type === 10}
                <!-- Group Frame / Fill Preview -->
                <div
                  class="relative p-2"
                  style="width: {Math.min(260, Math.max(60, drawItem.W || 200))}px; height: {Math.min(180, Math.max(40, drawItem.H || 120))}px; border-radius: 8px; border: {drawItem.Type === 9 ? `2px solid ${drawItem.Color}` : 'none'}; background-color: {drawItem.Type === 10 ? drawItem.Color : 'rgba(23,23,23,0.05)'};"
                >
                  <span
                    class="absolute bottom-1 right-2 text-xs font-bold"
                    style="color: {previewDark ? '#eee' : '#333'}; font-size: {drawItem.Size || 11}px;"
                  >
                    {drawItem.Text || "Group Title"}
                  </span>
                </div>
              {/if}
            </div>

            <!-- Size / Dimension Badge -->
            <div class="absolute bottom-2 left-3 text-[11px] text-gray-400 bg-black/30 backdrop-blur-sm px-2 py-0.5 rounded">
              {#if drawItem.Type === 2 || drawItem.Type === 4}
                サイズ: {drawItem.Size || 16} px
              {:else if drawItem.Type === 5}
                サイズ: {(drawItem.Size || 16) * 10} × {(drawItem.Size || 16) * 10} px (設定値: {drawItem.Size || 16})
              {:else if drawItem.Type === 6}
                サイズ: {drawItem.H || 120} × {drawItem.H || 120} px
              {:else if drawItem.Type === 7 || drawItem.Type === 8}
                サイズ: {(drawItem.H || 80) * 4} × {drawItem.H || 80} px
              {:else}
                サイズ: {drawItem.W || 220} × {drawItem.H || 84} px
              {/if}
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex justify-end space-x-2 pt-2 border-t dark:border-gray-700">
        <GradientButton
          shadow
          color="blue"
          type="button"
          onclick={save}
          size="xs"
        >
          <Icon path={icons.mdiContentSave} size={1} class="mr-1" />
          {$_("DrawItem.Save") || "保存"}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          size="xs"
          color="lime"
          onclick={() => {
            showHelp = true;
          }}
        >
          <Icon path={icons.mdiHelp} size={1} class="mr-1" />
          {$_("DrawItem.Help") || "ヘルプ"}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} class="mr-1" />
          {$_("DrawItem.Cancel") || "キャンセル"}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>

<Help bind:show={showHelp} page="drawitem" />

