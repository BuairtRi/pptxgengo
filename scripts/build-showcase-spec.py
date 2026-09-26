#!/usr/bin/env python3
"""Generate the tracked, illustrative showcase spec from pinned local assets.
No slides or pixels are authored here: pptxcompose performs native construction.
"""
import json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
assets=json.loads((ROOT/'library/showcase/assets.json').read_text())['assets']
A={a['role']:a for a in assets if a['role'] not in ('icon','arrow')}
A.update({a['id']:a for a in assets})
NAVY='#070154';BLUE='#0047FF';PALE='#E8EEF8';SLATE='#50658E';WHITE='#FFFFFF'
def R(x,y,w,h):return dict(x=x,y=y,width=w,height=h)
def text(id,txt,x,y,w,h,size=14,bold=False,color=NAVY,align='left',bg='',over=()):
 d=dict(id=id,kind='text',bounds=R(x,y,w,h),text=txt,font_face='Arial',font_size_pt=size,bold=bold,foreground=color,align=align,valign='top')
 if bg:d['background']=bg
 if over:d['allow_overlap']=list(over)
 return d
def surface(id,x,y,w,h,c=PALE,over=()):return dict(id=id,kind='surface',bounds=R(x,y,w,h),background=c,allow_overlap=list(over))
def line(id,x,y,w,h,c='#CED7E6',over=()):return dict(id=id,kind='line',bounds=R(x,y,w,h),foreground=c,line_width_pt=1,allow_overlap=list(over))
def image(id,a,x,y,w,over=()):return dict(id=id,kind='image',bounds=R(x,y,w,w*a['height_px']/a['width_px']),asset_path='samples/showcase/'+a['path'],asset_sha256=a['sha256'],alt_text=a['label'],allow_overlap=list(over))
def slide(id,title,role,takeaway):
 return dict(id=id,title=title,width_pt=960,height_pt=540,title_bounds=R(42,35,876,64),title_font_face='Arial',title_font_size_pt=26,title_bold=True,title_foreground=NAVY,pods=[],canvas=[],role=role,takeaway=takeaway,notes='Illustrative scenario. The organization, scope and planning assumptions are synthetic. '+takeaway)
def accent(id,target,mode):
 a=A[mode]
 return dict(id=id,target=target,mode=mode,asset_path='samples/showcase/'+a['path'],asset_sha256=a['sha256'],alpha_bounds=R(0,0,1,1),padding_x_pt=4,padding_y_pt=3,offset_x_pt=0,offset_y_pt=-1 if mode=='underline' else 0,stroke_height_pt=3 if mode=='underline' else 0)
slides=[]
s=slide('cover','', 'Introduce the proposal and its illustrative scope','A staged approach makes the first investment decision explicit.')
s['canvas']=[text('kicker','ILLUSTRATIVE PROPOSAL',42,74,430,22,12,True,BLUE),text('cover-title','Software\nmodernization',42,128,450,110,40,True),text('cover-subtitle','A staged approach to discovery,\ndelivery and operational readiness',42,264,450,72,19),text('disclosure','Synthetic scenario for a regional service company\nSeptember 2026',42,378,450,48,12,False,SLATE),image('cover-photo',A['photo'],518,143,400)]
slides.append(s)
s=slide('decision','The first decision is a bounded discovery', 'Set a specific executive decision','Discovery should resolve the choices needed to authorize a first delivery slice.')
s['canvas']=[text('ask-heading','Decision requested',42,130,370,32,20,True),text('ask','Authorize discovery planning and nominate a business owner. Confirm the scope and fee after reviewing available evidence and stakeholder access.',42,181,370,130,17),text('scope-heading','Discovery needs to establish',472,130,446,32,20,True),text('scope','Critical customer journeys and their system dependencies\n\nOwnership of customer and case data\n\nConstraints on deployment, recovery and service continuity',472,181,446,200,15),line('divider',441,128,0,275),text('decision-note','The target design and delivery commitment follow the evidence review.',42,423,876,35,17,True,BLUE)]
slides.append(s)
s=slide('numbered','Three workstreams turn discovery into a decision', 'Demonstrate preferred N2 explanation rows','Each workstream contributes a concrete input to the investment decision.')
s['cards']=[dict(id=i,kind='numbered',bounds=R(42,y,876,102),number=n,title=t,body=b,surface='surface.light',accent=BLUE) for i,y,n,t,b in [
 ('estate',125,'01','Application and interface evidence','Map critical journeys, integration ownership and change dependencies. Confirm the interfaces that the first delivery slice must preserve.'),
 ('design',241,'02','Target design and release boundaries','Define a small end-to-end slice. Review data ownership, security constraints and the path to a controlled release.'),
 ('investment',357,'03','Delivery scope and decision rights','Agree acceptance evidence, participation and the next funding decision. Record assumptions that would change scope or sequencing.')]]
