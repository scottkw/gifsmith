// Details pane formatters (UI-SPEC "Details pane"). No Svelte or DOM imports so node --test can load it.
import type { Gif } from './api';
import { providerName } from './errors.ts';

const TAGS_SHOWN = 12;

// '' = unknown: the Details row is omitted.
export function fileSize(bytes?: number): string {
  if (!bytes) return '';
  return bytes >= 1_000_000 ? `${(bytes / 1e6).toFixed(1)} MB` : `${Math.max(1, Math.round(bytes / 1000))} KB`;
}

export function dimensions(w?: number, h?: number): string {
  return w && h && w > 0 && h > 0 ? `${w} × ${h}` : '';
}

// Display text only; the click opens the URL Go holds, never this string.
export function linkText(url: string): string {
  return url.replace(/^https:\/\//, '');
}

export function tagsSummary(tags?: string[] | null): string {
  if (!tags?.length) return '';
  const more = tags.length - TAGS_SHOWN;
  return tags.slice(0, TAGS_SHOWN).join(', ') + (more > 0 ? ` +${more} more` : '');
}

export function providerLine(gif: Gif, favorited: boolean): string {
  return providerName(gif.provider) + (gif.isClip ? ' · MP4' : '') + (favorited ? ' · In Favorites' : '');
}
