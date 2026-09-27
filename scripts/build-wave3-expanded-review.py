#!/usr/bin/env python3
"""Prepare Wave 3 visual-review inputs; this does not render or qualify them.

The proposal bundle must have been assembled by pptxlib from the current durable
values/contracts. Reject stale inputs instead of silently combining revisions.
"""
import argparse
import copy
import hashlib
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT/'library/diagram-components'

def read(path):
    return json.loads(Path(path).read_text())

def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()

def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False)+'\n')

def load_slide(path, identity):
    return copy.deepcopy(next(s for s in read(ROOT/path)['slides'] if s['id']==identity))

def claim(slide):
    return next(c for c in slide['canvas'] if c['id']=='claim')

def verified_proposal(bundle):
    config=ROOT/'library/proposal/assembly.json'
    trace=read(bundle/'assembly-trace.json')
    spec=read(bundle/'spec.json')
    if trace['config_sha256']!=sha(config) or trace['spec_sha256']!=sha(bundle/'spec.json'):
        raise ValueError('stale assembly config/spec; rebuild the proposal through pptxlib')
    if trace['narrative_sha256']!=sha(config.parent/trace['narrative_path']):
        raise ValueError('stale proposal narrative')
    for s in trace['slides']:
        if s['values_sha256']!=sha(config.parent/s['values_path']):
            raise ValueError('stale semantic values: '+s['slide_id'])
        c=ROOT/'library/contracts'/(s['contract_id'].split('/')[-1]+'.json')
        if s['selection']['contract_sha256']!=sha(c):
            raise ValueError('stale library contract: '+s['contract_id'])
        contract=read(c)
        if contract['composition']['spec_sha256']!=sha(ROOT/contract['composition']['spec_path']):
            raise ValueError('stale template: '+s['contract_id'])
    return spec,trace

def phrase_extensions():
    path='library/diagram-components/accent-review.json'
    originals=read(ROOT/path)['slides']
    added=[]
    checks=[]
    for mode in ('underline','highlight'):
        base=next(s for s in originals if s['id']==mode+'-source-copy')
        moved=copy.deepcopy(base);moved['id']=mode+'-position-shift'
        claim(moved)['bounds']['x']+=12;claim(moved)['bounds']['y']+=42
        moved['notes']+=' Only claim-frame origin moves +12pt x/+42pt y. Text, width, font, phrase selector and optical calibration stay unchanged.'
        added.append(moved)
        checks.append(dict(kind='translation',control=base['id'],variant=moved['id'],
                           target='claim',delta_pt=dict(x=12,y=42),
                           invariant='Native phrase fragments and accent visible bounds move by exactly the same delta; artwork hash, size and calibration are unchanged.'))
    repeated=load_slide(path,'highlight-source-copy');repeated['id']='highlight-repeated'
    claim(repeated)['text']='West Monroe frames the question. West Monroe tests the evidence.'
    claim(repeated)['bounds']['width']=650
    repeated['accents'][0]['occurrence']=2
    added.append(repeated)
    wrapped=load_slide(path,'highlight-source-copy');wrapped['id']='highlight-per-line'
    claim(wrapped)['text']='West Monroe';claim(wrapped)['bounds']['width']=135
    wrapped['accents'][0]['multiline']='per_line'
    added.append(wrapped)
    ranged=load_slide(path,'underline-source-copy');ranged['id']='underline-unicode-range'
    claim(ranged)['text']='Résumé: a pivotal moment; café follows.'
    claim(ranged)['bounds']['width']=700
    a=ranged['accents'][0];a['start_rune']=claim(ranged)['text'].index(a['phrase'])
    a['end_rune']=a['start_rune']+len(a['phrase'])
    added.append(ranged)
    rich=load_slide(path,'highlight-source-copy');rich['id']='highlight-rich-runs'
    t=claim(rich);t.pop('text');t.pop('font_size_pt');t.pop('font_face');t.pop('bold');t.pop('foreground')
    def run(identity,body,bold=False):
        return dict(id=identity,text=body,font_face='Arial',font_size_pt=24,bold=bold,foreground='#070154')
    t['paragraphs']=[dict(id='claim-line',align='left',runs=[run('lead','Work with '),run('brand','West Monroe',True),run('tail',' to test the evidence.')])]
    added.append(rich)
    manual=load_slide('library/diagram-components/accent-manual.json','ambiguous-manual-stage')
    multiline=copy.deepcopy(wrapped);multiline['id']='highlight-multiline-manual'
    multiline['accents'][0]['multiline']='manual'
    multiline['accents'][0]['staging']=dict(asset_bounds=dict(x=600,y=235,width=180,height=85),
        note_bounds=dict(x=600,y=350,width=284,height=110),
        note='Highlight the two-line phrase “West Monroe” near the upper left. Choose one mark per line or revise the emphasis treatment during visual review; the current slide is unfinished.')
    multiline['canvas']=[c for c in multiline['canvas'] if c['id']!='explanation']
    added.append(multiline)
    negative=copy.deepcopy(manual);negative['id']='ambiguous-without-staging'
    negative['accents'][0].pop('staging')
    return added,checks,negative

