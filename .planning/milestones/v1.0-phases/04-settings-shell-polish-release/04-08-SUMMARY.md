---
phase: 04-settings-shell-polish-release
plan: 08
subsystem: infra
tags: [github-actions, ci, notarytool, stapler, codesign, release, gatekeeper]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "darwin:release task and scripts/verify-release.sh (04-05); TestRealDMG/TestReleaseJSON (04-07); notarized 1.0.0 draft (04-06)"
provides:
  - ".github/workflows/ci.yml: test job on push/PR to main, macos-15, read-only, no credentials"
  - ".github/workflows/release.yml: vX.Y.Z tag (never v1.0.0) builds, signs, notarizes, staples, verifies and drafts a release; workflow_dispatch dry run"
  - "scripts/release-version.sh: tag/version-source/notes gate"
  - "scripts/notarize.sh + scripts/stapler-retry.sh: Accepted-checked notarization and retried stapler lookups (local and CI)"
  - "docs/releasing.md: maintainer procedure for 1.0.1+"
  - "release environment with five secrets, deploying only from main and v* tags"
affects: [phase-04-verification, release, 1.0.1]

actuals:
  tokens: 5052
  tasks: 3
  commits: 4
plan_head_before: 7f0ad38cd51e458b23d71572a79a0aaf394411b5
plan_head_after: 438ed772091fd54e45a7a3c50a941e5711ec569c

tech-stack:
  added: []
  patterns:
    - "CI release reuses the local `wails3 task darwin:release` by recreating the same keychain identity and notary profile name on the runner"
    - "Every `xcrun stapler` call goes through scripts/stapler-retry.sh (5 tries, 20 s apart)"

key-files:
  created:
    - .github/workflows/ci.yml
    - .github/workflows/release.yml
    - scripts/release-version.sh
    - scripts/notarize.sh
    - scripts/stapler-retry.sh
    - docs/releasing.md
  modified:
    - go.mod
    - build/darwin/Taskfile.yml
    - scripts/verify-release.sh

key-decisions:
  - "go.mod pins `toolchain go1.27.2` (go directive stays 1.25.0) so setup-go v7 builds CI with the Go that built 1.0.0"
  - "pnpm pinned to 12.6.0 (the actual local version) instead of the plan's 12.14.0, so CI matches the version that installed the lockfile locally"
  - "App notarization moved out of wails3's --notarize into scripts/notarize.sh: notarytool --output-format json, require status Accepted, staple with retries"
  - "All stapler staple/validate calls retry 5x/20 s via scripts/stapler-retry.sh after GitHub's macOS runners dropped 3 of 7 ticket lookups (NSURLError -1005, error 68)"

patterns-established:
  - "Dry-run evidence: bin/ci-dry-run.log (passing run), bin/ci-dry-run-attempt{1,2,3}.log (the stapler failures)"

requirements-completed: [RELS-07]

coverage:
  - id: D1
    description: "CI test job on push/PR to main: frozen pnpm install, frontend build, gofmt, go vet, go test -race, golangci-lint, server build, node tests, svelte-check, smoke"
    requirement: RELS-07
    verification:
      - kind: integration
        ref: "gh run 39446862985 (main f6e3e40) success; also 39438996425 (b91dd66) and 39445350392 (6f4cc2c)"
        status: pass
      - kind: integration
        ref: "Task 1 verify chain locally (VERIFY_ALL_OK)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Release workflow dry run: keychain + notary profile, darwin:release, both notarizations Accepted, staples, RELEASE VERIFY OK (notarized), TestRealDMG, no release created"
    requirement: RELS-07
    verification:
      - kind: integration
        ref: "gh run 39446861711 (workflow_dispatch, main f6e3e40) success; bin/ci-dry-run.log"
        status: pass
    human_judgment: false
  - id: D3
    description: "Version gate refuses mismatched tags, bare versions, extra parts, mismatched sources and missing notes"
    requirement: RELS-07
    verification:
      - kind: unit
        ref: "bash scripts/release-version.sh cases in Task 1 verify plus scratch-copy negative tests"
        status: pass
    human_judgment: false
  - id: D4
    description: "Tag push creates a draft with exactly the DMG and .sha256 and TestReleaseJSON passes on it"
    requirement: RELS-07
    verification:
      - kind: other
        ref: "jq expressions checked read-only against the v1.0.0 draft; full path first runs on the v1.0.1 tag"
        status: unknown
    human_judgment: true
    rationale: "The draft step only runs on a real vX.Y.Z tag push; no tag may be pushed before 1.0.1 is ready"

duration: 91min
completed: 2026-10-20
status: complete
---

