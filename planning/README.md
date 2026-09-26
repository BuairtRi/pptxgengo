# Source corpus and inventory policy

The current implementation sequence is in [IMPLEMENTATION_PLAN.md](../IMPLEMENTATION_PLAN.md).
Inventory describes what exists; visual approval and reusable contracts are separate work.

| Source | Slides | Native layout parts | Masters | Availability |
| --- | ---: | ---: | ---: | --- |
| Graphics and Layouts | 166 | 33 | 1 | Local source, ignored |
| UHG Fabric Platforming proposal | 85 | 185 | 7 | Local source, ignored |
| EnableComp proposal | 35 | 23 | 1 | Local source, ignored |
| Software modernization campaign pick deck | 83 | 72 | 2 | Tracked source in `samples/` |

- [Source registry](source-registry.json): four sources, 369 slides; initial explicit allowlist.
  Across slide parts: 10,102 raw object nodes, including 315 native group
  containers. These include nested descendants and repeated designs; they are
  not counts of distinct reusable components.
- [Modernization inventory](modernization-inventory.md), with compact machine-readable records alongside it.
- [Asset catalog audit](asset-catalog-audit.md).
- [Shape and layout execution detail](shape-layout-execution.md).
- [Reconstruction checkpoint](RECONSTRUCTION_CHECKPOINT.md) and [QA findings](QA_ERRORS.md).

Only `samples/software modernization campaign pick deck v1 - Repaired.pptx` is
committed under `samples/`. Other originals, generated PPTX/PDF/PNG files, scene
projects and raw inspection outputs remain local and ignored. They have not been
deleted. Curated manifests and durable documentation belong outside `samples/`.
The SQLite catalog and previews will be rebuildable local outputs; approved
descriptions, ratings, source hashes and contracts must remain versioned files.

Local inventories of the original three sources are under
`samples/inspection/{graphics-and-layouts,uhg,enablecomp}/`. Their 286 slide
records report explicit properties, nested objects, notes, dependencies and
local transforms. They do not resolve all inherited styles or certify fit.
EnableComp slides 24–35 are hidden. Hidden status does not exclude a slide from
inventory. OLE/alternate representations require explicit capability records.

The branding workspace is `/Users/rscott/Documents/branding`; this path is a
local configuration value, not a requirement for another user's checkout.
Its 166-slide stock deck is distinct from the older 179-slide skill reference.
Production source registration should use hashes plus configurable locations,
never an unrestricted recursive scan of experiment output folders.

Recreate a structural inventory using a new output path:

```sh
python3 scripts/inventory-pptx.py \
  "samples/software modernization campaign pick deck v1 - Repaired.pptx" \
  /tmp/modernization-inventory.json
```

The committed deck permits a portable extraction example. Repeating UHG visual
benchmarks requires the separately supplied local source and macOS PowerPoint.
