// Action menu table (UI-SPEC "Action list"). No Svelte or DOM imports so node --test can load it.
import type { Gif } from './api';

export type ActionId =
  | 'copy'
  | 'paste'
  | 'copyLink'
  | 'pasteLink'
  | 'copyMarkdown'
  | 'pasteMarkdown'
  | 'copyPage'
  | 'open'
  | 'download'
  | 'copySquare'
  | 'pasteSquare'
  | 'favorite'
  | 'details'
  | 'removeRecent'
  | 'clearRecents';
export type ActionCtx = { gif: Gif; tab: 'provider' | 'favorites' | 'recents'; inDetails: boolean; favorited: boolean };
export type ActionRow = { id: ActionId; label: string; hint: string; keys: string; group: number; destructive: boolean };

export function buildActions(_ctx: ActionCtx): ActionRow[] {
  return [];
}

export function filterActions(rows: ActionRow[], _text: string): ActionRow[] {
  return rows;
}
