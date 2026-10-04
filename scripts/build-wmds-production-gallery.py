#!/usr/bin/env python3
"""Package a source-pinned current gallery; native previews and review are optional evidence.

Input is the Go library-source-reference directory and the corresponding bound
library-reference directory. No sample content is substituted for caller content.
Historical v2 paired evidence is retained in its own gallery, never relabeled v3.
"""
import argparse, copy, hashlib, html, json, shutil
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def read(p): return json.loads(p.read_text())
def write(p, obj):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(obj, indent=2, ensure_ascii=False)+'\n')
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def require(c, message):
    if not c: raise ValueError(message)
def one(doc, slide):
    d=copy.deepcopy({k:v for k,v in doc.items() if k!='slides'}); d['slides']=[copy.deepcopy(slide)]; return d

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--source',type=Path,required=True)
    ap.add_argument('--bound',type=Path,required=True)
    ap.add_argument('--out',type=Path,default=ROOT/'release/catalog')
    ap.add_argument('--bundle',type=Path,default=ROOT/'library/wm-design-system/v3')
    ap.add_argument('--review',type=Path,help='Optional explicit per-template native review receipt')
    ap.add_argument('--version',help='Explicit candidate package version; defaults to release/VERSION')
    args=ap.parse_args()
    version=args.version or (ROOT/'release/VERSION').read_text().strip()
    doc=read(args.source/'compiled-document.json'); report=read(args.source/'layout-report.json')
    catalog=read(args.source/'catalog.json'); bound=read(args.bound/'template-content.json')
    bundle=read(args.bundle/'bundle.json')
    revision=bundle['source_revision'];commit=bundle['source_commit']
    require(report['source_revision']==revision,'reference does not match bundle revision')
    require(len(doc['slides'])==len(catalog),'reference/catalog counts differ')
    contracts={x['key']:x for x in catalog}
    require(len(contracts)==len(catalog),'duplicate canonical keys')
    source_bindings={s['template_binding']['template']:s['template_binding'] for s in doc['slides']}
    require(len(source_bindings)==len(doc['slides']) and set(source_bindings)==set(contracts),
            'source slide identities do not match source catalog')
    require(report['source_commit']==commit,'source reference commit mismatch')
    require(sha(args.source/'library-reference.pptx')==report['pptx_sha256'],
            'source reference output hash mismatch')
    for key,b in source_bindings.items():
        c=contracts[key]
        require(b['source_revision']==revision and b['source_file']==c['source_file'] and
                b['source_sha256']==c['source_sha256'],f'{key}: source binding identity mismatch')
    bykey={s['template']:s for s in bound['slides']}
    require(len(bykey)==len(bound['slides']),'duplicate bound specimen keys')
    expected={k for k,c in contracts.items() if c.get('status','active')!='deprecated'}
    require(set(bykey)==expected,'bound specimen keys do not match active source catalog')
    bound_doc=read(args.bound/'compiled-document.json')
    bound_report=read(args.bound/'layout-report.json')
    require(bound_report['source_revision']==revision and bound_report['source_commit']==commit,
            'bound reference source identity mismatch')
    bound_bindings={s['template_binding']['template']:s['template_binding'] for s in bound_doc['slides']}
    require(len(bound_bindings)==len(bound_doc['slides']) and set(bound_bindings)==expected,
            'bound reference identities do not match source catalog')
    for key,b in bound_bindings.items():
        c=contracts[key]
        require(b['source_revision']==revision and b['source_file']==c['source_file'] and
                b['source_sha256']==c['source_sha256'],f'{key}: bound source identity mismatch')
    require(sha(args.bound/'library-reference.pptx')==bound_report['pptx_sha256'],
            'bound reference output hash mismatch')
    reviews=read(args.review) if args.review else {}
    require(set(reviews)<=set(contracts),'review has unknown template identities')
    prepared=[];designs=[]
    old_index=args.out/'design-system/index.json'
    for page,s in enumerate(doc['slides'],1):
        binding=s['template_binding'];key=binding['template']; c=contracts[key]
        require(c['source_revision']==revision and binding['source_sha256']==c['source_sha256'],f'{key}: source identity mismatch')
        require('/'.join(key.split('/'))==key and all(x and x not in ('.','..') for x in key.split('/')),f'{key}: unsafe key')
        source=args.bundle/'source'/c['source_file']
        require(sha(source)==c['source_sha256'],f'{key}: source hash drift')
        rel=Path('design-system')/key
        item={'template':key,'name':c['name'],'family':c['family'],'status':c.get('status','active'),
              'source_commit':commit,'source_revision':revision,'source_page':page,
              'native_review':'not_reviewed','contract':str(rel/'contract.json'),
              'source_foundation':str(rel/'source.foundation.json'),
              'slot_count':len(c.get('slots',[])),'array_count':len(c.get('arrays',[])),
              'discovery':c.get('discovery',{}),'replaced_by':c.get('replaced_by')}
        png=args.source/'native-pages'/f'slide-{page:03d}.png'
        review=reviews.get(key,{})
        if png.exists():
            item['source_preview']=str(rel/'source.png');item['source_preview_sha256']=sha(png)
            prepared.append((args.out/rel/'source.png',png,None))
            item['native_review']='native_exported_visual_review_pending'
            if review.get('status')=='reviewed_source_specimen':
                require(review.get('preview_sha256')==sha(png),f'{key}: review preview hash mismatch')
                item['native_review']='reviewed_source_specimen'
        else:
            require(not review,f'{key}: review exists without native preview')
        if key in bykey:
            item['source_values']=str(rel/'source-values.json')
            prepared.append((args.out/rel/'source-values.json',None,one(bound,bykey[key])))
        prepared.extend([(args.out/rel/'contract.json',None,c),(args.out/rel/'source.foundation.json',None,one(doc,s))])
        designs.append(item)
    # Inputs validated before updating the current projection.
    if old_index.exists() and read(old_index).get('source_revision')!=revision:
        previous=read(old_index)['source_revision'].rsplit('.',1)[-1]
        require(previous in ('v1','v2','v3','v4'),'unknown historical library revision')
        historical=args.out/('design-system-'+previous)
        if not historical.exists():
            shutil.copytree(args.out/'design-system',historical)
            def historic_links(value):
                if isinstance(value,dict):return {k:historic_links(v) for k,v in value.items()}
                if isinstance(value,list):return [historic_links(v) for v in value]
                if isinstance(value,str) and value.startswith('design-system/'):
                    return historical.name+'/'+value[len('design-system/'):]
                return value
            write(historical/'index.json',historic_links(read(historical/'index.json')))
        page=args.out/'design-system.html'
        historical_page=args.out/('design-system-'+previous+'.html')
        if page.exists() and not historical_page.exists():
            historical_page.write_text(page.read_text().replace('design-system/',historical.name+'/'))
    for p,src,value in prepared:
        p.parent.mkdir(parents=True,exist_ok=True)
        if src: shutil.copy2(src,p)
        else: write(p,value)
    active=sum(d['status']!='deprecated' for d in designs)
    reviewed=sum(d['native_review']=='reviewed_source_specimen' for d in designs)
    index={'schema':'pptxgengo.wmds-release-gallery.v1','version':version,'source_revision':revision,'source_commit':commit,
           'entries':len(designs),'active':active,'deprecated':len(designs)-active,
           'qualification':{'reviewed_source_specimens':reviewed,'arbitrary_content_qualified':False,
                            'scope':'Individual source specimens only. Supplied copy requires a separate fit/build and native review.'},
           'designs':designs}
    write(args.out/'design-system/index.json',index)
    encoded=json.dumps(index,ensure_ascii=False).replace('<',chr(92)+'u003c').replace('>',chr(92)+'u003e').replace('&',chr(92)+'u0026')
    page='''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>WM design system · VERSION</title>
<style>*{box-sizing:border-box}body{margin:0;background:#f5f7fb;color:#070154;font:16px/1.5 system-ui}header{padding:24px;background:#070154;color:white}header a{color:white;margin-right:20px}main{max-width:1440px;margin:auto;padding:24px}input,select{font:inherit;padding:10px;margin:8px 8px 8px 0}input{min-width:300px}article{background:white;padding:24px;margin:20px 0;border:1px solid #ced7e6}article img{width:100%;max-width:960px;display:block}h2{margin-top:0}a{color:#0047ff;margin-right:16px}.meta{color:#52678f}.pending{padding:30px;background:#eef2f8}pre{white-space:pre-wrap;overflow-wrap:anywhere}.toolbar{position:sticky;top:0;background:#f5f7fb}</style>
<header><h1>West Monroe design system</h1><p>COUNT designs · VERSION · pinned source REVISION</p><a href="index.html">Catalog</a><a href="templates.html">Legacy templates</a><a href="components.html">Components</a>HISTORY</header>
<main><p>Search by scenario or the shape of your content. Templates, frames and components are persistent library definitions. Maintain a deck in a single <code>deck.yaml</code>; custom layouts live in that deck's local templates.</p>
<p>Preview/review status is specific to the illustrated source specimen. Build actual-content alternatives with <code>pptxgengo design library-fit</code>; review new text and images before sharing.</p>
<div class="toolbar"><input id="q" placeholder="Search designs, roles or structures"><select id="family"><option value="">All families</option></select><span id="count"></span></div><div id="designs"></div></main>
<script type="application/json" id="data">DATA</script><script>
const data=JSON.parse(document.getElementById('data').textContent),q=document.getElementById('q'),family=document.getElementById('family'),list=document.getElementById('designs');
function el(tag,text){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e;}function link(label,path){const a=el('a',label);a.href=path;return a;}
[...new Set(data.designs.map(d=>d.family))].sort().forEach(f=>{const o=el('option',f);o.value=f;family.append(o)});
function paint(){const term=q.value.toLowerCase().trim();const shown=data.designs.filter(d=>(!family.value||d.family===family.value)&&(!term||JSON.stringify([d.template,d.name,d.family,d.discovery.content_roles,d.discovery.structures,d.discovery.visual_forms]).toLowerCase().includes(term)));list.replaceChildren();document.getElementById('count').textContent=shown.length+' designs';for(const d of shown){const a=el('article');a.append(el('h2',d.name));const m=el('p',d.template+' · '+d.native_review.replaceAll('_',' '));m.className='meta';a.append(m);if(d.source_preview){const im=document.createElement('img');im.src=d.source_preview;im.alt=d.name+' source specimen';im.loading='lazy';a.append(im)}else{const p=el('p','Native preview pending. Editable source and content contract are available.');p.className='pending';a.append(p)}const links=el('p');links.append(link('Content contract',d.contract),link('Editable composition',d.source_foundation));if(d.source_values)links.append(link('Example content',d.source_values));a.append(links,el('p',(d.discovery.content_roles||[]).join(' · ')),el('pre','pptxgengo design library-inspect --id '+d.template));if(d.replaced_by)a.append(el('p','Retained for compatibility; prefer '+d.replaced_by));list.append(a)}}q.addEventListener('input',paint);family.addEventListener('input',paint);paint();</script></html>'''
    history=''.join('<a href="'+p.name+'">Historical '+html.escape(p.stem.removeprefix('design-system-'))+' specimens</a>' for p in sorted(args.out.glob('design-system-v*.html')))
    for k,v in [('VERSION',html.escape(version)),('COUNT',str(len(designs))),('REVISION',html.escape(revision)),('HISTORY',history),('DATA',encoded)]:page=page.replace(k,v)
    (args.out/'design-system.html').write_text(page)
    root=args.out/'index.json';idx=read(root) if root.exists() else {'schema':'pptxgengo.release-catalog.v2'}
    idx['version']=version;idx['design_system']={k:index[k] for k in ['entries','active','deprecated','source_revision','source_commit','qualification']}
    idx['design_system'].update(gallery='design-system.html',index='design-system/index.json')
    write(root,idx)
    print(f'Packaged {len(designs)} definitions, {reviewed} reviewed source previews: {args.out}')
if __name__=='__main__':main()
