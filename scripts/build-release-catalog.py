#!/usr/bin/env python3
"""Build portable fixed-template and supported-component galleries for local release."""
import json, shutil, re, html, hashlib, subprocess, sys
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'release/catalog'
OLD_GALLERY=ROOT/'samples/template-expansion/rollout/gallery-v3/index.json'
REQ=ROOT/'samples/requested-templates/gallery/index.json'
COMP=ROOT/'library/templates/requested-components.json'
CHECK=ROOT/'library/adaptive/checkpoint.json'

def read(p): return json.loads(p.read_text())
def copy(src,dst):
    if Path(src).resolve()==Path(dst).resolve(): return
    dst.parent.mkdir(parents=True,exist_ok=True); shutil.copy2(src,dst)
def esc(x): return html.escape(str(x or ''),quote=True)
def slug(x): return re.sub(r'[^a-zA-Z0-9._-]+','-',x).strip('-')
def main():
    version=(ROOT/'release/VERSION').read_text().strip()
    old_gallery=read(OLD_GALLERY); new=read(REQ); occurrences=read(COMP)['components']; adaptive=read(CHECK)
    old_by={x['id']:x for x in old_gallery['templates']}
    old_source_by=old_by
    if OUT.exists(): shutil.rmtree(OUT)
    OUT.mkdir(parents=True,exist_ok=True)
    req_by={x['template_id']:x for x in new['templates']}
    ids=set(old_by)|set(req_by)
    if len(ids)!=101: raise ValueError(f'Expected 101 unique source templates; got {len(ids)}')
    templates=[]
    for tid in sorted(ids):
        a=old_by.get(tid); a_source=old_source_by.get(tid); b=req_by.get(tid)
        td=OUT/'templates'/tid; td.mkdir(parents=True,exist_ok=True)
        if a:
            item=dict(a); source=(OLD_GALLERY.parent/a_source['source_png']).resolve()
            source_rel=f'templates/{tid}/source.png'; copy(source,OUT/source_rel)
            adapted_rel=None
            if a_source.get('adapted_png'):
                adapted_rel=f'templates/{tid}/adapted.png'; copy((OLD_GALLERY.parent/a_source['adapted_png']).resolve(),OUT/adapted_rel)
            contract_src=(OLD_GALLERY.parent/a_source['contract']).resolve(); values_src=(OLD_GALLERY.parent/a_source['values']).resolve()
            contract_rel=f'templates/{tid}/contract.json'; values_rel=f'templates/{tid}/values.json'
            copy(contract_src,OUT/contract_rel); copy(values_src,OUT/values_rel)
            contract_data=read(contract_src)
            item.update(source_png=source_rel,adapted_png=adapted_rel,contract=contract_rel,values=values_rel,
                        slots=len(contract_data.get('slots',{})),profiles=list(contract_data.get('profiles',{})),zones=list(contract_data.get('zones',{})),
                        controls={'slots':len(contract_data.get('slots',{})),'profiles':list(contract_data.get('profiles',{})),'zones':list(contract_data.get('zones',{})),'geometry':'fixed source geometry'})
            item['source_native_rendered']=True
            item['review_status']='Source preview and one illustrative adaptation are listed; use source-fit review before reuse.'
        else:
            assert b
            srcdir=ROOT/'samples/requested-templates/gallery'/f'templates/{tid}'
            source_rel=f'templates/{tid}/source.png'; copy(srcdir/'source.png',OUT/source_rel)
            contract_rel=f'templates/{tid}/contract.json'; values_rel=f'templates/{tid}/values.json'
            copy(srcdir/'contract.json',OUT/contract_rel); copy(srcdir/'source-values.json',OUT/values_rel)
            for filename in ['implementation.json','customization.json']:
                if (srcdir/filename).exists(): copy(srcdir/filename,td/filename)
            item={'id':tid,'name':b['name'],'source':b['source'],'category':b['category'],'lane':'requested_source',
                  'altitude':['overview'],'density':['medium'],'source_png':source_rel,'adapted_png':None,
                  'contract':contract_rel,'values':values_rel,'slots':b.get('controls',{}).get('text_slots',0),
                  'profiles':b.get('controls',{}).get('profiles',[]),'zones':[],'limitations':b.get('limitations',[]),
                  'source_native_rendered':b.get('source_native_rendered',False),'adaptation_qualified':False,
                  'review_status':'Source rendered; reviewed example does not qualify arbitrary content'}
        # Merge exact requested-source component/control metadata without changing qualification.
        if b:
            if a:
                srcdir=ROOT/'samples/requested-templates/gallery'/f'templates/{tid}'
                for filename in ['implementation.json','customization.json']:
                    if (srcdir/filename).exists(): copy(srcdir/filename,td/filename)
            item['requested_source']=True; item['source_project']=b.get('source_project')
            item['source_scene']=b.get('source_scene'); item['implementation']=f'templates/{tid}/implementation.json'
            item['customization']=f'templates/{tid}/customization.json'
            item['components']=[]
            for rawc in b.get('components',[]):
                cc=dict(rawc)
                if cc.get('contract'):
                    cc['contract']=f"components/{slug(cc['id'])}/contract.json"
                    cc['contract_path']=cc['contract']
                item['components'].append(cc)
            item['controls_detail']=b.get('controls',{})
            item['limitations']=list(dict.fromkeys((item.get('limitations') or [])+(b.get('limitations') or [])))
            item['adaptation_qualified']=bool(a and item.get('adaptation_qualified'))
        item['command']=f'pptxgengo template inspect --id {tid}'
        item['customization_boundary']='Text and declared style fields only; source positions, card/row counts and artwork remain fixed unless a separately documented structural component route changes them.'
        templates.append(item)
    # Build component-focused previews using source-object rectangles over the exact source preview.
    bytemplate={x['id']:x for x in templates}
    comp_rows=[]
    for c in occurrences:
        t=bytemplate[c['template_id']]
        preview_svg=t['source_png']
        out_contract=None
        if c.get('contract') and (ROOT/c['contract']).exists():
            out_contract=f"components/{slug(c['id'])}/contract.json"; copy(ROOT/c['contract'],OUT/out_contract)
        # Locate exact binding guide for source IDs and derive source-object focus boxes (16:9 EMU canvas).
        project, num=c['source'].rsplit(':',1)
        candidates=[ROOT/'samples/requested-templates/sources'/project/f'binding-guide-{int(num):03}.json',
                    ROOT/'samples/template-expansion/rollout/sources'/project/f'binding-guide-{int(num):03}.json']
        guide=next((p for p in candidates if p.exists()),None); objmap={}
        if guide:
            try: objmap={str(o.get('object_id')):o for o in read(guide).get('objects',[])}
            except Exception: pass
        boxes=[]; nested=False
        pngw,pngh=1920,1080
        for oid in c.get('object_ids',[]):
            o=objmap.get(str(oid),{}); geom={x.get('property'):x.get('value') for x in o.get('geometry',[])}
            anc=(o.get('geometry') or [{}])[0].get('ancestors',[])
            if 'p:grpSp' in anc: nested=True; continue
            try:
                x=float(geom['transform.off.x']); y=float(geom['transform.off.y']); w=float(geom['transform.ext.cx']); h=float(geom['transform.ext.cy'])
                # Exact object-to-preview mapping on the source's 16:9 slide canvas.
                sx=pngw/12192000; sy=pngh/6858000
                boxes.append([round(x*sx,2),round(y*sy,2),round(w*sx,2),round(h*sy,2)])
            except Exception: pass
        image='../'+t['source_png']
        row=dict(c); row.update(preview=preview_svg,contract_path=out_contract,contract_available=bool(out_contract),
                               visual_only=not bool(out_contract),focus_object_boxes=len(boxes),focus_boxes=boxes,nested_geometry_present=nested,
                               template_preview=t['source_png'])
        comp_rows.append(row)
    # Gauge example is a distinct executable source component workflow, outside the 132 source-group occurrences.
    gauge_image='assets/source-gauge-controls.png'; copy(ROOT/'samples/requested-templates/gallery/gauge/source-gauge-controls.png',OUT/gauge_image)
    gauge_values='assets/gauge-values.json'; copy(ROOT/'samples/requested-templates/gallery/gauge/values.json',OUT/gauge_values)
    gauge_doc=ROOT/'library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/gauge-authoring.md'
    if gauge_doc.exists(): copy(gauge_doc,OUT/'assets/gauge-authoring.md')
    gauge={'id':'T045-gauge-controls','name':'Original five-cell gauge controls','template_id':'t045-graphics-and-layouts-045','source':'graphics-and-layouts:045',
           'preview':gauge_image,'values':gauge_values,'contract_available':True,'kind':'separate_structural_component_example',
           'description':'Five native semicircular gauges; the command independently selects highlighted source cells and pointer positions. Cell count, shape and upper scale artwork remain fixed.',
           'command':'pptxgengo template inspect --id t045-graphics-and-layouts-045; then follow assets/gauge-authoring.md',
           'controls':['highlight_cells: distinct positions 1–5','pointer_cell: position 1–5; may differ from highlighted cells','highlight_color: pink, blue, navy or gray']}
    # Package bounded, visually reviewed new-slide composition examples and evidence separately.
    family_rows=[]; famdir=OUT/'compose'; famdir.mkdir(exist_ok=True); evidence_assets={}
    for fam,record in adaptive.get('families',{}).items():
        for ex in record.get('examples',[]):
            er=dict(ex); prefix=f'compose/{slug(ex["id"])}'; family_rows.append(er)
            er['source_paths']={'render':er.pop('render_png',None),'spec':er.pop('spec_path',None),'values':er.pop('values_path',None),'evidence':er.pop('evidence',{})}
            src=er['source_paths']['render']
            if src and (ROOT/src).exists():
                er['render_asset']=prefix+'.png'; copy(ROOT/src,OUT/er['render_asset'])
            for key in ['spec','values']:
                src=er['source_paths'][key]
                if src and (ROOT/src).exists():
                    out=f'{prefix}-{key}.json'; copy(ROOT/src,OUT/out); er[key+'_asset']=out
            for key,path in er['source_paths']['evidence'].items():
                if path and (ROOT/path).exists():
                    if path not in evidence_assets:
                        suff=Path(path).suffix or '.bin'; hx=hashlib.sha256(path.encode()).hexdigest()[:12]
                        out=f'compose/evidence/{hx}-{Path(path).name}'; copy(ROOT/path,OUT/out); evidence_assets[path]=out
                    er[key+'_asset']=evidence_assets[path]
            er['limitations']=ex.get('limitations',[])
    recipes=[]
    definitions=[
        ('cards','Numbered and metric cards','library/dynamic-components/cards.json','library/dynamic-components/cards.md','Number/title/body or value/label; explicit bounds; light/dark profiles; surface, foreground and accent colors.','Fixed Arial styles and bounded text fields. This is an illustrative recipe; changed content needs native measurement.'),
        ('pods','Delivery pods and role tiles','library/dynamic-components/pods.json','library/dynamic-components/README.md','Ordered role lists, role counts, pod placement and native editable boxes.','Measured layout primitives; counts and copy must fit the chosen space.'),
        ('team','Teams, phases and reporting lines','library/dynamic-components/team.json','library/dynamic-components/team.md','Pods, role tiles, phase bands, staffing legend and explicit reporting connections.','Staffing colors carry legend meaning. Routes and text require native review after changes.'),
        ('layouts','Measured grids and panels','library/layout-components/controls.json','library/layout-components/README.md','Grid/panel geometry, spacing, padding and declared style tokens.','New measured compositions, not automatic reshaping of imported templates.'),
        ('visual','Rich text, images and native shapes','library/visual-components/qualification.json','library/visual-components/README.md','Native text runs, image placement/cropping and supported shape presets.','Examples have individual review scope; content and assets remain author supplied.'),
        ('dense','Dense proposal compositions','library/showcase/dense-deck.json','library/showcase/README.md','Five proposal patterns: approach, phase detail, workflow matrix, delivery organization and roadmap.','Reviewed specimens, not universal layouts. Preserve details and remeasure new copy.'),
    ]
    for key,name,spec,guide,controls,limits in definitions:
        asset=f'recipes/{key}.json'; copy(ROOT/spec,OUT/asset)
        recipes.append(dict(id=key,name=name,spec=asset,installed_spec=spec,installed_guide=guide,controls=controls,limitations=limits))
    # Index data retains template and component records as distinct catalog types.
    payload={'schema':'pptxgengo.release-catalog.v2','version':version,
             'catalog_scope':{'template_count':len(templates),'component_occurrence_count':len(comp_rows),'executable_component_occurrences':sum(not x['visual_only'] for x in comp_rows),'visual_only_component_occurrences':sum(x['visual_only'] for x in comp_rows),'unique_component_designs':None,
                              'component_count_note':'Source component groups are listed as template-bound occurrences; 132 entries are not 132 unique reusable designs.'},
             'qualification':{'source_contracts':len(templates),'reviewed_examples':65,'direct_native_exports':60,'geometry_replays':5,'arbitrary_content_qualified':0},
             'templates':templates,'components':comp_rows,'standalone_component_examples':[gauge],'compose_families':{k:{'status':v.get('status'),'example_count':v.get('example_count'),'qualification_boundary':'Exact bounded examples only; arbitrary input and source-template restructuring are not qualified.'} for k,v in adaptive.get('families',{}).items()},'compose_examples':family_rows,'compose_recipes':recipes}
    (OUT/'index.json').write_text(json.dumps(payload,indent=2,ensure_ascii=False)+'\n')
    common='''<style>*{box-sizing:border-box}body{margin:0;background:#f4f5f8;color:#211b48;font:16px/1.5 system-ui,sans-serif}header{background:#10004b;color:#fff;padding:28px max(24px,5vw)}h1{margin:0;font-size:32px}header p{max-width:1050px}.nav{display:flex;gap:18px;flex-wrap:wrap;margin-top:12px}.nav a{color:#fff}main{max-width:1500px;margin:auto;padding:24px}.intro{background:#fff;border:1px solid #ddd;padding:18px;margin:12px 0 22px}.toolbar{position:sticky;top:0;z-index:3;background:white;display:flex;gap:10px;flex-wrap:wrap;padding:12px;border-bottom:1px solid #ccc}input,select{padding:9px;border:1px solid #bbc;border-radius:4px;font:inherit}input[type=checkbox]{min-width:0;flex:initial;width:auto}input{min-width:230px;flex:1}.count{align-self:center}article,.tile{background:white;border:1px solid #ddd;border-radius:4px;padding:18px;margin:0 0 18px}h2{font-size:21px;margin:0 0 8px}h3{font-size:18px}.meta{font-size:13px;color:#5b5871}.tags{display:flex;gap:5px;flex-wrap:wrap}.tag{background:#ece9f6;padding:3px 8px;border-radius:3px;font-size:12px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(540px,1fr));gap:14px}.grid article{margin:0}.preview{width:100%;border:1px solid #ddd;display:block;background:#f8f8fa}.focus{position:relative;width:100%;aspect-ratio:16/9}.focus img{width:100%;height:100%;object-fit:contain;display:block}.focus svg{position:absolute;inset:0;width:100%;height:100%;pointer-events:none}.previewbox{margin:12px 0}.pair{display:grid;grid-template-columns:1fr 1fr;gap:12px}.pair img{width:100%;border:1px solid #ddd}.pair figure{margin:0;min-width:0}.jump{display:flex;gap:16px;flex-wrap:wrap;background:white;padding:12px;border:1px solid #ddd}.jump a{font-weight:650}.command{background:#f3f3f8;padding:12px;overflow-wrap:anywhere}a{color:#4b269f}details{margin-top:12px}summary{font-weight:650;cursor:pointer}li{margin:5px 0}.section{border-top:2px solid #d4d1df;padding-top:22px;margin-top:32px}.limit{color:#564f6b}@media(max-width:800px){main{padding:12px}.pair{grid-template-columns:1fr}.grid{grid-template-columns:minmax(0,1fr)}.toolbar input[type=search],.toolbar input:not([type]){min-width:0}header{padding:22px}}</style>'''
    nav=f'<div class="nav"><a href="index.html">Catalog home</a><a href="templates.html">Fixed-layout templates ({len(templates)})</a><a href="components.html">Component groups</a></div>'
    (OUT/'templates.html').write_text(template_page(version,templates,common,nav))
    (OUT/'components.html').write_text(component_page(version,comp_rows,gauge,family_rows,recipes,common,nav))
    (OUT/'index.html').write_text(home_page(version,len(templates),len(comp_rows),common,nav))
    subprocess.run([sys.executable,str(ROOT/'scripts/build-wmds-release-gallery.py'),'--out',str(OUT)],check=True)
    print(f'Built local {version}: {len(templates)} templates, {len(comp_rows)} source-bound component occurrences, {sum(not x["visual_only"] for x in comp_rows)} executable, {sum(x["visual_only"] for x in comp_rows)} visual-only, {len(family_rows)} compose examples.')

