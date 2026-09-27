# Wave 3 diagram foundation

The compose plan now retains measured container and cell geometry as `layout_ports`. A port ID is the container ID or `container/cell`; each entry records slide-space bounds and `top`, `bottom`, `left`, `right`, and `center` points. Nested panel translation is resolved before these coordinates are published. Unknown connection targets and unknown edge names fail validation.

`connections` accept `reporting`, `flow`, `dependency`, `advisory`, and `annotation`. Reporting remains acyclic and renders as a solid line. Dependency renders as a solid line with a terminal triangle. Advisory uses dashes; annotation uses dots. Non-flow relations use deterministic orthogonal editable line segments. They route around measured text and image frames and use `clearance_pt`, `preferred_direction`, `bend_penalty_pt`, `from_offset_pt`, and `to_offset_pt`. The default bend penalty is 24 pt. The route plan records `endpoint_behavior: "editable_segment"`; moving a shape in PowerPoint does not automatically reroute its lines.

A direct `flow` relation requires aligned `right` and `left` ports. It creates a native editable `rightArrow` in the declared gap and records a semantic connection with `strategy: "direct_right_arrow"`. `gap_margin_pt` defaults to 2 pt and `arrow_height_pt` to 13.724 pt. The planner rejects a gap too small for the arrow, overlap with measured text/image or another arrow, and nonaligned flow ports. The arrow is a separate editable shape, not an attached PowerPoint connector.

Example connection between two measured layout cells:

```json
{
  "id": "step-1-to-2",
  "from": "process/step-1",
  "to": "process/step-2",
  "relationship": "flow",
  "from_anchor": "right",
  "to_anchor": "left",
  "color": "#070154",
  "width_pt": 1,
  "clearance_pt": 2,
  "gap_margin_pt": 2,
  "arrow_height_pt": 13.724
}
```

The `paths` primitive accepts an ordered `labels` array, bounds, gap, typography, node colors and arrow settings. It expands to a measured one-row `ContainerSpec` with equal-width cells and explicit flow connections. A `parent_id` nests it within a panel. This can represent two parallel process rows without individual cell coordinates. The UHG36 reference comparison and native visual gate remain separate from these geometry tests.

Uniform canvas text uses a 3:1 contrast threshold at 18 pt regular or 14 pt bold, and 4.5:1 otherwise, following [WCAG 2.2 Understanding Contrast (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html). This preserves the source UHG36 decorative number pink `#F900D3` at 16 pt bold; it does not certify the whole source slide against WCAG.
