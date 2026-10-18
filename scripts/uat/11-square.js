// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/11-square.js
// Headless only, fixture keys only. The server build runs platform.Fake (copies never reach the
// real pasteboard; Accessibility is untrusted). ⌘K → "square" → ↵ crops the 48x32 Finer fixture
// to a square in Go and confirms with the T1s toast.
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

// Copy Square GIF: T1s toast.
const T1S = "Copied square GIF — press ⌘V to paste";
await menuRun("square");
await page.locator("footer .center.success", { hasText: T1S }).waitFor({ timeout: 10000 });
const msg = (await center.innerText()).trim();
if (!msg.includes(T1S)) fail(`footer is "${msg}", want "${T1S}"`);

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
