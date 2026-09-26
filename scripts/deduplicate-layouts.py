#!/usr/bin/env python3
"""Conservative native-content deduplication and structural layout candidate queue.

No visual certification or inherited-style resolution is implied. The candidate
queue is deliberately separate from confirmed native-content equivalence.
"""
import argparse
from collections import Counter, defaultdict
import hashlib
import json
from pathlib import Path
import posixpath
import os
import tempfile
import xml.etree.ElementTree as ET
from zipfile import ZipFile

ALGORITHM = 'native-layout-dedup-4'
A = 'http://schemas.openxmlformats.org/drawingml/2006/main'
P = 'http://schemas.openxmlformats.org/presentationml/2006/main'
R = 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'
NS = {'p': P, 'a': A, 'r': R}
KINDS = {'sp', 'pic', 'graphicFrame', 'cxnSp', 'grpSp', 'contentPart'}
REL_SKIP = {'notesSlide', 'slide'}  # Preserved as occurrence metadata, not display dependencies.

def digest(value):
    return hashlib.sha256(value if isinstance(value, bytes) else json.dumps(value, sort_keys=True, separators=(',', ':')).encode()).hexdigest()

def tag(e):
    return e.tag.rsplit('}', 1)[-1]

def relations(z, part):
    path = posixpath.join(posixpath.dirname(part), '_rels', posixpath.basename(part) + '.rels')
    if path not in z.namelist():
        return {}
    result = {}
    for n in ET.fromstring(z.read(path)):
        d = dict(n.attrib)
        if d.get('TargetMode') != 'External':
            d['resolved'] = posixpath.normpath(posixpath.join(posixpath.dirname(part), d['Target'])).lstrip('/')
            if d['resolved'] not in z.namelist():
                raise ValueError(f'Missing relationship target: {part} -> {d["resolved"]}')
        result[d['Id']] = d
    return result

def dependency_digest(z, start, cache):
    """Conservative byte closure; master/layout cycles traversed once.

    Resource XML retains incidental IDs: this can miss duplicates, never justify
    stripping unknown rendering-affecting metadata to increase a match count.
    """
    if start in cache:
        return cache[start]
    pending, visited, rows = [start], set(), []
    while pending:
        part = pending.pop()
        if part in visited:
            continue
        visited.add(part)
        edges = []
        for rel in relations(z, part).values():
            role = rel['Type'].rsplit('/', 1)[-1]
            if role in REL_SKIP:
                continue
            if rel.get('TargetMode') == 'External':
                edges.append([role, 'external', rel['Target']])
            else:
                dest = rel['resolved']
                edges.append([role, digest(z.read(dest))])
                pending.append(dest)
        rows.append([digest(z.read(part)), sorted(edges)])
    result = digest(sorted(rows))
    cache[start] = result
    return result

def canonical_slide(z, part, cache, presentation_style):
    root = ET.fromstring(z.read(part))
    rels = relations(z, part)
    ids = {n.get('id'): str(i) for i, n in enumerate(root.iter()) if tag(n) == 'cNvPr'}
    flags = set()
    for n in root.iter():
        if tag(n) in {'AlternateContent', 'oleObj', 'graphicData', 'contentPart', 'videoFile', 'audioFile'}:
            flags.add('complex_or_opaque_content')
        if tag(n) == 'fld':
            flags.add('dynamic_field')
        if tag(n) in {'timing', 'transition'}:
            flags.add('animation_or_transition')
    def walk(n):
        attrs = []
        for k, v in n.attrib.items():
            if tag(n) == 'cNvPr' and k == 'name':
                continue  # Nonvisual occurrence metadata remains in original.
            if tag(n) == 'cNvPr' and k == 'id' and not flags:
                v = ids[v]
            elif tag(n) in {'stCxn', 'endCxn'} and k == 'id' and not flags:
                v = ids.get(v, v)
            elif k.startswith('{' + R + '}'):
                rel = rels.get(v)
                if not rel:
                    raise ValueError(f'Unresolved relationship ID {part}:{v}')
                role = rel['Type'].rsplit('/', 1)[-1]
                if rel.get('TargetMode') == 'External':
                    v = role + ':' + rel['Target']
                elif role in REL_SKIP:
                    v = role + ':' + rel['resolved']
                else:
                    v = role + ':' + dependency_digest(z, rel['resolved'], cache)
            attrs.append([k, v])
        # Whitespace in XML formatting is irrelevant; text-run spaces are not.
        t = n.text if n.tag == '{' + A + '}t' else (n.text or '').strip()
        return [n.tag, sorted(attrs), t or '', [walk(c) for c in n]]
    deps = []
    for rel in rels.values():
        role = rel['Type'].rsplit('/', 1)[-1]
        if role not in REL_SKIP:
            deps.append([role, rel.get('Target') if rel.get('TargetMode') == 'External' else dependency_digest(z, rel['resolved'], cache)])
    return digest([walk(root), sorted(deps), presentation_style]), sorted(flags)

