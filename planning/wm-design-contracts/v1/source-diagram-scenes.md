# Native source diagrams, sequences and people

## Executable scene planners

The Go source-scene adapter implements `block`, `frame`, `chevron`, `textarrow`, `connector`, `container`, `cylinder`, `node`, `layerrow`, `matrix`, `beforeafter`, `stepper`, `vstepper`, `phasehead`, `phases`, `timeaxis`, `gantt`, `swimlane`, `pyramid`, `legend`, `pod`, `person`, `role`, `orgchart`, and `governance`.

The three handler entry points accept the verified source node, persistent key overlay, surface and an owned zone. They return an ordered native plan; they do not mutate the frozen snapshot. Shapes use editable DrawingML geometry, text uses the v2 Go engine and the normal IBM font names, photos use the pinned media helper, and recursively delegated bullet plans retain nested native group ownership. Reporting and connector paths are native line geometry; attachments to other shapes are not asserted.

Unknown render fields, enums, palette references, duplicate keys, invalid source IDs, intervals, unsupported marker shapes and text fit failures return errors. Content arrays use exact source JSON pointer keys; source references without an overlay have source ordinal identities. Gantt gate `key` is the frozen boolean emphasis flag and remains compatible with the source representation; a string key may supply identity instead. A typed binder overlay provides gate identity when the emphasis flag is present.

## Source contexts and layout

Frozen coordinates and local spacing define geometry. The adapter explicitly applies source typography contexts: semibold node/role/person/step titles, semibold small org roles, 12.5pt Gantt lane headings, 7.5pt period sublabels, 8pt tracked rotated group labels, Mono body funnel descriptions, and the source chip/label contexts. These are implementation contexts awaiting native qualification; implementation is not a typography qualification claim.

Source flex stacks are replaced with measured Go layouts. Text is rejected if it cannot fit its owned capacity; no font shrinking, silent text clipping or source content deletion is performed. The cylinder native `can` preset preserves its source outer rectangle; its rim representation is the native Office can geometry rather than the source SVG arc serialization. Matrix cell surfaces and outlines retain their explicit source dimensions.

Gantt item labels use shaped advances, replacing the source browser `tw` approximation. Label occupancy participates in deterministic stable track packing; each lane has at least42pt height and additional tracks consume30pt. Progress, soft boundaries and tentative intervals use native diagonal hatch strokes. Bar/event/tag labels must remain in the timeline, and body/legend bounds must remain in the supplied owned zone. The derived legend uses measured horizontal wrapping. Native group labels and governance escalation labels use uniform text rotation.

Swimlane paths retain the source routing rules and endpoint IDs. A source route crossing an unrelated step is rejected with the endpoint and obstacle identities; it is not silently rerouted. Public coordinate connectors retain their supplied path and include a warning that automatic obstacle routing and attachment after editing are not claimed.

## Remaining source resolutions and qualification

The source `plan/gantt` has a known body/legend envelope conflict before any Go metrics are applied. Measured packing may expose additional overflow on long late-period labels. These errors require an explicit template contract resolution or shorter authored content; a wider or taller invisible container is not an accepted correction. Source phase columns, narrow governance chips/decision columns, role boxes, person rows and compact cylinder labels remain subject to measured fit.

`orgchart` follows the live frozen renderer's162×54pt role boxes and108pt vertical pitch, resolving the previously recorded198pt metadata conflict in favor of executable source geometry. Heading and body scope exceptions such as an absent-header phasehead at y54 remain compiler/frame responsibilities.

`DiagramSceneReference(year)` provides four focused synthetic pages covering editable diagram geometry, keyed sequences, reporting lines and measured Gantt packing. The parent compiler's full family reference provides frozen template coverage. No tests, native capture or qualification status changes are part of this implementation.
