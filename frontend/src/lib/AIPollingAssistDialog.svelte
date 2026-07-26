<script lang="ts">
  import { marked } from "marked";
  import { Modal, GradientButton, Alert, Textarea, Badge } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import DOMPurify from "dompurify";
  import { _ } from "svelte-i18n";
  import { onDestroy, createEventDispatcher } from "svelte";
  import { LLMAssistPolling } from "../../wailsjs/go/main/App";

  import neko1 from "../assets/images/neko_anm1.png";
  import neko2 from "../assets/images/neko_anm2.png";
  import neko3 from "../assets/images/neko_anm3.png";
  import neko4 from "../assets/images/neko_anm4.png";
  import neko5 from "../assets/images/neko_anm5.png";
  import neko6 from "../assets/images/neko_anm6.png";
  import neko7 from "../assets/images/neko_anm7.png";

  export let show = false;
  export let nodeID = "";

  const dispatch = createEventDispatcher();

  let loading = false;
  let prompt = "";
  let error = "";
  let result: any = null;

  const nekos = [neko1, neko2, neko3, neko4, neko5, neko6, neko7];
  let nekoNo = 0;
  let timer: any = undefined;

  function startAnimation() {
    if (timer) return;
    timer = setInterval(() => {
      nekoNo = (nekoNo + 1) % nekos.length;
    }, 200);
  }

  function stopAnimation() {
    if (timer) {
      clearInterval(timer);
      timer = undefined;
    }
  }

  onDestroy(stopAnimation);

  $: quickTags = [
    { label: $_("AIPollingAssist.TagHttpLabel"), text: $_("AIPollingAssist.TagHttpText") },
    { label: $_("AIPollingAssist.TagCiscoCpuLabel"), text: $_("AIPollingAssist.TagCiscoCpuText") },
    { label: $_("AIPollingAssist.TagTlsLabel"), text: $_("AIPollingAssist.TagTlsText") },
    { label: $_("AIPollingAssist.TagSyslogLabel"), text: $_("AIPollingAssist.TagSyslogText") },
    { label: $_("AIPollingAssist.TagPingLabel"), text: $_("AIPollingAssist.TagPingText") },
  ];

  const setQuickTag = (tagText: string) => {
    prompt = tagText;
  };

  const analyze = async () => {
    loading = true;
    startAnimation();
    error = "";
    result = null;
    try {
      const resp = await LLMAssistPolling(nodeID, prompt);
      if (resp.Error) {
        error = resp.Error;
      } else {
        result = resp;
      }
    } catch (e: any) {
      error = e?.message || String(e);
    } finally {
      loading = false;
      stopAnimation();
    }
  };

  const applySetting = () => {
    if (!result) return;
    dispatch("apply", { polling: result });
    show = false;
  };

  $: renderedAdvice = DOMPurify.sanitize(marked.parse(result?.Advice || "") as string);
</script>

