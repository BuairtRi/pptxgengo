#!/usr/bin/env python3
"""Build a portable, evidence-aware gallery for adaptive deck review records."""
from __future__ import annotations
import argparse, hashlib, html, json, re, shutil
from collections import defaultdict
from pathlib import Path

STATUS={'pending_native','native_verified','visually_reviewed','needs_revision','blocked'}

def read_json(path): return json.loads(Path(path).read_text(encoding='utf-8'))
def safe(s):
    x=re.sub(r'[^A-Za-z0-9._-]+','-',str(s or 'item')).strip('-')[:80]
    return x or 'item'
def resolve(raw,root):
    if not raw:return None
    p=Path(raw)
    if not p.is_absolute():p=root/p
    return p.resolve()
def copy_input(raw,root,out,relbase,cache):
    p=resolve(raw,root)
    if not p or not p.is_file(): return None
    digest=hashlib.sha256(p.read_bytes()).hexdigest()
    key=(str(p),digest)
    if key in cache:return cache[key]
    dest=out/relbase/f'{safe(p.stem)}-{digest[:10]}{p.suffix.lower()}'
    dest.parent.mkdir(parents=True,exist_ok=True)
    shutil.copy2(p,dest)
    cache[key]=dest.relative_to(out).as_posix()
    return cache[key]
def file_digest(raw,root):
    p=resolve(raw,root)
    if not p or not p.is_file(): return None
    return hashlib.sha256(p.read_bytes()).hexdigest()
def flat(v,prefix=''):
    if isinstance(v,dict):
        out={}
        for k,x in v.items():out.update(flat(x,f'{prefix}.{k}' if prefix else str(k)))
        return out
    if isinstance(v,list):
        out={}
        for i,x in enumerate(v):out.update(flat(x,f'{prefix}[{i}]'))
        if not v:out[prefix]='[]'
        return out
    return {prefix:v}
def compact_cap(item,root=None,out=None,cache=None):
    fixed=item.get('fixed_source',{})
    slots=fixed.get('source_bound_slots',[])
    preview=(item.get('source_identity',{}).get('preview') or {})
    local_preview=copy_input(preview.get('path'),root,out,'assets/images',cache) if root and out and cache is not None else None
    return {'template_id':item.get('template_id'),'name':item.get('name'),'source':item.get('source'),'category':item.get('category'),'lane':item.get('lane'),'source_identity':item.get('source_identity',{}),'source_preview':{'original_path':preview.get('path'),'sha256':preview.get('sha256'),'local_path':local_preview,'availability':'copied' if local_preview else 'unavailable'},'semantic_summary':item.get('semantic_summary',{}),'fixed_source':{'adaptation_mode':fixed.get('adaptation_mode'),'qualification':fixed.get('qualification'),'slot_count':len(slots),'text_binding_count':sum(s.get('binding_count',0) for s in slots),'slots':slots,'source_bound_color_roles':fixed.get('source_bound_color_roles',{}),'source_bound_style_profiles':fixed.get('source_bound_style_profiles',{}),'fixed_areas':fixed.get('fixed_areas',[]),'retained_source_content':fixed.get('retained_source_content',[])},'adaptive_family_recommendation':item.get('adaptive_family_recommendation',{}),'adaptive_family_qualification':item.get('adaptive_family_qualification',{})}
