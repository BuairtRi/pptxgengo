#!/usr/bin/env python3
"""Compare native text bounds with fixed native shape frames.

This reports measured geometric containment only. It does not estimate copy
capacity, prove raster ink containment, or approve a template for adaptation.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import math
import posixpath
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path


def rect_from(row: dict) -> tuple[float, float, float, float, float, float, float, float, float]:
    """Return shape rect, rotation and margins; tolerate legacy frame reports."""
    sf = row.get("shape_frame") or {}
    left = sf.get("left", row.get("shape_left"))
    top = sf.get("top", row.get("shape_top"))
    width = sf.get("width", row.get("frame_width"))
    height = sf.get("height", row.get("frame_height"))
    rotation = row.get("rotation", sf.get("rotation_degrees"))
    margins = row.get("margins")
    if margins is None:
        margins = {"left": row.get("margin_left"), "right": row.get("margin_right"),
                   "top": row.get("margin_top"), "bottom": row.get("margin_bottom")}
    ml, mr, mt, mb = (margins.get(k) for k in ("left", "right", "top", "bottom"))
    vals = (left, top, width, height, rotation, ml, mr, mt, mb)
    if any(v is None or not isinstance(v, (int, float)) or not math.isfinite(v) for v in vals):
        raise ValueError("shape frame, rotation, or margins are missing/non-finite")
    if width <= 0 or height <= 0 or min(ml, mr, mt, mb) < 0 or ml + mr >= width or mt + mb >= height:
        raise ValueError("invalid shape frame or inner margins")
    return float(left), float(top), float(width), float(height), float(rotation), float(ml), float(mr), float(mt), float(mb)


def text_rect(row: dict) -> tuple[float, float, float, float]:
    b = row.get("text_bounds")
    if isinstance(b, dict):
        vals = (b.get("left"), b.get("top"), b.get("width"), b.get("height"))
    else:
        vals = (row.get("text_left"), row.get("text_top"), row.get("text_width"), row.get("text_height"))
    if any(v is None or not isinstance(v, (int, float)) or not math.isfinite(v) for v in vals):
        raise ValueError("text bounds are missing/non-finite")
    if vals[2] < 0 or vals[3] < 0:
        raise ValueError("negative text bounds")
    return tuple(float(v) for v in vals)  # type: ignore[return-value]


def inside(row: dict, tolerance: float) -> tuple[bool, dict]:
    x, y, w, h, angle, ml, mr, mt, mb = rect_from(row)
    tx, ty, tw, th = text_rect(row)
    # AppleScript text-range bounds are axis-aligned in slide coordinates.
    # Inverse-rotate their four corners into the shape's local frame. Using
    # the bounding rectangle is conservative for rotated text.
    cx, cy = x + w / 2, y + h / 2
    a = math.radians(-angle)
    ca, sa = math.cos(a), math.sin(a)
    corners = []
    for px, py in ((tx, ty), (tx + tw, ty), (tx, ty + th), (tx + tw, ty + th)):
        dx, dy = px - cx, py - cy
        corners.append((cx + dx * ca - dy * sa, cy + dx * sa + dy * ca))
    min_x, max_x = min(p[0] for p in corners), max(p[0] for p in corners)
    min_y, max_y = min(p[1] for p in corners), max(p[1] for p in corners)
    frame_inner = {"left": x + ml, "top": y + mt, "right": x + w - mr, "bottom": y + h - mb}
    text_local = {"left": min_x, "top": min_y, "right": max_x, "bottom": max_y}
    overflow = {
        "left": max(0.0, frame_inner["left"] - min_x),
        "top": max(0.0, frame_inner["top"] - min_y),
        "right": max(0.0, max_x - frame_inner["right"]),
        "bottom": max(0.0, max_y - frame_inner["bottom"]),
    }
    ok = all(v <= tolerance for v in overflow.values())
    return ok, {"inner_frame": frame_inner, "text_bounds_local": text_local, "overflow": overflow,
                "rotation_degrees": angle, "tolerance": tolerance}


def rows_from(doc: dict) -> list[dict]:
    for key in ("shapes", "measurements"):
        if isinstance(doc.get(key), list):
            return doc[key]
    raise ValueError("input needs a `shapes` or `measurements` array")


def native_top_level(deck_path: str) -> dict[tuple[str, str], dict]:
    """Read top-level OOXML shape transforms to pin frame identity."""
    ns = {"p": "http://schemas.openxmlformats.org/presentationml/2006/main",
          "a": "http://schemas.openxmlformats.org/drawingml/2006/main"}
    result: dict[tuple[str, str], dict] = {}
    with zipfile.ZipFile(deck_path) as zf:
        rel_root = ET.fromstring(zf.read("ppt/_rels/presentation.xml.rels"))
        rels = {r.get("Id"): r.get("Target") for r in list(rel_root)}
        pres = ET.fromstring(zf.read("ppt/presentation.xml"))
        sld_ids = pres.find("p:sldIdLst", ns)
        if sld_ids is None:
            raise ValueError("candidate PPTX has no presentation slide order")
        slides = []
        for sld in list(sld_ids):
            rid = sld.get("{http://schemas.openxmlformats.org/officeDocument/2006/relationships}id")
            target = rels.get(rid)
            if not target:
                raise ValueError("presentation slide relationship is missing")
            slides.append(posixpath.normpath(target.lstrip("/") if target.startswith("/") else posixpath.join("ppt", target)))
        drawable = {"sp", "grpSp", "pic", "graphicFrame", "cxnSp"}
        for slide_no, part in enumerate(slides, 1):
            root = ET.fromstring(zf.read(part))
            tree = root.find(".//p:spTree", ns)
            if tree is None:
                continue
            top = [n for n in list(tree) if n.tag.rsplit("}", 1)[-1] in drawable]
            for shape_no, node in enumerate(top, 1):
                kind = node.tag.rsplit("}", 1)[-1]
                nv = None
                for path in ("p:nvSpPr/p:cNvPr", "p:nvGrpSpPr/p:cNvPr", "p:nvPicPr/p:cNvPr", "p:nvGraphicFramePr/p:cNvPr", "p:nvCxnSpPr/p:cNvPr"):
                    nv = node.find(path, ns)
                    if nv is not None:
                        break
                if nv is None:
                    continue
                name = nv.get("name", "")
                xfrm = node.find("p:spPr/a:xfrm", ns)
                if xfrm is None:
                    xfrm = node.find("p:grpSpPr/a:xfrm", ns)
                if xfrm is None:
                    xfrm = node.find("p:xfrm", ns)
                off = xfrm.find("a:off", ns) if xfrm is not None else None
                ext = xfrm.find("a:ext", ns) if xfrm is not None else None
                if off is None or ext is None:
                    continue
                rot = float(xfrm.get("rot", "0")) / 60000.0
                paras = node.findall("p:txBody/a:p", ns)
                para_text = []
                for para in paras:
                    chunks = []
                    for child in list(para):
                        kind_name = child.tag.rsplit("}", 1)[-1]
                        if kind_name == "br":
                            chunks.append("\n")
                        elif kind_name in ("r", "fld"):
                            tnode = child.find("a:t", ns)
                            if tnode is not None and tnode.text:
                                chunks.append(tnode.text)
                    para_text.append("".join(chunks))
                full_text = "\n".join(para_text).rstrip("\r\n")
                result[(str(slide_no), str(shape_no))] = {"name": name, "kind": kind,
                    "emu": tuple(float(off.get(k)) for k in ("x", "y")) + tuple(float(ext.get(k)) for k in ("cx", "cy")),
                    "rotation": rot, "text": full_text}
    return result


def key(row: dict) -> tuple[str, str]:
    return str(row.get("slide", row.get("slide_index", ""))), str(row.get("shape_path", row.get("shape_name", row.get("name", ""))))


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("report", help="JSON emitted by measure-template-frames.applescript")
    ap.add_argument("--scene", help="pinned native frame report from the untouched source deck")
    ap.add_argument("--native-deck", help="candidate PPTX; verifies top-level shape order/name/frame from OOXML")
    ap.add_argument("--tolerance", type=float, default=0.15, help="containment tolerance in report coordinate units (default: 0.15)")
    ap.add_argument("--strict", action="store_true", help="return nonzero for definitive overflow as well as malformed evidence")
    ap.add_argument("--out", help="write JSON result here (default: stdout)")
    args = ap.parse_args()
    if args.tolerance < 0 or not math.isfinite(args.tolerance):
        ap.error("--tolerance must be finite and nonnegative")
    try:
        report = json.loads(Path(args.report).read_text())
        rows = rows_from(report)
        scene = json.loads(Path(args.scene).read_text()) if args.scene else None
        reference = rows_from(scene) if scene else None
        ooxml = native_top_level(args.native_deck) if args.native_deck else None
    except (OSError, json.JSONDecodeError, ValueError, TypeError, KeyError, zipfile.BadZipFile, ET.ParseError) as e:
        failure = json.dumps({"schema": "pptxgengo.template-fit-check.v1", "status": "invalid_evidence", "error": str(e)}, indent=2) + "\n"
        if args.out:
            Path(args.out).write_text(failure)
        else:
            sys.stdout.write(failure)
        return 2

    refmap = {key(r): r for r in reference or []}
    scene_identity_ok = True
    if scene is not None:
        source_id = report.get("source_scene_sha256")
        pinned_id = scene.get("source_scene_sha256")
        scene_identity_ok = bool(source_id and pinned_id and source_id == pinned_id)
    findings = []
    overflow_count = inconclusive_count = 0
    for row in rows:
        identity = {"slide": row.get("slide", row.get("slide_index")), "shape_path": row.get("shape_path"), "name": row.get("name", row.get("shape_name"))}
        if row.get("error"):
            findings.append({**identity, "status": "inconclusive", "reason": "native measurement error preserved", "measurement_error": row["error"]})
            inconclusive_count += 1
            continue
        if row.get("text") in (None, ""):
            continue
        baseline = refmap.get(key(row)) if reference is not None else None
        if reference is not None and not scene_identity_ok:
            findings.append({**identity, "status": "inconclusive", "reason": "measured and pinned source scene hashes are missing or differ"})
            inconclusive_count += 1
            continue
        if reference is not None and baseline is None:
            findings.append({**identity, "status": "inconclusive", "reason": "shape path missing from pinned source frame report"})
            inconclusive_count += 1
            continue
        nested = row.get("coordinate_scope") in ("group_local", "table_cell", "nested_unknown") or "." in str(row.get("shape_path", "")) or ".r" in str(row.get("shape_path", ""))
        if nested and (baseline is None or not (row.get("transform_chain_verified") is True and baseline.get("transform_chain_verified") is True)):
            findings.append({**identity, "status": "inconclusive", "reason": "group/table frame lacks a pinned, verified transform chain"})
            inconclusive_count += 1
            continue
        try:
            if ooxml is not None:
                path = str(row.get("shape_path", ""))
                if not path.isdigit():
                    raise ValueError("native-deck verification currently supports top-level shapes only")
                shape = ooxml.get((str(identity["slide"]), path))
                if shape is None:
                    raise ValueError("shape path is absent from candidate OOXML")
                if shape["name"] != identity.get("name"):
                    raise ValueError("native shape name does not match report identity")
                normalize = lambda s: str(s or "").replace("\r\n", "\n").replace("\r", "\n").replace("\x0b", "\n").rstrip("\n")
                if normalize(shape["text"]) != normalize(row.get("text")):
                    raise ValueError("candidate OOXML text does not match native measurement text")
                x, y, w, h = shape["emu"]
                reported = row.get("shape_frame") or {"left": row.get("shape_left"), "top": row.get("shape_top"), "width": row.get("frame_width"), "height": row.get("frame_height")}
                # AppleScript PowerPoint coordinates are normally points; use
                # the native deck to identify the conversion, not to guess it.
                scales = (1.0, 1.0 / 12700.0)
                matched = [scale for scale in scales if all(abs(float(reported.get(k, reported.get(alias))) - raw * scale) <= 0.25 for k, alias, raw in zip(("left", "top", "width", "height"), ("shape_left", "shape_top", "frame_width", "frame_height"), (x, y, w, h)))]
                if len(matched) != 1:
                    raise ValueError("candidate OOXML frame does not uniquely match reported native frame/units")
                scale = matched[0]
                reported_rotation = row.get("rotation", (row.get("shape_frame") or {}).get("rotation_degrees"))
                if reported_rotation is not None and abs(float(reported_rotation) - shape["rotation"]) > 0.001:
                    raise ValueError("candidate OOXML rotation differs from native report")
                row = dict(row)
                row["rotation"] = shape["rotation"]
                row["coordinate_scope"] = "slide"
                # Normalize the measured frame into OOXML-proven coordinates.
                row["shape_frame"] = {"left": x * scale, "top": y * scale, "width": w * scale,
                                      "height": h * scale, "rotation_degrees": shape["rotation"]}
                row["shape_left"] = None
                row["shape_top"] = None
                row["frame_width"] = None
                row["frame_height"] = None
            if row.get("rotation") is None and (row.get("shape_frame") or {}).get("rotation_degrees") is None:
                raise ValueError("rotation is missing; cannot assume zero")
            margins = row.get("margins")
            if margins is None and not all(k in row for k in ("margin_left", "margin_right", "margin_top", "margin_bottom")):
                raise ValueError("one or more text margins are missing; cannot assume zero")
            if baseline is not None:
                # A pinned native snapshot resolves identity only when object
                # path, label, geometry, rotation, and margins remain exact.
                for field in ("name", "coordinate_scope", "rotation", "frame_width", "frame_height", "inner_width", "inner_height", "shape_left", "shape_top", "margins", "shape_frame"):
                    if row.get(field) is not None and baseline.get(field) is not None and row.get(field) != baseline.get(field):
                        raise ValueError(f"pinned source geometry changed: {field}")
            fits, metrics = inside(row, args.tolerance)
            status = "fits" if fits else "overflow"
            findings.append({**identity, "status": status, "metrics": metrics})
            if not fits:
                overflow_count += 1
        except (ValueError, TypeError, KeyError) as e:
            findings.append({**identity, "status": "inconclusive", "reason": str(e)})
            inconclusive_count += 1
    definitive = sum(f["status"] in ("fits", "overflow") for f in findings)
    status = "overflow" if overflow_count else ("inconclusive" if inconclusive_count or not definitive else "fits")
    result = {"schema": "pptxgengo.template-fit-check.v1", "status": status,
              "qualification": "measurement_only", "source_presentation": report.get("presentation"),
              "source_scene_sha256": report.get("source_scene_sha256"), "checked_text_shapes": definitive,
              "overflow_count": overflow_count, "inconclusive_count": inconclusive_count,
              "tolerance": args.tolerance,
              "scope": "Native text-bound containment in the inner shape frame; not optical ink, collision, or copy capacity approval.",
              "pinned_scene_report": args.scene, "findings": findings}
    result["candidate_native_deck"] = args.native_deck
    result["native_frame_report_sha256"] = hashlib.sha256(Path(args.report).read_bytes()).hexdigest()
    if args.native_deck:
        result["candidate_native_deck_sha256"] = hashlib.sha256(Path(args.native_deck).read_bytes()).hexdigest()
    payload = json.dumps(result, indent=2, ensure_ascii=False) + "\n"
    if args.out:
        Path(args.out).write_text(payload)
    else:
        sys.stdout.write(payload)
    return 1 if args.strict and overflow_count else 0


if __name__ == "__main__":
    raise SystemExit(main())
