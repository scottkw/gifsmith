<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import {
    cancel,
    copy,
    hide,
    latest,
    lastProvider,
    paste,
    providerError,
    providers,
    requestAccessibility,
    search,
    setLastProvider,
    trending,
    type Gif,
    type Page,
    type ProviderErr,
    type ProviderStatus,
  } from './lib/api';
  import { errorState, pagingMessage, providerName } from './lib/errors';
  import { move, nearEnd, routeKey } from './lib/grid';
  import { SLOTS, slotById, type SlotId } from './lib/providers';
  import Footer from './components/Footer.svelte';
  import ProviderTabs from './components/ProviderTabs.svelte';
  import ResultGrid from './components/ResultGrid.svelte';
  import StateMessage from './components/StateMessage.svelte';

  const DEBOUNCE_MS = 450;
  const HIDE_AFTER_COPY_MS = 400;
  const COPY_FAILED = "Couldn't copy this GIF. The download failed. Try again or pick another GIF.";
  const CLIPS_COPY = 'Copying Clips (MP4) arrives in the next update.';
  const AX_FAILED = "Couldn't ask macOS for Accessibility access. Try again.";
  const BUSY = 'another action is still running'; // Go ErrBusy: a second call while one runs is ignored

  type Status = 'idle' | 'loading' | 'refreshing' | 'ready' | 'empty' | 'error';
  type Paging = 'idle' | 'loading' | 'end' | 'error';
  // UI state is held per provider id, so one provider's failure never touches another's.
  type PState = {
    status: Status;
    q: string; // query of the last first-page request; '' = trending (S7 copy, Try Again)
    items: Gif[];
    next: string; // provider cursor; "" = no more pages
    seq: number; // bumped on every result-set replacement, so a late page is dropped, not appended
    more: ((cursor: string) => Promise<Page | undefined>) | null;
    paging: Paging;
    err: ProviderErr | null; // first-page error (S9–S12)
    errAt: number;
    pageErr: ProviderErr | null; // page 2+ error (S9a/S12a)
    pageErrAt: number;
  };
  const fresh = (): PState => ({
    status: 'idle',
    q: '',
    items: [],
    next: '',
    seq: 0,
    more: null,
    paging: 'idle',
    err: null,
    errAt: 0,
    pageErr: null,
    pageErrAt: 0,
  });
  const states = $state<Record<SlotId, PState>>({ giphy: fresh(), clips: fresh(), klipy: fresh(), finergifs: fresh() });
  let active = $state<SlotId>('giphy');
  // Key state comes only from GifService.Providers() (booleans); null until the first answer.
  let statuses = $state<Record<string, ProviderStatus> | null>(null);
  const cur = $derived(states[active]);
  const slot = $derived(slotById(active));
  const clipsEnabled = $derived(statuses?.clips?.clips === 'enabled');
  const giphyKey = $derived(!!statuses?.giphy?.configured);

  let input: HTMLInputElement;
  let query = $state('');
  let selected = $state(-1);
  let status = $state(''); // footer center status line
  let tone = $state<'' | 'success' | 'error'>('');
  let copying = $state(false); // one Copy/Paste in flight; a second Enter would paste twice
  let needsAccessibility = $state(false);
  // Paste hides the panel before ⌘V; a copied-only fallback message must survive to the next show.
  let keepStatus = false;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let now = $state(Date.now()); // clock for retry countdowns; set whenever an error arrives

  const footerCenter = $derived(status || cur.items[selected]?.title || '');
  const busy = $derived(cur.status === 'loading' || cur.status === 'refreshing');
  // S8: Finer Gifs has no trending list, so a short query shows a prompt and makes no request.
  const finerIdle = $derived(active === 'finergifs' && [...query.trim()].length < 2);
  // Key slots without a key wait for the setup card; Clips waits for its probe (plan 02-06).
  const needsKey = $derived(!!slot.keyProvider && statuses !== null && !statuses[active]?.configured);
  // Error copy comes only from errors.ts rows, never from the caught error's text.
  const errView = $derived(cur.status === 'error' && cur.err ? errorState(cur.err, now, cur.errAt) : null);
  const pageView = $derived(
    cur.paging === 'error' && cur.pageErr
      ? { note: pagingMessage(cur.pageErr, now, cur.pageErrAt), state: errorState(cur.pageErr, now, cur.pageErrAt) }
      : null,
  );
  const retryAt = $derived(errView?.retryAt ?? pageView?.state?.retryAt);

  // Rate-limit countdown: refresh the copy once a minute and exactly when the window ends.
  $effect(() => {
    const at = retryAt;
    if (!at) return;
    const tick = () => (now = Date.now());
    const iv = setInterval(tick, 60_000);
    const end = setTimeout(tick, Math.max(0, at - Date.now()) + 50);
    return () => {
      clearInterval(iv);
      clearTimeout(end);
    };
  });

  const byId = (list: ProviderStatus[]) => Object.fromEntries(list.map((p) => [p.id, p]));
  async function refreshProviders() {
    statuses = byId(await providers());
  }

  onMount(async () => {
    input.focus();
    Events.On('panel:show', () => {
      now = Date.now();
      if (!keepStatus) setStatus('');
      keepStatus = false;
      needsAccessibility = false;
      input.focus();
      input.select();
    });
    Events.On('panel:status', (e) => setStatus(String(e.data)));
    Events.On('providers:changed', refreshProviders);
    const [list, last] = await Promise.all([providers(), lastProvider()]);
    const st = byId(list);
    // Persisted slot; first run (or an unavailable Clips) opens on GIPHY. Set with statuses in one step.
    active = SLOTS.some((s) => s.id === last) && !(last === 'clips' && st.clips?.clips !== 'enabled') ? (last as SlotId) : 'giphy';
    statuses = st;
    load(active);
  });

  function setStatus(text: string, t: typeof tone = '') {
    status = text;
    tone = t;
  }

  // A transient status (copy result, error) clears on the next keystroke or selection change.
  function clearTransient() {
    if (!copying) setStatus('');
  }

  function select(i: number) {
    if (i !== selected) clearTransient();
    selected = i;
  }

  // The query a provider should show: 2+ code points search, anything shorter is trending ('').
  function wanted(): string {
    const q = query.trim();
    return [...q].length >= 2 ? q : '';
  }

  // No search while an IME composition is open; compositionend calls this again.
  function onInput(e?: Event) {
    if ((e as InputEvent | undefined)?.isComposing) return;
    clearTransient();
    clearTimeout(timer);
    if (!wanted()) show(active); // trending or S8 at once, no debounce
    else timer = setTimeout(() => load(active), DEBOUNCE_MS);
  }

  // Fetches p's first page for the current query: search, trending, or (Finer Gifs) nothing.
  function load(p: SlotId) {
    clearTimeout(timer);
    if (p === 'clips' ? !clipsEnabled : !!slotById(p).keyProvider && !statuses?.[p]?.configured) return;
    const q = wanted();
    cancel(`first:${p}`);
    if (!q && p === 'finergifs') setResults(p, null, null);
    else runFirst(p, q);
  }

  // Shows p's own state, fetching only when what it holds belongs to another query.
  function show(p: SlotId) {
    const s = states[p];
    if (s.err?.kind === 'key_rejected') return; // S13: the card stays until a new key is saved
    if (s.status !== 'idle' && s.q === wanted()) return;
    load(p);
  }

  function switchTo(p: SlotId) {
    if (p !== active) {
      active = p;
      clearTransient();
      selected = states[p].items.length ? 0 : -1;
      setLastProvider(p);
      show(p);
    }
    tick().then(() => input.focus());
  }

  // S2 (nothing shown yet) or S4 (previous results stay, spinner only) until the latest call lands.
  async function runFirst(p: SlotId, q: string) {
    const s = states[p];
    const call = (cursor: string) => (q ? search(p, q, cursor) : trending(p, cursor));
    s.q = q;
    s.status = s.items.length ? 'refreshing' : 'loading';
    try {
      const page = await latest(`first:${p}`, call(''));
      if (!page) return; // cancelled or superseded: the newer call owns the state
      setResults(p, page, (cursor) => latest(`page:${p}`, call(cursor)));
    } catch (e) {
      setResults(p, null, null);
      s.status = 'error';
      s.err = providerError(e) ?? { kind: 'http', provider: p };
      now = s.errAt = Date.now();
    }
  }

  // A new result set (or none): first tile selected, paging restarts.
  function setResults(p: SlotId, page: Page | null, more: PState['more']) {
    const s = states[p];
    cancel(`page:${p}`);
    s.seq++;
    s.items = page?.items ?? [];
    s.next = page?.next ?? '';
    s.more = more;
    s.paging = page && !s.next ? 'end' : 'idle';
    s.pageErr = null;
    s.err = null;
    s.status = !page ? 'idle' : s.items.length ? 'ready' : 'empty';
    if (p === active) selected = s.items.length ? 0 : -1;
  }

  function retry() {
    runFirst(active, cur.q);
  }

  // One page request at a time; a failed page keeps loaded results until Try Again.
  async function loadMore() {
    const s = states[active];
    if (s.paging !== 'idle' || !s.next || !s.more) return;
    const mine = s.seq;
    s.paging = 'loading';
    try {
      const page = await s.more(s.next);
      if (!page || mine !== s.seq) return;
      s.items.push(...(page.items ?? []));
      s.next = page.next;
      s.paging = s.next ? 'idle' : 'end';
    } catch (e) {
      if (mine !== s.seq) return;
      s.paging = 'error';
      s.pageErr = providerError(e);
      now = s.pageErrAt = Date.now();
    }
  }

  function retryPage() {
    cur.paging = 'idle';
    loadMore();
  }

  // Every routed key is preventDefault-ed: arrows must not move the caret, and an Esc the page
  // lets through would reach the native panel's cancelOperation.
  function onKeydown(e: KeyboardEvent) {
    const a = routeKey(e);
    if (a === 'none') return;
    e.preventDefault();
    switch (a) {
      case 'left':
      case 'right':
      case 'up':
      case 'down':
        if (!cur.items.length) return;
        select(move(selected, a, cur.items.length));
        if (nearEnd(selected, cur.items.length)) loadMore();
        return;
      case 'copy':
      case 'paste':
        // ↵ with nothing selected runs a visible Try Again.
        if (selected < 0) {
          if (a === 'copy' && errView?.tryAgain) retry();
          return;
        }
        run(a);
        return;
      case 'escape':
        if (query) {
          query = '';
          onInput();
        } else hide(); // no-op in the server build; nothing here depends on it
        return;
      // 'actions' (⌘K, Phase 3 action menu) and 'tab': bound, no effect yet; slot1–slot4: onMainKeydown.
    }
  }

  // ⌘1–⌘4 from any field in the panel (search input or key field) switch provider.
  function onMainKeydown(e: KeyboardEvent) {
    const a = routeKey(e);
    if (!a.startsWith('slot')) return;
    e.preventDefault();
    switchTo(SLOTS[Number(a.slice(4)) - 1].id);
  }

  async function run(action: 'copy' | 'paste') {
    const g = cur.items[selected];
    if (copying || !g) return;
    if (g.isClip) {
      setStatus(CLIPS_COPY);
      return;
    }
    copying = true;
    setStatus('Copying…');
    try {
      if (action === 'copy') {
        const r = await copy(g);
        setStatus(r.message, 'success');
        setTimeout(hide, HIDE_AFTER_COPY_MS);
      } else {
        const r = await paste(g);
        setStatus(r.message);
        needsAccessibility = r.needsAccessibility;
        keepStatus = r.copied && !r.pasted && !r.needsAccessibility;
      }
    } catch (err) {
      // Compared, never rendered: Go's ErrBusy is the one error that is silently ignored.
      if ((err as Error)?.message === BUSY) setStatus('');
      else setStatus(COPY_FAILED, 'error');
    } finally {
      copying = false;
    }
  }

  async function enableAutoPaste() {
    try {
      await requestAccessibility();
      setStatus(
        'After turning Gifsmith on, press ⌘Enter again. If it still says it is not allowed, quit Gifsmith from the menu bar and open it again.',
      );
      needsAccessibility = false;
    } catch {
      setStatus(AX_FAILED, 'error');
    }
    input.focus();
  }
