<script lang="ts">
  import { Modal, GradientButton, Label, Input } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import { Icon } from "mdi-svelte-ts";
  import { _ } from "svelte-i18n";
  import {
    GetBackImage,
    SetBackImage,
    GetImage,
    SelectFile,
    GetMapConf,
  } from "../../wailsjs/go/main/App";
  import type { datastore } from "../../wailsjs/go/models";

  export let show: boolean = false;
  export let onSave: (() => void) | undefined = undefined;

  let backImage: datastore.BackImageEnt = {
    X: 0,
    Y: 0,
    Width: 100,
    Height: 100,
    Path: "",
  };
  let imageData: string = "";
  let imgNaturalW = 0;
  let imgNaturalH = 0;
  let lockRatio = true;

  let mapSizeX = 2500;
  let mapSizeY = 5000;

  const updateImageNaturalSize = (src: string, forceReset = false) => {
    if (!src) {
      imgNaturalW = 0;
      imgNaturalH = 0;
      return;
    }
    const img = new Image();
    img.onload = () => {
      imgNaturalW = img.naturalWidth;
      imgNaturalH = img.naturalHeight;
      if (
        forceReset ||
        !backImage.Width ||
        !backImage.Height ||
        backImage.Width <= 0 ||
        backImage.Height <= 0 ||
        (backImage.Width === 100 && backImage.Height === 100)
      ) {
        backImage.Width = img.naturalWidth;
        backImage.Height = img.naturalHeight;
      }
    };
    img.src = src;
  };

  const loadData = async () => {
    try {
      const conf = await GetMapConf();
      switch (conf.MapSize) {
        case 1:
          mapSizeX = 2894;
          mapSizeY = 4093;
          break;
        case 2:
          mapSizeX = 4093;
          mapSizeY = 2894;
          break;
        default:
          mapSizeX =
            typeof window !== "undefined" && window.screen.width > 4000
              ? 5000
              : 2500;
          mapSizeY = 5000;
      }

      backImage = await GetBackImage();
      if (backImage.Path) {
        imageData = await GetImage(backImage.Path);
        updateImageNaturalSize(imageData);
      } else {
        imageData = "";
        imgNaturalW = 0;
        imgNaturalH = 0;
      }
    } catch (e) {
      console.error(e);
    }
  };

  $: if (show) {
    loadData();
  }

  const handleWidthChange = (e: Event) => {
    const val = parseInt((e.target as HTMLInputElement).value, 10) || 0;
    backImage.Width = val;
    if (lockRatio && imgNaturalW > 0 && imgNaturalH > 0 && val > 0) {
      backImage.Height = Math.round((val * imgNaturalH) / imgNaturalW);
    }
  };

  const handleHeightChange = (e: Event) => {
    const val = parseInt((e.target as HTMLInputElement).value, 10) || 0;
    backImage.Height = val;
    if (lockRatio && imgNaturalW > 0 && imgNaturalH > 0 && val > 0) {
      backImage.Width = Math.round((val * imgNaturalW) / imgNaturalH);
    }
  };

  const resetToOriginalSize = () => {
    if (imgNaturalW > 0 && imgNaturalH > 0) {
      backImage.Width = imgNaturalW;
      backImage.Height = imgNaturalH;
    }
  };

  const setScaleMultiplier = (mult: number) => {
    if (imgNaturalW > 0 && imgNaturalH > 0) {
      backImage.Width = Math.round(imgNaturalW * mult);
      backImage.Height = Math.round(imgNaturalH * mult);
    }
  };

  const fitToMapWidth = () => {
    if (mapSizeX > 0) {
      backImage.Width = mapSizeX;
      if (imgNaturalW > 0 && imgNaturalH > 0) {
        backImage.Height = Math.round((mapSizeX * imgNaturalH) / imgNaturalW);
      }
    }
  };

  const centerOnMap = () => {
    backImage.X = Math.max(0, Math.round((mapSizeX - backImage.Width) / 2));
    backImage.Y = Math.max(0, Math.round((mapSizeY - backImage.Height) / 2));
  };

  const alignTopLeft = () => {
    backImage.X = 0;
    backImage.Y = 0;
  };

  const selectImage = async () => {
    const p = await SelectFile($_("Map.BackImage"), true);
    if (p) {
      backImage.Path = p;
      imageData = await GetImage(p);
      updateImageNaturalSize(imageData, true);
    }
  };

  const saveBackImage = async () => {
    show = false;
    backImage.Height = Number(backImage.Height) || 0;
    backImage.Width = Number(backImage.Width) || 0;
    backImage.X = Number(backImage.X) || 0;
    backImage.Y = Number(backImage.Y) || 0;
    await SetBackImage(backImage);
    if (onSave) onSave();
  };

  const clearBackImage = async () => {
    show = false;
    backImage.Path = "";
    imageData = "";
    imgNaturalW = 0;
    imgNaturalH = 0;
    backImage.Height = 0;
    backImage.Width = 0;
    backImage.X = 0;
    backImage.Y = 0;
    await SetBackImage(backImage);
    if (onSave) onSave();
  };

  $: previewScale = Math.min(
    360 / (mapSizeX || 2500),
    160 / (mapSizeY || 5000)
  );
  $: previewCanvasW = Math.max(
    40,
    Math.round((mapSizeX || 2500) * previewScale)
  );
  $: previewCanvasH = Math.max(
    40,
    Math.round((mapSizeY || 5000) * previewScale)
  );
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full"
>
  <div class="flex flex-col space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-700 pb-3">
      <div class="flex items-center gap-2">
        <Icon path={icons.mdiImage} size={0.9} class="text-sky-500" />
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">
          {$_("Map.BackImage")}
        </h3>
        {#if imgNaturalW > 0}
          <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">
            ({$_("Map.OriginalSize")}: {imgNaturalW} × {imgNaturalH} px)
          </span>
        {/if}
      </div>
      {#if backImage.Path}
        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-sky-100 text-sky-800 dark:bg-sky-900/60 dark:text-sky-300 border border-sky-300 dark:border-sky-700">
          {$_("Map.BackImageConfigured")}
        </span>
      {/if}
    </div>

    <!-- Position and Dimensions -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <Label class="space-y-1 text-xs">
        <span class="text-slate-600 dark:text-slate-300 font-medium">X (px)</span>
        <Input
          class="h-8 text-right text-xs font-mono"
          type="number"
          min={0}
          max={10000}
          bind:value={backImage.X}
          size="sm"
        />
      </Label>
      <Label class="space-y-1 text-xs">
        <span class="text-slate-600 dark:text-slate-300 font-medium">Y (px)</span>
        <Input
          class="h-8 text-right text-xs font-mono"
          type="number"
          min={0}
          max={10000}
          bind:value={backImage.Y}
          size="sm"
        />
      </Label>
      <Label class="space-y-1 text-xs">
        <div class="flex items-center justify-between">
          <span class="text-slate-600 dark:text-slate-300 font-medium">{$_("DrawItem.Width")} (px)</span>
          <button
            type="button"
            onclick={() => (lockRatio = !lockRatio)}
            title={lockRatio ? $_("Map.UnlockRatio") : $_("Map.LockRatio")}
            class="p-0.5 rounded hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors {lockRatio ? 'text-sky-500' : 'text-slate-400'}"
          >
            <Icon path={lockRatio ? icons.mdiLink : icons.mdiLinkOff} size={0.7} />
          </button>
        </div>
        <Input
          class="h-8 text-right text-xs font-mono"
          type="number"
          min={1}
          max={10000}
          value={backImage.Width}
          oninput={handleWidthChange}
          size="sm"
        />
      </Label>
      <Label class="space-y-1 text-xs">
        <span class="text-slate-600 dark:text-slate-300 font-medium">{$_("DrawItem.Height")} (px)</span>
        <Input
          class="h-8 text-right text-xs font-mono"
          type="number"
          min={1}
          max={10000}
          value={backImage.Height}
          oninput={handleHeightChange}
          size="sm"
        />
      </Label>
    </div>

    <!-- Quick Presets -->
    {#if backImage.Path}
      <div class="flex flex-wrap items-center gap-1.5 py-1 text-xs border-t border-b border-slate-100 dark:border-slate-800">
        <span class="text-slate-500 dark:text-slate-400 font-medium mr-1">
          {$_("Map.Presets")}:
        </span>
        <button
          type="button"
          onclick={resetToOriginalSize}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          100% ({$_("Map.OriginalSize")})
        </button>
        <button
          type="button"
          onclick={() => setScaleMultiplier(0.5)}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          50%
        </button>
        <button
          type="button"
          onclick={() => setScaleMultiplier(2.0)}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          200%
        </button>
        <button
          type="button"
          onclick={fitToMapWidth}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          {$_("Map.FitMapWidth")}
        </button>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <button
          type="button"
          onclick={alignTopLeft}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          (0, 0)
        </button>
        <button
          type="button"
          onclick={centerOnMap}
          class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs transition-colors"
        >
          {$_("Map.CenterMap")}
        </button>
      </div>

      <!-- Placement Preview Box -->
      <div class="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/60 p-3 space-y-2">
        <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
          <span class="font-medium flex items-center gap-1.5">
            <Icon path={icons.mdiAspectRatio} size={0.7} class="text-sky-500" />
            {$_("Map.PlacementPreview")}
          </span>
          <span class="font-mono text-[11px]">
            {$_("Map.MapCanvas")} {mapSizeX} × {mapSizeY} px
          </span>
        </div>

        {#if backImage.Width <= 0 || backImage.Height <= 0}
          <div class="text-xs text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800 rounded-lg p-2">
            {$_("Map.BackImageZeroWarning")}
          </div>
        {/if}

        <!-- Miniature Canvas Container -->
        <div class="relative w-full h-44 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-200 dark:bg-slate-950/80 overflow-hidden flex items-center justify-center p-2">
          <!-- Canvas Boundary -->
          <div
            class="relative bg-white dark:bg-slate-900 border-2 border-slate-400 dark:border-slate-600 rounded shadow overflow-hidden transition-all"
            style="width: {previewCanvasW}px; height: {previewCanvasH}px;"
          >
            <!-- Placed Image -->
            <div
              class="absolute border-2 border-sky-500 bg-sky-500/20 overflow-hidden transition-all flex items-center justify-center"
              style="
                left: {((Number(backImage.X) || 0) / mapSizeX) * 100}%;
                top: {((Number(backImage.Y) || 0) / mapSizeY) * 100}%;
                width: {((Number(backImage.Width) || 0) / mapSizeX) * 100}%;
                height: {((Number(backImage.Height) || 0) / mapSizeY) * 100}%;
              "
            >
              {#if imageData}
                <img
                  src={imageData}
                  alt="BackImage"
                  class="w-full h-full object-fill opacity-75 pointer-events-none"
                />
              {/if}
            </div>
          </div>
        </div>
      </div>
    {:else}
      <!-- Empty state when no image is selected -->
      <div class="flex flex-col items-center justify-center p-8 rounded-xl border-2 border-dashed border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/40 text-slate-500 dark:text-slate-400 space-y-2">
        <Icon path={icons.mdiImageOutline} size={2} class="text-slate-400" />
        <p class="text-xs">{$_("Map.BackImageSelect")}</p>
      </div>
    {/if}

    <!-- Footer Action Buttons -->
    <div class="flex items-center justify-between pt-2 border-t border-slate-200 dark:border-slate-700">
      <div class="flex items-center gap-2">
        <GradientButton
          shadow
          type="button"
          size="xs"
          color="blue"
          class="flex items-center gap-1.5"
          onclick={selectImage}
        >
          <Icon path={icons.mdiFolderOpen} size={0.7} />
          {$_("Map.BackImageSelect")}
        </GradientButton>

        {#if backImage.Path}
          <GradientButton
            color="red"
            type="button"
            size="xs"
            class="flex items-center gap-1.5"
            onclick={clearBackImage}
          >
            <Icon path={icons.mdiDelete} size={0.7} />
            {$_("Map.BackImageClear")}
          </GradientButton>
        {/if}
      </div>

      <div class="flex items-center gap-2">
        <GradientButton
          color="blue"
          type="button"
          size="xs"
          class="flex items-center gap-1.5"
          onclick={saveBackImage}
        >
          <Icon path={icons.mdiContentSave} size={0.7} />
          {$_("Map.Save")}
        </GradientButton>
        <GradientButton
          shadow
          color="teal"
          type="button"
          size="xs"
          class="flex items-center gap-1.5"
          onclick={() => (show = false)}
        >
          <Icon path={icons.mdiCancel} size={0.7} />
          {$_("Map.Cancel")}
        </GradientButton>
      </div>
    </div>
  </div>
</Modal>
