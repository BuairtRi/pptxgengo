#!/usr/bin/env python3
"""Candidate dense proposal variants; every changed page still needs native QA.

Keep historically qualified fixtures untouched. These variants give substantive
prose its own measured zones instead of forcing it into node or timeline labels.
The fictional brief supplies example copy, not proof of a measured fit envelope.
"""
import copy
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
NAVY, BLUE, PALE, WHITE = '#070154', '#0047FF', '#E8EEF8', '#FFFFFF'
NARRATIVE = {s['id']: s for s in json.loads((ROOT/'library/proposal/narrative.json').read_text())['slides']}

def rect(x, y, w, h):
    return dict(x=x, y=y, width=w, height=h)

def text(identity, body, bounds, size=11, bold=False, color=NAVY):
    return dict(id=identity, kind='text', text=body, bounds=bounds,
                font_face='Arial', font_size_pt=size, bold=bold,
                foreground=color, align='left', valign='top', inset_x=0, inset_y=0)

def panel(identity, bounds, heading, paragraphs, size=11):
    blocks = [dict(id='heading', text=heading, font_face='Arial', font_size_pt=12,
                   bold=True, foreground=NAVY)]
    blocks += [dict(id=f'detail-{i+1}', text=p, font_face='Arial', font_size_pt=size,
                    foreground=NAVY, gap_before_pt=7) for i,p in enumerate(paragraphs)]
    return dict(id=identity, bounds=bounds, padding=dict(top=10,right=11,bottom=10,left=11),
                columns=[dict(weight=1)], rows=[dict(fixed_pt=bounds['height']-20)],
                surface=PALE, cells=[dict(id='content',row=0,column=0,
                padding=dict(top=0,right=0,bottom=0,left=0),blocks=blocks)])

def fixture(path, identity):
    return copy.deepcopy(next(s for s in json.loads((ROOT/path).read_text())['slides'] if s['id']==identity))

def bind(slide, narrative_id):
    n=NARRATIVE[narrative_id]
    slide['title']=n['assertion_title']
    slide['role']=n['role']; slide['takeaway']=n['takeaway']
    slide['notes']='Synthetic proposal capacity variant. Example copy is from library/proposal/brief.md. Native fit, final geometry and visual qualification pending; no source-fidelity claim.'
    return n['required_detail'], n['qualifications']

def process():
    s=fixture('library/diagram-components/review.json','process-changed-content')
    d,q=bind(s,'parallel-delivery-paths')
    # Retain the two editable five-step paths, with prose below each path and a
    # separate ownership column. Never expand the short node labels into prose.
    keep={'wm-logo','footer-copy','page','synthetic-kicker','synthetic-title'}
    s['canvas']=[c for c in s['canvas'] if c['id'] in keep]
    for c in s['canvas']:
        if c['id']=='synthetic-kicker': c['bounds']=rect(36,25,888,14);c['font_size_pt']=9
        if c['id']=='synthetic-title': c['text']=s['title'];c['bounds']=rect(36,44,888,62)
    s['title']='' # Visible title is the canvas title, bound explicitly in contract.
    s['canvas'] += [text('synthetic-one','01  Case decision path',rect(324,125,600,20),14,True),
                    text('synthetic-two','02  Change and release path',rect(324,295,600,20),14,True),
                    text('case-controls',' '.join(d[:3]),rect(324,220,600,61)),
                    text('release-controls',' '.join(d[3:6]),rect(324,390,600,61))]
    for p,y,labels in zip(s['paths'],[152,322],[
        ['Validate intake','Establish case ID','Assign owner','Decide case state','Route exceptions'],
        ['Log change','Review controls','Prove readiness','Steering gate','Release / recover']]):
        p['bounds']=rect(324, y, 600, 57);p['labels']=labels
    s['layouts']=[panel('decision-ownership',rect(36,125,266,191),'Accountability',d[6:],11),
                  panel('implementation-boundaries',rect(36,330,266,151),'Discovery boundaries',q,11)]
    return s,'parallel-delivery-paths'

