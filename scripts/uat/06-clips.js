// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips.js
// Headless only, fixture keys only. Fresh state: saving the GIPHY key runs the Clips probe
// (fixture answers 200), the Clips pill turns on, ⌘2 shows clip tiles under the GIPHY mark
// with the MP4 badge, ⌘K lists MP4 rows, and Enter on a clip copies its MP4 (platform.Fake).
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

// MP4 badge (aria-hidden, on --scrim) and the tile's accessible name.
const badge = page.locator("#tile-0 .badge");
if ((await badge.innerText()).trim() !== "MP4" || !(await badge.isVisible())) fail("first Clips tile has no visible MP4 badge");
if ((await badge.getAttribute("aria-hidden")) !== "true") fail("MP4 badge is not aria-hidden");
if (!(await page.locator("#tile-0").getAttribute("aria-label")).endsWith(", MP4")) fail("Clips tile aria-label lacks ', MP4'");
const bg = await badge.evaluate((el) => getComputedStyle(el).backgroundColor);
if (bg !== "rgba(0, 0, 0, 0.6)") fail("MP4 badge background " + bg);

// ⌘K: MP4 rows, no square rows.
const dialog = page.locator("[role=dialog][aria-label=Actions]");
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
const labels = (await dialog.locator("[role=option]").allInnerTexts()).map((t) => t.split("\n")[0].trim());
if (!labels.includes("Copy MP4") || !labels.includes("Paste MP4")) fail("menu rows " + JSON.stringify(labels));
if (labels.some((l) => l.includes("Square") || l.includes("GIF"))) fail("Clips menu offers a GIF row: " + JSON.stringify(labels));
await page.keyboard.press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });

// Enter copies the clip's MP4 (T1m); the E8 placeholder is gone.
await input.press("Enter");
await page.locator("footer .center.success", { hasText: "Copied MP4 — press ⌘V to paste" }).waitFor({ timeout: 8000 });
if ((await footer.innerText()).includes("next update")) fail("E8 placeholder still shown");

// Clips search goes to the Clips endpoint too.
await input.fill("cats");
await page.waitForFunction("(document.querySelector('#tile-0 img')?.src ?? '').includes('/media/clip-')", null, { timeout: 8000 });

// CR-02: a GIPHY key rejected on the Clips tab shows a way back, not a blank area.
await input.fill("revoked");
const content = page.locator("section.content");
await content.locator("h2", { hasText: "GIPHY rejected your saved key" }).waitFor({ timeout: 8000 });
// WR-01: with an empty query GIPHY still holds its cached trending grid (the real revoked-key case:
// nothing forces a new GIPHY request), so the key field must come from the carried-over rejection.
await input.fill("");
await content.locator("h2", { hasText: "GIPHY rejected your saved key" }).waitFor({ timeout: 3000 });
await page.getByRole("button", { name: "Go to GIPHY" }).click();
await placeholder("Search GIPHY");
await page.locator(".card input").waitFor({ timeout: 3000 });
await page.locator(".card input").fill("uat-key-456");
await page.getByRole("button", { name: "Save Key" }).click();
await footer.filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
await page.waitForFunction("!document.querySelectorAll('[role=tab]')[1].hasAttribute('aria-disabled')", null, { timeout: 5000 });
await page.keyboard.press("Meta+2");
await placeholder("Search GIPHY Clips");
await page.waitForFunction("(document.querySelector('#tile-0 img')?.src ?? '').includes('/media/clip-')", null, { timeout: 8000 });

// WR-02: the persisted Clips tab is restored (here the probe already passed in this server process).
await page.reload();
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await placeholder("Search GIPHY Clips");
await page.waitForFunction("(document.querySelector('#tile-0 img')?.src ?? '').includes('/media/clip-')", null, { timeout: 8000 });

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
