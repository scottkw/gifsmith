<script lang="ts">
  // One Settings row (UI-SPEC Layout): 144px label, 16px gap, 272px control; helper text, or an
  // inline error in its place.
  import type { Snippet } from 'svelte';

  let {
    label,
    id,
    helper = '',
    error = '',
    children,
  }: { label: string; id: string; helper?: string; error?: string; children: Snippet } = $props();
</script>

<div class="row">
  <span class="label" {id}>{label}</span>
  <div class="control">
    <div class="line">{@render children()}</div>
    {#if error}
      <p class="error" role="alert">
        <svg viewBox="0 0 16 16" aria-hidden="true">
          <circle cx="8" cy="8" r="6.25" />
          <path d="M8 5v3.5" />
          <circle class="dot" cx="8" cy="11" r="0.4" />
        </svg>
        {error}
      </p>
    {:else if helper}
      <p class="helper">{helper}</p>
    {/if}
  </div>
</div>

<style>
  .row {
    display: grid;
    grid-template-columns: 144px 272px;
    gap: 16px;
    align-items: start;
  }
  .row + :global(.row) {
    margin-top: 8px;
  }
  .label,
  .line {
    min-height: 32px;
    display: flex;
    align-items: center;
  }
  .label {
    font: var(--body);
    color: var(--text);
  }
  .line {
    gap: 8px;
    flex-wrap: wrap;
  }
  p {
    margin: 4px 0 0;
    font: var(--label);
  }
  .helper {
    color: var(--text-secondary);
  }
  .error {
    display: flex;
    gap: 4px;
    color: var(--destructive);
  }
  .error svg {
    width: 12px;
    height: 12px;
    flex: none;
    margin-top: 2px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .error .dot {
    fill: currentColor;
  }
</style>
