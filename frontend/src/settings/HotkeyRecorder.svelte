<script lang="ts">
  // Click-to-record shortcut field (UI-SPEC "Recorder field"). Gifsmith's hotkeys are suspended for
  // the whole recording so the current combo is captured, not fired; every way out resumes them.
  // The field shows a new combo only after Go registered it.
  import { resumeHotkeys, setHotkey, settingsError, suspendHotkeys, type SettingsSnapshot } from '../lib/api';
  import { accelFromEvent, glyphs } from '../lib/hotkeys';

  let {
    slot,
    label,
    value,
    describedby,
    onsaved,
    onfail,
  }: {
    slot: string;
    label: string;
    value: string; // canonical accelerator, '' = not set
    describedby: string;
    onsaved: (s: SettingsSnapshot) => void;
    onfail: (kind: string, accel: string, otherSlot: string) => void; // '' kind clears the row's message
  } = $props();

  let recording = $state(false);
  let held = $state('');
  let saving = false;

  function start() {
    if (recording) return;
    recording = true;
    held = '';
    onfail('', '', '');
    suspendHotkeys();
  }

  function stop() {
    if (!recording) return;
    recording = false;
    held = '';
    resumeHotkeys();
  }

  async function save(accel: string) {
    saving = true;
    try {
      onsaved(await setHotkey(slot, accel));
    } catch (err) {
      const se = settingsError(err);
      onfail(se?.kind ?? 'hotkey_register_failed', accel, se?.slot ?? '');
    } finally {
      saving = false;
      stop();
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if (!recording) {
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault();
        start();
      }
      return;
    }
    const mods = e.metaKey || e.ctrlKey || e.altKey || e.shiftKey;
    if (e.key === 'Tab' && !mods) return stop(); // focus moves on
    e.preventDefault();
    if (saving) return;
    if (e.key === 'Escape' && !mods) return stop();
    const r = accelFromEvent(e);
    if (r.kind === 'held') held = r.glyphs;
    else if (r.kind === 'accel') save(r.accel);
  }

  function onkeyup(e: KeyboardEvent) {
    const r = accelFromEvent(e);
    if (recording && r.kind === 'held') held = r.glyphs;
  }
</script>

<div
  class="recorder"
  class:recording
  class:muted={recording ? !held : !value}
  role="button"
  tabindex="0"
  aria-label="{label} shortcut"
  aria-describedby={describedby}
  onclick={start}
  {onkeydown}
  {onkeyup}
  onblur={stop}
>
  {#if recording}{held || 'Type shortcut…'}{:else}{value ? glyphs(value) : 'Not set'}{/if}
</div>

<style>
  .recorder {
    box-sizing: border-box;
    width: 160px;
    height: 32px;
    flex: none;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--control-border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    font: var(--body);
    cursor: default;
    user-select: none;
  }
  .muted {
    color: var(--text-secondary);
  }
  /* The accent ring stays up for the whole recording, also when it was started by mouse. */
  .recorder:focus-visible,
  .recording {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
</style>
