#!/usr/bin/env python3
"""Emit source-bound structural occurrences from registered PPTX inventories."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import posixpath
import tempfile
import xml.etree.ElementTree as ET
from zipfile import ZipFile

NS = {
    'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
    'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships',
}
OBJECT_KINDS = {'sp', 'pic', 'graphicFrame', 'cxnSp', 'grpSp', 'contentPart'}
STYLE_FLAGS = [
    'inherited-layout-master-theme-styles-unresolved',
    'group-transforms-not-flattened',
    'semantic-role-unclassified',
]


def sha256(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()


def atomic_create(path, writer):
    """Atomically create path without overwriting an existing destination."""
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp_name = tempfile.mkstemp(prefix=f'.{path.name}.', suffix='.tmp', dir=path.parent)
    try:
        with os.fdopen(fd, 'w', encoding='utf-8', newline='\n') as stream:
            writer(stream)
            stream.flush()
            os.fsync(stream.fileno())
        # Hard-link installation is atomic and fails if the destination exists.
        os.link(temp_name, path)
        os.unlink(temp_name)
    except Exception:
        try:
            os.unlink(temp_name)
        except FileNotFoundError:
            pass
        raise


def part_relationships(zf, part):
    relpart = posixpath.join(posixpath.dirname(part), '_rels', posixpath.basename(part) + '.rels')
    if relpart not in zf.namelist():
        return []
    root = ET.fromstring(zf.read(relpart))
    out = []
    for rel in root:
        target = rel.attrib.get('Target', '')
        if rel.attrib.get('TargetMode') == 'External':
            out.append(dict(rel.attrib))
        else:
            resolved = posixpath.normpath(posixpath.join(posixpath.dirname(part), target)).lstrip('/')
            out.append({**rel.attrib, 'resolved_part': resolved})
    return out


def actual_slides(zf):
    prespart = 'ppt/presentation.xml'
    root = ET.fromstring(zf.read(prespart))
    relmap = {r['Id']: r for r in part_relationships(zf, prespart)}
    result = []
    for number, node in enumerate(root.findall('p:sldIdLst/p:sldId', NS), 1):
        rel = relmap[node.get('{' + NS['r'] + '}id')]
        part = rel['resolved_part']
        sroot = ET.fromstring(zf.read(part))
        result.append({'slide_number': number, 'part': part,
                       'hidden': sroot.attrib.get('show') in {'0', 'false'}})
    return result


def flatten(items, parent_path=''):
    for index, item in enumerate(items, 1):
        path = f'{parent_path}/{index}' if parent_path else str(index)
        yield path, item
        yield from flatten(item.get('children', []), path)


def object_counts(parts):
    counts = {'object_nodes': 0, 'groups': 0, 'top_level_objects': 0}
    for part in parts:
        counts['object_nodes'] += sum(1 for _, _ in flatten(part.get('objects', [])))
        counts['groups'] += sum(1 for _, obj in flatten(part.get('objects', [])) if obj['kind'] == 'grpSp')
        counts['top_level_objects'] += len(part.get('objects', []))
    return counts


def duplicate_ids(part):
    ids = [obj.get('id') for _, obj in flatten(part.get('objects', [])) if obj.get('id') is not None]
    seen, dup = set(), set()
    for ident in ids:
        if ident in seen:
            dup.add(str(ident))
        seen.add(ident)
    return dup


def record(source, file_hash, part, occ_type, fields, object_path=None):
    hprefix = file_hash[:12]
    suffix = f'{part}:{object_path}' if object_path is not None else part
    return {
        'occurrence_id': f"{source['source_id']}:{hprefix}:{suffix}",
        'source_id': source['source_id'], 'source_sha256': file_hash,
        'source_role': source.get('role'), 'source_part': part,
        'occurrence_type': occ_type, **fields,
    }


def emit_part(records, source, file_hash, item, category, slide_info=None):
    part = item['part']
    hidden = slide_info['hidden'] if slide_info else None
    common = {'native_category': category, 'part_name': item.get('name', ''),
              'review_status': 'structural', 'semantic_role': 'unclassified',
              'unresolved_style_flags': list(STYLE_FLAGS)}
    if slide_info:
        common.update(slide_number=slide_info['slide_number'], hidden=hidden,
                      title_candidate=item.get('title_candidate', ''))
    records.append(record(source, file_hash, part, category[:-1] if category.endswith('s') else category,
                          {**common, 'object_count': len(list(flatten(item.get('objects', []))))}))
    dups = duplicate_ids(item)
    def walk(items, parent_path=''):
        for index, obj in enumerate(items, 1):
            path = f'{parent_path}/{index}' if parent_path else str(index)
            is_group = obj['kind'] == 'grpSp'
            fields = {
                'native_category': category, 'slide_number': slide_info['slide_number'] if slide_info else None,
                'hidden': hidden, 'object_path': path, 'parent_object_path': parent_path or None,
                'source_shape_id': obj.get('id'), 'source_shape_name': obj.get('name'),
                'source_shape_id_duplicate_within_part': str(obj.get('id')) in dups if obj.get('id') is not None else False,
                'object_kind': obj['kind'], 'parent_group_id': obj.get('parent_group_id'),
                'placeholder': obj.get('placeholder'), 'local_transform': obj.get('transform'),
                'text': obj.get('text'), 'observed_characters': obj.get('observed_characters'),
                'explicit_style': {'font_faces': obj.get('explicit_font_faces', []),
                                   'font_sizes_pt': obj.get('explicit_font_sizes_pt', []),
                                   'text_body_properties': obj.get('text_body_properties', [])},
                'table_cells': obj.get('table_cells', []), 'graphic_types': obj.get('graphic_types', []),
                'ole_program_ids': obj.get('ole_program_ids', []),
                'relationship_ids': obj.get('relationship_ids', []),
                'native_group': is_group, 'review_status': 'structural',
                'semantic_role': 'unclassified', 'unresolved_style_flags': list(STYLE_FLAGS),
            }
            records.append(record(source, file_hash, part, 'group' if is_group else 'shape', fields, path))
            if is_group:
                walk(obj.get('children', []), path)
    walk(item.get('objects', []))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--registry', required=True, type=Path)
    ap.add_argument('--out', required=True, type=Path)
    ap.add_argument('--report', required=True, type=Path)
    args = ap.parse_args()
    registry = json.loads(args.registry.read_text(encoding='utf-8'))
    records = []
    summary = []
    for source in registry['sources']:
        binary = Path(source['path_hint'])
        raw_path = Path(source['raw_inventory_hint'])
        if not binary.is_file() or not raw_path.is_file():
            raise FileNotFoundError(f"Missing source or raw inventory for {source['source_id']}: {binary}, {raw_path}")
        file_hash = sha256(binary)
        if file_hash != source['source_sha256']:
            raise ValueError(f"Source hash mismatch for {source['source_id']}: registry={source['source_sha256']} actual={file_hash}")
        inv = json.loads(raw_path.read_text(encoding='utf-8'))
        if inv.get('source_sha256') != file_hash:
            raise ValueError(f"Raw inventory hash mismatch for {source['source_id']}: inventory={inv.get('source_sha256')} actual={file_hash}")
        with ZipFile(binary) as package:
            slides = actual_slides(package)
        if len(slides) != len(inv['slides']):
            raise ValueError(f"Slide count mismatch for {source['source_id']}: OOXML={len(slides)} inventory={len(inv['slides'])}")
        hidden_nums = [s['slide_number'] for s in slides if s['hidden']]
        if hidden_nums != source.get('hidden_slides', []):
            raise ValueError(f"Hidden-slide registry mismatch for {source['source_id']}: registry={source.get('hidden_slides', [])} actual={hidden_nums}")
        slide_by_part = {s['part']: s for s in slides}
        slide_object_counts = object_counts(inv['slides'])
        reg_count = source.get('slide_object_node_count')
        if reg_count is not None and slide_object_counts['object_nodes'] != reg_count:
            raise ValueError(f"Slide object count mismatch for {source['source_id']}: registry={reg_count} inventory={slide_object_counts['object_nodes']}")
        if source.get('slide_count') != len(inv['slides']) or source.get('native_layout_count') != len(inv['native_layouts']) or source.get('master_count') != len(inv['masters']):
            raise ValueError(f"Part count mismatch for {source['source_id']}")
        for item in inv['slides']:
            info = slide_by_part.get(item['part'])
            if not info:
                raise ValueError(f"Presentation slide part not found for {source['source_id']}: {item['part']}")
            emit_part(records, source, file_hash, item, 'slides', info)
        for item in inv['native_layouts']:
            emit_part(records, source, file_hash, item, 'native_layouts')
        for item in inv['masters']:
            emit_part(records, source, file_hash, item, 'masters')
        summary.append({'source_id': source['source_id'], 'sha256': file_hash,
                        'slides': len(inv['slides']), 'hidden': hidden_nums,
                        'slide_objects': slide_object_counts,
                        'native_layouts': len(inv['native_layouts']),
                        'layout_objects': object_counts(inv['native_layouts']),
                        'masters': len(inv['masters']), 'master_objects': object_counts(inv['masters']),
                        'occurrences': sum(1 + len(list(flatten(x.get('objects', [])))) for x in inv['slides'] + inv['native_layouts'] + inv['masters'])})
    def write_jsonl(stream):
        for item in records:
            stream.write(json.dumps(item, ensure_ascii=False, separators=(',', ':')) + '\n')
    def write_report(stream):
        stream.write('# Source occurrence catalog\n\n')
        stream.write('Generated from the registered source binaries and raw structural inventories. Every binary SHA-256 was checked against both the registry and its raw inventory before export. Occurrences preserve source identity and local hierarchy; they do not assert semantic deduplication or design approval.\n\n')
        stream.write('| Source | Slides (hidden) | Slide object nodes / groups | Native layouts (object nodes / groups) | Masters (object nodes / groups) | Emitted occurrence records | SHA-256 prefix |\n')
        stream.write('|---|---:|---:|---:|---:|---:|---|\n')
        for x in summary:
            stream.write(f"| {x['source_id']} | {x['slides']} ({len(x['hidden'])}) | {x['slide_objects']['object_nodes']} / {x['slide_objects']['groups']} | {x['native_layouts']} ({x['layout_objects']['object_nodes']} / {x['layout_objects']['groups']}) | {x['masters']} ({x['master_objects']['object_nodes']} / {x['master_objects']['groups']}) | {x['occurrences']} | `{x['sha256'][:12]}` |\n")
        stream.write(f"\nTotal emitted occurrence records: **{len(records)}**. Output JSONL: `samples/catalog/occurrences.jsonl`.\n\n")
        stream.write('## Counting and identity definitions\n\n')
        stream.write('- A slide occurrence is one record for each slide part referenced by the presentation, including hidden slides. Hidden status comes from the slide part root `p:sld/@show`, after resolving presentation relationships; absent `show` is visible.\n')
        stream.write('- A shape/group occurrence is one record per inventoried object node in each part. The node set is `sp`, `pic`, `graphicFrame`, `cxnSp`, `grpSp`, and `contentPart`, recursively including descendants inside groups. Group containers receive their own record.\n')
        stream.write('- Native layout and master occurrence records are emitted once per inventoried native layout/master part, along with one record per object node in that part. Counts are package-part counts and are not deduplicated across inheritance.\n')
        stream.write('- Object paths are one-based sibling positions through the inventoried hierarchy, independent of source shape IDs. Source IDs and duplicate-within-part flags are retained as evidence; IDs are not assumed globally unique or semantically stable.\n')
        stream.write('- `observed_characters` and extracted text are source observations, not text-capacity claims. Explicit font face/size/body properties are retained as found in the object. Inherited style resolution and group-transform flattening remain unresolved. Roles are `unclassified`; readiness is `structural`.\n')
    atomic_create(args.out, write_jsonl)
    atomic_create(args.report, write_report)
    print(json.dumps({'sources': len(summary), 'occurrences': len(records), 'output': str(args.out), 'report': str(args.report)}))

if __name__ == '__main__':
    main()
