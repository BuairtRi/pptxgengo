# V11 density contrast triage — complete audit

Historical machine summary: [`contrast-audit-4fce.json`](contrast-audit-4fce.json), source pin `4fce3cd7ecdfaf8daadfeb14bf6e4b596d731c9a`. This superseded audit retains the original full-report SHA-256, specimen/check counts and all 286 failing checks. Passing objects have been omitted. The original complete sweep covered emitted manual-diagram labels, rich-text runs and native-chart axis/data options. Current qualification uses [`contrast-audit-3c56-final.json`](contrast-audit-3c56-final.json), with zero failures; the recommendations below describe the historical investigation.

- 1,947 specimens; 62,850 text/role checks; 286 failures; zero coverage errors.
- 46 unique failing templates. By tier: 66 comfortable, 110 compact, 110 dense. 28 templates include comfortable failures; 18 fail only at compact/dense.
- Ink/background pairs: `#FFFFFF` on `#F900D3` (196), `#F900D3` on `#FFFFFF` (84), and `#97A4BA` on `#FFFFFF` (6).
- Roles: label 105, small 81, body 68, subhead 22, number 10.

## Templates with comfortable-tier failures (28)

`activities/subtrack-timeline`; `architecture/layer-map`, `messy-hotspots`, `messy-landscape`, `site-rollout`, `today-vs-target`; `context/three-zones`; `cutover/wave-matrix`; `frameworks/layer-table`; `from-to/rows`; `maturity/ai-beyond`, `insights`; `pathways/two-lanes`; `phases/activities-outcomes`; `plan/gantt`; `process/current-future`, `current-future-right`; `road-fork/decision-chosen`; `road/here`; `roadmap/dependencies`, `swimlane-bars`; `team/pods-pairs`, `pods-pairs-nav`; `transformation/before-after`; `value-bridge/levers`, `levers-split`; `value-levers/tree`; `value-phases/self-funding`.

## Compact/dense-only failures (18)

`agenda/schedule`, `agenda/schedule-nav`, `agenda/schedule-right`, `agenda/schedule-split`, `agenda/schedule-tall`; `guide/deck-overview`; `key-message/lead-questions`, `lead-questions-icons`, `lead-questions-nav`, `lead-questions-split`, `lead-questions-two`; `roadmap/compact-narrative-right`, `confidence-levels`, `now-next-later`, `outcomes-left`; `sequence/evidence-to-model`; `status/steering-update`; `venn/four-text`.

## Confirmed source/color paths and fix direction

- Agenda schedule family: frozen source `source/templates/library/covers.json` uses `keyInk: "callout"` on light. The Go schedule primitive passes this value through on the source surface (`internal/wmdesign/scene_primitives.go`, schedule case); tokens resolve it to magenta on white. Comfortable 14pt bold passes the large-text threshold; compact 12pt and dense 11pt fail. Change small schedule key ink to `primary` (navy) while retaining bold weight.
- Guide/deck overview and other text with explicit `ink: "callout"` on light have the same magenta-on-white problem at small sizes. Change those small text roles to `primary`.
- Activity G1–G4 and architecture hotspot number blocks use `surface: "callout"`; Go renders `display` ink on that surface (`internal/wmdesign/scene_diagrams.go`, block case), producing white on magenta. To preserve the magenta shape, add an explicit block `ink: "primary"` source option and pass it through the renderer; tokens define callout-primary as navy on magenta. Switching the surface to inverse/subtle is simpler but changes the visual accent.
- The six `road-fork/decision-chosen` failures are two inactive milestone numerals at each body density. `roadForkPin` correctly uses `#97A4BA` for the inactive pin outline but also uses it for numeral ink (`internal/wmdesign/scene_roadfork.go`, `roadForkPin`). Keep the gray outline and render numeral ink `#070154` on the white pin.

## CLI scope and verification status

The CLI exposes `measure-style` for one source role/text measurement and density/scope selection. Its JSON is explicitly `measurement_only`; it does not claim template fit or perform the 62,850-case contrast audit. There is no CLI command for whole-template contrast reporting in the listed `pptxdesign` verbs. Density build warnings are emitted separately by build paths.

Previously completed matching focused checks passed: `go test -short ./internal/deckproject ./cmd/pptxdesign`; focused density/capacity tests in `internal/wmdesign`; and `git diff --check`. This checkpoint ran no new tests or long jobs.
