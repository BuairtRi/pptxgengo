#!/usr/bin/env python3
"""Replay candidate predictions against preserved v1 evidence (development fit,
not independent v2 validation). No captured data or packet hashes are rewritten.
"""
import argparse,hashlib,json
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--baseline',required=True,type=Path)
p.add_argument('--capture-dir',required=True,type=Path)
p.add_argument('--candidate',required=True,type=Path)
p.add_argument('--out',required=True,type=Path)
a=p.parse_args()
load=lambda f:json.loads(f.read_text())
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
v1,v2=load(a.baseline/'probes.json'),load(a.candidate/'probes.json')
receipt=load(a.capture_dir/'capture-receipt.json');raw=load(a.capture_dir/'native.json')
assert v1['engine']=='wmds-go-foundation.v1' and v2['engine']=='wmds-go-foundation.v2'
assert sha(a.baseline/'typography-probes.pptx')==v1['deck_sha256']==receipt['deck_sha256']
assert sha(a.baseline/'probes.json')==receipt['manifest_sha256']
assert sha(a.capture_dir/'native.json')==receipt['native_sha256']
assert sha(a.candidate/'typography-probes.pptx')==v2['deck_sha256']
assert not a.out.exists()
native={(q['slide_index'],q['shape_name']):q for q in raw['measurements']}
candidates={(q['slide'],q['id']):q for q in v2['probes']}
rows=[]
for q in v1['probes']:
 c=candidates[q['slide'],q['id']];old,new=q['record'],c['record'];n=native[q['slide'],q['id']]
 assert old['rect']==new['rect'] and old['layout']['displayed']==new['layout']['displayed'] and old['layout']['style']==new['layout']['style'] and old['layout']['font']==new['layout']['font']
 groups=[]
 for char in n['characters']:
  y=char['bounds']['top'];g=next((g for g in groups if abs(g['top']-y)<=.1),None)
  if g is None:g=dict(top=y,text='');groups.append(g)
  g['text']+=char['text']
 observed=[g['text'].strip() for g in sorted(groups,key=lambda g:g['top']) if g['text'].strip()]
 predicted=[l['text'].strip() for l in new['layout']['lines'] if l['text'].strip()]
 content=n['text'].replace('\r\n','\n').replace('\r','\n')==new['layout']['displayed']
 paragraphs=len(n['paragraphs'])==len(new['layout']['displayed'].split('\n'))
 delta=n['text_bounds']['height']-new['layout']['estimated_occupied_height_pt']
 rows.append(dict(id=q['id'],predicted_lines=predicted,native_lines=observed,wrap_match=predicted==observed,
  historical_content_match=content,historical_paragraph_count_match=paragraphs,
  height_comparable=content and paragraphs and predicted==observed,height_delta_pt=delta))
valid=[r for r in rows if r['height_comparable']]
result=dict(schema='pptxgengo.wmds-development-replay.v1',status='development_fit_not_independent_validation',
 provenance=dict(baseline_manifest_sha256=sha(a.baseline/'probes.json'),candidate_manifest_sha256=sha(a.candidate/'probes.json'),
 native_sha256=sha(a.capture_dir/'native.json'),receipt_sha256=sha(a.capture_dir/'capture-receipt.json'),analyzer_sha256=sha(Path(__file__))),
 summary=dict(probe_count=len(rows),wrap_matches=sum(r['wrap_match'] for r in rows),comparable_height_count=len(valid),
 height_within_half_point=sum(abs(r['height_delta_pt'])<=.5 for r in valid),max_comparable_height_error_pt=max(abs(r['height_delta_pt']) for r in valid)),
 limits=['Candidate vertical anchors were fitted from this v1 evidence.',
 'Historical trailing-break content defect is retained and excluded from comparable heights.',
 'The candidate deck still requires independent native character capture; file identity is not established by replay.'],probes=rows)
a.out.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result['summary'],indent=2))
