# Patterson history and cleanup audit

- **8 local successful preferred builds + 3 relocated proof builds = 11 recorded successful IDs.** They produce 7 distinct PPTX byte versions.
- **6 distinct successful content/layout states: initial success plus 5 later edits.** The extra byte version changes only eight generated section UUIDs. Relocations/restoration are reproductions, not layout edits.
- **3 native batches: v1, v3, v5.** First flagged21 pages, next flagged6 already-seen pages, final accepted39. This is27 repair-required page observations across21 unique pages, not an atomic error total. Two feedback cycles used four versioned repair scripts plus a v3 refinement. v2/v4 have no native PNG batches.
- Initial authoring has a lower bound of **24 under-height allocations on13 pages**, comparing identical nodearguments/sourcevalues/widths to actual engine measurements. This is not a failed-command count. No complete compilerfailureledger exists.
- Final preservation audits have zero missing businessvalues/notes, retaining6hidden/15note-bearing/76 originalmedia. Source22 received an extra table/milestone semantics repair. Seven finalchanged pages are not seven errors. Originaldraft placeholders/mismatchedcontext were preserved.
- Alternatives are excluded: seven local builds, four layoutstates, three revisions; four native flaggedpage observations on threeunique pages, now allfour accepted.

## Python scripts

There are12: two authoring, oneinitialfit, onemapping, twoaudits, fourpreferredrepairs andtwoalternative repairs. None is needed for a current GoCLI build. They are forensic conversionhistory and can be archived.

## Minimum active trees

Retain the unchanged preferred `deck.yaml`, `context/project.md`, `toolchain-final.lock.json`, `state.json`, `assets/`, currentbuild `build-20261004T005614-ea93f276f609e455`, the25 exact ancestry snapshots and finalprovenance/qualification receipts. Alternativeprefix is `alternatives/profile-led-layout/`; currentbuild is `build-20261004T005614-d5ecda66ff304801`, with unchanged YAML/context/lock/state/assets and its bio snapshot. Detailed paths are in the JSON.

Use exact qualified8c481 runtime/bundle/fonts/calibration from the accepted offlineZIP. All72 registry assets remain required, even if some are not shown; retain76 rawsource media for provenance. Keep onecurrentbuild/state perproject forbaseline/export protection.

Archive the wholeworkspace first with hashes; retain originalfinalZIPs and all originalPPTXs. Historical builds/reviews/drafts/helpers/measurements/oldlocks/unusedinspectioncontracts/stale MAPPING prose and12scripts can then move out. Re-export filteredprojects to newZIPpaths; verify payloads and fresh noGo whole-PPTX equality before updating manifests. Preserve raw historicalqualification paths using an explicit archiveindex. No existingsource/delivery files were changed by thisaudit.

These counts describe the prior mechanics/native-fit process. They do not establish good semantic design; the requested new design reassessment is a separate pass.
