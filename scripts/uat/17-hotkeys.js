// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/17-hotkeys.js
// Headless only, against the -tags server build: its Registrar is an in-memory fake, so nothing is
// registered with macOS and no key press leaves the page. Key presses go to the headless page only.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push(String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
const openPane = async () => {
  await B.getByRole("tab", { name: "Hotkeys" }).click();
  await B.getByRole("heading", { name: "Hotkeys" }).waitFor({ timeout: 10000 });
};
await openPane();

const open = B.getByRole("button", { name: "Open Gifsmith shortcut", exact: true });
const reset = B.getByRole("button", { name: "Reset to Default" });
const text = async (loc) => (await loc.innerText()).trim();
const waitText = async (loc, want, ms = 5000) => {
  const end = Date.now() + ms;
  while (Date.now() < end) {
    if ((await text(loc)) === want) return;
    await B.waitForTimeout(50);
  }
  fail(`expected "${want}", got "${await text(loc)}"`);
};

await waitText(open, "⌃⌥⌘G");
if (!(await reset.isDisabled())) fail("Reset to Default enabled on the default");
await B.getByText("macOS doesn't warn when another app uses the same shortcut. Use Test to check yours.").waitFor();
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-default.png");

// Record ⌃⌥⌘H: the field shows the new combo only after Go accepted it.
await open.click();
await waitText(open, "Type shortcut…");
// No ⌘/⌥/⌃: H1, and the field keeps recording.
await B.keyboard.press("KeyH");
await B.getByText("Include ⌘, ⌥ or ⌃ in the shortcut.").waitFor({ timeout: 3000 });
await waitText(open, "Type shortcut…");
await B.keyboard.down("Control");
await B.keyboard.down("Alt");
await B.keyboard.down("Meta");
await waitText(open, "⌃⌥⌘"); // held modifiers, live
await B.keyboard.press("KeyH");
await B.keyboard.up("Meta");
await B.keyboard.up("Alt");
await B.keyboard.up("Control");
await waitText(open, "⌃⌥⌘H");
if (await reset.isDisabled()) fail("Reset to Default disabled after a rebind");
if (await B.getByText("Include ⌘, ⌥ or ⌃ in the shortcut.").count()) fail("H1 still shown after a successful rebind");
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-rebound.png");

// Persisted: a reload reads it back from Go.
await B.reload();
await openPane();
await waitText(open, "⌃⌥⌘H");

await reset.click();
await waitText(open, "⌃⌥⌘G");
if (!(await reset.isDisabled())) fail("Reset to Default enabled after reset");

// Backspace in a recording Open field is ignored: it keeps recording (Open can only be reset).
await open.click();
await B.keyboard.press("Backspace");
await waitText(open, "Type shortcut…");
await B.keyboard.press("Escape");
await waitText(open, "⌃⌥⌘G");

// Favorites and Recents start unset, with Test and Clear disabled.
const fav = B.getByRole("button", { name: "Open Favorites shortcut", exact: true });
const rec = B.getByRole("button", { name: "Open Recents shortcut", exact: true });
const favTest = B.getByRole("button", { name: "Test Open Favorites shortcut" });
const favClear = B.getByRole("button", { name: "Clear Open Favorites shortcut" });
await waitText(fav, "Not set");
await waitText(rec, "Not set");
for (const b of [favTest, favClear, B.getByRole("button", { name: "Test Open Recents shortcut" }), B.getByRole("button", { name: "Clear Open Recents shortcut" })]) {
  if (!(await b.isDisabled())) fail("Test/Clear enabled on an unset row");
}

await fav.click();
await B.keyboard.press("Control+Alt+Meta+KeyF");
await waitText(fav, "⌃⌥⌘F");
if ((await favTest.isDisabled()) || (await favClear.isDisabled())) fail("Favorites Test/Clear disabled once set");

// Duplicate inside Gifsmith (H2) and a macOS shortcut (H3): Recents stays unset.
await rec.click();
await B.keyboard.press("Control+Alt+Meta+KeyF");
await B.getByText("⌃⌥⌘F is already used for Open Favorites.").waitFor({ timeout: 5000 });
await waitText(rec, "Not set");
await rec.click();
await B.keyboard.press("Meta+KeyQ");
await B.getByText("⌘Q is a macOS shortcut. Choose another one.").waitFor({ timeout: 5000 });
await waitText(rec, "Not set");
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-errors.png");

// Test: the server build's fake never fires, so the row reports H7 after 5 s. Esc cancels listening.
await favTest.click();
await B.getByText("Press ⌃⌥⌘F now…").waitFor({ timeout: 3000 });
await B.keyboard.press("Escape");
await B.getByText("Press ⌃⌥⌘F now…").waitFor({ state: "detached", timeout: 3000 });
await favTest.click();
await B.getByText("Press ⌃⌥⌘F now…").waitFor({ timeout: 3000 });
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-test.png");
await B.getByText("Nothing happened. Another app may be using ⌃⌥⌘F. Try a different shortcut.").waitFor({ timeout: 7000 });

await favClear.click();
await waitText(fav, "Not set");
if (!(await favClear.isDisabled())) fail("Clear enabled after clearing");

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
