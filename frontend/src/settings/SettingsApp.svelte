<script lang="ts">
  // Settings window root (UI-SPEC "Layout"): vertical sidebar tablist + the active pane.
  // The page is never reloaded, so the active pane survives hide/show within a launch.
  import { Events } from '@wailsio/runtime';
  import { onMount, tick, untrack } from 'svelte';
  import type { SettingsSnapshot } from '../lib/api';
  import GeneralPane from './GeneralPane.svelte';
  import SearchPane from './SearchPane.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  // UI-SPEC order (General, Keys, Hotkeys, Search, Downloads, About); only built panes are listed.
  const PANES = [
    { id: 'general', name: 'General', component: GeneralPane },
    { id: 'search', name: 'Search', component: SearchPane },
  ];

  let active = $state(0);
  let snap = $state(untrack(() => initial)); // a pane mounts with the latest snapshot
  const tabs: HTMLButtonElement[] = [];

  onMount(() => Events.On('settings:changed', (e) => (snap = e.data as SettingsSnapshot)));

  // ↑ ↓ move between panes and show them at once, clamped; the selected tab is the only tab stop.
  async function onkeydown(e: KeyboardEvent) {
    const d = e.key === 'ArrowDown' ? 1 : e.key === 'ArrowUp' ? -1 : 0;
    if (!d) return;
    e.preventDefault();
    active = Math.min(PANES.length - 1, Math.max(0, active + d));
    await tick();
    tabs[active].focus();
  }
</script>

<div class="window">
  <div class="sidebar" role="tablist" aria-orientation="vertical" aria-label="Settings" tabindex="-1" {onkeydown}>
    {#each PANES as p, i (p.id)}
      <button
        bind:this={tabs[i]}
        role="tab"
        id="tab-{p.id}"
        aria-selected={i === active}
        aria-controls="pane-{p.id}"
        tabindex={i === active ? 0 : -1}
        onclick={() => (active = i)}>{p.name}</button
      >
    {/each}
  </div>
  {#each PANES as p, i (p.id)}
    {#if i === active}
      <div class="pane" role="tabpanel" id="pane-{p.id}" aria-labelledby="tab-{p.id}">
        <h1>{p.name}</h1>
        <p.component initial={snap} />
      </div>
    {/if}
  {/each}
</div>

<style>
  .window {
    display: flex;
    height: 100%;
  }
  .sidebar {
    box-sizing: border-box;
    width: 160px;
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px;
    overflow: hidden;
    background: var(--surface);
    border-right: 1px solid var(--hairline);
  }
  .sidebar button {
    height: 32px;
    padding: 0 8px;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--text);
    font: var(--body);
    text-align: left;
    cursor: default;
  }
  .sidebar button:hover,
  .sidebar button[aria-selected='true'] {
    background: var(--surface-active);
  }
  .pane {
    box-sizing: border-box;
    width: 480px;
    padding: 24px;
    overflow-y: auto;
    background: var(--bg);
  }
  h1 {
    margin: 0 0 16px;
    font: var(--heading);
    color: var(--text);
  }
</style>