def template_page(version,rows,css,nav):
 data=json.dumps(rows,ensure_ascii=False).replace('<','\\u003c')
 return f'''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Source templates · {version}</title>{css}<header><h1>Fixed-layout source templates</h1><p>{version} · All {len(rows)} source contracts, searchable by title, source, ID and category.</p>{nav}</header><main><section class="intro"><strong>What this gallery supports:</strong> filling named content slots and explicit bounded style profiles on the original fixed scene. It does not qualify arbitrary copy capacity or change row, card, shape or connector counts. Choose a template close to your content; run a build and inspect fit before use.</section><div class="toolbar"><input id="q" placeholder="Search name, ID, source, category" aria-label="Search"><select id="source"><option value="">All sources</option></select><select id="category"><option value="">All categories</option></select><span class="count" id="count"></span></div><div id="cards" class="grid"></div></main><script>const rows={data};const esc=s=>String(s??'').replace(/[&<>\"']/g,c=>({{'&':'&amp;','<':'&lt;','>':'&gt;','\"':'&quot;',"'":'&#39;'}}[c]));for(const [id,key] of [['source','source'],['category','category']])for(const v of [...new Set(rows.map(x=>key==='source'?String(x[key]||'').split(':')[0]:x[key]).filter(Boolean))].sort()){{let o=document.createElement('option');o.value=v;o.textContent=v;document.getElementById(id).append(o)}}function draw(){{let q=document.querySelector('#q').value.toLowerCase(),s=document.querySelector('#source').value,c=document.querySelector('#category').value;let a=rows.filter(x=>(!s||String(x.source).split(':')[0]===s)&&(!c||x.category===c)&&JSON.stringify([x.id,x.name,x.source,x.category]).toLowerCase().includes(q));document.querySelector('#count').textContent=a.length+' / '+rows.length;document.querySelector('#cards').innerHTML=a.map(x=>`<article><h2>${{esc(x.name)}}</h2><div class="meta">${{esc(x.id)}} · ${{esc(x.source)}} · ${{esc(x.category)}}</div><div class="tags"><span class="tag">${{esc(x.slots)}} content slots</span><span class="tag">fixed geometry</span>${{(x.profiles||[]).map(p=>`<span class="tag">style: ${{esc(p)}}</span>`).join('')}}</div><div class="previewbox">${{x.adapted_png?`<div class="pair"><figure><small>Source</small><a href="${{esc(x.source_png)}}"><a href="${{esc(x.source_png)}}"><img loading="lazy" src="${{esc(x.source_png)}}" alt="Source slide"></a></a></figure><figure><small>Reviewed illustrative content</small><a href="${{esc(x.adapted_png)}}"><a href="${{esc(x.adapted_png)}}"><img loading="lazy" src="${{esc(x.adapted_png)}}" alt="Reviewed adaptation"></a></a></figure></div>`:`<a href="${{esc(x.source_png)}}"><a href="${{esc(x.source_png)}}"><img class="preview" loading="lazy" src="${{esc(x.source_png)}}" alt="Original source preview"></a></a>`}}</div><p class="meta">${{esc(x.review_status||'Source preview')}}</p><p><a href="${{esc(x.values)}}" download>Example values</a> · <a href="${{esc(x.contract)}}">Contract</a></p><details><summary>Controls and limits</summary><p>${{esc(x.customization_boundary)}}</p><ul>${{(x.limitations||[]).map(z=>`<li>${{esc(z)}}</li>`).join('')}}</ul></details><p class="command"><code>${{esc(x.command)}}</code></p></article>`).join('')}}for(const x of ['q','source','category'])document.getElementById(x).addEventListener('input',draw);draw();</script></html>'''

