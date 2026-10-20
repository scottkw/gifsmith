<script lang="ts">
  import type { Gif } from '../lib/api';

  let {
    gif,
    index,
    selected,
    overlay,
    provider = '',
    favorite = false,
    still = false,
    onselect,
    oncopy,
    onmenu,
  }: {
    gif: Gif;
    index: number;
    selected: boolean;
    overlay: boolean; // busy overlay on the selected tile
    provider?: string; // library tiles name their provider (sections are not announced)
    favorite?: boolean; // star badge
    still?: boolean; // Reduce Motion: a first-frame still instead of the animated preview
    onselect: (i: number) => void;
    oncopy: (i: number) => void;
    onmenu: (i: number) => void; // right-click: select and open the action menu (P15)
  } = $props();

  // Keyed by src so a tile reused for another result after a new search loads again.
  let brokenSrc = $state('');
  const broken = $derived(brokenSrc === gif.previewUrl);

  // Draws src's first frame once, scaled like object-fit: contain, then drops the Image so nothing
  // keeps decoding. The URL is the provider's, unaltered (results carry no still rendition).
  // Cross-origin drawing only taints the canvas; nothing reads it back.
  function firstFrame(canvas: HTMLCanvasElement, src: string) {
    let img: HTMLImageElement | null = null;
    const draw = (s: string) => {
      if (img) img.onload = img.onerror = null;
      const im = (img = new Image());
      im.onload = () => {
        img = null;
        const r = window.devicePixelRatio || 1;
        const w = (canvas.width = Math.round(canvas.clientWidth * r));
        const h = (canvas.height = Math.round(canvas.clientHeight * r));
        const k = Math.min(w / im.naturalWidth, h / im.naturalHeight);
        const dw = im.naturalWidth * k;
        const dh = im.naturalHeight * k;
        canvas.getContext('2d')?.drawImage(im, (w - dw) / 2, (h - dh) / 2, dw, dh);
      };
      im.onerror = () => {
        img = null;
        brokenSrc = s;
      };
      im.src = s;
    };
    draw(src);
    return {
      update: draw,
      destroy() {
        if (img) img.onload = img.onerror = null;
        img = null;
      },
    };
  }
</script>

<!-- Keyboard selection goes through the combobox input (aria-activedescendant), never tile focus. -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  class="tile"
  id="tile-{index}"
  role="option"
  tabindex="-1"
  aria-selected={selected}
  aria-label={(gif.title || `GIF ${index + 1}`) +
    (provider ? `, ${provider}` : '') +
    (gif.isClip ? ', MP4' : '') +
    (favorite ? ', favorite' : '')}
  onclick={() => onselect(index)}
  ondblclick={() => oncopy(index)}
  oncontextmenu={(e) => {
    e.preventDefault();
    onmenu(index);
  }}
>
  {#if broken}
    <svg class="broken" viewBox="0 0 16 16" aria-hidden="true">
      <circle cx="8" cy="8" r="6.25" />
      <path d="M8 4.75v3.75M8 11h.01" />
    </svg>
  {:else if still}
    <canvas use:firstFrame={gif.previewUrl} aria-hidden="true"></canvas>
  {:else}
    <img
      src={gif.previewUrl}
      alt=""
      decoding="async"
      draggable="false"
      onerror={() => (brokenSrc = gif.previewUrl)}
    />
  {/if}
  {#if gif.isClip}
    <span class="badge" aria-hidden="true">MP4</span>
  {/if}
  {#if favorite}
    <span class="fav" aria-hidden="true">
      <svg viewBox="0 0 16 16"><path d="M8 1.5l1.9 4.1 4.5.5-3.4 3 1 4.4L8 11.3l-4 2.2 1-4.4-3.4-3 4.5-.5z" /></svg>
    </span>
  {/if}
  {#if overlay && selected}
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
  img,
  canvas {
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
  .badge {
    position: absolute;
    top: 4px;
    left: 4px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    background: var(--scrim);
    color: #fff;
    font: var(--label);
  }
  .fav {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--scrim);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .fav svg {
    width: 12px;
    height: 12px;
    fill: #fff;
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
