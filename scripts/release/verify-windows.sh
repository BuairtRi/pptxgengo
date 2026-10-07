#!/usr/bin/env bash
set -euo pipefail
binary="${1:?}"; ca="${2:?}"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
osslsigncode verify -CAfile "$ca" -TSA-CAfile "$ca" -in "$binary" > "$tmp/verify.txt" 2>&1
grep -Fq 'Subject: /C=US/ST=ca/L=San Pablo/O=Ryne Scott/CN=Ryne Scott' "$tmp/verify.txt"
grep -Fq 'Timestamp Server Signature verification: ok' "$tmp/verify.txt"
grep -Fq 'Signature verification: ok' "$tmp/verify.txt"
cp "$binary" "$tmp/tampered.exe"
# Modify a byte inside the PE image; a broken signature must be rejected.
byte="$(od -An -tu1 -j 4096 -N 1 "$tmp/tampered.exe" | tr -d '[:space:]')"
printf -v replacement '\\%03o' "$((255 - 10#$byte))"
printf '%b' "$replacement" | dd of="$tmp/tampered.exe" bs=1 seek=4096 conv=notrunc status=none
if osslsigncode verify -CAfile "$ca" -TSA-CAfile "$ca" -in "$tmp/tampered.exe" >/dev/null 2>&1; then
  echo 'Tampered Authenticode payload verified' >&2; exit 1
fi
echo "Verified publisher, signature, timestamp and tamper rejection: $(basename "$binary")"
