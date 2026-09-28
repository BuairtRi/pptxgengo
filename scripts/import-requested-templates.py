#!/usr/bin/env python3
"""Register explicitly selected source layouts, preserving native objects and runs.

Consumes curated source reviews and already extracted, hash-pinned scene projects.
Does not infer visual qualification, capacities, or portable component behavior.
"""
import argparse
import hashlib
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def read(p): return json.loads((ROOT / p).read_text())
def sha(p): return hashlib.sha256((ROOT / p).read_bytes()).hexdigest()
def write(p, x):
    p = ROOT / p
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(x, indent=2, ensure_ascii=False) + '\n')
def slug(s): return re.sub('[^a-z0-9]+', '_', s.lower()).strip('_')[:65]
def attr(n, key): return next((a['value'] for a in n.get('attributes', []) if a['name'] == key), '')
def local(n): return n.get('name', '').split(':')[-1]
def child(n, name): return next((c for c in n.get('children', []) if local(c) == name), {})
def objects(scene):
    out = []
    def walk(n, parent=None):
        if local(n) in ('sp', 'grpSp', 'pic', 'cxnSp', 'graphicFrame'):
            nv = next((c for c in n.get('children', []) if local(c).startswith('nv')), {})
            props = child(nv, 'cNvPr'); ident = attr(props, 'id')
            if ident:
                out.append(dict(object_id=ident, name=attr(props, 'name'), kind=local(n), parent=parent))
                parent = ident
        for c in n.get('children', []): walk(c, parent)
    walk(scene)
    return out

