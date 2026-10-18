// Hotkey recorder logic (UI-SPEC "Hotkeys"): KeyboardEvent → Wails accelerator, glyphs.
export type KeyLike = { key: string; code: string; ctrlKey: boolean; altKey: boolean; metaKey: boolean; shiftKey: boolean };
export type Recorded = { kind: 'accel'; accel: string } | { kind: 'held'; glyphs: string } | { kind: 'no_modifier' } | { kind: 'unsupported' };

export const DEFAULT_OPEN = 'Cmd+Ctrl+Option+G';
export const DENYLIST = new Set<string>();
export const accelFromEvent = (e: KeyLike): Recorded => ({ kind: 'unsupported' });
export const glyphs = (accel: string) => '';
export const canonical = (accel: string) => accel;
export const hasModifier = (accel: string) => true;
