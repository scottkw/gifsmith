<script lang="ts">
  // Footer (32px): vendor attribution left, status or selected title center, key hints right.
  // A toast may carry one action link and its shortcut hint after the message.
  let {
    attribution,
    hints = 'provider',
    text,
    tone,
    needsAccessibility,
    onenable,
    actionLabel,
    actionHint,
    onaction,
  }: {
    attribution: 'giphy' | 'klipy' | 'finer' | ''; // '' = nothing to attribute (empty library tab)
    hints?: 'provider' | 'library' | 'details';
    text: string;
    tone: '' | 'success' | 'error';
    needsAccessibility: boolean;
    onenable: () => void;
    actionLabel?: string;
    actionHint?: string;
    onaction?: () => void;
  } = $props();
</script>

<footer>
  <span class="attribution">
    {#if attribution === 'giphy'}
      <!-- Official GIPHY mark, unmodified (no filter, opacity or crop), not a link. -->
      <picture>
        <source srcset="/attribution/giphy-powered-dark.png" media="(prefers-color-scheme: dark)" />
        <img src="/attribution/giphy-powered-light.png" alt="Powered by GIPHY" height="16" />
      </picture>
    {:else if attribution === 'klipy'}
      Powered by KLIPY
    {:else if attribution === 'finer'}
      GIFs from The Finer Gifs Club
    {/if}
  </span>
  <span class="center {tone}">
    <span class="msg">
      {#if tone === 'success'}
        <svg class="check" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>
      {/if}
      {text}
    </span>
    {#if actionLabel && onaction}
      <button class="link" onclick={onaction}>{actionLabel}</button>
      {#if actionHint}<span class="hint">{actionHint}</span>{/if}
    {/if}
  </span>
  {#if needsAccessibility}
    <button class="secondary" onclick={onenable}>Enable Automatic Paste…</button>
  {:else if hints === 'details'}
    <!-- Details: "↵ Copy   ⌘K Actions   Esc Back" -->
    <span class="hints"><b>↵</b> Copy&nbsp;&nbsp;&nbsp;<b>⌘K</b> Actions&nbsp;&nbsp;&nbsp;<b>Esc</b> Back</span>
  {:else if hints === 'library'}
    <span class="hints"><b>↵</b> Copy&nbsp;&nbsp;&nbsp;<b>⌘K</b> Actions&nbsp;&nbsp;&nbsp;<b>⌘⌫</b> Remove</span>
  {:else}
    <span class="hints"><b>↵</b> Copy&nbsp;&nbsp;&nbsp;<b>⌘↵</b> Paste&nbsp;&nbsp;&nbsp;<b>⌘K</b> Actions</span>
  {/if}
</footer>

<style>
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
  .hints,
  button {
    flex: none;
  }
  .attribution {
    display: flex;
    align-items: center;
  }
  img {
    display: block;
    height: 16px;
    width: auto;
  }
  /* The message ellipsizes; the action link and its hint never shrink. */
  .center {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
  }
  .msg {
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .link {
    flex: none;
    margin-left: 8px;
    padding: 0;
    border: 0;
    background: none;
    color: var(--accent);
    cursor: pointer;
  }
  .link:hover {
    text-decoration: underline;
  }
  .hint {
    flex: none;
    margin-left: 4px;
    color: var(--text-tertiary);
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
  button {
    font: var(--label);
  }
  /* P2: secondary compact button. */
  .secondary {
    height: 24px;
    padding: 0 8px;
    border: 0;
    border-radius: 6px;
    background: var(--surface-active);
    color: var(--text);
    font: 600 12px/16px var(--font);
    cursor: pointer;
  }
</style>
