#!/usr/bin/env python3
"""Build the explicit manual-placement fallback fixture for UHG slide 5.

Assets remain native SVG pictures on the right pasteboard. Instructions are
native off-slide text shapes and speaker notes. The visible slide has no
guessed underline or arrow placement. This is a fixture, not a general CLI.
"""
from pathlib import Path
import json
import shutil
import subprocess
from xml.dom import minidom
from xml.etree import ElementTree as ET
from html import escape

root = Path(__file__).resolve().parent.parent
base = root / 'samples/reconstruction'
source = base / 'parameterized-scene'
project = base / 'work/manual-placement'
if project.exists():
    raise SystemExit('Manual placement fixture already exists')
shutil.copytree(source, project)
slide_path = project / 'slides/uhg-005.json'
slide = json.loads(slide_path.read_text())
intents = [
    {'object_id': '5', 'asset': 'image77.svg', 'name': 'Underline', 'y': 40,
     'instruction': 'MANUAL: Place the staged navy underline beneath “pivotal moment” in the title. Align the visible stroke to those words on their rendered line, not to the bottom of the title box. The SVG has large transparent vertical padding. Keep it clear of the next line.'},
    {'object_id': '20', 'asset': 'image86.svg', 'name': 'Curved arrow', 'y': 210,
     'instruction': 'MANUAL: Place the staged navy dashed arrow in the whitespace between the title and OUR UNDERSTANDING panel, curving down and right toward the panel. Preserve the source curve, rotation and aspect ratio. Keep the tail and tip clear of glyphs and icons; review the whitespace visually.'},
]
for intent in intents:
    for b in slide['bindings']:
        if b.get('object_id') == intent['object_id']:
            if b['property'] == 'transform.off.x': b['value'] = str(1000 * 12700)
            if b['property'] == 'transform.off.y': b['value'] = str(intent['y'] * 12700)


def node(element):
    if element.nodeType == element.TEXT_NODE:
        return {'text': element.data}
    out = {'name': element.tagName}
    attrs = [{'name': a.name, 'value': a.value} for a in element.attributes.values() if not a.name.startswith('xmlns')]
    if attrs: out['attributes'] = attrs
    kids = [node(x) for x in element.childNodes if x.nodeType in (x.ELEMENT_NODE, x.TEXT_NODE)]
    if kids: out['children'] = kids
    return out


def find(n, name):
    if n.get('name') == name: return n
    for child in n.get('children', []):
        result = find(child, name)
        if result is not None: return result


tree = find(slide['scene'], 'p:spTree')
for i, intent in enumerate(intents):
    y = (intent['y'] - 40) * 12700
    xml = f'''<p:sp xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
<p:nvSpPr><p:cNvPr id="{1001+i}" name="MANUAL PLACEMENT: {intent['name']}"/><p:cNvSpPr txBox="1"/><p:nvPr/></p:nvSpPr>
<p:spPr><a:xfrm><a:off x="12700000" y="{y}"/><a:ext cx="4445000" cy="508000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln><a:noFill/></a:ln></p:spPr>
<p:txBody><a:bodyPr wrap="square"/><a:lstStyle/><a:p><a:r><a:rPr sz="1000"><a:solidFill><a:srgbClr val="070154"/></a:solidFill><a:latin typeface="Arial"/></a:rPr><a:t>{escape(intent['instruction'])}</a:t></a:r></a:p></p:txBody></p:sp>'''
    tree['children'].append(node(minidom.parseString(xml).documentElement))
slide_path.write_text(json.dumps(slide, indent=2) + '\n')

namespaces = {'p': 'http://schemas.openxmlformats.org/presentationml/2006/main', 'a': 'http://schemas.openxmlformats.org/drawingml/2006/main'}
for prefix, uri in namespaces.items(): ET.register_namespace(prefix, uri)
notes_path = project / 'resources/ppt/notesSlides/notesSlide4.xml'
notes = ET.parse(notes_path)
for shape in notes.findall('.//p:sp', namespaces):
    ph = shape.find('p:nvSpPr/p:nvPr/p:ph', namespaces)
    if ph is not None and ph.get('type') == 'body':
        body = shape.find('p:txBody', namespaces)
        for intent in intents:
            p = ET.SubElement(body, '{'+namespaces['a']+'}p')
            r = ET.SubElement(p, '{'+namespaces['a']+'}r')
            t = ET.SubElement(r, '{'+namespaces['a']+'}t')
            t.text = intent['instruction'] + ' Asset is staged off-slide to the right; zoom out in Normal view.'
        break
else:
    raise RuntimeError('Source notes body placeholder missing')
notes.write(notes_path, encoding='UTF-8', xml_declaration=True)
(project / 'placement-todo.json').write_text(json.dumps({'status': 'manual_required', 'slide_role': 'Frame the pivotal decision to establish Fabric as a governed enterprise platform.', 'reason': 'Demonstration of intentional manual-placement fallback; no automatic guess.', 'intents': intents}, indent=2) + '\n')
subprocess.run(['go', 'run', './cmd/pptxscene', 'build', '--project', str(project), '--slides', '5', '--freeze-slide-numbers', '--out', str(base / 'output/UHG-5-manual-placement.pptx')], check=True)
