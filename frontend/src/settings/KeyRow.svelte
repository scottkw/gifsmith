<script lang="ts">
  // One provider card on the Keys pane (UI-SPEC "Keys"): status, add/change editor, Test, and an
  // inline remove confirmation. The key exists here only while typed; Go returns the last 4 at most.
  import { tick } from 'svelte';
  import { openExternal, providerError, removeKey, saveKey, testKey, type KeyStatus } from '../lib/api';
  import { keyMessage, keyRemovedText, keyRemoveFailedText, keySavedText, testKeyMessage } from '../lib/errors';

  let {
    provider,
    label,
    description,
    consequence,
    keyUrl,
    status,
  }: { provider: 'giphy' | 'klipy'; label: string; description: string; consequence: string; keyUrl: string; status: KeyStatus } = $props();

  let mode = $state<'view' | 'edit' | 'confirm'>('view');
  let value = $state('');
  let verifying = $state(false);
  let testing = $state(false);
  let message = $state('');
  let tone = $state<'success' | 'secondary' | 'error'>('success');
  let field = $state<HTMLInputElement>();
  let cancelBtn = $state<HTMLButtonElement>();
  let actions = $state<HTMLDivElement>();
  let msgTimer: ReturnType<typeof setTimeout> | undefined;

  const saved = $derived(status.status !== 'not_set');

  function say(text: string, t: typeof tone) {
    clearTimeout(msgTimer);
    message = text;
    tone = t;
    if (t !== 'error' && text) msgTimer = setTimeout(() => (message = ''), 4000);
  }

  async function open(m: 'edit' | 'confirm') {
    say('', 'success');
    mode = m;
    await tick();
    (m === 'edit' ? field : cancelBtn)?.focus();
  }

  async function close() {
    mode = 'view';
    value = '';
    await tick();
    actions?.querySelector<HTMLButtonElement>('button')?.focus();
  }

  async function submit(e: Event) {
    e.preventDefault();
    const key = value.trim();
    if (!key || verifying) return;
    verifying = true;
    say('', 'success');
    try {
      await saveKey(provider, key); // a real trending call with this key, then the Keychain
      verifying = false;
      await close();
      say(keySavedText(label), 'success');
    } catch (err) {
      verifying = false;
      say(keyMessage(providerError(err), label, Date.now()), 'error');
      await tick();
      field?.focus();
      field?.select();
    }
  }

  async function test() {
    if (testing) return;
    testing = true;
    say('', 'success');
    try {
      await testKey(provider);
      say(testKeyMessage(null, label, Date.now()), 'success');
    } catch (err) {
      say(testKeyMessage(providerError(err) ?? { kind: 'http', provider }, label, Date.now()), 'error');
    } finally {
      testing = false;
    }
  }

  async function remove() {
    try {
      await removeKey(provider);
      await close();
      say(keyRemovedText(label), 'secondary');
    } catch {
      await close();
      say(keyRemoveFailedText(), 'error');
    }
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key !== 'Escape' || e.isComposing || mode === 'view' || verifying) return;
    e.preventDefault();
    close();
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div class="card" role="group" aria-labelledby="key-{provider}-title" {onkeydown}>
  <div class="title">
    <span id="key-{provider}-title" class="name">{label}</span>
    {#if testing}
      <span class="status muted"><span class="spinner" aria-hidden="true"></span> Testing…</span>
    {:else if status.status === 'saved'}
      <span class="status ok">
        <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>
        Saved · ends in {status.last4}
      </span>
    {:else if status.status === 'rejected'}
      <span class="status bad">
        <svg viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="6.25" /><path d="M8 5v3.5" /><circle class="dot" cx="8" cy="11" r="0.4" /></svg>
        Rejected
      </span>
    {:else}
      <span class="status muted">Not set</span>
    {/if}
  </div>
  <p class="desc">{description}</p>

  {#if mode === 'edit'}
    <form class="actions" onsubmit={submit}>
      <input
        bind:this={field}
        bind:value
        type="text"
        placeholder="Paste your {label} API key"
        aria-label="{label} API key"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        readonly={verifying}
        class:invalid={tone === 'error' && message}
      />
      <button class="btn-primary" type="submit" disabled={verifying || !value.trim()}>
        {#if verifying}<span class="spinner" aria-hidden="true"></span> Verifying…{:else}Save Key{/if}
      </button>
      <button class="btn-secondary" type="button" disabled={verifying} onclick={close}>Cancel</button>
    </form>
  {:else if mode === 'confirm'}
    <div class="confirm">
      <p class="ask">Remove your {label} key?</p>
      <p class="desc">{consequence}</p>
      <div class="buttons end">
        <button class="btn-secondary" bind:this={cancelBtn} onclick={close}>Cancel</button>
        <button class="btn-destructive" onclick={remove}>Remove Key</button>
      </div>
    </div>
  {:else}
    <div class="actions" bind:this={actions}>
      <div class="buttons">
        {#if saved}
          <button class="btn-secondary" onclick={() => open('edit')}>Change Key…</button>
          <button class="btn-secondary" aria-label="Test {label} key" disabled={testing} onclick={test}>Test</button>
          <button class="btn-secondary" aria-label="Remove {label} key" onclick={() => open('confirm')}>Remove</button>
        {:else}
          <button class="btn-primary" onclick={() => open('edit')}>Add Key…</button>
        {/if}
      </div>
      <a
        href={keyUrl}
        onclick={(e) => {
          e.preventDefault();
          openExternal(keyUrl);
        }}>Get a free key ↗</a
      >
    </div>
  {/if}

  <div class="message {tone}" role={tone === 'error' ? 'alert' : 'status'}>{message}</div>
</div>

<style>
  .card {
    box-sizing: border-box;
    width: 432px;
    padding: 16px;
    border-radius: 8px;
    background: var(--surface);
  }
  .title {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .name {
    font: 600 14px/20px var(--font);
    color: var(--text);
  }
  .status {
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--label);
  }
  .ok {
    color: var(--success);
  }
  .bad {
    color: var(--destructive);
  }
  .muted {
    color: var(--text-secondary);
  }
  .status svg {
    width: 12px;
    height: 12px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .status .dot {
    fill: currentColor;
  }
  p {
    margin: 0;
  }
  .desc {
    margin-top: 4px;
    font: var(--label);
    color: var(--text-secondary);
  }
  .actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin: 16px 0 0;
  }
  .buttons {
    display: flex;
    gap: 8px;
  }
  .end {
    justify-content: flex-end;
    margin-top: 16px;
  }
  .confirm {
    margin-top: 16px;
  }
  .ask {
    font: var(--body);
    color: var(--text);
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
  a {
    font: var(--body);
    color: var(--accent);
    text-decoration: none;
    white-space: nowrap;
  }
  a:hover {
    text-decoration: underline;
  }
  /* Wraps instead of UI-SPEC's one-line ellipsis: K2, K3 and K-T2 run 430-470px at 12px, wider
     than the 400px line, and a cut-off error hides what to do next. */
  .message {
    min-height: 16px;
    margin-top: 8px;
    font: var(--label);
  }
  .message.success {
    color: var(--success);
  }
  .message.secondary {
    color: var(--text-secondary);
  }
  .message.error {
    color: var(--destructive);
  }
</style>
