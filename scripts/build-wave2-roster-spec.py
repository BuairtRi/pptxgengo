#!/usr/bin/env python3
"""Append a synthetic 19-person-card roster fixture to the people review spec.

The current people-review.json stays untouched: this writes a sibling 3-slide
spec and a separate one-slide fixed-size long-role stress spec.
"""
from __future__ import annotations

import copy
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BASE = ROOT / "library/visual-components/people-review.json"
OUT = ROOT / "library/visual-components/people-review-plus-roster.json"
STRESS_OUT = ROOT / "library/visual-components/people-role-stress.json"

NAVY = "#070154"
BLUE = "#0047FF"
SLATE = "#50658E"
PALE = "#E8EEF8"
WHITE = "#FFFFFF"


def rect(x, y, w, h):
    return {"x": x, "y": y, "width": w, "height": h}


def file_hash(path):
    return hashlib.sha256((ROOT / path).read_bytes()).hexdigest()


def canvas_text(id, bounds, text, size, color=NAVY, bold=False, *, background="", align="left", valign="top", peers=None):
    obj = {"id": id, "kind": "text", "bounds": bounds, "text": text,
           "font_face": "Arial", "font_size_pt": size, "bold": bold,
           "foreground": color, "align": align, "valign": valign,
           "inset_x": 0, "inset_y": 0}
    if background:
        obj["background"] = background
    if peers:
        obj["allow_overlap"] = peers
    return obj


def footer(page):
    logo = "samples/showcase/assets/wm_h_pos_clr_rgb_august2024.png"
    return [
        {"id": "footer-band", "kind": "surface", "bounds": rect(0, 504, 960, 36), "background": PALE,
         "allow_overlap": ["wm-logo", "footer-copy", "page-number"]},
        {"id": "wm-logo", "kind": "image", "bounds": rect(36, 511, 104, 21.727),
         "asset_path": logo, "asset_sha256": file_hash(logo),
         "alt_text": "West Monroe horizontal positive color logo", "image_fit": "contain",
         "allow_overlap": ["footer-band"]},
        canvas_text("footer-copy", rect(174, 516, 700, 17),
                    "© 2026 West Monroe Partners. Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.",
                    8, NAVY, peers=["footer-band"]),
        canvas_text("page-number", rect(892, 516, 32, 17), str(page), 9, NAVY,
                    align="right", peers=["footer-band"]),
    ]


CORE = [
    ("01", "Delivery sponsor"), ("02", "Program coordination lead"), ("03", "Solution architecture lead"),
    ("04", "Data platform lead"), ("05", "Integration lead"), ("06", "Operations readiness lead"),
    ("07", "Security and controls lead"), ("08", "Test coordination lead"), ("09", "Change coordination lead"),
    ("10", "Data governance lead"),
]
SPECIALISTS = [
    ("11", "Service operations adviser"), ("12", "Workflow design adviser"), ("13", "Information security adviser"),
    ("14", "Records and retention adviser"), ("15", "Integration adviser"), ("16", "Data quality adviser"),
    ("17", "Testing adviser"), ("18", "Accessibility adviser"), ("19", "Support transition adviser"),
]


def tile_items(prefix, number, role, x, y, long_role=False):
    frame_id = f"{prefix}-{number}-card"
    initial_id = f"{prefix}-{number}-initials"
    name_id = f"{prefix}-{number}-name"
    role_id = f"{prefix}-{number}-role"
    return [
        {"id": frame_id, "kind": "surface", "bounds": rect(x, y, 172.8, 36), "background": PALE,
         "allow_overlap": [initial_id, name_id, role_id]},
        canvas_text(initial_id, rect(x, y, 36, 36), number, 11, WHITE, True,
                    background=BLUE, align="center", valign="middle", peers=[frame_id]),
        canvas_text(name_id, rect(x + 43.2, y + 3.6, 125.8, 12), f"Illustrative Person {number}",
                    9, NAVY, True, peers=[frame_id]),
        canvas_text(role_id, rect(x + 43.2, y + 18, 125.8, 14.4), role,
                    9, NAVY, peers=[frame_id]),
    ]


def panel_layout(id, y, height, heading, bullets):
    blocks = [{"id": f"{id}-heading", "text": heading, "font_face": "Arial", "font_size_pt": 11,
               "bold": True, "foreground": NAVY, "align": "left", "min_height_pt": 16}]
    for i, value in enumerate(bullets, 1):
        blocks.append({"id": f"{id}-item-{i}", "text": value, "font_face": "Arial", "font_size_pt": 10,
                       "foreground": NAVY, "marker": "•", "marker_width_pt": 14,
                       "left_inset_pt": 14, "align": "left", "min_height_pt": 18})
    return {
        "id": id, "bounds": rect(590, y, 334, height),
        "columns": [{"fixed_pt": 334}], "rows": [{"fixed_pt": height}],
        "padding": {"top": 0, "right": 0, "bottom": 0, "left": 0},
        "cells": [{"id": f"{id}-content", "row": 0, "column": 0,
                   "background": PALE,
                   "padding": {"top": 9, "right": 10, "bottom": 8, "left": 10},
                   "blocks": blocks}],
    }