slides.append(s)
s=slide('metrics','Planning assumptions set a clear review boundary','Demonstrate metric components and composition-owned dividers','These example inputs organize the proposal; discovery must validate them.')
s['cards']=[dict(id=i,kind='metric',bounds=R(x,160,260,145),value=v,label=l,surface=PALE) for i,x,v,l in [('duration',42,'12','Weeks in the example workplan'),('pods',350,'2','Outcome delivery pods'),('gates',658,'4','Decision and readiness gates')]]
s['canvas']=[text('assumption','Illustrative planning inputs. These are not measured client results or a delivery commitment.',42,113,876,30,13,False,SLATE),line('metric-divider-1',329,169,0,128),line('metric-divider-2',637,169,0,128),text('baseline-heading','Benefits need an agreed baseline',42,352,876,30,20,True),text('baseline-body','Discovery will identify data sources and owners for change lead time, failed changes, interface incidents and duplicate records. Targets follow that baseline.',42,395,876,68,16)]
slides.append(s)
s=slide('architecture','Managed interfaces separate change responsibilities','Explain the conceptual architecture','Clear boundaries let the delivery team change a capability without redesigning every adjacent system.')
s['canvas']=[image('database-icon',A['icon-data-database'],58,140,74),image('cloud-icon',A['icon-cloud'],58,292,74),text('data-heading','Governed data',158,138,298,30,20,True),text('data-body','Business owners define customer and case meaning. Published interfaces make those definitions available to consuming services.',158,181,298,94,14),text('runtime-heading','Runtime and operations',158,298,298,30,20,True),text('runtime-body','Shared identity, telemetry and release controls support each application team. Recovery procedures accompany each release.',158,341,298,95,14),line('architecture-divider',485,127,0,325)]
for i,(label,y) in enumerate([('Customer and colleague channels',137),('Customer and case services',215),('API and event interfaces',293),('Data products and shared runtime',371)]):
 s['canvas'].append(text('layer-'+str(i),label,532,y,386,50,15,True,WHITE if i%2==0 else NAVY,'center',NAVY if i%2==0 else PALE))
 if i<3:s['canvas'].append(line('join-'+str(i),725,y+50,0,28,NAVY,['layer-'+str(i),'layer-'+str(i+1)]))
s['canvas'].append(text('architecture-note','Conceptual design, subject to discovery and platform constraints',532,449,386,30,10,False,SLATE))
slides.append(s)
s=slide('roadmap','A first release creates evidence for later waves','Show an editable illustrative Gantt-style workplan','Acceptance evidence governs expansion rather than elapsed time alone.')
s['canvas']=[text('schedule-note','Illustrative 12-week sequence. Timing depends on scope, access and team availability.',42,107,876,30,12,False,SLATE)]
for j,label in enumerate(['Weeks 1–3','Weeks 4–6','Weeks 7–9','Weeks 10–12']):s['canvas'].append(text('period-'+str(j),label,248+j*167,153,160,27,13,True,align='center'))
rows=[('Discover and align',1,0,'Evidence and scope agreed'),('Prove a vertical slice',2,1,'End-to-end acceptance'),('Expand priority capabilities',2,2,'Release readiness'),('Stabilize and transition',1,3,'Operational handoff')]
for i,(label,dur,start,gate) in enumerate(rows):
 y=207+i*63
 s['canvas'].append(text('work-'+str(i),label,42,y,190,39,13,True))
 # A deliberately overlapping expansion follows the proven slice.
 bw=min(dur,4-start)*167-12
 s['canvas'].append(text('bar-'+str(i),gate,248+start*167,y,bw,38,12,True,WHITE,'center',BLUE if i%2 else NAVY))
 s['canvas'].append(line('row-rule-'+str(i),42,y+49,876,0))
