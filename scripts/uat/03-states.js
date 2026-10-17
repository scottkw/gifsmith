// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/03-states.js
// Drives every content-area state through the fixture's scenario words (internal/provider/fakeapi).
// A 429 blocks the provider in Go (search.Service), so the rate-limit scenarios run last, in an
// order where each one's assertions hold inside the previous one's block window.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");

const input = page.locator("input[role=combobox]");
const content = page.locator("section.content");
const tryAgain = page.getByRole("button", { name: "Try Again", exact: true });
// UI copy carries only the provider name and an HTTP status: never a URL or the fixture host.
const noLeak = async () => {
  const text = await page.evaluate(() => document.body.innerText);
  for (const leak of ["127.0.0.1", "/finer/", "http"]) if (text.includes(leak)) fail(`page text contains "${leak}"`);
};
const step = async (q, ...texts) => {
  await input.fill(q);
  for (const t of texts) await content.filter({ hasText: t }).waitFor({ timeout: 6000 });
  await noLeak();
  return content.innerText();
};
const visible = async (loc) => (await loc.count()) > 0 && (await loc.first().isVisible());
const scrollToEnd = async (text) => {
  const row = page.locator("#grid", { hasText: text });
  for (let i = 0; i < 40 && !(await row.count()); i++) {
    await page.locator("#grid").evaluate((g) => (g.scrollTop = g.scrollHeight));
    await page.waitForTimeout(150);
  }
  if (!(await row.count())) fail(`never saw "${text}" at the end of the grid`);
};

// S8: Finer Gifs with nothing typed (and a single character) makes no request.
await step("", "Search The Office GIFs", "The Finer Gifs Club has no trending list. Type at least 2 characters to search.");
await step("d", "Search The Office GIFs");

// S7
await step("zzempty", "No GIFs found", "Nothing on The Finer Gifs Club matches “zzempty”");
if (await page.locator("#grid").count()) fail("S7 rendered a grid");
await step("", "Search The Office GIFs");

// S2: first page in flight with nothing to show -> 8 skeleton tiles + search-bar spinner; then S11.
await input.fill("slow");
await page.locator(".skeleton > div").nth(7).waitFor({ timeout: 2000 });
if (!(await visible(page.locator("header [aria-label=Loading]")))) fail("S2 without the search-bar spinner");
await content.filter({ hasText: "The Finer Gifs Club is taking too long" }).waitFor({ timeout: 6000 });
if (!(await visible(tryAgain))) fail("S11 without Try Again");
if (await page.locator("header [aria-label=Loading]").count()) fail("spinner still showing after S11");

// S12 / E5f and the 2xx E5 variant.
await step("boom", "The Finer Gifs Club seems to be down", "Its server isn't responding right now.");
if (!(await visible(tryAgain))) fail("E5f without Try Again");
const blank = await step("blank", "The Finer Gifs Club isn't responding", "The Finer Gifs Club returned an error.");
if (blank.includes("(HTTP")) fail("2xx error shows an HTTP parenthetical: " + blank);

// S4: a new query while results show keeps them visible with only the spinner, no dimming.
await step("dwight", "");
await page.locator("#tile-3 img").waitFor({ timeout: 6000 });
await input.fill("dwight slow");
await page.locator("header [aria-label=Loading]").waitFor({ timeout: 2000 });
await page.waitForTimeout(700);
if (!(await visible(page.locator("#tile-0")))) fail("S4 hid the previous results");
const opacity = await page.locator("#grid").evaluate((g) => getComputedStyle(g).opacity);
if (opacity !== "1") fail(`S4 dimmed the grid (opacity ${opacity})`);
await content.filter({ hasText: "is taking too long" }).waitFor({ timeout: 6000 });

// Superseded search: the slow call is cancelled, only the latest result renders.
await input.fill("slow again");
await page.waitForTimeout(700); // past the 450 ms debounce: the slow call is in flight
await input.fill("dwight");
await page.locator("#tile-3 img").waitFor({ timeout: 6000 });
await page.waitForTimeout(3500); // past the 3 s client timeout of the superseded call
if ((await content.innerText()).includes("taking too long")) fail("a superseded search rendered its error");

// S12a: page 2 fails, results stay, the end row offers Try Again.
await step("dwight pageboom", "");
await page.locator("#tile-3 img").waitFor({ timeout: 6000 });
await scrollToEnd("Couldn't load more.");
if (!(await visible(page.locator("#grid").getByRole("button", { name: "Try Again" })))) fail("S12a without Try Again");

// S9 with an unknown retry: Try Again shows; pressing it inside the 60 s local block costs nothing
// and lands on the same state.
await step("noretry", "The Finer Gifs Club request limit reached", "Try again later this hour.");
if (!(await visible(tryAgain))) fail("unknown-retry E1 without Try Again");
await tryAgain.click();
await page.waitForTimeout(300);
await content.filter({ hasText: "Try again later this hour." }).waitFor({ timeout: 3000 });
await saveScreenshot(await page.screenshot(), "gifsmith-03-ratelimit-unknown.png");

// Let the 60 s block from "noretry" run out (MinBlock).
await page.waitForTimeout(61_000);

// S9a: page 2 hits a fresh 429 with Retry-After: 120 -> E2 in the end row, no Try Again.
await step("dwight pagelimit", "");
await page.locator("#tile-3 img").waitFor({ timeout: 6000 });
await scrollToEnd("Request limit reached. More results in 2 min.");
if (await page.locator("#grid").getByRole("button", { name: "Try Again" }).count()) fail("S9a offered Try Again while blocked");

// S9 with a known retry: served by the Go block the 429 above started (no request reaches the fixture).
await step("ratelimit", "The Finer Gifs Club request limit reached", "Try again in 2 minutes (at ");
if (await tryAgain.count()) fail("known-retry E1 shows Try Again before the time passes");
await saveScreenshot(await page.screenshot(), "gifsmith-03-ratelimit.png");

await noLeak();
console.log("UAT OK");
