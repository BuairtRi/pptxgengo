#!/usr/bin/env python3
"""Validate independent source reviews and compile the template discovery catalog.

This publishes inventory candidates, never executable/qualified contracts.
"""
import argparse
from collections import Counter, defaultdict
import hashlib
import html
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def read(path):
    return json.loads((ROOT / path).read_text())

def digest(path):
    return hashlib.sha256((ROOT / path).read_bytes()).hexdigest()

def artifact(path):
    return dict(path=str(path), sha256=digest(path))

def require(condition, message):
    if not condition:
        raise ValueError(message)

def write(path, value):
    (ROOT / path).parent.mkdir(parents=True, exist_ok=True)
    (ROOT / path).write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n')

def compile_catalog(inputs, allow_metadata=False):
    taxonomy = read('library/templates/taxonomy.json')
    registry = read('planning/source-registry.json')
    sources = {s['source_id']: s for s in registry['sources']}
    expected = {f'{s["source_id"]}:{n:03}' for s in sources.values() for n in range(1, s['slide_count'] + 1)}
    decisions = read('library/layout-decisions.json')
    reference_shortlist = read('library/reference-shortlist.json')
    preferences = read('library/reference-preferences.json')
    require(preferences['shortlist_sha256'] == digest('library/reference-shortlist.json'), 'stale user reference preferences')
    preference_by_component = {p['component_id']:p for p in preferences['choices']}
    preferences_by_source = defaultdict(list)
    for ref in reference_shortlist['references']:
        choice = preference_by_component.get(ref['component_id'])
        if choice:
            preferences_by_source[f'{ref["source_id"]}:{ref["slide_number"]:03}'].append(dict(choice, scope='component_in_source_slide_not_whole_template_approval'))
    established = {m: g['id'] for g in decisions['groups'] for m in g['members']}
    rows, definitions, sequences, gaps = {}, defaultdict(list), [], []
    for path in inputs:
        doc = read(path)
        require(doc['schema'] == 'pptxgengo.template-review.v1', f'bad schema: {path}')
        for sid, value in doc['source_hashes'].items():
            require(sid in sources and value == sources[sid]['source_sha256'], f'stale source: {sid}')
        for r in doc['slides']:
            key = f'{r["source_id"]}:{r["slide_number"]:03}'
            require(key in expected and key not in rows, f'unknown/duplicate slide: {key}')
            require(r['source_id'] in doc['source_hashes'], f'missing source hash: {key}')
            for field, allowed in [('category', taxonomy['categories']), ('altitude', taxonomy['altitudes']), ('density', taxonomy['densities'])]:
                require(r[field] in allowed, f'{key}: invalid {field}: {r[field]}')
            require(r['disposition'] in ['prioritize', 'candidate', 'reference_only', 'avoid'], f'bad disposition: {key}')
            require(r['review_status'] in ['visual_reviewed', 'metadata_only'], f'bad review state: {key}')
            if r['review_status'] == 'visual_reviewed':
                preview = Path(r['preview_path'])
                require(preview == Path(f'samples/template-expansion/render/{r["source_id"]}/slide-{r["slide_number"]:03}.png'), f'incorrect source preview: {key}')
                require(digest(preview) == r['preview_sha256'], f'stale preview: {key}')
            require(r.get('rationale') and r.get('pattern'), f'missing reasoning: {key}')
            if key in established:
                require(r['family_id'] == established[key], f'{key}: preserve established family {established[key]}')
            r = dict(r, id=key, source_sha256=sources[r['source_id']]['source_sha256'], review_artifact=artifact(path))
            rows[key] = r
        for family in doc['families']:
            definitions[family['id']].append(dict(family, review_path=str(path)))
        sequences.extend(dict(s, review_path=str(path)) for s in doc.get('sequences', []))
        gaps.extend(dict(review_path=str(path), detail=g) for g in doc.get('gaps', []))
    require(set(rows) == expected, f'coverage mismatch: missing={sorted(expected-set(rows))}, extra={sorted(set(rows)-expected)}')
    curation_path = 'library/templates/curation-decisions.json'
    if (ROOT / curation_path).exists():
        curation = read(curation_path)
        require(set(curation) == {'schema', 'decisions'} and curation['schema'] == 'pptxgengo.template-curation.v1', 'invalid curation schema')
        curated = set()
        required_set = {'category', 'altitude', 'density', 'pattern', 'disposition', 'component_needs'}
        for decision in curation['decisions']:
            require(set(decision) == {'source', 'preview', 'reviewer', 'rationale', 'set'}, 'invalid curation decision fields')
            key = decision['source']
            require(key in rows and key not in curated, f'unknown/duplicate curation source: {key}')
            curated.add(key)
            row = rows[key]
            expected_preview = f'samples/template-expansion/render/{row["source_id"]}/slide-{row["slide_number"]:03}.png'
            require(set(decision['preview']) == {'path', 'sha256'} and decision['preview']['path'] == expected_preview, f'incorrect curation preview: {key}')
            require(digest(expected_preview) == decision['preview']['sha256'], f'stale curation preview: {key}')
            require(isinstance(decision['reviewer'], str) and decision['reviewer'].strip() and isinstance(decision['rationale'], str) and decision['rationale'].strip(), f'missing curation reviewer/reason: {key}')
            changes = decision['set']
            require(set(changes) in (required_set, required_set | {'secondary_categories'}), f'invalid curation fields: {key}')
            for field, allowed in [('category', taxonomy['categories']), ('altitude', taxonomy['altitudes']), ('density', taxonomy['densities'])]:
                require(changes[field] in allowed, f'{key}: invalid curated {field}')
            require(changes['disposition'] in ['prioritize', 'candidate', 'reference_only', 'avoid'], f'{key}: invalid curated disposition')
            require(isinstance(changes['pattern'], str) and changes['pattern'].strip(), f'{key}: empty curated pattern')
            require(isinstance(changes['component_needs'], list) and all(isinstance(v, str) and v.strip() for v in changes['component_needs']), f'{key}: invalid curated component needs')
            secondaries = changes.get('secondary_categories', [])
            require(isinstance(secondaries, list) and len(secondaries) == len(set(secondaries)) and all(v in taxonomy['categories'] and v != changes['category'] for v in secondaries), f'{key}: invalid secondary categories')
            row['worker_review'] = {field: row[field] for field in ('category', 'altitude', 'density', 'pattern', 'disposition', 'component_needs', 'review_status', 'preview_path', 'preview_sha256')}
            row['curation'] = dict(reviewer=decision['reviewer'], rationale=decision['rationale'], preview=decision['preview'], artifact=artifact(curation_path))
            row.update(changes)
            row['review_status'] = 'visual_reviewed'
            row['preview_path'] = decision['preview']['path']
            row['preview_sha256'] = decision['preview']['sha256']
    require(allow_metadata or all(r['review_status']=='visual_reviewed' for r in rows.values()), 'individual visual review is incomplete; metadata-only drafts cannot publish the final catalog')
    # Only explicit decisions can merge independently proposed families.
    merge_path = 'library/templates/cross-slice-decisions.json'
    merge_doc = read(merge_path) if (ROOT / merge_path).exists() else {'merges': []}
    aliases = {}
    for d in merge_doc['merges']:
        require(d.get('rationale') and d.get('reviewed_sources'), 'merge lacks review evidence')
        require(d['canonical'] in definitions, 'unknown canonical family')
        for old in d['aliases']:
            require(old in definitions and old not in aliases and old != d['canonical'], 'bad duplicate alias')
            aliases[old] = d['canonical']
    require(not set(aliases.values()).intersection(aliases), 'chained merges unsupported')
    grouped = defaultdict(list)
    for row in rows.values():
        grouped[aliases.get(row['family_id'], row['family_id'])].append(row)
    active_definitions = set(grouped) | set(aliases)
    require(set(definitions) == active_definitions, f'unused/missing family definitions: {sorted(set(definitions) ^ active_definitions)}')
    families = []
    for ident, members in sorted(grouped.items()):
        require(ident in definitions, f'family definition missing: {ident}')
        defs = definitions[ident]
        d = defs[0]
        require(isinstance(d.get('name'), str) and d['name'].strip() and isinstance(d.get('priority_reason'), str) and d['priority_reason'].strip(), f'missing family name/reason: {ident}')
        require(d['adaptation_complexity'] in ['low','medium','high'], f'bad complexity: {ident}')
        member_ids = {m['id'] for m in members}
        rep = d['representative']
        require(rep in member_ids, f'representative outside family: {ident}')
        related_defs = [candidate for name, candidates in definitions.items() if name == ident or aliases.get(name) == ident for candidate in candidates]
        declared = {m for candidate in related_defs for m in candidate['members']}
        require(declared == member_ids, f'family membership conflict: {ident}; extra={sorted(declared-member_ids)} missing={sorted(member_ids-declared)}')
        representative = rows[rep]
        family_name = representative['pattern'] if 'curation' in representative else d['name']
        family = dict(id='template-family:' + ident, source_family_id=ident, kind='template_family', name=family_name,
            source_family_name=d['name'],
            representative=rep, members=sorted(member_ids), categories=sorted({c for m in members for c in [m['category'], *m.get('secondary_categories', [])]}),
            altitudes=sorted({m['altitude'] for m in members}), densities=sorted({m['density'] for m in members}),
            source_ids=sorted({m['source_id'] for m in members}), primary_category=representative['category'],
            disposition='prioritize' if any(m['disposition']=='prioritize' for m in members) else representative['disposition'],
            component_needs=sorted({c for m in members for c in m.get('component_needs', [])}),
            suggested_slots=d.get('suggested_slots', []), adaptation_complexity=d['adaptation_complexity'],
            rationale=d['priority_reason'], dedup_rationale=[x['dedup_rationale'] for x in defs],
            variant_notes=[x.get('variant_notes', []) for x in defs], aliases=sorted(k for k,v in aliases.items() if v==ident),
            preview=dict(path=representative['preview_path'],sha256=representative['preview_sha256']),
            review_status='visual_reviewed' if all(m['review_status']=='visual_reviewed' for m in members) else 'partially_reviewed',
            qualification_state='inventory', executable=False, adaptation_qualified=False,
            design_preference='unreviewed', selection_authority='agent_nomination_not_user_preference',
            measured_capacity=None, source_reviews=sorted({m['review_artifact']['path'] for m in members}))
        family['component_preferences'] = [dict(p, source=member) for member in sorted(member_ids) for p in preferences_by_source[member]]
        family['representative_contains_avoided_component'] = any(p['preference']=='avoid' for p in preferences_by_source[rep])
        families.append(family)
    inputs = [artifact(p) for p in inputs] + [artifact('planning/source-registry.json'), artifact('library/layout-decisions.json'),artifact('library/templates/taxonomy.json')]
    if (ROOT/merge_path).exists():inputs.append(artifact(merge_path))
    if (ROOT/curation_path).exists():inputs.append(artifact(curation_path))
    inputs += [artifact('library/reference-shortlist.json'), artifact('library/reference-preferences.json')]
    return dict(schema='pptxgengo.template-catalog.v1', inputs=inputs,
        summary=dict(source_slides=len(rows), visual_reviewed=sum(r['review_status']=='visual_reviewed' for r in rows.values()),
                     classification_families=len(families), repeated_occurrences=len(rows)-len(families),
                     adaptation_qualified=0, executable_templates_in_this_catalog=0,
                     category_counts=dict(Counter(f['primary_category'] for f in families))),
        families=families, slides=[rows[k] for k in sorted(rows)], observed_sequences=sequences, gaps=gaps,
        caveat='Classification families may need further cross-family review. Existing executable contracts are separate; inventory never grants adaptation approval.')