</script>

<!-- Clicks anywhere except a text field keep focus in the search input (the grid never takes focus). -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<main onmousedown={(e) => e.target instanceof HTMLInputElement || e.preventDefault()} onkeydown={onMainKeydown}>
  <header class="search">
    <svg class="icon" viewBox="0 0 16 16" aria-hidden="true">
      <circle cx="7" cy="7" r="4.5" />
      <path d="M10.5 10.5 14 14" />
    </svg>
    <input
      bind:this={input}
      bind:value={query}
      oninput={onInput}
      oncompositionend={() => onInput()}
      onkeydown={onKeydown}
      placeholder={slot.placeholder}
      autocomplete="off"
      spellcheck="false"
      role="combobox"
      aria-expanded={cur.items.length > 0}
      aria-controls="grid"
      aria-activedescendant={selected >= 0 ? `tile-${selected}` : undefined}
    />
    {#if busy}
      <span class="spinner busy" role="status" aria-label="Loading"></span>
    {/if}
  </header>

  <ProviderTabs {active} {clipsEnabled} onswitch={switchTo} />

  <section class="content" aria-busy={statuses === null}>
    {#if statuses === null || needsKey}
      <!-- needsKey: the setup card (Task 2) -->
    {:else if active === 'clips' && !giphyKey}
      <StateMessage
        heading="Clips use your GIPHY key"
        body="Add a GIPHY key first. Gifsmith checks it for Clips access automatically."
        button="Add GIPHY Key"
        primary
        onbutton={() => switchTo('giphy')}
      />
    {:else if active === 'clips' && !clipsEnabled}
      <div class="skeleton" aria-hidden="true">
        {#each Array(8), i (i)}
          <div></div>
        {/each}
      </div>
    {:else if finerIdle}
      <StateMessage
        heading="Search The Office GIFs"
        body="The Finer Gifs Club has no trending list. Type at least 2 characters to search."
      />
    {:else if cur.items.length}
      {#key `${active}:${cur.seq}`}
        <ResultGrid
          items={cur.items}
          {selected}
          {copying}
          paging={cur.paging}
          pagingNote={pageView?.note}
          pagingRetry={pageView?.state?.tryAgain ?? true}
          onselect={select}
          oncopy={(i) => {
            select(i);
            run('copy');
          }}
          onneedmore={loadMore}
          onretrypage={retryPage}
        />
      {/key}
    {:else if cur.status === 'loading'}
      <div class="skeleton" aria-hidden="true">
        {#each Array(8), i (i)}
          <div></div>
        {/each}
      </div>
    {:else if cur.status === 'empty'}
      <StateMessage
        heading="No GIFs found"
        body="Nothing on {providerName(active)} matches “{cur.q}”. Try different words, or switch providers with ⌘1–⌘4."
      />
    {:else if errView}
      <StateMessage
        icon
        heading={errView.heading}
        body={errView.body}
        button={errView.tryAgain ? 'Try Again' : ''}
        onbutton={retry}
      />
    {/if}
  </section>

  <Footer attribution={slot.attribution} text={footerCenter} {tone} {needsAccessibility} onenable={enableAutoPaste} />
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
  .busy {
    flex: none;
    color: var(--text-secondary);
  }
  .content {
    flex: 1;
    min-height: 0;
  }
  /* S2: 8 skeleton tiles (2 rows) in grid geometry. */
  .skeleton {
    display: grid;
    grid-template-columns: repeat(4, 166px);
    grid-auto-rows: 144px;
    gap: 8px;
    padding: 16px;
  }
  .skeleton div {
    border-radius: 8px;
    background: var(--surface);
    animation: pulse 1.2s ease-in-out infinite;
  }
  @keyframes pulse {
    50% {
      opacity: 0.6;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .skeleton div {
      animation: none;
    }
  }
</style>
