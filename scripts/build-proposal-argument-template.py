#!/usr/bin/env python3
"""Dense, editable opening argument candidate; native qualification required."""
import copy,json
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'library/proposal/argument-template.json'
navy='#070154'; pale='#E8EEF8';blue='#0047FF'
base=json.loads((ROOT/'library/diagram-components/accent-review.json').read_text())['slides'][4]
s=copy.deepcopy(base)
s.update(id='dense-argument',role='Authorize a bounded discovery and first-release definition phase',
         takeaway='Prove one observable journey and its operating controls before expanding the program.',
         notes='Fictional proposal template using source-pinned WM highlight and original WM footer. Original UHG6 highlight calibration is a candidate here, not proof of native fit or source-slide identity.')
s['canvas']=copy.deepcopy(base['canvas'][:4])

def text(id,t,x,y,w,h,size,bold=False,color=navy):
 return dict(id=id,kind='text',bounds=dict(x=x,y=y,width=w,height=h),text=t,font_face='Arial',font_size_pt=size,bold=bold,foreground=color,align='left',valign='top')
s['canvas'] += [
 text('kicker','ILLUSTRATIVE PROPOSAL • DECISION CONTEXT',36,24,888,16,10,True,blue),
 text('claim','One observable end-to-end journey is the right first investment',36,47,888,58,24,True),
 text('context','SYNTHETIC BASELINE: 18,000 cases/month • 22% manual handoffs • 14% incomplete intake • definitions and sample coverage need validation',36,112,888,25,10.5),
]
columns=[
 ('understand','OWNERSHIP IS THE FIRST CONSTRAINT',
  'A portal replacement would leave manual exceptions, duplicate case state, and difficult audit reconstruction in place. Four teams currently exchange status in spreadsheets across three downstream systems.',
  'Establish one case identifier, accountable owners, and explicit escalation paths. Keep alternate intake routes available while discovery confirms where rework and decision delays occur.',
  'Evidence to collect',
  'A sampled baseline, journey and ownership map, and interface/dependency register. The assumed rates above are not an independently validated business case.'),
 ('prove','PROVE A COMPLETE, BOUNDED JOURNEY',
  'Select one thin release that connects reliable intake, a visible service queue, a supported integration, and an auditable case history. Include exceptions, access controls, and operational telemetry in the first proof.',
  'Existing processing systems retain transaction ownership. Versioned interfaces and event contracts connect the journey; real-time updates and platform consolidation are not assumed.',
  'Evidence to produce',
  'A prioritized backlog, agreed acceptance measures, test results, control findings, and working runbooks. Additional journeys and integrations require separate scope decisions.'),
 ('decide','EXPAND WHEN THE EVIDENCE SUPPORTS IT',
  'Authorize four weeks of discovery and first-release definition, subject to access readiness and agreement on acceptance measures. Use the results to confirm the release scope, dependencies, and delivery sequence.',
  'Pilot readiness requires accepted test evidence, resolved critical control findings, executable runbooks, and a rehearsed rollback. Weak evidence or changed conditions trigger a scope adjustment.',
  'Decision to make',
  'Proceed, hold, or adjust at explicit steering gates. The 20-week sequence is illustrative; named staffing, allocations, availability, commercial terms, and dates remain to be agreed.'),
]
cells=[]
for i,(id,heading,a,b,label,evidence) in enumerate(columns):
 for row,blocks,bg in [
  (0,[dict(id='heading',text=heading,font_face='Arial',font_size_pt=12,bold=True,foreground='#FFFFFF',align='left',min_height_pt=31)],navy),
  (1,[dict(id='mechanism',text=a,font_face='Arial',font_size_pt=10.5,foreground=navy,align='left',min_height_pt=83),dict(id='boundary',text=b,font_face='Arial',font_size_pt=10.5,foreground=navy,align='left',gap_before_pt=7,min_height_pt=73)],pale),
  (2,[dict(id='evidence-label',text=label,font_face='Arial',font_size_pt=10.5,bold=True,foreground=navy,align='left',min_height_pt=14),dict(id='evidence',text=evidence,font_face='Arial',font_size_pt=10,foreground=navy,align='left',gap_before_pt=3,min_height_pt=56)],'#FFFFFF')]:
  cells.append(dict(id=f'{id}-{row}',row=row,column=i,blocks=blocks,padding=dict(top=8,right=10,bottom=7,left=10),background=bg))
s['layouts']=[dict(id='argument-panels',bounds=dict(x=36,y=145,width=888,height=317),columns=[dict(fixed_pt=282)]*3,rows=[dict(fixed_pt=46),dict(fixed_pt=178),dict(fixed_pt=93)],column_gap_pt=21,cells=cells)]
s['canvas'].append(text('qualification','ILLUSTRATIVE SCENARIO • Candidate targets and schedule are planning hypotheses, not performance or delivery commitments.',36,478,888,18,10,False))
s['accents'][0].update(target='claim',phrase='end-to-end journey')
OUT.write_text(json.dumps(dict(schema='pptxgengo.compose-spec.v1',slides=[s]),indent=2,ensure_ascii=False)+'\n')
print(OUT.relative_to(ROOT))
