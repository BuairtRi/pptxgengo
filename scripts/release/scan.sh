#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
bash scripts/release/install-tools.sh --bin-dir "$tmp/tools" --only syft,grype
scan_env=(env -i PATH="$tmp/tools:$PATH" HOME="$tmp/home" GRYPE_CHECK_FOR_APP_UPDATE=false GRYPE_DB_CACHE_DIR="$tmp/db")
for name in HTTPS_PROXY HTTP_PROXY NO_PROXY SSL_CERT_FILE SSL_CERT_DIR; do
  if [[ -n "${!name:-}" ]]; then scan_env+=("$name=${!name}"); fi
done
"${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml db update
"${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml db status -o json > "$tmp/db.json"
jq -e '.valid == true and .built != null' "$tmp/db.json" >/dev/null
passed=true
evaluated="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
for target in darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64; do
  extension=tar.gz; [[ "$target" == windows-* ]] && extension=zip
  archive="dist/final/pptxgengo-$release_version-$target.$extension"
  release_ci extract "$archive" "$tmp/$target"
  release_ci verify-binaries "$tmp/$target" "$target" "$release_version" "$release_commit"
  raw="$tmp/$target.sbom.json"
  env -i PATH="$PATH" HOME="$tmp/home" SYFT_CHECK_FOR_APP_UPDATE=false "$tmp/tools/syft" scan "dir:$tmp/$target" --source-name "$(basename "$archive")" --source-version "$release_version" -o "cyclonedx-json=$raw"
  sha="$(sha256sum "$archive" | cut -d' ' -f1)"
  jq --arg sha "$sha" --arg version "$release_version" --arg evaluated "$evaluated" '
    .metadata.timestamp=$evaluated | .metadata.component.name="pptxgengo" |
    .metadata.component.version=$version | .metadata.component.hashes=[{alg:"SHA-256",content:$sha}]
  ' "$raw" > "$archive.sbom.json"
  jq -e '.bomFormat == "CycloneDX" and (.components | length) > 0' "$archive.sbom.json" >/dev/null
  "${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml "sbom:$archive.sbom.json" -o json --file "$archive.vulnerabilities.json"
  # Suppressions, missing/stale DBs and scanner errors cannot become a pass.
  if ! jq -e '(.descriptor.db.status.valid == true) and ([.matches[] | select(.vulnerability.severity == "Critical")] | length == 0)' "$archive.vulnerabilities.json" >/dev/null; then passed=false; fi
done
if [[ "$PPTXGENGO_OFFLINE_MODEL" == true ]]; then
  bash scripts/release/scan-model.sh dist/final "$release_version" "$release_commit"
fi

jq -n --argjson passed "$passed" --arg evaluated "$evaluated" --arg version "$release_version" --arg commit "$release_commit" --slurpfile db "$tmp/db.json" \
  '{schema:"pptxgengo.security-policy/v1",passed:$passed,version:$version,commit:$commit,evaluated_at:$evaluated,blocking_severity:"Critical",suppressions:[],maximum_db_age_hours:120,database:$db[0]}' > dist/final/security-policy.json
[[ "$passed" == true ]] || { echo 'Release security policy failed; retained reports explain findings' >&2; exit 1; }
release_ci seal dist/final "$release_version" "$release_commit"
