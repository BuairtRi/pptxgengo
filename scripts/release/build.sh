#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
if [[ "${CI:-}" == true ]]; then require_release_identity; fi
target="${1:?usage: build.sh OS-ARCH}"
case "$target" in linux-amd64|linux-arm64|windows-amd64|windows-arm64|darwin-amd64|darwin-arm64) ;; *) exit 2 ;; esac
os="${target%-*}"; arch="${target#*-}"
[[ "$(go env GOVERSION)" == go1.27.2 ]] || { echo 'Release requires Go 1.27.2' >&2; exit 1; }
tmp="$(mktemp -d)"; trap 'chmod -R u+w "$tmp"; rm -rf "$tmp"' EXIT
out="$release_root/dist/unsigned/$target"
[[ ! -e "$out" ]] || { echo 'Unsigned output must be new' >&2; exit 1; }
for pass in first second; do
  mkdir -p "$tmp/$pass/bin"
  # Separate build caches make the comparison independent of compiled objects.
  env GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 GOTOOLCHAIN=local GOCACHE="$tmp/cache-$pass" \
    go build -mod=readonly -buildvcs=false -trimpath -ldflags "-s -w -buildid= -X main.version=$release_version -X main.releaseIdentity=pptxgengo-build:$release_version:$release_commit:$target" \
    -o "$tmp/$pass/bin/" ./cmd/pptxgengo ./cmd/pptxdesign ./cmd/wmdsdocs
done
diff -r "$tmp/first/bin" "$tmp/second/bin"
mkdir -p "$out"
cp -R "$tmp/first/bin" "$out/bin"
release_ci evidence "$out" "$target" "$release_version" "$release_commit" reproducible
echo "Two independent builds matched: $target"
