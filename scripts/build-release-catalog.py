#!/usr/bin/env python3
"""Freeze the reviewed template images and authoring inputs into a portable HTML catalog."""
import hashlib
import json
import shutil
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / 'samples/template-expansion/rollout/gallery-v3'
OUT = ROOT / 'release/catalog'

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    OUT.mkdir(parents=True, exist_ok=True)
    rows = json.loads((SOURCE / 'index.json').read_text())['templates']
    checkpoint = json.loads((ROOT / 'library/templates/rollout/checkpoint.json').read_text())
    evidence = {x['template_id']: x for x in checkpoint['templates']}
    exported = []
    for row in rows:
        item = dict(row)
        target = OUT / 'templates' / row['id']
        target.mkdir(parents=True, exist_ok=True)
        contract = json.loads((SOURCE / row['contract']).read_text())
        e = evidence[row['id']]
        for key, filename in [('source_png','source.png'), ('adapted_png','adapted.png'), ('contract','contract.json'), ('values','values.json')]:
            src = (SOURCE / row[key]).resolve()
            expected = e.get({'adapted_png':'png_sha256', 'contract':'contract_sha256', 'values':'example_values_sha256'}.get(key,''))
            if expected and digest(src) != expected:
                raise ValueError(f'Changed reviewed input: {src}')
            shutil.copy2(src, target / filename)
            item[key] = str((target / filename).relative_to(OUT))
        for key in ['pptx','pdf','manifest','input_directory']:
            item.pop(key, None)
        item['slots'] = len(contract.get('slots',{}))
        item['profiles'] = list(contract.get('profiles',{}))
        item['zones'] = list(contract.get('zones',{}))
        item['native_review'] = 'Picture geometry replay; generated PPTX reopening pending' if any('Opening/repair validation' in s for s in item['limitations']) else 'Direct native PowerPoint open/export'
        item['command'] = f"pptxgengo template build-review --id {row['id']} --values /absolute/path/values.json --out /absolute/path/new-review"
        item['artifact_sha256'] = {p.name:digest(p) for p in target.iterdir() if p.is_file()}
        exported.append(item)
    payload = {'schema':'pptxgengo.release-catalog.v1','version':'0.1.0-local.3','templates':exported,'qualification':{'reviewed_examples':65,'direct_native_exports':60,'geometry_replays':5,'arbitrary_content_qualified':0}}
    (OUT/'index.json').write_text(json.dumps(payload,indent=2)+'\n')
    page = '''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>West Monroe presentation library</title>
<style>*{box-sizing:border-box}body{margin:0;background:#f3f4f8;color:#171449;font:16px/1.5 system-ui,sans-serif}header{background:#10004b;color:white;padding:32px max(24px,5vw)}h1{margin:0;font-size:32px}header p{max-width:1040px}.stats{display:flex;gap:12px;flex-wrap:wrap}.stats span{border:1px solid #807299;padding:7px 12px}.intro,main{max-width:1500px;margin:auto;padding:24px}.intro{background:white;margin-top:24px}code{font:13px/1.6 ui-monospace,monospace;overflow-wrap:anywhere}.toolbar{position:sticky;top:0;background:#fff;z-index:2;display:flex;flex-wrap:wrap;gap:12px;padding:14px 24px;border-bottom:1px solid #ddd}input,select,button{padding:9px;font:inherit;border:1px solid #bdc0ce;border-radius:4px}input{flex:1;min-width:240px}.count{align-self:center}article{background:white;border:1px solid #ddd;margin-bottom:24px;padding:22px}h2{font-size:21px;margin:0 0 8px}.meta{font-size:14px;color:#595674}.pair{display:grid;grid-template-columns:1fr 1fr;gap:16px;margin:18px 0}figure{margin:0}img{width:100%;border:1px solid #ddd}figcaption{font-size:13px;font-weight:600;margin-bottom:8px}a{color:#4b26a5}details{margin:12px 0}summary{cursor:pointer;font-weight:600}li{margin:6px 0}.tag{display:inline-block;background:#ece8fa;font-size:13px;padding:3px 8px;margin:4px 4px 0 0}.command{background:#f5f4fa;padding:12px}nav{display:flex;flex-wrap:wrap;gap:16px}.replay{color:#895106}@media(max-width:800px){.pair{grid-template-columns:1fr}header{padding:24px}main{padding:12px}}</style>
<header><h1>West Monroe presentation library</h1><p>Local release 0.1.0-local.3 · Browse source designs, compare reviewed adaptations, and choose a starting point for new content.</p><div class="stats"><span>65 source-based templates</span><span>60 direct native exports</span><span>5 accent geometry replays</span></div><p>These are fixed-geometry editing contracts. Reviewed examples demonstrate specific content; new copy still needs fit and visual review. Five accent examples await reopening of their generated PPTX files. Arbitrary content capacity has not been qualified.</p></header>
<section class="intro"><h2>Start a slide</h2><ol><li>Find a layout matching the slide’s role, density and number of content zones.</li><li>Download its values JSON. Replace content while preserving each slot’s binding structure.</li><li>Run the build command below. Inspect retained logos, labels and artwork; review text fit before client use.</li></ol><p><strong>Need a different number of boxes, rows or phases?</strong> Use the separate component composition workflow where supported. This release does not automatically restructure all imported templates. Ask Codex to use the <strong>west-monroe-presentations</strong> skill and its component-composition reference.</p><p>Browse from the shell with <code>pptxgengo catalog --open</code>. Inspect slots with <code>pptxgengo template inspect --id TEMPLATE_ID</code>.</p></section>
<div class="toolbar"><input id="q" aria-label="Search" placeholder="Search title, ID, source, category…"><select id="category" aria-label="Category"><option value="">All categories</option></select><select id="density" aria-label="Density"><option value="">All densities</option></select><span id="count" class="count"></span></div><main id="cards"></main>
<script>const data=PAYLOAD;const esc=s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));for(const k of ['category','density'])for(const v of [...new Set(data.flatMap(x=>Array.isArray(x[k])?x[k]:[x[k]]))].sort()){const o=document.createElement('option');o.value=v;o.textContent=v.replaceAll('_',' ');document.getElementById(k).append(o)}
function draw(){const q=document.getElementById('q').value.toLowerCase(),c=document.getElementById('category').value,d=document.getElementById('density').value;const rows=data.filter(x=>(!c||x.category===c)&&(!d||x.density.includes(d))&&JSON.stringify([x.id,x.name,x.source,x.category,x.altitude,x.density]).toLowerCase().includes(q));document.getElementById('count').textContent=rows.length+' / 65';document.getElementById('cards').innerHTML=rows.map(x=>`<article id="${esc(x.id)}"><h2>${esc(x.name)}</h2><div class="meta">${esc(x.id)} · ${esc(x.source)} · ${esc(x.category.replaceAll('_',' '))}</div><div>${[...x.altitude,...x.density,x.slots+' content slots','Fixed source geometry',...(x.profiles.length?['Style profiles: '+x.profiles.join(', ')]:[])].map(s=>'<span class="tag">'+esc(s)+'</span>').join('')}</div><div class="pair"><figure><figcaption>Source reference</figcaption><a href="${esc(x.source_png)}"><img loading="lazy" src="${esc(x.source_png)}" alt="Source for ${esc(x.id)}"></a></figure><figure><figcaption>Reviewed content adaptation</figcaption><a href="${esc(x.adapted_png)}"><img loading="lazy" src="${esc(x.adapted_png)}" alt="Reviewed adaptation for ${esc(x.id)}"></a></figure></div><p class="meta ${x.native_review.startsWith('Picture')?'replay':''}">${esc(x.native_review)}</p><details><summary>Customization scope and caveats</summary><p>Named text slots are available; style changes require explicit contract roles/profiles. Geometry, item counts and connectors remain fixed unless a separate component workflow implements the requested change.</p><ul>${x.limitations.map(s=>'<li>'+esc(s)+'</li>').join('')}</ul></details><details><summary>Build this template</summary><p>Save and edit the values JSON, then replace the paths in this command:</p><p class="command"><code>${esc(x.command)}</code></p><nav><a href="${esc(x.values)}" download>Download values JSON</a><a href="${esc(x.contract)}">Inspect contract</a></nav></details></article>`).join('')};for(const id of ['q','category','density'])document.getElementById(id).addEventListener('input',draw);draw();</script></html>'''
    (OUT/'index.html').write_text(page.replace('PAYLOAD',json.dumps(exported).replace('<','\\u003c')))
    print(f'{len(exported)} templates frozen at {OUT}')

if __name__ == '__main__':
    main()