def source_values(contract, bindings):
    out = {}
    for name, slot in contract['slots'].items():
        bs = [bindings[b] for b in slot['binding_ids']]
        if slot.get('value_format') == 'paragraphs':
            paras=[]; last=None
            for b in bs:
                key=b['node_path'][:-3]
                if last != key: paras.append({'runs':[]});last=key
                paras[-1]['runs'].append(dict(binding_id=b['binding_id'],text=b['value']))
            out[name]={'paragraphs':paras}
        else: out[name]=[b['value'] for b in bs]
    return {'slots':out}

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--reviews',nargs='+',default=['planning/requested-templates/uhg-review.json','planning/requested-templates/ai-review.json'])
    ap.add_argument('--source', choices=['uhg','ai-accelerator'], help='Import one source while the other review is being prepared')
    args=ap.parse_args()
    request=read('planning/requested-templates/request.json')
    reviews={}
    for p in args.reviews:
        doc=read(p)
        for r in doc['slides']: reviews[(doc['source_id'],r['slide_number'])]=r
    assignments=read('library/templates/rollout/assignments.json')
    entries=assignments['entries']; lookup={e['representative']:e for e in entries}
    taxonomy=read('library/templates/taxonomy.json')
    selected=[]; component_rows=[]
    for source in request['sources']:
        if args.source and source['source_id'] != args.source: continue
        sid=source['source_id']; project=source['project']
        assert sha(source['path'])==source['sha256'],f'Live source changed since snapshot: {sid}'
        assert read(project+'/manifest.json')['source_sha256']==source['sha256']
        # Existing UHG contracts point to the expanded project; scene bytes must
        # match exactly. Historical source projects and evidence stay untouched.
        for e in entries:
            if e['representative'].split(':')[0]==sid:
                scene=project+'/slides/'+Path(e['source_scene']).name
                assert sha(scene)==sha(e['source_scene']),e['template_id']
                e['source_project']=project; e['source_scene']=scene
        for page,n in enumerate(source['slides'],1):
            review=reviews[(sid,n)]; key=f'{sid}:{n:03}'
            assert review['category'] in taxonomy['categories'],review['category']
            scene_path=f'{project}/slides/uhg-{n:03}.json'
            scene=read(scene_path); guide=read(f'{project}/binding-guide-{n:03}.json')
            bindings={b['binding_id']:b for b in scene['bindings']}
            preview=f'samples/requested-templates/render/{sid}/slide-{page:03}.png'
            assert (ROOT/preview).exists(),preview
            existed=key in lookup and lookup[key]['rank']<=65
            if key in lookup:
                e=lookup[key]; directory=e['implementation_directory']; contract=read(directory+'/contract.json')
            else:
                rank=max(e['rank'] for e in entries)+1
                ident=f't{rank:03}-{sid}-{n:03}'
                directory=f'library/templates/rollout/user_selected/{ident}'
                e=dict(rank=rank,id=f'template-family:user-selected:{key}',name=review['name'],category=review['category'],representative=key,altitudes=[review.get('altitude','detail')],densities=[review.get('density','dense')],adaptation_complexity='source_import',qualification_state='source_layout',template_id=ident,lane='user_selected',status='implemented',source_project=project,source_scene=scene_path,implementation_directory=directory,selection_authority='user_explicit_slide_number')
                entries.append(e);lookup[key]=e
                slots={}; owned=[]; title=str(review.get('title_object',''))
                names=review.get('object_slots',{})
                excluded=[]
                for obj in guide['objects']:
                    if not obj['text']: continue
                    ident=obj['object_id']; name=obj['name']; text=''.join(b['value'] for b in obj['text'])
                    if 'Footer' in name or 'Slide Number' in name or text.startswith('©') or text.strip().lower()=='west monroe' or (sid=='ai-accelerator' and n==2 and ident in ['34','35','36']):
                        excluded.append(ident);continue
                    base=names.get(ident) or ('title' if ident==title else slug(name) if not re.fullmatch(r'(Text|TextBox|Content Placeholder|Title|Google Shape)[\s;\d;p]*',name) else '')
                    if not base: base=slug(' '.join(text.split()[:6])) or f'content_{ident}'
                    slot=base if base not in slots else f'{base}_{ident}'
                    slots[slot]=dict(binding_ids=[b['id'] for b in obj['text']],description=f'{name}, native object {ident}. Source: {text[:240]}. Preserve ordered runs and existing line breaks.')
                    owned.append(ident)
                roles={}; source_profile={}
                for ident in map(str,review.get('accent_objects',[])):
                    obj=next((o for o in guide['objects'] if o['object_id']==ident),None)
                    if not obj: raise ValueError(f'{key} missing accent object {ident}')
                    # Only explicit direct shape fills, never text or line fills.
                    colors=[b for b in obj['colors'] if b['property']=='fill.srgbClr.val' and [s.split(':')[-1] for s in b['ancestors'][-3:]]==['spPr','solidFill','srgbClr']]
                    for i,b in enumerate(colors):
                        role=f'accent_{ident}_{i+1}'
                        roles[role]=dict(binding_ids=[b['id']],description=f'Decorative accent fill in native object {ident}; no surrounding surface/text change.')
                        source_profile[role]=b['value']
                    if colors and ident not in owned: owned.append(ident)
                profiles={}
                if roles:
                    profiles['source']=source_profile
                    for name,color in [('gray','CED7E6'),('navy','070154'),('blue','0047FF'),('pink','F900D3')]:profiles[name]={r:color for r in roles}
                contract=dict(schema='pptxgengo.component-contract.v1',id='component-contract:'+e['template_id'],component_id='component:source-template:'+e['template_id'],source_sha256=source['sha256'],slide=n,scene_sha256=sha(scene_path),object_ids=owned,slots=slots,roles=roles,profiles=profiles,constraints=['Preserve original native geometry, object order, typography, image crops, resource relationships and master/layout.','Slot arrays preserve source rich-run cardinality and existing tab/line-break sequence.','No automatic reflow, item-count changes, image replacement, or arbitrary content fit qualification.','Color profiles affect only explicitly listed decorative bindings; they do not recolor the whole template.'])
                metadata=dict(schema='pptxgengo.template-implementation.v1',template_id=e['template_id'],family_id=e['id'],source=key,source_sha256=source['sha256'],scene_sha256=sha(scene_path),preview=dict(path=preview,sha256=sha(preview)),lane=e['lane'],category=e['category'],altitude=e['altitudes'][0],density=e['densities'][0],role=review['name'],takeaway='Author supplies a content-specific takeaway when adapting this source layout.',components=review['components'],slot_notes=dict(coverage='Native text in all selected content objects, including table cells and grouped objects where bindings exist.',rich_runs='Arrays preserve source run boundaries and typography.',source_values='Original words retained, including source facts and existing spelling.'),style_notes=['Source fonts, colors and geometry retained by default.','Named color profiles only affect individually mapped decorative objects.'] if roles else ['Source typography and styling retained; no blanket color profile declared.'],opaque_areas=['Images, embedded charts/think-cell/OLE data and master/layout artwork remain original resources. Their pixels or data are not changed by text slots.'],retained_source_content=['Original customer/staff names, testimonials, metrics, photos and claims remain source material, not approved claims for a new deck.','Footer, legal text and page number remain original.'],limitations=review.get('limitations',[])+contract['constraints'][1:3],status='implemented',adaptation_mode='fixed_source_geometry',native_fit='not_measured',adaptation_qualified=False,example_kind='source_reference',changed_content_coverage=dict(editable_text_bindings=sum(len(s['binding_ids']) for s in slots.values()),text_objects=len(slots),opaque_artwork_editable=False),excluded_footer_objects=excluded)
                write(directory+'/contract.json',contract);write(directory+'/implementation.json',metadata)
                write(directory+'/example-values.json',source_values(contract,bindings))
            if e['rank'] > 65:
                metadata=read(directory+'/implementation.json')
                metadata['components']=review['components']
                write(directory+'/implementation.json',metadata)
            values=source_values(contract,bindings);write(directory+'/source-values.json',values)
            # Actual native groups and curated visual groups remain distinct from
            # portable/reflowable components. Emit executable subcontracts where
            # the template exposes relevant text/style bindings.
            ids={o['object_id'] for o in objects(scene['scene'])}
            components=[]
            for group in review['components']:
                gids=list(map(str,group['object_ids']));assert set(gids)<=ids,(key,group['id'],set(gids)-ids)
                gs={name:s for name,s in contract['slots'].items() if {bindings[b].get('object_id') for b in s['binding_ids']}<=set(gids)}
                gr={name:r for name,r in contract['roles'].items() if r.get('binding_ids') and {bindings[b].get('object_id') for b in r['binding_ids']}<=set(gids) and not r.get('resource_targets')}
                comp=dict(id=e['template_id']+':'+group['id'],name=group['id'].replace('_',' '),purpose=group.get('purpose',group['id']),source=key,template_id=e['template_id'],object_ids=gids,slots=list(gs),profiles=list(contract['profiles']) if gr else [],geometry='original_source_group',portable=False,variable_count=False,accent='accent' in group['id'] or 'arrow' in group['id'])
                if gs or gr:
                    cc=dict(contract,id='component-contract:'+e['template_id']+':'+group['id'],component_id='component:source-group:'+e['template_id']+':'+group['id'],object_ids=sorted({bindings[b]['object_id'] for field in [*gs.values(),*gr.values()] for b in field['binding_ids']}),slots=gs,roles=gr,profiles={name:{k:v for k,v in profile.items() if k in gr} for name,profile in contract['profiles'].items()} if gr else {})
                    cc.pop('zones',None)
                    rel=directory+'/components/'+group['id']+'.json';write(rel,cc);comp['contract']=rel
                components.append(comp);component_rows.append(comp)
            controls=dict(text_slots=len(contract['slots']),text_bindings=sum(len(s['binding_ids']) for s in contract['slots'].values()),profiles=list(contract['profiles']),font='preserved_source_typography',item_count='fixed',geometry='preserved',accents='retained; re-anchor separately if emphasized copy changes',components=components)
            write(directory+'/customization.json',controls)
            selected.append(dict(template_id=e['template_id'],source=key,name=review['name'],existing_contract=existed,category=e['category'],implementation_directory=directory,source_project=project,source_scene=scene_path,preview=dict(path=preview,sha256=sha(preview)),controls=controls,limitations=review.get('limitations',[]),source_native_rendered=True,adaptation_qualified=False))
    assignments['target']=len(entries);assignments['scheduling']='Original 65 contracts plus explicitly requested exact source layouts; native review serialized.'
    write('library/templates/rollout/assignments.json',assignments)
    write('library/templates/requested-templates.json',dict(schema='pptxgengo.requested-template-catalog.v1',requested_count=len(selected),templates=selected,total_registered_templates=len(entries),scope='Exact source layouts; source values retained. Native rendering does not qualify arbitrary replacement content.'))
    write('library/templates/requested-components.json',dict(schema='pptxgengo.source-component-catalog.v1',components=component_rows,scope='Curated component occurrences with exact source bindings, not a deduplicated count of unique designs or portable dynamic components.'))
    print(f'{len(selected)} requested templates registered; {len(entries)} total; {len(component_rows)} source component occurrences')

if __name__=='__main__':main()
