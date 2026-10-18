// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/08-action-menu.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop. The
// server build runs platform.Fake, so copies never reach the real pasteboard.
// ⌘K opens the action menu beside the selected tile with the filter focused and the first row
// highlighted; ↓ ↓ ↵ runs Copy Link through the same dispatcher as ⌘⇧C; Esc closes it with the
// selection kept and focus back in the search input.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await page.keyboard.press("Meta+4");
await page.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
const input = page.locator("header input[role=combobox]");
await input.fill("dwight");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });

const dialog = page.locator("[role=dialog][aria-label=Actions]");
const filter = dialog.locator("input[role=combobox]");
const options = dialog.locator("[role=option]");
const highlighted = async () => {
  const id = await filter.getAttribute("aria-activedescendant");
  const sel = dialog.locator("[role=option][aria-selected=true]");
  if ((await sel.count()) !== 1) fail(`${await sel.count()} highlighted rows`);
  if ((await sel.getAttribute("id")) !== id) fail(`aria-activedescendant ${id} is not the highlighted row`);
  return (await sel.locator(".label").innerText()).trim();
};
const isFocused = (loc) => loc.evaluate((el) => document.activeElement === el);

// Open: filter focused, first row highlighted. Finer Gifs has no page URL: 6 rows.
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
if (!(await isFocused(filter))) fail("filter input not focused");
if ((await filter.getAttribute("placeholder")) !== "Search actions") fail("filter placeholder");
if ((await highlighted()) !== "Copy GIF") fail("first row not highlighted");
const labels = (await options.allInnerTexts()).map((t) => t.split("\n")[0].trim());
const want = ["Copy GIF", "Paste GIF", "Copy Link", "Paste Link", "Copy Markdown", "Paste Markdown"];
if (JSON.stringify(labels) !== JSON.stringify(want)) fail("rows " + JSON.stringify(labels));
if ((await options.nth(2).getAttribute("aria-keyshortcuts")) !== "Meta+Shift+C") fail("Copy Link aria-keyshortcuts");

// ↓ ↓ ↵ runs Copy Link and closes the menu first.
await filter.press("ArrowDown");
await filter.press("ArrowDown");
if ((await highlighted()) !== "Copy Link") fail("ArrowDown x2 did not reach Copy Link");
await filter.press("Enter");
await dialog.waitFor({ state: "detached", timeout: 2000 });
await page.locator("footer .center.success", { hasText: "Copied link" }).waitFor({ timeout: 5000 });
if (!(await isFocused(input))) fail("focus not back in the search input after running a row");

// Esc closes in one press even with filter text; selection and query unchanged.
await input.press("ArrowRight"); // clears the toast
await input.press("ArrowLeft");
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
await filter.type("li");
await filter.press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });
if ((await input.getAttribute("aria-activedescendant")) !== "tile-0") fail("selection changed");
if ((await input.inputValue()) !== "dwight") fail("query changed");
if (!(await isFocused(input))) fail("focus not back in the search input after Esc");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