def build(review_path,out_path,root,caps_path=None):
    review_path=Path(review_path).resolve(); out=Path(out_path).resolve(); root=Path(root).resolve()
    if out.exists():raise SystemExit(f'refusing to overwrite existing gallery: {out}')
    review=read_json(review_path)
    if not isinstance(review,dict) or not isinstance(review.get('slides'),list):raise SystemExit('review JSON must be an object with slides[]')
    if caps_path is None:caps_path=root/'library/adaptive/capabilities.json'
    else:
        caps_path=Path(caps_path); caps_path=caps_path if caps_path.is_absolute() else root/caps_path
    caps=read_json(caps_path); caprows=caps.get('items',[])
    cap_by_id={x.get('template_id'):x for x in caprows}
    out.mkdir(parents=True)
    cache={}; prepared=[]
    for i,row in enumerate(review['slides']):
        if not isinstance(row,dict):raise SystemExit(f'slides[{i}] must be an object')
        status=row.get('status','pending_native')
        if status not in STATUS:raise SystemExit(f"slides[{i}].status must be one of {sorted(STATUS)}")
        sid=str(row.get('id') or f'slide-{i+1:03d}')
        fam=str(row.get('family') or 'unmapped')
        variant=str(row.get('variant') or 'normal')
        tid=row.get('source_template_id')
        source_path=row.get('source_png')
        if not source_path and tid in cap_by_id:
            source_path=cap_by_id[tid].get('source_identity',{}).get('preview',{}).get('path')
        source_rel=copy_input(source_path,root,out,'assets/images',cache)
        render_rel=copy_input(row.get('render_png'),root,out,'assets/images',cache)
        spec_rel=copy_input(row.get('spec_path'),root,out,'assets/inputs',cache)
        values_rel=copy_input(row.get('values_path'),root,out,'assets/inputs',cache)
        evidence=row.get('evidence')
        evidence_links={}
        if isinstance(evidence,dict):
            for key,value in evidence.items():
                if isinstance(value,str):
                    copied=copy_input(value,root,out,'assets/evidence',cache)
                    if copied:evidence_links[key]=copied
        hashes=row.get('hashes') or {}
        integrity_mismatches=[]
        # A supplied hash is an identity claim for that exact artifact. Keep
        # the record visible, but never present a stale visually reviewed row
        # as reviewed when any referenced artifact has changed.
        hashed_inputs=[('spec_path','spec_sha256',row.get('spec_path')),
                       ('values_path','values_sha256',row.get('values_path')),
                       ('render_png','render_sha256',row.get('render_png'))]
        if isinstance(evidence,dict):
            hashed_inputs.extend([('evidence.deck','deck_sha256',evidence.get('deck')),
                                  ('evidence.native_verification','native_verification_sha256',evidence.get('native_verification')),
                                  ('evidence.fit_report','fit_report_sha256',evidence.get('fit_report'))])
        for label,key,raw in hashed_inputs:
            expected=hashes.get(key)
            if expected:
                actual=file_digest(raw,root)
                if actual != expected:
                    integrity_mismatches.append({'artifact':label,'path':raw,'expected_sha256':expected,'actual_sha256':actual})
        effective=status
        missing=[]
        if not source_rel:missing.append('source reference PNG')
        if not render_rel:missing.append('adaptive render PNG')
        if status == 'visually_reviewed' and integrity_mismatches: effective='needs_revision'
        elif status == 'visually_reviewed' and (not render_rel or not source_rel):effective='native_verified' if render_rel else 'pending_native'
        elif status == 'native_verified' and not render_rel:effective='pending_native'
        note=''
        if effective!=status:
            if integrity_mismatches:
                note='Gallery status downgraded to needs_revision because supplied artifact hashes do not match current files: '+', '.join(x['artifact'] for x in integrity_mismatches)+'. Re-run verification and visual review for these exact files.'
            else:
                note='Gallery status downgraded because '+(' and '.join(missing)+' unavailable. Visual comparison cannot be claimed.' if missing else 'required review evidence is unavailable.')
        prepared.append({'id':sid,'family':fam,'variant':variant,'source_template_id':tid,'source_png':source_rel,'render_png':render_rel,'spec_path':spec_rel,'values_path':values_rel,'archived_values_path':row.get('archived_values_path'),'declared_status':status,'review_state':effective,'status_note':note,'integrity_mismatches':integrity_mismatches,'findings':row.get('findings',[]),'controls':row.get('controls',{}),'limitations':row.get('limitations',[]),'slide_index':row.get('slide_index'),'evidence':evidence,'evidence_links':evidence_links,'hashes':hashes,'capability':compact_cap(cap_by_id[tid],root,out,cache) if tid in cap_by_id else None,'_source_input':source_path,'_render_input':row.get('render_png')})
    # Compare every nonbaseline variant against the normal record for the same source and family.
    baselines={}
    for r in prepared:
        key=(r['source_template_id'] or r['id'],r['family'])
        if r['variant'].lower() in {'normal','base','baseline'} and key not in baselines:baselines[key]=r
    for r in prepared:
        base=baselines.get((r['source_template_id'] or r['id'],r['family']))
        if r['variant'].lower() not in {'normal','base','baseline'} and base:
            a,b=flat(base.get('controls',{})),flat(r.get('controls',{})); changes=[]
            for k in sorted(set(a)|set(b)):
                if a.get(k)!=b.get(k):changes.append({'path':k,'normal':a.get(k),'variant':b.get(k)})
            r['control_changes_from_normal']=changes
            r['compared_to_variant']=base['variant']
        else:r['control_changes_from_normal']=[];r['compared_to_variant']=None
        r.pop('_source_input',None);r.pop('_render_input',None)
    compact_caps=[compact_cap(x,root,out,cache) for x in caprows]
    data={'schema':'pptxgengo.adaptive-review-gallery.v1','review_source':review_path.name,'capability_catalog':'library/adaptive/capabilities.json','summary':{'review_row_count':len(prepared),'render_available_count':sum(bool(x['render_png']) for x in prepared),'pending_native_count':sum(x['review_state']=='pending_native' for x in prepared),'visually_reviewed_count':sum(x['review_state']=='visually_reviewed' for x in prepared),'capability_record_count':len(caprows),'source_preview_available_count':sum(x['source_preview']['availability']=='copied' for x in compact_caps)},'filters':{'families':sorted({x['family'] for x in prepared}),'variants':sorted({x['variant'] for x in prepared}),'states':sorted(STATUS)},'slides':prepared,'capabilities':compact_caps,'family_builders':caps.get('family_builders',{}),'qualification_summary':caps.get('qualification_summary',{})}
    (out/'index.json').write_text(json.dumps(data,indent=2,ensure_ascii=False)+'\n',encoding='utf-8')
    html_doc=PAGE.replace('__DATA__',json.dumps(data,ensure_ascii=False).replace('</','<\\/').replace('\u2028','\\u2028').replace('\u2029','\\u2029'))
    (out/'index.html').write_text(html_doc,encoding='utf-8')
    return data

