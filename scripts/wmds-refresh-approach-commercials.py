#!/usr/bin/env python3
"""Build paired reference contracts for 19 added Approach/Commercials templates.

All alternate facts and prices are illustrative synthetic examples. The closed
catalog determines exact slots, arrays and navigation; pinned source stays intact.
"""
import argparse
import copy
import json
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
GROUP = ROOT / 'samples/wmds-refresh-slice3-20261002/work/approach-commercials'
BUNDLE = ROOT / 'library/wm-design-system/v2'
ENGINE = 'wmds-go-foundation.v2'
TITLES = {
 'pricing/capacity-split': 'One team, one monthly run-rate of $180K',
 'pricing/options-nav': 'Choose the engagement that fits regional rollout',
 'roadmap/staggered-phases-nav': 'Nine months from build to Lakeview ownership',
 'runbook/cutover-timeline': 'Four decision gates govern the [[regional]] cutover',
 'runbook/escalation': 'Four levels route each [[risk]] to a clear decision',
 'runbook/escalation-flow': 'Each missed threshold moves a [[risk]] up one level',
 'runbook/escalation-split': 'Four levels route each risk to a clear decision',
 'runbook/go-no-go': 'Wave 2 opens only when every criterion is [[green]]',
 'runbook/overview': 'One runbook governs every [[wave 2]] cutover decision',
 'runbook/roles': 'Six roles own every [[wave 2]] decision, with deputies',
 'runbook/roles-raci': 'Every regional step has one [[accountable]] owner',
 'runbook/step-detail': 'Step 1 freezes [[source postings]] before data moves',
 'runbook/step-detail-data': 'Tick every [[reconciliation]] before wave 2 loads',
 'runbook/step-detail-gate': 'Tick every criterion before wave 2 opens the [[ledger]]',
 'runbook/step-list': 'Steps 1 to 4 secure the source and [[move]] balances',
 'runbook/step-list-split': 'Steps 1 to 4 secure the source and [[move]] balances',
 'runbook/step-table': 'Eight steps move wave 2 from freeze to [[go-live]]',
 'runbook/step-table-nav': 'Eight steps move wave 2 from freeze to [[go-live]]',
 'scope/service-value-nav': 'Extend regional scope as adoption grows',
}
# Retain headings and terminology; replace substantive actions, evidence,
# personnel, conditions, dates, commercial assumptions and numeric examples.
COPY = {
 'Move close and consolidation to the new ledger over one weekend, with a tested way back at every step.':
 'Move regional close to the new ledger over one weekend, retaining a verified recovery path at each step.',
 'General ledger and subledgers': 'Regional ledger and subledgers',
 'Bank and intercompany interfaces': 'Treasury and intercompany feeds',
 'Excludes planning and reporting': 'Excludes forecasting and analytics',
 'Dress rehearsal signed off': 'Wave 2 rehearsal approved',
 'Rollback plan approved': 'Recovery plan signed off',
 'Day 1 postings reconciled': 'First-day entries reconciled',
 'Hypercare handover accepted': 'Service handover accepted',
 'Freeze legacy postings': 'Freeze source postings',
 'Stop every posting to the legacy ledger so the extract is one stable cut of the books.':
 'Suspend source postings so every regional balance comes from the same approved extract.',
 'Review the evidence for every criterion and decide whether the ledger opens to users.':
 'Review regional evidence against each gate and authorize opening the ledger to wave 2 users.',
 'Load the extracted open balances by entity and prove they match the legacy books.':
 'Load regional open balances by entity and reconcile each total to the approved source extract.',
 'Announce the freeze on the bridge': 'Confirm the freeze on the bridge',
 'Close and post open batches': 'Approve and post open batches',
 'Lock subledgers and the calendar': 'Lock regional posting periods',
 'Run the subledger lock report': 'Verify the posting-lock report',
 'Publish the freeze notice': 'Send the regional freeze notice',
 'No manual journals after freeze': 'No journals after the cutover lock',
 'Exceptions need the Cutover lead': 'Exceptions need the Release lead',
 'Late items are queued, not posted': 'Queue late items for first-day review',
 'Record the freeze time on the log': 'Log the freeze time and approver',
 'Lock report shows no open batches': 'Lock report confirms every batch closed',
 'Controller confirms the freeze on the log': 'Controller approves the lock in the log',
 'Cutover lead releases step 2': 'Release lead authorizes step 2',
 'Close open batches': 'Approve open batches',
 'Lock subledgers': 'Lock posting periods',
 'Take the final backup': 'Take a recovery snapshot',
 'Snapshot the legacy ledger': 'Snapshot the source ledger',
 'Verify checksums': 'Validate snapshot checksums',
 'Extract open balances': 'Extract regional balances',
 'Run the extract jobs': 'Run approved extract jobs',
 'Tie to the trial balance': 'Match the approved trial balance',
 'Log variances': 'Record regional variances',
 'Load the new ledger': 'Load the target ledger',
 'Load by entity': 'Load regional entities',
 'Check counts and totals': 'Reconcile counts and totals',
 'Park rejected rows': 'Quarantine rejected rows',
 'Load by entity, parents last': 'Load entities; parent groups last',
 'Run the count and total checks': 'Run regional count and total checks',
 'Log each variance': 'Record each variance and owner',
 'Every check inside tolerance': 'Each regional check within tolerance',
 'Variances explained and signed': 'Variances documented and approved',
 'Data lead releases step 5': 'Data lead authorizes step 5',
 'Controller opens the ledger': 'Controller releases the ledger',
 'Support lead sends the notice': 'Service lead sends the notice',
 'Log the decision and time': 'Record decision, approver and time',
 'Cutover lead invokes rollback': 'Release lead starts recovery',
 'Escalate to level 3': 'Page the level 3 decision owners',
 'Reset the date at steering': 'Agree a new date with steering',
 'Dress rehearsal passed with no open defects': 'Regional rehearsal passed; no defects open',
 'Balances tie to the legacy trial balance': 'Balances match the source trial balance',
 'Interfaces tested end to end': 'Regional feeds tested end to end',
 'Rollback rehearsed within the window': 'Recovery rehearsed inside the window',
 'Users trained and access confirmed': 'Regional users trained and access ready',
 'Support bridge staffed for 36 hours': 'Service bridge staffed for 36 hours',
 'Audit sign-off on migration controls': 'Audit approves regional migration controls',
 'Rehearsal log, run 3': 'Rehearsal log, run 4',
 'Reconciliation report': 'Regional tie-out report',
 'Interface test pack': 'Regional feed test pack',
 'Timed rollback record': 'Timed recovery record',
 'Training roster, 94%': 'Training roster, 98%',
 'Rota signed by leads': 'Service rota signed by leads',
 'Control walkthrough memo': 'Regional control review memo',
 'Balances tie to legacy': 'Balances match the source',
 'Interfaces pass end to end': 'Regional feeds pass end to end',
 'Rollback fits the window': 'Recovery fits the window',
 'Support bridge staffed': 'Service bridge staffed',
 'Signed rota': 'Approved service rota',
 'Row count by entity': 'Record count by region',
 'Debit and credit totals': 'Regional debit and credit totals',
 'Open items by entity': 'Open items by region',
 'Intercompany balances': 'Regional intercompany balances',
 '412,806': '386,204', '$1.284bn': '$1.106bn', '9,418': '8,206', '$38.2m': '$32.6m',
 'Sign the rollback plan': 'Approve the recovery plan',
 'Open the support bridge': 'Open the regional service bridge',
 'Lock legacy; take backup': 'Lock source; take snapshot',
 'Extract, load and reconcile': 'Extract and reconcile each region',
 'Switch interfaces; test posting': 'Switch feeds; validate posting',
 'Leads sign off balances': 'Owners approve regional balances',
 'Watch posting and match rates': 'Monitor posting and exception rates',
 'Triage defects every 2 hours': 'Review defects every 2 hours',
 'Runs the bridge; calls go or no-go': 'Runs the bridge; authorizes release',
 'Signs balances; opens the ledger': 'Approves balances; releases the ledger',
 'Runs loads and interface switch': 'Runs loads and regional feed switch',
 'User comms and hypercare': 'Regional user comms and hypercare',
 'The runbook and the bridge': 'The regional runbook and bridge',
 'Balances and ledger opening': 'Regional balances and ledger release',
 'Loads, jobs and environments': 'Regional loads and environments',
 'Interface switch and testing': 'Regional feed switch and testing',
 'Extracts and reconciliation': 'Regional extracts and tie-out',
 'Calls go or no-go; invokes rollback': 'Authorizes release; starts recovery',
 'Pauses a load; restores a snapshot': 'Stops a load; restores its snapshot',
 'Repoints interfaces to legacy': 'Returns feeds to source endpoints',
 'Stops a load on a variance': 'Holds a load on a regional variance',
 'Sends notices; triages defects': 'Notifies regions; triages defects',
 'Close diagnostic': 'Regional close review',
 'Target design and roadmap': 'Regional design and roadmap',
 'Wave 1 build': 'Wave 2 deployment',
 'Diagnose only': 'Assess the region',
 'Diagnose and build': 'Assess and deploy',
 'Embedded pod': 'Regional delivery pod',
 '$288K': '$240K', '$720K + $210K/mo': '$600K + $180K/mo', '$240K/mo': '$210K/mo',
 '$210K / mo': '$180K / mo', '$210K': '$180K', '$840K': '$720K',
 'Expenses capped at 8% of fees': 'Expenses capped at 6% of fees',
 '4 weeks': '5 weeks', '8 weeks fixed, then 16 weeks capacity': '10 weeks fixed, then 14 weeks capacity',
 'Rolling, 3-month minimum': 'Rolling, 4-month minimum',
 'Foundation': 'Regional foundation',
 'Architecture approved': 'Regional design approved',
 'Scale and early adoption': 'Regional scale and adoption',
 'Three early adopters': 'Four regional adopters',
 'Architecture and guardrails': 'Regional design controls',
 'Targeted release': 'Regional release',
 'Launch': 'Regional launch',
 'Onboarding and handover': 'Wave 2 onboarding',
 'Ongoing project and product management': 'Regional project and product management',
 'Early-adopter expansion': 'Regional adopter expansion',
 'Assess and select three more adopter teams': 'Assess and select four regional teams',
 'Workload-specific design and implementation': 'Regional workload design and delivery',
 'Provisioning and operational validation': 'Provision and validate regional operations',
 'Workload hypercare and enhancements': 'Regional hypercare and improvements',
 'Twice the workloads onboarded': 'Four more regional teams onboarded',
 'Higher value per workload': 'More regional value per workload',
 'Faster user adoption': 'Faster regional user adoption',
 'Change management ownership': 'Regional change ownership',
 'Lead end-to-end change for platform rollout': 'Lead change across the regional rollout',
 'Readiness assessment and champion program': 'Regional readiness and champion network',
 'Reusable onboarding and adoption playbook': 'Reusable regional adoption playbook',
 'Program communications and socialization': 'Regional communications and feedback',
 'Stickier adoption': 'Sustained regional adoption',
 'Less ongoing user support': 'Lower regional support demand',
 'Faster onboarding of future teams': 'Faster onboarding of later regions',
 'Cloud engineering surge': 'Regional engineering surge',
 'Continued cloud architecture and engineering': 'Regional cloud architecture and engineering',
 'Post-deployment environment optimization': 'Optimize regional environments after launch',
 'Ad hoc performance and scaling troubleshooting': 'Resolve regional performance and scale issues',
 'More capacity for workload growth': 'Capacity for regional workload growth',
 'Faster cost optimization after launch': 'Earlier regional cost optimization',
}

