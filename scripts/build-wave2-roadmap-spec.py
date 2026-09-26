#!/usr/bin/env python3
"""Author bounded source-derived roadmap shape and interval fixtures."""
from pathlib import Path
import copy,json
R=Path(__file__).resolve().parents[1];D=R/'library/visual-components';D.mkdir(exist_ok=True)
c=json.loads((R/'library/component-contracts/wave2-roadmap-contract.json').read_text())['observed']
base=json.loads((R/'library/layout-components/controls.json').read_text())['slides'][0]
NAVY='#070154';SLATE='#50658E';PALE='#CED7E6';LIGHT='#E8EEF8';BLUE='#0047FF';MAGENTA='#F900D3'
def rect(v):return dict(zip(['x','y','width','height'],v))
def text(id,words,bounds,size=11,bold=False,color=NAVY,align='left',italic=False,peers=[]):
 d=dict(id=id,kind='text',bounds=rect(bounds),align=align,valign='middle',inset_x=0,inset_y=0,layer=30,allow_overlap=peers)
 if color=='#FFFFFF':d['contrast_background']=NAVY
 if italic:d['paragraphs']=[dict(id='p1',align=align,runs=[dict(id='r1',text=words,font_face='Arial',font_size_pt=size,bold=bold,italic=True,foreground=color)])]
 else:d.update(text=words,font_face='Arial',font_size_pt=size,bold=bold,foreground=color)
 return d
def sourcebox(item,group):
 off=group['off_emu'];ext=group['ext_emu'];co=group['chOff_emu'];ce=group['chExt_emu'];lo=item['local_off_emu'];le=item['ext_emu']
 return [(off[0]+(lo[0]-co[0])*ext[0]/ce[0])/12700,(off[1]+(lo[1]-co[1])*ext[1]/ce[1])/12700,le[0]*ext[0]/ce[0]/12700,le[1]*ext[1]/ce[1]/12700]
def shape(id,b,preset='homePlate',fill=NAVY,pattern=False,layer=10,peers=[]):
 d=dict(id=id,kind='shape',bounds=rect(b),preset=preset,layer=layer,allow_overlap=peers)
 if preset=='homePlate':d['adjustments']={'adj':39542}
 if pattern:d['pattern']=dict(preset='wdUpDiag',foreground=fill,background='#FFFFFF')
 else:d['background']=fill
 return d