PAGE=r'''<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Adaptive slide review</title><style>
:root{--navy:#070154;--blue:#0047ff;--pale:#e8eef8;--ink:#17214a;--line:#d9deea;--muted:#5c6680;--green:#e8f5ee;--amber:#fff4d8;--red:#ffe7e7}*{box-sizing:border-box}body{margin:0;background:#f6f7fb;color:var(--ink);font:15px/1.48 -apple-system,BlinkMacSystemFont,"Segoe UI",Arial,sans-serif}header{background:var(--navy);color:white;padding:22px clamp(16px,4vw,56px)}header h1{font-size:26px;line-height:1.2;margin:0 0 5px}header p{margin:0;color:#dce2ff;max-width:1050px}.wrap{max-width:1500px;margin:auto;padding:20px clamp(12px,3vw,42px) 60px}.summary{display:flex;gap:10px;flex-wrap:wrap;margin:0 0 18px}.metric{background:white;border:1px solid var(--line);border-radius:10px;padding:10px 14px;min-width:145px}.metric strong{font-size:20px;display:block;color:var(--navy)}.tabs{display:flex;gap:8px;margin:14px 0}.tabs button,.filters button{border:1px solid #c7cee0;background:white;color:var(--navy);border-radius:7px;padding:8px 12px;cursor:pointer;font-weight:650}.tabs button.active,.filters button.active{background:var(--navy);color:white;border-color:var(--navy)}.filters{display:flex;gap:8px;flex-wrap:wrap;align-items:center;background:white;padding:12px;border:1px solid var(--line);border-radius:10px;margin-bottom:18px}.filters label{font-size:12px;color:var(--muted);font-weight:700}.filters select,.filters input{display:block;padding:7px 9px;border:1px solid var(--line);border-radius:6px;min-width:150px;color:var(--ink);background:white}.case{background:white;border:1px solid var(--line);border-radius:12px;margin:0 0 18px;overflow:hidden;box-shadow:0 2px 8px #1a21500a}.casehead{padding:14px 18px;border-bottom:1px solid var(--line);display:flex;gap:14px;align-items:flex-start;justify-content:space-between}.casehead h2{font-size:18px;margin:0;color:var(--navy)}.meta{font-size:12px;color:var(--muted);margin-top:4px}.badge{display:inline-block;border-radius:30px;padding:3px 9px;font-size:11px;font-weight:700;white-space:nowrap;background:var(--amber);color:#694b00}.badge.visually_reviewed{background:var(--green);color:#165a37}.badge.native_verified{background:#e9f0ff;color:#123c8c}.badge.needs_revision,.badge.blocked{background:var(--red);color:#8d1e1e}.pair{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:14px}.pane{border:1px solid var(--line);border-radius:9px;overflow:hidden;background:#fff}.pane h3{font-size:12px;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);padding:8px 10px;margin:0;border-bottom:1px solid var(--line)}.frame{min-height:150px;background:#f0f2f7;display:flex;justify-content:center;align-items:center}.frame img{display:block;width:100%;height:auto;object-fit:contain}.missing{color:var(--muted);padding:24px;text-align:center}.caption{padding:7px 10px;font-size:11px;color:var(--muted);border-top:1px solid var(--line);overflow-wrap:anywhere}.detailgrid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:0 14px 14px}.detail{border:1px solid var(--line);border-radius:8px;padding:10px;min-width:0}.detail h3{font-size:13px;margin:0 0 7px;color:var(--navy)}.detail p{margin:4px 0}.detail pre{margin:0;white-space:pre-wrap;overflow-wrap:anywhere;max-height:300px;overflow:auto;background:#f7f8fb;padding:9px;border-radius:6px;font-size:11px}.pills{display:flex;flex-wrap:wrap;gap:5px}.pill{font-size:11px;background:#f0f2f8;border-radius:4px;padding:2px 6px}.warning{background:var(--amber);padding:8px 12px;font-size:12px}.catalog{display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:12px}.cap{background:white;border:1px solid var(--line);border-radius:10px;padding:13px}.cap h3{margin:0;color:var(--navy);font-size:15px}.cap p{margin:5px 0;font-size:12px}.cap details{margin-top:8px}.cap summary{cursor:pointer;color:var(--blue);font-weight:650}.empty{background:white;padding:24px;border-radius:10px;color:var(--muted)}[hidden]{display:none!important}@media(max-width:780px){.pair,.detailgrid{grid-template-columns:1fr}.casehead{display:block}.casehead .badge{margin-top:7px}}
.cappreview{margin:8px -2px;background:#f0f2f7;border:1px solid var(--line);border-radius:7px;min-height:120px;display:flex;align-items:center;justify-content:center}.cappreview img{display:block;width:100%;max-height:220px;object-fit:contain}
</style></head><body><header><h1>Adaptive slide review</h1><p>Source slides are reference context. Adaptive slides are new semantic compositions. This gallery reports the supplied review state and never upgrades missing native renders into reviewed work.</p></header><main class="wrap"><section id="summary" class="summary"></section><nav class="tabs"><button data-tab="reviews" class="active">Review variants</button><button data-tab="catalog">65 source capabilities</button></nav><section id="filters" class="filters"></section><section id="reviews"></section><section id="catalog" hidden></section></main><script>
const DATA=__DATA__;
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const pretty=o=>JSON.stringify(o,null,2);
const capById=new Map(DATA.capabilities.map(x=>[x.template_id,x]));
const reviewSection=document.querySelector('#reviews'), catalogSection=document.querySelector('#catalog'), filters=document.querySelector('#filters');
let tab='reviews', family='all', variant='all', state='all', query='';
function select(label,id,values,selected){return `<label>${label}<select id="${id}"><option value="all">All</option>${values.map(x=>`<option ${x===selected?'selected':''} value="${esc(x)}">${esc(x)}</option>`).join('')}</select></label>`}
function renderFilters(){if(tab==='reviews'){filters.innerHTML=select('Family','family',DATA.filters.families,family)+select('Variant','variant',DATA.filters.variants,variant)+select('Review state','state',DATA.filters.states,state)+`<label>Search<input id="query" value="${esc(query)}" placeholder="slide, template, finding"></label>`;for(const k of ['family','variant','state'])filters.querySelector('#'+k).onchange=e=>{({family:v=>family=v,variant:v=>variant=v,state:v=>state=v})[k](e.target.value);render()};filters.querySelector('#query').oninput=e=>{query=e.target.value;renderReviews()};}else{filters.innerHTML=`<label>Taxonomy hint<select id="mapping"><option value="all">All</option><option value="category_candidate_requires_review">Candidate: structural review needed</option><option value="unmapped">No taxonomy hint</option></select></label><label>Family hint<select id="cfamily"><option value="all">All</option>${Object.keys(DATA.family_builders).map(x=>`<option>${esc(x)}</option>`).join('')}</select></label><label>Search<input id="cquery" placeholder="template ID, source, slot"></label>`;filters.querySelector('#mapping').onchange=renderCatalog;filters.querySelector('#cfamily').onchange=renderCatalog;filters.querySelector('#cquery').oninput=renderCatalog}}
function imgpane(title,path,caption){return `<div class="pane"><h3>${title}</h3><div class="frame">${path?`<img src="${esc(path)}" loading="lazy" alt="${title}">`:`<div class="missing">${title==='Adaptive render'?'Native render pending or image unavailable':'Source preview unavailable'}</div>`}</div><div class="caption">${esc(caption||'')}</div></div>`}
function stateBadge(s){return `<span class="badge ${esc(s)}">${esc(s.replaceAll('_',' '))}</span>`}
function card(r){const c=r.capability||capById.get(r.source_template_id)||{};const name=c.name||r.source_template_id||r.id;const diff=r.control_changes_from_normal||[];const changeHtml=diff.length?`<div><strong>${diff.length} control field(s) differ from ${esc(r.compared_to_variant)}</strong><pre>${esc(pretty(diff))}</pre></div>`:`<p>No matching normal variant is available for this comparison.</p>`;const evidenceLinks=Object.entries(r.evidence_links||{}).map(([k,v])=>`<a href="${esc(v)}">${esc(k)}</a>`).join(' · ');return `<article class="case"><div class="casehead"><div><h2>${esc(r.id)} · ${esc(r.family)} · ${esc(r.variant)}</h2><div class="meta">${esc(name)} · source template ${esc(r.source_template_id||'not linked')} · source category ${esc(c.category||'unknown')} ${r.slide_index!==null&&r.slide_index!==undefined?`· slide ${esc(r.slide_index)}`:''}</div></div>${stateBadge(r.review_state)}${r.declared_status!==r.review_state?`<span class="badge">declared ${esc(r.declared_status)}</span>`:''}</div>${r.status_note?`<div class="warning">${esc(r.status_note)}</div>`:''}<div class="pair">${imgpane('Source reference',r.source_png,`Fixed-source context. No source-layout transfer or fidelity is claimed.`)}${imgpane('Adaptive render',r.render_png,`Variant: ${r.variant}; state: ${r.review_state}`)}</div><div class="detailgrid"><section class="detail"><h3>Controls</h3><pre>${esc(pretty(r.controls||{}))}</pre></section><section class="detail"><h3>Controls delta vs normal</h3>${changeHtml}</section><section class="detail"><h3>Findings</h3>${Array.isArray(r.findings)&&r.findings.length?`<ul>${r.findings.map(x=>`<li>${esc(typeof x==='string'?x:pretty(x))}</li>`).join('')}</ul>`:'<p>No findings recorded.</p>'}</section><section class="detail"><h3>Limitations, evidence and inputs</h3>${r.slide_index!==null&&r.slide_index!==undefined?`<p>Slide index: ${esc(r.slide_index)}</p>`:""}${evidenceLinks?`<p class="pills">${evidenceLinks}</p>`:''}${r.evidence?`<details><summary>Evidence provenance paths</summary><pre>${esc(pretty(r.evidence))}</pre></details>`:""}${r.hashes?`<details><summary>Hashes</summary><pre>${esc(pretty(r.hashes))}</pre></details>`:""}${r.limitations?.length?`<ul>${r.limitations.map(x=>`<li>${esc(typeof x==='string'?x:pretty(x))}</li>`).join('')}</ul>`:'<p>No limitations supplied.</p>'}<div class="pills">${r.spec_path?`<a href="${esc(r.spec_path)}">spec JSON</a>`:''} ${r.values_path?`<a href="${esc(r.values_path)}">values JSON</a>`:''}</div></section></div></article>`}
function renderReviews(){const rows=DATA.slides.filter(r=>(family==='all'||r.family===family)&&(variant==='all'||r.variant===variant)&&(state==='all'||r.review_state===state)&&(!query||JSON.stringify(r).toLowerCase().includes(query.toLowerCase())));reviewSection.innerHTML=rows.length?rows.map(card).join(''):'<div class="empty">No review rows match these filters.</div>'}
function renderCatalog(){const mapping=filters.querySelector('#mapping')?.value||'all', f=filters.querySelector('#cfamily')?.value||'all', q=filters.querySelector('#cquery')?.value?.toLowerCase()||'';const rows=DATA.capabilities.filter(c=>(mapping==='all'||c.adaptive_family_recommendation?.status===mapping)&&(f==='all'||c.adaptive_family_recommendation?.family===f)&&(!q||JSON.stringify(c).toLowerCase().includes(q)));catalogSection.innerHTML=`<p class="note">Taxonomy hints are discovery aids only. Structural suitability is unreviewed; the five example patterns do not imply that other source templates can be adapted.</p><div class="catalog">${rows.map(c=>{const x=c.fixed_source||{};const slots=x.slots||[];const rec=c.adaptive_family_recommendation||{};const preview=c.source_preview||{};return `<article class="cap"><h3>${esc(c.template_id)} · ${esc(c.name)}</h3><div class="cappreview">${preview.local_path?`<img src="${esc(preview.local_path)}" loading="lazy" alt="Source slide reference for ${esc(c.template_id)}">`:'<span class="missing">Source preview unavailable</span>'}</div><p>Source preview · fixed-source reference only; no adaptive layout or fidelity claim</p><p>${esc(c.source)} · ${esc(c.category)} · ${esc(c.lane)}</p><p>${rec.status==='category_candidate_requires_review'?`Taxonomy discovery hint: ${esc(rec.family)} (structural suitability unreviewed)`:'No current taxonomy hint for the builder families'}</p><p>Fixed source: ${esc(x.qualification||'contract info unavailable')} · ${esc(x.slot_count||slots.length)} slots · ${esc(x.text_binding_count||slots.reduce((n,s)=>n+(s.binding_count||0),0))} bindings</p><p>Adaptive review: ${esc(c.adaptive_family_qualification?.state||'pending native review')}</p><details><summary>Slots, source roles, profiles, fixed areas</summary><pre>${esc(pretty({slots,roles:x.source_bound_color_roles,profiles:x.source_bound_style_profiles,fixed_areas:x.fixed_areas,retained_source_content:x.retained_source_content}))}</pre></details></article>`}).join('')}</div>`}
function render(){renderFilters();reviewSection.hidden=tab!=='reviews';catalogSection.hidden=tab!=='catalog';if(tab==='reviews')renderReviews();else renderCatalog()}
document.querySelectorAll('.tabs button').forEach(b=>b.onclick=()=>{tab=b.dataset.tab;document.querySelectorAll('.tabs button').forEach(x=>x.classList.toggle('active',x===b));render()});
const s=DATA.summary;document.querySelector('#summary').innerHTML=[['Review variants',s.review_row_count],['Rendered variants',s.render_available_count],['Pending native',s.pending_native_count],['Adaptive visually reviewed',s.visually_reviewed_count],['Fixed-source capabilities',s.capability_record_count]].map(([k,v])=>`<div class="metric"><strong>${esc(v)}</strong>${esc(k)}</div>`).join('');render();
</script></body></html>'''

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--review',required=True);ap.add_argument('--out',required=True);ap.add_argument('--root',default=str(Path(__file__).resolve().parents[1]));ap.add_argument('--capabilities',default='library/adaptive/capabilities.json');a=ap.parse_args();d=build(a.review,a.out,a.root,a.capabilities);print(f"wrote {len(d['slides'])} review rows and {len(d['capabilities'])} capability records to {Path(a.out).resolve()}")
if __name__=='__main__':main()
