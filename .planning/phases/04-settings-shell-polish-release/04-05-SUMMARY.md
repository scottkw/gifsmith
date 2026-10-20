---
phase: 04-settings-shell-polish-release
plan: 05
subsystem: release
tags: [codesign, dmg, notarytool, wails-task, swift-appkit, svelte, about, readme]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "services.Version and update.DMGAssetName (04-07); Klipy file-action gate (04-01/04-04); General pane, theme lib and GIPHY marks (04-01); smoke/UAT loop (04-04)"
provides:
  - "wails3 task darwin:release:build: Developer ID-signed (hardened runtime) bin/release/Gifsmith.app + Gifsmith-1.0.0-arm64.dmg (+ .sha256), verified"
  - "wails3 task darwin:release: the same plus notarize + staple of app and DMG, gated on the gifsmith-notary profile, verified with --notarized"
  - "scripts/verify-release.sh [--notarized] {dir}: RELEASE VERIFY OK / RELEASE VERIFY FAIL: {check}"
  - "scripts/dmg-background.swift + build/darwin/dmg-background.png (540x380) / @2x (1080x760)"
  - "Info.plist / Info.dev.plist / config.yml at 1.0.0; LICENSE in Contents/Resources and in the DMG"
  - "Snapshot.version; Settings → About pane; ABOUT_URLS in the openExternal allow-list"
  - "version_test.go: Info.plist == services.Version, Taskfile DMG_NAME == update.DMGAssetName, A2 names in LICENSE and About"
  - "README rewrite (install, verify, keys, Settings, Updating, attribution, known limits); docs/release-notes-1.0.0.md"
affects: [04-06]

actuals:
  tokens: 14600
  tasks: 4
  commits: 4
plan_head_before: c31c9205abb7beeaba65871c53c5bf858520e4b3
plan_head_after: d073df8bd63c4048a3b18498316fa50ae65ac918

tech-stack:
  added: []
  patterns:
    - "One internal release:assemble task with NOTARIZE=true|false shared by release:build and release"
    - "DMG background committed as @1x/@2x PNGs, combined into a two-representation TIFF at build time (tiffutil -cathidpicheck)"
    - "verify-release.sh mounts the DMG read-only with -nobrowse -noautoopen (never shown in Finder) and always detaches via trap"

key-files:
  created:
    - scripts/verify-release.sh
    - scripts/dmg-background.swift
    - build/darwin/dmg-background@2x.png
    - internal/services/version_test.go
    - frontend/src/settings/AboutPane.svelte
    - frontend/public/appicon.png
    - scripts/uat/19-about.js
    - docs/release-notes-1.0.0.md
  modified:
    - build/darwin/Taskfile.yml
    - build/darwin/Info.plist
    - build/darwin/Info.dev.plist
    - build/config.yml
    - build/darwin/dmg-background.png
    - internal/services/settings.go
    - frontend/bindings/gifsmith/internal/services/models.ts
    - frontend/src/lib/providers.ts
    - frontend/src/lib/api.ts
    - frontend/src/settings/SettingsApp.svelte
    - frontend/index.html
    - README.md
    - .planning/phases/04-settings-shell-polish-release/04-UI-SPEC.md

key-decisions:
  - "The DMG uses a two-representation TIFF built from the committed @1x/@2x PNGs at release time (the leaanthony/dmg library copies any image file as .background.<ext>), so Retina Finder windows get the sharp background"
  - "Notarized mode signs once with --hardened-runtime --notarize (wails3 tool sign re-signs before notarizing anyway), instead of a plain sign followed by a second notarizing sign"
  - "All six unreferenced Wails template public assets (wails.png, svelte.svg, Inter-Medium.ttf, style.css, bg-desktop.jpg, bg-mobile.jpg) were deleted; about 700 KB no longer embedded in the binary"
  - "The gifsmith-notary profile works (notarytool history exit 0, 2026-10-19); the notarized release run itself stays in 04-06"

patterns-established:
  - "Release metadata (VERSION, DMG_NAME) lives in build/darwin/Taskfile.yml and a Go test pins it to services.Version and the updater's asset name"

requirements-completed: [RELS-03, RELS-04]

