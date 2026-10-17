// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/05-setup.js
// Headless only. Fixture keys only: the fake API accepts "uat-key-123", rejects "bad*", and treats
// "slow-key-1" / "ratelimit-key" as a timeout / 429. Never clicks "Get a free key ↗": in the server
// build Browser.OpenURL would open the user's real default browser, so only its href is asserted.
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
const content = page.locator("section.content");
const footer = page.locator("footer");
const field = page.locator(".card input");
const save = page.getByRole("button", { name: "Save Key" });
const message = page.locator("#setup-message");
const focused = (loc) => loc.evaluate((el) => document.activeElement === el);
const heading = (t) => page.locator(".card h2", { hasText: t }).waitFor({ timeout: 5000 });
const say = (t) => message.filter({ hasText: t }).waitFor({ timeout: 8000 });
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });
const tiles = async (n) => {
  await page.locator(`#tile-${n - 1} img`).waitFor({ timeout: 8000 });
};
const noKeyInPage = async () => {
  const text = await page.evaluate(() => document.body.innerText + document.body.innerHTML);
  for (const k of ["uat-key-123", "bad-key-123"]) if (text.includes(k)) fail(`page contains ${k}`);
};

// S1 first run: GIPHY card, search disabled with its placeholder, Save Key disabled, key field focused.
await heading("Add your GIPHY key");
if (!(await input.isDisabled())) fail("search input enabled under the setup card");
await placeholder("Search GIPHY");
if (!(await save.isDisabled())) fail("Save Key enabled with an empty field");
if (!(await focused(field))) fail("key field not focused");
const fieldAttrs = await field.evaluate((el) => ({
  type: el.type,
  sec: getComputedStyle(el).webkitTextSecurity,
  ac: el.getAttribute("autocomplete"),
  cap: el.getAttribute("autocapitalize"),
  sp: el.getAttribute("spellcheck"),
  ph: el.placeholder,
  h: el.offsetHeight,
}));
if (JSON.stringify(fieldAttrs) !== JSON.stringify({ type: "text", sec: "disc", ac: "off", cap: "off", sp: "false", ph: "Paste your GIPHY API key", h: 32 }))
  fail("key field attributes " + JSON.stringify(fieldAttrs));
const link = page.getByRole("link", { name: "Get a free key ↗" });
if ((await link.getAttribute("href")) !== "https://developers.giphy.com/dashboard/") fail("GIPHY key link href");
const card = await page.locator(".card").boundingBox();
const area = await content.boundingBox();
if (card.y + card.height > area.y + area.height) fail("setup card runs past the content area");
if (!(await footer.locator('img[alt="Powered by GIPHY"]').count())) fail("GIPHY mark missing under the setup card");
await saveScreenshot(await page.screenshot(), "gifsmith-05-setup-s1.png");

// Tab order inside the card: field -> Save Key -> link -> Try Finer Gifs -> field.
await field.fill("x");
const order = [];
for (let i = 0; i < 4; i++) {
  await page.keyboard.press("Tab");
  order.push(await page.evaluate(() => document.activeElement.textContent.trim() || document.activeElement.tagName));
}
if (order.join("|") !== "Save Key|Get a free key ↗|Try Finer Gifs (no key)|INPUT") fail("Tab order " + order.join("|"));
// Esc clears a non-empty key field (an empty one hides the panel: a no-op here).
await page.keyboard.press("Escape");
if ((await field.inputValue()) !== "") fail("Esc did not clear the key field");

// S1a then K3: the timeout scenario keeps "Verifying…" up for the 3 s client timeout.
await field.fill("slow-key-1");
await field.press("Enter");
await page.getByRole("button", { name: "Verifying…" }).waitFor({ timeout: 2000 });
if (!(await field.evaluate((el) => el.readOnly))) fail("key field not readonly while verifying");
if (await link.isDisabled()) fail("link disabled while verifying");
await say("Couldn't reach GIPHY to check the key. Check your connection, then try again.");

// K4: a 429 with Retry-After 120.
await field.fill("ratelimit-key");
await field.press("Enter");
await say("GIPHY's request limit was hit while checking the key. Try again in 2 minutes.");

// K2 (S1b): rejected key; field keeps its text, is focused and selected, border destructive; editing clears.
await field.fill("bad-key-123");
await field.press("Enter");
await say("GIPHY rejected that key. Check that you copied the whole key, then try again.");
const sel = await field.evaluate((el) => [document.activeElement === el, el.selectionStart, el.selectionEnd, el.value.length]);
if (!sel[0] || sel[1] !== 0 || sel[2] !== sel[3] || sel[3] !== 11) fail("rejected key not focused and selected " + JSON.stringify(sel));
const border = await field.evaluate((el) => getComputedStyle(el).borderTopColor);
const destructive = await page.evaluate(() => {
  const d = document.createElement("div");
  d.style.color = "var(--destructive)";
  document.body.appendChild(d);
  const c = getComputedStyle(d).color;
  d.remove();
  return c;
});
if (border !== destructive) fail(`field border ${border}, want ${destructive}`);
await saveScreenshot(await page.screenshot(), "gifsmith-05-setup-k2.png");
await page.keyboard.type("z");
if ((await message.innerText()).trim()) fail("editing did not clear the message");

// S1c: saved -> GIPHY trending, K5 in the footer, focus to search, key gone from the page.
await field.fill("uat-key-123");
await save.click();
await footer.filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
await tiles(4);
if (await field.count()) fail("key field still present after save");
if (!(await focused(input))) fail("search input not focused after save");
await noKeyInPage();
await saveScreenshot(await page.screenshot(), "gifsmith-05-setup-saved.png");

// S13: the stored key is rejected during a search -> the card again with K6.
await input.fill("revoked");
await heading("Add your GIPHY key");
await say("GIPHY rejected your saved key. It may have been revoked. Paste a new key to continue.");
await saveScreenshot(await page.screenshot(), "gifsmith-05-setup-k6.png");

// Klipy without a key: same card, the disabled search keeps "Search KLIPY".
await page.keyboard.press("Meta+3");
await heading("Add your Klipy key");
await placeholder("Search KLIPY");
if (!(await input.isDisabled())) fail("search input enabled under the Klipy card");
if ((await link.getAttribute("href")) !== "https://partner.klipy.com") fail("Klipy key link href");
await field.fill("bad-key-123");
await field.press("Enter");
await say("Klipy rejected that key. Check that you copied the whole key, then try again.");
await field.fill("uat-key-123");
await save.click();
await footer.filter({ hasText: "Klipy key saved" }).waitFor({ timeout: 8000 });
await tiles(4);
if (!(await footer.innerText()).includes("Powered by KLIPY")) fail("Klipy attribution");
await noKeyInPage();

// Per-provider state: GIPHY still shows its K6 card.
await page.keyboard.press("Meta+1");
await heading("Add your GIPHY key");
await say("GIPHY rejected your saved key.");

// Try Finer Gifs (no key) -> Finer Gifs, S8, remembered.
await page.getByRole("button", { name: "Try Finer Gifs (no key)" }).click();
await placeholder("Search The Finer Gifs Club");
await content.filter({ hasText: "Search The Office GIFs" }).waitFor({ timeout: 5000 });
if (!(await focused(input))) fail("search input not focused after Try Finer Gifs");
await page.waitForTimeout(300);
await page.reload();
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await placeholder("Search The Finer Gifs Club");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
