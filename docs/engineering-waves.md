# Engineering waves — qualification record

Date: 2026-10-04. Release: **0.1.0-local.12/v5**.

Authorized scope: Waves 1–3 and bounded Wave 4 work from
[engineering asks](skill-planning/engineering-asks.md). Edited-PowerPoint
reconciliation, automatic project import, wireframe authoring/design lint and a
new Go preview renderer remain deferred. The presentation skill belongs to a
separate agent.

## Implemented work

| Area | Implemented | Qualification limits |
| --- | --- | --- |
| Native export | Canonical paths and file identity, isolated bounded worker, exact task cleanup, operational doctor, page subsets, contact sheets, dialog probe | Unit/race checks pass; actual export from this caller fails Apple Event dispatch |
| Authoring | Source-pinned aliases, descriptions, decorative-slot filtering, 44 family recipes, scaffold comments, coverage report | Inferred metadata; human semantic review remains open |
| Capacity | Font-measured Latin estimates and geometric line budgets where supported; explicit unsupported state | Advisory, with native fit reported separately |
| Template matching | Strict semantic page input, full active-library search, explicit leaf disposition and gaps, ready slide YAML/PPTX, optional native comparison | Limited structured-role adapters; three or four complete candidates are not guaranteed |
| Assets | Grouped icon variants, photo descriptions/tags, verified original paths/hashes, browsable gallery, project image registration/focus metadata | Focus is advisory; registered originals remain authoritative |
| Slide operations | Add/move/remove/hide/show, stable IDs, section handling, shared mutation guard, receipts/history, safe swap proposals/apply | Unmapped copy removal requires an explicit flag; unfilled required fields still block |
| Review and checks | Native evidence attachments, distinct rendered/reviewed/accepted states, titles, composition/evidence checks, outline/content/deck HTML packets | Missing or stale native evidence never counts as acceptance |
| Bounded Wave 4 | Read-only source inventories and explicit unmapped project references; retired-route guidance | No edited-PPTX reconciliation or automatic semantic import |

Sol 6.1 handled native export and semantic metadata. Luna handled bounded CLI
commands, assets and inventories. The coordinator handled integration, safe
mutations, matching, review evidence, packaging and qualification. Independent
reviews covered identity/cleanup, no-loss mapping, attachment integrity, asset
originals, source pins and unsupported-case reporting; identified issues were
remediated before packaging.

## Library and semantic coverage

The release retains the frozen v5 source pin
`d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf`: 587 templates, 586 active. Original
geometry, canonical stock values and existing source preview qualification are
preserved.

- 20,529 authoring slots; 317 empty decorative slots identified.
- Zero raw `block_N`/`nodeNN` aliases; **431 templates still have ambiguous
  semantic mappings**. A friendly alias does not establish understanding of its
  business role. **Zero templates have human semantic-metadata review**.
- 5,137 slots have supported capacity estimates; **15,392 slots remain
  unsupported**, principally component internals. Unsupported slots do not claim
  measured fit.
- The three-control page fixture finds one complete candidate (`cards/3`) and
  reports gaps for the other 585 active templates. A title-only fixture returns
  zero candidates, an error exit and a retained report. No copy is invented,
  silently dropped, split or shortened to manufacture alternatives.
- Rich heat-map, Gantt and commercial/economics schemas need dedicated content
  adapters and broader reviewed metadata before broad automatic matching can be
  claimed. Ask 2's three/four-candidate acceptance and ask 3's human-readable
  coverage acceptance are **not fully met**.

The upstream checkout now contains **602 templates** (15 additions and one
existing revision). [The intake record](engineering-new-templates.md) describes
the required capabilities and migration gates. These templates are not yet in
the installed v5 library.

## Verification

- Full normal suite passes in an isolated checkout containing only this wave's
  changes. All 586 active stock examples round-trip with identical canonical
  values. Committed Software (83), Patterson (39) and DentalXChange (75) projects
  compile after relocation without rewriting their resource pins.
- The calibration test's stale DentalXChange count was corrected from 57 to the
  committed project's 75. An unrelated in-progress edit to DentalXChange slide
  41 is excluded from qualification and left intact in the main workspace.
- Race suite: all packages except `deckproject` passed the first full run;
  `deckproject` exceeded Go's default ten-minute timeout. Final package rerun
  uses a 30-minute timeout; final result pending.
- Disposable packaged-CLI workflow from outside the checkout: init/split,
  content matching, add/move/remove/hide/show, same-template swap, original image
  registration, check/build/titles, source inventory and deck review pass.
  Selected measurement exposed hyphenated IDs being parsed as page ranges;
  exact-ID resolution now precedes range parsing, numeric page behavior is
  preserved, and normal/race regressions plus a real selected measurement pass.
  Original sample sources and final deliverables are untouched.
- Packager validates all 587 accepted source previews, 1,760 linked artifacts
  and the release manifest. Gallery raster thumbnails reduce the asset browser
  from roughly 101 MB to 5.8 MB; original registered artwork is retained once.
- Real PowerPoint 16.113.3 doctor: app/version, staging writability and
  Swift/PDFKit pass. Console user is logged in, but this caller's NSWorkspace
  exposes no GUI applications. Operational Apple Events fail −10827; a direct
  verified-PID probe fails −600. Automation/file-access permission is **unknown**,
  not established as denied. A selected-page render fails in 3.83 seconds with
  an explicit diagnostic, no output PDF/PNGs or success receipt, and unchanged
  source hash. Exact task cleanup was confirmed via the PowerPoint Window menu.
- Fresh Mach probes confirm the caller's bootstrap port has a dead-name right
  and no send right. Commands still descend from the September 30 managed
  daemon (Codex 0.159.3), while bundled CLI is 0.160.0. Installed CLI help
  supports `app-server daemon update --from-cli` and `restart`. The operator
  recovery procedure is in the command guide; no runtime restart, TCC change or
  alternative-session launch was performed. Recovery remains unverified.
- Native export and appearance are **not qualified in this agent session**.
  HTML packet structure/integrity tests pass; interactive HTML visual review was
  unavailable because the browser security policy rejected local-file access.
- Global CLI installation and isolated-workspace cleanup: pending final checks.

No new Python conversion helpers were created. The existing release assembler
retains its packaging Python; implemented authoring, image thumbnails, inventory,
matching, mutation and review functionality lives in Go.

## Remaining friction observed

- Project compiler pins intentionally require explicit re-pinning for a new
  executable; updating the installed CLI does not silently rewrite final decks.
- A generated matcher file uses `candidate-001` as its stable ID. `slide add`
  requires the requested ID to match; changing its ID before insertion is still
  a manual step when choosing a meaningful project identity.
- Native export needs a working desktop caller; diagnostics shorten failed
  attempts but cannot establish that access for this session.
- Complex semantic schemas and component capacity need further implementation
  and review. Explicit gaps make these visible instead of selecting a simple
  layout with lost content.

See [the command guide](engineering-cli.md) for supported operations and examples.
