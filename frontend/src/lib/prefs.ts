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
export const DEFAULT_ACTIONS: { id: ActionId; label: string }[] = [];

type ActionPrefs = { primaryAction: string; secondaryAction: string; klipyFileActions: boolean };

// The action ↵ (primary) or ⌘↵ (secondary) runs on g.
export function effectiveAction(g: Gif | undefined, which: 'primary' | 'secondary', p: ActionPrefs): ActionId {
  return 'copy';
}

// Footer short label of an action (UI-SPEC "Footer hints").
export const shortLabel = (id: ActionId): string => '';
