// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/18-grid-size.js
// Headless only, against the -tags server build. Page A is the panel, page B the Settings window's
// page. Search results and Trending sizes lay the panel out in 5 / 4 / 3 columns (112 / 144 / 192 px
// tiles); the action menu opens right or left by room, stays 8px inside and its bottom at or above
// 440 (P19); right-click selects a tile and opens the menu (P15). Fixture key only.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
const input = A.locator("header input[role=combobox]");

// Columns from the grid's computed template; the tile box is measured on #tile-0.
const GRID_JS = `(() => {
  const row = document.querySelector("#grid .tiles");
  const t = document.querySelector("#tile-0")?.getBoundingClientRect();
  return { cols: row ? getComputedStyle(row).gridTemplateColumns.split(" ").length : 0, w: t ? Math.round(t.width) : 0, h: t ? Math.round(t.height) : 0 };
})()`;
const layout = () => A.evaluate(GRID_JS);
const expectGrid = async (what, cols, h) => {
  await A.waitForFunction(`(() => { const g = ${GRID_JS}; return g.cols === ${cols} && g.h === ${h}; })()`, null, { timeout: 5000 }).catch(async () =>
    fail(`${what}: grid ${JSON.stringify(await layout())}, want ${cols} columns of ${h}px`),
  );
};

await A.keyboard.press("Meta+4");
await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
await input.fill("dwight");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await expectGrid("Medium search", 4, 144);
const medium = await layout();
if (Math.abs(medium.w - 166) > 1) fail("Medium tile width " + medium.w);

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Search", exact: true }).click();
await B.getByRole("heading", { name: "Search" }).waitFor({ timeout: 10000 });
const size = (group, name) => B.getByRole("radiogroup", { name: group, exact: true }).getByRole("radio", { name, exact: true });
if (!(await size("Search results", "Medium").isChecked()) || !(await size("Trending", "Medium").isChecked())) fail("grid sizes are not Medium by default");
await B.locator("text=Small fits 5 GIFs per row, Medium 4, Large 3.").waitFor({ timeout: 3000 });

await size("Search results", "Small").check();
await expectGrid("Small search", 5, 112);
await size("Search results", "Large").check();
await expectGrid("Large search", 3, 192);
await saveScreenshot(await A.screenshot(), "gifsmith-18-grid-large.png");

// Large: the 3rd tile's menu opens left of it (no room on the right), 8..712 and bottom <= 440.
await A.bringToFront();
const dialog = A.locator("[role=dialog][aria-label=Actions]");
const menuBox = async (what) => {
  await input.press("Meta+k");
  await dialog.waitFor({ timeout: 2000 });
  const d = await dialog.boundingBox();
  if (d.x < 8 - 0.5 || d.x + d.width > 712 + 0.5) fail(`${what}: menu spans ${d.x}..${d.x + d.width}`);
  if (d.y + d.height > 440 + 0.5) fail(`${what}: menu bottom ${d.y + d.height} covers the footer`);
  await dialog.locator("input[role=combobox]").press("Escape");
  await dialog.waitFor({ state: "detached", timeout: 2000 });
  return d;
};
await input.press("ArrowRight");
await input.press("ArrowRight");
await A.locator("#tile-2[aria-selected=true]").waitFor({ timeout: 2000 });
const third = await A.locator("#tile-2").boundingBox();
const d3 = await menuBox("Large 3rd tile");
if (Math.abs(d3.x + d3.width - (third.x - 8)) > 1 && d3.x !== 8) fail(`Large 3rd tile: menu at ${d3.x}, tile at ${third.x}`);
// A tile on the next row, whose top sits low in the panel: the menu moves up to clear the footer.
await input.press("ArrowDown");
await A.locator("#tile-5[aria-selected=true]").waitFor({ timeout: 2000 });
await A.waitForTimeout(200); // scrolled into view
await menuBox("Large lower-row tile");

// Right-click selects the tile and opens the menu beside it.
await A.locator("#tile-1").click({ button: "right" });
await A.locator("#tile-1[aria-selected=true]").waitFor({ timeout: 2000 });
await dialog.waitFor({ timeout: 2000 });
await dialog.locator("input[role=combobox]").press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });

// GIPHY trending keeps Medium until Trending changes; Favorites/Recents follow Search results.
await A.keyboard.press("Meta+1");
const field = A.locator(".card input");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await expectGrid("GIPHY trending", 4, 144);
await size("Trending", "Small").check();
await expectGrid("Small trending", 5, 112);
await A.bringToFront();
await input.press("Meta+Shift+F"); // favorite the selected GIF so Favorites has a grid
await A.locator("footer .center.success", { hasText: "Added to Favorites" }).waitFor({ timeout: 5000 });
await A.keyboard.press("Meta+5");
await A.locator("#tile-0").waitFor({ timeout: 5000 });
await expectGrid("Favorites", 3, 192);

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
