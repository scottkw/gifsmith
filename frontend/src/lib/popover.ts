// Action menu placement (UI-SPEC "Action menu popover"). No Svelte or DOM imports so node --test can load it.
export const MENU_W = 320;
const GAP = 8; // anchor offset and window inset

// Columns 1-2 open right of the tile, 3-4 left of it; top aligns with the tile, then everything is
// clamped to stay GAP inside the window.
export function placeMenu(
  tile: { x: number; y: number; width: number; height: number },
  col: number,
  menuH: number,
  win = { w: 720, h: 480 },
): { left: number; top: number } {
  const left = col < 2 ? tile.x + tile.width + GAP : tile.x - GAP - MENU_W;
  const clamp = (v: number, max: number) => Math.max(GAP, Math.min(v, max));
  return { left: clamp(left, win.w - GAP - MENU_W), top: clamp(tile.y, win.h - GAP - menuH) };
}

// In Details the menu hangs 16px inside the content area's top-right corner (content top = search 48 + tabs 36).
const CONTENT_TOP = 84;
const DETAILS_INSET = 16;
export function placeInDetails(_menuH: number): { left: number; top: number } {
  return { left: 720 - DETAILS_INSET - MENU_W, top: CONTENT_TOP + DETAILS_INSET };
}
