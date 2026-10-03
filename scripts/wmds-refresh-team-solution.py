#!/usr/bin/env python3
"""Generate source/changed-content pairs for the 19 added Team/Solution designs.
All copy is synthetic specimen content. Source bundle and v1 remain unchanged.
No tests or native automation are invoked. Output must be a new directory.
"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
INVENTORY = ROOT / 'planning/wm-design-contracts/v2/source-update-2026-10-02.json'
BUNDLE = ROOT / 'library/wm-design-system/v2'
ENGINE = 'wmds-go-foundation.v2'
REPLACEMENTS = {
    'Northfield': 'Lakeview',
    'Jordan ': 'Alex ', 'Jordan Ellis': 'Alex Morgan', 'Sam Ortiz': 'Jamie Chen', 'Priya Kapoor': 'Taylor Reed',
    'Priya Raman': 'Casey Patel', 'Sam Ibarra': 'Morgan Lee', 'Tomas Berg': 'Avery Cole',
    'Aisha Okafor': 'Riley Stone', 'Helen Brandt': 'Robin Hart', 'Ravi Menon': 'Drew Park',
    'Claire Dubois': 'Cameron Bell', 'Jonas Weber': 'Quinn James',
    'Dana Whitfield': 'Alex Carter', 'Kiran Mehta': 'Jamie Wells', 'Julia Ansel': 'Taylor Quinn',
    'Mike Odell': 'Casey Brooks', 'Jag Soni': 'Morgan Dale', 'Albert Reyes': 'Avery Lane',
    'Scott Alder': 'Riley Ford', 'Dan Calloway': 'Robin Shaw', 'Brett Marlow': 'Drew Kent',
    'Anneke Groot': 'Cameron West', 'Tanner Zell': 'Quinn Marsh', 'Greta Movsky': 'Jordan Finch',
    'Morgan Reyes': 'Blake Nolan',
    '30+': '25+', 'Thirty years': 'Twenty-five years', '14-entity': '12-entity',
    '3,000': '2,500', '40 teams': '35 teams', '12 enterprise': '10 enterprise',
    '3 FTE · 50%': '2 FTE · 60%', '0.5 FTE': '0.6 FTE', 'CFO · 0.1 FTE': 'CFO · 0.2 FTE',
    '50–75%': '40–60%', '(5%)': '(10%)', '2 × Accountants': '2 × Controllers',
    '2 × Data engineers': '3 × Data engineers',
    'Close cockpit and mobile approvals': 'Close dashboard and exception approvals',
    'Business rules and posting services': 'Workflow rules and validation services',
    'One governed path to close data': 'One governed path to journal data',
    'ERP, sub-ledgers and bank feeds': 'ERP, reconciliations and bank feeds',
    'Cloud landing zone, identity, network and storage': 'Cloud foundation, identity, network and recovery',
    'Each layer serves every close process, so one moves without a rebuild.':
        'Shared layers serve each close process, so changes stay local.',
    'Build once': 'Reuse by design', 'Shared controls': 'Common controls',
    'Observability': 'Monitoring', 'Audit trail': 'Evidence trail',
    '11% of lines miscoded': '9% of lines miscoded',
    'Six days waiting to approve': 'Five days waiting to approve',
    'Intelligent capture codes 95% of lines': 'Intelligent capture codes 96% of lines',
    'Cycle time 14 days to 5 days': 'Cycle time 12 days to 5 days',
    '1.5 days': '1.0 day', '2.0 days': '2.5 days', '8.0 days': '6.5 days',
    '12,400': '10,800', '11,900': '10,500', '11%': '9%', '9%': '8%',
    'Approval holds 57% of the 14-day cycle and four of nine handoffs: redesign it first.':
        'Approval holds 54% of the 12-day cycle and four of nine handoffs: redesign it first.',
    'Approval holds most of the 14-day invoice cycle': 'Approval holds most of the 12-day invoice cycle',
    'Capital planning': 'Investment planning', 'Workforce planning': 'Capacity planning',
    'ERP billing module': 'ERP receivables',
    'Cash positioning': 'Cash forecasting', 'Debt and investment': 'Debt management',
    'Posted liability': 'Approved liability', 'Accrual report': 'Reconciliation report',
    'ERP master data': 'Vendor master data', 'Budget owners': 'Cost-center owners',
    'Vendor onboarding and disputes': 'Vendor setup and contract disputes',
    'A vendor invoice arrives': 'An approved invoice arrives',
    'Payment is posted and reconciled': 'Payment clears and is reconciled',
    'Planning, coordination, dependencies, approvals and RAID.':
        'Plans delivery, tracks dependencies and manages approvals and RAID.',
    'Priorities, backlog, acceptance and operating-model calls.':
        'Owns priorities, backlog, acceptance and service decisions.',
    'Provide oversight and direction': 'Review outcomes and direction',
    'Make strategic decisions': 'Approve strategic decisions',
    'Own quality and escalation': 'Own quality and key escalations',
    'Own outcomes, scope, quality and client relationships.':
        'Own scope, quality and outcomes.',
    'Oversee pod execution and manage escalations.':
        'Review pod delivery and resolve cross-team escalations.',
    'Champion early adopters, review risks, remove blockers.':
        'Champion adoption, review risks and remove delivery blockers.',
    'Define target architecture, standards and interoperability.':
        'Define architecture and integration standards.',
    'Define the service catalog': 'Agree the service catalog',
    'Set the standards': 'Publish the standards',
    'The platform becomes an IT silo': 'The service becomes an IT silo',
    'Catalog goes stale within months': 'Catalog ownership is unclear',
    'Standards are ignored in practice': 'Standards lack named owners',
    'Exceptions outnumber patterns': 'Exceptions grow without review',
    'Costs rise faster than usage': 'Costs grow without guardrails',
    'Value is never shown to funders': 'Benefits lack named owners',
}
# These pairs exercise short bullets/body copy without increasing source capacities.
EXACT = {
    'process/sipoc': {'title': 'A SIPOC fixes invoice scope before detailed mapping'},
    'process/hierarchy': {'title': 'Wave one covers two priority finance processes'},
    'process/current-future': {'title': 'Redesign removes four invoice handoffs'},
    'workstreams/five-narrative': {'title': 'Five workstreams establish an owned platform service'},
    'readiness/six-criteria-split': {'title': 'Six criteria confirm launch readiness',
        'node07.items.item03': 'Approved patterns are reusable',
        'node13.items.item02': 'Alerts and runbooks are approved'},
}

def write(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')

def run(cli, command, *args):
    p = subprocess.run([str(cli), command, '--bundle', str(BUNDLE), '--engine', ENGINE, *map(str,args)],
                       cwd=ROOT, text=True, capture_output=True)
    return {'returncode': p.returncode, 'stdout': p.stdout, 'stderr': p.stderr}

def alternate(source, family):
    s = copy.deepcopy(source)
    s['id'] += '-alternate'
    slots = s['values']['slots']
    changed = []
    for name, value in list(slots.items()):
        if not isinstance(value, str) or name.endswith(('.photo','.icon')):
            continue
        new = value
        # Exact replacements are ordered first; then shorter lexical replacements.
        if value in REPLACEMENTS:
            new = REPLACEMENTS[value]
        else:
            for old, replacement in sorted(REPLACEMENTS.items(), key=lambda item: -len(item[0])):
                new = new.replace(old, replacement)
        if new != value:
            slots[name] = new
            changed.append(name)
    for name,value in EXACT.get(s['template'], {}).items():
        if name in slots and slots[name] != value:
            slots[name] = value
            changed.append(name)
    # Readiness guidance: explicitly change two operative criteria, not just headers.
    if s['template'] == 'readiness/six-criteria-split':
        body_names = [n for n in slots if n.startswith('node') and n.endswith('.body')]
        if len(body_names) >= 2:
            for name, value in zip(body_names, ['Owners approve the launch evidence.', 'Support can resolve priority incidents.']):
                slots[name] = value
                changed.append(name)
    # Align initials with each changed synthetic name without touching photos.
    for name,value in list(slots.items()):
        if name.endswith('.name') and name in changed:
            initials = name[:-4] + 'initials'
            if initials in slots:
                slots[initials] = ''.join(v[0] for v in value.split()[:2])
                changed.append(initials)
    for arr, values in s['values'].get('keys', {}).items():
        s['values']['keys'][arr] = [f'changed-{i+1:02d}' for i in range(len(values))]
    if 'nav' in s['values']:
        s['values']['nav'] = {'items': [
            {'key':'context','label':'Context'}, {'key':'solution','label':'Design'},
            {'key':'delivery','label':'Delivery'}, {'key':'team','label':'Team'}],
            'active':'team' if family == 'team' else 'solution'}
    body = sorted(set(n for n in changed if n.startswith('node')))
    if len(body) < 2:
        raise ValueError(f'{s["template"]}: fewer than two changed body slots: {body}')
    return s, sorted(set(changed))

def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--cli', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True, help='NEW output directory')
    a = p.parse_args()
    if a.out.exists():
        p.error('output directory must not already exist')
    a.out.mkdir(parents=True)
    inventory = json.loads(INVENTORY.read_text())
    templates = [t for t in inventory['templates'] if t['change']=='added' and t['family'] in ('team','solution')]
    assert len(templates)==19
    all_source, all_paired, all_foundation, receipts = [], [], [], []
    foundation_header = None
    for t in templates:
        d = a.out / t['key'].replace('/','--')
        d.mkdir()
        r = {'template': t['key'], 'family': t['family'], 'source_commit': inventory['source_commit'],
             'source': run(a.cli,'library-reference','--template-keys',t['key'],'--out',d/'source')}
        if r['source']['returncode']:
            r['status']='source_generation_failed'
            receipts.append(r); print('FAIL source',t['key'],r['source']['stderr'].strip()); continue
        content = json.loads((d/'source/template-content.json').read_text())
        source = content['slides'][0]
        alt, changed = alternate(source,t['family'])
        paired = dict(content, slides=[source,alt])
        write(d/'bound-content.json',paired)
        r['changed_slots']=changed
        r['paired']=run(a.cli,'template','--spec',d/'bound-content.json','--out',d/'paired')
        if r['paired']['returncode']:
            r['status']='changed_content_generation_failed'
            receipts.append(r); print('FAIL alternate',t['key'],r['paired']['stderr'].strip()); continue
        doc=json.loads((d/'paired/compiled-document.json').read_text())
        write(d/'combined.foundation.json',doc)
        all_source.append(source); all_paired.extend([source,alt]); all_foundation.extend(doc['slides'])
        foundation_header=doc
        r['status']='generated_native_review_pending'
        r['slide_count']=2
        r['powerpoint_verified']=False
        r['native_visual_reviewed']=False
        r['pptx_sha256']=hashlib.sha256((d/'paired/bound-templates.pptx').read_bytes()).hexdigest()
        receipts.append(r);write(d/'generation-receipt.json',r); print('PASS',t['key'])
    write(a.out/'generation-receipts.json',receipts)
    write(a.out/'bound-content.json',{'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':all_paired})
    if foundation_header:
        write(a.out/'combined.foundation.json',dict(foundation_header,slides=all_foundation))
    summary={'selected':len(templates),'generated':len(all_source),'slides':len(all_foundation),
             'tests_added_or_run':False,'native_review':'pending','families':['team','solution']}
    write(a.out/'summary.json',summary)
    print(json.dumps(summary))
    return 0 if len(all_source)==19 else 1

if __name__=='__main__':
    sys.exit(main())
