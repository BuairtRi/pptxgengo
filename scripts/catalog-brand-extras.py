#!/usr/bin/env python3
"""Ingest local font files and explicitly path-identified illustration candidates.

Uses only the standard library; does not download, infer rights, or modify sources.
"""
from __future__ import annotations
import argparse
import collections
import hashlib
import json
import os
import struct
import uuid
from pathlib import Path

FONT_EXTS={'.ttf','.otf','.woff','.woff2'}
ILLUST_EXTS={'.svg','.ai','.eps','.pdf','.png','.jpg','.jpeg','.webp'}


def digest(path):
    h=hashlib.sha256()
    with path.open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''): h.update(b)
    return h.hexdigest()

def decode_name(raw, platform, enc):
    try:
        if platform in (0,3): return raw.decode('utf-16-be').rstrip('\0')
        if platform==1: return raw.decode('mac_roman').rstrip('\0')
    except (UnicodeDecodeError,LookupError): pass
    return None

def sfnt_info(path):
    """Safely extract selected SFNT name IDs and OS/2 fsType; no font execution."""
    try:
        data=path.read_bytes()
        if len(data)<12 or data[:4] not in (b'\x00\x01\x00\x00',b'OTTO',b'true',b'typ1'):
            return {'font_name':None,'family_name':None,'style':None,'fsType':None,'sfnt_status':'unsupported-or-invalid-sfnt'}
        n=struct.unpack_from('>H',data,4)[0]
        if 12+16*n>len(data): return {'font_name':None,'family_name':None,'style':None,'fsType':None,'sfnt_status':'invalid-table-directory'}
        tables={}
        for i in range(n):
            off=12+i*16; tag=data[off:off+4].decode('ascii','replace')
            _,to,ln=struct.unpack_from('>III',data,off+4)
            if to<=len(data) and ln<=len(data)-to: tables[tag]=(to,ln)
        names={}
        if 'name' in tables:
            to,ln=tables['name']; block=data[to:to+ln]
            if len(block)>=6:
                count,stringoff=struct.unpack_from('>HH',block,2)
                for j in range(count):
                    p=6+j*12
                    if p+12>len(block): break
                    plat,enc,lang,nid,nlen,noff=struct.unpack_from('>HHHHHH',block,p)
                    a=stringoff+noff; b=a+nlen
                    if nid not in (1,2,4,6,16,17) or b>len(block): continue
                    val=decode_name(block[a:b],plat,enc)
                    if val:
                        # Prefer typographic family/subfamily then Unicode/Windows.
                        rank=(0 if nid in (16,17) else 1 if nid in (1,2,4) else 2, 0 if plat in (0,3) else 1)
                        names.setdefault(nid,[]).append((rank,val))
        def best(nid):
            vals=names.get(nid,[])
            return min(vals,key=lambda x:x[0])[1] if vals else None
        family=best(16) or best(1)
        style=best(17) or best(2)
        full=best(4)
        fs=None
        if 'OS/2' in tables:
            to,ln=tables['OS/2']
            if ln>=10 and to+10<=len(data): fs=struct.unpack_from('>H',data,to+8)[0]
        return {'font_name':full or (family+' '+style if family and style else family), 'family_name':family,'style':style,'fsType':fs,'sfnt_status':'parsed'}
    except (OSError,struct.error,OverflowError):
        return {'font_name':None,'family_name':None,'style':None,'fsType':None,'sfnt_status':'read-error'}

def rights_from_fstype(value):
    if value is None: return {'embedding_rights':'UNKNOWN','embedding_flags':[]}
    flags=[]
    if value & 0x0002: flags.append('restricted_license_embedding')
    if value & 0x0004: flags.append('preview_and_print_embedding')
    if value & 0x0008: flags.append('editable_embedding')
    if value & 0x0100: flags.append('no_subsetting')
    if value & 0x0200: flags.append('bitmap_embedding_only')
    if value==0: flags=['fsType_zero']
    unknown=value & ~0x030e
    if unknown: flags.append(f'unknown_bits_0x{unknown:04x}')
    return {'embedding_rights':'UNKNOWN','embedding_flags':flags,'embedding_fstype':value}