def roster_slide(page=3):
    canvas = [
        canvas_text("roster-kicker", rect(36, 105, 700, 13),
                    "ILLUSTRATIVE • PROPOSED ROLES ONLY • NO REAL PEOPLE DEPICTED", 9, SLATE, True),
        canvas_text("core-heading", rect(36, 124, 540, 14), "CORE TEAM • 10 ILLUSTRATIVE TILES", 10, NAVY, True),
        canvas_text("specialist-heading", rect(36, 307, 540, 14), "SPECIALIST ADVISERS • 9 ILLUSTRATIVE TILES", 10, NAVY, True),
    ]
    xs = [36.25, 217.15, 398.8]
    core_ys = [143, 184, 225, 266]
    for i, (number, role) in enumerate(CORE):
        row, col = divmod(i, 3)
        canvas.extend(tile_items("core", number, role, xs[col], core_ys[row]))
    specialist_ys = [326, 367, 408]
    for i, (number, role) in enumerate(SPECIALISTS):
        row, col = divmod(i, 3)
        canvas.extend(tile_items("specialist", number, role, xs[col], specialist_ys[row]))

    layouts = [
        panel_layout("core-responsibilities", 143, 159, "Illustrative core responsibilities", [
            "Agree pilot scope, decision points and accountable owners.",
            "Coordinate architecture, data and integration questions.",
            "Review testing, control and release evidence with stakeholders.",
            "Plan support, user preparation and transition activities.",
        ]),
        panel_layout("specialist-responsibilities", 326, 118, "Illustrative specialist input", [
            "Review workflow and policy assumptions.",
            "Flag security, records and interface dependencies.",
            "Identify evidence needed for discovery and validation.",
        ]),
    ]
    return {
        "id": "illustrative-19-tile-roster",
        "title": "Illustrative roster separates ten core roles from nine specialist inputs",
        "width_pt": 960, "height_pt": 540, "title_bounds": rect(36, 44, 888, 52),
        "title_font_face": "Arial", "title_font_size_pt": 22, "title_bold": True,
        "title_foreground": NAVY, "pods": [],
        "role": "ILLUSTRATIVE SYNTHETIC ROSTER / CARDINALITY FIXTURE",
        "takeaway": "The page preserves the 10-core / 9-specialist arrangement while replacing portraits and names with neutral placeholders.",
        "notes": "All 19 tiles are illustrative placeholders for layout and cardinality review. Number labels are identifiers only; role labels describe proposed responsibilities, not credentials, staffing commitments, completed work or actual named people. Arrangement follows UHG44’s 10 core tiles (three columns for three rows, then one left-aligned tile) plus nine specialist tiles in a 3×3 matrix, with responsibility lists alongside. Each card is fixed at 172.8×36 pt and uses 9 pt Arial, matching the measured UHG44 source text size; the neutral initials block substitutes for source portrait art. The current compose contract does not yet link a portrait/image slot with both rich name/role runs, so this fixture uses separate native shapes. Source card typography/margins/colors are recorded in library/component-contracts/wave2-fixture-contracts.json. No source portraits or source identities are used.",
        "canvas": canvas + footer(page), "layouts": layouts,
    }


def stress_slide():
    role = "Integration and data governance coordination lead for the proposed pilot"
    canvas = [
        canvas_text("stress-kicker", rect(36, 105, 700, 13), "ILLUSTRATIVE STRESS FIXTURE • NO REAL PERSON DEPICTED", 9, SLATE, True),
    ]
    canvas.extend(tile_items("stress", "X1", role, 36.25, 143, long_role=True))
    return {
        "id": "long-role-fixed-card-stress",
        "title": "Long role text is measured at fixed source-sized card geometry",
        "width_pt": 960, "height_pt": 540, "title_bounds": rect(36, 44, 888, 52),
        "title_font_face": "Arial", "title_font_size_pt": 22, "title_bold": True,
        "title_foreground": NAVY, "pods": [],
        "role": "BOUNDED LONG-ROLE STRESS CASE",
        "takeaway": "Keep the source card dimensions and 9 pt role type; report a fit failure rather than shrinking the role text.",
        "notes": "One synthetic 172.8×36 pt tile stresses a deliberately long proposed-role label at the same 9 pt Arial source size. The fixture has fixed text bounds and no auto-fit or font reduction. Native measurement must either demonstrate fit at this exact size or report the bounded role text as overflow with an actionable explanation. This is not a claimed successful fit and is not a real person or staffing credential.",
        "canvas": canvas + footer(1),
    }


base = json.loads(BASE.read_text())
base["slides"].append(roster_slide())
OUT.write_text(json.dumps(base, ensure_ascii=False, indent=2) + "\n")
stress = {"schema": "pptxgengo.compose-spec.v1", "slides": [stress_slide()]}
STRESS_OUT.write_text(json.dumps(stress, ensure_ascii=False, indent=2) + "\n")
print(OUT.relative_to(ROOT))
print(STRESS_OUT.relative_to(ROOT))