def roadmap():
    s=fixture('library/visual-components/review.json','roadmap-interval-variant')
    d,q=bind(s,'roadmap-and-gates')
    s['canvas']=[c for c in s['canvas'] if c['id'] in {'kicker','footer-band','wm-logo','footer-copy','page'}]
    s['layouts']=[]
    s['canvas'].append(text('sequence-explanation',d[1],rect(36,110,888,31),11))
    x0,step=292,31.6
    for i in range(20):
        x=x0+i*step
        s['canvas'].append(dict(id=f'week-{i+1}',kind='surface',bounds=rect(x,149,step,22),background=PALE))
        t=text(f'week-{i+1}-number',str(i+1),rect(x+1,153,step-2,14),10,True)
        t.update(align='center',allow_overlap=[f'week-{i+1}'])
        s['canvas'].append(t)
    s['canvas'].append(text('axis-label','ILLUSTRATIVE WEEK',rect(36,153,240,14),10,True,BLUE))
    stages=[('Discovery and definition',1,4),('Build and prove',5,10),
            ('Pilot and adapt',11,14),('Scale and transition',15,20)]
    for i,(label,start,end) in enumerate(stages):
        y=183+i*33
        s['canvas'].append(text(f'phase-{i+1}-label',label,rect(36,y+3,242,21),11,True))
        # Native editable time bar: endpoints are exactly the week boundaries.
        s['canvas'].append(dict(id=f'phase-{i+1}-bar',kind='surface',
            bounds=rect(x0+(start-1)*step,y,(end-start+1)*step,23),background=BLUE if i%2==0 else NAVY))
        for boundary in (4,10,14,20):
            # Small discrete ticks avoid decorative guides crossing the bars.
            s['canvas'].append(dict(id=f'phase-{i+1}-tick-{boundary}',kind='line',
                bounds=rect(x0+boundary*step,y+25,0,4),foreground='#536A92',line_width_pt=.5))
    # Gate language is legible prose under the diagram, not a 170-character
    # milestone caption. Dates remain explicitly conditional.
    s['layouts']=[panel('readiness',rect(36,329,282,148),'01  Mobilize and sequence',[d[0],d[2]],10.5),
                  panel('pilot-gate',rect(339,329,282,148),'02  Authorize pilot',[d[4],d[3]],10.5),
                  panel('expansion-gate',rect(642,329,282,148),'03  Decide expansion',[d[5],q[1]],10.5)]
    s['canvas'].append(text('date-qualification',q[0],rect(36,481,888,17),10,True))
    return s,'roadmap-and-gates'

def team():
    s=fixture('library/diagram-components/dense-review.json','dense-delivery-relationships')
    d,q=bind(s,'team-and-accountability')
    edits={
        'client-panel-header':('Accountability and delivery coverage',None),
        'client-h-0':('Investment and scope',rect(654,169,258,17)),
        'client-b-0':(d[0],rect(654,189,258,32)),
        'client-h-1':('Cross-team coordination',rect(654,230,258,17)),
        'client-b-1':(d[1],rect(654,250,258,28)),
        'client-h-2':('Architecture and control decisions',rect(654,285,258,17)),
        'client-b-2':(d[2],rect(654,305,258,43)),
        'client-h-3':('Three delivery pods',rect(654,355,258,17)),
        'client-b-3':(' '.join(d[3:5]),rect(654,375,258,50)),
    }
    for c in s['canvas']:
        if c['id'] in edits:
            c['text'],bounds=edits[c['id']]
            if bounds:c['bounds']=bounds
    s['layouts']=s.get('layouts',[])+[
        panel('staffing-boundary',rect(36,125,177,156),'Responsibility model',[d[6],q[0]],10.5)]
    # The architect's route to shared roles crosses y289: reserve that corridor.
    # Qualifications below the two individual contributors stop before the legend.
    s['canvas'].append(text('client-readiness',q[1],rect(36,426,222,31),10.5))
    s['canvas'].append(text('shared-specialists',d[5],rect(36,460,222,25),10.5))
    # Role boxes stay at their existing measured font. Pod three reflects the
    # narrative rather than inheriting an unrelated data-only role label.
    s['pods'][0]['title']='Journey pod';s['pods'][1]['title']='Integration pod';s['pods'][2]['title']='Platform pod'
    return s,'team-and-accountability'

def roster():
    s=fixture('library/visual-components/review.json','illustrative-19-tile-roster')
    d,q=bind(s,'role-roster')
    for c in s['canvas']:
        if c['id'].endswith('-name'):
            number=c['id'].split('-')[1];c['text']='Role category '+number
    core=s['layouts'][0]['cells'][0]['blocks']
    core[0]['text']='Core coverage and allocation'
    for b,body in zip(core[1:],[d[0],d[1],q[0],'Confirm accountable owners for delivery, architecture, controls and service readiness.']):b['text']=body
    s['canvas'].append(text('staffing-agreement',' '.join([d[2],d[3],q[1]]),rect(36,455,888,42),10.5))
    return s,'role-roster'

