// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_UPDATE=tamper bash scripts/smoke-browser-loop.sh --uat scripts/uat/20-update-fail.js
// Headless only, against the -tags server build: the fixture offers 9.9.0 with a digest that does
// not match the download, so Go deletes the file and opens nothing (U-E3).
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
await B.getByRole("button", { name: "Check for Updates…", exact: true }).click();
await row.getByText("Gifsmith 9.9.0 is available. You have 1.0.0.", { exact: true }).waitFor({ timeout: 5000 });
await B.getByRole("button", { name: "Download and Open…", exact: true }).click();
const alert = row.locator("[role=alert]");
await alert.filter({ hasText: "The download didn't pass the signature check, so Gifsmith deleted it. Nothing was opened." }).waitFor({ timeout: 10000 });
await B.getByRole("button", { name: "Download and Open…", exact: true }).waitFor({ timeout: 3000 });
if (await row.locator("progress").count()) fail("a progress bar remains after the failure");
await row.getByText("Gifsmith 9.9.0 is available. You have 1.0.0.", { exact: true }).waitFor({ timeout: 3000 });
await saveScreenshot(await B.screenshot(), "gifsmith-20-update-fail.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
