# Single-shape editable card pilot

Source builds support an explicit `wmds/component/editable-card`
component. One native filled rectangle contains two native paragraphs: a title
using the source `subhead` / `primary` role and a body using `body` / `secondary`.
There is no separate surface/text group. A 12-point inset, eight-point card gap,
and six-point source paragraph margin define this opt-in design. This authored
component remains deliberately narrow. Separately, the v4.2.1 `native-v1`
profile converts eligible stock scene cards, including titleless body cards,
multi-paragraph cards, supported rich inline runs, and simple outlined or
deemphasis surfaces when measured styles and positions fit. Decorated/stateful
or incompatible cards retain their structure with explicit reasons; see the
[catalog profile](native-editing-profile.md).

## Authoring

A local template can declare `card_title` and `card_body` as required string
zones and bind them explicitly:

```yaml
nodes:
  - id: source-card
    kind: component
    placement: {zone: body, span: {start: 1, count: 6, y: 36, h: 200}}
    definition: {scope: shared, id: wmds/component/editable-card}
    arguments:
      surface: subtle
      title: {binding: card_title}
      body: {binding: card_body}
```

Both fields must contain nonempty plain single-paragraph copy. Natural wrapping
is measured. Explicit line breaks, tabs, interpreted markup, footnotes, outline
surfaces, unknown style options and non-fitting copy are refused. The component
measures title and body independently using the source typography and density.
The candidate engine serializes their distinct paragraph/run styles inside the
same native shape; it records pending desktop qualification in the build report.

## Source correspondence and reconciliation

The object map records `pptxgengo.editable-card.v1`, the explicit title/body
source slots and exact native paragraph addresses. Field identity includes its
semantic role. Neither copy equality nor position is used to infer a binding.
Literal, repeated or ambiguous bindings remain manual.

Existing `project editability` and `project reconcile propose/apply` commands
can inspect this object and review supported copy changes independently for each
role. The baseline reader checks recorded paragraph metadata against the actual
PowerPoint XML. Added/deleted paragraphs, changed title/body role formatting,
extra runs and unsupported copy remain manual. Geometry changes are listed for
review even when supported copy can be adopted; adoption changes named source
text only. Reviewed packets retain the edited deck and baseline, and repeated
adoption reuses the existing receipt. Older single-field object maps retain their
whole-object comparison behavior.

## Qualification scope

Headless real-build checks verify one filled native object, separate title/body
addresses and styles, twelve-point native insets, three-tier measured fitting,
conflict handling, duplicate-binding refusal, metadata/format/topology guards,
each role's reviewed adoption/replay/rebuild and retained historical baselines.
These fixtures use synthetic copy. They do not qualify PowerPoint selection,
whole-card movement, resizing, alignment, native typography parity, Windows or
human ergonomics. Native diagram and table evidence is recorded separately in
[native editing pilot](native-editing-pilot.md).

Desktop qualification for this standalone component should use a separate owned copy under the user's
Documents folder, select the whole native rectangle, edit each paragraph, move,
align and resize it, Save As, then inspect lineage and supported reconciliation
on both Mac and Windows. Those family-specific results are distinct from the
profile's source-plan catalog comparison.

The 2026-10-07 comparison demo preserves the reference plain card's title/body
positions at Comfortable, Compact and Dense. The real Mac editing task moved
the entire Compact card 47pt and changed its body independently; all generation
identities survived saving. Resizing, Windows behavior and exhaustive catalog
desktop equivalence remain unqualified. See [component demo](native-component-demo.md).
