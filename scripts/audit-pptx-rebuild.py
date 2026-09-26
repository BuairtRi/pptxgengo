#!/usr/bin/env python3
"""Structural OOXML audit for selected source/rebuilt PowerPoint slides.

This compares editable package structure and resource bytes. It does not measure
rendering or imply pixel/raster equivalence.
"""
from __future__ import annotations

import argparse
import collections
import copy
import hashlib
import json
import re
import sys
import zipfile
import posixpath
from pathlib import Path, PurePosixPath
from typing import Any
from xml.etree import ElementTree as ET

P = "http://schemas.openxmlformats.org/presentationml/2006/main"
A = "http://schemas.openxmlformats.org/drawingml/2006/main"
R = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
PKG_REL = "http://schemas.openxmlformats.org/package/2006/relationships"
NS = {"p": P, "a": A, "r": R, "rel": PKG_REL}
SLIDES_DEFAULT = [5, 8, 11, 12, 14, 24, 28, 36, 38, 43, 44, 67]


def q(uri: str, local: str) -> str:
    return f"{{{uri}}}{local}"


def local(tag: str) -> str:
    return tag.split("}", 1)[-1]


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


class Package:
    def __init__(self, path: Path):
        self.path = path
        self.zf = zipfile.ZipFile(path)
        self.names = set(self.zf.namelist())
        self._xml: dict[str, ET.Element] = {}
        self._rels: dict[str, dict[str, dict[str, str]]] = {}
        self._parts = self.presentation_slides()

    def read(self, part: str) -> bytes:
        return self.zf.read(part)

    def xml(self, part: str) -> ET.Element:
        if part not in self._xml:
            self._xml[part] = ET.fromstring(self.read(part))
        return self._xml[part]

    def rel_part(self, source_part: str) -> str:
        p = PurePosixPath(source_part)
        return str(p.parent / "_rels" / (p.name + ".rels"))

    def relationships(self, source_part: str) -> dict[str, dict[str, str]]:
        if source_part in self._rels:
            return self._rels[source_part]
        part = self.rel_part(source_part)
        values: dict[str, dict[str, str]] = {}
        if part in self.names:
            root = self.xml(part)
            for rel in root.findall(q(PKG_REL, "Relationship")):
                values[rel.get("Id", "")] = dict(rel.attrib)
        self._rels[source_part] = values
        return values

    def resolve_target(self, source_part: str, target: str) -> str:
        if target.startswith("/"):
            return target.lstrip("/")
        return posixpath.normpath(str(PurePosixPath(source_part).parent.joinpath(target)))

    def presentation_slides(self) -> list[str]:
        presentation = "ppt/presentation.xml"
        rels = self.relationships(presentation)
        root = self.xml(presentation)
        result = []
        for sid in root.findall(".//" + q(P, "sldId")):
            rid = sid.get(q(R, "id"), "")
            rel = rels.get(rid)
            if rel and rel.get("Target"):
                result.append(self.resolve_target(presentation, rel["Target"]))
        return result

    def slide_part(self, ordinal: int) -> str:
        if ordinal < 1 or ordinal > len(self._parts):
            raise ValueError(f"slide ordinal {ordinal} is out of range for {self.path} ({len(self._parts)} slides)")
        return self._parts[ordinal - 1]

    def rel_target(self, source_part: str, rid: str) -> tuple[str | None, dict[str, str] | None]:
        rel = self.relationships(source_part).get(rid)
        if not rel:
            return None, None
        target = rel.get("Target", "")
        if rel.get("TargetMode") == "External":
            return target, rel
        return self.resolve_target(source_part, target), rel

    def target_identity(self, source_part: str, rid: str) -> str:
        target, rel = self.rel_target(source_part, rid)
        if target is None or rel is None:
            return "MISSING:" + rid
        if rel.get("TargetMode") == "External":
            return "external:" + target
        if target.endswith(".xml") and "/slides/slide" in target:
            return "slide-part:" + PurePosixPath(target).name
        if target in self.names:
            content = self.read(target)
            if "/media/" in target:
                return "sha256:" + sha(content)
        return "part:" + PurePosixPath(target).name


