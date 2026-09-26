#!/usr/bin/env python3
"""Catalog native groups and explicitly selected compositions without authoring edits.

Geometry is consumed from the hash-matched geometry catalog. Structural families
are candidates, not approved semantic components or content-capacity guarantees.
"""
import argparse
from collections import Counter, defaultdict
import hashlib
import json
import os
from pathlib import Path
import tempfile
import xml.etree.ElementTree as ET
from zipfile import ZipFile

A = 'http://schemas.openxmlformats.org/drawingml/2006/main'
P = 'http://schemas.openxmlformats.org/presentationml/2006/main'
R = 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'
NS = {'a': A, 'p': P}
KINDS = {'sp', 'pic', 'grpSp', 'graphicFrame', 'cxnSp', 'contentPart'}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tag(n):
    return n.tag.rsplit('}', 1)[-1]


def walk(tree, prefix=''):
    index = 0
    for node in tree:
        if tag(node) not in KINDS:
            continue
        index += 1
        path = f'{prefix}/{index}' if prefix else str(index)
        yield path, node
        if tag(node) == 'grpSp':
            yield from walk(node, path)


def union(boxes):
    if not boxes or any(b is None for b in boxes):
        return None
    left = min(b['x'] for b in boxes); top = min(b['y'] for b in boxes)
    right = max(b['x'] + b['width'] for b in boxes)
    bottom = max(b['y'] + b['height'] for b in boxes)
    return {'x': left, 'y': top, 'width': right-left, 'height': bottom-top}


def style_signature(node):
    # Candidate comparison only: source hash scopes unresolved inheritance and
    # relationship IDs. Text values/positions differ; run/style structure remains.
    def canon(n):
        if tag(n) in KINDS or tag(n) == 'xfrm':
            return None
        attrs = {k: v for k, v in n.attrib.items() if not
                 (tag(n) == 'cNvPr' and k in {'id', 'name'})}
        return [n.tag, sorted(attrs.items()), '' if tag(n) == 't' else (n.text or '').strip(),
                [v for c in n if (v := canon(c)) is not None]]
    return sha(json.dumps([canon(c) for c in node], sort_keys=True).encode())


def paragraphs(node):
    result = []
    for p in node.findall('.//a:p', NS):
        result.append({'text': ''.join(t.text or '' for t in p.findall('.//a:t', NS)),
                       'xml': ET.tostring(p, encoding='unicode')})
    return result


def color_observations(node):
    """Preserve explicit color expressions, including transforms, without inheritance guesses."""
    colors = []
    color_tags = {'srgbClr', 'schemeClr', 'prstClr', 'sysClr', 'scrgbClr', 'hslClr'}
    def visit(n, path):
        if tag(n) in color_tags:
            colors.append({'xml_path': path, 'type': tag(n), 'attributes': dict(n.attrib),
                           'transforms': [{'type': tag(c), 'attributes': dict(c.attrib)} for c in n]})
        for index, child in enumerate(n):
            visit(child, f'{path}/{tag(child)}[{index}]')
    visit(node, tag(node))
    return colors


