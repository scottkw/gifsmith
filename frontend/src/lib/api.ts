// Thin facade over the generated bindings. Provider HTTP happens in Go; this only calls it.
import { ActionService, GifService, PanelService } from '../../bindings/gifsmith/internal/services';
import type { Gif, Page } from '../../bindings/gifsmith/internal/provider';

export type { Gif, Page };

export const search = (p: string, q: string, cursor: string) => GifService.Search(p, q, cursor);
export const trending = (p: string, cursor: string) => GifService.Trending(p, cursor);

// Go's provider.Error arrives as err.cause (Options.MarshalError); anything else has no cause.
export type ProviderErr = { kind: string; provider: string; status?: number; retryAfterSec?: number };
export function providerError(e: unknown): ProviderErr | null {
  const c = (e as { cause?: unknown } | null)?.cause as ProviderErr | undefined;
  return c && typeof c.kind === 'string' ? c : null;
}

export const copy = (g: Gif) => ActionService.Copy(g);
export const paste = (g: Gif) => ActionService.Paste(g);
export const requestAccessibility = () => ActionService.RequestAccessibility();
export const hide = () => PanelService.Hide();
