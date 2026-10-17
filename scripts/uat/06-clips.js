// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips.js
// Headless only, fixture keys only. Fresh state: saving the GIPHY key runs the Clips probe
// (fixture answers 200), the Clips pill turns on, ⌘2 shows clip tiles under the GIPHY mark,
// and Enter on a clip shows E8 without copying.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });

const input = page.locator("input[role=combobox]");
const footer = page.locator("footer");
const clipsTab = page.locator("[role=tab]").nth(1);
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });

if ((await clipsTab.getAttribute("aria-disabled")) !== "true") fail("Clips enabled before any GIPHY key");
await page.locator(".card input").fill("uat-key-123");
await page.getByRole("button", { name: "Save Key" }).click();
await footer.filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
await page.waitForFunction("!document.querySelectorAll('[role=tab]')[1].hasAttribute('aria-disabled')", null, { timeout: 5000 });
if (await clipsTab.getAttribute("title")) fail("enabled Clips pill still has the unavailable title");

await page.keyboard.press("Meta+2");
await placeholder("Search GIPHY Clips");
await page.locator("#tile-3 img").waitFor({ timeout: 8000 });
if (!(await page.locator("#tile-0 img").getAttribute("src")).includes("/media/clip-1-200w.gif")) fail("Clips tile is not a clip preview");
if (!(await footer.locator('img[alt="Powered by GIPHY"]').count())) fail("Clips grid without the GIPHY mark");
await saveScreenshot(await page.screenshot(), "gifsmith-06-clips.png");

// E8: Enter and ⌘Enter on a clip tile keep the panel and make no copy.
for (const key of ["Enter", "Meta+Enter"]) {
  await input.press(key);
  await footer.filter({ hasText: "Copying Clips (MP4) arrives in the next update." }).waitFor({ timeout: 3000 });
  if (!(await page.locator("#tile-0 img").count())) fail(`${key} left the Clips grid`);
}

// Clips search goes to the Clips endpoint too.
await input.fill("cats");
await page.waitForFunction("(document.querySelector('#tile-0 img')?.src ?? '').includes('/media/clip-')", null, { timeout: 8000 });

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
