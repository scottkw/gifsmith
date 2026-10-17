// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/05-tabs.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop.
// Fresh state (the harness wipes bin/uat-config): four fixed tabs, GIPHY first, per-slot
// placeholder and attribution, ⌘1–⌘4 switching, last provider persisted, per-provider state.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
const ready = () => page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await ready();

const input = page.locator("input[role=combobox]");
const content = page.locator("section.content");
const footer = page.locator("footer");
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });
const activeTab = () => page.locator("[role=tab][aria-selected=true]").innerText();

// Four pills, fixed order, ⌘1–⌘4 hints, none in the Tab order.
const tabs = await page.locator("[role=tab]").evaluateAll((els) =>
  els.map((e) => {
    const hint = e.querySelector(".hint").textContent;
    return { label: e.textContent.replace(hint, "").trim(), hint, tabindex: e.tabIndex };
  }),
);
const want = [
  ["GIPHY", "⌘1"],
  ["Clips", "⌘2"],
  ["Klipy", "⌘3"],
  ["Finer Gifs", "⌘4"],
];
if (tabs.length !== 4) fail(`${tabs.length} tabs`);
want.forEach(([label, hint], i) => {
  if (tabs[i].label !== label || tabs[i].hint !== hint) fail(`tab ${i}: ${JSON.stringify(tabs[i])}`);
  if (tabs[i].tabindex !== -1) fail(`tab ${label} is in the Tab order`);
});
const clips = page.locator("[role=tab]").nth(1);
if ((await clips.getAttribute("aria-disabled")) !== "true") fail("Clips not aria-disabled without a GIPHY key");
if ((await clips.getAttribute("title")) !== "Clips need a GIPHY key approved for Clips") fail("Clips title");

// First run: GIPHY active with the official mark.
if (!(await activeTab()).startsWith("GIPHY")) fail("first run not on GIPHY: " + (await activeTab()));
await placeholder("Search GIPHY");
const mark = footer.locator('img[alt="Powered by GIPHY"]');
if (!(await mark.getAttribute("src")).endsWith("/attribution/giphy-powered-light.png")) fail("mark src");
await page.waitForFunction("document.querySelector('footer img').naturalWidth === 641", null, { timeout: 5000 });
if ((await mark.boundingBox()).height !== 16) fail("mark height");
if (await footer.locator("a").count()) fail("attribution is a link");

await page.keyboard.press("Meta+4");
await placeholder("Search The Finer Gifs Club");
if (!(await footer.innerText()).includes("GIFs from The Finer Gifs Club")) fail("Finer attribution");
await content.filter({ hasText: "Search The Office GIFs" }).waitFor({ timeout: 5000 });
if (await footer.locator("img").count()) fail("GIPHY mark on Finer Gifs");

await page.keyboard.press("Meta+3");
await placeholder("Search KLIPY");
if (!(await footer.innerText()).includes("Powered by KLIPY")) fail("Klipy attribution");

await page.keyboard.press("Meta+2");
await placeholder("Search GIPHY Clips");
await content.filter({ hasText: "Clips use your GIPHY key" }).waitFor({ timeout: 5000 });
if (!(await footer.locator('img[alt="Powered by GIPHY"]').count())) fail("Clips without the GIPHY mark");
await saveScreenshot(await page.screenshot(), "gifsmith-05-clips-c1.png");

// Clicking a pill switches and leaves focus in the search input.
await page.locator("[role=tab]", { hasText: "Finer Gifs" }).click();
await placeholder("Search The Finer Gifs Club");
if (!(await input.evaluate((el) => document.activeElement === el))) fail("tab click left focus off the search input");

// Persisted through SetLastProvider: a reload opens on Finer Gifs.
await page.waitForTimeout(300);
await page.reload();
await ready();
if (!(await activeTab()).startsWith("Finer Gifs")) fail("reload opened on " + (await activeTab()));
await placeholder("Search The Finer Gifs Club");

// Per-provider state: Finer's grid survives a trip to GIPHY and back.
await input.fill("dwight");
await page.locator("#tile-3 img").waitFor({ timeout: 10000 });
await page.keyboard.press("Meta+1");
await placeholder("Search GIPHY");
if (await page.locator("#tile-0").count()) fail("Finer results shown on GIPHY");
if ((await input.inputValue()) !== "dwight") fail("query not kept across the switch");
await page.keyboard.press("Meta+4");
await placeholder("Search The Finer Gifs Club");
await page.locator("#tile-3 img").waitFor({ timeout: 2000 });
if (!(await footer.innerText()).includes("Office moment 1")) fail("Finer selection not restored");
await saveScreenshot(await page.screenshot(), "gifsmith-05-tabs.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
