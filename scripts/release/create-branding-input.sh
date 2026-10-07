#!/usr/bin/env bash
# Run on the machine holding the registered branding originals. The producer
# validates bytes via the normal frozen-release installer before exporting.
set -euo pipefail
repo="$(cd "$(dirname "$0")/../.." && pwd)"
output="${1:?usage: create-branding-input.sh NEW.tar.gz}"
[[ ! -e "$output" ]] || { echo 'Output must be new' >&2; exit 1; }
output="$(cd "$(dirname "$output")" && pwd)/$(basename "$output")"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
bash "$repo/scripts/install-local-release.sh" --stage-only "$tmp/release"
# Match the safe regular-file archive contract used by the consumer. Python
# is already required by the source packaging installer.
python3 - "$tmp/release/branding" "$output" <<'PY'
import gzip,tarfile,sys
from pathlib import Path
root=Path(sys.argv[1])
with open(sys.argv[2],'xb') as dst:
 with gzip.GzipFile(fileobj=dst,mode='wb',mtime=0,filename='') as compressed:
  with tarfile.open(fileobj=compressed,mode='w') as archive:
   for p in sorted(root.rglob('*')):
    if p.is_dir():continue
    if not p.is_file() or p.is_symlink():raise SystemExit('branding input must contain regular files')
    info=archive.gettarinfo(str(p),arcname=p.relative_to(root).as_posix())
    info.mtime=0;info.uid=info.gid=0;info.uname=info.gname='';info.mode=0o644
    with p.open('rb') as f:archive.addfile(info,f)
PY
shasum -a 256 "$output"
echo 'Upload privately to S3 or GitLab Generic Packages; set the protected URL and SHA-256 variables.'
