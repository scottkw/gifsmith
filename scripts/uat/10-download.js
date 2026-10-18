// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/10-download.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop. The
// server build saves into bin/uat-downloads (never ~/Downloads) and runs platform.Fake, so
// nothing reaches Finder. ⌘S downloads and keeps the panel as it was, ⌘R never reloads the
// page, and HTML media shows F3.
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
const input = page.locator("input[role=combobox]");
const center = page.locator("footer .center");
await input.fill("dwight");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await input.press("ArrowRight");
await page.locator("#tile-1[aria-selected=true]").waitFor({ timeout: 2000 });

// The server build's folder display name is "uat-downloads"; the desktop build says "Downloads".
const DONE = "Downloaded to uat-downloads";
await input.press("Meta+s");
await page.locator("footer .center.success", { hasText: DONE }).waitFor({ timeout: 10000 });
if (!(await page.locator("footer .center.success svg.check").count())) fail("no check icon on the download toast");
const msg = (await center.innerText()).trim();
if (msg !== DONE) fail(`toast message is "${msg}", want "${DONE}"`);
if (!(await page.locator("#tile-1[aria-selected=true]").count())) fail("selection changed after ⌘S");
if (!(await page.locator("#tile-0 img").count())) fail("grid gone after ⌘S");
if ((await input.inputValue()) !== "dwight") fail("query changed after ⌘S");

// ⌘R is captured: the page must not reload.
await page.evaluate("window.__marker = 1");
await input.press("Meta+r");
await page.waitForTimeout(300);
if ((await page.evaluate("window.__marker")) !== 1) fail("⌘R reloaded the page");

const F3 = "The Finer Gifs Club sent something that isn't a GIF. Nothing was saved. Try another GIF.";
await input.fill("htmlmedia");
await page.locator("#tile-0[aria-selected=true] svg.broken").waitFor({ timeout: 10000 });
await input.press("Meta+s");
await page.locator("footer .center.error", { hasText: F3 }).waitFor({ timeout: 10000 });
if (await page.locator("footer .center button").count()) fail("an error toast offers an action");

if (errors.length) fail("page error: " + errors[0]);
console.log(`center after F3: ${await center.innerText()}`);
console.log("UAT OK");
