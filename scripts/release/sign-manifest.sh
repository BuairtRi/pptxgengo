#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
release_ci verify dist/final "$release_version" "$release_commit"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
bash scripts/release/install-tools.sh --bin-dir "$tmp/tools" --only cosign
export PATH="$tmp/tools:$PATH"
bash scripts/release/sigstore.sh initialize --home "$tmp/trust"
bash scripts/release/sigstore.sh sign --home "$tmp/trust" --blob dist/final/manifest.json --bundle dist/final/manifest.sigstore.json --version "$release_version"
