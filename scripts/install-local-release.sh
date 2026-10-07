#!/usr/bin/env bash
set -euo pipefail
export PYTHONDONTWRITEBYTECODE=1
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
published_version="$(cat "$repo_root/release/VERSION")"
version="$published_version"
published_bundle="$(cat "$repo_root/release/default-bundle.txt")"
documentation_input="$repo_root/wmds-docs/site"
if [[ -f "$repo_root/release/default-docs.txt" ]]; then
  documentation_relative="$(cat "$repo_root/release/default-docs.txt")"
  documentation_input="$repo_root/$documentation_relative"
fi
bundle_revision="$published_bundle"
catalog_input="$repo_root/library/wm-design-system/$published_bundle/catalog"
verification_input=""
stage_only=false
install_skill=true
stage_destination=""
release_parent="${HOME}/.local/share/pptxgengo/releases"
launcher="${HOME}/.local/bin/pptxgengo"
skill_link="${CODEX_HOME:-${HOME}/.codex}/skills/west-monroe-presentations"
usage() { echo "usage: scripts/install-local-release.sh [--cli-only] [--stage-only NEW_DIRECTORY] [--bundle $published_bundle] [--catalog DIRECTORY] [--version VERSION] [--verification RECEIPT]" >&2; }
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
if [[ ! "$published_bundle" =~ ^v[1-9][0-9]*$ || "$bundle_revision" != "$published_bundle" ]]; then echo "only the current published library is packaged" >&2; exit 1; fi
bundle_input="$repo_root/library/wm-design-system/$bundle_revision"
if [[ ! "$version" =~ ^0\.1\.0-local\.[0-9]+(-candidate)?$ && ! "$version" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$ ]]; then echo "invalid package version: $version" >&2; exit 1; fi
catalog_input="$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$catalog_input")"
if [[ "$stage_only" == false && ( "$version" != "$published_version" || "$catalog_input" != "$bundle_input/catalog" || -n "$verification_input" ) ]]; then
  echo "version/catalog/verification overrides require --stage-only" >&2; exit 1
fi
release_dir="$release_parent/$version"
if [[ "$stage_only" == true ]]; then
  release_dir="$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$stage_destination")"
  release_parent="$(dirname "$release_dir")"
fi
if [[ -e "$release_dir" || -L "$release_dir" ]]; then echo "release already exists: $release_dir" >&2; exit 1; fi
if [[ "$stage_only" == false && -e "$launcher" && ! -L "$launcher" ]]; then echo "launcher exists and is not a symlink: $launcher" >&2; exit 1; fi

# Gallery labels identify the original export; source pins identify the library.
python3 - "$bundle_input" "$catalog_input" "$verification_input" "$bundle_revision" "$repo_root" "$documentation_input" <<'PY'
import json, sys
from pathlib import Path
sys.path.insert(0, str(Path(sys.argv[5]) / 'release'))
from package_files import file_digest, validate_design_docs
bundle_root, catalog = map(Path, sys.argv[1:3])
bundle = json.loads((bundle_root / 'bundle.json').read_text())
index = json.loads((catalog / 'design-system/index.json').read_text())
landing = json.loads((catalog / 'index.json').read_text())
source = json.loads((bundle_root / 'source/templates/catalog.json').read_text())['templates']
documentation = Path(sys.argv[6]).resolve()
if not documentation.is_relative_to(Path(sys.argv[5]).resolve()):
    raise SystemExit('release documentation must be inside the repository')
validate_design_docs(documentation, bundle_root)
count = bundle['template_count']
source_status = {r['key']: r.get('status', 'active') for r in source}
if bundle['source_revision'] != 'wmds-library.' + sys.argv[4] or count <= 0 or len(source) != count or len(source_status) != count:
    raise SystemExit('installer requires the current complete source bundle')
if index['entries'] != count or len(index['designs']) != count or len({r['template'] for r in index['designs']}) != count:
    raise SystemExit('gallery must contain one unique specimen for every source template')
