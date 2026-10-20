<script lang="ts">
  import { onMount, tick, untrack } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import {
    actionError,
    cancel,
    checkClips,
    clearRecents,
    copy,
    copySquare,
    copyText,
    download,
    openPage,
    openSettings,
    openSource,
    hide,
    latest,
    lastProvider,
    loadLibrary,
    openExternal,
    paste,
    pasteSquare,
    pasteText,
    providerError,
    providers,
    removeRecent,
    removeSearch,
    requestAccessibility,
    saveSearch,
    search,
    setFavorite,
    setLastProvider,
    showDownload,
    trending,
    undoLibrary,
    type Gif,
    type Library,
    type Page,
    type ProviderErr,
    type ProviderStatus,
    type Result,
    type SettingsSnapshot,
    type TextKind,
  } from './lib/api';
  import { actionErrorText, errorState, libraryErrorText, pagingMessage, providerName, revokedMessage } from './lib/errors';
  import { buildActions, type ActionId, type ActionRow } from './lib/actions';
  import { COLS, move, nearEnd, routeKey } from './lib/grid';
  import { flatItems, moveSectioned, sections } from './lib/library';
  import { placeInDetails, placeMenu } from './lib/popover';
  import { CHIP_MAX, SAVE_IDLE_MS, chipsVisible, shouldSave } from './lib/searches';
  import { applyTheme } from './lib/theme';
  import { CLIPS_ACCESS_MAILTO, LIBRARY_TABS, SLOTS, libraryTab, slotById, type LibraryId, type SlotId } from './lib/providers';
  import ActionMenu from './components/ActionMenu.svelte';
  import DetailsPane from './components/DetailsPane.svelte';
  import LibraryGrid from './components/LibraryGrid.svelte';
  import Footer from './components/Footer.svelte';
  import ProviderTabs from './components/ProviderTabs.svelte';
  import RecentSearches from './components/RecentSearches.svelte';
  import ResultGrid from './components/ResultGrid.svelte';
  import SetupCard from './components/SetupCard.svelte';
  import StateMessage from './components/StateMessage.svelte';

  const DEBOUNCE_MS = 450;
  const HIDE_AFTER_COPY_MS = 400;
  const KEY_SAVED_MS = 2000;
  const DOWNLOAD_TOAST_MS = 5000;
  const CROP_BUSY_MS = 300; // a square crop shows its busy state only if it is still running then (D-16)
  const LIBRARY_SKELETON_MS = 300; // a library read slower than this shows the S2 skeleton
  const FAVORITE_TOAST_MS = 2000;
  const FAVORITES_MAX = 1000; // library.FavoritesMax
  const UNDO_TOAST_MS = 5000;
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
    stale: boolean; // fetched with old provider parameters: refetch the first page when next shown
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
    stale: false,
  });
  const states = $state<Record<SlotId, PState>>({ giphy: fresh(), clips: fresh(), klipy: fresh(), finergifs: fresh() });
  let { initial }: { initial: SettingsSnapshot } = $props();
  // Provider parameters of the last snapshot; a change marks every provider tab stale.
  const paramsOf = (s: SettingsSnapshot) => [s.giphyRating, s.pageSize, s.giphyLang, s.klipyCountry].join('|');
  let params = untrack(() => paramsOf(initial));
  function onSettings(s: SettingsSnapshot) {
    applyTheme(s.theme);
    if (paramsOf(s) === params) return;
    params = paramsOf(s);
    for (const st of Object.values(states)) st.stale = st.status !== 'idle'; // results on screen stay until then
  }
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
  // A toast may offer one action (link + shortcut hint); it goes whenever the status changes.
  type ToastAction = { label: string; hint: string; run: () => void };
  let toastAction = $state<ToastAction | null>(null);
  let toastTimer: ReturnType<typeof setTimeout> | undefined;
  let copying = $state(false); // one Copy/Paste in flight; a second Enter would paste twice
  let overlay = $state(false); // the selected tile's busy overlay; for a crop it waits CROP_BUSY_MS
  let needsAccessibility = $state(false);
  let downloadFolder: string | undefined; // display name from the last download, for F4
  // Paste hides the panel before ⌘V; a copied-only fallback message must survive to the next show.
  let keepStatus = false;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let hideTimer: ReturnType<typeof setTimeout> | undefined; // pending hide after a copy
  let now = $state(Date.now()); // clock for retry countdowns; set whenever an error arrives
  let menuOpen = $state(false); // ⌘K action menu
  let menuRows = $state<ActionRow[]>([]);
  let menuAt = $state({ left: 0, top: 0 });
  let menu = $state<ActionMenu>();
  // ⌘I Details shows the selected item of the current grid over the content area (D-10).
  let detailsOpen = $state(false);

  // Favorites / Recents (⌘5 / ⌘6) replace the provider grid while libTab is set; not persisted (D-22).
  let libTab = $state<LibraryId | null>(null);
  let library = $state<Library | null>(null); // null until the first load answers
  let libSlow = $state(false); // the load has not answered within LIBRARY_SKELETON_MS
  let libSeq = 0;
  const keyOf = (g: Gif) => `${g.provider}:${g.id}`;
  const noop = () => {};
  // The query filters library titles locally: no request, no debounce (D-06).
  const libSections = $derived(libTab && library ? sections(library[libTab].map((e) => e.gif), query) : []);
  const libItems = $derived(flatItems(libSections));
  const favKeys = $derived(new Set(library?.favorites.map((e) => keyOf(e.gif)) ?? []));
  // What the selection indexes: the library's flat item list, or the provider grid.
  const items = $derived(libTab ? libItems : cur.items);
  const libName = $derived(libTab ? libraryTab(libTab).label : '');
  const detail = $derived(detailsOpen ? items[selected] : undefined);

  const footerCenter = $derived(status || items[selected]?.title || '');
  const busy = $derived(!libTab && (cur.status === 'loading' || cur.status === 'refreshing'));
  // Library attribution follows the selected item's provider; nothing selected, nothing shown.
  const attribution = $derived(
    !libTab ? slot.attribution : items[selected] ? slotById(items[selected].provider).attribution : '',
  );
  // S8: Finer Gifs has no trending list, so a short query shows a prompt and makes no request.
  const finerIdle = $derived(active === 'finergifs' && [...query.trim()].length < 2);
  // S1: a key slot without a key; S13: its stored key was rejected (key_missing: Go lost it meanwhile).
  const card = $derived(
    !libTab &&
      !!slot.keyProvider &&
      statuses !== null &&
      (!statuses[active]?.configured || cur.err?.kind === 'key_rejected' || cur.err?.kind === 'key_missing'),
  );
  // Recent-search chips (D-12): provider tab, empty box, no setup or Clips card. chip = keyboard selection.
  const chips = $derived(
    chipsVisible({
      tab: libTab ?? 'provider',
      query,
      card: card || (active === 'clips' && !clipsEnabled),
      searches: library?.searches[active],
    })
      ? library!.searches[active].slice(0, CHIP_MAX)
      : [],
  );
  let chip = $state(-1);
  const gridShown = $derived(!finerIdle && cur.items.length > 0);
  // D-13/D-20 idle save: the time of the last keystroke and the last provider:query saved by it.
  let lastKeyAt = 0;
  let saveTimer: ReturnType<typeof setTimeout> | undefined;
  let savedKey = '';
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
    const before = clipsState;
    statuses = byId(await providers());
    // IN-04: Clips just became enabled, so an error stored before (e.g. GIPHY offline at boot) is stale.
    // Only on enabled: clearing on 'unknown' would re-probe after every failed probe.
    if (clipsEnabled && clipsState !== before && states.clips.status === 'error') setResults('clips', null, null);
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
      menuOpen = detailsOpen = false; // never reopen on their own
      chip = -1;
      now = Date.now();
      // A toast with an action (the download toast) outlives a hide inside its 5 s: a TCC prompt can hide the panel.
      if (!keepStatus && !toastAction) setStatus('');
      keepStatus = false;
      needsAccessibility = false;
      refreshLibrary(); // a copy or paste while hidden may have changed Recents
      if (!libTab && states[active].stale) show(active);
      if (card) setup?.focus();
      else {
        input.focus();
        input.select();
      }
    });
    Events.On('panel:status', (e) => setStatus(String(e.data)));
    Events.On('providers:changed', refreshProviders);
    Events.On('settings:changed', (e) => onSettings(e.data as SettingsSnapshot));
    // Open Favorites / Open Recents hotkeys: show that tab, or hide when it was already showing.
    Events.On('panel:open-tab', (e) => {
      const { tab, wasVisible } = e.data as { tab: LibraryId; wasVisible: boolean };
      if (wasVisible && libTab === tab) hide();
      else switchLib(tab);
    });
    const [list, last] = await Promise.all([providers(), lastProvider()]);
    const st = byId(list);
    // Persisted slot; first run (or Clips without a GIPHY key) opens on GIPHY. Set with statuses in one step.
    active = SLOTS.some((s) => s.id === last) && !(last === 'clips' && !st.giphy?.configured) ? (last as SlotId) : 'giphy';
    statuses = st;
    if (active === 'clips' && !clipsEnabled) bootClips();
    else load(active);
    refreshLibrary();
  });

  // Reads the saved lists (local files only). The selected item stays selected when it is still there.
  async function refreshLibrary(want?: string) {
    const seq = ++libSeq;
    const slow = setTimeout(() => (libSlow = true), LIBRARY_SKELETON_MS);
    try {
      const snap = await loadLibrary();
      if (seq === libSeq) applyLibrary(snap, want);
    } catch {
      // Go never fails Load (unreadable files are named in the snapshot); a dead bridge shows the load error.
      if (seq === libSeq && !library) library = { favorites: [], recents: [], searches: {}, unreadable: ['favorites', 'recents'], canUndo: false };
    } finally {
      clearTimeout(slow);
      if (seq === libSeq) libSlow = false;
    }
  }

  // Applies a snapshot and reselects `want` (default: the selected item); if it is gone, the
  // selection clamps, which lands on the next item in flat order, else the previous one.
  function applyLibrary(snap: Library, want?: string) {
    const prev = want ?? (libTab && items[selected] ? keyOf(items[selected]) : '');
    library = snap;
    if (!libTab) return;
    const i = libItems.findIndex((g) => keyOf(g) === prev);
    selected = i >= 0 ? i : Math.min(Math.max(selected, 0), libItems.length - 1);
    if (!libItems[selected]) detailsOpen = false;
  }

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

  // A later status replaces an earlier one, its action and its expiry; ms > 0 clears it after ms.
  function setStatus(text: string, t: typeof tone = '', action: ToastAction | null = null, ms = 0) {
    status = text;
    tone = t;
    toastAction = action;
    clearTimeout(toastTimer);
    if (ms) toastTimer = setTimeout(() => setStatus(''), ms);
  }

  // Running a toast's action (click or its shortcut) clears the toast.
  function runToastAction() {
    const a = toastAction;
    setStatus('');
    a?.run();
  }

  // A transient status (copy result, error) clears on the next keystroke or selection change.
  // An Undo toast stays for its 5 s: the removal itself moved the selection (D-08).
  function clearTransient() {
    if (!copying && !undoVisible()) setStatus('');
  }

  const undoVisible = () => toastAction?.label === 'Undo';

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
    chip = -1;
    lastKeyAt = Date.now();
    savedKey = '';
    clearTimeout(saveTimer);
    if (libTab) {
      selected = libItems.length ? 0 : -1;
      return;
    }
    if (!wanted()) show(active); // trending or S8 at once, no debounce
    else {
      timer = setTimeout(() => load(active), DEBOUNCE_MS);
      saveTimer = setTimeout(() => maybeSave(active), SAVE_IDLE_MS + 10);
    }
  }

  // Saves the current query for p once 1.5 s have passed since the last keystroke and its first page
  // has landed with a result, whichever comes last (D-13, D-20). Library filters are never saved.
  function maybeSave(p: SlotId) {
    const s = states[p];
    const q = wanted();
    if (libTab || p !== active || !q || s.q !== q || s.status !== 'ready' || savedKey === `${p}:${q}`) return;
    if (!shouldSave({ sinceKeyMs: Date.now() - lastKeyAt, firstPageItems: s.items.length, query: q })) return;
    savedKey = `${p}:${q}`;
    // Go applies the list rules (dedupe, cap); a failed save only means no chip.
    saveSearch(p, q).then(() => refreshLibrary(), noop);
  }

  // ↵ on a chip: the query becomes the chip text and the search fires at once (no debounce).
  function runChip(q: string) {
    chip = -1;
    query = q;
    clearTransient();
    load(active);
    input.focus();
  }

  // ⌘⌫ on a chip: removed silently (no toast, no Undo, D-22); the selection stays in the row.
  async function removeChip() {
    const q = chips[chip];
    try {
      library = await removeSearch(active, q);
    } catch {
      return;
    }
    const n = Math.min(library.searches[active]?.length ?? 0, CHIP_MAX);
    chip = Math.min(chip, n - 1);
    if (chip < 0 && gridShown) selected = 0;
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

  // Shows p's own state, fetching only when what it holds belongs to another query or is stale.
  function show(p: SlotId) {
    const s = states[p];
    if (s.err?.kind === 'key_rejected') return; // S13: the card stays until a new key is saved
    const stale = s.stale;
    s.stale = false;
    if (!stale && s.status !== 'idle' && s.q === wanted()) return;
    load(p);
  }

  function switchTo(p: SlotId) {
    menuOpen = detailsOpen = false;
    chip = -1;
    if (p !== active || libTab) {
      if (p !== active) setLastProvider(p);
      active = p;
      libTab = null;
      clearTransient();
      selected = states[p].items.length ? 0 : -1;
      if (p === 'clips' && clipsState === 'unknown') bootClips();
      else show(p);
    }
    tick().then(() => card || input.focus()); // a setup card focuses its own key field
  }

  // ⌘5 / ⌘6: the library renders from the saved snapshot; reloaded on every switch (D-09).
  function switchLib(id: LibraryId) {
    menuOpen = detailsOpen = false;
    chip = -1;
    clearTimeout(timer); // a pending provider search must not fire for the filter text
    if (libTab !== id) {
      libTab = id;
      clearTransient();
      selected = libItems.length ? 0 : -1;
    }
    refreshLibrary(''); // '' = no item to keep: the first item of the fresh list
    tick().then(() => input.focus());
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
      maybeSave(p);
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
      menuOpen = detailsOpen = false; // its item is gone
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
  // ⌘⌫ and ⌘Z are the exception: they are captured only when they act (a library item is selected;
  // an Undo toast is visible), so otherwise they stay native editing and undo in the input (Pitfall 10).
  function onKeydown(e: KeyboardEvent) {
    const a = routeKey(e);
    // A printable key closes Details and goes on to the input as usual.
    if (detail && a === 'none' && e.key.length === 1 && !e.metaKey && !e.ctrlKey) detailsOpen = false;
    if (a === 'none') return;
    if (chip >= 0 && chipKey(a)) return e.preventDefault();
    if (a === 'remove' && !(libTab && items[selected])) return;
    if (a === 'undo' && !undoVisible()) return;
    e.preventDefault();
    switch (a) {
      case 'left':
      case 'right':
      case 'up':
      case 'down':
        // In Details ← → step through the grid's items (clamped) and the pane follows; ↑ ↓ do nothing.
        if (detail && (a === 'up' || a === 'down')) return;
        // ↑ from the grid's first row, or ↓ with no grid, selects the first chip.
        if (chips.length && !detail && ((a === 'up' && gridShown && selected < COLS) || (a === 'down' && !gridShown))) {
          chip = 0;
          select(-1);
          return;
        }
        if (!items.length) return;
        if (libTab) {
          select(moveSectioned(selected, a, libSections.map((s) => s.items.length)));
          return;
        }
        select(move(selected, a, items.length));
        if (nearEnd(selected, items.length)) loadMore();
        return;
      case 'favorite':
      case 'details':
        runAction(a);
        return;
      case 'remove':
        runAction('removeRecent'); // on the Favorites tab this unfavorites
        return;
      case 'undo':
        runToastAction();
        return;
      case 'copy':
      case 'paste':
        // ↵ with nothing selected runs a visible Try Again.
        if (selected < 0) {
          if (a === 'copy' && libTab && library?.unreadable.includes(libTab)) refreshLibrary();
          else if (a === 'copy' && !libTab && errView?.tryAgain) retry();
          return;
        }
        runAction(a);
        return;
      case 'copyLink':
      case 'copyMarkdown':
      case 'open':
      case 'download':
        runAction(a);
        return;
      case 'actions':
        toggleMenu();
        return;
      case 'escape':
        if (detail) detailsOpen = false; // selection kept
        else if (query) {
          query = '';
          onInput();
        } else hide(); // no-op in the server build; nothing here depends on it
        return;
      case 'reveal':
        // Show in Finder while the download toast is up; otherwise captured and ignored (never reloads).
        if (toastAction?.hint === '⌘R') runToastAction();
        return;
      // 'tab': bound, no effect; slot1–slot4: onMainKeydown.
    }
  }

  // Keys while a chip is selected; false = not a chip key (normal handling, e.g. Esc, ⌘1–⌘6).
  function chipKey(a: string): boolean {
    switch (a) {
      case 'left':
      case 'right':
        chip = Math.max(0, Math.min(chips.length - 1, chip + (a === 'left' ? -1 : 1)));
        return true;
      case 'down':
        // Back to the first tile; with no grid (S8 or a message state) nothing happens.
        if (gridShown) {
          chip = -1;
          select(0);
        }
        return true;
      case 'up':
      case 'paste':
      case 'actions':
        return true;
      case 'copy':
        runChip(chips[chip]);
        return true;
      case 'remove':
        removeChip();
        return true;
    }
    return false;
  }

  // ⌘1–⌘6 from any field in the panel (search input or key field) switch provider or library tab.
  function onMainKeydown(e: KeyboardEvent) {
    const a = routeKey(e);
    // ⌘, from any panel context (search, setup card key field, action menu): Go hides the panel first.
    if (a === 'settings') {
      e.preventDefault();
      openSettings();
    } else if (a.startsWith('slot')) {
      e.preventDefault();
      const n = Number(a.slice(4));
      if (n <= SLOTS.length) switchTo(SLOTS[n - 1].id);
      else switchLib(n === 5 ? 'favorites' : 'recents');
      // The menu's filter has focus while it is open: its ⌘ shortcuts (⌘K, ⌘↵, ⌘⇧C, …) arrive here.
    } else if (menuOpen && e.metaKey && e.target !== input) onKeydown(e);
  }

  // ⌘K: open beside the selected tile, or close. No tile on screen (empty, loading, message, setup card): nothing.
  async function toggleMenu() {
    if (menuOpen) return closeMenu();
    const g = items[selected];
    const tile = document.getElementById(`tile-${selected}`);
    if (!g || !(tile || detail)) return;
    const inDetails = !!detail;
    menuRows = buildActions({ gif: g, tab: libTab ?? 'provider', inDetails, favorited: favKeys.has(keyOf(g)) });
    menuOpen = true;
    await tick(); // measure the rendered menu, then place it before the browser paints
    if (inDetails || !tile) {
      menuAt = placeInDetails(menu!.height());
      menu!.focus();
      return;
    }
    // Library sections restart the column count at each section's first tile.
    const col = (selected - (libSections.findLast((s) => s.start <= selected)?.start ?? 0)) % COLS;
    menuAt = placeMenu(tile.getBoundingClientRect(), col, menu!.height(), { w: innerWidth, h: innerHeight });
    menu!.focus();
  }

  function closeMenu() {
    menuOpen = false;
    input.focus();
  }

  // The one dispatcher for the menu and every shortcut. The menu closes before anything runs.
  function runAction(id: ActionId) {
    if (menuOpen) closeMenu();
    const g = items[selected];
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
        if (g?.pageUrl) openItem(g, openPage);
        return;
      case 'details':
        if (g) detailsOpen = !detailsOpen; // the panel stays open
        return;
      case 'download':
        downloadItem();
        return;
      case 'copySquare':
        act(copySquare, false, true);
        return;
      case 'pasteSquare':
        act(pasteSquare, true, true);
        return;
      case 'favorite':
        toggleFavorite();
        return;
      case 'removeRecent':
        removeSelected();
        return;
      case 'clearRecents':
        clearAll();
        return;
    }
  }

  // ⌘⇧F / the menu row, on any grid: add (2 s toast) or remove (5 s Undo toast). The panel stays open.
  async function toggleFavorite() {
    const g = items[selected];
    if (!g || !library) return;
    const k = keyOf(g);
    const on = !favKeys.has(k) && libTab !== 'favorites';
    const full = library.favorites.length >= FAVORITES_MAX; // Go drops the oldest to make room (P23)
    const lib = await mutate(() => setFavorite(g, on), 'favorites', on ? k : undefined);
    if (on && lib) setStatus(full ? 'Added to Favorites. Oldest favorite removed (limit 1,000).' : 'Added to Favorites', 'success', null, FAVORITE_TOAST_MS);
    else removed(lib, 'Removed from Favorites', 'favorites', k);
  }

  // A removal offers Undo; one that changed nothing (canUndo false: Go held no such item, WR-03) has
  // cleared Go's Undo slot, so an earlier Undo toast would restore nothing: clear it (P24, IN-06).
  function removed(lib: Library | null, text: string, list: LibraryId, k: string) {
    if (!lib) return;
    if (lib.canUndo) undoToast(text, list, k);
    else if (toastAction?.label === 'Undo') setStatus('');
  }

  // ⌘⌫ on a library tab: Remove from Recents, or unfavorite on the Favorites tab. No confirmation.
  async function removeSelected() {
    const g = items[selected];
    if (!g || !libTab) return;
    if (libTab === 'favorites') return toggleFavorite();
    removed(await mutate(() => removeRecent(g), 'recents'), 'Removed from Recents', 'recents', keyOf(g));
  }

  async function clearAll() {
    removed(await mutate(clearRecents, 'recents'), 'Cleared Recents', 'recents', '');
  }

  // T8–T10: Undo restores the latest removal and reselects it ('' = the first item, after Clear Recents).
  function undoToast(text: string, list: LibraryId, k: string) {
    const run = () => void mutate(undoLibrary, list, k);
    setStatus(text, '', { label: 'Undo', hint: '⌘Z', run }, UNDO_TOAST_MS);
  }

  // Runs one library change, applies its fresh lists and returns them; on failure F5 shows, the
  // lists stay as they were and it returns null.
  async function mutate(call: () => Promise<Library>, list: LibraryId, want?: string): Promise<Library | null> {
    try {
      const lib = await call();
      applyLibrary(lib, want);
      return lib;
    } catch {
      setStatus(libraryErrorText(list), 'error');
      return null;
    }
  }

  function run(action: 'copy' | 'paste') {
    act(action === 'copy' ? copy : paste, action === 'paste');
  }

  // Go builds the text from the selected item; the webview only names the kind.
  function textAction(kind: TextKind, pasteIt = false) {
    if (pasteIt && kind !== 'page') act((g) => pasteText(g, kind), true, false, 'text');
    else act((g) => copyText(g, kind), false, false, 'text'); // there is no Paste Page Link
  }

  // One copy or paste of the selected item at a time. A copy confirms in the footer and hides the
  // panel; a paste has hidden it already, unless it fell back to copying.
  // A square crop shows the overlay and `Cropping…` only if it is still running after 300 ms.
  async function act(call: (g: Gif) => Promise<Result>, pasteIt: boolean, crop = false, failAs: 'copy' | 'text' = 'copy') {
    const g = items[selected];
    if (copying || !g) return;
    copying = true;
    // Copying from a searched provider grid saves that search (D-13); trending and library filters never do.
    const p = active;
    const searched = libTab ? '' : states[p].q;
    const busy = () => {
      overlay = true;
      setStatus(crop ? 'Cropping…' : 'Copying…');
    };
    const busyTimer = crop ? setTimeout(busy, CROP_BUSY_MS) : undefined;
    if (!crop) busy();
    try {
      const r = await call(g);
      if (searched) saveSearch(p, searched).then(() => refreshLibrary(), noop);
      if (!pasteIt) {
        setStatus(r.message, 'success');
        scheduleHide();
      } else {
        setStatus(r.message, r.pasted ? 'success' : '');
        needsAccessibility = r.needsAccessibility;
        keepStatus = r.copied && !r.pasted && !r.needsAccessibility;
      }
    } catch (err) {
      actionFailed(err, g, failAs);
    } finally {
      clearTimeout(busyTimer);
      copying = overlay = false;
    }
  }

  // Copy actions confirm in the footer, then hide the panel; panel:show cancels a pending hide.
  function scheduleHide() {
    clearTimeout(hideTimer);
    hideTimer = setTimeout(hide, HIDE_AFTER_COPY_MS);
  }

  // Fixed copy keyed by Go's error kind, never the error's text. A busy refusal is silently ignored.
  function actionFailed(err: unknown, g: Gif, action: 'copy' | 'download' | 'text' = 'copy') {
    const kind = actionError(err)?.kind ?? '';
    if (kind === 'busy') setStatus('');
    else setStatus(actionErrorText(kind, g, action, downloadFolder), 'error');
  }

  // Saves the selected item into the download folder; the panel stays open (D-04). Shares the
  // copy lock and busy overlay: one file action at a time.
  async function downloadItem() {
    const g = items[selected];
    if (copying || !g) return;
    copying = overlay = true;
    setStatus('Downloading…');
    try {
      const r = await download(g);
      downloadFolder = r.folder;
      // A Finder failure is logged in Go; the toast simply clears (UI-SPEC has no copy for it).
      const reveal = () => void showDownload().catch(() => {});
      setStatus(`Downloaded to ${r.folder}`, 'success', { label: 'Show in Finder', hint: '⌘R', run: reveal }, DOWNLOAD_TOAST_MS);
    } catch (err) {
      actionFailed(err, g, 'download');
    } finally {
      copying = overlay = false;
    }
  }

  // Success shows no toast: the browser takes focus and HideOnFocusLost hides the panel.
  async function openItem(g: Gif, call: (g: Gif) => Promise<void>) {
    try {
      await call(g);
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
      placeholder={libTab ? libraryTab(libTab).placeholder : slot.placeholder}
      disabled={card}
      autocomplete="off"
      spellcheck="false"
      role="combobox"
      aria-expanded={items.length > 0}
      aria-controls="grid"
      aria-activedescendant={chip >= 0 ? `chip-${chip}` : selected >= 0 && !detail ? `tile-${selected}` : undefined}
    />
    {#if busy}
      <span class="spinner busy" role="status" aria-label="Loading"></span>
    {/if}
  </header>

  <ProviderTabs {active} lib={libTab} {clipsEnabled} onswitch={switchTo} onlib={switchLib} />

  <section class="content" aria-busy={statuses === null || (!!libTab && !library)}>
    <!-- Details covers the content area; the grid stays mounted underneath so ← → move its selection. -->
    {#if detail}
      <div class="details-layer">
        <DetailsPane
          gif={detail}
          favorited={favKeys.has(keyOf(detail))}
          onopenpage={() => openItem(detail, openPage)}
          onopensource={() => openItem(detail, openSource)}
        />
      </div>
    {/if}
    {#if chips.length}
      <RecentSearches searches={chips} selected={chip} onrun={runChip} />
    {/if}
    <div class="view">
      {#if libTab}
        {#if !library}
          {#if libSlow}
            <div class="skeleton" aria-hidden="true">
              {#each Array(8), i (i)}
                <div></div>
              {/each}
            </div>
          {/if}
        {:else if library.unreadable.includes(libTab)}
          <StateMessage
            icon
            heading="Couldn't load {libName}"
            body="Gifsmith couldn't read its saved list. Try again, or quit and reopen Gifsmith."
            button="Try Again"
            onbutton={() => refreshLibrary()}
          />
        {:else if !library[libTab].length && libTab === 'favorites'}
          <StateMessage
            heading="No favorites yet"
            body="Press ⌘⇧F on any GIF to save it here. Favorites open instantly, even offline or when a provider's request limit is reached."
          />
        {:else if !library[libTab].length}
          <StateMessage heading="No recent GIFs" body="GIFs you copy or paste show up here, newest first." />
        {:else if !libItems.length}
          <StateMessage heading="No matches" body="Nothing in {libName} matches “{query.trim()}”. Press Esc to clear the search." />
        {:else}
          <!-- Remounted when the row layout changes: the virtualizer reads row sizes only on a count change. -->
          {#key `${libTab}:${libSections.map((s) => s.items.length).join(',')}`}
            <LibraryGrid
              sections={libSections}
              {selected}
              {overlay}
              favorites={favKeys}
              badges={libTab !== 'favorites'}
              onselect={select}
              oncopy={(i) => {
                select(i);
                run('copy');
              }}
            />
          {/key}
        {/if}
      {:else if statuses === null}
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
            {overlay}
            favorites={favKeys}
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
    </div>
  </section>

  {#if menuOpen}
    <ActionMenu bind:this={menu} rows={menuRows} {...menuAt} onrun={runAction} onclose={closeMenu} />
  {/if}

  <Footer
    {attribution}
    hints={detail ? 'details' : libTab ? (items[selected] ? 'library' : 'none') : 'provider'}
    text={footerCenter}
    {tone}
    {needsAccessibility}
    onenable={enableAutoPaste}
    actionLabel={toastAction?.label}
    actionHint={toastAction?.hint}
    onaction={runToastAction}
    onback={() => {
      detailsOpen = false; // the Esc path: selection kept
      input.focus();
    }}
  />
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
  /* Opts out of the shared focus ring: the combobox always holds focus (P1). */
  input:focus-visible {
    outline: 0;
  }
  /* P7: stale query text behind a setup card reads as inactive. */
  input:disabled {
    color: var(--text-tertiary);
  }
  input::placeholder {
    color: var(--text-secondary);
  }
  .busy {
    flex: none;
    color: var(--text-secondary);
  }
  .content {
    position: relative;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }
  /* The grid or state below the chip row: 364px, or 332px under chips. */
  .view {
    flex: 1;
    min-height: 0;
  }
  .details-layer {
    position: absolute;
    inset: 0;
    z-index: 1;
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
