#!/usr/bin/env python3
"""Compare archived, validated native probes with Go predictions; write new output.

Checks receipt, deck, spec, manifest, adapter and font provenance. Native line
boxes and PDF glyph baselines are separate contracts. Does not invoke PowerPoint.
"""
import argparse
import hashlib
import json
from pathlib import Path
import statistics
import struct
import unicodedata

ROOT = Path(__file__).resolve().parents[1]
METRICS = ('offset_x_pt', 'offset_y_pt', 'rendered_width_pt', 'rendered_height_pt')


def load(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(ok, message):
    if not ok:
        raise ValueError(message)


def normalized(text):
    return ''.join(unicodedata.normalize('NFKC', text).split())


def stats(values):
    return dict(count=len(values), mean_absolute_delta_pt=statistics.mean(map(abs, values)),
                max_absolute_delta_pt=max(map(abs, values)),
                within_0_01_pt_count=sum(abs(v) <= .01 for v in values),
                underprediction_over_0_01_pt_count=sum(v < -.01 for v in values)) if values else dict(count=0)


def default_axes(path, index):
    # Native CoreText omits default axes. Read the selected collection face's
    # fvar defaults so omission is checked against file metadata, not guessed.
    raw = path.read_bytes()
    start = 0
    if raw[:4] == b'ttcf':
        count = struct.unpack_from('>I', raw, 8)[0]
        require(index < count, 'Font collection face missing')
        start = struct.unpack_from('>I', raw, 12 + 4*index)[0]
    tables = {}
    for i in range(struct.unpack_from('>H', raw, start+4)[0]):
        tag, _, offset, length = struct.unpack_from('>4sIII', raw, start+12+16*i)
        tables[tag] = raw[offset:offset+length]
    fv = tables.get(b'fvar')
    if not fv:
        return {}
    offset, _, count, size = struct.unpack_from('>HHHH', fv, 4)
    return {fv[offset+i*size:offset+i*size+4].decode('ascii'):
            struct.unpack_from('>i', fv, offset+i*size+8)[0]/65536 for i in range(count)}


def check_fonts(native, candidate):
    key = lambda f: (f['family'].casefold(), f['style'])
    actual = {key(f): f for f in native['environment']['fonts']}
    expected = {key(f): f for f in candidate['fonts']}
    require(actual.keys() == expected.keys(), 'Native/Go font family/style sets differ')
    for k, g in expected.items():
        n = actual[k]
        require(n['sha256'] == g['sha256'], 'Native/Go font hash differs: '+str(k))
        path = Path(g['file'])
        require(sha(path) == g['sha256'], 'Current font changed: '+str(k))
        axes = default_axes(path, g['face_index'])
        axes.update({int(tag).to_bytes(4, 'big').decode('ascii'): value
                     for tag, value in n.get('variations', {}).items()})
        require(axes == g.get('axes', {}), 'Native/Go font axes differ: '+str(k))


def compare(data, name, candidate, path, receipt):
    folder = data/name
    native = load(folder/'native.json')
    manifest = load(folder/'manifest.json')
    old = load(folder/'v3-go.json')
    cases_path = ROOT/'library/dynamic-components'/('font-reference/cases.json' if name == 'reference' else 'font-wrap-calibration/heldout/cases.json')
    cases = {c['id']: c for c in load(cases_path)['cases']}
    require(sha(folder/'native.json') == receipt['evidence_sha256'], 'Native evidence hash mismatch')
    for k, artifact_path in [('spec_sha256', folder/'spec.json'), ('manifest_sha256', folder/'manifest.json'), ('deck_sha256', folder/manifest['deck_file'])]:
        require(native[k] == receipt[k] == sha(artifact_path), k+' mismatch')
    require(native['deck_sha256'] == manifest['deck_sha256'], 'Manifest deck hash mismatch')
    require(native['spec_sha256'] == manifest['spec_sha256'] == candidate['spec_sha256'] == old['spec_sha256'], 'Go spec mismatch')
    require(native['adapter_sha256'] == native['environment']['adapter_sha256'], 'Adapter provenance mismatch')
    require(not candidate['powerpoint_verified'] and not old['powerpoint_verified'], 'Go claims native verification')
    check_fonts(native, candidate)
    require(old['fonts'] == candidate['fonts'], 'Candidate font resolution changed')
    require(len(manifest['requests']) == len(native['native']['measurements']) == len(candidate['requests']) == receipt['request_count'], 'Count mismatch')
    ids = {q['id'] for q in manifest['requests']}
    require(ids == candidate['requests'].keys() == native['measurements']['by_request_id'].keys(), 'Request set mismatch')
    rows = []
    for i, (q, slide, raw) in enumerate(zip(manifest['requests'], manifest['slides'], native['native']['measurements'])):
        rid = q['id']
        element = slide['elements'][0]
        require(element['measurement_id'] == slide['id'] == rid, 'Request order mismatch')
        require(raw['slide_index'] == i+1 and element['name'] == raw['shape_name'], 'Native shape/order mismatch')
        for k, rk in [('x','left'),('y','top'),('width','width'),('height','height')]:
            require(abs(element['frame'][k]-raw['shape_frame'][rk]) < .0001, 'Native raw units differ from slide points')
        source = '\n'.join(''.join(r['text'] for r in p['runs']) for p in q['paragraphs']) if q.get('paragraphs') else q['text']
        require(raw['text'].replace('\r\n','\n').replace('\r','\n') == source, 'Native source mismatch')
        groups = {}
        for c in raw['characters']:
            if c['text'].strip():
                require(c['bounds']['width'] > 0 and c['bounds']['height'] > 0, 'Missing occupied character bounds')
                top = round(c['bounds']['top']-raw['shape_frame']['top'], 4)
                groups.setdefault(top, []).append(c)
        native_lines = [dict(text=''.join(c['text'] for c in chars), y_pt=y,
            height_pt=max(c['bounds']['height'] for c in chars)) for y, chars in sorted(groups.items())]
        now = candidate['requests'][rid]
        visible = [l for l in now['lines'] if l['text'].strip()]
        before = old['requests'][rid]
        # This pass must not change wrapping or horizontal fit decisions.
        fields = ('text','start_rune','end_rune','x_pt','advance_pt')
        require([[l[k] for k in fields] for l in now['lines']] == [[l[k] for k in fields] for l in before['lines']], 'Candidate wrapping/advance changed')
        require(now.get('boundary_warnings', []) == before.get('boundary_warnings', []), 'Wrap warning changed')
        match = [normalized(l['text']) for l in visible] == [normalized(l['text']) for l in native_lines]
        actual = native['measurements']['by_request_id'][rid]
        bounds = raw['text_bounds']
        reconstructed = dict(offset_x_pt=bounds['left']-raw['shape_frame']['left'],
            offset_y_pt=bounds['top']-raw['shape_frame']['top'], rendered_width_pt=bounds['width'], rendered_height_pt=bounds['height'])
        require(all(abs(actual.get(k, 0)-reconstructed[k]) < .0001 for k in METRICS), 'Converted native bounds mismatch')
        delta = lambda report: {k: report['measurements']['by_request_id'][rid].get(k,0)-actual.get(k,0) for k in METRICS}
        row = dict(case_id=q['slide_id'], family=cases[q['slide_id']]['family'], category=cases[q['slide_id']]['category'],
            request_id=rid, line_breaks_match=match, native=actual, go=candidate['measurements']['by_request_id'][rid],
            v3_delta_pt=delta(old), delta_pt=delta(candidate), native_lines=native_lines,
            go_lines=[{k:l[k] for k in ('text','y_pt','height_pt','baseline_pt')} for l in visible],
            height_warnings=now.get('height_warnings', []), boundary_warnings=now.get('boundary_warnings', []),
            horizontal_contract_note='Go includes bullet glyph/ink; native character union excludes bullet glyph' if any(p.get('bullet') for p in q.get('paragraphs', [])) else 'Go includes glyph ink horizontally; native uses character advance bounds')
        if match:
            row['line_top_delta_pt'] = [l['y_pt']-n['y_pt'] for l,n in zip(visible,native_lines)]
            row['line_height_delta_pt'] = [l['height_pt']-n['height_pt'] for l,n in zip(visible,native_lines)]
        rows.append(row)
    return dict(name=name, count=len(rows), native_sha256=sha(folder/'native.json'),
        candidate_sha256=sha(path), case_metadata_sha256=sha(cases_path),
        line_break_match_count=sum(r['line_breaks_match'] for r in rows),
        before={k:stats([r['v3_delta_pt'][k] for r in rows]) for k in METRICS},
        after={k:stats([r['delta_pt'][k] for r in rows]) for k in METRICS},
        matching_wrap_height=dict(before=stats([r['v3_delta_pt']['rendered_height_pt'] for r in rows if r['line_breaks_match']]),
            after=stats([r['delta_pt']['rendered_height_pt'] for r in rows if r['line_breaks_match']])),
        line_tops=stats([v for r in rows for v in r.get('line_top_delta_pt', [])]),
        line_heights=stats([v for r in rows for v in r.get('line_height_delta_pt', [])]), rows=rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--data', type=Path, default=ROOT/'library/dynamic-components/font-native-calibration')
    parser.add_argument('--reference-go', type=Path, required=True)
    parser.add_argument('--additional-go', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    receipt = load(args.data/'capture-receipt.json')
    require(sha(args.data/'prepared.json') == receipt['prepared_receipt_sha256'], 'Prepared receipt mismatch')
    records = {s['name']: s for s in receipt['sets']}
    candidate_path = dict(reference=args.reference_go, additional=args.additional_go)
    candidates = {k:load(v) for k,v in candidate_path.items()}
    require(candidates['reference']['engine'] == candidates['additional']['engine'], 'Candidate engine mismatch')
    report = dict(schema='pptxgengo.font-native-comparison.v1', engine=candidates['reference']['engine'],
        height_policy=candidates['reference']['height_policy'], powerpoint_verified=False,
        capture_receipt_sha256=sha(args.data/'capture-receipt.json'), font_hashes_and_axes_match=True,
        raw_coordinate_scale=1, scope='393 validated native probes; layout predictions, not final-deck verification',
        sets=[compare(args.data,k,candidates[k],candidate_path[k],records[k]) for k in ('reference','additional')])
    # Preflight repeats retain their hashes even though the main comparison uses
    # the 393 unique corpus requests only.
    preflight = args.data/'trailing-break'
    evidence = load(preflight/'native.json'); manifest = load(preflight/'manifest.json')
    require(sha(preflight/'native.json') == records['trailing-break']['evidence_sha256'], 'Preflight changed')
    for key,path in [('spec_sha256',preflight/'spec.json'),('manifest_sha256',preflight/'manifest.json'),('deck_sha256',preflight/manifest['deck_file'])]:
        require(evidence[key] == records['trailing-break'][key] == sha(path), 'Preflight artifact changed')
    with args.out.open('x') as stream:
        json.dump(report, stream, indent=2, ensure_ascii=False, allow_nan=False); stream.write('\n')
    print(args.out)
    for s in report['sets']:
        print(s['name'], 'wrapping', str(s['line_break_match_count'])+'/'+str(s['count']), 'matched-wrap height', s['matching_wrap_height']['after'])


if __name__ == '__main__':
    main()
