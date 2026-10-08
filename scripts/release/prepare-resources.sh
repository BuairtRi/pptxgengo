#!/usr/bin/env bash
# Resource inclusion is explicit and recorded in the signed release manifest.
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
policy="${PPTXGENGO_BROWSING_POLICY:-required}"
case "$policy" in required|deferred|templates-only) ;; *) echo 'Unknown browsing policy' >&2; exit 1 ;; esac
case "${PPTXGENGO_PACKAGE_KIND:-cli-only}" in full|cli-only) ;; *) echo 'Unknown package kind' >&2; exit 1 ;; esac
if [[ "$policy" == deferred ]]; then
  [[ "${PPTXGENGO_PACKAGE_KIND:-cli-only}" == cli-only ]] || { echo 'Deferred browsing is allowed only for CLI-only releases' >&2; exit 1; }
  release_ci prepare-resource-policy dist/resources "$release_version" "$release_commit"
  exit 0
fi
if [[ "$policy" == templates-only ]]; then
  [[ "${PPTXGENGO_PACKAGE_KIND:-cli-only}" == cli-only ]] || { echo 'Template-only browsing is supported only for CLI-only releases' >&2; exit 1; }
  export WMDS_BROWSING_AS_OF="${WMDS_BROWSING_AS_OF:-${CI_PIPELINE_CREATED_AT:0:10}}"
  [[ "${WMDS_BROWSING_AS_OF:-}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || { echo 'Set WMDS_BROWSING_AS_OF to the release review date (YYYY-MM-DD)' >&2; exit 1; }
  [[ -n "${CI_PIPELINE_CREATED_AT:-}" && "$WMDS_BROWSING_AS_OF" == "${CI_PIPELINE_CREATED_AT:0:10}" ]] || { echo 'Template catalog date must match the authoritative GitLab pipeline creation date' >&2; exit 1; }
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  bundle_revision="$(cat release/default-bundle.txt)"
  bundle_source="library/wm-design-system/$bundle_revision"
  go build -buildvcs=false -trimpath -ldflags "-X main.version=$release_version -X main.releaseIdentity=$release_commit" -o "$tmp/pptxdesign" ./cmd/pptxdesign
  "$tmp/pptxdesign" browsing-library --kind templates --frames "${WMDS_BROWSING_FRAMES:-catalog}" --media-policy placeholders --bundle "$bundle_source" --as-of "${WMDS_BROWSING_AS_OF:-${CI_PIPELINE_CREATED_AT:0:10}}" --out "$tmp/template-browsing"
  mkdir -p dist/resources/browsing "dist/resources/$bundle_source"
  cp "$tmp/template-browsing/template-library.pptx" dist/resources/browsing/
  cp "$tmp/template-browsing/browsing-manifest.json" dist/resources/browsing/template-library.manifest.json
  dest="dist/resources/$bundle_source"
  cp "$bundle_source/bundle.json" "$bundle_source/README.md" "$bundle_source/inventory.json" "$dest/"
  for part in source fonts typography assets; do [[ ! -e "$bundle_source/$part" ]] || cp -R "$bundle_source/$part" "$dest/"; done
  mkdir -p "$dest/catalog"
  cp "$bundle_source/catalog/index.json" "$bundle_source/catalog/publication-report.json" "$bundle_source/catalog/design-system.html" "$dest/catalog/"
  cp -R "$bundle_source/catalog/design-system" "$dest/catalog/"
  if [[ -f "$tmp/template-browsing/native-editing-coverage.json" ]]; then cp "$tmp/template-browsing/native-editing-coverage.json" dist/resources/browsing/; fi
  "$tmp/pptxdesign" library-index --bundle "$dest" --gallery "$dest/catalog" --out "$dest/library.sqlite" >/dev/null
  python3 - "$dest" "$bundle_revision" <<'PY'
import json, sqlite3, sys
from pathlib import Path
bundle=Path(sys.argv[1]); revision=sys.argv[2]
relative=Path('library/wm-design-system')/revision
with sqlite3.connect(bundle/'library.sqlite') as db:
    row=db.execute("SELECT value FROM meta WHERE key='report'").fetchone()
    if row is None: raise SystemExit('template catalog SQLite report missing')
    report=json.loads(row[0]); report['path']=(relative/'library.sqlite').as_posix()
    report['options']['bundle']=relative.as_posix(); report['options']['gallery']=(relative/'catalog').as_posix()
    db.execute("UPDATE meta SET value=? WHERE key='report'",(json.dumps(report,separators=(',',':')),))
PY
  # Confirm ordinary runtime discovery works after moving the complete source
  # bundle; queries must not depend on the build checkout or staging path.
  mkdir -p "$tmp/relocated/bin" "$tmp/relocated/library/wm-design-system"
  cp "$tmp/pptxdesign" "$tmp/relocated/bin/pptxdesign"
  cp -R "$dest" "$tmp/relocated/library/wm-design-system/$bundle_revision"
  (cd /tmp && "$tmp/relocated/bin/pptxdesign" library-find --query 'weekly status' --kinds template --summary > "$tmp/relocated-query.json")
  python3 - "$tmp/relocated-query.json" <<'PY'
import json,sys
if not json.load(open(sys.argv[1])).get('matches'): raise SystemExit('relocated template catalog search returned no results')
PY
  python3 - "$tmp/template-browsing/browsing-manifest.json" dist/resources <<'PY'
import hashlib, json, os, sys
from pathlib import Path
manifest=json.loads(Path(sys.argv[1]).read_text())
root=Path(sys.argv[2])
files={}
for subtree in ('browsing','library'):
    base=root/subtree
    for path in sorted(base.rglob('*')):
        if path.is_symlink(): raise SystemExit(f'linked resource refused: {path}')
        if path.is_file():
            rel=path.relative_to(root).as_posix()
            files[rel]=hashlib.sha256(path.read_bytes()).hexdigest()
bundle=Path(manifest['source_revision'].removeprefix('wmds-library.'))
out={'schema':'pptxgengo.template-catalog-release.v1','files_sha256':files,'bundle':bundle.name,'source_revision':manifest['source_revision'],'source_commit':manifest['source_commit'],'bundle_sha256':manifest['bundle_sha256'],'compiler':manifest['compiler'],'release_identity':manifest['release_identity'],'as_of':manifest['as_of'],'pipeline_created_at':os.environ['CI_PIPELINE_CREATED_AT'],'editing_profile':manifest['editing_profile'],'media_policy':manifest['media_policy']}
(root/'template-catalog-manifest.json').write_text(json.dumps(out,indent=2)+'\n')
PY
  release_ci prepare-resource-policy dist/resources "$release_version" "$release_commit"
  exit 0
fi
[[ "${WMDS_BRANDING_ARCHIVE_SHA256:-}" =~ ^[a-f0-9]{64}$ ]] || { echo 'Browsing-inclusive releases require immutable WMDS_BRANDING_ARCHIVE_SHA256' >&2; exit 1; }
[[ "${WMDS_BRANDING_ARCHIVE_URL:-}" == https://* ]] || { echo 'Private branding input requires WMDS_BRANDING_ARCHIVE_URL using HTTPS' >&2; exit 1; }
[[ "${WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256:-}" =~ ^[a-f0-9]{64}$ ]] || { echo 'Reusable browsing deck requires immutable WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256 and real operator-approved slide revisions' >&2; exit 1; }
[[ "${WMDS_FINISHED_LIBRARY_ARCHIVE_URL:-}" == https://* ]] || { echo 'Private authored-slide input requires WMDS_FINISHED_LIBRARY_ARCHIVE_URL using HTTPS' >&2; exit 1; }
export WMDS_BROWSING_AS_OF="${WMDS_BROWSING_AS_OF:-${CI_PIPELINE_CREATED_AT:0:10}}"
[[ "${WMDS_BROWSING_AS_OF:-}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || { echo 'Set WMDS_BROWSING_AS_OF to the release review date (YYYY-MM-DD); CI never grants content approval' >&2; exit 1; }
[[ "$WMDS_BROWSING_AS_OF" == "${CI_PIPELINE_CREATED_AT:0:10}" ]] || { echo 'Browsing freshness date must match the authoritative GitLab pipeline creation date' >&2; exit 1; }
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
fetch_private() {
  local input_url="$1" input_sha="$2" input_name="$3"
  local headers=()
  # A protected masked URL can be presigned S3. Never send CI tokens outside
  # this exact private project's Generic Package endpoint; no redirect follows.
  if [[ "$input_url" == https://gitlab.samcott.com/api/v4/projects/17/packages/generic/* ]]; then headers=(-H "JOB-TOKEN: $CI_JOB_TOKEN"); fi
  curl --fail --proto '=https' --tlsv1.2 --retry 3 "${headers[@]}" --output "$tmp/$input_name.tar.gz" "$input_url"
  printf '%s  %s\n' "$input_sha" "$tmp/$input_name.tar.gz" | sha256sum -c -
  release_ci extract "$tmp/$input_name.tar.gz" "$tmp/$input_name"
}
fetch_private "$WMDS_BRANDING_ARCHIVE_URL" "$WMDS_BRANDING_ARCHIVE_SHA256" branding
fetch_private "$WMDS_FINISHED_LIBRARY_ARCHIVE_URL" "$WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256" finished-library
export WMDS_BRANDING_ROOT="$tmp/branding"
bundle_revision="$(cat release/default-bundle.txt)"
go build -buildvcs=false -trimpath -ldflags "-X main.version=$release_version -X main.releaseIdentity=$release_commit" -o "$tmp/pptxdesign" ./cmd/pptxdesign
"$tmp/pptxdesign" browsing-library --kind templates --frames "${WMDS_BROWSING_FRAMES:-catalog}" --media-policy originals --bundle "library/wm-design-system/$bundle_revision" --as-of "$WMDS_BROWSING_AS_OF" --out "$tmp/template-browsing"
"$tmp/pptxdesign" browsing-library --kind reusable --bundle "library/wm-design-system/$bundle_revision" --as-of "$WMDS_BROWSING_AS_OF" --finished-library "$tmp/finished-library" --out "$tmp/reusable-browsing"
if [[ "${PPTXGENGO_PACKAGE_KIND:-cli-only}" == full ]]; then
  # Stage-only verifies registered originals and documentation before packaging.
  bash scripts/install-local-release.sh --stage-only "$PWD/dist/resources" --version "$release_version"
else
  [[ "${PPTXGENGO_PACKAGE_KIND:-cli-only}" == cli-only ]] || { echo 'Unknown package kind' >&2; exit 1; }
  mkdir -p dist/resources
fi
mkdir -p dist/resources/browsing
cp "$tmp/template-browsing/template-library.pptx" dist/resources/browsing/
cp "$tmp/template-browsing/browsing-manifest.json" dist/resources/browsing/template-library.manifest.json
cp "$tmp/reusable-browsing/reusable-slides.pptx" dist/resources/browsing/
cp "$tmp/reusable-browsing/browsing-manifest.json" dist/resources/browsing/reusable-slides.manifest.json
# Hash actual deck bytes, not just input manifests. Include this inventory in
# each platform archive, the final release checksums and Sigstore attestation.
python3 - <<'PY'
import hashlib, json, os
from pathlib import Path
root=Path('dist/resources')
inputs={'branding_archive_sha256':os.environ['WMDS_BRANDING_ARCHIVE_SHA256'],'finished_library_archive_sha256':os.environ['WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256'],'as_of':os.environ['WMDS_BROWSING_AS_OF'],'pipeline_created_at':os.environ['CI_PIPELINE_CREATED_AT']}
common=None
for name in ('template-library.manifest.json','reusable-slides.manifest.json'):
    path=root/'browsing'/name
    data=json.loads(path.read_text());data['release_inputs']=inputs
    pins={key:data[key] for key in ('bundle_sha256','source_revision','source_commit','source_files','compiler','release_identity')}
    if common is not None and common != pins: raise SystemExit('Generated browsing decks disagree on exact tagged bundle/source/compiler/release pins')
    common=pins
    path.write_text(json.dumps(data,separators=(',',':'))+'\n')
files={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'browsing').iterdir()) if p.is_file()}
(root/'browsing-manifest.json').write_text(json.dumps({'schema':'pptxgengo.release-browsing-files.v1','files_sha256':files,**common,**inputs},indent=2)+'\n')
local=root/'release-manifest.json'
if local.exists():
    data=json.loads(local.read_text());data['files_sha256'].update(files);data['files_sha256']['browsing-manifest.json']=hashlib.sha256((root/'browsing-manifest.json').read_bytes()).hexdigest();data['file_count']=len(data['files_sha256']);local.write_text(json.dumps(data,indent=2)+'\n')
PY
release_ci prepare-resource-policy dist/resources "$release_version" "$release_commit"
