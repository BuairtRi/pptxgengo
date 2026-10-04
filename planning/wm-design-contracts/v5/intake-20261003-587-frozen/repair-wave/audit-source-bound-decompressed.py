"""Extend the original audit with strict decompressed XLSX resource comparison.

Only ZIP container metadata and the two build timestamps in XLSX core metadata
are excluded. Worksheet, chart, style, relationship and other member bytes stay
exact. The original audit script and receipt are retained unchanged.
"""
import hashlib
import io
import json
import posixpath
import re
import sys
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

NS = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}
SHA = lambda data: hashlib.sha256(data).hexdigest()


def core_metadata(data):
    """Replace only created/modified text; retain every other metadata byte."""
    root = ET.fromstring(data)
    namespace = 'http://purl.org/dc/terms/'
    for field in ('created', 'modified'):
        nodes = root.findall('{%s}%s' % (namespace, field))
        if len(nodes) != 1:
            raise ValueError('expected one XLSX build timestamp: ' + field)
        # Match the qualified tag and preserve its exact prefix/attributes.
        pattern = rb'(<([A-Za-z_][\w.-]*):' + field.encode() + rb'\b[^>]*>)([^<]*)(</\2:' + field.encode() + rb'>)'
        data, count = re.subn(pattern, lambda match: match[1] + b'BUILD_TIMESTAMP' + match[4], data)
        if count != 1:
            raise ValueError('unexpected XLSX timestamp serialization: ' + field)
    return data


def xlsx_members(data, normalize=False):
    with zipfile.ZipFile(io.BytesIO(data)) as package:
        names = package.namelist()
        if len(names) != len(set(names)):
            raise ValueError('duplicate XLSX member names')
        rows = []
        for name in sorted(names):
            member = package.read(name)
            if normalize and name == 'docProps/core.xml':
                member = core_metadata(member)
            rows.append((name, SHA(member)))
        return rows


def identities(directory):
    doc = json.loads((directory / 'compiled-document.json').read_text())
    result = {slide['template_binding']['template']: page
              for page, slide in enumerate(doc['slides'], 1)}
    if len(result) != len(doc['slides']):
        raise ValueError('duplicate template identities')
    return result


def visual_xml(archive, part):
    root = ET.fromstring(archive.read(part))
    root.find('p:cSld', NS).attrib.pop('name', None)
    for shape in root.findall('.//p:sp', NS):
        name = shape.find('p:nvSpPr/p:cNvPr', NS)
        if name is not None and name.get('name') == 'wm.page':
            for text in shape.findall('.//a:t', NS):
                text.text = 'PAGE_ORDINAL'
    for name in root.findall('.//p:cNvPr', NS):
        name.attrib.pop('name', None)
    return ET.tostring(root)


def resources(archive, part, packages, visited=None):
    visited = set() if visited is None else visited
    if part in visited:
        return []
    visited.add(part)
    rels = posixpath.join(posixpath.dirname(part), '_rels',
                         posixpath.basename(part) + '.rels')
    if rels not in archive.namelist():
        return []
    rows = []
    for rel in ET.fromstring(archive.read(rels)):
        if rel.get('Type').endswith('/notesSlide'):
            continue
        if rel.get('TargetMode') == 'External':
            rows.append((rel.get('Id'), rel.get('Type'), rel.get('Target'), []))
            continue
        target = rel.get('Target')
        target = target.lstrip('/') if target.startswith('/') else posixpath.normpath(
            posixpath.join(posixpath.dirname(part), target))
        data = archive.read(target)
        if target.endswith('.xlsx'):
            packages[target] = {'raw_sha256': SHA(data),
                                'raw_members': xlsx_members(data),
                                'normalized_members': xlsx_members(data, True)}
            digest = SHA(json.dumps(packages[target]['normalized_members'], separators=(',', ':')).encode())
        else:
            digest = SHA(data)
        rows.append((rel.get('Id'), rel.get('Type'), digest,
                     resources(archive, target, packages, visited)))
    return sorted(rows)


def audit(production, output):
    source, bound = production / 'source', production / 'bound'
    source_keys, bound_keys = identities(source), identities(bound)
    missing_source = set(bound_keys) - set(source_keys)
    if missing_source:
        raise ValueError('bound identities absent from source: ' + str(sorted(missing_source)))
    rows, sa, sb = [], {}, {}
    with zipfile.ZipFile(source / 'library-reference.pptx') as a, zipfile.ZipFile(bound / 'library-reference.pptx') as b:
        for key, bound_page in bound_keys.items():
            source_page = source_keys[key]
            sp, bp = f'ppt/slides/slide{source_page}.xml', f'ppt/slides/slide{bound_page}.xml'
            slide_equal = visual_xml(a, sp) == visual_xml(b, bp)
            resource_equal = resources(a, sp, sa) == resources(b, bp, sb)
            rows.append({'template': key, 'source_page': source_page, 'bound_page': bound_page,
                         'visual_xml_equivalent': slide_equal, 'resources_equivalent': resource_equal,
                         'rendered_parts_equivalent': slide_equal and resource_equal})
    packages = []
    if sa.keys() != sb.keys():
        raise ValueError('XLSX resource paths differ')
    for part in sorted(sa):
        left, right = sa[part], sb[part]
        lm, rm = dict(left['raw_members']), dict(right['raw_members'])
        packages.append({'part': part, 'source_raw_sha256': left['raw_sha256'],
                         'bound_raw_sha256': right['raw_sha256'],
                         'decompressed_member_differences': [name for name in sorted(lm.keys() | rm.keys()) if lm.get(name) != rm.get(name)],
                         'normalized_members_equivalent': left['normalized_members'] == right['normalized_members'],
                         'members': [{'part': name, 'source_sha256': digest, 'bound_sha256': dict(right['normalized_members']).get(name)}
                                     for name, digest in left['normalized_members']]})
    receipt = {'schema': 'pptxgengo.wmds-source-bound-render-equivalence.v2',
               'production': str(production),
               'source_pptx_sha256': SHA((source / 'library-reference.pptx').read_bytes()),
               'bound_pptx_sha256': SHA((bound / 'library-reference.pptx').read_bytes()),
               'normalization': ['nonvisual slide name', 'nonvisual object names', 'visible footer page ordinal only',
                                 'XLSX ZIP container framing/metadata', 'XLSX docProps/core.xml created/modified timestamp text only'],
               'resources': 'Recursive internal non-notes resources; exact member names/bytes for XLSX except the two timestamp fields; exact external URLs.',
               'source': len(source_keys), 'bound': len(bound_keys),
               'source_only': sorted(source_keys.keys() - bound_keys.keys()),
               'equivalent': sum(row['rendered_parts_equivalent'] for row in rows),
               'different': [row['template'] for row in rows if not row['rendered_parts_equivalent']],
               'xlsx_packages': packages, 'pages': rows}
    output.write_text(json.dumps(receipt, indent=2) + '\n')
    print(json.dumps({key: value for key, value in receipt.items() if key not in ('pages', 'xlsx_packages')}, indent=2))


if __name__ == '__main__':
    audit(Path(sys.argv[1]), Path(sys.argv[2]))
