#!/usr/bin/env python3
"""Dense editable diagram fixtures; content is explicitly illustrative."""
import copy,json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];OUT=ROOT/'library/diagram-components'
NAVY='#070154';BLUE='#0047FF';PALE='#E8EEF8';WHITE='#FFFFFF'
def rect(x,y,w,h):return dict(x=x,y=y,width=w,height=h)
def text(id,body,bounds,size=11,bold=False,color=NAVY):return dict(id=id,kind='text',text=body,bounds=bounds,font_face='Arial',font_size_pt=size,bold=bold,foreground=color,align='left',valign='top')
def block(id,body,size=11,bold=False,color=NAVY,gap=0):return dict(id=id,text=body,font_face='Arial',font_size_pt=size,bold=bold,foreground=color,gap_before_pt=gap)
def main():
 dense=json.loads((ROOT/'library/showcase/dense-deck.json').read_text())
 team=copy.deepcopy(next(s for s in dense['slides'] if s['id']=='delivery-organization'))
 team['id']='dense-delivery-relationships'
 team['notes']+=' Wave3: preserves 21 illustrative role tiles and three dynamic pods. Three counterpart rules become typed advisory/dependency/annotation relations. This is a changed synthetic composition, not UHG43 reconstruction.'
 team['canvas']=[c for c in team['canvas'] if not c['id'].startswith('counterpart-')]
 for i,(a,b,rel) in enumerate([('sponsor','client-sponsor','advisory'),('lead','client-lead','dependency'),('architect','client-owner','annotation')]):
  team['connections'].append(dict(id='counterpart-'+str(i),**{'from':a,'to':b},relationship=rel,from_anchor='right',to_anchor='left',color=NAVY,width_pt=1,clearance_pt=4))
 for c in team['canvas']:
  if c['id']=='staffing-head':c['text']='RELATIONSHIPS AND STAFFING'
  if c['id']=='staffing-body':c['text']='Solid: reporting • Arrow: approval\nDashed: advisory • Dotted: annotation\nStaffing colors follow the legend.'
  if c['id']=='page':c['text']='2'
 footer=copy.deepcopy([c for c in team['canvas'] if c['id'] in ['footer-band','wm-logo','footer-copy','page']])
 for c in footer:
  if c['id']=='page':c['text']='1'
 arch=dict(id='dense-layered-architecture',title='Separate shared platform controls from domain delivery\nso each release has clear ownership and observable evidence',width_pt=960,height_pt=540,title_bounds=rect(36,44,888,60),title_font_face='Arial',title_font_size_pt=23,title_bold=True,title_foreground=NAVY,role='ILLUSTRATIVE TARGET ARCHITECTURE',takeaway='Domain teams own workflow outcomes while shared services provide consistent controls and evidence.',notes='Fictional case-flow modernization architecture. Editable nested native shapes and measured text. Inspired by UHG14 hierarchy only; does not claim to extract the baked raster diagram.',pods=[],canvas=footer+[text('kicker','ILLUSTRATIVE • FICTIONAL CASE-FLOW MODERNIZATION',rect(36,22,888,16),10,True,BLUE)],layouts=[],connections=[])
 # The parent owns every architecture cell and governance band; translated variants
 # change only the parent bounds, never individual child coordinates.
 arch['layouts'].append(dict(id='platform',bounds=rect(226,117,698,373),padding=dict(top=8,right=8,bottom=8,left=8),columns=[dict(weight=1)],rows=[dict(fixed_pt=20)],surface=PALE,layer=1,cells=[]))
 for identity,y,body in [('governance-band',0,'GOVERNANCE  •  decision rights  •  exception review  •  release evidence'),('security-band',325,'SHARED CONTROLS  •  identity and access  •  auditability  •  data protection')]:
  arch['layouts'].append(dict(id=identity,parent_id='platform',bounds=rect(0,y,682,32),padding=dict(top=7,right=8,bottom=5,left=8),columns=[dict(weight=1)],rows=[dict(fixed_pt=20)],surface=NAVY,layer=1,cells=[dict(id='statement',row=0,column=0,padding=dict(top=0,right=0,bottom=0,left=0),blocks=[block('label',body,11,True,WHITE)])]))
 levels=[('experience',46,'01  WORKFLOW EXPERIENCE',[
 ('intake','Case intake','Capture request context, confirm completeness and show the requester what happens next.'),
 ('workbench','Decision workbench','Bring evidence and policy guidance together; record the decision and unresolved questions.'),
 ('communications','Service communications','Make status, owner and next action visible to requesters and downstream service teams.')]),
 ('services',142,'02  DOMAIN SERVICES',[
 ('orchestration','Case orchestration','Coordinate handoffs and exception paths with explicit service ownership and recovery steps.'),
 ('rules','Decision services','Version decision rules, expose review points and retain the rationale for each outcome.'),
 ('integration','Managed integration','Publish interface contracts and handle retries, failures and dependencies consistently.')]),
 ('foundation',238,'03  PLATFORM FOUNDATION',[
 ('information','Governed information','Define trusted records, lineage and access boundaries; validate quality at critical handoffs.'),
 ('engineering','Engineering enablement','Automate deployment checks and release evidence; make changes reproducible and reversible.'),
 ('operations','Platform operations','Monitor workflow health and service reliability; connect alerts to accountable responders.')])]
 for ident,y,label,nodes in levels:
  arch['canvas'].append(text(ident+'-guide',label,rect(36,125+y,174,32),12,True,BLUE))
  explanation={'experience':'Make work and decisions legible to the people who use the service.','services':'Keep workflow rules and integrations modular enough to change safely.','foundation':'Provide shared capabilities that can be measured and operated consistently.'}[ident]
  arch['canvas'].append(text(ident+'-why',explanation,rect(36,164+y,174,48),11))
  cells=[]
  for col,(cid,heading,body) in enumerate(nodes):
   cells.append(dict(id=cid,row=0,column=col,padding=dict(top=7,right=8,bottom=6,left=8),background=WHITE,blocks=[block('heading',heading,12,True),block('detail',body,10.5,gap=4)]))
  arch['layouts'].append(dict(id=ident,parent_id='platform',bounds=rect(0,y,682,72),padding=dict(top=0,right=0,bottom=0,left=0),columns=[dict(weight=1)]*3,column_gap_pt=23,rows=[dict(fixed_pt=72)],layer=1,cells=cells))
 # Vertical dependencies occupy the reserved inter-layer gaps; they do not cross labels.
 for i,(upper,lower) in enumerate(zip(levels,levels[1:])):
  for col in range(3):
   a=upper[0]+'/'+upper[3][col][0];b=lower[0]+'/'+lower[3][col][0]
   arch['connections'].append(dict(id=f'dependency-{i}-{col}',**{'from':a,'to':b},relationship='dependency',from_anchor='bottom',to_anchor='top',color=NAVY,width_pt=1,clearance_pt=3,preferred_direction='vertical'))
 moved=copy.deepcopy(arch);moved['id']='dense-architecture-parent-shift';moved['layouts'][0]['bounds']['x']-=12;moved['layouts'][0]['bounds']['y']-=6
 moved['notes']+=' Translation stress: parent shifts -12pt x/-6pt y; descendants and connectors follow without edits.'
 # Stress deliberately overruns one detail while retaining point size and all words.
 negative=copy.deepcopy(arch);negative['id']='dense-architecture-overflow'
 negative['layouts'][3]['cells'][0]['blocks'][1]['text']+=' The case retains all supporting evidence and reviewer qualifications. '*12
 OUT.mkdir(parents=True,exist_ok=True)
 for name,slides in [('dense-review',[arch,team]),('dense-translation',[moved]),('dense-overflow',[negative]),('dense-qualification',[arch,team,moved,negative])]:
  (OUT/(name+'.json')).write_text(json.dumps(dict(schema=dense['schema'],slides=slides),indent=2)+'\n')
 print('Wrote dense 9-node layered architecture and 21-role delivery organization, translation and overflow fixtures.')
if __name__=='__main__':main()
