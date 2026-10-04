"""Compare rendered XML/resources; exclude only documented nonvisual metadata."""
import hashlib
import json
import posixpath
import sys
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

production = Path(sys.argv[1])
output = Path(sys.argv[2])
ns = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}
sha = lambda data: hashlib.sha256(data).hexdigest()

def identities(directory):
    doc = json.loads((directory / 'compiled-document.json').read_text())
    return {slide['template_binding']['template']: page
            for page, slide in enumerate(doc['slides'], 1)}

def visual_xml(archive, part):
    root = ET.fromstring(archive.read(part))
    root.find('p:cSld', ns).attrib.pop('name', None)
    for shape in root.findall('.//p:sp', ns):
        name = shape.find('p:nvSpPr/p:cNvPr', ns)
        if name is not None and name.get('name') == 'wm.page':
            for text in shape.findall('.//a:t', ns):
                text.text = 'PAGE_ORDINAL'
    for name in root.findall('.//p:cNvPr', ns):
        name.attrib.pop('name', None)
    return ET.tostring(root)

def resources(archive, part, visited=None):
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
        rows.append((rel.get('Id'), rel.get('Type'), sha(archive.read(target)),
                     resources(archive, target, visited)))
    return sorted(rows)

source, bound = production / 'source', production / 'bound'
source_keys, bound_keys = identities(source), identities(bound)
rows = []
with zipfile.ZipFile(source / 'library-reference.pptx') as a, zipfile.ZipFile(bound / 'library-reference.pptx') as b:
    for key, bound_page in bound_keys.items():
        source_page = source_keys[key]
        sp, bp = f'ppt/slides/slide{source_page}.xml', f'ppt/slides/slide{bound_page}.xml'
        equal = visual_xml(a, sp) == visual_xml(b, bp) and resources(a, sp) == resources(b, bp)
        rows.append({'template': key, 'source_page': source_page, 'bound_page': bound_page,
                     'rendered_parts_equivalent': equal})
receipt = {'schema': 'pptxgengo.wmds-source-bound-render-equivalence.v1',
           'source_pptx_sha256': sha((source / 'library-reference.pptx').read_bytes()),
           'bound_pptx_sha256': sha((bound / 'library-reference.pptx').read_bytes()),
           'normalization': ['nonvisual slide name', 'nonvisual object names', 'visible footer page ordinal only'],
           'resources': 'Internal non-notes relationships and recursive target bytes compared; external URLs compared.',
           'source': len(source_keys), 'bound': len(bound_keys),
           'equivalent': sum(row['rendered_parts_equivalent'] for row in rows),
           'different': [row['template'] for row in rows if not row['rendered_parts_equivalent']],
           'pages': rows}
output.write_text(json.dumps(receipt, indent=2) + '\n')
print(json.dumps({key: value for key, value in receipt.items() if key != 'pages'}, indent=2))
