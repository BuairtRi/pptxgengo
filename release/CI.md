# Private, tag-driven release pipeline

All downloads are in the **private GitLab project** `riscott/pptxgengo` (project
17). GitHub remains the canonical source/review mirror. No GitHub release assets
are created. The first CI release is a CLI-only prerelease because registered
branding originals are currently on another machine.

## Cut a release

1. Land the release changes on `main` and wait for its developer and security jobs.
2. Create an annotated immutable tag, `vX.Y.Z-rc.N` for a prerelease or `vX.Y.Z`
   for a stable release. Push the tag through `origin`, which has both GitHub and
   GitLab push URLs. Only maintainers can create protected `v*` tags.
3. Watch the tag pipeline in GitLab. Publication is automatic only after every
   required gate succeeds. A failed gate never creates a GitLab Release.
4. Download from GitLab **Deploy → Releases**. Generic Package storage retains
   release bytes beyond the 14-day diagnostic CI artifacts.

```sh
# Example for the next new stable version; never recreate an existing tag.
git tag -a v4.1.1 -m 'Release v4.1.1'
git push origin refs/tags/v4.1.1
```

Never move or reuse a release tag. A retry can reuse already-uploaded identical
bytes, but publication refuses to replace bytes at an existing package coordinate.
If a new build/signature is needed, cut another version.

The pipeline builds all three commands (`pptxgengo`, `pptxdesign`, `wmdsdocs`) for
macOS, Linux and Windows, each for amd64 and arm64. Unsigned binaries must match
across two builds with separate caches. Each includes a version/commit/target
marker, checked alongside Go build metadata and per-executable SHA-256 evidence.
Signing changes the binary hashes; signed hashes and original reproducibility
proofs are retained separately.

## Signing and release gates

- Windows cross-builds on Linux, then Jsign uses Azure Artifact Signing with a
  short-lived GitLab OIDC credential. Linux independently verifies Authenticode,
  the Ryne Scott publisher, certificate chain, timestamp and tamper rejection.
- macOS binaries use the protected `macos` shell runner, Developer ID Application
  `Ryne Scott (VZJU7JS89T)`, hardened runtime and secure timestamps. Apple must
  report `Accepted` for every submitted binary before packaging. A separate job
  verifies final archive binaries and queries Apple's live notarization status.
  Bare CLIs and ZIPs cannot be stapled; first-use Gatekeeper checks need Apple's
  online ticket lookup. This pipeline does not claim offline stapling.
- Linux downloads run an amd64 CLI smoke probe. The arm64 build has archive,
  identity and checksum checks; it is not yet a native arm64 runtime result.
- All branches run redacted Gitleaks and reachable Go vulnerability checks for
  Linux, Darwin and Windows. JSON vulnerability output is explicitly inspected:
  `govulncheck` JSON mode alone does not return failure for findings.
- Final signed archives, rather than the source checkout, receive CycloneDX
  SBOMs and Grype reports. Critical findings block release; no vulnerability
  suppressions are configured. The database must be valid and at most 120 hours
  old. The source gate independently blocks every reachable Go vulnerability.
- The manifest binds archives, SBOMs, scans, build proofs and notary evidence to
  the tag and source commit. Private Fulcio/Rekor/TSA services sign it using the
  GitLab `sigstore` ID token, with a pinned TUF root, inclusion proof and RFC3161
  timestamp. This signature covers Linux binaries through their archive hashes.
- Publication re-verifies the manifest signature, enforces project privacy,
  downloads and hashes every uploaded file, then creates the GitLab Release.

Scanner/signer versions and download SHA-256 pins are in `tool-pins.json`.
Container images are pinned to the existing home-lab image digests. Apple and
Azure signing variables were copied as protected variables from ATV2; Azure has
its own federation constrained to project 17 and `v*` tags. Values remain outside
Git. The macOS keychain/notary profile stays on the existing protected runner.

The final privacy guard uses `RELEASE_PRIVACY_READ_TOKEN`, a project 17 token
with Guest access and only `read_api`. GitLab's project metadata endpoint does
not accept CI job tokens. This protected, masked, hidden variable is scoped to
the `release-publication` environment and expires on 2027-10-06; rotate it before
then. It cannot write packages or releases. Publication writes still use the
ephemeral `CI_JOB_TOKEN`. Both project and package registry privacy must pass.

## Verify a download

Download `manifest.json`, `manifest.sigstore.json`, and the desired archive and
reports from the same GitLab Release. On Linux x86_64, from a trusted source
checkout (not from an unverified archive):

