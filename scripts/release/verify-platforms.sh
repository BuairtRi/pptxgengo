#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
platform="${1:?usage: verify-platforms.sh windows|darwin|linux}"
release_ci verify dist/final "$release_version" "$release_commit"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
if [[ "$platform" == windows ]]; then bash scripts/release/install-windows-ca.sh --bundle-out "$tmp/ca.pem"; fi
for arch in amd64 arm64; do
  target="$platform-$arch"; extension=tar.gz; [[ "$platform" == windows ]] && extension=zip
  release_ci extract "dist/final/pptxgengo-$release_version-$target.$extension" "$tmp/$target"
  release_ci verify-binaries "$tmp/$target" "$target" "$release_version" "$release_commit"
  for tool in pptxgengo pptxdesign wmdsdocs; do
    case "$platform" in
      windows) bash scripts/release/verify-windows.sh "$tmp/$target/bin/$tool.exe" "$tmp/ca.pem" ;;
      darwin) bash scripts/release/verify-macos.sh "$tmp/$target/bin/$tool" ;;
    esac
  done
  if [[ "$platform" == "$(go env GOHOSTOS)" && "$arch" == "$(go env GOHOSTARCH)" ]]; then
    [[ "$("$tmp/$target/bin/pptxgengo" --version)" == "$release_version" ]]
    "$tmp/$target/bin/pptxdesign" asset-catalog > "$tmp/assets.json"
  fi
  if [[ "$platform" == darwin ]]; then
    keychain="$HOME/Library/Keychains/${MACOS_SIGN_KEYCHAIN:?}"
    security unlock-keychain -p "${MACOS_SIGN_KEYCHAIN_PASSWORD:?}" "$keychain"
    id="$(plutil -extract id raw -o - "dist/final/$target.notarization.json")"
    xcrun notarytool info "$id" --keychain-profile "${MACOS_NOTARY_KEYCHAIN_PROFILE:?}" --keychain "$keychain" --output-format json > "$tmp/notary.json"
    [[ "$(plutil -extract status raw -o - "$tmp/notary.json")" == Accepted ]]
  fi
done
