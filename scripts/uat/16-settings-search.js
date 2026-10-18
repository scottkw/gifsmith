// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/16-settings-search.js
// Headless only, against the -tags server build: fixture keys, fixture API, never the real Keychain.
// Page A is the panel, page B the Settings window's page. The fixture tags GIPHY titles with
// non-default request parameters ("[rating=g] "), so a title shows what Go actually sent.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.locator(".card input").fill("uat-key-123");
await A.getByRole("button", { name: "Save Key" }).click();
await A.locator("#tile-0").waitFor({ timeout: 8000 });
const firstTitle = () => A.locator("#tile-0").getAttribute("aria-label");
if ((await firstTitle()).startsWith("[rating=")) fail("default request carried a rating tag: " + (await firstTitle()));

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Search" }).click();
await B.getByRole("heading", { name: "Search" }).waitFor({ timeout: 10000 });
const rating = B.getByRole("combobox", { name: "Content rating" });
if ((await rating.inputValue()) !== "pg-13") fail("default rating is " + (await rating.inputValue()));
const labels = await rating.evaluate((el) => [...el.options].map((o) => o.text).join(","));
if (labels !== "G,PG,PG-13,R") fail("rating options " + labels);
await B.locator("text=Rating and language apply to GIPHY and Clips.").waitFor({ timeout: 3000 });
await rating.selectOption("g");
await B.waitForTimeout(300);

// Go flushed GIPHY's cache, so the reloaded panel's trending request carries rating=g.
await A.reload();
await A.locator("#tile-0").waitFor({ timeout: 8000 });
await A.waitForFunction(`(document.querySelector('#tile-0')?.getAttribute('aria-label') || '').startsWith('[rating=g] ')`, null, { timeout: 5000 }).catch(async () =>
  fail("after choosing G the first title is " + (await firstTitle())),
);
await saveScreenshot(await B.screenshot(), "gifsmith-16-settings-search.png");

await B.reload();
await B.getByRole("tab", { name: "Search" }).click();
if ((await B.getByRole("combobox", { name: "Content rating" }).inputValue()) !== "g") fail("G not kept after reload");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
