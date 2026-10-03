#!/usr/bin/env python3
"""Create 42 source/alternate pairs for the complete v2 Team/Solution families.
Reuses reviewed slice3 additions and explicit body copy for remaining designs.
All claims and identities are synthetic layout specimens. No tests/native UI.
"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
BUNDLE = ROOT / 'library/wm-design-system/v2'
ENGINE = 'wmds-go-foundation.v2'
REVIEWED = ROOT / 'samples/wmds-refresh-slice3-20261002/work/team-solution/bound-content.json'
INVENTORY = ROOT / 'planning/wm-design-contracts/v2/source-update-2026-10-02.json'
# These bodies share reviewed source layouts and corresponding closed slot names.
REVIEWED_SIBLINGS = {
    'architecture/layers': 'architecture/layers-nav',
    'bios/leadership-specialists': 'bios/leadership-specialists-nav',
    'bios/three': 'bios/three-nav',
    'readiness/six-criteria': 'readiness/six-criteria-split',
    'roles/by-phase': 'roles/by-phase-nav',
    'team/org-chart': 'team/org-chart-nav',
    'team/org-roles': 'team/org-roles-nav',
    'team/pods': 'team/pods-nav',
    'team/pods-pairs': 'team/pods-pairs-nav',
    'team/roster': 'team/roster-nav',
}
# Every value below is a deliberate per-template content substitution. Geometry,
# typography, structural indices, asset IDs and semantic edges stay fixed.
COPY = {
    'architecture/layer-map': {
        'node03.text':'BI, developer, sharing and reporting experiences',
        'node09.text':'Reusable ingestion, interoperability and operations',
        'node17.rows.item03.cells.item02':'Intake',
        'node17.rows.item04.cells.item03':'Access checks',
    },
    'architecture/nested': {
        'node01.bullets.item02':'Bank transaction feeds',
        'node03.label':'Lakeview cloud tenant',
        'node06.text':'Validation engine',
        'node08.text':'Close dashboard',
    },
    'architecture/product': {
        'node01.bullets.item02.text':'Specs, guides, wiki',
        'node07.bullets.item02.text':'Issue and delivery boards',
        'node17.text':'15 months',
        'node19.text':'12+',
    },
    'architecture/reference': {
        'node02.items.item02':'Code promotion reviewed',
        'node04.items.item03':'Reviewed at each program meeting',
        'node29.text':'Observability',
        'node33.text':'One evidence model',
    },
    'bio-full/portrait-list': {
        'node02.text':'Jamie Chen',
        'node04.body':'Jamie leads the diagnostic and designs rules, routing and the review calendar for wave 1.',
        'node07.items.item01':'Rules engines for 4 clients',
        'node07.items.item02':'Cut one close from 12 days to 7',
    },
    'bio-full/portrait-quote': {
        'node02.text':'Alex Morgan',
        'node05.body':'Alex sets direction, chairs the steering committee and is accountable for the five-day close.',
        'node06.body':'Twenty-five years in health-system finance, most spent redesigning the close.',
        'node08.items.item01':'25+ close transformations',
        'node08.items.item02':'12-entity finance redesign',
    },
    'context/build-sustain': {
        'node01.body.item01.bullets.item03.text':'and versioned artifacts.',
        'node05.text':'Retrieve rules, process context and open questions for a bounded design, build or cutover task.',
        'node11.text':'Connect changes and incidents to documented intent, architecture and tests for impact analysis.',
    },
    'context/three-zones': {
        'node04.text':'Source code, data and controls, with confidence.',
        'node08.text':'Approved outcomes, design decisions and open questions.',
        'node39.text':'Traceable stories and test cases',
    },
    'governance/stack': {
        'node01.tiers.item01.decisions.item01':'Budget, scope and gates',
        'node01.tiers.item02.decisions.item01':'Plan and dependencies',
        'node01.tiers.item03.decisions.item01':'Design and acceptance',
        'node02.items.item02.text':'Lakeview',
    },
    'pathways/two-lanes': {
        'node01.duration':'Weeks 36–43',
        'node04.items.item02':'SLA measured on every request',
        'node04.items.item05':'Exceptions reviewed monthly',
        'node14.text':'Design review',
    },
    'patterns/three-rows': {
        'node05.items.item01':'Data movement controls',
        'node05.items.item02':'Semantic layer ownership',
        'node10.items.item02':'Metadata and lineage checks',
        'node15.items.item03':'Release standards',
    },
    'process/swimlane': {
        'node01.steps.item02.text':'Validate entry rules',
        'node01.steps.item04.text':'Correct and repost',
        'node01.painNote':'Routing delay raised in interviews',
    },
    'workstreams/five-with-risks': {
        'node04.items.item01':'Name the priority workloads',
        'node10.items.item02':'Agree placement guardrails',
        'node28.items.item02':'Track benefit against plan',
        'node35.body.item01.p':'Costs grow without guardrails',
        'node37.body.item01.p':'Benefits lack named owners',
    },
}
# Keep alternate identities consistent with the explicitly substituted bodies.
# Source headers remain unchanged. These are specimen corrections, not resizing.
HEADER_COPY = {
    'architecture/nested': 'Everything runs in one governed Lakeview tenant',
    'bio-full/portrait-list': 'Jamie Chen designs the new close',
    'bio-full/portrait-quote': 'Alex Morgan leads the engagement',
    'readiness/six-criteria': 'The platform is generally available when Lakeview can run it as one owned service',
    'team/pods': 'Three pods work beside four Lakeview roles',
}

def write(path, value):
    path.write_text(json.dumps(value,indent=2)+'\n')

def run(cli, command, *args):
    p=subprocess.run([str(cli),command,'--bundle',str(BUNDLE),'--engine',ENGINE,*map(str,args)],
                     cwd=ROOT,text=True,capture_output=True)
    return {'returncode':p.returncode,'stdout':p.stdout,'stderr':p.stderr}

def changed_values(source, reviewed, added):
    key=source['template']
    if key in added:
        original,alternate=reviewed[key]
        # Keep reviewed source and body copy, navigation and registered photo IDs.
        return copy.deepcopy(original),copy.deepcopy(alternate),'slice3_reviewed_added_pair'
    alternate=copy.deepcopy(source)
    alternate['id']+='-alternate'
    alternate['content_kind']='synthetic_example'
    slots=alternate['values']['slots']
    if key in REVIEWED_SIBLINGS:
        prior=reviewed[REVIEWED_SIBLINGS[key]][1]
        prior_slots=prior['values']['slots']
        # Corresponding BODY slots only. Source copy and geometry are retained;
        # named alternate header corrections below keep identity copy consistent.
        overrides={name:value for name,value in prior_slots.items()
                   if name.startswith('node') and name in slots and value!=slots[name]}
        origin='reviewed_sibling_body:'+REVIEWED_SIBLINGS[key]
    else:
        overrides=COPY[key]
        origin='explicit_per_template_body_copy'
    for name,value in overrides.items():
        if name not in slots:
            raise ValueError(f'{key}: unknown explicit slot {name}')
        slots[name]=value
    if key in HEADER_COPY:
        slots['title']=HEADER_COPY[key]
    for name,keys in alternate['values'].get('keys',{}).items():
        alternate['values']['keys'][name]=[f'integrated-{i+1:02}' for i in range(len(keys))]
    return source,alternate,origin

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--cli',type=Path,required=True)
    p.add_argument('--out',type=Path,required=True,help='NEW directory')
    a=p.parse_args()
    if a.out.exists():p.error('output directory must not already exist')
    a.out.mkdir(parents=True)
    inventory=json.loads(INVENTORY.read_text()); changes={t['key']:t['change'] for t in inventory['templates']}
    catalog_result=run(a.cli,'library-catalog')
    if catalog_result['returncode']:raise RuntimeError(catalog_result['stderr'])
    catalog=json.loads(catalog_result['stdout'])
    templates=[t for t in catalog if t['family'] in ('team','solution')]
    if len(templates)!=42:raise ValueError(f'expected42 got{len(templates)}')
    write(a.out/'catalog.json',templates)
    added={t['key'] for t in templates if changes[t['key']]=='added'}
    reviewed={}
    for s in json.loads(REVIEWED.read_text())['slides']:
        reviewed.setdefault(s['template'],[None,None])[int(s['id'].endswith('-alternate'))]=s
    receipts=[];all_foundation=[];all_content=[];foundation_header=None
    for t in templates:
        key=t['key'];folder=a.out/key.replace('/','--');folder.mkdir()
        receipt={'template':key,'family':t['family'],'change':changes[key],
                 'content_kinds':['synthetic_example','synthetic_example'],
                 'source_commit':inventory['source_commit'],'native_visual_review':'pending'}
        if key in added:
            source=copy.deepcopy(reviewed[key][0])
        else:
            result=run(a.cli,'library-reference','--template-keys',key,'--out',folder/'source')
            receipt['source_generation']=result
            if result['returncode']:
                receipt['status']='source_generation_failed';receipts.append(receipt);write(folder/'generation-receipt.json',receipt)
                print('FAIL source',key,result['stderr'].strip(),flush=True);continue
            source=json.loads((folder/'source/template-content.json').read_text())['slides'][0]
        source,alternate,origin=changed_values(source,reviewed,added)
        source['id']='integrated-'+key.replace('/','--')+'-source'
        alternate['id']='integrated-'+key.replace('/','--')+'-alternate'
        body_changes={name:{'source':source['values']['slots'][name],'alternate':value}
                      for name,value in alternate['values']['slots'].items()
                      if name.startswith('node') and value!=source['values']['slots'][name]}
        if len(body_changes)<2:raise ValueError(f'{key}: at least2 substantive body changes required')
        receipt['specimen_origin']=origin;receipt['changed_slots']=body_changes
        receipt['changed_header_slots']={name:{'source':source['values']['slots'][name],'alternate':value}
                      for name,value in alternate['values']['slots'].items()
                      if not name.startswith('node') and value!=source['values']['slots'][name]}
        receipt['array_keys']={'source':source['values'].get('keys',{}),'alternate':alternate['values'].get('keys',{})}
        receipt['navigation']={'source':source['values'].get('nav'),'alternate':alternate['values'].get('nav')}
        packet={'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':[source,alternate]}
        write(folder/'bound-content.json',packet)
        result=run(a.cli,'template','--spec',folder/'bound-content.json','--out',folder/'paired')
        receipt['paired_generation']=result
        if result['returncode']:
            receipt['status']='paired_generation_failed';print('FAIL paired',key,result['stderr'].strip(),flush=True)
        else:
            doc=json.loads((folder/'paired/compiled-document.json').read_text());foundation_header=doc
            write(folder/'combined.foundation.json',doc)
            all_foundation.extend(doc['slides']);all_content.extend(packet['slides'])
            receipt['status']='generated_native_review_pending';receipt['slide_count']=2
            receipt['pptx_sha256']=hashlib.sha256((folder/'paired/bound-templates.pptx').read_bytes()).hexdigest()
            print('PASS',key,flush=True)
        receipts.append(receipt);write(folder/'generation-receipt.json',receipt)
    write(a.out/'generation-receipts.json',receipts)
    write(a.out/'bound-content.json',{'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':all_content})
    if foundation_header:write(a.out/'combined.foundation.json',dict(foundation_header,slides=all_foundation))
    summary={'templates':len(templates),'generated':len(all_content)//2,'slides':len(all_foundation),
             'reviewed_added_pairs_reused':19,'remaining_designs':23,'tests_added_or_run':False,'native_review':'pending'}
    write(a.out/'summary.json',summary);print(json.dumps(summary))
    return int(len(all_foundation)!=84)

if __name__=='__main__':sys.exit(main())