slides.append(s)
s=slide('comparison','Release strategy depends on service constraints','Compare two delivery options without presuming a decision','Discovery should select a release strategy using continuity and dependency evidence.')
xs=[42,262,590];ws=[220,328,328]
headers=['Decision factor','Incremental replacement','Coordinated cutover']
for j in range(3):s['canvas'].append(text('table-head-'+str(j),headers[j],xs[j],130,ws[j]-2,40,14,True,WHITE,'left',NAVY))
data=[('Service continuity','Move one bounded journey at a time','Plan a single transition window'),('Dependency pattern','Stable interfaces isolate each release','Shared dependencies move together'),('Rollback approach','Return traffic to the existing path','Restore the agreed pre-cutover state'),('Evidence to review','Routing, data coexistence and monitoring','Rehearsal, reconciliation and recovery')]
for i,row in enumerate(data):
 y=177+i*68
 for j,t in enumerate(row):
  s['canvas'].append(text(f'cell-{i}-{j}',t,xs[j]+8,y+7,ws[j]-20,51,12,j==0))
 s['canvas'].append(line('table-rule-'+str(i),42,y+61,876,0))
slides.append(s)
# Compact team derived from the validated team model; positions remain explicit.
s=slide('team','Delivery pods own bounded capabilities','Demonstrate roles, variable pods, staffing semantics and routes','Shared direction and clear pod boundaries connect ownership to delivery.')
s['roles']=[dict(id='sponsor',label='Engagement lead',background='staffing.wm_full_time',foreground='auto',bounds=R(374,112,212,34),font_face='Arial',font_size_pt=11,bold=True,horizontal_inset_pt=7,vertical_inset_pt=4),dict(id='architect',label='Platform architect',background='staffing.wm_part_time',foreground='auto',bounds=R(374,172,212,34),font_face='Arial',font_size_pt=11,bold=True,horizontal_inset_pt=7,vertical_inset_pt=4)]
for id,x,title,labels in [('foundation',42,'Platform enablement',['Platform lead','Operations engineer']),('customer',350,'Customer and case',['Delivery lead','Software engineer','Client product owner']),('integration',658,'Integration and data',['Delivery lead','Data engineer','Quality engineer'])]:
 s['pods'].append(dict(id=id,title=title,bounds=dict(x=x,y=246,width=260,max_height_pt=178),layout=dict(columns=1,gap_pt=6,padding_pt=10,title_gap_pt=10,min_tile_width_pt=95),style=dict(font_face='Arial',font_size_pt=11,bold=True,title_font_face='Arial',title_font_size_pt=13,title_bold=True,tile_min_height_pt=30,horizontal_inset_pt=7,vertical_inset_pt=4,paragraph_gap_pt=0,surface='surface.light',foreground='auto',title_foreground='auto'),roles=[dict(id=id+'-'+str(i),label=l,background='staffing.client_part_time' if 'Client' in l else 'staffing.wm_full_time',foreground='auto') for i,l in enumerate(labels)]))
