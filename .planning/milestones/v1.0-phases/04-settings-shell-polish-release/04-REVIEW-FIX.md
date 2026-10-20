---
phase: 04-settings-shell-polish-release
fixed_at: 2026-10-20T17:20:00Z
review_path: .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
iteration: 3
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 4: Code Review Fix Report (iteration 3, final)

**Fixed at:** 2026-10-20T17:20:00Z
**Source review:** .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
**Iteration:** 3 (the iteration-2 report is kept in `04-REVIEW-FIX.iter2.md`)

**Summary:**
- Findings in scope: 3. These are WR-01 and WR-02, plus IN-12, which the orchestrator added. The other Info items were out of scope.
- Fixed: 3
- Skipped: 0

`workflow.use_worktrees` is `false`. Every edit and commit was made in the main checkout on
`gsd/phase-04-settings-shell-polish-release`, and every gate below ran there.

## Fixed Issues

### WR-01: The "already has a release" gate can no longer see a draft release

**Files modified:** `.github/workflows/release.yml`, `docs/releasing.md`
**Commit:** f196276
**Applied fix:**
- **New first step in `sign`.** The step "No existing release (drafts included)" runs before checkout and before the keychain step, and only when `github.event_name == 'push'`, so dispatch dry runs skip it. It runs `gh api "repos/$REPO/releases" --paginate --jq '.[] | select(.tag_name == env.TAG) | .id'` with the job's `contents: write` token, which can see drafts. It fails if that returns any id.
  - It fails closed: the default `bash -eo pipefail` fails the step if `gh api` errors.
  - It also covers "Re-run failed jobs", which skips `build`.
- **`build` gate.** It keeps its `gh release view` check as an early stop for published releases. It is renamed "no published release", and a comment explains that a read-only token never sees drafts.
- **`docs/releasing.md`.** Updated to match. Before re-running, delete the old draft.

I checked the jq query against the real repo with a read-only GET. `v1.0.0`, which is currently a draft, returned its id, and `v9.9.9` returned nothing. `actionlint` v1.7.12 is clean.
**Status:** fixed: requires human verification. The workflow was not run on GitHub Actions. The next tag run will exercise it.

### WR-02: `release:app` refuses its own second run, because `go.sum` is not tidy

**Files modified:** `go.sum`, `.github/workflows/ci.yml`
**Commit:** 80f1c17
**Applied fix:**
- **`go.sum`.** I ran `go mod tidy`. The diff is exactly the two `github.com/stretchr/objx v0.5.2` lines, `h1:` and `/go.mod`. `go.mod` is byte-identical to before (checked with `cmp`), and `go mod tidy -diff` now exits 0.
- **CI.** `ci.yml`'s test job has a new `go mod tidy` step, `go mod tidy -diff`, placed before `go vet`.

**Local release path proof:**
- **`bin/release` protected.** I moved `bin/release` aside first and restored it afterwards. The DMG's SHA-256 is still `a3129b96641fc156859af9ad96bc8f76e6cc4428abe056aa8f23f355089395d4`, and its xattrs are intact.
- **Two builds in a row.** I ran `wails3 task darwin:release:app` twice. Both runs exited 0, and each ran `go mod tidy`. Neither hit "Tracked files differ from HEAD".
- **Clean tree.** `git status` showed no tracked changes after either run, and `release-stamp.sh check bin/gifsmith.app` passes.
- **`darwin:release` / `darwin:release:sign`.** Not run.

**The old "`git checkout -- go.sum` after wails3 tasks" rule is no longer needed.** After both builds and the bindings regeneration, `git status` is clean apart from the existing untracked `.planning` files and `.ignore`.
**Status:** fixed

### IN-12: The panel's `setLastProvider` rejection is unhandled while `settings.json` is unreadable

**Files modified:** `frontend/src/App.svelte`
**Commit:** 9408093
**Applied fix:** `setLastProvider(p).catch(() => {})` now has a one-line comment saying that Settings › General reports `prefs_unreadable`. This is the same fire-and-forget pattern as `openSettings()` and `revealUpdates()` in the same file.
**Status:** fixed

## Verification (all run in the main checkout, after the three commits)

- `gofmt -l main.go wire_desktop.go wire_server.go internal`: clean.
- `go mod tidy -diff`: clean.
- `go vet . ./internal/...`: OK.
- `go test -race -count=1 ./internal/...`: every package passes. The only other output is the usual `ld` macOS-version warnings.
- `golangci-lint run . ./internal/...`: 0 issues.
- `CGO_ENABLED=0 go build -tags server -o bin/gifsmith-server .`: OK.
- `node --test --test-reporter=tap src/lib/*.test.ts`: 103/103.
- `svelte-check --tsconfig ./tsconfig.json`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
- Full UAT set, `--uat` for all 35 `scripts/uat/*.js`: 35/35 pass.
- `actionlint` v1.7.12 on `.github/workflows/*.yml`: clean.
- Bindings regenerated with the Taskfile's command (`wails3 generate bindings -f '' -clean=true -ts -i`): no diff. `go.sum` did not need restoring.
- `wails3 task darwin:release:app`, run twice: both pass, and the tree stays clean (see WR-02).

Not verified:
- The new `sign` step on GitHub Actions.
- Native WKWebView behaviour. No desktop input was allowed.

Residual limits:
- **Draft cleanup is manual.** The draft check refuses, but it does not delete or replace an existing draft. You have to remove an old draft yourself before re-running.
- **Untracked files.** `release:app`'s clean-tree check still covers tracked files only. This is unchanged and marked with a `shortcut:` comment.

---

_Fixed: 2026-10-20T17:20:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 3_
