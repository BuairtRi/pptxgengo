# Choosing templates

Use templates by default. Search the library thoroughly before considering a custom slide, and record what you considered for every slide in the [composition log](project-templates.md#composition-logyaml).

## Know what's in the library

Templates are grouped into families:

The current V11 gallery contains 649 source templates: 648 active, one deprecated,
including 29 workshop layouts. The latest additions include five pillar layouts,
seven branching roadmaps and six narrative roadmaps. Search workshop scenarios,
facilitation structures and branch/decision relationships as well as familiar
card layouts. Do not infer capability from a template's name alone.

| Group | Families |
| --- | --- |
| Openers | covers, understanding |
| Content | core, argument, evidence, diagrams, heatmaps, venn |
| Technology | architecture, solution, lifecycle, software, modernization, maturity, decisions |
| Delivery | approach, team-curves, interviews, status, runbooks, change |
| Commercial and firm | commercials, offers, value, team, about, proof |
| Workshops and roadmaps | workshops, roadmaps; search facilitation, exercises, branching and narrative sequences |

Templates come in two tiers. **Core** templates are a slide type in its simplest useful form. **Working** templates are dense slides rebuilt from real proposals, such as phase detail, RACI and fee summaries. Dense content usually belongs in a working template, not a custom slide.

Each template record has a **purpose** (what the slide is for), **uses** (the components it contains), **slots** (what content it takes) and a **budget** (word and mark limits). Read all four before choosing.

Browse visually with `pptxgengo catalog --design-system --open`. Look at the gallery before relying only on search results; seeing the template is often faster than reading its record.

## Search by scenario and by shape

For each page, describe two things separately:

- **Scenario:** what the page is for. "Weekly status", "phase 2 approach", "team and roles".
- **Shape:** what the content looks like. "Four parallel points and one key message", "five sequential stages with an owner each", "three options compared on four criteria".

Search both ways, because a template built for a different scenario often has exactly the right shape. Labels such as "problem", "goal" or "recommendation" are hints, not filters.

```sh
pptxgengo design library-find --query 'weekly status' --kinds template --limit 10 --summary
pptxgengo design library-find --roles point,key-message --items 4 --item-role point --kinds template --summary
pptxgengo design library-find --structures comparison --visual-forms table --kinds template --summary
pptxgengo design library-inspect --id cards/3 --summary
pptxgengo design library-preview --id cards/3
```

- Always pass `--kinds template`; otherwise icons and components mix in.
- `library-find --summary` returns ranked candidates with purpose, content groups and verified `screenshot_paths`. Open the screenshots. Results with no match are dropped, so a short list is honest; reword or search by shape (`--structures`) instead. Words aren't split ("heatmap" finds nothing; "heat map" works). Scores are coarse and ties sort alphabetically, so look at several candidates. `--include-weak` lists everything, for browsing only.
- `library-inspect --summary` returns slots and types, item counts, frame and zones, and the zone-to-binding map.
- `library-authoring --template KEY` lists each slot's readable alias, description and approximate capacity (characters and lines). Read it before writing copy. Estimates use the authored typography and identify unsupported internals explicitly; changed density and actual text require `--check-fit`, a build and native review. Capacity is advisory, not a character limit or proof of fit.
- More flags are in the [CLI reference](cli-reference.md#find-and-understand-templates).

### Keyword ranking in the next source build

After building a new index, `library-find --retrieval keyword` uses BM25 over
names, purposes, relationships and authoring metadata. The older metadata mode
remains the default. See `docs/semantic-template-discovery.md` in the toolkit
source for index creation and migration commands; stable v4.1.0 lacks these flags.
Use `--require-shape` when all supplied roles, structures, visual forms and exact
source group counts must match. Read `structural_status` and source count scope;
neither a lexical rank nor a matching count establishes content fit. Inspect
and build the actual copy. Model-backed semantic/hybrid ranking remains pending.

Consider at least three candidates per page, or record why fewer exist.

## Judge candidates as a designer

For each candidate, ask:

1. **Does its shape match the content's relationships?** A sequence needs a sequence template; a comparison needs a comparison template. Equal boxes for unequal ideas is a mismatch.
2. **Does it hold the story unit?** Can the claim, its evidence and its caveat stay together on it?
3. **Does the count match its contract?** Fixed arrays retain their exact slot
   count. Use a sibling item-count variation or a supported variable list when
   needed; density does not add cards, columns or stages.
4. **Does the copy fit the budget?** If not, can the copy be tightened without losing meaning, or is there a roomier variant (a working-tier version, a tall footer, a different frame)?

5. **Does it suit the reading mode?** An emailed deck needs room for explanation; a presented deck can be sparser.
6. **Does it match the rest of the deck?** Similar content should use the same template family throughout.

For wordier content, also evaluate the supported [slide density presets](typography-density.md).
Automatic body fitting may produce a denser candidate, with warnings; inspect
reading quality before choosing it. Fixed item counts, geometry and source limits
still apply. An overview should not silently become a dense detail slide.

Then decide:

- **Use it as is.**
- **Use a variation:** change card treatment, frame or count. This is still the template.
- **Tighten the copy** to fit, without losing meaning. Ask the operator if the change is material.
- **Derive a local template** when it is about 80% right, and change only what's missing ([editing slides](editing-slides.md#change-a-slides-layout), [custom slide design](custom-slide-design.md)).
- **Reject it**, with a reason about structure or argument.

A template that fits only by removing essential explanation, forcing unequal ideas into equal boxes, hiding a qualification, or splitting the argument badly is overfitting. Show the operator the tradeoff before applying it. When every candidate fails, follow [custom slide design](custom-slide-design.md).

## Match page content automatically

`library-match` takes one page's content by role. It returns **ready** slides for templates it can fill completely, and ranked **needs_copy** drafts for templates that take all your content but have empty slots left.

```yaml
# page.yaml: block style
title: Three controls make the review repeatable
eyebrow: Approach            # include it: almost every template has a section label
relationship: parallel       # parallel | sequence | comparison | cycle | quantities | table | ownership
items:
  - lead: Trace evidence
    text: Link each claim to the material that supports it.
  - lead: Review visibly
    text: Give reviewers the actual content and audience context.
  - lead: Keep the source
    text: Record each decision in the deck project.
```

- **Item fields:** `lead`, `text`, `value`, `label`, `source`, `owner`, `duration`, `state`, `objective`, `id`, and `activities` (a list of items, one level deep) for phases.
- **Page fields:** `title` (required), `eyebrow`, `source`, `relationship`, `items`, `callout` (one item), or instead of `items` a `comparison` block for options against criteria:

  ```yaml
  relationship: comparison
  comparison:
    criterion_label: Criterion
    weight_label: Weight
    criteria:
      - {label: Cost, weight: "30%"}
      - {label: Fit, weight: "40%"}
    options:
      - {name: Build, values: ["2", "4"]}
      - {name: Buy, values: ["4", "3"]}
    legend: 1 = weak, 5 = strong
  ```

```sh
pptxgengo design library-match --page page.yaml --limit 4 --out ./candidates > ./candidates.log
```

- **Ready** means every value landed in a slot, every required slot is filled, and the slide built. Extra items no longer need an exact count: the first N items fill a larger group and the leftover slots make it a draft.
- **What can map:** bounded adapters for parallel groups, sequences, one-level
  phase activities and rectangular comparisons. `owner`, `duration`, `state`
  and callouts map when the selected source contract declares suitable editable
  roles. Many complex diagrams still produce drafts or explicit gaps. Inspect
  each candidate's mapped/unmapped fields; no arbitrary semantic rewrite is
  implied. Step state values must match the selected template's contract.
- **Output folder:** `match-report.json`, `page.json`, `candidate-NNN.yaml` with `.pptx` and `.layout.json` for each ready slide, a combined `candidates.pptx`, and up to `--limit` `candidate-NNN.needs-copy.yaml` drafts. `NNN` is evaluation order, not a rank.
- **Choosing a draft:** drafts are sorted by `near_miss_rank`, which only counts mapped fields, so judge the fit yourself and open each template's screenshot. Ignore drafts whose template doesn't express the page's relationship.
- **Finishing a draft:** fill every `""` slot with copy written for that template (check `library-authoring --template KEY` for sizes), delete the two draft header comment lines, then add it with `--check-fit` (below). Unfilled drafts are rejected.
- Exit code 1 means no ready slide; drafts and the report are still in the output folder. `--templates k1,k2` tests specific templates; `--render` adds PowerPoint renders and a contact sheet.

Add a ready slide or a finished draft without editing its ID:

```sh
pptxgengo design project slide add --project ./client-deck --file ./candidates/candidate-027.yaml \
  --as three-controls --after situation --check-fit
```

## Build two to four alternatives with real content

When the choice matters, or Ri hasn't seen this kind of page in the deck yet, build two to four alternatives with the same real copy and compare the renders. Make them meaningfully different: a diagram-led version and a text-dense version, or two different templates.

Use `library-match` when it can fill the candidates. Otherwise scaffold each candidate template with `project scaffold --stock`, fill them with the same copy, build them in a scratch copy of the project (or with `library-fit`), and render them together with `--contact-sheet`.

`library-fit` builds several candidates from one file in `pptxgengo.wmds-template-document.v1` format. Each candidate has a unique ID, a template key and its own complete `supplied_content` values. Keep the argument and evidence the same across candidates.

```sh
pptxgengo design library-fit --spec alternatives.json --out ./candidate-review
```

The output has a deck, layout report and foundation per candidate, plus `fit-report.json` with `passed` or `failed` for each. A finished run does not mean every candidate passed. Passing Go layout still leaves native rendering and visual review to do.

Show the operator the rendered alternatives with what each emphasizes, its fit caveats and your recommendation. Record the candidates, the choice and the reasons in the composition log.

An exact item count in a search result is the template's fixed topology, and word budgets are advisory; neither shows that your copy fits.

### Reuse authored finished content in the next source build

A `finished-slide` is content-complete material with explicit owner/revision and
reuse policy. It is distinct from an empty stock template. Search a private index
with `--kinds finished-slide`; inspect its approval, freshness, package path and
preview. Inserting a supported revision with `project slide insert` creates an
independent copy and writes library lineage plus your composition rationale.
Draft reuse requires `--allow-draft`. Existing composition entries and exact
compiler/template pins are required; expired/deprecated content is refused.
Adapted copy needs the destination deck's evidence and review process.
Stable v4.1.0 lacks this feature. See `docs/finished-slides.md` for the implemented
source commands, supported dependencies and remaining curation/native work.
