#!/usr/bin/env python3
"""QA-only control: render regenerated target slide bodies in the full source deck.

Unselected source slides are untouched. This fixture is NOT the reconstructed
deliverable. It isolates PowerPoint's deck-context-dependent export behavior.
"""
import json
from pathlib import Path
import subprocess
import zipfile


def run(args):
    subprocess.run([str(a) for a in args], check=True)


root = Path(__file__).resolve().parent.parent
base = root / 'samples/reconstruction'
numbers = [8, 11, 12, 14, 24, 28, 36, 38, 43, 44, 67, 5]
out = base / 'work/source-context-validation.pptx'
pdf = base / 'reference/source-context-validation.pdf'
qa = base / 'qa/source-context'
if out.exists() or pdf.exists() or qa.exists():
    raise SystemExit('Validation evidence already exists; refusing overwrite')
replacement = {}
for n in numbers:
    scene = json.loads((base / f'parameterized-scene/slides/uhg-{n:03d}.json').read_text())
    part = scene['part']
    rel = str(Path(part).parent / '_rels' / (Path(part).name + '.rels'))
    with zipfile.ZipFile(base / f'work/uhg-{n:03d}-bound.pptx') as compiled:
        for name in (part, rel):
            replacement[name] = compiled.read(name)
source = root / 'samples/UHG Fabric Platforming RFP Response - July 2026.pptx'
with zipfile.ZipFile(source) as original, zipfile.ZipFile(out, 'w', zipfile.ZIP_DEFLATED) as dest:
    for entry in original.infolist():
        dest.writestr(entry.filename, replacement.get(entry.filename, original.read(entry.filename)))
run(['osascript', root / 'scripts/export-powerpoint.applescript', out, pdf])
run(['swift', root / 'scripts/render-pdf.swift', pdf, qa / 'png', *numbers])
results = []
for n in numbers:
    run(['go', 'run', './cmd/pptxdiff', '--reference', base / f'reference/png/slide-{n:03d}.png',
         '--candidate', qa / f'png/slide-{n:03d}.png', '--out', qa / f'{n:03d}-diff'])
    result = json.loads((qa / f'{n:03d}-diff/report.json').read_text())
    results.append({'slide': n, **result})
(qa / 'report.json').write_text(json.dumps({'purpose': __doc__, 'results': results}, indent=2) + '\n')
print(json.dumps([{'slide': r['slide'], 'mismatch_pixels': r['exact_mismatch_count']} for r in results]))
