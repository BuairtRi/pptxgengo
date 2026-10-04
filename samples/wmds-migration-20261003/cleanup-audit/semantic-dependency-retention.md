# Current semantic projects: retention boundaries

Read-only dependency inventory; no files moved or deleted.

## Minimum retained working set

Keep each current preferred/alternative project's `deck.yaml`, pinned toolchain lock, project state, current build directory, registered asset files, referenced context/page-brief files, and local-template definition snapshots. Keep final delivery PPTX/PDF/source ZIPs, delivery manifest, page map, source-preservation and native acceptance/carry receipts, and the successful relocated no-Go rebuild proof. Archive predecessor builds, native repair histories and authoring helpers recoverably after the final dependency closure and receipts have been verified.

Current projects include Patterson `semantic-remap` and `semantic-remap/alternatives/story-layouts`; Dental `semantic-remap/project-v2` and its maintained alternative; Software `semantic-remap/preferred-83` and its current alternatives. Software native qualification is still in progress: retain its source inventory, original native PNGs and active authoring lanes until final integration/acceptance.

Patterson dependencies: preferred 72 asset entries, 26 local templates, 13 referenced shared-parent snapshots and 39 page briefs; alternatives one asset entry, three local templates/snapshots and four page briefs. Software's initial frozen 83-page build has 76 registered assets. Recompute dependency closure from the final frozen source rather than relying on counts from the earlier 65-page draft. Software's frozen-v4 inline `brief` strings are editorial prose rather than real file dependencies; the corrected contract requires relative page-brief paths before adopting a newer loader.

## Scripts

The migration inventory contains **60 one-off Python helpers**. No Python helper is required by the current semantic Go project builds. Archive these with the recoverable history; do not remove any immutable runtime/source-tool dependency.

- Retain untracked `scripts/build-wmds-production-gallery.py`: `scripts/install-local-release.sh:426` copies it into release staging, and historical staged manifests pin it.
- Archive untracked `scripts/snapshot-wmds-intake.py` with intake observation history: no current installer/project runtime dependency was found.
- Preserve all **110 tracked files under `scripts/`** and immutable upstream snapshot tools such as `source/tools/build_explorations.py`.

## Incoming gallery history

The old feedback/gallery trees and predecessor v4/v5 production builds are archive candidates after the final v5-production-v3c gallery dependencies are retained. Keep the accepted 587-preview catalog, matching source/bound artifacts, native hash receipts, frozen bundles v1–v5, font/calibration/provenance files, and the SQLite index with its gallery/bundle references. Any duplicate PNG directory may be archived only when its exact hashes remain available and receipt paths resolve to the retained catalog or recoverable archive.

## Archive gate

Create the external recoverable archive and per-file hashes first. Restore the explicit current dependency closure and final delivery allowlist; verify receipt links and run relocated no-Go rebuild comparisons after restoration. Do not move active lanes, compatibility bundles, original input decks or the qualified gallery. The larger machine-readable inventory is `dependency-inventory-before-cleanup.json`; earlier Software draft counts must be refreshed at its final freeze.
