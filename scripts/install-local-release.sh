#!/usr/bin/env bash
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
published_version="$(cat "$repo_root/release/VERSION")"
version="$published_version"
bundle_revision=v5
catalog_input="$repo_root/library/wm-design-system/v5/catalog"
verification_input=""
stage_only=false
install_skill=true
stage_destination=""
release_parent="${HOME}/.local/share/pptxgengo/releases"
launcher="${HOME}/.local/bin/pptxgengo"
skill_link="${HOME}/.codex/skills/west-monroe-presentations"
usage() { echo "usage: scripts/install-local-release.sh [--cli-only] [--stage-only NEW_DIRECTORY] [--bundle v5] [--catalog DIRECTORY] [--version VERSION] [--verification RECEIPT]" >&2; }
while [[ $# -gt 0 ]]; do
  if [[ "$1" == --cli-only ]]; then install_skill=false; shift; continue; fi
  if [[ $# -lt 2 ]]; then usage; exit 1; fi
  case "$1" in
    --stage-only) stage_only=true; stage_destination="$2" ;;
    --bundle) bundle_revision="$2" ;;
    --catalog) catalog_input="$2" ;;
    --version) version="$2" ;;
    --verification) verification_input="$2" ;;
    *) usage; exit 1 ;;
  esac
  shift 2
done
if [[ "$bundle_revision" != v5 ]]; then echo "only the current v5 library is packaged" >&2; exit 1; fi
if [[ ! "$version" =~ ^0\.1\.0-local\.[0-9]+(-candidate)?$ ]]; then echo "invalid package version: $version" >&2; exit 1; fi
catalog_input="$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$catalog_input")"
if [[ "$stage_only" == false && ( "$version" != "$published_version" || "$catalog_input" != "$repo_root/library/wm-design-system/v5/catalog" || -n "$verification_input" ) ]]; then
  echo "version/catalog/verification overrides require --stage-only" >&2; exit 1
fi
release_dir="$release_parent/$version"
if [[ "$stage_only" == true ]]; then
  release_dir="$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$stage_destination")"
  release_parent="$(dirname "$release_dir")"
fi
if [[ -e "$release_dir" || -L "$release_dir" ]]; then echo "release already exists: $release_dir" >&2; exit 1; fi
if [[ "$stage_only" == false && -e "$launcher" && ! -L "$launcher" ]]; then echo "launcher exists and is not a symlink: $launcher" >&2; exit 1; fi
if [[ "$stage_only" == false && "$install_skill" == true && -e "$skill_link" && ! -L "$skill_link" ]]; then echo "skill path exists and is not a symlink: $skill_link" >&2; exit 1; fi

# Gallery labels identify the original export; source pins identify the library.
python3 - "$repo_root/library/wm-design-system/v5" "$catalog_input" "$verification_input" <<'PY'
import hashlib, json, sys
from pathlib import Path
bundle_root, catalog = map(Path, sys.argv[1:3])
bundle = json.loads((bundle_root / 'bundle.json').read_text())
index = json.loads((catalog / 'design-system/index.json').read_text())
landing = json.loads((catalog / 'index.json').read_text())
if bundle['source_revision'] != 'wmds-library.v5' or bundle['template_count'] != 587:
    raise SystemExit('installer requires the current v5 bundle')
if index['entries'] != 587 or len(index['designs']) != 587 or len({r['template'] for r in index['designs']}) != 587:
    raise SystemExit('gallery must contain 587 unique source specimens')
if index['active'] != 586 or index['deprecated'] != 1 or not (catalog / 'design-system.html').is_file():
    raise SystemExit('gallery counts or HTML missing')
for field in ('source_revision', 'source_commit'):
    if index[field] != bundle[field] or landing['design_system'][field] != bundle[field]:
        raise SystemExit(f'gallery source mismatch: {field}')
