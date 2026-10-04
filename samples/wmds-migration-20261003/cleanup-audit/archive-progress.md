# Limited recoverable cleanup

Three authorized cleanup steps are complete. Current presentation projects and the accepted v5-production-v3c gallery remain in place.

| Archived group | Files | Exact bytes | Recovery manifest |
| --- | ---: | ---: | --- |
| Ten superseded incoming roots | 4,171 | 1,785,874,904 | `/Users/rscott/Projects/pptxgengo-artifact-archive/20261003-wmds-cleanup/incoming-history-recovery-manifest.json` |
| Sixty migration Python helpers plus intake snapshot helper | 61 | 494,255 | `/Users/rscott/Projects/pptxgengo-artifact-archive/20261003-wmds-cleanup/python-helper-recovery-manifest.json` |
| Rejected migration entries outside retained semantic trees | 4,163 | 1,915,225,098 | `/Users/rscott/Projects/pptxgengo-artifact-archive/20261003-wmds-cleanup/rejected-migrations-recovery-manifest.json` |

Every archived file has its old path, archive path, size, and matching pre/post SHA256 recorded. No files were deleted.

All 1,760 SQLite gallery artifact references, including 587 previews, resolve and match their pinned hashes after the incoming move. The accepted gallery PPTX retains SHA256 `657b6fee2c976057467cbc513c14906291fe222532cb086db55e4a7b6fd03b21`; SQLite bytes also remain unchanged.

The helper inventory contained exactly 60 migration paths. Current source/runtime reference searches found no dependency on those helpers or `scripts/snapshot-wmds-intake.py`. All 110 tracked scripts, the installer-dependent `scripts/build-wmds-production-gallery.py`, and five frozen bundle `source/tools/build_explorations.py` snapshots remain present with unchanged hashes. No migration `.py` helpers remain in the working tree.

The rejected migration scope contains 36 entries: all immediate Software/Patterson children except `semantic-remap`, plus the old `dentalxchange` tree. Both audit agents cleared their legacy dependencies; upstream copied its last selection matrix before the move. Six current projects pass source/pin checks before and after the archive. All 334 recorded required source, lock, asset, context and contract files remain present with unchanged hashes. All semantic trees, `dental-semantic-remap`, shared `source-inventory`, and original input decks remain in place.

Total archived across these three steps: **8,395 files, 3,701,594,257 bytes**. Machine-readable summaries: `incoming-history-archive-result.json`, `python-helper-archive-result.json` and `rejected-migrations-archive-result.json`.

Final verification also covers Software case alternative, Patterson and Dental package predecessors. Across all six recovery groups: **8,431 file records, 3,835,263,308 bytes**, all current hashes verified. See [recovery-current-verification.json](recovery-current-verification.json). Two Dental records are explicitly identified as generated structured projections; their raw predecessor YAML files are separately retained. [The final inventory](dependency-inventory-final.md) records all seven maintained projects and 244 real brief files. Pending audits have been retired to `history/`.
