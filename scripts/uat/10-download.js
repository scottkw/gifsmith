// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/10-download.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop. The
// server build saves into bin/uat-downloads (never ~/Downloads) and runs platform.Fake, so
// nothing reaches Finder. ⌘S downloads and keeps the panel as it was, the toast offers Show in
// Finder (click or ⌘R) for 5 s, ⌘R never reloads the page, and HTML media shows F3.
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
const link = page.locator("footer .center button.link");
const toast = page.locator("footer .center.success", { hasText: DONE });
const download = async () => {
  await input.press("Meta+s");
  await toast.waitFor({ timeout: 10000 });
};
const gone = async (what) => {
  await toast.waitFor({ state: "detached", timeout: 2000 }).catch(() => fail(`${what}: the download toast is still up`));
  if (await page.locator("footer .center.error").count()) fail(`${what}: error row ${await center.innerText()}`);
  if (await link.count()) fail(`${what}: the action link is still up`);
};
const notReloaded = async (what) => {
  if ((await page.evaluate("window.__marker")) !== 1) fail(`${what}: ⌘R reloaded the page`);
};
await page.evaluate("window.__marker = 1");

// ⌘S: message, then the Show in Finder link and the ⌘R hint; the panel state is unchanged.
await download();
if (!(await page.locator("footer .center.success svg.check").count())) fail("no check icon on the download toast");
const msg = (await page.locator("footer .center .msg").innerText()).trim();
if (msg !== DONE) fail(`toast message is "${msg}", want "${DONE}"`);
if ((await link.innerText()).trim() !== "Show in Finder") fail("no Show in Finder link");
if ((await page.locator("footer .center .hint").innerText()).trim() !== "⌘R") fail("no ⌘R hint");
const layout = await page.evaluate(`(() => {
  const s = (q) => getComputedStyle(document.querySelector(q));
  return [s("footer .center button.link").flexShrink, s("footer .center .hint").flexShrink, s("footer .center .msg").textOverflow];
})()`);
if (JSON.stringify(layout) !== JSON.stringify(["0", "0", "ellipsis"])) fail("toast layout " + JSON.stringify(layout));
if (!(await page.locator("#tile-1[aria-selected=true]").count())) fail("selection changed after ⌘S");
if (!(await page.locator("#tile-0 img").count())) fail("grid gone after ⌘S");
if ((await input.inputValue()) !== "dwight") fail("query changed after ⌘S");

// Click Show in Finder: Go reveals the file (platform.Fake here) and the toast clears.
await link.click();
await gone("click Show in Finder");

// ⌘R while the toast is visible does the same, without reloading.
await download();
await input.press("Meta+r");
await gone("⌘R");
await notReloaded("⌘R with the toast up");

// The next keystroke (a selection change) clears it.
await download();
await input.press("ArrowLeft");
await gone("ArrowLeft");

// After 5 s the toast is gone and ⌘R does nothing.
await download();
await page.waitForTimeout(5300);
await gone("5 s expiry");
await input.press("Meta+r");
await page.waitForTimeout(300);
await notReloaded("⌘R with no toast");
if ((await center.innerText()).trim() !== "Office moment 1") fail(`⌘R with no toast changed the footer: ${await center.innerText()}`);

const F3 = "The Finer Gifs Club sent something that isn't a GIF. Nothing was saved. Try another GIF.";
await input.fill("htmlmedia");
await page.locator("#tile-0[aria-selected=true] svg.broken").waitFor({ timeout: 10000 });
await input.press("Meta+s");
await page.locator("footer .center.error", { hasText: F3 }).waitFor({ timeout: 10000 });
if (await page.locator("footer .center button").count()) fail("an error toast offers an action");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
