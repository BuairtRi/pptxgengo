#!/usr/bin/env python3
"""Measure phrases in PowerPoint, solve accents, and publish a new review bundle.

Requires macOS PowerPoint. Only top-level, unrotated text and native pictures are
supported. Failed/ambiguous measurements remain explicit manual-review items.
Never rewrites source text or moves its frame to accommodate an accent.
"""
import argparse
import copy
import hashlib
import json
import posixpath
from pathlib import Path
import shutil
import subprocess
import zipfile


def read(p):
    return json.loads(Path(p).read_text())


def write(p, value):
    Path(p).write_text(json.dumps(value, indent=2) + '\n')


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def artifact(p):
    return {'path': str(Path(p).resolve()), 'sha256': sha(p)}


def run(args):
    return subprocess.check_output(list(map(str, args)), text=True, timeout=150)


def attr(n, k):
    return next((a['value'] for a in n.get('attributes', []) if a['name'] == k), '')


def child(n, name):
    return next((c for c in n.get('children', []) if c.get('name', '').split(':')[-1] == name), None)


def walk(n):
    yield n
    for c in n.get('children', []):
        yield from walk(c)


def object_id(n):
    for tag in ('nvSpPr', 'nvPicPr', 'nvGraphicFramePr', 'nvCxnSpPr', 'nvGrpSpPr'):
        nv = child(n, tag)
        if nv is not None:
            return attr(child(nv, 'cNvPr') or {}, 'id')
    return ''


def objects(scene):
    tree = child(child(scene, 'cSld'), 'spTree')
    return [n for n in tree.get('children', []) if object_id(n)]


def one_object(scene, oid):
    matches = [(i + 1, n) for i, n in enumerate(objects(scene)) if object_id(n) == str(oid)]
    if len(matches) != 1:
        raise ValueError(f'Object {oid} absent, nested, or ambiguous')
    return matches[0]


def restored(doc):
    s = copy.deepcopy(doc['scene'])
    for b in doc['bindings']:
        n = s
        for i in b['node_path']:
            n = n['children'][i]
        if b.get('attribute'):
            a = next(a for a in n['attributes'] if a['name'] == b['attribute'])
            if a['value'] != '__BINDING:' + b['binding_id']:
                raise ValueError('Binding sentinel mismatch')
            a['value'] = b['value']
        else:
            if n['text'] != '__BINDING:' + b['binding_id']:
                raise ValueError('Text binding sentinel mismatch')
            n['text'] = b['value']
    return s


def geometry(n):
    xf = child(child(n, 'spPr') or {}, 'xfrm')
    if xf is None:
        raise ValueError('Explicit transform required')
    off, ext = child(xf, 'off'), child(xf, 'ext')
    return xf, [int(attr(a, k)) / 12700 for a, k in ((off, 'x'), (off, 'y'), (ext, 'cx'), (ext, 'cy'))]


def normalized(text):
    return ' '.join(text.split())


def full_text(n):
    return normalized(' '.join(''.join(x.get('text', '') if x.get('name', '') == '' else ' ' if x.get('name') == 'a:br' else '' for x in walk(p)) for p in walk(n) if p.get('name') == 'a:p'))


