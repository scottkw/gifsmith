<script lang="ts">
  // Hotkeys pane (UI-SPEC "Hotkeys"): one stacked row per global shortcut. Go owns validation and
  // registration; a row shows a new combo only once Go accepted it. Test listens 5 s for the
  // hotkey:test event Go emits when that hotkey's callback fires (the only way to see a clash).
  import { Events } from '@wailsio/runtime';
  import { onMount } from 'svelte';
  import { setHotkey, settingsError, testHotkey, type SettingsSnapshot } from '../lib/api';
  import { hotkeyErrorText } from '../lib/errors';
  import { canonical, DEFAULT_OPEN, glyphs, SLOT_LABELS, type Slot } from '../lib/hotkeys';
  import HotkeyRecorder from './HotkeyRecorder.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  const ROWS = (['open', 'favorites', 'recents'] as const).map((slot) => ({ slot, label: SLOT_LABELS[slot] }));
  const TEST_MS = 5000;

  // Only SettingsApp's snapshot (settings:changed, and a re-read on show), never a call's result:
  // responses are not ordered against events, so a late one could show an older binding.
  const hk = $derived(initial.hotkeys);
  let msg = $state<Record<string, { text: string; tone: 'secondary' | 'success' | 'error' }>>({});
  let listening = $state<Slot | null>(null);
  let listenTimer: ReturnType<typeof setTimeout> | undefined;
  let offTest: (() => void) | undefined;

  onMount(() => {
    // A stored combo macOS refused at launch is not bound (the row shows Not set): say why (H4b).
    for (const { slot } of ROWS) {
      const refused = initial.hotkeysFailed[slot];
      if (refused) say(slot, hotkeyErrorText('hotkey_register_failed_new', refused, ''), 'error');
    }
    return stopListening;
  });

  const say = (slot: Slot, text: string, tone: 'secondary' | 'success' | 'error' = 'secondary') => (msg[slot] = { text, tone });

  // '' clears the row's message (a recording started or succeeded); otherwise H1–H4b under the row.
  // The field keeps its combo either way.
  function fail(slot: Slot, kind: string, accel: string, other = '') {
    if (!kind) stopListening(); // hotkeys are suspended while recording: a Test would report H7 falsely
    say(slot, kind ? hotkeyErrorText(kind, accel, SLOT_LABELS[other as Slot] ?? '') : '', kind ? 'error' : 'secondary');
  }

  async function change(slot: Slot, accel: string) {
    stopListening();
    say(slot, '');
    try {
      await setHotkey(slot, accel); // a change arrives as settings:changed
    } catch (err) {
      const se = settingsError(err);
      fail(slot, se?.kind ?? 'hotkey_register_failed', accel, se?.slot);
    }
  }

  function stopListening() {
    clearTimeout(listenTimer);
    offTest?.();
    offTest = undefined;
    if (listening) say(listening, '');
    listening = null;
  }

  // H5 while listening, then H6 when Go reports the press, else H7 after 5 s.
  async function test(slot: Slot) {
    stopListening();
    const accel = hk[slot] ?? '';
    try {
      await testHotkey(slot);
    } catch {
      return say(slot, hotkeyErrorText('test_failed', accel, ''), 'error');
    }
    listening = slot;
    say(slot, `Press ${glyphs(accel)} now…`);
    offTest = Events.On('hotkey:test', (e) => {
      if (e.data !== slot) return;
      stopListening();
      say(slot, 'It works.', 'success');
    });
    listenTimer = setTimeout(() => {
      stopListening();
      say(slot, hotkeyErrorText('test_failed', accel, ''), 'error');
    }, TEST_MS);
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key !== 'Escape' || !listening) return;
    e.preventDefault();
    stopListening();
  }
</script>

<svelte:window {onkeydown} />

{#each ROWS as r, i (r.slot)}
  <div class="settings-section">
    <span class="label" id="hk-{r.slot}-label">{r.label}</span>
    <div class="line">
      <HotkeyRecorder
        slot={r.slot}
        label={r.label}
        value={hk[r.slot] ?? ''}
        describedby="hk-{r.slot}-msg"
        onfail={(kind, accel, other) => fail(r.slot, kind, accel, other)}
      />
      <button class="btn-secondary" aria-label="Test {r.label} shortcut" disabled={!hk[r.slot]} onclick={() => test(r.slot)}>Test</button>
      {#if r.slot === 'open'}
        <button
          class="btn-secondary"
          aria-label="Reset to Default, {r.label} shortcut"
          disabled={!!hk.open && canonical(hk.open) === DEFAULT_OPEN}
          onclick={() => change('open', DEFAULT_OPEN)}>Reset to Default</button
        >
      {:else}
        <button class="btn-secondary" aria-label="Clear {r.label} shortcut" disabled={!hk[r.slot]} onclick={() => change(r.slot, '')}>Clear</button>
      {/if}
    </div>
    <p class="msg {msg[r.slot]?.tone ?? 'secondary'}" id="hk-{r.slot}-msg" role={msg[r.slot]?.tone === 'error' ? 'alert' : 'status'}>
      {#if msg[r.slot]?.tone === 'error'}
        <svg viewBox="0 0 16 16" aria-hidden="true">
          <circle cx="8" cy="8" r="6.25" />
          <path d="M8 5v3.5" />
          <circle class="dot" cx="8" cy="11" r="0.4" />
        </svg>
      {/if}{msg[r.slot]?.text ?? ''}
    </p>
    {#if i === ROWS.length - 1}
      <p class="msg secondary">macOS doesn't warn when another app uses the same shortcut. Use Test to check yours.</p>
    {/if}
  </div>
{/each}

<style>
  .label {
    display: block;
    font: var(--body);
    color: var(--text);
  }
  .line {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
  }
  .msg {
    display: flex;
    gap: 4px;
    width: 432px;
    margin: 4px 0 0;
    font: var(--label);
  }
  .msg:empty {
    margin: 0;
  }
  .secondary {
    color: var(--text-secondary);
  }
  .success {
    color: var(--success);
  }
  .error {
    color: var(--destructive);
  }
  svg {
    width: 12px;
    height: 12px;
    flex: none;
    margin-top: 2px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .dot {
    fill: currentColor;
  }
</style>
