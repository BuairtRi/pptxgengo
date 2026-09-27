#!/usr/bin/env python3
"""Build bounded Wave 2 follow-up fixtures; originals remain in the v7 checkpoint."""
import copy
import json
import subprocess
import sys
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'library/visual-components'
NS = {'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
      'p': 'http://schemas.openxmlformats.org/presentationml/2006/main'}


def run(text, identity, bold=False, color='#070154', size=12, underline=False):
    return dict(id=identity, text=text, font_face='Arial', font_size_pt=size,
                bold=bold, foreground=color, underline=underline)


def paragraph(identity, runs, bullet=None, after=3):
    result = dict(id=identity, align='left', space_after_pt=after, runs=runs)
    if bullet:
        result['bullet'] = dict(character=bullet, margin_left_pt=23.04, hanging_pt=22.5)
    return result


def rect(x, y, w, h):
    return dict(x=x, y=y, width=w, height=h)


def main():
    for script in ['build-visual-wave2-spec.py', 'build-wave2-people-spec.py', 'build-wave2-response-spec.py']:
        subprocess.run([sys.executable, str(ROOT / 'scripts' / script)], cwd=ROOT, check=True)
    deliverable = json.loads((OUT / 'deliverable-review.json').read_text())['slides'][0]
    bio = json.loads((OUT / 'people-review.json').read_text())['slides'][1]
    response = json.loads((OUT / 'response-review.json').read_text())['slides'][0]
    # Promote only this follow-up fixture to the new native picture contract.
    border_ids = {c['id'] for c in deliverable['canvas'] if c['id'].startswith('source-border-')}
    deliverable['canvas'] = [c for c in deliverable['canvas'] if c['id'] not in border_ids]
    for c in deliverable['canvas']:
        if 'allow_overlap' in c:
            c['allow_overlap'] = [peer for peer in c['allow_overlap'] if peer not in border_ids]
        if c['id'].startswith('source-pic-'):
            c.update(outline_color='#CED7E6', outline_width_pt=.75)
    deliverable['notes'] += ' Follow-up: pictures use native 0.75pt #CED7E6 outlines rather than four separate line objects.'
    manifest = json.loads((ROOT / 'samples/visual-wave2/response-assets.json').read_text())
    assets = manifest['items']
    for c in response['canvas']:
        matching = next((a for a in assets if a['derived_sha256'] == c.get('asset_sha256')), None)
        if matching:
            c['fallback_asset_path'] = c['asset_path']
            c['fallback_asset_sha256'] = c['asset_sha256']
            c['asset_path'] = matching['source_file']
            c['asset_sha256'] = matching['source_sha256']
    response['notes'] = 'ILLUSTRATIVE CHANGED-CONTENT FIXTURE. EnableComp5 icons are original pinned native SVGs with explicit pinned AppKit PNG fallbacks. Both are embedded and hashed; vector artwork remains a picture, not editable individual paths. Need/response copy and source triangle-to-rightArrow adaptation follow the v7 checkpoint; no whole-slide identity claim.'
    base = copy.deepcopy(deliverable)
    base.update(id='native-rich-bullet-control', title='Native bullets preserve structure through wrapping',
                role='SOURCE BULLET CONTROL AND CHANGED-CONTENT VARIANT',
                takeaway='A hanging indent keeps continuation lines aligned with the text while emphasis remains editable.',
                notes='Right: UHG67 industries panel source text, 12pt Arial, 23.04pt text margin and 22.5pt hanging indent, 3pt paragraph spacing. Heading 14pt with 6pt after. Panel geometry from layout 80; source has explicit Arial bullet font, this fixture follows Arial text font. Left: changed content with inline emphasis and three supported glyph variants. Not a whole-slide identity claim.')
    base['title_bounds']['width'] = 888
    base['canvas'] = [x for x in base['canvas'] if x['id'] in ['footer-band', 'wm-logo', 'footer-copy', 'page']]
    with zipfile.ZipFile(ROOT / 'samples/UHG Fabric Platforming RFP Response - July 2026.pptx') as z:
        root = ET.fromstring(z.read('ppt/slides/slide67.xml'))
    shape = next(s for s in root.findall('.//p:sp', NS) if s.find('.//p:cNvPr', NS).get('id') == '7')
    texts = [''.join(t.text or '' for t in p.findall('.//a:t', NS)) for p in shape.findall('.//a:p', NS)]
    source = [paragraph('source-heading', [run(texts[0], 'heading', True, '#50658E', 14, True)], after=6)]
    source += [paragraph(f'source-{i}', [run(t, 'text')], '•') for i, t in enumerate(texts[1:6], 1)]
    base['canvas'] += [
        dict(id='source-panel', kind='surface', bounds=rect(635.951, 106.09, 272.424, 178.91), background='#E8EEF8', allow_overlap=['source-bullets']),
        dict(id='source-bullets', kind='text', bounds=rect(650.351, 120.49, 243.624, 150.11), paragraphs=source, align='left', valign='top', inset_x=0, inset_y=0, layer=10),
        dict(id='variant-panel', kind='surface', bounds=rect(36, 106.09, 560, 266), background='#E8EEF8', allow_overlap=['variant-bullets']),
        dict(id='variant-bullets', kind='text', bounds=rect(52, 122.09, 528, 234), align='left', valign='top', inset_x=0, inset_y=0, layer=10,
             paragraphs=[
                 paragraph('variant-heading', [run('Proposed delivery responsibilities', 'text', True, size=14)], after=10),
                 paragraph('variant-1', [run('Establish evidence: ', 'lead', True), run('agree baseline measures, document their source and define how each measure informs a delivery decision.', 'body')], '•', after=9),
                 paragraph('variant-2', [run('Manage dependencies: ', 'lead', True, '#0047FF'), run('connect service ownership, access requirements and integration decisions before committing to the pilot sequence.', 'body')], '–', after=9),
                 paragraph('variant-3', [run('Review and adapt: ', 'lead', True), run('carry inline emphasis across wrapped lines while keeping the bullet and continuation text on their declared columns.', 'body')], '▪', after=9),
             ]),
        dict(id='source-label', kind='text', bounds=rect(635.951, 300, 272.424, 42), text='UHG67 source industries panel\nExact copy and indent geometry', font_face='Arial', font_size_pt=10, foreground='#50658E', align='left', valign='top'),
        dict(id='fixture-note', kind='text', bounds=rect(36, 398, 870, 55), text='The left panel is illustrative proposed content. The right panel reproduces a source component for typography comparison. Bullets and inline emphasis are native editable text; unsupported numbering and custom bullet fonts are rejected.', font_face='Arial', font_size_pt=11, foreground='#070154', align='left', valign='top'),
    ]
    slides = [base, bio, deliverable, response]
    for i, slide in enumerate(slides, 1):
        footer = next(c for c in slide['canvas'] if c['id'] in ['page', 'page-number'])
        footer['text'] = str(i)
    document = dict(schema='pptxgengo.compose-spec.v1', slides=slides)
    (OUT / 'followup-review.json').write_text(json.dumps(document, indent=2) + '\n')
    negative = copy.deepcopy(base)
    negative['id'] = 'native-bullet-overflow'
    negative['canvas'] = [copy.deepcopy(next(x for x in base['canvas'] if x['id'] == 'variant-bullets'))]
    negative['canvas'][0]['bounds']['height'] = 30
    (OUT / 'followup-negative.json').write_text(json.dumps(dict(schema=document['schema'], slides=[negative]), indent=2) + '\n')
    (OUT / 'followup-qualification.json').write_text(json.dumps(dict(schema=document['schema'], slides=slides + [negative]), indent=2) + '\n')
    print(OUT / 'followup-review.json')


if __name__ == '__main__':
    main()
