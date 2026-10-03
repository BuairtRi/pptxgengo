#!/usr/bin/env python3
"""Generate paired synthetic reference specimens for all added Proof/Evidence v2 templates.

Frozen bindings contain no specimen fallback. Blank thumbnail surfaces receive
explicit editable illustrations only in the composed reference documents.
"""
import argparse, copy, json, subprocess, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BUNDLE = ROOT / 'library/wm-design-system/v2'
INVENTORY = ROOT / 'planning/wm-design-contracts/v2/source-update-2026-10-02.json'


def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n')


def alternate(slot, template):
    v = copy.deepcopy(slot['synthetic_source_example'])
    name, pointer = slot['name'], slot['source_pointer']
    if name == 'source.text':
        return 'Illustrative only; not client results.'
    if isinstance(v, bool):
        return not v
    if isinstance(v, (int, float)):
        if pointer.endswith('/x') or pointer.endswith('/y'):
            return round(min(.94, max(.06, v + (.02 if pointer.endswith('/x') else -.02))), 3)
        if '/groups/' in pointer: return v
        if template.startswith('case-study/exhibit') and '.series.item01.values.' in name: return v - 1
        return v + 1
    if not isinstance(v, str):
        raise ValueError(f'unsupported slot {name}')
    # Preserve bounded status enums, WM assets and intentional empty sample cells.
    if v in ('on', 'risk', 'off', 'done', 'current', 'next', '') or v.startswith('photo-'):
        return v
    replacements = [('Northfield', 'Lakeview'), ('northfield', 'lakeview'),
                    ('15-day', '14-day'), ('15 days', '14 days'), ('from 15', 'from 14'),
                    ('6-day', '7-day'), ('6 days', '7 days'), ('under a week', 'under eight days'),
                    ('92%', '90%'), ('8%', '10%'), ('$2.1M', '$1.9M'), ('$2.2M', '$2.0M'),
                    ('38,000', '34,000'), ('Sep 2026', 'Oct 2026'), ('Aug 2026', 'Sep 2026'),
                    ('30 Sep', '7 Oct'), ('9 Oct', '12 Oct'), ('16 Oct', '19 Oct'),
                    ('30 Oct', '2 Nov'), ('13 Nov', '16 Nov'), ('day one', 'day two'),
                    ('day nine', 'day eight'), ('email', 'spreadsheets'),
                    ('five-day', 'six-day'), ('5-day', '6-day'), ('a daily cockpit', 'a shared cockpit'),
                    ('A daily view', 'A shared view'), ('Finance IT', 'Data lead'),
                    ('Controller', 'Finance lead'), ('Corporate controller', 'Group controller'),
                    ('PMO lead', 'PMO owner'), ('five entities', 'four entities'), ('days to 6', 'days to 7'), ('Day-one review', 'Day-two review'),
                    ('first day', 'second day'), ('Corporate closes in 6', 'Corporate closes in 7')]
    for old, new in replacements:
        v = v.replace(old, new)
    if template == 'chart/column-full-split':
        v = v.replace('Hospitals take 14 days.', 'Hospitals take 16 days.')
    if name == 'eyebrow':
        v = 'Illustrative · ' + ('Proof' if 'study' in v.lower() else 'Evidence')
    return v


def node(identity, raw, pointer):
    return {'id': identity, 'kind': 'scene',
            'rect': {'x_pt': 0, 'y_pt': 0, 'width_pt': 0, 'height_pt': 0},
            'scene': {'node': raw, 'source_pointer': pointer,
                      'resolutions': ['wmds.proof-evidence-illustrative-preview.v2']}}


