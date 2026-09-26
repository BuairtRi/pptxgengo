#!/usr/bin/env python3
"""Compare the local Wave 1 control and stress geometry after building."""
from pathlib import Path
import json,collections
R=Path(__file__).resolve().parents[1];B=R/'samples/component-adaptation/dynamic-pods';Q=R/'samples/layout-wave1'
m=json.loads((B/'wave1-layout-review/manifest.json').read_text());slides=m['slides'];plan=m['plan']['slides']
def canvas(i):return {x['id']:x for x in plan[i]['canvas']}
a,b,c,d=canvas(0),canvas(5),canvas(2),canvas(3)
moves=[]
for id,x in a.items():
 if id.startswith('delivery-phases/'):
  y=b[id];af,bf=x['bounds'],y['bounds'];moves.append(dict(id=id,dx=bf['x']-af['x'],dy=bf['y']-af['y'],dw=bf['width']-af['width'],dh=bf['height']-af['height']))
assert moves and all(abs(x['dx'])<1e-6 and abs(x['dy']-4)<1e-6 and abs(x['dw'])<1e-6 and abs(x['dh'])<1e-6 for x in moves)
control=c['workflow-rows/workflow-0-column-1']['bounds'];long=d['workflow-rows/workflow-0-column-1']['bounds']
assert long['height']>control['height']
row_heights=[d[f'workflow-rows/workflow-0-column-{j}']['bounds']['height'] for j in range(4)];assert max(row_heights)-min(row_heights)<1e-6
six=canvas(4);sixrows=[six[f'workflow-rows/workflow-{i}-column-0']['bounds'] for i in range(6)];assert len(sixrows)==6 and sixrows[-1]['y']+sixrows[-1]['height']<=479.0001
geometry=dict(parent_translation=dict(objects=len(moves),delta_y_pt=4,all_widths_and_heights_unchanged=True),longer_content=dict(control_row_height_pt=control['height'],variant_row_height_pt=long['height'],all_four_columns_share_height=True),six_rows=dict(row_count=6,bottom_pt=sixrows[-1]['y']+sixrows[-1]['height'],container_bottom_pt=479))
(Q/'geometry-checks.json').write_text(json.dumps(geometry,indent=2)+'\n')
old=json.loads((B/'proposal-complexity-v2/manifest.json').read_text())['slides'][:3]
copy=[]
for a,b in zip(old,slides[:3]):
 aa=collections.Counter(x['text'] for x in a['elements'] if x['kind']=='text');bb=collections.Counter(x['text'] for x in b['elements'] if x['kind']=='text');copy.append(dict(id=b['id'],text_multiset_unchanged=aa==bb,text_shapes=sum(bb.values())))
assert all(x['text_multiset_unchanged'] for x in copy)
(Q/'copy-preservation-final.json').write_text(json.dumps(copy,indent=2)+'\n');print(json.dumps(geometry,indent=2))
