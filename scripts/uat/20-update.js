// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/20-update.js
// Headless only, against the -tags server build: the update check asks the in-process fixture
// (default mode serves v9.9.0), never api.github.com. Page B is the Settings window's page.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

const row = B.locator(".row").filter({ has: B.locator("#updates") });
const checkBtn = B.getByRole("button", { name: "Check for Updates…", exact: true });
await checkBtn.waitFor({ timeout: 5000 });
const available = "Gifsmith 9.9.0 is available. You have 1.0.0.";

await checkBtn.click();
await row.getByText(available, { exact: true }).waitFor({ timeout: 5000 });
// A second check sends the ETag; the fixture answers 304 and Go reuses the cached release.
await checkBtn.click();
await B.waitForTimeout(300);
await row.getByText(available, { exact: true }).waitFor({ timeout: 5000 });
if (await row.locator("[role=alert]").count()) fail("error after the second check: " + (await row.locator("[role=alert]").innerText()));
await saveScreenshot(await B.screenshot(), "gifsmith-20-update-available.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
