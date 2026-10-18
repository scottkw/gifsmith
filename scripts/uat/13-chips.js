// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/13-chips.js
// Headless only, fixture keys only. The server build keeps searches.json in bin/uat-config (wiped by
// the smoke script each run), never the user's Application Support folder.
// A search is saved after 1.5 s without typing when it found something, or at once when a GIF from
// it is copied; an empty box shows the provider's saved searches as chips that rerun from the keyboard.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
const input = page.locator("header input[role=combobox]");
const row = page.locator("[role=listbox][aria-label='Recent searches']");
const chips = row.locator("[role=option]");
const chipTexts = async () => (await chips.allInnerTexts()).map((t) => t.trim());
const active = () => input.getAttribute("aria-activedescendant");
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });
// Types q, waits for its first page, then waits past the 1.5 s idle rule.
const searchAndIdle = async (q, ready = "#tile-0[aria-selected=true] img") => {
  await input.fill(q);
  await page.locator(ready).first().waitFor({ timeout: 10000 });
  await page.waitForTimeout(1700);
};
const clear = async () => {
  await input.press("Escape"); // clears the query (it is non-empty)
  if (await input.inputValue()) fail("Esc did not clear the query");
};

// 1. Finer Gifs: "office" is saved by the idle rule; the chip row shows above S8 (no grid).
await page.keyboard.press("Meta+4");
await placeholder("Search The Finer Gifs Club");
if (await row.count()) fail("chip row with no saved searches");
await searchAndIdle("office");
await clear();
await row.waitFor({ timeout: 5000 });
if (JSON.stringify(await chipTexts()) !== JSON.stringify(["office"])) fail("chips " + JSON.stringify(await chipTexts()));
if ((await chips.first().getAttribute("aria-label")) !== "Search again for office") fail("chip aria-label");
const rb = await row.boundingBox();
const vb = await page.locator(".view").boundingBox();
if (Math.round(rb.y) !== 84 || Math.round(vb.y) !== 116 || Math.round(vb.height) !== 332) fail(`row ${JSON.stringify(rb)} view ${JSON.stringify(vb)}`);
const s8 = await page.locator(".view .message").boundingBox();
if (Math.round(s8.y) !== 116) fail("S8 not directly below the chip row: " + JSON.stringify(s8));
// No grid: ↓ selects the first chip, ↓ again stays.
await input.press("ArrowDown");
if ((await active()) !== "chip-0") fail("↓ with no grid did not select chip 0: " + (await active()));
await input.press("ArrowDown");
if ((await active()) !== "chip-0") fail("↓ on a chip with no grid moved");
// ↵ reruns at once: query, results, row hidden, first tile selected.
await input.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
if ((await input.inputValue()) !== "office") fail("rerun query " + (await input.inputValue()));
if (await row.count()) fail("chip row shown with a query");

// 2. D-20: a search with no results is never saved.
await searchAndIdle("empty", ".view .message");
await clear();
await row.waitFor({ timeout: 5000 });
if ((await chipTexts()).includes("empty")) fail("a search with no results was saved");

// 3. Save on copy, before 1.5 s: "dwight" becomes the newest chip.
await input.fill("dwight");
const t0 = Date.now();
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await input.press("Enter");
await page.locator("footer .center.success", { hasText: "Copied" }).waitFor({ timeout: 10000 });
if (Date.now() - t0 >= 1500) fail("copy took past the idle window; the test proves nothing");
await input.fill("");
await page.waitForFunction("document.querySelector('#chip-0')?.title === 'dwight'", null, { timeout: 5000 });
if (JSON.stringify(await chipTexts()) !== JSON.stringify(["dwight", "office"])) fail("chips after copy " + JSON.stringify(await chipTexts()));

// 4. GIPHY: ↑ from the first row selects a chip; ← → clamp; ⌘⌫ removes silently; ↓ back to tile 0.
await page.keyboard.press("Meta+1");
const field = page.locator(".card input");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
if (await row.count()) fail("Finer chips shown on GIPHY");
await searchAndIdle("cats");
await searchAndIdle("dogs");
await clear();
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await row.waitFor({ timeout: 5000 });
if (JSON.stringify(await chipTexts()) !== JSON.stringify(["dogs", "cats"])) fail("GIPHY chips " + JSON.stringify(await chipTexts()));
if (Math.round((await page.locator("#grid").boundingBox()).height) !== 332) fail("grid not 332px under chips");
await input.press("ArrowRight"); // tile 1, still the first row
await input.press("ArrowUp");
if ((await active()) !== "chip-0") fail("↑ from the first row: " + (await active()));
if ((await chips.first().getAttribute("aria-selected")) !== "true") fail("chip 0 not marked selected");
if (await page.locator(".tile[aria-selected=true]").count()) fail("a tile stays selected with a chip selected");
const ring = await chips.first().evaluate((el) => getComputedStyle(el).outlineStyle + " " + getComputedStyle(el).outlineWidth);
if (ring !== "solid 2px") fail("chip ring " + ring);
await saveScreenshot(await page.screenshot(), "gifsmith-13-chips.png");
await input.press("ArrowRight");
await input.press("ArrowRight"); // clamped
if ((await active()) !== "chip-1") fail("→ did not clamp at chip 1: " + (await active()));
await input.press("Meta+Backspace");
await page.waitForFunction("document.querySelectorAll('[aria-label=\"Recent searches\"] [role=option]').length === 1", null, { timeout: 5000 });
if (JSON.stringify(await chipTexts()) !== JSON.stringify(["dogs"])) fail("after ⌘⌫ " + JSON.stringify(await chipTexts()));
if ((await active()) !== "chip-0") fail("selection after removal " + (await active()));
if ((await page.locator("footer .center").innerText()).trim()) fail("removal showed a toast");
if (await input.inputValue()) fail("⌘⌫ edited the input");
await input.press("ArrowDown");
if ((await active()) !== "tile-0") fail("↓ from a chip: " + (await active()));
// Typing clears the chip selection and the row hides once the query is non-empty.
await input.press("ArrowUp");
if ((await active()) !== "chip-0") fail("↑ again");
await input.press("a");
if ((await input.inputValue()) !== "a") fail("typed value " + (await input.inputValue()));
if (await row.count()) fail("chip row shown after typing");
if ((await active()) === "chip-0") fail("chip still selected after typing");
await clear();
await row.waitFor({ timeout: 5000 });

// 5. Library tabs never show chips.
await page.keyboard.press("Meta+6");
await placeholder("Search Recents");
if (await row.count()) fail("chips on Recents");
await page.keyboard.press("Meta+5");
await placeholder("Search Favorites");
if (await row.count()) fail("chips on Favorites");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