<Modal bind:open={show} size="lg" dismissable={false} class="w-full">
  <div class="flex flex-col max-h-[80vh]">
    <div class="flex items-center space-x-2 border-b pb-2 mb-3 dark:border-gray-700">
      <Icon path={icons.mdiAutoFix} size={1.2} class="text-pink-600 dark:text-pink-400" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {$_("AIPollingAssist.Title")}
      </h3>
    </div>

    <!-- 目的入力エリア -->
    <div class="mb-4 space-y-2 w-full">
      <label for="aiPurposeInput" class="block text-xs font-semibold text-gray-700 dark:text-gray-300">
        {$_("AIPollingAssist.PurposeLabel")}
      </label>
      <Textarea
        id="aiPurposeInput"
        rows="3"
        bind:value={prompt}
        placeholder={$_("AIPollingAssist.PurposePlaceholder")}
        class="w-full text-xs"
        style="width: 100%; min-height: 90px;"
      />
      <div class="flex flex-wrap gap-1 items-center mt-2">
        <span class="text-[11px] text-gray-500 mr-1">{$_("AIPollingAssist.QuickTags")}:</span>
        {#each quickTags as tag}
          <button
            type="button"
            onclick={() => setQuickTag(tag.text)}
            class="text-[11px] bg-pink-50 hover:bg-pink-100 text-pink-700 dark:bg-gray-700 dark:text-pink-300 px-2 py-0.5 rounded-full border border-pink-200 dark:border-gray-600 transition-colors"
          >
            {tag.label}
          </button>
        {/each}
      </div>
    </div>

    <div class="flex justify-start mb-3">
      <GradientButton
        shadow
        type="button"
        color="pink"
        disabled={loading}
        onclick={analyze}
        size="xs"
      >
        <Icon path={icons.mdiSparkles} size={1} class="mr-1" />
        {loading ? $_("AIPollingAssist.Analyzing") : $_("AIPollingAssist.Analyze")}
      </GradientButton>
    </div>

    <!-- アニメーション / 結果エリア -->
    {#if loading}
      <div class="flex flex-col items-center justify-center p-6 space-y-3 min-h-[160px] bg-gray-50 dark:bg-gray-800 rounded-lg">
        <div class="bg-white p-3 rounded-xl border border-gray-200 shadow-sm flex items-center justify-center">
          <img src={nekos[nekoNo]} alt="loading animation" class="h-12 w-auto object-contain" />
        </div>
        <p class="text-xs font-medium text-pink-600 dark:text-pink-400 animate-pulse">
          {$_("AIPollingAssist.Analyzing")}
        </p>
      </div>
    {:else}
      {#if error}
        <Alert color="red" dismissable class="mb-3">
          <div class="flex items-center space-x-2 text-xs">
            <Icon path={icons.mdiExclamation} size={1} />
            <span>{error}</span>
          </div>
        </Alert>
      {/if}

      <div class="overflow-y-auto p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 space-y-3 min-h-[160px]">
        {#if result}
          <!-- 設定プレビュー概要カード -->
          <div class="bg-white dark:bg-gray-900 p-3 rounded border border-pink-200 dark:border-pink-800 space-y-2">
            <div class="flex justify-between items-center border-b pb-1">
              <span class="font-bold text-sm text-pink-700 dark:text-pink-300">
                {result.Name || $_("AIPollingAssist.DefaultName")}
              </span>
              <div class="space-x-1">
                <Badge color="blue">{result.Type}</Badge>
                {#if result.Mode}
                  <Badge color="indigo">{result.Mode}</Badge>
                {/if}
                <Badge color={result.Level === "high" ? "red" : result.Level === "warn" ? "yellow" : "green"}>
                  {result.Level}
                </Badge>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs text-gray-600 dark:text-gray-300">
              <div><span class="font-semibold text-gray-500">Params:</span> {result.Params || "-"}</div>
              <div><span class="font-semibold text-gray-500">PollInt:</span> {result.PollInt} sec</div>
              {#if result.Filter}
                <div class="col-span-2"><span class="font-semibold text-gray-500">Filter:</span> <code>{result.Filter}</code></div>
              {/if}
              {#if result.Extractor}
                <div class="col-span-2"><span class="font-semibold text-gray-500">Extractor:</span> <code>{result.Extractor}</code></div>
              {/if}
              {#if result.Script}
                <div class="col-span-2"><span class="font-semibold text-gray-500">Script:</span> <code>{result.Script}</code></div>
              {/if}
            </div>
          </div>

          <!-- AI アドバイス -->
          <div>
            <h4 class="text-xs font-bold text-gray-700 dark:text-gray-300 mb-1 flex items-center">
              <Icon path={icons.mdiLightbulbOn} size={0.9} class="text-amber-500 mr-1" />
              {$_("AIPollingAssist.AdviceTitle")}
            </h4>
            <article class="prose prose-sm dark:prose-invert max-w-none text-xs bg-amber-50/50 dark:bg-gray-900/50 p-2.5 rounded border border-amber-200 dark:border-amber-900/50">
              {@html renderedAdvice}
            </article>
          </div>
        {:else if !error}
          <div class="text-gray-400 text-xs text-center py-8">
            {$_("AIPollingAssist.NoAdvice")}
          </div>
        {/if}
      </div>
    {/if}

    <!-- フッターボタン -->
    <div class="flex justify-end space-x-2 mt-4">
      {#if result && !loading}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={applySetting}
          size="xs"
        >
          <Icon path={icons.mdiCheck} size={1} class="mr-1" />
          {$_("AIPollingAssist.ApplySetting")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => (show = false)}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("AIPollingAssist.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>
