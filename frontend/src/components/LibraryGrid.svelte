<script lang="ts">
  // Favorites / Recents grid (UI-SPEC "Favorites and Recents tabs"): virtual rows of two heights,
  // a 32px provider header per section, then tile rows (tile + 8 px). Tiles carry their flat index.
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack } from 'svelte';
  import { rows as toRows, type Section } from '../lib/library';
  import { animates } from '../lib/motion';
  import Tile from './Tile.svelte';

  let {
    sections,
    cols,
    tile,
    selected,
    overlay,
    favorites,
    badges,
    reduced = false,
    onselect,
    oncopy,
  }: {
    sections: Section[];
    cols: number; // 3, 4 or 5 (the Search results grid size)
    tile: number; // tile height
    selected: number;
    overlay: boolean;
    favorites: Set<string>; // provider:id keys
    badges: boolean; // favorite badges (not on the Favorites tab)
    reduced?: boolean; // prefers-reduced-motion: only the selected tile animates
    onselect: (i: number) => void;
    oncopy: (i: number) => void;
  } = $props();

  let scrollEl: HTMLDivElement;
  const rows = $derived(toRows(sections, cols, tile + 8));

  // Sizes are read when the count changes; App remounts this grid when the row layout changes.
  const virt = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scrollEl,
    estimateSize: (i) => rows[i]?.size ?? 0,
    overscan: 2,
    paddingStart: 8,
    paddingEnd: 16,
    scrollPaddingStart: 16,
    scrollPaddingEnd: 16,
  });

  $effect(() => {
    const count = rows.length;
    untrack(() => $virt.setOptions({ count }));
  });
  $effect(() => {
    const row = rows.findIndex((r) => r.kind === 'tiles' && selected >= r.start && selected < r.start + r.items.length);
    if (row >= 0) untrack(() => $virt.scrollToIndex(row, { align: 'auto' }));
  });
</script>

<div class="scroller" bind:this={scrollEl} id="grid" role="listbox" aria-label="Results">
  <div class="inner" style="height: {$virt.getTotalSize()}px">
    {#each $virt.getVirtualItems() as v (v.index)}
      {@const row = rows[v.index]}
      <div class="row" style="height: {v.size}px; transform: translateY({v.start}px)">
        {#if row.kind === 'header'}
          {@const s = sections[row.section]}
          <div class="header" role="presentation">{s.label} <span class="count">· {s.items.length}</span></div>
        {:else}
          <div class="tiles" style="--cols: {cols}; --tile: {tile}px">
            {#each row.items as g, c (c)}
              {@const i = row.start + c}
              <Tile
                gif={g}
                index={i}
                selected={i === selected}
                still={!animates(i, selected, reduced)}
                {overlay}
                provider={sections[row.section].label}
                favorite={badges && favorites.has(`${g.provider}:${g.id}`)}
                {onselect}
                {oncopy}
              />
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .scroller {
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
  }
  .inner {
    position: relative;
    width: 100%;
  }
  .row {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
  }
  .header {
    box-sizing: border-box;
    height: 32px;
    padding: 8px 16px;
    font: var(--label);
    color: var(--text-secondary);
  }
  .count {
    color: var(--text-tertiary);
  }
  .tiles {
    box-sizing: border-box;
    width: 100%;
    display: grid;
    grid-template-columns: repeat(var(--cols), 1fr);
    grid-auto-rows: var(--tile);
    gap: 8px;
    padding: 0 16px;
  }
</style>