def replacement(value):
    if not isinstance(value, str):
        return value
    value = COPY.get(value, value)
    for old, new in [('Northfield', 'Lakeview'), ('Cutover lead', 'Release lead'),
                     ('cutover lead', 'release lead'), ('Support lead', 'Service lead'),
                     ('Program director', 'Delivery director'), ('program director', 'delivery director'),
                     ('Finance IT', 'Regional IT'), ('Shared services', 'Regional ops'),
                     ('Enterprise architecture', 'Platform office'), ('legacy', 'source'),
                     ('Fri ', 'Sat '), ('Sat 00:', 'Sun 00:'), ('Sat 02:', 'Sun 02:'),
                     ('Sat 03:', 'Sun 03:'), ('Sat 06:', 'Sun 06:')]:
        value = value.replace(old, new)
    return value

def run(cli, *args):
    completed = subprocess.run([str(cli), *map(str, args)], cwd=ROOT, capture_output=True, text=True)
    if completed.returncode:
        raise RuntimeError(completed.stderr or completed.stdout)
    return completed.stdout

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--cli', type=Path, default=ROOT/'pptxdesign')
    parser.add_argument('--out', type=Path, default=GROUP)
    opts = parser.parse_args()
    out = opts.out.resolve()
    out.mkdir(parents=True, exist_ok=True)
    audit = json.loads((ROOT/'planning/wm-design-contracts/v2/source-update-2026-10-02.json').read_text())
    keys = [t['key'] for t in audit['templates'] if t['change']=='added' and t['family'] in ('approach','commercials')]
    catalog = json.loads(run(opts.cli,'library-catalog','--bundle',BUNDLE,'--engine',ENGINE,'--template-keys',','.join(keys)))
    defs = {t['key']:t for t in catalog}
    content = {'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':[]}
    changed = {}
    for index, key in enumerate(keys,1):
        definition=defs[key]
        values={'slots':{s['name']:s['synthetic_source_example'] for s in definition['slots']},
                'keys':{a['name']:[f'source-{n+1:03d}' for n in range(a['count'])] for a in definition.get('arrays', [])}}
        if definition.get('nav'):
            values['nav']=copy.deepcopy(definition['nav']['synthetic_source_example'])
        base={'id':f'approach-commercials-{index:02d}-source','template':key,'content_kind':'synthetic_example','values':values}
        alt=copy.deepcopy(base)
        alt['id']=f'approach-commercials-{index:02d}-alternate'
        altvalues=alt['values']
        for name,value in altvalues['slots'].items():
            altvalues['slots'][name]=replacement(value)
        altvalues['slots']['title']=TITLES[key]
        for name,array in altvalues['keys'].items():
            altvalues['keys'][name]=[f'regional-{n+1:03d}' for n in range(len(array))]
        if definition.get('nav'):
            altvalues['nav']={'items':[{'key':'context','label':'Context'},{'key':'prepare','label':'Prepare'},
                                     {'key':'release','label':'Release'},{'key':'handover','label':'Handover'}],
                             'active':'release' if definition['family']=='approach' else 'prepare'}
        # Complete a partially checked wave 2 reconciliation with real caller
        # values, including intentional blanks in the still-incomplete rows.
        if key=='runbook/step-detail-data':
            for n in (1,2): altvalues['slots'][f'node08.rows.item{n:02d}.d']=True
            altvalues['slots']['node08.rows.item01.n']='386,204'
            altvalues['slots']['node08.rows.item02.n']='$1.106bn'
        if key=='runbook/step-detail-gate':
            for n in (1,2,4): altvalues['slots'][f'node08.rows.item{n:02d}.g']=True
        if key=='runbook/go-no-go':
            altvalues['slots']['node01.rows.item04.g']=True
            altvalues['slots']['node01.rows.item04.s']='on'
            altvalues['slots']['node01.rows.item07.s']='risk'
        if key=='pricing/capacity-split':
            altvalues['slots']['node02.rows.item04.alloc']=0.75
            altvalues['slots']['node02.rows.item07.alloc']=0.75
        changed[key]=[name for name,value in altvalues['slots'].items() if value!=values['slots'][name]]
        content['slides'].extend([base,alt])
    (out/'bound-content.json').write_text(json.dumps(content,indent=2)+'\n')
    (out/'changed-slots.json').write_text(json.dumps(changed,indent=2)+'\n')
    receipts=[]
    with tempfile.TemporaryDirectory(prefix='wmds-approach-commercials-') as temporary:
        temporary=Path(temporary)
        sweep=temporary/'source-sweep'
        run(opts.cli,'library-bound-sweep','--bundle',BUNDLE,'--engine',ENGINE,
            '--template-keys',','.join(keys),'--out',sweep)
        import shutil
        durable=out/'source-sweep'
        if durable.exists(): shutil.rmtree(durable)
        shutil.copytree(sweep,durable)
        source_receipts=json.loads((sweep/'generation-receipts.json').read_text())
        if len(source_receipts)!=len(keys) or any(r['status']=='generation_failed' for r in source_receipts):
            raise SystemExit('Source sweep failed; retained in source-sweep/generation-receipts.json.')
        for index,key in enumerate(keys,1):
            pair=copy.deepcopy(content)
            pair['slides']=content['slides'][2*(index-1):2*index]
            spec=temporary/f'pair-{index:02d}.json'
            spec.write_text(json.dumps(pair,indent=2)+'\n')
            target=temporary/f'generated-{index:02d}'
            try:
                output=run(opts.cli,'template','--bundle',BUNDLE,'--engine',ENGINE,'--spec',spec,'--out',target)
                compiled=json.loads((target/'compiled-document.json').read_text())
                if index==1: combined={**compiled,'slides':[]}
                combined['slides'].extend(compiled['slides'])
                receipt={'template':key,'status':'generated_native_review_pending','slide_count':2,
                         'changed_alternate_slots':changed[key],'source_revision':audit['source_commit']}
                # Durable individual packet aids root review and diagnosis.
                import shutil
                destination=out/f'pair-{index:02d}'
                if destination.exists(): shutil.rmtree(destination)
                shutil.copytree(target,destination)
            except Exception as error:
                receipt={'template':key,'status':'generation_failed','error':str(error)}
            receipts.append(receipt)
            print(key+': '+receipt['status']+((': '+receipt['error']) if 'error' in receipt else ''))
        (out/'generation-receipts.json').write_text(json.dumps(receipts,indent=2)+'\n')
        if any(r['status']=='generation_failed' for r in receipts):
            raise SystemExit('Generation failures retained in receipts; no complete combined deck emitted.')
        (out/'combined.foundation.json').write_text(json.dumps(combined,indent=2)+'\n')
        target=temporary/'combined'
        run(opts.cli,'build','--bundle',BUNDLE,'--engine',ENGINE,'--spec',out/'combined.foundation.json','--out',target)
        for name in ('foundation.json','layout-report.json','reference.pptx'):
            candidate=target/name
            if candidate.exists():
                import shutil
                shutil.copy2(candidate,out/name)
        print('Completed paired 38-slide packet. Native visual review pending.')

if __name__=='__main__': main()