def numeric(node, key, fallback=0):
    return float(node.get(key, fallback)) if node is not None else fallback

def object_features(root, size):
    """Comparison features only. Group-local geometry carries its ancestry.

    Unresolved geometry is marked. No flattened slide-space geometry is claimed.
    """
    features = []
    def walk(tree, ancestry=()):
        for child in tree:
            kind = tag(child)
            if kind not in KINDS:
                continue
            tr = child.find('a:xfrm', NS)
            if tr is None:
                tr = child.find('p:xfrm', NS)  # Native graphicFrame (tables/charts).
            if tr is None:
                tr = child.find('p:spPr/a:xfrm', NS)
            if tr is None:
                tr = child.find('p:grpSpPr/a:xfrm', NS)
            xy = tr.find('a:off', NS) if tr is not None else None
            wh = tr.find('a:ext', NS) if tr is not None else None
            box = [numeric(xy, 'x') / size[0], numeric(xy, 'y') / size[1], numeric(wh, 'cx') / size[0], numeric(wh, 'cy') / size[1]] if xy is not None and wh is not None else None
            attrs = dict(tr.attrib) if tr is not None else {}
            preset = child.find('p:spPr/a:prstGeom', NS)
            ph = next((n for n in child.iter() if tag(n) == 'ph'), None)
            styles = sorted({ET.tostring(n, encoding='unicode') for n in child.iter() if tag(n) in {'rPr', 'defRPr', 'bodyPr'}})
            table = child.find('.//a:tbl', NS)
            table_structure = None
            if table is not None:
                table_structure = {
                    'grid': [dict(n.attrib) for n in table.findall('a:tblGrid/a:gridCol', NS)],
                    'rows': [{'height': row.get('h'), 'cells': [dict(c.attrib) for c in row.findall('a:tc', NS)]}
                             for row in table.findall('a:tr', NS)]}
            graphic = child.find('a:graphic/a:graphicData', NS)
            custom = child.find('p:spPr/a:custGeom', NS)
            crop = child.find('p:blipFill/a:srcRect', NS)
            feature = {'kind': kind, 'box': box,
                       'table_structure': table_structure,
                       'graphic_type': graphic.get('uri') if graphic is not None else None,
                       'custom_geometry_hash': digest(ET.tostring(custom)) if custom is not None else None,
                       'crop': dict(crop.attrib) if crop is not None else None, 'transform_attributes': attrs,
                       'preset': preset.get('prst') if preset is not None else None,
                       'placeholder': dict(ph.attrib) if ph is not None else None,
                       'ancestry': list(ancestry), 'explicit_text_style_hash': digest(styles)}
            features.append(feature)
            if kind == 'grpSp':
                descriptor = digest(ET.tostring(tr)) if tr is not None else 'unresolved'
                walk(child, ancestry + (descriptor,))
    tree = root.find('p:cSld/p:spTree', NS)
    if tree is not None:
        walk(tree)
    return features

def geometry_key(features):
    return digest([{'kind': f['kind'], 'box': [round(v, 4) for v in f['box']] if f['box'] else None,
                    'transform': f['transform_attributes'], 'preset': f['preset'],
                    'placeholder': f['placeholder'], 'ancestry': f['ancestry'],
                    'table_structure': f['table_structure'], 'graphic_type': f['graphic_type'],
                    'custom_geometry_hash': f['custom_geometry_hash'], 'crop': f['crop']} for f in features])

def similarity(left, right):
    # Restrict to same object-kind cardinality; broader variants require review.
    if Counter(f['kind'] for f in left) != Counter(f['kind'] for f in right):
        return None
    if not left or any(f['box'] is None for f in left + right):
        return None
    if any(f['ancestry'] for f in left + right):
        return None  # No false slide-space assumptions for nested group transforms.
    def key(f):
        return (f['kind'], f['box'][1], f['box'][0], f['box'][3], f['box'][2])
    a, b = sorted(left, key=key), sorted(right, key=key)
    for l, r in zip(a, b):
        if any(l[k] != r[k] for k in ('table_structure', 'graphic_type', 'custom_geometry_hash', 'crop', 'preset', 'transform_attributes', 'placeholder')):
            return None
    errors = [max(abs(x-y) for x, y in zip(l['box'], r['box'])) for l, r in zip(a, b)]
    if max(errors) > .035 or sum(errors) / len(errors) > .01:
        return None
    return round(1 - sum(errors) / len(errors), 6)