def fill_preview(slide):
    """Add authored semantic miniatures to every placeholder thumbnail."""
    template = slide['template_binding']['template']
    alt = slide['id'].startswith('alternate-')
    if template == 'deliverables/annotated':
        fill_scorecard(slide, alt)
        return
    out = []
    for n in slide['nodes']:
        raw = n.get('scene', {}).get('node', {})
        out.append(n)
        if raw.get('type') != 'thumbnail' or raw.get('photo') or raw.get('src'):
            continue
        kind = raw.get('kind', 'text')
        raw['kind'] = 'blank'
        n['scene'].setdefault('resolutions', []).append('wmds.proof-evidence-illustrative-preview.v2')
        x, y, w, h = (raw[k] for k in ('x', 'y', 'w', 'h'))
        pad = 6 if w <= 126 else 12
        ix, iy, iw = x + pad, y + pad, w - 2 * pad
        base = 'preview-' + n['id']
        def add(key, val):
            out.append(node(base + '-' + key, val, '/specimens/' + template + '/' + key))
        def text(key, text, tx, ty, tw, style='label', ink='primary'):
            add(key, {'type':'text','x':tx,'y':ty,'w':tw,'style':style,'ink':ink,'text':text})
        def block(key, bx, by, bw, bh, surface):
            add(key, {'type':'block','x':bx,'y':by,'w':bw,'h':bh,'surface':surface})
        text('heading', {'table':'CLOSE','chart':'DAYS','diagram':'FLOW','text':'PLAN'}[kind], ix, iy, iw, 'label', 'emphasis')
        if kind == 'chart':
            label_space = 16 if w >= 162 else 0
            top, available = iy + 19 + label_space, h - pad * 2 - 23 - label_space
            col = iw / 3
            for i, days in enumerate([7, 5, 6] if alt else [6, 4, 5]):
                barh = available * days / 8
                block(f'bar-{i}', ix + i * col + 2, top + available - barh, col - 5, barh, 'strong' if i == 0 else 'subtle')
                if w >= 162:
                    text(f'value-{i}', str(days), ix + i*col, top+available-barh-16, col, 'small')
        elif kind == 'diagram':
            rh = min(20, max(12, (h - pad*2 - 23)/3))
            for i, title in enumerate(['POST','MATCH','REVIEW']):
                yy = iy + 20 + i*(rh+4)
                block(f'flow-{i}', ix, yy, iw, rh, 'subtle')
                text(f'label-{i}', title, ix+3, yy, iw-6, 'label')
        elif kind == 'table':
            available = h - pad*2 - 19
            rows = max(1, min(4, int(available//18)))
            labels = ['POST','MATCH','REVIEW','REPORT']
            for i in range(rows):
                yy = iy + 19 + i*18
                block(f'row-{i}', ix, yy, iw, 17, 'subtle' if i%2 == 0 else 'light')
                text(f'row-label-{i}', labels[i], ix+3, yy, iw-6, 'label')
        else:
            for i, title in enumerate(['1 · Scope','2 · Pilot','3 · Scale']):
                yy = iy+20+i*18
                if yy+14 <= y+h-pad:
                    text(f'phase-{i}', title, ix, yy, iw, 'small')
    slide['nodes'] = out


def fill_scorecard(slide, alt):
    content = []
    for n in slide['nodes']:
        raw = n['scene']['node']
        if n['id'] == 'node01': raw['kind'] = 'blank'
        elif n['id'] == 'z1': raw['h'] = 72
        elif n['id'] == 'z2': raw['h'] = 90
    def add(key, raw):
        content.append(node('preview-'+key, raw, '/specimens/annotated-scorecard/'+key))
    def text(key, copy, x, y, w, style='small', ink='primary'):
        add(key, {'type':'text','x':x,'y':y,'w':w,'style':style,'ink':ink,'text':copy})
    text('summary-label','Readiness / pilot decision',87,154,354,'label','emphasis')
    text('score','76 / 100' if alt else '72 / 100',87,170,144,'subhead','display')
    text('decision','Proceed with a limited pilot',249,173,192,'small','emphasis')
    text('target',f'Target: close the books in {6 if alt else 5} days',87,194,354)
    for y, key, heads in [(243,'steps',('Close step','Current','Target')),
                          (351,'actions',('Priority action','Owner','Due'))]:
        cols = [(87,192),(285,78),(369,72)]
        for j,(x,w) in enumerate(cols): text(f'{key}-head-{j}',heads[j],x,y,w,'label','emphasis')
        rows = [('Ledger reconciliation','3 days','2 days'),('Review and approvals','2 days','1 day'),
                ('Reporting and sign-off','2 days','3 days' if alt else '2 days')]
        if key == 'actions': rows = [('Resolve mapping gaps','Controller','Oct 09'),
            ('Confirm exception owners','Finance','Oct 12'),('Rehearse the pilot close','PMO','Oct 16')]
        for i,row in enumerate(rows):
            yy=y+17+i*18
            if i%2==0: add(f'{key}-fill-{i}',{'type':'block','x':84,'y':yy,'w':360,'h':18,'surface':'subtle'})
            for j,(x,w) in enumerate(cols): text(f'{key}-row-{i}-{j}',row[j],x,yy,w)
    slide['nodes'][1:1] = content


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--cli', default='/tmp/wmds-proof-evidence-pptxdesign')
    parser.add_argument('--out', type=Path, default=ROOT/'samples/wmds-refresh-slice3-20261002/work/proof-evidence')
    args = parser.parse_args()
    if args.out.exists(): raise SystemExit('output directory must be new')
    args.out.mkdir(parents=True)
    temporary = Path(tempfile.mkdtemp(prefix='wmds-proof-evidence-'))
    def run(route, *opts):
        result = subprocess.run([args.cli, route, '--bundle',str(BUNDLE),'--engine','wmds-go-foundation.v2', *map(str,opts)], cwd=ROOT, capture_output=True,text=True)
        return result
    inventory = json.loads(INVENTORY.read_text())
    keys = [t['key'] for t in inventory['templates'] if t['change']=='added' and t['family'] in ('proof','evidence')]
    catalog_run = run('library-catalog')
    if catalog_run.returncode: raise SystemExit(catalog_run.stderr)
    catalog = {t['key']:t for t in json.loads(catalog_run.stdout)}
    source_dir = temporary/'source'
    source_run = run('library-bound-sweep','--template-keys',','.join(keys),'--year',2026,'--out',source_dir)
    if source_run.returncode: raise SystemExit(source_run.stderr)
    source = json.loads((source_dir/'compiled-document.json').read_text())
    write(args.out/'source-generation-receipts.json',json.loads((source_dir/'generation-receipts.json').read_text()))
    combined = copy.deepcopy(source); combined['slides'] = []
    bound = {'schema':'pptxgengo.wmds-template-document.v1','year':2026,'slides':[]}
    receipts = []
    for index,key in enumerate(keys):
        contract = catalog[key]
        values = {'slots':{s['name']:alternate(s, key) for s in contract['slots']},
                  'keys':{a['name']:[f'bound-{i+1:02d}' for i in range(a['count'])] for a in contract.get('arrays',[])}}
        if key == 'case-study/what-we-did':
            # Retain account count and automated-matching meaning within the
            # source row's one-line description clearance above its separator.
            values['slots']['node05.body'] = 'The 12 busiest accounts match automatically.'
        if key == 'deliverables/annotated':
            values['slots'].update({
                'node05.callout.label':'1 · Readiness today',
                'node05.callout.text':'A headline score tied to the decision to launch a limited pilot.',
                'node06.callout.label':'2 · Close-step baseline',
                'node06.callout.text':'Each close step compared with the six-day target.',
                'node07.callout.label':'3 · Action ownership',
                'node07.callout.text':'The three priority gaps, with accountable owners and due dates.'})
        if key == 'deliverables/sample-grid-nav':
            replacements = {
                'Operating baseline and workflow findings':'Close baseline and workflow gaps',
                'Prioritized redesign portfolio':'Ranked improvement portfolio',
                'Mobilization and adoption plan':'Pilot readiness and adoption plan',
                'Value realization and governance':'Value tracking and leadership routines',
                'Workflow baselines, process maps, friction findings and capability gaps, owner-validated.':'Close baselines, workflow maps and control gaps, reviewed with accountable owners.',
                'Ranked redesign opportunities, future-state workflows and value cases sized with Finance.':'Ranked opportunities, future-state handoffs and value cases reviewed with Finance.',
                'Roadmap, integration readiness, controls, stage gates and adoption actions for each business unit.':'Pilot waves, control checks, stage gates and adoption actions for each business unit.'}
            for name, value in values['slots'].items():
                if value in replacements: values['slots'][name] = replacements[value]
        if key == 'deliverables/index':
            replacements = {'Operating baseline':'Workflow baseline', 'Redesign portfolio':'Opportunity portfolio',
                            'Adoption plan':'Mobilization plan', 'KPI scorecard':'Value scorecard',
                            'Finance leadership':'Finance directors', 'Steering committee':'Executive sponsors',
                            'Business unit leads':'Operations leads', 'CFO and controllers':'CFO and finance leads'}
            for name, value in values['slots'].items():
                if value in replacements: values['slots'][name] = replacements[value]
        if key == 'deliverables/walkthrough':
            for name, value in values['slots'].items():
                if value == 'Baseline': values['slots'][name] = 'Close baseline'
                elif value == 'Findings': values['slots'][name] = 'Priority gaps'
                elif value == 'Redesign': values['slots'][name] = 'Future flow'
                elif value == 'Roadmap': values['slots'][name] = 'Pilot roadmap'
        if contract.get('nav'):
            nav = copy.deepcopy(contract['nav']['synthetic_source_example'])
            for item in nav['items']:
                if item['label']=='Deliverables': item['label']='Outputs'
                elif item['label']=='Context': item['label']='Overview'
            values['nav']=nav
        s = {'id':'alternate-'+key.replace('/','-'),'template':key,'content_kind':'synthetic_example','values':values}
        bound['slides'].append(s)
        per = args.out/key.replace('/','-')
        write(per/'bound-content.json',{'schema':bound['schema'],'year':2026,'slides':[s]})
        bound_dir = temporary/f'bound-{index}'
        result = run('template','--spec',per/'bound-content.json','--out',bound_dir)
        receipt={'template':key,'binding_slot_count':len(contract['slots']),'array_counts':{a['name']:a['count'] for a in contract.get('arrays',[])},
                 'source_status':'generation_pending','alternate_status':'generation_failed','native_review':'pending','tests_run':False}
        source_slide = next(sl for sl in source['slides'] if sl['template_binding']['template']==key)
        source_slide = copy.deepcopy(source_slide); source_slide['id']='source-'+key.replace('/','-')
        fill_preview(source_slide)
        if result.returncode:
            receipt['error']=result.stderr.strip()
            pair=copy.deepcopy(source); pair['slides']=[source_slide]
        else:
            alt_doc=json.loads((bound_dir/'compiled-document.json').read_text())
            alt_slide=alt_doc['slides'][0];fill_preview(alt_slide)
            pair=copy.deepcopy(source);pair['slides']=[source_slide,alt_slide]
            receipt['alternate_status']='generated_native_review_pending'
            receipt['changed_body_slots']=[sl['name'] for sl in contract['slots'] if sl['name'].startswith('node') and values['slots'][sl['name']] != sl['synthetic_source_example']]
        write(per/'combined.foundation.json',pair)
        gen_dir=temporary/f'pair-{index}'
        gen=run('build','--spec',per/'combined.foundation.json','--out',gen_dir)
        if gen.returncode:
            receipt['pair_status']='generation_failed';receipt['pair_error']=gen.stderr.strip()
        else:
            receipt['source_status']='generated_native_review_pending';receipt['pair_status']='generated_native_review_pending'
            for filename in ('reference.pptx','layout-report.json'):
                (per/filename).write_bytes((gen_dir/filename).read_bytes())
            combined['slides'].extend(pair['slides'])
        write(per/'generation-receipt.json',receipt);receipts.append(receipt)
        print(key,receipt['pair_status'],receipt.get('error',''),receipt.get('pair_error',''))
    write(args.out/'bound-content.json',bound)
    write(args.out/'combined.foundation.json',combined)
    write(args.out/'generation-receipts.json',receipts)
    write(args.out/'contracts.json',[catalog[k] for k in keys])
    final=run('build','--spec',args.out/'combined.foundation.json','--out',temporary/'combined')
    if final.returncode: raise SystemExit(final.stderr)
    for filename in ('reference.pptx','layout-report.json'):
        (args.out/filename).write_bytes((temporary/'combined'/filename).read_bytes())
    print(f'Generated {len(combined["slides"])} paired specimen slides. Native review pending.')

if __name__ == '__main__': main()
