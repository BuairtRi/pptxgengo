#!/usr/bin/env python3
"""Package the fully reviewed 167-design WMDS gallery, with editable local inputs.
Pure stdlib. Does not generate slides, edit images or perform native review.
--allow-pending is an explicit developer preview; normal release requires every
source/alternate pair to be accepted_paired_specimens.
"""
import argparse
import copy
import hashlib
import html
import json
from pathlib import Path
import re
import shutil
import sys

ROOT = Path(__file__).resolve().parents[1]
SOURCE_COMMIT = '7bdcaee5030a12275a1f881a8542f4d302d207df'
SOURCE_REVISION = 'wmds-library.v2'
ACCEPTED = 'accepted_paired_specimens'
EXPECTED = 167
KEY = re.compile(r'^[a-z0-9][a-z0-9-]*/[a-z0-9][a-z0-9-]*$')
NAV_LINK = '<a href="design-system.html">WM design system (166 active designs)</a>'


def read(path):
    with path.open(encoding='utf-8') as f:
        return json.load(f)


def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n', encoding='utf-8')


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def one_slide(document, slide):
    # Preserve all document-level fields and explicit composition nodes. Preview
    # PNGs are separate assets; no slide is flattened into a picture here.
    result = copy.deepcopy(document)
    result['slides'] = [copy.deepcopy(slide)]
    return result


def planned_nav_updates(out):
    """Preflight exact nav insertions before writing any release artifacts."""
    result = {}
    for name in ('index.html', 'templates.html', 'components.html'):
        path = out / name
        if not path.exists():
            continue
        source = path.read_text(encoding='utf-8')
        if 'href="design-system.html"' in source:
            continue
        opening = '<div class="nav">'
        require(source.count(opening) == 1,
                f'{path}: expected one exact catalog navigation opening')
        start = source.index(opening) + len(opening)
        end = source.find('</div>', start)
        require(end >= start, f'{path}: catalog navigation closing tag missing')
        require('<div' not in source[start:end], f'{path}: nested navigation is unsupported')
        result[path] = source[:end] + NAV_LINK + source[end:]
    return result


