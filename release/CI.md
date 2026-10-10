# Private, tag-driven release pipeline

All downloads are in the **private GitLab project** `riscott/pptxgengo` (project
17). GitHub remains the canonical source/review mirror. No GitHub release assets
are created. All builds, tests, security scans and release operations run in
GitLab; GitHub Actions is disabled. The v4.3.1 source selects V12 (735 templates, 734 active; 804 browsing slides).
The last independently verified stable package is v4.2.1. Its protected
GitLab pipeline generated template browsing resources without requiring the
private original-photo/branding collection or a reusable-slide inventory.

## Cut a release

1. Land release changes on `main` and wait for every applicable GitLab gate,
   including Linux and protected Mac native CLI/model checks.
2. Create an annotated immutable tag, `vX.Y.Z-rc.N` for a prerelease or `vX.Y.Z`
   for a stable release. Push the tag through `origin`, which has both GitHub and
   GitLab push URLs. Only maintainers can create protected `v*` tags.
3. Watch the tag pipeline in GitLab. Publication is automatic only after every
   required gate succeeds. A failed gate never creates a GitLab Release.
4. Download from GitLab **Deploy → Releases**. Generic Package storage retains
   release bytes beyond the 14-day diagnostic CI artifacts.

```sh
# Example for the next new stable version; never recreate an existing tag.
git tag -a v4.2.2 -m 'Release v4.2.2'
git push origin refs/tags/v4.2.2
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

Protected tag pipelines require Linux tests/security, six-target compilation,
and protected Mac ARM64 unit/installer/model checks before resource or
binary builds. Windows native execution and Intel Mac execution remain separate
opt-in private GitLab lanes pending their runners; compilation is not runtime
qualification. See [CI policy](../docs/testing.md).

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
- PR/main pipelines run redacted Gitleaks; protected tags also require reachable Go vulnerability checks for
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
  --version v4.2.1
```

The signature verifier requires the exact project/tag certificate identity, issuer,
Rekor evidence and signed timestamp. Its bundle verification is offline after
initializing trust. Then compare archive SHA-256 with the signed manifest.
`SHA256SUMS` is a convenience inventory; verify the signed manifest first.

## Optional offline search model

Protected tags, nightly schedules and selected on-demand pipelines run
`security:offline-model`. Ordinary branch pushes create no pipeline. It verifies the pinned complete package from the offline model
runtime gate, creates one deterministic ZIP and scans its actual extracted files.
Compiled-in identity, artifact hashes/sizes, license and source attribution must
match; caller-provided manifest hashes cannot override pins. File/count/size
bounds and closed inventory apply before archive extraction.

`PPTXGENGO_OFFLINE_MODEL=true` (default) attaches this ZIP separately from all six
CLI archives. Set it to `false` to omit that release attachment. Assemble copies
verified bytes/evidence; final scan rechecks and rescans the final ZIP. Seal
requires its file evidence, SBOM and vulnerability report. The private Sigstore
manifest and private Generic Package publication cover all four files. No model
weights enter Git, GitHub release assets or platform executable archives.

Model weights/tokenizers are platform-independent data files; Authenticode and
Developer ID/notarization remain for executables. Model scans record explicit
hashed file components, Apache-2.0 license and source attribution, and enforce
the same fresh-database/Critical/no-suppression policy. They do not prove model
safety, retrieval quality or native qualification. The `v0.0.0` branch fixture
is never a release and expires as a private CI artifact after three days.

Download the signed release manifest, signature bundle and model ZIP from the
same private release, verify the signature/ZIP hash, then extract to a separate
model directory. Normal searches remain offline and never fetch weights. Build
source-bound embeddings for the exact selected library. This change adds future
CI automation; it does not modify released v4.1.0 or cut a new tag.

## Template browsing resources (v4.2.1 policy)

The protected release pipeline sets `PPTXGENGO_PACKAGE_KIND=cli-only` and
`PPTXGENGO_BROWSING_POLICY=templates-only`. This produces six platform archives
with the three CLI tools plus a closed, hashed template authoring catalog. The
`cli-only` package-kind label preserves its binary/install format; it does not
mean the template resources are omitted.

Each archive contains the 717-slide template browsing deck, provenance manifest,
`native-editing-coverage.json` for all 649 catalog templates, the v11 template
source bundle, fonts, catalog and freshly regenerated `library.sqlite`. SQLite
is the search index; templates render from the bundled source and compiler inputs.
Installer relocation checks exercise catalog search, and strict archive
verification rejects stale, extra or missing files. Existing pinned frame marks
are included as template prerequisites.

Generation uses `--editing-profile native-v1` and
`--media-policy placeholders`. Synthetic diagrams represent external photos,
icons and artwork; they are identified as placeholders and are not original
brand assets. The package does not require unavailable private photo/branding
originals or a finished reusable-slide inventory. The separate content-complete
reusable browsing deck remains deferred. Legacy `required` policy and its
two-deck/private-input validation remain supported independently; `deferred`
remains available for older CLI-only packages. v4.2.0 remains immutable and does
not gain these resources retroactively.

Pipeline 21443 for tag `v4.2.1` at `66229bd1` completed all 28 jobs successfully
in 793 seconds. Resource generation recorded 2,663 files and a regenerated
SQLite index with 649 templates. Private GitLab holds CI artifacts and release
packages; no GitHub release assets are created.

## Signing and qualification status

The v4.2.1 release completed Windows Azure Artifact Signing and macOS Developer
ID signing with Apple notarization acceptance, followed by signed archive
verification and private manifest attestation. Windows runtime execution and
interactive native PowerPoint qualification remain pending. Cross-compilation,
Authenticode verification and signing do not establish Windows runtime or desktop
Office qualification; release metadata keeps both qualification flags false.

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
Go 1.27.2 from official archives with checked-in SHA-256 pins before scanning or
building. macOS uses the same verified archive approach. GOTOOLCHAIN stays local.
