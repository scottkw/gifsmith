// Recent-search chips (UI-SPEC "Recent-search chip row", D-12, D-13, D-20). No Svelte or DOM imports
// so node --test can load it. Go owns the list rules (dedupe, cap 20, 100 runes).
export const CHIP_MAX = 8;
export const SAVE_IDLE_MS = 1500;

export type ChipCtx = { tab: 'provider' | 'favorites' | 'recents'; query: string; card: boolean; searches?: string[] };

// Library tabs never show chips; a setup or Clips card replaces the grid area entirely.
export function chipsVisible(c: ChipCtx): boolean {
  return c.tab === 'provider' && !c.query.trim() && !c.card && !!c.searches?.length;
}

// D-13/D-20: saved after 1.5 s without typing, and only when the search found something.
export function shouldSave(s: { sinceKeyMs: number; firstPageItems: number; query: string }): boolean {
  return s.sinceKeyMs >= SAVE_IDLE_MS && s.firstPageItems >= 1 && [...s.query.trim()].length >= 2;
}
