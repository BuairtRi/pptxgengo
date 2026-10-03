#!/usr/bin/env python3
"""Produce 69 source/changed-content pairs for integrated Proof/Evidence/Argument.

Reviewed additions are reused verbatim. Other alternates use explicit per-key
copy decisions; there is no generic replacement, geometry shrinking or fallback.
"""
import argparse, copy, hashlib, importlib.util, json, shutil, subprocess, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BUNDLE = ROOT/'library/wm-design-system/v2'
FAMILIES = ('proof','evidence','argument')
OVERLAYS = {
 'about/glance': {'node01.text':'We pair industry expertise with technology delivery to turn practical decisions into measurable value.', 'node03.body.item01.bullets.item02':'Operations and platforms'},
 'capability-map/layered-bands': {'node23.rows.item01.cells.item01':'Eligibility verification','node23.rows.item01.cells.item03':'Claims tracking'},
 'capability-table/icon-rows': {'node01.rows.item01.a':'Structure interviews and surface evidence gaps for expert validation.','node01.rows.item01.h':'Experts confirm the context and meaning.'},
 'case-studies/cards-quotes': {'node01.body.item01.p':'Mapped 1,300 reports to a new model, automated generation and review, and finished eight months early.','node02.body.item01.p':'Built a governed data platform for analytics and AI, with reusable access patterns and quality checks.'},
 'case-studies/three': {'node01.case.challenge':'Intercompany matching delayed a 15-day close.','node01.case.did.item01':'Automated account matching'},
 'case-studies/two': {'node01.case.challenge':'Intercompany matching delayed a 15-day close.','node02.case.did.item01':'Shared consolidation model'},
 'case-study/metrics-exhibits': {'node02.text':'A national payer wanted faster data use cases, but legacy platforms and manual handoffs limited its growth.','node04.text':'Built the business case and a governed multi-tenant platform, then defined access patterns and a staged retirement plan.'},
 'case-study/quote': {'node02.body':'Manual intercompany matching extended the close to 15 days.','node03.body':'Automated matching, posting controls and review from day one.'},
 'case-study/rail': {'node02.body':'Manual intercompany matching extended the close to 15 days.','node03.body':'Automated matching, posting controls and review from day one.'},
 'chart/column-full': {'node01.series.item01.values.item03':14,'node01.series.item01.values.item04':14},
 'chart/column-rail': {'node01.series.item02.values.item02':5,'node01.series.item02.values.item03':6},
 'comparison/us-versus-others': {'node01.text':'Our teams combine finance, data and change expertise with the engineering depth needed to deliver a working platform.','node02.text':'We pair with your team through architecture, controls, engineering and operations, transferring the capability to run the result.'},
 'confidential/notice': {'node03.body':'Share only with people evaluating this proposal.','node04.body':'Direct questions to your West Monroe engagement lead.'},
 'decision/buy-build-economics': {'node04.items.item01':'Requirements are standard','node08.items.item01':'Workflows differentiate you'},
 'deliverables/sample-grid': {'node02.text':'Close baselines, workflow maps and control gaps, reviewed with accountable owners.','node06.text':'Ranked opportunities, future-state handoffs and value cases reviewed with Finance.'},
 'flow/challenges-to-outcomes': {'node07.body.item01.p':'Posts payments at receipt','node08.body.item01.p':'Flags posting errors'},
 'frameworks/layer-table': {'node06.text':'Outcome-led planning; continuous discovery','node05.sub':'Why this matters'},
 'from-to/cards': {'node01.body.item01.bullets.item02':'Exceptions surface late','node03.body.item01.bullets.item02':'Exceptions assigned that day'},
 'from-to/rows': {'node01.rows.item02.item01':'Intercompany matched in spreadsheets','node01.rows.item02.item02':'Balances matched automatically each day'},
 'goal-steps/goal-band-three-steps': {'node03.text':'Choose one participant-owned opportunity that Product values and Engineering can advance in a demonstrable slice today.','node06.items.item01.body.item01.p':'Product brings up to three candidates with supporting evidence.'},
 'history/axis': {'node02.body':'Mapped the close in 14 entities and identified late review.','node03.body':'Introduced automated matching across 11 entities.'},
 'hub-spokes/eight-spokes': {'node01.body.item01.p':'Reusable pipelines bring varied sources into a common data model.','node03.body.item01.p':'Parameters define the source, timing and destination, with every run logged.'},
 'image-text/frame-right': {'node01.text':'One team owns the exception queue','node02.text':'Work is routed to available capacity. Controllers review the common queue each morning so entities can move together.'},
 'image-text/square-left': {'node03.text':'Controllers share one view of open reconciliations, unmatched balances and review queues across all 14 business units.','node04.items.item01.text':'Everyone can see status without a meeting.'},
 'method/accelerator-steps': {'node01.body.item04.bullets.item01':'Validated workflows, handoffs and exceptions','node01.body.item04.bullets.item02':'Prioritized simplification and automation moves'},
 'objective/numbered-and-panel': {'node02.items.item01.text':'Drafting, coding and testing move faster, enabling smaller slices with evidence at each review.','node02.items.item02.text':'Unclear intent and acceptance criteria create rework that spreads just as quickly through delivery.'},
 'offerings/three-panels': {'node24.items.item01':'Care operations modernization','node26.items.item02':'Patient experience design'},
 'outcomes/three-cases': {'node01.body.item01.p':'A critical close application was assessed and rebuilt in less than half the usual delivery time.','node10.body.item01.p':'A claims platform was modernized for six hospitals while billing operations continued.'},
 'pillars/four-why-matters': {'node01.body.item01.bullets.item01':'Work follows business value and delivery risk.','node05.body.item01.bullets.item03':'Rebuild only with purpose.'},
 'problem/stat-callout': {'node02.body.item01.p':'Legacy systems add cost and risk while slowing new products and developer productivity.','node03.body.item01.p':'Lift-and-shift copies existing constraints when the process and architecture need to change.'},
 'quadrant/strong': {'node02.body':'Posting rules reduce manual checks.','node03.body':'Earlier review prevents repeat errors.'},
 'quadrant/subtle': {'node02.body':'Three initiatives repay the investment in wave 1.','node03.body':'A sponsor must own the value and effort.'},
 'rfp-map/appendix': {'node01.rows.item01.how.item01':'Executive sponsor named in week 1','node01.rows.item03.how.item02':'Daily close-status summary'},
 'session/stepper-and-guidance': {'node02.steps.item01.text':'Facilitators agree the objective and decision criteria; teams bring real constraints and context.','node02.steps.item02.text':'Product and engineering work in focused tracks while facilitators manage the timeboxes.'},
 'stats/circled-headline': {'node04.body.item01.p':'Capacity released in the first two waves funds the later work. The illustrated investment is recovered in 14 months across the business units.','node04.label':'How value is realized'},
 'stats/four-metrics': {'node07.text':'Automated matching clears balances before period end. Review at posting routes exceptions while they can still be corrected, reducing the manual work that slows the close.','node08.text':'The illustrated changes release about 11,000 hours a year across 14 units and support sign-off on day five without adding headcount.'},
 'status/four-panel': {'node03.rows.item01.ent':'4','node03.rows.item02.ent':'10'},
 'takeaway-rail/metrics-rail': {'node05.text':'Automated matching clears balances throughout the month, reducing manual reconciliation. Review at posting catches errors while they are cheaper to fix. Together, the changes give controllers more time to analyze results and less time chasing close status.','node08.text':'A quarter of delay adds about four months to payback.'},
 'transformation/before-after': {'node01.rows.item02.item01':'Intercompany matched in spreadsheets','node01.rows.item02.item02':'Balances matched automatically each day'},
 'transformation/pain-to-theme': {'node07.text':'Posting rules flag errors while entries can still be corrected.','node12.text':'Daily matching routes exceptions to an accountable owner.'},
 'value/capacity-conversion': {'node03.items.item01.text':'Confirm volumes, effort, loaded cost and adoption.','node03.items.item02.text':'Measure touch time, cycle time and exceptions in use.'},
}


