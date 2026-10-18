// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_AUTOSTART=fail bash scripts/smoke-browser-loop.sh --uat scripts/uat/15-settings-startup-fail.js
// Headless only. The server build's in-memory login item fails every Enable/Disable (G2).
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push(String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

const login = B.getByRole("checkbox", { name: "Open Gifsmith at login" });
await login.click();
await B.getByRole("alert").filter({ hasText: "Couldn't change Open at login. Try again." }).waitFor({ timeout: 5000 });
if (await login.isChecked()) fail("checkbox did not revert after a failed Enable");
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-startup-fail.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
