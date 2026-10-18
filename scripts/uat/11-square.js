// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/11-square.js
// Headless only, fixture keys only. The server build runs platform.Fake (copies never reach the
// real pasteboard; Accessibility is untrusted). ⌘K → "square" → ↵ crops the 48x32 Finer fixture
// to a square in Go and confirms with the T1s toast; a fixture crop finishes well inside 300 ms, so
// neither the busy overlay nor "Cropping…" may appear. Paste Square GIF falls back to the T5a
// copy-only message. GIPHY "hugesize" (claimed 16 MB) shows F7 with the grid kept; Clips menus
// have no square rows.
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
const dialog = page.locator("[role=dialog][aria-label=Actions]");
const filter = dialog.locator("input[role=combobox]");
const center = page.locator("footer .center");

// Runs the first menu row matching text on the selected tile.
const menuRun = async (text) => {
  await input.press("Meta+k");
  await dialog.waitFor({ timeout: 2000 });
  await filter.type(text);
  await filter.press("Enter");
  await dialog.waitFor({ state: "detached", timeout: 2000 });
};

await input.fill("widemedia");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
if (!(await page.locator("#tile-0 img").getAttribute("src")).includes("/media/wide-1")) fail("widemedia did not return wide fixtures");

// Records whether the busy overlay or "Cropping…" ever shows (the 300 ms rule).
await page.evaluate(`(() => {
  window.__busySeen = [];
  new MutationObserver(() => {
    if (document.querySelector(".tile .overlay")) window.__busySeen.push("overlay");
    if ((document.querySelector("footer .center")?.textContent ?? "").includes("Cropping")) window.__busySeen.push("Cropping…");
  }).observe(document.body, { subtree: true, childList: true, characterData: true });
})()`);

// Copy Square GIF: T1s toast.
const T1S = "Copied square GIF — press ⌘V to paste";
await menuRun("square");
await page.locator("footer .center.success", { hasText: T1S }).waitFor({ timeout: 10000 });
const msg = (await center.innerText()).trim();
if (!msg.includes(T1S)) fail(`footer is "${msg}", want "${T1S}"`);
const seen = await page.evaluate("window.__busySeen");
if (seen.length) fail("a fast crop showed its busy state: " + JSON.stringify(seen));

// Paste Square GIF: the server Fake is untrusted, so it copies only (T5a) and offers Enable.
await menuRun("paste square");
await page.locator("footer .center", { hasText: "Copied — press ⌘V to paste. To paste automatically" }).waitFor({ timeout: 10000 });
if (!(await page.getByRole("button", { name: "Enable Automatic Paste…" }).count())) fail("no Enable Automatic Paste… after Paste Square GIF");

// GIPHY: a claimed size over 15 MB is refused with F7; the panel and grid stay.
await page.keyboard.press("Meta+1");
await page.locator(".card input").fill("uat-key-123");
await page.getByRole("button", { name: "Save Key" }).click();
await page.locator("footer", { hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 }); // trending
await input.fill("hugesize");
// Same fixture ids as trending, so wait out the 450 ms debounce and the search itself.
await page.waitForTimeout(700);
await page.locator("section.content[aria-busy=false] #tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await menuRun("copy square");
const F7 = "This GIF is too large to crop (over 15 MB). Press ↵ to copy it as is.";
await page.locator("footer .center.error", { hasText: F7 }).waitFor({ timeout: 5000 });
if (!(await page.locator("#tile-0 img").count())) fail("grid gone after F7");
if ((await input.inputValue()) !== "hugesize") fail("query changed after F7");

// Clips: no square rows.
await page.waitForFunction("!document.querySelectorAll('[role=tab]')[1].hasAttribute('aria-disabled')", null, { timeout: 5000 });
await page.keyboard.press("Meta+2");
await page.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search GIPHY Clips'", null, { timeout: 5000 });
await page.locator("#tile-0 img").waitFor({ timeout: 8000 });
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
const labels = (await dialog.locator("[role=option]").allInnerTexts()).map((t) => t.split("\n")[0].trim());
if (!labels.includes("Copy MP4") || labels.some((l) => l.includes("Square"))) fail("Clips menu rows " + JSON.stringify(labels));
await page.keyboard.press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
