# Portable component contracts

`schema.json` defines the versioned `pptxgengo.library-component.v1` record. One JSON file per contract is the durable source of truth. `pptxlib index` verifies each template, asset, preview and proof hash before projecting metadata into a new SQLite index. The SQLite file is rebuildable and should never be treated as the approval record.

A contract pins a source identity and slide, a compose template and slide ID, semantic content roles, slot JSON pointers into that slide, cardinality and copy bounds, tested transforms, semantic style variants, design preference, qualification state, proof/failure artifacts, previews and provenance. `adaptation_qualified` requires hash-pinned `native_fit`, `changed_content`, `stress_negative` and `visual_review` evidence roles. Preference remains separate from qualification.

Slots accept only strings or bounded arrays of strings. Every slot has a positive `max_chars` cap; array slots apply it to each item and also have explicit item count bounds. These are editorial input limits, not proof that changed copy fits. A pointer may target content such as `text`, `label`, `title`, `role`, `takeaway`, `labels`, `roles`, `paragraphs` or an accent `phrase`; geometry pointers are rejected. Array slots may use a contract-owned `item_template` with `item_value_pointer`, `item_id_pointer` and `item_id_prefix` for repeated role objects. Values files provide no coordinates, typography or arbitrary colors. A selected semantic style variant replaces exact token names only in color fields with contract-owned `#RRGGBB` values; text, IDs and asset paths are untouched. Unqualified variants need the experimental override.

Example commands from the repository root:

```sh
go run ./cmd/pptxlib index --out /tmp/pptxlib-new.sqlite
go run ./cmd/pptxlib find --index /tmp/pptxlib-new.sqlite --inventory --query governance
go run ./cmd/pptxlib find --index /tmp/pptxlib-new.sqlite --query process
go run ./cmd/pptxlib inspect --index /tmp/pptxlib-new.sqlite --id CONTRACT_ID
go run ./cmd/pptxlib preview --index /tmp/pptxlib-new.sqlite --id CONTRACT_ID
go run ./cmd/pptxlib instantiate --index /tmp/pptxlib-new.sqlite --id CONTRACT_ID --values values.json --out NEW_OUTPUT_DIR
```

Default `find` only returns `adaptation_qualified` contracts and excludes avoided references. `--inventory` searches the existing catalog for exploration and explicitly labels results as unproven. `instantiate` requires `--allow-unqualified` for a candidate or measured fixture experiment. It writes `spec.json` and `selection-trace.json` with an unmeasured changed-copy status. The existing `pptxcompose` native probe, measurement, fit, build, verify and visual review sequence still governs the output.

Narrative input uses `pptxgengo.proposal-narrative.v1`: a brief, slide claims and evidence. The loader verifies source packet hashes and exact quote text against the pinned source file. A values file may reference a narrative path and slide ID; its assertion title, role and takeaway must match the resulting slide exactly at the contract-owned `narrative_bindings` pointers, and required detail and qualifications must be present in supplied slot values. The trace records narrative and source hashes. Synthetic claims require an explicitly synthetic brief. No claim is promoted by component selection.

`assembly-schema.json` defines a deck as an ordered list of `{contract_id, values_path, slide_id}` entries with an optional `narrative_path`. Relative paths resolve from the config directory. Slide IDs must be unique. A contract may declare `composition.pagination_binding` to an existing string page label; assembly writes the sequential page number only at that pointer. The page pointer cannot overlap a content slot. Contracts without the binding retain their existing page text.

```sh
go run ./cmd/pptxlib assemble --index /tmp/pptxlib-new.sqlite --config assembly.json --out NEW_DECK_DIR --allow-unqualified
```

Assembly validates each slide through the same instantiation path, checks the narrative when supplied, and publishes the complete directory in one rename. It writes `spec.json`, `assembly-trace.json`, and per-slide specs and selection traces under `slides/`. The assembly trace pins config, narrative, original values, contract, source, and template hashes. Candidate contracts require `--allow-unqualified`. Changed copy still needs the native fit, build, verify, and visual review sequence.
