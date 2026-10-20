// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/19-about.js
// Headless, against the -tags server build. Settings → About: copy, version, links (never
// clicked: Browser.OpenURL would open the user's real browser), the theme-resolved GIPHY mark
// and scrolling at the 640×520 window size.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push(String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

const openAbout = async () => {
  await B.getByRole("tab", { name: "About" }).click();
  await B.getByRole("heading", { name: "About" }).waitFor({ timeout: 5000 });
};
await openAbout();

const COPY = [
  "Version 1.0.0",
  "Search GIFs from the menu bar and copy or paste them into any app.",
  "Gifsmith is a port of the GIF Search extension for Raycast by Joseph Schmitt and contributors: tm.wrnr, j3lte, samuelkraft, ridemountainpig, spookywy, seanparkross, YourMCGeek, thomas, tim_gailey, clins1994, 0xdhrv, Olli0103 and alexi.build. Used under the MIT License, © 2021 Raycast.",
  "Search by GIPHY and KLIPY. GIFs also from The Finer Gifs Club.",
  "© 2026 Ken Scott. MIT License.",
];
const text = (await B.locator("[role=tabpanel]").innerText()).replace(/\s+/g, " ");
for (const c of COPY) if (!text.includes(c)) fail("About lacks: " + c);

const LINKS = {
  "Website ↗": "https://github.com/scottkw/gifsmith",
  "Report an Issue ↗": "https://github.com/scottkw/gifsmith/issues",
  "View License ↗": "https://github.com/scottkw/gifsmith/blob/main/LICENSE",
  "GIF Search extension ↗": "https://github.com/raycast/extensions/tree/main/extensions/gif-search",
};
for (const [name, href] of Object.entries(LINKS)) {
  const link = B.getByRole("link", { name, exact: true });
  if ((await link.count()) !== 1) fail("link missing: " + name);
  if ((await link.getAttribute("href")) !== href) fail(`${name} href ${await link.getAttribute("href")}`);
}

const markSrc = () => B.locator('[role=tabpanel] img[alt="Powered by GIPHY"]').getAttribute("src");
if (!(await markSrc()).endsWith("-light.png")) fail("Light mark src " + (await markSrc()));
if ((await B.locator('[role=tabpanel] img[alt=""]').getAttribute("src")) !== "/appicon.png") fail("app icon");
await saveScreenshot(await B.screenshot(), "gifsmith-19-about-light.png");

// The pane scrolls rather than clipping; the last line is reachable.
const scroll = await B.evaluate(() => {
  const pane = document.querySelector("[role=tabpanel]");
  pane.scrollTop = pane.scrollHeight;
  const lastLine = [...pane.querySelectorAll("p")].pop();
  return {
    overflowY: getComputedStyle(pane).overflowY,
    fits: pane.scrollHeight >= pane.clientHeight,
    reachable: lastLine.getBoundingClientRect().bottom <= pane.getBoundingClientRect().bottom + 1,
    last: lastLine.textContent.trim(),
  };
});
if (scroll.overflowY !== "auto" || !scroll.fits || !scroll.reachable) fail("pane scroll " + JSON.stringify(scroll));
if (scroll.last !== "© 2026 Ken Scott. MIT License.") fail("last line " + scroll.last);

await B.getByRole("tab", { name: "General" }).click();
await B.getByRole("radio", { name: "Dark" }).click();
await B.waitForFunction(`document.documentElement.dataset.theme === 'dark'`, null, { timeout: 3000 });
await openAbout();
if (!(await markSrc()).endsWith("-dark.png")) fail("Dark mark src " + (await markSrc()));
await saveScreenshot(await B.screenshot(), "gifsmith-19-about-dark.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
