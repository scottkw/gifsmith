// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_UPDATE=current bash scripts/smoke-browser-loop.sh --uat scripts/uat/20-update-current.js
// Headless only, against the -tags server build: the fixture has no published release (404),
// so Gifsmith is up to date and the panel never announces anything.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];

const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

const row = B.locator(".row").filter({ has: B.locator("#updates") });
await B.getByRole("button", { name: "Check for Updates…", exact: true }).click();
await row.getByText("Gifsmith 1.0.1 is up to date.", { exact: true }).waitFor({ timeout: 5000 });
if (await B.getByRole("button", { name: "Download and Open…", exact: true }).count()) fail("Download and Open… shown when up to date");
if (await row.locator("[role=alert]").count()) fail("error: " + (await row.locator("[role=alert]").innerText()));
await A.waitForTimeout(2000);
if ((await A.locator("footer").innerText()).includes("is available")) fail("the panel announced an update");
await saveScreenshot(await B.screenshot(), "gifsmith-20-update-current.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