coverage:
  - id: D1
    description: "darwin:release:build builds, Developer ID-signs (hardened runtime, no entitlements) and verifies Gifsmith.app and Gifsmith-1.0.0-arm64.dmg (volume Gifsmith, LICENSE, Applications link, .background, TeamIdentifier=S2K7P43927) plus its SHA-256; designated requirement equals the ~/Applications dev install's"
    requirement: RELS-01
    verification:
      - kind: integration
        ref: "wails3 task darwin:release:build (bin/release-build.log: RELEASE VERIFY OK), run twice (task verify + tracer gate)"
        status: pass
    human_judgment: false
  - id: D2
    description: "darwin:release notarizes and staples app and DMG and requires spctl 'source=Notarized Developer ID' and stapler validate; gated on the gifsmith-notary profile"
    requirement: RELS-01
    verification:
      - kind: other
        ref: "wails3 task -dry darwin:release (step order); xcrun notarytool history --keychain-profile gifsmith-notary (exit 0)"
        status: pass
    human_judgment: true
    rationale: "The notarized run is plan 04-06's; only a real Apple submission proves notarization passes"
  - id: D3
    description: "DMG window look: new background (no Wails logo) under Wails' fixed icon layout, sharp on Retina"
    verification: []
    human_judgment: true
    rationale: "Opening the DMG shows a Finder window on the user's desktop, which executors never do; visual check belongs to the end-of-phase UAT"
  - id: D4
    description: "Settings → About: icon, Gifsmith, Version 1.0.0 from services.Version, A1-A4, links through the fixed allow-list, theme-resolved GIPHY mark, scrolls at 640x520"
    requirement: RELS-03
    verification:
      - kind: unit
        ref: "internal/services/version_test.go#TestVersionMatchesInfoPlist,TestVersionDMGAssetName,TestVersionInSnapshot,TestLicenseCredits"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/19-about.js"
        status: pass
    human_judgment: false
  - id: D5
    description: "README (install, verify, keys, Settings, Updating, attribution, known limits incl. Klipy file actions disabled) and docs/release-notes-1.0.0.md"
    requirement: RELS-04
    verification:
      - kind: other
        ref: "Task 3 <verify> greps (Powered By GIPHY, Search KLIPY, Joseph Schmitt, shasum -a 256 -c, api.github.com, Check for updates automatically, macOS 13)"
        status: pass
    human_judgment: false

duration: 12min
completed: 2026-10-20
status: complete
---

# Phase 4 Plan 05: Release Pipeline, About and README Summary

**One command (`wails3 task darwin:release:build`) now produces a Developer ID-signed, hardened-runtime arm64 `Gifsmith.app` and `Gifsmith-1.0.0-arm64.dmg` (new background, LICENSE, checksum), checked by `scripts/verify-release.sh`. `darwin:release` adds notarization and stapling and requires the gifsmith-notary profile, which works. Settings → About credits the Raycast extension's authors, and the README and 1.0.0 release notes cover install, keys, updates, attribution and known limits.**

## Performance

- **Duration:** about 12 min
- **Started:** 2026-10-20T02:48:51Z
- **Completed:** 2026-10-20T03:00:27Z
- **Tasks:** 4 of 4 (Task 4 satisfied by evidence)
- **Files modified:** 27 (664 insertions, 386 deletions, including 6 deleted template assets)

## Accomplishments

- `darwin:release:build` packages a production arm64 build, copies it to `bin/release/Gifsmith.app`, Developer ID-signs it with the hardened runtime, builds the DMG (`wails3 tool package`, volume `Gifsmith`, with LICENSE), renames it to `Gifsmith-1.0.0-arm64.dmg`, codesigns it with a timestamp, writes the `.sha256` from inside `bin/release`, and runs `verify-release.sh`. It is guarded so it only ever removes a directory ending in `/release`.
- `darwin:release` runs the same steps plus notarize + staple for the app (`wails3 tool sign --notarize`) and the DMG (`notarytool submit --wait`, `stapler staple`). Its precondition is a working `gifsmith-notary` profile, and it verifies with `--notarized`.
- `verify-release.sh` checks: strict deep codesign, the Developer ID authority, the runtime flag, no entitlements, version 1.0.0, LSUIElement, LICENSE in Resources, an arm64-only executable, the DMG signature and `TeamIdentifier=S2K7P43927`. It also mounts the DMG read-only to check `Gifsmith.app`, the `Applications` → `/Applications` symlink, LICENSE and `.background`, then checks the checksum. With `--notarized` it adds spctl (exec: `source=Notarized Developer ID`; DMG open) and stapler validate for both. It confirms the designated requirement matches `~/Applications/Gifsmith.app`, so Accessibility grants carry over.
- The new DMG background follows the UI-SPEC: flat #F2F2F7, a 2px #86868B arrow between (151,164) and (388,164), and "Drag Gifsmith to Applications" 48px from the bottom. It is drawn by `scripts/dmg-background.swift` at @1x and @2x.
- Settings → About: icon, name, `Version 1.0.0` (from `Snapshot.version` = `services.Version`), A1 to A4, Website / Report an Issue / View License, the Credits section with the GIF Search extension link, and the GIPHY mark for the resolved theme. The four URLs are the only additions to `openExternal`'s allow-list.
- README rewritten and `docs/release-notes-1.0.0.md` added; the favicon now uses `/appicon.png`, and the UI-SPEC DMG row is amended per D-22.

