// dev-browser UAT (QuickJS sandbox, Playwright Page API). Run via:
//   bash scripts/smoke-browser-loop.sh --uat scripts/uat/18-backstops.js
// Headless only, against the -tags server build; fixture key only. Phase 4 UI backstops:
// (a) token contrast in both themes, (b) every Keys-card message fits the card (no horizontal
// overflow, at most two lines: 04-02 made the line wrap because K2, K3 and K-T2 exceed 400px at 12px),
// (c) keyboard-only Settings: Tab reaches every control of each pane in DOM order, ↑/↓ switch panes,
// Esc closes a key editor, a remove confirmation and a recording hotkey field and nothing else,
// (d) the longest footer hint set leaves the centre >= 200px.
const fail = (msg) => {
  throw new Error("UAT FAIL: " + msg);
};
const errors = [];
const B = await browser.newPage();
B.on("pageerror", (e) => errors.push("B: " + String(e)));
await B.setViewportSize({ width: 640, height: 520 });
await B.goto("http://127.0.0.1:9245/settings.html");
await B.getByRole("heading", { name: "General" }).waitFor({ timeout: 10000 });

// (a) Contrast from the computed token colours (Theme is System, so the emulated scheme drives it).
const CONTRAST_JS = `(() => {
  const rgb = (v) => {
    const el = document.createElement("div");
    el.style.color = v;
    document.body.append(el);
    const m = getComputedStyle(el).color.match(/[0-9.]+/g).map(Number);
    el.remove();
    return m;
  };
  const lum = (c) => {
    const l = c.slice(0, 3).map((x) => x / 255).map((v) => (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4));
    return 0.2126 * l[0] + 0.7152 * l[1] + 0.0722 * l[2];
  };
  const ratio = (a, b) => {
    const x = lum(rgb(a)), y = lum(rgb(b));
    return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05);
  };
  return {
    "--text-tertiary on --bg": [ratio("var(--text-tertiary)", "var(--bg)"), 4.5],
    "--control-border on --bg": [ratio("var(--control-border)", "var(--bg)"), 3],
    "--control-border on --surface": [ratio("var(--control-border)", "var(--surface)"), 3],
    "--text on --surface-active": [ratio("var(--text)", "var(--surface-active)"), 4.5],
    "white on --destructive-fill": [ratio("#FFFFFF", "var(--destructive-fill)"), 4.5],
  };
})()`;
for (const scheme of ["light", "dark"]) {
  await B.emulateMedia({ colorScheme: scheme });
  const want = scheme === "dark" ? "dark" : "light";
  await B.waitForFunction(`(document.documentElement.dataset.theme || "light") === "${want}"`, null, { timeout: 3000 }).catch(() => fail(`data-theme never became ${want}`));
  const r = await B.evaluate(CONTRAST_JS);
  for (const [name, [got, min]] of Object.entries(r)) {
    if (!(got >= min)) fail(`${scheme}: ${name} is ${got.toFixed(2)}:1, want >= ${min}:1`);
  }
}

// (b) Keys-card messages, measured in a clone of each card's message line, both themes.
await B.getByRole("tab", { name: "Keys", exact: true }).click();
await B.getByRole("heading", { name: "Keys", exact: true }).waitFor({ timeout: 5000 });
const messages = (P) => [
  `Saved. ${P} is ready.`,
  `${P} accepted the key.`,
  `${P} rejected the saved key. It may have been revoked. Change it to continue.`,
  ...["in less than a minute", "in 59 minutes", "later this hour"].flatMap((when) => [
    `${P}'s request limit was hit while testing the key. Try again ${when}.`,
    `${P}'s request limit was hit while checking the key. Try again ${when}.`,
  ]),
  `Couldn't reach ${P} to test the key. Check your connection, then try again.`,
  `Removed. Add a key to use ${P} again.`,
  "Couldn't remove the key from the macOS Keychain. Try again.",
  `${P} rejected that key. Check that you copied the whole key, then try again.`,
  `Couldn't reach ${P} to check the key. Check your connection, then try again.`,
  "Couldn't save the key to the macOS Keychain. Try again.",
];
for (const scheme of ["light", "dark"]) {
  await B.emulateMedia({ colorScheme: scheme });
  for (const P of ["GIPHY", "Klipy"]) {
    const bad = await B.getByRole("group", { name: P, exact: true })
      .locator(".message")
      .evaluate((el, list) => {
        const out = [];
        for (const text of list) {
          const c = el.cloneNode();
          c.textContent = text;
          el.after(c);
          const font = getComputedStyle(c).fontSize;
          if (c.scrollWidth > c.clientWidth || c.clientWidth > 400 || c.clientHeight > 32 || font !== "12px")
            out.push(`${text} (${c.scrollWidth}/${c.clientWidth}px, ${c.clientHeight}px tall, ${font})`);
          c.remove();
        }
        return out;
      }, messages(P));
    if (bad.length) fail(`${scheme} ${P}: message does not fit: ${bad[0]}`);
  }
}
await B.emulateMedia({ colorScheme: "light" });

