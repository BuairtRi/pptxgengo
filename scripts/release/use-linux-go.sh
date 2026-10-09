#!/usr/bin/env bash
# Source this: the shared signing/security images retain an older Go toolchain.
set -euo pipefail
case "$(uname -m)" in
  x86_64) host_arch=amd64; go_sha=ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 ;;
  aarch64|arm64) host_arch=arm64; go_sha=94f3e30b8e374bc285e7dadc11e0865726b9bc6e85b841ccceaabc0214c6b7c8 ;;
  *) echo 'Unsupported Linux CI architecture' >&2; exit 1 ;;
esac
release_go_dir="$(mktemp -d)"
trap 'rm -rf "$release_go_dir"' EXIT
archive="go1.27.2.linux-$host_arch.tar.gz"
curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --output "$release_go_dir/$archive" "https://go.dev/dl/$archive"
printf '%s  %s\n' "$go_sha" "$release_go_dir/$archive" | sha256sum -c -
tar -xzf "$release_go_dir/$archive" -C "$release_go_dir"
export GOROOT="$release_go_dir/go"
export PATH="$GOROOT/bin:$PATH"
[[ "$(go version)" == "go version go1.27.2 linux/$host_arch" ]]
