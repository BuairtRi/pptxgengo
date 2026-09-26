#!/usr/bin/env python3
"""Record the accepted Wave 1 fixture from local native and visual evidence."""
from pathlib import Path
import collections,hashlib,json
R=Path(__file__).resolve().parents[1]
B=R/'samples/component-adaptation/dynamic-pods'
Q=R/'samples/layout-wave1'
def read(p):return json.loads(Path(p).read_text())
def rec(p):
 p=Path(p);data=p.read_bytes();return dict(path=str(p.relative_to(R)),sha256=hashlib.sha256(data).hexdigest(),size_bytes=len(data))
def write(p,obj):Path(p).write_text(json.dumps(obj,indent=2)+'\n')
spec=R/'library/layout-components/review.json';deck=B/'wave1-layout-review/wave1-layout-review.pptx'
m=read(B/'wave1-layout-review/manifest.json');v=read(B/'wave1-layout-review-verification.json')
assert m['spec_sha256']==v['spec_sha256']==rec(spec)['sha256']
assert m['deck_sha256']==v['deck_sha256']==rec(deck)['sha256']
fit=read(Q/'review-fit.json');assert fit['planner_passed'] and fit['overflow_count']==0 and fit['layout_failure_count']==0
bad=read(Q/'overflow-fit.json');assert not bad['planner_passed']
bad_cells={x.get('cell_id') for x in bad['layout_failures']}
assert {'workflow-0-column-1','workflow-1-column-1'}.issubset(bad_cells)
reviews=read(Q/'visual-review.json');assert [x['id'] for x in reviews]==[x['id'] for x in read(spec)['slides']]
assert all(x['reviewed'] and x['accepted'] for x in reviews)
checks=read(Q/'qa/cache-integrity-results.json');assert all(x['passed'] for x in checks)
copy=read(Q/'copy-preservation-final.json');assert all(x['text_multiset_unchanged'] for x in copy)
glyphs=read(Q/'native-copy-comparison.json');assert all(not x['geometry_changes'] for x in glyphs)
reuse=read(Q/'cache-reuse-check.json');assert reuse['changed_contracts']==1 and reuse['all_cached_status']=='all_requests_cached'
rows={(x['slide_index'],x['shape_name']):x for x in v['native']['measurements']}
max_frame=max_overflow=0
for i,slide in enumerate(m['slides'],1):
 for e in slide['elements']:
  n=rows[(i,e['name'])];f=e['frame'];nf=n['shape_frame']
  max_frame=max(max_frame,*[abs(f[a]-nf[b]) for a,b in [('x','left'),('y','top'),('width','width'),('height','height')]])
  if e['kind']=='text':
   t=n['text_bounds'];ix=e['inset_x'];iy=e['inset_y'];max_overflow=max(max_overflow,f['x']+ix-t['left'],f['y']+iy-t['top'],t['left']+t['width']-f['x']-f['width']+ix,t['top']+t['height']-f['y']-f['height']+iy)
assert max_frame<=.12 and max_overflow<=.15
report=dict(schema='pptxgengo.layout-fixture-proof.v1',status='native_fixture_verified_not_general_layout_approval',date='2026-09-26',slides=len(m['slides']),objects=len(rows),text_zones=fit['fixed_zone_count'],layout_failures=0,max_frame_delta_pt=max_frame,max_text_overflow_pt=max_overflow,environment=v['environment'],controls='Accepted dense showcase first three pages; not a new UHG pixel-reconstruction claim',intentional_changes=['Five blue Evidence / output labels changed to navy after actual-surface contrast check found 4.33:1','Native shape segmentation changes: cell backgrounds and text are separate editable objects'],copy_preservation=copy,native_control_comparison=glyphs,visual_review=reviews,cache_reuse=reuse,cache_integrity_checks=checks,overflow_cells=sorted(x for x in bad_cells if x),geometry=read(Q/'geometry-checks.json'),regression='go test ./internal/compose ./cmd/pptxcompose passed; go build ./... passed',limitations=['Uniform-style Arial blocks; rich text is Wave 2','Semantic parent-child ownership is in the spec, not native PowerPoint drag behavior','Explicit bounds/minimums; no automatic page splitting or broad layout selection','Editable native shapes, not a native table model','Fixture helper resolves named style tokens; no general runtime token registry','Cache is local and conservatively bound to compiled CLI; changed binaries require fresh observations','Known-deck layout text recovery only; formatting/layout reconciliation remains later work','General direct/row-rule/accumulated layer range validation remains incomplete; exact fixture values are verified','Row spans use final-row growth; parent-cell fill inheritance and concurrent cache import are not qualified'],inputs=[rec(spec),rec(R/'library/layout-components/controls.json'),rec(R/'library/layout-components/variants.json'),rec(R/'library/layout-components/overflow.json'),rec(R/'library/layout-components/tokens.json')],code=[rec(R/p) for p in ['cmd/pptxcompose/main.go','cmd/pptxcompose/cache.go','cmd/pptxcompose/recover.go','internal/compose/layout.go','internal/compose/canvas.go','internal/compose/planner.go','internal/compose/fit_report.go','internal/compose/layout_test.go','cmd/pptxcompose/cache_test.go','scripts/compose-environment.swift','scripts/build-layout-wave1-spec.py','scripts/report-layout-wave1.py','scripts/check-layout-wave1-geometry.py','scripts/check-layout-wave1-cache.py','scripts/check-layout-wave1-copy.py']],artifacts=[rec(p) for p in [deck,B/'wave1-layout-review.pdf',B/'wave1-layout-review/manifest.json',B/'wave1-layout-review-verification.json',B/'wave1-release-evidence.json',B/'wave1-cache-edit-evidence.json',Q/'review-fit.json',Q/'overflow-fit.json',Q/'visual-review.json',Q/'geometry-checks.json',Q/'copy-preservation-final.json',Q/'native-copy-comparison.json',Q/'cache-reuse-check.json',Q/'qa/cache-integrity-results.json']]+[rec(B/'wave1-layout-review-render'/f'slide-{i:03}.png') for i in range(1,7)])
write(R/'library/layout-components/proof.json',report)
print(json.dumps({k:report[k] for k in ['slides','objects','text_zones','layout_failures','max_frame_delta_pt','max_text_overflow_pt','cache_reuse']},indent=2))