for row in index['designs']:
    relative = Path(row['source_preview'])
    if relative.is_absolute() or '..' in relative.parts or row['native_review'] != 'reviewed_source_specimen':
        raise SystemExit(f'unreviewed or nonlocal source specimen: {row["template"]}')
    if hashlib.sha256((catalog / relative).read_bytes()).hexdigest() != row['source_preview_sha256']:
        raise SystemExit(f'source preview drift: {row["template"]}')
if index['qualification']['reviewed_source_specimens'] != 587:
    raise SystemExit('native specimen qualification count mismatch')
if sys.argv[3]:
    receipt = json.loads(Path(sys.argv[3]).read_text())
    if any(receipt[field] != bundle[field] for field in ('source_revision', 'source_commit')):
        raise SystemExit('verification source mismatch')
    if receipt['reviewed_pages'] != 587 or len(receipt['pages']) != 587:
        raise SystemExit('verification page count mismatch')
    if {p['template']: p['sha256'] for p in receipt['pages']} != {r['template']: r['source_preview_sha256'] for r in index['designs']}:
        raise SystemExit('verification preview hashes differ')
PY
mkdir -p "$release_parent"
stage="$(mktemp -d "$release_parent/.${version}.stage.XXXXXXXX")"
trap 'rm -r "$stage"' EXIT
mkdir -p "$stage/bin"
cd "$repo_root"
for tool in pptxgengo pptxdesign; do
  echo "building $tool" >&2
  if [[ "$tool" == pptxgengo ]]; then
    go build -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/bin/$tool" "./cmd/$tool"
  else
    go build -trimpath -ldflags '-s -w' -o "$stage/bin/$tool" "./cmd/$tool"
  fi
done
python3 - "$repo_root" "$stage" "$version" "$catalog_input" "$verification_input" <<'PY'
import hashlib, json, os, shutil, subprocess, sys
from pathlib import Path
src, dst = map(Path, sys.argv[1:3])
version, selected_catalog = sys.argv[3], Path(sys.argv[4])
bundle_relative = Path('library/wm-design-system/v5')
bundle = dst / bundle_relative
shutil.copytree(src / bundle_relative, bundle, ignore=shutil.ignore_patterns('catalog', 'library.sqlite'))
shutil.copytree(selected_catalog, bundle / 'catalog')
for path in ('skills/west-monroe-presentations', 'schemas', 'examples/deck-project', 'examples/local-composition'):
    shutil.copytree(src / path, dst / path)
for path in ('release/README.md', 'release/VERSION', 'library/README.md', 'cmd/pptxdesign/README.md',
             'internal/deckproject/README.md', 'docs/semantic-template-discovery.md',
             'docs/engineering-cli.md', 'docs/engineering-waves.md', 'docs/engineering-new-templates.md',
             'docs/engineering-followups.md', 'docs/testing.md', 'docs/local-agent-macos-recovery.md',
             'docs/skill-planning/deck-source-contract.md', 'docs/skill-planning/catalog-discovery-contract.md',
             'scripts/build-wmds-production-gallery.py', 'scripts/export-powerpoint.applescript',
             'scripts/render-pdf.swift', 'scripts/render-contact-sheet.swift'):
    target = dst / path
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src / path, target)
if sys.argv[5]: shutil.copy2(sys.argv[5], dst / 'release/verification-wmds-v5.json')
(dst / 'VERSION').write_text(version + '\n')
(dst / 'release/VERSION').write_text(version + '\n')
(dst / 'release/default-bundle.txt').write_text('v5\n')
# Package registered artwork once for authoring outside the checkout.
assets = json.loads(subprocess.run([str(dst / 'bin/pptxdesign'), 'asset-catalog'], check=True, capture_output=True, text=True).stdout)
branding_root = Path(os.environ.get('WMDS_BRANDING_ROOT') or Path.home() / 'Documents/branding')
seen = {}
for asset in assets:
    relative = Path(asset['path'])
    target = (dst / 'branding' / relative).resolve()
    if relative.is_absolute() or str(relative) != asset['path'] or not target.is_relative_to(dst.resolve()):
        raise ValueError(f'nonlocal registered artwork: {relative}')
    if relative in seen:
        if seen[relative] != asset['sha256']: raise ValueError(f'conflicting artwork: {relative}')
        continue
    payload = branding_root / relative
    if hashlib.sha256(payload.read_bytes()).hexdigest() != asset['sha256']: raise ValueError(f'artwork hash mismatch: {relative}')
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(payload, target)
    seen[relative] = asset['sha256']