## Task Commits

1. **Task 1 (tracer): signed arm64 DMG pipeline**: `d397eee` (feat). Tracer gate: the full `<verify>` was re-run after the commit and passed again (`RELEASE VERIFY OK`, `shasum -c` OK, go.sum clean, bundle 1.0.0).
2. **Task 2: About pane and the version test**: `51d0388` (test, RED), `917a3a9` (feat, GREEN)
3. **Task 3: README, release notes, favicon, UI-SPEC amendment**: `d073df8` (docs)
4. **Task 4: notary profile checkpoint**: no commit. Satisfied by evidence (below).

### Release build evidence (Task 1)

```
$ ls -l bin/release
-rw-r--r--@ 1 ken  staff  6547470 Oct 19 21:53 Gifsmith-1.0.0-arm64.dmg
-rw-r--r--  1 ken  staff       91 Oct 19 21:53 Gifsmith-1.0.0-arm64.dmg.sha256
drwxr-xr-x  3 ken  staff       96 Oct 19 21:52 Gifsmith.app
```

The DMG is 6.5 MB (6,547,470 bytes; `du` reports 9.0M because of the custom file icon's resource fork). The tail of verify-release.sh in `bin/release-build.log`:

```
release: designated => identifier "com.kenscott.gifsmith" and anchor apple generic and certificate 1[field.1.2.840.113635.100.6.2.6] /* exists */ and certificate leaf[field.1.2.840.113635.100.6.1.13] /* exists */ and certificate leaf[subject.OU] = S2K7P43927
dev:     designated => identifier "com.kenscott.gifsmith" and anchor apple generic and ... leaf[subject.OU] = S2K7P43927
RELEASE VERIFY OK
```

`hdiutil` prints a deprecation notice for `attach -readonly -nobrowse -mountpoint` (it suggests `diskutil image attach`). It is informational; the attach works.

### TDD notes (Task 2)

- RED (`51d0388`): `TestVersionInSnapshot` failed on its assertion (`snapshot version = <nil>, want "1.0.0"`). `TestLicenseCredits` failed on its About assertions: AboutPane.svelte did not exist, so every A2 name was reported missing from About; LICENSE already had all of them. `TestVersionMatchesInfoPlist` and `TestVersionDMGAssetName` passed at RED because Task 1 had already set the plist to 1.0.0 and added DMG_NAME. They stand as drift guards. Semantic assessment: the two target tests failed for the intended reasons. There were no compile, load or fixture failures. The evidence is `go test -v` text, a format the `tdd-red-evidence` classifier does not parse; 04-07 handled Go RED the same way.
- GREEN (`917a3a9`): all four pass.

### Task 4: notary profile checkpoint (gate="blocking-human")

The user created the profile on 2026-10-19 (`xcrun notarytool store-credentials gifsmith-notary …` → "Credentials saved to Keychain"). The executor ran `xcrun notarytool history --keychain-profile gifsmith-notary >/dev/null 2>&1` and it exited **0**, so the profile works. The profile's contents were never printed. As the plan directs, no notarization was submitted here; the notarized `wails3 task darwin:release` run belongs to 04-06. RELS-01 stays open until spctl reports `source=Notarized Developer ID`.

## Verification (plan level)

- `wails3 task darwin:release:build` ended with `RELEASE VERIFY OK` (twice), and `git diff --quiet -- go.sum` passed after `git checkout -- go.sum`.
- `gofmt -l` reported nothing and `go vet . ./internal/...` passed. `go test -race ./internal/services -run 'Version|License' -v` passed 4 of 4, and `go test -race ./internal/... -count=1` passed in every package.
- `CGO_ENABLED=0 go build -tags server .` and `pnpm --dir frontend run build` passed. `pnpm run test` reported 101 pass and 0 fail; svelte-check reported 0 errors and 0 warnings.
- `--uat scripts/uat/19-about.js` printed UAT OK. As regression checks after the template assets were deleted, `15-settings-general.js` and `01-tracer.js` also printed UAT OK.
- `wails3 task -dry darwin:release` shows the notarized step order: sign+notarize+staple app → DMG → codesign → notarytool submit → stapler staple → checksum → `verify-release.sh --notarized`.

## Template assets (Task 3, Chesterton)

`grep -rn` across the repo (excluding node_modules, dist, bin and .planning) found these references:

| Asset | Referenced by | Action |
|---|---|---|
| frontend/public/wails.png | only the index.html favicon, now `/appicon.png` | deleted |
| frontend/public/svelte.svg | nothing | deleted |
| frontend/public/Inter-Medium.ttf | nothing | deleted |
| frontend/public/style.css | nothing loads it; build/Taskfile.yml's optional template Puppertino task only uses its `<link>` as an awk anchor, and that link is not in index.html | deleted |
| frontend/public/bg-desktop.jpg, bg-mobile.jpg | only the deleted style.css | deleted |
| frontend/public/attribution/* | Footer, About, README | kept |

Info.plist's `This is a comment` is now `Gifsmith 1.0.0`. In config.yml, `Some Product Comments` is now the A1 tagline and `0.1.0` is now `1.0.0`. `common:update:build-assets` was never run.

## Decisions Made

See `key-decisions` in the frontmatter. None of them changes the UI contract beyond the planned D-22 amendment.

## Deviations from Plan

**1. [Rule 1 - Simplification of a planned step] Single sign step in notarized mode**
- **Found during:** Task 1
- **Issue:** The plan signed with `--hardened-runtime` and then signed again with `--notarize`. `wails3 tool sign --notarize` already runs codesign with `--options runtime` before it submits (sign.go), so the first signature would be overwritten at once.
- **Fix:** One templated line, `wails3 tool sign … --hardened-runtime{{if NOTARIZE}} --notarize --keychain-profile "gifsmith-notary"{{end}}`. The resulting signature is the same.
- **Files modified:** build/darwin/Taskfile.yml
- **Commit:** d397eee

**2. [Plan option taken] TIFF background**
- The plan said to pass a two-representation TIFF if the DMG library accepts one. leaanthony/dmg copies any background file as `.background.<ext>` (hdiutil_darwin.go `copyBackground`), so the release task builds `bin/dmg-background.tiff` with `tiffutil -cathidpicheck` and passes that. Only the PNGs are committed.

**Total deviations:** 1 simplification, plus 1 plan option chosen. **Impact:** none on outputs or scope.

## Issues Encountered

None.

## Known Stubs

None.

## Threat Flags

None. The only new surface is the four About URLs, and they go through the fixed allow-list (T-04-23). verify-release.sh enforces T-04-21 and T-04-22 (signature, timestamped DMG signature, checksum, no entitlements). The Taskfile names only the profile `gifsmith-notary` (T-04-20).

## Human-check items (deferred to the end of the phase)

1. Open `bin/release/Gifsmith-1.0.0-arm64.dmg` in Finder. Expect a 540×380 window with the new background (no Wails logo), the arrow running between the Gifsmith icon on the left and the Applications alias on the right, and the text sharp on a Retina display. Eject it afterwards.
2. On the signed install, check Settings → About in Light and Dark. Each About link should open the right page in the browser, and the GIPHY mark should switch with the theme.
3. 04-06: run `wails3 task darwin:release` (notarized). Expect `RELEASE VERIFY OK (notarized)`, then continue with 04-06's install and update checks.

## Next Phase Readiness

- Ready for 04-06: the notarized pipeline is wired and the notary profile works.
- RELS-01 is open until 04-06's notarized run passes.

---
*Phase: 04-settings-shell-polish-release*
*Completed: 2026-10-20*

## Self-Check: PASSED
