# Wave 3 first slice: measured process paths with container anchors

## Recommendation

Implement one bounded reusable **process path** slice from UHG slide 36: a measured row of labeled nodes with explicit right-arrow connectors, nested in a parent panel. Use UHG36 as the source-control geometry and content reference, then exercise the same component with changed synthetic content and cardinality/fit stress. Use UHG14 only as a companion container-positioning control: its architecture board is a raster image and its nearby curved/dotted arrow is pinned SVG artwork, not an editable semantic diagram. Do not present an editable redraw as a faithful extraction of UHG14.

## Source evidence

Both reference slides are from `UHG Fabric Platforming RFP Response - July 2026.pptx` (SHA-256 `b0f254ed7739768d0f345257689264393049d06cdd3849a177fc764f3d348d99`).

- **UHG36**, `samples/reconstruction/reference/png/slide-036.png`; source XML `ppt/slides/slide36.xml`; sidecar `samples/inspection/uhg/sidecars/slide-036.md`. Its main grouped panel contains two five-node horizontal paths, four right-arrow preset shapes per path, headings, and a result band. Resolved node frames are approximately 105.185 × 60.996 pt; left edges are 324.210, 448.409, 572.607, 696.806, and 821.004 pt. Row tops are 169.039 and 272.268 pt. The roughly 19 pt gaps leave about 3.3 pt around each 12.375 × 13.724 pt arrow. The full source group uses anisotropic scaling, so that group transform must not be mistaken for a safe text-preserving resize rule.
- **UHG14**, `samples/reconstruction/reference/png/slide-014.png`; source XML `ppt/slides/slide14.xml`; sidecar `samples/inspection/uhg/sidecars/slide-014.md`. The large architecture board is `Picture15` (`ppt/media/image124.png`) at about `[471.29,58.34,379.34,404.35]` pt. The small preview is `Picture11` (`image123.png`) at about `[819.07,22.78,113.06,120.04]` pt. `Graphic 17` references `ppt/media/image86.svg` at about `[852.18,132.80,70.29,70.29]` pt. The diagram labels are baked into the image; no internal editable node structure or semantic arrow endpoint is available from this slide.
- The existing catalog seed `component:slice-proposal-uhg-governance-pathway` identifies UHG36's two pathway instances and the evidence PNG `samples/component-expansion/proposals/uhg/slide-036.png` (SHA-256 `276d4c9d6f564132520a2dc45bcb25a872fcfe241300d6da8de028997310d30f`). Its current slots (`path_title`, `result`, `monitoring_step`) are preservation-only and do not yet define node/connector roles or dynamic capacity.

## Smallest API additions

Current `ContainerSpec`/`CellSpec` grids already support nesting, padding, weighted/fixed tracks, row rules, and measured text-fit reporting. Layout lowering resolves parent-relative geometry, but does not expose stable resolved anchors for containers/cells. Current connections target only root roles or whole pods, accept only the `reporting` relationship, and use orthogonal line routes. Canvas lines are straight horizontal/vertical rules; no editable arrow preset or arrowhead is emitted.

Add only the following for this slice:

1. **Resolved layout ports.** Give a container and cell stable IDs plus named edge ports (`left`, `right`, `top`, `bottom`, optionally centered). Resolve these after measured layout and parent translation; report their final slide-space coordinates in the plan/fit evidence. Reject unknown targets and ports.
2. **Direct flow connectors.** Add a typed `flow` connector that can target those ports, with native editable right-arrow geometry. First-slice routing is direct edge-to-edge within a declared inter-node gap; it must fail clearly when the gap cannot hold the declared arrow/margins. Keep general obstacle routing, curves, and SVG-like hand-drawn paths out of scope.
3. **Path row primitive.** A path is an ordered node list plus explicit inter-node connectors, contained by a parent panel. Node count and content may vary; geometry is resolved from measured node widths/gaps and returns per-node fit evidence. Parent translation moves nodes and connectors together.

Keep the source-control UHG36 typography, colors, and geometry fixed to observed values. A changed-content instance may use the same five-node/two-row structure with clearly synthetic labels and a clearly illustrative result; do not invent business outcomes. UHG14's PNG/SVG controls may test pinned-art placement inside a translated container, but must remain identified as raster/vector artwork with no inferred semantic links.

## Required fixture set and acceptance evidence

1. **Source control:** reproduce the two selected UHG36 paths and result area with unchanged source text/style/geometry; compare to its reference render and retain source-object provenance. This is source-instance evidence, not proof of broad component approval.
2. **Changed-content fixture:** two five-step synthetic paths in the same bounded component, with neutral process wording and a plainly labeled illustrative result. Keep source and synthetic content visibly distinguished in fixture metadata and slide labeling.
3. **Stress fixtures:** exercise both four- and six-node paths, a long node label, and parent translation. Each must either fit using measured geometry or fail with an explicit overflow/gap/fit reason—never silently shrink fonts. Verify connectors stay centered in their declared gaps and attach to resolved node ports after translation.
4. **UHG14 positioning control:** retain its pinned image and SVG asset identities/hashes, test only whole-container translation/placement bounds, and state that the board's internal labels/arrows are not editable or semantically anchored.

Acceptance is based on source comparison, measured fit reports, connector endpoints/gap clearance, and translation invariance. It does not equate a passing geometry check with visual quality or general template approval.
