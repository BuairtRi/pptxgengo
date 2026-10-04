# Template mapping and designer gaps

71 of 75 slides (94.7%) use genuine shared v5 templates. Shared definitions are referenced by ID and are not copied or geometrically modified. Four slides use three existing local definitions.

## Designer gaps

| Slides | Local definition | Required library capability |
| --- | --- | --- |
|8|stage-coverage|Seven lanes × six PDLC stages, non-calendar spans, configurable ownership legend. Stock swimlane lacks the required rows and stages.|
|36,54|role-ownership|Nine roles × six artifacts with the source L/C/A/I meanings. Stock RACI has fewer roles and different letter semantics.|
|45|activation-chart|Editable stacked columns: four series × three categories, exact numeric zeros. Stock column rail exposes two unstacked series.|

Slide 11 (and 43) uses stock capability-heat/dense. The visible legend treats codes as categories, not measured maturity scores. Slide 7 uses the stock interviews-readout/compact roster; dates and follow-up were not supplied. Slide 35 uses stock runbook/step-table to link artifacts to stages and review decisions. Slide 41 uses stock offers-onepager/three-band. Roadmaps use stock roadmap/grid-rows and roadmap/outcomes-left. The source mainly specifies Now/Next/Future horizons and six-week phase groups, rather than per-task start/end dates; no task dates were invented to populate Gantt bars. The seven-lane PDLC stage-coverage diagram remains an explicit designer gap.

## Slide map

