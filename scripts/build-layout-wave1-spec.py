#!/usr/bin/env python3
"""Instantiate three named proposal layouts from the accepted dense-copy fixture.

The CLI owns measured block/row geometry. This file specifies component tracks,
styles, slots and content; it does not position each cell or each bullet.
"""
from pathlib import Path
import copy,json,hashlib
R=Path(__file__).resolve().parents[1]
T=json.loads((R/'library/layout-components/tokens.json').read_text())
C=T['colors']; P=T['spacing_pt']; F=T['typography']
source=json.loads((R/'library/showcase/dense-deck.json').read_text())
def rect(x,y,w,h):return dict(x=x,y=y,width=w,height=h)
def pad(t=0,r=0,b=0,l=0):return dict(top=t,right=r,bottom=b,left=l)
def block(id,text,size=10,bold=False,fg=None,minh=0,gap=0,marker='',align='left'):
 d=dict(id=id,text=text,font_face=F['family'],font_size_pt=size,bold=bold,foreground=fg or C['ink'],min_height_pt=minh,gap_before_pt=gap,align=align)
 if marker:d.update(marker=marker,marker_width_pt=P['bullet_indent'])
 return d
def track(h):return dict(fixed_pt=h)
def cell(id,row,col,blocks,padding=None,bg=''):
 d=dict(id=id,row=row,column=col,blocks=blocks,padding=padding or pad())
 if bg:d['background']=bg
 return d
def grid(id,bounds,widths,rows,cells,gap=0,rowgap=0,bg='',parent=''):
 d=dict(id=id,bounds=bounds,columns=[track(w) for w in widths],rows=rows,cells=cells,column_gap_pt=gap,row_gap_pt=rowgap,padding=pad(),layer=10)
 if bg:d['surface']=bg
 if parent:d['parent_id']=parent
 return d
def content(s):return {c['id']:c for c in s['canvas']}
def text(q,key):return q[key]['text']
def inherited(s,keep):
 s=copy.deepcopy(s);s['canvas']=[c for c in s['canvas'] if c['id'] in keep or c['id'].startswith('footer') or c['id']=='page-number'];s['layouts']=[];return s
# The source spec uses footer IDs below; preserve the entire shared footer by region.
def base(index,keep):
 src=source['slides'][index];out=inherited(src,keep)
 out['canvas']=[copy.deepcopy(c) for c in src['canvas'] if c['id'] in keep or c['bounds']['y']>=504]
 return out,content(src)

s,q=base(0,{'kicker','engagement','extension'})
cells=[]
for i in range(5):
 cells.append(cell(f'phase-{i}-number',0,i,[block('number',text(q,f'phase-{i}'),25,True,C['white'],32)],pad(5,8,5,8),C['muted'] if i==4 else C['ink']))
 bs=[block('heading',text(q,f'heading-{i}'),12,True,minh=34,align='center'),block('timing',text(q,f'dates-{i}'),9.5,fg=C['muted'],minh=17,gap=2,align='center'),block('activities-label','Activities',11,True,minh=18,gap=7)]
 hs=[38,50,38] if i==1 else [38,38,50]
 for j,h in enumerate(hs):bs.append(block(f'activity-{j}',text(q,f'activities-{i}-text-{j}'),minh=h,gap=6 if j==0 else 3,marker='•'))
 cells.append(cell(f'phase-{i}-activities',1,i,bs,pad(9,8,2,8),C['surface']))
 cells.append(cell(f'phase-{i}-output',2,i,[block('label','Evidence / output',11,True,C['ink'],18),block('evidence',text(q,f'output-{i}'),minh=49,gap=4)],pad(10,8,7,8),C['rule']))
s['layouts']=[grid('delivery-phases',rect(36,139,888,357),[168]*5,[track(42),track(227),track(88)],cells,gap=P['phase_gutter'])]
controls=[s]

