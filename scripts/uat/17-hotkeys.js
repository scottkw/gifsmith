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

const open = B.getByRole("button", { name: "Open Gifsmith shortcut" });
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
await saveScreenshot(await B.screenshot(), "gifsmith-17-hotkeys-rebound.png");

// Persisted: a reload reads it back from Go.
await B.reload();
await openPane();
await waitText(open, "⌃⌥⌘H");

await reset.click();
await waitText(open, "⌃⌥⌘G");
if (!(await reset.isDisabled())) fail("Reset to Default enabled after reset");

if (errors.length) fail("page errors: " + errors.join("; "));
console.log("UAT OK");