| Slide | Source | Template | Editable file |
| --- | --- | --- | --- |
|1|Operating model 1|shared:cover/photo|[YAML](slides/001-dxc-001.yaml)|
|2|Operating model 2|shared:agenda/index|[YAML](slides/002-dxc-002.yaml)|
|3|Operating model 3|shared:runbook/step-table|[YAML](slides/003-dxc-003.yaml)|
|4|Operating model 4|shared:divider/panel-edge|[YAML](slides/004-dxc-004.yaml)|
|5|Operating model 5|shared:divider/panel-edge|[YAML](slides/005-dxc-005.yaml)|
|6|Operating model 6|shared:understanding-situation/bands|[YAML](slides/006-dxc-006.yaml)|
|7|Operating model 7|shared:interviews-readout/compact|[YAML](slides/007-dxc-007.yaml)|
|8|Operating model 8|local:stage-coverage|[YAML](slides/008-dxc-008.yaml)|
|9|Operating model 9|shared:divider/panel-edge|[YAML](slides/009-dxc-009.yaml)|
|10|Operating model 10|shared:cards/3|[YAML](slides/010-dxc-010.yaml)|
|11|Operating model 11|shared:capability-heat/dense|[YAML](slides/011-dxc-011.yaml)|
|12|Operating model 12|shared:runbook/step-table|[YAML](slides/012-dxc-012.yaml)|
|13|Operating model 13|shared:cards/3|[YAML](slides/013-dxc-013.yaml)|
|14|Operating model 14|shared:pillars/two-categories-six|[YAML](slides/014-dxc-014.yaml)|
|15|Operating model 15|shared:text/three-column|[YAML](slides/015-dxc-015.yaml)|
|16|Operating model 16|shared:text/three-column|[YAML](slides/016-dxc-016.yaml)|
|17|Operating model 17|shared:cards/3|[YAML](slides/017-dxc-017.yaml)|
|18|Operating model 18|shared:roadmap/grid-rows|[YAML](slides/018-dxc-018.yaml)|
|19|Operating model 19|shared:text/three-column|[YAML](slides/019-dxc-019.yaml)|
|20|Operating model 20|shared:roadmap/grid-rows|[YAML](slides/020-dxc-020.yaml)|
|21|Operating model 21|shared:roadmap/outcomes-left|[YAML](slides/021-dxc-021.yaml)|
|22|Operating model 22|shared:roadmap/outcomes-left|[YAML](slides/022-dxc-022.yaml)|
|23|Operating model 23|shared:roadmap/outcomes-left|[YAML](slides/023-dxc-023.yaml)|
|24|Operating model 24|shared:understanding-situation/bands|[YAML](slides/024-dxc-024.yaml)|
|25|Operating model 25|shared:workstreams/five-with-risks|[YAML](slides/025-dxc-025.yaml)|
|26|Operating model 26|shared:pillars/two-categories-six|[YAML](slides/026-dxc-026.yaml)|
|27|Operating model 27|shared:pillars/two-categories-six|[YAML](slides/027-dxc-027.yaml)|
|28|Operating model 28|shared:pillars/two-categories-six|[YAML](slides/028-dxc-028.yaml)|
|29|Operating model 29|shared:pillars/two-categories-six|[YAML](slides/029-dxc-029.yaml)|
|30|Operating model 30|shared:pillars/two-categories-six|[YAML](slides/030-dxc-030.yaml)|
|31|Operating model 31|shared:pillars/two-categories-six|[YAML](slides/031-dxc-031.yaml)|
|32|Operating model 32|shared:runbook/roles|[YAML](slides/032-dxc-032.yaml)|
|33|Operating model 33|shared:divider/inverse|[YAML](slides/033-dxc-033.yaml)|
|34|Operating model 34|shared:runbook/roles|[YAML](slides/034-dxc-034.yaml)|
|35|Operating model 35|shared:runbook/step-table|[YAML](slides/035-dxc-035.yaml)|
|36|Operating model 36|local:role-ownership|[YAML](slides/036-dxc-036.yaml)|
|37|Operating model 37|shared:divider/panel-edge|[YAML](slides/037-dxc-037.yaml)|
|38|Operating model 38|shared:offers-onepager/nav|[YAML](slides/038-dxc-038.yaml)|
|39|Operating model 39|shared:offers-onepager/panel-right|[YAML](slides/039-dxc-039.yaml)|
|40|Operating model 40|shared:offers-onepager/panel-right|[YAML](slides/040-dxc-040.yaml)|
|41|Operating model 41|shared:offers-onepager/three-band|[YAML](slides/041-dxc-041.yaml)|
|42|Operating model 42|shared:understanding-situation/bands|[YAML](slides/042-dxc-042.yaml)|
|43|Operating model 43|shared:capability-heat/dense|[YAML](slides/043-dxc-043.yaml)|
|44|Operating model 44|shared:runbook/step-table|[YAML](slides/044-dxc-044.yaml)|
|45|Operating model 45|local:activation-chart|[YAML](slides/045-dxc-045.yaml)|
|46|Operating model 46|shared:workstreams/five-with-risks|[YAML](slides/046-dxc-046.yaml)|
|47|Operating model 47|shared:workstreams/five-with-risks|[YAML](slides/047-dxc-047.yaml)|
|48|Operating model 48|shared:activities/by-phase-table|[YAML](slides/048-dxc-048.yaml)|
|49|Operating model 49|shared:pillars/two-categories-six|[YAML](slides/049-dxc-049.yaml)|
|50|Operating model 50|shared:pillars/two-categories-six|[YAML](slides/050-dxc-050.yaml)|
|51|Operating model 51|shared:pillars/two-categories-six|[YAML](slides/051-dxc-051.yaml)|
|52|Operating model 52|shared:pillars/two-categories-six|[YAML](slides/052-dxc-052.yaml)|
|53|Operating model 53|shared:runbook/roles|[YAML](slides/053-dxc-053.yaml)|
|54|Operating model 54|local:role-ownership|[YAML](slides/054-dxc-054.yaml)|
|55|Operating model 55|shared:stats/four-metrics-right|[YAML](slides/055-dxc-055.yaml)|
|56|Operating model 56|shared:roadmap/grid-rows|[YAML](slides/056-dxc-056.yaml)|
|57|Operating model 57|shared:offers-onepager/nav|[YAML](slides/057-dxc-057.yaml)|
|58|Engineering 1|shared:cover/grid|[YAML](slides/058-dxc-engineering-findings-cover.yaml)|
|59|Engineering 2|shared:cards/narrative-2x3-tall|[YAML](slides/059-dxc-engineering-six-findings.yaml)|
|60|Engineering 3|shared:text/four-column-tall|[YAML](slides/060-dxc-engineering-four-moves.yaml)|
|61|Engineering 4|shared:method/accelerator-steps|[YAML](slides/061-dxc-core-seven-gates.yaml)|
|62|Engineering 5|shared:activities/by-phase-table|[YAML](slides/062-dxc-engineering-proposed-scorecard.yaml)|
|63|Engineering 6|shared:cards/narrative-bullets-2x2|[YAML](slides/063-dxc-engineering-next-phase-adds.yaml)|
|64|Engineering 7|shared:cover/grid|[YAML](slides/064-dxc-engineering-appendix.yaml)|
|65|Engineering 8|shared:method/accelerator-steps|[YAML](slides/065-dxc-core-current-seven-weak-points.yaml)|
|66|Engineering 9|shared:comms/plan-table|[YAML](slides/066-dxc-core-stabilization-action-plan.yaml)|
|67|Engineering 10|shared:capacity/tiers|[YAML](slides/067-dxc-claude-agile-ceremonies.yaml)|
|68|Engineering 11|shared:comms/plan-table|[YAML](slides/068-dxc-claude-agile-action-plan.yaml)|
|69|Engineering 12|shared:method/accelerator-steps|[YAML](slides/069-dxc-qa-seven-layers.yaml)|
|70|Engineering 13|shared:comms/plan-table|[YAML](slides/070-dxc-qa-action-plan.yaml)|
|71|Engineering 14|shared:activities/by-phase-table|[YAML](slides/071-dxc-engineering-ten-metrics.yaml)|
|72|Engineering 15|shared:runbook/step-table|[YAML](slides/072-dxc-engineering-findings-traceability.yaml)|
|73|Engineering 16|shared:roadmap/grid-rows|[YAML](slides/073-dxc-engineering-six-week-roadmap.yaml)|
|74|Engineering 17|shared:assumptions/detailed-impact|[YAML](slides/074-dxc-engineering-risks-dependencies.yaml)|
|75|Engineering 18|shared:guide/section-overview|[YAML](slides/075-dxc-engineering-internal-proposed-edits.yaml)|