def five_phase():
    s=fixture('library/layout-components/review.json','five-phase')
    d,q=bind(s,'five-phase-sequence')
    for c in s['canvas']:
        if c['id']=='engagement':c['text']='ILLUSTRATIVE WEEKS 1–20'
        if c['id']=='extension':c['text']='SUBJECT TO READINESS'
    layout=s['layouts'][0]
    layout['bounds']['height']=318
    layout['rows']=[dict(fixed_pt=42),dict(fixed_pt=200),dict(fixed_pt=76)]
    heads=['MOBILIZE','DISCOVER &\nDEFINE','BUILD &\nPROVE','PILOT &\nADAPT','SCALE &\nTRANSITION']
    times=['Pre-engagement','Weeks 1–4','Weeks 5–10','Weeks 11–14','Weeks 15–20']
    gates=['Named owners and access readiness authorize discovery.',
           'Agree the baseline and acceptance measures before implementation.',
           'Use test and control evidence to assess pilot readiness.',
           'Resolve findings before a steering expansion decision.',
           'Accept operational ownership before further expansion.']
    outputs=['Sponsor and owner map; access readiness; first journey hypothesis.',
             'Validated baseline; thin-release definition; acceptance checklist.',
             'Implemented journey; interface contract; telemetry and test evidence.',
             'Cohort observations; baseline comparison; owned readiness findings.',
             'Transfer acceptance; prioritized next journey; steering decision.']
    for i in range(5):
        cell=next(c for c in layout['cells'] if c['id']==f'phase-{i}-activities')
        cell['blocks']=[dict(id='heading',text=heads[i],font_face='Arial',font_size_pt=12,bold=True,foreground=NAVY,min_height_pt=30,align='center'),
            dict(id='timing',text=times[i],font_face='Arial',font_size_pt=9.5,foreground='#50658E',min_height_pt=15,gap_before_pt=2,align='center'),
            dict(id='activity',text=d[i],font_face='Arial',font_size_pt=10,foreground=NAVY,gap_before_pt=10),
            dict(id='progression',text=gates[i],font_face='Arial',font_size_pt=10,bold=True,foreground=NAVY,gap_before_pt=9)]
        out=next(c for c in layout['cells'] if c['id']==f'phase-{i}-output')
        out['padding']=dict(top=8,right=8,bottom=6,left=8)
        out['blocks'][0]['min_height_pt']=16
        out['blocks'][1]['text']=outputs[i];out['blocks'][1]['min_height_pt']=0
    s['canvas'].append(text('timing-and-scope-qualification',' '.join(q),rect(36,465,888,33),10.5))
    return s,'five-phase-sequence'

def discovery():
    s=fixture('library/visual-components/review.json','phase-detail-visual')
    d,q=bind(s,'discovery-detail')
    s.update(title_bounds=rect(36,44,888,62),title_font_face='Arial',title_font_size_pt=23,title_bold=True,title_foreground=NAVY)
    kept=[]
    for c in s['canvas']:
        identity=c['id']
        if identity in {'footer-band','wm-logo','footer-copy','page'}:kept.append(c)
        elif identity.startswith('adapted-'):
            c['bounds']['y']+=70;kept.append(c)
        elif identity=='source-artwork-note':
            c['bounds']=rect(716,324,208,14);kept.append(c)
    s['canvas']=kept+[text('kicker','ILLUSTRATIVE DISCOVERY • WEEKS 1–4',rect(36,25,888,14),9,True,BLUE),
        text('design-example-heading','Attributed design examples',rect(716,119,208,19),12,True),
        text('design-example-attribution',' '.join([d[5],q[1]]),rect(716,347,208,145),10.5)]
    sidebar=panel('phase-definition',rect(36,125,212,371),'Discover and define',
        [s['takeaway'],d[4],q[0]],11)
    sidebar['surface']=NAVY
    for b in sidebar['cells'][0]['blocks']:b['foreground']=WHITE
    sidebar['cells'][0]['blocks'][0]['font_size_pt']=18
    s['layouts']=[sidebar]
    # Four workstreams pair the exact discovery activity with its evidence output.
    rows=[('Journey and ownership',d[0],'Journey map; state owners; handoffs and exception decisions.'),
          ('Baseline validation',d[1],'Sampling notes; measure definitions; confirmed baseline owners.'),
          ('Systems and dependencies',d[2],'Interface register; legacy constraints; dependencies and owners.'),
          ('First-release definition',d[3],'Thin-release backlog; supported integration; control and acceptance checklist.')]
    cells=[]
    for i,(heading,activity,output) in enumerate(rows):
        for col,body,label in [(0,activity,heading),(1,output,'Evidence / output')]:
            cells.append(dict(id=f'workstream-{i}-'+('activities' if col==0 else 'outputs'),row=i,column=col,
                padding=dict(top=7,right=9,bottom=7,left=9),background=PALE if col==1 else WHITE,
                blocks=[dict(id='heading',text=label,font_face='Arial',font_size_pt=10.5,bold=True,foreground=NAVY),
                        dict(id='detail',text=body,font_face='Arial',font_size_pt=10,foreground=NAVY,gap_before_pt=6)]))
    s['layouts'].append(dict(id='discovery-workstreams',bounds=rect(266,125,432,371),
        columns=[dict(fixed_pt=265),dict(fixed_pt=163)],column_gap_pt=4,
        rows=[dict(fixed_pt=90)]*4,row_gap_pt=3,padding=dict(top=0,right=0,bottom=0,left=0),cells=cells))
    return s,'discovery-detail'

