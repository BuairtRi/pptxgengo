# Finding and selecting reference variants

The first shortlist contains 10 source examples across three semantic families.
It is a reviewed starting set; the wider catalog still contains 244 compositions.

## Visual review

Open the locally generated [reference gallery](../samples/component-adaptation/reference-variants-reviewed.html).
Each card includes an enlarged source region, full-slide context, exact source
position, named slots, intended use and dependencies. The previews come from
hash-verified native PowerPoint renders. A crop can include neighboring pixels;
it does not prove that the component has been isolated or transplanted.

Use short reference codes in feedback, such as “Prefer N2 for explanations; keep
N5 for readiness gates.” Alternatively mark preferences and export the JSON file.
The page does not auto-save. Exported preferences are separate from technical
approval and must be reviewed before updating durable catalog preferences.

| Code | Reference | Source position | Intended use |
|---|---|---|---|
| M1 | Metric card | Modernization 57 | Value plus short label; first implementation |
| M2 | Metric with context | Modernization 56 | Metric plus explanatory evidence |
| M3 | Metric strip cell | Graphics and Layouts 153 | Repeated statistics within a shared band |
| N1 | Numbered key-point panel | Graphics and Layouts 109 | Title and three supporting points; first implementation |
| N2 | Numbered explanation row | Modernization 8 | General steps, reasons and workstreams |
| N3 | Evidence and measurement row | Modernization 11 | Claim, evidence and success measure |
| N4 | Decision option card | Modernization 38 | Short side-by-side options |
| N5 | Evidence gate card | UHG 33 | Evidence, exit criteria and decision owner |
| P1 | Two-role delivery pod | UHG 43 | Lead plus one role; first implementation |
| P2 | Three-role delivery pod | UHG 43 | Lead plus two roles; first implementation |

These are source positions, not footer numbers. Numbered-card N2 is the strongest
general-purpose visual alternative; N1 is a deliberately simple first fixture
for preserving multiple text paragraphs. Recommendation does not imply user
design approval.

## CLI retrieval

```sh
python3 scripts/catalog-reference-review.py find --family numbered-card --query evidence
python3 scripts/catalog-reference-review.py find --family team-pod --query 'three roles'
python3 scripts/catalog-reference-review.py gallery --out samples/my-reference-gallery.html
```

The curated finder matches all query words against reviewed descriptions. It is
not embedding search or a learned relevance model. For wider discovery, use the
SQLite catalog and the full component gallery:

```sh
python3 scripts/catalog-index.py search \
  --db samples/component-expansion/catalog-v2.sqlite \
  --kind component --query 'metric' --limit 20
```

The agent workflow is: understand the slide's role and content fields; retrieve
candidates; compare actual renderings; recommend two or three suitable structures;
then choose or adapt a reference. Color alternatives belong to the chosen
structure unless they encode a different meaning. UHG pod colors encode staffing
ownership/commitment and cannot be treated as interchangeable decoration.

## First executable contracts

[M1](component-contracts/metric-card.json),
[N1](component-contracts/numbered-card.json),
[P1](component-contracts/pod-two.json) and
[P2](component-contracts/pod-three.json) now expose specific text bindings through
[`pptxcomponent`](../cmd/pptxcomponent/README.md). N1's visible numeral has been
explicitly verified and added to the contract; the earlier catalog listed only
heading and body slots.

Current editing scope: update text within the original source geometry, preserving
paragraph/run topology and staffing semantics. Proposed neutral/subtle/inverse
styles, dynamic role counts, automatic overflow rejection, cross-slide composition
and adaptation approval remain subsequent gates.

## Review limitations

The source previews were visually inspected. Gallery output hashes and JavaScript
syntax were checked. Interactive browser review is still pending: automatic
approval review previously rejected opening local HTML through the available
browser-control path. No alternative browser route was used.
