#!/bin/bash
# Capture an immutable prepared packet from a normal Terminal whose PowerPoint
# automation already works. Does not save, close or export any presentation.
set -euo pipefail
if [[ $# != 1 ]]; then echo 'Usage: capture-wmds-typography.sh PACKET-DIR' >&2; exit 64; fi
wmds_root="$(cd -- "$1" && pwd)"
wmds_run=""
trap 'wmds_exit_code=$?; if [[ $wmds_exit_code -ne 0 && -n "$wmds_run" ]]; then echo "Capture stopped (exit $wmds_exit_code); retained diagnostics: $wmds_run" >&2; for wmds_log in "$wmds_run"/*.log; do if [[ -s "$wmds_log" ]]; then echo "$(basename -- "$wmds_log"):" >&2; cat "$wmds_log" >&2; fi; done; fi' EXIT
python3 - "$wmds_root" <<'PY'
import hashlib,json,sys
from pathlib import Path
r=Path(sys.argv[1]);p=json.loads((r/'prepared.json').read_text())
for name,expected in p['artifacts'].items():
    if hashlib.sha256((r/name).read_bytes()).hexdigest()!=expected:raise SystemExit('Packet drift: '+name)
print('Prepared packet hashes match.')
PY
echo 'Checking live PowerPoint automation (next number is the open presentation count)...'
/usr/bin/osascript -e 'tell application "/Applications/Microsoft PowerPoint.app" to get count of presentations'
wmds_run="$(mktemp -d "$wmds_root/runs/capture.XXXXXXXX")"
export SWIFT_MODULECACHE_PATH="$wmds_run/swift-cache"
export CLANG_MODULE_CACHE_PATH="$wmds_run/clang-cache"
echo "Capture directory: $wmds_run"
# Reject a same-named deck from another directory or an edited in-memory copy.
# The presentation check is read-only and runs before and after measurement.
cat > "$wmds_run/check-open.applescript" <<'APPLESCRIPT'
on run argv
    set expectedPath to item 1 of argv
    set matchCount to 0
    set matchedName to ""
    tell application "/Applications/Microsoft PowerPoint.app"
        -- Enumerate by index: avoid Office's whose-filter behavior and do not
        -- assume whether the presentation name includes its file extension.
        repeat with presentationIndex from 1 to (count of presentations)
            set p to presentation presentationIndex
            set livePath to my normalizePath(full name of p)
            if livePath is expectedPath then
                set matchCount to matchCount + 1
                if saved of p is false then error "The probe presentation has unsaved changes; reopen the unedited packet" number 66
                set matchedName to name of p
            end if
        end repeat
    end tell
    if matchCount is not 1 then error "Expected exactly one saved open presentation at packet path " & expectedPath & "; found " & matchCount number 65
    if matchedName is "" then error "The matching presentation has an empty name" number 67
    return matchedName
end run

on normalizePath(livePath)
    try
        return POSIX path of (livePath as alias)
    on error
        return livePath as text
    end try
end normalizePath
APPLESCRIPT
echo 'Locating the saved packet presentation by full path...'
/usr/bin/osascript "$wmds_run/check-open.applescript" "$wmds_root/typography-probes.pptx" > "$wmds_run/presentation-name-before.txt" 2> "$wmds_run/presentation-before.log"
wmds_presentation_name="$(cat "$wmds_run/presentation-name-before.txt")"
echo "Found presentation: $wmds_presentation_name"
echo 'Inspecting the font environment...'
/usr/bin/swift "$wmds_root/inspect-wmds-fonts.swift" < "$wmds_root/probes.json" > "$wmds_run/environment-before.json" 2> "$wmds_run/environment-before.log"
# Require an already-open exact filename. Opening and visually reviewing the
# prepared deck is explicit in the packet README; no other open deck is touched.
wmds_probe_count="$(python3 - "$wmds_root/probes.json" <<'PYCOUNT'
import json,sys
print(len(json.load(open(sys.argv[1]))['probes']))
PYCOUNT
)"
echo "Capturing $wmds_probe_count text objects, character by character; this can take several minutes..."
if ! /usr/bin/osascript "$wmds_root/measure-wmds-text.applescript" "$wmds_presentation_name" > "$wmds_run/native.json" 2> "$wmds_run/native.log"; then
    exit 1
fi
echo 'Character capture returned; checking environment, saved presentation and packet hashes...'
/usr/bin/swift "$wmds_root/inspect-wmds-fonts.swift" < "$wmds_root/probes.json" > "$wmds_run/environment-after.json" 2> "$wmds_run/environment-after.log"
/usr/bin/osascript "$wmds_run/check-open.applescript" "$wmds_root/typography-probes.pptx" > "$wmds_run/presentation-name-after.txt" 2> "$wmds_run/presentation-after.log"
python3 - "$wmds_root" "$wmds_run" <<'PY'
import datetime,hashlib,json,sys
from pathlib import Path
r,o=map(Path,sys.argv[1:]);sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
prepared=json.loads((r/'prepared.json').read_text())
for name,expected in prepared['artifacts'].items():
    if sha(r/name)!=expected:raise SystemExit('Packet changed during capture: '+name)
a=json.loads((o/'environment-before.json').read_text());b=json.loads((o/'environment-after.json').read_text())
if a!=b:raise SystemExit('Environment changed during capture')
m=json.loads((r/'probes.json').read_text());n=json.loads((o/'native.json').read_text())
live_name=(o/'presentation-name-before.txt').read_text().strip()
if live_name!=(o/'presentation-name-after.txt').read_text().strip() or live_name!=n['presentation']:raise SystemExit('Presentation identity changed during capture')
expected={(p['slide'],p['id']):p for p in m['probes']}
actual={(p['slide_index'],p['shape_name']):p for p in n['measurements']}
if len(actual)!=len(n['measurements']) or set(actual)!=set(expected):raise SystemExit('Native shape inventory mismatch')
normalize=lambda s:(s or '').replace('\r\n','\n').replace('\r','\n')
content_mismatches=[]
for key,p in expected.items():
    observed=normalize(actual[key]['text']);authored=p['record']['layout']['displayed']
    if observed!=authored:content_mismatches.append(dict(id=p['id'],expected=authored,observed=observed))
receipt=dict(schema='pptxgengo.wmds-native-capture.v1',completed_at=datetime.datetime.now(datetime.timezone.utc).isoformat(),
    deck_sha256=m['deck_sha256'],manifest_sha256=sha(r/'probes.json'),prepared_sha256=sha(r/'prepared.json'),
    native_sha256=sha(o/'native.json'),environment_sha256=sha(o/'environment-before.json'),probe_count=len(expected),
    presentation_path=str(r/'typography-probes.pptx'),presentation_saved_before_and_after=True,
    presentation_name=live_name,
    state='captured_with_content_mismatches' if content_mismatches else 'captured',
    content_matches=len(expected)-len(content_mismatches),content_mismatches=content_mismatches,
    identity_scope='Saved exact-path presentation checked before/after; native content and environment checked. Exact PowerPoint font file access is not observed.')
with (o/'capture-receipt.json').open('x') as f:json.dump(receipt,f,indent=2);f.write('\n')
print('Capture complete: '+str(o))
if content_mismatches:
    print('Qualification failed content controls (raw evidence retained): '+', '.join(c['id'] for c in content_mismatches))
    print('This is a captured diagnostic result, not a native qualification pass.')
PY
