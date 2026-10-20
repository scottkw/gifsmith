// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_HOTKEY_FAIL=Cmd+Ctrl+Option+J bash scripts/smoke-browser-loop.sh --uat scripts/uat/17-hotkeys-fail.js
// Headless only. The server build's fake Registrar refuses ⌃⌥⌘J, as macOS would for a failed
// RegisterEventHotKey: Go re-registers the old binding and the field keeps it (H4).
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push(String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Hotkeys" }).click();
await B.getByRole("heading", { name: "Hotkeys" }).waitFor({ timeout: 10000 });

const open = B.getByRole("button", { name: "Open Gifsmith shortcut", exact: true });
const text = async () => (await open.innerText()).trim();
if ((await text()) !== "⌃⌥⌘G") fail("Open shows " + (await text()));

await open.click();
await B.keyboard.press("Control+Alt+Meta+KeyJ");
await B.getByText("macOS couldn't register ⌃⌥⌘J. Your previous shortcut still works.").waitFor({ timeout: 5000 });
if ((await text()) !== "⌃⌥⌘G") fail("after the failed rebind Open shows " + (await text()));
const role = await B.locator("#hk-open-msg").getAttribute("role");
if (role !== "alert") fail("H4 role " + role);
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-fail.png");

// Go kept the old binding: a reload still shows it.
await B.reload();
await B.getByRole("tab", { name: "Hotkeys" }).click();
await open.waitFor({ timeout: 10000 });
if ((await text()) !== "⌃⌥⌘G") fail("after reload Open shows " + (await text()));

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
