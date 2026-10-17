// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/02-keyboard.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop.
// Arrows move the selection (not the caret), Esc clears then hides (a no-op in the server
// build), ⌘K is bound with no effect, and focus never leaves the search input.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
const input = page.locator("input[role=combobox]");
await input.fill("dwight");
await page.locator("#tile-3 img").waitFor({ timeout: 10000 });

const active = () => input.getAttribute("aria-activedescendant");
const caret = () => input.evaluate((el) => el.selectionStart);
if ((await active()) !== "tile-0") fail("initial selection " + (await active()));
if ((await input.getAttribute("aria-controls")) !== "grid") fail("aria-controls");
if ((await page.locator("#grid").getAttribute("role")) !== "listbox") fail("grid role");
const caret0 = await caret();

for (const [key, want] of [
  ["ArrowRight", "tile-1"],
  ["ArrowDown", "tile-5"],
  ["ArrowUp", "tile-1"],
  ["ArrowLeft", "tile-0"],
  ["ArrowLeft", "tile-0"], // clamped at the first tile
]) {
  await input.press(key);
  if ((await active()) !== want) fail(`${key}: aria-activedescendant ${await active()}, want ${want}`);
  if ((await page.locator("#" + want).getAttribute("aria-selected")) !== "true") fail(`${want} not aria-selected`);
}
if ((await input.inputValue()) !== "dwight") fail("query changed by arrows");
if ((await caret()) !== caret0) fail(`caret moved: ${caret0} -> ${await caret()}`);

// Clicking a tile selects it and keeps focus in the input.
await page.locator("#tile-2").click();
if ((await active()) !== "tile-2") fail("click did not select tile-2");
if (!(await input.evaluate((el) => document.activeElement === el))) fail("click took focus from the input");

// ↓ into the last 2 rows of the first page (24) loads page 2 and keeps the selection.
for (let i = 0; i < 4; i++) await input.press("ArrowDown");
if ((await active()) !== "tile-18") fail("ArrowDown x4 from tile-2: " + (await active()));
await page.waitForFunction("parseFloat(document.querySelector('#grid > div').style.height) > 16 + 6 * 152 + 16 + 48", null, { timeout: 10000 });
if ((await active()) !== "tile-18") fail("selection lost when page 2 appended");

const before = await page.locator("main").innerHTML();
await input.press("Meta+k");
await input.press("Tab");
if ((await page.locator("main").innerHTML()) !== before) fail("⌘K/Tab changed the UI");
if (!(await input.evaluate((el) => document.activeElement === el))) fail("Tab moved focus");

await input.press("Escape");
if ((await input.inputValue()) !== "") fail("Esc did not clear the query");
if (await page.locator("[role=option]").count()) fail("grid still shown after clearing the query");
await input.press("Escape"); // empty query: PanelService.Hide (no-op in the server build)
await page.waitForTimeout(300);
if (!(await input.evaluate((el) => document.activeElement === el))) fail("input lost focus after Esc");
if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
