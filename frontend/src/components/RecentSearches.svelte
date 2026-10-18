<script lang="ts">
  // Recent-search chips (UI-SPEC "Recent-search chip row", D-12): shown above the grid while the box
  // is empty. Keyboard selection lives in App (the input's aria-activedescendant); hover never moves it.
  let {
    searches,
    selected,
    onrun,
  }: {
    searches: string[];
    selected: number; // -1 = none
    onrun: (q: string) => void;
  } = $props();

  $effect(() => {
    document.getElementById(`chip-${selected}`)?.scrollIntoView({ inline: 'nearest', block: 'nearest' });
  });
</script>

<div class="chips" role="listbox" aria-label="Recent searches">
  {#each searches as q, i (q)}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="chip"
      id="chip-{i}"
      role="option"
      tabindex="-1"
      aria-selected={i === selected}
      aria-label="Search again for {q}"
      title={q}
      onclick={() => onrun(q)}
    >
      <svg viewBox="0 0 16 16" aria-hidden="true">
        <circle cx="8" cy="8" r="6.25" />
        <path d="M8 4.75V8l2.25 1.5" />
      </svg>
      <span class="label">{q}</span>
    </div>
  {/each}
</div>

<style>
  /* 32px in the layout; the 4px below overlaps the grid's top padding so the selection ring is not clipped. */
  .chips {
    box-sizing: border-box;
    flex: none;
    height: 36px;
    margin-bottom: -4px;
    padding: 8px 16px 4px;
    display: flex;
    gap: 8px;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: none;
  }
  .chips::-webkit-scrollbar {
    display: none;
  }
  .chip {
    box-sizing: border-box;
    flex: none;
    height: 24px;
    padding: 0 8px;
    border-radius: 12px;
    background: var(--surface);
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--label);
    color: var(--text-secondary);
    cursor: pointer;
  }
  .chip:hover {
    background: var(--surface-active);
  }
  .chip[aria-selected='true'] {
    background: var(--surface-active);
    color: var(--text);
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  svg {
    width: 12px;
    height: 12px;
    flex: none;
    fill: none;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .label {
    max-width: 160px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
