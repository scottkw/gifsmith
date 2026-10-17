// Provider error -> UI-SPEC copy rows. No Svelte or DOM imports so node --test can load it.
import type { ProviderErr } from './api';

export type ErrorState = { id: 'E1' | 'E3' | 'E4' | 'E5' | 'E5f'; heading: string; body: string; tryAgain: boolean; retryAt?: number };

export function providerName(id: string): string {
  return id;
}

export function errorState(e: ProviderErr, now: number, receivedAt = now): ErrorState | null {
  return { id: 'E5', heading: '', body: '', tryAgain: true };
}

export function pagingMessage(e: ProviderErr, now: number, receivedAt = now): string {
  return '';
}
