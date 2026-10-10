<script context="module" lang="ts">
  const STORAGE_KEY = "twsnmp_mibtree_expansion";

  function loadExpansionState(): Map<string, boolean> {
    const map = new Map<string, boolean>();
    map.set(".1", true); // Default root expanded
    try {
      if (typeof localStorage !== "undefined") {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw) {
          const obj = JSON.parse(raw);
          if (typeof obj === "object" && obj !== null) {
            for (const [k, v] of Object.entries(obj)) {
              map.set(k, Boolean(v));
            }
          }
        }
      }
    } catch {}
    return map;
  }

  export const globalExpansionState = loadExpansionState();

  export function saveExpansionState() {
    try {
      if (typeof localStorage !== "undefined") {
        const obj: Record<string, boolean> = {};
        for (const [k, v] of globalExpansionState.entries()) {
          if (v) obj[k] = true;
        }
        localStorage.setItem(STORAGE_KEY, JSON.stringify(obj));
      }
    } catch {}
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Tooltip } from "flowbite-svelte";

  export let tree: any;
  export let onselect: ((name: string, oid: string) => void) | undefined = undefined;
  export let expansionState: Map<string, boolean> = globalExpansionState;

  $: ({ oid, name, MIBInfo, children, count, forceExpand } = tree || {});
  const dispatch = createEventDispatcher();

  let mibInfoTooltip = "";
  let type = "";
  $: {
    if (MIBInfo) {
      mibInfoTooltip = MIBInfo.Description || "";
      type = ":" + (MIBInfo.Type || "");
    } else {
      mibInfoTooltip = "";
      type = "";
    }
  }

  let expanded = false;
  $: {
    if (forceExpand) {
      expanded = true;
    } else if (oid && expansionState.has(oid)) {
      expanded = !!expansionState.get(oid);
    } else if (oid === ".1") {
      expanded = true;
    } else {
      expanded = false;
    }
  }

  const toggleExpansion = (e: MouseEvent) => {
    e.stopPropagation();
    expanded = !expanded;
    if (oid) {
      expansionState.set(oid, expanded);
      saveExpansionState();
    }
  };

  const handleSelect = (e: MouseEvent | KeyboardEvent) => {
    e.stopPropagation();
    if (name) {
      dispatch("select", name);
      if (onselect) {
        onselect(name, oid);
      }
    }
  };

  $: arrowDown = expanded;
</script>

{#if tree}
  <ul>
    <li>
      {#if children && children.length > 0}
        <!-- Branch node with children -->
        <span class="inline-flex items-center gap-0.5 py-0.5 rounded text-slate-700 dark:text-slate-200">
          <!-- Arrow toggle button -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <span
            class="arrow p-1 hover:text-cyan-500 rounded cursor-pointer select-none"
            class:arrowDown
            onclick={toggleExpansion}
            title={expanded ? "Collapse" : "Expand"}
            role="button"
            tabindex="0"
            onkeydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                toggleExpansion(e as any);
              }
            }}
          >
            &#x25b6;
          </span>

          <!-- Object name (single click selection) -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <span
            class="cursor-pointer py-0.5 px-1 rounded hover:bg-sky-200 dark:hover:bg-sky-900 transition-colors"
            onclick={handleSelect}
            ondblclick={handleSelect}
            role="button"
            tabindex="0"
            onkeydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleSelect(e);
              }
            }}
          >
            <span class="font-semibold text-teal-700 dark:text-teal-300">{name}</span>
            <span class="text-slate-500 dark:text-slate-400">({oid}{type})</span>
            {#if count > 0}
              <span class="text-teal-600 font-bold">: {count}</span>
            {/if}
          </span>
          {#if mibInfoTooltip}
            <Tooltip trigger="hover">
              <pre class="font-sans whitespace-pre-wrap leading-relaxed max-h-56 overflow-y-auto text-xs">{mibInfoTooltip}</pre>
            </Tooltip>
          {/if}
        </span>

        {#if expanded}
          {#each children as child}
            <svelte:self
              tree={child}
              {expansionState}
              {onselect}
              on:select
            />
          {/each}
        {/if}
      {:else}
        <!-- Leaf node without children -->
        <span class="inline-flex items-center gap-0.5 py-0.5 rounded text-slate-700 dark:text-slate-200">
          {#if type == ":Notification"}
            <span class="no-arrow text-red-500 font-bold select-none">*</span>
          {:else}
            <span class="no-arrow select-none"></span>
          {/if}

          <!-- Object name (single click selection) -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <span
            class="cursor-pointer py-0.5 px-1 rounded hover:bg-sky-200 dark:hover:bg-sky-900 transition-colors"
            onclick={handleSelect}
            ondblclick={handleSelect}
            role="button"
            tabindex="0"
            onkeydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleSelect(e);
              }
            }}
          >
            <span class="font-semibold text-teal-700 dark:text-teal-300">{name}</span>
            <span class="text-slate-500 dark:text-slate-400">({oid}{type})</span>
            {#if count > 0}
              <span class="text-teal-600 font-bold">: {count}</span>
            {/if}
          </span>
          {#if mibInfoTooltip}
            <Tooltip trigger="hover">
              <pre class="font-sans whitespace-pre-wrap leading-relaxed max-h-56 overflow-y-auto text-xs">{mibInfoTooltip}</pre>
            </Tooltip>
          {/if}
        </span>
      {/if}
    </li>
  </ul>
{/if}

<style>
  ul {
    margin: 0;
    list-style: none;
    padding-left: 1.2rem;
    user-select: none;
  }
  .no-arrow {
    display: inline-block;
    width: 1.2rem;
    text-align: center;
  }
  .arrow {
    cursor: pointer;
    display: inline-block;
    width: 1.2rem;
    text-align: center;
    transition: transform 0.15s ease;
  }
  .arrowDown {
    transform: rotate(90deg);
  }
</style>
