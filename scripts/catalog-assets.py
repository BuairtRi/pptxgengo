#!/usr/bin/env python3
"""Ingest local West Monroe brand assets and a hosted asset inventory as JSONL.

No downloads are performed. Existing Markdown photo sidecars are stored verbatim.
"""
from __future__ import annotations
import argparse
import collections
import hashlib
import json
import os
import re
import struct
import sys
import tempfile
import uuid
import xml.etree.ElementTree as ET
from pathlib import Path

IMAGE_EXTS = {'.png', '.jpg', '.jpeg', '.gif', '.webp', '.tif', '.tiff', '.bmp', '.svg', '.eps'}
TEXT_EXTS = {'.md', '.txt'}

def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()

def dimensions(path: Path):
    """Return dimensions only when safely derivable with stdlib; otherwise nulls."""
    try:
        ext = path.suffix.lower()
        if ext == '.png':
            with path.open('rb') as f:
                head = f.read(24)
            if head[:8] == b'\x89PNG\r\n\x1a\n' and head[12:16] == b'IHDR':
                return list(struct.unpack('>II', head[16:24]))
        if ext in {'.jpg', '.jpeg'}:
            with path.open('rb') as f:
                if f.read(2) != b'\xff\xd8': return None
                while True:
                    b=f.read(1)
                    if not b: return None
                    if b != b'\xff': continue
                    while b == b'\xff':
                        b=f.read(1)
                        if not b: return None
                    marker=b[0]
                    if marker in (0xd8,0xd9) or 0xd0 <= marker <= 0xd7: continue
                    raw=f.read(2)
                    if len(raw)<2: return None
                    length=int.from_bytes(raw,'big')
                    if length < 2: return None
                    if marker in (0xc0,0xc1,0xc2,0xc3,0xc5,0xc6,0xc7,0xc9,0xca,0xcb,0xcd,0xce,0xcf):
                        seg=f.read(5)
                        if len(seg)<5: return None
                        h=int.from_bytes(seg[1:3],'big'); w=int.from_bytes(seg[3:5],'big')
                        return [w,h]
                    f.seek(length-2,1)
        if ext == '.svg':
            root=ET.parse(path).getroot(); w=root.get('width'); h=root.get('height')
            def px(v):
                m=re.fullmatch(r'\s*(\d+(?:\.\d+)?)\s*(?:px)?\s*',v or '')
                return int(float(m.group(1))) if m else None
            wd,ht=px(w),px(h)
            if wd and ht: return [wd,ht]
    except (OSError, ValueError, ET.ParseError, struct.error):
        pass
    return None

def svg_viewbox(path: Path):
    if path.suffix.lower() != '.svg': return None
    try:
        return ET.parse(path).getroot().get('viewBox')
    except (OSError, ET.ParseError):
        return None

def dimensions_source(path: Path, dims):
    if not dims: return 'unknown'
    ext=path.suffix.lower()
    return {'.png':'png-ihdr','.jpg':'jpeg-sof','.jpeg':'jpeg-sof','.svg':'svg-numeric-width-height'}.get(ext,'unknown')

def category_for(rel: str) -> str:
    p=rel.lower().replace('\\','/'); name=Path(p).name.lower()
    if '/logos/' in p or p.startswith(('assets/logos/','logos/')): return 'logo'
    if '/icons/' in p or p.startswith(('assets/icons/','icons/')): return 'icon'
    if '/illustrations/' in p or p.startswith(('assets/illustrations/','illustrations/')): return 'illustration'
    if '/graphics/' in p or 'handdrawn' in p or '/accents/' in p or 'arrow' in name or 'highlight' in name: return 'accent'
    if 'west monroe photos/' in p or p.startswith('photos/') or '/photography/' in p or '/images/' in p: return 'photo'
    return 'other'

def source_ref(source, path, **extra):
    d={'source':source,'path':path}; d.update(extra); return d