if {r['template']: r['status'] for r in index['designs']} != source_status:
    raise SystemExit('gallery template identities or lifecycle states differ from source')
active = sum(status != 'deprecated' for status in source_status.values())
if index['active'] != active or index['deprecated'] != count - active or not (catalog / 'design-system.html').is_file():
    raise SystemExit('gallery counts or HTML missing')
for field in ('source_revision', 'source_commit'):
    if index[field] != bundle[field]: raise SystemExit(f'gallery source mismatch: {field}')
for field in ('source_revision', 'source_commit', 'entries', 'active', 'deprecated', 'qualification'):
    if landing['design_system'][field] != index[field]: raise SystemExit(f'gallery landing metadata mismatch: {field}')
for row in index['designs']:
    relative = Path(row['source_preview'])
    if relative.is_absolute() or '..' in relative.parts or not (catalog / relative).resolve().is_relative_to(catalog.resolve()) or row['native_review'] != 'reviewed_source_specimen':
        raise SystemExit(f'unreviewed or nonlocal source specimen: {row["template"]}')
    if file_digest(catalog / relative) != row['source_preview_sha256']:
        raise SystemExit(f'source preview drift: {row["template"]}')
if index['qualification']['reviewed_source_specimens'] != count or index['qualification']['arbitrary_content_qualified'] is not False:
    raise SystemExit('native specimen qualification count mismatch')
if sys.argv[3]:
    receipt = json.loads(Path(sys.argv[3]).read_text())
    if any(receipt[field] != bundle[field] for field in ('source_revision', 'source_commit')):
        raise SystemExit('verification source mismatch')
    if receipt['reviewed_pages'] != count or len(receipt['pages']) != count:
        raise SystemExit('verification page count mismatch')
    if {p['template']: p['sha256'] for p in receipt['pages']} != {r['template']: r['source_preview_sha256'] for r in index['designs']}:
        raise SystemExit('verification preview hashes differ')
PY
mkdir -p "$release_parent"
stage="$(mktemp -d "$release_parent/.${version}.stage.XXXXXXXX")"
trap 'rm -r "$stage"' EXIT
mkdir -p "$stage/bin"
cd "$repo_root"
for tool in pptxgengo pptxdesign wmdsdocs; do
  echo "building $tool" >&2
  go build -buildvcs=false -trimpath -ldflags "-s -w -X main.version=$version" -o "$stage/bin/$tool" "./cmd/$tool"
done
python3 - "$repo_root" "$stage" "$version" "$catalog_input" "$verification_input" "$bundle_revision" "$documentation_input" <<'PY'
import json, os, shutil, subprocess, sys
from pathlib import Path
src, dst = map(Path, sys.argv[1:3])
sys.path.insert(0, str(src / 'release'))
from package_files import clone_or_copy, file_digest, validate_asset_gallery
version, selected_catalog = sys.argv[3], Path(sys.argv[4])
bundle_relative = Path('library/wm-design-system') / sys.argv[6]
bundle = dst / bundle_relative
shutil.copytree(src / bundle_relative, bundle, ignore=shutil.ignore_patterns('catalog', 'library.sqlite'))
shutil.copytree(selected_catalog, bundle / 'catalog',
                ignore=lambda path, names: ['assets'] if Path(path) == selected_catalog else [])
for path in ('skills/west-monroe-presentations', 'schemas', 'examples/deck-project', 'examples/local-composition'):
    shutil.copytree(src / path, dst / path)
