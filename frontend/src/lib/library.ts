// Favorites and Recents grids: provider sections in slot order, never blended (D-05, D-06).
// Pure functions, no Svelte or DOM imports, so node --test can load them.
import type { Gif } from './api';
import type { Direction } from './grid.ts';
import { SLOTS, type SlotId } from './providers.ts';

// start is the flat index of the section's first item (the selection index runs through all sections).
export type Section = { provider: SlotId; label: string; items: Gif[]; start: number };
export type Row =
  | { kind: 'header'; section: number; size: number }
  | { kind: 'tiles'; section: number; start: number; items: Gif[]; size: number };

export const HEADER_ROW = 32;

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

// Virtual rows: each section is a header row, then its tiles in rows of cols, tileRow px apart
// (tile height + 8 gap).
export function rows(secs: Section[], cols: number, tileRow: number): Row[] {
  const out: Row[] = [];
  secs.forEach((s, section) => {
    out.push({ kind: 'header', section, size: HEADER_ROW });
    for (let i = 0; i < s.items.length; i += cols) {
      out.push({ kind: 'tiles', section, start: s.start + i, items: s.items.slice(i, i + cols), size: tileRow });
    }
  });
  return out;
}

// Library-tab grid movement (UI-SPEC): ← → step through the flat order across sections; ↑ ↓ move
// by row inside a section (Phase 2 rule); ↓ from a section's last row goes to the same column of
// the next section's first row, ↑ from a first row to the previous section's last row, clamped.
// sizes are the item counts per section (none is 0).
export function moveSectioned(index: number, dir: Direction, sizes: number[], cols: number): number {
  const total = sizes.reduce((a, b) => a + b, 0);
  if (total <= 0) return 0;
  if (dir === 'left') return Math.max(0, index - 1);
  if (dir === 'right') return Math.min(total - 1, index + 1);
  let s = 0;
  let start = 0;
  while (index >= start + sizes[s]) start += sizes[s++];
  const off = index - start;
  const col = off % cols;
  const n = sizes[s];
  if (dir === 'down') {
    if (Math.floor(off / cols) < Math.floor((n - 1) / cols)) return start + Math.min(off + cols, n - 1);
    return s + 1 < sizes.length ? start + n + Math.min(col, sizes[s + 1] - 1) : index;
  }
  if (off >= cols) return index - cols;
  if (s === 0) return index;
  const pn = sizes[s - 1];
  return start - pn + Math.min(Math.floor((pn - 1) / cols) * cols + col, pn - 1);
}
