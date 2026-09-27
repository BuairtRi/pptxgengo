# Source-bound component editing

`pptxcomponent` exposes reviewed text slots, color roles, and bounded empty text
zones in an extracted native scene project. It retains the complete source slide
and resources. It does not prove text fit.

## Example

Run from the repository root. All output paths must be new.

```sh
go run ./cmd/pptxscene extract \
  --source 'samples/software modernization campaign pick deck v1 - Repaired.pptx' \
  --slides 57 --out samples/my-metric-source

go run ./cmd/pptxcomponent inspect \
  --project samples/my-metric-source \
  --contract library/component-contracts/metric-card.json

go run ./cmd/pptxcomponent apply \
  --project samples/my-metric-source \
  --contract library/component-contracts/metric-card.json \
  --values library/component-contracts/metric-card.example.json \
  --out samples/my-metric-edited

go run ./cmd/pptxscene build --project samples/my-metric-edited \
  --out samples/my-metric.pptx --freeze-slide-numbers
```

The example copy is a synthetic edit fixture, not approved client evidence.

## Contract and values

Four reviewed contracts bind M1, N1, P1 and P2 from
[the reference shortlist](../../library/reference-shortlist.json).
Extract source slide 109 from Graphics and Layouts for N1, and source slide 43
from UHG for P1/P2. Source slide positions differ from printed footer numbers in
the modernization deck.

Contracts pin the source package identity and the exact serialized source scene.
Each slot lists specific text binding IDs, in source run/paragraph order. A legacy
value is an array of text segments with exactly that cardinality. For example:

```json
{
  "slots": {
    "key_point_heading": ["ALIGN ON VALUE"],
    "bullet_copy": [
      "Confirm target outcomes",
      "Agree success measures",
      "Prioritize the roadmap"
    ],
    "number": ["01"]
  }
}
```

For a rich text slot, use the explicit paragraph model returned by `inspect` as
`slot_paragraphs`. Each paragraph contains ordered runs and their pinned binding
IDs. Supply one text value per run; the engine preserves each source run's style:

```json
{
  "slots": {
    "headline": {
      "paragraphs": [
        {"runs": [
          {"binding_id": "b/4/text/0.0.4.2.2.0.1.0/text", "text": "Plan "},
          {"binding_id": "b/4/text/0.0.4.2.2.1.1.0/text", "text": "delivery"}
        ]}
      ]
    }
  }
}
```

The binding IDs above illustrate the JSON shape; use actual IDs from `inspect`.
Set `"value_format": "paragraphs"` on a reviewed slot to require this format and
reject legacy arrays for that slot.
Omitted slots retain their source content. Unknown slots, incorrect paragraph or
run counts, reordered IDs, empty text, control characters, and stale scenes fail
before output is written. Text is never distributed across runs automatically.
The output is a new scene project with a `component-application.json` change log.
To iterate, apply revised values to the pinned original project. Rebasing a
contract onto a returned/modified scene requires an explicit new review.

Reviewed color roles may bind `fill.srgbClr.val` (`color_kind: "rgb"`, the default)
or `fill.schemeClr.val` (`color_kind: "scheme"`). A named profile must supply every
role. RGB values are six hex digits; scheme values are supported theme tokens such
as `accent1`, `tx2`, and `bg1`. A `source` profile for a scheme role must exactly
match its pinned source tokens. Roles address exact node paths because an object
ID alone may repeat in alternate content. A role may also contain
`resource_targets` for retained XML parts such as chart user shapes. Each target
pins a local `resources/` part, its SHA-256, exact node path, and source color
value. The engine changes only that color in the copied output project and logs
the input and output resource hashes. Unbound chart XML, images, and other uses
of the theme color stay untouched.

An optional contract `zones` map can name an existing empty `a:p` in a selected
shape's text body via `paragraph_path`. Each zone pins a same-slide text binding as
`style_donor_binding_id`, or supplies a complete explicit `style` with
`font_size` (hundredths of a point), `typeface`, and `color_rgb`. A donor may have
bounded style overrides. Values use `"zones": {"name": "New text"}`. The engine
rejects a nonempty target and retains original binding IDs while updating paths
shifted by the inserted run. For freeforms whose text rotates with the artwork,
`mode: "overlay_textbox"` clones a pinned same-slide donor text shape, gives it a
unique `overlay_object_id`, and places it in an explicit four-value `frame_emu`
inside the empty target shape's source bounds. The original shape and paragraph
remain intact. Created zones and overlay frames are recorded in the application
log. Native rendering must confirm visual fit.

`inspect` reports the contract, paragraph/run model, and all explicit bindings for
its selected objects.
These include typography and inset declarations, which can contain unused list
levels and theme references. They are not resolved effective typography.

## Fit and promotion

Successful application proves bounded native text replacement. Every application
reports `requires_native_measurement_and_visual_review` and
`adaptation_approved: false`. Character counts are not fit limits. PowerPoint may
wrap, overflow, shrink or substitute fonts. Render and measure the resulting
content in its actual source container before using it in a deliverable.

The source hash in the project manifest is provenance; it does not revalidate
every retained resource against the original package. Use an unmodified project
extracted from the registered source. Scene hash validation protects the reviewed
slot mapping. Contracts are reviewed repository inputs, not untrusted templates.
