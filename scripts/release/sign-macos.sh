#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
[[ "$(uname -s)" == Darwin ]] || exit 1
identity='Developer ID Application: Ryne Scott (VZJU7JS89T)'
[[ "${MACOS_CI_ENABLED:-}" == 1 && "${MACOS_SIGN_ENABLED:-}" == 1 && "${MACOS_SIGN_IDENTITY:-}" == "$identity" ]] || exit 1
keychain="$HOME/Library/Keychains/${MACOS_SIGN_KEYCHAIN:?}"
security find-identity -v -p codesigning "$keychain" | grep -Fq "$identity"
security unlock-keychain -p "${MACOS_SIGN_KEYCHAIN_PASSWORD:?}" "$keychain"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/notarize"
for target in darwin-amd64 darwin-arm64; do
  input="dist/unsigned/$target"; output="dist/signed/$target"
  release_ci verify-binaries "$input" "$target" "$release_version" "$release_commit"
  [[ ! -e "$output" ]] || exit 1
  mkdir -p "$output/bin" "$tmp/notarize/$target"
  for tool in pptxgengo pptxdesign wmdsdocs; do
    binary="$output/bin/$tool"
    cp "$input/bin/$tool" "$binary"
    codesign --force --timestamp --options runtime --identifier "net.scottwebworks.pptxgengo.$tool" --sign "$identity" --keychain "$keychain" "$binary"
    bash scripts/release/verify-macos.sh "$binary"
    cp "$binary" "$tmp/notarize/$target/$tool"
  done
done
/usr/bin/ditto -c -k --sequesterRsrc --keepParent "$tmp/notarize" "$tmp/notarize.zip"
xcrun notarytool submit "$tmp/notarize.zip" --keychain-profile "${MACOS_NOTARY_KEYCHAIN_PROFILE:?}" --keychain "$keychain" --wait --timeout 30m --output-format json > "$tmp/notary.json"
[[ "$(plutil -extract status raw -o - "$tmp/notary.json")" == Accepted ]] || { cat "$tmp/notary.json"; exit 1; }
notary_id="$(plutil -extract id raw -o - "$tmp/notary.json")"
xcrun notarytool info "$notary_id" --keychain-profile "$MACOS_NOTARY_KEYCHAIN_PROFILE" --keychain "$keychain" --output-format json > "$tmp/notary-live.json"
[[ "$(plutil -extract status raw -o - "$tmp/notary-live.json")" == Accepted ]] || exit 1
for target in darwin-amd64 darwin-arm64; do
  cp "$tmp/notary.json" "dist/signed/$target/notarization.json"
  release_ci evidence "dist/signed/$target" "$target" "$release_version" "$release_commit" notarized "dist/unsigned/$target/build-evidence.json"
done
# A ZIP or bare CLI cannot be stapled. Apple retains tickets for every signed
# executable; online Gatekeeper lookup is required on first use.