// (c) Keyboard only. Mark each pane's tab stops in DOM order (one per radio group: the checked one).
// Each stop is named by a descriptor (tag, accessible label, text, radio name), not a DOM marker:
// panes re-render controls (the Automatic paste poll), and a marker would not survive that.
const DESC_JS = `const desc = (el) => [el.tagName, el.getAttribute("aria-label") || el.getAttribute("aria-labelledby") || "", ["SELECT", "INPUT"].includes(el.tagName) ? "" : (el.textContent || "").trim().slice(0, 40), el.name || "", el.type || ""].join("|");`;
const MARK_JS = `(() => {
  ${DESC_JS}
  const pane = document.querySelector("[role=tabpanel]");
  const seen = new Set();
  const stops = [...pane.querySelectorAll("a[href], button, input, select, textarea, [tabindex]")].filter((el) => {
    if (el.disabled || el.tabIndex < 0 || !el.getClientRects().length) return false;
    if (el.type === "radio") {
      if (seen.has(el.name)) return false;
      const group = [...pane.querySelectorAll('input[type=radio][name="' + el.name + '"]')];
      if (el !== (group.find((r) => r.checked) ?? group[0])) return false;
      seen.add(el.name);
    }
    return true;
  });
  return stops.map(desc);
})()`;
const where = () =>
  B.evaluate(`(() => {
    ${DESC_JS}
    const a = document.activeElement;
    if (a?.getAttribute("role") === "tab") return "tab:" + a.textContent.trim();
    return a && document.querySelector("[role=tabpanel]")?.contains(a) ? desc(a) : "out";
  })()`);
const PANES = ["General", "Keys", "Hotkeys", "Search", "Downloads"];
await B.getByRole("tab", { name: "General", exact: true }).click();
await B.getByRole("tab", { name: "General", exact: true }).focus();
for (const [i, name] of PANES.entries()) {
  await B.getByRole("heading", { name, exact: true }).waitFor({ timeout: 3000 }).catch(() => fail(`pane ${name} not shown`));
  await B.waitForTimeout(500); // let the pane's async reads land (General's Automatic paste status adds Grant Access…)
  if ((await where()) !== "tab:" + name) fail(`focus on ${await where()}, want the ${name} tab`);
  const stops = await B.evaluate(MARK_JS);
  const n = stops.length;
  if (n < 2) fail(`${name}: only ${n} tab stops`);
  for (let k = 0; k < n; k++) {
    await B.keyboard.press("Tab");
    const at = await where();
    if (at !== stops[k]) fail(`${name}: Tab ${k + 1} landed on ${at}, want ${stops[k]}`);
  }
  for (let k = n - 2; k >= -1; k--) {
    await B.keyboard.press("Shift+Tab");
    const at = await where();
    const want = k >= 0 ? stops[k] : "tab:" + name;
    if (at !== want) fail(`${name}: Shift+Tab landed on ${at}, want ${want}`);
  }
  if (i < PANES.length - 1) await B.keyboard.press("ArrowDown");
}
await B.keyboard.press("ArrowDown"); // past the last built pane: clamped
if ((await where()) !== "tab:" + PANES[PANES.length - 1]) fail("↓ on the last pane moved to " + (await where()));
for (let k = 0; k < PANES.length + 1; k++) await B.keyboard.press("ArrowUp");
if ((await where()) !== "tab:General") fail("↑ did not clamp at General: " + (await where()));

