#!/usr/bin/env python3
"""Build full-family paired, synthetic, closed-content WMDS specimens.

Reuses reviewed slice3 content, with explicit per-key overrides for older designs.
No generic text substitution, truncation, scaling or font shrinking is performed.
"""
import argparse
import copy
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT=Path(__file__).resolve().parents[1]
BUNDLE=ROOT/'library/wm-design-system/v2'
ENGINE='wmds-go-foundation.v2'
OUT=ROOT/'samples/wmds-refresh-slice4-20261002/work/approach-openers-commercials'
FAMILIES=('approach','commercials','openers')
# Exact caller-owned slot changes for designs without a prior alternate.
OVERRIDES={
 'agenda/index': {'node02.text':'Regional rollout decisions',
                  'node04.items.item01.title':'Regional close priorities',
                  'node04.items.item02.title':'Release operating model',
                  'node04.items.item03.title':'Regional delivery plan'},
 'agenda/sessions': {'title':'Agenda for the regional design day',
                     'node07.items.item01':'Finance: bring regional close risks',
                     'node07.items.item02':'Data: confirm regional source readiness',
                     'node08.text':'One regional scope with named decision owners',
                     'node11.items.item01':'Finance: approve regional acceptance rules',
                     'node12.text':'Approved regional rules and a tested data sample',
                     'node25.text':'Outcome: one regional scope, a verified data sample and a named owner for each decision.'},
 'closing/tagline': {'node03.text':'Kick off regional wave 2 in February',
                     'node04.text':'Regional lead · Finance Transformation',
                     'node05.text':'regional-example@westmonroe.com'},
 'cover/grid': {'node02.text':'Illustrative proposal · Lakeview Health',
                'node03.text':'A six-day close that [[pays]] for itself',
                'node04.text':'Regional record-to-report redesign for 12 units',
                'node05.stat':'13 → 6'},
 'cover/photo': {'node02.text':'Illustrative proposal for Lakeview Health',
                 'node03.text':'Close regional books in [[six days]]',
                 'node05.text':'West Monroe · Regional Finance',
                 'node07.text':'Regional delivery lead · illustrative',
                 'node08.text':'Regional process lead · illustrative',
                 'node09.text':'delivery-example@westmonroe.com',
                 'node10.text':'process-example@westmonroe.com'},
 'divider/full-photo': {'node03.text':'04','node04.text':'How we release wave 2'},
 'divider/inverse': {'node02.text':'Section 06 · Delivery','node03.text':'Who will [[lead the release]]'},
 'divider/light': {'node01.text':'05','node02.text':'Regional value case','node03.text':'What the region [[gains]]'},
 'divider/panel-edge': {'node04.text':'02','node05.text':'Regional scope and decisions'},
 'divider/panel-photo': {'node02.text':'[[03]]','node03.text':'Regional release perspective'},
 'key-message/statement': {'node01.text':'Regional release principle',
                           'node02.text':'Review moves to the [[start]] of regional close',
                           'node03.text':'Regional exceptions reach owners on day one, so day six is for approval rather than discovery.'},
 'narrative/lead-in': {'title':'Three causes explain regional close delay',
                      'source.text':'Illustrative regional scenario; synthetic figures for layout review.',
                      'node01.text':'The regional close takes 13 days. Seven are spent waiting for reviews and reconciliations.',
                      'node02.items.item01.text':'Regional balance-sheet accounts are reconciled after close rather than during the month.',
                      'node02.items.item02.text':'Twelve regional entities match balances through email and spreadsheets.',
                      'node02.items.item03.text':'Regional controllers first review most entries on day ten.'},
 'quote/inverse': {'node01.text':'Illustrative regional perspective',
                   'node02.by':'Regional controller · synthetic example',
                   'node02.text':'We need regional exceptions routed to owners early, so the last close day is for approval.'},
 'quote/light': {'node01.text':'Illustrative regional perspective',
                 'node02.by':'Regional controller · synthetic example',
                 'node02.text':'We need regional exceptions routed to owners early, so the last close day is for approval.'},
 'cutover/wave-matrix': {'title':'Regional cutover uses capability slices; people own every release decision',
                        'node02.text':'Identify viable regional slices, map shared dependencies and verify recovery options.',
                        'node04.text':'Map regional data and draft controls, acceptance checks, notices and support plans.',
                        'node06.text':'Track regional outcomes, route exceptions and trace failures to their dependencies.',
                        'node08.rows.item01.label':'Regional close and consolidation'},
 'phase-detail/rail-cohort': {'node03.text':'Regional Platform Foundation',
                             'node24.text':'Three regional workloads prove core patterns',
                             'node25.text':'Choose regional pilots to test each architecture pattern and document reusable operating controls.',
                             'node27.body.item01.p':'Build regional analytics on governed warehouse data.',
                             'node31.body.item01.p':'Rebuild regional dashboards on curated lakehouse data.',
                             'node35.body.item01.p':'Deliver regional ingestion, curation and BI for one use case.'},
 'phase-detail/rail-matrix': {'node03.text':'Regional Scale and Adoption',
                             'node09.text':'Prove that Lakeview can assess, onboard and support regional workloads through one repeatable process.',
                             'node13.rows.item01.w':'Check regional goals, readiness and dependencies',
                             'node13.rows.item02.e':'Approved regional design and delivery backlog'},
 'phase-detail/rail-table': {'node03.text':'Regional Platform Foundation',
                            'node09.text':'Validate regional architecture, access controls and onboarding patterns for the first finance workloads.',
                            'node11.items.item04':'Three regional pilots selected',
                            'node12.rows.item01.a.item01':'Define regional architecture and capacity',
                            'node12.rows.item02.a.item02':'Set regional health and cost telemetry',
                            'node16.body.item01.bullets.item01':'Source access is ready'},
 'phase-detail/team-handoff': {'title':'Phase 3 hands regional ownership to Lakeview',
                              'node04.body.item01.bullets.item02':'Lakeview on call',
                              'node05.series.item02.name':'Lakeview effort',
                              'node06.text':'West Monroe reduces its driving role as Lakeview takes ownership. After week 30, regional staff lead more delivery work than West Monroe.',
                              'node12.text':'Lakeview team',
                              'node13.text':'Approving regional standards and overseeing the next adopter cohort.',
                              'node14.text':'Regional staff learn support routines and assume delivery ownership.'},
 'phase-gate/evidence-matrix': {'title':'Regional adopter evidence decides readiness for general availability',
                               'node03.items.item01':'Regional patterns proven',
                               'node10.items.item03':'Regional runbooks tested',
                               'node17.items.item03':'Regional reviews done',
                               'node26.text':'Support accepted by Lakeview',
                               'node31.items.item03':'Regional value proven'},
 'phases/activities-outcomes': {'title':'Our [[three-phase]] approach establishes a Lakeview-owned regional service',
                               'node05.items.item01':'Define regional architecture',
                               'node05.items.item05':'Choose regional adopter workloads',
                               'node12.items.item04':'Onboard regional adopters',
                               'node19.items.item02':'Hand ownership to Lakeview',
                               'node21.items.item01':'Regional platform ready to scale'},
 'phases/four': {'title':'Four phases move regional close from [[13 days to 6]]',
                 'node01.phases.item01.activities.item01':'Regional kickoff and interviews',
                 'node01.phases.item02.objective':'Know where the 13 days go.',
                 'node01.phases.item03.objective':'A six-day regional close, designed.',
                 'node01.phases.item04.deliverables.item01':'$3.6M regional value case'},
 'phases/summary-bands': {'title':'Three phases define the regional release roadmap',
                         'node03.body.item01.bullets.item02':'Interview regional finance leads',
                         'node04.body.item01.bullets.item01':'Regional baseline and benchmarks',
                         'node07.body.item01.bullets.item02':'Set regional savings and value targets',
                         'node11.body.item01.bullets.item04':'Name regional owners and dependencies',
                         'node12.body.item01.bullets.item02':'Regional initiative owner register'},
 'plan/gantt': {'title':'Eight weeks to an approved regional roadmap',
                'node01.groups.item01.lanes.item01.items.item01.label':'Regional close mapping',
                'node01.groups.item01.lanes.item02.items.item01.label':'Regional design sprints',
                'node01.groups.item02.lanes.item01.sub':'12 regional units',
                'node01.groups.item02.lanes.item01.items.item01.label':'Regional systems',
                'node01.groups.item03.lanes.item01.items.item02.label':'Regional working sessions'},
 'pricing/capacity': {'title':'One regional team, one monthly run-rate of $180K',
                      'node01.lines.item01.item02':'$180K / mo','node01.total':'$720K',
                      'node01.terms.item03':'Expenses capped at 6% of fees',
                      'node02.rows.item01.alloc':0.5,'node02.runRate.value':'$180K'},
 'pricing/fixed-fee': {'title':'A fixed fee of $600K, paid as milestones are accepted',
                       'source.text':'Illustrative estimate; synthetic pricing excludes expenses.',
                       'node01.rows.item01.a':'$60K','node01.rows.item01.c':'$60K',
                       'node01.rows.item02.a':'$180K','node01.rows.item02.c':'$240K',
                       'node01.rows.item03.a':'$240K','node01.rows.item03.c':'$480K',
                       'node01.rows.item04.a':'$120K','node01.rows.item04.c':'$600K',
                       'node01.rows.item05.a':'$600K','node02.total':'$600K',
                       'node02.lines.item01.item02':'$240K','node02.lines.item02.item02':'$360K',
                       'node02.terms.item02':'Expenses capped at 6% of fees'},
}
# Equivalent body-slot contracts reuse already-reviewed explicit content, with
# source geometry still selected by each template key. This is not string editing.
EQUIVALENT={
 'agenda/schedule':'agenda/schedule-nav',
 'key-message/stat':'key-message/stat-nav',
 'roadmap/staggered-phases':'roadmap/staggered-phases-nav',
 'pricing/options':'pricing/options-nav',
 'scope/service-value':'scope/service-value-nav',
}

