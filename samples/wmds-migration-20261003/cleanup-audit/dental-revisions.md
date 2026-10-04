# Dental conversion revisions and cleanup

The retained Dental lane has **11 successful local preferred builds** across six named versions: v1 **1**, v2 **4**, v3 **1**, v4 **3**, v5 **1**, v6 **1**. All11 PPTX and official canonical-source hashes match their receipts. **Relocation rebuilds are excluded**: two separate offline verification builds cover preferred and alternative. Five alternative local builds are also separate.

The28 saved preferred YAML/predecessor copies deduplicate to **12 distinct parsed authoring documents**, preserving every authored field. Eleven successful builds have11 distinct official compiler canonical hashes. These are retained evidence counts; intermediate in-place edits are not a complete revision history.

## Errors and repair rounds

- **31 native page-level issue events on30 slides:**30 initial v3 findings and one residual v4c page4 rule/eyebrow collision. A page finding can contain several defects;31 is not an atomic bug total.
- **Two native visual repair waves:** bulk v4 repair, then the single-rule v5 repair. Final v6 native verification accepted the result.
- Initial findings included invisible text on dark panels, decorations obscuring copy, clipping/clearance, semantic marker placement, legend distinction and an intra-word title wrap. Type categories overlap.
- **285 pre-native allocation adjustments** (185 v1 +100 v2) are recorded compiler text-capacity mutations, not285 actual native failures.
- **Two runtime data defects** on source45: category uppercasing and six numeric-zero workbook cells serialized blank. Both are repaired. Alternative builds propagate fixes rather than adding preferred errors.

Exact issue findings, page lists, receipt hashes, build hashes, source snapshot groups and package inventory are in [dental-revisions.json](dental-revisions.json).

## Minimal retained working set

Keep `dentalxchange/delivery-final/` plus one maintained preferred source (`conversion-v6`) and one maintained alternative (`alternative-deliverable-bands-v5`). The immutable final source ZIPs already include the deck source, assets, parent contracts, frozen compiler/library, lock, current build and provenance/acceptance evidence. Both relocated57-slide rebuilds matched the delivered PPTX bytes exactly.

Archive the full migration tree with its relative paths before restoring this working set. Older versions, native PNG rounds, diagnostic exports and **seven one-time Python helpers** belong in the archive. Those helpers are not required to rebuild the final decks. Preserve frozen upstream `runtime/library/wm-design-system/pinned/source/tools/build_explorations.py` and all pins in the immutable packages. Keep evidence links resolvable through the archive. Leave the original input deck untouched.

Only these two new audit files were written; source projects and delivered outputs were not changed.
