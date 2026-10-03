# Slice 4: integrated review and portable authoring

Source: `wmds-library.v2`, frozen at
`7bdcaee5030a12275a1f881a8542f4d302d207df`.

All **167 designs** generate: 166 active plus the retained deprecated
`from-to/rows` compatibility entry. Every design has a source specimen and a
meaningful alternate, yielding **334 accepted native specimens**. The three
parallel family reviews covered every page, with focused full-size inspection
of dense tables, text, marks, imagery and people components. Acceptance is
bounded to these specimens. Variable counts, arbitrary replacement copy and
exact PowerPoint font-file selection remain separate qualification work.

## Corrections and native evidence

Native PowerPoint exported both review rounds using **Best for printing** locally.
Three layout findings and five alternate identity headings were corrected:

- `phase-gate/evidence-matrix`: four synthetic evidence bullets shortened to
  one line, restoring the gap before fixed exit labels.
- `case-studies/cards-quotes`: flowing content explicitly ends at315pt, leaving
  9pt before fixed platform captions. The new optional card `contentBottom`
  preserves outer geometry and rejects excess flowing content. The alternate
  paragraph was shortened without changing the intended outcome.
- `transformation/pain-to-theme`: existing fixed-height text support now reserves
  24pt for headings and18pt for captions through a v2 named refinement. The
  alternate caption stays on one line inside its shaded row.
- Five alternate headings now match their fictional body identities: nested
  architecture, two portrait bios, six readiness criteria and team pods.

All eight changed native pages were inspected again. The other326 PNGs were
byte-identical to their accepted initial renders; slide XML/relationships also
remained identical. No fonts were shrunk or pinned source files modified.
The inverse divider retains the approved **54pt / three-grid-step photo offset**
to the left of its shorter whiteboard and the reserved left text column.

The final **167-slide library** includes one source example per design. Its
native PDF body text, line bounds and extracted font sizes match all167 accepted
paired-source pages exactly; footer numbering is regenerated for the shorter deck.

Durable evidence:

- [Integrated receipt](reference-review/slice4.json)
- [Approach, Openers and Commercials](reference-review/slice4-families/approach-openers-commercials.json)
- [Proof, Evidence and Argument](reference-review/slice4-families/proof-evidence-argument.json)
- [Team and Solution](reference-review/slice4-families/team-solution.json)

Local artifacts live in `samples/wmds-refresh-slice4-20261002/`. The deliverable
is `final/WMDS-template-library.pptx` and its native PDF. Paired review decks,
initial evidence and reproduction inputs are retained separately.

## Portable local.6 release

`pptxgengo design` defaults to bundled v2 source and `wmds-go-foundation.v2`.
Standalone developer defaults remain intact; the installed route accepts explicit
`--bundle v1` for earlier contracts. The package includes pinned static font
files, calibration, all696 registered artwork keys /695 distinct payload paths,
closed content contracts, editable specimen downloads and review receipts.
Artwork resolves inside the installed release, independently of the checkout
or a home branding directory. Hash verification remains mandatory.

The separate `catalog --design-system` gallery exposes166 active designs and
one hidden-by-default retired design. It supports family/update/search filters,
paired native previews, exact fields and list counts, and editable content or
illustrated composition downloads. The two retained typed card-row contracts
expose their own fields and exact3/4 card counts. The legacy101 source contracts
and132 component occurrences keep their separate gallery identities and scope.

The local release is staged without changing global launcher/skill links. See
[release evidence](../../../release/verification-wmds-v2.json) for actual package
operations and limits. Browser policy blocked the file:// gallery preview;
packaged links/assets were checked directly. No tests or negative controls were run.

## Reproduce

The family scripts retain deliberate per-template bindings and exact changed-slot
receipts. They require the prior reviewed slice3 content as provenance input.
The integration script consumes those three complete family packets:

```sh
python3 scripts/integrate-wmds-slice4.py
pptxdesign build --bundle library/wm-design-system/v2 \
  --engine wmds-go-foundation.v2 \
  --spec samples/wmds-refresh-slice4-20261002/reference.foundation.json \
  --out /tmp/wmds-integrated-new
```

Regeneration resets native acceptance to pending; it does not confer the prior
review on changed inputs. Release gallery generation requires an accepted packet.
For normal authoring, use the packaged gallery downloads and the installed skill's
design-system authoring reference. Build to new output directories and visually
review replacement content before sharing. Rebuilding a client deck is a consumer
of this library, not an additional prerequisite for the library refresh.
