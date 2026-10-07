#!/usr/bin/env bash
# Sign and verify release evidence with the organization-owned private
# Sigstore deployment. The timestamp authority is deliberately in-cluster;
# the protected signing job must run on the in-cluster Linux runner.

set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
policy_file=${PPTXGENGO_SIGSTORE_POLICY_FILE:-$script_dir/../../release/sigstore-policy.json}
command_name=${1:-}
shift || true
home_dir=''
blob=''
bundle=''
version=''

die() {
  printf 'sigstore-release: %s\n' "$1" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || die "'$1' is required on PATH"
}

usage() {
  cat >&2 <<'EOF'
usage:
  sigstore-release.sh initialize --home <directory>
  sigstore-release.sh sign --home <directory> --blob <file> --bundle <file> --version <vX.Y.Z>
  sigstore-release.sh verify --home <directory> --blob <file> --bundle <file> --version <vX.Y.Z>

sign reads the GitLab ID token from SIGSTORE_ID_TOKEN and never accepts it as
a command-line argument.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --home) home_dir=${2:-}; shift 2 ;;
    --blob) blob=${2:-}; shift 2 ;;
    --bundle) bundle=${2:-}; shift 2 ;;
    --version) version=${2:-}; shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *) die "unknown argument: $1" ;;
  esac
done

need base64
need cosign
need jq
[[ -s "$policy_file" ]] || die 'Sigstore policy is missing'

mirror=$(jq -er '.tuf.mirror' "$policy_file")
root_url=$(jq -er '.tuf.root_url' "$policy_file")
root_sha256=$(jq -er '.tuf.root_sha256' "$policy_file")
issuer=$(jq -er '.identity.oidc_issuer' "$policy_file")
audience=$(jq -er '.identity.audience' "$policy_file")
project_id=$(jq -er '.identity.project_id' "$policy_file")
project_path=$(jq -er '.identity.project_path' "$policy_file")
fulcio=$(jq -er '.signing.fulcio' "$policy_file")
rekor=$(jq -er '.signing.rekor' "$policy_file")
tsa=$(jq -er '.signing.tsa' "$policy_file")
cosign_version=$(jq -er '.cosign.version' "$policy_file")

[[ "$mirror" == 'https://tuf.samcott.com' && "$root_url" == "$mirror/1.root.json" ]] || die 'unexpected TUF endpoint policy'
[[ "$root_sha256" =~ ^[0-9a-f]{64}$ ]] || die 'malformed TUF root SHA-256'
[[ "$issuer" == 'https://gitlab.samcott.com' && "$audience" == 'sigstore' ]] || die 'unexpected OIDC policy'
[[ "$project_id" == '17' && "$project_path" == 'riscott/pptxgengo' ]] || die 'unexpected GitLab project policy'
[[ "$fulcio" == 'https://fulcio.samcott.com' && "$rekor" == 'https://rekor.samcott.com' ]] || die 'unexpected Fulcio/Rekor policy'
[[ "$tsa" == 'http://tsa-server.sigstore-system.svc.cluster.local/api/v1/timestamp' ]] || die 'unexpected TSA policy'
[[ "$(cosign version --json | jq -er '.gitVersion')" == "$cosign_version" ]] || die 'cosign version does not match reviewed policy'

expected_identity() {
  [[ "$version" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$ ]] || die "unsupported release version: $version"
  local template
  template=$(jq -er '.identity.certificate_identity_template' "$policy_file")
  printf '%s\n' "${template//\{\{version\}\}/$version}"
}

trusted_root_path() {
  local found
  found=$(find "$home_dir/.sigstore/root" -type f -name trusted_root.json -print -quit 2>/dev/null || true)
  [[ -n "$found" ]] || die 'TUF did not deliver trusted_root.json'
  printf '%s\n' "$found"
}

signing_config_path() {
  local found
  found=$(find "$home_dir/.sigstore/root" -type f -name signing_config.v0.2.json -print -quit 2>/dev/null || true)
  [[ -n "$found" ]] || die 'TUF did not deliver signing_config.v0.2.json'
  printf '%s\n' "$found"
}

validate_trust_targets() {
  local trusted_root signing_config
  trusted_root=$(trusted_root_path)
  signing_config=$(signing_config_path)
  jq -e '.mediaType == "application/vnd.dev.sigstore.trustedroot+json;version=0.1" and (.certificateAuthorities | length) > 0 and (.ctlogs | length) > 0 and (.tlogs | length) > 0 and (.timestampAuthorities | length) > 0' "$trusted_root" >/dev/null || die 'unexpected trusted-root target'
  jq -e '.mediaType == "application/vnd.dev.sigstore.signingconfig.v0.2+json" and (.caUrls | any(.[]; .url == "https://fulcio.samcott.com")) and (.rekorTlogUrls | any(.[]; .url == "https://rekor.samcott.com")) and (.tsaUrls | any(.[]; .url == "http://tsa-server.sigstore-system.svc.cluster.local/api/v1/timestamp"))' "$signing_config" >/dev/null || die 'TUF signing configuration does not match policy'
}