def component_page(version,rows,gauge,families,recipes,css,nav):
 recipe_html='<section class="section" id="measured-recipes"><h2>Measured component recipes</h2><p>Editable building blocks and dense composition examples. Read the named guide under the installed root from <code>pptxgengo paths</code>. Copy a spec, supply your content, then use compose probe → measure → build → verify. These are separate from the 132 source group occurrences.</p><div class="grid">'+''.join('<article><h3>'+esc(r['name'])+'</h3><p>'+esc(r['controls'])+'</p><p class="limit">'+esc(r['limitations'])+'</p><p><a href="'+esc(r['spec'])+'" download>Recipe JSON</a></p><p class="meta">Installed guide: <code>'+esc(r['installed_guide'])+'</code></p><p class="command"><code>pptxgengo compose probe --spec &quot;$release_root/'+esc(r['installed_spec'])+'&quot; --out /absolute/path/new-probe</code></p></article>' for r in recipes)+'</div></section>'
 data=json.dumps(rows,ensure_ascii=False).replace('<','\\u003c'); famdata=json.dumps(families,ensure_ascii=False).replace('<','\\u003c'); gd=json.dumps(gauge,ensure_ascii=False).replace('<','\\u003c')
 return f'''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Component groups · {version}</title>{css}<header><h1>Component groups and composition</h1><p>{version} · Source-bound component occurrences and separately labeled new-slide builders.</p>{nav}</header><main><section class="intro"><strong>How to read the count:</strong> the source inventory contains 132 template-bound group occurrences: 124 with executable component contracts and 8 visual-only references. Repeated group names across source slides are not unique reusable designs. A supported occurrence is bound to its original source project/objects and fixed geometry; portability to other decks is not claimed. Text and declared styles may change only as its contract permits; shapes, counts, relationships and artwork are otherwise retained.</section><nav class="jump"><a href="#measured-recipes">Measured recipes</a> · <a href="#source-groups">Source component groups</a><a href="#t045-gauge">T045 gauge example</a><a href="#new-slide-composition">New-slide composition builders</a></nav><section class="section" id="t045-gauge"><h2>Original source component example: T045 gauges</h2><article><div class="pair"><img src="{gauge['preview']}" alt="Five original source gauges"><div><p>{gauge['description']}</p><div class="tags">{''.join(f'<span class="tag">{esc(t)}</span>' for t in gauge['controls'])}</div><p><a href="{gauge['values']}">Gauge values example</a> · <a href="assets/gauge-authoring.md">Instructions</a></p><p class="command"><code>{esc(gauge['command'])}</code></p></div></div></article></section><section class="section" id="new-slide-composition"><h2>Supported new-slide composition families</h2><p class="limit">These five semantic builders create new editable slides inspired by source patterns; they do not import or restructure the source slide. Only the exact reviewed examples and stated controls are represented. Arbitrary content capacity is not qualified.</p><div id="families" class="grid"></div></section>{recipe_html}<section class="section" id="source-groups"><h2>Source-bound component occurrences</h2><div class="toolbar"><input id="q" placeholder="Search component, source, template, slot" aria-label="Search"><select id="source"><option value="">All sources</option></select><label><input type="checkbox" id="showVisual"> Show 8 visual-only references</label><span id="count" class="count"></span></div><div id="cards" class="grid"></div></section></main><script>const rows={data},familyRows={famdata},gauge={gd};const esc=s=>String(s??'').replace(/[&<>\"']/g,c=>({{'&':'&amp;','<':'&lt;','>':'&gt;','\"':'&quot;',"'":'&#39;'}}[c]));const fams=[...new Set(familyRows.map(x=>x.family))];document.querySelector('#families').innerHTML=fams.map(f=>{{let a=familyRows.filter(x=>x.family===f),sample=a.find(x=>x.variant==='normal')||a[0];return `<article><h3>${{esc(f[0].toUpperCase()+f.slice(1))}} · ${{a.length}} reviewed examples</h3><p>Bounded builder examples. Preview shows one reviewed variant; the other variants and controls are enumerated below.</p><img class="preview" loading="lazy" src="${{esc(sample.render_asset||'')}}"><p class="meta">Source pattern: ${{esc(sample.source_template_id)}} · example status: ${{esc(sample.status)}}</p><details><summary>Reviewed variants, controls and evidence</summary><ul>${{a.map(x=>`<li><strong>${{esc(x.variant)}}</strong>: ${{esc(JSON.stringify(x.controls||{{}}))}} · <a href="${{esc(x.render_asset||'')}}">render</a>${{x.spec_asset?` · <a href="${{esc(x.spec_asset)}}">spec</a>`:''}}${{x.values_asset?` · <a href="${{esc(x.values_asset)}}">values</a>`:''}}${{x.native_verification_asset?` · <a href="${{esc(x.native_verification_asset)}}">native verification</a>`:''}}<br>${{(x.limitations||[]).map(esc).join(' ')}}</li>`).join('')}}</ul></details></article>`}}).join('');for(const v of [...new Set(rows.map(x=>x.source.split(':')[0]))].sort()){{let o=document.createElement('option');o.value=v;o.textContent=v;document.querySelector('#source').append(o)}}function draw(){{let q=document.querySelector('#q').value.toLowerCase(),s=document.querySelector('#source').value,show=document.querySelector('#showVisual').checked;let a=rows.filter(x=>(show||!x.visual_only)&&(!s||x.source.split(':')[0]===s)&&JSON.stringify([x.id,x.name,x.purpose,x.source,x.template_id,x.slots]).toLowerCase().includes(q));document.querySelector('#count').textContent=a.length+' / '+rows.length+' occurrences';document.querySelector('#cards').innerHTML=a.map(x=>`<article><h3>${{esc(x.name)}}</h3><div class="meta">${{esc(x.id)}} · ${{esc(x.source)}} · ${{esc(x.template_id)}}</div><div class="focus"><img loading="lazy" src="${{esc(x.preview)}}" alt="Source template preview"><svg viewBox="0 0 1920 1080" preserveAspectRatio="none" aria-label="Magenta outlines mark object IDs in this source-bound group">${{(x.focus_boxes||[]).map(b=>`<rect x="${{b[0]}}" y="${{b[1]}}" width="${{b[2]}}" height="${{b[3]}}" rx="2" fill="none" stroke="#f900d3" stroke-width="5" stroke-dasharray="9 6"/>`).join('')}}</svg></div><p>${{esc(x.purpose)}}</p><div class="tags"><span class="tag">${{x.contract_available?'Executable source contract':'Visual-only reference'}}</span><span class="tag">${{x.object_ids.length}} source objects</span><span class="tag">${{x.slots.length}} text slots</span><span class="tag">fixed count and geometry</span><span class="tag">${{x.focus_object_boxes}} source-object bounds outlined${{x.nested_geometry_present?'; nested group bounds omitted':''}}</span></div><details><summary>Scope and inspect</summary><p><strong>Source objects:</strong> ${{esc(x.object_ids.join(', '))}}</p><p><strong>Bound slots:</strong> ${{esc(x.slots.join(', ')||'none; source artwork retained')}}</p><p><strong>Styles:</strong> ${{esc(x.profiles.join(', ')||'no style profiles declared')}}</p><p>Source-bound group; its source project and slide geometry are required. This group is not qualified as a portable component or for arbitrary resizing/count changes.</p>${{x.contract_path?`<p><a href="${{esc(x.contract_path)}}">Inspect component contract</a></p><p class="command"><code>pptxgengo component inspect --project "$root/${{esc(x.source_project||'SOURCE_PROJECT')}}" --contract "$root/${{esc(x.contract_path)}}"</code></p>`:'<p>No executable component contract; visual reference only.</p>'}}</details></article>`).join('')}}for(const x of ['q','source','showVisual'])document.getElementById(x).addEventListener('input',draw);draw();</script></html>'''

