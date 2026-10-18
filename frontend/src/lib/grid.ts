// Grid selection math and key routing. No Svelte or DOM imports so node --test can load it.
export const COLS = 4;

export type Direction = 'left' | 'right' | 'up' | 'down';
export type KeyAction =
  | Direction
  | 'copy'
  | 'paste'
  | 'copyLink'
  | 'escape'
  | 'actions'
  | 'tab'
  | 'slot1'
  | 'slot2'
  | 'slot3'
  | 'slot4'
  | 'none';

// ← → wrap across rows and clamp at the ends; ↑ on the first row stays; ↓ with no tile directly
// below goes to the last tile when it sits on a later row.
export function move(index: number, key: Direction, count: number): number {
  if (count <= 0) return 0;
  const last = count - 1;
  switch (key) {
    case 'left':
      return Math.max(0, index - 1);
    case 'right':
      return Math.min(last, index + 1);
    case 'up':
      return index >= COLS ? index - COLS : index;
    case 'down':
      if (index + COLS <= last) return index + COLS;
      return Math.floor(index / COLS) < Math.floor(last / COLS) ? last : index;
  }
}

// The next page loads once the selection is within the last 2 rows.
export function nearEnd(index: number, count: number): boolean {
  return index >= count - 2 * COLS;
}

const ARROWS: Record<string, Direction> = { ArrowLeft: 'left', ArrowRight: 'right', ArrowUp: 'up', ArrowDown: 'down' };

export function routeKey(e: {
  key: string;
  metaKey: boolean;
  shiftKey: boolean;
  isComposing: boolean;
  keyCode: number;
  repeat: boolean;
}): KeyAction {
  // Keys inside an open IME composition (WebKit: keyCode 229) belong to the IME: arrows pick
  // candidates, Enter commits, Esc cancels.
  if (e.isComposing || e.keyCode === 229) return 'none';
  if (e.key === 'Enter') return e.repeat ? 'none' : e.metaKey ? 'paste' : 'copy';
  if (e.key === 'Escape') return 'escape';
  if (e.key === 'Tab') return 'tab';
  if (e.metaKey && e.key.toLowerCase() === 'k') return 'actions';
  // ⌘⇧ letters may arrive as "C" or "c"; plain ⌘C stays native copy in the input.
  if (e.metaKey && e.shiftKey && e.key.toLowerCase() === 'c') return 'copyLink';
  if (e.metaKey && ['1', '2', '3', '4'].includes(e.key)) return `slot${e.key}` as KeyAction;
  return ARROWS[e.key] ?? 'none';
}