initialize() {
  [[ -n "$home_dir" ]] || die '--home is required'
  mkdir -p "$home_dir"
  chmod 700 "$home_dir"
  HOME="$home_dir" cosign initialize --mirror "$mirror" --root "$root_url" --root-checksum "$root_sha256"
  validate_trust_targets
  printf 'sigstore-release: initialized TUF trust for private release services\n'
}

decode_jwt_payload() {
  local identity_jwt=$1 payload padding
  payload=${identity_jwt#*.}
  [[ "$payload" != "$identity_jwt" && "$payload" == *.* ]] || return 1
  payload=${payload%%.*}
  padding=$(( (4 - ${#payload} % 4) % 4 ))
  if (( padding > 0 )); then
    payload="${payload}$(printf '=%.0s' $(seq 1 "$padding"))"
  fi
  printf '%s' "$payload" | tr '_-' '/+' | base64 -d 2>/dev/null
}

validate_identity_token() {
  local claims expected_sub config_ref
  [[ -n "${SIGSTORE_ID_TOKEN-}" ]] || die 'SIGSTORE_ID_TOKEN is required'
  claims=$(decode_jwt_payload "$SIGSTORE_ID_TOKEN") || die 'cannot decode SIGSTORE_ID_TOKEN claims'
  config_ref=$(expected_identity)
  config_ref=${config_ref#https://}
  expected_sub="project_path:${project_path}:ref_type:tag:ref:${version}"
  jq -e --arg issuer "$issuer" --arg audience "$audience" --arg project_id "$project_id" --arg project_path "$project_path" --arg version "$version" --arg config_ref "$config_ref" --arg expected_sub "$expected_sub" --arg commit "${CI_COMMIT_SHA:-}" '.iss == $issuer and (.aud == $audience or ((.aud | type) == "array" and (.aud | index($audience)) != null)) and (.project_id | tostring) == $project_id and .project_path == $project_path and .ref_type == "tag" and .ref == $version and (.ref_protected | tostring) == "true" and .ci_config_ref_uri == $config_ref and .sub == $expected_sub and ($commit == "" or .sha == $commit)' <<<"$claims" >/dev/null || die 'GitLab ID-token claims do not match policy'
}

validate_bundle() {
  [[ -s "$bundle" ]] || die 'Sigstore bundle was not written'
  jq -e '.mediaType == "application/vnd.dev.sigstore.bundle.v0.3+json" and (.verificationMaterial.tlogEntries | length) > 0 and (.verificationMaterial.timestampVerificationData.rfc3161Timestamps | length) > 0 and (.verificationMaterial.certificate.rawBytes | length) > 0 and (.messageSignature.signature | length) > 0' "$bundle" >/dev/null || die 'bundle lacks certificate, Rekor, signature, or RFC3161 timestamp evidence'
}

verify_bundle() {
  [[ -n "$home_dir" && -f "$blob" && -n "$bundle" && -n "$version" ]] || die '--home, --blob, --bundle, and --version are required'
  validate_trust_targets
  validate_bundle
  local trusted_root identity
  trusted_root=$(trusted_root_path)
  identity=$(expected_identity)
  env HOME="$home_dir" HTTP_PROXY='http://127.0.0.1:9' HTTPS_PROXY='http://127.0.0.1:9' ALL_PROXY='http://127.0.0.1:9' http_proxy='http://127.0.0.1:9' https_proxy='http://127.0.0.1:9' all_proxy='http://127.0.0.1:9' no_proxy='' cosign verify-blob --bundle "$bundle" --trusted-root "$trusted_root" --certificate-identity "$identity" --certificate-oidc-issuer "$issuer" --use-signed-timestamps "$blob"
  printf 'sigstore-release: verified offline bundle for %s\n' "$identity"
}

sign_blob() {
  [[ -n "$home_dir" && -f "$blob" && -n "$bundle" && -n "$version" ]] || die '--home, --blob, --bundle, and --version are required'
  [[ "${CI_COMMIT_REF_PROTECTED:-}" == 'true' && "${CI_COMMIT_TAG:-}" == "$version" ]] || die 'signing requires the matching protected GitLab tag'
  [[ "${CI_PROJECT_ID:-}" == "$project_id" && "${CI_PROJECT_PATH:-}" == "$project_path" ]] || die 'signing job project does not match policy'
  [[ "${CI_SERVER_URL:-}" == "$issuer" ]] || die 'signing job issuer does not match policy'
  validate_trust_targets
  validate_identity_token
  local trusted_root signing_config token_file
  trusted_root=$(trusted_root_path)
  signing_config=$(signing_config_path)
  token_file="$home_dir/.sigstore-id-token"
  umask 077
  printf '%s' "$SIGSTORE_ID_TOKEN" > "$token_file"
  trap 'rm -f -- "$token_file"' EXIT
  HOME="$home_dir" cosign sign-blob --identity-token "$token_file" --bundle "$bundle" --signing-config "$signing_config" --trusted-root "$trusted_root" --yes "$blob" >/dev/null
  rm -f -- "$token_file"
  trap - EXIT
  verify_bundle
}

case "$command_name" in
  initialize) initialize ;;
  sign) sign_blob ;;
  verify) verify_bundle ;;
  *) usage; exit 2 ;;
esac