def home_page(version,nt,nc,css,nav):
 return f'''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>West Monroe presentation toolkit · {version}</title>{css}<header><h1>West Monroe presentation toolkit</h1><p>Local release {version} · choose a fixed-layout template or a supported component workflow.</p>{nav}</header><main><section class="intro"><h2>Choose an authoring route</h2><div class="grid"><article><h3>WM design system</h3><p>Browse 166 active modern slide designs, compare reviewed examples, and download editable content or illustrated compositions. One retired design remains available for compatibility.</p><p><a href="design-system.html">Browse the WM design system →</a></p><code>pptxgengo catalog --design-system --open</code></article><article><h3>1 · Fill a source template</h3><p>Browse {nt} source slide contracts. Replace declared text and bounded style values while retaining the original slide layout. Counts and geometry stay fixed.</p><p><a href="templates.html">Browse all source templates →</a></p><code>pptxgengo catalog --templates --open</code></article><article><h3>2 · Edit a supported component or compose a new slide</h3><p>Browse source-bound groups with explicit object mapping, the T045 gauge-control example, and five bounded semantic builders for new slides.</p><p><a href="components.html">Browse component groups and composition →</a></p><code>pptxgengo catalog --components --open</code></article></div><p><strong>Support boundary:</strong> the {nt} templates are usable fixed-layout starting points; arbitrary content capacity is not qualified. The {nc} component rows are source-bound occurrences, not unique portable designs. The five semantic builders create a new composition inspired by a source pattern and are not a claim of arbitrary-content support.</p></section><section class="section"><h2>From terminal</h2><p><code>pptxgengo catalog --templates</code> lists or prints the template gallery path; add <code>--open</code> to open it. Use <code>pptxgengo catalog --components --open</code> to open component workflows. Inspect a layout with <code>pptxgengo template inspect --id TEMPLATE_ID</code>.</p><p>All gallery assets are local and linked relative to this page.</p></section></main></html>'''
if __name__=='__main__': main()