def ingest(a):
    brand=Path(a.brand_root).expanduser().resolve(); invpath=Path(a.hosted_inventory).expanduser().resolve()
    if not brand.is_dir(): raise ValueError(f'Brand root is not a directory: {brand}')
    inventory=json.loads(invpath.read_text(encoding='utf-8'))
    out=Path(a.out); report=Path(a.report)
    if out.exists() or report.exists(): raise SystemExit('refusing to overwrite existing output/report')
    recs={}; errors=[]; font_paths=[]; illustration_paths=[]
    for path in sorted(brand.rglob('*')):
        if not path.is_file(): continue
        ext=path.suffix.lower(); rel=path.relative_to(brand).as_posix()
        if ext in FONT_EXTS:
            font_paths.append((path,rel)); continue
        # Only explicit illustration directory/name path metadata qualifies. Icons,
        # hand-drawn arrows, accents, and generic graphics are not illustrations.
        low=rel.lower()
        if ext in ILLUST_EXTS and ('/illustrations/' in '/'+low or 'illustration' in path.stem.lower()):
            illustration_paths.append((path,rel))
    for path,rel in font_paths+illustration_paths:
        try: h=digest(path); size=path.stat().st_size
        except OSError as e: errors.append(f'{rel}: {e}'); continue
        ref={'source':'branding-filesystem','path':rel,'root':str(brand)}
        if path.suffix.lower() in FONT_EXTS:
            info=sfnt_info(path); rights=rights_from_fstype(info.pop('fsType'))
            rec={'id':'sha256:'+h,'kind':'font','category':'font','content_sha256':h,'size_bytes':size,
                 'font_name':info['font_name'],'family_name':info['family_name'],'style':info['style'],
                 'font_metadata_status':info['sfnt_status'],**rights,'readiness':'discovered',
                 'origin':'local-branding-font-file','source_refs':[ref],'aliases':[rel]}
        else:
            descriptions=[]
            side=path.with_name(path.name+'.md')
            if side.is_file():
                try:
                    descriptions.append({'path':side.relative_to(brand).as_posix(), 'text':side.read_bytes().decode('utf-8'),
                                         'authority':'existing-branding-sidecar', 'preservation':'verbatim decoded UTF-8 text'})
                except (OSError,UnicodeError) as e:
                    errors.append(f'{side.relative_to(brand).as_posix()}: {e}')
            rec={'id':'sha256:'+h,'kind':'asset','category':'illustration','content_sha256':h,'size_bytes':size,
                 'readiness':'discovered','origin':'local-explicit-illustration-path','source_refs':[ref],'aliases':[rel],
                 'description_sources':descriptions,'visual_description':None,'metadata_status':'path-identified candidate; visual review required'}
        if rec['id'] in recs:
            old=recs[rec['id']]
            old['aliases'].append(rel); old['source_refs'].append(ref)
        else: recs[rec['id']]=rec
    # Hosted inventory has no explicit illustration path. Do not classify handdrawn
    # accents as illustration candidates; report the observed gap instead.
    out.parent.mkdir(parents=True,exist_ok=True); report.parent.mkdir(parents=True,exist_ok=True)
    rows=sorted(recs.values(),key=lambda r:r['id'])
    extcounts=collections.Counter(Path(rel).suffix.lower() for _,rel in font_paths)
    s=inventory.get('assets',[])
    remote_illustrations=[x['path'] for x in s if 'illustration' in x['path'].lower()]
    nfont=sum(r['kind']=='font' for r in rows); nillustr=sum(r['category']=='illustration' for r in rows)
    # Stage, then exclusively link both outputs. Roll back a partial pair on errors.
    staged=[]; committed=[]
    try:
        for target,data in ((out,''.join(json.dumps(r,ensure_ascii=False,separators=(',',':'))+'\n' for r in rows)),(report,None)):
            if data is None: continue
            tmp=target.with_name(target.name+'.stage-'+uuid.uuid4().hex)
            fd=os.open(tmp,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o644)
            with os.fdopen(fd,'w',encoding='utf-8',newline='\n') as f:f.write(data)
            staged.append((tmp,target))
        illustration_dir=brand/'assets'/'illustrations'
        illustration_dir_exists=illustration_dir.is_dir()
        candidate_paths=[rel for _,rel in illustration_paths]
        report_text=f'''# Brand extras ingestion report

Read-only discovery run; source files and the existing asset manifest were not modified. No network requests or downloads were made.

## Runtime findings

- Configured branding root `{brand}`: local font files found: {len(font_paths)} ({nfont} unique SHA-256 records); extensions: {dict(extcounts)}.
- Configured branding root `{brand}`: explicit local illustration candidates: {len(illustration_paths)} files ({nillustr} unique SHA-256 records). Candidate paths: {', '.join(candidate_paths) if candidate_paths else 'none'}. Candidate rule: image/vector file inside an `illustrations` path or filename explicitly containing `illustration`; general icons, arrows, graphic accents, photos, and templates are not labeled illustrations.
- Configured hosted inventory `{invpath}`: {len(s)} records; illustration-named paths: {len(remote_illustrations)}. Paths: {', '.join(remote_illustrations) if remote_illustrations else 'none'}.
- Fonts have `readiness: discovered`; SFNT family/full name, subfamily/style, and OS/2 `fsType` were extracted when safely available. `embedding_rights` remains `UNKNOWN`; parsed flags report bitfield indicators, not a permission opinion. WOFF/WOFF2, if encountered, remain identifiable by file/hash but metadata is unsupported unless safely parseable.
- Every record includes a content SHA-256, byte size, local provenance path, and aliases for same-hash files. IDs use `sha256:<digest>` like physical assets; this supplementary manifest is a separate source and should be merged without overwriting existing rows.

## Illustration gap and static context

For this run, `{illustration_dir}` {'exists and contains no candidate files' if illustration_dir_exists and not illustration_paths else ('does not exist' if not illustration_dir_exists else f'has {len(illustration_paths)} candidate file(s)')}. The runtime count above describes the configured sources only and should not be read as a permanent claim about other roots or future inventory versions. Hosted hand-drawn graphics remain categorized separately.

Static local context checked: the current branding workspace includes an Illustration guide describing Adobe Illustrator artwork and linking to an external S3 sample image. That remote image was not downloaded and is not a local source asset. Likely places to inspect manually include the linked Illustration Suite/source library and supplied PowerPoint templates/decks for embedded vector artwork; likely source formats include `.ai`, `.eps`, `.svg`, and high-resolution raster. These are leads, not discovered catalog assets.

## Excluded file types

- Scan target types: TTF, OTF, WOFF/WOFF2 font files; SVG, AI, EPS, PDF, PNG, JPG/JPEG, WEBP only when explicitly illustration-path identified.
- Existing presentation/document sources, ordinary icons, logos, accent graphics, and photographs remain outside this supplementary manifest.

## Read errors

'''+('\n'.join('- '+x for x in errors) if errors else '- None.')+'\n'
        tmp=report.with_name(report.name+'.stage-'+uuid.uuid4().hex)
        fd=os.open(tmp,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o644)
        with os.fdopen(fd,'w',encoding='utf-8',newline='\n') as f:f.write(report_text)
        staged.append((tmp,report))
        for tmp,target in staged: os.link(tmp,target); committed.append(target)
        for tmp,_ in staged: tmp.unlink()
    except Exception:
        for target in committed:
            try: target.unlink()
            except OSError: pass
        for tmp,_ in staged:
            try: tmp.unlink()
            except OSError: pass
        raise
    print(json.dumps({'font_files':len(font_paths),'font_records':nfont,'illustration_files':len(illustration_paths),'illustration_records':nillustr,'out':str(out),'report':str(report)},indent=2))

def main():
    p=argparse.ArgumentParser(description=__doc__); sub=p.add_subparsers(dest='cmd',required=True)
    q=sub.add_parser('ingest'); q.add_argument('--brand-root',required=True); q.add_argument('--hosted-inventory',required=True); q.add_argument('--out',required=True); q.add_argument('--report',required=True); q.set_defaults(fn=ingest)
    a=p.parse_args(); a.fn(a)
if __name__=='__main__': main()
