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

// Download and Open…: record every status text and whether a progress bar appeared.
await B.evaluate(() => {
  const row = document.querySelector("#updates").closest(".row");
  window.__upd = { texts: [], progress: false };
  const note = () => {
    const s = row.querySelector("[role=status]");
    if (s && window.__upd.texts.at(-1) !== s.textContent.trim()) window.__upd.texts.push(s.textContent.trim());
    if (row.querySelector("progress")) window.__upd.progress = true;
  };
  new MutationObserver(note).observe(row, { subtree: true, childList: true, characterData: true, attributes: true });
});
await B.getByRole("button", { name: "Download and Open…", exact: true }).click();
const opened = "The Gifsmith 9.9.0 disk image is open. Quit Gifsmith, then drag the new Gifsmith to Applications and choose Replace.";
await row.getByText(opened, { exact: true }).waitFor({ timeout: 10000 });
const seen = await B.evaluate(() => window.__upd);
if (!seen.texts.some((t) => t.startsWith("Downloading Gifsmith 9.9.0…"))) fail("no Downloading status: " + JSON.stringify(seen.texts));
if (!seen.texts.some((t) => /Downloading Gifsmith 9\.9\.0… \d+%/.test(t))) fail("no percent: " + JSON.stringify(seen.texts));
if (!seen.texts.includes("Checking the download's signature…")) fail("no verifying status: " + JSON.stringify(seen.texts));
if (!seen.progress) fail("no progress bar seen");
if (await B.getByRole("button", { name: "Download and Open…", exact: true }).count()) fail("Download and Open… still shown");
if (await row.locator("[role=alert]").count()) fail("error after download: " + (await row.locator("[role=alert]").innerText()));
await saveScreenshot(await B.screenshot(), "gifsmith-20-update-opened.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
