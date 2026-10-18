// Action menu placement. No Svelte or DOM imports so node --test can load it.
export const MENU_W = 320;

export function placeMenu(
  _tile: { x: number; y: number; width: number; height: number },
  _col: number,
  _menuH: number,
  _win = { w: 720, h: 480 },
): { left: number; top: number } {
  return { left: 0, top: 0 };
}
