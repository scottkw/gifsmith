// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/15-settings-general.js
// Headless only, against the -tags server build (platform.Fake, in-memory autostart): never the
// real NSApp appearance or login items. Page A is the panel, page B the Settings window's page.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
// Save a fixture key (as 05-setup.js) so the GIPHY tab shows results and its footer mark.
await A.locator(".card input").fill("uat-key-123");
await A.getByRole("button", { name: "Save Key" }).click();
await A.locator("footer").filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });
const mark = A.locator('footer img[alt="Powered by GIPHY"]');
const markIs = (want) =>
  A.waitForFunction(
    `(document.querySelector('footer img[alt="Powered by GIPHY"]')?.getAttribute('src') || '').endsWith(${JSON.stringify(want)})`,
    null,
    { timeout: 3000 },
  );

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");

const themeIs = (page, want) =>
  page.waitForFunction(`(document.documentElement.dataset.theme || '') === ${JSON.stringify(want)}`, null, { timeout: 3000 });
const radio = (page, name) => page.getByRole("radio", { name });

await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
if (!(await radio(B, "System").isChecked())) fail("System is not the default theme");
// System resolves to the (light) headless appearance on both pages.
await themeIs(A, "light");
await themeIs(B, "light");
await markIs("giphy-powered-light.png");

await radio(B, "Dark").click();
await themeIs(B, "dark");
await themeIs(A, "dark"); // settings:changed reached the panel page
const bg = await A.evaluate(() => getComputedStyle(document.body).backgroundColor);
if (bg !== "rgb(30, 30, 30)") fail("panel background in Dark: " + bg);
await markIs("giphy-powered-dark.png");
await saveScreenshot(await A.screenshot(), "gifsmith-15-panel-dark.png");
await B.waitForTimeout(200);
await B.reload();
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
if (!(await radio(B, "Dark").isChecked())) fail("Dark not kept after reload");
await themeIs(B, "dark");

// Startup: off by default (OS state from the in-memory login item); checking it persists.
const login = B.getByRole("checkbox", { name: "Open Gifsmith at login" });
if (await login.isChecked()) fail("Open at login checked by default");
await login.check();
await B.waitForTimeout(300);
await B.reload();
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });
if (!(await login.isChecked())) fail("Open at login not kept after reload");
// Automatic paste: the server build's platform.Fake is not trusted.
await B.locator("text=Off").first().waitFor({ timeout: 5000 });
await B.getByRole("button", { name: "Grant Accessibility access" }).waitFor({ timeout: 5000 });
if ((await B.getByRole("button", { name: "Grant Accessibility access" }).innerText()).trim() !== "Grant Access…") fail("Grant button text");
await B.locator("text=Paste actions need Accessibility access. Copying always works without it.").waitFor({ timeout: 3000 });
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-general.png");

await radio(B, "Light").click();
await themeIs(B, "light");
await themeIs(A, "light");
await markIs("giphy-powered-light.png");
if (!(await mark.count())) fail("GIPHY mark missing");
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-light.png");

// Sidebar keyboard: the selected tab is the only tab stop; ↑/↓ move and clamp at the ends; Tab goes on into the pane.
const tab = B.getByRole("tab", { name: "General" });
if ((await tab.getAttribute("aria-selected")) !== "true") fail("General tab not selected");
await tab.focus();
const last = await B.getByRole("tab").count();
for (const [k, want] of [["ArrowUp", 0], ...Array.from({ length: last }, (_, i) => ["ArrowDown", Math.min(i + 1, last - 1)]), ["ArrowUp", last - 2]]) {
  await B.keyboard.press(k);
  const at = await B.evaluate(() => [...document.querySelectorAll("[role=tab]")].indexOf(document.activeElement));
  if (at !== Math.max(want, 0)) fail(`${k}: focus on tab ${at}, want ${want}`);
}
for (let i = 0; i < last; i++) await B.keyboard.press("ArrowUp");
if (!(await tab.evaluate((el) => document.activeElement === el && el.getAttribute("aria-selected") === "true"))) fail("ArrowUp did not return to General");
await B.keyboard.press("Tab");
const tabbed = await B.evaluate(() => document.activeElement && document.activeElement.getAttribute("type") + ":" + document.activeElement.name);
if (tabbed !== "radio:theme") fail("Tab from the sidebar reached " + tabbed);
const overflow = await B.evaluate(() => [
  getComputedStyle(document.querySelector("[role=tabpanel]")).overflowY,
  getComputedStyle(document.querySelector("[role=tablist]")).overflowY,
]);
if (overflow[0] !== "auto" || overflow[1] === "auto" || overflow[1] === "scroll") fail("pane/sidebar overflow " + overflow.join("/"));
await saveScreenshot(await B.screenshot(), "gifsmith-15-settings-sidebar.png");

// ⌘, in the panel asks Go to open Settings (server build: the shell hides/shows nothing real).
await A.bringToFront();
await A.locator("input[role=combobox]").focus();
await A.keyboard.press("Meta+Comma");
await A.waitForTimeout(500);
const center = await A.locator("footer .center").evaluate((el) => ({ cls: el.className, text: el.innerText }));
if (center.cls.includes("error")) fail("Cmd+, showed an error toast: " + center.text);
await B.bringToFront();

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
