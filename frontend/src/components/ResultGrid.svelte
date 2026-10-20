<script lang="ts">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack } from 'svelte';
  import type { Gif } from '../lib/api';
  import { animates } from '../lib/motion';
  import Tile from './Tile.svelte';

  type Paging = 'idle' | 'loading' | 'end' | 'error';
  let {
    items,
    cols,
    tile,
    selected,
    overlay,
    favorites = new Set<string>(),
    paging,
    pagingNote = "Couldn't load more.",
    pagingRetry = true,
    reduced = false,
    onselect,
    oncopy,
    onneedmore,
    onretrypage,
  }: {
    items: Gif[];
    cols: number; // 3, 4 or 5 (grid size)
    tile: number; // tile height; rows are tile + 8 apart
    selected: number;
    overlay: boolean; // busy overlay on the selected tile
    favorites?: Set<string>; // provider:id keys that get the star badge
    paging: Paging;
    pagingNote?: string; // E6, or E2 while a 429 window is open (S9a)
    pagingRetry?: boolean; // false while the provider is still blocked
    reduced?: boolean; // prefers-reduced-motion: only the selected tile animates
    onselect: (i: number) => void;
    oncopy: (i: number) => void;
    onneedmore: () => void;
    onretrypage: () => void;
  } = $props();

  const EXTRA = 48; // S5 / S6 / S12a row

  let scrollEl: HTMLDivElement;
  const tileRows = $derived(Math.ceil(items.length / cols));
  // S6 only once 2+ rows are loaded; idle shows no extra row.
  const extra = $derived(
    paging === 'loading' || paging === 'error' || (paging === 'end' && tileRows >= 2) ? paging : null,
  );

  const virt = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scrollEl,
    estimateSize: (i) => (i < tileRows ? tile + 8 : EXTRA), // App remounts this grid on a size change
    overscan: 2,
    paddingStart: 16,
    paddingEnd: 16,
    scrollPaddingStart: 16,
    scrollPaddingEnd: 16,
  });

  // setOptions/scrollToIndex write the store this component reads; untrack avoids an effect loop.
  $effect(() => {
    const count = tileRows + (extra ? 1 : 0);
    untrack(() => $virt.setOptions({ count }));
  });
  $effect(() => {
    const row = Math.floor(selected / cols);
    untrack(() => $virt.scrollToIndex(row, { align: 'auto' }));
  });
  // Sentinel: the last tile row is mounted (visible or within overscan).
  $effect(() => {
    if (tileRows > 0 && $virt.getVirtualItems().some((r) => r.index === tileRows - 1)) untrack(onneedmore);
  });
</script>

<div class="scroller" bind:this={scrollEl} id="grid" role="listbox" aria-label="Results">
  <div class="inner" style="height: {$virt.getTotalSize()}px">
    {#each $virt.getVirtualItems() as row (row.index)}
      <div class="row" style="height: {row.size}px; transform: translateY({row.start}px)">
        {#if row.index < tileRows}
          <div class="tiles" style="--cols: {cols}; --tile: {tile}px">
            {#each items.slice(row.index * cols, row.index * cols + cols) as g, c (c)}
              {@const i = row.index * cols + c}
              <Tile
                gif={g}
                index={i}
                selected={i === selected}
                still={!animates(i, selected, reduced)}
                {overlay}
                favorite={favorites.has(`${g.provider}:${g.id}`)}
                {onselect}
                {oncopy}
              />
            {/each}
          </div>
        {:else if extra === 'loading'}
          <span class="spinner" role="status" aria-label="Loading"></span>
        {:else if extra === 'end'}
          <span class="note">End of results</span>
        {:else if extra === 'error'}
          <span class="note secondary">{pagingNote}</span>
          {#if pagingRetry}
            <button class="link" onclick={onretrypage}>Try Again</button>
          {/if}
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
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
  }
  .tiles {
    align-self: flex-start;
    box-sizing: border-box;
    width: 100%;
    display: grid;
    grid-template-columns: repeat(var(--cols), 1fr);
    grid-auto-rows: var(--tile);
    gap: 8px;
    padding: 0 16px;
  }
  .note {
    font: var(--label);
    color: var(--text-tertiary);
  }
  .note.secondary {
    color: var(--text-secondary);
  }
  .link {
    margin-left: 4px;
    padding: 0;
    border: 0;
    background: none;
    font: var(--label);
    color: var(--accent);
    cursor: pointer;
  }
  .link:hover {
    text-decoration: underline;
  }
</style>
