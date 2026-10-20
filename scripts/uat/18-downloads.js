// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_PICK_FOLDER=bin/uat-picked bash scripts/smoke-browser-loop.sh --uat scripts/uat/18-downloads.js
// Headless only, against the -tags server build. Choose… "picks" FAKE_PICK_FOLDER (never the real
// picker); downloads land in bin/uat-downloads or bin/uat-picked, never ~/Downloads.
// Page A is the panel, page B the Settings window's page.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.keyboard.press("Meta+4");
await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
const input = A.locator("input[role=combobox]");
await input.fill("dwight");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
const downloadsTo = async (folder) => {
  await input.press("Meta+s");
  await A.locator("footer .center.success", { hasText: "Downloaded to " + folder }).waitFor({ timeout: 10000 }).catch(async () =>
    fail(`⌘S: footer reads "${await A.locator("footer .center").innerText()}", want Downloaded to ${folder}`),
  );
  await input.press("ArrowRight"); // any keystroke clears the toast
  await input.press("ArrowLeft");
};

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Downloads", exact: true }).click();
await B.getByRole("heading", { name: "Downloads" }).waitFor({ timeout: 10000 });
const display = (name) => B.getByRole("group", { name: "Download folder: " + name, exact: true });
const reset = B.getByRole("button", { name: "Reset download folder to Downloads" });
await display("uat-downloads").waitFor({ timeout: 3000 });
const path = B.locator(".display .path");
const title = await path.getAttribute("title");
if (!title.endsWith("/bin/uat-downloads") || !title.startsWith("/")) fail("path title " + title);
const pathStyle = await path.evaluate((el) => [getComputedStyle(el).whiteSpace, getComputedStyle(el).textOverflow].join(","));
if (pathStyle !== "nowrap,ellipsis") fail("path is not one line with an ellipsis: " + pathStyle);
if (!(await reset.isDisabled())) fail("Reset to Downloads is enabled on the default folder");

await B.getByRole("button", { name: "Choose download folder" }).click();
await display("uat-picked").waitFor({ timeout: 3000 });
if (await reset.isDisabled()) fail("Reset to Downloads is disabled after choosing a folder");
await saveScreenshot(await B.screenshot(), "gifsmith-18-downloads.png");

await A.bringToFront();
await downloadsTo("uat-picked");

await reset.click();
await display("uat-downloads").waitFor({ timeout: 3000 });
if (!(await reset.isDisabled())) fail("Reset to Downloads is enabled after resetting");
await A.bringToFront();
await downloadsTo("uat-downloads");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
