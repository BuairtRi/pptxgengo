#!/usr/bin/env bash
set -euo pipefail

version=0.1.0-local.3
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
release_parent="${HOME}/.local/share/pptxgengo/releases"
release_dir="${release_parent}/${version}"
launcher="${HOME}/.local/bin/pptxgengo"
skill_link="${HOME}/.codex/skills/west-monroe-presentations"

if [[ -e "$release_dir" ]]; then
  echo "release already exists: $release_dir" >&2
  exit 1
fi
if [[ -e "$launcher" && ! -L "$launcher" ]]; then
  echo "launcher exists and is not a symlink: $launcher" >&2
  exit 1
fi
if [[ -e "$skill_link" && ! -L "$skill_link" ]]; then
  echo "skill path exists and is not a symlink: $skill_link" >&2
  exit 1
fi
if [[ ! -f "$repo_root/release/catalog/index.html" ]]; then
  echo "missing release/catalog/index.html" >&2
  exit 1
fi
python3 - "$repo_root/release/catalog/index.json" "$repo_root/release/catalog/index.html" "$repo_root/release/README.md" "$version" <<'PY'
import json
from pathlib import Path
import sys
index, html, readme = map(Path, sys.argv[1:4])
version = sys.argv[4]
if json.loads(index.read_text()).get('version') != version or version not in html.read_text() or version not in readme.read_text():
    raise SystemExit(f'catalog and release README must declare {version}')
PY

mkdir -p "$release_parent" "$(dirname "$launcher")"
stage="$(mktemp -d "$release_parent/.${version}.stage.XXXXXXXX")"
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/bin"

cd "$repo_root"
for tool in pptxgengo pptxtemplate pptxcompose pptxscene pptxcomponent pptxlib pptxanchor pptxdiff pptxadapt; do
  echo "building $tool" >&2
  if [[ "$tool" == pptxgengo ]]; then
    go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/bin/$tool" "./cmd/$tool"
  else
    go build -trimpath -ldflags '-s -w' -o "$stage/bin/$tool" "./cmd/$tool"
  fi
done

python3 - "$repo_root" "$stage" "$version" <<'PY'
import hashlib
import json
from pathlib import Path
import shutil
import sys

src, dst = map(Path, sys.argv[1:3])
version = sys.argv[3]

def copy(path):
    path = Path(path)
    if path.is_absolute() or '..' in path.parts:
        raise ValueError(f'nonlocal release input: {path}')
    source, target = src / path, dst / path
    if not source.is_file():
        raise FileNotFoundError(source)
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, target)

shutil.copytree(src / 'library', dst / 'library')
shutil.copytree(src / 'release/catalog', dst / 'catalog')
shutil.copytree(src / 'skills/west-monroe-presentations', dst / 'skills/west-monroe-presentations')
shutil.copytree(src / 'planning/release-0.1', dst / 'planning/release-0.1')
shutil.copytree(src / 'planning/adaptive', dst / 'planning/adaptive')
for release_doc in ('release/README.md', 'release/cleanup.json'):
    if (src / release_doc).is_file():
        copy(release_doc)

assignments = json.loads((src / 'library/templates/rollout/assignments.json').read_text())['entries']
for project in sorted({row['source_project'] for row in assignments}):
    shutil.copytree(src / project, dst / project)
for row in assignments:
    metadata = json.loads((src / row['implementation_directory'] / 'implementation.json').read_text())
    copy(metadata['preview']['path'])

manifest = json.loads((src / 'library/catalog-manifest.json').read_text())
for artifact in manifest['artifacts'].values():
    copy(artifact['path'])
catalog = json.loads((src / 'library/templates/catalog.json').read_text())
for artifact in catalog['inputs']:
    copy(artifact['path'])
copy('planning/source-registry.json')

