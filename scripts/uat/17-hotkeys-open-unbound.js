// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_HOTKEY_FAIL=Cmd+Ctrl+Option+G bash scripts/smoke-browser-loop.sh --uat scripts/uat/17-hotkeys-open-unbound.js
// Headless only. The fake Registrar refuses the default ⌃⌥⌘G, so Open is unbound from launch (H4b).
// Reset to Default must stay available as the way back (WR-06); here it is refused again.
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
if ((await open.innerText()).trim() !== "Not set") fail("Open shows " + (await open.innerText()));
await B.getByText("macOS couldn't register ⌃⌥⌘G. Try a different shortcut.").waitFor({ timeout: 5000 });

const reset = B.getByRole("button", { name: "Reset to Default, Open Gifsmith shortcut" });
if (await reset.isDisabled()) fail("Reset to Default is disabled while Open is unbound");
await reset.click();
await B.getByText("macOS couldn't register ⌃⌥⌘G. Try a different shortcut.").waitFor({ timeout: 5000 });
if ((await open.innerText()).trim() !== "Not set") fail("after the refused reset Open shows " + (await open.innerText()));
if (await reset.isDisabled()) fail("Reset to Default disabled after the refused retry");
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-open-unbound.png");

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
