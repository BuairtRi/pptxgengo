# Editing individual slides

Keep the source and generated baseline together. Author source changes in the project; for PowerPoint edits, use a separate working copy and the reconciliation workflow below. After a source change: update the slide's composition-log entry, run `project check`, `project build`, `project measure --report BUILD/layout-report.json --slides <ids>`, and render the changed pages.

Commands run as `pptxgengo design project … --project PATH`. Use block-style YAML in slide files and patches. In flow style (`{…}`), quote any value containing a comma; an unquoted comma breaks the map, sometimes with a misleading `binding.unsupported_field` error.

## Put each slide in its own file

```sh
pptxgengo design project split --project ./client-deck --bundle v11
```

Split once, early. `deck.yaml` becomes an ordered index of slide files; notes move to `notes/`, local templates to `slides/templates/`. Shared-template slides get readable `content` with `bindings` to the template's slots. The deck's output doesn't change.

Use the bundle matching the project's lock; V11 is the current source for new
projects. Upgrade an older project with [project migrate](cli-reference.md#move-a-project-to-a-new-cli-version)
before applying V11 density. Supplying a bundle during split enables stock aliases.

## Change a slide's copy

Edit `slides/<id>.yaml` directly. Change the text under `content`, keep every key and every array's item count, and set `content_kind: supplied_content` once the copy is real. This is the simplest route for one slide.

Use a patch when changing several slides at once:

```sh
pptxgengo design project edit --project ./client-deck --patch slide-edits.yaml
```

```yaml
situation:
  content:
    eyebrow: What we heard
    headline: Your PMs spend two days on each PRD while interviews go unsynthesized
    cards:
      - key: prd
        title: PRDs
        body: Each takes about two days to write.
      - key: interviews
        title: Interviews
        body: Recordings pile up unsynthesized.
      - key: roadmap
        title: Roadmap
        body: Priorities go to whoever argues hardest.
```

- Patch fields: `content`, `values`, `bindings`, `template`, `brief`. Each named field is replaced completely; arrays never merge. `content_kind` can't be patched; edit the file.
- Shared-slide edits and swaps regenerate capacity comments once for the selected
  template, preserving human comments on retained fields. Estimates are advisory;
  after direct YAML or density edits, inspect `library-authoring --template KEY`
  and the build report rather than treating an old comment as a fit guarantee.
- `[[…]]` in a title applies the highlight mark to one to four words, once per slide.
- Add `--check-fit` to `edit` (and `slide add`) to run the layout check before anything is written; a failure leaves the source unchanged. Without it, only `project build` checks fit (`title has 2 lines, capacity 1`).

## Adjust textual density

Edit `density`, `header_density` and `auto_density` directly beside the slide's
`template` field. These are slide metadata, not content bindings or patch fields.
V11 can step the entire body through Comfortable / Compact / Dense and reports
every adjustment; the header remains Comfortable unless explicitly changed.
Use `auto_density: false` to keep a deliberate tier. Source-owned limits still
apply. Read [typography density](typography-density.md) before changing the tier,
then rebuild and review legibility. Geometry and fixed item counts do not expand.

## Add a slide

1. Get a slide file, either way:
   - `pptxgengo design project scaffold --stock --template KEY --id ID --out new-slide.yaml`. The file uses the template unchanged, with readable field names and a comment under each field giving its description and approximate capacity. Example copy is marked `synthetic_example`.
   - A ready `library-match` candidate, or a finished `needs_copy` draft ([template selection](template-selection.md#match-page-content-automatically)).
2. Replace the copy and set `content_kind: supplied_content`.
3. Write the slide's composition-log entry (`project check` fails without one once the log exists).
4. Add it:

   ```sh
   pptxgengo design project slide add --project ./client-deck --file new-slide.yaml --as phases --after situation --check-fit
   pptxgengo design project slide add --project ./client-deck --file new-slide.yaml --as phases --into-section approach --check-fit
   ```

   `--as` sets the slide's ID (the input file is left unchanged); without it, `--id` must equal the file's `id`. `--before`, `--after` or `--into-section` places it; `--into-section` alone makes it the section's first slide. With no position it goes at the end of the deck. In a split project the file is copied to `slides/<id>.yaml`.
5. Check and build.

Slide IDs: letters, digits, `.`, `_`, `-`; start with a letter or digit. Choose it at add time; to rename a slide, remove it and add its file again with `--as NEW-ID`, then update the composition log.

## Move, hide, show or remove a slide

```sh
pptxgengo design project slide move --project ./client-deck --id phases --before recommendation
pptxgengo design project slide hide --project ./client-deck --id phases
pptxgengo design project slide show --project ./client-deck --id phases
pptxgengo design project slide remove --project ./client-deck --id phases
```

- **Move:** needs `--before`, `--after` or `--into-section`. `--into-section SECTION` alone makes the slide that section's first slide; with `--before`/`--after` the position must be inside the section. A plain `--before <first slide of a section>` puts the slide at the end of the previous section. If the moving slide starts a section, add `--reanchor`.
- **Hide/show:** toggles `hidden: true`; the slide stays in the deck and its section.
- **Remove:** the slide file and its notes stay on disk, unreferenced, and the receipt lists them. Sections re-anchor automatically. The last remaining slide can't be removed. To bring a slide back, add its retained file (`--id ID --file slides/<id>.yaml`); a changed copy must be added with `--as NEW-ID`.
- Each operation writes a receipt under `decisions/`. Update the composition log (remove obsolete entries) and rebuild.

## Change a slide's layout

Pick the smallest change that works and log it.

**1. A different shared template.** Start with a proposal:

```sh
pptxgengo design project swap --project ./client-deck --slide situation --template cards/4 > swap.json
```

The proposal lists `mapped`, `unmapped` and `missing` fields and contains a ready `patch`. Fields carry over by role and position, including across different item counts (three cards into a four-card template maps all three plus headline and eyebrow). Different structures (cards to a statement, phases to cards) carry over little beyond the headline.

- `--apply` works only when nothing is missing. Leftover (unmapped) copy needs `--allow-unmapped`, which deletes it.
- When fields are missing, take `.patch` from the proposal, fill every blank with copy written for the new layout, and run `project edit --patch FILE --check-fit`. `swap` itself doesn't check fit.
- Swaps to `*-nav` templates and from local templates aren't supported; use a full patch.

Update the composition log's `chosen_template` before checking.

**2. A local derivative of a shared template** (the "80% right" case):

```sh
pptxgengo design project scaffold --bundle v11 --template lifecycle/three-phases \
  --reason 'Add a fourth phase' --out /tmp/phases-four.json
```

Save the output's `template` object as `slides/templates/phases-four.yaml`, register it under `local_templates`, and point the slide at `template: {scope: local, id: phases-four}` with real `values`. `--omit-nodes` drops source nodes. This works for source-scene templates; typed templates (`cards/3`, `cards/4`) can't be derived and must stay shared. Then modify the definition ([source format](source-format.md#local-templates), [custom slide design](custom-slide-design.md)).

**3. Detach** a slide's shared template into a local one, keeping its readable copy: `project detach --slide ID --as NEW --reason R`. Capacity comments are dropped, and decorative slots may appear as empty entries; leave them empty. Works for source-scene templates, not `cards/3` or `cards/4`. Use detach to keep an existing slide's copy; use option 2 when starting fresh.

**4. Fork** a local template so one slide can diverge: `project fork --template ID --as NEW --slides IDS --reason R`.

Editing a local template changes every slide that uses it.

## Add or change an image

```sh
pptxgengo design project asset add --project ./client-deck --id workshop-photo \
  --file ./approved/workshop.jpg --description "Client workshop, June 2026" --focus 0.5,0.4
```

Then put the asset ID in the slide's image field (`photo: workshop-photo`; use `project:workshop-photo` if it collides with a library asset key). Builds shrink JPEGs to 220 pixels per inch at their placed size and keep the original in `assets/originals`; `layout-report.json` (`media_optimization.parts`) shows what happened to each image. PNGs aren't resized, so shrink large PNGs before registering. Library photos and icons: see [assets](assets.md).

## Speaker notes

Use `notes_file: notes/<slide-id>.md` (or inline `notes: |`, not both). Notes never substitute for visible explanation in an emailed deck.

## Sections and dividers

```sh
pptxgengo design project section list --project ./client-deck
pptxgengo design project section add --project ./client-deck --id findings --title Findings --before results \
  --divider divider/panel-edge --divider-photo photo-abstract-cubes
pptxgengo design project section rename --project ./client-deck --id findings --title Recommendations
pptxgengo design project section remove --project ./client-deck --id findings
```

- `--before` names the section's first slide. Adding the first section midway creates an `Opening` section for earlier slides.
- `--divider` inserts a visible divider slide (`SECTION-divider`, or `--divider-slide-id`). `panel-edge`, `panel-photo` and `full-photo` take the title and `--divider-photo`; other dividers need `--divider-values FILE.json`.
- `rename` changes only the section name; `remove` keeps all slides.

## Check, build and review the change

```sh
pptxgengo design project check --project ./client-deck
pptxgengo design project build --project ./client-deck
pptxgengo design project titles --project ./client-deck
pptxgengo design project measure --report ./client-deck/builds/<build-id>/layout-report.json --slides phases,situation
pptxgengo design render --pptx ./client-deck/builds/<build-id>/deck.pptx --out ./review-<n> --png --slides 4,6 --contact-sheet
```

Look at every changed page at full size, then record the review with `project attach-render` ([review packets](review-packets.md#record-native-review)). Changing a slide invalidates that slide's approvals; `project status` lists them.

## Rebuild or update an existing deck

1. Inventory the deck: `pptxgengo design source-inventory --in existing.pptx --out ./inventory`. Read `source_inventory.md` for each slide's title, text, tables, charts, images and notes.
2. For each source slide, decide its job and content relationship, then find a template ([template selection](template-selection.md)): write a page spec and run `library-match`, or search and scaffold by hand.
3. Rewrite the copy following the voice references, add each slide with `project slide add`, and log each mapping decision in the composition log, citing the source slide number.
4. Re-run `source-inventory --in existing.pptx --out ./inventory-2 --project ./client-deck` to see source and project slides side by side, and confirm nothing was dropped. The tool never maps them for you.

## If someone edited the PowerPoint

Never overwrite a hand-edited deck or an immutable build. Keep the edited copy
separately and record who changed what in `project.md`.

For a working copy of a receipt-backed project build, use reviewed text reconciliation:

```sh
pptxgengo design project reconcile propose --project ./deck \
  --edited './colleague edited.pptx' --out ./new-text-review
pptxgengo design project reconcile adopt --project ./deck \
  --packet ./new-text-review --decisions ./review-decisions.yaml
```

Read `report.json`, review baseline/current YAML/native values and all manual
items, and create explicit decisions with the actual actor, report hash,
proposal IDs, `use_native` or `keep_yaml`, and reasons. Only supported
`native_only` and `conflict` proposals are selectable. Do not invent decisions
or operator acceptance. Adoption retains the edited deck and predecessor source,
preserves unsupported changes for review, and requires a new build and visual
review. Inspect `report.json` and the CLI's decision template for required fields. A successful adoption does not approve the edited layout or establish desktop fidelity.

For older baselines without lineage or unsupported fields, use
`source-inventory`, inspect the edited deck and reconcile reviewed changes into
source manually. A modified build baseline is an integrity blocker; do not
bypass its receipt or remove its lock.


## Inventory native editing structure

Use `project editability --project PATH` to inspect receipt-pinned ownership,
Selection Pane labels, nesting and source fields. The report describes object
structure; perform and review the actual editing task in PowerPoint separately.
Under `native-v1`, a list can reflow in one box, a simple card can move as one
shape, and a table can be selected and edited through its cells. Extra copy may
need the box resized. Decorated components can retain multiple objects; inspect
their retention reason before changing the layout.

Keep the generated baseline unchanged and save native edits in a separate copy.
Review supported text proposals before adoption. Combined list/card fields,
geometry, table structural edits and unsupported formatting require manual
source updates and a fresh build/render review.


## Reuse maintained authored slides

The CLI can publish, find, preview and insert operator-provided closed authored
revisions. No curated reusable-slide inventory or content-complete browsing deck
is bundled; use this route only when an actual approved revision is available.
Use the distinct `finished-slide` entity kind, inspect reuse scope/freshness and
exact template/toolchain pins, and review the actual supplied copy and evidence.
Keep closed packages private: an exact Markdown registry can include material
beyond the selected claim.

`project slide insert --project PATH --package DIR --id FRESH_ID --rationale
REASON` creates independent item, asset and claim identities, registers exact
evidence files and records composition/library lineage. `--allow-draft` permits
unapproved draft work explicitly. Missing or incompatible dependencies refuse
insertion; do not strip evidence to make a package fit. Adaptations require the
destination deck's copy/evidence review, build, fit checks and visual review.

Only record real operator decisions using `project slide review-reuse --package
DIR --decision FILE.json --out NEW_DIR`. Approval names the exact revision,
source/preview/report hashes, reviewer, date, reason, scope and expiry policy.
The command retains the raw decision and predecessor manifest in a new revision.
It does not authenticate the reviewer or establish native acceptance. Never
invent approval or an initial curated content set. Deprecation/draft revisions
retain decision history; earlier inserted copies keep their original lineage.
