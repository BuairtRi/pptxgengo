#!/usr/bin/env python3
"""Extract all shortlisted source scenes and write per-object binding guides.

This prepares local source dependencies, not editable implementations or fit proof.
Original source hashes are verified. Existing projects are preserved and checked.
"""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def read(path):
    return json.loads(path.read_text())

def guide(scene_path, source):
    scene = read(scene_path)
    objects = {}
    paragraph_paths = {}
    for b in scene['bindings']:
        ident = b.get('object_id', '')
        obj = objects.setdefault(ident, dict(object_id=ident, name=b.get('object_name', ''), text=[], paragraphs=[], geometry=[], colors=[], typography=[]))
        node = scene['scene']
        ancestors = [node.get('name', '')]
        paragraph = None
        for depth, index in enumerate(b['node_path']):
            node = node['children'][index]
            ancestors.append(node.get('name', ''))
            if node.get('name', '').split(':')[-1] == 'p':
                paragraph = tuple(b['node_path'][:depth + 1])
        prop = b['property']
        item = dict(id=b['binding_id'], property=prop, value=b['value'], node_path=b['node_path'], ancestors=ancestors)
        if prop == 'text':
            obj['text'].append(item)
            key = (ident, paragraph)
            if key not in paragraph_paths:
                paragraph_paths[key] = dict(binding_ids=[], source_runs=[], source_text='')
                obj['paragraphs'].append(paragraph_paths[key])
            p = paragraph_paths[key]
            p['binding_ids'].append(b['binding_id'])
            p['source_runs'].append(b['value'])
            p['source_text'] += b['value']
        elif 'Clr' in prop:
            obj['colors'].append(item)
        elif prop.startswith('transform.'):
            obj['geometry'].append(item)
        else:
            obj['typography'].append(item)
    return dict(scene_sha256=digest(scene_path), source_sha256=source['source_sha256'], source=f"{source['source_id']}:{scene['source_slide_number']:03}", objects=list(objects.values()))

def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--scene-bin', default='/tmp/pptxscene-rollout')
    args = ap.parse_args()
    entries = read(ROOT / 'library/templates/rollout/assignments.json')['entries']
    for source in read(ROOT / 'planning/source-registry.json')['sources']:
        selected = [e for e in entries if e['representative'].split(':')[0] == source['source_id']]
        if not selected:
            continue
        source_path = ROOT / source['path_hint']
        if digest(source_path) != source['source_sha256']:
            raise ValueError('Source hash changed: ' + source['source_id'])
        project = ROOT / selected[0]['source_project']
        numbers = sorted({int(e['representative'].split(':')[1]) for e in selected})
        if not project.exists():
            subprocess.run([args.scene_bin, 'extract', '--source', str(source_path), '--slides', ','.join(map(str, numbers)), '--out', str(project)], check=True)
        manifest = read(project / 'manifest.json')
        if manifest['source_sha256'] != source['source_sha256']:
            raise ValueError('Existing project source mismatch: ' + str(project))
        for number in numbers:
            scene_path = project / f'slides/uhg-{number:03}.json'
            result = guide(scene_path, source)
            (project / f'binding-guide-{number:03}.json').write_text(json.dumps(result, indent=2) + '\n')
        print(source['source_id'], len(numbers), 'source scenes/guides ready')

if __name__ == '__main__':
    main()