# Phase 4 Plan 08: GitHub Actions CI and tag-triggered draft releases Summary

**CI runs the full suite on macos-15 for every push/PR to main, and a vX.Y.Z tag runs the unchanged `wails3 task darwin:release` on a runner with a temporary Developer ID keychain and notary API key, then drafts the release; proven by a green workflow_dispatch dry run (both notarizations Accepted, RELEASE VERIFY OK (notarized), TestRealDMG PASS) after adding stapler retries.**

## Performance

- **Duration:** about 91 min, including four dry runs and the user's Task 2
- **Started:** about 2026-10-20T13:50Z
- **Completed:** 2026-10-20T15:21:13Z
- **Tasks:** 3 (Task 2 by the user)
- **Files modified:** 9

## Accomplishments

- **CI on main is green.** ci.yml runs on push and PR to main with read-only permissions and no credentials. Run 39446862985 on f6e3e40 passed: Go 1.27.2 via the toolchain line, pnpm 12.6.0, `0 issues.`, `SMOKE OK`.
- **The release dry run is green.** https://github.com/scottkw/gifsmith/actions/runs/39446861711 (main f6e3e40, 3m48s):
  - App notarization `ba1290bc-5768-4bcc-8d13-b2132b7d57a9` and DMG notarization `872d7e5c-9c7e-4d14-bc3c-652d312fad14` were both `"status":"Accepted"`.
  - Both staples and both validates worked on the first try; no retry lines appeared.
  - `source=Notarized Developer ID` was printed for the app and the DMG, followed by `RELEASE VERIFY OK (notarized)`.
  - `--- PASS: TestRealDMG`, then `dry run: no release created`.
  - The log contains no `PRIVATE KEY` line, and secrets show only as `***`.
- **Releases are unchanged.** The only release is the `v1.0.0` draft (isDraft true), and the remote has no tags.
- **Task 2 (user):** the `release` environment exists with deployment policies `main,v*` and the five secret names `MACOS_CERT_P12_BASE64,MACOS_CERT_PASSWORD,NOTARY_ISSUER_ID,NOTARY_KEY_ID,NOTARY_KEY_P8_BASE64`. The orchestrator checked both. No required reviewer was added.
- **Two probe-flagged assumptions held on the runner:**
  - notarytool finds the `gifsmith-notary` profile without a keychain flag once the temporary keychain is the default and the only one on the search list.
  - The runner trusts the Developer ID G2 intermediate, and Gatekeeper assessments are enabled.

  The plan's NOTARY_KEYCHAIN and G2-import fallbacks were not needed.

## Task Commits

1. **Task 1 pre-step (orchestrator decision): Go toolchain pin.** `bc6e453` (chore)
2. **Task 1: CI workflow, release workflow, version gate, releasing doc.** `4324ef1` (feat)
3. **Task 2: release environment and secrets.** Done by the user; no commit.
4. **Task 3 fix 1: notarize via scripts/notarize.sh with an Accepted check and staple retries.** `3d960ff` (fix)
5. **Task 3 fix 2: retry every stapler lookup through scripts/stapler-retry.sh.** `438ed77` (fix)

The orchestrator merged these to main (b91dd66, 6f4cc2c, f6e3e40) and pushed with the user's approval.

## Files Created/Modified

- `.github/workflows/ci.yml`: test job on macos-15. Actions pinned by SHA, no credentials. Mirrors the local suite and smoke run.
- `.github/workflows/release.yml`:
  - triggers on tags `v[0-9]+.[0-9]+.[0-9]+` excluding `v1.0.0`, and on workflow_dispatch;
  - runs in environment `release`;
  - order: gate, Gatekeeper, temporary keychain and notary profile, `wails3 task darwin:release`, TestRealDMG, then the draft plus TestReleaseJSON (on tag push only) or the dry-run line (on dispatch);
  - cleanup always runs.
- `scripts/release-version.sh`: the tag must equal `v` plus services.Version; services.Version must match the Taskfile VERSION, the verify-release.sh VERSION and both Info.plist keys; the release notes must exist.
- `scripts/notarize.sh`: `notarytool submit --wait --output-format json`, requires `Accepted`, then staples through the helper.
- `scripts/stapler-retry.sh`: `xcrun stapler {staple|validate}`, 5 tries, 20 s apart.
- `build/darwin/Taskfile.yml`:
  - in `release:assemble`, wails3 now only signs (identity and hardened runtime, unchanged);
  - the app zip and the DMG are notarized with `notarize.sh`;
  - `release:build` behaves as before.
