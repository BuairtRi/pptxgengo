#!/usr/bin/env python3
"""Compare native text geometry against the earlier dense proposal controls."""
from pathlib import Path
import collections,json
R=Path(__file__).resolve().parents[1]
B=R/'samples/component-adaptation/dynamic-pods'
def read(p):return json.loads(p.read_text())
def load(name):
 m=read(B/name/'manifest.json');v=read(B/(name+'-verification.json'))
 rows={(r['slide_index'],r['shape_name']):r for r in v['native']['measurements']}
 result=[]
 for i,s in enumerate(m['slides'][:3],1):
  groups=collections.defaultdict(list)
  for e in s['elements']:
   if e['kind']!='text':continue
   key=(e['text'],e['font_face'],e['font_size'],e.get('bold',False))
   groups[key].append((e,rows[i,e['name']]))
  for g in groups.values():g.sort(key=lambda x:(x[0]['frame']['x'],x[0]['frame']['y']))
  result.append(groups)
 return result
old,new=load('proposal-complexity-v2'),load('wave1-layout-review')
report=[]
for i,(a,b) in enumerate(zip(old,new),1):
 assert a.keys()==b.keys()
 changes=[];count=0;maximum=0
 for k,aa in a.items():
  bb=b[k];assert len(aa)==len(bb)
  for (oe,on),(ne,nn) in zip(aa,bb):
   count+=1
   delta={d:nn['text_bounds'][d]-on['text_bounds'][d] for d in ['left','top','width','height']}
   maximum=max(maximum,*[abs(d) for d in delta.values()])
   if max(abs(d) for d in delta.values())>.12:changes.append(dict(text=k[0],delta_pt=delta))
 report.append(dict(slide=i,text_blocks=count,max_glyph_delta_pt=maximum,geometry_changes=changes))
p=R/'samples/layout-wave1/native-copy-comparison.json';p.write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))
