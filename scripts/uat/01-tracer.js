// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/01-tracer.js
// Search renders an animated 4-column grid of fixture GIFs; Enter stages the GIF and,
// in the CGO-free server build, the pasteboard stub rejects it, so the footer shows E7.
const E7 = "Couldn't copy this GIF. The download failed. Try again or pick another GIF.";
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
// Fresh state opens on GIPHY (plan 02-05); these checks run on Finer Gifs (⌘4).
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await page.keyboard.press("Meta+4");
await page.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
const input = page.locator("input[role=combobox]");
if ((await input.getAttribute("placeholder")) !== "Search The Finer Gifs Club") fail("placeholder");

await input.fill("dwight");
await page.locator("[role=option] img").nth(3).waitFor({ timeout: 10000 });
await page.waitForFunction("document.querySelector('#tile-0 img')?.naturalWidth > 0", null, { timeout: 10000 });
const src = await page.locator("#tile-0 img").getAttribute("src");
if (!/^http:\/\/127\.0\.0\.1:\d+\/media\/office-1\.gif$/.test(src)) fail("tile-0 src " + src);
if ((await page.locator("#tile-0").getAttribute("aria-selected")) !== "true") fail("tile-0 not selected");

// UI-SPEC geometry: 166x144 tiles, 8px gap, 16px padding, 4 columns.
const b0 = await page.locator("#tile-0").boundingBox();
const b3 = await page.locator("#tile-3").boundingBox();
const b4 = await page.locator("#tile-4").boundingBox();
if (b0.width !== 166 || b0.height !== 144) fail(`tile size ${b0.width}x${b0.height}`);
if (b0.x !== 16 || b3.y !== b0.y || b3.x !== 16 + 3 * 174 || b4.y !== b0.y + 152) fail("grid layout");
const fit = await page.locator("#tile-0 img").evaluate((img) => getComputedStyle(img).objectFit);
if (fit !== "contain") fail("object-fit " + fit);

const footer = page.locator("footer");
if (!(await footer.innerText()).includes("GIFs from The Finer Gifs Club")) fail("attribution");
if (!(await footer.innerText()).includes("Office moment 1")) fail("footer title");
await saveScreenshot(await page.screenshot(), "gifsmith-01-grid.png");

await input.press("Enter");
await page.locator("footer .center.error", { hasText: E7 }).waitFor({ timeout: 10000 });
await saveScreenshot(await page.screenshot(), "gifsmith-01-copy-error.png");

// Typing clears the transient error and the footer falls back to the selected title.
await input.press("End");
await input.type("s");
if ((await footer.innerText()).includes(E7)) fail("E7 survived a keystroke");

console.log("UAT OK");
