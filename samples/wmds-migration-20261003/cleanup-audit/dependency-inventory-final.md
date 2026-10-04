# Final maintained dependency inventory

Frozen after the final Dental and Software source cleanup. All seven projects pass frozen v4 source/pin checks and the current repository loader. Every listed source, lock, asset, context, brief and parent contract file is present with its current SHA256 recorded in `dependency-inventory-final.json`.

| Maintained project | Slides | Assets | Local definitions | Real brief files |
| --- | ---: | ---: | ---: | ---: |
| Software preferred | 83 | 38 | 35 | 83 |
| Software core alternatives | 2 | 0 | 2 | 2 |
| Software case alternatives | 2 | 0 | 2 | 2 |
| Patterson preferred | 39 | 72 | 26 | 39 |
| Patterson alternatives | 4 | 1 | 3 | 4 |
| Dental preferred | 57 | 6 | 57 | 57 |
| Dental alternative | 57 | 6 | 60 | 57 |

All **244** slide records across these projects have real brief files; alternative decks repeat original sources. The three originals retain their exact hashes and contain 179 original slides. Software's authoritative alternative sources are the `core-alternatives-2/maintained` and `case-alternatives-2/maintained` directories.

`final-retained-gallery-verification.json` verifies all 1,760 SQLite artifact references, including 587 native preview PNGs, and the unchanged accepted gallery PPTX. `recovery-current-verification.json` verifies six archive groups and 8,431 file records, including Dental's 22 predecessor records. Two Dental records are explicitly marked generated structured projections; the raw predecessor YAML files are separately retained.

`dependency-inventory-before-cleanup.json` remains the original dependency snapshot. Superseded pending audits and their preliminary check files have been retired to `history/`. The final seven-project check files are in `final-project-checks-v2/`.
