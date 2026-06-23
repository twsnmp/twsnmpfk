<script lang="ts">
  import { onDestroy, untrack } from "svelte";
  import { CodeJar as CreateCodeJar } from "codejar";
  import { withLineNumbers as _withLineNumbers } from "codejar/linenumbers";

  interface Props {
    element?: HTMLElement;
    class?: string;
    style?: string;
    addClosing?: boolean;
    catchTab?: boolean;
    history?: boolean;
    indentOn?: RegExp;
    preserveIdent?: boolean;
    spellcheck?: boolean;
    tab?: string;
    withLineNumbers?: boolean;
    highlight?: (code: string, syntax?: string) => string;
    syntax?: string;
    value?: string;
  }

  let {
    element = $bindable(undefined),
    class: className = "",
    style = "",
    addClosing = false,
    catchTab = false,
    history = true,
    indentOn = /{$/,
    preserveIdent = true,
    spellcheck = false,
    tab = "\t",
    withLineNumbers = undefined,
    highlight = undefined,
    syntax = undefined,
    value = $bindable("")
  }: Props = $props();

  let editorElement: HTMLElement;
  let jar: any;
  let ignoreUpdate = false;

  function wrapHighlight(hl: any, syn: any, wln: any) {
    const _highlight = hl
      ? (el: HTMLElement) => {
          el.innerHTML = hl(el.textContent ?? "", syn);
        }
      : () => {};
    return wln ? _withLineNumbers(_highlight) : _highlight;
  }

  function handleKeyDownCapture(e: KeyboardEvent) {
    if (e.isComposing || e.keyCode === 229) {
      e.stopPropagation();
    }
  }

  function initJar(
    hl: any,
    syn: any,
    wln: any,
    val: string,
    opts: {
      addClosing: boolean;
      catchTab: boolean;
      history: boolean;
      indentOn: RegExp;
      preserveIdent: boolean;
      spellcheck: boolean;
      tab: string;
    }
  ) {
    destroyJar();

    if (!editorElement) return;

    editorElement.addEventListener("keydown", handleKeyDownCapture, { capture: true });

    jar = CreateCodeJar(editorElement, wrapHighlight(hl, syn, wln), opts);
    
    // Set initial value
    jar.updateCode(val);

    jar.onUpdate((code: string) => {
      ignoreUpdate = true;
      value = code;
    });

    if (element === undefined) {
      element = editorElement;
    }
  }

  function destroyJar() {
    if (editorElement) {
      editorElement.removeEventListener("keydown", handleKeyDownCapture, { capture: true });
    }
    if (jar) {
      jar.destroy();
      jar = undefined;
      // Remove linenumbers wrapper if it was added by codejar/linenumbers
      const wrap = editorElement?.parentElement;
      if (wrap && wrap.classList.contains("codejar-wrap")) {
        const parent = wrap.parentElement;
        if (parent) {
          editorElement.style.padding = "";
          parent.appendChild(editorElement);
          wrap.remove();
        }
      }
    }
  }

  // Manage lifecycle and changes in Svelte 5
  $effect(() => {
    // Re-initialize if element or highlighting options change
    if (editorElement) {
      const hl = highlight;
      const wln = withLineNumbers;
      const syn = syntax;
      const val = untrack(() => value);
      const opts = {
        addClosing,
        catchTab,
        history,
        indentOn,
        preserveIdent,
        spellcheck,
        tab
      };
      untrack(() => {
        initJar(hl, syn, wln, val, opts);
      });
    }
  });

  $effect(() => {
    // Update code if value prop changes externally
    if (jar && !ignoreUpdate) {
      if (value !== jar.toString()) {
        jar.updateCode(value);
      }
    }
    ignoreUpdate = false;
  });

  $effect(() => {
    // Update options if they change
    if (jar) {
      jar.updateOptions({
        addClosing,
        catchTab,
        history,
        indentOn,
        preserveIdent,
        spellcheck,
        tab
      });
    }
  });

  onDestroy(() => {
    destroyJar();
  });
</script>

<pre
  bind:this={editorElement}
  class="{syntax ? `language-${syntax}` : ''} {className}"
  style={style ? style : ""}
><code class={syntax ? `language-${syntax}` : ''}>{#if highlight}{@html highlight(value, syntax)}{:else}{value}{/if}</code></pre>

<style>
  pre[class*="language-"] {
    background-color: #f9fafb !important; /* bg-gray-50 */
    color: #111827 !important; /* text-gray-900 */
    border: 1px solid #d1d5db !important; /* border-gray-300 */
    border-radius: 0.5rem !important; /* rounded-lg */
    padding: 0.5rem !important; /* p-2 */
    font-size: 0.875rem !important; /* text-sm */
    text-shadow: none !important;
    overflow: auto;
  }

  pre[class*="language-"]:focus {
    outline: none !important;
    border-color: #3b82f6 !important; /* focus:border-blue-500 */
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.5) !important; /* focus:ring-blue-500 */
  }

  :global(.dark) pre[class*="language-"] {
    background-color: #374151 !important; /* bg-gray-700 */
    color: #f9fafb !important; /* text-gray-50 */
    border-color: #4b5563 !important; /* border-gray-600 */
  }

  :global(.dark) pre[class*="language-"]:focus {
    border-color: #3b82f6 !important;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.5) !important;
  }

  pre[class*="language-"] :global(.token) {
    background: none !important;
  }

  /* Dark mode token overrides */
  :global(.dark) pre[class*="language-"] :global(.token.comment),
  :global(.dark) pre[class*="language-"] :global(.token.prolog),
  :global(.dark) pre[class*="language-"] :global(.token.doctype),
  :global(.dark) pre[class*="language-"] :global(.token.cdata) {
    color: #999999 !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.punctuation) {
    color: #cccccc !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.property),
  :global(.dark) pre[class*="language-"] :global(.token.tag),
  :global(.dark) pre[class*="language-"] :global(.token.boolean),
  :global(.dark) pre[class*="language-"] :global(.token.number),
  :global(.dark) pre[class*="language-"] :global(.token.constant),
  :global(.dark) pre[class*="language-"] :global(.token.symbol),
  :global(.dark) pre[class*="language-"] :global(.token.deleted) {
    color: #f08d49 !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.selector),
  :global(.dark) pre[class*="language-"] :global(.token.attr-name),
  :global(.dark) pre[class*="language-"] :global(.token.string),
  :global(.dark) pre[class*="language-"] :global(.token.char),
  :global(.dark) pre[class*="language-"] :global(.token.builtin),
  :global(.dark) pre[class*="language-"] :global(.token.inserted) {
    color: #7ec699 !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.operator),
  :global(.dark) pre[class*="language-"] :global(.token.entity),
  :global(.dark) pre[class*="language-"] :global(.token.url) {
    color: #a67f59 !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.atrule),
  :global(.dark) pre[class*="language-"] :global(.token.attr-value),
  :global(.dark) pre[class*="language-"] :global(.token.keyword) {
    color: #cc99cd !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.function),
  :global(.dark) pre[class*="language-"] :global(.token.class-name) {
    color: #f8c555 !important;
  }
  :global(.dark) pre[class*="language-"] :global(.token.regex),
  :global(.dark) pre[class*="language-"] :global(.token.important),
  :global(.dark) pre[class*="language-"] :global(.token.variable) {
    color: #e90 !important;
  }
</style>