def parse_slide_map(value: str | None, slides: list[int]) -> dict[int, int]:
    result = {n: n for n in slides}
    if not value:
        return result
    for pair in value.split(","):
        try:
            a, b = pair.split(":", 1)
            result[int(a)] = int(b)
        except Exception as exc:
            raise ValueError(f"invalid --slide-map entry {pair!r}; expected source:target") from exc
    return result


def shape_tree(root: ET.Element) -> ET.Element | None:
    return root.find(".//" + q(P, "spTree"))


def normalize_tree(tree: ET.Element, pkg: Package, slide_part: str) -> str:
    root = copy.deepcopy(tree)
    rels = pkg.relationships(slide_part)
    # Relationship IDs are package-local. Replace them with stable semantic targets.
    for node in root.iter():
        for attr, value in list(node.attrib.items()):
            if attr.startswith("{" + R + "}"):
                rel = rels.get(value)
                if rel:
                    node.set(attr, pkg.target_identity(slide_part, value))
                else:
                    node.set(attr, "MISSING:" + value)
        # PPT field metadata and an equivalent ordinary run serialize differently.
        # Normalize a:fld to a:r and discard generated field id/type only.
        if node.tag == q(A, "fld") and node.get("type") == "slidenum":
            node.tag = q(A, "r")
            node.attrib.clear()
            # The freeze step removes field-only paragraph properties because
            # an ordinary run cannot contain a:pPr.
            for child in list(node):
                if child.tag == q(A, "pPr"):
                    node.remove(child)
    return ET.tostring(root, encoding="unicode", short_empty_elements=True)


def shape_identity(el: ET.Element) -> str:
    props = el.find(".//" + q(P, "cNvPr"))
    name = props.get("name", "") if props is not None else ""
    return f"{local(el.tag)}::{name}"


def text_info(shape_tree_el: ET.Element) -> dict[str, Any]:
    text_shapes = 0
    text_nodes = 0
    characters = 0
    paragraph_count = 0
    spans = 0
    direct_size = direct_typeface = 0
    para_size = para_typeface = 0
    inherited_size = inherited_typeface = 0
    payload: list[str] = []
    for shape in shape_tree_el.iter():
        txbody = shape.find(q(P, "txBody"))
        if txbody is None:
            continue
        text_shapes += 1
        paragraph_count += len(txbody.findall(".//" + q(A, "p")))
        for text in txbody.findall(".//" + q(A, "t")):
            text_nodes += 1
            characters += len(text.text or "")
            payload.append(text.text or "")
        for span in list(txbody.findall(".//" + q(A, "r"))) + list(txbody.findall(".//" + q(A, "fld"))):
            if span.find(q(A, "t")) is None:
                continue
            spans += 1
            rpr = span.find(q(A, "rPr"))
            has_direct_size = rpr is not None and "sz" in rpr.attrib
            has_direct_typeface = rpr is not None and rpr.find(q(A, "latin")) is not None and bool(rpr.find(q(A, "latin")).get("typeface"))
            if has_direct_size:
                direct_size += 1
            if has_direct_typeface:
                direct_typeface += 1
            para = next((p for p in span.iterancestors() if p.tag == q(A, "p")), None) if hasattr(span, "iterancestors") else None
            # ElementTree has no iterancestors; find its containing paragraph explicitly.
            if para is None:
                for p in txbody.findall(".//" + q(A, "p")):
                    if span in list(p.iter()):
                        para = p
                        break
            defpr = para.find(q(A, "pPr") + "/" + q(A, "defRPr")) if para is not None else None
            has_para_size = defpr is not None and "sz" in defpr.attrib
            has_para_typeface = defpr is not None and defpr.find(q(A, "latin")) is not None and bool(defpr.find(q(A, "latin")).get("typeface"))
            if has_para_size:
                para_size += 1
            if has_para_typeface:
                para_typeface += 1
            if not has_direct_size and not has_para_size:
                inherited_size += 1
            if not has_direct_typeface and not has_para_typeface:
                inherited_typeface += 1
    return {
        "text_shapes": text_shapes, "text_nodes": text_nodes, "characters": characters,
        "text_payload_sha256": sha("\u241e".join(payload).encode("utf-8")),
        "paragraphs": paragraph_count, "text_spans": spans,
        "typography": {
            "run_direct_font_size": direct_size,
            "run_size_from_paragraph_default": para_size,
            "run_size_inherited_above_paragraph": inherited_size,
            "run_direct_typeface": direct_typeface,
            "run_typeface_from_paragraph_default": para_typeface,
            "run_typeface_inherited_above_paragraph": inherited_typeface,
            "note": "Inheritance counts indicate absent direct/paragraph-default attributes; the script does not resolve placeholder, layout, master, or theme typography."
        }
    }


