#!/usr/bin/env python3
"""Render verified picture-only accent changes through an already-open review.

This fallback replays only picture frame changes, exports PDF, and restores the
original in-memory frames. It never saves or closes the source presentation.
The manifest records this method explicitly; it is not a reopen/repair check of
the newly built PPTX. It refuses unsaved source presentations or scene changes
other than the pinned picture frames.
"""
import argparse
import copy
import importlib.util
import json
from pathlib import Path
import shutil

spec = importlib.util.spec_from_file_location('accents', Path(__file__).with_name('adapt-template-accents.py'))
a = importlib.util.module_from_spec(spec)
spec.loader.exec_module(a)


def q(s):
    return '"' + str(s).replace('\\', '\\\\').replace('"', '\\"') + '"'


def replay(presentation, jobs, pdf, allow_restored=False):
    lines = ['with timeout of 120 seconds', 'tell application "Microsoft PowerPoint"', f'set p to presentation {q(presentation)}', '-- Replay touches only the verified picture frames; original presentation remains open and unsaved changes are preserved.', 'set snapshots to {}', 'try']
    if not allow_restored:
        lines.insert(3, 'if not (saved of p) then error "Source review has unsaved edits; replay refused"')
    for j in jobs:
        old, new = j['old'], j['new']
        lines += [f'set sh to shape {q(j["name"])} of slide {j["slide"]} of p', f'if name of sh is not {q(j["name"])} then error "Native picture name differs"', f'if (count of (every shape of slide {j["slide"]} of p whose name is {q(j["name"])})) is not 1 then error "Ambiguous picture name"', f'if (z order position of sh) is not {j["shape"]} then error "Native picture order differs"']
        for prop, val in zip(('left position','top','width','height'), old):
            lines += [f'if abs((get {prop} of sh) - {val}) > 0.05 then error "Native picture geometry differs"']
        lines += ['set end of snapshots to {sh, left position of sh, top of sh, width of sh, height of sh, lock aspect ratio of sh, z order position of sh}', 'set lock aspect ratio of sh to false']
        for prop, val in zip(('width','height','left position','top'), (new[2],new[3],new[0],new[1])):
            lines += [f'set {prop} of sh to {val}']
        for prop, val in zip(('left position','top','width','height'), new):
            lines += [f'if abs((get {prop} of sh) - {val}) > 0.05 then error ("Native picture did not accept {prop}: " & (get {prop} of sh) & " vs {val}")']
        if j['shape'] != j['new_shape']:
            direction = 'send shape backward' if j['new_shape'] < j['shape'] else 'bring shape forward'
            lines += [f'repeat {abs(j["shape"]-j["new_shape"])} times', f'z order sh z order position {direction}', 'end repeat']
    restore = ['repeat with snap in snapshots', 'set sh to item 1 of snap', 'set lock aspect ratio of sh to false', 'set width of sh to item 4 of snap', 'set height of sh to item 5 of snap', 'set left position of sh to item 2 of snap', 'set top of sh to item 3 of snap', 'set lock aspect ratio of sh to item 6 of snap', 'repeat while (z order position of sh) < (item 7 of snap)', 'z order sh z order position bring shape forward', 'end repeat', 'repeat while (z order position of sh) > (item 7 of snap)', 'z order sh z order position send shape backward', 'end repeat', 'end repeat']
    lines += [f'save p in (POSIX file {q(pdf)}) as save as PDF', 'on error msg number num'] + restore + ['error msg number num', 'end try'] + restore + ['end tell', 'end timeout', 'on abs(v)', 'if v < 0 then return -v', 'return v', 'end abs']
    # `my abs` avoids resolving the handler inside PowerPoint's dictionary.
    return '\n'.join(lines).replace('if abs(', 'if my abs(')


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('bundle', type=Path)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--name', required=True)
    p.add_argument('--scene-bin', required=True)
    p.add_argument('--id', help='render only this template')
    p.add_argument('--allow-restored-task-session', action='store_true', help='only for this tool’s prior task session after verifying its restored frames; never for unsaved user edits')
    x = p.parse_args()
    root = Path(__file__).resolve().parents[1]
    src, out = x.bundle.resolve(), x.out.resolve()
    if out.exists() or out.is_relative_to(src) or src.is_relative_to(out):
        p.error('out must be a new independent directory')
    if not x.name or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in x.name):
        p.error('unsafe output prefix')
    application = a.read(src / 'accent-application.json')
    if application['status'] != 'completed_pending_visual_review':
        p.error('all accents must be applied before replay')
    input_manifest = application['input_manifest']
    if a.sha(input_manifest['path']) != input_manifest['sha256']:
        raise ValueError('Input manifest changed')
    original = a.read(input_manifest['path'])
    source_bundle = Path(input_manifest['path']).parent
    original_native = a.read(source_bundle / 'native-render.json')
    record = a.read(src / 'review-bundle.json')
    if x.id:
        record['templates'] = [r for r in record['templates'] if r['template_id']==x.id]
        for d in record['decks']:
            d['pages'] = [r for r in d['pages'] if r['template_id']==x.id]
        record['decks'] = [d for d in record['decks'] if d['pages']]
        application['templates'] = [r for r in application['templates'] if r['template_id']==x.id]
        if not record['templates']:
            raise ValueError('Template not in bundle')
    shutil.copytree(src, out, ignore=shutil.ignore_patterns('native-render.json', 'render'))
    record['render_method'] = 'verified_picture_geometry_replay_in_existing_native_presentation'
    record['render_limitations'] = ['Native PDF verifies the picture-only geometry changes. Opening/repair validation of the newly built PPTX remains pending while PowerPoint rejects new opens. Original frames restored in memory; source PPTX files never saved.']
    record['status'] = 'native_render_in_progress'
    record['visual_review'] = 'pending'
    try:
        for deck in record['decks']:
            base = next(d for d in original['decks'] if d['path'] == deck['path'])
            opened = next(d for d in original_native['decks'] if d['path'] == deck['path'])
            native_input = root / opened['native_pptx']
            if a.sha(native_input) != base['sha256']:
                raise ValueError('Native source deck changed')
            if not native_input.resolve().is_relative_to((root / 'samples/visual-wave3').resolve()) or not native_input.name.startswith('template-rollout-'):
                raise ValueError('Replay is restricted to established template-rollout task copies')
            source = base['pages'][0]['source'].split(':')[0]
            jobs = []
            for row in application['templates']:
                matches = [page for page in deck['pages'] if page['template_id'] == row['template_id']]
                if not matches:
                    continue
                page = matches[0]
                proof_path = Path(row['evidence']['path'])
                if a.sha(proof_path) != row['evidence']['sha256']:
                    raise ValueError('Accent evidence changed')
                proof = a.read(proof_path)
                for key in ('scene','measurement','placement','asset','native_deck'):
                    if a.sha(proof[key]['path']) != proof[key]['sha256']:
                        raise ValueError('Pinned accent input changed')
                if a.sha(row['output_scene']['path']) != row['output_scene']['sha256']:
                    raise ValueError('Accent output changed')
                before = a.restored(a.read(proof['scene']['path']))
                after = a.restored(a.read(row['output_scene']['path']))
                index, oldpic = a.one_object(before, proof['picture_object_id'])
                newindex, newpic = a.one_object(after, proof['picture_object_id'])
                oldxf, oldg = a.geometry(oldpic)
                newxf, newg = a.geometry(newpic)
                if (a.attr(oldxf,'rot') or '0') != (a.attr(newxf,'rot') or '0'):
                    raise ValueError('Replay requires unchanged source rotation')
                for tag in ('off','ext'):
                    old = a.child(oldxf,tag)
                    old.clear(); old.update(copy.deepcopy(a.child(newxf,tag)))
                # Solver may emit explicit zero rotation and move a highlight
                # immediately behind its text. Verify precisely those changes.
                oldxf['attributes'] = copy.deepcopy(newxf.get('attributes', []))
                if not oldxf['attributes']:
                    oldxf.pop('attributes', None)
                oldtree = a.child(a.child(before,'cSld'),'spTree')
                newtree = a.child(a.child(after,'cSld'),'spTree')
                oldtree['children'].remove(oldpic)
                target_index = next(i for i,n in enumerate(newtree['children']) if n is newpic)
                oldtree['children'].insert(target_index,oldpic)
                if before != after:
                    raise ValueError('Scene has changes other than the pinned picture frame')
                name = a.attr(a.child(a.child(oldpic,'nvPicPr'),'cNvPr'),'name')
                jobs.append({'slide': proof['native_slide'], 'shape': index, 'new_shape':newindex, 'name': name, 'old': oldg, 'new': newg})
                page['pdf_page'] = proof['native_slide']
            # Package the full source page set, matching the replay session.
            build_record = copy.deepcopy(base)
            a.build_deck(x.scene_bin, out / 'projects' / source, out / deck['path'], build_record)
            deck.update(sha256=build_record['sha256'], native_build_report=build_record['native_build_report'], expected_pdf_pages=base['expected_pdf_pages'])
            stem = x.name + '-' + source
            pptx = root / 'samples/visual-wave3' / (stem + '.pptx')
            pdf = pptx.with_suffix('.pdf')
            if pptx.exists() or pdf.exists():
                raise FileExistsError(stem)
            shutil.copy2(out / deck['path'], pptx)
            script = out / ('replay-' + source + '.applescript')
            script.write_text(replay(native_input.name, jobs, pdf, x.allow_restored_task_session))
            a.run(['osascript', script])
            if not pdf.exists() or not pdf.read_bytes().startswith(b'%PDF-'):
                raise ValueError('Native export did not create PDF')
            count = int(a.run(['swift','-e','import PDFKit; import Foundation; print(PDFDocument(url: URL(fileURLWithPath: CommandLine.arguments[1]))!.pageCount)',pdf]).strip())
            if count != deck['expected_pdf_pages']:
                raise ValueError('Native page count differs')
            render = out / 'render' / stem
            a.run(['swift', root / 'scripts/render-pdf.swift', pdf, render, *[p['pdf_page'] for p in deck['pages']]])
            deck.update(native_pptx=str(pptx.relative_to(root)),native_pdf=str(pdf.relative_to(root)),native_pdf_sha256=a.sha(pdf),render_directory=str(render.relative_to(root)),actual_pdf_pages=count,replay_source=a.artifact(native_input),replay_script=a.artifact(script))
            for page in deck['pages']:
                png = render / f"slide-{page['pdf_page']:03d}.png"
                page.update(png=str(png.relative_to(root)),png_sha256=a.sha(png))
        record['status'] = 'native_rendered_pending_visual_review'
        a.write(out / 'review-bundle.json', {k:v for k,v in record.items() if k!='status'})
    except Exception as error:
        record.update(status='native_render_failed',failure=str(error))
        a.write(out / 'native-render.json',record)
        raise
    a.write(out / 'native-render.json',record)
    print(out / 'native-render.json')


if __name__ == '__main__':
    main()
