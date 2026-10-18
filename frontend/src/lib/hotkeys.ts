// Hotkey recorder logic (UI-SPEC "Hotkeys"): KeyboardEvent → Wails accelerator, glyphs, and the
// checks the recorder makes before asking Go (Go re-validates every accelerator and is authoritative).
export type KeyLike = { key: string; code: string; ctrlKey: boolean; altKey: boolean; metaKey: boolean; shiftKey: boolean };
export type Recorded = { kind: 'accel'; accel: string } | { kind: 'held'; glyphs: string } | { kind: 'no_modifier' } | { kind: 'unsupported' };

export const DEFAULT_OPEN = 'Cmd+Ctrl+Option+G';

// Display order ⌃⌥⇧⌘, each with its Wails modifier name.
const MODS = [
  ['ctrlKey', 'Ctrl', '⌃'],
  ['altKey', 'Option', '⌥'],
  ['shiftKey', 'Shift', '⇧'],
  ['metaKey', 'Cmd', '⌘'],
] as const;

// KeyboardEvent.code → Wails key name for everything but letters and digits. Never '+' (it splits
// the accelerator) and never 'super' (Wails renders it as Cmd).
const CODES: Record<string, string> = {
  Space: 'Space',
  ArrowUp: 'Up',
  ArrowDown: 'Down',
  ArrowLeft: 'Left',
  ArrowRight: 'Right',
  Enter: 'Return',
  Tab: 'Tab',
  Equal: '=',
  Minus: '-',
  BracketRight: ']',
  BracketLeft: '[',
  Quote: "'",
  Semicolon: ';',
  Backslash: '\\',
  Comma: ',',
  Slash: '/',
  Period: '.',
  Backquote: '`',
};

function keyName(e: KeyLike): string | null {
  // Letters from e.key: Wails binds them by the active layout's label. ⌥ turns e.key into © etc.,
  // so fall back to the QWERTY position in e.code (may differ on other layouts; Test catches it).
  if (/^[a-z]$/i.test(e.key)) return e.key.toUpperCase();
  let m = /^Key([A-Z])$/.exec(e.code);
  if (m) return m[1];
  m = /^Digit(\d)$/.exec(e.code);
  if (m) return m[1];
  m = /^F(\d{1,2})$/.exec(e.code);
  if (m && +m[1] >= 1 && +m[1] <= 20) return `F${+m[1]}`;
  return CODES[e.code] ?? null;
}

export function accelFromEvent(e: KeyLike): Recorded {
  if (['Meta', 'Control', 'Alt', 'Shift'].includes(e.key)) {
    return { kind: 'held', glyphs: MODS.filter(([k]) => e[k]).map(([, , g]) => g).join('') };
  }
  const key = keyName(e);
  if (!key) return { kind: 'unsupported' };
  if (!e.ctrlKey && !e.altKey && !e.metaKey) return { kind: 'no_modifier' };
  return { kind: 'accel', accel: [...MODS.filter(([k]) => e[k]).map(([, n]) => n), key].join('+') };
}

const parse = (accel: string) => {
  const parts = accel.split('+');
  return { mods: new Set(parts.slice(0, -1).map((m) => m.toLowerCase())), key: parts[parts.length - 1] };
};

const KEY_GLYPHS: Record<string, string> = { SPACE: 'Space', UP: '↑', DOWN: '↓', LEFT: '←', RIGHT: '→', RETURN: '↩', TAB: '⇥' };

// 'Cmd+Ctrl+Option+G' → '⌃⌥⌘G'.
export function glyphs(accel: string): string {
  if (!accel) return '';
  const { mods, key } = parse(accel);
  const k = key.toUpperCase();
  return MODS.filter(([, n]) => mods.has(n.toLowerCase())).map(([, , g]) => g).join('') + (KEY_GLYPHS[k] ?? k);
}

// Go's canonical form: modifiers sorted (Cmd, Ctrl, Option, Shift), upper-cased key.
export function canonical(accel: string): string {
  const { mods, key } = parse(accel);
  const names = MODS.map(([, n]) => n).filter((n) => mods.has(n.toLowerCase())).sort();
  return [...names, key.toUpperCase()].join('+');
}

export const hasModifier = (accel: string) => ['cmd', 'ctrl', 'option'].some((m) => parse(accel).mods.has(m));

// macOS shortcuts Gifsmith refuses (H3), canonical; mirrors Go's systemHotkeys.
export const DENYLIST = new Set(['Cmd+Q', 'Cmd+W', 'Cmd+H', 'Cmd+M', 'Cmd+TAB', 'Cmd+SPACE', 'Ctrl+SPACE', 'Cmd+Shift+3', 'Cmd+Shift+4', 'Cmd+Shift+5']);