def read_unique_jsonl(path, id_key):
    rows = {}
    for lineno, line in enumerate(path.read_text(encoding='utf-8').splitlines(), 1):
        if not line.strip():
            continue
        row = json.loads(line)
        ident = row.get(id_key)
        if not isinstance(ident, str) or not ident:
            raise ValueError(f'{path}:{lineno}: missing non-empty {id_key}')
        if ident in rows:
            raise ValueError(f'{path}:{lineno}: duplicate {id_key}: {ident}')
        rows[ident] = row
    return rows


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('registry', 'geometry', 'occurrences', 'seeds', 'out', 'report'):
        parser.add_argument('--'+name, type=Path, required=True)
    args = parser.parse_args()
    if args.out.exists() or args.report.exists():
        parser.error('Output/report exists; choose new paths')
    registry = json.loads(args.registry.read_text())
    sources = {}
    for source in registry['sources']:
        source_id = source.get('source_id')
        if not source_id or source_id in sources:
            raise ValueError(f'Duplicate or missing registry source_id: {source_id!r}')
        sources[source_id] = source
    source_hashes = {s['source_id']: s['source_sha256'] for s in sources.values()}
    seeds = json.loads(args.seeds.read_text())
    if seeds.get('source_hashes') != source_hashes:
        raise ValueError('Seed source hashes do not match registry')
    geometry = read_unique_jsonl(args.geometry, 'occurrence_id')
    occurrences = read_unique_jsonl(args.occurrences, 'occurrence_id')
    for label, rows in (('geometry', geometry), ('occurrence', occurrences)):
        for ident, row in rows.items():
            source_id = row.get('source_id')
            if source_id not in sources or row.get('source_sha256') != source_hashes.get(source_id):
                raise ValueError(f'Stale or unknown {label} source hash: {ident}')
    objects = {}; slides = {}; geometries = {}; native_groups = []; expected_slide_object_ids = set()
    for source in sources.values():
        source_id = source['source_id']; source_hash = source['source_sha256']
        if sha(Path(source['path_hint']).read_bytes()) != source_hash:
            raise ValueError('Source binary changed: '+source_id)
        inv = json.loads(Path(source['raw_inventory_hint']).read_text())
        if inv['source_sha256'] != source_hash:
            raise ValueError('Raw inventory changed: '+source_id)
        with ZipFile(source['path_hint']) as package:
            for slide in inv['slides']:
                key = (source_id, slide['slide_number']); slides[key] = slide
                tree = ET.fromstring(package.read(slide['part'])).find('p:cSld/p:spTree', NS)
                if tree is None:
                    continue
                for path, node in walk(tree):
                    ident = f"{source_id}:{source_hash[:12]}:{slide['part']}:{path}"
                    objkey = (*key, path); objects[objkey] = node
                    expected_slide_object_ids.add(ident)
                    g = geometry.get(ident); occ = occurrences.get(ident)
                    if not g or not occ or g.get('source_sha256') != source_hash or occ.get('source_sha256') != source_hash:
                        raise ValueError('Missing or stale geometry/occurrence: '+ident)
                    for label, row in (('geometry', g), ('occurrence', occ)):
                        if (row.get('source_id') != source_id or row.get('source_part') != slide['part']
                                or row.get('slide_number') != slide['slide_number']
                                or row.get('object_path') != path):
                            raise ValueError(f'Mismatched {label} identity fields: {ident}')
                    if occ.get('native_category') != 'slides' or occ.get('occurrence_type') not in {'shape', 'group'}:
                        raise ValueError('Unexpected occurrence kind for slide object: '+ident)
                    geometries[objkey] = g
                    if tag(node) == 'grpSp':
                        native_groups.append((key, path))
    if set(geometry) != expected_slide_object_ids:
        missing = expected_slide_object_ids - set(geometry)
        extra = set(geometry) - expected_slide_object_ids
        raise ValueError(f'Geometry coverage mismatch: missing={len(missing)} extra={len(extra)}')
    if not expected_slide_object_ids.issubset(occurrences):
        missing = expected_slide_object_ids - set(occurrences)
        raise ValueError(f'Occurrence coverage mismatch: missing={len(missing)}')
    records = []
    def compose(key, paths, ident, name, origin, slots=None, seed=None, instance=None):
        if not isinstance(paths, list) or not paths or any(not isinstance(p, str) or not p for p in paths):
            raise ValueError('Composition paths must be a non-empty list of object paths: '+ident)
        if slots is not None and not isinstance(slots, list):
            raise ValueError('Composition slots must be a list: '+ident)
        if len(paths) != len(set(paths)):
            raise ValueError('Repeated composition path: '+ident)
        if any(p != q and q.startswith(p+'/') for p in paths for q in paths):
            raise ValueError('Composition includes both group and its descendant: '+ident)
        selected = {k[2]: node for k, node in objects.items() if k[:2] == key and
                    any(k[2] == path or k[2].startswith(path+'/') for path in paths)}
        for path in paths:
            if path not in selected:
                raise ValueError(f'Unknown selected object {key} {path}')
        gs = {p: geometries[(*key, p)] for p in selected}
        leaf_paths = [p for p, n in selected.items() if tag(n) != 'grpSp']
        # Include group frames AND descendants so an intentionally large frame
        # and content extending beyond it are both retained as source evidence.
        bounds = union([g['bounds_emu'] for g in gs.values()])
        features = []
        for path in leaf_paths:
            g = gs[path]; b = g['bounds_emu']
            relative = ([round(b['x']-bounds['x'], 2), round(b['y']-bounds['y'], 2),
                         round(b['width'], 2), round(b['height'], 2)] if bounds and b else None)
            polygon = [[round(x-bounds['x'], 2), round(y-bounds['y'], 2)]
                       for x, y in g['polygon_emu']] if bounds and g.get('polygon_emu') else None
            features.append([tag(selected[path]), relative, polygon, style_signature(selected[path])])
        # Do not group unresolved geometry candidates together simply because
        # they all have null bounds. Relation IDs remain scoped to source package.
        signature_data = [sources[key[0]]['source_sha256'], features,
                          None if bounds else ident]
        fingerprint = sha(json.dumps(signature_data, sort_keys=True).encode())
        slot_rows = []
        for slot in slots or []:
            if not isinstance(slot, dict) or slot.get('kind') not in {'text', 'image'}:
                raise ValueError('Unsupported or malformed slot kind: '+ident)
            path = slot.get('object_path')
            if not isinstance(path, str) or not path:
                raise ValueError('Slot has missing object_path: '+ident)
            if path not in selected:
                raise ValueError('Slot is outside composition: '+ident+':'+path)
            node = selected[path]; g = gs[path]; occ = occurrences[g['occurrence_id']]
            if slot.get('occurrence_id') is not None and slot['occurrence_id'] != g['occurrence_id']:
                raise ValueError('Slot occurrence_id is stale: '+ident+':'+path)
            if slot['kind'] == 'text' and (tag(node) != 'sp' or not node.findall('./p:txBody/a:p', NS)):
                raise ValueError('Text slot must reference a native text shape with text paragraphs: '+ident+':'+path)
            if slot['kind'] == 'text' and not (occ.get('text') or '').strip():
                if slot.get('allow_empty') is not True or not slot.get('empty_role_rationale'):
                    raise ValueError('Empty text slot requires an explicit role rationale: '+ident+':'+path)
            if slot['kind'] == 'image' and tag(node) != 'pic':
                raise ValueError('Image slot is not a native picture: '+ident+':'+path)
            slot_rows.append({**slot, 'occurrence_id': g['occurrence_id'],
                              'bounds_emu': g['bounds_emu'], 'geometry_status': g['geometry_status'],
                              'observed_text': occ.get('text'),
                              'observed_characters': occ.get('observed_characters'),
                              'explicit_style': occ.get('explicit_style'),
                              'paragraphs': paragraphs(node) if slot['kind'] == 'text' else [],
                              'effective_typography': 'unresolved', 'capacity_status': 'not_measured'})
        text = '\n'.join(occurrences[gs[p]['occurrence_id']].get('text') or '' for p in leaf_paths).strip()
        observed_colors = []
        for path in leaf_paths:
            colors = color_observations(selected[path])
            if colors:
                observed_colors.append({'object_path': path, 'colors': colors})
        return {'id': ident, 'kind': 'component', 'name': name, 'title': name,
                'source_id': key[0], 'source_sha256': sources[key[0]]['source_sha256'],
                'slide_number': key[1], 'source_part': slides[key]['part'],
                'origin': origin, 'seed_id': seed['id'] if seed else None,
                'selected_object_paths': paths,
                'member_occurrence_ids': [gs[p]['occurrence_id'] for p in selected],
                'leaf_object_count': len(leaf_paths), 'text': text,
                'bounds_emu': bounds, 'structural_fingerprint': fingerprint,
                'candidate_family_id': 'component-structure:'+fingerprint[:24],
                'slots': slot_rows, 'geometry_status': 'resolved' if bounds else 'unresolved',
                'readiness': 'slot_candidates' if slots else 'discovered',
                'design_preference': 'unrated', 'content_approval': 'unreviewed',
                'supported_operations': ['inspect', 'preserve_source'],
                'purpose': seed.get('purpose') if seed else None,
                'canonical_family_id': seed.get('canonical_family_id') if seed else None,
                'proposed_style_options': seed.get('proposed_style_options', []) if seed else [],
                'style_application_status': seed.get('style_application_status', 'unmapped') if seed else 'unmapped',
                'source_variant': {k: v for k, v in (instance or {}).items()
                                   if k not in {'slots', 'object_paths', 'source_id', 'slide_number'}},
                'observed_colors': observed_colors,
                'color_resolution': 'explicit_expressions_only_inheritance_unresolved',
                'adaptation_constraints': seed.get('adaptation_constraints', []) if seed else [],
                'limitations': ['No adaptation or rendered fit approval',
                                'Inherited typography remains unresolved',
                                'Bounds enclose source frames, not visible ink or safe text zones',
                                'Native groups can contain unrelated elements',
                                'Structural matches are scoped to source package and preserve source order']}
    for key, path in native_groups:
        g = geometries[(*key, path)]
        records.append(compose(key, [path], 'component-occurrence:'+g['occurrence_id'],
                               'Native group '+path, 'native_group'))
    seed_ids = set()
    for seed in seeds['seeds']:
        if not isinstance(seed.get('id'), str) or not seed['id'] or seed['id'] in seed_ids:
            raise ValueError(f'Duplicate or missing seed ID: {seed.get("id")!r}')
        seed_ids.add(seed['id'])
        for index, inst in enumerate(seed.get('instances', []), 1):
            source_id = inst.get('source_id')
            if source_id not in sources:
                raise ValueError(f'Unknown seed source: {source_id!r}')
            if inst.get('source_sha256') is not None and inst['source_sha256'] != source_hashes[source_id]:
                raise ValueError(f'Stale seed instance source hash: {seed["id"]}')
            key = (source_id, inst.get('slide_number'))
            if key not in slides:
                raise ValueError(f'Unknown seed instance slide: {seed["id"]} {key}')
            slots = inst.get('slots', [])
            records.append(compose(key, inst.get('object_paths', []), inst.get('instance_id', seed['id']+f':instance-{index:03d}'),
                                   seed.get('name') or seed['id'], 'curated_composition_candidate', slots, seed, inst))
    if len({r['id'] for r in records}) != len(records):
        raise ValueError('Duplicate component IDs')
    families = defaultdict(list)
    for record in records:
        if record['origin'] == 'native_group':
            families[record['structural_fingerprint']].append(record['id'])
    counts = {'native_group_occurrences': len(native_groups),
              'native_group_structural_buckets': len(families),
              'multi_member_structural_buckets': sum(len(m)>1 for m in families.values()),
              'curated_component_seeds': len(seeds['seeds']),
              'curated_composition_instances': len(records)-len(native_groups),
              'canonical_semantic_families': len({r['canonical_family_id'] for r in records if r['canonical_family_id']}),
              'slot_candidates': sum(len(r['slots']) for r in records),
              'unresolved_bounds': sum(r['bounds_emu'] is None for r in records),
              'approved_reusable_components': 0}
    report = '# Component candidate catalog\n\n| Measure | Count |\n|---|---:|\n'
    report += ''.join(f'| {k.replace("_", " ")} | {v} |\n' for k, v in counts.items())
    report += '\nNative group occurrences and curated compositions are separate records and may overlap.\n'
    report += 'Structural buckets are source-scoped comparison candidates; no semantic merge is automatic.\n'
    report += 'Every selected source object and slot resolves to a hash-matched occurrence and geometry record.\n'
    report += 'Text slots retain observed text, explicit styles and paragraph XML. Observations are not capacity limits.\n'
    report += 'Explicit color expressions retain XML locations and transforms; theme inheritance is not resolved. Semantic style options are contracts only, with no automatic object-role bindings or recoloring.\n'
    report += '\nOutput: `'+str(args.out)+'`. Input hashes:\n\n'
    for name in ('registry','geometry','occurrences','seeds'):
        path = getattr(args,name); report += f'- `{path}`: `{sha(path.read_bytes())}`\n'
    report += '\n## Seed families\n\n'+''.join(f"- **{s['name']}** — {s['purpose']}\n" for s in seeds['seeds'])
    report += '\n## Next gate\n\nResolve effective typography, measure rendered text, verify overlaps and dependencies, then prove content edits before promoting a component. All current seeds support inspection/source preservation only.\n'
    outputs = [(args.out, ''.join(json.dumps(r,ensure_ascii=False)+'\n' for r in records)),(args.report,report)]
    staged=[];installed=[]
    try:
        for target,content in outputs:
            target.parent.mkdir(parents=True,exist_ok=True)
            fd,name=tempfile.mkstemp(prefix='.'+target.name+'.',dir=target.parent); staged.append(Path(name))
            with os.fdopen(fd,'w',encoding='utf-8') as stream:
                stream.write(content);stream.flush();os.fsync(stream.fileno())
        for path,(target,_) in zip(staged,outputs):
            os.link(path,target);installed.append(target)
    except Exception:
        for target in installed:target.unlink(missing_ok=True)
        raise
    finally:
        for path in staged:path.unlink(missing_ok=True)
    print(json.dumps(counts))

if __name__ == '__main__':
    main()
