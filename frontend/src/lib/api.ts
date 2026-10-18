// Thin facade over the generated bindings. Provider HTTP happens in Go; this only calls it.
import { ActionService, GifService, PanelService, SettingsService, type ProviderStatus } from '../../bindings/gifsmith/internal/services';
import type { Gif, Page } from '../../bindings/gifsmith/internal/provider';
import { Browser, CancelError, type CancellablePromise } from '@wailsio/runtime';
import { CLIPS_ACCESS_MAILTO, GIPHY_KEY_URL, KLIPY_KEY_URL } from './providers';

export type { Gif, Page, ProviderStatus };

export const search = (p: string, q: string, cursor: string) => GifService.Search(p, q, cursor);
export const trending = (p: string, cursor: string) => GifService.Trending(p, cursor);
// Key state arrives only as booleans; no call returns a key.
export const providers = async () => (await GifService.Providers()) ?? [];
export const lastProvider = () => SettingsService.LastProvider();
export const setLastProvider = (p: string) => SettingsService.SetLastProvider(p);
// Verifies with a real trending call, then stores in the Keychain; rejects with a provider cause.
export const saveKey = (p: string, key: string) => SettingsService.SaveKey(p, key);
// Re-probes Clips once when its result is unknown; resolves to the state, rejects with the probe's cause.
export const checkClips = () => GifService.CheckClips();

// The webview may only ask the OS to open these fixed URLs.
const EXTERNAL = new Set([GIPHY_KEY_URL, KLIPY_KEY_URL, CLIPS_ACCESS_MAILTO]);
export function openExternal(url: string) {
  if (!EXTERNAL.has(url)) throw new Error('refusing to open an unlisted URL');
  return Browser.OpenURL(url);
}

// One in-flight call per slot ("first:<provider>", "page:<provider>"): a new call cancels the
// previous one (its Go ctx and HTTP request). A cancelled or superseded call resolves to
// undefined and must not render; only the slot's latest call yields a result or an error.
const slots = new Map<string, CancellablePromise<unknown>>();

export function latest<T>(slot: string, p: CancellablePromise<T>): Promise<T | undefined> {
  slots.get(slot)?.cancel();
  slots.set(slot, p);
  const mine = () => slots.get(slot) === p;
  return p.then(
    (v) => {
      if (!mine()) return undefined;
      slots.delete(slot); // settled: a later cancel(slot) must not send a stray CancelCall
      return v;
    },
    (e) => {
      if (e instanceof CancelError || !mine()) return undefined;
      slots.delete(slot);
      throw e;
    },
  );
}

export function cancel(slot: string) {
  slots.get(slot)?.cancel();
  slots.delete(slot);
}

// Go's provider.Error arrives as err.cause (Options.MarshalError); anything else has no cause.
export type ProviderErr = { kind: string; provider: string; status?: number; retryAfterSec?: number };
export function providerError(e: unknown): ProviderErr | null {
  const c = (e as { cause?: unknown } | null)?.cause as ProviderErr | undefined;
  return c && typeof c.kind === 'string' ? c : null;
}

// Go's services.ActionError arrives as err.cause too: { kind } only.
export const actionError = (e: unknown): { kind: string } | null => providerError(e);

export const copy = (g: Gif) => ActionService.Copy(g);
// Go builds the text from g; the webview only names the kind.
export const copyText = (g: Gif, kind: 'link') => ActionService.CopyText(g, kind);
export const paste = (g: Gif) => ActionService.Paste(g);
export const requestAccessibility = () => ActionService.RequestAccessibility();
export const hide = () => PanelService.Hide();