def artifact_paths(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key.endswith('path') and isinstance(child, str) and child.startswith(('library/', 'samples/')):
                yield child
            yield from artifact_paths(child)
    elif isinstance(value, list):
        for child in value:
            yield from artifact_paths(child)

for contract in (src / 'library/contracts').glob('*.json'):
    for path in artifact_paths(json.loads(contract.read_text())):
        copy(path)
for recipe in (src / 'library').rglob('*.json'):
    try:
        doc = json.loads(recipe.read_text())
    except (ValueError, UnicodeDecodeError):
        continue
    if isinstance(doc, dict) and doc.get('schema') == 'pptxgengo.compose-spec.v1':
        for path in artifact_paths(doc):
            copy(path)

# Keep the exact accepted adaptive examples inspectable in the next release.
# The checkpoint names the required evidence; no broad samples tree is copied.
checkpoint = json.loads((src / 'library/adaptive/checkpoint.json').read_text())
evidence_paths = {}
def add_evidence(path, expected_hash=None):
    if (not isinstance(path, str) or '\\' in path or
            any(part in ('', '.', '..') for part in path.split('/')) or
            not path.startswith(('library/', 'samples/adaptive/'))):
        raise ValueError(f'invalid adaptive evidence path: {path!r}')
    prior = evidence_paths.get(path)
    if prior and expected_hash and prior != expected_hash:
        raise ValueError(f'conflicting adaptive evidence hash: {path}')
    evidence_paths[path] = expected_hash or prior

example_groups = list(checkpoint['families'].values())
example_groups.extend(checkpoint.get('accents', {}).values())
for group in example_groups:
    for example in group.get('examples', []):
        hashes = example.get('hashes', {})
        evidence = example.get('evidence', {})
        add_evidence(example['spec_path'], hashes.get('spec_sha256'))
        if example.get('values_path'):
            add_evidence(example['values_path'], hashes.get('values_sha256'))
        if example.get('render_png'):
            add_evidence(example['render_png'], hashes.get('render_sha256'))
        for name, path in evidence.items():
            if isinstance(path, str):
                add_evidence(path, hashes.get(name + '_sha256'))

for path, expected_hash in sorted(evidence_paths.items()):
    if expected_hash:
        actual_hash = hashlib.sha256((src / path).read_bytes()).hexdigest()
        if actual_hash != expected_hash:
            raise ValueError(f'adaptive evidence hash mismatch: {path}')
    copy(path)

# The capability records link to these exact implementation schema sources.
capabilities = json.loads((src / 'library/adaptive/capabilities.json').read_text())
for family in capabilities['family_builders'].values():
    path = family['schema_path']
    if not isinstance(path, str) or not path.startswith('internal/adapt/') or not path.endswith('.go'):
        raise ValueError(f'invalid adaptive schema source path: {path!r}')
    copy(path)

for path in [
    'scripts/adapt-template-accents.py',
    'scripts/export-powerpoint.applescript',
    'scripts/measure-pptx-text.applescript',
    'scripts/extract-emf-bitmap.py',
    'scripts/svg-visible-bounds.swift',
    'scripts/measure-compose-text.applescript',
    'scripts/compose-environment.swift',
    'scripts/measure-template-frames.applescript',
    'scripts/check-template-fit.py',
    'scripts/render-template-review.py',
    'scripts/build-template-review-gallery.py',
    'scripts/render-open-accent-review.py',
    'scripts/render-pdf.swift',
]:
    copy(path)
(dst / 'VERSION').write_text(version + '\n')
PY

"$stage/bin/pptxlib" index --root "$stage" --out "$stage/library/catalog-library.sqlite" >/dev/null

mv "$stage" "$release_dir"
trap - EXIT
link_stage="${launcher}.tmp.$$"
ln -s "$release_dir/bin/pptxgengo" "$link_stage"
mv -f "$link_stage" "$launcher"
mkdir -p "$(dirname "$skill_link")"
skill_link_stage="${skill_link}.tmp.$$"
ln -s "$release_dir/skills/west-monroe-presentations" "$skill_link_stage"
mv -fh "$skill_link_stage" "$skill_link"
echo "installed $version at $release_dir"
echo "launcher: $launcher"
echo "skill: $skill_link"
