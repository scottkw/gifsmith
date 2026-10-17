<script lang="ts">
  // Centered content-area message (UI-SPEC S7–S12): optional alert icon, heading, body, one button (secondary, or primary for C1).
  let {
    heading,
    body,
    icon = false,
    button = '',
    primary = false,
    onbutton,
  }: { heading: string; body: string; icon?: boolean; button?: string; primary?: boolean; onbutton?: () => void } =
    $props();
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
  {#if button}
    <button class:primary onclick={onbutton}>{button}</button>
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
  button {
    margin-top: 16px;
    height: 32px;
    padding: 0 16px;
    border: 0;
    border-radius: 6px;
    background: var(--surface-active);
    color: var(--text);
    font: 600 14px/1 var(--font);
    cursor: pointer;
  }
  button.primary {
    background: var(--accent-fill);
    color: #ffffff;
  }
</style>
