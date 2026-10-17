# Feature Landscape

**Domain:** macOS menu-bar / hotkey GIF search launcher (port of Raycast "GIF Search")
**Project:** Gifsmith
**Researched:** 2026-10-17
**Overall confidence:** HIGH for the Raycast inventory (read from source) and provider rules (read from vendor docs). MEDIUM for competitor feature claims (reviews and store pages). LOW where marked.

Confidence tags used below: **[SRC]** read from the extension source (HIGH). **[DOC]** read from a vendor's official docs (HIGH for Klipy, read as raw markdown. MEDIUM for GIPHY, read through a summarizing fetch). **[WEB]** third-party or store-page claim (MEDIUM). **[INF]** my inference or design recommendation (LOW to MEDIUM, a judgement call and not a fact).

---

## 1. Parity Inventory: what the Raycast extension actually does

Source: `gif-search` v. changelog through 2026-09-30, about 2k lines of TypeScript. Everything in this section is **[SRC]** unless noted.

### 1.1 Commands and entry points

| Raycast surface | Behavior | Gifsmith equivalent |
|---|---|---|
| `search` (view) | The main grid with a provider dropdown in the search bar | The hotkey panel |
| `favorites` (view) | Opens straight into the all-provider Favorites grid. Has its own hotkey or alias | A separate hotkey or panel mode, or a menu-bar item |
| `recents` (view) | Same idea, for Recents | Same |
| `search-gifs` AI tool | Search returning media and page links | **Out of scope** (PROJECT.md) |

### 1.2 Search view behavior

- **Provider switcher:** GIPHY GIFs, GIPHY Clips, Klipy, Finer Gifs Club. A second section holds the two virtual views, Favorites and Recent. Each provider has a logo icon.
- **Selected provider persists** between invocations (`useCachedState("service")`). The default is GIPHY.
- **Placeholder** changes per provider: "Search for GIFs on {provider}", "Search favorites", "Search recents".
- **Search is throttled.**
- **Empty search** shows three sections stacked in one grid:
  1. **Favorites** for the current provider, capped to the first 2 rows (`columns * 2`).
  2. **Recent** for the current provider, same cap, with anything already in Favorites removed.
  3. **Trending** from the provider.
- **With a search term** there is one section titled `GIF results for "{term}"`. The Favorites and Recent sections disappear.
- **Pagination** is infinite scroll. The page size is `maxResults` (default 20, and an empty or invalid value falls back to 20).
  - Paging cursor: GIPHY and Finer Gifs use offset. Klipy uses a `next` cursor. `hasMore` is derived from the cursor, or from `results.length === limit` when there is no cursor.
  - Results are de-duplicated by ID.
  - The previous results stay visible while a new query loads (`keepPreviousData`).
- **Finer Gifs has no trending.** Its trending call returns `[]`, so an empty search shows only Favorites and Recent plus the empty state.
- **Empty and error states** (five distinct states):
  - "Enter a search above to get started…"
  - "No GIFs were found."
  - A provider error message (the full error text in the empty state, plus a failure toast "Could not load GIFs from X").
  - "Add some GIFs to your Favorites first…"
  - "Work with some GIFs first…"
- **Favorites and Recent virtual views:**
  - They show **all providers**, grouped into one section per provider.
  - The search box becomes a client-side title filter.
  - There is no pagination.
  - A provider whose lookup fails does not blank the others (partial-failure toast). If every provider fails, the last good cache stays on screen.
- **Grid density:** the columns are small 8, medium 5, large 3. There are two preferences, one for search results and one for the empty-search view (Trending, Favorites, Recents). The thumbnail source depends on size. Large uses a bigger rendition, and small uses the preview GIF.
- **Grid item:** a thumbnail plus a title. There is no attribution or provider badge on the grid item.

### 1.3 Actions (13) and their shortcuts

Raycast shows these in an action panel (⌘K). The first action is the one that runs on ⏎.

