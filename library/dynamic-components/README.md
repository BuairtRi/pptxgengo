# Dynamic roles, pods and team compositions

This is the first dynamic implementation of the user's P1 direction. A pod accepts
an arbitrary nonempty `roles[]` list, with stable IDs independent of its labels.
Native PowerPoint rectangles and text remain editable. This is a bounded engineering
implementation, not general approval of the component catalog or a pixel-perfect
reconstruction of UHG43.

## Workflow

For generation without PowerPoint, see the opt-in
[pure Go font layout prototype](go-layout.md) and its
[font layout example](go-font-layout.json). Go measurement results are
predictions; native final verification remains a separate command.
The [font layout reference set](font-layout-reference.md) supplies the frozen
IBM Plex Sans, Arial and IBM Plex Mono corpus and PowerPoint PDF comparison.
The [v2 wrapping calibration](font-wrap-calibration.md) records the shared
advance rule, additional comparisons and boundary warnings.
The [v3 height calibration](font-height-calibration.md) records native character
height controls, PDF baseline spacing comparisons and height review warnings.

Run from the repository root on macOS with Microsoft PowerPoint installed:

```sh
go build -o /tmp/pptxcompose ./cmd/pptxcompose
/tmp/pptxcompose probe --spec library/dynamic-components/pods.json --out samples/pod-probes
/tmp/pptxcompose measure --bundle samples/pod-probes --out samples/pod-evidence.json
/tmp/pptxcompose build --spec library/dynamic-components/pods.json --bundle samples/pod-probes --evidence samples/pod-evidence.json --out samples/pod-output
/tmp/pptxcompose verify --bundle samples/pod-output --out samples/pod-verification.json
```

Every output is new-only. Bundle directories contain a PPTX named after the
directory and a `manifest.json`. Close a presentation with the same filename before
measuring: the CLI refuses an already-open copy, which may contain unsaved edits.
The native adapter reads and leaves the generated presentation open. It does not
save source decks. Supply `--adapter` if running outside the repository root.

1. **Probe:** validates the spec, resolves semantic colors/contrast and emits an
   isolated text shape for every feasible role width and title. Fonts, weight,
   paragraph spacing, insets and wrapping are explicit.
2. **Measure:** opens the generated probe deck in PowerPoint; measures the union
   of native non-whitespace character bounds (retaining whole-range bounds for diagnosis); checks text, actual font and known shape frames; records raw native
   measurements and hashes of the spec, PPTX and manifest. Known frame dimensions
   calibrate the returned geometry to points.
3. **Build:** requires matching evidence, reconstructs dimensions from its raw
   native records, and computes each row from its tallest role. Auto mode selects
   the fewest columns that fit. Explicit columns remain a constraint. No text
   shortening, font shrinking or character-count fit estimate is used.
4. **Verify:** opens the final deck and checks actual frames, font, text and text
   bounds against inner safe zones, plus native fills, text colors and margins.
   A failed check retains a `.failed.json`
   diagnostic. Then export through PowerPoint and inspect each slide visually;
   passing measurement alone does not establish design quality.

Evidence guards against accidental stale inputs; it is not a cryptographically
attested measurement service. Use the shipped adapter and retain the generated
files. A custom adapter is a trusted local integration.

## Geometry and style contract

- All input/output geometry is in points, relative to the slide's top-left corner.
- Row-major ordering preserves roles as declared. An uneven final row starts at
  the left. Pod height is content-dependent, capped by `max_height_pt`.
- The plan records outer bounds and five midpoint anchors for each role and pod.
  The optional [team composition fields](team.md) add standalone roles, shared
  phase backgrounds, a derived staffing legend and obstacle-aware reporting lines.
  Native line segments are editable but do not reroute when manually moved.
- `staffing.wm_full_time` is navy `#070154`; `staffing.wm_part_time` is blue
  `#0047FF`; `staffing.client_part_time` is magenta `#F900D3`. These tokens encode
  staffing meaning and must be accompanied by a legend in a client-facing deck.
- `foreground: auto` chooses the higher-contrast navy or white. Explicit colors
  are checked too. A 4.5:1 minimum is an engineering policy for this component,
  using the [W3C relative-luminance formula](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html).
- Typography requires explicit font families, point sizes and weights, and zero
  pod `paragraph_gap_pt`. Installed families such as IBM Plex Sans are supported;
  see [font selection and measurement](fonts.md). Rich canvas/layout paragraphs
  have explicit per-run fonts. General inheritance from arbitrary source
  masters, mixed-format pod labels and text rotation remain unsupported. Negative geometry, duplicate IDs,
  pod collisions and text that cannot fit are errors.
- The provided 12pt role and 14pt pod-title styles are proposed fixture dimensions.
  Original UHG43 uses smaller 9pt roles; these examples are not source-sized clones.

The fixture is illustrative, with no client staffing commitments. It covers role
counts 1/2/3/5/8, repeated labels with distinct IDs, semantic color choices, one and
two columns, auto layout under a height constraint, and wrapped role/pod titles.

The [P1 source evidence](source-reference.json) retains the original XML and source
hash for the container and two roles. [Native QA results](proof-report.json)
record the exact reviewed artifacts.

## Team composition fixture

Use [team.json](team.json) in the same workflow for two additional examples: a
three-pod reporting tree with a shared phase surface, and a reporting path around
standalone specialists. See [the team contract and limits](team.md) and
[its separate native proof](team-proof.json). The earlier pod proof remains a
historical record of its exact code and measurement adapter.

## Additional authoring components

[Numbered rows and metrics](cards.md) add measured card contracts with light/dark
profiles. [Canvas elements and accents](canvas.md) support varied new layouts,
pinned PNG/JPEG assets, measured whole-text emphasis, and bounded recovery of
colleague text edits. The [showcase](../showcase/README.md) combines these paths
into a proposal-style review deck.

- [Fresh native font calibration](font-native-calibration.md): 393 validated
  PowerPoint probes, v4 line allocation corrections, and the focused follow-up.

- [Focused font results](font-native-focused-results.md): 45 additional controls
  and successful native verification of the two-slide Go example.
