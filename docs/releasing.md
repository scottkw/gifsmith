# Releasing Gifsmith

1.0.0 shipped from a local, notarized build (`wails3 task darwin:release`). From 1.0.1 on,
GitHub Actions builds every release: a `vX.Y.Z` tag runs `.github/workflows/release.yml`, which
builds, signs, notarizes, staples and verifies the DMG and creates a **draft** release. You publish
it by hand. Pushes and pull requests to `main` run the test suite (`.github/workflows/ci.yml`).

## Cutting a release

1. Bump the version everywhere (the release job refuses any mismatch):
   - `internal/services/updates.go`: `const Version`
   - `build/darwin/Info.plist`: `CFBundleShortVersionString` and `CFBundleVersion`
   - `build/darwin/Taskfile.yml`: `VERSION`
   - `scripts/verify-release.sh`: `VERSION`
   - `build/config.yml`: `info.version`
2. Add `docs/release-notes-X.Y.Z.md` (it becomes the release body).
3. Check locally: `bash scripts/release-version.sh vX.Y.Z` prints `X.Y.Z`.
4. Merge to `main` and wait for CI to go green.
5. Tag `main` and push the tag:

   ```sh
   git fetch origin
   git tag vX.Y.Z origin/main
   git push origin vX.Y.Z
   ```

6. The run pauses before the `sign` job until you approve it: Actions, the Release run, Review
   deployments, `release`, Approve and deploy (see Protecting the signing secrets).
7. When the run is green, open the draft on GitHub (Releases), check the DMG, its `.sha256` and
   the notes, then publish. Installed copies see the update once it is published.

The `build` job refuses, before any credential is touched, when the tag is not
`vMAJOR.MINOR.PATCH`, does not match the version sources, has no release notes, is not on `main`,
or already has a published release. Its read-only token cannot see drafts, so the `sign` job's
first step, before the keychain, refuses a tag that already has any release, draft included. That
also covers "Re-run failed jobs", which skips `build`: delete the old draft before re-running.
`v1.0.0` is excluded from the tag trigger. The run has two jobs on separate runners:

- `build` has no environment, no secrets and a read-only token. It installs dependencies, builds
  the app (`wails3 task darwin:release:app`) and uploads it as an artifact.
- `sign` runs in the `release` environment on a fresh runner. It installs only the pinned `wails3`,
  downloads the app, creates the signing keychain, then signs, notarizes and verifies it
  (`darwin:release:sign`) and deletes the keychain before `go test` compiles anything. No
  dependency install or build runs on this runner, and nothing there executes the downloaded app.

`release:app` refuses a tree whose tracked files differ from `HEAD` and stamps the bundle with the
commit and a digest of its contents (`scripts/release-stamp.sh`). `release:sign` refuses a bundle
whose stamp does not match `HEAD` and its contents, such as one left by `install:dev`. Locally,
`wails3 task darwin:release` runs both in your own shell.

## Dry run

Builds, signs, notarizes and verifies on a runner without creating a release. Like a tag run, its
gate refuses before touching any credential when the ref is not on `main`, as long as the ref's own
`release.yml` is unedited (see Protecting the signing secrets):

```sh
gh workflow run release.yml --repo scottkw/gifsmith --ref main
```

A dry run pauses before `sign` for your approval, like a tag run. The log ends with
`dry run: no release created`.

## Protecting the signing secrets

The gate in `release.yml` (version, `main` ancestry, no existing release) guards against mistakes,
not attacks. Both a tag push and `workflow_dispatch` run the `release.yml` of the ref being built,
so a tag or branch that carries an edited workflow can drop the gate and still reach the `release`
environment. The real protection lives in the repository settings, outside any workflow file
(both applied 2026-10-20):

- Tag ruleset "Release tags: maintainer only" (Settings, Rules, Rulesets), active: creating,
  updating and deleting `refs/tags/v*` is blocked for everyone except repository admins (the
  bypass list is the Repository admin role, always). Keep the admin role to the maintainer.
- The `release` environment (Settings, Environments, release) has `scottkw` as a required
  reviewer, so every tag or dispatch run pauses before the `sign` job, the only job that can read
  the secrets, until you approve it in the Actions UI. Its deployment branches and tags stay
  limited to `main` and `v*` tags. Leave "Prevent self-review" off, or you cannot approve your own
  runs.

Before approving a run, check that its commit is on `main` and that `.github/workflows/release.yml`
at that commit is the one on `main`.

## Secrets

Five secrets live in the GitHub environment `release` (Settings, Environments, release), which
deploys only from `main` and `v*` tags and needs your approval (see Protecting the signing
secrets). Only the release workflow's keychain step reads them.

| Secret | Holds |
|--------|-------|
| `MACOS_CERT_P12_BASE64` | base64 of the "Developer ID Application: Ken Scott (S2K7P43927)" certificate and private key, exported from Keychain Access as .p12 |
| `MACOS_CERT_PASSWORD` | the export password of that .p12 |
| `NOTARY_KEY_P8_BASE64` | base64 of the App Store Connect Team API key (.p8) used for notarization |
| `NOTARY_KEY_ID` | that API key's Key ID |
| `NOTARY_ISSUER_ID` | the App Store Connect Issuer ID |

### Rotating

- Signing certificate: export a fresh .p12 from Keychain Access, then
  `base64 -i <file> | gh secret set MACOS_CERT_P12_BASE64 --env release --repo scottkw/gifsmith`
  and `gh secret set MACOS_CERT_PASSWORD --env release --repo scottkw/gifsmith` (hidden prompt).
  Delete the exported file afterwards.
- Notary API key: create a new Team key in App Store Connect (Users and Access, Integrations,
  App Store Connect API), set `NOTARY_KEY_P8_BASE64`, `NOTARY_KEY_ID` (and `NOTARY_ISSUER_ID` if it
  changed) the same way, then revoke the old key in App Store Connect.
- Run a dry run after any rotation.