def invoke(cli,*args):
    run=subprocess.run([str(cli),*map(str,args)],cwd=ROOT,capture_output=True,text=True)
    if run.returncode: raise RuntimeError(run.stderr or run.stdout)
    return run.stdout

def examples(definition):
    values={'slots':{s['name']:s['synthetic_source_example'] for s in definition.get('slots',[])},
            'keys':{a['name']:[f'original-{i+1:03d}' for i in range(a['count'])] for a in definition.get('arrays',[])}}
    if definition.get('nav'): values['nav']=copy.deepcopy(definition['nav']['synthetic_source_example'])
    return values

def reused_alternates():
    values={}
    for family in ('approach-commercials','argument-openers'):
        path=ROOT/f'samples/wmds-refresh-slice3-20261002/work/{family}/bound-content.json'
        for slide in json.loads(path.read_text())['slides']:
            if 'alternate' in slide['id']: values[slide['template']]=slide['values']
    return values

def flatten(value,prefix=''):
    if isinstance(value,dict):
        result={}
        for name,item in value.items(): result.update(flatten(item,f'{prefix}.{name}' if prefix else name))
        return result
    if isinstance(value,list):
        result={}
        for i,item in enumerate(value): result.update(flatten(item,f'{prefix}[{i}]'))
        return result
    return {prefix:value}