- `scripts/verify-release.sh`: both `stapler validate` calls go through the helper; every other check and output line is unchanged.
- `docs/releasing.md`: version bump list, tag, draft, publish, dry run, the five secret names and how to rotate them.
- `go.mod`: `toolchain go1.27.2`.

## Decisions Made

See key-decisions in the frontmatter. In short: go.mod carries a toolchain pin for parity, pnpm matches the local 12.6.0, and notarization and stapling moved into two small scripts so they can check the result and retry.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] pnpm version**
- **Found during:** Task 1.
- **Issue:** D-40 called 12.14.0 "the local version", but `pnpm -v` reports 12.6.0.
- **Fix:** both workflows pin `pnpm@12.6.0`.
- **Committed in:** 4324ef1.

**2. [Rule 1 - Bug] Relative DMG path in go test env**
- **Found during:** Task 1.
- **Issue:** `go test` runs in `internal/update`, so a relative `UPDATE_REAL_DMG` path would not resolve.
- **Fix:** the workflow uses `$PWD/bin/release/...`.
- **Committed in:** 4324ef1.

**3. [Orchestrator decision] go.mod `toolchain go1.27.2`**
- go.mod is an expected changed file, committed on its own before the parity checks.
- **Committed in:** bc6e453.

### User-approved fixes during Task 3

**4. Stapling flake on GitHub's macOS runners.** The first three dry runs failed on stapler lookups to Apple's ticket service (`https://api.apple-cloudkit.com/.../com.apple.gk.ticket-delivery/.../records/lookup`): NSURLError -1005 "The network connection was lost.", stapler error 68, about 7 s into each lookup. 3 of 7 lookups failed.
- **Attempt 1** (run 39440751169, attempt 1, b91dd66): the app was notarized and Accepted (`936551e9-…`), then `stapler staple` on the app failed inside wails3's `--notarize`.
- **Attempt 2** (same run, attempt 2): the app was Accepted (`f28296a3-…`) and stapled; the DMG was Accepted (`cffd2900-…`), then `stapler staple` on the DMG failed.
  - **Fix 1** (3d960ff, user "add retries"): `scripts/notarize.sh` replaces wails3's `--notarize`. It checks for Accepted and retries the staple.
- **Attempt 3** (run 39445349341, 6f4cc2c): both artifacts were Accepted (`cc61477a-…`, `13ce4fd9-…`) and both stapled on the first try. Then `stapler validate` on the DMG in verify-release.sh failed.
  - **Fix 2** (438ed77, user "fix, push and run"): `scripts/stapler-retry.sh` is shared by `notarize.sh` and both validate calls in verify-release.sh. It covers every `stapler` call in the repo; no other call sites exist.
- **Attempt 4** (run 39446861711, f6e3e40): green, with no retries needed.
- The logs are in bin/ci-dry-run-attempt{1,2,3}.log and bin/ci-dry-run.log (gitignored).

**Total deviations:** 2 auto-fixed (Rule 1), 1 orchestrator decision, 2 user-approved fixes.
**Impact on plan:** The plan wanted the local release path left unchanged. It now notarizes through `notarize.sh`, which does the same ditto, submit, staple sequence plus the Accepted check and retries. The local `wails3 task darwin:release:build` still passes, and the shipped 1.0.0 still verifies with `--notarized`; its DMG SHA-256 is unchanged at a3129b96…395d4.

## Issues Encountered

- The stapler flakes described above.
- `shellcheck` warns SC2010 at `scripts/verify-release.sh:71` (`ls | grep`). The line predates this plan and was left alone.

## Deferred Items

- The draft-release step (`gh release create`, the draft-view assertion, TestReleaseJSON on the draft) runs only on a real tag push. It first runs for v1.0.1. The jq expressions were checked read-only against the v1.0.0 draft.
- Optional: add a required reviewer to the `release` environment (Task 2 step 6, not done).
- If the runner's stapler failures ever outlast the 5×20 s retries, the fallback is to skip CI-side `stapler validate` and validate locally before publishing.

## User Setup Required

None remaining. The user created the API key, the .p12 and the `release` environment secrets in Task 2.

## Next Phase Readiness

1.0.1 can be released by following docs/releasing.md: bump the version, add notes, merge with CI green, push the tag, then check and publish the draft. The v1.0.0 draft is untouched and still unpublished.

## Self-Check: PASSED

- Files exist: .github/workflows/ci.yml, .github/workflows/release.yml, scripts/release-version.sh, scripts/notarize.sh, scripts/stapler-retry.sh, docs/releasing.md.
- Commits are ancestors of HEAD: bc6e453, 4324ef1, 3d960ff, 438ed77.
- Remote state: CI 39446862985 success; dry run 39446861711 success; releases `v1.0.0:true` only; no tags.
