#!/usr/bin/env python3
"""Compare pinned WMDS controls with separate PDF and native contracts.

PDF selections are not PowerPoint character bounds. This tool never upgrades
the authoring engine's qualification or fits correction constants automatically.
"""
import argparse
import hashlib
import json
import math
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path


def require(ok, message):
    if not ok:
        raise ValueError(message)


def load(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def text(value):
    return (value or '').replace('\r\n', '\n').replace('\r', '\n')


def font_name(value):
    return value.split('+')[-1]


def stats(values):
    return dict(count=len(values), max_absolute_pt=max(map(abs, values), default=0),
                mean_absolute_pt=sum(map(abs, values))/len(values) if values else 0)


def package_paragraph_contracts(root, manifest):
    """Independent saved DrawingML check; no native observation is inferred."""
    if manifest['engine'] != 'wmds-go-foundation.v2':
        return {}
    ns = {'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
          'p': 'http://schemas.openxmlformats.org/presentationml/2006/main'}
    rows, slides = {}, {}
    with zipfile.ZipFile(root/'typography-probes.pptx') as deck:
        for probe in manifest['probes']:
            slide = probe['slide']
            if slide not in slides:
                slides[slide] = ET.fromstring(deck.read(f'ppt/slides/slide{slide}.xml'))
            shapes = [sp for sp in slides[slide].findall('.//p:sp', ns)
                      if sp.find('p:nvSpPr/p:cNvPr', ns).get('name') == probe['id']]
            require(len(shapes) == 1, 'Package shape identity mismatch: '+probe['id'])
            paras = shapes[0].findall('p:txBody/a:p', ns)
            layout = probe['record']['layout']; style = layout['style']; font = layout['font']
            displayed = '\n'.join(''.join(t.text or '' for t in para.findall('.//a:t', ns)) for para in paras)
            require(displayed == layout['displayed'], 'Package text mismatch: '+probe['id'])
            require(len(paras) == len(displayed.split('\n')), 'Package paragraph mismatch: '+probe['id'])
            defaults_match = True
            for para in paras:
                end = para.find('a:endParaRPr', ns)
                if end is None:
                    defaults_match = False; continue
                latin = end.find('a:latin', ns); color = end.find('a:solidFill/a:srgbClr', ns)
                defaults_match = defaults_match and (
                    latin is not None and latin.get('typeface') == font['pptx_typeface'] and
                    color is not None and color.get('val') == probe['record']['color'] and
                    end.get('b') == str(int(font['pptx_bold'])) and
                    end.get('i') == str(int(font['pptx_italic'])) and end.get('kern') == '0' and
                    abs(float(end.get('sz', '-1'))/100-style['size']) <= .001 and
                    abs(float(end.get('spc', '-999999'))/100-style['tracking_pt']) <= .001)
            require(defaults_match, 'Package paragraph defaults mismatch: '+probe['id'])
            rows[slide, probe['id']] = dict(paragraph_count=len(paras), end_defaults_match=defaults_match)
    return rows


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--packet', required=True, type=Path)
    parser.add_argument('--capture-dir', type=Path)
    parser.add_argument('--out', required=True, type=Path)
    parser.add_argument('--require-parity', action='store_true')
    args = parser.parse_args()
    require(not args.out.exists(), 'Output must not exist')
    root = args.packet
    m = load(root/'probes.json')
    require(m['schema'] == 'pptxgengo.wmds-typography-probes.v1', 'Unknown probe schema')
    require(m['deck_sha256'] == sha(root/'typography-probes.pptx'), 'Deck drift')
    selection, bases = load(root/'pdf-text.json'), load(root/'baselines.json')
    export = load(root/'pdf-export.json')
    pdf_hash = sha(root/'typography-native.pdf')
    require(selection['contract'] == 'powerpoint-pdf-selection-bounds.v1', 'Unknown PDF selection contract')
    require(bases['contract'] == 'powerpoint-pdf-text-show-origins.v1', 'Unknown PDF baseline contract')
    require(selection['pdf_sha256'] == bases['pdf_sha256'] == export['pdf_sha256'] == pdf_hash, 'PDF drift')
    require(export['deck_sha256'] == m['deck_sha256'], 'PDF export/deck provenance mismatch')
    package_paras = package_paragraph_contracts(root, m)
    count = math.ceil(len(m['probes'])/3)
    require(len(bases['pages']) == selection['page_count'] == len(selection['pages']) == count, 'Page count mismatch')
    ids = {(p['slide'], p['id']) for p in m['probes']}
    require(len(ids) == len(m['probes']), 'Duplicate probe identity')
    native, environment = None, load(root/'font-environment.json')
    receipt = None
    if args.capture_dir:
        cap = args.capture_dir
        receipt = load(cap/'capture-receipt.json')
        require(receipt['schema'] == 'pptxgengo.wmds-native-capture.v1', 'Unknown capture schema')
        require(receipt.get('presentation_saved_before_and_after') is True, 'Saved presentation checks missing')
        require(receipt['deck_sha256'] == m['deck_sha256'] and receipt['manifest_sha256'] == sha(root/'probes.json'), 'Capture/deck mismatch')
        require(receipt['prepared_sha256'] == sha(root/'prepared.json'), 'Prepared receipt drift')
        prepared = load(root/'prepared.json')
        for name, expected_hash in prepared['artifacts'].items():
            require(sha(root/name) == expected_hash, 'Prepared artifact drift: '+name)
        require(receipt['native_sha256'] == sha(cap/'native.json'), 'Native capture drift')
        require(receipt['environment_sha256'] == sha(cap/'environment-before.json'), 'Capture environment drift')
        require(load(cap/'environment-before.json') == load(cap/'environment-after.json'), 'Capture environment changed')
        environment = load(cap/'environment-before.json')
        raw = load(cap/'native.json')
        require(raw['schema'] in ('pptxgengo.compose-text-measurement.v8','pptxgengo.wmds-text-measurement.v9'), 'Unknown native schema')
        native = {(q['slide_index'], q['shape_name']): q for q in raw['measurements']}
        require(len(native) == len(raw['measurements']) and set(native) == ids, 'Native inventory mismatch')
    rows, baseline_deltas, width_deltas, leading_deltas = [], [], [], []
    for q in m['probes']:
        record = q['record']; rect = record['rect']; layout = record['layout']
        y, h = rect['y_pt'], rect['height_pt']
        page = selection['pages'][q['slide']-1]
        require(page['width_pt'] == 960 and page['height_pt'] == 540, 'PDF page geometry mismatch')
        # Selection boxes can start above a text frame at tight leading. Their
        # vertical centers assign them to one of the separated control rows.
        lines = [l for l in page['lines'] if y <= l['bounds']['y']+l['bounds']['height']/2 < y+h]
        lines.sort(key=lambda l: l['bounds']['y'])
        predicted = [l for l in layout['lines'] if l['text'].strip()]
        actual_text = [text(l['text']).strip() for l in lines]
        expected_text = [l['text'].strip() for l in predicted]
        origins = sorted(set(round(o['baseline_pt'], 4) for o in bases['pages'][q['slide']-1]['origins'] if y <= o['baseline_pt'] < y+h))
        require(len(origins) == len(lines), 'PDF origin/selection disagreement: '+q['id'])
        require(''.join(actual_text).replace(' ', '') == ''.join(expected_text).replace(' ', ''), 'PDF content mismatch: '+q['id'])
        wrap_match = expected_text == actual_text
        row = dict(id=q['id'], case=q['case'], style=layout['style']['id'],
                   predicted_lines=expected_text, pdf_lines=actual_text, wrap_match=wrap_match,
                   pdf_baselines_from_frame_pt=[o-y for o in origins],
                   pdf_postscript_present=layout['font']['postscript_name'] in {font_name(n) for n in page['font_names']},
                   native_checks=None)
        if wrap_match:
            bd = [o-y-p['baseline_pt'] for o, p in zip(origins, predicted)]
            wd = [l['bounds']['width']-p['advance_pt'] for l, p in zip(lines, predicted)]
            row.update(baseline_delta_pt=bd, pdf_selection_width_delta_pt=wd,
                       baseline_within_half_point=all(abs(d) <= .5 for d in bd))
            baseline_deltas.extend(bd); width_deltas.extend(wd)
            pitches = [(origins[i]-origins[i-1])-(predicted[i]['baseline_pt']-predicted[i-1]['baseline_pt']) for i in range(1,len(origins))]
            row['baseline_pitch_delta_pt'] = pitches
            leading_deltas.extend(pitches)
        if native is not None:
            n = native[q['slide'], q['id']]
            content_match = text(n['text']) == layout['displayed']
            frame = n['shape_frame']; st = layout['style']; identity = layout['font']
            frame_match = all(abs(frame[k]-rect[v]) <= .15 for k, v in [('left','x_pt'),('top','y_pt'),('width','width_pt'),('height','height_pt')]) and abs(frame['rotation_degrees']) <= .001
            margin_match = all(abs(v) <= .01 for v in n['margins'].values())
            font_match = all(c['font_name'] == identity['pptx_typeface'] and c['bold'] == identity['pptx_bold'] and c['italic'] == identity['pptx_italic'] and abs(c['font_size_pt']-st['size']) <= .01 for c in n['characters'])
            # Paragraph returns are removed by split; visible characters are retained.
            character_inventory_match = ''.join(''.join(c['text'] for c in n['characters']).split()) == ''.join(text(n['text']).split())
            para_defaults_match = all(p.get('font_name') == identity['pptx_typeface'] and p.get('font_size_pt') is not None and abs(p['font_size_pt']-st['size'])<=.01 and p.get('bold') == identity['pptx_bold'] and p.get('italic') == identity['pptx_italic'] for p in n['paragraphs']) if raw['schema']=='pptxgengo.wmds-text-measurement.v9' else None
            para_match = all(not p['line_rule_within'] and abs(p['space_within']-st['leading']) <= .01 and abs(p['space_before_pt']) <= .01 and abs(p['space_after_pt']) <= .01 and not p['bullet_visible'] for p in n['paragraphs'])
            groups = []
            for c in n['characters']:
                cy = c['bounds']['top']
                group = next((g for g in groups if abs(g['top']-cy) <= .1), None)
                if group is None:
                    group = dict(top=cy, text=''); groups.append(group)
                group['text'] += c['text']
            native_lines = [g['text'].strip() for g in sorted(groups,key=lambda g:g['top']) if g['text'].strip()]
            raw_bounds = n['text_bounds']
            tb = raw_bounds
            if m['engine'] == 'wmds-go-foundation.v2':
                # Empty terminal TextRange snapshots and paragraph separators are
                # preserved in raw data but are not visible character occupancy.
                visible = [c['bounds'] for c in n['characters'] if text(c['text']).strip()]
                if visible:
                    left=min(b['left'] for b in visible);top=min(b['top'] for b in visible)
                    tb=dict(left=left,top=top,width=max(b['left']+b['width'] for b in visible)-left,
                            height=max(b['top']+b['height'] for b in visible)-top)
                else:
                    tb=dict(left=frame['left'],top=frame['top'],width=0,height=0)
            overflow = max(0, frame['left']-tb['left'], frame['top']-tb['top'], tb['left']+tb['width']-frame['left']-frame['width'], tb['top']+tb['height']-frame['top']-frame['height'])
            occupied_delta = tb['height']-layout['estimated_occupied_height_pt']
            top_delta = tb['top']-frame['top']-layout.get('occupied_top_pt',0)
            native_wrap_match = native_lines == expected_text
            authored_count = len(layout['displayed'].split('\n'))
            direct_count_match = authored_count == len(n['paragraphs'])
            expected_collection_count = authored_count
            terminal_omitted = False
            paragraph_contents_match = None
            package_contract = package_paras.get((q['slide'], q['id']))
            if m['engine'] == 'wmds-go-foundation.v2' and raw['schema'] == 'pptxgengo.wmds-text-measurement.v9':
                # In this captured PowerPoint v9 adapter data the collection
                # covers characters and excludes a zero-length final paragraph.
                # Preserve authored/raw content; require exact snapshot content
                # reconstruction plus the independently checked saved package.
                terminal_omitted = layout['displayed'].endswith('\n')
                expected_collection_count -= int(terminal_omitted)
                paragraph_contents_match = ''.join(text(p.get('content')) for p in n['paragraphs']) == text(n['text'])
                paragraph_count_match = (content_match and paragraph_contents_match and
                    len(n['paragraphs']) == expected_collection_count and
                    package_contract['paragraph_count'] == authored_count)
            else:
                paragraph_count_match = direct_count_match
            row['native_checks'] = dict(content_match=content_match,observed_content=text(n['text']),
                expected_paragraph_count=len(layout['displayed'].split('\n')),observed_paragraph_count=len(n['paragraphs']),
                paragraph_count_match=paragraph_count_match,
                authored_count_equals_native_collection_count=direct_count_match,
                expected_native_collection_count=expected_collection_count,
                paragraph_snapshot_contents_match_range=paragraph_contents_match,
                terminal_empty_paragraph_not_enumerated=terminal_omitted,
                terminal_empty_defaults_native_observation='not enumerated; saved package defaults checked' if terminal_omitted else 'not applicable',
                package_paragraph_contract=package_contract,
                frame_match=frame_match,margins_match=margin_match,font_attributes_match=font_match,
                character_inventory_match=character_inventory_match,paragraph_font_defaults_match=para_defaults_match,
                paragraph_attributes_match=para_match,lines=native_lines,wrap_match=native_wrap_match,
                character_union_height_pt=tb['height'],raw_character_union_bounds=raw_bounds,comparison_character_union_bounds=tb,
                occupied_union_policy='visible_non_whitespace_characters' if m['engine']=='wmds-go-foundation.v2' else 'raw_adapter_union',
                predicted_occupied_height_delta_pt=occupied_delta,predicted_occupied_top_delta_pt=top_delta,
                occupied_height_comparable=content_match and paragraph_count_match and native_wrap_match,
                occupied_height_within_half_point=abs(occupied_delta)<=.5,occupied_top_within_half_point=abs(top_delta)<=.5,positive_overflow_pt=overflow,
                no_overflow=overflow<=.15)
        rows.append(row)
    wrap_fails = [r['id'] for r in rows if not r['wrap_match']]
    baseline_fails = [r['id'] for r in rows if r.get('baseline_within_half_point') is False]
    font_fails = [r['id'] for r in rows if not r['pdf_postscript_present']]
    native_fails = [] if native is None else [r['id'] for r in rows if not all(r['native_checks'][k] for k in ['content_match','paragraph_count_match','frame_match','margins_match','font_attributes_match','paragraph_attributes_match','wrap_match','occupied_height_within_half_point','no_overflow'])]
    if native is not None and m['engine']=='wmds-go-foundation.v2':
        native_fails = sorted(set(native_fails+[r['id'] for r in rows if not r['native_checks']['character_inventory_match'] or r['native_checks']['paragraph_font_defaults_match'] is not True or not r['native_checks']['occupied_top_within_half_point']]))
    content_fails = [] if native is None else [r['id'] for r in rows if not r['native_checks']['content_match']]
    result = dict(schema='pptxgengo.wmds-typography-comparison.v1',profile=m['profile'],engine=m['engine'],
        qualification='native_parity_failed' if wrap_fails or baseline_fails or font_fails or native_fails else ('native_controls_passed_identity_unqualified' if native is not None else 'character_capture_and_identity_pending'),
        provenance=dict(deck_sha256=m['deck_sha256'],manifest_sha256=sha(root/'probes.json'),pdf_sha256=pdf_hash,
                        export_receipt_sha256=sha(root/'pdf-export.json'),capture_receipt=receipt,
                        analyzer_sha256=sha(Path(__file__).resolve())),
        summary=dict(probe_count=len(rows),wrap_matches=len(rows)-len(wrap_fails),wrap_failures=wrap_fails,
                     baseline_failures=baseline_fails,pdf_font_failures=font_fails,native_failures=native_fails,
                     native_content_matches=None if native is None else len(rows)-len(content_fails),native_content_failures=content_fails,
                     native_occupied_height_delta_on_matching_content_and_lines=None if native is None else stats([r['native_checks']['predicted_occupied_height_delta_pt'] for r in rows if r['native_checks']['occupied_height_comparable']]),
                     native_wrap_matches=None if native is None else sum(r['native_checks']['wrap_match'] for r in rows),
                     native_occupied_height_matches=None if native is None else sum(r['native_checks']['occupied_height_comparable'] and r['native_checks']['occupied_height_within_half_point'] for r in rows),
                     native_occupied_top_delta=None if native is None else stats([r['native_checks']['predicted_occupied_top_delta_pt'] for r in rows]),
                     native_positive_overflow_max_pt=None if native is None else max(r['native_checks']['positive_overflow_pt'] for r in rows),
                     terminal_empty_paragraphs_not_enumerated=None if native is None else sum(r['native_checks']['terminal_empty_paragraph_not_enumerated'] for r in rows),
                     baseline_delta=stats(baseline_deltas),pdf_selection_width_delta_diagnostic=stats(width_deltas),
                     baseline_pitch_delta=stats(leading_deltas),
                     rejection_control_count=len(m['rejections']),native_character_capture_present=native is not None),
        native_control_parity_passed=native is not None and not (wrap_fails or baseline_fails or font_fails or native_fails),
        qualified_envelope=False,exact_native_font_file_identity_verified=False,
        font_environment=environment,probes=rows,
        limits=['Only this bounded uniform Latin control set was compared; no general typography envelope is qualified.',
                'PDF selection width/height are not PowerPoint TextRange or native character bounds.',
                'CoreText resolver fingerprints are independent evidence, not proof of actual PowerPoint font file access.',
                'PDF font checks establish expected PostScript-name presence in each page resource dictionary, not a per-character font identity.',
                'First baseline and occupied height remain provisional; no correction is fitted by this command.',
                'Mixed runs, complex scripts, rotated text and footnotes are outside the current authoring API.',
                'v2/v9 terminal empty paragraphs are checked by exact native content plus saved package structure; the scripting collection omits the zero-length terminal entry in this capture.',
                'Final zero-length paragraph font defaults are not independently enumerated by this native API; saved DrawingML defaults are checked.',
                'v2 visible occupied bounds exclude whitespace, paragraph separators and empty terminal character snapshots; raw unions are retained.'])
    with args.out.open('x') as output:
        json.dump(result, output, indent=2);output.write('\n')
    print(json.dumps(dict(qualification=result['qualification'],summary=result['summary']),indent=2))
    if args.require_parity and result['qualification'] != 'qualified_envelope':
        raise SystemExit(2)


if __name__ == '__main__':
    main()
