// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/14-backstops.js
// Phase 3 UI backstops at 720x480 in light and dark: a11y contrast (scrim, selected chip/menu row,
// destructive rows), keyboard-only reach of the action menu, Details and chips, long text (setup
// card message slot P8, Details title clamp, chip ellipsis, toast ellipsis), plus the Phase 2
// polish checks (P1-P3, P6, P7). Fixture keys only; no external links clicked; downloads land in
// bin/uat-downloads (server build).
const page = await browser.newPage();
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
page.on("pageerror", (e) => errors.push(String(e)));

await page.setViewportSize({ width: 720, height: 480 });
await page.goto("http://127.0.0.1:9245/");
await page.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
const input = page.locator("header input[role=combobox]");
const dialog = page.locator("[role=dialog][aria-label=Actions]");
const filter = dialog.locator("input[role=combobox]");
const details = page.locator("#details");
const chipRow = page.locator("[role=listbox][aria-label='Recent searches']");
const css = (loc, prop) => loc.evaluate((el, p) => getComputedStyle(el)[p], prop);
const token = (v) =>
  page.evaluate((name) => {
    const d = document.createElement("div");
    d.style.color = `var(${name})`;
    document.body.appendChild(d);
    const c = getComputedStyle(d).color;
    d.remove();
    return c;
  }, v);
const focused = () => page.evaluate(() => document.activeElement?.getAttribute("aria-label") || document.activeElement?.tagName);
const inputFocused = () => input.evaluate((el) => document.activeElement === el);

// P6 / P7 / P1 on the first-run GIPHY card: disabled Save Key without opacity, disabled search
// input text in --text-tertiary, the key field keeps the shared ring.
const field = page.locator(".card input");
await field.waitFor({ timeout: 5000 });
const save = page.getByRole("button", { name: "Save Key" });
if (!(await save.isDisabled())) fail("Save Key enabled with an empty field");
if ((await css(save, "opacity")) !== "1") fail("disabled Save Key uses opacity");
if ((await css(save, "backgroundColor")) !== (await token("--surface-active"))) fail("disabled Save Key background");
if ((await css(save, "color")) !== (await token("--text-tertiary"))) fail("disabled Save Key text");
if ((await css(input, "color")) !== (await token("--text-tertiary"))) fail("disabled search input text " + (await css(input, "color")));
await field.focus();
if ((await css(field, "outlineStyle")) !== "solid" || (await css(field, "outlineWidth")) !== "2px") fail("key field lost the focus ring");

// P8: every K-row, K6 and F9 string fits the card's one-line message slot (<= 500px at 12px).
await page.keyboard.press("Meta+3");
await field.fill("bad-key");
await field.press("Enter");
const slot = page.locator("#setup-message");
await slot.filter({ hasText: "Klipy rejected that key" }).waitFor({ timeout: 8000 });
const P8 = [];
for (const name of ["GIPHY", "Klipy"]) {
  P8.push(
    `${name} rejected that key. Check that you copied the whole key, then try again.`,
    `${name}'s request limit was hit while checking the key. Try again in less than a minute.`,
    `${name}'s request limit was hit while checking the key. Try again in 59 minutes.`,
    `${name}'s request limit was hit while checking the key. Try again later this hour.`,
    `Couldn't reach ${name} to check the key. Check your connection, then try again.`,
    `${name} rejected your saved key. It may have been revoked. Paste a new key to continue.`,
  );
}
P8.push("Couldn't save the key to the macOS Keychain. Try again.");
let widest = 0;
for (const s of P8) {
  const m = await slot.evaluate((el, text) => {
    const node = [...el.childNodes].reverse().find((n) => n.nodeType === 3 && n.data.trim());
    node.data = text;
    const r = document.createRange();
    r.selectNodeContents(el);
    const t = document.createRange();
    t.selectNodeContents(node);
    return { w: r.getBoundingClientRect().width, tw: t.getBoundingClientRect().width, sw: el.scrollWidth, cw: el.clientWidth, font: getComputedStyle(el).fontSize };
  }, s);
  // UI-SPEC P8 measures the string itself (K6 = 493px); the 12px icon and 4px gap sit in the
  // slot's remaining 12px, which scrollWidth <= clientWidth proves.
  widest = Math.max(widest, m.tw);
  if (m.font !== "12px" || m.tw > 500 || m.sw > m.cw) fail(`P8 "${s}": ${JSON.stringify(m)}`);
}
console.log(`P8 widest ${widest.toFixed(1)}px`);

