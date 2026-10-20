// Reduce Motion (UI-SPEC): with it on, only the selected tile animates (none when nothing is
// selected); without it every tile does.
export const animates = (index: number, selected: number, reduced: boolean): boolean => !reduced || index === selected;
