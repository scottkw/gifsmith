// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   UAT_SEED_SETTINGS='{not json' bash scripts/smoke-browser-loop.sh --uat scripts/uat/15-settings-unreadable.js
// settings.json is corrupt (WR-03): General names the problem, a change says why it did not save,
// and Reset Settings… (after a confirm) moves the file aside so changes save again. The harness
// then checks the corrupt bytes were kept as settings.json.<time>.bak.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push(String(e)));
await B.setViewportSize({ width: 640, height: 560 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

const banner = B.getByRole("alert").filter({ hasText: "Gifsmith couldn't read its settings file" });
await banner.waitFor({ timeout: 5000 });

// A change is refused with the cause, not "Try again".
await B.getByRole("radio", { name: "Dark" }).click();
await B.getByRole("alert").filter({ hasText: "Couldn't save. Gifsmith can't read its settings file. See General." }).waitFor({ timeout: 5000 });
if (!(await B.getByRole("radio", { name: "System" }).isChecked())) fail("theme did not revert after the refused write");

// Reset asks first; Cancel keeps everything.
await B.getByRole("button", { name: "Reset Settings…" }).click();
await B.getByRole("alert").filter({ hasText: "Reset all settings to their defaults?" }).waitFor({ timeout: 5000 });
const focused = await B.evaluate(() => document.activeElement?.textContent?.trim());
if (focused !== "Cancel") fail("confirm did not focus Cancel: " + focused);
await B.getByRole("button", { name: "Cancel" }).click();
await banner.waitFor({ timeout: 5000 });
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-unreadable.png");

await B.getByRole("button", { name: "Reset Settings…" }).click();
await B.getByRole("button", { name: "Reset Settings", exact: true }).click();
await banner.waitFor({ state: "detached", timeout: 5000 });

// Changes save again.
await B.getByRole("radio", { name: "Dark" }).click();
await B.waitForFunction(`document.documentElement.dataset.theme === 'dark'`, null, { timeout: 3000 });
if (await B.getByRole("alert").filter({ hasText: "Couldn't save" }).count()) fail("save error after the reset");
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-unreadable-reset.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