def shortlist(catalog, size):
    # Round-robin category coverage. Selection is a transparent implementation queue,
    # not a claim of human preference, completed adaptation or final uniqueness.
    eligible = [f for f in catalog['families'] if f['review_status']=='visual_reviewed' and f['disposition'] in ['prioritize','candidate'] and f['primary_category']!='reference_instruction' and not f['representative_contains_avoided_component']]
    rank = lambda f:(f['disposition']!='prioritize', {'low':0,'medium':1,'high':2}[f['adaptation_complexity']], f['id'])
    buckets = defaultdict(list)
    for f in sorted(eligible,key=rank):buckets[f['primary_category']].append(f)
    chosen=[]
    batch_path=ROOT/'library/templates/first-adaptation-batch.json'
    if batch_path.exists():
        batch=json.loads(batch_path.read_text())
        for item in batch['templates']:
            matches=[f for f in eligible if item['source'] in f['members']]
            require(len(matches)==1, f'first-batch source must resolve to one reviewed eligible family: {item["source"]}')
            f=matches[0]
            if f not in chosen:
                chosen.append(f)
                buckets[f['primary_category']].remove(f)
    while len(chosen)<size and any(buckets.values()):
        for category in sorted(buckets):
            if buckets[category] and len(chosen)<size:chosen.append(buckets[category].pop(0))
    return dict(schema='pptxgengo.template-release-shortlist.v1', target=size, selected=len(chosen),
        status='implementation_candidates_not_qualified_templates', selection_method='Root first adaptation batch, then category round-robin; agent-prioritized candidates then complexity. No user preference is inferred.',
        category_counts=dict(Counter(f['primary_category'] for f in chosen)),
        entries=[dict(rank=i+1,id=f['id'],name=f['name'],category=f['primary_category'],representative=f['representative'],altitudes=f['altitudes'],densities=f['densities'],adaptation_complexity=f['adaptation_complexity'],next_step='Bind slots/components/styles and qualify changed-content fixtures',qualification_state='inventory') for i,f in enumerate(chosen)],
        gaps=catalog['gaps'])

