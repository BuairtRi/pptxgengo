#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version="$(cat "$repo_root/release/VERSION")"
if [[ ! "$version" =~ ^0\.1\.0-local\.[0-9]+$ ]]; then
  echo "invalid release/VERSION: $version" >&2
  exit 1
fi
release_parent="${HOME}/.local/share/pptxgengo/releases"
release_dir="${release_parent}/${version}"
launcher="${HOME}/.local/bin/pptxgengo"
skill_link="${HOME}/.codex/skills/west-monroe-presentations"
stage_only=false
if [[ $# -gt 0 ]]; then
  if [[ $# -ne 2 || "$1" != "--stage-only" ]]; then
    echo "usage: scripts/install-local-release.sh [--stage-only NEW_DIRECTORY]" >&2
    exit 1
  fi
  stage_only=true
  release_dir="$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$2")"
  release_parent="$(dirname "$release_dir")"
fi

if [[ -e "$release_dir" || -L "$release_dir" ]]; then
  echo "release already exists: $release_dir" >&2
  exit 1
fi
if [[ "$stage_only" == false && -e "$launcher" && ! -L "$launcher" ]]; then
  echo "launcher exists and is not a symlink: $launcher" >&2
  exit 1
fi
if [[ "$stage_only" == false && -e "$skill_link" && ! -L "$skill_link" ]]; then
  echo "skill path exists and is not a symlink: $skill_link" >&2
  exit 1
fi
python3 - "$repo_root/release/catalog" "$repo_root/release/README.md" "$version" <<'PY'
import json
from pathlib import Path
import sys
catalog, readme = map(Path, sys.argv[1:3])
version = sys.argv[3]
pages = [catalog / name for name in ('index.html', 'templates.html', 'components.html')]
if any(not p.is_file() for p in [catalog / 'index.json', *pages]):
    raise SystemExit('catalog landing, both galleries, and index.json are required')
if json.loads((catalog / 'index.json').read_text()).get('version') != version or any(version not in p.read_text() for p in pages) or version not in readme.read_text():
    raise SystemExit(f'catalog and release README must declare {version}')
PY

mkdir -p "$release_parent"
if [[ "$stage_only" == false ]]; then
  mkdir -p "$(dirname "$launcher")"
fi
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
shutil.copytree(src / 'planning/requested-templates', dst / 'planning/requested-templates')
for release_doc in ('release/README.md', 'release/cleanup.json', 'release/verification.json'):
    if (src / release_doc).is_file():
        copy(release_doc)

assignments = json.loads((src / 'library/templates/rollout/assignments.json').read_text())['entries']
for project in sorted({row['source_project'] for row in assignments}):
    shutil.copytree(src / project, dst / project)
for row in assignments:
    metadata = json.loads((src / row['implementation_directory'] / 'implementation.json').read_text())
    copy(metadata['preview']['path'])

# Retain the exact two user-requested sources and the reviewed native gauge
# example as hash-pinned evidence, without copying entire samples trees.
request = json.loads((src / 'planning/requested-templates/request.json').read_text())
for source in request['sources']:
    path = source['path']
    if hashlib.sha256((src / path).read_bytes()).hexdigest() != source['sha256']:
        raise ValueError(f'requested source hash mismatch: {path}')
    copy(path)
gauge_review = json.loads((src / 'planning/requested-templates/gauge-review.json').read_text())
for artifact in gauge_review['artifacts']:
    path = artifact['path']
    if hashlib.sha256((src / path).read_bytes()).hexdigest() != artifact['sha256']:
        raise ValueError(f'gauge evidence hash mismatch: {path}')
    copy(path)

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

# Retain the bounded font example's measurements and rendered review alongside
# the authoring fixture. Its evidence is separate from older template proofs.
font_proof_path = src / 'library/dynamic-components/fonts-proof.json'
if font_proof_path.is_file():
    font_proof = json.loads(font_proof_path.read_text())
    for artifact in font_proof['artifacts']:
        path = artifact['path']
        actual_hash = hashlib.sha256((src / path).read_bytes()).hexdigest()
        if actual_hash != artifact['sha256']:
            raise ValueError(f'font evidence hash mismatch: {path}')
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
    'scripts/build-release-catalog.py',
    'scripts/build-requested-template-gallery.py',
    'scripts/import-requested-templates.py',
]:
    copy(path)
(dst / 'VERSION').write_text(version + '\n')
PY

"$stage/bin/pptxlib" index --root "$stage" --out "$stage/library/catalog-library.sqlite" >/dev/null

python3 - "$stage" "$version" <<'PY'
import hashlib
from html.parser import HTMLParser
import json
from pathlib import Path
import re
import subprocess
import sys
from urllib.parse import unquote, urlsplit

root = Path(sys.argv[1])
version = sys.argv[2]
assignments = json.loads((root / 'library/templates/rollout/assignments.json').read_text())
if assignments['target'] != 101 or len(assignments['entries']) != 101:
    raise SystemExit('release requires exactly 101 registered templates')
catalog = root / 'catalog'
index = json.loads((catalog / 'index.json').read_text())
if index.get('version') != version or len(index.get('templates', [])) != 101 or len(index.get('components', [])) != 132:
    raise SystemExit('release catalog version/template/component count mismatch')
if sum(row.get('contract_available') is True for row in index['components']) != 124:
    raise SystemExit('release gallery must identify all 124 executable component contracts')
requested = json.loads((root / 'library/templates/requested-templates.json').read_text())
components = json.loads((root / 'library/templates/requested-components.json').read_text())
if len(requested['templates']) != 44 or len(components['components']) != 132:
    raise SystemExit('requested template/component catalog mismatch')
if sum('contract' in row for row in components['components']) != 124:
    raise SystemExit('release requires 124 executable component subcontracts')
for row in components['components']:
    if row.get('contract') and not (root / row['contract']).is_file():
        raise SystemExit(f"missing component contract: {row['contract']}")
result = subprocess.run([str(root / 'bin/pptxtemplate'), 'list', '--root', str(root)], capture_output=True, text=True, check=True)
rows = json.loads(result.stdout)
bad = [row for row in rows if row.get('status') != 'bindings_inspected' or row.get('issue')]
if len(rows) != 101 or bad:
    details = '; '.join(f"{row.get('id')}: {row.get('status')}: {row.get('issue')}" for row in bad[:12])
    raise SystemExit(f'packaged template discovery is incomplete ({len(rows)}/101; {len(bad)} invalid): {details}')

class LocalLinks(HTMLParser):
    def __init__(self, page):
        super().__init__()
        self.page = page
    def handle_starttag(self, tag, attrs):
        for key, value in attrs:
            if key not in ('href', 'src') or not value:
                continue
            parsed = urlsplit(value)
            if parsed.scheme or parsed.netloc or value.startswith('#'):
                continue
            target = (self.page.parent / unquote(parsed.path)).resolve()
            if not target.is_relative_to(catalog.resolve()) or not target.is_file():
                raise SystemExit(f'broken catalog local asset: {self.page.name}: {value}')

for name in ('index.html', 'templates.html', 'components.html'):
    page = catalog / name
    if not page.is_file():
        raise SystemExit(f'missing gallery page {name}')
    LocalLinks(page).feed(page.read_text())
def catalog_assets(value):
    if isinstance(value, dict):
        for child in value.values():
            yield from catalog_assets(child)
    elif isinstance(value, list):
        for child in value:
            yield from catalog_assets(child)
    elif isinstance(value, str) and value.startswith(('templates/', 'components/', 'assets/', 'compose/', 'recipes/')):
        yield value

for value in catalog_assets(index):
    path = value.split('?', 1)[0].split('#', 1)[0]
    if not (catalog / path).is_file():
        raise SystemExit(f'broken catalog index asset: {path}')

skill = root / 'skills/west-monroe-presentations'
if not (skill / 'SKILL.md').is_file():
    raise SystemExit('missing packaged skill')
for name in ('template-authoring.md', 'component-customization.md', 'compose-authoring.md'):
    if not (skill / 'references' / name).is_file():
        raise SystemExit(f'missing skill reference {name}')
for doc in skill.rglob('*.md'):
    for link in re.findall(r'\]\(([^)]+)\)', doc.read_text()):
        parsed = urlsplit(link)
        if parsed.scheme or parsed.netloc or link.startswith('#'):
            continue
        target = (doc.parent / unquote(parsed.path)).resolve()
        if not target.is_relative_to(skill.resolve()) or not target.is_file():
            raise SystemExit(f'broken skill link: {doc.relative_to(root)}: {link}')

files = {}
for path in sorted(root.rglob('*')):
    if path.is_file():
        files[str(path.relative_to(root))] = hashlib.sha256(path.read_bytes()).hexdigest()
(root / 'release-manifest.json').write_text(json.dumps({
    'schema': 'pptxgengo.local-release-manifest.v1',
    'version': version,
    'file_count': len(files),
    'files_sha256': files,
}, indent=2) + '\n')
print(f'validated {len(rows)} templates, {len(components["components"])} component occurrences, {len(files)} release files')
PY

mv "$stage" "$release_dir"
trap - EXIT
if [[ "$stage_only" == true ]]; then
  echo "staged $version at $release_dir"
  exit 0
fi
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
