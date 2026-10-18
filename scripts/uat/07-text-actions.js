// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/07-text-actions.js
// Headless only: keys go to the page in the dev-browser Chromium, never to the desktop. The
// server build runs platform.Fake, so copies never reach the real pasteboard.
// ⌘⇧C copies the link (plain ⌘C stays native), a copy whose media is HTML shows F3, ⌘⇧M copies
// Markdown, and ⌘O opens a GIPHY page through Go (Fake records it, nothing opens) and does nothing
// on Klipy (no page URL). Fixture key uat-key-123 only.
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
const footer = page.locator("footer");
await input.fill("dwight");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });

// Plain ⌘C belongs to the input (native copy): no link toast.
await input.press("Meta+c");
await page.waitForTimeout(300);
if ((await footer.innerText()).includes("Copied link")) fail("⌘C ran Copy Link");

await input.press("Meta+Shift+C");
await page.locator("footer .center.success", { hasText: "Copied link" }).waitFor({ timeout: 5000 });
if (!(await page.locator("footer .center.success svg.check").count())) fail("no check icon on Copied link");

const F3 = "The Finer Gifs Club sent something that isn't a GIF. Nothing was saved. Try another GIF.";
await input.fill("htmlmedia");
await page.locator("#tile-0[aria-selected=true] svg.broken").waitFor({ timeout: 10000 });
await input.press("Enter");
await page.locator("footer .center.error", { hasText: F3 }).waitFor({ timeout: 10000 });

const field = page.locator(".card input");
const quiet = async (what) => {
  // Let "… key saved" (2 s) clear first, so any toast after this is the action's own.
  await page.locator("footer .center.success").waitFor({ state: "detached", timeout: 5000 });
  await input.press("Meta+o");
  await page.waitForTimeout(400);
  if (await page.locator("footer .center.error").count()) fail(`${what}: ⌘O showed an error: ${await footer.innerText()}`);
  if (await page.locator("footer .center.success").count()) fail(`${what}: ⌘O showed a toast: ${await footer.innerText()}`);
};

// GIPHY: Markdown and Open in Browser.
await page.keyboard.press("Meta+1");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await input.press("Meta+Shift+M");
await page.locator("footer .center.success", { hasText: "Copied Markdown" }).waitFor({ timeout: 5000 });
await input.press("ArrowRight"); // clears the toast
await quiet("GIPHY");

// Klipy: no page URL, so ⌘O does nothing.
await page.keyboard.press("Meta+3");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await quiet("Klipy");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
