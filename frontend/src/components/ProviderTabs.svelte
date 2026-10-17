<script lang="ts">
  // Four fixed pills (UI-SPEC Provider Tabs Contract). Not in the Tab order: ⌘1–⌘4 is the keyboard path.
  import { SLOTS, type SlotId } from '../lib/providers';

  let { active, clipsEnabled, onswitch }: { active: SlotId; clipsEnabled: boolean; onswitch: (id: SlotId) => void } =
    $props();
</script>

<div class="tabs" role="tablist">
  {#each SLOTS as s (s.id)}
    {@const off = s.id === 'clips' && !clipsEnabled}
    <button
      role="tab"
      tabindex="-1"
      aria-selected={s.id === active}
      aria-disabled={off ? 'true' : undefined}
      title={off ? 'Clips need a GIPHY key approved for Clips' : undefined}
      class:active={s.id === active}
      class:off
      onclick={() => onswitch(s.id)}
    >
      {s.label}<span class="hint">{s.shortcut}</span>
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
  button.off {
    color: var(--text-tertiary);
  }
  .hint {
    font: var(--label);
    color: var(--text-tertiary);
  }
</style>
