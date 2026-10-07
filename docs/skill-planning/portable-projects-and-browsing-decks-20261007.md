# Portable projects and packaged browsing decks

Requested by Ri on 2026-10-07. This extends the
[product enhancements handoff](product-enhancements-handoff-20261006.md).
The requirements below are implementation scope, not completion claims.

## Two regenerated PowerPoint libraries

Generate these during private GitLab release CI and include both in installation
archives:

1. A complete template library: native editable placeholder slides for every
   template in the pinned library, organized with family dividers and a
   “How to use this deck” component. Include supported frame/rail combinations
   and other declared variants, with explicit coverage identifying each entry.
2. A separate library of versioned reusable authored slides, organized for
   browsing and copy/paste. Preserve content, assets, source/revision provenance
   and visible approval/freshness information. Do not present placeholder
   specimens as authored reusable slides.

Colleagues must be able to copy/paste slides using PowerPoint without invoking
the CLI. Preserve native editability, theme/font requirements and useful names.
Dividers/instructions must explain inherited formatting, placeholders,
approved reuse scope and the provenance limits of a copied slide.

Use exact bundle/content pins rather than fixed catalog counts. Produce a
coverage manifest and hashes alongside both decks. Release CI must regenerate
them when inputs change and scan/attest the final archived bytes. All artifacts
remain private in GitLab. Missing private assets/content must be actionable
release blockers; an incomplete deck cannot claim complete coverage.

The existing library-reference/frame-reference and finished-slide compilation
are starting points. Enumerate valid frame/rail combinations from contracts;
do not silently omit valid variants or force incompatible combinations. The
display policy for retained reusable revisions is pending: the proposed default
is every revision, grouped by stable identity and labeled with its actual status.

## Consistent portable project layout

Projects should have a predictable layout:

    deck.yaml
    slides/
      <stable-slide-id>.yaml
      templates/
        <local-template-id>.yaml
    assets/
    versions/
      <numbered-version>/
    composition-log.yaml
    toolchain.lock.json

Use one authoritative slide order in deck.yaml; filenames and numbered deck
versions must not replace stable semantic slide identities. Shared template
definitions remain pinned references. Keep local template files under
slides/templates and owned original/derived assets under assets.

The proposed numbered-version contract retains source, assets, template pins and
generated deck together, with a visible current version. Never overwrite a
retained version or rewrite immutable build baselines. Specify the exact schema,
number allocation, latest pointer, concurrent publication behavior and recovery
before implementing the command. Existing build receipts/history remain valid.

Provide a CLI workflow to create a complete ZIP for colleagues and OneDrive
collaboration. Extend existing maintainer export where useful. Archives need
relative paths, clear open/resume instructions and an inventory/hash manifest.
Separate client-only deck delivery from private source/evidence handoff.
Document runtime/font/branding requirements and explicit toolchain migration;
portability must not be confused with changing a project's executable pin.

OneDrive is file synchronization, not a source transaction or lock service.
Detect conflicting copies, interrupted sync and changed predecessors. Preserve
both collaborators' source and native edits; never silently pick one. Define
existing-project migration with a dry run and retain preimages/receipts.

Update CLI help, presentation skill references and offline colleague guides
together. An installed version must expose commands that its documentation
describes; proposed commands must stay clearly labeled until implemented.

## PowerPoint qualification location

Use an owned stable folder under ~/Documents/pptxgengo-qualification for Mac
desktop fixtures rather than global temporary folders. Ri is granting PowerPoint
Full Disk Access. Permission state alone is not a qualification result: record
actual native open/Save As/render output, app version and file hashes.
Bound waits, close only exact task copies and preserve failed evidence.

## Acceptance evidence

- Coverage proves every pinned template and supported frame/rail variant is
  present, without silent omissions, and both browsing decks open in PowerPoint.
- A colleague copies a template and an authored slide into another deck with
  expected copy, native objects and assets on Mac and Windows.
- Release archives contain both decks, matching coverage pins, hashes and signed
  release inventory.
- New projects use the declared layout; migration preserves older project
  semantics, approvals, lineage and immutable history.
- Two numbered source/deck versions remain independent and inspectable.
- A colleague extracts a complete ZIP into a different path and can resume,
  inspect provenance and rebuild with the stated runtime requirements.
- Concurrent/conflicting OneDrive changes are detected without losing either
  source history or edited PowerPoint files.
