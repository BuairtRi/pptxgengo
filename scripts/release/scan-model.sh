#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
[[ $# == 3 ]] || { echo 'scan-model.sh DIRECTORY VERSION COMMIT' >&2; exit 1; }
dir="$1"; version="$2"; commit="$3"
release_ci verify-model "$dir" "$version" "$commit"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
bash scripts/release/install-tools.sh --bin-dir "$tmp/tools" --only syft,grype
scan_env=(env -i PATH="$tmp/tools:$PATH" HOME="$tmp/home" GRYPE_CHECK_FOR_APP_UPDATE=false GRYPE_DB_CACHE_DIR="$tmp/db")
for key in HTTPS_PROXY HTTP_PROXY NO_PROXY SSL_CERT_FILE SSL_CERT_DIR; do
  if [[ -n "${!key:-}" ]]; then scan_env+=("$key=${!key}"); fi
done
"${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml db update
"${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml db status -o json > "$tmp/db.json"
jq -e '.valid == true and .built != null' "$tmp/db.json" >/dev/null
archive="$dir/pptxgengo-$version-offline-model.zip"
release_ci extract "$archive" "$tmp/model"
env -i PATH="$PATH" HOME="$tmp/home" SYFT_CHECK_FOR_APP_UPDATE=false "$tmp/tools/syft" scan "dir:$tmp/model" --source-name "$(basename "$archive")" --source-version "$version" -o "cyclonedx-json=$tmp/model.sbom.json"
sha="$(sha256sum "$archive" | cut -d' ' -f1)"
evaluated="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
# Weights/tokenizer are data files. Record their pins and attribution even when
# Syft finds no executable dependencies; a scan is not model safety qualification.
jq --arg sha "$sha" --arg version "$version" --arg evaluated "$evaluated" --slurpfile evidence "$archive.evidence.json" '
  .metadata.timestamp=$evaluated | .metadata.component.name="pptxgengo-offline-model" |
  .metadata.component.version=$version | .metadata.component.hashes=[{alg:"SHA-256",content:$sha}] |
  .components = ((.components // []) + [$evidence[0].identity.artifacts[] | {type:"file",name:.file,version:$evidence[0].identity.revision,hashes:[{alg:"SHA-256",content:.sha256}],licenses:[{license:{id:"Apache-2.0"}}],externalReferences:[{type:"distribution",url:$evidence[0].source}]}])
' "$tmp/model.sbom.json" > "$archive.sbom.json"
jq -e '.bomFormat == "CycloneDX" and (.components | length) >= 3' "$archive.sbom.json" >/dev/null
"${scan_env[@]}" "$tmp/tools/grype" --config release/grype.yaml "sbom:$archive.sbom.json" -o json --file "$archive.vulnerabilities.json"
jq -e '(.descriptor.db.status.valid == true) and ([.matches[] | select(.vulnerability.severity == "Critical")] | length == 0)' "$archive.vulnerabilities.json" >/dev/null
release_ci verify-model "$dir" "$version" "$commit"
