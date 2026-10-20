// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/16-settings-keys.js
// Headless only, against the -tags server build: fixture keys (bad-key-123 rejected, uat-key-123
// accepted) in an in-memory store, never the real Keychain. Never clicks "Get a free key ↗".
// Page A is the panel, page B the Settings window's page.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.locator(".card input").waitFor({ timeout: 5000 }); // GIPHY setup card: no key yet

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Keys", exact: true }).click();
await B.getByRole("heading", { name: "Keys", exact: true }).waitFor({ timeout: 10000 });

// The webview must never hold a key outside the field being typed in.
const noKeyInPage = async (when) => {
  const html = await B.evaluate(() => document.documentElement.outerHTML + document.body.innerText);
  for (const k of ["uat-key-123", "bad-key-123"]) if (html.includes(k)) fail(`${k} in the Settings DOM ${when}`);
};
// Every message is shown whole inside the card (UI-SPEC backstop): no horizontal overflow, at most two lines.
const fits = async (card) => {
  const [sw, cw, h] = await card.locator(".message").evaluate((el) => [el.scrollWidth, el.clientWidth, el.clientHeight]);
  if (sw > cw || h > 32) fail(`message does not fit: ${sw}px in ${cw}px, ${h}px tall`);
};
const giphy = B.getByRole("group", { name: "GIPHY" });
const klipy = B.getByRole("group", { name: "Klipy" });
const order = await B.locator(".card .name").allInnerTexts();
if (order.join(",") !== "GIPHY,Klipy") fail("card order " + order.join(","));
for (const c of [giphy, klipy]) await c.getByText("Not set").waitFor({ timeout: 5000 });
await giphy.getByText("Used for GIPHY and GIPHY Clips. Free keys allow 100 requests per hour.").waitFor();
await klipy.getByText("Used for Klipy. Free keys allow 100 requests per hour.").waitFor();
if ((await giphy.getByRole("link", { name: "Get a free key ↗" }).getAttribute("href")) !== "https://developers.giphy.com/dashboard/") fail("GIPHY key link");
if ((await klipy.getByRole("link", { name: "Get a free key ↗" }).getAttribute("href")) !== "https://partner.klipy.com") fail("Klipy key link");
await saveScreenshot(await B.screenshot(), "gifsmith-16-keys-empty.png");

// Add: the field takes focus; Save Key is disabled while empty; a rejected key keeps Not set.
await giphy.getByRole("button", { name: "Add Key…" }).click();
const field = giphy.getByRole("textbox", { name: "GIPHY API key" });
if (!(await field.evaluate((el) => document.activeElement === el))) fail("key field not focused");
if ((await field.getAttribute("placeholder")) !== "Paste your GIPHY API key") fail("placeholder");
const save = giphy.getByRole("button", { name: "Save Key" });
if (!(await save.isDisabled())) fail("Save Key enabled with an empty field");
// Esc closes the editor and restores the row.
await B.keyboard.press("Escape");
await giphy.getByRole("button", { name: "Add Key…" }).waitFor({ timeout: 3000 });
await giphy.getByRole("button", { name: "Add Key…" }).click();
await field.fill("bad-key-123");
await save.click();
await giphy.getByText("GIPHY rejected that key. Check that you copied the whole key, then try again.").waitFor({ timeout: 8000 });
await fits(giphy);
await giphy.getByText("Not set").waitFor();
await field.fill("uat-key-123");
await field.press("Enter"); // ↵ submits
await giphy.getByText("Saved. GIPHY is ready.").waitFor({ timeout: 8000 });
await giphy.getByText("Saved · ends in -123").waitFor({ timeout: 5000 });
await noKeyInPage("after saving");
await A.locator(".card input").waitFor({ state: "detached", timeout: 5000 }); // the panel got providers:changed
await saveScreenshot(await B.screenshot(), "gifsmith-16-keys-saved.png");

// Test: one uncached request; the result shows in the message line.
await giphy.getByRole("button", { name: "Test GIPHY key" }).click();
await giphy.getByText("GIPHY accepted the key.").waitFor({ timeout: 8000 });
await fits(giphy);

// Remove: inline confirmation, focus on Cancel, Esc = Cancel.
await giphy.getByRole("button", { name: "Remove GIPHY key" }).click();
await giphy.getByText("Remove your GIPHY key?").waitFor({ timeout: 3000 });
await giphy.getByText("GIPHY and Clips search stop working until you add a key again.").waitFor();
const cancel = giphy.getByRole("button", { name: "Cancel" });
if (!(await cancel.evaluate((el) => document.activeElement === el))) fail("focus not on Cancel");
await saveScreenshot(await B.screenshot(), "gifsmith-16-keys-confirm.png");
await B.keyboard.press("Escape");
await giphy.getByRole("button", { name: "Change Key…" }).waitFor({ timeout: 3000 });
await giphy.getByText("Saved · ends in -123").waitFor();
await giphy.getByRole("button", { name: "Remove GIPHY key" }).click();
await giphy.getByRole("button", { name: "Remove Key" }).click();
await giphy.getByText("Removed. Add a key to use GIPHY again.").waitFor({ timeout: 5000 });
await giphy.getByText("Not set").waitFor({ timeout: 5000 });
await A.locator(".card input").waitFor({ timeout: 5000 }); // the panel's GIPHY tab shows its setup card again
await noKeyInPage("after removing");

// Klipy is independent: saving it leaves GIPHY Not set.
await klipy.getByRole("button", { name: "Add Key…" }).click();
await klipy.getByRole("textbox", { name: "Klipy API key" }).fill("uat-key-123");
await klipy.getByRole("button", { name: "Save Key" }).click();
await klipy.getByText("Saved · ends in -123").waitFor({ timeout: 8000 });
await giphy.getByText("Not set").waitFor();
await noKeyInPage("after saving Klipy");
await saveScreenshot(await B.screenshot(), "gifsmith-16-keys-klipy.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
