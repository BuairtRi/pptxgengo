# Human-editable source format

One ordered `deck.yaml` is authoritative. YAML comments are permitted; duplicate keys, YAML merge keys, anchors, aliases, unknown runtime fields and undeclared values are rejected. Build reports point to source locations. Edit the source, run `project check`, then build an immutable new version.

## Shared template slide

```yaml
schema: pptxgengo.deck-document.v1
id: controls-deck
title: Review controls
year: 2026
toolchain:
  lockfile: toolchain.lock.json
slides:
  - id: three-controls
    content_kind: supplied_content
    template: {scope: shared, id: cards/3}
    values:
      eyebrow: Delivery controls
      title: Three controls make review repeatable
      cards:
        - key: evidence
          title: Trace evidence
          body: Link each claim to the material supporting it.
        - key: review
          title: Review visibly
          body: Give reviewers the actual content and audience context.
        - key: maintain
          title: Keep the source
          body: Record the decisions in the maintained deck project.
```

These statements are illustrative workflow copy. For client claims, provide indexed evidence and qualifications. `content_kind` distinguishes `supplied_content` from `synthetic_example`; it does not by itself verify a claim.

Canonical shared template IDs are opaque keys such as `cards/3` or `architecture/layers-nav`, not scenario names guessed from a title. Optional shared `revision` is the decimal template revision; the toolchain lock separately pins the source bundle/files and engine. `values` must match that template's exact typed or named-slot contract. No missing-content specimen fallback exists.

## Assets and local templates

`assets` maps authored IDs to a registered `registry_id` or an original project-relative `path` and optional SHA256/provenance. Local file assets live inside the project; originals are retained independently from generated deck media. Slide values and local nodes refer to authored IDs. Shared template media fields use the contract's registered asset keys; inspect the contract before assuming project-local replacements are supported.

`local_templates` defines reusable project designs independently from visible slide copy. Each declares:

- frame reference: `wmds/frame/{none,left,right,nav}-{compact,tall}`;
- optional `frame_options` for declared split/navigation/surface/source allocations;
- grid reference: `wmds/grid/12-columns`;
- typed content `zones`, with required/schema/capacity notes;
- stable node IDs and bindings to slide values.

Node kinds are `text`, `box`, `image`, `rule`, `group`, `component`, `composite`. Placement uses body/rail/short_body/tall_body zones and an explicit rectangle or grid span. Local title/eyebrow/source roles populate frame chrome. The component adapter set is bounded: consult the current checker/CLI; a catalog source example does not create an arbitrary runtime component adapter.

To change shared geometry, create a local design with provenance pointing to its parent and reason. Preserve original shared identity and pin. A local template revision is the canonical definition SHA256. Editing a reusable local definition affects its dependent slides and approval hashes. If only one page should change, clone its local definition to a new stable local ID.

The compiler expands YAML into the native foundation document and Go layout engine. A maintained project keeps authored source, lock, assets, context and generated receipt/object map together. Bitwise reproduction depends on pinned source, fonts/assets, runtime/executable and build inputs; it is not promised across arbitrary platforms/toolchains.

## Deliberate local design changes

```sh
pptxgengo design project fork --project ./client-deck --template editorial-photo --as editorial-detail --slides page-detail --reason 'Give the detail page more text space'
pptxgengo design project detach --project ./client-deck --bundle v3 --slide page-detail --as detached-detail --reason 'Adjust this source layout locally'
```

`fork` clones a local definition and keeps its frozen ancestor snapshot; selected
slides must belong to the parent design. `detach` converts a supported generic
source-scene shared slide to a reusable local definition, retains authored values,
frame chrome and key identity, and records source pins. Typed retained card rows
and unsupported chrome/out-of-zone geometry remain shared and are rejected with
an explicit limit. A rejected conversion does not overwrite the source deck.
Inspect the result and run `project check` before the next build.

Mutations preserve an exact predecessor YAML snapshot and a decision receipt.
They currently rewrite YAML formatting/comments through the typed model; use
the retained predecessor for any editorial comments that need recovery. These
operations create local source changes, not shared-library patches or implicit
qualification.

## Local component adapters

Local definitions can use the component adapters that the selected compiler and
bundle support, which may include `funnel`, `cycle`, `road`, `gauge`, `bracket`,
`scorelegend`, `venn` and `maturity`. Check the active toolchain's checker or
help for the adapter list; catalog availability alone does not establish an
adapter. Character limits in example projects are local source constraints, not
a general content-fit guarantee. Some bundles expose slim frames; inspect the
selected bundle before assuming a frame variant exists.

Use executable IDs such as `wmds/component/venn` and the four special typed
definitions listed in the runtime reference. Semantic catalog names such as
`diagram.cycle` are discovery identities, not universal executable definition
IDs. Local allocations contain the complete diagram ink and labels; the
compiler reserves maturity endpoint stroke padding and plans label headroom
inside the declared component allocation.
