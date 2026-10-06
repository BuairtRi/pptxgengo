# Additional contrast repairs for design review

Current source `4fce3cd7ecdfaf8daadfeb14bf6e4b596d731c9a`: 46 templates with 286 failing text checks across 1,947 density specimens; zero coverage gaps. The four initial repairs pass. This is an inventory for source design decisions, not a core CLI implementation defect.

Candidate treatment: keep geometry and Magenta fills; use Grounded for affected normal text consistently across all tiers. Diagram blocks need an independent ink property to preserve their fill. Keep de-emphasized decision-pin rings and change their numeral ink. No wider source color edits have been applied yet.

| Template | Failing tiers | Runs | Roles |
| --- | --- | ---: | --- |
| `activities/subtrack-timeline` | comfortable, compact, dense | 12 | small |
| `agenda/schedule` | compact, dense | 10 | body |
| `agenda/schedule-nav` | compact, dense | 10 | body |
| `agenda/schedule-right` | compact, dense | 10 | body |
| `agenda/schedule-split` | compact, dense | 10 | body |
| `agenda/schedule-tall` | compact, dense | 10 | body |
| `architecture/layer-map` | comfortable, compact, dense | 9 | small |
| `architecture/messy-hotspots` | comfortable, compact, dense | 27 | label |
| `architecture/messy-landscape` | comfortable, compact, dense | 18 | label |
| `architecture/site-rollout` | comfortable, compact, dense | 3 | label |
| `architecture/today-vs-target` | comfortable, compact, dense | 6 | label |
| `context/three-zones` | comfortable, compact, dense | 6 | label |
| `cutover/wave-matrix` | comfortable, compact, dense | 9 | small |
| `frameworks/layer-table` | comfortable, compact, dense | 3 | body |
| `from-to/rows` | comfortable, compact, dense | 12 | label |
| `guide/deck-overview` | compact, dense | 10 | number |
| `key-message/lead-questions` | compact, dense | 2 | subhead |
| `key-message/lead-questions-icons` | compact, dense | 2 | subhead |
| `key-message/lead-questions-nav` | compact, dense | 2 | subhead |
| `key-message/lead-questions-split` | compact, dense | 2 | subhead |
| `key-message/lead-questions-two` | compact, dense | 2 | subhead |
| `maturity/ai-beyond` | comfortable, compact, dense | 3 | label |
| `maturity/insights` | comfortable, compact, dense | 3 | label |
| `pathways/two-lanes` | comfortable, compact, dense | 3 | small |
| `phases/activities-outcomes` | comfortable, compact, dense | 3 | label |
| `plan/gantt` | comfortable, compact, dense | 3 | label |
| `process/current-future` | comfortable, compact, dense | 6 | body |
| `process/current-future-right` | comfortable, compact, dense | 6 | small |
| `road-fork/decision-chosen` | comfortable, compact, dense | 6 | label |
| `road/here` | comfortable, compact, dense | 3 | label |
| `roadmap/compact-narrative-right` | compact, dense | 2 | subhead |
| `roadmap/confidence-levels` | compact, dense | 2 | subhead |
| `roadmap/dependencies` | comfortable, compact, dense | 3 | small |
| `roadmap/now-next-later` | compact, dense | 2 | subhead |
| `roadmap/outcomes-left` | compact, dense | 2 | subhead |
| `roadmap/swimlane-bars` | comfortable, compact, dense | 6 | small |
| `sequence/evidence-to-model` | compact, dense | 2 | subhead |
| `status/steering-update` | compact, dense | 6 | body |
| `team/pods-pairs` | comfortable, compact, dense | 9 | small |
| `team/pods-pairs-nav` | comfortable, compact, dense | 9 | small |
| `transformation/before-after` | comfortable, compact, dense | 12 | label |
| `value-bridge/levers` | comfortable, compact, dense | 3 | small |
| `value-bridge/levers-split` | comfortable, compact, dense | 6 | small |
| `value-levers/tree` | comfortable, compact, dense | 3 | body |
| `value-phases/self-funding` | comfortable, compact, dense | 6 | small |
| `venn/four-text` | compact, dense | 2 | subhead |