def gallery_page(index, version, allow_pending):
    # Inline JSON makes filtering work under file:// without fetch/CORS.
    encoded = json.dumps(index, ensure_ascii=False, separators=(',', ':'))
    encoded = encoded.replace('<', '\\u003c').replace('>', '\\u003e').replace('&', '\\u0026')
    warning = ('<p class="dev">Developer preview: some examples await PowerPoint review.</p>'
               if allow_pending and index['qualification']['accepted_pairs'] != EXPECTED else '')
    return '''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>West Monroe design system · VERSION</title>
<style>
*{box-sizing:border-box}body{margin:0;color:#070154;background:#f5f7fb;font:16px/1.5 system-ui,sans-serif}
header{background:#070154;color:white;padding:26px max(24px,4vw)}h1{margin:0;font-size:32px}header p{max-width:1000px}.nav{display:flex;gap:20px;flex-wrap:wrap}.nav a{color:white}
main{max-width:1600px;margin:auto;padding:24px}.intro{background:white;padding:20px;border:1px solid #ced7e6;margin-bottom:20px}.dev{background:#ffdc00;color:#070154;padding:10px}
.toolbar{display:flex;gap:12px;align-items:center;flex-wrap:wrap;position:sticky;top:0;background:white;padding:14px;z-index:3;border:1px solid #ced7e6}
input[type=search]{flex:1;min-width:240px}input,select{font:inherit;padding:9px;border:1px solid #aab5c9;border-radius:3px}.toolbar label{white-space:nowrap}.count{font-size:14px}
article{background:white;border:1px solid #ced7e6;margin-top:20px;padding:20px}h2{font-size:23px;margin:0 0 6px}h3{font-size:17px;margin:18px 0 8px}.meta{color:#52678f;font-size:14px}.tags{display:flex;gap:7px;flex-wrap:wrap;margin:8px 0}.tag{background:#e7eef8;padding:3px 8px;font-size:13px}
.pair{display:grid;grid-template-columns:1fr 1fr;gap:18px}figure{margin:0;min-width:0}figure img{width:100%;display:block;border:1px solid #ced7e6}figcaption{padding:8px 0;font-weight:650}
a{color:#0047ff}.links{display:flex;gap:15px;flex-wrap:wrap}.controls{display:grid;grid-template-columns:1fr 1fr;gap:18px}.controls p{margin:8px 0}.command{background:#f3f6fa;padding:12px;white-space:pre-wrap;overflow-wrap:anywhere;font:13px/1.6 ui-monospace,monospace}
details{margin-top:12px}summary{cursor:pointer;font-weight:650}.slotlist{max-height:280px;overflow:auto;font-size:13px}.slotlist li{margin:4px 0}.empty{padding:28px}.deprecated{border-top:4px solid #ffdc00}
@media(max-width:850px){main{padding:12px}.pair,.controls{grid-template-columns:1fr}header{padding:22px}input[type=search]{min-width:0;width:100%}.toolbar{position:static}}
</style></head><body><header><h1>West Monroe design system</h1>
<p>Choose a slide design, compare two examples, and start with editable content.</p>
<div class="nav"><a href="index.html">Catalog home</a><a href="templates.html">Source templates</a><a href="components.html">Components</a><a href="design-system.html">WM design system</a></div>
</header><main><section class="intro"><strong>166 active designs · local release VERSION</strong>
<p>The source and changed-content examples were reviewed in PowerPoint. They show the intended layout and sample content. Review your own text and images for fit before sharing.</p>
<p>Download an example’s content to fill its template, or its illustrated slide to keep the sample’s editable diagrams and composed previews.</p>WARNING</section>
<div class="toolbar"><input id="search" type="search" placeholder="Search designs" aria-label="Search designs"><select id="family" aria-label="Design family"><option value="">All families</option></select>
<select id="change" aria-label="Design update"><option value="">All updates</option><option value="added">Added</option><option value="revised">Revised</option><option value="unchanged">Unchanged</option></select>
<label><input id="deprecated" type="checkbox"> Show retired design</label><span class="count" id="count"></span></div>
<div id="designs"></div></main><script type="application/json" id="gallery-data">DATA</script>
<script>
'use strict';
const data=JSON.parse(document.getElementById('gallery-data').textContent);
const q=document.getElementById('search'),family=document.getElementById('family'),change=document.getElementById('change'),retired=document.getElementById('deprecated'),list=document.getElementById('designs');
function el(tag,text,cls){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e;}
function link(label,path){const a=el('a',label);a.href=path;return a;}
function command(parent,text){parent.append(el('pre',text,'command'));}
[...new Set(data.designs.map(d=>d.family))].sort().forEach(f=>{const o=el('option',f[0].toUpperCase()+f.slice(1));o.value=f;family.append(o);});
function paint(){const term=q.value.trim().toLowerCase();const shown=data.designs.filter(d=>(retired.checked||d.status!=='deprecated')&&(!family.value||d.family===family.value)&&(!change.value||d.change===change.value)&&(!term||[d.template,d.name,d.family].join(' ').toLowerCase().includes(term)));
list.replaceChildren();document.getElementById('count').textContent=shown.length+' designs';
for(const d of shown){const a=el('article',undefined,d.status==='deprecated'?'deprecated':'');a.id=d.template.replaceAll('/','--');a.append(el('h2',d.name));a.append(el('div',d.template+' · '+d.family,'meta'));
const tags=el('div',undefined,'tags');for(const t of [d.change,d.status==='deprecated'?'Retired':'Active',d.native_review==='accepted_paired_specimens'?'Two examples reviewed':'Review pending'])tags.append(el('span',t,'tag'));a.append(tags);
if(d.replaced_by)a.append(el('p','Use '+d.replaced_by+' for new slides.'));
const pair=el('div',undefined,'pair');for(const kind of ['source','alternate']){const f=el('figure');const im=document.createElement('img');im.src=d[kind+'_preview'];im.alt=d.name+' · '+(kind==='source'?'source example':'changed-content example');im.loading='lazy';f.append(im,el('figcaption',kind==='source'?'Source example':'Changed-content example'));pair.append(f);}a.append(pair);
const controls=el('div',undefined,'controls');for(const kind of ['source','alternate']){const c=el('section');c.append(el('h3',kind==='source'?'Use the source example':'Use the changed-content example'));const links=el('div',undefined,'links');links.append(link('Download content',d[kind+'_values']),link('Download illustrated slide',d[kind+'_foundation']));c.append(links);
command(c,'pptxgengo design template --spec '+kind+'-values.json --out NEW-DIR');command(c,'pptxgengo design build --spec '+kind+'.foundation.json --out NEW-DIR');controls.append(c);}a.append(controls);
if(d.value_schema_kind==='typed_values'){const counts=Object.entries(d.exact_counts).map(([name,count])=>'exactly '+count+' '+name).join(' · ');a.append(el('p',d.typed_fields.length+' editable content fields · '+counts+'. Keep each card’s key, title and body; review replacement content for fit.','meta'));}
else a.append(el('p',d.slot_count+' editable content fields · '+d.array_count+' fixed lists or row groups. Keep the design’s item counts and layout; review replacement content for fit.','meta'));
const detail=el('details');detail.append(el('summary','Content fields'));const items=el('ul',undefined,'slotlist');if(d.value_schema_kind==='typed_values'){for(const f of d.typed_fields)items.append(el('li',f.name+' · '+f.description));for(const [name,count] of Object.entries(d.exact_counts))items.append(el('li',name+' · exactly '+count+' items'));}
else for(const s of d.slots)items.append(el('li',s.name+' · '+s.kind+(s.allow_empty?' · may be empty':'')));detail.append(items,link('Download design contract',d.contract));a.append(detail);list.append(a);}
if(!shown.length)list.append(el('p','No designs match these filters.','empty'));}
for(const x of [q,family,change,retired])x.addEventListener('input',paint);paint();
</script></body></html>'''.replace('VERSION', html.escape(version)).replace('WARNING', warning).replace('DATA', encoded)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--packet',type=Path,default=ROOT/'samples/wmds-refresh-slice4-20261002')
    parser.add_argument('--out',type=Path,default=ROOT/'release/catalog')
    parser.add_argument('--version',default=None)
    parser.add_argument('--allow-pending',action='store_true',help='Explicit developer gallery; never implies native acceptance')
    args=parser.parse_args()
    version=args.version or (ROOT/'release/VERSION').read_text(encoding='utf-8').strip()
    require(bool(version),'release version is empty')
    packet=args.packet.resolve();out=args.out.resolve()
    document=read(packet/'reference.foundation.json');manifest=read(packet/'manifest.json')
    catalog=read(packet/'catalog.json');bound=read(packet/'bound-content.json')
    require(isinstance(catalog,list) and len(catalog)==EXPECTED,'catalog must contain exactly167 designs including retired')
    require(isinstance(manifest,list) and len(manifest)==EXPECTED,'manifest must contain exactly167 paired design records')
    require(document.get('schema')=='pptxgengo.wmds-foundation.v1','unexpected foundation schema')
    require(bound.get('schema')=='pptxgengo.wmds-template-document.v1','unexpected bound-content schema')
    require(len(document.get('slides',[]))==EXPECTED*2,'foundation must contain334 specimens')
    require(len(bound.get('slides',[]))==EXPECTED*2,'bound content must contain334 specimens')
    bundle=read(ROOT/'library/wm-design-system/v2/bundle.json')
    require(bundle.get('source_commit')==SOURCE_COMMIT and bundle.get('source_revision')==SOURCE_REVISION,'frozen bundle source commit/revision mismatch')
    contracts={c['key']:c for c in catalog};require(len(contracts)==EXPECTED,'catalog has duplicate keys')
    require(sum(c.get('status')!='deprecated' for c in catalog)==166,'expected166 active designs and one retired design')
    records={m['template']:m for m in manifest};require(set(records)==set(contracts) and len(records)==EXPECTED,'manifest/catalog keys must match exactly')
    accepted=sum(m.get('native_review')==ACCEPTED for m in manifest)
    require(args.allow_pending or accepted==EXPECTED,'all167 paired specimens must be accepted; use --allow-pending only for developer preview')
    bound_by_id={s['id']:s for s in bound['slides']};require(len(bound_by_id)==EXPECTED*2,'bound specimen IDs must be unique')
    nav_updates=planned_nav_updates(out)
    root_index_path=out/'index.json'
    root_index=read(root_index_path) if root_index_path.exists() else {'schema':'pptxgengo.release-catalog.v2','version':version}
    require(isinstance(root_index,dict),'release root index must be an object')
    used_pages=set();designs=[];prepared=[]
    for key in sorted(contracts):
        require(KEY.fullmatch(key) is not None,f'unsafe template key:{key}')
        contract=contracts[key];record=records[key]
        require(contract.get('source_revision')==SOURCE_REVISION,f'{key}: wrong source revision')
        frozen=ROOT/'library/wm-design-system/v2/source'/contract['source_file']
        require(frozen.is_file() and sha(frozen)==contract.get('source_sha256'),f'{key}: source hash does not match pinned commit')
        require(record.get('family')==contract.get('family'),f'{key}: family mismatch')
        require(record.get('source_commit')==SOURCE_COMMIT,f'{key}: source commit missing or mismatch')
        lifecycle=record.get('status')
        require(lifecycle==contract.get('status','active'),f'{key}: lifecycle mismatch')
        change=record.get('change')
        require(change in ('added','revised','unchanged','deprecated'),f'{key}: manifest must supply an explicit change state:{change}')
        relative=Path('design-system')/key
        item={'template':key,'name':contract.get('name',key),'family':contract['family'],'status':lifecycle,'change':change,
              'native_review':record.get('native_review','pending'),'replaced_by':contract.get('replaced_by'),
              'contract':(relative/'contract.json').as_posix(),'slot_count':len(contract.get('slots',[])),
              'array_count':len(contract.get('arrays',[])),'slots':contract.get('slots',[])}
        value_schema=contract.get('value_schema',{})
        if value_schema.get('kind')=='typed_values':
            require(key in ('cards/3','cards/4'),f'{key}: unsupported retained typed content schema')
            fields=value_schema.get('fields');counts=value_schema.get('exact_counts')
            require(isinstance(fields,dict) and set(fields)=={'cards','eyebrow','title'},f'{key}: unexpected typed card fields')
            require(all(isinstance(v,str) and v for v in fields.values()),f'{key}: typed field descriptions missing')
            require(counts=={'cards':int(key.split('/')[1])},f'{key}: exact card count mismatch')
            item['value_schema_kind']='typed_values'
            item['typed_fields']=[{'name':name,'description':description} for name,description in fields.items()]
            item['exact_counts']=copy.deepcopy(counts)
            item['slot_count']=len(fields)
            item['array_count']=len(counts)
        for kind in ('source','alternate'):
            page=record.get(kind+'_page')
            require(type(page) is int and 1<=page<=EXPECTED*2,f'{key}: invalid {kind} page')
            require(page not in used_pages,f'duplicate specimen page:{page}');used_pages.add(page)
            slide=document['slides'][page-1];binding=slide.get('template_binding',{})
            require(binding.get('template')==key and binding.get('source_revision')==SOURCE_REVISION,f'{key}: page{page} binding mismatch')
            require(binding.get('source_sha256')==contract['source_sha256'],f'{key}: page{page} source hash mismatch')
            require(slide['id'] in bound_by_id,f'{key}: page{page} bound specimen ID missing')
            values=bound_by_id[slide['id']]
            require(values.get('template')==key,f'{key}: bound template mismatch')
            png=packet/'native-pages'/f'slide-{page:03d}.png'
            require(png.is_file(),f'{key}: missing native preview:{png}')
            item[kind+'_page']=page;item[kind+'_preview']=(relative/(kind+'.png')).as_posix()
            item[kind+'_values']=(relative/(kind+'-values.json')).as_posix()
            item[kind+'_foundation']=(relative/(kind+'.foundation.json')).as_posix()
            item[kind+'_preview_sha256']=sha(png)
            prepared.append((out/relative/(kind+'.png'),png,
                             out/relative/(kind+'-values.json'),one_slide(bound,values),
                             out/relative/(kind+'.foundation.json'),one_slide(document,slide)))
        designs.append(item)
    require(len(used_pages)==EXPECTED*2,'all334 native specimen pages must be represented')
    index={'schema':'pptxgengo.wmds-release-gallery.v1','version':version,'source_revision':SOURCE_REVISION,
           'source_commit':SOURCE_COMMIT,'entries':EXPECTED,'active':166,'deprecated':1,
           'qualification':{'accepted_pairs':accepted,'reviewed_specimens':accepted*2,'arbitrary_content_qualified':False,
                            'scope':'Two synthetic sample slides per design; review replacement content for fit.'},
           'developer_preview':args.allow_pending and accepted!=EXPECTED,'designs':designs}
    # All input validation completes before any release/catalog mutations.
    out.mkdir(parents=True,exist_ok=True)
    for png_out,png_in,values_out,values_doc,foundation_out,foundation_doc in prepared:
        png_out.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(png_in,png_out)
        write(values_out,values_doc);write(foundation_out,foundation_doc)
    for contract in catalog:write(out/'design-system'/contract['key']/'contract.json',contract)
    write(out/'design-system/index.json',index)
    (out/'design-system.html').write_text(gallery_page(index,version,args.allow_pending),encoding='utf-8')
    root_index['version']=version
    root_index['design_system']={'entries':EXPECTED,'active':166,'deprecated':1,'gallery':'design-system.html',
                                'index':'design-system/index.json','source_revision':SOURCE_REVISION,'source_commit':SOURCE_COMMIT,
                                'accepted_pairs':accepted,'arbitrary_content_qualified':False}
    write(root_index_path,root_index)
    for path,source in nav_updates.items():path.write_text(source,encoding='utf-8')
    print(f'Packaged {EXPECTED} designs / {EXPECTED*2} editable specimens: {out / "design-system.html"}')
    return 0

if __name__=='__main__':
    try:
        sys.exit(main())
    except (ValueError,KeyError,OSError,json.JSONDecodeError) as error:
        print(f'WMDS gallery: {error}',file=sys.stderr)
        sys.exit(1)