s,q=base(1,{'deliverable-title','deliverable-caption','scope-heading','scope-text'})
sidebar_specs=[('phase-kicker',18,0),('phase-number',22,11),('phase-title',65,11),('phase-dates',25,14),('objective-head',23,27),('objective',133,11),('exit-head',23,12)]
bs=[]
for key,h,gap in sidebar_specs:
 c=q[key];bs.append(block(key,c['text'],c['font_size_pt'],c['bold'],C['white'],h,gap))
for j,h in enumerate([16,16,28]):bs.append(block(f'exit-{j}',text(q,f'exit-text-{j}'),10.5,fg=C['white'],minh=h,gap=9 if j==0 else 3,marker='•'))
s['layouts'].append(grid('phase-sidebar',rect(0,0,252,504),[252],[track(504)],[cell('content',0,0,bs,pad(28,18,6,36),C['ink'])]))
s['layouts'].append(grid('workstream-headers',rect(270,28,426,30),[272,152],[track(30)],[cell('activities',0,0,[block('heading','Activities',12,True,minh=16)],pad(7,8,7,8),C['surface']),cell('products',0,1,[block('heading','Work products',12,True,minh=16)],pad(7,8,7,8),C['surface'])],gap=2))
cells=[]
for i in range(4):
 acts=[block('heading',text(q,f'activity-label-{i}'),10.5,True,minh=16)]
 products=[block('heading',text(q,f'product-label-{i}'),10,True,minh=27)]
 for j in range(3):acts.append(block(f'activity-{j}',text(q,f'activity-row-{i}-text-{j}'),minh=24,gap=4 if j==0 else 3,marker='•'))
 for j in range(2):products.append(block(f'product-{j}',text(q,f'product-row-{i}-text-{j}'),minh=32,gap=3,marker='•'))
 cells.extend([cell(f'workstream-{i}-activities',i,0,acts,pad(0,9,8 if i<3 else 6,8)),cell(f'workstream-{i}-products',i,1,products,pad(0,8,9 if i<3 else 7,8))])
work=grid('workstream-detail',rect(270,67,426,422),[272,152],[dict(min_pt=106,max_pt=120)]*3+[dict(min_pt=104,max_pt=120)],cells,gap=2)
work['row_rule']=dict(color=C['rule'],width_pt=.65,offset_pt=-6)
s['layouts'].append(work)
s['canvas'].append(copy.deepcopy(q['vertical-divider']))
s['canvas'][-1]['allow_overlap']=['workstream-detail/row-rule/1','workstream-detail/row-rule/2','workstream-detail/row-rule/3']
s['layouts'].append(grid('acceptance',rect(716,99,208,156),[208],[track(48)]*3,[cell(f'view-{i}',i,0,[block('description',text(q,f'acceptance-{i}'),minh=38)],pad(5,8,5,8),C['surface']) for i in range(3)],rowgap=6))
bs=[block('heading',text(q,'assumptions-heading'),12,True,C['emphasis'],21)]
for j in range(3):bs.append(block(f'assumption-{j}',text(q,f'assumptions-text-{j}'),minh=37,gap=7 if j==0 else 3,marker='•'))
s['layouts'].append(grid('scope-assumptions',rect(716,265,208,160),[208],[track(160)],[cell('content',0,0,bs,pad(9,8,6,8),C['surface'])]))
controls.append(s)

s,q=base(2,{'kicker','matrix-takeaway'})
widths=[134,234,308,196]
heads=[cell(f'column-{i}',0,i,[block('heading',text(q,f'matrix-head-{i}'),11,True,C['white'] if i==0 else C['ink'],16)],pad(6,7,6,7),C['ink'] if i==0 else C['surface']) for i in range(4)]
h=grid('workflow-headers',rect(36,127,888,28),widths,[track(28)],heads);h['column_gaps_pt']=[4,6,6];s['layouts'].append(h)
cells=[]
for i in range(5):
 for j in range(4):cells.append(cell(f'workflow-{i}-column-{j}',i,j,[block('content',text(q,f'matrix-{i}-{j}'),10.5,j==0,C['white'] if j==0 else C['ink'],minh=0)],pad(P['matrix_vertical'],8,P['matrix_vertical'],8),[C['ink'],C['surface'],C['proposed'],C['measure']][j]))
