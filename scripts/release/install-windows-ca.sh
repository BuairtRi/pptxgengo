#!/usr/bin/env bash
# Produce a private CA bundle for Linux-side Authenticode verification. The
# committed Microsoft root is fingerprint-pinned before it is appended; this
# avoids trusting an ambient or silently changed certificate store entry.
set -euo pipefail

readonly script_dir="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly root_cert="$script_dir/../../trust/microsoft-identity-verification-root-ca-2020.pem"
readonly expected_sha256="5367F20C7ADE0E2BCA790915056D086B720C33C1FA2A2661ACF787E3292E1270"

if [[ "${1:-}" != "--bundle-out" || -z "${2:-}" || "$#" -ne 2 ]]; then
  echo "usage: $0 --bundle-out PATH" >&2
  exit 2
fi
bundle_out="$2"

for command in openssl sha256sum; do
  command -v "$command" >/dev/null 2>&1 || { echo "error: $command is required" >&2; exit 1; }
done
[[ "$(uname -s)" == "Linux" ]] || { echo "error: Linux is required" >&2; exit 1; }
[[ -f "$root_cert" ]] || { echo "error: missing committed Microsoft trust root" >&2; exit 1; }

actual_sha256="$(openssl x509 -in "$root_cert" -noout -fingerprint -sha256 | cut -d= -f2 | tr -d ':[:space:]' | tr '[:lower:]' '[:upper:]')"
[[ "$actual_sha256" == "$expected_sha256" ]] || { echo "error: Microsoft trust root fingerprint mismatch" >&2; exit 1; }
[[ "$(openssl x509 -in "$root_cert" -noout -subject)" == *"Microsoft Identity Verification Root Certificate Authority 2020"* ]] || {
  echo "error: unexpected Microsoft trust root subject" >&2
  exit 1
}

system_bundle="${SSL_CERT_FILE:-/etc/ssl/certs/ca-certificates.crt}"
[[ -f "$system_bundle" ]] || { echo "error: system CA bundle is missing" >&2; exit 1; }
mkdir -p "$(dirname "$bundle_out")"
cp "$system_bundle" "$bundle_out"
printf '\n' >> "$bundle_out"
cat "$root_cert" >> "$bundle_out"
chmod 0644 "$bundle_out"
echo "wrote Microsoft Trusted Signing verification CA bundle"
