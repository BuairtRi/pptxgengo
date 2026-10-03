#!/usr/bin/env python3
"""Apply the October 2 review corrections to the composed reference specimen.

The pinned WMDS source stays unchanged. The deliverable preview is illustrative
native composition, not default content inserted into a bound template.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PATH = ROOT / "samples/wmds-refresh-slice2-20261002/reference.foundation.json"
doc = json.loads(PATH.read_text())

for slide in doc["slides"]:
    binding = slide.get("template_binding", {})
    template = binding.get("template")
    if template == "divider/inverse":
        slide["library_chrome"]["whiteboard"] = [
            {"x": 561, "y": 54, "w": 342, "h": 396, "fade": "none"}]
        for node in slide["nodes"]:
            scene = node["scene"]
            raw = scene["node"]
            if node["id"] == "node01":
                raw.update(x=507, y=126, size=270)
            elif node["id"] in ("node02", "node03"):
                raw["w"] = 432
                raw["y"] = 216 if node["id"] == "node02" else 234
            resolution = "wmds.inverse-divider-separated-columns.v2"
            if resolution not in scene.setdefault("resolutions", []):
                scene["resolutions"].append(resolution)
    elif template == "deliverables/annotated":
        alternate = slide["id"].startswith("alternate")
        target_days = "6" if alternate else "5"
        slide["nodes"] = [n for n in slide["nodes"] if not n["id"].startswith("preview-")]
        for node in slide["nodes"]:
            raw = node["scene"]["node"]
            if node["id"] == "node01":
                raw["kind"] = "blank"
                node["scene"]["resolutions"] = ["wmds.annotated-scorecard-specimen.v2"]
            elif node["id"] == "z1":
                raw["h"] = 72
            elif node["id"] == "z2":
                raw["h"] = 90

        content = []

        def add(key, raw):
            content.append({"id": "preview-" + key, "kind": "scene",
                            "rect": {"x_pt": 0, "y_pt": 0, "width_pt": 0, "height_pt": 0},
                            "scene": {"node": raw,
                                      "source_pointer": "/specimens/annotated-scorecard/" + key,
                                      "resolutions": ["wmds.annotated-scorecard-specimen.v2"]}})

        def text(key, copy, x, y, w, style="small", ink="primary"):
            add(key, {"type": "text", "x": x, "y": y, "w": w,
                      "style": style, "ink": ink, "text": copy})

        text("summary-label", "Readiness / pilot decision", 87, 154, 354, "label", "emphasis")
        text("score", "76 / 100" if alternate else "72 / 100", 87, 170, 144, "subhead", "display")
        text("decision", "Proceed with a limited pilot", 249, 173, 192, "small", "emphasis")
        text("target", f"Target: close the books in {target_days} days", 87, 194, 354)

        for y, key, headings in [(243, "steps", ("Close step", "Current", "Target")),
                                 (351, "actions", ("Priority action", "Owner", "Due"))]:
            for j, (x, w) in enumerate(((87, 192), (285, 78), (369, 72))):
                text(f"{key}-head-{j}", headings[j], x, y, w, "label", "emphasis")
            rows = [("Ledger reconciliation", "3 days", "2 days"),
                    ("Review and approvals", "2 days", "1 day"),
                    ("Reporting and sign-off", "2 days", "2 days" if target_days == "5" else "3 days")]
            if key == "actions":
                rows = [("Resolve mapping gaps", "Controller", "Oct 09"),
                        ("Confirm exception owners", "Finance", "Oct 12"),
                        ("Rehearse the pilot close", "PMO", "Oct 16")]
            for i, row in enumerate(rows):
                ry = y + 17 + i * 18
                if i % 2 == 0:
                    add(f"{key}-fill-{i}", {"type": "block", "x": 84, "y": ry,
                                            "w": 360, "h": 18, "surface": "subtle"})
                for j, (x, w) in enumerate(((87, 192), (285, 78), (369, 72))):
                    text(f"{key}-row-{i}-{j}", row[j], x, ry, w, "small", "primary")

        # Paint the page first, preview content next, then frames and annotations.
        slide["nodes"][1:1] = content

PATH.write_text(json.dumps(doc, indent=2) + "\n")
