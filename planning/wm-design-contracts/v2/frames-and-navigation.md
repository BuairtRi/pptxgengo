# Refreshed source contract: split frames and navigation

Source revision: `wmds-library.v2`. Typography remains the existing candidate Go
engine; no font alias or per-deck character capture is introduced.

## Source and bundle selection

Both frozen v1 and v2 bundle/inventory digests are registered. Loading checks the
matching pair, every frozen source file and every packaged font/logo asset.
An override must match the selected revision. Source revision/commit are emitted
in inspection and layout reports; bindings identify the revision as well as the
source hash. Unknown source fields remain errors.

Lifecycle fields `rev`, `added`, `revised`, `status` and `replacedBy` are consumed
and returned by the library catalog. CLI catalog selection excludes deprecated
entries by default; `--include-deprecated` shows them. Explicit bindings retain
access to the deprecated `from-to/rows`; its replacement is reported.

## Split frame geometry

`FrameRequest.split` accepts exactly the following values, from the pinned source:

| Split | Short column | Tall column |
|---|---|---|
| tall-right | 57–327 | 345–903 |
| tall-left | 633–903 | 57–615 |
| tall-right-narrow | 57–615 | 633–903 |
| tall-left-narrow | 345–903 | 57–327 |

All units are points. The gutter is 18. Tall content starts at y 36. Split is
allowed with no rail or nav; panel rails, appendix density and headerless split
requests are rejected. Compact/tall footers and zero/one/two source lines retain
their declared body-bottom reservation.

A 270-wide short column uses heading 24/30 with rules at 108/126/162 for one/two/
three title lines. A 558-wide short column uses title 32/36 with rules at
108/144/180. Short body starts 18 below the rule. Text does not shrink to fit.

Resolved frames report `header`, `short_body` and `tall_body`. `body` is their
enclosing content area and does not authorize content in the gutter or short-column
header. Direct native nodes may name `scope: "short"` or `"tall"`; otherwise their
horizontal span selects a zone. Source scene bounds must fit one content zone.
Declared media retains the existing full-canvas bleed policy. Coordinate-only
connectors are planned from their actual points, then checked against split zones.

Title, eyebrow, header rule, default whiteboard patch, stamp and source use the
short column. Footer rule, legal line, logo and slide number keep the full frame
width and ordinary footer locations. The default dot patch starts 36 left of the
short column, bounded at x 3 (or 57 for nav).

## Navigation content

Source `slide.nav` uses labels and a zero-based active index. The compiler converts
that to keyed native tabs. Caller bindings use the following content-only object
alongside the existing required `slots` and `keys`:

```json
{
  "slots": {"...": "all declared scalar values"},
  "keys": {"...": ["all declared array keys"]},
  "nav": {
    "items": [
      {"key": "context", "label": "Context"},
      {"key": "solution", "label": "Solution"},
      {"key": "delivery", "label": "Delivery"}
    ],
    "active": "solution"
  }
}
```

Nav is required for a nav template and forbidden for other templates. There must
be 2–6 tabs, unique valid keys, nonempty single-line labels, and an active key in
the list. Array order controls position; keys control shape identity. Nav count is
independent of the fixture's original four tabs. Labels are measured in the
rotated native tab capacity; overflow fails rather than shrinking. No sample
labels are supplied for missing bound content.

Tabs remain x 21–39 with 6 pt gaps. Height is
`(footerBodyBottom - 36 - 6*(count-1))/count`, independent of source-line reservation.
Active fill/text are Grounded/White, inactive fill/text Light Gray/Dark Gray.
V2 tabs use the pinned renderer's dedicated `.gtab span` typography: Plex Mono
8 pt, weight 600 and 0.1em tracking (0.8 pt), rather than the generic 9 pt label.
The CSS has no padding; the native fit check reserves2 pt at each rotated end.
The old v1 native-tab behavior remains tied to the old bundle.

Binding assignments record caller labels, selected key, source pointers and the
key overlay. Native shapes are named `nav.<key>` and `nav.label.<key>`.

## Revision handling and current boundary

The revised agenda already authors an offset panel/photo, so the v1 additional
photo refinement is skipped for v2. The revised pillars already author their
hand-drawn arrows, so the v1 connector-replacement refinement is also skipped.
Other unchanged templates retain their named v1 refinements.

The unchanged typed `cards/3` and `cards/4` APIs are available in either revision.
Slice 2 implements revised metric contracts, table checkbox, quadrant marker and
named annotations; see [component contracts](components-and-revised-bindings.md).
Unsupported source fields still fail. Catalog availability remains separate from
manual specimen review and whole-library qualification.

## Reference artifacts

`frame-reference` generates 12 controls covering all four splits, both footers,
one/two/three-line titles, source reservation, 2/6 navigation tabs and split + nav.
Five focused library bindings cover agenda, key-message/stat, roster, case-study
exhibit and portrait/nav. A sixth bound example supplies six caller tabs and a
different active key. These form the 18-slide review specimen.

No tests were added or run. Compilation, artifact generation and native review
are recorded separately from a reusable content qualification envelope.
