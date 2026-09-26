#!/usr/bin/env python3
"""Author a proposal-density benchmark. All visible text remains native/editable.

Reference patterns: EnableComp 3/4, UHG 28/38/43. Content is synthetic;
these are new compositions, not claims of pixel-identical reconstruction.
Geometry is explicit in points. Native PowerPoint measurement is required.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
NAVY, BLUE, PALE, SLATE, WHITE = '#070154', '#0047FF', '#E8EEF8', '#50658E', '#FFFFFF'
PINK, LILAC, RULE = '#F900D6', '#CFDAFF', '#CED7E6'
assets = json.loads((ROOT/'library/showcase/assets.json').read_text())['assets']
logo = next(a for a in assets if a['role'] == 'logo')

def rect(x,y,w,h): return dict(x=x,y=y,width=w,height=h)
def tx(id,txt,x,y,w,h,size=10.5,bold=False,fg=NAVY,bg='',align='left',pad=0,py=0):
    d = dict(id=id,kind='text',bounds=rect(x,y,w,h),text=txt,font_face='Arial',font_size_pt=size,bold=bold,foreground=fg,align=align,valign='top',inset_x=pad,inset_y=py)
    if bg: d['background']=bg
    return d
def surface(id,x,y,w,h,bg,children):
    return dict(id=id,kind='surface',bounds=rect(x,y,w,h),background=bg,allow_overlap=children)
def line(id,x,y,w,h,fg=RULE,over=()):
    return dict(id=id,kind='line',bounds=rect(x,y,w,h),foreground=fg,line_width_pt=.65,allow_overlap=list(over))
def slide(id,kicker,title,takeaway):
    return dict(id=id,title=title,width_pt=960,height_pt=540,title_bounds=rect(36,44,888,62),title_font_face='Arial',title_font_size_pt=23,title_bold=True,title_foreground=NAVY,pods=[],canvas=[tx('kicker',kicker,36,25,888,14,9,True,SLATE)],role=kicker,takeaway=takeaway,notes='Illustrative software modernization proposal. All scope, roles, durations and measures are synthetic planning examples. '+takeaway)
def cell(s,id,txt,x,y,w,h,bg=PALE,size=10.5,bold=False,fg=NAVY,pad=8,py=7,align='left'):
    s['canvas'].append(tx(id,txt,x,y,w,h,size,bold,fg,bg,align,pad,py))
def bullets(s,id,items,x,y,w,heights,size=10.5,fg=NAVY,bg=''):
    for i,(t,h) in enumerate(zip(items,heights)):
        s['canvas'] += [tx(f'{id}-dot-{i}','•',x,y,7,h,size,False,fg,bg),tx(f'{id}-text-{i}',t,x+12,y,w-12,h,size,False,fg,bg)]
        y += h+3
def panel(s,id,x,y,w,h,color,build):
    start=len(s['canvas']);build()
    children=[c['id'] for c in s['canvas'][start:]]
    s['canvas'].insert(start,surface(id,x,y,w,h,color,children))

slides=[]
s=slide('five-phase','PROPOSED DELIVERY APPROACH','A staged modernization program proves the release model,\nthen expands capability and transfers operational ownership','Each phase pairs detailed activities with an explicit evidence package and a decision to proceed.')
s['canvas'] += [tx('engagement','ILLUSTRATIVE FIRST RELEASE — WEEKS 1–12',36,112,700,18,10.5,True),tx('extension','SUBJECT TO READINESS',752,112,172,18,9,True,SLATE)]
phases=[
('00','DISCOVER &\nMOBILIZE','Pre-engagement',['Inventory critical journeys, applications, interfaces and known service constraints.','Identify business owners, technical counterparts and access dependencies.','Confirm the evidence needed to scope discovery and agree decision rights.'],'A prioritized evidence register, stakeholder map, access plan and confirmed discovery questions.'),
('01','ALIGN &\nVALIDATE','Weeks 1–3',['Map current journeys and trace the interfaces and data they rely on.','Agree capability boundaries, data ownership and nonfunctional acceptance criteria.','Select the first vertical slice and establish the release and recovery approach.'],'A scoped vertical slice, architecture decisions, baseline measures and an agreed acceptance plan.'),
('02','BUILD &\nPROVE','Weeks 4–7',['Implement the journey across application, integration and data boundaries.','Exercise security, observability, reconciliation and recovery procedures.','Demonstrate working software with business users and resolve acceptance gaps.'],'An end-to-end working slice, test evidence, operating runbook and prioritized findings for expansion.'),
('03','RELEASE &\nSTABILIZE','Weeks 8–12',['Rehearse cutover and rollback with service owners before release approval.','Monitor customer outcomes, interface errors and operational workload.','Address defects, confirm support readiness and validate release controls.'],'A controlled production release, accepted service evidence and a documented operational handoff.'),
('04','EXPAND &\nTRANSFER','Subsequent waves',['Prioritize the next capabilities using value, dependency and risk evidence.','Reuse validated patterns while adapting to each journey’s constraints.','Transfer ownership, sustain measurement and retire legacy paths when safe.'],'A sequenced capability roadmap, ownership transition plan and criteria for legacy decommissioning.')]
for i,(n,t,d,items,out) in enumerate(phases):
    x=36+i*180
    cell(s,f'phase-{i}',n,x,139,168,42,SLATE if i==4 else NAVY,25,True,WHITE,8,5)
    def body(i=i,x=x,t=t,d=d,items=items,out=out):
        s['canvas'] += [tx(f'heading-{i}',t,x+8,190,152,34,12,True,align='center'),tx(f'dates-{i}',d,x+8,226,152,17,9.5,False,SLATE,align='center'),tx(f'activity-{i}','Activities',x+8,250,152,18,11,True)]
        bullets(s,f'activities-{i}',items,x+8,274,152,[38,50,38] if i==1 else [38,38,50],10)
    panel(s,f'body-{i}',x,181,168,227,PALE,body)
    def output(i=i,x=x,out=out):
        s['canvas'] += [tx(f'output-label-{i}','Evidence / output',x+8,418,152,18,11,True,BLUE),tx(f'output-{i}',out,x+8,440,152,49,10)]
    panel(s,f'output-surface-{i}',x,408,168,88,RULE,output)
slides.append(s)

s=slide('phase-detail','PHASE 02 — BUILD & PROVE','','The first slice must prove architecture, engineering, governance and operations together before expansion.')
s['canvas']=[]
# The phase detail uses a dedicated sidebar instead of an ordinary title zone.
def sidebar():
    s['canvas'] += [tx('phase-kicker','OUR APPROACH',36,28,198,18,10,True,WHITE,NAVY),tx('phase-number','PHASE 02',36,57,198,22,16,True,WHITE,NAVY),tx('phase-title','Build & prove\nthe first slice',36,90,198,65,25,True,WHITE,NAVY),tx('phase-dates','Weeks 4–7',36,169,198,25,15,True,WHITE,NAVY),tx('objective-head','Objective',36,221,198,23,14,True,WHITE,NAVY),tx('objective','Demonstrate that one priority customer journey can operate across application, integration and data boundaries with agreed service controls. Use the working slice to validate reusable patterns, expose constraints and establish the evidence needed for a controlled release.',36,255,198,133,12,False,WHITE,NAVY),tx('exit-head','Phase exit evidence',36,400,198,23,14,True,WHITE,NAVY)]
    bullets(s,'exit',['Journey accepted by the owner','Recovery procedures exercised','Release gaps owned and prioritized'],36,432,198,[16,16,28],10.5,WHITE,NAVY)
panel(s,'sidebar',0,0,252,504,NAVY,sidebar)
cell(s,'activity-header','Activities',270,28,272,30,PALE,12,True)
cell(s,'product-header','Work products',544,28,152,30,PALE,12,True)
rows=[
('ARCHITECTURE',['Validate service and data boundaries through the end-to-end journey.','Resolve identity, integration and deployment decisions with platform owners.','Record reusable patterns and constraints that affect subsequent waves.'],['Architecture decisions and interface contracts','Reusable reference patterns']),
('ENGINEERING',['Build the vertical slice with automated checks and deployment controls.','Exercise reconciliation, telemetry, security and recovery scenarios.','Resolve defects and document gaps that would prevent a controlled release.'],['Working software and automated test evidence','Release and recovery procedures']),
('GOVERNANCE',['Confirm business ownership, access policies and sensitive-data handling.','Define acceptance authority, escalation paths and exception decisions.','Review readiness evidence with the service owner and record conditions.'],['Ownership and decision register','Readiness review and risk actions']),
('OPERATIONS & ADOPTION',['Validate runbooks with the people who will support the service.','Observe users completing the journey; capture process and training gaps.','Prioritize findings and confirm which patterns are ready to reuse.'],['Operating runbook and support model','User feedback and expansion backlog'])]
for i,(h,acts,products) in enumerate(rows):
    y=67+i*106
    s['canvas'] += [tx(f'activity-label-{i}',h,278,y,255,16,10.5,True),tx(f'product-label-{i}',h,552,y,136,27,10,True)]
    bullets(s,f'activity-row-{i}',acts,278,y+20,255,[24,24,24],10)
    bullets(s,f'product-row-{i}',products,552,y+30,136,[32,32],10)
    if i<3:s['canvas'].append(line(f'row-{i}',270,y+100,426,0))
s['canvas'].append(line('vertical-divider',543,28,0,461,over=['row-0','row-1','row-2']))
s['canvas'] += [tx('deliverable-title','Acceptance package',716,29,208,23,13,True,BLUE),tx('deliverable-caption','Three views of the same release decision',716,58,208,29,10,False,SLATE)]
for i,(h,t) in enumerate([('Business acceptance','Journey outcomes, exception handling and process ownership'),('Technical readiness','Security, interface integrity, deployment and recovery evidence'),('Service readiness','Monitoring, runbooks, support coverage and escalation paths')]):
    y=99+i*54
    cell(s,f'acceptance-{i}',h+'\n'+t,716,y,208,48,PALE,10,False,NAVY,8,5)
def assumptions():
    s['canvas'].append(tx('assumptions-heading','Assumptions',724,274,192,21,12,True,BLUE))
    bullets(s,'assumptions',['Named owners and a representative user group are available for reviews.','Required environments, test data and interfaces can be accessed.','No unresolved security constraint prevents the selected journey.'],724,302,192,[37,37,37],10)
panel(s,'assumptions-surface',716,265,208,160,PALE,assumptions)
s['canvas'] += [tx('scope-heading','Variable scope',716,435,208,20,12,True,BLUE),tx('scope-text','Additional interfaces, complex data remediation or broader process change require explicit scope and sequencing decisions.',716,462,208,37,10)]
slides.append(s)

s=slide('workflow-matrix','MODERNIZATION OPPORTUNITY','Modernization should remove friction across the service journey\nwhile preserving ownership, controls and measurable outcomes','Link each workflow change to specific control evidence and a measure that can be baselined.')
xs=[36,174,414,728];ws=[134,234,308,196]
for i,h in enumerate(['Workflow','Current constraints','Proposed operating model','Measures to baseline']):
    cell(s,f'matrix-head-{i}',h,xs[i],127,ws[i],28,NAVY if i==0 else PALE,11,True,WHITE if i==0 else NAVY,7,6)
data=[
('01  Customer intake','Teams re-enter customer and case details across channels. Incomplete requests create clarification loops and inconsistent handoffs.','Capture required information once, validate it at entry and route exceptions to a named owner. Retain human review where judgment or policy requires it.','Completion time; rework rate; first-pass completeness; avoidable handoffs'),
('02  Case preparation','Specialists gather documents and reference data from multiple systems. The evidence trail depends on local files and individual knowledge.','Assemble a traceable case record through managed interfaces. Show source provenance, missing evidence and reconciliation status before the case advances.','Preparation time; missing evidence; duplicate records; reconciliation exceptions'),
('03  Decision & action','Work queues hide dependency and approval status. Escalations rely on messages, and the rationale for decisions can be difficult to recover.','Make state, ownership and approval rules explicit. Route exceptions by reason, capture decision rationale and retain an auditable record of human approvals.','Queue age; approval time; exception volume; reopened cases; audit completeness'),
('04  Fulfilment & service','Downstream changes require coordinated manual updates. Failures emerge through user reports, and teams reconcile outcomes after the event.','Publish validated events and managed APIs. Monitor end-to-end completion, expose failed handoffs and provide controlled retry, rollback and recovery procedures.','Failed handoffs; recovery time; reconciliation lag; repeat contacts'),
('05  Improvement & control','Reporting is assembled retrospectively. Product, service and business owners use different measures to prioritize fixes and investment.','Create a shared measurement cadence with agreed definitions and owners. Use observed outcomes to prioritize the backlog, assess controls and select the next wave.','Change lead time; service incidents; adoption; support demand; cost to serve')]
for i,row in enumerate(data):
    y=161+i*64
    for j,t in enumerate(row):
        cell(s,f'matrix-{i}-{j}',t,xs[j],y,ws[j],62,[NAVY,PALE,LILAC,'#FCE5F8'][j],10.5,j==0,WHITE if j==0 else NAVY,8,5.5)
s['canvas'].append(tx('matrix-takeaway','Targets follow an agreed baseline; the measures above are candidate definitions, not claimed benefits.',36,486,888,16,10,True,BLUE))
slides.append(s)

s=slide('delivery-organization','TEAM STRUCTURE','Scale delivery through accountable pods, shared specialists\nand explicit client ownership of the release decision','Reporting, shared capabilities, time-limited roles and client decision rights must be legible on one page.')
def role(id,label,x,y,w=106,h=32,kind='wm_full_time'):
    s.setdefault('roles',[]).append(dict(id=id,label=label,bounds=rect(x,y,w,h),background='staffing.'+kind,foreground='auto',font_face='Arial',font_size_pt=9.5,bold=True,horizontal_inset_pt=6,vertical_inset_pt=4))
for id,label,x,y,kind in [('sponsor','Executive\nsponsor',236,125,'wm_part_time'),('client-sponsor','Client executive\nsponsor',408,125,'client_part_time'),('lead','Engagement\nlead',236,180,'wm_full_time'),('client-lead','Client program\nlead',408,180,'client_part_time'),('architect','Enterprise\narchitect',236,235,'wm_full_time'),('client-owner','Client platform\nowner',408,235,'client_part_time')]:role(id,label,x,y,kind=kind)
for i,(id,label,kind) in enumerate([('infra','Infrastructure\narchitect','wm_part_time'),('change','Change &\ntraining lead','wm_part_time'),('program','Program\nmanager','wm_full_time'),('fabric','Solution\narchitect','wm_full_time'),('governance','Governance\nspecialist','wm_part_time')]):
    # Align managers directly above their children; preserve stable request order.
    x=152 if id=='program' else (268 if id=='change' else 36+i*116)
    role(id,label,x,311,kind=kind)
for a,b in [('sponsor','lead'),('lead','architect')]+[('architect',k) for k in ['infra','change','program','fabric','governance']]:
    s.setdefault('connections',[]).append(dict(id=a+'-'+b,**{'from':a,'to':b},relationship='reporting',from_anchor='bottom',to_anchor='top',color=NAVY,width_pt=.75,clearance_pt=4))
for i in range(3):s['canvas'].append(line('counterpart-'+str(i),342,141+i*55,66,0,NAVY,['sponsor','client-sponsor','lead','client-lead','architect','client-owner']))
role('infra-engineer','Platform\nengineer',36,381)
role('analyst','Program analyst\n& release control',152,381)
for a,b in [('infra','infra-engineer'),('program','analyst')]:s['connections'].append(dict(id=a+'-'+b,**{'from':a,'to':b},relationship='reporting',from_anchor='bottom',to_anchor='top',color=NAVY,width_pt=.75,clearance_pt=4))
for i,(id,title,labels) in enumerate([('pod-1','Journey pod',['Delivery lead','Software engineer']),('pod-2','Integration pod',['Delivery lead','Integration engineer','Quality engineer']),('pod-3','Data pod',['Delivery lead','Data engineer','Quality engineer'])]):
    x=276+i*114
    s['pods'].append(dict(id=id,title=title,bounds=dict(x=x,y=374,width=108,max_height_pt=111),layout=dict(columns=1,gap_pt=4,padding_pt=5,title_gap_pt=5,min_tile_width_pt=95),style=dict(font_face='Arial',font_size_pt=8.5,bold=True,title_font_face='Arial',title_font_size_pt=9.5,title_bold=True,tile_min_height_pt=23,horizontal_inset_pt=4,vertical_inset_pt=3,paragraph_gap_pt=0,surface='surface.light',foreground='auto',title_foreground='auto'),roles=[dict(id=id+'-'+str(j),label=t,background='staffing.wm_full_time',foreground='auto') for j,t in enumerate(labels)]))
    s['connections'].append(dict(id='fabric-'+id,**{'from':'fabric','to':id},relationship='reporting',from_anchor='bottom',to_anchor='top',color=NAVY,width_pt=.75,clearance_pt=4))
def leadership():
    cell(s,'client-panel-header','Client leadership & decision rights',642,125,282,32,NAVY,12,True,WHITE,10,8)
    entries=[('Executive sponsor','Sets strategic direction, resolves organizational blockers and approves major investment decisions.'),('Program lead','Owns internal coordination, dependencies, stakeholder availability and escalation of program risks.'),('Platform / service owner','Prioritizes the backlog, accepts service evidence and owns the support model and release decision.'),('Business journey owners','Validate outcomes, approve process changes and nominate users for acceptance and adoption activity.')]
    for i,(h,t) in enumerate(entries):
        y=169+i*65
        s['canvas'] += [tx(f'client-h-{i}',h,654,y,258,18,11,True),tx(f'client-b-{i}',t,654,y+22,258,35,10.5)]
    s['canvas'] += [tx('staffing-head','Staffing and transition assumptions',654,429,258,18,10.5,True,BLUE),tx('staffing-body','Pod count and participation vary by scope. Specialist support is shared; client ownership continues after delivery.',654,452,258,36,10)]
panel(s,'client-panel',642,125,282,371,PALE,leadership)
s['legend']=dict(bounds=rect(36,487,574,15),font_face='Arial',font_size_pt=8.5,bold=True,foreground=NAVY,swatch_size_pt=9,gap_pt=5,item_gap_pt=16)
slides.append(s)

s=slide('release-roadmap','ILLUSTRATIVE RELEASE & READINESS PLAN','Sequence overlapping workstreams around evidence gates,\nwith operational readiness continuing through each release','The roadmap connects delivery timing to dependencies, ownership and explicit decisions.')
s['canvas'].append(tx('timeline-note','Illustrative month-level plan: first release by Month 3; expansion follows accepted evidence. Validate dates in discovery.',36,114,888,18,10,False,SLATE))
labelw=174;start=216;col=43
for j in range(12):cell(s,f'month-{j}',str(j+1),start+j*col,143,col-1,26,NAVY if j<9 else PALE,10.5,True,WHITE if j<9 else NAVY,0,6,'center')
cell(s,'workstream-header','Workstream / accountable owner',36,143,174,26,PALE,10,True, NAVY,6,6)
cell(s,'gate-header','Evidence / decision gate',750,143,174,26,PALE,10,True,NAVY,6,6)
rows=[
('Discover & align\nBusiness / architecture',0,1,'M1','Priority journeys, evidence owners and acceptance criteria agreed.'),
('Foundation controls\nPlatform enablement',0,2,'M1–2','Environment, deployment and recovery controls exercised.'),
('First slice & acceptance\nJourney pod',1,2,'M2–3','Business owner accepts the complete journey and exception paths.'),
('Release & stabilize\nService owner',2,1,'M3','Operational readiness, rollback and support coverage approved.'),
('Capability expansion\nDelivery pods',3,9,'Reuse and adapt accepted patterns','Each capability meets service and data acceptance criteria.'),
('Data & reconciliation\nData owner',0,12,'Ownership, quality and migration controls','Data definitions, reconciliation and retention decisions maintained.'),
('Adoption & transition\nBusiness / operations',1,11,'Training, adoption and ownership transfer','Named service owners can operate, recover and improve the service.')]
for i,(lab,a,d,label,gate) in enumerate(rows):
    y=178+i*39
    s['canvas'] += [tx(f'work-{i}',lab,42,y+3,165,32,9.5,True),tx(f'gate-{i}',gate,756,y+3,162,33,9)]
    cell(s,f'bar-{i}',label,start+a*col,y+4,d*col-5,28,SLATE if i in (1,5) else (BLUE if i==6 else NAVY),9.5,True,WHITE,5,6,'center')
    s['canvas'].append(line(f'roadmap-row-{i}',36,y+37,888,0))
def dependencies():
    s['canvas'] += [tx('dependency-head','Decision dependencies',46,466,175,19,11,True,BLUE),tx('dependency-body','Scope precedes build • Recovery evidence precedes release • Accepted patterns precede expansion • Demonstrated service ownership precedes handoff',231,466,681,27,10)]
panel(s,'dependency-band',36,457,888,40,PALE,dependencies)
slides.append(s)

footer='© 2026 West Monroe Partners. Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.'
for i,s in enumerate(slides,1):
    s['canvas'] += [surface('footer-band',0,504,960,36,PALE,['wm-logo','footer-copy','page']),dict(id='wm-logo',kind='image',bounds=rect(36,511,104,104*logo['height_px']/logo['width_px']),asset_path='samples/showcase/'+logo['path'],asset_sha256=logo['sha256'],alt_text=logo['label']),tx('footer-copy',footer,174,516,700,17,8),tx('page',str(i),892,516,32,17,9,align='right')]
    s['notes'] += ' Reference pattern: '+['EnableComp 3','UHG 28','EnableComp 4','UHG 43','UHG 38'][i-1]+'. New synthetic copy and explicit geometry; not a source reconstruction.'
out=ROOT/'library/showcase/dense-deck.json'
out.write_text(json.dumps(dict(schema='pptxgengo.compose-spec.v1',slides=slides),indent=2)+'\n')
print(f'{len(slides)} dense slides -> {out}')