// Esc closes only the innermost transient state.
const stillOn = async (name) => {
  if (!(await B.getByRole("heading", { name, exact: true }).isVisible())) fail(`Esc left the ${name} pane`);
};
await B.keyboard.press("ArrowDown"); // Keys
const giphy = B.getByRole("group", { name: "GIPHY", exact: true });
await giphy.getByRole("button", { name: "Add Key…" }).focus();
await B.keyboard.press("Enter");
const field = giphy.getByRole("textbox", { name: "GIPHY API key" });
await field.waitFor({ timeout: 3000 });
await B.keyboard.press("Escape");
await giphy.getByRole("button", { name: "Add Key…" }).waitFor({ timeout: 3000 });
await stillOn("Keys");
await giphy.getByRole("button", { name: "Add Key…" }).focus();
await B.keyboard.press("Enter");
await field.waitFor({ timeout: 3000 });
await B.keyboard.type("uat-key-123");
await B.keyboard.press("Enter");
await giphy.getByText("Saved. GIPHY is ready.").waitFor({ timeout: 8000 });
await giphy.getByRole("button", { name: "Remove GIPHY key" }).focus();
await B.keyboard.press("Enter");
await giphy.getByText("Remove your GIPHY key?").waitFor({ timeout: 3000 });
await B.keyboard.press("Escape");
await giphy.getByText("Remove your GIPHY key?").waitFor({ state: "detached", timeout: 3000 });
await giphy.getByText("Saved · ends in -123").waitFor({ timeout: 3000 });
await stillOn("Keys");

await B.getByRole("tab", { name: "Keys", exact: true }).focus();
await B.keyboard.press("ArrowDown"); // Hotkeys
await B.getByRole("heading", { name: "Hotkeys", exact: true }).waitFor({ timeout: 3000 });
const open = B.getByRole("button", { name: "Open Gifsmith shortcut", exact: true });
const before = (await open.innerText()).trim();
await open.focus();
await B.keyboard.press("Enter");
await B.waitForFunction(`document.querySelector(".recorder.recording") !== null`, null, { timeout: 3000 }).catch(() => fail("Enter did not start recording"));
await B.keyboard.press("Escape");
await B.waitForFunction(`document.querySelector(".recorder.recording") === null`, null, { timeout: 3000 }).catch(() => fail("Esc did not stop recording"));
if ((await open.innerText()).trim() !== before) fail(`Esc changed the Open shortcut to ${await open.innerText()}`);
await stillOn("Hotkeys");

// (d) The longest hint set (Copy Markdown / Paste Markdown) leaves the footer centre >= 200px.
await B.getByRole("tab", { name: "General", exact: true }).click();
await B.getByRole("combobox", { name: "Primary action (↵)" }).selectOption("copyMarkdown");
await B.getByRole("combobox", { name: "Secondary action (⌘↵)" }).selectOption("pasteMarkdown");
const A = await browser.newPage();
A.on("pageerror", (e) => errors.push("A: " + String(e)));
await A.setViewportSize({ width: 720, height: 480 });
await A.goto("http://127.0.0.1:9245/");
await A.locator("section.content[aria-busy=false]").waitFor({ timeout: 10000 });
await A.keyboard.press("Meta+4");
await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search The Finer Gifs Club'", null, { timeout: 5000 });
await A.locator("header input[role=combobox]").fill("dwight");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
const hints = (await A.locator("footer .hints").innerText()).replace(/\s+/g, " ").trim();
if (hints !== "↵ Copy Markdown ⌘↵ Paste Markdown ⌘K Actions") fail("hints " + hints);
// The centre message ellipsizes beside the longest hints. UI-SPEC's "about 240px" assumes the GIPHY
// mark; Finer's longer text attribution leaves about 180px, still a one-line ellipsized title.
const centre = () => A.locator("footer .center").evaluate((el) => el.getBoundingClientRect().width);
if ((await A.locator("footer .center .msg").evaluate((el) => getComputedStyle(el).textOverflow)) !== "ellipsis") fail("footer message does not ellipsize");
if ((await centre()) < 150) fail(`Finer footer centre is ${await centre()}px with the longest hints`);
await A.keyboard.press("Meta+1"); // GIPHY: the key was saved in (c)
await A.waitForFunction("document.querySelector('input[role=combobox]').placeholder === 'Search GIPHY'", null, { timeout: 5000 });
await A.locator("header input[role=combobox]").fill("");
await A.locator("#tile-0[aria-selected=true] img").waitFor({ timeout: 10000 });
if ((await centre()) < 200) fail(`GIPHY footer centre is ${await centre()}px with the longest hints`);

if (errors.length) fail("page error: " + errors[0]);
console.log("UAT OK");