g=grid('workflow-rows',rect(36,161,888,318),widths,[dict(min_pt=62,max_pt=92)]*5,cells,rowgap=2);g['column_gaps_pt']=[4,6,6];s['layouts'].append(g)
controls.append(s)

# Changed-content fixtures. IDs differ; the native cache must bind by complete
# rendering contract and retain the original observation provenance.
variants=[]
v=copy.deepcopy(controls[2]);v['id']='workflow-four-long';v['layouts'][-1]['rows']=v['layouts'][-1]['rows'][:4];v['layouts'][-1]['cells']=[c for c in v['layouts'][-1]['cells'] if c['row']<4]
v['layouts'][-1]['cells'][1]['blocks'][0]['text']+=' Teams must also confirm ownership when a request spans multiple service lines and approval policies.'
variants.append(v)
v=copy.deepcopy(controls[2]);v['id']='workflow-six';g=v['layouts'][-1];g['rows']=[dict(min_pt=51,max_pt=70)]*6
rows=[('01  Intake','Confirm required case information and accountable owner.','Validate once; route incomplete cases to a named queue.','Completeness; intake time'),('02  Preparation','Collect evidence across service and data boundaries.','Show source, freshness and unresolved evidence gaps.','Missing evidence; preparation time'),('03  Decision','Make approval authority and dependencies visible.','Record rationale and exceptions with each decision.','Approval time; reopened cases'),('04  Fulfilment','Coordinate changes across downstream services.','Publish controlled events with recovery procedures.','Failed handoffs; recovery time'),('05  Monitoring','Observe results after each release.','Track service, control and customer outcomes together.','Service incidents; adoption'),('06  Improvement','Assign follow-up actions and prioritize next changes.','Review measured results before funding the next wave.','Change lead time; support demand')]
g['cells']=[]
for i,row in enumerate(rows):
 for j,t in enumerate(row):g['cells'].append(cell(f'workflow-{i}-column-{j}',i,j,[block('content',t,10.5,j==0,C['white'] if j==0 else C['ink'])],pad(5.5,8,5.5,8),[C['ink'],C['surface'],C['proposed'],C['measure']][j]))
variants.append(v)
v=copy.deepcopy(controls[0]);v['id']='phases-parent-move';v['layouts'][0]['parent_id']='phase-parent';v['layouts'].insert(0,grid('phase-parent',rect(0,4,960,500),[960],[track(500)],[]));variants.append(v)
for filename,slides in [('controls.json',controls),('variants.json',variants)]:
 (R/'library/layout-components'/filename).write_text(json.dumps(dict(schema=source['schema'],slides=slides),indent=2)+'\n')
negative=copy.deepcopy(controls[2]);negative['id']='workflow-overflow';negative['layouts'][-1]['cells'][1]['blocks'][0]['text']+=' '+('Each owner must review and resolve evidence gaps before accepting the case. '*12).strip();negative['layouts'][-1]['cells'][5]['blocks'][0]['text']+=' '+('Every exception requires an accountable approver and a documented recovery path. '*12).strip()
(R/'library/layout-components/overflow.json').write_text(json.dumps(dict(schema=source['schema'],slides=[negative]),indent=2)+'\n')
review=copy.deepcopy(controls+variants)
for i,slide in enumerate(review,1):
 for item in slide['canvas']:
  if item['id']=='page':item['text']=str(i)
 slide['notes']+=' Layout fixture: '+('control' if i<=3 else ['four rows with longer content','six rows with new content','nested parent moved four points'][i-4])+'.'
(R/'library/layout-components/review.json').write_text(json.dumps(dict(schema=source['schema'],slides=review),indent=2)+'\n')
local=R/'samples/layout-wave1';local.mkdir(parents=True,exist_ok=True)
(local/'stress-probes-spec.json').write_text(json.dumps(dict(schema=source['schema'],slides=review+[negative]),indent=2)+'\n')
print('wrote controls, variants, review and explicit overflow fixture')