def fill_thumbnails(slide):
    """Explicit illustrative native compositions, never missing-content fallback."""
    template=slide['template_binding']['template']
    output=[]
    for original in slide['nodes']:
        output.append(original)
        raw=original.get('scene',{}).get('node',{})
        if raw.get('type')!='thumbnail' or raw.get('photo') or raw.get('src'): continue
        kind=raw.get('kind','text'); raw['kind']='blank'
        resolution='wmds.integrated-approach-illustrative-miniature.v2'
        original['scene'].setdefault('resolutions',[]).append(resolution)
        x,y,w,h=(raw[n] for n in ('x','y','w','h')); ix,iy,iw=x+6,y+6,w-12
        def add(name,value):
            output.append({'id':'preview-'+original['id']+'-'+name,'kind':'scene',
                'rect':{'x_pt':0,'y_pt':0,'width_pt':0,'height_pt':0},
                'scene':{'node':value,'source_pointer':'/specimens/'+template+'/'+name,'resolutions':[resolution]}})
        def text(name,copy,ty):
            add(name,{'type':'text','x':ix+3,'y':ty,'w':iw-6,'style':'source','ink':'primary','text':copy})
        text('heading','PLATFORM FLOW' if kind=='diagram' else 'RELEASE PLAN',iy)
        if kind=='diagram':
            row=(h-12-15-6)/3
            for n,copy in enumerate(('SOURCE','CURATE','SERVE')):
                top=iy+15+n*(row+3)
                add('flow-'+str(n),{'type':'block','x':ix,'y':top,'w':iw,'h':row,'surface':'subtle'})
                text('label-'+str(n),copy,top+(row-12)/2)
        else:
            for n,copy in enumerate(('1 · Define scope','2 · Pilot controls','3 · Release')):
                text('phase-'+str(n),copy,iy+15+15*n)
    slide['nodes']=output
    return any(n['id'].startswith('preview-') for n in output)

