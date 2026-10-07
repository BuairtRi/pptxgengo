#!/usr/bin/env bash
set -euo pipefail
out="$(mktemp -d "${TMPDIR:-/tmp}/pptxgengo-cross-build.XXXXXX")"
trap 'rm -rf "$out"' EXIT
export CGO_ENABLED=0
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  export GOOS="${target%/*}" GOARCH="${target#*/}"
  mkdir -p "$out/$GOOS-$GOARCH"
  go build -trimpath -o "$out/$GOOS-$GOARCH/" ./cmd/pptxgengo ./cmd/pptxdesign ./cmd/wmdsdocs
  for package in internal/installstate internal/deckproject scripts/cmd/search-benchmark; do
    go test -c -o "$out/$GOOS-$GOARCH/${package##*/}.test" "./$package"
  done
  printf 'Cross-build verified %s; native execution is separate\n' "$target"
done
