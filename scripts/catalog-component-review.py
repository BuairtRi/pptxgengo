#!/usr/bin/env python3
"""Build a local, source-verified HTML gallery for curated component candidates."""
import argparse
import hashlib
import html
import json
import os
from pathlib import Path
import tempfile


def read_jsonl(path):
    with path.open(encoding='utf-8') as stream:
        for line_number, line in enumerate(stream, 1):
            if line.strip():
                try:
                    yield json.loads(line)
                except json.JSONDecodeError as exc:
                    raise ValueError(f'{path}:{line_number}: {exc}') from exc


def sha256(path):
    digest = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            digest.update(block)
    return digest.hexdigest()


def esc(value):
    return html.escape('' if value is None else str(value), quote=True)


def root_path(path):
    return (Path.cwd() / path).resolve() if not path.is_absolute() else path.resolve()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--components', type=Path, required=True)
    ap.add_argument('--geometry', type=Path, required=True)
    ap.add_argument('--previews', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    args = ap.parse_args()
    if args.out.exists():
        ap.error(f'refusing to overwrite {args.out}')

    comps = [c for c in read_jsonl(args.components) if c.get('origin') == 'curated_composition_candidate']
    if not comps:
        raise ValueError('no curated composition candidates')
    if len({c['id'] for c in comps}) != len(comps):
        raise ValueError('duplicate component IDs')
    needed = {oid for c in comps for oid in c.get('member_occurrence_ids', [])}
    needed.update(s['occurrence_id'] for c in comps for s in c.get('slots', []))
    geometry = {}
    for row in read_jsonl(args.geometry):
        oid = row['occurrence_id']
        if oid in needed:
            if oid in geometry:
                raise ValueError(f'duplicate geometry occurrence: {oid}')
            geometry[oid] = row
    missing = needed - geometry.keys()
    if missing:
        raise ValueError(f'missing geometry refs: {sorted(missing)[:8]}')

    manifest = json.loads(args.previews.read_text(encoding='utf-8'))
    source_hashes = {s['source_id']:s['source_sha256'] for s in manifest['sources']}
    previews = manifest['previews']
    preview_cache = {}
    cards = []
    unresolved = 0
    for c in comps:
        sid, source_hash = c['source_id'], c['source_sha256']
        if source_hashes.get(sid) != source_hash:
            raise ValueError(f"source manifest hash mismatch: {c['id']}")
        members = []
        for oid in c['member_occurrence_ids']:
            row = geometry[oid]
            if (row['source_id'],row['source_sha256'],row['source_part'],row['slide_number']) != (sid,source_hash,c['source_part'],c['slide_number']):
                raise ValueError(f"member geometry identity mismatch: {c['id']} {oid}")
            expected = f"{sid}:{source_hash[:12]}:{c['source_part']}:{row['object_path']}"
            if oid != expected:
                raise ValueError(f'noncanonical occurrence ID: {oid}')
            members.append(row)
        member_ids = set(c['member_occurrence_ids'])
        for slot in c.get('slots', []):
            oid = slot['occurrence_id']
            if oid not in member_ids or geometry[oid]['object_path'] != slot['object_path']:
                raise ValueError(f"slot geometry mismatch: {c['id']} {oid}")
        preview_key = f"{sid}:{c['slide_number']:03d}"
        preview = previews.get(preview_key)
        if preview is None or preview.get('source_slide_number') != c['slide_number']:
            raise ValueError(f"missing preview entry: {c['id']} {preview_key}")
        png = root_path(Path(preview['path']))
        if not png.is_file():
            raise FileNotFoundError(f'missing preview file: {png}')
        if png not in preview_cache:
            preview_cache[png] = sha256(png)
        if preview_cache[png] != preview['sha256']:
            raise ValueError(f'preview hash mismatch: {png}')
        sizes = {(r['slide_size_emu']['width'],r['slide_size_emu']['height']) for r in members}
        if len(sizes) != 1:
            raise ValueError(f"inconsistent slide dimensions: {c['id']}")
        width,height = sizes.pop()
        if width <= 0 or height <= 0:
            raise ValueError(f"invalid slide dimensions: {c['id']}")
        polygons = []
        for index,row in enumerate(members):
            poly = row.get('polygon_emu')
            if row['geometry_status'] != 'resolved' or not poly:
                unresolved += 1
                continue
            if len(poly) != 4 or any(len(p) != 2 for p in poly):
                raise ValueError(f"invalid frame polygon: {row['occurrence_id']}")
            pts = ' '.join(f'{float(x):.3f},{float(y):.3f}' for x,y in poly)
            label = f"{row['object_path']} · {row['object_kind']}"
            polygons.append(f'<polygon points="{esc(pts)}" class="frame" data-path="{esc(row["object_path"])}"><title>{esc(label)}</title></polygon>')
        relative = os.path.relpath(png, start=root_path(args.out.parent)).replace(os.sep,'/')
        image_html = (f'<div class="slide"><img src="{esc(relative)}" alt="Unmodified rendered source slide {esc(c["slide_number"])}" '
                      f'width="{int(preview["size_pixels"][0])}" height="{int(preview["size_pixels"][1])}">'
                      f'<svg class="overlay" viewBox="0 0 {width} {height}" preserveAspectRatio="none" aria-label="Object frame overlays">'
                      + ''.join(polygons) + '</svg></div>')
        rows = []
        for slot in c.get('slots', []):
            observed = slot.get('observed_text') or ''
            capacity = slot.get('measured_capacity')
            capacity_text = 'null (unmeasured)' if capacity is None else str(capacity)
            rows.append('<tr>' + ''.join(f'<td>{esc(v)}</td>' for v in (
                slot.get('name'),slot.get('object_path'),slot.get('kind'),
                observed,slot.get('observed_characters',len(observed)),capacity_text,
                slot.get('geometry_status'))) + '</tr>')
        slot_table = ('<table><thead><tr><th>Slot</th><th>Path</th><th>Kind</th><th>Observed text</th>'
                      '<th>Characters</th><th>Capacity</th><th>Geometry</th></tr></thead><tbody>'
                      + (''.join(rows) if rows else '<tr><td colspan="7">No named slots</td></tr>') + '</tbody></table>')
        constraints = ''.join(f'<li>{esc(x)}</li>' for x in c.get('adaptation_constraints', []))
        limits = ''.join(f'<li>{esc(x)}</li>' for x in c.get('limitations', []))
        unresolved_members = [r['object_path'] for r in members if r['geometry_status'] != 'resolved']
        status = ('Unresolved member frames: ' + ', '.join(unresolved_members)) if unresolved_members else f'{len(polygons)} member frames overlaid'
        cards.append(f'''<article class="card" id="{esc(c['id'])}">
<header><div><h2>{esc(c.get('name') or c['id'])}</h2><p class="meta">{esc(c['id'])} · {esc(sid)} slide {esc(c['slide_number'])} · {esc(status)}</p></div><span class="pill">{esc(c.get('readiness'))}</span></header>
{image_html}
<p class="purpose">{esc(c.get('purpose') or 'Purpose not classified.')}</p>
{slot_table}
<details><summary>Readiness and constraints</summary><p>Design preference: {esc(c.get('design_preference'))} · Content approval: {esc(c.get('content_approval'))} · Supported operations: {esc(', '.join(c.get('supported_operations', [])))}</p><h3>Adaptation constraints</h3><ul>{constraints or '<li>None recorded</li>'}</ul><h3>Limitations</h3><ul>{limits or '<li>None recorded</li>'}</ul></details>
</article>''')

    page = f'''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Curated component review</title>
<style>
:root{{color-scheme:light;font-family:system-ui,-apple-system,Segoe UI,sans-serif;color:#17232d;background:#f2f4f6}}
*{{box-sizing:border-box}}body{{margin:0}}.top{{position:sticky;top:0;z-index:2;background:#fff;border-bottom:1px solid #cbd3da;padding:16px 24px;display:flex;align-items:center;justify-content:space-between;gap:20px}}
h1{{font-size:22px;margin:0}}.top p{{margin:5px 0 0;color:#53626e;font-size:13px}}label{{white-space:nowrap;font-weight:600}}main{{max-width:1250px;margin:22px auto;padding:0 18px;display:grid;gap:24px}}
.card{{background:#fff;border:1px solid #d2d9df;border-radius:12px;overflow:hidden;box-shadow:0 2px 9px #18212b0d}}.card header{{display:flex;justify-content:space-between;align-items:start;padding:17px 20px;gap:16px}}
h2{{font-size:20px;margin:0 0 4px}}.meta{{font-size:12px;color:#60707e;margin:0;overflow-wrap:anywhere}}.pill{{font-size:12px;border-radius:99px;padding:5px 9px;background:#eaf0f5;white-space:nowrap}}
.slide{{position:relative;width:100%;background:#111;line-height:0}}.slide img{{display:block;width:100%;height:auto}}.overlay{{position:absolute;inset:0;width:100%;height:100%;pointer-events:none;overflow:visible}}
.frame{{fill:#ffcc0030;stroke:#f7b500;stroke-width:18000;vector-effect:non-scaling-stroke;stroke-width:2.5px}}body.hide-overlays .overlay{{display:none}}
.purpose{{margin:16px 20px;color:#34434d}}table{{border-collapse:collapse;width:calc(100% - 40px);margin:0 20px 18px;font-size:13px}}th,td{{padding:8px;border-bottom:1px solid #e1e6ea;text-align:left;vertical-align:top;overflow-wrap:anywhere}}th{{background:#f3f6f8}}td:nth-child(4){{white-space:pre-line;max-width:350px}}details{{border-top:1px solid #e1e6ea;padding:12px 20px 18px;font-size:13px}}summary{{cursor:pointer;font-weight:650}}li{{margin:4px 0}}details h3{{font-size:13px;margin-bottom:4px}}
</style></head><body><div class="top"><div><h1>Curated component review</h1><p>{len(cards)} source-bound candidates · Preview PNGs are unmodified · Yellow outlines show object frames, not visible ink or fit guarantees</p></div><label><input id="toggle" type="checkbox" checked> Show frame overlays</label></div>
<main>{''.join(cards)}</main><script>document.getElementById('toggle').addEventListener('change',e=>document.body.classList.toggle('hide-overlays',!e.target.checked));</script></body></html>'''
    args.out.parent.mkdir(parents=True, exist_ok=True)
    fd, temp = tempfile.mkstemp(prefix='.'+args.out.name+'.',suffix='.tmp',dir=args.out.parent)
    try:
        with os.fdopen(fd,'w',encoding='utf-8') as stream:
            stream.write(page)
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temp,args.out)
    finally:
        os.unlink(temp)
    print(json.dumps({'components':len(cards),'unresolved_member_frames':unresolved,'verified_previews':len(preview_cache),'output':str(args.out)}))


if __name__ == '__main__':
    main()