def main():
    ap=argparse.ArgumentParser()
    ap.add_argument('--cli',type=Path,default=ROOT/'pptxdesign')
    ap.add_argument('--out',type=Path,default=OUT)
    args=ap.parse_args(); out=args.out.resolve(); out.mkdir(parents=True,exist_ok=True)
    catalog=json.loads(invoke(args.cli,'library-catalog','--bundle',BUNDLE,'--engine',ENGINE,'--include-deprecated'))
    selected=[t for t in catalog if t['family'] in FAMILIES]
    previous=reused_alternates()
    typed=json.loads((ROOT/'samples/wmds-templates-20261002/final/template-content.json').read_text())['slides']
    content={'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':[]}
    changes=[]
    for index,definition in enumerate(selected,1):
        key=definition['key']; origin='explicit_per_template_alternate'
        if key in ('cards/3','cards/4'):
            pair=[copy.deepcopy(s) for s in typed if s['template']==key]
            source,alt=pair[:2]
            origin='previously_reviewed_typed_card_pair'
        else:
            source={'template':key,'content_kind':'synthetic_example','values':examples(definition)}
            alt=copy.deepcopy(source)
            equivalent=EQUIVALENT.get(key,key)
            if equivalent in previous:
                origin='reviewed_slice3_alternate' if equivalent==key else f'reviewed_equivalent:{equivalent}'
                alt['values']=copy.deepcopy(previous[equivalent])
                if not definition.get('nav'): alt['values'].pop('nav',None)
            else:
                if key not in OVERRIDES: raise ValueError(f'Missing explicit alternate: {key}')
                for name,value in OVERRIDES[key].items():
                    if name not in alt['values']['slots']: raise ValueError(f'Unknown override: {key}/{name}')
                    alt['values']['slots'][name]=value
            for name,array in alt['values']['keys'].items():
                alt['values']['keys'][name]=[f'regional-{i+1:03d}' for i in range(len(array))]
        source['id']=f'aoc-{index:03d}-source'; alt['id']=f'aoc-{index:03d}-alternate'
        source['content_kind']=alt['content_kind']='synthetic_example'
        a,b=flatten(source['values']),flatten(alt['values'])
        changed=[{'slot':n,'source':a.get(n),'alternate':v} for n,v in b.items() if a.get(n)!=v and '.keys.' not in n and not n.startswith('keys.')]
        body=[c for c in changed if c['slot'] not in ('slots.title','slots.eyebrow','slots.source.text','title','eyebrow','source') and not c['slot'].startswith('nav.')]
        if not body: raise ValueError(f'No substantive alternate body change: {key}')
        changes.append({'template':key,'family':definition['family'],'source_kind':'synthetic_example',
                        'alternate_kind':'synthetic_example','alternate_origin':origin,'changed_slots':changed,
                        'body_changed_slots':[c['slot'] for c in body],
                        'status':definition.get('status','active')})
        content['slides'].extend((source,alt))
    (out/'bound-content.json').write_text(json.dumps(content,indent=2)+'\n')
    (out/'changed-slot-receipts.json').write_text(json.dumps(changes,indent=2)+'\n')
    combined={'schema':'pptxgengo.wmds-foundation.v1','year':2026,'slides':[]}; receipts=[]
    with tempfile.TemporaryDirectory(prefix='wmds-integrated-aoc-') as tmp:
        tmp=Path(tmp)
        for index,definition in enumerate(selected,1):
            key=definition['key']; pair={**content,'slides':content['slides'][2*(index-1):2*index]}
            spec=tmp/f'pair-{index:03d}.json'; spec.write_text(json.dumps(pair,indent=2)+'\n')
            generated=tmp/f'pair-{index:03d}'
            try:
                invoke(args.cli,'template','--bundle',BUNDLE,'--engine',ENGINE,'--spec',spec,'--out',generated)
                compiled=json.loads((generated/'compiled-document.json').read_text())
                has_preview=any([fill_thumbnails(slide) for slide in compiled['slides']])
                (generated/'combined.foundation.json').write_text(json.dumps(compiled,indent=2)+'\n')
                if has_preview:
                    specimen=tmp/f'specimen-{index:03d}'
                    invoke(args.cli,'build','--bundle',BUNDLE,'--engine',ENGINE,
                        '--spec',generated/'combined.foundation.json','--out',specimen)
                    shutil.copy2(specimen/'reference.pptx',generated/'bound-templates.pptx')
                    shutil.copy2(specimen/'layout-report.json',generated/'layout-report.json')
                combined['slides'].extend(compiled['slides'])
                destination=out/f'pair-{index:03d}'
                if destination.exists(): shutil.rmtree(destination)
                shutil.copytree(generated,destination)
                receipt={'template':key,'family':definition['family'],'status':'generated_native_review_pending',
                         'slide_ids':[s['id'] for s in compiled['slides']],
                         'changed_body_slots':changes[index-1]['body_changed_slots']}
            except Exception as error:
                receipt={'template':key,'status':'generation_failed','error':str(error)}
            receipts.append(receipt); print(key+': '+receipt['status']+((': '+receipt['error']) if 'error' in receipt else ''),flush=True)
        (out/'generation-receipts.json').write_text(json.dumps(receipts,indent=2)+'\n')
        if len(combined['slides'])!=2*len(selected): raise SystemExit('Incomplete paired generation; failures retained.')
        (out/'combined.foundation.json').write_text(json.dumps(combined,indent=2)+'\n')
        generated=tmp/'combined'
        invoke(args.cli,'build','--bundle',BUNDLE,'--engine',ENGINE,'--spec',out/'combined.foundation.json','--out',generated)
        for name in ('reference.pptx','foundation.json','layout-report.json'):
            shutil.copy2(generated/name,out/name)
    print(f'Completed {len(selected)} templates / {len(combined["slides"])} paired slides. Native review pending.')

if __name__=='__main__': main()
