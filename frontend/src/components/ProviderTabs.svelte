<script lang="ts">
  // Four fixed provider pills, a divider, then the library pills (UI-SPEC Tabs row).
  // Not in the Tab order: ⌘1–⌘6 is the keyboard path.
  import { LIBRARY_TABS, SLOTS, type LibraryId, type SlotId } from '../lib/providers';

  let {
    active,
    lib,
    clipsEnabled,
    onswitch,
    onlib,
  }: {
    active: SlotId;
    lib: LibraryId | null; // a library tab is showing instead of the provider grid
    clipsEnabled: boolean;
    onswitch: (id: SlotId) => void;
    onlib: (id: LibraryId) => void;
  } = $props();
</script>

<div class="tabs" role="tablist">
  {#each SLOTS as s (s.id)}
    {@const off = s.id === 'clips' && !clipsEnabled}
    <button
      role="tab"
      tabindex="-1"
      aria-selected={s.id === active && !lib}
      aria-disabled={off ? 'true' : undefined}
      title={off ? 'Clips need a GIPHY key approved for Clips' : undefined}
      class:active={s.id === active && !lib}
      class:off
      onclick={() => onswitch(s.id)}
    >
      {s.label}<span class="hint">{s.shortcut}</span>
    </button>
  {/each}
  <span class="divider" aria-hidden="true"></span>
  {#each LIBRARY_TABS as t (t.id)}
    <button role="tab" tabindex="-1" aria-selected={t.id === lib} class:active={t.id === lib} onclick={() => onlib(t.id)}>
      {#if t.star}
        <svg class="star" viewBox="0 0 16 16" aria-hidden="true"
          ><path d="M8 1.5l1.9 4.1 4.5.5-3.4 3 1 4.4L8 11.3l-4 2.2 1-4.4-3.4-3 4.5-.5z" /></svg
        >
      {/if}{t.label}<span class="hint">{t.shortcut}</span>
    </button>
  {/each}
</div>

<style>
  .tabs {
    box-sizing: border-box;
    height: 36px;
    flex: none;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 0 8px;
    border-bottom: 1px solid var(--hairline);
    white-space: nowrap;
  }
  button {
    height: 28px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--text-secondary);
    font: var(--body);
    cursor: pointer;
  }
  button.active {
    background: var(--surface-active);
    color: var(--text);
  }
  button.off:not(.active) {
    color: var(--text-tertiary);
  }
  .divider {
    flex: none;
    width: 1px;
    height: 16px;
    margin: 0 4px;
    background: var(--hairline);
  }
  /* The label's 8px gap is the pill's flex gap; the star sits 4px before its label. */
  .star {
    width: 12px;
    height: 12px;
    margin-right: -4px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.25;
    stroke-linejoin: round;
  }
  .hint {
    font: var(--label);
    color: var(--text-tertiary);
  }
</style>
