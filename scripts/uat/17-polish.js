// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/17-polish.js
// Headless only, Finer Gifs fixture (no key). The server build runs platform.Fake (untrusted), so
// ⌘↵ copies only and shows T5a; nothing reaches the real pasteboard or desktop.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));
await page.setViewportSize({ width: 720, height: 480 });
await page.emulateMedia({ colorScheme: "light" });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
const input = page.locator("input[role=combobox]");
const center = page.locator("footer .center");

// P13: the footer centre is a polite live region.
if ((await center.getAttribute("role")) !== "status" || (await center.getAttribute("aria-live")) !== "polite") fail("footer centre is not a polite live region");

// P17: --text-tertiary is #737378 in light.
const tertiary = await page.evaluate(() => {
  const el = document.createElement("span");
  el.style.color = "var(--text-tertiary)";
  document.body.append(el);
  const c = getComputedStyle(el).color;
  el.remove();
  return c;
});
if (tertiary !== "rgb(115, 115, 120)") fail("--text-tertiary " + tertiary);

await page.keyboard.press("Meta+4");
await page.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
await input.fill("dwight");
await page.waitForFunction("document.querySelector('#tile-0 img')?.naturalWidth > 0", null, { timeout: 10000 });

// P16: Details' Esc Back is a button; clicking it closes Details.
await input.press("Meta+i");
const details = page.locator("#details");
await details.waitFor({ timeout: 3000 });
const back = page.getByRole("button", { name: "Close Details (Esc)" });
await back.waitFor({ timeout: 3000 });
await back.click();
await details.waitFor({ state: "detached", timeout: 3000 });
if ((await page.locator("#tile-0").getAttribute("aria-selected")) !== "true") fail("selection lost after closing Details");

// P14: T5a is the T1 text; the Enable button carries the explanation in its title.
await input.press("Meta+Enter");
await center.locator(".msg", { hasText: "Copied — press ⌘V to paste" }).waitFor({ timeout: 10000 });
const msg = (await center.locator(".msg").innerText()).trim();
if (msg !== "Copied — press ⌘V to paste") fail("T5a text: " + msg);
const enable = page.getByRole("button", { name: "Enable Automatic Paste…" });
if ((await enable.getAttribute("title")) !== "Turn on Accessibility for Gifsmith so pastes happen automatically") fail("Enable title");
// P21: the check icon's gap is a margin, not a text node (↵ copy shows the success toast T1).
await input.press("Enter");
await page.locator("footer .center.success .check").waitFor({ timeout: 10000 });
const gap = await center.locator(".check").evaluate((el) => [getComputedStyle(el).marginRight, el.nextSibling?.textContent ?? ""]);
if (gap[0] !== "4px" || /^\s/.test(gap[1])) fail("toast icon gap " + JSON.stringify(gap));
await saveScreenshot(await page.screenshot(), "gifsmith-17-polish-t5a.png");

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
