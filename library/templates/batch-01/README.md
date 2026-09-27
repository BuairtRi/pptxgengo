# Architecture/product batch 01

Five source-bound component contracts expose 101 semantic slots over 158 existing
text bindings. All five passed `pptxcomponent inspect`. They preserve source
geometry and rich-text segment counts. They are **not adaptation-qualified**.

| Reference | Arrangement | Slots | Text bindings |
|---|---|---:|---:|
| UHG 13 | Layer legend beside architecture map | 12 | 12 |
| Modernization 44 | Inputs, platform, integrations and features | 27 | 62 |
| Modernization 46 | Evidence progression and action panels | 17 | 17 |
| EnableComp 23 | Product proposition and numbered method | 27 | 30 |
| UHG 12 | Capability pillars and risk band | 18 | 37 |

UHG 13's main architecture map is opaque artwork. Its adjacent layer text is
editable; arbitrary changes to the map require separate component reconstruction.
No semantic color profiles are declared in these five contracts yet.

## Recreate local scene dependencies

The extracted projects and inspection output are ignored local artifacts. The
contracts pin source and scene hashes. Obtain the exact source decks named in
`planning/source-registry.json`; only the modernization deck is tracked in Git.
Run from the repository root with new output directories:

```sh
go build -o /tmp/pptxscene-template-expansion ./cmd/pptxscene
go build -o /tmp/pptxcomponent-template-expansion ./cmd/pptxcomponent
python3 - <<'PY'
import hashlib, json, subprocess
from pathlib import Path
sources = {s['source_id']: s for s in json.load(open('planning/source-registry.json'))['sources']}
for source_id, slides, destination in [
    ('uhg', '12,13', 'uhg-scenes'),
    ('software-modernization', '44,46', 'modernization-scenes'),
    ('enablecomp', '23', 'enablecomp-scenes'),
]:
    source = sources[source_id]
    path = Path(source['path_hint'])
    assert hashlib.sha256(path.read_bytes()).hexdigest() == source['source_sha256'], source_id
    subprocess.run(['/tmp/pptxscene-template-expansion', 'extract', '--source', str(path),
                    '--slides', slides, '--out', 'samples/template-expansion/batch-01/' + destination], check=True)
PY
```

The extractor currently uses `uhg-NNN.json` scene filenames for all sources;
the manifest/source hash establishes source identity. Do not rename scene files.

Inspect an individual contract:

```sh
/tmp/pptxcomponent-template-expansion inspect \
  --project samples/template-expansion/batch-01/uhg-scenes \
  --contract library/templates/batch-01/contracts/layered-architecture-overview.json
```

`manifest.json` maps every contract to its source project and scene.
`inspection-report.json` pins the inspected contracts and local inspection outputs.
Source copy is retained as an inspection reference, not approved content for a new
client deck. Rich-text slots accept an array matching their listed binding order.

## Remaining acceptance gates

1. Populate realistic new content without inheriting client-specific claims.
2. Bind semantic colors/assets and define supported cardinality and text envelopes.
3. Render and measure in native PowerPoint, reusing `samples/visual-wave3`.
4. Review every changed slide for typography, margins, relationships and source art.
5. Record supported transformations and explicit failure behavior before promotion.