def ingest(args):
    out=Path(args.out); report=Path(args.report)
    for p in (out, report):
        if p.exists(): raise SystemExit(f'refusing to overwrite existing output: {p}')
    brand=Path(args.brand_root).expanduser().resolve()
    inv_path=Path(args.hosted_inventory).expanduser().resolve()
    inv=json.loads(inv_path.read_text(encoding='utf-8'))
    records={}; local_hash_paths=collections.defaultdict(list)
    sidecar_count=0; scan_errors=[]
    # Sidecars are attached verbatim to their paired photo record. They are read as
    # UTF-8 text and JSON-escaped by the serializer; no normalization is applied.
    for base in (brand/'assets', brand/'West Monroe Photos'):
        if not base.exists(): continue
        for path in sorted(base.rglob('*')):
            if not path.is_file() or path.suffix.lower() not in IMAGE_EXTS: continue
            if path.name.startswith('.'): continue
            try:
                digest=sha256(path); dims=dimensions(path)
            except OSError as e:
                scan_errors.append(f'{path}: {e}'); digest=None; dims=None
            rel=path.relative_to(brand).as_posix()
            key=('sha256:'+digest) if digest else ('local:'+rel)
            cat=category_for(rel)
            rec=records.setdefault(key, {'id':key,'kind':'asset','category':cat,'categories':[cat],'category_status':'path-inferred','source_refs':[],
                'content_sha256':digest,'dimensions_px':dims,'dimensions_source':dimensions_source(path,dims),'svg_viewbox':svg_viewbox(path),'description_sources':[],
                'remote_metadata':[],'aliases':[],'variant_group_id':None,'variant_status':'unconfirmed'})
            if cat not in rec['categories']: rec['categories'].append(cat)
            ref=source_ref('branding-filesystem',rel,root_label=base.name)
            if ref not in rec['source_refs']: rec['source_refs'].append(ref)
            rec['aliases'].append(rel)
            local_hash_paths[digest].append((rel,key)) if digest else None
            side=path.with_name(path.name+'.md')
            if side.is_file():
                try: text=side.read_bytes().decode('utf-8'); sidecar_count+=1
                except (OSError,UnicodeError) as e:
                    scan_errors.append(f'{side}: {e}'); continue
                desc={'path':side.relative_to(brand).as_posix(),'text':text,'authority':'existing-branding-sidecar','preservation':'verbatim decoded UTF-8 text'}
                if desc not in rec['description_sources']: rec['description_sources'].append(desc)
    # Hosted records remain individually source-scoped: no byte identity is known
    # without downloading. Exact local path matches are recorded only as evidence.
    remote_local_exact=0
    for item in inv.get('assets',[]):
        rpath=item['path']; rid='hosted:'+rpath
        cat=category_for(rpath)
        rec={'id':rid,'kind':'asset','category':cat,'categories':[cat],'category_status':'path-inferred','source_refs':[source_ref('wm-brand-assets-inventory',rpath,canonical_url=inv.get('assetHostDefault','').rstrip('/')+'/'+rpath)],
             'content_sha256':None,'dimensions_px':None,'dimensions_source':'unknown','svg_viewbox':None,'description_sources':[],
             'remote_metadata':[{k:item.get(k) for k in ('fileName','directory','extension','assetType','sizeBytes','modifiedAt')}],
             'aliases':[],'variant_group_id':None,'variant_status':'unconfirmed',
             'metadata_availability':{'file_metadata':True,'visual_description':False,'alt_text':False,'keywords':False,'pixel_dimensions':False,'content_hash':False,'downloaded':False}}
        # Local filename/path resemblance is not a match; exact relative path only.
        local_candidate=brand/rpath
        if local_candidate.is_file():
            try:
                h=sha256(local_candidate)
                rec['local_hash_evidence']=h
                remote_local_exact+=1
            except OSError: pass
        records[rid]=rec
    # Group exact local duplicate hashes by one canonical record and retain refs.
    # records above already keyed by hash, so aliases/source refs remain together.
    out.parent.mkdir(parents=True,exist_ok=True); report.parent.mkdir(parents=True,exist_ok=True)
    ordered=sorted(records.values(),key=lambda x:x['id'])
    counts=collections.Counter(cat for r in ordered for cat in r['categories'])
    source_counts=collections.Counter(ref['source'] for r in ordered for ref in r['source_refs'])
    dup_hashes={h:paths for h,paths in local_hash_paths.items() if len(paths)>1}
    report_text=f'''# Asset ingestion report

Run mode: bounded local ingestion; no network requests or asset downloads.

## Counts

- Hosted inventory records read: {len(inv.get('assets', []))}
- Canonical JSONL asset records written: {len(ordered)} (local exact-byte duplicates collapse to SHA-256 IDs; hosted entries remain source-scoped)
- Existing paired photo sidecars preserved verbatim: {sidecar_count}
- Local duplicate-hash groups: {len(dup_hashes)}
- Exact hosted-path/local-file candidates: {remote_local_exact}

| Category | Catalog records |
|---|---:|
'''+''.join(f'| {k} | {counts[k]} |\n' for k in ('photo','icon','illustration','logo','accent','other'))+'''
## Provenance and coverage

- Local records point to paths relative to the configured branding root and carry SHA-256 when readable. PNG/JPEG dimensions are parsed with the Python standard library; SVG pixel dimensions are used only when explicit numeric px or unitless width/height are present. SVG viewBox is retained separately as coordinate metadata, never converted to pixels. Unsupported/unreadable formats have null dimensions; no values are guessed.
- Sidecar text is stored as decoded UTF-8 exactly as read and JSON-escaped in JSONL; it is not parsed, summarized, or rewritten. Provenance identifies the relative sidecar path and `existing-branding-sidecar` authority.
- Hosted rows retain source inventory metadata and canonical hosted URL. Since binaries are not downloaded, their hash and dimensions are null; remote visual descriptions, alt text, and keywords are marked unavailable. Filename-derived descriptions are not invented.
- Distinct source files sharing a local hash are represented by one canonical hash ID with every path in `aliases` and `source_refs`. Hosted/local matches are not claimed from basename similarity. Similar filename variants are retained as separate rows with `variant_status: unconfirmed` and no inferred family link.
- Categories are path-inferred and retained as a list when duplicate hashes have aliases from multiple categories. Illustration coverage is measured from files actually found; an empty category is reported as zero. `other` counts unclassified image files; fonts and templates are outside the scan.

## Errors

'''+('\n'.join('- '+x for x in scan_errors) if scan_errors else '- None.')+'\n'
    # Stage both artifacts with exclusive temporary creation. Commit via hard links
    # so existing destinations are never overwritten, and roll back a partial pair.
    staged=[]; committed=[]
    try:
        for target, data in ((out, '\n'.join(json.dumps(row,ensure_ascii=False,separators=(',',':')) for row in ordered)+'\n'), (report, report_text)):
            tmp=target.with_name(target.name+'.stage-'+uuid.uuid4().hex)
            fd=os.open(tmp,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o644)
            with os.fdopen(fd,'w',encoding='utf-8',newline='\n') as f: f.write(data)
            staged.append((tmp,target))
        for tmp,target in staged:
            os.link(tmp,target); committed.append(target)
        for tmp,_ in staged: tmp.unlink()
    except Exception:
        for target in committed:
            try: target.unlink()
            except OSError: pass
        for tmp,_ in staged:
            try: tmp.unlink()
            except OSError: pass
        raise
    print(json.dumps({'records':len(ordered),'hosted':len(inv.get('assets',[])),'sidecars_preserved':sidecar_count,'categories':dict(counts),'out':str(out),'report':str(report)},indent=2))

def main():
    p=argparse.ArgumentParser(description=__doc__); sub=p.add_subparsers(dest='command',required=True)
    q=sub.add_parser('ingest',help='ingest local assets plus hosted inventory metadata without downloads')
    q.add_argument('--brand-root',required=True); q.add_argument('--hosted-inventory',required=True)
    q.add_argument('--out',required=True); q.add_argument('--report',required=True)
    q.set_defaults(func=ingest)
    args=p.parse_args(); args.func(args)
if __name__=='__main__': main()
