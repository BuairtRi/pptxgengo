#!/usr/bin/env python3
"""Analyze the archived focused capture and verified example without live apps.

Writes a new report. Does not promote Go predictions into native evidence.
"""
import argparse
import importlib.util
import json
from pathlib import Path
import xml.etree.ElementTree as ET
from zipfile import ZipFile

ROOT = Path(__file__).resolve().parents[1]
module = importlib.util.spec_from_file_location('native_comparison', ROOT/'scripts/compare-font-native.py')
native_comparison = importlib.util.module_from_spec(module)
module.loader.exec_module(native_comparison)
load, sha, require, stats = (getattr(native_comparison, key) for key in ('load', 'sha', 'require', 'stats'))


def theme_font(deck):
    with ZipFile(deck) as z:
        root = ET.fromstring(z.read('ppt/theme/theme1.xml'))
        ns = {'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}
        return root.find('a:themeElements/a:fontScheme/a:minorFont/a:latin', ns).attrib['typeface']


def verify_artifacts(folder, evidence, manifest):
    for key, path in [('spec_sha256', folder/'spec.json'), ('manifest_sha256', folder/'manifest.json'),
                      ('deck_sha256', folder/manifest['deck_file'])]:
        require(evidence[key] == sha(path), key+' mismatch')
    require(manifest['spec_sha256'] == evidence['spec_sha256'] and manifest['deck_sha256'] == evidence['deck_sha256'], 'Manifest provenance mismatch')
    require(evidence['adapter_sha256'] == evidence['environment']['adapter_sha256'], 'Adapter mismatch')
    require(evidence['native']['visible_slide_count'] == len(manifest['slides']), 'Slide count mismatch')


def lines(row):
    groups = {}
    for c in row['characters']:
        if c['text'].strip():
            top = round(c['bounds']['top']-row['shape_frame']['top'], 4)
            groups.setdefault(top, []).append(c)
    return [dict(text=''.join(c['text'] for c in chars), y_pt=y,
                 height_pt=max(c['bounds']['height'] for c in chars),
                 x_pt=min(c['bounds']['left'] for c in chars)-row['shape_frame']['left'],
                 right_pt=max(c['bounds']['left']+c['bounds']['width'] for c in chars)-row['shape_frame']['left'])
            for y, chars in sorted(groups.items())]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--data', type=Path, default=ROOT/'library/dynamic-components/font-native-calibration/focused')
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args(); data = args.data
    receipt = load(data/'capture-receipt.json')
    require(receipt['evidence_sha256'] == sha(data/'native.json'), 'Focused evidence changed')
    require(receipt['prepared_receipt_sha256'] == sha(data/'prepared.json'), 'Prepared receipt changed')
    evidence, manifest, go = [load(data/p) for p in ('native.json','manifest.json','v4-go.json')]
    require(not go['powerpoint_verified'], 'Go incorrectly claims native verification')
    verify_artifacts(data, evidence, manifest)
    native_comparison.check_fonts(evidence, go)
    require(go['spec_sha256'] == evidence['spec_sha256'], 'Go source mismatch')
    require(len(manifest['requests']) == len(evidence['native']['measurements']) == len(go['requests']) == receipt['request_count'] == 45, 'Focused count mismatch')
    rows = []
    for index,(q,slide,raw) in enumerate(zip(manifest['requests'],manifest['slides'],evidence['native']['measurements'])):
        element = slide['elements'][0]; rid=q['id']
        require(element['measurement_id'] == rid and element['name'] == raw['shape_name'] and raw['slide_index'] == index+1, 'Focused order mismatch')
        for k,rk in [('x','left'),('y','top'),('width','width'),('height','height')]:
            require(abs(element['frame'][k]-raw['shape_frame'][rk]) < .0001, 'Raw coordinate scale mismatch')
        source='\n'.join(''.join(r['text'] for r in p['runs']) for p in q['paragraphs'])
        require(source == raw['text'].replace('\r','\n'), 'Source text mismatch')
        native_lines=lines(raw); current=go['requests'][rid]; visible=[l for l in current['lines'] if l['text'].strip()]
        match=[native_comparison.normalized(l['text']) for l in visible] == [native_comparison.normalized(l['text']) for l in native_lines]
        actual=evidence['measurements']['by_request_id'][rid]; predicted=go['measurements']['by_request_id'][rid]
        row=dict(case_id=q['slide_id'],request_id=rid,category='boundary' if '-boundary-' in q['slide_id'] else 'spacing',
            line_breaks_match=match,native=actual,go=predicted,native_lines=native_lines,
            delta_pt={k:predicted.get(k,0)-actual.get(k,0) for k in native_comparison.METRICS},
            height_warnings=current.get('height_warnings',[]),boundary_warnings=current.get('boundary_warnings',[]))
        if match:row['line_top_delta_pt']=[l['y_pt']-n['y_pt'] for l,n in zip(visible,native_lines)]
        rows.append(row)
    example=data/'example-v4'
    final, fm, fg = [load(example/p) for p in ('native-verification.json','manifest.json','go-layout.json')]
    verify_artifacts(example, final, fm); native_comparison.check_fonts(final,fg)
    require(not fg['powerpoint_verified'] and fm['go_layout'] == fg, 'Example prediction provenance changed')
    require(len(final['native']['measurements']) == sum(len(s['elements']) for s in fm['slides']), 'Example shape count mismatch')
    require(len(final['measurements']['by_request_id']) == len(fg['requests']), 'Example text count mismatch')
    # The successful CLI verify already checked native styles, paragraphs,
    # frames and planner fit. Independently audit the saved safe-zone bounds.
    example_rows=[]
    for slide in fm['slides']:
        for element in slide['elements']:
            if element['kind']!='text':continue
            observed=[r for r in final['native']['measurements'] if r['slide_index']==fm['slides'].index(slide)+1 and r['shape_name']==element['name']]
            require(len(observed)==1, 'Example shape missing/duplicated')
            raw=observed[0];frame=element['frame'];bounds=raw['text_bounds'];ix=element.get('inset_x',0);iy=element.get('inset_y',0)
            fits=(bounds['left']>=frame['x']+ix-.15 and bounds['top']>=frame['y']+iy-.15 and bounds['left']+bounds['width']<=frame['x']+frame['width']-ix+.15 and bounds['top']+bounds['height']<=frame['y']+frame['height']-iy+.15)
            require(fits, 'Verified example overflows: '+element['name'])
            rid=element['measurement_id'];nl=lines(raw);gl=[l for l in fg['requests'][rid]['lines'] if l['text'].strip()]
            example_rows.append(dict(slide_id=slide['id'],shape_name=element['name'],request_id=rid,native_safe_zone_fits=True,
                line_breaks_match=[native_comparison.normalized(l['text']) for l in gl]==[native_comparison.normalized(l['text']) for l in nl]))
    expanded=[r for r in rows if r['height_warnings']]
    report=dict(schema='pptxgengo.focused-font-analysis.v1',engine=go['engine'],powerpoint_verified=False,
        native_capture_completed=True,font_hashes_styles_and_axes_match=True,raw_coordinate_scale=1,
        capture_receipt_sha256=sha(data/'capture-receipt.json'),native_sha256=sha(data/'native.json'),
        theme_body_font=theme_font(data/manifest['deck_file']),case_count=len(rows),line_break_match_count=sum(r['line_breaks_match'] for r in rows),
        expanded_spacing_case_count=len(expanded),expanded_spacing_height=stats([r['delta_pt']['rendered_height_pt'] for r in expanded]),
        matching_wrap_line_tops=stats([v for r in rows for v in r.get('line_top_delta_pt',[])]),
        remaining_wrap_differences=[r['case_id'] for r in rows if not r['line_breaks_match']],rows=rows,
        example=dict(engine=fg['engine'],native_final_verification_passed=True,verified_deck_sha256=final['deck_sha256'],
            verified_manifest_sha256=final['manifest_sha256'],evidence_sha256=sha(example/'native-verification.json'),
            measured_at=final['measured_at'],slide_count=len(fm['slides']),shape_count=len(final['native']['measurements']),text_shape_count=len(example_rows),
            font_families=sorted({f['family'] for f in final['environment']['fonts']}),all_text_safe_zones_fit=True,
            line_break_match_count=sum(r['line_breaks_match'] for r in example_rows),rows=example_rows,
            scope='Successful native CLI verification of this exact v4 deck; no visual design review or arbitrary-deck qualification'),
        limitations=['Expanded-spacing terminal height remains conservative; source paragraph-end markers omit explicit font families.',
            'Theme inheritance is a candidate explanation for differing terminal extents between packets, not established causality.',
            'Boundary observations do not establish a general shaping/kerning correction.'])
    with args.out.open('x') as f:json.dump(report,f,indent=2,ensure_ascii=False,allow_nan=False);f.write('\n')
    print(args.out)
    print('Focused line breaks:',report['line_break_match_count'],'/',len(rows),'expanded spacing:',report['expanded_spacing_height'])
    print('Native example verified:',report['example']['slide_count'],'slides,',len(example_rows),'text shapes; all native safe zones fit.')


if __name__=='__main__':main()
