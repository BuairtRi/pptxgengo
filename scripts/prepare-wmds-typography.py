#!/usr/bin/env python3
"""Pin the capture tools beside a newly generated WMDS typography fixture."""
import argparse
import hashlib
import json
import math
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--packet', required=True, type=Path)
a = p.parse_args()
root = a.packet.resolve()
scripts = Path(__file__).resolve().parent
manifest = json.loads((root/'probes.json').read_text())
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
if manifest['schema'] != 'pptxgengo.wmds-typography-probes.v1' or sha(root/'typography-probes.pptx') != manifest['deck_sha256']:
    raise SystemExit('Invalid or changed typography fixture')
if (root/'prepared.json').exists():
    raise SystemExit('Packet already prepared; prepare a new directory')
names = ['capture-wmds-typography.sh', 'inspect-wmds-fonts.swift', 'measure-wmds-text.applescript',
         'compare-wmds-typography.py', 'read-font-reference-pdf.swift', 'read-font-reference-baselines.swift']
for name in names:
    target = 'capture.sh' if name == 'capture-wmds-typography.sh' else name
    data = (scripts/name).read_bytes()
    if name == 'measure-wmds-text.applescript':
        data = data.replace(b'tell application "Microsoft PowerPoint"', b'tell application "/Applications/Microsoft PowerPoint.app"')
    with (root/target).open('xb') as output:
        output.write(data)
(root/'runs').mkdir(exist_ok=False)
artifacts = {name: sha(root/name) for name in ['typography-probes.pptx', 'probes.json', 'capture.sh']+names[1:]}
if manifest['engine'] == 'wmds-go-foundation.v2':
    calibration = scripts.parent/'library/wm-design-system/typography-v2-candidate/calibration.json'
    with (root/'calibration.json').open('xb') as output:
        output.write(calibration.read_bytes())
    artifacts['calibration.json'] = sha(root/'calibration.json')
    source_files = sorted((scripts.parent/'internal/wmdesign').glob('*.go')) + [scripts.parent/'cmd/pptxdesign/main.go',scripts.parent/'go.mod',scripts.parent/'go.sum']
    provenance = dict(schema='pptxgengo.wmds-generator-snapshot.v1',engine=manifest['engine'],
        scope='Source snapshot at packet preparation; Go binary is a repository development build.',
        source_sha256={str(f.relative_to(scripts.parent)):sha(f) for f in source_files})
    with (root/'generator-source.json').open('x') as output:
        json.dump(provenance,output,indent=2);output.write('\n')
    artifacts['generator-source.json'] = sha(root/'generator-source.json')
receipt = dict(schema='pptxgengo.wmds-prepared-capture.v1',artifacts=artifacts,
               generator_manifest_sha256=sha(root/'probes.json'), capture_tool_revision=4,
               probe_count=len(manifest['probes']), slide_count=math.ceil(len(manifest['probes'])/3),
               scope='Capture tools and the exact generated deck. PDF evidence is independently hashed by pdf-export.json.')
with (root/'prepared.json').open('x') as output:
    json.dump(receipt,output,indent=2);output.write('\n')
instructions = '''# WMDS native typography capture

1. Open this packet's `typography-probes.pptx` in PowerPoint. The title bar should
   show **typography-probes** and the deck should have 25 slides. Do not edit it.
   Close any other deck with the same filename before opening this one.
2. Run `bash /ABSOLUTE/PACKET/PATH/capture.sh` from your usual Terminal.
   The packet directory is passed as the argument shown above. It captures 75
   objects and does not save, export or close any deck. Each run gets a new
   directory under `runs`.
3. Supply the resulting capture directory to the comparison command:

   `python3 /ABSOLUTE/PACKET/PATH/compare-wmds-typography.py --packet /ABSOLUTE/PACKET/PATH --capture-dir /PATH/TO/RUN --out /PATH/TO/NEW-COMPARISON.json`

The comparison also needs `typography-native.pdf`, `pdf-text.json`,
`baselines.json`, `pdf-export.json`, and `font-environment.json` in this packet.
Those are separate native PDF evidence, not fabricated character bounds.
The current packet includes them when prepared by the implementation session.
Generation alone does not qualify the engine. Expected failures are retained.
'''.replace('/ABSOLUTE/PACKET/PATH',str(root)).replace('25 slides',str(math.ceil(len(manifest['probes'])/3))+' slides').replace('75\n   objects',str(len(manifest['probes']))+'\n   objects')
instructions = instructions.replace('bash '+str(root)+'/capture.sh', 'bash '+str(root)+'/capture.sh '+str(root))
with (root/'CAPTURE.md').open('x') as output:
    output.write(instructions)
print('Prepared: '+str(root))
print('Run: bash '+str(root)+'/capture.sh '+str(root))
