#!/usr/bin/env python3
"""Publish the explicitly requested source layouts and their real controls."""
import hashlib
import html
import json
import os
import shutil
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'samples/requested-templates/gallery'
def read(p):return json.loads((ROOT/p).read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def copy(src,dest):
    p=OUT/dest;p.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(ROOT/src,p);return dest

def main():
    cat=read('library/templates/requested-templates.json');rows=[]
    for row in sorted(cat['templates'],key=lambda x:(not x['source'].startswith('ai-'),x['source'])):
        r=dict(row);ident=r['template_id'];directory=Path(r['implementation_directory'])
        assert sha(ROOT/r['preview']['path'])==r['preview']['sha256']
        r['image']=copy(r['preview']['path'],f'templates/{ident}/source.png')
        for key,filename in [('values','source-values.json'),('contract','contract.json'),('metadata','implementation.json'),('customization','customization.json')]:
            r[key]=copy(directory/filename,f'templates/{ident}/{filename}')
        r['components']=[]
        for c in r['controls']['components']:
            c=dict(c)
            if 'contract' in c:c['contract']=copy(c['contract'],f'templates/{ident}/components/{Path(c["contract"]).name}')
            r['components'].append(c)
        r['command']=f'go run ./cmd/pptxtemplate build-review --id {ident} --values /absolute/path/source-values.json --out /absolute/path/new-slide'
        rows.append(r)
    decks=[]
    for source in ['ai','uhg']:
        for ext in ['pptx','pdf']:
            p=f'samples/visual-wave3/requested-{source}-source-v1.{ext}'
            dest=copy(p,f'decks/{source}-selected-layouts.{ext}')
            decks.append(dict(source=source,format=ext,path=dest,sha256=sha(ROOT/p)))
    gauge_review=ROOT/'planning/requested-templates/gauge-review.json'
    gauge_html=''
    if gauge_review.exists():
        g=json.loads(gauge_review.read_text())
        assert g['status']=='native_rendered_and_root_visually_reviewed'
        for a in g['artifacts']: assert sha(ROOT/a['path'])==a['sha256']
        img=copy('samples/requested-templates/render/gauge-v3/slide-001.png','gauge/source-gauge-controls.png')
        pptx=copy('samples/visual-wave3/requested-gauge-v3.pptx','gauge/source-gauge-controls.pptx')
        values=copy('samples/requested-templates/gauge/values.json','gauge/values.json')
        gauge_html=f'<article class="card"><h2>Original gauge · component customization</h2><p>Original T045 gauge structure, with selected cells and an independently controlled pointer. Defaults align pointer and highlight. This is a component example within the original template.</p><img src="{img}" alt="Five original semicircular gauges with independently selected highlights and pointers"><nav><a href="{pptx}">Gauge example PPTX</a><a href="{values}">Gauge settings JSON</a></nav><p>Top to bottom: pink cell 1 and pointer 1; blue cells 2–3 and pointer 5; gray cell 3 and pointer 3; navy cell 4 and pointer 2; blue cell 5 and pointer 5. Two top scale bands remain original.</p></article>'
    payload=dict(schema='pptxgengo.requested-template-gallery.v1',templates=rows,decks=decks,total_registered_templates=cat['total_registered_templates'],scope=cat['scope'])
    (OUT/'index.json').write_text(json.dumps(payload,indent=2)+'\n')
    nav='<a href="#gauge">Original gauge controls</a> '+ ' '.join(f'<a href="{d["path"]}">{"Lab" if d["source"]=="ai" else "UHG"} layouts · {d["format"].upper()}</a>' for d in decks)
    page='''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Selected West Monroe layouts</title>
<style>*{box-sizing:border-box}body{margin:0;background:#f0f2f7;color:#070154;font:16px/1.5 system-ui,sans-serif}header{background:#070154;color:white;padding:32px 5vw}h1{margin:0}header p{max-width:1100px}a{color:#0047ff}header a{color:#fff}nav{display:flex;gap:20px;flex-wrap:wrap}.toolbar{position:sticky;top:0;background:white;display:flex;gap:16px;flex-wrap:wrap;padding:16px 5vw;z-index:5;border-bottom:1px solid #ced7e6}input,select{padding:10px;font:inherit;border:1px solid #aeb7cf}input{flex:1;min-width:250px}main{max-width:1440px;margin:auto;padding:24px}.card{background:white;border:1px solid #ced7e6;margin:0 0 28px;padding:24px}h2{margin:0;font-size:23px}.meta{color:#53668d;font-size:14px}.card img{width:100%;margin:16px 0;border:1px solid #e6eaf2}.tags{display:flex;gap:8px;flex-wrap:wrap}.tag{background:#e8edf5;padding:4px 10px;font-size:13px}summary{cursor:pointer;font-weight:600;margin-top:18px}table{border-collapse:collapse;width:100%;margin:16px 0;font-size:14px}td,th{border:1px solid #ced7e6;padding:8px;text-align:left;vertical-align:top}code{white-space:pre-wrap;overflow-wrap:anywhere;font-size:13px}.note{padding:16px;background:#e8edf5}button{font:inherit;padding:8px;cursor:pointer}@media(max-width:700px){main{padding:10px}.card{padding:12px}table{display:block;overflow:auto}}</style>
<header><h1>Your selected layout templates</h1><p>44 exact source layouts: 11 from the AI Product and Engineering Executive Accelerator Lab and 33 from UHG. Their original designs and content are preserved. The repository now registers 101 source templates, including the 8 selections already in the library.</p><p>Choose a layout, download its content fields, and adapt the existing components. Counts and geometry stay fixed unless the documented component command supports a specific change.</p><nav>DECK_LINKS</nav></header>
<div class="toolbar"><input id="q" aria-label="Search layouts" placeholder="Search source slide, template name, category or component…"><select id="source" aria-label="Source"><option value="">Both sources</option><option value="ai-accelerator">Accelerator Lab</option><option value="uhg">UHG</option></select><select id="category" aria-label="Category"><option value="">All categories</option></select><span id="count"></span></div><main><p class="note">These are original source layouts, not rewritten proposals. Names, photos, metrics, quotes and other source claims still require review when reused. Original typography is retained. New copy requires text-fit and visual review. The installed stable release is unchanged; commands below run from the repository checkout.</p><section id="cards"></section><section id="gauge">GAUGE_HTML</section></main>
<script>const data=PAYLOAD;const esc=s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));for(const v of [...new Set(data.map(x=>x.category))].sort()){const o=document.createElement('option');o.value=v;o.textContent=v.replaceAll('_',' ');document.querySelector('#category').append(o)}
function render(){const q=document.querySelector('#q').value.toLowerCase(),s=document.querySelector('#source').value,c=document.querySelector('#category').value;const rows=data.filter(r=>(!s||r.source.startsWith(s+':'))&&(!c||r.category===c)&&JSON.stringify([r.name,r.source,r.template_id,r.category,r.components.map(c=>c.name)]).toLowerCase().includes(q));document.querySelector('#count').textContent=rows.length+' / '+data.length;document.querySelector('#cards').innerHTML=rows.map(r=>`<article class="card" id="${esc(r.template_id)}"><h2>${esc(r.name)}</h2><p class="meta">${esc(r.source)} · ${esc(r.template_id)} · ${r.existing_contract?'Existing exact template':'New exact template'}</p><div class="tags"><span class="tag">${esc(r.category.replaceAll('_',' '))}</span><span class="tag">${r.controls.text_slots} text slots</span><span class="tag">${r.controls.text_bindings} editable text runs</span><span class="tag">Original layout preserved</span></div><a href="${esc(r.image)}"><img loading="lazy" src="${esc(r.image)}" alt="Original layout ${esc(r.source)}"></a><nav><a href="${esc(r.values)}" download>Original content JSON</a><a href="${esc(r.contract)}">Template contract</a><a href="${esc(r.customization)}">Customization map</a></nav><details><summary>Components and styling options</summary><p>Template profiles: ${r.controls.profiles.length?esc(r.controls.profiles.join(', ')):'Original source styling'}. Profiles apply only to the contract’s named bindings. Fonts, spacing and component counts retain the original values.</p><table><thead><tr><th>Component group</th><th>Text fields</th><th>Supported color profiles</th><th>Source component contract</th></tr></thead><tbody>${r.components.map(c=>`<tr><td>${esc(c.name)}</td><td>${esc(c.slots.join(', ')||'Retained artwork')}</td><td>${esc(c.profiles.join(', ')||'Source styling')}</td><td>${c.contract?`<a href="${esc(c.contract)}">Inspect controls</a>`:'Retained source resource'}</td></tr>`).join('')}</tbody></table><p>These component groups are tied to their original slide. Their mapped fields are executable controls; they are not claims that the entire group can automatically reflow or move between unrelated decks.</p></details><details><summary>Use this layout</summary><pre><code>${esc(r.command)}</code></pre><p>For an unchanged source reference: <code>go run ./cmd/pptxtemplate build-review --id ${esc(r.template_id)} --source-values --out /absolute/path/new-reference</code></p></details>${r.limitations.length?`<details><summary>Source-specific caveats</summary><ul>${r.limitations.map(x=>`<li>${esc(x)}</li>`).join('')}</ul></details>`:''}</article>`).join('')};for(const id of ['q','source','category'])document.querySelector('#'+id).addEventListener('input',render);render();</script></html>'''
    (OUT/'index.html').write_text(page.replace('DECK_LINKS',nav).replace('GAUGE_HTML',gauge_html).replace('PAYLOAD',json.dumps(rows).replace('<','\\u003c')))
    print(OUT/'index.html')
if __name__=='__main__':main()
