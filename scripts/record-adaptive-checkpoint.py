#!/usr/bin/env python3
"""Record explicitly reviewed examples after checking their native evidence identities.

This does not perform or infer visual review. The caller supplies the IDs it has
actually inspected. A successful native verify writes the evidence file only
after its checks pass; failures use a separate .failed.json file.
"""
import argparse
import hashlib
import json
from pathlib import Path


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--root', type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument('--review', default='samples/adaptive/review.json')
    parser.add_argument('--checkpoint', default='library/adaptive/checkpoint.json')
    parser.add_argument('--visually-reviewed-ids', nargs='+', required=True)
    args = parser.parse_args()
    root = args.root.resolve()
    review_path = root / args.review
    checkpoint_path = root / args.checkpoint
    review = json.loads(review_path.read_text())
    requested = set(args.visually_reviewed_ids)
    rows = {row['id']: row for row in review['slides']}
    if len(rows) != len(review['slides']):
        raise ValueError('Review IDs must be unique')
    if missing := requested - rows.keys():
        raise ValueError(f'Unknown review IDs: {sorted(missing)}')
    checkpoint = json.loads(checkpoint_path.read_text())
    for ident in sorted(requested):
        row = rows[ident]
        evidence = row['evidence']
        paths = {
            'spec': root / row['spec_path'],
            'deck': root / evidence['deck'],
            'native_verification': root / evidence['native_verification'],
            'render': root / row['render_png'],
            'fit_report': root / evidence['fit_report'],
        }
        if row.get('values_path'):
            paths['values'] = root / row['values_path']
        hashes = {key + '_sha256': digest(path) for key, path in paths.items()}
        native = json.loads(paths['native_verification'].read_text())
        fit = json.loads((root / evidence['fit_report']).read_text())
        if native.get('schema') != 'pptxgengo.compose-evidence.v1':
            raise ValueError(f'{ident}: not successful native evidence')
        for key in ('spec_sha256', 'deck_sha256'):
            if native.get(key) != hashes[key]:
                raise ValueError(f'{ident}: native {key} mismatch')
        if fit.get('spec_sha256') != hashes['spec_sha256']:
            raise ValueError(f'{ident}: fit spec mismatch')
        if not fit.get('planner_passed') or fit.get('overflow_count') != 0 or fit.get('layout_failure_count') != 0:
            raise ValueError(f'{ident}: fit/planner did not pass')
        spec = json.loads(paths['spec'].read_text())
        slide_index = row['slide_index']
        if spec['slides'][slide_index - 1]['id'] != ident:
            raise ValueError(f'{ident}: slide index mismatch')
        if native.get('native', {}).get('visible_slide_count') != len(spec['slides']):
            raise ValueError(f'{ident}: native slide count mismatch')
        row['hashes'] = hashes
        row['status'] = 'visually_reviewed'
        native_finding = 'Native fit: no overflow; final PowerPoint verification passed.'
        row['findings'] = [native_finding] + [x for x in row.get('findings', []) if not x.startswith('Native fit:')]
        group_key = 'accents' if row['family'] == 'accent_arrow' else 'families'
        group = checkpoint.setdefault(group_key, {}).setdefault(row['family'], {})
        accepted = {x['id']: x for x in group.get('examples', [])}
        accepted[ident] = {k: v for k, v in row.items() if k not in ('source_png', 'findings')}
        group.update(status='bounded_examples_visually_reviewed',
                     example_count=sum(x.get('status') == 'visually_reviewed' for x in accepted.values()),
                     examples=list(accepted.values()))
    # Validate all requested rows before writing either record.
    review_path.write_text(json.dumps(review, indent=2) + '\n')
    checkpoint_path.write_text(json.dumps(checkpoint, indent=2) + '\n')
    print(f'Recorded {len(requested)} exact reviewed examples; broader scope unchanged.')


if __name__ == '__main__':
    main()
