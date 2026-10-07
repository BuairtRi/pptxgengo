#!/usr/bin/env bash
# Install only reviewed, checksum-pinned release-security tools.  Release jobs
# invoke this even when the base image already has a tool, so an image update
# cannot silently substitute a scanner or signer binary.
set -euo pipefail

readonly script_dir="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly repo_root="$(CDPATH= cd -- "$script_dir/../.." && pwd)"
pins_file="$repo_root/release/tool-pins.json"
bin_dir="$repo_root/.release-tools/bin"
only=""

usage() {
  echo "usage: $0 [--pins FILE] [--bin-dir DIR] [--only name[,name...]]" >&2
  exit 2
}

while (($#)); do
  case "$1" in
    --pins) pins_file="${2:-}"; shift 2 ;;
    --bin-dir) bin_dir="${2:-}"; shift 2 ;;
    --only) only="${2:-}"; shift 2 ;;
    -h|--help) usage ;;
    *) usage ;;
  esac
done

for command in curl jq sha256sum tar; do
  command -v "$command" >/dev/null 2>&1 || { echo "error: required command $command is not on PATH" >&2; exit 1; }
done
[[ "$(uname -s)" == "Linux" && "$(uname -m)" == "x86_64" ]] || {
  echo "error: pptxgengo Linux release-security tool installer supports only Linux x86_64" >&2
  exit 1
}
[[ -f "$pins_file" ]] || { echo "error: missing release tool pins: $pins_file" >&2; exit 1; }
[[ "$(jq -r '.schema' "$pins_file")" == "pptxgengo.release-tool-pins/v1" ]] || {
  echo "error: unsupported pptxgengo release tool pin schema" >&2
  exit 1
}

names=(cosign grype jsign syft)
if [[ -n "$only" ]]; then
  IFS=',' read -r -a names <<<"$only"
fi
[[ "${#names[@]}" -gt 0 ]] || { echo "error: no tools selected" >&2; exit 1; }

tmp="$(mktemp -d)"
cleanup() { rm -rf -- "$tmp"; }
trap cleanup EXIT
mkdir -p "$bin_dir"

pin() {
  local name="$1" field="$2"
  jq -er --arg name "$name" --arg field "$field" '
    .tools[] | select(.name == $name) | .[$field]
  ' "$pins_file"
}

for name in "${names[@]}"; do
  case "$name" in cosign|grype|jsign|syft) ;; *) echo "error: unsupported tool $name" >&2; exit 1 ;; esac
  version="$(pin "$name" version)"
  filename="$(pin "$name" filename)"
  digest="$(pin "$name" sha256)"
  [[ "$digest" =~ ^[0-9a-f]{64}$ ]] || { echo "error: malformed $name SHA-256 pin" >&2; exit 1; }
  archive="$tmp/$filename"

  case "$name" in
    cosign)
      [[ "$filename" == "cosign-linux-amd64" ]] || { echo "error: malformed cosign asset" >&2; exit 1; }
      url="https://github.com/sigstore/cosign/releases/download/v${version}/${filename}"
      ;;
    jsign)
      [[ "$filename" == "jsign-${version}.jar" ]] || { echo "error: malformed jsign asset" >&2; exit 1; }
      url="https://github.com/ebourg/jsign/releases/download/${version}/${filename}"
      ;;
    syft|grype)
      [[ "$filename" == "${name}_${version}_linux_amd64.tar.gz" ]] || { echo "error: malformed $name asset" >&2; exit 1; }
      url="https://github.com/anchore/${name}/releases/download/v${version}/${filename}"
      ;;
  esac
  curl --fail --location --silent --show-error --retry 3 --retry-all-errors --output "$archive" "$url"
  printf '%s  %s\n' "$digest" "$archive" | sha256sum --check --status || {
    echo "error: $name checksum does not match the committed pin" >&2
    exit 1
  }

  case "$name" in
    cosign)
      install -m 0755 "$archive" "$bin_dir/cosign"
      actual="$($bin_dir/cosign version --json | jq -er '.gitVersion | ltrimstr("v")')"
      ;;
    jsign)
      install -m 0644 "$archive" "$bin_dir/jsign.jar"
      actual="$(java -jar "$bin_dir/jsign.jar" --version | tr -d '\r' | awk '{print $NF}')"
      ;;
    syft|grype)
      tar -xzf "$archive" -C "$tmp" "$name"
      install -m 0755 "$tmp/$name" "$bin_dir/$name"
      actual="$($bin_dir/$name version -o json | jq -er '.version')"
      ;;
  esac
  [[ "$actual" == "$version" ]] || {
    echo "error: installed $name version $actual does not match committed $version" >&2
    exit 1
  }
done

echo "installed checksum-pinned pptxgengo release tools in $bin_dir"
