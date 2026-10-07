# Native component comparison and release review

## Release scope, 2026-10-07

The next release focuses on enhancements already implemented, portable deck
projects, and native editing ergonomics. Reusable-slide inventory, revision
metadata, production browsing-library distribution, and the authoritative
branding/graphics/photo distribution bundle are deferred to the following
release. New branding inputs are not a prerequisite for this comparison.

The authoritative branding package means a reviewed asset bundle suitable for
distribution. It would provide exact input provenance for complete branded
libraries and graphics. It is separate from the native object editing work.
Future small graphics may be packaged; a large photo dataset needs a separate
distribution design. No S3 upload or URL policy is established here.

Release packaging must explicitly record these deferrals while retaining the
existing build, signing/notarization, security, model and installer gates. The
current unconditional browsing-resource release requirement still needs that
scoping change before a new tag. No release is cut as part of this demo.

## Comparison

Nine pages compare the existing renderer (left) with the native editing
candidate (right), using identical source copy, allocations and density:

| Family | Existing | Candidate |
| --- | --- | --- |
| Flat plain lists | Separate marker and text shapes | One native bulleted text box |
| Plain title/body cards | Grouped surface and separate text | One filled rectangle with distinct title/body paragraphs |
| Plain tables | Native table within a group | Directly selectable native table |

Each family appears at Comfortable, Compact and Dense. The fixture uses the
existing pinned V11 repository bundle and fonts; it contains synthetic copy.
The local Documents directory retains exact build receipts, object map, source,
composition log, pinned dependencies, native inventory and hashes. Private
PowerPoints/screenshots are not committed to Git or published to GitHub.

## Real Mac observations

PowerPoint opened the comparison without a repair prompt. All nine pages were
visually inspected. Initial card text positions, wrapping, table content and
styling match the reference examples; square bullet glyph size was corrected
after the first comparison exposed font substitution. This is visual review,
not a claim of pixel identity or exhaustive stock-template equivalence.

The final private edited copy records these UI actions:

- Comfortable list: extend the second bullet from two to four lines; following
  paragraphs reflow automatically. Enter inserts a fourth native square bullet
  with the same spacing and text font.
- Compact card: edit Update to Refresh in the body while preserving the title,
  and move the entire filled native object down 47pt.
- Compact table: select a cell directly and change Review to Reviewed, retaining
  the five other cells.

Read-only package verification retained all **93 generation object identities**,
the card's size/X position, all candidate top-level object structures, and the
immutable baseline receipt. Reconciliation exposes native edits for review; it
does not silently infer bullet identities or alter the source.

Windows, resizing/alignment, decorated tables, rich/nested lists and general
stock-template migration remain open. User demo review precedes migration and
release packaging. Do not represent these observations as release approval.

## Repeat the fixture

From the managed development checkout, use new private Documents destinations:

```sh
PPTXGENGO_NATIVE_COMPONENT_DEMO_OUT="$HOME/Documents/pptxgengo-qualification/new-component-demo" \
  go test -count=1 -timeout=3m -run '^TestNativeEditabilityComponentComparisonDemo$' ./internal/deckproject
```

Retain the generated baseline and make an editable copy named
`native-editing-demo-edited.pptx`. After performing the exact tasks above:

```sh
PPTXGENGO_NATIVE_COMPONENT_DEMO_VERIFY_ROOT="$HOME/Documents/pptxgengo-qualification/new-component-demo" \
  go test -count=1 -timeout=3m -run '^TestNativeEditabilityComponentComparisonSuppliedEdits$' ./internal/deckproject
```

The verifier writes a new `desktop-verification.json` and refuses overwriting
existing evidence. It does not control PowerPoint or confer visual acceptance.
Ordinary Make lanes clear both opt-in variables.
