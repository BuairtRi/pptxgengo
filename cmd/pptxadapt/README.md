# Semantic slide-family adaptation (development)

`pptxadapt` compiles content relationships into a standard measured `pptxcompose`
spec. Imported source templates remain a separate fixed-geometry editing route.
The compiler never estimates text capacity and does not mutate source decks.

```sh
go build -o /tmp/pptxadapt-dev ./cmd/pptxadapt
go build -o /tmp/pptxcompose-adaptive-dev ./cmd/pptxcompose
/tmp/pptxadapt-dev capabilities --id t020-graphics-and-layouts-062
/tmp/pptxadapt-dev compile --spec library/adaptive/architecture-standard.json \
  --out /path/to/new-compilation
```

A compilation bundle contains the exact semantic input, `spec.json`, a report
of resolved controls and limits, and SHA-256 identities. Outputs must be new.
A successful compilation validates structure, not native fit or visual quality.

## Native workflow

Use one fixed compose binary throughout a measurement run. Keep task outputs
separate from the frozen installed release. Reuse the established native folder:

```sh
/tmp/pptxcompose-adaptive-dev probe --spec /path/to/new-compilation/spec.json \
  --cache /path/to/native-cache --out /path/to/new-probe
/tmp/pptxcompose-adaptive-dev measure --bundle /path/to/new-probe \
  --cache /path/to/native-cache \
  --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3 \
  --out /path/to/new-evidence.json
/tmp/pptxcompose-adaptive-dev fit-report --spec /path/to/new-compilation/spec.json \
  --cache /path/to/native-cache --out /path/to/new-fit.json
/tmp/pptxcompose-adaptive-dev build --spec /path/to/new-compilation/spec.json \
  --cache /path/to/native-cache --out /path/to/new-deck
/tmp/pptxcompose-adaptive-dev verify --bundle /path/to/new-deck \
  --native-workspace /Users/rscott/Projects/pptxgengo/samples/visual-wave3 \
  --out /path/to/new-verification.json
```

Skip measurement when probe reports that all requests are cached and creates no
PPTX. A written fit report is not necessarily a pass: check overflow, layout
failures and planner errors. Native operations are serial. Leave user decks alone.
Export the final task deck, inspect every page, and record findings against the
exact spec and render hashes.

## Shared content and style contract

Each adaptive slide has `id`, `family`, `title`, `role`, `takeaway`, `content`, and
optional `template_id`, `style`, `bounds`, and `footer`. Empty footer text is omitted. The template ID documents the reference
pattern; it does not cause source OOXML import or promise pixel identity.

Families: `roadmap`, `architecture`, `process`, `team`, `comparison`. The explicit
content schemas and examples live in `library/adaptive/`. All provided claims,
roles and numeric values are synthetic capability examples.

Style profiles are `neutral`, `subtle`, `inverse`. Explicit point sizes are
`title_font_pt` (18–36), `body_font_pt` (9–18), `label_font_pt` (9–18). These are
input ranges, not capacity guarantees. Use `font_face` for any installed explicit family, for example `IBM Plex Sans`.
Arial remains the default. Native measurement checks family/style availability
and fingerprints the selected font files; see [font support](../../library/dynamic-components/fonts.md).
The compiler never reduces a font size automatically to hide overflow.

`style.colors` accepts semantic roles: `text.primary`, `text.secondary`,
`surface`, `surface.muted`, `accent`, `state.active`, `state.inactive`, `border`.
Values are explicit `#RRGGBB`. Default colors follow the current WM palette;
custom values are caller choices and receive contrast checks, not brand approval.
Surface-specific foregrounds resolve to readable navy/white where required.
Staffing tokens retain their existing ownership meaning and require a legend.

`bounds` is the body region in points, inside x=40..920 and y=115..485 on a
960×540 slide. Moving or resizing it moves/recomputes the family geometry. The
shared heading, logo and footer remain in the surrounding chrome.

The capability inventory distinguishes the 65 original editing contracts from
family category hints. A category hint requires structural inspection; new content
still requires measurement and review. Consult the current checkpoint before claiming
any family or variant has completed native review.

The source dispatcher now routes `pptxgengo adapt` to this CLI for the next packaged
release. The currently installed 0.1.0-local.3 remains unchanged and does not
include this development command.
