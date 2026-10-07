# Native deck, slide and shape lineage

Project builds now carry `pptxgengo.native-lineage.v1` in the immutable object map
and native identity tags in the generated PowerPoint. This extends the
[native text field baseline](native-field-mapping.md). [Three-way text proposals and reviewed adoption](text-reconciliation.md) are
implemented with a bounded plain-text contract.

## Identity representation

The tags use PresentationML `custDataLst` and related `tagLst` parts, the standard
customer-data facility for presentations, slides and shape nonvisual properties.
See Microsoft's [customer-data definition](https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.presentation.customerdatalist?view=openxml-3.0.1).
These are native PowerPoint tags, separate from editable Selection Pane names.

Tokens are uppercase SHA256 strings because the PowerPoint
[Tags API stores values in uppercase](https://learn.microsoft.com/en-us/office/vba/api/powerpoint.shape.tags).
Names and Unicode authored IDs are not uppercased. A token contains no source
copy, evidence, ownership metadata, or filesystem path.

| Owner | Tags | Baseline meaning |
| --- | --- | --- |
| Presentation | `PPTXGENGO_SCHEMA`, `PPTXGENGO_DECK`, `PPTXGENGO_BUILD` | Schema 1; exact deck; deterministic source/lock/native generation |
| Slide common data | `PPTXGENGO_SLIDE`, `PPTXGENGO_BUILD` | Authored slide identity at that generation |
| Each shape's own nonvisual properties | `PPTXGENGO_SHAPE`, `PPTXGENGO_BUILD` | Exact baseline object, independent of later name/number/address |

The generation token hashes the lineage schema, deck token, canonical source
hash, exact lockfile hash and unstamped native PPTX hash. Execution-specific
`build_id` values remain in receipts: identical source/runtime/native generations
retain identical PowerPoint bytes and tag identity. The receipt also records
`native_lineage_build_token` and hashes the tagged deck and object map.

The sidecar stores slide-token to authored-ID mappings, object `shape_token`
and `native_parent_token`, and the exact generation inputs. Shape tokens identify
objects within that baseline generation. They are not promised to survive a new
build, a compiler migration or moving an object to another authored slide.

Stamping inserts only tags, relationships and content types. Visible XML is
preserved, with empty nonvisual containers expanded as necessary. Original
paragraphs/runs, source fields, formatting fingerprints and template pins remain
in the object map. Connectors also retain their own nonvisual identities.

## Reading an edited deck

`InspectNativeLineage` requires a trusted baseline object map. Its caller must
verify that sidecar against the immutable receipt and baseline artifacts before
using the report for reconciliation. Tags establish identity continuity, not a
cryptographic signature or proof of unchanged content.

The reader follows the live presentation's relationships and slide list, then
reads each shape's own tag reference, including children of groups. Numeric shape
IDs, names, ZIP part paths, current text and slide positions are diagnostics;
none are used to guess a source match. Slide reordering and native renaming do
not discard a surviving token.

A wrong presentation/deck/generation, malformed relationship, external tag
reference, incorrect tag content type or corrupted baseline token fails. Older
baselines without lineage require a new build before this route can be used.

The report explicitly lists untagged, unmatched, missing and duplicated slides
or shapes; movement between authored slides; changes of native kind; and changed
group ownership. Missing objects never imply deleting source. A duplicated token
cannot select one of its copies. Consumers must retain these issues and exclude
ambiguous objects from proposals.

Geometry, formatting and source-field comparison belong to the next three-way
layer. Identity survival alone does not establish supported edit semantics,
visual acceptance or permission to adopt content.

## Bounds and preservation

Reading extracts no files. It rejects duplicate/unsafe ZIP paths, nonregular or
encrypted parts, external/escaping relationships, duplicate tag names and
relationship IDs, XML directives and duplicate XML attributes. Bounds are:

- 512 MiB compressed and aggregate declared/decompressed package size;
- 64 MiB per part, 40,000 ZIP entries and 10,000 native shapes;
- XML depth 100 and 200,000 elements per XML part;
- tag names 256 characters and values 4,096 characters.

Normal reads and comparisons do not change the edited deck or baseline.
The writer accepts the project's generated package and refuses existing customer
data or tag-part/relationship collisions rather than overwrite it.

## Evidence and remaining qualification

Focused checks cover deterministic repeated builds, exact baseline stamp/read,
visible XML preservation, names/numeric IDs/text changed independently of tokens,
slide reordering, missing/duplicate/untagged shapes, ungrouped ownership, wrong
baseline identity, external/malformed tags, wrong content types and package/XML
bounds. These are generated fixtures and package mutations.

An October 7 UTC trial on macOS 27.0.1 with PowerPoint 16.113.4
(`16.113.26100421`) opened a separate owned synthetic copy. The Save As AppleEvent
timed out with `-1712` and produced no output. The owned presentation was closed
without saving; no native identity-survival qualification is claimed from that
trial. Windows desktop Save As/edit/reorder/duplicate/delete/ungroup trials remain
pending when the interactive runner is available.

Two explicit test hooks support desktop evidence without starting Office in the
headless development lane:

```sh
# OUT must be a new directory; generates a pinned synthetic baseline and hashes.
PPTXGENGO_NATIVE_LINEAGE_FIXTURE_OUT=/new/owned/fixture \
  go test ./internal/deckproject -run '^TestNativeLineageDesktopFixture$' -count=1

# After a native trial saves a new file, verify tokens and retain a JSON report.
PPTXGENGO_NATIVE_LINEAGE_FIXTURE=/owned/fixture \
PPTXGENGO_NATIVE_LINEAGE_EDITED=/owned/trial-output.pptx \
  go test ./internal/deckproject -run '^TestNativeLineageDesktopEdited$' -count=1
```

The edited-file hook checks fixture integrity and refuses replacing an existing
inspection report. Its scope is identity survival only. Actual editing ergonomics,
text proposals/adoption, native visual comparison and both-platform qualification
remain separate requirements in the handoff.