def build_deck(scene_bin, project, path, deck):
    path.unlink(missing_ok=True)
    Path(str(path) + '.build.json').unlink(missing_ok=True)
    run([scene_bin, 'build', '--project', project, '--out', path, '--slides', ','.join(str(p['source_slide_number']) for p in deck['pages']), '--freeze-slide-numbers'])
    report = Path(str(path) + '.build.json')
    deck.update(sha256=sha(path), native_build_report={'path': report.name, 'sha256': sha(report)})


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root', type=Path, default=Path(__file__).resolve().parents[1])
    ap.add_argument('--bundle', type=Path, required=True)
    ap.add_argument('--intents', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--name', required=True)
    ap.add_argument('--template-bin', required=True)
    ap.add_argument('--anchor-bin', required=True)
    ap.add_argument('--scene-bin', required=True)
    ap.add_argument('--id', action='append')
    args = ap.parse_args()
    root, src, out = args.root.resolve(), args.bundle.resolve(), args.out.resolve()
    if out.is_relative_to(src) or src.is_relative_to(out):
        ap.error('output and input bundle must not contain one another')
    if out.exists() or not args.name or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.name):
        ap.error('out must be new; name must be a unique safe filename prefix')
    raw = read(args.intents)
    intents = {r['template_id']: r for r in raw['intents'] if not args.id or r['template_id'] in args.id}
    record = read(src / 'review-bundle.json')
    chosen = {t['template_id'] for t in record['templates']} & intents.keys()
    if not chosen:
        ap.error('No accent intents match this bundle')
    for t in record['templates']:
        for a in t['inputs']:
            p = (src / a['path']).resolve()
            if not p.is_relative_to(src) or sha(p) != a['sha256']:
                raise ValueError('Frozen bundle input changed')
    for d in record['decks']:
        p = (src / d['path']).resolve()
        if not p.is_relative_to(src) or sha(p) != d['sha256']:
            raise ValueError('Frozen native deck changed')
    native_dir = root / 'samples/visual-wave3'
    if not native_dir.is_dir():
        raise ValueError('Established PowerPoint working folder is missing')
    # Copy provenance and editable native projects; publish the manifest last.
    out.mkdir(parents=True)
    for name in ('inputs', 'projects', 'applications'):
        if (src / name).exists():
            shutil.copytree(src / name, out / name)
    evidence_dir = out / 'accent-evidence'
    evidence_dir.mkdir()
    shutil.copy2(args.intents, evidence_dir / 'intents.json')
    record['templates'] = [t for t in record['templates'] if t['template_id'] in chosen]
    record['decks'] = [d for d in record['decks'] if any(p['template_id'] in chosen for p in d['pages'])]
    application = {'schema': 'pptxgengo.accent-application.v1', 'status': 'in_progress', 'input_manifest': artifact(src / 'review-bundle.json'), 'intents': artifact(evidence_dir / 'intents.json'), 'templates': []}
    write(out / 'accent-application.json', application)
    try:
        for deck in record['decks']:
            source = deck['pages'][0]['source'].split(':')[0]
            project = out / 'projects' / source
            # Rebuild every original page before selecting the accent subset.
            # Deterministic ZIP bytes must equal the pinned native source deck.
            verified = evidence_dir / ('verified-' + deck['path'])
            original_hash = deck['sha256']
            check_record = copy.deepcopy(deck)
            build_deck(args.scene_bin, project, verified, check_record)
            if sha(verified) != original_hash:
                raise ValueError('Editable project no longer regenerates the pinned native deck')
            deck['pages'] = [p for p in deck['pages'] if p['template_id'] in chosen]
            for i, p in enumerate(deck['pages']):
                p['measurement_slide'] = p['pdf_page']
                p['pdf_page'] = i + 1
            deck['expected_pdf_pages'] = len(deck['pages'])
            source = deck['pages'][0]['source'].split(':')[0]
            project = out / 'projects' / source
            deck_path = out / deck['path']
            # Reuse the already-open, hash-verified native source when available.
            # This avoids new PowerPoint file-access prompts for measurement.
            native = None
            if (src / 'native-render.json').exists():
                native_record = read(src / 'native-render.json')
                candidates = [d for d in native_record['decks'] if d['path'] == deck['path'] and d['sha256'] == original_hash]
                if len(candidates) == 1:
                    candidate = root / candidates[0]['native_pptx']
                    if sha(candidate) == original_hash:
                        exists = run(['osascript', '-e', 'on run argv\n tell application "Microsoft PowerPoint" to return exists presentation (item 1 of argv)\nend run', candidate.name]).strip()
                        if exists == 'true':
                            native = candidate
            if native is None:
                native = native_dir / f'{args.name}-measure-{source}.pptx'
                if native.exists():
                    raise FileExistsError(native)
                shutil.copy2(verified, native)
                run(['osascript', root / 'scripts/export-powerpoint.applescript', native, native.with_suffix('.pdf')])
            for page in deck['pages']:
                tid = page['template_id']
                print('Measuring ' + tid, flush=True)
                intent = intents[tid]
                folder = evidence_dir / tid
                folder.mkdir()
                scene_path = project / 'slides' / f"uhg-{page['source_slide_number']:03d}.json"
                shutil.copy2(scene_path, folder / 'input-scene.json')
                doc = read(scene_path)
                scene = restored(doc)
                index, text_shape = one_object(scene, intent['text_object_id'])
                _, picture = one_object(scene, intent['picture_object_id'])
                tx, tg = geometry(text_shape)
                px, pg = geometry(picture)
                phrase = intent['phrase']
                row = {'template_id': tid, 'phrase': phrase, 'mode': intent['mode'], 'text_unchanged': True, 'text_frame_unchanged': True}
                application['templates'].append(row)
                try:
                    measurement = json.loads(run(['osascript', root / 'scripts/measure-pptx-text.applescript', native.name, page['measurement_slide'], index, phrase]))
                    write(folder / 'measurement-raw.json', measurement)
                    name = attr(child(child(text_shape, 'nvSpPr'), 'cNvPr'), 'name')
                    frame = measurement.get('shape_bounds', {})
                    if measurement['shape_name'] != name or normalized(measurement['text']) != full_text(text_shape) or any(abs(frame.get(k, float('inf')) - v) > .05 for k, v in zip(('left','top','width','height'), tg)):
                        raise ValueError('Native measurement does not match scene identity/text/frame')
                    if measurement.get('rotation_degrees') is None:
                        if attr(tx, 'rot') not in ('', '0'):
                            raise ValueError('Rotated text requires manual placement')
                        measurement.update(rotation_degrees=0, rotation_source='pinned_scene_OOXML_after_native_name_text_frame_match; native_deck_rechecked_by_apply_accent')
                    else:
                        measurement['rotation_source'] = 'native_PowerPoint'
                    write(folder / 'measurement.json', measurement)
                    embeds = {attr(n, 'r:embed') for n in walk(picture) if attr(n, 'r:embed')}
                    parts = []
                    for rel in doc['relationships']['children']:
                        if attr(rel, 'Id') in embeds:
                            target = attr(rel, 'Target')
                            parts.append(posixpath.normpath(posixpath.join(posixpath.dirname(doc['part']), target)).lstrip('/'))
                    parts.sort(key=lambda p: (not p.endswith('.svg'), p))
                    asset_part = next(p for p in parts if Path(p).suffix.lower() in ('.svg', '.png', '.emf'))
                    asset = folder / Path(asset_part).name
                    with zipfile.ZipFile(native) as z:
                        asset.write_bytes(z.read(asset_part))
                    measured_asset = asset
                    if asset.suffix.lower() == '.emf':
                        measured_asset = folder / 'embedded-bitmap.png'
                        extraction = json.loads(run(['python3', root / 'scripts/extract-emf-bitmap.py', asset, measured_asset]))
                        write(folder / 'bitmap-extraction.json', extraction)
                    bounds = json.loads(run(['swift', root / 'scripts/svg-visible-bounds.swift', '--resolution', '4096', measured_asset]))
                    bounds['file'] = str(asset)
                    bounds['measurement_asset'] = artifact(measured_asset)
                    write(folder / 'asset-bounds.json', bounds)
                    command = [args.anchor_bin, '--mode', intent['mode'], '--measurement', folder / 'measurement.json', '--asset', folder / 'asset-bounds.json', '--units', 'points']
                    if intent['mode'] == 'underline':
                        command += ['--container-aspect', pg[2] / pg[3], '--asset-rotation-deg', int(attr(px, 'rot') or 0) / 60000]
                    else:
                        command += ['--asset-rotation-deg', int(attr(px, 'rot') or 0) / 60000]
                    for k, v in intent.get('parameters', {}).items():
                        command += ['--' + k.replace('_', '-'), v]
                    placement = json.loads(run(command))
                    write(folder / 'placement.json', placement)
                    if placement['status'] != 'placed':
                        raise ValueError(placement.get('operator_note', 'Solver requires manual placement'))
                    proof = {'schema': 'pptxgengo.accent-evidence.v1', 'scene': artifact(folder / 'input-scene.json'), 'measurement': artifact(folder / 'measurement.json'), 'placement': artifact(folder / 'placement.json'), 'asset': artifact(asset), 'asset_part': asset_part, 'native_deck': artifact(native), 'native_slide': page['measurement_slide'], 'text_object_id': str(intent['text_object_id']), 'picture_object_id': str(intent['picture_object_id']), 'phrase': phrase}
                    write(folder / 'evidence.json', proof)
                    applied = folder / 'output-scene.json'
                    run([args.template_bin, 'apply-accent', '--evidence', folder / 'evidence.json', '--out', applied])
                    shutil.copy2(applied, scene_path)
                    row.update(status='applied_pending_native_visual_review', evidence=artifact(folder / 'evidence.json'), output_scene=artifact(applied))
                except (ValueError, subprocess.SubprocessError, StopIteration) as error:
                    # Retain original artwork and record a precise placement instruction.
                    # This does not pretend an unresolved accent has been positioned.
                    row.update(status='manual_required', reason=str(error), operator_note=f"Place existing picture {intent['picture_object_id']} {('behind' if intent['mode']=='highlight' else 'under')} the words {phrase!r} in text shape {intent['text_object_id']}. Original artwork retained; this output is not visually approved.")
                write(out / 'accent-application.json', application)
            build_deck(args.scene_bin, project, deck_path, deck)
        application['status'] = 'completed_pending_visual_review' if all(t['status'].startswith('applied') for t in application['templates']) else 'manual_review_required'
        write(out / 'accent-application.json', application)
        record['accent_application'] = artifact(out / 'accent-application.json')
        record['visual_review'] = 'pending'
        write(out / 'review-bundle.json', record)
    except Exception as error:
        application.update(status='failed', failure=str(error))
        write(out / 'accent-application.json', application)
        raise
    print(json.dumps(application, indent=2))


if __name__ == '__main__':
    main()
