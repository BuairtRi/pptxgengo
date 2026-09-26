#!/usr/bin/env python3
"""Export source-bound slide-object geometry from registered PPTX originals.

One JSONL row is emitted for every slide object node, including group containers.
Coordinates describe transformed rectangular object frames, not painted pixels.
"""
import argparse
from collections import Counter
import hashlib
import json
import math
import os
from pathlib import Path
import posixpath
import tempfile
import xml.etree.ElementTree as ET
from zipfile import ZipFile

NS = {'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
      'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
KINDS = {'sp', 'pic', 'graphicFrame', 'cxnSp', 'grpSp', 'contentPart'}
EMU_INCH = 914400
I = (1., 0., 0., 1., 0., 0.)


def local(tag):
    return tag.rsplit('}', 1)[-1]


def sha256(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def atomic_create(path, writer):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp = tempfile.mkstemp(prefix='.' + path.name + '.', suffix='.tmp', dir=path.parent)
    try:
        with os.fdopen(fd, 'w', encoding='utf-8', newline='\n') as stream:
            writer(stream)
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temp, path)  # Exclusive installation; never replace a prior artifact.
    finally:
        os.unlink(temp)


def rels(zf, part):
    rp = posixpath.join(posixpath.dirname(part), '_rels', posixpath.basename(part) + '.rels')
    if rp not in zf.namelist():
        return []
    out = []
    for rel in ET.fromstring(zf.read(rp)):
        d = dict(rel.attrib)
        if d.get('TargetMode') != 'External':
            d['resolved_part'] = posixpath.normpath(posixpath.join(posixpath.dirname(part), d['Target'])).lstrip('/')
        out.append(d)
    return out


def related(zf, part, suffix):
    matches = [r['resolved_part'] for r in rels(zf, part)
               if r.get('Type', '').endswith('/' + suffix) and 'resolved_part' in r]
    return matches[0] if len(matches) == 1 else None


def objects(root):
    tree = root.find('p:cSld/p:spTree', NS)
    return [n for n in tree if local(n.tag) in KINDS] if tree is not None else []


def walk(nodes, prefix=''):
    for index, node in enumerate(nodes, 1):
        path = prefix + '/' + str(index) if prefix else str(index)
        yield path, node
        if local(node.tag) == 'grpSp':
            yield from walk([n for n in node if local(n.tag) in KINDS], path)


def nonvisual(node):
    for child in node:
        if local(child.tag).startswith('nv'):
            prop = next((n for n in child if local(n.tag) == 'cNvPr'), None)
            ph = next((n for n in child.iter() if local(n.tag) == 'ph'), None)
            return ((dict(prop.attrib) if prop is not None else {}),
                    (dict(ph.attrib) if ph is not None else None))
    return {}, None


def xfrm(node):
    for child in node:
        if local(child.tag) == 'xfrm':
            return child
    for child in node:
        if local(child.tag) in {'spPr', 'grpSpPr'}:
            found = child.find('a:xfrm', NS)
            if found is not None:
                return found
    return None


def raw_xfrm(node):
    return None if node is None else {'attributes': dict(node.attrib),
                                     **{local(n.tag): dict(n.attrib) for n in node}}


def num(d, key):
    try:
        return int(d[key])
    except (KeyError, TypeError, ValueError):
        return None


def mul(a, b):
    """Affine a after b, represented by (a,b,c,d,tx,ty)."""
    A, B, C, D, X, Y = a
    e, f, g, h, u, v = b
    return (A*e+C*f, B*e+D*f, A*g+C*h, B*g+D*h,
            A*u+C*v+X, B*u+D*v+Y)


def apply(m, x, y):
    a,b,c,d,u,v = m
    return (a*x+c*y+u, b*x+d*y+v)


def centered_transform(off, ext, attributes):
    """Frame-local rectangle to parent coordinates; DrawingML positive rot is clockwise onscreen."""
    ox, oy, w, h = off[0], off[1], ext[0], ext[1]
    angle = (num(attributes, 'rot') or 0) / 60000 * math.pi / 180
    cos, sin = math.cos(angle), math.sin(angle)
    fx = -1 if attributes.get('flipH') in {'1', 'true'} else 1
    fy = -1 if attributes.get('flipV') in {'1', 'true'} else 1
    cx, cy = ox + w/2, oy + h/2
    return mul((1,0,0,1,cx,cy),
               mul((cos,sin,-sin,cos,0,0),
                   mul((fx,0,0,fy,0,0), (1,0,0,1,-w/2,-h/2))))


def frame_parts(transform):
    if transform is None:
        return None
    off, ext = transform.get('off'), transform.get('ext')
    if off is None or ext is None:
        return None
    vals = (num(off, 'x'), num(off, 'y'), num(ext, 'cx'), num(ext, 'cy'))
    return vals if None not in vals else None


def geometry(transform, parent_matrix, is_group):
    vals = frame_parts(transform)
    if vals is None or parent_matrix is None:
        return None, None
    x,y,w,h = vals
    if w < 0 or h < 0:
        return None, None
    own = centered_transform((x,y), (w,h), transform['attributes'])
    matrix = mul(parent_matrix, own)
    polygon = [apply(matrix, px, py) for px,py in ((0,0),(w,0),(w,h),(0,h))]
    child_matrix = None
    if is_group:
        co, ce = transform.get('chOff'), transform.get('chExt')
        if co is not None and ce is not None:
            ch = (num(co,'x'), num(co,'y'), num(ce,'cx'), num(ce,'cy'))
            if None not in ch and ch[2] != 0 and ch[3] != 0:
                child_matrix = mul(matrix, (w/ch[2],0,0,h/ch[3],-ch[0]*w/ch[2],-ch[1]*h/ch[3]))
    return polygon, child_matrix


def ph_match(root, placeholder):
    if root is None or placeholder is None:
        return None
    wanted_idx = placeholder.get('idx')
    wanted_type = placeholder.get('type', 'obj')
    choices = []
    # A nested placeholder needs its layout/master ancestor group matrix too.
    # Only top-level placeholders can safely donate a raw local transform.
    for node in objects(root):
        _, ph = nonvisual(node)
        if ph is None:
            continue
        if ph.get('idx') == wanted_idx and ph.get('type', 'obj') == wanted_type:
            choices.append(node)
    return choices[0] if len(choices) == 1 else None


def explicit_styles(node):
    def attributes(path):
        return [dict(n.attrib) for n in node.findall(path, NS)]
    return {
        'bodyPr': attributes('./p:txBody/a:bodyPr'),
        'listStyle': [ET.tostring(n, encoding='unicode') for n in node.findall('./p:txBody/a:lstStyle', NS)],
        'paragraph_properties': attributes('./p:txBody/a:p/a:pPr'),
        'run_properties': attributes('./p:txBody/a:p/a:r/a:rPr'),
        'default_run_properties': attributes('.//a:defRPr'),
        'font_faces': sorted({n.get('typeface') for n in node.iter() if n.get('typeface')}),
        'font_sizes_pt': sorted({int(n.get('sz'))/100 for n in node.iter()
                                 if local(n.tag) in {'rPr','defRPr','endParaRPr'} and (n.get('sz') or '').isdigit()}),
    }


def text_of(node):
    return '\n'.join(''.join(t.text or '' for t in p.iter() if local(t.tag) == 't')
                     for p in node.findall('.//a:p', NS))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--registry', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--report', type=Path, required=True)
    args = ap.parse_args()
    for dest in (args.out, args.report):
        if dest.exists():
            ap.error(f'refusing to overwrite {dest}')
    registry = json.loads(args.registry.read_text(encoding='utf-8'))
    rows, summary = [], []
    for source in registry['sources']:
        binary = Path(source['path_hint'])
        invpath = Path(source['raw_inventory_hint'])
        actual_hash = sha256(binary)
        if actual_hash != source['source_sha256']:
            raise ValueError(f"source hash mismatch: {source['source_id']}")
        inv = json.loads(invpath.read_text(encoding='utf-8'))
        if inv['source_sha256'] != actual_hash:
            raise ValueError(f"inventory hash mismatch: {source['source_id']}")
        size = inv['slide_size_emu']
        sw, sh = int(size['cx']), int(size['cy'])
        stats = Counter()
        with ZipFile(binary) as zf:
            pres = ET.fromstring(zf.read('ppt/presentation.xml'))
            actual_size = pres.find('p:sldSz', NS)
            if actual_size is None or (int(actual_size.get('cx')), int(actual_size.get('cy'))) != (sw, sh):
                raise ValueError(f"slide size mismatch: {source['source_id']}")
            presrels = {r['Id']: r for r in rels(zf, 'ppt/presentation.xml')}
            slides = []
            for number, sid in enumerate(pres.findall('p:sldIdLst/p:sldId', NS), 1):
                part = presrels[sid.get('{' + NS['r'] + '}id')]['resolved_part']
                slides.append((number, part))
            if [part for _,part in slides] != [s['part'] for s in inv['slides']]:
                raise ValueError(f"slide ordering mismatch: {source['source_id']}")
            if len(slides) != source['slide_count']:
                raise ValueError(f"slide count mismatch: {source['source_id']}")
            for number, part in slides:
                matrices.clear()
                root = ET.fromstring(zf.read(part))
                layoutpart = related(zf, part, 'slideLayout')
                layout = ET.fromstring(zf.read(layoutpart)) if layoutpart else None
                masterpart = related(zf, layoutpart, 'slideMaster') if layoutpart else None
                master = ET.fromstring(zf.read(masterpart)) if masterpart else None
                hidden = root.get('show') in {'0','false'}
                if hidden != (number in source.get('hidden_slides', [])):
                    raise ValueError(f"hidden status mismatch: {source['source_id']} slide {number}")
                invslide = inv['slides'][number-1]
                invnodes = dict(walk_inventory(invslide['objects']))
                for path, node, parent_path, parent_matrix, ancestors in recursive(objects(root)):
                    kind = local(node.tag)
                    props, ph = nonvisual(node)
                    observed = invnodes.get(path)
                    if observed is None or observed['kind'] != kind or observed.get('id') != props.get('id'):
                        raise ValueError(f"object inventory mismatch: {source['source_id']}:{part}:{path}")
                    transform_node = xfrm(node)
                    transform = raw_xfrm(transform_node)
                    provenance = {'kind':'slide', 'part':part, 'object_path':path} if transform is not None else None
                    nested_placeholder = transform is None and ph is not None and parent_path is not None
                    if transform is None and ph is not None and not nested_placeholder:
                        for fallback_root, fallback_part, level in ((layout,layoutpart,'layout'), (master,masterpart,'master')):
                            matched = ph_match(fallback_root, ph)
                            if matched is not None and xfrm(matched) is not None:
                                transform = raw_xfrm(xfrm(matched))
                                provenance = {'kind':level, 'part':fallback_part,
                                              'placeholder':nonvisual(matched)[1]}
                                break
                    polygon, child_matrix = geometry(transform, parent_matrix, kind == 'grpSp')
                    bounds = None
                    if polygon is not None:
                        xs,ys = zip(*polygon)
                        bounds = {'x':min(xs), 'y':min(ys), 'width':max(xs)-min(xs), 'height':max(ys)-min(ys)}
                    own_id = f"{source['source_id']}:{actual_hash[:12]}:{part}:{path}"
                    descendant_paths = [p for p in invnodes if p.startswith(path + '/')] if kind == 'grpSp' else []
                    row = {
                        'occurrence_id':own_id, 'source_id':source['source_id'], 'source_sha256':actual_hash,
                        'source_part':part, 'slide_number':number, 'hidden':hidden,
                        'object_path':path, 'parent_object_path':parent_path,
                        'ancestor_group_paths':list(ancestors), 'z_order_sibling':int(path.split('/')[-1]),
                        'shape_id':props.get('id'), 'shape_name':props.get('name'), 'object_kind':kind,
                        'placeholder':ph, 'text':observed.get('text'),
                        'preset_geometry':next((n.get('prst') for n in node.iter() if local(n.tag)=='prstGeom'), None),
                        'slide_size_emu':{'width':sw,'height':sh},
                        'slide_size_inches':{'width':sw/EMU_INCH,'height':sh/EMU_INCH},
                        'local_transform':raw_xfrm(transform_node),
                        'effective_transform':transform, 'transform_provenance':provenance,
                        'transform_attributes':transform.get('attributes') if transform else None,
                        'geometry_flags':({
                            'flip_horizontal':transform['attributes'].get('flipH') in {'1','true'},
                            'flip_vertical':transform['attributes'].get('flipV') in {'1','true'},
                            'rotation_degrees':(num(transform['attributes'],'rot') or 0)/60000,
                            'placeholder_transform_inherited':provenance is not None and provenance['kind'] != 'slide',
                        } if transform else None),
                        'polygon_emu':[[x,y] for x,y in polygon] if polygon else None,
                        'bounds_emu':bounds,
                        'polygon_inches':[[x/EMU_INCH,y/EMU_INCH] for x,y in polygon] if polygon else None,
                        'bounds_inches':{k:v/EMU_INCH for k,v in bounds.items()} if bounds else None,
                        'geometry_status':'resolved' if polygon else 'unresolved',
                        'geometry_note':None if polygon else (
                            'unresolved ancestor group transform' if parent_matrix is None else
                            'nested placeholder inheritance unsupported' if nested_placeholder else
                            'missing or incomplete transform'),
                        'explicit_text_style':explicit_styles(node) if kind != 'grpSp' else None,
                        'inherited_typography_status':'unresolved',
                        'descendant_occurrence_ids':[f"{source['source_id']}:{actual_hash[:12]}:{part}:{p}" for p in descendant_paths],
                    }
                    rows.append(row)
                    stats['objects'] += 1
                    stats['resolved' if polygon else 'unresolved'] += 1
                    stats['groups'] += kind == 'grpSp'
                    stats['inherited_transforms'] += provenance is not None and provenance['kind'] != 'slide'
                    if kind == 'grpSp' and child_matrix is None:
                        stats['unresolved_group_child_frames'] += 1
                    # The recursive traversal consumes this map to propagate composed group matrices.
                    matrices[path] = child_matrix
                if len(invnodes) != sum(1 for p in rows if p['source_id']==source['source_id'] and p['source_part']==part):
                    raise ValueError(f"object count mismatch: {source['source_id']}:{part}")
        if stats['objects'] != source['slide_object_node_count']:
            raise ValueError(f"registered object count mismatch: {source['source_id']}")
        summary.append({'source_id':source['source_id'], **dict(stats)})
    def write_jsonl(stream):
        for row in rows:
            stream.write(json.dumps(row, ensure_ascii=False, separators=(',',':')) + '\n')
    def write_report(stream):
        stream.write('# Slide geometry catalog\n\n')
        stream.write('Generated from source-hash-verified PPTX originals in `planning/source-registry.json`; source files were read only. One JSONL record represents one slide object node, including native group containers. Occurrence IDs match `catalog-occurrences.py`.\n\n')
        stream.write('| Source | Objects | Resolved frames | Unresolved frames | Groups | Inherited transforms | Groups without child map |\n|---|---:|---:|---:|---:|---:|---:|\n')
        for s in summary:
            stream.write(f"| {s['source_id']} | {s['objects']} | {s['resolved']} | {s['unresolved']} | {s['groups']} | {s.get('inherited_transforms',0)} | {s.get('unresolved_group_child_frames',0)} |\n")
        totals = Counter()
        for s in summary: totals.update({k:v for k,v in s.items() if k!='source_id'})
        stream.write(f"| **Total** | **{totals['objects']}** | **{totals['resolved']}** | **{totals['unresolved']}** | **{totals['groups']}** | **{totals['inherited_transforms']}** | **{totals['unresolved_group_child_frames']}** |\n\n")
        stream.write('## Meaning and limits\n\n')
        stream.write('- `polygon_emu` gives the four transformed corners of the object frame, ordered top-left, top-right, bottom-right, bottom-left in local coordinates. `bounds_emu` is its axis-aligned enclosing box. Inch values divide EMU by 914400. Rotated frames and fills are not treated as exact painted regions.\n')
        stream.write('- Group child frames compose `off/ext/chOff/chExt`, horizontal and vertical flips, and rotation through every ancestor. Descendant IDs and ancestry retain native grouping.\n')
        stream.write('- A missing slide transform can inherit a unique matching placeholder transform from its linked layout, then master. Ambiguous or incomplete transforms remain null. Provenance records the source of each effective transform.\n')
        stream.write('- Text and selected explicit text properties are observations. Typography inherited from layouts, masters, and themes remains unresolved. Shapes may also paint outside their rectangular frame; no render or clipping calculation was performed.\n')
    atomic_create(args.out, write_jsonl)
    try:
        atomic_create(args.report, write_report)
    except Exception:
        args.out.unlink()
        raise
    print(json.dumps({'rows':len(rows),'sources':len(summary),'summary':summary,'out':str(args.out),'report':str(args.report)}))


def walk_inventory(nodes, prefix=''):
    for index,node in enumerate(nodes,1):
        path = prefix+'/'+str(index) if prefix else str(index)
        yield path,node
        yield from walk_inventory(node.get('children',[]),path)


def recursive(nodes, prefix='', parent_matrix=I, ancestors=()):
    """Yield pre-order; caller deposits the preceding group's child matrix in matrices."""
    for index,node in enumerate(nodes,1):
        path = prefix+'/'+str(index) if prefix else str(index)
        yield path,node,(prefix or None),parent_matrix,ancestors
        if local(node.tag)=='grpSp':
            yield from recursive([n for n in node if local(n.tag) in KINDS],path,matrices.get(path),ancestors+(path,))


matrices = {}


if __name__ == '__main__':
    main()
