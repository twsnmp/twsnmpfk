<script lang="ts">
  import {
    Modal,
    GradientButton,
    Textarea,
    Alert,
    Spinner,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { lang } from "../i18n/i18n";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import Reveal from "reveal.js";
  import Highlight from "reveal.js/plugin/highlight/highlight";
  import Markdown from "reveal.js/plugin/markdown/markdown";
  import "reveal.js/dist/reveal.css";
  import "reveal.js/dist/theme/black.css";
  import "reveal.js/plugin/highlight/monokai.css";
  import { SendFeedback } from "../../wailsjs/go/main/App";

  export let page = "";
  export let show: boolean = false;
  let reveal: any = undefined;
  let helpUrl = "";
  let showFeedback = false;
  let feedback = "";
  let feedbackError = false;
  let sending = false;

  const onOpen = async () => {
    helpUrl = `help/${lang}/${page}.md`;
    await tick();
    reveal = new Reveal({
      plugins: [Markdown, Highlight],
      hash: true,
      center: false,
    });
    reveal.initialize();
  };

  const close = () => {
    if (reveal) {
      reveal.destroy();
      reveal = undefined;
    }
    show = false;
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh] bg-gray-800 help"
  on:open={onOpen}
>
  <div class="reveal max-h-[90%]">
    <div class="slides">
      <section data-markdown={helpUrl} data-separator-vertical="^\n>>>\n" />
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      color="blue"
      type="button"
      class="mr-2"
      size="xs"
      on:click={() => {
        reveal.toggleOverview();
      }}
    >
      <Icon path={icons.mdiGrid} size={1} />
      {$_("Help.Overiview")}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="lime"
      class="mr-2"
      size="xs"
      on:click={() => {
        BrowserOpenURL(
          `https://lhx98.linkclub.jp/twise.co.jp/download/twsnmpfk_${lang}.pdf`
        );
      }}
    >
      <Icon path={icons.mdiFilePdfBox} size={1} />
      {$_("Help.Manual")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      class="mr-2"
      size="xs"
      on:click={() => {
        showFeedback = true;
      }}
    >
      <Icon path={icons.mdiChat} size={1} />
      {$_("Help.Feedback")}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      size="xs"
      on:click={close}
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("Help.Close")}
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showFeedback} size="md" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Help.Feedback")}
    </h3>
    {#if feedbackError}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {$_("Help.feedbackError")}
        </div>
      </Alert>
    {/if}
    <Textarea rows="10" bind:value={feedback}></Textarea>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="red"
        type="button"
        size="xs"
        on:click={async () => {
          if (feedback) {
            sending = true;
            feedbackError = false;
            if (await SendFeedback(feedback)) {
              showFeedback = false;
              sending = false;
              return;
            }
          }
          sending = false;
          feedbackError = true;
        }}
      >
        {#if sending}
          <Spinner class="me-3" size="4" />
        {:else}
          <Icon path={icons.mdiSend} size={1} />
        {/if}
        {$_("Help.Send")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        size="xs"
        on:click={() => {
          showFeedback = false;
        }}
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<style global>
  .help .reveal img {
    margin: 0 auto;
  }

  .help .reveal pre.code-wrapper {
    background-color: white;
  }
</style>
