// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/16-settings-search.js
// Headless only, against the -tags server build: fixture keys, fixture API, never the real Keychain.
// Page A is the panel, page B the Settings window's page. The fixture tags GIPHY titles with
// non-default request parameters ("[rating=g] "), so a title shows what Go actually sent.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.locator(".card input").fill("uat-key-123");
await A.getByRole("button", { name: "Save Key" }).click();
await A.locator("#tile-0").waitFor({ timeout: 8000 });
const firstTitle = () => A.locator("#tile-0").getAttribute("aria-label");
if ((await firstTitle()).startsWith("[rating=")) fail("default request carried a rating tag: " + (await firstTitle()));

const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("tab", { name: "Search" }).click();
await B.getByRole("heading", { name: "Search" }).waitFor({ timeout: 10000 });
const optionsLoaded = () => B.waitForFunction("document.querySelectorAll('select option').length > 40", null, { timeout: 5000 });
await optionsLoaded(); // Choices() arrives after mount
const rating = B.getByRole("combobox", { name: "Content rating" });
if ((await rating.inputValue()) !== "pg-13") fail("default rating is " + (await rating.inputValue()));
const labels = await rating.evaluate((el) => [...el.options].map((o) => o.text).join(","));
if (labels !== "G,PG,PG-13,R") fail("rating options " + labels);
await B.locator("text=Rating and language apply to GIPHY and Clips.").waitFor({ timeout: 3000 });
await rating.selectOption("g");
await B.waitForTimeout(300);

// Go flushed GIPHY's cache and the panel marked its tabs stale: switching back to GIPHY refetches.
const titleStarts = (prefix, what) =>
  A.waitForFunction(`(document.querySelector('#tile-0')?.getAttribute('aria-label') || '').startsWith(${JSON.stringify(prefix)})`, null, { timeout: 5000 }).catch(
    async () => fail(what + ": first title is " + (await firstTitle())),
  );
const backToGiphy = async () => {
  await A.keyboard.press("Meta+4");
  await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
  await A.keyboard.press("Meta+1");
};
await backToGiphy();
await titleStarts("[rating=g] ", "after choosing G");
await saveScreenshot(await B.screenshot(), "gifsmith-16-settings-search.png");

// GIFs per page: 36 → the refetched first page has 9 rows of 4 (row 152px, 16px padding each end).
const pageSize = B.getByRole("combobox", { name: "GIFs per page" });
if ((await pageSize.inputValue()) !== "24") fail("default page size is " + (await pageSize.inputValue()));
if ((await pageSize.evaluate((el) => [...el.options].map((o) => o.text).join(","))) !== "24,36,48") fail("page size options");
await B.locator("text=Bigger pages use fewer requests. Free GIPHY and Klipy keys allow 100 per hour.").waitFor({ timeout: 3000 });
await pageSize.selectOption("36");
await B.waitForTimeout(300);
await backToGiphy();
await A.waitForFunction("document.querySelector('#grid .inner')?.style.height === '1400px'", null, { timeout: 5000 }).catch(async () =>
  fail("GIPHY first page after 36: grid height " + (await A.locator("#grid .inner").evaluate((el) => el.style.height))),
);

// Language: Match System names the server build's fallback; German reaches GIPHY search.
const lang = B.getByRole("combobox", { name: "Language" });
const langText = await lang.evaluate((el) => [...el.options].map((o) => o.text));
if (langText[0] !== "Match System (English)") fail("first language option " + langText[0]);
if (!langText.includes("German")) fail("German missing from " + langText.join(","));
const sorted = langText.slice(1);
if (sorted.join("|") !== [...sorted].sort((x, y) => x.localeCompare(y, "en")).join("|")) fail("languages not sorted by name");
await rating.selectOption("pg-13");
await lang.selectOption("de");
await B.waitForTimeout(300);
await A.bringToFront();
await A.locator("input[role=combobox]").fill("cat");
await titleStarts("[lang=de] ", "GIPHY search with German");

const country = B.getByRole("combobox", { name: "Country" });
const first = await country.evaluate((el) => el.options[0].text);
if (first !== "Match System (United States)") fail("first country option " + first);
if ((await country.inputValue()) !== "") fail("country is not Match System by default");

await B.reload();
await B.getByRole("tab", { name: "Search" }).click();
await B.getByRole("heading", { name: "Search" }).waitFor({ timeout: 10000 });
await optionsLoaded();
for (const [name, want] of [["Content rating", "pg-13"], ["GIFs per page", "36"], ["Language", "de"]]) {
  const got = await B.getByRole("combobox", { name }).inputValue();
  if (got !== want) fail(`${name} after reload is ${got}, want ${want}`);
}
await saveScreenshot(await B.screenshot(), "gifsmith-16-settings-search-2.png");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
