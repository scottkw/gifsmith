<script lang="ts">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack } from 'svelte';
  import type { Gif } from '../lib/api';
  import Tile from './Tile.svelte';

  type Paging = 'idle' | 'loading' | 'end' | 'error';
  let {
    items,
    selected,
    copying,
    paging,
    onselect,
    oncopy,
    onneedmore,
    onretrypage,
  }: {
    items: Gif[];
    selected: number;
    copying: boolean;
    paging: Paging;
    onselect: (i: number) => void;
    oncopy: (i: number) => void;
    onneedmore: () => void;
    onretrypage: () => void;
  } = $props();

  const COLS = 4;
  const ROW = 152; // UI-SPEC: 144 tile + 8 gap
  const EXTRA = 48; // S5 / S6 / S12a row

  let scrollEl: HTMLDivElement;
  const tileRows = $derived(Math.ceil(items.length / COLS));
  // S6 only once 2+ rows are loaded; idle shows no extra row.
  const extra = $derived(
    paging === 'loading' || paging === 'error' || (paging === 'end' && tileRows >= 2) ? paging : null,
  );

  const virt = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scrollEl,
    estimateSize: (i) => (i < tileRows ? ROW : EXTRA),
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
    const row = Math.floor(selected / COLS);
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
          <div class="tiles">
            {#each items.slice(row.index * COLS, row.index * COLS + COLS) as g, c (c)}
              {@const i = row.index * COLS + c}
              <Tile gif={g} index={i} selected={i === selected} {copying} {onselect} {oncopy} />
            {/each}
          </div>
        {:else if extra === 'loading'}
          <span class="spinner" role="status" aria-label="Loading"></span>
        {:else if extra === 'end'}
          <span class="note">End of results</span>
        {:else if extra === 'error'}
          <span class="note secondary">Couldn't load more.</span>
          <button class="link" onmousedown={(e) => e.preventDefault()} onclick={onretrypage}>Try Again</button>
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
    grid-template-columns: repeat(4, 166px);
    grid-auto-rows: 144px;
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
</style>