| Action | Shortcut in extension | Behavior details |
|---|---|---|
| **Copy GIF** (default) | ⌘⌥C | Resolve the file, write the **file** to the pasteboard, record as recent, **close the window**, toast `Copied GIF "name"` |
| **Paste GIF** | ⌘⌥P | Resolve the file and paste it into the frontmost app. It does **not** copy first, and it does not close the window first |
| Copy GIF Link | none | Copies `gif_url` with the query string stripped |
| Paste GIF Link | ⌘⌥L | Same URL, pasted |
| Copy GIF Markdown | ⌘⇧⏎ | `![title](gif_url-without-query)` |
| Paste GIF Markdown | ⌘⇧P | Same string, pasted |
| Copy GIF Square | none | Center-crop to a square, re-encode as a GIF, copy the file. Keeps frame delays and loop count. macOS only |
| Paste GIF Square | none | Same, pasted |
| Add or Remove Favorite | ⌘⇧F | Toggles based on current state. Removing evicts the cached file |
| Remove from Recents | ⌃⇧R | Shown only if the GIF is in Recents |
| View GIF Details | ⌘⇧D | Pushes the details view. Hidden inside the details view itself |
| Copy Page Link | Raycast's `Common.Copy` | Only when the item has a page URL |
| Open in Browser | ⌘⇧B | Only when the item has a page URL |
| Download GIF | ⌘⌥D | Saves to the download folder. The toast offers "Open File" (⌘O) and "Show in Finder" (⌘⇧O) |

Behaviors worth carrying over:

- **Default-action preference** has 13 choices (all of the above, including Toggle Favorite and View Details). Whichever is chosen is **moved to the top of the panel** and becomes ⏎. If it lived in a lower section, it becomes its own first section.
- **Usage tracking:** every successful action (copy, paste, link, open, details push) adds the GIF to Recents and bumps its frecency. A tracking failure must never fail the user's action. The code comments say so, and it is tested.
- **Hide filename** replaces the file name with `gif.gif` or `gif.mp4` on copy and download. It is a display name only and must never be the cache key.
- **Download names:** a download never overwrites. It claims the name by exclusive create and then writes `name (1).gif`, `name (2).gif` and so on. A partial download removes the file it created. The default folder is `~/Downloads`, overridable.
- **Clips:** GIPHY Clips download the best MP4 (1080p, then 720p, then 360p, then the GIF). The file keeps its `.mp4` extension. "Copy GIF" on a clip therefore puts an **MP4 file** on the clipboard. Link and Markdown actions use the GIF rendition.
- **Error taxonomy:** timeout at 15 seconds with a "server may be slow" message, "Could not reach X. Check your connection, VPN, proxy, or firewall", and HTTP status failures. They all surface as toasts with a provider name.

### 1.4 Details view