def slide_summary(pkg: Package, ordinal: int) -> dict[str, Any]:
    part = pkg.slide_part(ordinal)
    root = pkg.xml(part)
    tree = shape_tree(root)
    if tree is None:
        raise ValueError(f"no p:spTree in {part}")
    types = collections.Counter(local(el.tag) for el in tree.iter())
    identities = collections.Counter(shape_identity(el) for el in tree.iter() if local(el.tag) in {"sp", "pic", "graphicFrame", "grpSp", "cxnSp"})
    text = text_info(tree)
    rels = pkg.relationships(part)
    image_targets = []
    rel_types = collections.Counter()
    for rid, rel in rels.items():
        typ = rel.get("Type", "").rsplit("/", 1)[-1]
        rel_types[typ] += 1
        if typ in {"image", "hdphoto"}:
            target, _ = pkg.rel_target(part, rid)
            if target and target in pkg.names:
                image_targets.append({"target": target, "sha256": sha(pkg.read(target))})
    layout_part = related_part(pkg, part, "slideLayout")
    master_part = related_part(pkg, layout_part, "slideMaster") if layout_part else None
    theme_part = related_part(pkg, master_part, "theme") if master_part else None
    resources = {}
    for role, resource in (("layout", layout_part), ("master", master_part), ("theme", theme_part)):
        resources[role] = {"part": resource, "sha256": sha(pkg.read(resource)) if resource and resource in pkg.names else None}
    return {
        "slide_ordinal": ordinal, "part": part,
        "hidden_show_0": root.get("show") == "0",
        "counts": {
            "shapes": types["sp"], "pictures": types["pic"], "tables": types["tbl"],
            "groups": types["grpSp"], "connectors": types["cxnSp"],
            "graphic_frames": types["graphicFrame"], "media_relationships": len(image_targets),
            "media_hashes": [i["sha256"] for i in image_targets],
            "native_components": {k: types[k] for k in ("sp", "pic", "graphicFrame", "tbl", "grpSp", "cxnSp", "oleObj", "chart") if types[k]},
        },
        "shape_bindings": identities,
        "text": text,
        "relationship_types": dict(rel_types),
        "inherited_resources": resources,
        "shape_tree_xml": normalize_tree(tree, pkg, part),
    }


def related_part(pkg: Package, source_part: str, rel_suffix: str) -> str | None:
    if not source_part:
        return None
    for rid, rel in pkg.relationships(source_part).items():
        if rel.get("Type", "").endswith("/" + rel_suffix):
            target, _ = pkg.rel_target(source_part, rid)
            return target
    return None


