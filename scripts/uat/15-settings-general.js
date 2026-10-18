// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/15-settings-general.js
// Headless only, against the -tags server build (platform.Fake, in-memory autostart): never the
// real NSApp appearance or login items. Page B is the Settings window's page.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");

const themeIs = (page, want) =>
  page.waitForFunction(`(document.documentElement.dataset.theme || '') === ${JSON.stringify(want)}`, null, { timeout: 3000 });
const radio = (page, name) => page.getByRole("radio", { name });

await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
if (!(await radio(B, "System").isChecked())) fail("System is not the default theme");

await radio(B, "Dark").click();
await themeIs(B, "dark");
await B.waitForTimeout(200);
await B.reload();
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
if (!(await radio(B, "Dark").isChecked())) fail("Dark not kept after reload");
await themeIs(B, "dark");

await radio(B, "Light").click();
await themeIs(B, "light");
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-light.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
