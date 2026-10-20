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

6. If the `release` environment has a required reviewer, approve the run (Actions, the Release
   run, Review deployments).
7. When the run is green, open the draft on GitHub (Releases), check the DMG, its `.sha256` and
   the notes, then publish. Installed copies see the update once it is published.

The `build` job refuses, before any credential is touched, when the tag is not
`vMAJOR.MINOR.PATCH`, does not match the version sources, has no release notes, is not on `main`,
or already has a release. `v1.0.0` is excluded from the tag trigger. The run has two jobs on separate runners:

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

Builds, signs, notarizes and verifies on a runner without creating a release. Like a tag run, it
refuses before touching any credential when the ref is not on `main`:

```sh
gh workflow run release.yml --repo scottkw/gifsmith --ref main
```

The log ends with `dry run: no release created`.

## Secrets

Five secrets live in the GitHub environment `release` (Settings, Environments, release), which
deploys only from `main` and `v*` tags. Only the release workflow's keychain step reads them.

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
