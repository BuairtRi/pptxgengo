#!/usr/bin/env bash
# Source this so its verified toolchain is used throughout the signing job.
set -euo pipefail
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/pptxgengo-go.XXXXXX")"
trap 'rm -rf "$temporary_dir"' EXIT
archive=go1.27.1.darwin-arm64.tar.gz
curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$temporary_dir/$archive" "https://go.dev/dl/$archive"
printf '%s  %s\n' ee215d57e0ec269c60cc9ceca68e6bda321ba9ee5afe24f4b0988703c2d87d12 "$temporary_dir/$archive" | shasum -a 256 -c -
tar -xzf "$temporary_dir/$archive" -C "$temporary_dir"
export GOROOT="$temporary_dir/go"
export PATH="$GOROOT/bin:$PATH"
[[ "$(go version)" == 'go version go1.27.1 darwin/arm64' ]]
