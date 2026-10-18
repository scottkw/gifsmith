// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   UAT_SEED_FAVORITES='{not json' bash scripts/smoke-browser-loop.sh --uat scripts/uat/12-library-unreadable.js
// favorites.json is corrupt (Pitfall 11): the Favorites tab shows the load error with Try Again,
// ⌘⇧F shows F5 and changes nothing, and Recents keeps working. The harness then checks the file
// bytes are unchanged.
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
const content = page.locator("section.content");

await page.keyboard.press("Meta+5");
await content.filter({ hasText: "Couldn't load Favorites" }).waitFor({ timeout: 5000 });
if (!(await content.innerText()).includes("Gifsmith couldn't read its saved list. Try again, or quit and reopen Gifsmith."))
  fail("load error body");
await page.getByRole("button", { name: "Try Again" }).click();
await content.filter({ hasText: "Couldn't load Favorites" }).waitFor({ timeout: 5000 }); // still corrupt
await saveScreenshot(await page.screenshot(), "gifsmith-12-unreadable.png");

// ⌘⇧F on a provider tile: F5, no badge.
await page.keyboard.press("Meta+4");
await input.fill("dwight");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
await input.press("Meta+Shift+F");
await page.locator("footer .center.error", { hasText: "Couldn't save your Favorites. Try again." }).waitFor({ timeout: 5000 });
if (await page.locator(".tile .fav").count()) fail("star badge after a refused favorite");

// Recents is a separate file and still works.
await input.press("Enter");
await page.locator("footer .center.success", { hasText: "Copied" }).waitFor({ timeout: 10000 });
await input.fill("");
await page.keyboard.press("Meta+6");
await page.locator(".header", { hasText: "Finer Gifs · 1" }).waitFor({ timeout: 5000 });

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
