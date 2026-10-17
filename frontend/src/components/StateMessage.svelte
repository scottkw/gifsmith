<script lang="ts">
  // Centered content-area message (UI-SPEC S7–S12, C1, C2): optional alert icon, heading, body, an optional
  // primary link (C2's external link, an <a> styled as a button) and one button (secondary, or primary for C1).
  let {
    heading,
    body,
    icon = false,
    button = '',
    primary = false,
    onbutton,
    link = '',
    href = '',
    onlink,
  }: {
    heading: string;
    body: string;
    icon?: boolean;
    button?: string;
    primary?: boolean;
    onbutton?: () => void;
    link?: string;
    href?: string;
    onlink?: () => void;
  } = $props();
</script>

<div class="message" role="status">
  {#if icon}
    <svg class="alert" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="9.25" />
      <path d="M12 7.5v5.5" />
      <circle class="dot" cx="12" cy="16.5" r="0.6" />
    </svg>
  {/if}
  <h2>{heading}</h2>
  <p>{body}</p>
  {#if link || button}
    <div class="actions">
      {#if link}
        <a
          class="primary"
          {href}
          onclick={(e) => {
            e.preventDefault();
            onlink?.();
          }}>{link}</a
        >
      {/if}
      {#if button}
        <button class:primary onclick={onbutton}>{button}</button>
      {/if}
    </div>
  {/if}
</div>

<style>
  .message {
    box-sizing: border-box;
    max-width: 432px;
    margin: 0 auto;
    padding-top: 32px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
  .alert {
    width: 24px;
    height: 24px;
    margin-bottom: 8px;
    fill: none;
    stroke: var(--text-secondary);
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .alert .dot {
    fill: var(--text-secondary);
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
  .actions {
    margin-top: 16px;
    display: flex;
    gap: 8px;
  }
  button,
  a {
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    text-decoration: none;
    height: 32px;
    padding: 0 16px;
    border: 0;
    border-radius: 6px;
    background: var(--surface-active);
    color: var(--text);
    font: 600 14px/1 var(--font);
    cursor: pointer;
  }
  .primary {
    background: var(--accent-fill);
    color: #ffffff;
  }
</style>