def motion_extensions(proposal):
    arrow=load_slide('library/diagram-components/arrow-review.json','single-arrow-0.75-+0')
    moved=copy.deepcopy(arrow);moved['id']='single-arrow-translated-anchors'
    changed=copy.deepcopy(arrow);changed['id']='single-arrow-target-moved'
    for s,ids,dx,dy in [(moved,{'source','target'},12,20),(changed,{'target'},20,8)]:
        for c in s['canvas']:
            if c['id'] in ids:c['bounds']['x']+=dx;c['bounds']['y']+=dy
        s['notes']+=' Endpoint-motion experiment: only named source/target text frames move. Arrow image coordinates are recalculated from the declared ports.'
    team=copy.deepcopy(next(s for s in proposal['slides'] if s['id']=='team-and-accountability'))
    team['id']='dense-team-control-highlight'
    accent=load_slide('library/diagram-components/accent-review.json','highlight-source-copy')['accents'][0]
    accent.update(id='control-emphasis',target='client-h-2',phrase='control decisions')
    team['accents']=[accent]
    team['notes']+=' Dense accent experiment: highlight the control-decision phrase in the client accountability column. Small-type optical calibration is a candidate, not an accepted style.'
    shifted=copy.deepcopy(team);shifted['id']='dense-team-highlight-body-shift'
    delta=dict(x=4,y=-6)
    fixed={'footer-band','footer-copy','wm-logo','page','page-number','kicker'}
    def move(bounds):
        bounds['x']+=delta['x'];bounds['y']+=delta['y']
    for c in shifted['canvas']:
        if c['id'] not in fixed:move(c['bounds'])
    for kind in ('roles','pods'):
        for c in shifted.get(kind,[]):move(c['bounds'])
    for c in shifted.get('layouts',[]):
        if not c.get('parent_id'):move(c['bounds'])
    if shifted.get('legend'):move(shifted['legend']['bounds'])
    shifted['notes']+=' Diagram-body origins shift +4pt x/-6pt y; title/footer stay fixed. Roles, pods, root layouts and relationship endpoints move together; the phrase accent follows its text.'
    checks=[dict(kind='translation',control=arrow['id'],variant=moved['id'],delta_pt=dict(x=12,y=20),
                 target='source,target,relationship',invariant='Both native endpoint frames and artwork visible endpoints translate equally; span/scale/rotation remain unchanged.'),
            dict(kind='endpoint_change',control=arrow['id'],variant=changed['id'],target='target',delta_pt=dict(x=20,y=8),
                 invariant='Source frame stays fixed, target frame moves by the declared delta, and arrow geometry recalculates within its candidate span/rotation/clearance contract.'),
            dict(kind='translation',control=team['id'],variant=shifted['id'],target='client-h-2,control-emphasis',delta_pt=delta,
                 invariant='All diagram-body objects/ports translate equally; fixed title/footer remain unchanged; native phrase and highlight translate together without new collision.')]
    return [moved,changed,team,shifted],checks