```sh
bash scripts/release/install-tools.sh --bin-dir /tmp/pptx-release-tools --only cosign
export PATH="/tmp/pptx-release-tools:$PATH"
bash scripts/release/sigstore.sh initialize --home /tmp/pptx-release-trust
bash scripts/release/sigstore.sh verify --home /tmp/pptx-release-trust \
  --blob downloads/manifest.json --bundle downloads/manifest.sigstore.json \
  --version v4.1.0
```

The signature verifier requires the exact project/tag certificate identity, issuer,
Rekor evidence and signed timestamp. Its bundle verification is offline after
initializing trust. Then compare archive SHA-256 with the signed manifest.
`SHA256SUMS` is a convenience inventory; verify the signed manifest first.

## Bring private branding into CI

The branding files are deliberately external inputs. On the machine containing
`~/Documents/branding` (or set `WMDS_BRANDING_ROOT`), use the same source revision:

```sh
scripts/release/create-branding-input.sh /private/path/branding.tar.gz
```

This stages and verifies the existing source release, then archives its verified
registered originals. Upload the archive **privately** to S3 or this project's
Generic Packages. Do not add branding archives to GitHub or the source tree.
Configure these **protected** GitLab variables:

| Variable | Value |
| --- | --- |
| `PPTXGENGO_PACKAGE_KIND` | `full` |
| `WMDS_BRANDING_ARCHIVE_URL` | masked HTTPS presigned S3 URL, or project 17 Generic Package URL |
| `WMDS_BRANDING_ARCHIVE_SHA256` | exact archive digest from the producer |

Cut a new tag. CI verifies the archive hash, rejects unsafe members, and runs the
existing source/gallery/asset/documentation verifier in stage-only mode. Full
archives pair these resources with final signed platform binaries and regenerate
package hashes. Windows receives its existing user installer, smoke harness and
offline guides; macOS/Linux can run directly after adding the extracted `bin` to
PATH, with font installation as a separate local action. Full mode fails if the
input is absent or differs; it cannot silently downgrade to CLI-only mode.

Full package assembly is implemented, but cannot be qualified until the private
input exists. Signing a CLI does not qualify arbitrary presentation content.

## Windows desktop qualification

A Windows node is expected shortly. It is needed for execution and interactive
PowerPoint COM validation, not compilation or Authenticode signing. The existing
opt-in job requires runner tags `windows` and `pptxgengo-windows-native` and must
run from the signed-in desktop, not Session 0. Until that lane runs, release
metadata keeps `windows_runtime_qualified` and `native_powerpoint_qualified` false.
The first CI prerelease therefore must not be promoted as a fully qualified
Windows presentation package.

## slotctl

The project is registered under key `pptxgengo` with `.slotctl.yaml`, no shared
provider claims, and preparation/preflight tasks calling the same Make targets
used by CI. The installed CLI lives outside this shell's PATH:

```sh
/Users/rscott/.local/share/slotctl-bootstrap/bin/slotctl \
  --config /Users/rscott/.config/slotctl/config.yaml \
  --repository pptxgengo slot create --name pptx-dev --target main
```

The primary/default branch is now `main`; `master` remains available. At setup,
all 32 daemon allocations were occupied and capacity cleanup was being handled
elsewhere. The release implementation used an isolated ordinary worktree while
waiting. No daemon restart, registry editing or retirement of other work was done.

Capacity cleanup and the shared daemon's upgrade to 0.0.50 allowed the original
operation `slot-create/d800a961c03c481d0962291749384319` to resume successfully.
The prepared managed slot `pptx-release-packaging` is intentionally stopped
(there are no service claims). Its immutable identity is
`bc7d977e-dc1d-413e-a898-117b82119e9e`, machine index 2. Preparation completed
dependency download and the three-command build; diagnosis confirmed its
worktree and canonical identity match. Future slots should target `main`.

References: [GitLab releases](https://docs.gitlab.com/user/project/releases/),
[GitLab Azure OIDC](https://docs.gitlab.com/ci/cloud_services/azure/),
[Go vulnerability checker](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck),
and [Sigstore blob signing](https://docs.sigstore.dev/cosign/signing/signing_with_blobs/).

The legacy library index now uses the same pure Go SQLite driver as the unified
index; distribution binaries do not require a separate system sqlite3 CLI.

The pinned shared security/signing images contain Go 1.26.5. Linux jobs install
Go 1.27.1 from official archives with checked-in SHA-256 pins before scanning or
building. macOS uses the same verified archive approach. GOTOOLCHAIN stays local.
