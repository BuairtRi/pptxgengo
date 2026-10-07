#!/usr/bin/env bash
# Browsing decks are private presentation inputs in every installation archive.
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
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
"$tmp/pptxdesign" browsing-library --kind templates --frames "${WMDS_BROWSING_FRAMES:-catalog}" --bundle "library/wm-design-system/$bundle_revision" --as-of "$WMDS_BROWSING_AS_OF" --out "$tmp/template-browsing"
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
files={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((root/'browsing').iterdir()) if p.is_file()}
(root/'browsing-manifest.json').write_text(json.dumps({'schema':'pptxgengo.release-browsing-files.v1','files_sha256':files,'branding_archive_sha256':os.environ['WMDS_BRANDING_ARCHIVE_SHA256'],'finished_library_archive_sha256':os.environ['WMDS_FINISHED_LIBRARY_ARCHIVE_SHA256'],'as_of':os.environ['WMDS_BROWSING_AS_OF']},indent=2)+'\n')
local=root/'release-manifest.json'
if local.exists():
    data=json.loads(local.read_text());data['files_sha256'].update(files);data['files_sha256']['browsing-manifest.json']=hashlib.sha256((root/'browsing-manifest.json').read_bytes()).hexdigest();data['file_count']=len(data['files_sha256']);local.write_text(json.dumps(data,indent=2)+'\n')
PY
