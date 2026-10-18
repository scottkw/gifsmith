<script lang="ts">
  // ⌘K action menu (UI-SPEC "Action menu popover" / "Action menu behavior", D-01..D-04). Rows come
  // from lib/actions.ts and run through App's dispatcher via onrun; ⌘ shortcuts bubble up to App.
  import { filterActions, type ActionId, type ActionRow } from '../lib/actions';

  let {
    rows,
    left,
    top,
    onrun,
    onclose,
  }: {
    rows: ActionRow[];
    left: number;
    top: number;
    onrun: (id: ActionId) => void;
    onclose: () => void;
  } = $props();

  let root: HTMLDivElement;
  let field: HTMLInputElement;
  let text = $state('');
  let hi = $state(0);
  const shown = $derived(filterActions(rows, text));
  const filtering = $derived(!!text.trim());

  export function focus() {
    field.focus();
  }
  export function height() {
    return root.offsetHeight;
  }

  function onkeydown(e: KeyboardEvent) {
    // ⌘ shortcuts (⌘K, ⌘↵, ⌘⇧C, …) belong to App; IME keys belong to the IME.
    if (e.metaKey || e.isComposing || e.keyCode === 229) return;
    switch (e.key) {
      case 'ArrowDown':
        hi = Math.min(hi + 1, shown.length - 1);
        break;
      case 'ArrowUp':
        hi = Math.max(hi - 1, 0);
        break;
      case 'Enter':
        if (shown[hi] && !e.repeat) onrun(shown[hi].id);
        break;
      case 'Escape':
        onclose();
        break;
      case 'Tab': // focus stays in the filter
        break;
      default:
        return;
    }
    e.preventDefault();
  }
</script>

<div class="menu" role="dialog" aria-label="Actions" style="left: {left}px; top: {top}px" bind:this={root}>
  <div class="filter">
    <input
      bind:this={field}
      bind:value={text}
      oninput={() => (hi = 0)}
      {onkeydown}
      placeholder="Search actions"
      autocomplete="off"
      spellcheck="false"
      role="combobox"
      aria-controls="action-list"
      aria-expanded="true"
      aria-activedescendant={shown[hi] ? `action-${shown[hi].id}` : undefined}
    />
  </div>
  <div class="list" id="action-list" role="listbox" aria-label="Actions">
    {#each shown as r, i (r.id)}
      {#if !filtering && i > 0 && r.group !== shown[i - 1].group}
        <div class="sep" aria-hidden="true"></div>
      {/if}
      <div
        id="action-{r.id}"
        class="row"
        class:hi={i === hi}
        role="option"
        aria-selected={i === hi}
        aria-keyshortcuts={r.keys || undefined}
      >
        <span class="label">{r.label}</span>
        {#if r.hint}<span class="hint" aria-hidden="true"><b>{r.hint}</b></span>{/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .menu {
    position: fixed;
    z-index: 10;
    box-sizing: border-box;
    width: 320px;
    background: var(--bg);
    border: 1px solid var(--hairline);
    border-radius: 12px;
    box-shadow: var(--shadow-popover);
  }
  .filter {
    box-sizing: border-box;
    height: 40px;
    display: flex;
    align-items: center;
    padding: 0 16px;
    border-bottom: 1px solid var(--hairline);
  }
  input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: 0;
    padding: 0;
    background: transparent;
    color: var(--text);
    font: var(--body);
    caret-color: var(--accent);
  }
  input::placeholder {
    color: var(--text-secondary);
  }
  .list {
    padding: 8px;
  }
  .row {
    box-sizing: border-box;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 0 8px;
    border-radius: 6px;
    color: var(--text);
    font: var(--body);
    cursor: default;
  }
  .row.hi {
    background: var(--surface-active);
  }
  .label {
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .hint {
    flex: none;
    font: var(--label);
    color: var(--text-tertiary);
  }
  .hint b {
    font-weight: 400;
    color: var(--text-secondary);
  }
  .sep {
    height: 1px;
    margin: 4px 0;
    background: var(--hairline);
  }
</style>
