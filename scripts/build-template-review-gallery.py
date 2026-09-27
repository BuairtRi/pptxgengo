#!/usr/bin/env python3
"""Build a local source/adaptation gallery with explicit, evidence-bound QA states."""
import argparse
import hashlib
import json
import os
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--native', type=Path, action='append', required=True,
                        help='native-render.json; later entries supersede earlier ones')
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    out = args.out.resolve()
    out.mkdir(parents=True, exist_ok=True)
    rows = json.loads((root / 'library/templates/rollout/assignments.json').read_text())['entries']
    evidence = {}
    reviews = {}
    for p in (root / 'library/templates/rollout/reviews').glob('*.json'):
        r = json.loads(p.read_text())
        for t in r['templates']:
            reviews[(str((root / r['native_render']).resolve()), t['template_id'])] = t
    for path in args.native:
        path = path.resolve()
        manifest = json.loads(path.read_text())
        templates = {t['template_id']: t for t in manifest['templates']}
        for deck in manifest['decks']:
            if 'render_directory' not in deck:
                continue
            if sha(root / deck['native_pptx']) != deck['sha256'] or sha(root / deck['native_pdf']) != deck['native_pdf_sha256']:
                raise ValueError(f"Changed native artifact: {deck['path']}")
            for page in deck['pages']:
                png = root / deck['render_directory'] / f"slide-{page['pdf_page']:03d}.png"
                actual_hash = sha(png)
                if page.get('png_sha256', actual_hash) != actual_hash:
                    raise ValueError(f'Changed render: {png}')
                review = reviews.get((str(path), page['template_id']), {})
                if review and review.get('png_sha256') != actual_hash:
                    raise ValueError(f'Stale review: {png}')
                evidence[page['template_id']] = dict(
                    png=str(png), pdf=deck['native_pdf'], pptx=deck['native_pptx'],
                    pdf_page=page['pdf_page'], manifest=str(path),
                    snapshot=templates[page['template_id']], review=review, render_limitations=manifest.get('render_limitations', []))
    def link(p):
        return os.path.relpath(root / p, out)
    data = []
    for row in rows:
        directory = root / row['implementation_directory']
        m = json.loads((directory / 'implementation.json').read_text())
        if sha(root / m['preview']['path']) != m['preview']['sha256']:
            raise ValueError(f"Changed source preview: {row['template_id']}")
        e = evidence.get(row['template_id'])
        item = dict(id=row['template_id'], name=row['name'], category=row['category'], lane=row['lane'],
                    source=row['representative'], altitude=row['altitudes'], density=row['densities'],
                    source_png=link(m['preview']['path']), status='pending_native_render',
                    limitations=m['limitations'] + m.get('opaque_areas', []) + m.get('retained_source_content', []), findings=[], resolved_findings=[], current_inputs_match=None,
                    contract=link(directory / 'contract.json'), values=link(directory / 'example-values.json'))
        if e:
            snapshot = e['snapshot']
            native_dir = Path(e['manifest']).parent
            matches = []
            for inp in snapshot['inputs']:
                name = Path(inp['path']).name
                if name in ('contract.json', 'values.json'):
                    live = directory / ('example-values.json' if name == 'values.json' else name)
                    matches.append(sha(live) == inp['sha256'])
                frozen = native_dir / inp['path']
                if sha(frozen) != inp['sha256']:
                    raise ValueError(f'Changed input snapshot: {frozen}')
            review = e['review']
            item.update(adapted_png=link(e['png']), pptx=link(e['pptx']), pdf=link(e['pdf']) + '#page=' + str(e['pdf_page']),
                        manifest=link(e['manifest']), status=review.get('status', 'pending_visual_review'),
                        findings=review.get('findings', []), resolved_findings=review.get('resolved_findings', []),
                        limitations=snapshot['limitations'] + snapshot.get('opaque_areas', []) + snapshot.get('retained_source_content', []) + e.get('render_limitations', []), current_inputs_match=len(matches) == 2 and all(matches),
                        input_directory=link(native_dir / 'inputs' / row['template_id']))
        data.append(item)
    (out / 'index.json').write_text(json.dumps({'schema':'pptxgengo.template-gallery.v1', 'adaptation_qualified':False, 'templates':data}, indent=2) + '\n')
    payload = json.dumps(data).replace('<', '\\u003c')
    page = '''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Template adaptation review</title><style>
*{box-sizing:border-box}body{margin:0;background:#f3f5fa;color:#101044;font:16px system-ui,sans-serif}header{padding:26px 32px;background:#0b075b;color:white}h1{margin:0 0 8px;font-size:28px}header p{max-width:1100px;line-height:1.5;margin:8px 0}.toolbar{position:sticky;top:0;z-index:2;padding:15px 32px;display:flex;gap:12px;flex-wrap:wrap;background:white;border-bottom:1px solid #cdd5e4}input,select{font:inherit;padding:8px;border:1px solid #aab3c9;border-radius:4px}input{min-width:290px}.count{align-self:center;margin-left:auto}main{padding:24px 32px}.card{background:white;padding:22px;margin:0 0 24px;border:1px solid #d3daea}.top{display:flex;justify-content:space-between;gap:15px}.card h2{font-size:21px;margin:0 0 8px}.meta{color:#546581;font-size:14px}.badge{padding:7px 10px;background:#e8edf7;white-space:nowrap;font-size:14px;height:fit-content}.needs_revision{background:#ffe9e9;color:#901313}.example_visually_reviewed{background:#e1f3ed;color:#135a42}.pair{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin:16px 0}.pair img{width:100%;border:1px solid #d3daea}.pair a{display:block}figcaption{font-size:14px;margin-bottom:8px;font-weight:600}figure{margin:0}.links{display:flex;gap:18px;flex-wrap:wrap;margin-top:12px}a{color:#164ee3}li{margin:8px 0}details{margin-top:14px;line-height:1.45}.stale{color:#8b4705;font-size:14px} @media(max-width:900px){.pair{grid-template-columns:1fr}.top{display:block}.badge{display:inline-block;margin:8px 0}main{padding:16px}.toolbar{padding:12px}}
</style><header><h1>Template adaptation review</h1><p>65 shortlisted designs · source reference beside native PowerPoint output</p><p>These examples preserve source geometry and assets. A reviewed example demonstrates that specific content; it does not establish arbitrary text capacity, dynamic reflow, or a client-ready deck. Source-specific artwork and facts may remain.</p></header>
<section class="toolbar"><input id="search" aria-label="Search templates" placeholder="Search category, template, source…"><select id="lane" aria-label="Workstream"><option value="">All workstreams</option></select><select id="status" aria-label="Review state"><option value="">All review states</option></select><span id="count" class="count"></span></section><main id="cards"></main>
<script>const data=DATA;const labels={pending_native_render:'Awaiting native render',pending_visual_review:'Awaiting visual review',needs_revision:'Needs revision',example_visually_reviewed:'Example visually reviewed'};
const esc=s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
for(const key of ['lane','status'])for(const v of [...new Set(data.map(d=>d[key]))].sort()){const o=document.createElement('option');o.value=v;o.textContent=labels[v]||v.replaceAll('_',' ');document.getElementById(key).append(o)}
function draw(){const q=document.getElementById('search').value.toLowerCase(),lane=document.getElementById('lane').value,state=document.getElementById('status').value;const rows=data.filter(d=>(!lane||d.lane===lane)&&(!state||d.status===state)&&JSON.stringify([d.id,d.name,d.source,d.category,d.altitude,d.density]).toLowerCase().includes(q));document.getElementById('count').textContent=rows.length+' / '+data.length+' templates';document.getElementById('cards').innerHTML=rows.map(d=>`<article class="card" id="${esc(d.id)}"><div class="top"><div><h2>${esc(d.name)}</h2><div class="meta">${esc(d.id)} · ${esc(d.category)} · ${esc(d.altitude.join(', '))} · ${esc(d.density.join(', '))}</div></div><span class="badge ${esc(d.status)}">${esc(labels[d.status])}</span></div>${d.current_inputs_match===false?'<p class="stale">Authoring inputs have changed since this render. This image and review apply to the frozen snapshot linked below.</p>':''}<div class="pair"><figure><figcaption>Source · ${esc(d.source)}</figcaption><a href="${esc(d.source_png)}"><img loading="lazy" src="${esc(d.source_png)}" alt="Source ${esc(d.source)}"></a></figure><figure><figcaption>Adapted example · native PowerPoint</figcaption>${d.adapted_png?`<a href="${esc(d.adapted_png)}"><img loading="lazy" src="${esc(d.adapted_png)}" alt="Adapted ${esc(d.id)}"></a>`:'<p>Native render pending.</p>'}</figure></div>${d.findings.length?'<strong>Review findings</strong><ul>'+d.findings.map(f=>'<li>'+esc(f)+'</li>').join('')+'</ul>':''}${d.resolved_findings.length?'<details><summary>Corrections confirmed in this render</summary><ul>'+d.resolved_findings.map(f=>'<li>'+esc(f.resolution)+'</li>').join('')+'</ul></details>':''}<details><summary>Supported boundary and retained content</summary><ul>${d.limitations.map(s=>'<li>'+esc(s)+'</li>').join('')}</ul></details><nav class="links">${d.pptx?`<a href="${esc(d.pptx)}">PowerPoint</a><a href="${esc(d.pdf)}">PDF page</a><a href="${esc(d.manifest)}">Render evidence</a><a href="${esc(d.input_directory)}/values.json">Frozen values</a>`:''}<a href="${esc(d.contract)}">Current contract</a><a href="${esc(d.values)}">Current values</a></nav></article>`).join('')};for(const id of ['search','lane','status'])document.getElementById(id).addEventListener('input',draw);draw();</script></html>'''
    (out / 'index.html').write_text(page.replace('DATA', payload, 1))
    print(out / 'index.html')


if __name__ == '__main__':
    main()
