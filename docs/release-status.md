# Release status and qualification

## v4.3.0 candidate

The semantic composition capabilities and specialized skill runbooks are merged
to main. [The completion matrix](v4.3.0-completion-matrix.json) records all 29
frozen V11 runtime families and the named native qualification cases; see the
[delivery record](v4.3.0-delivery.md) for scope and limitations. This does not
claim exhaustive desktop acceptance of every browsing slide.

Publication remains pending the final GitLab gates and independent download
verification. The release compiler is pinned to Go 1.27.2 after pre-tag standard
library vulnerability findings. The template browsing deck, SQLite catalog,
registered schematic SVG originals and installable presentation skill are in
scope. Windows runtime/PowerPoint, content-complete reusable slides and external
branding/photo distribution remain deferred. The newer documentation-site
catalog is independently versioned; this release retains the frozen V11 runtime.

## v4.2.1

The [private GitLab release](https://gitlab.samcott.com/riscott/pptxgengo/-/releases/v4.2.1)
was cut at commit `66229bd14118e1fa031e25d9d9c7e610598a6001`.
[Pipeline 21443](https://gitlab.samcott.com/riscott/pptxgengo/-/pipelines/21443)
passed all 28 gates in 793 seconds on 2026-10-08 UTC. The reviewed PR tier passed
four jobs; the merged-main tier passed four jobs in 146 seconds. GitHub hosts
source review/merge, with GitHub Actions disabled; GitLab hosts all CI and artifacts.

Six installation archives cover macOS, Linux and Windows, amd64 and arm64.
Each includes three executables, a generated 717-slide template browsing deck,
its manifest/coverage, pinned V11 source/catalog/fonts and a regenerated SQLite
index of 649 templates. The closed template-resource inventory has 2,663 files
per archive. Source revision is `wmds-library.v11`, upstream commit
`3c56d842ba3abb5f24eb33cf0082be7a7a67f116`.

The signed package policy is `cli-only` plus `templates-only`; `cli-only` does
not mean these particular archives lack authoring resources. They omit the skill,
optional upstream docs site, source examples/scripts, private branding/photo
originals and curated reusable content. `pptxgengo paths` reports conventional
optional paths even when those files are absent. Browse the included template
deck or specimen gallery; install the skill separately and verify optional paths.

Browsing uses explicit synthetic media placeholders and native-v1. Frame marks
already pinned in V11 are retained; a placeholder deck does not qualify final
branding or supplied content. The content-complete reusable browsing deck and
operator-approved reusable inventory remain deferred.

## Native structure

New v2 projects persist native-v1; existing projects without a profile retain
stock structure. Rendering converts eligible components without changing frozen
source bundle bytes. The browsing deck contains 586 native list boxes, 96 native
card shapes and 206 native tables. The 682 converted list/card objects matched
stock displayed text and measured absolute line positions. That comparison is
source-plan evidence, not exhaustive desktop/pixel qualification.

The operator approved the component demo. Full-catalog PowerPoint review,
Windows interactive runtime/Office qualification and acceptance across the
current native template families remain pending. The retained two-slide macOS
Save As/edit/reorder fixture passed its bounded checks; it does not qualify the
full catalog. See [round-trip evidence](native-roundtrip.md). A build or structural inventory never
records an operator's review automatically.

## Signing and independent verification

macOS binaries use the existing Ryne Scott Developer ID and accepted notarization
for both architectures. Windows binaries use Azure Artifact Signing with
Authenticode verification. Linux artifact integrity is covered by the private
Sigstore-signed release manifest. Final archive/model SBOMs, vulnerability policy,
binary hashes, platform evidence, Rekor and RFC3161 timestamp proof are retained.
Apple ticket lookup is online; CLI binaries do not carry a stapled application ticket.

After publication, all 35 linked files were downloaded independently; all 32
signed file hashes and SHA256SUMS matched. The manifest bundle verified offline
against the reviewed private Sigstore trust and exact v4.2.1 identity. All six
archives passed binary/resource verification, SQLite integrity/count checks and
resource inventory hashes. All six macOS tool signatures and tamper rejection
passed. A private stage-only macOS installation and default catalog queries from
a relocated/staged release passed without changing active user settings.

Linux ARM64 installation ran on the Mac mini Linux runner (24), completing in
58 seconds. PR/main/tag lanes omit race/performance work; nightly and explicit
on-demand lanes retain it. See [CI cadence](ci-cadence.md).

## Earlier releases and documentation updates

v4.2.0 at `f6555c6fc5a3fc851d9c5cbcba53eb723925cc61` published signed binaries
and the optional offline model, but no browsing decks, fonts, authoring bundle
or SQLite. v4.2.1 adds those template resources. Older v4.1.0 and local-version
qualification ledgers remain historical and must not be used to infer current
package contents or desktop acceptance.

Documentation updates after this tag can be installed separately from the
executable. They do not change signed release bytes or a project pin. See the
[documentation index](README.md) for current operator/maintainer routing.
