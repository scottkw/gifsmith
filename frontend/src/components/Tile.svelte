<script lang="ts">
  import type { Gif } from '../lib/api';

  let {
    gif,
    index,
    selected,
    copying,
    onselect,
    oncopy,
  }: {
    gif: Gif;
    index: number;
    selected: boolean;
    copying: boolean;
    onselect: (i: number) => void;
    oncopy: (i: number) => void;
  } = $props();

  // Keyed by src so a tile reused for another result after a new search loads again.
  let brokenSrc = $state('');
  const broken = $derived(brokenSrc === gif.previewUrl);
</script>

<!-- Keyboard selection goes through the combobox input (aria-activedescendant), never tile focus;
     mousedown preventDefault keeps focus in the input when a tile is clicked. -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  class="tile"
  id="tile-{index}"
  role="option"
  tabindex="-1"
  aria-selected={selected}
  aria-label={gif.title || `GIF ${index + 1}`}
  onmousedown={(e) => e.preventDefault()}
  onclick={() => onselect(index)}
  ondblclick={() => oncopy(index)}
>
  {#if broken}
    <svg class="broken" viewBox="0 0 16 16" aria-hidden="true">
      <circle cx="8" cy="8" r="6.25" />
      <path d="M8 4.75v3.75M8 11h.01" />
    </svg>
  {:else}
    <img
      src={gif.previewUrl}
      alt=""
      decoding="async"
      draggable="false"
      onerror={() => (brokenSrc = gif.previewUrl)}
    />
  {/if}
  {#if copying && selected}
    <div class="overlay"><span class="spinner" role="status" aria-label="Loading"></span></div>
  {/if}
</div>

<style>
  .tile {
    position: relative;
    border-radius: 8px;
    background: var(--surface);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .tile:hover {
    background: var(--surface-active);
  }
  .tile[aria-selected='true'] {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
  .broken {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: var(--text-tertiary);
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.4);
    color: #fff;
  }
</style>
