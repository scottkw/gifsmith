<script lang="ts">
  // Inline setup card (UI-SPEC S1/S1a/S1b/S13, copy K0–K6) for the GIPHY and Klipy slots.
  // The key exists here only while it is typed; it is cleared after a successful save.
  import { onMount, tick, untrack } from 'svelte';
  import { hide, openExternal, providerError, saveKey } from '../lib/api';
  import { keyMessage } from '../lib/errors';
  import type { Slot } from '../lib/providers';

  let {
    slot,
    initialMessage = '',
    onsaved,
    ontryfiner,
  }: { slot: Slot; initialMessage?: string; onsaved: (p: 'giphy' | 'klipy') => void; ontryfiner: () => void } = $props();

  const COPY = {
    giphy: {
      body: 'GIPHY needs a free API key of your own. It takes about 2 minutes, and free keys allow 100 requests per hour.',
      step1: '1. Click “Get a free key” and create an app on GIPHY\'s developer site.',
      step2: "2. Copy the app's API key.",
    },
    klipy: {
      body: 'Klipy needs a free API key of your own. It takes about 2 minutes, and new keys allow 100 requests per hour.',
      step1: '1. Click “Get a free key” and create a platform in the Klipy Partner Panel.',
      step2: '2. Copy its API key.',
    },
  };

  const p = $derived(slot.keyProvider!);
  const label = $derived(slot.keyLabel!);
  const copy = $derived(COPY[p]);

  let card: HTMLDivElement;
  let field: HTMLInputElement;
  let value = $state('');
  let verifying = $state(false);
  let message = $state(untrack(() => initialMessage)); // App re-keys the card when this changes

  export function focus() {
    field?.focus();
  }

  onMount(focus);

  async function submit(e?: Event) {
    e?.preventDefault();
    const key = value.trim();
    if (!key || verifying) return;
    verifying = true;
    message = '';
    try {
      await saveKey(p, key); // SettingsService.SaveKey(p, key): a trending call with this key, then the Keychain
      value = '';
      onsaved(p);
    } catch (err) {
      message = keyMessage(providerError(err), label, Date.now());
      verifying = false;
      await tick();
      field.focus();
      field.select();
    }
  }

  function onFieldKeydown(e: KeyboardEvent) {
    if (e.key !== 'Escape' || e.isComposing) return;
    e.preventDefault();
    if (value) {
      value = '';
      message = '';
    } else hide();
  }

  // Tab cycles inside the card: key field -> Save Key -> Get a free key -> Try Finer Gifs.
  function onCardKeydown(e: KeyboardEvent) {
    if (e.key !== 'Tab') return;
    const els = [...card.querySelectorAll<HTMLElement>('input, a[href], button:not(:disabled)')];
    const i = els.indexOf(document.activeElement as HTMLElement);
    if (!e.shiftKey && i === els.length - 1) {
      e.preventDefault();
      els[0].focus();
    } else if (e.shiftKey && i <= 0) {
      e.preventDefault();
      els[els.length - 1].focus();
    }
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="card" bind:this={card} onkeydown={onCardKeydown} role="group" aria-labelledby="setup-heading">
  <h2 id="setup-heading">Add your {label} key</h2>
  <p>{copy.body}</p>
  <div class="steps">
    <p>{copy.step1}</p>
    <p>{copy.step2}</p>
    <p>3. Reopen Gifsmith with ⌃⌥⌘G, paste the key here, and press Save Key.</p>
  </div>
  <form class="row" onsubmit={submit}>
    <input
      bind:this={field}
      bind:value
      oninput={() => (message = '')}
      onkeydown={onFieldKeydown}
      class:invalid={message}
      type="text"
      placeholder="Paste your {label} API key"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      readonly={verifying}
      aria-label="{label} API key"
      aria-invalid={message ? 'true' : undefined}
      aria-describedby="setup-message"
    />
    <button class="btn-primary" type="submit" disabled={verifying || !value.trim()}>
      {#if verifying}<span class="spinner" aria-hidden="true"></span> Verifying…{:else}Save Key{/if}
    </button>
  </form>
  <div class="message" id="setup-message" role="alert">
    {#if message}
      <svg viewBox="0 0 16 16" aria-hidden="true">
        <circle cx="8" cy="8" r="6.25" />
        <path d="M8 5v3.5" />
        <circle class="dot" cx="8" cy="11" r="0.4" />
      </svg>
      {message}
    {/if}
  </div>
  <div class="links">
    <a
      href={slot.getKeyUrl}
      onclick={(e) => {
        e.preventDefault();
        openExternal(slot.getKeyUrl!);
      }}>Get a free key ↗</a
    >
    <button class="btn-secondary" type="button" onclick={ontryfiner}>Try Finer Gifs (no key)</button>
  </div>
</div>

<style>
  /* 560, not UI-SPEC's 456: at 456 the steps wrap and the card ran 21px into the footer, and
     K2–K6 overflowed the one-line message slot. At 560 they fit on one line (K6 is 493px). */
  .card {
    box-sizing: border-box;
    width: 560px;
    margin: 24px auto 0;
    padding: 24px;
    border-radius: 12px;
    background: var(--surface);
  }
  h2 {
    margin: 0 0 8px;
    font: var(--heading);
    color: var(--text);
  }
  p {
    margin: 0;
    font: var(--body);
    color: var(--text-secondary);
  }
  .steps {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin: 16px 0;
  }
  .row {
    display: flex;
    gap: 8px;
    margin: 0;
  }
  input {
    flex: 1;
    min-width: 0;
    box-sizing: border-box;
    height: 32px;
    padding: 0 8px;
    border: 1px solid var(--control-border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    font: 400 14px/1 var(--font-mono);
    caret-color: var(--accent);
    -webkit-text-security: disc;
  }
  input.invalid {
    border-color: var(--destructive);
  }
  .message {
    height: 16px;
    margin: 8px 0;
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--label);
    color: var(--destructive);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .message svg {
    width: 12px;
    height: 12px;
    flex: none;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .message .dot {
    fill: currentColor;
  }
  .links {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  a {
    font: var(--body);
    color: var(--accent);
    text-decoration: none;
  }
  a:hover {
    text-decoration: underline;
  }
</style>
