<script lang="ts">
  // Footer (32px): vendor attribution left, status or selected title center, key hints right.
  let {
    attribution,
    text,
    tone,
    needsAccessibility,
    onenable,
  }: {
    attribution: 'giphy' | 'klipy' | 'finer';
    text: string;
    tone: '' | 'success' | 'error';
    needsAccessibility: boolean;
    onenable: () => void;
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
    {:else}
      GIFs from The Finer Gifs Club
    {/if}
  </span>
  <span class="center {tone}">
    {#if tone === 'success'}
      <svg class="check" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>
    {/if}
    {text}
  </span>
  {#if needsAccessibility}
    <button onclick={onenable}>Enable Automatic Paste…</button>
  {:else}
    <span class="hints"><b>↵</b> Copy&nbsp;&nbsp;&nbsp;<b>⌘↵</b> Paste</span>
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
  button {
    font: var(--label);
  }
</style>
