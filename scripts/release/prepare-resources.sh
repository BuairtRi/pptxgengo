#!/usr/bin/env bash
# Activated only once the other machine supplies private branding originals.
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
[[ "${PPTXGENGO_PACKAGE_KIND:-cli-only}" == full ]] || { echo 'CLI-only release: branding resources not requested'; exit 0; }
[[ "${WMDS_BRANDING_ARCHIVE_SHA256:-}" =~ ^[a-f0-9]{64}$ ]] || { echo 'Missing immutable branding archive SHA-256' >&2; exit 1; }
[[ "${WMDS_BRANDING_ARCHIVE_URL:-}" == https://* ]] || { echo 'Private branding input must use HTTPS' >&2; exit 1; }
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
# A protected masked URL can be a presigned private S3 URL. Only exact same
# project Generic Package URLs receive the CI token; never forward it to S3.
headers=()
if [[ "$WMDS_BRANDING_ARCHIVE_URL" == https://gitlab.samcott.com/api/v4/projects/17/packages/generic/* ]]; then headers=(-H "JOB-TOKEN: $CI_JOB_TOKEN"); fi
curl --fail --proto '=https' --tlsv1.2 --retry 3 "${headers[@]}" --output "$tmp/branding.tar.gz" "$WMDS_BRANDING_ARCHIVE_URL"
printf '%s  %s\n' "$WMDS_BRANDING_ARCHIVE_SHA256" "$tmp/branding.tar.gz" | sha256sum -c -
release_ci extract "$tmp/branding.tar.gz" "$tmp/branding"
export WMDS_BRANDING_ROOT="$tmp/branding"
# Stage-only verifies every registered original, gallery link, documentation
# source and package hash. It never changes this runner's CLI or skill links.
bash scripts/install-local-release.sh --stage-only "$PWD/dist/resources" --version "$release_version"
