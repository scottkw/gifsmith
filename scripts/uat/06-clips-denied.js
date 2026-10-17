// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips-denied.js
// Headless only, fixture keys only. The fixture answers 403 on /v1/clips, so a key that works for
// GIFs gets C2. Never clicks "Email clips@giphy.com ↗": Browser.OpenURL would open the user's mail
// app, so only its href is asserted.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });

const content = page.locator("section.content");
const footer = page.locator("footer");
const placeholder = (want) =>
  page.waitForFunction(`document.querySelector('input[role=combobox]').placeholder === ${JSON.stringify(want)}`, null, {
    timeout: 5000,
  });

await page.locator(".card input").fill("uat-key-123");
await page.getByRole("button", { name: "Save Key" }).click();
await footer.filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });

await page.keyboard.press("Meta+2");
await placeholder("Search GIPHY Clips");
await content.locator("h2", { hasText: "Clips aren't enabled for your key" }).waitFor({ timeout: 8000 });
const body = await content.locator("p").innerText();
if (
  body !==
  "Your GIPHY key works for GIFs, but GIPHY has to approve it for Clips. Email clips@giphy.com to request access, then reopen Gifsmith."
)
  fail("C2 body: " + body);
const email = page.getByRole("link", { name: "Email clips@giphy.com ↗" });
if ((await email.getAttribute("href")) !== "mailto:clips@giphy.com?subject=GIPHY%20Clips%20API%20access%20request")
  fail("mailto href " + (await email.getAttribute("href")));
const back = page.getByRole("button", { name: "Back to GIPHY" });
if (!(await back.isVisible())) fail("Back to GIPHY not visible");
if ((await page.locator("[role=tab]").nth(1).getAttribute("aria-disabled")) !== "true") fail("Clips pill enabled for a denied key");
if (!(await footer.locator('img[alt="Powered by GIPHY"]').count())) fail("C2 without the GIPHY mark");
await saveScreenshot(await page.screenshot(), "gifsmith-06-clips-c2.png");

await back.click();
await placeholder("Search GIPHY");
await page.locator("#tile-3 img").waitFor({ timeout: 8000 });
if (!(await page.locator("#tile-0 img").getAttribute("src")).includes("/media/giphy-")) fail("Back to GIPHY did not show the GIPHY grid");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