shutil.copytree(Path(sys.argv[7]), dst / 'wmds-docs/site')
shutil.copy2(src / 'wmds-docs/README.md', dst / 'wmds-docs/README.md')
for path in ('release/README.md', 'release/VERSION', 'release/package_files.py', 'library/README.md', 'cmd/pptxdesign/README.md',
             'internal/deckproject/README.md', 'docs/semantic-template-discovery.md',
             'docs/engineering-cli.md', 'docs/engineering-waves.md', 'docs/engineering-new-templates.md',
             'docs/engineering-followups.md', 'docs/testing.md', 'docs/installation.md', 'docs/local-agent-macos-recovery.md',
             'docs/skill-planning/deck-source-contract.md', 'docs/skill-planning/catalog-discovery-contract.md',
             'scripts/export-powerpoint.applescript',
             'scripts/render-pdf.swift', 'scripts/render-contact-sheet.swift'):
    target = dst / path
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src / path, target)
if sys.argv[5]: shutil.copy2(sys.argv[5], dst / f'release/verification-wmds-{sys.argv[6]}.json')
(dst / 'VERSION').write_text(version + '\n')
(dst / 'release/VERSION').write_text(version + '\n')
(dst / 'release/default-bundle.txt').write_text(sys.argv[6] + '\n')
# Package registered artwork once for authoring outside the checkout.
assets = json.loads(subprocess.run([str(dst / 'bin/pptxdesign'), 'asset-catalog'], check=True, capture_output=True, text=True).stdout)
branding_root = Path(os.environ.get('WMDS_BRANDING_ROOT') or Path.home() / 'Documents/branding')
seen = {}
copy_methods = {'cloned': 0, 'copied': 0}
for asset in assets:
    relative = Path(asset['path'])
    target = (dst / 'branding' / relative).resolve()
    if relative.is_absolute() or str(relative) != asset['path'] or not target.is_relative_to(dst.resolve()):
        raise ValueError(f'nonlocal registered artwork: {relative}')
    if relative in seen:
        if seen[relative] != asset['sha256']: raise ValueError(f'conflicting artwork: {relative}')
        continue
    payload = branding_root / relative
    if file_digest(payload) != asset['sha256']: raise ValueError(f'artwork hash mismatch: {relative}')
    target.parent.mkdir(parents=True, exist_ok=True)
    method = clone_or_copy(payload, target)
    if file_digest(target) != asset['sha256']: raise ValueError(f'packaged artwork hash mismatch: {relative}')
    copy_methods[method] += 1
    seen[relative] = asset['sha256']
(dst / 'branding/registry.json').write_text(json.dumps(assets, indent=2) + '\n')
photo_snapshot = json.loads((src / 'internal/wmdesign/photo_registry.json').read_text())
(dst / 'branding/photo-registry.json').write_text(json.dumps(photo_snapshot, indent=2) + '\n')
asset_gallery = selected_catalog / 'assets'
if asset_gallery.exists():
    concepts, variants = validate_asset_gallery(asset_gallery, assets, photo_snapshot)
    shutil.copytree(asset_gallery, bundle / 'catalog/assets')
    validate_asset_gallery(bundle / 'catalog/assets', assets, photo_snapshot)
    print(f'reused verified asset gallery: {concepts} concepts, {variants} variants')
print(f'packaged artwork: {copy_methods["cloned"]} APFS clones, {copy_methods["copied"]} copies')
PY
"$stage/bin/pptxdesign" library-index --bundle "$stage/library/wm-design-system/$bundle_revision" --gallery "$stage/library/wm-design-system/$bundle_revision/catalog" --out "$stage/library/wm-design-system/$bundle_revision/library.sqlite" >/dev/null
if [[ ! -d "$stage/library/wm-design-system/$bundle_revision/catalog/assets" ]]; then
  WMDS_BRANDING_ROOT="$stage/branding" "$stage/bin/pptxdesign" asset-gallery --out "$stage/library/wm-design-system/$bundle_revision/catalog/assets" >/dev/null