def run(registry, project_root):
    sources, slides = [], []
    for source in registry['sources']:
        path = project_root / source['path_hint']
        if digest(path.read_bytes()) != source['source_sha256']:
            raise ValueError(f'Source hash mismatch: {source["source_id"]}')
        inv = json.loads((project_root / source['raw_inventory_hint']).read_text())
        if inv['source_sha256'] != source['source_sha256']:
            raise ValueError('Inventory identity mismatch')
        with ZipFile(path) as z:
            pres = ET.fromstring(z.read('ppt/presentation.xml'))
            size_node = pres.find('p:sldSz', NS)
            size = [int(size_node.get('cx')), int(size_node.get('cy'))]
            style = pres.find('p:defaultTextStyle', NS)
            pres_props = {'size': size, 'default_text_style': ET.tostring(style, encoding='unicode') if style is not None else None,
                          'attributes': dict(pres.attrib)}
            cache = {}
            pres_props['global_resources'] = sorted([
                [rel['Type'], rel['Target'] if rel.get('TargetMode') == 'External' else dependency_digest(z, rel['resolved'], cache)]
                for rel in relations(z, 'ppt/presentation.xml').values()
                if rel['Type'].rsplit('/', 1)[-1] not in {'slide', 'slideMaster', 'notesMaster'}])
            for s in inv['slides']:
                exact, flags = canonical_slide(z, s['part'], cache, pres_props)
                feats = object_features(ET.fromstring(z.read(s['part'])), size)
                if any(f['box'] is None for f in feats):
                    flags.append('inherited_geometry_unresolved')
                if any(f['ancestry'] for f in feats):
                    flags.append('group_geometry_local')
                slides.append({'id': f'{source["source_id"]}:{s["slide_number"]:03d}',
                               'source_id': source['source_id'], 'source_sha256': source['source_sha256'],
                               'slide_number': s['slide_number'], 'part': s['part'],
                               'hidden': s['attributes'].get('show') in ('0', 'false'),
                               'object_count': len(feats), 'native_content_fingerprint': exact,
                               'geometry_fingerprint': geometry_key(feats), 'flags': sorted(set(flags)),
                               'preview_status': 'not_rendered_by_this_run', '_features': feats})
        sources.append({'source_id': source['source_id'], 'sha256': source['source_sha256'], 'slides': len(inv['slides'])})
    exact_groups = defaultdict(list)
    for s in slides:
        exact_groups[s['native_content_fingerprint']].append(s)
    groups = []
    for fingerprint, members in sorted(exact_groups.items()):
        # Source order favors stock designs; preserve user-selected exemplars as aliases.
        canonical = members[0]['id']
        gid = 'native-' + fingerprint[:20]
        groups.append({'id': gid, 'representative': canonical, 'members': [s['id'] for s in members],
                       'status': 'native_content_equivalent_unrendered' if len(members) > 1 else 'singleton',
                       'visual_review': 'pending', 'semantic_layout_id': None})
        for s in members:
            s['native_group_id'] = gid
            s['representative'] = canonical
    reps = [s for s in slides if s['id'] == s['representative']]
    geometry_buckets = defaultdict(list)
    for slide in reps:
        geometry_buckets[slide['geometry_fingerprint']].append(slide)
    queue = []
    for members in geometry_buckets.values():
        for r in members[1:]:
            l = members[0]
            queue.append({'left': l['id'], 'right': r['id'], 'geometry_similarity': 1.0,
                          'reason': 'same_rounded_geometry_candidate', 'status': 'review_required',
                          'auto_merge': False, 'warnings': sorted(set(l['flags'] + r['flags'] +
                              ['effective_styles_and_content_roles_not_resolved', 'geometry_quantized_for_candidate_retrieval']))})
    # One direct representative-to-member comparison per exact geometry bucket;
    # this deduplicates review requests, never merges semantic layout contracts.
    bucket_reps = [members[0] for members in geometry_buckets.values()]
    near = []
    for i, l in enumerate(bucket_reps):
        for r in bucket_reps[i+1:]:
            score = similarity(l['_features'], r['_features'])
            if score is not None:
                near.append({'left': l['id'], 'right': r['id'], 'geometry_similarity': score,
                             'reason': 'near_geometry_candidate', 'status': 'review_required',
                             'auto_merge': False, 'warnings': sorted(set(l['flags'] + r['flags'] +
                                 ['effective_styles_and_content_roles_not_resolved', 'z_order_requires_review']))})
    near.sort(key=lambda p: (-p['geometry_similarity'], p['left'], p['right']))
    degree = Counter()
    for pair in near:
        if degree[pair['left']] >= 3 or degree[pair['right']] >= 3:
            continue
        queue.append(pair)
        degree[pair['left']] += 1
        degree[pair['right']] += 1
    queue.sort(key=lambda p: (-p['geometry_similarity'], p['left'], p['right']))
    for s in slides:
        del s['_features']
    return {'schema_version': 1, 'algorithm': ALGORITHM,
            'limitations': ['Not visual certification', 'Raw dependency bytes cause conservative missed matches',
                           'Group transforms and inherited styles are not fully resolved',
                           'Shared geometry proposes review; it does not confirm shared layouts',
                           'Counts are native-content groups, not measured unique semantic layouts'],
            'summary': {'source_slides': len(slides), 'native_content_groups': len(groups),
                        'duplicate_native_content_occurrences': len(slides)-len(groups),
                        'multi_member_native_groups': sum(len(g['members']) > 1 for g in groups),
                        'geometry_candidate_buckets': len(geometry_buckets),
                        'near_candidate_pairs_before_queue_limit': len(near), 'review_queue_pairs': len(queue),
                        'unique_semantic_layout_count': None},
            'sources': sources, 'native_groups': groups, 'review_queue': queue, 'slides': slides}

