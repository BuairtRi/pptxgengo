#!/usr/bin/env python3
"""Qualify catalog geometry with a fixed pptxdesign binary and private output.

Uses controlled XML edits. Optional native rendering is distinct from desktop
editing/Save As qualification. Originals and closed review packets are retained.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import xml.etree.ElementTree as ET
import zipfile

NS = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}


def transforms(path):
    with zipfile.ZipFile(path) as archive:
        tree = ET.fromstring(archive.read('ppt/slides/slide1.xml'))
    result, order = {}, {}

    def walk(parent, owner=''):
        for node in parent:
            kind = node.tag.split('}')[-1]
            if kind not in ('sp', 'grpSp', 'pic', 'graphicFrame', 'cxnSp'):
                continue
            nv = next(child for child in node if child.tag.split('}')[-1].startswith('nv'))
            name = nv.find('p:cNvPr', NS).get('name')
            if name in result:
                raise ValueError('duplicate native name: ' + name)
            xfrm = node.find('p:grpSpPr/a:xfrm', NS) if kind == 'grpSp' else node.find('p:spPr/a:xfrm', NS)
            if xfrm is None:
                xfrm = node.find('p:xfrm', NS)
            if xfrm is None:
                raise ValueError('missing transform: ' + name)
            attrs = {key: value for key, value in xfrm.attrib.items()
                     if (key == 'rot' and int(value) != 0)
                     or (key in ('flipH', 'flipV') and value not in ('0', 'false'))
                     or key not in ('rot', 'flipH', 'flipV')}
            result[name] = {'kind': kind, 'parent': owner, 'attrs': attrs,
                            'children': [(child.tag.split('}')[-1], child.attrib) for child in xfrm]}
            order.setdefault(owner, []).append(name)
            if kind == 'grpSp':
                walk(node, name)
    walk(tree.find('p:cSld/p:spTree', NS))
    return result, order


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--cli', type=Path, required=True, help='fixed development pptxdesign executable')
    parser.add_argument('--bundle', type=Path, required=True, help='ordinary V11 bundle directory, no symlinks')
    parser.add_argument('--out', type=Path, required=True, help='new private directory; use Documents for PowerPoint')
    parser.add_argument('--structures', action='store_true', help='qualify mapped copies and deletion together with transforms')
    parser.add_argument('--containment', action='store_true', help='also declare Services padding and verify refused source/native escapes')
    parser.add_argument('--native', action='store_true', help='also render original/edited/rebuilt in local PowerPoint')
    args = parser.parse_args()
    cli, bundle, root = args.cli.resolve(), args.bundle.resolve(), args.out.absolute()
    root.mkdir(parents=False, exist_ok=False)
    logs = root / 'logs'
    logs.mkdir()
    project = root / 'project'
    summary = {'schema': 'pptxgengo.geometry-roundtrip-qualification.v1',
               'cli_sha256': hashlib.sha256(cli.read_bytes()).hexdigest(),
               'execution': 'controlled XML edits; PowerPoint editing/Save As not executed',
               'project': str(project), 'cases': {}}

    def run(name, *arguments):
        command = [str(cli), *map(str, arguments)]
        process = subprocess.run(command, capture_output=True, text=True, timeout=150)
        (logs / (name + '.stdout')).write_text(process.stdout)
        (logs / (name + '.stderr')).write_text(process.stderr)
        (logs / (name + '.command.json')).write_text(json.dumps(command, indent=2) + '\n')
        if process.returncode:
            summary['failure'] = {'step': name, 'stderr_log': str(logs / (name + '.stderr'))}
            (root / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
            raise RuntimeError(name + ': ' + process.stderr)
        return json.loads(process.stdout)

    run('create', 'project', 'create', '--out', project, '--id', 'geometry-demo',
        '--title', 'Architecture geometry demo', '--bundle', bundle, '--template', 'architecture/nested')
    run('detach', 'project', 'detach', '--project', project, '--slide', 'first-slide',
        '--as', 'custom-architecture', '--reason', 'Catalog geometry round-trip qualification', '--bundle', bundle)
    if args.containment:
        arguments = ['project', 'diagram', 'contain', '--project', project, '--slide', 'first-slide',
                     '--bundle', bundle, '--nodes', 'node06,node07,node08', '--container', 'node05',
                     '--padding', '12', '--padding-top', '28', '--padding-bottom', '4',
                     '--actor', 'Controlled XML qualification operator', '--reason', 'Keep services inside Services']
        original = (project / 'slides/first-slide.yaml').read_bytes()
        preview = run('contain-preview', *arguments)
        if preview['applied'] or original != (project / 'slides/first-slide.yaml').read_bytes():
            raise AssertionError('containment preview wrote source')
        run('contain-apply', *arguments, '--apply')
    baseline = run('baseline', 'project', 'build', '--project', project, '--bundle', bundle)
    run('inspect', 'project', 'diagram', 'inspect', '--project', project, '--slide', 'first-slide', '--bundle', bundle)
    cases = root / 'xml-cases'
    process = subprocess.run([sys.executable, str(Path(__file__).with_name('geometry-catalog-fixtures.py')),
                              '--project', str(project), '--out', str(cases)],
                             capture_output=True, text=True, check=True, timeout=30)
    (logs / 'fixtures.stdout').write_text(process.stdout)
    if args.containment:
        def refused(name, *arguments):
            command = [str(cli), *map(str, arguments)]
            before = (project / 'slides/first-slide.yaml').read_bytes()
            process = subprocess.run(command, capture_output=True, text=True, timeout=150)
            (logs / (name + '.stdout')).write_text(process.stdout)
            (logs / (name + '.stderr')).write_text(process.stderr)
            (logs / (name + '.command.json')).write_text(json.dumps(command, indent=2) + '\n')
            if not process.returncode or 'containment' not in process.stderr:
                raise AssertionError(name + ': escape was not refused by containment: ' + process.stderr)
            if before != (project / 'slides/first-slide.yaml').read_bytes():
                raise AssertionError(name + ': refused escape wrote source')
        patch = root / 'escape-patch.json'
        patch.write_text(json.dumps({'schema': 'pptxgengo.diagram-patch.v1', 'actor': 'qualification',
                                    'reason': 'Verify inner-container refusal',
                                    'operations': [{'action': 'move', 'id': 'node06', 'dy_pt': 60}]}))
        refused('source-escape', 'project', 'diagram', 'patch', '--project', project, '--slide', 'first-slide',
                '--bundle', bundle, '--patch', patch, '--apply')
        baseline_file = project / 'builds' / baseline['build_id'] / 'deck.pptx'
        escape_file = root / 'native-escape.pptx'
        with zipfile.ZipFile(baseline_file) as archive:
            tree = ET.fromstring(archive.read('ppt/slides/slide1.xml'))
            target = next(node for node in tree.findall('.//p:grpSp', NS)
                          if node.find('p:nvGrpSpPr/p:cNvPr', NS).get('name') == 'node06')
            off = target.find('p:grpSpPr/a:xfrm/a:off', NS)
            off.set('y', str(int(off.get('y')) + 60 * 12700))
            with zipfile.ZipFile(escape_file, 'w') as output:
                for info in archive.infolist():
                    output.writestr(info, ET.tostring(tree, encoding='utf-8', xml_declaration=True)
                                    if info.filename == 'ppt/slides/slide1.xml' else archive.read(info.filename))
        escape = run('native-escape-propose', 'project', 'reconcile', 'propose', '--project', project,
                     '--geometry', '--bundle', bundle, '--edited', escape_file, '--out', root / 'escape-review')
        decision = root / 'escape-decisions.json'
        decision.write_text(json.dumps({'schema': 'pptxgengo.text-review-decisions.v1',
                                       'report_sha256': escape['report_sha256'], 'actor': 'qualification',
                                       'decisions': [{'field_id': field['id'], 'action': 'use_native',
                                                      'reason': 'Verify refusal before source writes'}
                                                     for field in escape['report']['geometry'] if field['status'] == 'native_only']}))
        refused('native-escape-adopt', 'project', 'reconcile', 'adopt', '--project', project,
                '--bundle', bundle, '--packet', root / 'escape-review', '--decisions', decision)
        summary['containment_source_and_native_escape_refused_before_writes'] = True
    supported = None
    selected_case = 'combined-inherited-tags' if args.structures else 'combined-supported'
    mapping = root / 'structure-map.json'
    mapping.write_text(json.dumps({'schema': 'pptxgengo.native-structure-map.v1', 'copies': [
        {'slide_id': 'first-slide', 'native_object': 'added-service', 'source_node': 'node07', 'new_node': 'monitoring'}]}, indent=2) + '\n')
    for name in json.loads(process.stdout)['cases']:
        extra = ['--structure-map', mapping] if args.structures and name in ('add', 'combined', 'add-inherited-tags', 'combined-inherited-tags') else []
        result = run(name + '-propose', 'project', 'reconcile', 'propose', '--project', project,
                     '--geometry', '--bundle', bundle, '--edited', cases / (name + '.pptx'),
                     '--out', root / (name + '-review'), *extra)
        report = result['report']
        summary['cases'][name] = {'counts': report['counts'],
                                 'manual_kinds': sorted(set(item['kind'] for item in report['manual_review']))}
        if name == selected_case:
            supported = result
    expected_geometry = 3 if args.structures else 4
    expected_structure = 2 if args.structures else 0
    if not supported or supported['report']['counts'].get('geometry_native_only') != expected_geometry:
        raise AssertionError('unexpected supported geometry count')
    if supported['report']['counts'].get('structure_native_only', 0) != expected_structure:
        raise AssertionError('unexpected supported structural count')
    if supported['report']['manual_review']:
        raise AssertionError('supported geometry has unresolved manual items')
    if summary['cases']['remove']['counts'].get('structure_native_only') != 1:
        raise AssertionError('whole native deletion not proposed')
    if not args.structures and not summary['cases']['add']['manual_kinds']:
        raise AssertionError('unmapped native copy was silently ignored')
    decisions = {'schema': 'pptxgengo.text-review-decisions.v1', 'report_sha256': supported['report_sha256'],
                 'actor': 'Controlled XML qualification operator',
                 'decisions': [{'field_id': field['id'], 'action': 'use_native',
                                'reason': 'Accept this controlled catalog geometry edit for qualification'}
                               for field in supported['report']['geometry'] + supported['report'].get('structure', []) if field['status'] == 'native_only']}
    decision_path = root / 'geometry-decisions.json'
    decision_path.write_text(json.dumps(decisions, indent=2) + '\n')
    adopted = run('adopt', 'project', 'reconcile', 'adopt', '--project', project, '--bundle', bundle,
                  '--packet', root / (selected_case + '-review'), '--decisions', decision_path)
    if len(adopted['changed']) != expected_geometry + expected_structure or adopted['remaining_field_ids'] or adopted['manual_review']:
        raise AssertionError('adoption is partial or unresolved')
    run('adopt-idempotent', 'project', 'reconcile', 'adopt', '--project', project, '--bundle', bundle,
        '--packet', root / (selected_case + '-review'), '--decisions', decision_path)
    rebuilt = run('rebuilt', 'project', 'build', '--project', project, '--bundle', bundle)
    edited_file = cases / (selected_case + '.pptx')
    rebuilt_file = project / 'builds' / rebuilt['build_id'] / 'deck.pptx'
    edited_transforms, edited_order = transforms(edited_file)
    rebuilt_transforms, rebuilt_order = transforms(rebuilt_file)
    if args.structures:
        edited_transforms = {name.replace('added-service', 'monitoring', 1):
                             {**item, 'parent': item['parent'].replace('added-service', 'monitoring', 1)}
                             for name, item in edited_transforms.items()}
        edited_order = {owner.replace('added-service', 'monitoring', 1):
                        [name.replace('added-service', 'monitoring', 1) for name in names]
                        for owner, names in edited_order.items()}
        current = run('inspect-after', 'project', 'diagram', 'inspect', '--project', project,
                      '--slide', 'first-slide', '--bundle', bundle)
        if any(node['id'] == 'node08' for node in current['nodes']) or not any(node['id'] == 'monitoring' for node in current['nodes']):
            raise AssertionError('native copy/deletion did not persist in source')
        summary['whole_component_deletion_and_mapped_copy_passed'] = True
        if args.containment:
            rules = {rule['member']: rule for rule in current['containment']}
            if set(rules) != {'node06', 'node07', 'monitoring'} or rules['monitoring']['container'] != 'node05':
                raise AssertionError('copy inheritance or deletion pruning failed')
            if current.get('overlaps'):
                raise AssertionError('unexpected sibling allocation overlap')
            summary['containment_clearances_and_copy_inheritance_passed'] = True
    if edited_transforms != rebuilt_transforms or edited_order != rebuilt_order:
        raise AssertionError('rebuilt transforms or paint order differ from edited geometry')
    summary.update(objects_compared=len(edited_transforms), all_transforms_equal=True,
                   all_paint_orders_equal=True, rebuilt=str(rebuilt_file), edited=str(edited_file))
    run('version', 'project', 'version', 'save', '--project', project,
        '--actor', 'Controlled XML qualification operator', '--message', 'Retain reconciled native geometry')
    run('share', 'project', 'share', '--project', project, '--out', root / 'project.zip')
    run('extract', 'project', 'share-extract', '--archive', root / 'project.zip', '--out', root / 'extracted')
    original_slide = (project / 'slides/first-slide.yaml').read_bytes()
    if original_slide != (root / 'extracted/slides/first-slide.yaml').read_bytes():
        raise AssertionError('shared slide YAML differs after extraction')
    summary['version_share_extraction_passed'] = True
    if args.native:
        staging = root / 'powerpoint-staging'
        run('render-doctor', 'render-doctor', '--staging-dir', staging, '--timeout', '30s', '--json')
        files = {'baseline': project / 'builds' / baseline['build_id'] / 'deck.pptx',
                 'edited': edited_file, 'rebuilt': rebuilt_file}
        pngs = {}
        for name, file in files.items():
            output = root / (name + '-render')
            receipt = run(name + '-render', 'render', '--pptx', file, '--out', output, '--png', '--pdf',
                          '--staging-dir', staging, '--timeout', '90s')
            pngs[name] = (output / receipt['pngs'][0]['path']).read_bytes()
        if pngs['edited'] != pngs['rebuilt']:
            raise AssertionError('PowerPoint PNGs differ; inspect the native renders')
        summary['native_render'] = {'renderer': 'Microsoft PowerPoint and local PNG rasterizer',
                                    'edited_rebuilt_png_bytes_equal': True,
                                    'sha256': hashlib.sha256(pngs['edited']).hexdigest(),
                                    'visual_inspection': 'required separately'}
    (root / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, indent=2), flush=True)


if __name__ == '__main__':
    main()
