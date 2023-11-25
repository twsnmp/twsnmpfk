<script lang="ts">
  import { Modal, GradientButton } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
  import { lang } from "../i18n/i18n";
  import Reveal from "reveal.js";
  import Highlight from "reveal.js/plugin/highlight/highlight";
  import Markdown from "reveal.js/plugin/markdown/markdown";
  import "reveal.js/dist/reveal.css";
  import "reveal.js/dist/theme/black.css";
  import "reveal.js/plugin/highlight/monokai.css";

  export let page = "";
  let show: boolean = false;
  let reveal: Reveal.Api | undefined = undefined;
  let helpUrl = "";

  const dispatch = createEventDispatcher();

  onMount(async () => {
    helpUrl = `help/${lang}/${page}.md`;
    show = true;
    await tick();
    reveal = new Reveal({
      plugins: [Markdown, Highlight],
    });
    reveal.initialize();
  });

  const close = () => {
    if (reveal) {
      reveal.destroy();
      reveal = undefined;
    }
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full min-h-[90vh] bg-gray-800"
>
  <div class="reveal max-h-[90%]">
    <div class="slides">
      <section data-markdown={helpUrl} />
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      size="xs"
      color="lime"
      class="ml-2"
      on:click={() => {
        BrowserOpenURL(
          `https://lhx98.linkclub.jp/twise.co.jp/download/twsnmpfk_${lang}.pdf`
        );
      }}
    >
      <Icon path={icons.mdiHelp} size={1} />
      {$_('Help.Manual')}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={close}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_('Help.Close')}
    </GradientButton>
  </div>
</Modal>
