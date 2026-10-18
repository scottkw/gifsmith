<script lang="ts">
  // Hotkeys pane (UI-SPEC "Hotkeys"): one stacked row per global shortcut. Go owns validation and
  // registration; a row shows a new combo only once Go accepted it.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { setHotkey, type SettingsSnapshot } from '../lib/api';
  import { canonical, DEFAULT_OPEN, glyphs } from '../lib/hotkeys';
  import HotkeyRecorder from './HotkeyRecorder.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  type Slot = 'open';
  const ROWS: { slot: Slot; label: string }[] = [{ slot: 'open', label: 'Open Gifsmith' }];

  let hk = $state(untrack(() => ({ ...initial.hotkeys })));
  let msg = $state<Record<string, { text: string; tone: 'secondary' | 'success' | 'error' }>>({});

  onMount(() => Events.On('settings:changed', (e) => (hk = { ...(e.data as SettingsSnapshot).hotkeys })));

  function saved(s: SettingsSnapshot) {
    hk = { ...s.hotkeys };
  }

  function fail(slot: Slot, kind: string, accel: string) {
    msg[slot] = kind ? { text: `Couldn't set ${glyphs(accel)}. Your previous shortcut still works.`, tone: 'error' } : { text: '', tone: 'secondary' };
  }

  async function reset() {
    msg.open = { text: '', tone: 'secondary' };
    try {
      saved(await setHotkey('open', DEFAULT_OPEN));
    } catch {
      fail('open', 'hotkey_register_failed', DEFAULT_OPEN);
    }
  }
</script>

{#each ROWS as r, i (r.slot)}
  <div class="settings-section">
    <span class="label" id="hk-{r.slot}-label">{r.label}</span>
    <div class="line">
      <HotkeyRecorder
        slot={r.slot}
        label={r.label}
        value={hk[r.slot] ?? ''}
        describedby="hk-{r.slot}-msg"
        onsaved={saved}
        onfail={(kind, accel) => fail(r.slot, kind, accel)}
      />
      {#if r.slot === 'open'}
        <button class="btn-secondary" disabled={canonical(hk.open || DEFAULT_OPEN) === DEFAULT_OPEN} onclick={reset}>Reset to Default</button>
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
