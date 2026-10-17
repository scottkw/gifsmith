// Grid selection math and key routing. No Svelte or DOM imports so node --test can load it.
export const COLS = 4;

export type Direction = 'left' | 'right' | 'up' | 'down';
export type KeyAction = Direction | 'copy' | 'paste' | 'escape' | 'actions' | 'tab' | 'none';

export function move(_index: number, _key: Direction, _count: number): number {
  return -1; // RED skeleton
}

export function nearEnd(_index: number, _count: number): boolean {
  return false; // RED skeleton
}

export function routeKey(_e: { key: string; metaKey: boolean; isComposing: boolean; keyCode: number; repeat: boolean }): KeyAction {
  return 'none'; // RED skeleton
}
