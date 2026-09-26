#!/usr/bin/env python3
"""Build and query a rebuildable SQLite/FTS5 catalog for assets and deck occurrences."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import sqlite3
import sys
import tempfile

MAX_LIMIT = 100


def file_hash(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def jsonl(path):
    with path.open(encoding='utf-8') as stream:
        for lineno, line in enumerate(stream, 1):
            if not line.strip():
                continue
            try:
                row = json.loads(line)
            except json.JSONDecodeError as exc:
                raise ValueError(f'{path}:{lineno}: invalid JSON: {exc}') from exc
            if not isinstance(row, dict):
                raise ValueError(f'{path}:{lineno}: JSONL record must be an object')
            yield row


def first_description_title(row):
    for desc in row.get('description_sources') or []:
        for line in (desc.get('text') or '').splitlines():
            line = line.strip()
            if line.startswith('# '):
                return line[2:].strip()
            m = re.match(r'\*\*Title:\*\*\s*(.+)', line, re.I)
            if m:
                return m.group(1).strip()
    for ref in row.get('source_refs') or []:
        name = Path(ref.get('path', '')).name
        if name:
            return name
    return row.get('id', '')


def asset_item(row):
    descriptions = row.get('description_sources') or []
    text_blocks = []
    for desc in descriptions:
        text = desc.get('text') or ''
        if text:
            text_blocks.append(text)
    context = []
    for ref in row.get('source_refs') or []:
        context.extend(str(ref.get(k, '')) for k in ('source', 'path', 'canonical_url') if ref.get(k))
    context.extend(str(x) for x in row.get('aliases') or [])
    for meta in row.get('remote_metadata') or []:
        context.extend(str(meta.get(k, '')) for k in ('fileName', 'directory', 'extension', 'assetType') if meta.get(k))
    context.extend(str(x) for x in row.get('categories') or [])
    body = '\n'.join(context + text_blocks)
    return {'id': row.get('id'), 'kind': row.get('kind') or 'asset', 'source_id': None,
            'category': row.get('category') or ','.join(row.get('categories') or []), 'title': first_description_title(row), 'body': body,
            'json': row}


def occurrence_item(row, aggregated_slide_text=None):
    text = row.get('text') or ''
    cells = '\n'.join(str(x) for x in row.get('table_cells') or [] if x)
    extra = []
    for key in ('source_shape_name', 'title_candidate', 'part_name', 'object_kind'):
        value = row.get(key)
        if value and value != text:
            extra.append(str(value))
    body_parts = extra + [text, cells]
    if row.get('occurrence_type') == 'slide' and aggregated_slide_text:
        body_parts.append('\n'.join(aggregated_slide_text))
    body = '\n'.join(part for part in body_parts if part).strip()
    title = row.get('source_shape_name') or row.get('title_candidate') or row.get('part_name') or row.get('object_kind') or row.get('occurrence_type')
    return {'id': row.get('occurrence_id'), 'kind': row.get('occurrence_type') or 'occurrence',
            'source_id': row.get('source_id'), 'category': row.get('native_category'),
            'title': title, 'body': body, 'json': row}



def component_item(row):
    parts = [row.get(k) for k in ('name', 'title', 'purpose', 'text', 'seed_id', 'canonical_family_id')]
    parts.extend(row.get('adaptation_constraints') or [])
    parts.extend(row.get('proposed_style_options') or [])
    parts.append(json.dumps(row.get('source_variant') or {}))
    return {'id': row.get('id'), 'kind': row.get('kind') or 'component',
            'source_id': row.get('source_id'), 'category': row.get('origin'),
            'title': row.get('name') or row.get('title') or row.get('id'),
            'body': '\n'.join(str(x) for x in parts if x), 'json': row}


def font_item(row):
    refs = row.get('source_refs') or []
    body = [row.get('font_name'), row.get('family_name'), row.get('style')]
    body.extend(row.get('aliases') or [])
    for ref in refs:
        body.extend(str(v) for v in ref.values() if v)
    body.append(row.get('readiness'))
    return {'id': row.get('id'), 'kind': 'font', 'source_id': None,
            'category': row.get('category') or 'font',
            'title': row.get('font_name') or row.get('family_name') or row.get('id'),
            'body': '\n'.join(str(x) for x in body if x), 'json': row}


def semantic_item(row):
    return {'id': row['id'], 'kind': row['kind'], 'source_id': None,
            'category': row.get('readiness'), 'title': row.get('name') or row['id'],
            'body': json.dumps(row, ensure_ascii=False), 'json': row}


def enrich_geometry(occurrence_rows, geometry_rows, geometry_path):
    expected = {row['occurrence_id']: row for row in occurrence_rows
                if row.get('native_category') == 'slides' and row.get('occurrence_type') in {'shape', 'group'}}
    geometry = {}
    for lineno, row in enumerate(geometry_rows, 1):
        ident = row.get('occurrence_id')
        if not ident:
            raise ValueError(f'{geometry_path}:{lineno}: geometry row missing occurrence_id')
        if ident in geometry:
            raise ValueError(f'{geometry_path}:{lineno}: duplicate geometry occurrence_id {ident}')
        source = expected.get(ident)
        if source is None:
            raise ValueError(f'{geometry_path}:{lineno}: geometry refers to non-slide or unknown occurrence {ident}')
        for field, source_field in (('source_sha256', 'source_sha256'), ('source_part', 'source_part'),
                                    ('object_path', 'object_path')):
            if row.get(field) != source.get(source_field):
                raise ValueError(f'{geometry_path}:{lineno}: {field} mismatch for {ident}')
        if row.get('source_id') != source.get('source_id') or row.get('slide_number') != source.get('slide_number'):
            raise ValueError(f'{geometry_path}:{lineno}: source/slide mismatch for {ident}')
        geometry[ident] = row
    missing = set(expected) - set(geometry)
    if missing:
        raise ValueError(f'{geometry_path}: missing geometry for {len(missing)} slide shape/group occurrences; first: {sorted(missing)[0]}')
    for ident, row in expected.items():
        row['geometry'] = geometry[ident]
    return len(geometry)


def validate_components(component_rows, occurrence_rows, components_path):
    occurrences = {row.get('occurrence_id'): row for row in occurrence_rows
                   if row.get('native_category') == 'slides' and row.get('occurrence_type') in {'shape', 'group'}}
    seen_ids = set()
    for lineno, row in enumerate(component_rows, 1):
        if row.get('kind') != 'component':
            raise ValueError(f'{components_path}:{lineno}: expected component kind')
        ident = row.get('id')
        if not isinstance(ident, str) or not ident:
            raise ValueError(f'{components_path}:{lineno}: component missing non-empty id')
        if ident in seen_ids:
            raise ValueError(f'{components_path}:{lineno}: duplicate component id {ident}')
        seen_ids.add(ident)
        member_ids = row.get('member_occurrence_ids')
        if not isinstance(member_ids, list) or not member_ids:
            raise ValueError(f'{components_path}:{lineno}: component {ident} has no member occurrences')
        if len(set(member_ids)) != len(member_ids):
            raise ValueError(f'{components_path}:{lineno}: component {ident} repeats a member occurrence')
        members = []
        for member_id in member_ids:
            member = occurrences.get(member_id)
            if member is None:
                raise ValueError(f'{components_path}:{lineno}: component {ident} references unknown/non-slide occurrence {member_id}')
            if any(member.get(field) != row.get(field) for field in
                   ('source_id', 'source_sha256', 'slide_number', 'source_part')):
                raise ValueError(f'{components_path}:{lineno}: component {ident} member crosses source/slide/hash boundary')
            members.append(member)
        member_set = set(member_ids)
        slot_names = set()
        for slot in row.get('slots') or []:
            name = slot.get('name')
            if not isinstance(name, str) or not name or name in slot_names:
                raise ValueError(f'{components_path}:{lineno}: missing or repeated slot name for {ident}')
            slot_names.add(name)
            slot_ref = slot.get('occurrence_id')
            if not slot_ref or slot_ref not in member_set:
                raise ValueError(f'{components_path}:{lineno}: component {ident} slot reference is not in member_occurrence_ids')


def load_dedup(path, occurrence_slides):
    data = json.loads(path.read_text(encoding='utf-8'))
    rows = []
    seen_refs = set()
    slide_refmap = {}
    for key, slide in occurrence_slides.items():
        slide_refmap[key] = slide
    def require_slide(ref, where):
        if ref not in slide_refmap:
            raise ValueError(f'{path}: {where} references unknown source slide {ref!r}')
        seen_refs.add(ref)
        return slide_refmap[ref]
    for slide in data.get('slides', []):
        ref = f"{slide.get('source_id')}:{int(slide.get('slide_number', 0)):03d}"
        occ = require_slide(ref, 'slides entry')
        title = occ.get('title_candidate') or f"{slide.get('source_id')} slide {slide.get('slide_number')}"
        body = '\n'.join([occ.get('text') or '', ' '.join(slide.get('flags') or []), slide.get('preview_status') or '']).strip()
        rows.append({'id': slide.get('id'), 'kind': 'dedup_slide', 'source_id': slide.get('source_id'),
                     'category': 'slide', 'title': title, 'body': body,
                     'json': {**slide, 'resolved_occurrence_id': occ['occurrence_id']}})
    for group in data.get('native_groups', []):
        members = group.get('members') or []
        resolved = [require_slide(ref, f"native group {group.get('id')}")['occurrence_id'] for ref in members]
        rep = group.get('representative')
        if rep:
            require_slide(rep, f"native group {group.get('id')} representative")
        rows.append({'id': group.get('id'), 'kind': 'native_group', 'source_id': None,
                     'category': group.get('status'), 'title': group.get('id'),
                     'body': ' '.join(members), 'json': {**group, 'resolved_occurrence_ids': resolved}})
    for index, review in enumerate(data.get('review_queue', []), 1):
        left = require_slide(review.get('left'), f'review queue row {index} left')
        right = require_slide(review.get('right'), f'review queue row {index} right')
        ident = f"review:{review.get('left')}:{review.get('right')}"
        rows.append({'id': ident, 'kind': 'review_queue', 'source_id': None,
                     'category': review.get('status'), 'title': f"{review.get('left')} ↔ {review.get('right')}",
                     'body': ' '.join([review.get('reason', ''), *review.get('warnings', [])]),
                     'json': {**review, 'resolved_left_occurrence_id': left['occurrence_id'],
                              'resolved_right_occurrence_id': right['occurrence_id']}})
    if len(seen_refs) != len(data.get('slides', [])):
        raise ValueError(f'{path}: failed to resolve every referenced slide')
    return rows, data


def load_decisions(path, occurrence_slides, dedup_data):
    data = json.loads(path.read_text(encoding='utf-8'))
    rows = []
    member_owner = {}
    group_ids = set()
    occurrence_hashes = {}
    for slide in occurrence_slides.values():
        source_id = slide.get('source_id')
        source_hash = slide.get('source_sha256')
        if source_id in occurrence_hashes and occurrence_hashes[source_id] != source_hash:
            raise ValueError(f'{path}: inconsistent occurrence source hashes for {source_id}')
        occurrence_hashes[source_id] = source_hash
    dedup_hashes = {source.get('source_id'): source.get('sha256')
                    for source in dedup_data.get('sources', [])}
    for source_id in set(occurrence_hashes) | set(dedup_hashes):
        if occurrence_hashes.get(source_id) != dedup_hashes.get(source_id):
            raise ValueError(f'{path}: occurrence/dedup source hash mismatch for {source_id}')
    supplied_hashes = data.get('source_hashes')
    if not isinstance(supplied_hashes, dict):
        raise ValueError(f'{path}: source_hashes must be an object mapping source_id to SHA-256')
    if supplied_hashes != occurrence_hashes or supplied_hashes != dedup_hashes:
        raise ValueError(f'{path}: source_hashes do not match occurrence and dedup source hashes')
    def resolve(ref, where):
        if ref not in occurrence_slides:
            raise ValueError(f'{path}: {where} references unknown source slide {ref!r}')
        return occurrence_slides[ref]
    for index, group in enumerate(data.get('groups', []), 1):
        group_id = group.get('id')
        if not isinstance(group_id, str) or not group_id:
            raise ValueError(f'{path}: group {index} missing non-empty id')
        if group_id in group_ids:
            raise ValueError(f'{path}: duplicate layout decision group id {group_id!r}')
        group_ids.add(group_id)
        members = group.get('members')
        if not isinstance(members, list) or not members:
            raise ValueError(f'{path}: group {group_id} must have non-empty members')
        resolved_members = [resolve(ref, f'group {group_id} member')['occurrence_id'] for ref in members]
        rep = group.get('representative')
        if not rep or rep not in members:
            raise ValueError(f'{path}: group {group_id} representative must be one of its members')
        resolved_rep = resolve(rep, f'group {group_id} representative')['occurrence_id']
        if len(set(members)) != len(members):
            raise ValueError(f'{path}: group {group_id} contains duplicate members')
        disposition = group.get('disposition')
        if disposition not in {'shared_layout_family', 'distinct'}:
            raise ValueError(f'{path}: group {group_id} has invalid disposition {disposition!r}')
        for ref in members:
            if ref in member_owner:
                raise ValueError(f'{path}: overlapping group membership for {ref}: {member_owner[ref]} and {group_id}')
            member_owner[ref] = group_id
        searchable = [group_id] + [str(group.get(k, '')) for k in ('disposition', 'reuse_scope', 'rationale')]
        searchable.extend(str(x) for x in group.get('variants') or [])
        searchable.append(json.dumps(group.get('reviewed_previews') or [], ensure_ascii=False))
        searchable.extend(members)
        rows.append({'id': group_id, 'kind': 'layout_family',
                     'source_id': rep.split(':', 1)[0] if rep else None,
                     'category': group.get('disposition'),
                     'title': group.get('rationale') or group_id,
                     'body': ' '.join(x for x in searchable if x),
                     'json': {**group, 'resolved_representative_occurrence_id': resolved_rep,
                              'resolved_member_occurrence_ids': resolved_members}})
    return rows, data


def temp_path_for(path):
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, name = tempfile.mkstemp(prefix=f'.{path.name}.', suffix='.tmp', dir=path.parent)
    os.close(fd)
    os.unlink(name)
    return Path(name)


def atomic_install(staged, destination):
    os.link(staged, destination)
    os.unlink(staged)


def build(args):
    inputs = {'occurrences': args.occurrences, 'assets': args.assets, 'dedup': args.dedup}
    for name in ('decisions', 'geometry', 'components', 'extras'):
        path = getattr(args, name)
        if path:
            inputs[name] = path
    for name, path in inputs.items():
        if not path.is_file():
            raise FileNotFoundError(f'{name} input not found: {path}')
    if args.out.exists():
        raise FileExistsError(f'refusing to overwrite existing database: {args.out}')
    occurrence_rows = list(jsonl(args.occurrences))
    asset_rows = list(jsonl(args.assets))
    geometry_rows = list(jsonl(args.geometry)) if args.geometry else []
    component_rows = list(jsonl(args.components)) if args.components else []
    extras_rows = list(jsonl(args.extras)) if args.extras else []
    style_ids = {r['id'] for r in extras_rows if r.get('kind') == 'style_profile'}
    family_ids = {r['id'] for r in extras_rows if r.get('kind') == 'component_family'}
    for row in component_rows + extras_rows:
        if not set(row.get('proposed_style_options') or []).issubset(style_ids):
            raise ValueError('Unknown component style reference: '+str(row.get('id')))
        if row.get('canonical_family_id') and row['canonical_family_id'] not in family_ids:
            raise ValueError('Unknown component family reference: '+str(row.get('id')))
    slides = {}
    for row in occurrence_rows:
        if row.get('occurrence_type') == 'slide':
            key = f"{row.get('source_id')}:{int(row.get('slide_number', 0)):03d}"
            if key in slides:
                raise ValueError(f'duplicate slide occurrence mapping {key}')
            slides[key] = row
    dedup_rows, dedup_data = load_dedup(args.dedup, slides)
    decision_rows = []
    decisions_data = None
    geometry_count = enrich_geometry(occurrence_rows, geometry_rows, args.geometry) if args.geometry else 0
    if args.components:
        validate_components(component_rows, occurrence_rows, args.components)
    component_items = [component_item(row) for row in component_rows]
    extras_items = []
    for row in extras_rows:
        if row.get('kind') == 'asset':
            extras_items.append(asset_item(row))
        elif row.get('kind') == 'font':
            extras_items.append(font_item(row))
        elif row.get('kind') == 'component_family':
            components_by_id = {r['id']: r for r in component_rows}
            if not row.get('instances'):
                raise ValueError('Component family has no instances: '+row['id'])
            for inst in row['instances']:
                component = components_by_id.get(inst['id'])
                if not component or component.get('canonical_family_id') != row['id']:
                    raise ValueError('Family member is missing or mismatched: '+inst['id'])
                if (component['source_id'], component['slide_number']) != (inst['source_id'], inst['slide_number']):
                    raise ValueError('Family member source identity mismatch: '+inst['id'])
                if row['source_hashes'].get(component['source_id']) != component['source_sha256']:
                    raise ValueError('Family source hash mismatch: '+row['id'])
            extras_items.append(semantic_item(row))
        elif row.get('kind') == 'dynamic_recipe':
            if row.get('readiness') != 'native_fixture_verified' or row.get('adaptation_approved') is not False:
                raise ValueError('Dynamic recipe must retain bounded fixture status: '+str(row.get('id')))
            spec_path = Path(row.get('spec_path', ''))
            proof_path = Path(row.get('proof_path', ''))
            if not spec_path.is_file() or file_hash(spec_path) != row.get('spec_sha256'):
                raise ValueError('Dynamic recipe spec hash mismatch: '+str(row.get('id')))
            if not proof_path.is_file() or file_hash(proof_path) != row.get('proof_sha256'):
                raise ValueError('Dynamic recipe proof hash mismatch: '+str(row.get('id')))
            spec = json.loads(spec_path.read_text())
            if row.get('slide_id') not in {x.get('id') for x in spec.get('slides', [])}:
                raise ValueError('Dynamic recipe slide is missing: '+str(row.get('id')))
            proof = json.loads(proof_path.read_text())
            if proof.get('status') != 'native_fixture_proof_not_general_catalog_approval':
                raise ValueError('Dynamic recipe proof status invalid: '+str(row.get('id')))
            if not any(x.get('path') == row['spec_path'] and x.get('sha256') == row['spec_sha256'] for x in proof.get('inputs', [])):
                raise ValueError('Dynamic recipe proof belongs to another spec: '+str(row.get('id')))
            if not any(x.get('id') == row['slide_id'] and x.get('reviewed') is True and x.get('accepted') is True for x in proof.get('visual_review', [])):
                raise ValueError('Dynamic recipe slide has no accepted visual review: '+str(row.get('id')))
            extras_items.append(semantic_item(row))
        elif row.get('kind') == 'style_profile':
            if not row.get('tokens') or row.get('readiness') != 'proposed_contract':
                raise ValueError('Malformed style contract: '+str(row.get('id')))
            extras_items.append(semantic_item(row))
        else:
            raise ValueError(f"{args.extras}: unsupported extra kind {row.get('kind')!r}")
    if args.decisions:
        if not args.decisions.is_file():
            raise FileNotFoundError(f'decisions input not found: {args.decisions}')
        decision_rows, decisions_data = load_decisions(args.decisions, slides, dedup_data)
    slide_text = {}
    for row in occurrence_rows:
        if row.get('slide_number') is not None and row.get('occurrence_type') in {'shape', 'group'}:
            key = f"{row.get('source_id')}:{int(row.get('slide_number', 0)):03d}"
            text = row.get('text') or ''
            if text:
                slide_text.setdefault(key, []).append(text)
            for cell in row.get('table_cells') or []:
                if cell:
                    slide_text.setdefault(key, []).append(str(cell))
    for dedup_row in dedup_rows:
        if dedup_row['kind'] == 'dedup_slide':
            source_slide = f"{dedup_row['source_id']}:{int(dedup_row['json']['slide_number']):03d}"
            dedup_row['body'] = '\n'.join(slide_text.get(source_slide, []) + [dedup_row['body']]).strip()
    items = [occurrence_item(x, slide_text.get(f"{x.get('source_id')}:{int(x.get('slide_number') or 0):03d}")) for x in occurrence_rows]
    items.extend(asset_item(x) for x in asset_rows)
    items.extend(dedup_rows)
    items.extend(decision_rows)
    items.extend(component_items)
    items.extend(extras_items)
    seen = set()
    for row in items:
        if not isinstance(row['id'], str) or not row['id']:
            raise ValueError(f"item missing non-empty ID (kind={row['kind']!r})")
        if row['id'] in seen:
            raise ValueError(f"duplicate catalog item ID: {row['id']}")
        seen.add(row['id'])
    out = args.out
    staged_db = temp_path_for(out)
    staged_report = temp_path_for(args.report)
    installed_db = False
    try:
        conn = sqlite3.connect(staged_db)
        try:
            conn.execute('PRAGMA journal_mode=DELETE')
            conn.execute('PRAGMA synchronous=FULL')
            conn.execute('''CREATE TABLE items (
                id TEXT PRIMARY KEY, kind TEXT NOT NULL, source_id TEXT,
                category TEXT, title TEXT, body TEXT NOT NULL, json TEXT NOT NULL
            )''')
            conn.execute('CREATE INDEX items_kind_idx ON items(kind)')
            conn.execute('CREATE INDEX items_source_idx ON items(source_id)')
            conn.execute('''CREATE TABLE ingestion (
                input_name TEXT PRIMARY KEY, input_path TEXT NOT NULL,
                sha256 TEXT NOT NULL, record_count INTEGER NOT NULL,
                detail_json TEXT NOT NULL
            )''')
            try:
                conn.execute("CREATE VIRTUAL TABLE items_fts USING fts5(item_id UNINDEXED, title, body, tokenize='unicode61')")
            except sqlite3.OperationalError as exc:
                raise RuntimeError(f'SQLite FTS5 is unavailable: {exc}') from exc
            conn.executemany('INSERT INTO items VALUES (?, ?, ?, ?, ?, ?, ?)', [
                (x['id'], x['kind'], x['source_id'], x['category'], x['title'], x['body'],
                 json.dumps(x['json'], ensure_ascii=False, separators=(',', ':'))) for x in items])
            conn.executemany('INSERT INTO items_fts(item_id,title,body) VALUES(?,?,?)', [
                (x['id'], x['title'] or '', x['body']) for x in items])
            input_counts = {
                'occurrences': len(occurrence_rows), 'assets': len(asset_rows),
                'dedup': len(dedup_rows), 'decisions': len(decision_rows),
                'geometry': len(geometry_rows), 'components': len(component_rows),
                'extras': len(extras_rows),
            }
            input_details = {
                'occurrences': {'source_counts': source_counts(occurrence_rows)},
                'assets': {'kinds': count_values(asset_rows, 'kind'), 'categories': count_values(asset_rows, 'category')},
                'dedup': {'native_groups': len(dedup_data.get('native_groups', [])),
                          'slides': len(dedup_data.get('slides', [])), 'review_queue': len(dedup_data.get('review_queue', []))},
                'decisions': {'groups': len(decisions_data.get('groups', [])) if decisions_data else 0},
                'geometry': {'enriched_slide_shape_group_occurrences': geometry_count},
                'components': {'indexed_component_records': len(component_items)},
                'extras': {'kinds': count_values(extras_rows, 'kind')},
            }
            ingestion = [(name, path, input_counts[name], input_details[name]) for name, path in inputs.items()]
            conn.executemany('INSERT INTO ingestion VALUES (?,?,?,?,?)', [
                (name, str(path), file_hash(path), count, json.dumps(detail, ensure_ascii=False))
                for name, path, count, detail in ingestion])
            conn.execute('PRAGMA user_version=1')
            conn.commit()
            counts = dict(conn.execute('SELECT kind, COUNT(*) FROM items GROUP BY kind'))
            total = conn.execute('SELECT COUNT(*) FROM items').fetchone()[0]
            fts_total = conn.execute('SELECT COUNT(*) FROM items_fts').fetchone()[0]
            if total != len(items) or fts_total != total:
                raise ValueError(f'index count mismatch: items={total}, fts={fts_total}, input={len(items)}')
        finally:
            conn.close()
        with staged_report.open('w', encoding='utf-8', newline='\n') as stream:
            write_report(stream, args, inputs, input_counts, occurrence_rows, asset_rows, dedup_rows, decision_rows, component_rows, extras_rows, geometry_count, counts, len(items))
            stream.flush()
            os.fsync(stream.fileno())
        if args.report.exists():
            raise FileExistsError(f'refusing to overwrite existing report: {args.report}')
        atomic_install(staged_db, out)
        installed_db = True
        atomic_install(staged_report, args.report)
        print(json.dumps({'database': str(out), 'report': str(args.report), 'items': len(items), 'kinds': counts}))
    except Exception:
        for path in (staged_db, staged_report):
            try:
                path.unlink()
            except FileNotFoundError:
                pass
        if installed_db:
            try:
                out.unlink()
            except FileNotFoundError:
                pass
        raise


def count_values(rows, key):
    from collections import Counter
    return dict(sorted(Counter(str(row.get(key) or '') for row in rows).items()))


def source_counts(rows):
    from collections import Counter
    return {k: dict(v) for k, v in sorted(_source_counts(rows).items())}


def _source_counts(rows):
    from collections import defaultdict, Counter
    result = defaultdict(Counter)
    for row in rows:
        result[row.get('source_id', '')][row.get('occurrence_type', '')] += 1
    return result


def write_report(stream, args, inputs, input_counts, occurrence_rows, asset_rows, dedup_rows, decision_rows, component_rows, extras_rows, geometry_count, counts, total):
    stream.write('# Catalog index build report\n\n')
    stream.write('The rebuildable SQLite catalog was built from occurrence, asset, and dedup manifests plus any supplied decision, geometry, component, and extras manifests. Full records are preserved in the `items.json` column; FTS5 indexes item titles and searchable text. Slide occurrence search bodies include text aggregated from all shape and group descendants on that slide. Geometry is attached to matching slide shape/group occurrence records and adds no item rows. Components and optional layout-family decisions are separate indexed records; decisions do not replace native dedup review statuses. Source readiness/review fields are preserved without inferring approval.\n\n')
    stream.write('| Input | Path | SHA-256 | Records ingested |\n|---|---|---|---:|\n')
    for name, path in inputs.items():
        n = input_counts[name]
        stream.write(f'| {name} | `{path}` | `{file_hash(path)}` | {n} |\n')
    stream.write(f'\nTotal indexed items: **{total}**.\n\n')
    stream.write('| Kind | Items |\n|---|---:|\n')
    for kind, n in sorted(counts.items()):
        stream.write(f'| {kind} | {n} |\n')
    stream.write('\nDedup slide references were checked against source slide occurrences; component member and slot references were checked against slide occurrence IDs, source hashes, and slide boundaries. Geometry IDs were checked for unique exact coverage of slide shape/group occurrences. Every indexed ID is unique.\n')


def connect_db(path):
    conn = sqlite3.connect(path.resolve().as_uri() + '?mode=ro', uri=True)
    conn.row_factory = sqlite3.Row
    return conn


def search(args):
    if not 1 <= args.limit <= MAX_LIMIT:
        raise ValueError(f'--limit must be between 1 and {MAX_LIMIT}')
    try:
        conn = connect_db(args.db)
    except sqlite3.Error as exc:
        raise ValueError(f'cannot open database {args.db}: {exc}') from exc
    try:
        sql = '''SELECT i.id,i.kind,i.source_id,i.category,i.title,i.json,
                        snippet(items_fts,2,'[',']',' … ',12) AS snippet,
                        bm25(items_fts) AS rank
                 FROM items_fts JOIN items i ON i.id=items_fts.item_id
                 WHERE items_fts MATCH ?'''
        params = [args.query]
        if args.kind:
            sql += ' AND i.kind=?'
            params.append(args.kind)
        sql += ' ORDER BY rank, i.id LIMIT ?'
        params.append(args.limit)
        try:
            rows = conn.execute(sql, params).fetchall()
        except sqlite3.OperationalError as exc:
            raise ValueError(f'invalid FTS query: {exc}') from exc
        result = []
        for row in rows:
            record = json.loads(row['json'])
            result.append({'id': row['id'], 'kind': row['kind'], 'source_id': row['source_id'],
                           'category': row['category'], 'title': row['title'], 'snippet': row['snippet'],
                           'status': status_fields(record), 'rank': row['rank']})
        print(json.dumps({'query': args.query, 'kind': args.kind, 'results': result}, ensure_ascii=False, indent=2))
    finally:
        conn.close()


def status_fields(record):
    keys = ('review_status', 'readiness', 'technical_readiness', 'semantic_role', 'status', 'disposition', 'reuse_scope', 'visual_review', 'preview_status', 'variant_status')
    return {key: record[key] for key in keys if key in record}


def inspect(args):
    try:
        conn = connect_db(args.db)
    except sqlite3.Error as exc:
        raise ValueError(f'cannot open database {args.db}: {exc}') from exc
    try:
        row = conn.execute('SELECT id,kind,source_id,category,title,body,json FROM items WHERE id=?', (args.id,)).fetchone()
        if row is None:
            raise ValueError(f'item not found: {args.id}')
        out = dict(row)
        out['json'] = json.loads(out['json'])
        print(json.dumps(out, ensure_ascii=False, indent=2))
    finally:
        conn.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='command', required=True)
    b = sub.add_parser('build', help='build a new SQLite catalog database')
    b.add_argument('--occurrences', type=Path, required=True)
    b.add_argument('--assets', type=Path, required=True)
    b.add_argument('--dedup', type=Path, required=True)
    b.add_argument('--decisions', type=Path)
    b.add_argument('--geometry', type=Path)
    b.add_argument('--components', type=Path)
    b.add_argument('--extras', type=Path)
    b.add_argument('--out', type=Path, required=True)
    b.add_argument('--report', type=Path, default=Path('library/catalog-index-report.md'))
    b.set_defaults(func=build)
    s = sub.add_parser('search', help='search the FTS5 catalog')
    s.add_argument('--db', type=Path, required=True)
    s.add_argument('--query', required=True)
    s.add_argument('--kind')
    s.add_argument('--limit', type=int, default=10)
    s.set_defaults(func=search)
    i = sub.add_parser('inspect', help='display one item by exact ID')
    i.add_argument('--db', type=Path, required=True)
    i.add_argument('--id', required=True)
    i.set_defaults(func=inspect)
    args = parser.parse_args()
    try:
        args.func(args)
    except (OSError, ValueError, RuntimeError, sqlite3.Error) as exc:
        print(f'error: {exc}', file=sys.stderr)
        return 2
    return 0

if __name__ == '__main__':
    sys.exit(main())
