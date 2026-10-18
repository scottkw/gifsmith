<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import {
    actionError,
    cancel,
    checkClips,
    copy,
    copyText,
    openPage,
    hide,
    latest,
    lastProvider,
    openExternal,
    paste,
    pasteText,
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
    type Result,
    type TextKind,
  } from './lib/api';
  import { actionErrorText, errorState, pagingMessage, providerName, revokedMessage } from './lib/errors';
  import { buildActions, type ActionId, type ActionRow } from './lib/actions';
  import { COLS, move, nearEnd, routeKey } from './lib/grid';
  import { placeMenu } from './lib/popover';
  import { CLIPS_ACCESS_MAILTO, SLOTS, slotById, type SlotId } from './lib/providers';
  import ActionMenu from './components/ActionMenu.svelte';
  import Footer from './components/Footer.svelte';
  import ProviderTabs from './components/ProviderTabs.svelte';
  import ResultGrid from './components/ResultGrid.svelte';
  import SetupCard from './components/SetupCard.svelte';
  import StateMessage from './components/StateMessage.svelte';

  const DEBOUNCE_MS = 450;
  const HIDE_AFTER_COPY_MS = 400;
  const KEY_SAVED_MS = 2000;
  const AX_FAILED = "Couldn't ask macOS for Accessibility access. Try again.";

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
  // Clips probe state from Go: no_key | checking | enabled | disabled | unknown.
  const clipsState = $derived(statuses?.clips?.clips ?? 'no_key');
  const clipsEnabled = $derived(clipsState === 'enabled');
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
  let hideTimer: ReturnType<typeof setTimeout> | undefined; // pending hide after a copy
  let now = $state(Date.now()); // clock for retry countdowns; set whenever an error arrives
  let menuOpen = $state(false); // ⌘K action menu
  let menuRows = $state<ActionRow[]>([]);
  let menuAt = $state({ left: 0, top: 0 });
  let menu = $state<ActionMenu>();

  const footerCenter = $derived(status || cur.items[selected]?.title || '');
  const busy = $derived(cur.status === 'loading' || cur.status === 'refreshing');
  // S8: Finer Gifs has no trending list, so a short query shows a prompt and makes no request.
  const finerIdle = $derived(active === 'finergifs' && [...query.trim()].length < 2);
  // S1: a key slot without a key; S13: its stored key was rejected (key_missing: Go lost it meanwhile).
  const card = $derived(
    !!slot.keyProvider &&
      statuses !== null &&
      (!statuses[active]?.configured || cur.err?.kind === 'key_rejected' || cur.err?.kind === 'key_missing'),
  );
  const cardMessage = $derived(cur.err?.kind === 'key_rejected' ? revokedMessage(slot.keyLabel!) : '');
  let setup = $state<SetupCard>();
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
    // The Clips probe settled while Clips is on screen: its grid, or one re-probe for an unknown result.
    if (active === 'clips' && clipsEnabled) show('clips');
    else if (active === 'clips' && clipsState === 'unknown' && states.clips.status !== 'error') recheckClips();
  }

  // Choosing Clips while its result is unknown re-probes once (S2 skeleton meanwhile);
  // a repeat failure shows that error for "GIPHY Clips". Enabled loads via refreshProviders.
  let rechecking = false;
  async function recheckClips() {
    if (rechecking) return;
    rechecking = true;
    const s = states.clips;
    s.status = 'idle';
    s.err = null;
    try {
      if ((await checkClips()) === 'enabled') await refreshProviders();
    } catch (e) {
      s.status = 'error';
      s.err = providerError(e) ?? { kind: 'http', provider: 'clips' };
      now = s.errAt = Date.now();
    } finally {
      rechecking = false;
    }
  }

  onMount(async () => {
    input.focus();
    Events.On('panel:show', () => {
      clearTimeout(hideTimer); // a copy just before this show must not hide the panel again (02 IN-07)
      menuOpen = false; // never reopens on its own
      now = Date.now();
      if (!keepStatus) setStatus('');
      keepStatus = false;
      needsAccessibility = false;
      if (card) setup?.focus();
      else {
        input.focus();
        input.select();
      }
    });
    Events.On('panel:status', (e) => setStatus(String(e.data)));
    Events.On('providers:changed', refreshProviders);
    const [list, last] = await Promise.all([providers(), lastProvider()]);
    const st = byId(list);
    // Persisted slot; first run (or Clips without a GIPHY key) opens on GIPHY. Set with statuses in one step.
    active = SLOTS.some((s) => s.id === last) && !(last === 'clips' && !st.giphy?.configured) ? (last as SlotId) : 'giphy';
    statuses = st;
    if (active === 'clips' && !clipsEnabled) bootClips();
    else load(active);
  });

  // The Clips probe only means something after a GIPHY GIFs success with the key. Fetch GIPHY
  // trending (usually cached) behind the S2 skeleton to give it one, then re-probe once if that
  // did not start a probe; refreshProviders shows Clips when the probe lands.
  async function bootClips() {
    const s = states.clips;
    s.status = 'idle';
    s.err = null;
    try {
      await trending('giphy', '');
    } catch (e) {
      if (active !== 'clips' || clipsEnabled) return;
      // No probe will run: show GIPHY's error here (a rejected key gets the Go to GIPHY message).
      s.status = 'error';
      s.err = providerError(e) ?? { kind: 'http', provider: 'giphy' };
      now = s.errAt = Date.now();
      return;
    }
    if (active === 'clips' && !clipsEnabled) recheckClips();
  }

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
    menuOpen = false;
    if (p !== active) {
      active = p;
      clearTransient();
      selected = states[p].items.length ? 0 : -1;
      setLastProvider(p);
      if (p === 'clips' && clipsState === 'unknown') bootClips();
      else show(p);
    }
    tick().then(() => card || input.focus()); // a setup card focuses its own key field
  }

  // S1c: the card gives way at once to that provider's trending page (seeded by verification).
  async function keySaved(p: 'giphy' | 'klipy') {
    await refreshProviders();
    if (p === active) query = '';
    if (p === 'giphy') setResults('clips', null, null); // Clips use this key: drop the old key's results and errors
    setResults(p, null, null);
    runFirst(p, '');
    const text = `${slotById(p).keyLabel} key saved`;
    setStatus(text, 'success');
    setTimeout(() => status === text && setStatus(''), KEY_SAVED_MS);
    await tick();
    if (!card) input.focus();
  }

  // "Try Finer Gifs (no key)": the keyless provider with an empty query (S8), remembered.
  function tryFiner() {
    query = '';
    switchTo('finergifs');
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
    if (p === active) {
      selected = s.items.length ? 0 : -1;
      menuOpen = false; // its item is gone
    }
  }

  function retry() {
    if (active === 'clips' && !clipsEnabled) bootClips(); // GIPHY first: it may be what failed
    else runFirst(active, cur.q);
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
        runAction(a);
        return;
      case 'copyLink':
      case 'copyMarkdown':
      case 'open':
        runAction(a);
        return;
      case 'actions':
        toggleMenu();
        return;
      case 'escape':
        if (query) {
          query = '';
          onInput();
        } else hide(); // no-op in the server build; nothing here depends on it
        return;
      // 'tab': bound, no effect; slot1–slot4: onMainKeydown.
    }
  }

  // ⌘1–⌘4 from any field in the panel (search input or key field) switch provider.
  function onMainKeydown(e: KeyboardEvent) {
    const a = routeKey(e);
    if (a.startsWith('slot')) {
      e.preventDefault();
      switchTo(SLOTS[Number(a.slice(4)) - 1].id);
      // The menu's filter has focus while it is open: its ⌘ shortcuts (⌘K, ⌘↵, ⌘⇧C, …) arrive here.
    } else if (menuOpen && e.metaKey && e.target !== input) onKeydown(e);
  }

  // ⌘K: open beside the selected tile, or close. No tile on screen (empty, loading, message, setup card): nothing.
  async function toggleMenu() {
    if (menuOpen) return closeMenu();
    const g = cur.items[selected];
    const tile = document.getElementById(`tile-${selected}`);
    if (!g || !tile) return;
    menuRows = buildActions({ gif: g, tab: 'provider', inDetails: false, favorited: false });
    menuOpen = true;
    await tick(); // measure the rendered menu, then place it before the browser paints
    menuAt = placeMenu(tile.getBoundingClientRect(), selected % COLS, menu!.height(), { w: innerWidth, h: innerHeight });
    menu!.focus();
  }

  function closeMenu() {
    menuOpen = false;
    input.focus();
  }

  // The one dispatcher for the menu and every shortcut. The menu closes before anything runs.
  function runAction(id: ActionId) {
    if (menuOpen) closeMenu();
    const g = cur.items[selected];
    switch (id) {
      case 'copy':
      case 'paste':
        run(id);
        return;
      case 'copyLink':
        textAction('link');
        return;
      case 'pasteLink':
        textAction('link', true);
        return;
      case 'copyMarkdown':
        textAction('markdown');
        return;
      case 'pasteMarkdown':
        textAction('markdown', true);
        return;
      case 'copyPage':
        textAction('page');
        return;
      case 'open':
        // Only GIPHY and Clips items carry a page URL; for the rest ⌘O does nothing.
        if (g?.pageUrl) openItem(g);
        return;
    }
  }

  function run(action: 'copy' | 'paste') {
    act(action === 'copy' ? copy : paste, action === 'paste');
  }

  // Go builds the text from the selected item; the webview only names the kind.
  function textAction(kind: TextKind, pasteIt = false) {
    if (pasteIt && kind !== 'page') act((g) => pasteText(g, kind), true);
    else act((g) => copyText(g, kind), false); // there is no Paste Page Link
  }

  // One copy or paste of the selected item at a time. A copy confirms in the footer and hides the
  // panel; a paste has hidden it already, unless it fell back to copying.
  async function act(call: (g: Gif) => Promise<Result>, pasteIt: boolean) {
    const g = cur.items[selected];
    if (copying || !g) return;
    copying = true;
    setStatus('Copying…');
    try {
      const r = await call(g);
      if (!pasteIt) {
        setStatus(r.message, 'success');
        scheduleHide();
      } else {
        setStatus(r.message, r.pasted ? 'success' : '');
        needsAccessibility = r.needsAccessibility;
        keepStatus = r.copied && !r.pasted && !r.needsAccessibility;
      }
    } catch (err) {
      actionFailed(err, g);
    } finally {
      copying = false;
    }
  }

  // Copy actions confirm in the footer, then hide the panel; panel:show cancels a pending hide.
  function scheduleHide() {
    clearTimeout(hideTimer);
    hideTimer = setTimeout(hide, HIDE_AFTER_COPY_MS);
  }

  // Fixed copy keyed by Go's error kind, never the error's text. A busy refusal is silently ignored.
  function actionFailed(err: unknown, g: Gif) {
    const kind = actionError(err)?.kind ?? '';
    if (kind === 'busy') setStatus('');
    else setStatus(actionErrorText(kind, g), 'error');
  }

  // Success shows no toast: the browser takes focus and HideOnFocusLost hides the panel.
  async function openItem(g: Gif) {
    try {
      await openPage(g);
    } catch (err) {
      actionFailed(err, g);
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
<!-- A mousedown outside the open action menu closes it without running anything. -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<main
  onmousedown={(e) => {
    if (menuOpen && !(e.target as Element).closest('[role=dialog]')) closeMenu();
    if (!(e.target instanceof HTMLInputElement)) e.preventDefault();
  }}
  onkeydown={onMainKeydown}
>
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
      disabled={card}
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
    {#if statuses === null}
      <!-- waiting for Providers() -->
    {:else if card}
      {#key `${active}:${cardMessage}`}
        <SetupCard bind:this={setup} {slot} initialMessage={cardMessage} onsaved={keySaved} ontryfiner={tryFiner} />
      {/key}
    {:else if active === 'clips' && !giphyKey}
      <StateMessage
        heading="Clips use your GIPHY key"
        body="Add a GIPHY key first. Gifsmith checks it for Clips access automatically."
        button="Add GIPHY Key"
        primary
        onbutton={() => switchTo('giphy')}
      />
    {:else if active === 'clips' && (cur.err?.kind === 'key_rejected' || cur.err?.kind === 'key_missing')}
      <!-- Clips have no key card of their own: errorState() has no row for key errors. -->
      <StateMessage
        heading="GIPHY rejected your saved key"
        body="Clips use your GIPHY key. Paste a new key on the GIPHY tab to continue."
        button="Go to GIPHY"
        primary
        onbutton={() => {
          // GIPHY's own grid may still hold cached results for this key: carry the rejection over so its card shows.
          setResults('giphy', null, null);
          Object.assign(states.giphy, { status: 'error', err: { kind: 'key_rejected', provider: 'giphy' }, errAt: Date.now() });
          switchTo('giphy');
        }}
      />
    {:else if active === 'clips' && clipsState === 'disabled'}
      <StateMessage
        heading="Clips aren't enabled for your key"
        body="Your GIPHY key works for GIFs, but GIPHY has to approve it for Clips. Email clips@giphy.com to request access, then reopen Gifsmith."
        link="Email clips@giphy.com ↗"
        href={CLIPS_ACCESS_MAILTO}
        onlink={() => openExternal(CLIPS_ACCESS_MAILTO)}
        button="Back to GIPHY"
        onbutton={() => switchTo('giphy')}
      />
    {:else if active === 'clips' && !clipsEnabled && cur.status !== 'error'}
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

  {#if menuOpen}
    <ActionMenu bind:this={menu} rows={menuRows} {...menuAt} onrun={runAction} onclose={closeMenu} />
  {/if}

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
