#!/usr/bin/env bash
set -euo pipefail
binary="${1:?}"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
codesign --verify --strict --verbose=2 "$binary"
codesign -d --verbose=4 "$binary" 2> "$tmp/info"
grep -Fq 'Authority=Developer ID Application: Ryne Scott (VZJU7JS89T)' "$tmp/info"
grep -Fq 'TeamIdentifier=VZJU7JS89T' "$tmp/info"
grep -Fq "Identifier=net.scottwebworks.pptxgengo.$(basename "$binary")" "$tmp/info"
grep -Eq 'flags=.*\(runtime\)' "$tmp/info"
grep -Fq 'Timestamp=' "$tmp/info"
cp "$binary" "$tmp/tampered"
byte="$(od -An -tu1 -j 4096 -N 1 "$tmp/tampered" | tr -d '[:space:]')"
printf -v replacement '\\%03o' "$((255 - 10#$byte))"
printf '%b' "$replacement" | dd of="$tmp/tampered" bs=1 seek=4096 conv=notrunc 2>/dev/null
if codesign --verify --strict "$tmp/tampered" >/dev/null 2>&1; then echo 'Tampered Mach-O verified' >&2; exit 1; fi