(dst / 'branding/registry.json').write_text(json.dumps(assets, indent=2) + '\n')
PY
"$stage/bin/pptxdesign" library-index --bundle "$stage/library/wm-design-system/v5" --gallery "$stage/library/wm-design-system/v5/catalog" --out "$stage/library/wm-design-system/v5/library.sqlite" >/dev/null
WMDS_BRANDING_ROOT="$stage/branding" "$stage/bin/pptxdesign" asset-gallery --out "$stage/library/wm-design-system/v5/catalog/assets" >/dev/null
python3 - "$stage" "$version" "$release_dir" <<'PY'
import hashlib, json, sqlite3, subprocess, sys
from pathlib import Path
root, version, destination = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3])
bundle = root / 'library/wm-design-system/v5'
catalog, index_path = bundle / 'catalog', bundle / 'library.sqlite'
result = subprocess.run([str(root / 'bin/pptxdesign'), 'library-find', '--index', str(index_path), '--query', 'weekly status', '--kinds', 'template', '--summary'], cwd='/tmp', capture_output=True, text=True, check=True)
if not json.loads(result.stdout)['matches']: raise SystemExit('packaged discovery has no results')
with sqlite3.connect(index_path) as db:
    artifacts = list(db.execute('SELECT path,sha256 FROM artifacts'))
    if len(artifacts) != 1760: raise SystemExit('artifact count mismatch')
    for path, expected in artifacts:
        if hashlib.sha256((catalog / path).read_bytes()).hexdigest() != expected: raise SystemExit(f'artifact drift: {path}')
    # Persist final roots before the atomic move; staging roots are temporary.
    report = json.loads(db.execute("SELECT value FROM meta WHERE key='report'").fetchone()[0])
    report['path'] = str(destination / 'library/wm-design-system/v5/library.sqlite')
    report['options']['bundle'] = str(destination / 'library/wm-design-system/v5')
    report['options']['gallery'] = str(destination / 'library/wm-design-system/v5/catalog')
    db.execute("UPDATE meta SET value=? WHERE key='report'", (json.dumps(report),))
source = json.loads((bundle / 'bundle.json').read_text())
files = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
(root / 'release-manifest.json').write_text(json.dumps({
    'schema': 'pptxgengo.local-release-manifest.v1', 'version': version, 'selected_bundle': 'v5',
    'source_revision': source['source_revision'], 'source_commit': source['source_commit'],
    'template_count': 587, 'active_template_count': 586, 'native_reviewed_source_specimens': 587,
    'qualification': 'reviewed_source_specimens', 'file_count': len(files), 'files_sha256': files,
}, indent=2) + '\n')
print(f'validated v5: 587 designs, {len(artifacts)} artifact links, {len(files)} package files')
PY
mv "$stage" "$release_dir"
trap - EXIT
if [[ "$stage_only" == true ]]; then echo "staged $version at $release_dir"; exit 0; fi
mkdir -p "$(dirname "$launcher")"
ln -s "$release_dir/bin/pptxgengo" "${launcher}.tmp.$$"
mv -f "${launcher}.tmp.$$" "$launcher"
if [[ "$install_skill" == true ]]; then
  mkdir -p "$(dirname "$skill_link")"
  ln -s "$release_dir/skills/west-monroe-presentations" "${skill_link}.tmp.$$"
  mv -fh "${skill_link}.tmp.$$" "$skill_link"
fi
echo "installed $version at $release_dir"
