# Source-bound component editing

`pptxcomponent` exposes reviewed text slots in an extracted native scene project.
It is the first executable bridge from catalog references to editable components.
It retains the complete source slide and resources. Moving a component to a new
slide, changing its role count, and automatically proving text fit are future work.

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
Each slot lists specific text binding IDs, in source run/paragraph order. A value
is an array of text segments with exactly that cardinality. For example:

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

Omitted slots retain their source content. No runs, paragraphs, fonts, shapes,
geometry or line breaks are invented. Unknown slots, incorrect segment counts,
empty segments, control characters and stale scenes fail before output is written.
The output is a new scene project with a `component-application.json` change log.
To iterate, apply revised values to the pinned original project. Rebasing a
contract onto a returned/modified scene requires an explicit new review.

The engine also supports explicitly bound RGB color roles and complete named
profiles in reviewed contracts. A role can address a text, line or shape color;
`fill.srgbClr.val` is the low-level binding property for all three. Its precise
meaning must be established by inspecting the node path. Theme colors, inherited
fills and baked-in image colors cannot be recolored by this engine. **The four
initial contracts expose no style profiles.** The proposed library styles remain
pending inheritance resolution and visual proof.

`inspect` reports the contract and all explicit bindings for its selected objects.
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