def architecture():
    s=fixture('library/diagram-components/dense-review.json','dense-layered-architecture')
    d,q=bind(s,'architecture-and-controls')
    for c in s['canvas']:
        for name,delta in [('experience',6),('services',12),('foundation',18)]:
            if c['id'].startswith(name+'-'):c['bounds']['y']-=delta
        if c['id']=='services-guide':c['text']='02  ORCHESTRATION LAYER'
        for i,name in enumerate(['experience','services','foundation']):
            if c['id']==name+'-why':c['text']=d[i]
    labels={
        'intake':('Requester portal','Validate intake and show missing information while retaining alternate routes during transition.'),
        'workbench':('Agent workspace','Make case state, assigned queue, evidence and next action visible to the people doing the work.'),
        'communications':('Supervisor decisions','Review waiting decisions, escalation reasons and reassignment choices with accountable owners.'),
        'orchestration':('Exception routing','Distinguish missing information, waiting decisions, technical failure and genuine new demand.'),
        'rules':('Case state and rules',d[6]),
        'integration':('Integration events',d[7]),
        'information':('Case data and audit',d[5]),
        'engineering':('Identity and access','Apply identity, least privilege and retention controls across the case journey and its evidence.'),
        'operations':('Telemetry and support','Observe processing and recovery events; connect service evidence to accountable operations owners.')}
    for l in s['layouts']:
        if l['id']=='platform':l['bounds']['height']-=18
        elif l['id']=='security-band':l['bounds']['y']-=18;l['cells'][0]['blocks'][0]['text']=d[4]
        elif l['id']=='governance-band':l['cells'][0]['blocks'][0]['text']=d[3]
        elif l['id'] in ['experience','services','foundation']:
            l['bounds']['y']-={'experience':6,'services':12,'foundation':18}[l['id']]
            for c in l['cells']:
                c['blocks'][0]['text'],c['blocks'][1]['text']=labels[c['id']]
    s['canvas'].append(text('architecture-qualification',' '.join(q),rect(36,474,888,28),10.5))
    return s,'architecture-and-controls'

def needs_response():
    s=fixture('library/visual-components/review.json','enablecomp-response-fixture')
    d,q=bind(s,'needs-and-responses')
    labels=['Reliable intake','Clear ownership','Controlled integration','Evidence for operations','Sustainable change']
    heads=['1. Validate intake and preserve access','2. Make state and escalation ownership explicit',
           '3. Start with one supported integration','4. Separate evidence from exception reasons',
           '5. Agree support and transfer ownership']
    for c in s['canvas']:
        if c['id']=='section-response':c['text']='PROPOSED RESPONSE'
        if c['id']=='section-needs':c['text']='ILLUSTRATIVE NEEDS'
        for i in range(1,6):
            if c['id']==f'need-label-{i}':c['text']=labels[i-1]
            elif c['id']==f'row-copy-{i}':
                c['paragraphs'][0]['runs'][0]['text']=heads[i-1]
                c['paragraphs'][1]['runs'][0]['text']=d[i-1]+(' '+d[5] if i==1 else '')
                c['bounds']['height']=52
    s['canvas'].append(text('requirements-qualification',' '.join(q),rect(36,467,888,32),10.5))
    return s,'needs-and-responses'

def main():
    generated=[process(),roadmap(),team(),roster(),five_phase(),discovery(),architecture(),needs_response()]
    path=ROOT/'library/proposal/capacity-templates.json'
    path.write_text(json.dumps(dict(schema='pptxgengo.compose-spec.v1',slides=[s for s,_ in generated]),indent=2,ensure_ascii=False)+'\n')
    # Values are authored through exactly the same semantic bindings as the
    # portable contracts, rather than relying on geometry overrides at call time.
    spec=importlib.util.spec_from_file_location('contracts',ROOT/'scripts/build-library-contracts.py')
    module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
    for s,narrative_id in generated:
        _,values=module.semantic_slots(s)
        (ROOT/f'library/proposal/values/{narrative_id}.json').write_text(json.dumps(dict(slots=values,style_variant='source'),indent=2,ensure_ascii=False)+'\n')
    print('Generated eight candidate proposal capacity variants and semantic example values.')

if __name__=='__main__':main()
