#!/usr/bin/env python3
"""Inventory slide/layout structure for the product survey; no rendering or edits.

Only explicit OOXML properties are reported. Bounds are in each object's local
coordinate system; fonts/styles inherited from layouts, masters or themes are
not resolved. Observed text lengths are not capacity limits. This is an inventory,
not the proposed lossless importer or a visual QA tool.
"""

import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import posixpath
import xml.etree.ElementTree as ET
from zipfile import ZipFile

NS = {
    "a": "http://schemas.openxmlformats.org/drawingml/2006/main",
    "p": "http://schemas.openxmlformats.org/presentationml/2006/main",
    "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
}
OBJECTS = {"sp", "pic", "graphicFrame", "cxnSp", "grpSp", "contentPart"}


def local(tag):
    return tag.rsplit("}", 1)[-1]


def text(node):
    if node is None:
        return ""
    return "\n".join(
        "".join(n.text or "" for n in paragraph.iter() if local(n.tag) == "t")
        for paragraph in node.findall(".//a:p", NS)
    )


def relationships(zf, part):
    relpart = posixpath.join(posixpath.dirname(part), "_rels", posixpath.basename(part) + ".rels")
    if relpart not in zf.namelist():
        return []
    result = []
    for rel in ET.fromstring(zf.read(relpart)):
        entry = dict(rel.attrib)
        target = entry.get("Target", "")
        if entry.get("TargetMode") != "External":
            resolved = posixpath.normpath(posixpath.join(posixpath.dirname(part), target)).lstrip("/")
            entry["resolved_part"] = resolved
            entry["target_exists"] = resolved in zf.namelist()
        result.append(entry)
    return result


def objects(tree, parent=None):
    result = []
    for node in tree:
        kind = local(node.tag)
        if kind not in OBJECTS:
            continue
        props = next((n for child in node if local(child.tag).startswith("nv")
                      for n in child if local(n.tag) == "cNvPr"), None)
        placeholder = next((n for child in node if local(child.tag).startswith("nv")
                            for n in child.iter() if local(n.tag) == "ph"), None)
        xfrm = next((n for n in list(node) if local(n.tag) == "xfrm"), None)
        if xfrm is None:
            for child in node:
                if local(child.tag) in {"spPr", "grpSpPr"}:
                    xfrm = child.find("a:xfrm", NS)
                    if xfrm is not None:
                        break
        record = {
            "kind": kind,
            "id": props.get("id") if props is not None else None,
            "name": props.get("name") if props is not None else None,
            "parent_group_id": parent,
            "placeholder": dict(placeholder.attrib) if placeholder is not None else None,
            "transform": ({"attributes": dict(xfrm.attrib),
                           **{local(n.tag): dict(n.attrib) for n in xfrm}} if xfrm is not None else None),
        }
        if kind == "grpSp":
            record["children"] = objects(node, record["id"])
        else:
            record["text"] = text(node)
            record["observed_characters"] = len(record["text"])
            record["explicit_font_faces"] = sorted({n.get("typeface") for n in node.iter()
                                                    if n.get("typeface")})
            record["explicit_font_sizes_pt"] = sorted({int(n.get("sz")) / 100 for n in node.iter()
                                                       if local(n.tag) in {"rPr", "defRPr", "endParaRPr"}
                                                       and n.get("sz", "").isdigit()})
            record["text_body_properties"] = [dict(n.attrib) for n in node.findall(".//a:bodyPr", NS)]
            record["graphic_types"] = [n.get("uri") for n in node.findall(".//a:graphicData", NS)]
            record["ole_program_ids"] = sorted({n.get("progId", "") for n in node.iter()
                                                if local(n.tag) == "oleObj"})
            record["relationship_ids"] = sorted({v for n in node.iter() for k, v in n.attrib.items()
                                                 if k.startswith("{" + NS["r"] + "}")})
            record["table_cells"] = [text(n) for n in node.findall(".//a:tc", NS)]
        result.append(record)
    return result


def flatten(items):
    for item in items:
        yield item
        yield from flatten(item.get("children", []))


def part_inventory(zf, part):
    root = ET.fromstring(zf.read(part))
    tree = root.find("p:cSld/p:spTree", NS)
    items = objects(tree) if tree is not None else []
    flat = list(flatten(items))
    title = next((n.get("text", "") for n in flat if n.get("placeholder")
                  and n["placeholder"].get("type") in {"title", "ctrTitle"}), "")
    common = root.find("p:cSld", NS)
    return {
        "part": part,
        "name": common.get("name", "") if common is not None else "",
        "attributes": dict(root.attrib),
        "alternate_content_count": sum(local(n.tag) == "AlternateContent" for n in root.iter()),
        "title_candidate": title,
        "text_including_tables": text(root),
        "counts": dict(Counter(n["kind"] for n in flat)),
        "explicit_font_faces": sorted({face for n in flat for face in n.get("explicit_font_faces", [])}),
        "objects": items,
        "relationships": relationships(zf, part),
        "review_status": "structural-inventory-only",
    }


def inventory(path):
    with ZipFile(path) as zf:
        prespart = "ppt/presentation.xml"
        pres = ET.fromstring(zf.read(prespart))
        rels = {r["Id"]: r for r in relationships(zf, prespart)}
        slides = []
        for number, node in enumerate(pres.findall("p:sldIdLst/p:sldId", NS), 1):
            entry = part_inventory(zf, rels[node.get("{" + NS["r"] + "}id")]["resolved_part"])
            entry.update(slide_number=number, slide_id=node.get("id"),
                         inventory_id=f"{path.stem}:slide:{node.get('id')}")
            slides.append(entry)
        layouts = [part_inventory(zf, p) for p in sorted(zf.namelist())
                   if p.startswith("ppt/slideLayouts/") and p.endswith(".xml") and "/_rels/" not in p]
        masters = [part_inventory(zf, p) for p in sorted(zf.namelist())
                   if p.startswith("ppt/slideMasters/") and p.endswith(".xml") and "/_rels/" not in p]
        size = pres.find("p:sldSz", NS)
        return {
            "schema_version": 1,
            "source_file": path.name,
            "source_sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "slide_size_emu": dict(size.attrib) if size is not None else None,
            "limitations": ["No rendered review", "No inherited style resolution",
                            "Transforms are local; group transforms are not flattened",
                            "Observed text lengths are not capacity limits",
                            "IDs identify source objects, not persistent semantic bindings"],
            "slide_count": len(slides), "native_layout_count": len(layouts),
            "master_count": len(masters),
            "package_parts": zf.namelist(),
            "slides": slides, "native_layouts": layouts, "masters": masters,
        }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    data = inventory(args.source)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("x") as stream:
        json.dump(data, stream, indent=2, ensure_ascii=False)
        stream.write("\n")
    print(json.dumps({k: data[k] for k in ("source_file", "slide_count", "native_layout_count", "master_count")}))


if __name__ == "__main__":
    main()