def report(data):
    s = data['summary']
    lines = ['# Layout deduplication — first structural pass', '',
             f'Algorithm: `{ALGORITHM}`. Generated from the four hash-verified registered sources.', '',
             '| Measure | Count |', '| --- | ---: |',
             *[f'| {k.replace("_", " ")} | {v if v is not None else "Pending review"} |' for k,v in s.items()], '',
             'Native-content groups use normalized slide XML and a conservative resource closure,',
             'including inherited layout/master/theme bytes and presentation defaults. They are',
             'not pixel comparisons. The geometry queue abstracts literal text and image identity;',
             'every proposed layout merge remains pending style, role and visual review.', '',
             '## Native-content duplicate groups', '']
    for g in data['native_groups']:
        if len(g['members']) > 1:
            lines.append('- ' + ', '.join(g['members']))
    if not any(len(g['members']) > 1 for g in data['native_groups']):
        lines.append('No exact native-content groups found by this conservative algorithm.')
    lines += ['', '## First candidate pairs', '', '| Left | Right | Geometry score | Basis |', '| --- | --- | ---: | --- |']
    for p in data['review_queue'][:30]:
        lines.append(f'| {p["left"]} | {p["right"]} | {p["geometry_similarity"]:.4f} | {p["reason"]} |')
    lines += ['', '## Limitations and next gate', '',
              '- Identical native layout names or masters never establish a match.',
              '- Incidental dependency XML differences can hide real duplicates; no tolerance is silently used to confirm them.',
              '- Nested geometry remains local; near-geometry comparison skips nested groups.',
              '- Candidate edges do not form automatic transitive clusters.',
              '- All source occurrences and their own content/fit responsibilities remain intact.',
              '- Render representatives and ambiguous pairs; record explicit decisions before sharing semantic layout contracts.',
              '- The 250–300 layout estimate remains unverified. No reduction target was imposed.', '']
    return '\n'.join(lines)

def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--registry', type=Path, required=True)
    p.add_argument('--project-root', type=Path, default=Path.cwd())
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--report', type=Path, required=True)
    a = p.parse_args()
    if a.out.exists() or a.report.exists():
        p.error('Output/report exists; choose new paths to preserve prior evidence')
    data = run(json.loads(a.registry.read_text()), a.project_root)
    a.out.parent.mkdir(parents=True, exist_ok=True)
    a.report.parent.mkdir(parents=True, exist_ok=True)
    staged, installed = [], []
    try:
        for destination, content in [(a.out, json.dumps(data, indent=2) + '\n'), (a.report, report(data))]:
            fd, name = tempfile.mkstemp(prefix='.' + destination.name + '.', dir=destination.parent)
            staged.append(Path(name))
            with os.fdopen(fd, 'w') as f:
                f.write(content); f.flush(); os.fsync(f.fileno())
        for source, destination in zip(staged, (a.out, a.report)):
            os.link(source, destination)
            installed.append(destination)
    except Exception:
        for path in installed:
            path.unlink(missing_ok=True)
        raise
    finally:
        for path in staged:
            path.unlink(missing_ok=True)
    print(json.dumps(data['summary']))

if __name__ == '__main__':
    main()
