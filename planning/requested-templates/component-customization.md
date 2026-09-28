# Requested source-template customization and sample cleanup assessment

## Completed reversible archive

The root-level cleanup review archived 20 superseded adaptive probe directories
as a reversible move with compatibility symlinks. The exact per-file paths and
SHA-256 records are in [`cleanup.json`](cleanup.json): 36 files, approximately
9.5 MB. The move preserved existing paths through symlinks. It deliberately
kept `samples/adaptive/native-cache`, current accepted probes and evidence, all
source and working decks, and the active native workspace. The open AI working
deck was left untouched. Do not infer clearance for any additional samples from
this completed archive.

## Customization direction

The selected UHG and AI Accelerator slides are source designs to preserve. Use
the registered source-bound contracts as the supported customization surface:
edit only named text slots, exact color roles/profiles, and declared empty zones.
Retain hierarchy, geometry, charts, artwork, and source palette unless the
specific contract exposes an exact binding. Do not use global hex replacement.

The preferred visual roles for new components are neutral gray surfaces
(`#E8EEF8` / `#F4F6FA`), navy ink (`#070154`), blue emphasis (`#0047FF`), and
pink (`#F900D3`) as a restrained accent. Preserve evidence and state encodings.
Pink is already a client part-time semantic token in the team composition;
outside that meaning, it must remain decorative and must not imply status or
rating. Check contrast against the actual surface for each instance.

### Gauge treatment

Keep the target highlight and actual-value pointer as independent visual channels.
The T045 source comparison template retains its native gauge design but now has
a separate source-preserving `apply-gauge` operation. It accepts highlighted
cells and a pointer cell for each gauge, using discrete source positions 1–5. By
default, one highlighted cell also selects the pointer; an explicit pointer may
be chosen elsewhere. The source legend determines what those marks mean. Highlight
colors are gray, navy, blue, or pink. The five-cell arcs remain intact; the two
scale bands above the table are not controlled by this operation. Text-bound
min/max labels are editable but do not automatically calculate or move gauge
marks. Keep the chosen cells/pointer explicitly tied to the source facts and
legend. See the packaged T045 `gauge-authoring.md` for the exact schema and
command sequence.

The semantic comparison composer calculates target-band and actual-value pointer
positions from caller-supplied numeric inputs and supports linear and dial forms.
Target color follows `state.active`; the pointer uses contrast-resolved ink and
has no independent color role. This is a separate model gap for continuous
gauges, not for T045's discrete source gauge control.

### Accents

Keep highlight/underline artwork local to the component or exact measured phrase
it emphasizes. Accents are not full-slide templates or alternate page backgrounds.
The supported phrase workflow has bounded artwork, phrase, and geometry cases;
use manual placement or another design when it cannot anchor unambiguously.

See the installed skill's [component customization reference](../../skills/west-monroe-presentations/references/component-customization.md)
for route choice, supported profiles, palette roles, and gauge boundaries.

## Sample inventory snapshot

Read-only inventory on 2026-09-27:

- `samples/` contains about 9.1 GB, 77,193 files, and 420 non-lock `.pptx`
  files. The data is broadly ignored by Git, so ordinary status checks do not
  expose the consequences of moving or deleting it.
- Large clusters include `samples/template-expansion` (5.2 GB),
  `samples/proposal-authoring` (about 1.0 GB), `samples/reconstruction`
  (about 769 MB), `samples/library-wave4` (about 302 MB),
  `samples/component-adaptation` (about 297 MB), and `samples/adaptive`
  (about 180 MB). These include mixtures of source copies, native evidence,
  probes, superseded variants, and accepted review artifacts; size alone does
  not establish that any file is disposable.
- `samples/requested-templates` (about 179 MB) holds the active selected source
  extractions and review artifacts. The source decks declared in
  `planning/requested-templates/request.json` are hash-pinned inputs. Preserve
  those originals, all extracted projects/binding guides, and any working copies.
- Preserve Patterson and AI Accelerator working decks and open lock files;
  preserve `samples/visual-wave3` as the established native workspace. Do not
  clean source decks, working decks, render evidence, proof dependencies, or
  anything with a live PowerPoint session.
- `samples/_archive/` already contains dated manifests for prior archive actions.
  Keep those manifests and archived files immutable; use them as an audit model,
  not as authorization to extend that operation.

## Earlier candidate scan (historical)

Before the root-level closure review, a first-pass textual reference scan found
these generated adaptive probe/cache directories with no direct path or hash
hits in the manifests checked (about 20 MB combined):

```text
samples/adaptive/architecture-team-v2-probe
samples/adaptive/architecture-only6v5a-probe
samples/adaptive/team-only6v4-final-probe
samples/adaptive/comparison-linear-v5-probe
samples/adaptive/arrow-qualification-v2-probe
samples/adaptive/architecture-team-v1-probe
samples/adaptive/native-cache
samples/adaptive/process-variants-v3-probe
samples/adaptive/arrow-qualification-v3-probe
samples/adaptive/arrow-review-v1-probe
samples/adaptive/dial-variants-v1-probe
samples/adaptive/architecture-only6v5b-probe
samples/adaptive/arrow-qualification-v1-probe
samples/adaptive/roadmap-final-v4-probe
samples/adaptive/roadmap-comparison-v3-probe
samples/adaptive/team-only6v4-probe
samples/adaptive/team-only6v4-baseline-probe
samples/adaptive/roadmap-final-v6-probe
samples/adaptive/comparison-v4-probe
samples/adaptive/process-normal-v2-probe
samples/adaptive/architecture-only6v3-probe
samples/adaptive/roadmap-final-v2-probe
samples/adaptive/comparison-v2-probe
samples/adaptive/comparison-illustrative-v1-probe
samples/adaptive/architecture-only6v4-baseline-probe
samples/adaptive/process-normal-v1-probe
samples/adaptive/architecture-only6v4-probe
samples/adaptive/roadmap-final-v5-probe
samples/adaptive/architecture-team-v1-probe-cache
samples/adaptive/team-only6v3-probe
samples/adaptive/comparison-v3-probe
samples/adaptive/roadmap-final-v3-probe
samples/adaptive/roadmap-normal/probe-dialbin
samples/adaptive/comparison-normal-v2/probe-dialbin
```

The root-level review archived 20 directories from the scan. The remaining
entries—including `native-cache` and current evidence—were intentionally kept.
The historical scan is not approval to move those remaining entries; textual
search cannot rule out dynamic references or references outside this checkout.

For any broader cleanup, prepare a candidate manifest with each file's current
path, size, SHA-256, creation/version context, proposed archive path, and every
manifest/proof/document reference found in the repository and sample-side
metadata. Exclude source/work decks, active extraction projects,
`visual-wave3`, accepted evidence, all hash-pinned dependencies, and open
PowerPoint files. Do not infer supersession from a `v1`/`v2` name alone.

Any future review should focus on obsolete generated probe/build/render bundles
whose later version is the current evidence and whose earlier artifact hashes are
not referenced by any checkpoint, QA ledger, gallery, rollout manifest, or
reproducibility record. Keep each candidate set isolated in a new dated archive
directory with an immutable move/copy manifest. Verify archived hashes and all
references before considering removal of any original. Because the initial move
would change paths used by evidence, get a reference-closure review before
executing it. This document is an assessment only; no candidate is approved for
movement.
