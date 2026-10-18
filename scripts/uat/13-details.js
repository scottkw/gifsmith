// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/13-details.js
// Headless only, fixture keys only. The server build runs platform.Fake, so the Page and Source
// links open nothing on the desktop.
// ⌘I shows the selected item's Details from the metadata it already carries; ← → step items,
// Esc closes with the selection kept, ⌘K opens the menu in the content area's top-right corner
// without "Show Details", and a printable key closes Details and edits the query.
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
const details = page.locator("#details");
const term = (t) => details.locator("dt", { hasText: t });
const value = (t) => details.locator(`dt:text-is("${t}") + dd`);
const saveKey = async (slot) => {
  await page.keyboard.press(slot);
  const field = page.locator(".card input");
  await field.waitFor({ timeout: 5000 });
  await field.fill("uat-key-123");
  await field.press("Enter");
  await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
};

// 1. GIPHY trending, item 1: every field from the result, medium rendition as the preview.
await saveKey("Meta+1");
await input.press("Meta+i");
await details.waitFor({ timeout: 2000 });
const src = await details.locator(".preview img").getAttribute("src");
if (!src.endsWith("/media/giphy-1-200h.gif")) fail("preview src " + src);
if ((await details.locator(".preview img").getAttribute("alt")) !== "GIPHY moment 1") fail("preview alt");
if ((await details.locator("h2").innerText()).trim() !== "GIPHY moment 1") fail("title");
if ((await details.locator(".provider").innerText()).trim() !== "GIPHY") fail("provider line " + (await details.locator(".provider").innerText()));
if ((await value("Dimensions").innerText()).trim() !== "480 × 480") fail("dimensions " + (await value("Dimensions").innerText()));
if ((await value("File size").innerText()).trim() !== "6 KB") fail("file size " + (await value("File size").innerText()));
if (await term("Tags").count()) fail("GIPHY has no tags in the fixture");
const pageLink = value("Page").locator("button");
if (!(await pageLink.innerText()).startsWith("giphy.com/gifs/")) fail("page link text " + (await pageLink.innerText()));
if ((await pageLink.getAttribute("title")) !== "https://giphy.com/gifs/moment-giphy-1") fail("page link title");
const sourceLink = value("Source").locator("button");
if ((await sourceLink.innerText()).trim() !== "example.org/source/giphy-1") fail("source link " + (await sourceLink.innerText()));
const terms = (await details.locator("dt").allInnerTexts()).map((t) => t.trim());
if (JSON.stringify(terms) !== JSON.stringify(["Dimensions", "File size", "Page", "Source"])) fail("terms " + JSON.stringify(terms));
if (!(await page.locator("footer .hints").innerText()).includes("Esc Back")) fail("details footer hints");
const box = await details.boundingBox();
if (box.y !== 84 || box.height !== 364 || box.width !== 720) fail("details area " + JSON.stringify(box));
const pbox = await details.locator(".preview").boundingBox();
if (pbox.width !== 332 || pbox.height !== 332) fail("preview box " + JSON.stringify(pbox));
await saveScreenshot(await page.screenshot(), "gifsmith-13-details.png");

// Links open through Go (platform.Fake here): no error toast, Details stays.
await pageLink.click();
await sourceLink.click();
await page.waitForTimeout(300);
if (await page.locator("footer .center.error").count()) fail("link error: " + (await page.locator("footer").innerText()));
if (!(await details.isVisible())) fail("Details closed by a link click");

// 2. → shows item 2 (no source: term and value omitted); the grid selection follows.
await input.press("ArrowRight");
await details.locator("h2", { hasText: "GIPHY moment 2" }).waitFor({ timeout: 2000 });
if (await term("Source").count()) fail("Source row for item 2");
if ((await page.locator("#tile-1").getAttribute("aria-selected")) !== "true") fail("grid selection did not follow →");
await input.press("ArrowLeft");
await input.press("ArrowLeft"); // clamped at the first item
await details.locator("h2", { hasText: "GIPHY moment 1" }).waitFor({ timeout: 2000 });
await input.press("ArrowRight");
await details.locator("h2", { hasText: "GIPHY moment 2" }).waitFor({ timeout: 2000 });

// 3. Esc closes with the selection kept; ⌘I toggles.
await input.press("Escape");
await details.waitFor({ state: "detached", timeout: 2000 });
if ((await page.locator("#tile-1").getAttribute("aria-selected")) !== "true") fail("selection not kept on Esc");
if ((await input.getAttribute("aria-activedescendant")) !== "tile-1") fail("aria-activedescendant after Esc");
await input.press("Meta+i");
await details.waitFor({ timeout: 2000 });
await input.press("Meta+i");
await details.waitFor({ state: "detached", timeout: 2000 });

// 4. Show Details from the ⌘K menu.
const dialog = page.locator("[role=dialog][aria-label=Actions]");
const filter = dialog.locator("input[role=combobox]");
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
await filter.type("details");
await filter.press("Enter");
await details.waitFor({ timeout: 2000 });
await input.press("Escape");
await details.waitFor({ state: "detached", timeout: 2000 });

// 5. Klipy: tags, no Page row; ⌘K inside Details sits 16px inside the content area's top-right.
await saveKey("Meta+3");
await input.press("Meta+i");
await details.locator("h2", { hasText: "Klipy moment 1" }).waitFor({ timeout: 2000 });
if ((await value("Tags").innerText()).trim() !== "funny, cat") fail("tags " + (await value("Tags").innerText()));
if (await term("Page").count()) fail("Page row on Klipy");
if (!(await details.locator(".preview img").getAttribute("src")).endsWith("/media/klipy-1-md.gif")) fail("Klipy preview is not md");
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
const d = await dialog.boundingBox();
if (Math.abs(d.x + d.width - 704) > 1 || Math.abs(d.y - 100) > 1) fail("menu in Details at " + JSON.stringify(d));
const labels = (await dialog.locator("[role=option]").allInnerTexts()).map((t) => t.split("\n")[0].trim());
if (labels.includes("Show Details")) fail("Show Details listed inside Details");
await filter.press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });
if (!(await details.isVisible())) fail("Esc in the menu closed Details too");

// 6. A printable key closes Details and goes to the input.
await input.press("x");
await details.waitFor({ state: "detached", timeout: 2000 });
if (!(await input.inputValue()).endsWith("x")) fail("input value " + (await input.inputValue()));

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
