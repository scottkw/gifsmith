<script lang="ts">
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import {
    cancel,
    copy,
    hide,
    latest,
    paste,
    providerError,
    requestAccessibility,
    search,
    type Gif,
    type Page,
    type ProviderErr,
  } from './lib/api';
  import { errorState, pagingMessage, providerName } from './lib/errors';
  import { move, nearEnd, routeKey } from './lib/grid';
  import ResultGrid from './components/ResultGrid.svelte';
  import StateMessage from './components/StateMessage.svelte';

  const PROVIDER = 'finergifs'; // the active provider; the tabs in plan 02-05 make it switchable
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
    q: string; // query of the last first-page request (S7 copy, Try Again)
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
  const states = $state<Record<string, PState>>({ [PROVIDER]: fresh() });
  const cur = $derived(states[PROVIDER]);

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
  const finerIdle = $derived(PROVIDER === 'finergifs' && [...query.trim()].length < 2);
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

  onMount(() => {
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

  // No search while an IME composition is open; compositionend calls this again.
  function onInput(e?: Event) {
    if ((e as InputEvent | undefined)?.isComposing) return;
    clearTransient();
    clearTimeout(timer);
    const q = query.trim();
    if ([...q].length < 2) {
      cancel(`first:${PROVIDER}`);
      setResults(PROVIDER, null, null);
      return;
    }
    timer = setTimeout(() => runSearch(PROVIDER, q), DEBOUNCE_MS);
  }

  // S2 (nothing shown yet) or S4 (previous results stay, spinner only) until the latest call lands.
  async function runSearch(p: string, q: string) {
    const s = states[p];
    s.q = q;
    s.status = s.items.length ? 'refreshing' : 'loading';
    try {
      const page = await latest(`first:${p}`, search(p, q, ''));
      if (!page) return; // cancelled or superseded: the newer call owns the state
      setResults(p, page, (cursor) => latest(`page:${p}`, search(p, q, cursor)));
    } catch (e) {
      setResults(p, null, null);
      s.status = 'error';
      s.err = providerError(e) ?? { kind: 'http', provider: p };
      now = s.errAt = Date.now();
    }
  }

  // A new result set (or none): first tile selected, paging restarts.
  function setResults(p: string, page: Page | null, more: PState['more']) {
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
    if (p === PROVIDER) selected = s.items.length ? 0 : -1;
  }

  function retry() {
    if (cur.q) runSearch(PROVIDER, cur.q);
  }

  // One page request at a time; a failed page keeps loaded results until Try Again.
  async function loadMore() {
    const s = states[PROVIDER];
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
      // 'actions' (⌘K, Phase 3 action menu) and 'tab': bound, no effect yet.
    }
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
<main onmousedown={(e) => e.target instanceof HTMLInputElement || e.preventDefault()}>
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
      placeholder="Search The Finer Gifs Club"
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

  <section class="content">
    {#if finerIdle}
      <StateMessage
        heading="Search The Office GIFs"
        body="The Finer Gifs Club has no trending list. Type at least 2 characters to search."
      />
    {:else if cur.items.length}
      {#key cur.seq}
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
        body="Nothing on {providerName(PROVIDER)} matches “{cur.q}”. Try different words, or switch providers with ⌘1–⌘4."
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

  <footer>
    <span class="attribution">GIFs from The Finer Gifs Club</span>
    <span class="center {tone}">
      {#if tone === 'success'}
        <svg class="check" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>
      {/if}
      {footerCenter}
    </span>
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
  .center.success {
    color: var(--success);
  }
  .center.error {
    color: var(--destructive);
  }
  .check {
    width: 12px;
    height: 12px;
    vertical-align: -2px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
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