def load(path): return json.loads(Path(path).read_text())
def write(path,value):
    Path(path).parent.mkdir(parents=True,exist_ok=True)
    Path(path).write_text(json.dumps(value,indent=2,ensure_ascii=False)+'\n')


def source_caller(contract, baseline):
    """Reconstruct exact frozen content and preserve rendered source key identities."""
    overlay = {}
    for n in baseline['nodes']:
        overlay.update(n.get('scene', {}).get('keys', {}))
    values = {
        'slots': {slot['name']: copy.deepcopy(slot['synthetic_source_example']) for slot in contract['slots']},
        'keys': {a['name']: copy.deepcopy(overlay.get(a['source_pointer'],
                  [f'source-{i+1:03d}' for i in range(a['count'])])) for a in contract.get('arrays', [])}}
    if contract.get('nav'):
        values['nav'] = {'items': [{'key': x['id'], 'label': x['label']} for x in baseline['frame']['nav']],
                         'active': baseline['frame']['active']}
    return {'id': baseline['id'], 'template': contract['key'], 'content_kind': 'synthetic_example', 'values': values}


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--cli',default='/tmp/wmds-slice3-root-pptxdesign')
    parser.add_argument('--out',type=Path,default=ROOT/'samples/wmds-refresh-slice4-20261002/work/proof-evidence-argument')
    args=parser.parse_args()
    if args.out.exists(): raise SystemExit('Output directory must be new.')
    args.out.mkdir(parents=True)
    tmp=Path(tempfile.mkdtemp(prefix='wmds-integrated-pea-'))
    def run(route,*opts):
        return subprocess.run([args.cli,route,'--bundle',str(BUNDLE),'--engine','wmds-go-foundation.v2',*map(str,opts)],cwd=ROOT,capture_output=True,text=True)
    c=run('library-catalog','--include-deprecated')
    if c.returncode:raise SystemExit(c.stderr)
    catalog=[t for t in json.loads(c.stdout) if t['family'] in FAMILIES]
    keys=[t['key'] for t in catalog]
    source=run('library-sweep','--template-keys',','.join(keys),'--year',2026,'--out',tmp/'source')
    if source.returncode:raise SystemExit(source.stderr)
    source_receipts=load(tmp/'source/generation-receipts.json')
    write(args.out/'source-generation-receipts.json',source_receipts)
    if any(r['status']!='generated_native_review_pending' for r in source_receipts):
        raise SystemExit(source.stdout)
    source_doc=load(tmp/'source/compiled-document.json')
    reviewed={};reviewed_values={}
    for family in ['proof-evidence','argument-openers']:
        root=ROOT/'samples/wmds-refresh-slice3-20261002/work'/family
        for s in load(root/'combined.foundation.json')['slides']:
            key=s['template_binding']['template']
            reviewed.setdefault(key,[]).append(s)
        for s in load(root/'bound-content.json')['slides']:
            reviewed_values[s['template']]=s
    # Preview authoring is explicit specimen composition, separate from binding.
    spec=importlib.util.spec_from_file_location('proof_preview',ROOT/'scripts/wmds-refresh-proof-evidence.py')
    preview=importlib.util.module_from_spec(spec);spec.loader.exec_module(preview)
    combined=copy.deepcopy(source_doc);combined['slides']=[]
    bound={'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':[]}
    receipts=[]
    for index,t in enumerate(catalog):
        key=t['key'];per=args.out/key.replace('/','-');per.mkdir()
        baseline=copy.deepcopy(next(s for s in source_doc['slides'] if s['template_binding']['template']==key))
        baseline['id']='source-'+key.replace('/','-')
        receipt={'template':key,'family':t['family'],'lifecycle':t['status'],'replacement':t.get('replaced_by'),
            'source_kind':'synthetic_example','alternate_kind':'synthetic_example','native_review':'pending','tests_run':False}
        if key in reviewed and key in reviewed_values:
            pair=copy.deepcopy(reviewed[key]);
            if len(pair)!=2:raise ValueError(f'Expected reviewed pair for {key}')
            # Existing reviewed source and alternate carry their meaningful
            # preview amendments and caller copy exactly as accepted in slice3.
            baseline,alt=pair
            baseline['id']='source-'+key.replace('/','-');alt['id']='alternate-'+key.replace('/','-')
            caller=copy.deepcopy(reviewed_values[key]);caller['id']=alt['id']
            receipt['origin']='slice3_reviewed_pair'
        else:
            overlay=OVERLAYS.get(key)
            if overlay is None:raise ValueError(f'Missing explicit copy decisions for {key}')
            values={'slots':{s['name']:copy.deepcopy(s['synthetic_source_example']) for s in t['slots']},
                    'keys':{a['name']:[f'item-{i+1:02d}' for i in range(a['count'])] for a in t.get('arrays',[])}}
            if t.get('nav'):values['nav']=copy.deepcopy(t['nav']['synthetic_source_example'])
            for name,value in overlay.items():
                if name not in values['slots']:raise ValueError(f'Unknown {key} slot {name}')
                values['slots'][name]=value
            if 'source.text' in values['slots']:values['slots']['source.text']='Illustrative example; not client results.'
            # Explicit eyebrow indicator also marks slides with no source zone.
            if 'eyebrow' in values['slots']:values['slots']['eyebrow']='Illustrative · '+t['family'].title()
            caller={'id':'alternate-'+key.replace('/','-'),'template':key,'content_kind':'synthetic_example','values':values}
            write(per/'bound-content.json',{'schema':bound['schema'],'year':2026,'slides':[caller]})
            result=run('template','--spec',per/'bound-content.json','--out',tmp/f'bound-{index}')
            if result.returncode:
                receipt.update(generation='failed',error=result.stderr.strip());receipts.append(receipt);write(per/'generation-receipt.json',receipt);print(key,receipt['error']);continue
            alt=load(tmp/f'bound-{index}/compiled-document.json')['slides'][0]
            preview.fill_preview(baseline);preview.fill_preview(alt)
            receipt['origin']='slice4_explicit_caller_overlay'
        caller['content_kind']='synthetic_example'
        changed=[s['name'] for s in t['slots'] if caller['values']['slots'][s['name']]!=s['synthetic_source_example']]
        body=[s for s in changed if s.startswith('node') or s.startswith('z')]
        if not body:raise ValueError(f'{key} has no substantive changed body content')
        receipt.update(changed_slots=changed,changed_body_slots=body,slot_count=len(t['slots']),
             exact_array_counts={a['name']:a['count'] for a in t.get('arrays',[])},generation='pending')
        for specimen in (baseline, alt):
            specimen['template_binding']['slide_id']=specimen['id']
        receipt['compatibility_retained'] = t['status']=='deprecated'
        pair_doc=copy.deepcopy(source_doc);pair_doc['slides']=[baseline,alt]
        write(per/'combined.foundation.json',pair_doc)
        source_input=source_caller(t, baseline)
        write(per/'bound-content.json',{'schema':bound['schema'],'year':2026,'slides':[source_input,caller]})
        result=run('build','--spec',per/'combined.foundation.json','--out',tmp/f'pair-{index}')
        if result.returncode:receipt.update(generation='failed',error=result.stderr.strip())
        else:
            receipt['generation']='source_and_alternate_generated'
            for file in ['reference.pptx','layout-report.json']:shutil.copy2(tmp/f'pair-{index}'/file,per/file)
            combined['slides'] += pair_doc['slides'];bound['slides'].extend([source_input,caller])
        write(per/'generation-receipt.json',receipt);receipts.append(receipt)
        print(key,receipt['generation'])
    write(args.out/'combined.foundation.json',combined);write(args.out/'bound-content.json',bound)
    write(args.out/'generation-receipts.json',receipts);write(args.out/'contracts.json',catalog)
    if len(combined['slides'])!=2*len(catalog):raise SystemExit('Incomplete family reference; inspect generation receipts.')
    result=run('build','--spec',args.out/'combined.foundation.json','--out',tmp/'combined')
    if result.returncode:raise SystemExit(result.stderr)
    for file in ['reference.pptx','layout-report.json']:shutil.copy2(tmp/'combined'/file,args.out/file)
    write(args.out/'generation-summary.json',{'families':FAMILIES,'templates':len(catalog),'slides':len(combined['slides']),
        'source_passes':len(catalog),'alternate_passes':len(catalog),'reused_reviewed_pairs':sum(r['origin']=='slice3_reviewed_pair' for r in receipts),
        'native_review':'pending','tests_run':False,'pptx_sha256':hashlib.sha256((args.out/'reference.pptx').read_bytes()).hexdigest()})
    print(f'Generated {len(combined["slides"])} integrated paired specimens; native review pending.')

if __name__=='__main__':main()