s['phases']=[dict(id='delivery-phase',title='ILLUSTRATIVE DELIVERY SCOPE',members=['customer','integration'],bounds=R(338,234,592,216),padding_pt=12,label_height_pt=18,font_face='Arial',font_size_pt=10,bold=True,surface='#CCDAFF',foreground='auto')]
s['legend']=dict(bounds=R(42,468,876,16),font_face='Arial',font_size_pt=9,bold=True,foreground=NAVY,swatch_size_pt=10,gap_pt=6,item_gap_pt=24)
s['connections']=[dict(id=f'{a}-{b}',**{'from':a,'to':b},relationship='reporting',from_anchor='bottom',to_anchor='top',color=NAVY,width_pt=1.1,clearance_pt=6) for a,b in [('sponsor','architect'),('architect','foundation'),('architect','customer'),('architect','integration')]]
slides.append(s)
s=slide('readiness','Each gate needs evidence and a decision owner','Show segmented evidence-gate layouts at readable density','A gate closes when the owner accepts the evidence, not when a status turns green.')
for k,(x,heading,evidence,exit,owner) in enumerate([(42,'Design agreed','Journey map\nInterface boundaries\nData ownership decisions','A vertical slice has an agreed scope and acceptance plan','Business owner'),(350,'Release ready','End-to-end test results\nSecurity review\nRollback rehearsal','The release meets its service and recovery criteria','Service owner'),(658,'Handoff complete','Operating runbook\nTelemetry and alerts\nSupport knowledge transfer','Operations can support and recover the service','Operations lead')]):
 w=260;ids=[f'g{k}-{q}' for q in ['title','evidence','exit','owner']]
 s['canvas'].append(surface(f'g{k}-surface',x,135,w,310,PALE,ids))
 s['canvas'] += [text(ids[0],heading,x+16,153,w-32,36,18,True),text(ids[1],evidence,x+16,207,w-32,83,13),text(ids[2],exit,x+16,306,w-32,69,13,True),text(ids[3],owner,x+16,408,w-32,25,11,True,BLUE)]
 s['canvas'].append(line(f'g{k}-divider',x+16,294,w-32,0,'#CED7E6',[f'g{k}-surface']))
slides.append(s)
s=slide('learning','Review evidence before expanding the next wave','Demonstrate selected hand-drawn arrow compositions','The delivery team revisits scope after reviewing a working slice.')
arrow=A['arrow-wm_handdrawn_connecting_arrow_rgb_240912'];dashed=A['arrow-wm_handdrawn_dashed_arrow_rgb_240912']
s['canvas']=[image('review-loop',arrow,90,158,165,['review-label']),text('review-label','Review\nevidence',120,222,105,70,18,True,align='center'),text('learning-heading','The working slice informs the next decision',342,146,576,56,20,True),text('learning-body','Compare observed behavior with the agreed acceptance criteria. Record gaps in data, integration and service recovery before expanding scope.',342,215,576,94,16),image('optional-path',dashed,350,334,102),text('next-wave','Next-wave scope',480,408,385,30,20,True),text('learning-note','The illustrated feedback path is a proposed review step',42,464,876,24,11,False,SLATE)]
slides.append(s)
s=slide('next-step','', 'Close with a concrete working session','A short working session can identify the first evidence owners and shape discovery.')
s['canvas']=[text('hero-a','Evidence before expansion',42,68,876,62,32,True),text('hero-b','A focused working session',42,162,876,48,24,True),text('session-time','90 minutes',42,259,210,48,30,True,BLUE),text('agenda','Confirm the business outcomes\nMap the critical journeys and dependencies\nIdentify evidence owners and access\nAgree discovery participants and the decision path',300,253,618,150,17),text('closing','The sponsor nominates business, technology and operations counterparts.',42,435,876,39,16)]
s['accents']=[accent('evidence-highlight','hero-a','highlight'),accent('session-underline','hero-b','underline')]
slides.append(s)
# Consistent brand frame rebuilt with native elements and the official logo.
footer='© 2026 West Monroe Partners. Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.'
for n,s in enumerate(slides,1):
 frame=[surface('footer-band',0,504,960,36,'#E8EEF8',['footer-copy','page-number','wm-logo']),image('wm-logo',A['logo'],42,509,104),text('footer-copy',footer,177,512,698,19,8.1,False,NAVY),text('page-number',str(n),890,512,28,19,9,False,NAVY,'right')]
 s['canvas']=frame+s['canvas']
 s['notes']+=' Role: '+s['role']+' Takeaway: '+s['takeaway']
 s['notes']+=' Brand artwork source paths and hashes: library/showcase/assets.json. This sample does not assert actual client benefits.'
(ROOT/'library/showcase/deck.json').write_text(json.dumps(dict(schema='pptxgengo.compose-spec.v1',slides=slides),indent=2)+'\n')
print(f'{len(slides)} slides -> library/showcase/deck.json')