// Save the GIPHY key; create two recent searches (one longer than 160px) for the chip checks.
await page.keyboard.press("Meta+1");
await field.fill("uat-key-123");
await field.press("Enter");
await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
const LONG = "an extremely long recent search that will not fit";
for (const q of ["cats", LONG]) {
  await input.fill(q);
  await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
  await page.waitForTimeout(1700);
}
await input.press("Escape");
await chipRow.waitFor({ timeout: 5000 });
const longChip = page.locator("#chip-0");
if ((await longChip.getAttribute("title")) !== LONG) fail("long chip title");
const label = await longChip.locator(".label").evaluate((el) => ({ w: el.getBoundingClientRect().width, sw: el.scrollWidth, cw: el.clientWidth, to: getComputedStyle(el).textOverflow }));
if (label.w > 160.5 || label.sw <= label.cw || label.to !== "ellipsis") fail("chip label not ellipsized at 160px: " + JSON.stringify(label));

const ratios = [];
for (const scheme of ["light", "dark"]) {
  await page.emulateMedia({ colorScheme: scheme });
  await page.keyboard.press("Meta+1");
  await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });

  // a11y-contrast: white on --scrim composited over white and over --surface; --text on
  // --surface-active (selected chip, highlighted menu row); --destructive on --bg (destructive rows).
  const r = await page.evaluate(() => {
    const color = (v) => {
      const d = document.createElement("div");
      d.style.color = `var(${v})`;
      document.body.appendChild(d);
      const m = getComputedStyle(d).color.match(/[\d.]+/g).map(Number);
      d.remove();
      return m.length === 4 ? m : [...m, 1];
    };
    const over = ([r, g, b, a], [R, G, B]) => [r * a + R * (1 - a), g * a + G * (1 - a), b * a + B * (1 - a)];
    const lum = ([r, g, b]) => {
      const f = (c) => ((c /= 255) <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
      return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b);
    };
    const ratio = (a, b) => {
      const [x, y] = [lum(a), lum(b)].sort((p, q) => q - p);
      return (x + 0.05) / (y + 0.05);
    };
    const white = [255, 255, 255];
    const scrim = color("--scrim");
    return {
      "white/scrim-over-white": ratio(white, over(scrim, white)),
      "white/scrim-over-surface": ratio(white, over(scrim, color("--surface"))),
      "text/surface-active": ratio(color("--text"), color("--surface-active")),
      "destructive/bg": ratio(color("--destructive"), color("--bg")),
    };
  });
  for (const [pair, v] of Object.entries(r)) {
    ratios.push(`${scheme} ${pair} ${v.toFixed(2)}`);
    if (v < 4.5) fail(`${scheme} ${pair} contrast ${v.toFixed(2)} < 4.5`);
  }

  // a11y-keyboard: the menu, Details and chips from the keyboard alone; focus never leaves the input
  // (or the menu filter while the menu is open).
  if (!(await inputFocused())) fail(`${scheme}: focus not in the search input: ${await focused()}`);
  await input.press("Meta+k");
  await dialog.waitFor({ timeout: 2000 });
  if (!(await filter.evaluate((el) => document.activeElement === el))) fail(`${scheme}: menu filter not focused`);
  let prev = await filter.getAttribute("aria-activedescendant");
  for (let i = 0; i < 3; i++) {
    await filter.press("ArrowDown");
    const id = await filter.getAttribute("aria-activedescendant");
    const hi = await dialog.locator("[role=option][aria-selected=true]").getAttribute("id");
    if (!id || id === prev || id !== hi) fail(`${scheme}: aria-activedescendant ${id} (was ${prev}, highlighted ${hi})`);
    if (!(await filter.evaluate((el) => document.activeElement === el))) fail(`${scheme}: focus left the menu filter`);
    prev = id;
  }
  await filter.press("Escape");
  await dialog.waitFor({ state: "detached", timeout: 2000 });
  if (!(await inputFocused())) fail(`${scheme}: focus not back in the input after the menu`);
  await input.press("Meta+i");
  await details.waitFor({ timeout: 2000 });
  if (!(await inputFocused())) fail(`${scheme}: focus left the input in Details`);
  await input.press("Escape");
  await details.waitFor({ state: "detached", timeout: 2000 });
  await input.press("ArrowUp");
  if ((await input.getAttribute("aria-activedescendant")) !== "chip-0") fail(`${scheme}: ↑ did not reach the chips`);
  const chipBg = await css(page.locator("#chip-0"), "backgroundColor");
  if (chipBg !== (await token("--surface-active"))) fail(`${scheme}: selected chip background ${chipBg}`);
  await input.press("ArrowDown");
  if ((await input.getAttribute("aria-activedescendant")) !== "tile-0") fail(`${scheme}: ↓ did not leave the chips`);
  if (!(await inputFocused())) fail(`${scheme}: focus left the input on the chips`);

  // P1: the search input opts out of the shared ring.
  if ((await css(input, "outlineStyle")) !== "none") fail(`${scheme}: search input shows a focus ring`);

  // long-text: a 300-character Details title clamps to 3 lines.
  await input.press("Meta+i");
  await details.waitFor({ timeout: 2000 });
  const t = await details.locator("h2").evaluate((el) => {
    el.textContent = "Long title ".repeat(28).slice(0, 300);
    const lh = parseFloat(getComputedStyle(el).lineHeight);
    return { h: el.getBoundingClientRect().height, sh: el.scrollHeight, lh };
  });
  if (t.h > 3 * t.lh + 1 || t.sh <= t.h) fail(`${scheme}: Details title not clamped to 3 lines: ${JSON.stringify(t)}`);
  await saveScreenshot(await page.screenshot(), `gifsmith-14-details-long-${scheme}.png`);
  await input.press("Escape");
  await details.waitFor({ state: "detached", timeout: 2000 });

  // long-text: the toast message ellipsizes; the action link and its hint keep their width.
  await input.press("Meta+s");
  const action = page.locator("footer .center button", { hasText: "Show in Finder" });
  await action.waitFor({ timeout: 8000 });
  const before = await page.locator("footer .center").evaluate((el) => ({
    a: el.querySelector("button").getBoundingClientRect().width,
    h: el.querySelector(".hint").getBoundingClientRect().width,
  }));
  const after = await page.locator("footer .center").evaluate((el) => {
    const msg = el.querySelector(".msg");
    const node = [...msg.childNodes].reverse().find((n) => n.nodeType === 3 && n.data.trim());
    node.data = "Downloaded to " + "a very long folder name ".repeat(12);
    return {
      a: el.querySelector("button").getBoundingClientRect().width,
      h: el.querySelector(".hint").getBoundingClientRect().width,
      msw: msg.scrollWidth,
      mcw: msg.clientWidth,
      to: getComputedStyle(msg).textOverflow,
    };
  });
  if (after.msw <= after.mcw || after.to !== "ellipsis") fail(`${scheme}: toast message does not ellipsize: ${JSON.stringify(after)}`);
  if (Math.abs(after.a - before.a) > 0.5 || Math.abs(after.h - before.h) > 0.5) fail(`${scheme}: toast action/hint shrank: ${JSON.stringify({ before, after })}`);
  await saveScreenshot(await page.screenshot(), `gifsmith-14-toast-long-${scheme}.png`);
  await input.press("ArrowRight"); // a selection change does not clear the download toast; ⌘R would
  await input.press("ArrowLeft");

  // P3: the active Clips pill reads as active even when Clips are unavailable (FAKE_CLIPS=denied).
  await page.keyboard.press("Meta+2");
  await page.locator("section.content h2", { hasText: "Clips aren't enabled for your key" }).waitFor({ timeout: 8000 });
  const pill = page.locator("[role=tab]", { hasText: "Clips" });
  if ((await css(pill, "color")) !== (await token("--text"))) fail(`${scheme}: active Clips pill color ${await css(pill, "color")}`);
  if ((await css(pill, "backgroundColor")) !== (await token("--surface-active"))) fail(`${scheme}: active Clips pill background`);
  // P1: the C2 buttons get the shared ring on keyboard focus.
  const back = page.getByRole("button", { name: "Back to GIPHY" });
  await back.focus();
  await page.keyboard.press("Shift"); // keyboard modality for :focus-visible
  if ((await css(back, "outlineStyle")) !== "solid") fail(`${scheme}: Back to GIPHY has no focus ring`);
  await page.keyboard.press("Meta+1");
  await page.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
  const off = page.locator("[role=tab]", { hasText: "Clips" });
  if ((await css(off, "color")) !== (await token("--text-tertiary"))) fail(`${scheme}: inactive unavailable Clips pill not dimmed`);
  await input.focus();
}
console.log(ratios.join("; "));

// P2: the Enable Automatic Paste… button is a 24px secondary button (⌘↵ in the server build reports
// that Accessibility is needed; platform.Fake pastes nothing).
await input.press("Meta+Enter");
const enable = page.locator("footer button", { hasText: "Enable Automatic Paste…" });
await enable.waitFor({ timeout: 8000 });
const eb = await enable.boundingBox();
if (Math.round(eb.height) !== 24) fail("Enable Automatic Paste… height " + eb.height);
if ((await css(enable, "backgroundColor")) !== (await token("--surface-active"))) fail("Enable Automatic Paste… background");
if ((await css(enable, "fontWeight")) !== "600") fail("Enable Automatic Paste… weight");

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
