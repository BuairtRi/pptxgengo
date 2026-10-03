#!/usr/bin/env python3
"""Check slide template files (templates/library/*.json) against the system's rules.

Checks: known node types, lattice alignment of box edges, content inside the
content area and above the body bottom, word budget, one mark per slide, and the
required chrome (legal line and whiteboard grid).

Usage: python3 tools/check_templates.py [files...]   (default: every file in templates/library)
"""
import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
SRC = (ROOT / "explorations/components.src.html").read_text()
KNOWN = set(re.findall(r'case "([a-z]+)":', SRC))
BOXED = {"thumbnail", "card", "block", "surface", "container", "imageframe", "square", "chevron", "textarrow", "node", "frame", "layerrow", "role", "plane", "cylinder"}
FIVE_UP = [57 + 172.8 * k for k in range(5)]


def on_lattice(x):
    return abs(((x - 3) / 18) - round((x - 3) / 18)) < 1e-6


def on_five_up(x):
    """On the lattice measured from a 5-up column start (the start itself, or content padded inside it)."""
    return any(abs(((x - s) / 18) - round((x - s) / 18)) < 1e-6 and x >= s - 1e-6 for s in FIVE_UP)


MARK_EMPH = {"highlight", "underscore", "circle", "spark"}


def words(obj):
    n = 0
    if isinstance(obj, dict):
        for k, v in obj.items():
            if k in ("type", "style", "ink", "surface", "id", "src", "photo", "focus", "kind", "icon", "rule", "status", "mark", "emphasis", "variant", "layout", "preset", "header", "on", "rail", "footer", "railSurface", "density", "k", "labelPos", "placement", "arrow"):
                continue
            n += words(v)
    elif isinstance(obj, list):
        n += sum(words(x) for x in obj)
    elif isinstance(obj, str):
        n += len(re.sub(r"\[\[|\]\]|\[\^\d+\]", "", obj).split())
    return n


def check(t):
    issues, slide = [], t.get("slide", {})
    if slide.get("type") != "slide":
        issues.append("template must have a 'slide' node")
        return issues
    footer = slide.get("footer", "compact")
    bottom = 450 if footer == "tall" else 468
    src = slide.get("source") or {}
    lines = (1 if src.get("notes") else 0) + (1 if src.get("text") else 0)
    if lines:
        bottom = {1: 450, 2: 432}[min(lines, 2)]
    if t.get("status", "active") not in ("active", "deprecated"):
        issues.append(f"unknown status {t.get('status')}")
    if t.get("status") == "deprecated" and not t.get("replacedBy"):
        issues.append("a deprecated template names its replacement in replacedBy")
    if t.get("rev", 1) > 1 and not t.get("revised"):
        issues.append("a revised template (rev > 1) records its revised date")
    FR = json.loads((ROOT / "frames/v0/frames.json").read_text())
    if slide.get("split") and (slide["split"] not in FR.get("splits", {}) or slide["split"].startswith("_") or slide["split"] == "titleZone"):
        issues.append(f"unknown split frame {slide['split']}")
    if slide.get("split") and slide.get("rail", "none") not in ("none", "nav"):
        issues.append("split frames combine only with no rail or the nav rail")
    if slide.get("rail") == "nav" and not (slide.get("nav") or {}).get("items"):
        issues.append("a nav-rail slide names its sections in nav.items (and nav.active)")
    if slide.get("noLegal"):
        issues.append("noLegal is not allowed: the legal line is required on every slide")
    if slide.get("whiteboard") in (False, None, [], "none") and "whiteboard" in slide:
        issues.append("whiteboard is required on every slide (use the header default or name fields)")
    if any(n.get("type") == "whiteboard" for n in slide.get("body", [])):
        issues.append("put whiteboard fields on the slide's whiteboard property, not in the body")
    marks = 1 if slide.get("emphasis") else 0
    five_bands, twelve_bands = [], []

    def five_only(x):
        return on_five_up(x) and not on_lattice(x)
    for i, n in enumerate(slide.get("body", [])):
        tag = f"body[{i}] {n.get('type')}"
        if n.get("type") not in KNOWN:
            issues.append(f"{tag}: unknown node type")
        if n.get("type") == "mark" or n.get("type") == "annotation" or n.get("emphasis") in MARK_EMPH or n.get("circle"):
            marks += 1
        if n.get("type") in BOXED and all(k in n for k in ("x", "y")):
            x, y = n["x"], n["y"]
            w = n.get("w", n.get("size", 0))
            h = n.get("h", n.get("size", 0))
            full = x == 0 and y == 0 and w == 960
            bleed_x = x == 0  # a panel or photo may bleed off the slide's left edge
            if five_only(x) and slide.get("rail") in ("left", "right", "nav"):
                issues.append(f"{tag}: the 5-up grid is for full-width slides only")
            if five_only(x):
                five_bands.append((y, y + h))
            elif not full and not bleed_x and x not in (57,) and on_lattice(x):
                twelve_bands.append((y, y + h, x))
            if not full and ((not on_lattice(x) and not on_five_up(x) and not bleed_x) or y % 18):
                issues.append(f"{tag}: edge ({x},{y}) is off the 18 pt lattice")
            if not full and n.get("type") not in ("square", "imageframe") and h and y + h > bottom + 0.5 and not slide.get("rail") in ("left", "right"):
                issues.append(f"{tag}: bottom {y + h} is below the body bottom {bottom}")
            if not full and (x < 0 or x + w > 960):
                issues.append(f"{tag}: outside the slide")
    for (a0, a1) in five_bands:
        for (b0, b1, bx) in twelve_bands:
            if a0 < b1 and b0 < a1 and bx + 1e-6 < 903:
                issues.append(f"5-up band y {a0}–{a1} overlaps 12-column inner edges (x {bx}) at y {b0}–{b1}; separate the bands")
                break
        else:
            continue
        break
    budget = t.get("budget", {})
    wc = words(slide)
    if budget.get("words") and wc > budget["words"]:
        issues.append(f"{wc} words, over the budget of {budget['words']}")
    if marks > max(1, budget.get("marks", 1)):
        issues.append(f"{marks} marks, over the limit")
    return issues


def main():
    files = [pathlib.Path(a) for a in sys.argv[1:]] or sorted((ROOT / "templates/library").glob("[!_]*.json"))
    bad = 0
    for f in files:
        data = json.loads(f.read_text())
        for t in data.get("templates", []):
            issues = check(t)
            status = "ok" if not issues else f"{len(issues)} issue(s)"
            print(f"{f.name} · {t.get('id')}: {status}")
            for i in issues:
                print(f"    - {i}")
            bad += bool(issues)
    keys = {}
    for f in files:
        for t in json.loads(f.read_text()).get("templates", []):
            k = (t.get("id"), t.get("variant"))
            if k in keys:
                print(f"duplicate template key {k[0]}/{k[1]} in {f.name} and {keys[k]}")
                bad += 1
            keys[k] = f.name
    sys.exit(1 if bad else 0)


if __name__ == "__main__":
    main()
