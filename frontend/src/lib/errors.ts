// Provider error -> UI-SPEC copy rows (E1–E5f, E2, E6). No Svelte or DOM imports so node --test can
// load it. Copy is built only from the provider name and the HTTP status: never a URL, body or key.
import type { ProviderErr } from './api';

export type ErrorState = { id: 'E1' | 'E3' | 'E4' | 'E5' | 'E5f'; heading: string; body: string; tryAgain: boolean; retryAt?: number };

const NAMES: Record<string, string> = {
  giphy: 'GIPHY',
  clips: 'GIPHY Clips',
  klipy: 'Klipy',
  finergifs: 'The Finer Gifs Club',
};
const SWITCH = 'Other providers still work: ⌘1–⌘4.';
const LIMIT = 'Free keys allow 100 requests per hour.';

export function providerName(id: string): string {
  return NAMES[id] ?? 'The provider';
}

// Seconds left in a known Retry-After window, measured from when the error arrived; null = unknown.
function remaining(e: ProviderErr, now: number, receivedAt: number): number | null {
  const sec = e.retryAfterSec ?? 0;
  return sec > 0 ? sec - (now - receivedAt) / 1000 : null;
}

const clock = (t: number) => new Date(t).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });

// `receivedAt` is when the error came back; re-render with a later `now` to count the retry down.
export function errorState(e: ProviderErr, now: number, receivedAt = now): ErrorState | null {
  const name = providerName(e.provider);
  switch (e.kind) {
    case 'key_missing':
    case 'key_rejected':
      return null; // setup card territory (plan 02-05)
    case 'rate_limited': {
      const heading = `${name} request limit reached`;
      const left = remaining(e, now, receivedAt);
      if (left === null) return { id: 'E1', heading, body: `${LIMIT} Try again later this hour. ${SWITCH}`, tryAgain: true };
      const retryAt = receivedAt + (e.retryAfterSec ?? 0) * 1000;
      // Once the window has passed the button says it; the countdown sentence goes.
      if (left <= 0) return { id: 'E1', heading, body: `${LIMIT} ${SWITCH}`, tryAgain: true, retryAt };
      const n = Math.ceil(left / 60);
      const when = left < 60 ? 'in less than a minute' : `in ${n} ${n === 1 ? 'minute' : 'minutes'} (at ${clock(retryAt)})`;
      return { id: 'E1', heading, body: `${LIMIT} Try again ${when}. ${SWITCH}`, tryAgain: false, retryAt };
    }
  }
  const status = e.status ?? 0;
  if (e.provider === 'finergifs' && (e.kind === 'network' || (status >= 500 && status < 600))) {
    return {
      id: 'E5f',
      heading: 'The Finer Gifs Club seems to be down',
      body: "Its server isn't responding right now. This happens from time to time. Try again later, or switch providers with ⌘1–⌘4.",
      tryAgain: true,
    };
  }
  if (e.kind === 'network') {
    return { id: 'E3', heading: `Can't reach ${name}`, body: 'Check your connection, VPN, proxy, or firewall, then try again.', tryAgain: true };
  }
  if (e.kind === 'timeout') {
    return {
      id: 'E4',
      heading: `${name} is taking too long`,
      body: 'The server may be slow right now. Try again, or switch providers with ⌘1–⌘4.',
      tryAgain: true,
    };
  }
  // http, or anything unrecognised. A 2xx status means an empty or invalid body (D-23): no "(HTTP 200)".
  const code = status && (status < 200 || status >= 300) ? ` (HTTP ${status})` : '';
  return { id: 'E5', heading: `${name} isn't responding`, body: `${name} returned an error${code}. ${SWITCH}`, tryAgain: true };
}

// Row shown in place of the next page (S9a E2 / S12a E6).
export function pagingMessage(e: ProviderErr, now: number, receivedAt = now): string {
  if (e.kind !== 'rate_limited') return "Couldn't load more.";
  const left = remaining(e, now, receivedAt);
  if (left === null) return 'Request limit reached. Try again later this hour.';
  if (left <= 0) return "Couldn't load more.";
  return `Request limit reached. More results in ${Math.ceil(left / 60)} min.`;
}