def manual_loop():
    source=next(a for a in read(OUT/'arrow-assets.json')['assets'] if a['id']=='wm-handdrawn-connecting-arrow')
    module_spec=importlib.util.spec_from_file_location('arrow_fixtures',ROOT/'scripts/build-wave3-arrow-spec.py')
    helpers=importlib.util.module_from_spec(module_spec);module_spec.loader.exec_module(helpers)
    (ROOT/'samples/visual-wave3/arrow-ink'/source['asset_sha256']).mkdir(parents=True,exist_ok=True)
    fallback,digest=helpers.derived_fallback(source)
    s=load_slide('library/diagram-components/arrow-review.json','single-arrow-0.75-+0')
    s['id']='connecting-loop-manual-only'
    s['notes']='Original WM return-loop artwork, pinned by its source bytes. No automatic span or rotation envelope is declared. Manual-only policy must stage it with a specific target note.'
    for c in s['canvas']:
        if c['id']=='experiment-title':c['text']='Return-loop artwork requires optical placement review'
        elif c['id']=='description':c['text']='The original return-loop asset is available on the slide. Its tight endpoint span and large curved path need a human placement decision; no automatic placement calibration is claimed.'
        elif c['id']=='source':c['text']='Decision ready';c['bounds']=dict(x=36,y=175,width=220,height=40)
        elif c['id']=='target':c['text']='Review evidence and release';c['bounds']=dict(x=320,y=175,width=220,height=40)
    art={k:copy.deepcopy(source[k]) for k in ('id','asset_path','asset_sha256','intrinsic_width','intrinsic_height','alpha_bounds','tail','tip')}
    art.update(fallback_asset_path=fallback,fallback_asset_sha256=digest,
               endpoint_provenance='Observed source-path endpoints in arrow-assets.json; no native placement calibration. Manual-only artwork, no automatic span/rotation envelope.')
    s['artwork_arrows']=[dict(id='relationship',manual_only=True,
        **{'from':dict(target='source',port='right',gap_pt=20),'to':dict(target='target',port='left',gap_pt=20)},
        artwork=art,clearance_pt=2,staging=dict(asset_bounds=dict(x=640,y=175,width=200,height=160),
          note_bounds=dict(x=640,y=350,width=284,height=90),
          note='Use this return loop to connect “Decision ready” with “Review evidence and release.” Position the large curve around the content and align its right-facing tip during visual review. Endpoint spacing is unqualified.'))]
    return s

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--proposal-bundle',type=Path,required=True)
    args=parser.parse_args()
    bundle=args.proposal_bundle.resolve()
    proposal,trace=verified_proposal(bundle)
    extensions,checks,negative=phrase_extensions()
    motion,motion_checks=motion_extensions(proposal)
    checks+=motion_checks
    loop=manual_loop()
    schema=proposal['schema']
    write(OUT/'accent-extended.json',dict(schema=schema,slides=extensions))
    write(OUT/'accent-ambiguous-negative.json',dict(schema=schema,slides=[negative]))
    write(OUT/'placement-motion.json',dict(schema=schema,slides=motion))
    write(OUT/'loop-manual.json',dict(schema=schema,slides=[loop]))
    pages=[];sources=[]
    def add(s,source,category,acceptance):
        if s['id'] in [p['id'] for p in pages]:raise ValueError('duplicate slide '+s['id'])
        s=copy.deepcopy(s)
        # The expanded review has its own pagination; content/geometry stay fixed.
        for c in s.get('canvas',[]):
            if c['id'] in ('page','page-number') and c['kind']=='text':c['text']=str(len(pages)+1)
        pages.append(s)
        sources.append(dict(position=len(pages),slide_id=s['id'],source=source,category=category,
                            expected_status='manual_required' if category=='manual' else 'native_review_required',
                            acceptance=acceptance))
    for identity in ['five-phase-sequence','discovery-detail','journey-workflow',
                     'team-and-accountability','roadmap-and-gates',
                     'architecture-and-controls','parallel-delivery-paths']:
        add(next(s for s in proposal['slides'] if s['id']==identity),
            dict(path=str((bundle/'spec.json').relative_to(ROOT)),sha256=trace['spec_sha256']),
            'changed_dense_pattern','Native text fit, all required copy, correct hierarchy, padding, date/role semantics and clear relationships.')
    fixtures=[('accent-review.json','underline-source-copy','accent'),
              ('accent-review.json','underline-per-line','accent'),
              ('accent-review.json','highlight-larger-wrap','accent'),
              ('arrow-review.json','single-arrow-0.75-+0','arrow'),
              ('arrow-review.json','dashed-arrow-1.25--15','arrow'),
              ('arrow-review.json','right-angle-arrow-1-+15','arrow'),
              ('accent-manual.json','ambiguous-manual-stage','manual'),
              ('arrow-manual.json','arrow-direction-manual','manual')]
    for file,identity,category in fixtures:
        path=OUT/file
        add(load_slide(str(path.relative_to(ROOT)),identity),dict(path=str(path.relative_to(ROOT)),sha256=sha(path)),category,
            'Inspect intended phrase/endpoint, visible ink clearance, layer order, optical appeal and explicit unfinished note where staged.')
    write(OUT/'expanded-review.json',dict(schema=schema,slides=pages))
    controls=[load_slide('library/diagram-components/review.json','uhg36-source-control'),
              *read(OUT/'art-translation.json')['slides']]
    write(OUT/'source-controls.json',dict(schema=schema,slides=controls))
    qualification=[*read(OUT/'accent-qualification.json')['slides'],*extensions,*motion,
                   *read(OUT/'arrow-qualification.json')['slides'],loop]
    write(OUT/'accent-arrow-qualification.json',dict(schema=schema,slides=qualification))
    write(OUT/'expanded-review-plan.json',dict(schema='pptxgengo.wave3-review-plan.v1',
        status='prepared_not_native_qualified',review_spec=dict(path='library/diagram-components/expanded-review.json',sha256=sha(OUT/'expanded-review.json')),
        pages=sources,comparison_checks=checks,
        supplemental=[dict(path='library/diagram-components/'+name,sha256=sha(OUT/name)) for name in
            ['source-controls.json','accent-arrow-qualification.json','dense-translation.json','overflow.json','dense-overflow.json','accent-ambiguous-negative.json']],
        coverage_notes=[
            'The seven changed dense pages include the original five difficult patterns plus substantive architecture and parallel paths.',
            'UHG36 is a bounded geometry/copy reconstruction; source controls do not assert whole-slide pixel identity.',
            'UHG14 source board internals remain raster; only whole-image placement/translation is tested.',
            'Arrow variants use three automatic candidate asset styles; the return loop uses an explicit manual-only policy without invented span limits.',
            'All native measurement, final verification, per-page rendered review and independent acceptance remain required.'],
        negative_expectations=[dict(path='accent-ambiguous-negative.json',stage='probe',contains='choose an explicit occurrence or range'),
            dict(path='overflow.json',stage='native build/fit-report',expect='process text overflow; no silent shrink'),
            dict(path='dense-overflow.json',stage='native build/fit-report',expect='architecture cell overflow; no silent shrink')]))
    print(f'Prepared {len(pages)} expanded review pages, {len(controls)} source/art controls, {len(qualification)} accent/arrow cases and explicit negative expectations.')

if __name__=='__main__':main()