def make(changed=False):
 s=copy.deepcopy(base);s.update(id='roadmap-interval-variant' if changed else 'roadmap-shape-control',title='A sequenced release plan makes dependencies and transition decisions explicit' if changed else 'Editable roadmap ribbons and patterned extensions',layouts=[])
 s['canvas']=[x for x in s['canvas'] if x['id'] in ['kicker','footer-band','wm-logo','footer-copy','page']]
 next(x for x in s['canvas'] if x['id']=='kicker')['text']='ILLUSTRATIVE DELIVERY ROADMAP' if changed else 'SOURCE GEOMETRY CONTROL — UHG 38'
 s['notes']='ILLUSTRATIVE ROADMAP. Source UHG38 preset shapes, hatch pattern and arrow adjustment are retained. The month header is composed of editable native rectangles/text, not a PowerPoint table. Source groups are flattened; label geometry is explicitly authored and is not claimed pixel-identical. Milestone/interval changes are declared in this fixture, not inferred semantic dates.'
 grid=c['month_grid'];x,y,_,_=[v/12700 for v in grid['frame_off_ext_emu']];cw=grid['column_widths_emu'][0]/12700
 for i,label in enumerate(grid['header_labels']):
  bg=LIGHT if i>=9 else NAVY;id=f'month-{i+1}'
  s['canvas'].append(dict(id=id,kind='surface',bounds=rect([x+i*cw,y,cw,36]),background=bg))
  s['canvas'].append(text(id+'-label',label.replace(' ','\n'),[x+i*cw+4,y+3,cw-8,30],12,True,NAVY if i>=9 else '#FFFFFF','center',peers=[id]))
 labels=['Phase 01: Fabric Platform Foundation','Phase 02: Scalability & Early Adoption\n3 Early Adopters','Phase 03: General Availability & Optimization']
 if changed:labels=['Phase 01: Prove the release model','Phase 02: Expand adoption\nPriority journeys','Phase 03: Transfer ownership']
 for i,item in enumerate(c['phase_ribbons']['items']):
  b=sourcebox(item['base'],item['group']);t=sourcebox(item['tail'],item['group']);color=[PALE,SLATE,NAVY][i]
  if changed:
   # Duration variant: grow the first base and tail by half a month; move the
   # second interval one quarter-month without changing its shape dimensions.
   if i==0:b[2]+=cw*.5;t[0]+=cw*.5
   if i==1:b[0]+=cw*.25;t[0]+=cw*.25
  bid=f'phase-{i+1}-base';tid=f'phase-{i+1}-tail'
  s['canvas'] += [shape(tid,t,fill=color,pattern=True,layer=10),shape(bid,b,fill=color,layer=11,peers=[tid])]
  # Keep labels within the rectangular portion of the homePlate and away from tip.
  inset=4;tip=b[3]*.39542
  lbl=text(f'phase-{i+1}-label',labels[i],[b[0]+inset,b[1]+4,b[2]-tip-inset*2,b[3]-8],11,True,NAVY if i==0 else '#FFFFFF','center',peers=[bid,tid])
  lbl['contrast_background']=color
  if i==1:
   words=labels[i].split('\n');lbl.pop('text');lbl.pop('font_face');lbl.pop('font_size_pt');lbl.pop('bold');lbl.pop('foreground')
   lbl['paragraphs']=[dict(id=f'p{j}',align='center',runs=[dict(id='r',text=v,font_face='Arial',font_size_pt=11,bold=True,italic=j==1,foreground='#FFFFFF')]) for j,v in enumerate(words)]
  s['canvas'].append(lbl)
 milestone_labels=['Architecture Approved and Foundation Ready','Targeted Production Release for Early Adopters','GA Launch','UHG Ownership']
 if changed:milestone_labels=['Release model accepted','Priority journeys ready for release','Operating controls accepted','Client ownership confirmed']
 for i,m in enumerate(c['milestones']['items']):
  b=[v/12700 for v in m['star']['bounds_emu']];lb=[v/12700 for v in m['label']['bounds_emu']]
  if changed and i==0:b[0]+=cw*.5;lb[0]+=cw*.5
  sid=f'milestone-{i+1}';s['canvas'] += [shape(sid,b,preset='star5',fill=MAGENTA,layer=20,peers={0:['phase-1-tail'],2:['phase-2-tail'],3:['phase-3-tail']}.get(i,[])),text(sid+'-caption',milestone_labels[i],[lb[0],lb[1]-1,lb[2],max(lb[3],15)],10,italic=True)]
 # Two lower bars retain source geometry but use explicit illustrative copy.
 for id,b,fill,words in [('onboarding',[9.053*72,4.888*72,3.78*72,.55*72],BLUE,'Onboard subsequent journeys' if changed else 'GA Onboarding'),('ongoing',[.503*72,5.68*72,8.528*72,.55*72],LIGHT,'Ongoing: Delivery governance, measurement and service ownership' if changed else 'Ongoing: Project & Product Management from Enablement Team')]:
  s['canvas'] += [shape(id,b,fill=fill),text(id+'-label',words,[b[0]+5,b[1]+4,b[2]-b[3]*.39542-10,b[3]-8],11,True,'#FFFFFF' if fill==BLUE else NAVY,'center',peers=[id])]
 for element in s['canvas']:
  if element['id']=='onboarding-label':element['contrast_background']=BLUE
 s['canvas'] += [shape('hatch-key',[366,457,23,16],fill=SLATE,pattern=True),text('hatch-description','Potential extension; timing depends on acceptance and readiness',[394,455,465,20],10)]
 return s
slides=[make(),make(True)]
for i,s in enumerate(slides,1):next(c for c in s['canvas'] if c['id']=='page')['text']=str(i)
(D/'roadmap-review.json').write_text(json.dumps(dict(schema='pptxgengo.compose-spec.v1',slides=slides),indent=2)+'\n')
print(D/'roadmap-review.json')
