// Search-pane option lists (UI-SPEC "Search"): Go owns the allowed codes, the webview names them.
import type { ActionId } from './actions';
import type { Gif } from './api';

export type Option = { value: string; label: string };

const languages = new Intl.DisplayNames(['en'], { type: 'language' });
const regions = new Intl.DisplayNames(['en'], { type: 'region' });
// GIPHY's zh-CN/zh-TW are script choices and tl is Filipino; name them that way.
const LANG_ALIAS: Record<string, string> = { 'zh-CN': 'zh-Hans', 'zh-TW': 'zh-Hant', tl: 'fil' };

const languageName = (c: string) => languages.of(LANG_ALIAS[c] ?? c) ?? c;
const regionName = (c: string) => regions.of(c.toUpperCase()) ?? c;

// "Match System ({name})" first, then the codes by English name A–Z, ties by code.
function options(codes: string[], system: string, name: (c: string) => string): Option[] {
  const named = codes.map((value) => ({ value, label: name(value) }));
  named.sort((a, b) => a.label.localeCompare(b.label, 'en') || (a.value < b.value ? -1 : a.value > b.value ? 1 : 0));
  return [{ value: '', label: `Match System (${name(system)})` }, ...named];
}

export const languageOptions = (codes: string[], system: string) => options(codes, system, languageName);
export const countryOptions = (codes: string[], system: string) => options(codes, system, regionName);
export const ratingLabel = (v: string) => v.toUpperCase();

// ↵/⌘↵ choices in Settings order (UI-SPEC "Default actions"); Go stores and validates the id.
export const DEFAULT_ACTIONS: { id: ActionId; label: string; short: string }[] = [
  { id: 'copy', label: 'Copy GIF', short: 'Copy' },
  { id: 'paste', label: 'Paste GIF', short: 'Paste' },
  { id: 'copyLink', label: 'Copy Link', short: 'Copy Link' },
  { id: 'pasteLink', label: 'Paste Link', short: 'Paste Link' },
  { id: 'copyMarkdown', label: 'Copy Markdown', short: 'Copy Markdown' },
  { id: 'pasteMarkdown', label: 'Paste Markdown', short: 'Paste Markdown' },
  { id: 'download', label: 'Download GIF', short: 'Download' },
];

type ActionPrefs = { primaryAction: string; secondaryAction: string; klipyFileActions: boolean };
const FILE_ACTIONS: string[] = ['copy', 'paste', 'download'];

// The action ↵ (primary) or ⌘↵ (secondary) runs on g. While Klipy file actions are gated (RELS-04),
// a file action on a Klipy item runs as Copy Link (↵) or Paste Link (⌘↵); a Clip runs the chosen
// action, which Go performs on its MP4.
export function effectiveAction(g: Gif | undefined, which: 'primary' | 'secondary', p: ActionPrefs): ActionId {
  const id = (which === 'primary' ? p.primaryAction : p.secondaryAction) as ActionId;
  if (g?.provider === 'klipy' && !p.klipyFileActions && FILE_ACTIONS.includes(id)) return which === 'primary' ? 'copyLink' : 'pasteLink';
  return id;
}

// Footer short label of an action (UI-SPEC "Footer hints").
export const shortLabel = (id: ActionId): string => DEFAULT_ACTIONS.find((a) => a.id === id)?.short ?? id;

// Grid size (UI-SPEC "Grid size"): columns, tile height and skeleton tiles; '' or unknown = medium.
const GRID: Record<string, { cols: number; tile: number }> = {
  small: { cols: 5, tile: 112 },
  medium: { cols: 4, tile: 144 },
  large: { cols: 3, tile: 192 },
};
const grid = (size: string) => GRID[size] ?? GRID.medium;
export const colsFor = (size: string): number => grid(size).cols;
export const tileHeight = (size: string): number => grid(size).tile;
// S2: two rows of skeleton tiles.
export const skeletonCount = (size: string): number => 2 * grid(size).cols;
