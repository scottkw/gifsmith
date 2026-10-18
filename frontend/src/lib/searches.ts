// Recent-search chips (UI-SPEC "Recent-search chip row", D-12, D-13, D-20). No Svelte or DOM imports
// so node --test can load it. Go owns the list rules (dedupe, cap 20, 100 runes).
export const CHIP_MAX = 8;
export const SAVE_IDLE_MS = 1500;

export type ChipCtx = { tab: 'provider' | 'favorites' | 'recents'; query: string; card: boolean; searches?: string[] };

export function chipsVisible(_c: ChipCtx): boolean {
  return false;
}

export function shouldSave(_s: { sinceKeyMs: number; firstPageItems: number; query: string }): boolean {
  return false;
}
