<script lang="ts">
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { ActionService } from '../bindings/gifsmith/internal/services';
  import { search, type Gif } from './lib/api';

  const PROVIDER = 'finergifs';
  const DEBOUNCE_MS = 450;

  let input: HTMLInputElement;
  let query = $state('');
  let items = $state<Gif[]>([]);
  let selected = $state(-1);
  let status = $state('');
  let needsAccessibility = $state(false);
  // Paste hides the panel before ⌘V; a copied-only fallback message must survive to the next show.
  let keepStatus = false;
  let busy = false; // one Copy/Paste in flight; a second Enter would paste twice
  let timer: ReturnType<typeof setTimeout> | undefined;
  let seq = 0; // only the latest search may update the grid

  const footerCenter = $derived(status || items[selected]?.title || '');

  onMount(() => {
    input.focus();
    Events.On('panel:show', () => {
      if (!keepStatus) status = '';
      keepStatus = false;
      needsAccessibility = false;
      input.focus();
      input.select();
    });
    Events.On('panel:status', (e) => {
      status = String(e.data);
    });
  });

  function onInput() {
    clearTimeout(timer);
    const q = query.trim();
    if ([...q].length < 2) {
      seq++;
      items = [];
      selected = -1;
      return;
    }
    timer = setTimeout(() => runSearch(q), DEBOUNCE_MS);
  }

  async function runSearch(q: string) {
    const mine = ++seq;
    try {
      const page = await search(PROVIDER, q, '');
      if (mine !== seq) return;
      items = page.items ?? [];
      selected = items.length ? 0 : -1;
    } catch {
      if (mine !== seq) return;
      items = [];
      selected = -1;
      status = "Couldn't search The Finer Gifs Club.";
    }
  }

  async function onKeydown(e: KeyboardEvent) {
    // Enter that commits an IME composition (WebKit: keyCode 229) is not a submit.
    if (e.key !== 'Enter' || e.isComposing || e.keyCode === 229 || e.repeat) return;
    e.preventDefault();
    if (busy) return;
    busy = true;
    try {
      if (e.metaKey) {
        const r = await ActionService.Paste();
        status = r.message;
        needsAccessibility = r.needsAccessibility;
        keepStatus = r.copied && !r.pasted && !r.needsAccessibility;
      } else {
        status = (await ActionService.Copy()).message;
      }
    } catch (err) {
      status = String(err);
    } finally {
      busy = false;
    }
  }

  async function enableAutoPaste() {
    try {
      await ActionService.RequestAccessibility();
      status =
        'After turning Gifsmith on, press ⌘Enter again. If it still says it is not allowed, quit Gifsmith from the menu bar and open it again.';
      needsAccessibility = false;
    } catch (err) {
      status = String(err);
    }
    input.focus();
  }
</script>

<main>
  <header class="search">
    <svg class="icon" viewBox="0 0 16 16" aria-hidden="true">
      <circle cx="7" cy="7" r="4.5" />
      <path d="M10.5 10.5 14 14" />
    </svg>
    <input
      bind:this={input}
      bind:value={query}
      oninput={onInput}
      onkeydown={onKeydown}
      placeholder="Search The Finer Gifs Club"
      autocomplete="off"
      spellcheck="false"
      role="combobox"
      aria-expanded={items.length > 0}
      aria-controls="grid"
      aria-activedescendant={selected >= 0 ? `tile-${selected}` : undefined}
    />
  </header>

  <section class="content">
    <div class="grid" id="grid" role="listbox" aria-label="Results">
      {#each items as g, i (i)}
        <!-- Keyboard selection goes through the combobox input (aria-activedescendant), never tile focus. -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div
          class="tile"
          id="tile-{i}"
          role="option"
          tabindex="-1"
          aria-selected={i === selected}
          aria-label={g.title || `GIF ${i + 1}`}
          onclick={() => (selected = i)}
        >
          <img src={g.previewUrl} alt="" decoding="async" draggable="false" loading="lazy" />
        </div>
      {/each}
    </div>
  </section>

  <footer>
    <span class="attribution">GIFs from The Finer Gifs Club</span>
    <span class="center">{footerCenter}</span>
    {#if needsAccessibility}
      <button onclick={enableAutoPaste}>Enable Automatic Paste…</button>
    {:else}
      <span class="hints"><b>↵</b> Copy&nbsp;&nbsp;&nbsp;<b>⌘↵</b> Paste</span>
    {/if}
  </footer>
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .search {
    box-sizing: border-box;
    height: 48px;
    flex: none;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px;
    border-bottom: 1px solid var(--hairline);
  }
  .icon {
    width: 16px;
    height: 16px;
    flex: none;
    fill: none;
    stroke: var(--text-secondary);
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: 0;
    padding: 0;
    background: transparent;
    color: var(--text);
    font: var(--display);
    caret-color: var(--accent);
  }
  input::placeholder {
    color: var(--text-secondary);
  }
  .content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(4, 166px);
    grid-auto-rows: 144px;
    gap: 8px;
    padding: 16px;
  }
  .tile {
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
  .tile img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
  footer {
    box-sizing: border-box;
    height: 32px;
    flex: none;
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 0 16px;
    border-top: 1px solid var(--hairline);
    font: var(--label);
    color: var(--text-secondary);
  }
  .attribution,
  .hints {
    flex: none;
  }
  .center {
    flex: 1;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .hints {
    color: var(--text-tertiary);
  }
  .hints b {
    font-weight: 400;
    color: var(--text-secondary);
  }
  footer button {
    flex: none;
    font: var(--label);
  }
</style>
