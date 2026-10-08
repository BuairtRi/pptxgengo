#!/usr/bin/env python3
"""Make controlled XML working copies of a receipt-backed architecture/nested demo.

This is a qualification helper, not PowerPoint application evidence or adoption.
It does not alter the project, baseline, source, receipt or original archive.
"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import xml.etree.ElementTree as ET
import zipfile

NS = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
      'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
for prefix, uri in NS.items():
    ET.register_namespace(prefix, uri)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def group(tree, name):
    matches = [g for g in tree.findall('.//p:grpSp', NS)
               if g.find('p:nvGrpSpPr/p:cNvPr', NS).get('name') == name]
    if len(matches) != 1:
        raise ValueError('expected one catalog group: ' + name)
    return matches[0]


def transform(node):
    return node.find('p:grpSpPr/a:xfrm', NS)


def move(tree):
    off = transform(group(tree, 'node06')).find('a:off', NS)
    off.set('y', str(int(off.get('y')) + 12 * 12700))


def resize(tree):
    extent = transform(group(tree, 'node07')).find('a:ext', NS)
    extent.set('cy', str(int(extent.get('cy')) + 12 * 12700))


def remove(tree):
    node = group(tree, 'node08')
    for parent in tree.iter():
        if node in list(parent):
            parent.remove(node)
            return
    raise ValueError('missing removal parent')


def add(tree, inherited_tags=False):
    node = copy.deepcopy(group(tree, 'node07'))
    maximum = max(int(i.get('id')) for i in tree.findall('.//p:cNvPr', NS))
    for i, identity in enumerate(node.findall('.//p:cNvPr', NS), 1):
        identity.set('id', str(maximum + i))
        identity.set('name', identity.get('name').replace('node07', 'added-service'))
    for parent in node.iter():
        for child in list(parent):
            if not inherited_tags and child.tag == '{' + NS['p'] + '}custDataLst':
                parent.remove(child)
    for leaf in node.findall('.//a:t', NS):
        leaf.text = 'Monitoring'
    off = transform(node).find('a:off', NS)
    # Reuse the third service's row position. In the combined case that service
    # was removed; the add-only case intentionally retains the occupied slot.
    off.set('x', str(int(off.get('x')) + 160 * 12700))
    tree.find('p:cSld/p:spTree', NS).append(node)


def add_with_tags(tree):
    add(tree, inherited_tags=True)


def reorder(tree):
    node = group(tree, 'node06')
    shapes = tree.find('p:cSld/p:spTree', NS)
    shapes.remove(node)
    shapes.append(node)


def reroute(tree):
    node = group(tree, 'node12')
    # Move the existing vertical connector as one native grouped object.
    # A route with additional bends is a separate topology/geometry case.
    off = transform(node).find('a:off', NS)
    off.set('x', str(int(off.get('x')) + 18 * 12700))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--project', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    state = json.loads((args.project / 'state.json').read_text())
    build = args.project / 'builds' / state['current_build']
    receipt_bytes = (build / 'receipt.json').read_bytes()
    if digest(receipt_bytes) != state['receipt_sha256']:
        raise ValueError('receipt differs from the state pin')
    receipt = json.loads(receipt_bytes)
    baseline = (build / 'deck.pptx').read_bytes()
    if digest(baseline) != receipt['outputs']['deck.pptx']:
        raise ValueError('immutable baseline changed')
    canonical = json.loads((build / 'source.canonical.json').read_text())
    reference = canonical['slides'][0]['template'] if len(canonical['slides']) == 1 else {}
    local = canonical.get('local_templates', {}).get(reference.get('id'), {})
    parent = local.get('provenance', {}).get('parent', {})
    catalog_derived = reference.get('scope') == 'shared' and reference.get('id') == 'architecture/nested'
    catalog_derived = catalog_derived or (reference.get('scope') == 'local' and parent.get('scope') == 'shared' and parent.get('id') == 'architecture/nested')
    if not catalog_derived:
        raise ValueError('this fixture helper requires the one-slide architecture/nested catalog demo')
    args.out.mkdir(parents=False, exist_ok=False)
    cases = {'no-op': [], 'move': [move], 'resize': [resize], 'remove': [remove],
             'add': [add], 'reorder': [reorder], 'connector-move': [reroute],
             'combined-supported': [move, resize, reorder, reroute],
             'combined': [move, resize, remove, add, reorder, reroute],
             'add-inherited-tags': [add_with_tags],
             'combined-inherited-tags': [move, resize, remove, add_with_tags, reorder, reroute]}
    results = {}
    with zipfile.ZipFile(build / 'deck.pptx') as archive:
        original = archive.read('ppt/slides/slide1.xml')
        for name, actions in cases.items():
            destination = args.out / (name + '.pptx')
            if not actions:
                destination.write_bytes(baseline)
            else:
                tree = ET.fromstring(original)
                for action in actions:
                    action(tree)
                replacement = ET.tostring(tree, encoding='utf-8', xml_declaration=True)
                with zipfile.ZipFile(destination, 'w') as edited:
                    for info in archive.infolist():
                        edited.writestr(copy.copy(info), replacement if info.filename == 'ppt/slides/slide1.xml' else archive.read(info.filename))
            results[name] = {'file': destination.name, 'sha256': digest(destination.read_bytes()),
                             'execution': 'controlled XML edit; PowerPoint execution not recorded'}
    manifest = {'schema': 'pptxgengo.geometry-catalog-fixtures.v1', 'baseline_build': state['current_build'],
                'baseline_receipt_sha256': digest(receipt_bytes), 'baseline_pptx_sha256': digest(baseline),
                'catalog_template': 'architecture/nested', 'cases': results}
    (args.out / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    print(json.dumps(manifest, indent=2))


if __name__ == '__main__':
    main()