- An animated preview, plus a metadata panel:
  - Title.
  - Width and height in px, and file size in SI units.
  - "Created" (relative time) and "User" (GIPHY).
  - Source link (GIPHY's original source site, shown as its TLD).
  - Tag list.
  - **Source** (the provider name with its logo icon).
- Finer Gifs adds Season and Episode labels, and a TVDB link to the season.
- The same action panel is available, minus View Details.
- Klipy provides Created, tags and dimensions, but no user.

### 1.5 Favorites and Recents semantics

- **Storage** is per-provider ID lists (`{service}-favs`, `{service}-recent`). Only the IDs are stored. **Metadata is re-fetched from the provider on every view.** IDs are unique only within a provider.
- **Favorites append. Recents prepend.** Re-using a recent moves it to the top. **Recents have no cap** in the extension, so they grow forever.
- **Favorite media is cached on disk**, keyed by `sha1(service:id)` plus extension. Copy, paste and download use the cache first. Removing a favorite evicts it. Corrupt or missing cache entries fall back to a download.
- **Frecency** (from `@raycast/utils`, read from source):
  - Each use does `score = score * 0.5^(daysSinceLastUse/10) + 100`.
  - Items with a score sort above never-used items. Never-used items keep their original order.
  - Frecency applies only to the local sections (Favorites and Recents), and each is keyed by provider as a namespace.
  - **Quirks not to copy:** the sort compares the score as of the last visit and never decays it to "now". The Favorites and Recents virtual views share one namespace across providers, so ID collisions are possible. The changelog says Recents sort by most-recent first, yet the code frecency-sorts them as well.
- **Limits:**
  - Finer Gifs ID lookup requests `size=10`, so only 10 favorites would ever load. That is an existing bug in the lookup.
  - GIPHY ID lookup supports up to 100 IDs per call [DOC].

### 1.6 Preferences (Raycast) vs the PROJECT.md target

| Raycast preference | Values | In PROJECT.md? |
|---|---|---|
| Grid Item Size | small, medium, large | Yes |
| Grid Trending Item Size | small, medium, large | Yes (sizes) |
| Max Results | integer, default 20 | Yes |
| Default Action | 13 choices | Yes |
| GIF Download Path | directory | Yes |
| Hide Filename | checkbox | Yes |
| GIPHY Locale | 31 locales | Yes |
| Klipy Locale | 10 locales | Yes |
| (none) API keys | The extension uses Raycast's keyless proxy | PROJECT.md adds keys, **a new requirement not in the original** |

**Not in the extension, but needed for Gifsmith:** the global hotkey, API keys, launch at login, panel behavior, and a content-rating setting (see D-5).

**Gap flagged for requirements:** the extension's Klipy client speaks Klipy's **Tenor-compatible** response shape (`media_formats`, `itemurl`, `nanogif`). Klipy's own docs also expose a native API (`/api/v1/{app_key}/gifs/search`) [DOC]. Which one to target is a stack and architecture decision. The feature set below does not depend on the choice.

---

## 2. Ecosystem context (competitors and constraints)

### 2.1 Provider constraints that shape the feature set

| Fact | Source | Feature consequence |
|---|---|---|
| **Tenor's API was shut down on 2026-06-30** and new keys had been closed since Jan 2026. Discord, WhatsApp and Bluesky lost their pickers | [WEB] (several news reports agree) | Do **not** add Tenor. The extension already removed it (2026-06-30 changelog) |
| GIPHY beta keys: **100 calls per hour**, max 50 results per request. A production key needs review and is paid | [DOC] (GIPHY docs) | Debounce, cache trending, and don't spend calls on browsing local Favorites. **Every infinite-scroll page costs a call** |
| Klipy "Testing" keys: **100 requests per hour**. Production needs a request in the Partner Panel | [DOC] (Klipy getting-started) | Same budget. The key comes from partner.klipy.com, API Keys, create platform |
| **GIPHY requires a conspicuous "Powered By GIPHY" mark** where the API is used | [DOC] | Attribution UI is mandatory for GIPHY and Clips (see section 5) |
| **Klipy requires the search placeholder "Search KLIPY"**. The watermark and "Powered by KLIPY" mark are optional | [DOC] (read directly) | The extension's placeholder "Search for GIFs on Klipy" would **not** comply |
| Both vendors say: **preserve returned URLs and parameters**, **don't reorder or filter** Search and Trending results, and **don't mix their content with another provider's in one grid** | [DOC] | Per-provider sections only (the extension already does this). `stripQParams` on the media URL is a grey area (see PITFALLS) |
| Both vendors say: **don't store, mirror or cache media without approval**. Klipy also wants requests to originate from the end-user client | [DOC] | Temp-file staging for the clipboard is unavoidable. The persistent favorites file cache is a **ToS-risk** feature (see D-8) |
| GIPHY exposes autocomplete, related tags, trending searches, categories, a `rating` filter, lookup by ID, and analytics pingbacks. Klipy exposes categories, autocomplete, search suggestions, share-trigger and recents-per-customer | [DOC] | Enables the differentiators in section 4 |
| Zulip's integration docs say these keys "are not secrets" because every client gets a copy | [DOC] (Zulip docs) | Keychain is still fine, but a leaked key is not a security incident |

### 2.2 Competitor snapshot

| Product | What it has | Notes |
|---|---|---|
| **Raycast GIF Search** | Everything in section 1 | The baseline. Raycast-only, with no drag-out and no rating filter |
| **Quickgif** (Mac App Store) | Menu bar, **global hotkey**, **drag and drop into any app**, copy link instead of file, unique filenames. Switched to **Klipy** in v1.4.0 (June) | [WEB] **Bookmarks sit behind a $2.99 premium.** Favorites being free is a selling point |
| **Gnome** (May 2026, $7) | Menu-bar GIF finder with a **local library (tag, search)** and hotkey. Selecting copies the GIF | [WEB] Local-library angle. Out of scope here |
| **Grabbie, GIFme, GIF Bar, Hard G, PopKey** | Menu-bar GIPHY search, **drag and drop**, copy file or URL, favorites (Grabbie), auto-hide after an action | [WEB] Mostly GIPHY-only. Several are dated |
| **Alfred workflows** (mr-pennyworth, kejadlen) | Query to a result list to copy. GIPHY or GIPHY plus Tenor | [WEB] Thin. No grid, no favorites, no previews to speak of. The key is a workflow variable |
| **Gifox / GIPHY Capture** | GIF **recording** and sharing, not search | Different category (anti-feature here) |
| **Slack, Discord pickers** | Favorites (star), category tiles, search suggestions, trending | [INF] From general knowledge and not re-verified this session. Treat as LOW confidence |

**Takeaway:** the typical menu-bar competitor is "search GIPHY, click, copy or drag". Raycast's extension beats them on action breadth, defaults, favorites and recents. **Gifsmith's opportunity is Raycast parity as a standalone app, plus the things Raycast can't do**: drag-out, offline favorites, and a keyless first run.

---

## 3. Table Stakes

Missing any of these makes the product feel broken or incomplete. IDs are for the requirements doc.

| ID | Feature | Why expected | Complexity | Notes |
|---|---|---|---|---|
| T-1 | Global hotkey opens a floating panel. Esc or focus loss dismisses it. Search field focused on open | Core value | Med | The panel must appear over full-screen apps and restore focus to the previous app (native work). Hotkey rebinding is also required |
| T-2 | Menu-bar icon, no Dock icon (agent app), plus a menu with Settings, Favorites, Recents, Quit | Launcher convention | Low | |
| T-3 | Search and Trending for the four providers with a provider switcher that **remembers the last provider** | Parity | Med | Finer Gifs has no trending. Show Favorites and Recents instead |
| T-4 | Animated GIF grid with infinite scroll, a page size setting, and de-duplication | Parity | Med | Must also load more on **keyboard selection** (see KB-4) |
| T-5 | **Copy GIF as a file** to the pasteboard (default action) | Core value | Med | Needs a native file-URL pasteboard write. No Accessibility permission needed |
| T-6 | **Paste GIF into the frontmost app** | Parity, and the other half of the core value | High | Needs Accessibility permission, a CGEvent ⌘V, and focus restore. Gate it behind a clear permission prompt. Copy must still work without it |
| T-7 | Copy and paste for link and Markdown. Copy Page Link, Open in Browser, Download | Parity | Low | Reuse the extension's string formats |
| T-8 | Copy and paste **square-cropped** GIF, done natively | Parity | Med | Re-encode with ImageIO or Go `image/gif`. **Preserve frame delays and loop count** (the Swift script does). Clips are MP4, so see PITFALLS before porting |
| T-9 | **Action panel (⌘K)** listing every action with its shortcut, filterable by typing | 13 actions are undiscoverable otherwise | Med | Also lets users learn the default action |
| T-10 | **Configurable default (⏎) action**, defaulting to Copy GIF | Parity, and it is what lets people replace Raycast habits | Low | All 13 choices. The chosen one is promoted to the top of the panel |
| T-11 | Favorites and Recents **per provider**, with both views browsable (all providers, grouped by provider) | Parity | Med | See D-7 on storing metadata. Dedupe on (provider, id). **Cap Recents** (the extension never does) |
| T-12 | Frecency ordering for Favorites. Recents by recency | Parity | Low | Use half-life 10 days and +100 per visit, but **decay to now** when sorting |
| T-13 | GIF details view: animated preview, size, dimensions, created, user, source link, tags, provider | Parity, and "where is this GIF from?" is in the extension's FAQ | Med | Finer Gifs adds season and episode labels and a TVDB link |
| T-14 | Settings: API keys, hotkey, grid sizes (search and empty-state), max results, default action, download folder, hide filename, GIPHY and Klipy locales | Parity | Med | Locales: 31 for GIPHY and 10 for Klipy. The keys apply to GIPHY, Clips and Klipy |
| T-15 | **Keyless first run** that works (Finer Gifs) and a guided key setup for the keyed providers (section 6) | The BYOK requirement means a bad first run loses everyone | Med | See section 6 |
| T-16 | **Provider attribution** that meets each provider's rules (section 5) | ToS compliance. It also affects any later production-key application | Low to Med | Mandatory marks for GIPHY. "Search KLIPY" placeholder for Klipy |
| T-17 | Honest, specific **errors and rate-limit feedback**: invalid key, 429, timeout, offline. Never a blank grid | Parity (a whole changelog entry went into error handling) | Med | See section 6.4 |
| T-18 | Loading and empty states, plus a toast or HUD for each action result | Parity | Low | Panel closes after copy. A confirmation shows (sound or HUD) |
| T-19 | **Keyboard-first grid navigation** (section 7) | The product is a launcher | Med | Arrows, ⏎, ⌘⏎, ⌘K, Esc, provider switching by key |
| T-20 | Signed and notarized build, **retaining the original MIT attribution** | PROJECT.md constraint | Low | A license file and an About entry |

---

## 4. Differentiators

Not required for parity, but valued. Ordered roughly by value for cost.

| ID | Feature | Value | Complexity | Notes |
|---|---|---|---|---|
| D-1 | **Drag the selected GIF out** of the grid into any app (Slack, Messages, Mail) | The most common feature among Mac menu-bar competitors (Quickgif, GIF Bar, Hard G) [WEB], and Raycast can't do it | High | Needs a native file drag source (NSDraggingSource). A web view alone won't drag a file out. De-risk it in the spike if wanted |
| D-2 | **Paired shortcut for the sibling action**: ⏎ runs the default, ⌘⏎ runs its paste or copy twin | Makes "copy or paste?" a keypress and not a settings detour | Low | [INF] Raycast gives the second action ⌘⏎ by default. Treat that as MEDIUM confidence |
| D-3 | **Favorites and Recents that work offline and without a key** by snapshotting the GIF metadata locally | The extension re-fetches every view. With 100 calls per hour, that is a real budget drain and a failure mode when the key is missing or the network is down | Med | Store title, URLs, dimensions, tags, provider. Store **metadata, not media** (see D-8). This also fixes the Finer Gifs `size=10` lookup limit |
| D-4 | Search suggestions and autocomplete, trending searches, and **category tiles** on the empty state | Replaces a bare "Trending" with discoverability. Both vendors offer the endpoints [DOC] | Med | Costs API calls, so cache aggressively. Optional |
| D-5 | **Content rating setting** (GIPHY `rating`: g, pg, pg-13, r). Klipy filters are configured in its Partner Panel [DOC] | People paste GIFs into work chats. **Recommend defaulting to PG-13** | Low | [INF] The extension has no rating setting (the Raycast proxy decides). Verify the default in review |
| D-6 | **Recent searches** (history) with a quick recall | Cheap and convenient | Low | Not in Raycast. Keep it local, capped, and clearable |
| D-7 | Quick preview (Space) in a larger animated view, then act | Faster than opening the details view | Low to Med | [INF] |
| D-8 | **Bounded media cache for repeat copy**, with a size and age cap and a "Clear cache" button | Instant repeat copies, offline repeat copy of favorites | Med | **Both vendors forbid storing or mirroring media without approval [DOC].** Treat as opt-in or short-lived, and decide in a legal-risk review. Default plan: a small temp staging area only |
| D-9 | **Request-budget indicator** for the keyed providers ("62/100 this hour") | BYOK with a 100/hour cap makes rate limits a daily event. Shows why results stopped | Low to Med | [INF] Count locally. The 429 handler is the real fix (T-17) |
| D-10 | **Inline key paste and verify** in the first-run card, with a test call and Keychain storage | Shortens the worst moment of the product (section 6) | Med | |
| D-11 | Per-provider **hotkeys** or modes: one hotkey for Search, one for Favorites, one for Recents | Mirrors the extension's separate commands, where each can have its own hotkey | Low | |
| D-12 | Remember the **frontmost app** and show it ("Paste into Slack") on the paste action | Confidence that paste lands where expected | Low | |
| D-13 | Menu-bar **Recent GIFs submenu** for one-click copy | Low-friction | Low | Competitors like Grabbie and PopKey have a menu-bar-first feel |
| D-14 | Launch at login | Expected of menu-bar apps, but not essential to the first release | Low | |

---

## 5. Attribution UI (details)

Constraint source: GIPHY docs ([DOC], via summary) and Klipy docs ([DOC], read directly). **The extension does no attribution in its grid at all**, because Raycast's proxy handled the relationship. Gifsmith, using each user's own key, takes on the obligation.

| Provider | Requirement | Recommended UI |
|---|---|---|
| **GIPHY and GIPHY Clips** | "Conspicuously display 'Powered By GIPHY' attribution marks where the API is utilized" [DOC] | Show the official "Powered by GIPHY" mark in the **panel footer or search-bar accessory whenever the GIPHY or Clips provider is active**, including the empty and trending view. Use the official logo files. Also show it in the details view |
| **Klipy** | **Required:** the search input's default placeholder is "Search KLIPY". **Optional:** a watermark on shared content, and a "Powered by KLIPY" mark [DOC] | Set the placeholder to exactly `Search KLIPY` when Klipy is active. Add the footer "Powered by KLIPY" mark anyway (cheap goodwill) |
| **Finer Gifs Club** | No formal API terms found | A small footer credit with a link to thefinergifs.club. Courtesy only |
| All | Don't strip or alter returned URLs [DOC]. Don't blend providers in one grid [DOC] | One section per provider everywhere (including the Favorites and Recents views). Never a unified search |

Details view: keep the extension's "Source: {provider} with logo" row, and add GIPHY's user name when present. Use **official** logo assets. The extension's bundled PNGs are trademarked marks, so don't assume the MIT license covers them.

Confidence: HIGH that the rules exist. MEDIUM on exact GIPHY wording and size rules, because the page was read through a summarizer and the separate attribution guide wasn't read. **Re-read the current GIPHY attribution guide during the phase that builds this.**

---

## 6. First-run experience with BYO API keys

Nothing in the extension covers this (it is keyless via proxy), so this section is **[INF]** design guidance built on the vendor facts above.

### 6.1 Principles

1. **Never a dead end.** The first hotkey press must show something useful. Finer Gifs works with no key, but it is The Office quotes only and has no trending, so it is not a good first impression on its own.
2. **Keyed providers stay visible in the switcher**, with a small "key needed" badge. Hiding them hides the product's main value.
3. **Selecting a keyless-unavailable provider shows a setup card in the grid area** (not a modal and not a failed request).
4. **The key setup is inline.** The user shouldn't have to find Settings.

### 6.2 Recommended flow

1. **First launch:** a small welcome panel (once), covering the hotkey, and the **Accessibility permission only when the user first uses Paste** (not at launch).
2. **Default provider on a fresh install:** GIPHY, with the setup card showing, and a clear secondary button **"Try Finer Gifs (no key)"**. [INF] An alternative is to default to Finer Gifs and show a banner "Add a GIPHY key for more". I prefer the first option because the product's value is mainstream GIFs. Revisit with the user.
3. **Setup card contents:**
   - One line on why a key is needed ("GIPHY limits shared use. Your own free key takes about 2 minutes").
   - A button that **opens the vendor dashboard** (GIPHY developer dashboard, or partner.klipy.com for Klipy).
   - Three short numbered steps matching the vendor flow.
   - A **paste field** that **validates on paste** with a one-result test call. Show a spinner and then a green check. Save to Keychain.
   - A note on the limit: "Free keys allow 100 requests per hour. That is plenty for normal use."
4. **After success:** go straight to Trending for that provider.
5. **Failure messages must differ:**
   - **Invalid key** (401 or 403): "That key was rejected. Check you copied the whole key."
   - **Rate limited** (429): "You've used this hour's 100 requests. Resets within the hour." Show Favorites and Recents (works with D-3) and offer the other providers.
   - **Network or timeout:** the extension's existing wording.
6. **Vendor steps to confirm live during implementation (LOW confidence):**
   - GIPHY: create an app in the Developer Dashboard. Zulip's docs say to choose **SDK** as the product type for a picker [DOC], but this should be verified against the live dashboard, since the dashboard has changed over time.
   - Klipy: Partner Panel, API Keys, create a platform. The key starts in **Testing mode, 100 requests per hour** [DOC].
7. **Settings page** shows each key (masked), its status, a "Test" button, and "Remove".

### 6.3 Pitfalls specific to the first run

- Do not require **both** keys. One keyless provider and one keyed provider is enough to start.
- Do not ask for Accessibility at launch. Copy must work without it (PROJECT.md).
- Keys appear in the Klipy native URL path (`/api/v1/{app_key}/...`) [DOC]. Never log full URLs.
- Both vendors say keys are not secrets [DOC], so the cost of Keychain storage is only a small amount of native code. Do it, but don't over-invest.

### 6.4 Rate-limit behavior (applies after first run)

At 100 calls per hour with a page size of 20, a user who scrolls a lot or types without a debounce can run out within minutes.

- Debounce the search input, and cancel stale requests.
- Cache Trending per provider, with a TTL of a few minutes.
- Do not call the API when browsing Favorites or Recents (see D-3).
- Fetch the next page only when the user is near the end (see KB-4).
- Treat a 429 as a first-class state.

---

## 7. Keyboard-first grid navigation

The extension gets this from Raycast's `Grid` component for free. Gifsmith has to build it. Items marked **[Raycast]** describe Raycast's behavior from general knowledge (MEDIUM confidence).

| ID | Behavior | Notes |
|---|---|---|
| KB-1 | **The search field always has focus.** Typing anywhere edits the query. Arrow keys move the **grid selection** and never leave the field | [Raycast] works this way. Left and Right move the grid selection even when text is present. Offer ⌘← and ⌘→ or the Option variants for the text caret, or accept the loss |
| KB-2 | **Arrow keys move in 2D.** Up and Down move by one row, so the app must know the **actual column count** (8, 5 or 3, from the item size) | Compute from the rendered layout, not only from the setting. Hold-to-repeat must stay smooth with animated GIFs |
| KB-3 | **The first item is auto-selected** whenever results change for a new query. **Selection is kept** when more results are appended. Selection scrolls into view | Easy to get wrong with infinite scroll |
| KB-4 | **Load the next page based on the selected index** (for example, when the selection is within 1 or 2 rows of the end), not only on a scroll event | A keyboard user never scrolls. An IntersectionObserver alone leaves them stuck at the bottom. This is the most likely bug |
| KB-5 | ⏎ runs the default action. ⌘⏎ runs the sibling action (D-2). ⌘K opens the action panel | Esc closes the action panel, then the details view, then the app. Esc with a non-empty query first clears the query (a choice to confirm) |
| KB-6 | **Provider switching by key.** For example ⌘1 to ⌘4 for providers, and ⌘5 and ⌘6 for Favorites and Recents, or ⌘[ and ⌘]. Plus Raycast's search-bar dropdown opening with ⌘P | [Raycast] opens the search-bar dropdown with ⌘P. Gifsmith should choose its own scheme and show it in the UI |
| KB-7 | **Keep the extension's action shortcuts** (the table in section 1.3) as in-panel shortcuts | Users migrating from Raycast keep muscle memory. Check for conflicts (for example, ⌘⇧C for Copy Page Link) |
| KB-8 | A **visible selection ring** and an accessible label (the title) on each cell | VoiceOver and Reduce Motion matter. Honor the system setting to pause animation |
| KB-9 | **Animation budget:** only animate visible cells and use the small rendition in dense grids | An 8-column grid of 20+ animated GIFs is heavy. The extension uses a static-ish `preview_gif` for small |

---

## 8. Anti-Features

Deliberately not built.

| Anti-Feature | Why avoid | What to do instead |
|---|---|---|
| Tenor provider | The API shut down on 2026-06-30. No new keys [WEB] | Nothing. Revisit only if Google restores it |
| A hosted key proxy, or shared embedded keys | Out of scope in PROJECT.md. Requires servers and GIPHY production approval. The vendors also forbid routing requests through intermediaries without approval [DOC] | BYO keys with a good first run (section 6) |
| Calling Raycast's proxy (`gif-search.raycast.com`) | Not ours. Must never be called | Direct provider APIs |
| A unified cross-provider search or a mixed-provider grid | Both vendors prohibit mixing their content with other providers' in one result set [DOC]. It is also not how the extension works | One provider per grid. Favorites and Recents grouped in sections per provider |
| Re-ranking, filtering or reordering API search and trending results | Both vendors prohibit it [DOC] | Local ordering (frecency) applies to the user's own saved items only |
| GIF recording, creation, editing or uploading (Gifox and GIPHY Capture territory) | A different product. Large scope | Stay a search-and-send tool |
| A local GIF library with tags (Gnome's angle) | Scope creep. Competes with the core value | Favorites cover "my GIFs" |
| Stickers, memes and AI emoji (Klipy offers them) | Not in the extension. Dilutes the grid | Revisit later |
| AI tool or MCP server (`search-gifs`) | Out of scope in PROJECT.md | none |
| Accounts, cloud sync, telemetry, ads | No servers. Privacy. Klipy's ads API is optional [DOC] | Local-only data. Export or import of favorites is a possible later item |
| Mac App Store build | The sandbox blocks synthesized ⌘V, which Paste needs | Developer ID plus notarization |
| Windows in v1 | PROJECT.md. Native glue is per-platform | Keep the Go core portable |
| An in-app browser for GIPHY or Klipy pages | Unneeded | Open in the default browser |
| Unbounded Recents | The extension has this flaw. It grows forever | Cap it (for example, 100 per provider) with a "Clear recents" action |

---

## 9. Feature Dependencies

```
Hotkey panel (T-1) -> everything visible
Provider clients + keys (T-3, T-14) -> Search/Trending (T-3) -> Grid + paging (T-4) -> Keyboard nav (T-19, KB-4)
Keyless provider (Finer Gifs) -> First run (T-15)
Key storage (Keychain) + key validation -> Setup card (T-15, D-10) -> Rate-limit UX (T-17, D-9)

Copy GIF file (T-5) -> Action panel (T-9) -> Default action (T-10) -> Sibling shortcut (D-2)
Copy GIF file (T-5) -> Square crop (T-8)      (needs a GIF-type input. Clips need care)
Copy GIF file (T-5) -> Paste GIF (T-6)        (adds the Accessibility permission, CGEvent and focus restore)
Copy GIF file (T-5) -> Drag-out (D-1)         (same file-resolution path, plus a native drag source)
Download (T-7)  -> uses the same file resolver as T-5

Usage tracking -> Recents (T-11) -> Frecency (T-12)
Favorites (T-11) -> Metadata snapshot (D-3) -> works offline / no key / no API calls
Favorites (T-11) -> Media cache (D-8)         (optional. Needs a ToS decision first)

Provider attribution (T-16) -> Details view (T-13)   (the Source row)
Provider attribution (T-16) -> each provider's grid footer + placeholder
Categories / suggestions (D-4) -> extra API calls -> Rate-limit budget (D-9)
```

---

## 10. MVP Recommendation

**Spike first (as in PROJECT.md):** hotkey panel, copy GIF file, paste into TextEdit. This validates T-1, T-5 and T-6, the riskiest native parts.

**Prioritize (v1):**
1. T-1 to T-5, T-7, T-9, T-10, T-19: the core loop (hotkey, search, grid, keyboard navigation, copy, action panel, default action).
2. T-3, T-4, T-14, T-15, T-16, T-17: providers, keys, a first run that works, and attribution. **Bring the first-run and attribution work forward.** It is not polish. It determines whether the free public release works for strangers.
3. T-6 (paste): after copy is solid. Gate it behind the Accessibility prompt.
4. T-11, T-12, T-13: Favorites, Recents and details. Build them with **D-3 (metadata snapshot)** from the start. Retrofitting it changes the storage format.
5. T-8 (square): after copy and paste, because the Clips handling needs a decision.
6. T-18, T-20: polish and release.

**Cheap differentiators to include in v1:** D-2 (⌘⏎ sibling), D-5 (rating, default PG-13), D-6 (recent searches), D-10 (inline key verify), D-11 (separate hotkeys).

**Defer:**
- **D-1 drag-out:** the best differentiator versus competitors and the hardest. Plan it as a v1.x stretch once the file pipeline is stable.
- **D-4 categories and suggestions:** after the rate-limit budget behavior is proven.
- **D-8 media cache:** until the ToS question is decided.
- D-7, D-9, D-12, D-13, D-14: nice-to-haves.

**Open decisions for requirements:**
- Default provider on a fresh install (GIPHY with a setup card, or Finer Gifs with a banner).
- Whether Recents are frecency-ordered or strictly most-recent (the extension is inconsistent).
- Whether to keep the persistent favorites media cache (ToS risk).
- Whether to send GIPHY and Klipy analytics pingbacks (the endpoints exist [DOC]). They are not clearly required for a personal key and they send usage data. The default here is off.
- Square crop for Clips (see PITFALLS).

---

## Sources

- Extension source (read in full): `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/rc/extensions/gif-search` (`package.json`, `README.md`, `CHANGELOG.md`, `src/**`, `tests/**`). HIGH.
- Raycast frecency implementation: https://raw.githubusercontent.com/raycast/utils/main/src/useFrecencySorting.ts (HIGH, read directly).
- GIPHY API docs (attribution, beta limits, URL, ordering, mixing and caching rules): https://developers.giphy.com/docs/api and https://developers.giphy.com/docs/api/endpoint/ (MEDIUM, read through a summarizing fetch. The beta-key support page returned 403).
- Klipy docs, read as raw markdown: https://docs.klipy.com/attribution.md, https://docs.klipy.com/integration-requirements.md, https://docs.klipy.com/getting-started.md, https://docs.klipy.com/gifs-api/gifs-share-trigger-api.md, https://docs.klipy.com/llms.txt (HIGH).
- Zulip GIF picker integration docs (key setup for GIPHY, Klipy, Tenor): https://zulip.readthedocs.io/en/12.1/_sources/production/gif-picker-integrations.md.txt (MEDIUM).
- Tenor shutdown: https://9to5google.com/2026/06/30/google-tenor-api-gif-updates/, https://www.shacknews.com/article/149852/google-tenor-api-service-discontinued-twitter-discord-whatsapp-bluesky (MEDIUM, several outlets agree).
- Competitors: Quickgif https://apps.apple.com/app/id6744745027, Gnome https://macmagazine.com.br/post/2026/05/11/gnome-e-um-buscador-de-gifs-que-fica-na-barra-de-menus-do-macos/, https://maketecheasier.com/search-send-gifs-mac, Setapp and AlternativeTo listings for Gifox, Grabbie, GIFme, GIF Bar and Hard G (MEDIUM, store pages and reviews).
- Slack and Discord picker features and Raycast's default ⌘⏎ and ⌘P behavior: general knowledge, not verified this session (LOW to MEDIUM).
- Note on the confidence seam: `classify-confidence --provider webfetch` returns LOW by default. I upgraded the Klipy findings to HIGH because I read the vendor's own raw docs, and kept GIPHY at MEDIUM because it came through a summarizer.
