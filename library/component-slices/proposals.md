# Proposal component slice

13 source patterns; 40 examples (UHG 25, EnableComp 15). All selected source pages
were inspected by the primary agent in native PowerPoint previews. Source geometry
and selected text/image slots pass integrated ingestion. No adaptation is approved.

| Source pattern | Source | Slide | Examples |
|---|---|---:|---:|
| Platform data-layer module | uhg | 26 | 1 |
| Platform operations service module | uhg | 26 | 1 |
| Platform control-domain module | uhg | 26 | 1 |
| Evidence gate card | uhg | 33 | 5 |
| Team role node | uhg | 43 | 6 |
| Staffing availability key | uhg | 43 | 1 |
| Image-caption collage | uhg | 35 | 1 |
| Phase one process card | enablecomp | 28 | 4 |
| Sample deliverable card | enablecomp | 29 | 4 |
| ProcessIQ stage card | enablecomp | 23 | 7 |
| Accelerator benefit card | uhg | 31 | 4 |
| Roadmap phase bar | uhg | 38 | 3 |
| Governance workflow pathway | uhg | 36 | 2 |

## Corrections and constraints

- Removed draft native-table text slots: tables need a separate cell contract.
- Removed UHG29 groups that were mistaken for content panels; the selected groups
  were a screenshot frame and empty-text chevrons.
- Removed the UHG37 phase ribbon mislabeled as an operating handoff and standalone
  EnableComp30 labels mislabeled as full stages.
- EnableComp23 now pairs each numbered box (paths 9–15) with its stage caption
  (17–23), preserving stage color semantics.
- UHG35 is a three-screenshot collage with an italic caption, not a process node.
- UHG26 module colors are affected by a slide-wide translucent overlay. Isolating
  those groups requires resolving that context and the surrounding headers.
- UHG43 color and marker meanings distinguish staffing/ownership. Shared connectors
  and phase backgrounds are not owned by individual role nodes.
- Shared backdrops, relationship arrows and source commercial/content claims
  remain subject to layout dependency and content approval gates.

Every instance has source paths, preview hashes and retained constraints in the JSON.
