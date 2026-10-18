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

type Row = Omit<ActionRow, 'label' | 'destructive'> & {
  label: string | ((ctx: ActionCtx) => string);
  applies?: (ctx: ActionCtx) => boolean;
  destructive?: boolean;
};

const media = (verb: string) => (c: ActionCtx) => `${verb} ${c.gif.isClip ? 'MP4' : 'GIF'}`;
const hasPage = (c: ActionCtx) => !!c.gif.pageUrl;
// A Clip's link and Markdown use its GIF rendition; without one there is nothing to link.
const linkable = (c: ActionCtx) => !c.gif.isClip || !!c.gif.gifUrl;

// Locked order (D-02). Later plans add rows here once their handlers exist; the menu never changes.
const ROWS: Row[] = [
  { id: 'copy', label: media('Copy'), hint: '↵', keys: 'Enter', group: 1 },
  { id: 'paste', label: media('Paste'), hint: '⌘↵', keys: 'Meta+Enter', group: 1 },
  { id: 'copyLink', label: 'Copy Link', hint: '⌘⇧C', keys: 'Meta+Shift+C', group: 2, applies: linkable },
  { id: 'pasteLink', label: 'Paste Link', hint: '', keys: '', group: 2, applies: linkable },
  { id: 'copyMarkdown', label: 'Copy Markdown', hint: '⌘⇧M', keys: 'Meta+Shift+M', group: 2, applies: linkable },
  { id: 'pasteMarkdown', label: 'Paste Markdown', hint: '', keys: '', group: 2, applies: linkable },
  { id: 'copyPage', label: 'Copy Page Link', hint: '', keys: '', group: 2, applies: hasPage },
  { id: 'open', label: 'Open in Browser', hint: '⌘O', keys: 'Meta+O', group: 3, applies: hasPage },
  { id: 'download', label: media('Download'), hint: '⌘S', keys: 'Meta+S', group: 3 },
  // A Clip is an MP4: no square crop (D-15).
  { id: 'copySquare', label: 'Copy Square GIF', hint: '', keys: '', group: 4, applies: (c) => !c.gif.isClip },
  { id: 'pasteSquare', label: 'Paste Square GIF', hint: '', keys: '', group: 4, applies: (c) => !c.gif.isClip },
];

// Rows that do not apply are omitted, never greyed (D-03).
export function buildActions(ctx: ActionCtx): ActionRow[] {
  return ROWS.filter((r) => r.applies?.(ctx) ?? true).map((r) => ({
    id: r.id,
    label: typeof r.label === 'string' ? r.label : r.label(ctx),
    hint: r.hint,
    keys: r.keys,
    group: r.group,
    destructive: r.destructive ?? false,
  }));
}

export function filterActions(rows: ActionRow[], text: string): ActionRow[] {
  const t = text.trim().toLowerCase();
  return t ? rows.filter((r) => r.label.toLowerCase().includes(t)) : rows;
}