def compare_slide(src: dict[str, Any], dst: dict[str, Any]) -> dict[str, Any]:
    names_src, names_dst = src["shape_bindings"], dst["shape_bindings"]
    matched = sum((names_src & names_dst).values())
    count = sum(names_src.values())
    sxml, dxml = src.pop("shape_tree_xml"), dst.pop("shape_tree_xml")
    normalized_exact = sxml == dxml
    # Text payload remains useful even when object shape trees differ.
    result = {
        "source_slide": src["slide_ordinal"], "target_slide": dst["slide_ordinal"],
        "shape_tree_xml_exact_after_normalization": normalized_exact,
        "normalized_shape_tree_sha256": {"source": sha(sxml.encode()), "target": sha(dxml.encode())},
        "source_object_name_type_coverage": {
            "source_items": count, "matched_items": matched,
            "fraction": matched / count if count else 1.0,
            "unmatched_source": dict(names_src - names_dst),
            "unmatched_target": dict(names_dst - names_src),
        },
        "source": src, "target": dst,
        "media_sha256_exact_match_count": len(set(src["counts"]["media_hashes"]) & set(dst["counts"]["media_hashes"])),
        "media_sha256_source_count": len(src["counts"]["media_hashes"]),
        "media_sha256_target_count": len(dst["counts"]["media_hashes"]),
        "text_payload_exact": src["text"]["text_payload_sha256"] == dst["text"]["text_payload_sha256"],
        "typography_count_equality": src["text"]["typography"] == dst["text"]["typography"],
    }
    return result


def canonical_packages(source: Package, target: Package, slides: list[int], slide_map: dict[int, int]) -> dict[str, Any]:
    slide_reports = []
    for slide in slides:
        src = slide_summary(source, slide)
        dst = slide_summary(target, slide_map[slide])
        slide_reports.append(compare_slide(src, dst))
    # Resource bytes are compared by role for each selected slide and by media hash.
    roles = {}
    media_src, media_dst = set(), set()
    for sr in slide_reports:
        for side, accum in (("source", media_src), ("target", media_dst)):
            accum.update(sr[side]["counts"]["media_hashes"])
            for role, value in sr[side]["inherited_resources"].items():
                key = f"slide-{sr['source_slide']}:{role}"
                roles.setdefault(key, {})[side] = value
    role_comparison = {}
    for key, sides in roles.items():
        s, d = sides.get("source"), sides.get("target")
        role_comparison[key] = {"source": s, "target": d, "exact_bytes": bool(s and d and s.get("sha256") == d.get("sha256"))}
    return {
        "source_package": str(source.path), "target_package": str(target.path),
        "selected_source_slides": slides, "source_to_target_slide_map": slide_map,
        "presentation_slide_counts": {"source": len(source._parts), "target": len(target._parts), "target_unselected_ordinals": [i for i in range(1, len(target._parts) + 1) if i not in set(slide_map.values())]},
        "algorithm": {
            "normalization": ["relationship IDs are replaced by semantic target identities/content hashes; internal slide links use the referenced slide-part filename", "slidenum a:fld is normalized to an ordinary a:r; field-only attributes and child a:pPr are removed as in the rebuild freeze step", "p:sld show=0 is recorded per slide; slide-root presentation attributes are outside the compared p:spTree"],
            "limits": "Structural OOXML and resource-byte audit only. It does not establish raster, font-rendering, placement, clipping, or pixel fidelity. Typography counts report direct and paragraph-default properties; placeholder/layout/master/theme cascade is not resolved.",
        },
        "resource_byte_comparison": {
            "selected_media_unique_sha256_source": len(media_src), "selected_media_unique_sha256_target": len(media_dst),
            "selected_media_exact_sha256_intersection": len(media_src & media_dst),
            "layout_master_theme_by_slide": role_comparison,
        },
        "slides": slide_reports,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, required=True, help="reference/source PPTX")
    parser.add_argument("--target", type=Path, required=True, help="compiled rebuilt PPTX")
    parser.add_argument("--out", type=Path, required=True, help="JSON report path")
    parser.add_argument("--slides", default=",".join(map(str, SLIDES_DEFAULT)), help="source slide ordinals, comma-separated")
    parser.add_argument("--slide-map", help="optional source:target ordinal pairs, comma-separated (default identity)")
    args = parser.parse_args()
    try:
        slides = [int(x) for x in args.slides.split(",") if x]
        if not slides:
            raise ValueError("--slides must include at least one slide")
        slide_map = parse_slide_map(args.slide_map, slides)
        source, target = Package(args.source), Package(args.target)
        report = canonical_packages(source, target, slides, slide_map)
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        print(f"wrote structural audit for {len(slides)} slide(s) to {args.out}")
        return 0
    except Exception as exc:
        print(f"audit-pptx-rebuild: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