fi
python3 - "$stage" "$version" "$release_dir" "$bundle_revision" <<'PY'
import platform
from collections import Counter
import json, sqlite3, subprocess, sys
from pathlib import Path
root, version, destination = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3])
sys.path.insert(0, str(root / 'release'))
from package_files import file_digest, validate_asset_gallery, validate_design_docs
bundle_relative = Path('library/wm-design-system') / sys.argv[4]
bundle = root / bundle_relative
catalog, index_path = bundle / 'catalog', bundle / 'library.sqlite'
gallery = json.loads((catalog / 'design-system/index.json').read_text())
assets = json.loads((root / 'branding/registry.json').read_text())
photo_snapshot = json.loads((root / 'branding/photo-registry.json').read_text())
concepts, variants = validate_asset_gallery(catalog / 'assets', assets, photo_snapshot)
result = subprocess.run([str(root / 'bin/pptxdesign'), 'library-find', '--index', str(index_path), '--query', 'weekly status', '--kinds', 'template', '--summary'], cwd='/tmp', capture_output=True, text=True, check=True)
if not json.loads(result.stdout)['matches']: raise SystemExit('packaged discovery has no results')
with sqlite3.connect(index_path) as db:
    artifacts = list(db.execute('SELECT entity_id,role,path,sha256 FROM artifacts'))
    expected_artifacts = []
    for row in gallery['designs']:
        for role in ('source_preview', 'alternate_preview', 'source_values', 'alternate_values', 'source_foundation', 'alternate_foundation'):
            if row.get(role):
                relative = Path(row[role])
                if relative.is_absolute() or '..' in relative.parts or not (catalog / relative).resolve().is_relative_to(catalog.resolve()):
                    raise SystemExit(f'nonlocal gallery artifact: {relative}')
                expected_artifacts.append(('wmds/template/' + row['template'], role, row[role], file_digest(catalog / relative)))
    if Counter(artifacts) != Counter(expected_artifacts): raise SystemExit('indexed artifacts differ from gallery links')
    for _, _, path, expected in artifacts:
        if file_digest(catalog / path) != expected: raise SystemExit(f'artifact drift: {path}')
    # Persist final roots before the atomic move; staging roots are temporary.
    report = json.loads(db.execute("SELECT value FROM meta WHERE key='report'").fetchone()[0])
    report['path'] = str(destination / bundle_relative / 'library.sqlite')
    report['options']['bundle'] = str(destination / bundle_relative)
    report['options']['gallery'] = str(destination / bundle_relative / 'catalog')
    db.execute("UPDATE meta SET value=? WHERE key='report'", (json.dumps(report),))
source = json.loads((bundle / 'bundle.json').read_text())
docs_source = validate_design_docs(root / 'wmds-docs/site', bundle)
files = {str(p.relative_to(root)): file_digest(p) for p in sorted(root.rglob('*')) if p.is_file()}
(root / 'release-manifest.json').write_text(json.dumps({
    'schema': 'pptxgengo.local-release-manifest.v1', 'version': version, 'selected_bundle': sys.argv[4],
    'target_os': {'Darwin': 'darwin', 'Linux': 'linux'}.get(platform.system(), platform.system().lower()),
    'target_arch': {'aarch64': 'arm64', 'x86_64': 'amd64'}.get(platform.machine(), platform.machine()),
    'source_revision': source['source_revision'], 'source_commit': source['source_commit'],
    'template_count': source['template_count'], 'active_template_count': gallery['active'],
    'native_reviewed_source_specimens': gallery['qualification']['reviewed_source_specimens'],
    'asset_concepts': concepts, 'asset_variants': variants, 'photography_originals': len(photo_snapshot['photos']),
    'documentation_source_commit': docs_source['commit'], 'documentation_templates': docs_source['templates'],
    'qualification': 'reviewed_source_specimens', 'file_count': len(files), 'files_sha256': files,
}, indent=2) + '\n')
print(f'validated {sys.argv[4]}: {source["template_count"]} designs, {len(artifacts)} artifact links, {len(files)} package files')
PY
mv "$stage" "$release_dir"
trap - EXIT
if [[ "$stage_only" == true ]]; then echo "staged $version at $release_dir"; exit 0; fi
activation_args=(installation install --from "$release_dir" --root "$(dirname "$release_parent")" --bin-dir "$(dirname "$launcher")" --skill-dir "$skill_link")
if [[ "$install_skill" == false ]]; then activation_args+=(--skip-skill); fi
"$release_dir/bin/pptxgengo" "${activation_args[@]}"
echo "installed $version at $release_dir; open a new terminal and run pptxgengo installation doctor"
