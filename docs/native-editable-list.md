# Native unordered lists

`wmds/component/editable-list` places a complete flat unordered list in one
native PowerPoint text box. Every item is a paragraph with a styled square
bullet, hanging indent, source text typeface, body/small size and paragraph spacing.
Wrapping stays within that paragraph; the following paragraphs flow with it.
Markers are native bullet properties, not separate shapes or typed characters.

## Authoring

Use this component for new plain lists intended for PowerPoint editing:

```yaml
nodes:
  - id: review-list
    kind: component
    definition: {scope: shared, id: wmds/component/editable-list}
    placement:
      zone: body
      span: {start: 1, count: 6, y: 36, h: 216}
    arguments:
      size: body
      items:
        - Review the source
        - Confirm the decisions with the team
        - Share the deck
    keys:
      items: [review, confirm, share]
```

The existing shared frame/grid pins and normal local-template compilation
apply. A source scene can use `type: editable-list` with explicit `x/y/w/h`.
Items must be nonempty plain strings, at most 100. Nested lists, ordered lists,
lead/body objects, rich markup, tabs, newlines and footnotes are rejected.
`size` is `body` (default) or `small`. Stable authored item keys are required
when the component participates in keyed project compilation.

## Editing and source review

The box has zero text insets, native square bullets with explicit color/font,
EMU paragraph indents, and fixed font sizes (`a:noAutofit`). The compiler
measures initial wrapping and rejects overflow. It emits no hard wrap breaks.
PowerPoint owns subsequent paragraph reflow; longer or added content may
require increasing the one text box's height. Automatic font shrinking and
automatic box growth are not enabled.

The marker uses Wingdings' small solid square (U+F0A7), with its font size
calibrated to the source marker's ink width. IBM Plex supplies the text; it has
no square glyph and must not be relied on for the bullet. Wingdings is a desktop
Office prerequisite and is not redistributed. The last paragraph retains the
same after-spacing so pressing Enter inherits a consistent gap; trailing spacing
is not included in the painted initial fit. The Mac comparison covers all three
densities; Windows and non-Office consumers remain unqualified.

The maintained generation identity belongs to the text box. Individual
paragraphs do not contain durable item identities. Changing text, adding,
deleting or reordering bullets, and changing styles therefore require manual
source review. The CLI does not infer authored keys or silently adopt these
changes into YAML. Preserve an edited copy and reconcile it against its exact
receipt-pinned baseline.

Existing stock `bullets`, ordered and rich/decorative list renderers remain
available. This authored component is the authoring preference for new plain
lists. Separately, the v4.2.1 `native-v1` profile converts eligible stock flat
unordered lists, including supported measured lead/body and inline rich runs.
Nested, ordered, decorative or otherwise unsupported structures retain their
scene output with a reason. See the [catalog profile](native-editing-profile.md).
The profile's source-plan matching does not qualify desktop behavior on Windows.

## Qualification

Hermetic tests inspect actual generated PPTX structure: one top-level text box,
three bullet paragraphs, zero insets, explicit bullet styling, no hard breaks,
no autofit/shrink and retained shape tags through synthetic paragraph edits.
Planner tests cover all three densities and both text sizes, wrapped initial
copy and invalid/nonfitting inputs. These are structural checks, not evidence
of a user editing successfully in PowerPoint.

Prepare a synthetic fixture in a new private Documents directory:

```sh
PPTXGENGO_EDITABLE_LIST_PILOT_OUT="$HOME/Documents/pptxgengo-qualification/native-list-new" \
  go test -count=1 -timeout=3m -run '^TestNativeEditabilityEditableListActualPackage$' -v ./internal/deckproject
```

The parent directory must exist and the destination must be new. The fixture
contains immutable build evidence, source and hashes. Preparation opens no app.
Ordinary Make test lanes clear this opt-in output variable. macOS and Windows
qualification must check wrapping alignment, pressing Enter to add a bullet,
paragraph spacing, fixed fonts, resizing and Save As identity retention.

The 2026-10-07 [comparison demo](native-component-demo.md) records real Mac
wrapping and insertion: a two-line item became four lines, both following items
flowed down, and Enter produced a fourth native square-bullet paragraph. The
saved package retained every generation identity. Resizing remains pending.

The 2026-10-07 synthetic fixture's 15 retained files passed independent SHA256
verification. Its generated PPTX is
`23ebe2988583e6171d50cd9f240077e01a4d9c25194bde1321201083538e7a01`:
one list shape, three paragraphs and group depth zero. Computer Use rejected
PowerPoint access before opening the fixture, so no live editing result is
recorded. Native font glyph dimensions may differ from the older geometric
square markers even when their nominal point sizes match.
