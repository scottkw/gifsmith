// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/12-library.js
// Headless only, fixture keys only. The server build keeps its library files in bin/uat-config
// (wiped by the smoke script each run), never the user's Application Support folder.
// Copies land in Recents (⌘6), grouped by provider in slot order, newest first; Recents renders
// from the saved snapshot, so it still works while GIPHY is rate-limited.
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
const center = page.locator("footer .center");
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });
const headers = () => page.locator(".header").allInnerTexts();
const label = (i) => page.locator(`#tile-${i}`).getAttribute("aria-label");
const COPIED = "Copied — press ⌘V to paste";
// Enter on the selected tile; waits for the copy toast, then clears it with a fresh status.
const copySelected = async () => {
  await input.press("Enter");
  await page.locator("footer .center.success", { hasText: COPIED }).waitFor({ timeout: 10000 });
};
const recents = async () => {
  await page.keyboard.press("Meta+6");
  await placeholder("Search Recents");
};

// 1. Copy a Finer Gifs result, then ⌘6.
await page.keyboard.press("Meta+4");
await placeholder("Search The Finer Gifs Club");
await input.fill("dwight");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await copySelected();
await input.fill("");
await recents();
const pill = page.locator("[role=tab][aria-selected=true]");
if (!(await pill.innerText()).startsWith("Recents")) fail("active pill is " + (await pill.innerText()));
await page.locator(".header", { hasText: "Finer Gifs" }).waitFor({ timeout: 5000 });
if (JSON.stringify(await headers()) !== JSON.stringify(["Finer Gifs · 1"])) fail("headers " + JSON.stringify(await headers()));
if ((await label(0)) !== "Office moment 1, Finer Gifs") fail("tile 0 label " + (await label(0)));
if ((await page.locator("#tile-0").getAttribute("aria-selected")) !== "true") fail("first Recents tile not selected");
if (!(await page.locator("footer").innerText()).includes("GIFs from The Finer Gifs Club")) fail("Finer attribution on Recents");
if (!(await page.locator("footer .hints").innerText()).includes("⌘⌫ Remove")) fail("library footer hints");

// 2. Back to Finer Gifs, copy the second result: it becomes the newest Recents item.
await page.keyboard.press("Meta+4");
await placeholder("Search The Finer Gifs Club");
await input.fill("dwight");
await page.locator("#tile-1 img").waitFor({ timeout: 10000 });
await input.press("ArrowRight");
await page.locator("#tile-1[aria-selected=true]").waitFor({ timeout: 2000 });
await copySelected();
await input.fill("");
await recents();
await page.locator(".header", { hasText: "Finer Gifs · 2" }).waitFor({ timeout: 5000 });
if ((await label(0)) !== "Office moment 2, Finer Gifs" || (await label(1)) !== "Office moment 1, Finer Gifs")
  fail(`order ${await label(0)} / ${await label(1)}, want the newest first`);

// 3. Save the GIPHY fixture key, copy a GIPHY tile: the GIPHY section sits above Finer Gifs.
await page.keyboard.press("Meta+1");
await page.locator(".card input").fill("uat-key-123");
await page.getByRole("button", { name: "Save Key" }).click();
await page.locator("footer", { hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await copySelected();
await recents();
await page.locator(".header", { hasText: "GIPHY · 1" }).waitFor({ timeout: 5000 });
if (JSON.stringify(await headers()) !== JSON.stringify(["GIPHY · 1", "Finer Gifs · 2"])) fail("headers " + JSON.stringify(await headers()));
if ((await label(0)) !== "GIPHY moment 1, GIPHY") fail("tile 0 label " + (await label(0)));
if (!(await page.locator('footer img[alt="Powered by GIPHY"]').count())) fail("GIPHY mark for a selected GIPHY item");

// 4. GIPHY rate-limited: Recents still renders every item and copies from the saved snapshot.
await page.keyboard.press("Meta+1");
await placeholder("Search GIPHY");
await input.fill("ratelimit");
await page.locator("section.content", { hasText: "GIPHY request limit reached" }).waitFor({ timeout: 10000 });
await recents();
await input.fill(""); // the query is the Recents title filter; clearing it makes no request
await page.locator("#tile-2").waitFor({ timeout: 5000 });
if ((await page.locator(".tile").count()) !== 3) fail(`${await page.locator(".tile").count()} Recents tiles, want 3`);
await copySelected();
if (!(await center.innerText()).includes(COPIED)) fail("no copy toast from Recents while rate-limited");

await saveScreenshot(await page.screenshot(), "gifsmith-12-recents.png");
if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