def gallery(catalog, short, output):
    selected={f['id'] for f in short['entries']}
    import os
    cards=[]
    for f in catalog['families']:
        src=os.path.relpath(ROOT/f['preview']['path'], (ROOT/output).parent)
        data=html.escape(json.dumps(dict(categories=f['categories'],altitudes=f['altitudes'],densities=f['densities'],selected=f['id'] in selected)))
        desc=' '.join([f['name'],f['rationale'],*f['component_needs']])
        cards.append(f'<article data-tags="{data}" data-search="{html.escape(desc.lower())}"><a href="{html.escape(src)}"><img loading="lazy" src="{html.escape(src)}" alt="{html.escape(f["name"])}"></a><h2>{html.escape(f["name"])}</h2><p>{html.escape(f["representative"])} · {len(f["members"])} source occurrence(s)</p><p>{html.escape(" / ".join(f["categories"]))}<br>{html.escape(" / ".join(f["altitudes"]))} · {html.escape(" / ".join(f["densities"]))}</p><p>{html.escape(f["rationale"])}</p><small>{"Shortlisted · " if f["id"] in selected else ""}Inventory candidate · adaptation unproven</small></article>')
    select=lambda name,values:'<label>'+name+' <select id="'+name+'"><option value="">All</option>'+''.join('<option value="'+html.escape(v)+'">'+html.escape(v.replace('_',' ').title())+'</option>' for v in sorted(values))+'</select></label>'
    controls=select('category',{v for f in catalog['families'] for v in f['categories']})+select('altitude',{v for f in catalog['families'] for v in f['altitudes']})+select('density',{v for f in catalog['families'] for v in f['densities']})
    sequence_cards=[]
    for sequence in read('library/templates/sequence-patterns.json')['patterns']:
        steps=[]
        for step in sequence['steps']:
            label=step['role'].replace('-', ' ').title()+' ('+step['altitude']+')'
            if step.get('optional'):label+=' · optional'
            if step['repeat']!='once':label+=' · '+step['repeat'].replace('_',' ')+f" · {step.get('min_pages',1)}–{step.get('max_pages',1)} pages"
            steps.append('<li>'+html.escape(label)+'</li>')
        sequence_cards.append('<section><h3>'+html.escape(sequence['name'])+'</h3><ol>'+''.join(steps)+'</ol><p>'+html.escape(' '.join(sequence['continuity_rules']))+'</p></section>')
    sequences='<details><summary>Multi-slide narrative patterns · 4 proposed sequences</summary><p>These describe ordering and continuity. Automatic sequence expansion is not implemented.</p>'+''.join(sequence_cards)+'</details>'
    doc='''<!doctype html><html lang="en"><meta charset="utf-8"><title>Template inventory and release candidates</title><style>body{font:16px Arial;margin:28px;background:#f4f6fa;color:#070154}h1{font-size:30px}header{position:sticky;top:0;background:#f4f6fa;padding:12px 0;z-index:1}input,select{font:inherit;padding:8px;margin:6px}main{display:grid;grid-template-columns:repeat(auto-fill,minmax(380px,1fr));gap:22px}article{background:white;padding:14px;border:1px solid #ccd3df;border-radius:8px}article[hidden]{display:none}img{width:100%;height:auto}h2{font-size:19px}small{color:#536586}label{white-space:nowrap}</style><h1>Template inventory and release candidates</h1><p>Architecture, product, layers/components and narrative sequences are included. Altitude and density are independent. These previews are source designs, not newly qualified templates.</p>'''+sequences+'''<header><input id="search" placeholder="Search purpose or component" aria-label="Search">'''+controls+'''<label><input id="shortlisted" type="checkbox" checked>Release shortlist</label><strong id="count"></strong></header><main>'''+''.join(cards)+'''</main><script>function filter(){let n=0;for(const el of document.querySelectorAll('article')){const t=JSON.parse(el.dataset.tags);const q=document.querySelector('#search').value.toLowerCase();const ok=el.dataset.search.includes(q)&&(!document.querySelector('#shortlisted').checked||t.selected)&&['category','altitude','density'].every((k,i)=>{const v=document.getElementById(k).value;return !v||t[['categories','altitudes','densities'][i]].includes(v)});el.hidden=!ok;n+=ok}document.querySelector('#count').textContent=n+' families'}document.querySelectorAll('input,select').forEach(e=>e.addEventListener('input',filter));filter();</script></html>'''
    (ROOT/output).parent.mkdir(parents=True,exist_ok=True)
    (ROOT/output).write_text(doc)

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--inputs',nargs='+',default=['library/templates/reviews/slice-a.json','library/templates/reviews/slice-b.json','library/templates/reviews/slice-c.json'])
    ap.add_argument('--target',type=int,default=65)
    ap.add_argument('--allow-metadata',action='store_true',help='compile an explicitly incomplete working snapshot; exclude unreviewed families from shortlist')
    ap.add_argument('--out',default='library/templates/catalog.json')
    ap.add_argument('--shortlist',default='library/templates/release-shortlist.json')
    ap.add_argument('--gallery',default='samples/template-expansion/gallery.html')
    args=ap.parse_args()
    require(50<=args.target<=75,'release target must be 50..75 distinct candidates')
    cat=compile_catalog(args.inputs,args.allow_metadata); short=shortlist(cat,args.target)
    write(args.out,cat);write(args.shortlist,short);gallery(cat,short,args.gallery)
    print(json.dumps(dict(summary=cat['summary'],shortlisted=short['selected'],gallery=args.gallery),indent=2))

if __name__=='__main__':main()
