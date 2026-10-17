// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/02-grid.js
// The 60-item fixture pages 24 + 24 + 12 into a row-virtualized grid that ends in "End of results",
// with only the visible rows (plus overscan) mounted and no duplicate-key error.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("console", (m) => {
  if (m.type() === "error") errors.push(m.text());
});
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
// Fresh state opens on GIPHY (plan 02-05); these checks run on Finer Gifs (⌘4).
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await page.keyboard.press("Meta+4");
await page.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
await page.locator("input[role=combobox]").fill("dwight");
await page.locator("#tile-3 img").waitFor({ timeout: 10000 });

const mounted = () => page.locator("[role=option] img").count();
const first = await mounted();
if (first < 4 || first >= 24) fail(`first page mounted ${first} images, want 4..23 (virtualized)`);

const ids = new Set();
const end = page.locator("#grid", { hasText: "End of results" });
for (let i = 0; i < 40 && !(await end.count()); i++) {
  for (const id of await page.locator("[role=option]").evaluateAll((els) => els.map((e) => e.id))) ids.add(id);
  await page.locator("#grid").evaluate((g) => (g.scrollTop = g.scrollHeight));
  await page.waitForTimeout(150);
}
if (!(await end.count())) fail("never reached End of results");
for (const id of await page.locator("[role=option]").evaluateAll((els) => els.map((e) => e.id))) ids.add(id);
if (ids.size !== 60 || !ids.has("tile-0") || !ids.has("tile-59")) fail(`saw ${ids.size} distinct tiles, want tile-0..tile-59`);
await saveScreenshot(await page.screenshot(), "gifsmith-02-end.png");

// Back at the top with all 60 loaded, only a few rows are mounted.
await page.locator("#grid").evaluate((g) => (g.scrollTop = 0));
await page.waitForTimeout(150);
const top = await mounted();
if (top >= 60) fail(`${top} images mounted at the top with 60 loaded`);
if ((await page.locator("#tile-0").getAttribute("aria-selected")) !== "true") fail("selection lost across pages");

const dup = errors.filter((e) => e.includes("each_key_duplicate"));
if (dup.length) fail("duplicate key error: " + dup[0]);
console.log(`mounted at top: ${top} of 60; distinct tiles seen: ${ids.size}`);
console.log("UAT OK");
