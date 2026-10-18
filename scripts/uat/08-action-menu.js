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

// Open: filter focused, first row highlighted. Finer Gifs has no page URL: 9 rows.
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
if (!(await isFocused(filter))) fail("filter input not focused");
if ((await filter.getAttribute("placeholder")) !== "Search actions") fail("filter placeholder");
if ((await highlighted()) !== "Copy GIF") fail("first row not highlighted");
const labels = (await options.allInnerTexts()).map((t) => t.split("\n")[0].trim());
const want = ["Copy GIF", "Paste GIF", "Copy Link", "Paste Link", "Copy Markdown", "Paste Markdown", "Download GIF", "Copy Square GIF", "Paste Square GIF"];
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

const footer = page.locator("footer");
const openMenu = async () => {
  await input.press("Meta+k");
  await dialog.waitFor({ timeout: 2000 });
};
const closed = () => dialog.waitFor({ state: "detached", timeout: 2000 });
const noToast = async (what) => {
  await page.waitForTimeout(300);
  if (await page.locator("footer .center.success, footer .center.error").count()) fail(`${what}: footer ${await footer.innerText()}`);
};

// Filter: "mark" leaves the 2 Markdown rows, separators hidden, first match highlighted.
await openMenu();
if ((await dialog.locator(".sep").count()) !== 3) fail("Finer rows should have 3 separators (groups 1 to 4)");
await filter.type("mark");
if ((await options.count()) !== 2) fail(`"mark" left ${await options.count()} rows`);
if (await dialog.locator(".sep").count()) fail("separators shown while filtering");
if ((await highlighted()) !== "Copy Markdown") fail("first match not highlighted");

// No match: one "No matching actions" row that is not an option; ↵ runs nothing.
await filter.fill("zzz");
if (await options.count()) fail("options left for zzz");
const empty = dialog.locator(".empty");
if ((await empty.count()) !== 1 || (await empty.innerText()).trim() !== "No matching actions") fail("no-match row missing");
if (await filter.getAttribute("aria-activedescendant")) fail("aria-activedescendant set with no rows");
await filter.press("Enter");
if (!(await dialog.isVisible())) fail("↵ on no match closed the menu");
await noToast("↵ on no match");
await filter.press("Escape");
await closed();

// Mouse: hover moves the highlight, click runs the row and closes the menu.
await openMenu();
await options.nth(2).hover();
if ((await highlighted()) !== "Copy Link") fail("hover did not highlight Copy Link");
await options.nth(2).click();
await closed();
await page.locator("footer .center.success", { hasText: "Copied link" }).waitFor({ timeout: 5000 });
if (!(await isFocused(input))) fail("focus not in the search input after a click");
await input.press("ArrowRight"); // clears the toast
await input.press("ArrowLeft");

// A mousedown outside the popover closes it and runs nothing.
await openMenu();
await page.mouse.click(700, 24); // search bar, right of the input text
await closed();
await noToast("outside click");
if ((await input.getAttribute("aria-activedescendant")) !== "tile-0") fail("outside click changed the selection");

// Shortcuts still work while the menu is open.
await openMenu();
await filter.press("Meta+Shift+M");
await closed();
await page.locator("footer .center.success", { hasText: "Copied Markdown" }).waitFor({ timeout: 5000 });
await input.press("ArrowRight");
await input.press("ArrowLeft");
await openMenu();
await filter.press("Meta+Enter");
await closed();
await page.locator("footer button", { hasText: "Enable Automatic Paste…" }).waitFor({ timeout: 5000 });
if (!(await footer.innerText()).includes("press ⌘V to paste")) fail("⌘↵ from the menu: " + (await footer.innerText()));
await openMenu();
await filter.press("Meta+k");
await closed();

// Placement backstop: tile 0, and columns 1 and 4 on the bottom visible row, stay 8px inside 720x480.
const placed = async (tile, shot) => {
  await page.locator(`#${tile}[aria-selected=true]`).waitFor({ timeout: 2000 });
  await openMenu();
  const b = await dialog.boundingBox();
  if (!(b.x >= 8 && b.y >= 8 && b.x + b.width <= 712 && b.y + b.height <= 472)) fail(`${tile}: menu at ${JSON.stringify(b)}`);
  const t = await page.locator("#" + tile).boundingBox();
  const right = t.x < 360;
  if (right ? Math.abs(b.x - (t.x + t.width + 8)) > 1 : Math.abs(b.x + b.width - (t.x - 8)) > 1) fail(`${tile}: not beside the tile`);
  if (shot) await saveScreenshot(await page.screenshot(), shot);
  await filter.press("Escape");
  await closed();
};
await placed("tile-0");
await input.press("ArrowDown");
await input.press("ArrowDown");
await placed("tile-8", "gifsmith-08-menu-col1.png");
for (let i = 0; i < 3; i++) await input.press("ArrowRight");
await placed("tile-11", "gifsmith-08-menu-col4.png");

// Overflow (GIPHY: 11 rows + 3 separators > 224px): the list scrolls and keeps the highlight in view.
await page.keyboard.press("Meta+1");
const field = page.locator(".card input");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await openMenu();
if ((await options.count()) !== 11) fail(`GIPHY rows ${await options.count()}`);
const list = dialog.locator("[role=listbox]");
if (!(await list.evaluate((el) => el.scrollHeight > el.clientHeight && el.clientHeight <= 240))) fail("list does not scroll");
for (let i = 0; i < 12; i++) await filter.press("ArrowDown"); // clamped at the last row
if ((await highlighted()) !== "Paste Square GIF") fail("ArrowDown did not clamp at Paste Square GIF");
const inView = await dialog.locator("[role=option][aria-selected=true]").evaluate((el) => {
  const r = el.getBoundingClientRect();
  const l = el.parentElement.getBoundingClientRect();
  return r.top >= l.top && r.bottom <= l.bottom;
});
if (!inView) fail("highlighted row scrolled out of view");
await filter.press("Escape");
await closed();

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
