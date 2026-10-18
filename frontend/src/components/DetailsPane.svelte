<script lang="ts">
  // ⌘I Details (UI-SPEC "Details pane", D-10, D-11, D-21): built only from fields the item already
  // carries; no provider request. Page/Source open through Go, which re-validates the URL it holds.
  import type { Gif } from '../lib/api';
  import { dimensions, fileSize, linkText, providerLine, tagsSummary } from '../lib/format';

  let {
    gif,
    favorited,
    onopenpage,
    onopensource,
  }: {
    gif: Gif;
    favorited: boolean;
    onopenpage: () => void;
    onopensource: () => void;
  } = $props();

  const title = $derived(gif.title || 'Untitled GIF');
  // Fixed order; a field with no data is omitted with its term.
  const fields = $derived(
    [
      { term: 'Dimensions', value: dimensions(gif.width, gif.height) },
      { term: 'File size', value: fileSize(gif.size) },
      { term: 'Tags', value: tagsSummary(gif.tags) },
    ].filter((f) => f.value),
  );
</script>

<div class="details" id="details" role="region" aria-label="Details">
  <div class="preview">
    <img src={gif.mediumUrl || gif.previewUrl} alt={title} decoding="async" draggable="false" />
    {#if gif.isClip}
      <span class="badge" aria-hidden="true">MP4</span>
    {/if}
  </div>
  <div class="meta">
    <h2 title={title}>{title}</h2>
    <p class="provider">{providerLine(gif, favorited)}</p>
    <dl>
      {#each fields as f (f.term)}
        <dt>{f.term}</dt>
        <dd>{f.value}</dd>
      {/each}
      {#if gif.pageUrl}
        <dt>Page</dt>
        <dd><button class="link" title={gif.pageUrl} onclick={onopenpage}>{linkText(gif.pageUrl)}</button></dd>
      {/if}
      {#if gif.sourceUrl}
        <dt>Source</dt>
        <dd><button class="link" title={gif.sourceUrl} onclick={onopensource}>{linkText(gif.sourceUrl)}</button></dd>
      {/if}
    </dl>
  </div>
</div>

<style>
  .details {
    box-sizing: border-box;
    height: 100%;
    display: flex;
    gap: 16px;
    padding: 16px;
    background: var(--bg);
  }
  .preview {
    position: relative;
    flex: none;
    width: 332px;
    height: 332px;
    border-radius: 8px;
    background: var(--surface);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
  .badge {
    position: absolute;
    top: 4px;
    left: 4px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    background: var(--scrim);
    color: #fff;
    font-size: 12px;
    line-height: 16px;
  }
  .meta {
    width: 340px;
    min-width: 0;
    overflow-y: auto;
  }
  h2 {
    margin: 0;
    font: var(--heading);
    color: var(--text);
    overflow-wrap: anywhere;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    overflow: hidden;
  }
  .provider {
    margin: 8px 0 0;
    font: var(--label);
    color: var(--text-secondary);
  }
  dl {
    margin: 16px 0 0;
    display: grid;
    grid-template-columns: 80px minmax(0, 1fr);
    gap: 8px;
    align-items: baseline;
  }
  dt {
    font: var(--label);
    color: var(--text-secondary);
  }
  dd {
    margin: 0;
    min-width: 0;
    font: var(--body);
    color: var(--text);
    overflow-wrap: anywhere;
  }
  .link {
    display: block;
    max-width: 100%;
    padding: 0;
    border: 0;
    background: none;
    font: var(--body);
    color: var(--accent);
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
  }
  .link:hover {
    text-decoration: underline;
  }
</style>
