#!/usr/bin/env python3
"""Generate a local side-by-side layout review gallery from explicit preview paths."""
import argparse
import html
import hashlib
import json
import os
from pathlib import Path

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--dedup',type=Path,required=True)
    p.add_argument('--previews',type=Path,required=True)
    p.add_argument('--decisions',type=Path,required=True)
    p.add_argument('--out',type=Path,required=True)
    a=p.parse_args()
    if a.out.exists():p.error('Output exists')
    dedup=json.loads(a.dedup.read_text());previews=json.loads(a.previews.read_text())['previews']
    decisions=json.loads(a.decisions.read_text());by_member={}
    pair_decisions={(p['left'],p['right']):p for p in decisions.get('pair_decisions',[])}
    for group in decisions['groups']:
        for member in group['members']:by_member[member]=group
    def esc(v):return html.escape(str(v),quote=True)
    def img(ident):
        meta=previews.get(ident)
        if not meta:return '<p class="missing">Preview missing: '+esc(ident)+'</p>'
        path=Path(meta['path']).resolve()
        if not path.is_file():raise FileNotFoundError(path)
        if hashlib.sha256(path.read_bytes()).hexdigest()!=meta['sha256']:
            raise ValueError('Preview hash mismatch: '+str(path))
        url=Path(os.path.relpath(path,a.out.resolve().parent)).as_posix()
        return '<a href="'+esc(url)+'"><img loading="lazy" src="'+esc(url)+'" alt="'+esc(ident)+'"></a>'
    cards=[]
    for pair in dedup['review_queue']:
        left,right=pair['left'],pair['right'];g=by_member.get(left)
        matched=g and right in g['members']
        decision=pair_decisions.get((left,right))
        status=('Reviewed: shared family; variants retained' if matched else
                'Reviewed: keep separate' if decision and decision['decision']=='keep_separate' else
                'Review pending')
        reason=g['rationale'] if matched else decision['reason'] if decision else pair['reason']
        cards.append('<article><h2>'+esc(left)+' ↔ '+esc(right)+'</h2><p class="status">'+esc(status)+'</p><p>'+esc(reason)+'</p><div class="pair"><figure>'+img(left)+'<figcaption>'+esc(left)+'</figcaption></figure><figure>'+img(right)+'<figcaption>'+esc(right)+'</figcaption></figure></div></article>')
    doc='''<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Layout deduplication review</title><style>
body{font:16px system-ui;margin:0;color:#10113b;background:#eef1f7}header{padding:28px 4vw;background:#070154;color:white}main{padding:24px 3vw}article{background:white;border:1px solid #ccd3e1;border-radius:8px;margin-bottom:24px;padding:20px}h1{margin:0 0 10px}h2{font-size:18px}.status{font-weight:650;color:#0047ff}.pair{display:grid;grid-template-columns:1fr 1fr;gap:18px}figure{margin:0}img{width:100%;border:1px solid #ccd3e1}figcaption{font-size:13px;margin-top:8px}.missing{color:#97321e}@media(max-width:800px){.pair{grid-template-columns:1fr}}
</style><header><h1>Layout deduplication review</h1><p>Source occurrences remain intact. Shared families consolidate classification work; variants and content retain their own fit checks.</p><p>Native PowerPoint previews · 1920×1080 · Click a preview for full size</p></header><main>'''+''.join(cards)+'</main></html>'
    a.out.parent.mkdir(parents=True,exist_ok=True)
    with a.out.open('x') as f:f.write(doc)
    print(json.dumps({'pairs':len(cards),'out':str(a.out)}))
if __name__=='__main__':main()
