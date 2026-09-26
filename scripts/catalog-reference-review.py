#!/usr/bin/env python3
"""Find reviewed component references or build a visual selection gallery."""
import argparse
import base64
import hashlib
import html
import json
import os
from pathlib import Path
import sys
import tempfile


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def esc(value):
    return html.escape(str(value), quote=True)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('command', choices=['find', 'gallery'])
    ap.add_argument('--shortlist', type=Path, default=Path('library/reference-shortlist.json'))
    ap.add_argument('--family', help='metric-panel, numbered-card, or team-pod')
    ap.add_argument('--query', default='', help='all words must match reviewed descriptions')
    ap.add_argument('--components', type=Path, default=Path('samples/component-expansion/components-v2.jsonl'))
    ap.add_argument('--geometry', type=Path, default=Path('samples/catalog-next/geometry.jsonl'))
    ap.add_argument('--previews', type=Path, default=Path('library/render-manifest.json'))
    ap.add_argument('--sources', type=Path, default=Path('planning/source-registry.json'))
    ap.add_argument('--out', type=Path)
    args = ap.parse_args()
    shortlist = json.loads(args.shortlist.read_text())
    if shortlist['schema'] != 'pptxgengo.reference-shortlist.v1':
        raise ValueError('unsupported shortlist schema')
    refs = shortlist['references']
    if len({r['ref'] for r in refs}) != len(refs) or len({r['component_id'] for r in refs}) != len(refs):
        raise ValueError('duplicate reference or component')
    if args.family:
        family = args.family.removeprefix('component-family:')
        refs = [r for r in refs if r['family_id'] == 'component-family:' + family]
    terms = args.query.lower().split()
    refs = [r for r in refs if all(t in json.dumps(r).lower() for t in terms)]
    if args.command == 'find':
        print(json.dumps({'shortlist_sha256':digest(args.shortlist), 'references': refs}, indent=2))
        return
    if not args.out or args.out.exists():
        raise ValueError('--out must name a new HTML file')
    if digest(args.components) != shortlist['components_sha256']:
        raise ValueError('component input differs from reviewed shortlist')
    if digest(args.geometry) != shortlist['geometry_sha256']:
        raise ValueError('geometry input differs from reviewed shortlist')
    components = {r['id']:r for r in map(json.loads,args.components.read_text().splitlines())}
    geometry = {r['occurrence_id']:r for r in map(json.loads,args.geometry.read_text().splitlines())}
    manifest = json.loads(args.previews.read_text())
    sources = {s['source_id']:s for s in json.loads(args.sources.read_text())['sources']}
    images = {}
    verified_sources = set()
    cards = []
    for r in refs:
        c = components[r['component_id']]
        for key in ('source_id','source_sha256','slide_number'):
            if c[key] != r[key]:
                raise ValueError(f'{r["ref"]}: source mismatch')
        if c['canonical_family_id'] != r['family_id']:
            raise ValueError('family mismatch')
        source = sources[c['source_id']]
        if source['source_sha256'] != c['source_sha256']:
            raise ValueError('registry hash mismatch')
        if c['source_id'] not in verified_sources:
            if digest(Path(source['path_hint'])) != c['source_sha256']:
                raise ValueError(f'linked source deck hash mismatch: {source["path_hint"]}')
            verified_sources.add(c['source_id'])
        key = f'{c["source_id"]}:{c["slide_number"]:03}'
        p = manifest['previews'][key]
        if p['source_slide_number'] != c['slide_number']:
            raise ValueError('preview slide mismatch')
        source_manifest = next(s for s in manifest['sources'] if s['source_id']==c['source_id'])
        if source_manifest['source_sha256'] != c['source_sha256']:
            raise ValueError('preview source hash mismatch')
        png = Path(p['path'])
        if key not in images:
            if digest(png) != p['sha256']:
                raise ValueError(f'preview hash mismatch: {png}')
            images[key] = base64.b64encode(png.read_bytes()).decode()
        member_rows = [geometry[oid] for oid in c['member_occurrence_ids']]
        for g in member_rows:
            if any(g[k]!=c[k] for k in ('source_id','source_sha256','slide_number','source_part')) or g['geometry_status']!='resolved':
                raise ValueError('unresolved or mismatched component geometry')
        sizes = {(g['slide_size_emu']['width'],g['slide_size_emu']['height']) for g in member_rows}
        if len(sizes)!=1: raise ValueError('inconsistent slide sizes')
        sw,sh = sizes.pop()
        bounds = c['bounds_emu']; x,y,w,h = [bounds[k] for k in ('x','y','width','height')]
        if w<=0 or h<=0: raise ValueError('nonpositive component bounds')
        padding = min(w,h)*0.04
        viewport = f'{x-padding} {y-padding} {w+2*padding} {h+2*padding}'
        source_href = os.path.relpath(Path(source['path_hint']).resolve(),args.out.parent.resolve())
        frame = f'<rect x="{x}" y="{y}" width="{w}" height="{h}" fill="none" stroke="#F900D3" stroke-width="22000"/>'
        slots = ''.join(f'<tr><td>{esc(s["name"])}</td><td>{esc(s.get("observed_text",""))}</td></tr>' for s in c['slots'])
        constraints = ''.join(f'<li>{esc(v)}</li>' for v in c.get('adaptation_constraints',[]))
        badge = '<span class="badge">First implementation fixture</span>' if r['implementation_start'] else ''
        cards.append(f'''<article id="{esc(r['ref'])}" data-ref="{esc(r['ref'])}" data-id="{esc(c['id'])}" data-family="{esc(r['family_id'])}" data-search="{esc(json.dumps(r).lower())}">
<header><h2>{esc(r['ref'])} · {esc(r['title'])}</h2>{badge}<p>{esc(c['source_id'])} · source slide {c['slide_number']}</p></header>
<div class="crop"><svg viewBox="{viewport}" role="img" aria-label="Enlarged source region for {esc(r['title'])}"><use href="#image-{esc(key)}"/></svg></div>
<div class="content"><p><strong>Use for:</strong> {esc(r['use_when'])}</p><p>{esc(r['visual_structure'])}</p><p><strong>Content:</strong> {esc(r['content_density'])}. Capacity is not yet measured.</p><p class="recommendation">{esc(r['recommendation'])}</p>
<label>Preference <select class="preference"><option value="unreviewed">Unreviewed</option><option value="preferred">Prefer</option><option value="alternate">Keep as alternate</option><option value="avoid">Avoid</option></select></label>
<label class="note-label">Your note <textarea class="note" rows="2" placeholder="e.g. Use this for dense proposals"></textarea></label>
<details><summary>Full slide and component details</summary><svg class="context" viewBox="0 0 {sw} {sh}" role="img" aria-label="Full source slide; magenta frame marks component"><use href="#image-{esc(key)}"/>{frame}</svg>
<p>The enlarged region includes original pixels from surrounding slide objects. It is not an isolated or newly rendered component.</p><table><thead><tr><th>Catalog slot</th><th>Source content</th></tr></thead><tbody>{slots}</tbody></table><ul>{constraints}</ul><p><a href="{esc(source_href)}">Open source deck</a> (source position {c['slide_number']}; footer numbering may differ)</p><p class="identifier">{esc(c['id'])}</p></details></div></article>''')
    definitions = []
    for key,data in images.items():
        sid,slide = key.rsplit(':',1)
        c = next(components[r['component_id']] for r in refs if r['source_id']==sid and r['slide_number']==int(slide))
        g = geometry[c['member_occurrence_ids'][0]];sw,sh = g['slide_size_emu']['width'],g['slide_size_emu']['height']
        definitions.append(f'<image id="image-{esc(key)}" width="{sw}" height="{sh}" href="data:image/png;base64,{data}"/>')
    page = r'''<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Component reference shortlist</title><style>
*{box-sizing:border-box}body{margin:0;font:15px/1.5 system-ui,sans-serif;color:#070154;background:#f1f4f8}h1{margin:0;font-size:28px}h2{font-size:20px;margin:0}header p{margin:6px 0;color:#50658E}.intro{padding:24px;max-width:1450px;margin:auto}.filters{padding:14px 24px;background:#fff;display:flex;gap:18px;flex-wrap:wrap;align-items:center;border-block:1px solid #CED7E6}input,select,textarea,button{font:inherit;border:1px solid #98a7bd;border-radius:4px;padding:7px}button{cursor:pointer;background:#070154;color:white}main{max-width:1450px;margin:24px auto;padding:0 24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,410px),1fr));gap:24px;align-items:start}article{background:white;border:1px solid #CED7E6;border-radius:8px;overflow:hidden}article[hidden]{display:none}article:target{outline:3px solid #0047FF}article header,.content{padding:18px}.badge{font-size:12px;display:inline-block;padding:3px 7px;background:#E8EEF8;margin-top:8px}.crop{height:265px;background:#fafbfd;padding:16px;border-block:1px solid #e2e8ef}.crop svg{width:100%;height:100%}.context{width:100%;display:block;margin-top:12px}.content p{margin:0 0 12px}.recommendation{padding:10px;border-left:3px solid #0047FF;background:#f1f5ff}.note-label{display:block;margin-top:12px}textarea{display:block;width:100%;margin-top:4px}details{margin-top:18px}summary{cursor:pointer;font-weight:650}table{border-collapse:collapse;width:100%;font-size:12px}td,th{text-align:left;vertical-align:top;border-bottom:1px solid #ddd;padding:6px;white-space:pre-line}td{overflow-wrap:anywhere}ul{padding-left:18px}.identifier{font:11px monospace;overflow-wrap:anywhere}a{color:#0047FF}.intro p{max-width:1050px}#message{font-size:13px}
</style></head><body><svg width="0" height="0" aria-hidden="true" style="position:absolute"><defs>DEFINITIONS</defs></svg>
<div class="intro"><h1>Choose a reference by the job it needs to do</h1><p>Ten starting references across metric panels, numbered cards and delivery pods. Each enlargement uses the original PowerPoint render. Structural variants are listed separately; repeated content examples and color options stay within the same structure.</p><p>Use the codes M1–M3, N1–N5 and P1–P2 when giving feedback. Preferences describe your taste; they do not mark a component as technically approved. <strong>Export your choices before closing this page.</strong> No automatic saving or network requests.</p></div>
<div class="filters"><label>Family <select id="family"><option value="">All</option><option value="component-family:metric-panel">Metrics</option><option value="component-family:numbered-card">Numbered cards</option><option value="component-family:team-pod">Delivery pods</option></select></label><label>Find <input id="query" type="search" placeholder="evidence, short, three roles…"></label><label><input id="preferred" type="checkbox"> Preferred only</label><button id="export" type="button">Export choices</button><span id="count" aria-live="polite"></span><span id="message" role="status"></span></div><main>CARDS</main><script>
const cards=[...document.querySelectorAll('article')],family=document.getElementById('family'),query=document.getElementById('query'),preferred=document.getElementById('preferred');
function filter(){const terms=query.value.toLowerCase().split(/\s+/).filter(Boolean);let n=0;cards.forEach(c=>{c.hidden=!!((family.value&&c.dataset.family!==family.value)||terms.some(t=>!c.dataset.search.includes(t))||(preferred.checked&&c.querySelector('.preference').value!=='preferred'));if(!c.hidden)n++;});document.getElementById('count').textContent=n+' references shown';}
[family,query,preferred,...document.querySelectorAll('.preference')].forEach(x=>x.addEventListener('input',filter));filter();
document.getElementById('export').addEventListener('click',()=>{const output={schema:'pptxgengo.reference-preferences.v1',shortlist_sha256:'SHORTLIST_HASH',technical_approval:false,choices:cards.map(c=>({ref:c.dataset.ref,component_id:c.dataset.id,preference:c.querySelector('.preference').value,note:c.querySelector('.note').value}))};const blob=new Blob([JSON.stringify(output,null,2)+'\n'],{type:'application/json'});const url=URL.createObjectURL(blob),a=document.createElement('a');a.href=url;a.download='component-reference-preferences.json';document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(url),1000);document.getElementById('message').textContent='Choices exported as JSON.';});
</script></body></html>'''
    # Replace tokens before embedding any source-provided strings.
    page = page.replace('SHORTLIST_HASH',digest(args.shortlist)).replace('DEFINITIONS',''.join(definitions)).replace('CARDS',''.join(cards))
    args.out.parent.mkdir(parents=True,exist_ok=True)
    fd,temp = tempfile.mkstemp(dir=args.out.parent,prefix='.references-',suffix='.tmp')
    try:
        with os.fdopen(fd,'w') as stream: stream.write(page)
        os.link(temp,args.out)
    finally:
        os.unlink(temp)
    print(json.dumps({'references':len(refs),'verified_previews':len(images),'output':str(args.out)}))


if __name__=='__main__':
    try: main()
    except (OSError,ValueError,KeyError) as error:
        print(f'error: {error}',file=sys.stderr)
        sys.exit(2)
