#!/usr/bin/env python3
"""Integrate exact parallel source/alternate specimens; native review is separate."""
import argparse
import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
GROUPS = ('approach-openers-commercials', 'proof-evidence-argument', 'team-solution')


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--packet', type=Path, default=ROOT/'samples/wmds-refresh-slice4-20261002')
    parser.add_argument('--foundation-only', action='store_true', help='Prepare composition while source caller files are finishing')
    args = parser.parse_args()
    packet = args.packet
    inventory = read(ROOT/'planning/wm-design-contracts/v2/source-update-2026-10-02.json')
    changes = {t['key']: t for t in inventory['templates']}
    contracts = {t['key']: t for t in read(packet/'catalog.json')}
    assert len(contracts) == 167 and set(contracts) == set(changes)
    slides, values, manifest, seen = [], [], [], set()
    for group in GROUPS:
        folder = packet/'work'/group
        doc = read(folder/'combined.foundation.json')
        callers = {s['id']: s for s in read(folder/'bound-content.json')['slides']}
        assert len(doc['slides']) % 2 == 0
        for i in range(0, len(doc['slides']), 2):
            pair = doc['slides'][i:i+2]
            key = pair[0]['template_binding']['template']
            assert key not in seen and pair[1]['template_binding']['template'] == key
            seen.add(key)
            entry = contracts[key]
            record = dict(template=key, name=entry['name'], family=entry['family'],
                          status=entry.get('status', 'active'), change=changes[key]['change'],
                          source_commit=inventory['source_commit'], group=group,
                          source_page=len(slides)+1, alternate_page=len(slides)+2,
                          native_review='pending', replaced_by=entry.get('replaced_by'))
            for kind, slide in zip(('source', 'alternate'), pair):
                slide = copy.deepcopy(slide)
                old_id = slide['id']
                slide['id'] = key.replace('/', '--')+'-'+kind
                slide['template_binding']['slide_id'] = slide['id']
                assert slide['template_binding']['source_sha256'] == entry['source_sha256']
                slides.append(slide)
                if not args.foundation_only:
                    caller = copy.deepcopy(callers[old_id])
                    assert caller['template'] == key
                    caller['id'] = slide['id']
                    values.append(caller)
            manifest.append(record)
    assert seen == set(contracts) and len(slides) == 334
    write(packet/'reference.foundation.json', dict(schema='pptxgengo.wmds-foundation.v1', year=2026, slides=slides))
    write(packet/'manifest.json', manifest)
    if not args.foundation_only:
        assert len(values) == 334
        write(packet/'bound-content.json', dict(schema='pptxgengo.wmds-template-document.v1', year=2026, slides=values))
    print(f'Integrated {len(manifest)} designs / {len(slides)} specimens; native review pending.')


if __name__ == '__main__':
    main()
