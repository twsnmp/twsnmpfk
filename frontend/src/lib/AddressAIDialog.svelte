<script lang="ts">
  import { marked } from "marked";
  import { Modal, GradientButton, Alert } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import DOMPurify from "dompurify";
  import { _ } from "svelte-i18n";
  import { onDestroy } from "svelte";
  import { LLMExplainAddress, ExportMarkdown } from "../../wailsjs/go/main/App";
  import { ClipboardSetText } from "../../wailsjs/runtime";

  import neko1 from "../assets/images/neko_anm1.png";
  import neko2 from "../assets/images/neko_anm2.png";
  import neko3 from "../assets/images/neko_anm3.png";
  import neko4 from "../assets/images/neko_anm4.png";
  import neko5 from "../assets/images/neko_anm5.png";
  import neko6 from "../assets/images/neko_anm6.png";
  import neko7 from "../assets/images/neko_anm7.png";

  export let show = false;

  let loading = false;
  let content = "";
  let error = "";
  let copied = false;
  let hasAnalyzed = false;

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

  const analyze = async () => {
    loading = true;
    startAnimation();
    error = "";
    content = "";
    copied = false;
    try {
      const resp = await LLMExplainAddress();
      if (resp.Error) {
        error = resp.Error;
      } else {
        content = resp.Results;
      }
    } catch (e: any) {
      error = e?.message || String(e);
    } finally {
      loading = false;
      stopAnimation();
      hasAnalyzed = true;
    }
  };

  const copyReport = async () => {
    if (!content) return;
    try {
      await ClipboardSetText(content);
    } catch {
      await navigator.clipboard.writeText(content);
    }
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  };

  const exportReport = async () => {
    if (!content) return;
    try {
      const filename = "address_ai_explanation";
      await ExportMarkdown(filename, content);
    } catch (e: any) {
      error = e?.message || String(e);
    }
  };

  $: if (show && !hasAnalyzed && !loading) {
    analyze();
  }

  $: if (!show) {
    hasAnalyzed = false;
  }

  $: renderedContent = DOMPurify.sanitize(marked.parse(content || "") as string);
</script>

<Modal bind:open={show} size="lg" dismissable={false} class="w-full">
  <div class="flex flex-col max-h-[70vh]">
    <div class="flex items-center space-x-2 border-b pb-2 mb-3 dark:border-gray-700">
      <Icon path={icons.mdiBrain} size={1.2} class="text-purple-600 dark:text-purple-400" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {$_("AddressAIDialog.Title")}
      </h3>
    </div>

    {#if loading}
      <div class="flex flex-col items-center justify-center p-6 space-y-3 min-h-[180px]">
        <div class="bg-white p-3 rounded-xl border border-gray-200 shadow-sm flex items-center justify-center">
          <img src={nekos[nekoNo]} alt="loading animation" class="h-14 w-auto object-contain" />
        </div>
        <p class="text-xs font-medium text-gray-600 dark:text-gray-300">
          {$_("AddressAIDialog.Loading")}
        </p>
      </div>
    {:else}
      {#if error}
        <Alert color="red" dismissable class="mb-3">
          <div class="flex items-center space-x-2">
            <Icon path={icons.mdiExclamation} size={1} />
            <span>{error}</span>
          </div>
        </Alert>
      {/if}

      <div
        class="overflow-y-auto p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 min-h-[200px]"
      >
        {#if content}
          <article
            class="prose prose-sm md:prose-base dark:prose-invert max-w-none"
          >
            {@html renderedContent}
          </article>
        {:else if !error}
          <p class="text-gray-500 text-sm">No analysis content.</p>
        {/if}
      </div>
    {/if}

    <div class="flex justify-end space-x-2 mt-4">
      {#if content && !loading}
        <GradientButton
          shadow
          type="button"
          color="blue"
          onclick={copyReport}
          size="xs"
        >
          <Icon path={copied ? icons.mdiCheck : icons.mdiContentCopy} size={1} />
          {copied ? $_("AddressAIDialog.Copied") : $_("AddressAIDialog.Copy")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="lime"
          onclick={exportReport}
          size="xs"
        >
          <Icon path={icons.mdiDownload} size={1} />
          {$_("AddressAIDialog.Export")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="pink"
        disabled={loading}
        onclick={analyze}
        size="xs"
      >
        <Icon path={icons.mdiRefresh} size={1} />
        {$_("AddressAIDialog.ReAnalyze")}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => (show = false)}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("AddressAIDialog.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>
