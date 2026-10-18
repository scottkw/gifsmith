// Favorites and Recents grids: provider sections in slot order, never blended (D-05, D-06).
// Pure functions, no Svelte or DOM imports, so node --test can load them.
import type { Gif } from './api';
import { COLS } from './grid.ts';
import { SLOTS, type SlotId } from './providers.ts';

// start is the flat index of the section's first item (the selection index runs through all sections).
export type Section = { provider: SlotId; label: string; items: Gif[]; start: number };
export type Row =
  | { kind: 'header'; section: number; size: number }
  | { kind: 'tiles'; section: number; start: number; items: Gif[]; size: number };

export const HEADER_ROW = 32;
export const TILE_ROW = 152; // 144 tile + 8 gap

// Groups gifs (newest first) by provider; filter is a trimmed, case-insensitive title substring.
export function sections(gifs: Gif[], filter: string): Section[] {
  const f = filter.trim().toLowerCase();
  const keep = f ? gifs.filter((g) => (g.title ?? '').toLowerCase().includes(f)) : gifs;
  const out: Section[] = [];
  let start = 0;
  for (const s of SLOTS) {
    const items = keep.filter((g) => g.provider === s.id);
    if (!items.length) continue;
    out.push({ provider: s.id, label: s.label, items, start });
    start += items.length;
  }
  return out;
}

export const flatItems = (secs: Section[]): Gif[] => secs.flatMap((s) => s.items);

// Virtual rows: each section is a header row, then its tiles in rows of COLS.
export function rows(secs: Section[]): Row[] {
  const out: Row[] = [];
  secs.forEach((s, section) => {
    out.push({ kind: 'header', section, size: HEADER_ROW });
    for (let i = 0; i < s.items.length; i += COLS) {
      out.push({ kind: 'tiles', section, start: s.start + i, items: s.items.slice(i, i + COLS), size: TILE_ROW });
    }
  });
  return out;
}

// skeleton (RED)
export function moveSectioned(index: number, _dir: 'left' | 'right' | 'up' | 'down', _sizes: number[]): number {
  return index;
}
