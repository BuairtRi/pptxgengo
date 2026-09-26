#!/usr/bin/env python3
"""Compare delivered deck to source-slide control in the same shortened deck.

Only source slide-number fields are converted to literal text in the control,
matching the deliberate --freeze-slide-numbers behavior. Everything else in
each control slide body and its relationships comes from source ZIP bytes.
This control is QA evidence, not a reconstructed deliverable.
"""
import json
from pathlib import Path
import re
import subprocess
import zipfile

root = Path(__file__).resolve().parent.parent
base = root / 'samples/reconstruction'
slides = [8, 11, 12, 14, 24, 28, 36, 38, 43, 44]
out = base / 'work/delivery-source-control.pptx'
pdf = base / 'reference/delivery-source-control.pdf'
qa = base / 'qa/delivery-control'
if out.exists() or pdf.exists() or qa.exists():
    raise SystemExit('Control evidence already exists')
replacements = {}
with zipfile.ZipFile(root / 'samples/UHG Fabric Platforming RFP Response - July 2026.pptx') as source:
    for n in slides:
        scene = json.loads((base / f'parameterized-scene/slides/uhg-{n:03d}.json').read_text())
        part = scene['part']
        rel = str(Path(part).parent / '_rels' / (Path(part).name + '.rels'))
        text = source.read(part).decode()
        def freeze(m):
            children = re.sub(r'<a:pPr\b[^>]*(?:/>|>.*?</a:pPr>)', '', m.group(1), flags=re.S)
            return '<a:r>' + children + '</a:r>'
        text = re.sub(r'<a:fld\b[^>]*\btype="slidenum"[^>]*>(.*?)</a:fld>', freeze, text, flags=re.S)
        replacements[part] = text.encode()
        replacements[rel] = source.read(rel)
with zipfile.ZipFile(base / 'output/UHG-10-reconstructed.pptx') as candidate, zipfile.ZipFile(out, 'w', zipfile.ZIP_DEFLATED) as dest:
    for entry in candidate.infolist():
        dest.writestr(entry.filename, replacements.get(entry.filename, candidate.read(entry.filename)))

def run(args):
    subprocess.run([str(a) for a in args], check=True)

run(['osascript', root / 'scripts/export-powerpoint.applescript', out, pdf])
run(['swift', root / 'scripts/render-pdf.swift', pdf, qa / 'png', *range(1, 11)])
results = []
for i, n in enumerate(slides, 1):
    run(['go', 'run', './cmd/pptxdiff', '--reference', qa / f'png/slide-{i:03d}.png', '--candidate', base / f'qa/final/png/slide-{i:03d}.png', '--out', qa / f'{n:03d}-diff'])
    r = json.loads((qa / f'{n:03d}-diff/report.json').read_text())
    results.append({'slide': n, **r})
(qa / 'report.json').write_text(json.dumps(results, indent=2) + '\n')
print([(r['slide'], r['exact_mismatch_count']) for r in results])
