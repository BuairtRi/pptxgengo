#!/usr/bin/env python3
"""Qualify attached elbow authoring/reconciliation with private catalog projects.

Uses controlled XML edits. Does not establish PowerPoint GUI/Save As fidelity.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import xml.etree.ElementTree as ET
import zipfile

NS = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}


def route(path):
    with zipfile.ZipFile(path) as archive:
        tree = ET.fromstring(archive.read('ppt/slides/slide1.xml'))
    connector = next(node for node in tree.findall('.//p:cxnSp', NS)
                     if node.find('p:nvCxnSpPr/p:cNvPr', NS).get('name') == 'service-edge')
    return connector.find('p:spPr/a:prstGeom/a:avLst/a:gd', NS).get('fmla')


def edit_bend(original, destination, value):
    with zipfile.ZipFile(original) as archive:
        tree = ET.fromstring(archive.read('ppt/slides/slide1.xml'))
        connector = next(node for node in tree.findall('.//p:cxnSp', NS)
                         if node.find('p:nvCxnSpPr/p:cNvPr', NS).get('name') == 'service-edge')
        connector.find('p:spPr/a:prstGeom/a:avLst/a:gd', NS).set('fmla', 'val ' + str(value))
        with zipfile.ZipFile(destination, 'w') as output:
            for info in archive.infolist():
                output.writestr(info, ET.tostring(tree, encoding='utf-8', xml_declaration=True)
                                if info.filename == 'ppt/slides/slide1.xml' else archive.read(info.filename))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--cli', type=Path, required=True)
    parser.add_argument('--bundle', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True, help='new private Documents directory')
    args = parser.parse_args()
    cli, bundle, root = args.cli.resolve(), args.bundle.resolve(), args.out.absolute()
    root.mkdir(parents=False, exist_ok=False)
    summary = {'schema': 'pptxgengo.routed-connector-qualification.v1',
               'cli_sha256': hashlib.sha256(cli.read_bytes()).hexdigest(),
               'execution': 'controlled XML edits; native rendering/GUI Save As not executed', 'modes': {}}
    for mode in ('horizontal', 'vertical'):
        work = root / mode
        work.mkdir()
        logs = work / 'logs'
        logs.mkdir()
        project = work / 'project'

        def run(name, *arguments, refusal=None):
            command = [str(cli), *map(str, arguments)]
            process = subprocess.run(command, capture_output=True, text=True, timeout=150)
            (logs / (name + '.stdout')).write_text(process.stdout)
            (logs / (name + '.stderr')).write_text(process.stderr)
            (logs / (name + '.command.json')).write_text(json.dumps(command, indent=2) + '\n')
            if refusal:
                if not process.returncode or refusal not in process.stderr:
                    raise AssertionError(name + ': expected refusal ' + refusal + ': ' + process.stderr)
                return None
            if process.returncode:
                raise RuntimeError(name + ': ' + process.stderr)
            return json.loads(process.stdout)

        def inspect(name):
            return run(name, 'project', 'diagram', 'inspect', '--project', project,
                       '--slide', 'first-slide', '--bundle', bundle)

        def decisions(packet, destination):
            destination.write_text(json.dumps({'schema': 'pptxgengo.text-review-decisions.v1',
                                              'report_sha256': packet['report_sha256'], 'actor': 'qualification',
                                              'decisions': [{'field_id': field['id'], 'action': 'use_native',
                                                             'reason': 'Adopt controlled bend edit'}
                                                            for field in packet['report']['geometry'] if field['status'] == 'native_only']}, indent=2))

        run('create', 'project', 'create', '--out', project, '--id', 'elbow-demo',
            '--title', 'Architecture elbow demo', '--bundle', bundle, '--template', 'architecture/nested')
        run('detach', 'project', 'detach', '--project', project, '--slide', 'first-slide',
            '--as', 'local-architecture', '--reason', 'Qualify attached elbow routing', '--bundle', bundle)
        arguments = ['project', 'diagram', 'connect', '--project', project, '--slide', 'first-slide',
                     '--bundle', bundle, '--id', 'service-edge', '--from', 'node06', '--to', 'node07',
                     '--route', mode, '--bend', '.35', '--actor', 'qualification', '--reason', 'Show service relationship']
        before = (project / 'slides/first-slide.yaml').read_bytes()
        preview = run('connect-preview', *arguments)
        if preview['applied'] or before != (project / 'slides/first-slide.yaml').read_bytes():
            raise AssertionError('connector preview wrote source')
        run('connect-apply', *arguments, '--apply')
        patch = work / 'move.json'
        patch.write_text(json.dumps({'schema': 'pptxgengo.diagram-patch.v1', 'actor': 'qualification',
                                    'reason': 'Move service and recalculate attached path',
                                    'operations': [{'action': 'move', 'id': 'node07', 'dy_pt': 12}]}))
        run('move', 'project', 'diagram', 'patch', '--project', project, '--slide', 'first-slide',
            '--bundle', bundle, '--patch', patch, '--apply')
        observed = inspect('inspect-before')
        connection = observed['connections'][0]
        if len(connection['points']) != 4:
            raise AssertionError('missing calculated bend points')
        for point, endpoint in [(connection['points'][0], connection['from']), (connection['points'][-1], connection['to'])]:
            if abs(point[0] - endpoint['x_pt']) > .02 or abs(point[1] - endpoint['y_pt']) > .02:
                raise AssertionError('authored movement disconnected endpoint')
        run('contain', 'project', 'diagram', 'contain', '--project', project, '--slide', 'first-slide',
            '--bundle', bundle, '--nodes', 'service-edge', '--container', 'node05', '--padding', '12',
            '--padding-top', '28', '--padding-bottom', '4', '--actor', 'qualification',
            '--reason', 'Keep route inside Services', '--apply')
        baseline = run('baseline', 'project', 'build', '--project', project, '--bundle', bundle)
        original = project / 'builds' / baseline['build_id'] / 'deck.pptx'
        edited = work / 'edited.pptx'
        edit_bend(original, edited, 65000)
        packet = run('propose', 'project', 'reconcile', 'propose', '--project', project, '--bundle', bundle,
                     '--geometry', '--edited', edited, '--out', work / 'review')
        if packet['report']['counts'].get('geometry_native_only') != 1 or packet['report']['manual_review']:
            raise AssertionError('bend edit not a single supported geometry change')
        decision = work / 'decisions.json'
        decisions(packet, decision)
        run('adopt', 'project', 'reconcile', 'adopt', '--project', project, '--bundle', bundle,
            '--packet', work / 'review', '--decisions', decision)
        source = (project / 'slides/first-slide.yaml').read_bytes()
        run('adopt-idempotent', 'project', 'reconcile', 'adopt', '--project', project, '--bundle', bundle,
            '--packet', work / 'review', '--decisions', decision)
        if source != (project / 'slides/first-slide.yaml').read_bytes():
            raise AssertionError('repeat adoption changed source')
        rebuilt = run('rebuilt', 'project', 'build', '--project', project, '--bundle', bundle)
        rebuilt_file = project / 'builds' / rebuilt['build_id'] / 'deck.pptx'
        if route(edited) != route(rebuilt_file) or route(rebuilt_file) != 'val 65000':
            raise AssertionError('bend guide lost after rebuilding')
        final = inspect('inspect-after')
        final_connection = final['connections'][0]
        if connection['points'] == final_connection['points']:
            raise AssertionError('native bend did not change calculated path')
        # A horizontal elbow can escape Services while its endpoints remain fixed.
        if mode == 'horizontal':
            for case, value, refusal in [('inner-escape', 2000000, 'containment'), ('frame-escape', 100000000, 'frame zone')]:
                file = work / (case + '.pptx')
                edit_bend(rebuilt_file, file, value)
                result = run(case + '-propose', 'project', 'reconcile', 'propose', '--project', project,
                             '--bundle', bundle, '--geometry', '--edited', file, '--out', work / (case + '-review'))
                decision_file = work / (case + '-decisions.json')
                decisions(result, decision_file)
                run(case + '-adopt', 'project', 'reconcile', 'adopt', '--project', project, '--bundle', bundle,
                    '--packet', work / (case + '-review'), '--decisions', decision_file, refusal=refusal)
                if source != (project / 'slides/first-slide.yaml').read_bytes():
                    raise AssertionError('refused route wrote source')
        run('version', 'project', 'version', 'save', '--project', project,
            '--actor', 'qualification', '--message', 'Retain native elbow geometry')
        run('share', 'project', 'share', '--project', project, '--out', work / 'project.zip')
        run('extract', 'project', 'share-extract', '--archive', work / 'project.zip', '--out', work / 'extracted')
        if source != (work / 'extracted/slides/first-slide.yaml').read_bytes():
            raise AssertionError('shared YAML lost route metadata')
        summary['modes'][mode] = {'source_movement_keeps_endpoints': True, 'bend_roundtrip_and_idempotence': True,
                                  'version_share_extraction': True, 'rebuilt': str(rebuilt_file),
                                  'calculated_points': final_connection['points']}
    (root / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, indent=2))


if __name__ == '__main__':
    main()
