#!/usr/bin/env bash
# Source this: the shared signing/security images retain an older Go toolchain.
set -euo pipefail
case "$(uname -m)" in
  x86_64) host_arch=amd64; go_sha=63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445 ;;
  aarch64|arm64) host_arch=arm64; go_sha=3450b45a3f9ee8568792736a5c5e70a1f2e9b36c35a8f74958c03e51d7d92bec ;;
  *) echo 'Unsupported Linux CI architecture' >&2; exit 1 ;;
esac
release_go_dir="$(mktemp -d)"
trap 'rm -rf "$release_go_dir"' EXIT
archive="go1.27.1.linux-$host_arch.tar.gz"
curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$release_go_dir/$archive" "https://go.dev/dl/$archive"
printf '%s  %s\n' "$go_sha" "$release_go_dir/$archive" | sha256sum -c -
tar -xzf "$release_go_dir/$archive" -C "$release_go_dir"
export GOROOT="$release_go_dir/go"
export PATH="$GOROOT/bin:$PATH"
[[ "$(go version)" == "go version go1.27.1 linux/$host_arch" ]]
