> Repository-only legacy workflow. These source-scene/compose contracts are distinct from the released V11 `pptxgengo design` authoring route. Exact fixture evidence and historical checkpoints apply only to the artifacts named below. For current authoring use [the documentation index](../README.md).

# Proposal authoring through the library

Use this route for a new proposal or a supported layout adaptation. Run commands
from the repository root. Keep specs, values, narrative, source packet, and
selection traces together so later revisions reproduce the content decisions.

## Narrative before layout

Read the brief and identify the requested decision, audience, supporting evidence,
and qualifications. Use `pptxgengo.proposal-narrative.v1` (schema and examples in
`library/contracts/narrative-schema.json` and `library/proposal/narrative.json`).
Each slide needs an assertion title, role, one-sentence takeaway, required detail,
evidence references, qualifications, emphasis targets, and visual relationship.

For detailed proposals, keep mechanisms, ownership, dependencies, acceptance
conditions, and evidence. A brief headline plus three short bullets rarely has
the needed detail. Choose a dense pattern that supports the argument. If the
content exceeds its envelope, split the argument or choose another pattern;
retain necessary evidence and qualifications. Do not shorten substantive content
or reduce fonts simply to make a favorite template fit.

Claims must distinguish supported facts, hypotheses, unverified statements, and
synthetic examples. Pin source packet files and evidence with SHA-256. Exact
quotes must exist in the supplied source text and have a locator; the library
does not extract quotation text from compressed PPTX/PDF bytes. Preserve a pinned
text sidecar and attribution when necessary. Fictional people, allocations,
metrics, and scenarios remain visibly labeled in the deck.

## Discover, inspect, and choose

Build a CLI once for a qualification run. `pptxlib` reads portable contracts and a
rebuildable SQLite index over the existing inventory. `pptxcompose` measures and
renders chosen content. Preserve the same compose binary throughout probing,
measurement, build, and verification; rebuilding it changes the cache environment.

```sh
go build -o /tmp/pptxlib-authoring ./cmd/pptxlib
go build -o /tmp/pptxcompose-authoring ./cmd/pptxcompose
/tmp/pptxlib-authoring index --out samples/NEW_LIBRARY.sqlite
/tmp/pptxlib-authoring find --index samples/NEW_LIBRARY.sqlite --query workflow
/tmp/pptxlib-authoring inspect --index samples/NEW_LIBRARY.sqlite --id wm/workflow-matrix
/tmp/pptxlib-authoring preview --index samples/NEW_LIBRARY.sqlite --id wm/workflow-matrix
```

Default search returns adaptation-qualified contracts and excludes avoided
references. At the current candidate checkpoint, that search may be empty. Search
`--state measured_fixture` or `--state reviewed` for bounded experiments. Use
`--inventory --query ...` to explore raw assets/layouts/components, which are not
approved just because they appear in the index. Inspect purpose, roles, slot
bounds, measured envelope, transforms, preview, failures, and user preferences.
Preferred appearance is distinct from technical qualification. Record why the
selected pattern supports the slide's argument and evidence.

The initial thirteen contracts in `library/contracts/` are candidates. Their
historical fixture previews do not prove new content fits. Experimental use
requires the explicit `--allow-unqualified` flag and the full native QA sequence.
Do not change a contract's qualification state to bypass that sequence.

## Supply semantic values and assemble

Use a contract's named slots; example files are in `library/contract-examples/`.
Values provide text or bounded string arrays, never coordinates, typography, or
arbitrary JSON patches. Character caps are editorial input guards; native text
measurement remains decisive. Named style variants affect contract-owned colors.
Assets are pinned in the contract, preserving original bytes and source context.
If an image must change, define and qualify that asset binding rather than
rewriting the generated spec outside its contract.

For a single slide, values can specify `narrative_path` and `narrative_slide_id`.
For a deck, supply the narrative once in the assembly config:

```json
{
  "schema": "pptxgengo.library-assembly.v1",
  "narrative_path": "narrative.json",
  "slides": [
    {
      "contract_id": "wm/workflow-matrix",
      "values_path": "values/workflow.json",
      "slide_id": "journey-workflow"
    }
  ]
}
```

Paths in this config are relative to its directory. The narrative slide ID must
match the assembly slide ID. Required detail and qualifications must each appear
verbatim within supplied slot copy; do not split one required sentence across
multiple slots. The visible assertion title and metadata role/takeaway must match
at the contract-owned narrative bindings. Assembly rejects mismatches.

Review the meaning of each filled slot before treating assembly as a useful
draft. Exact required-detail matching proves coverage, not that the sentence
belongs under its heading. Replace inherited source-topic labels and boilerplate;
put evidence under evidence headings, decision requests under decision headings,
and global qualifications in a suitable qualification zone. Check phase dates
against the brief and the roadmap axis. Distinguish the decision requested now
from later pilot/expansion gates. Read joined rich-text runs as whole sentences.
If correct copy has no suitable zone, select or author a bounded layout variant
and retain its unqualified status until measured and visually reviewed.

```sh
/tmp/pptxlib-authoring assemble --index samples/NEW_LIBRARY.sqlite \
  --config library/proposal/assembly.json --out samples/NEW_ASSEMBLY \
  --allow-unqualified
```

Replace example paths with the actual task's new config/output paths. Assembly
produces a complete `spec.json`, individual slide specs, and selection/assembly
traces; it changes only semantic values, declared IDs, and declared page labels.
Repeated use of one contract requires distinct output slide IDs. Keep the config
and values as revision sources. Do not hand-position generated cells afterward.

## Measure, build, and review

Use the compose probe/measure/build/verify workflow in `workflow.md`. Serialize
all PowerPoint operations. Keep source decks and unsaved user decks open; close
only known saved task copies when required to reopen exact bytes for verification.
No desktop inspection connection or export timeout is evidence of a good render.

Native probes calculate text fit. Final verification checks the generated deck.
Export through PowerPoint and review every slide at presentation size, including
margins/padding, typography hierarchy, text overflow, relationship meaning,
photography/icon relevance, decorative ink, sources, dates, and staffing caveats.
Log errors and preserve rejected evidence. Qualification needs changed-content
and content/cardinality stress, including a useful negative failure. Unit tests,
character counts, or an unchanged copied fixture do not close this gate.

## WM accents and difficult diagrams

Read `library/diagram-components/README.md` and the pinned arrow catalog when
using new diagrams/accents. Phrase selectors use exact copy plus occurrence or
Unicode range; native character geometry identifies wrapped fragments. Use
`multiline: per_line` only when separate marks on the wrapped lines make sense.
Do not rewrite source wording to make an underline easier to place.

Highlights sit behind the intended text; underlines use explicit optical
calibration. Hand-drawn arrows use vetted visible tail/tip/tangent geometry,
uniform scale, and declared ports. The picture rectangle is not the endpoint.
Curved ink collision tiles are asset-contract metadata and still need optical
review. A stock PNG counterpart may have a different canvas from its SVG; use
only the pinned, verified fallback.

When automatic placement is unsupported, stage the selected asset and a precise
visible placement note in reserved space and mark the slide unfinished. Do not
present guessed placement as polished. Routes remain editable segments unless a
specific contract proves native attachment; moving a shape in PowerPoint does
not automatically reroute them.

## Revision boundary

Configuration changes rebuild through the same contract and QA path. Colleague
edit reconciliation beyond the existing bounded text-only command is deferred
with Wave 5. This workflow does not promise arbitrary returned-PPTX to semantic
YAML conversion or preservation of unsupported manual edits.
