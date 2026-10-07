#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
bash scripts/release/install-tools.sh --bin-dir "$tmp/tools" --only cosign
export PATH="$tmp/tools:$PATH"
bash scripts/release/sigstore.sh initialize --home "$tmp/trust"
bash scripts/release/sigstore.sh verify --home "$tmp/trust" --blob dist/final/manifest.json --bundle dist/final/manifest.sigstore.json --version "$release_version"
release_ci publish dist/final "$release_version" "$release_commit"
