// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/18-default-actions.js
// Headless only, against the -tags server build (platform.Fake: nothing reaches the pasteboard).
// Page A is the panel, page B the Settings window's page. ↵/⌘↵ run the actions chosen in General;
// choosing the other key's action swaps them; Klipy file actions are gated (RELS-04). Fixture key only.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.keyboard.press("Meta+4");
await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
const input = A.locator("header input[role=combobox]");
const hints = A.locator("footer .hints");
const toast = (text, tone = "success") => A.locator(`footer .center${tone ? "." + tone : ""}`, { hasText: text }).waitFor({ timeout: 5000 });
await input.fill("dwight");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
const hintText = async () => (await hints.innerText()).replace(/\s+/g, " ").trim();
if ((await hintText()) !== "↵ Copy ⌘↵ Paste ⌘K Actions") fail("default hints " + (await hintText()));

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
const primary = B.getByRole("combobox", { name: "Primary action (↵)" });
const secondary = B.getByRole("combobox", { name: "Secondary action (⌘↵)" });
if ((await primary.inputValue()) !== "copy" || (await secondary.inputValue()) !== "paste") fail("defaults " + (await primary.inputValue()) + "/" + (await secondary.inputValue()));
const opts = await primary.evaluate((el) => [...el.options].map((o) => o.text).join(","));
if (opts !== "Copy GIF,Paste GIF,Copy Link,Paste Link,Copy Markdown,Paste Markdown,Download GIF") fail("options " + opts);
await B.locator("text=Klipy GIFs always copy and paste as links.").waitFor({ timeout: 3000 });
await saveScreenshot(await B.screenshot(), "gifsmith-18-default-actions.png");

// Primary → Copy Link: the panel's hint follows and ↵ copies the link.
await primary.selectOption("copyLink");
await A.waitForFunction("document.querySelector('footer .hints')?.innerText.replace(/\\s+/g, ' ').trim() === '↵ Copy Link ⌘↵ Paste ⌘K Actions'", null, { timeout: 3000 }).catch(async () =>
  fail("hints after Primary Copy Link: " + (await hintText())),
);
await A.bringToFront();
await input.press("Enter");
await toast("Copied link");
await input.press("ArrowRight"); // clears the toast
await input.press("ArrowLeft");

// Secondary → Copy Link (the primary's action): the two swap.
await secondary.selectOption("copyLink");
await B.waitForFunction("document.querySelector('select[aria-labelledby=primary-label]').value === 'paste'", null, { timeout: 3000 }).catch(async () =>
  fail("primary after the swap is " + (await primary.inputValue())),
);
if ((await secondary.inputValue()) !== "copyLink") fail("secondary after the swap is " + (await secondary.inputValue()));

// Klipy while gated: no file rows; ↵ (Paste GIF chosen) runs Copy Link; ⌘S shows T11.
await A.bringToFront();
await A.keyboard.press("Meta+3");
const field = A.locator(".card input");
await field.waitFor({ timeout: 5000 });
await field.fill("uat-key-123");
await field.press("Enter");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
if ((await hintText()) !== "↵ Copy Link ⌘↵ Copy Link ⌘K Actions") fail("Klipy hints " + (await hintText()));
const dialog = A.locator("[role=dialog][aria-label=Actions]");
await input.press("Meta+k");
await dialog.waitFor({ timeout: 2000 });
const rows = dialog.locator("[role=option]");
const labels = (await rows.allInnerTexts()).map((t) => t.split("\n")[0].trim());
for (const gone of ["Copy GIF", "Paste GIF", "Download GIF", "Copy Square GIF", "Paste Square GIF"]) {
  if (labels.includes(gone)) fail(`${gone} listed on a gated Klipy item: ${JSON.stringify(labels)}`);
}
const copyLink = rows.filter({ hasText: "Copy Link" });
if ((await copyLink.locator(".hint").innerText()).trim() !== "↵") fail("no ↵ hint on Copy Link: " + (await copyLink.innerText()));
await dialog.locator("input[role=combobox]").press("Escape");
await dialog.waitFor({ state: "detached", timeout: 2000 });
await input.press("Enter");
await toast("Copied link");
await input.press("ArrowRight");
await input.press("ArrowLeft");
await input.press("Meta+s");
await toast("Klipy GIFs can't be downloaded yet.", "");
if (await A.locator("footer .center.error").count()) fail("T11 shown as an error");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
