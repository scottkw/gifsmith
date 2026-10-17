// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-backstops.js
// UI-SPEC backstops at 720x480 in light and dark: the longest state copy (E5f, C2, K6) fits the
// content area without clipping, and the text tokens meet 4.5:1. Fixture keys only; no external links clicked.
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });

const input = page.locator("input[role=combobox]");
const content = page.locator("section.content");
await page.locator(".card input").fill("uat-key-123");
await page.getByRole("button", { name: "Save Key" }).click();
await page.locator("footer").filter({ hasText: "GIPHY key saved" }).waitFor({ timeout: 8000 });

// The state block sits inside the content area and the area does not overflow.
async function fits(name, block, scheme) {
  const r = await content.evaluate((el, sel) => {
    const b = el.querySelector(sel).getBoundingClientRect();
    const c = el.getBoundingClientRect();
    return { sh: el.scrollHeight, ch: el.clientHeight, bottom: b.bottom, right: b.right, left: b.left, cBottom: c.bottom, cRight: c.right };
  }, block);
  if (r.sh > r.ch) fail(`${name} (${scheme}) overflows: scrollHeight ${r.sh} > clientHeight ${r.ch}`);
  if (r.bottom > r.cBottom || r.right > r.cRight || r.left < 0) fail(`${name} (${scheme}) runs outside the content area: ${JSON.stringify(r)}`);
  await saveScreenshot(await page.screenshot(), `gifsmith-06-backstop-${name}-${scheme}.png`);
}

const ratios = [];
for (const scheme of ["light", "dark"]) {
  await page.emulateMedia({ colorScheme: scheme });

  // E5f: The Finer Gifs Club down.
  await page.keyboard.press("Meta+4");
  await input.fill("boom");
  await content.locator("h2", { hasText: "The Finer Gifs Club seems to be down" }).waitFor({ timeout: 8000 });
  await fits("E5f", ".message", scheme);

  // C2: Clips denied for a key that works for GIFs.
  await page.keyboard.press("Meta+2");
  await content.locator("h2", { hasText: "Clips aren't enabled for your key" }).waitFor({ timeout: 8000 });
  await fits("C2", ".message", scheme);

  // K6: the stored GIPHY key rejected during a search (the card is sticky once shown).
  await page.keyboard.press("Meta+1");
  if (await input.isEnabled()) await input.fill("revoked");
  await page.locator("#setup-message", { hasText: "GIPHY rejected your saved key." }).waitFor({ timeout: 8000 });
  await fits("K6", ".card", scheme);

  // Contrast from computed styles: each token resolved to rgb through a probe element.
  const r = await page.evaluate(() => {
    const rgb = (v) => {
      const d = document.createElement("div");
      d.style.color = `var(${v})`;
      document.body.appendChild(d);
      const m = getComputedStyle(d).color.match(/[\d.]+/g).map(Number);
      d.remove();
      return m.slice(0, 3);
    };
    const lum = ([r, g, b]) => {
      const f = (c) => ((c /= 255) <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
      return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b);
    };
    const ratio = (a, b) => {
      const [x, y] = [lum(rgb(a)), lum(rgb(b))].sort((p, q) => q - p);
      return (x + 0.05) / (y + 0.05);
    };
    return {
      "text-secondary/bg": ratio("--text-secondary", "--bg"),
      "text-secondary/surface": ratio("--text-secondary", "--surface"),
      "accent/bg": ratio("--accent", "--bg"),
      "destructive/surface": ratio("--destructive", "--surface"),
    };
  });
  for (const [pair, v] of Object.entries(r)) {
    ratios.push(`${scheme} ${pair} ${v.toFixed(2)}`);
    if (v < 4.5) fail(`${scheme} ${pair} contrast ${v.toFixed(2)} < 4.5`);
  }
}
console.log(ratios.join("; "));

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
