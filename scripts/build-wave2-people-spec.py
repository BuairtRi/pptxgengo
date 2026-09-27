#!/usr/bin/env python3
"""Write the two-slide Wave 2 people/source-review compose spec."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ASSET_DIR = ROOT / "samples/visual-wave2/assets"
OUT = ROOT / "library/visual-components/people-review.json"


def rect(x, y, w, h):
    return {"x": x, "y": y, "width": w, "height": h}


def asset(file):
    path = ASSET_DIR / file
    return str(path.relative_to(ROOT)), hashlib.sha256(path.read_bytes()).hexdigest()


logo_path = "samples/showcase/assets/wm_h_pos_clr_rgb_august2024.png"
logo_hash = hashlib.sha256((ROOT / logo_path).read_bytes()).hexdigest()
cam44_path, cam44_hash = asset("412b8a8aff7a377698d2990cde0f1bffb942fd079fcb623f17f499b1f65e49fd.png")
cam67_path, cam67_hash = asset("a4f73cfb44e2fc08785b0fb653bc2d0fd25435082f0979c2241fd3a48486548b.png")

NAVY = "#070154"
BLUE = "#0047FF"
SLATE = "#50658E"
PALE = "#E8EEF8"
LAVENDER = "#CFDAFF"
WHITE = "#FFFFFF"
PINK = "#FCE5F8"


def footer(page):
    return [
        {"id": "footer-band", "kind": "surface", "bounds": rect(0, 504, 960, 36), "background": PALE,
         "allow_overlap": ["wm-logo", "footer-copy", "page-number"]},
        {"id": "wm-logo", "kind": "image", "bounds": rect(36, 511, 104, 21.727),
         "asset_path": logo_path, "asset_sha256": logo_hash,
         "alt_text": "West Monroe horizontal positive color logo", "image_fit": "contain",
         "allow_overlap": ["footer-band"]},
        {"id": "footer-copy", "kind": "text", "bounds": rect(174, 516, 700, 17),
         "text": "© 2026 West Monroe Partners. Reproduction and/or distribution without West Monroe Partners’ prior consent is prohibited.",
         "font_face": "Arial", "font_size_pt": 8, "foreground": NAVY, "align": "left", "valign": "top",
         "allow_overlap": ["footer-band"]},
        {"id": "page-number", "kind": "text", "bounds": rect(892, 516, 32, 17), "text": str(page),
         "font_face": "Arial", "font_size_pt": 9, "foreground": NAVY, "align": "right", "valign": "top",
         "allow_overlap": ["footer-band"]},
    ]


def text(id, bounds, value, size, color=NAVY, bold=False, align="left", bg=""):
    item = {"id": id, "kind": "text", "bounds": bounds, "text": value, "font_face": "Arial",
            "font_size_pt": size, "bold": bold, "foreground": color, "align": align,
            "valign": "top", "inset_x": 0, "inset_y": 0}
    if bg:
        item["background"] = bg
    return item


def image(id, bounds, path, sha, alt, mode="source_crop", crop=None, focal=None):
    item = {"id": id, "kind": "image", "bounds": bounds, "asset_path": path,
            "asset_sha256": sha, "alt_text": alt, "image_fit": mode}
    if crop is not None:
        item["image_crop"] = crop
    if focal is not None:
        item["focal_x"], item["focal_y"] = focal
    return item


def para(pid, *runs, align="left", before=0, after=0):
    return {"id": pid, "align": align, "space_before_pt": before, "space_after_pt": after,
            "runs": [{"id": rid, "text": value, "font_face": "Arial", "font_size_pt": size,
                      "bold": bold, "italic": italic, "underline": underline, "foreground": color}
                     for rid, value, size, color, bold, italic, underline in runs]}


def block(id, value, size, color=NAVY, *, bold=False, paragraphs=None, marker="", marker_width=0,
          min_h=0, max_h=0, gap=0, left=0, right=0, bg="", align="left"):
    out = {"id": id, "text": value, "font_face": "Arial", "font_size_pt": size, "bold": bold,
           "foreground": color, "align": align, "gap_before_pt": gap, "min_height_pt": min_h,
           "left_inset_pt": left, "right_inset_pt": right}
    if paragraphs:
        out["paragraphs"] = paragraphs
        out["font_face"] = ""
        out["font_size_pt"] = 0
        out["bold"] = False
        out["foreground"] = ""
    if marker:
        out["marker"] = marker
        out["marker_width_pt"] = marker_width
    if max_h:
        out["max_height_pt"] = max_h
    if bg:
        out["background"] = bg
    return out


slides = [
    {
        "id": "people-source-controls",
        "title": "Source portrait and roster treatments for review",
        "width_pt": 960, "height_pt": 540,
        "title_bounds": rect(36, 44, 888, 52), "title_font_face": "Arial", "title_font_size_pt": 23,
        "title_bold": True, "title_foreground": NAVY, "pods": [],
        "role": "SOURCE COMPONENT QUALIFICATION / CONTROL",
        "takeaway": "The source card crop is the reconstruction control; fit variants are explicit visual alternatives for comparison.",
        "notes": "Source-reference page only. Uses named source portrait Cam Cross solely to reproduce the source UHG44 roster tile and UHG67 crop. These people/image treatments are not synthetic personas and are not approved reusable components. UHG44 source slide 44 shape 15 + picture 69: tile frame 36.25,148.0109,172.8,36 pt; picture frame 36.25,148.0109,36,36 pt; srcRect l=.19194,t=.06977,r=.19402,b=.31619. Text Arial 9 pt; role/title exact source strings. Native source read: paragraph alignment left, before/after spacing zero, frame margins L43.200000762939/R3.599999904633/T3.599999904633/B3.599999904633 pt; first-character foreground RGB {7,1,84} (#070154); vertical anchor middle; line_rule_within true; space_within 1.0. Source content is one paragraph with a vertical-tab soft break between the name and role. This control uses separate name/role boxes to target measured glyph starts (name top 155.2109375 pt; role top 166.010940551758 pt), so it approximates rather than reproduces the original paragraph/run structure. UHG67 source slide 67 picture 12 layout 80 idx10: frame 126x125.1703 pt, crop l=.17779,t=.06677,r=.17779,b=.29327. Card fill #E8EEF8. Contain/cover items are comparison variants only; their behavior comes from the native compose fit modes, not the source deck.",
        "canvas": [
            text("control-kicker", rect(36, 111, 650, 14), "SOURCE MATERIAL • RECONSTRUCTION CONTROL", 9, SLATE, True),
            text("cam44-label", rect(36.25, 128, 198, 12), "UHG44 • source roster card", 9, SLATE, True),
            {"id": "cam44-card", "kind": "surface", "bounds": rect(36.25, 148.0109, 172.8, 36), "background": PALE,
             "allow_overlap": ["cam44-photo", "cam44-name", "cam44-role"]},
            image("cam44-photo", rect(36.25, 148.0109, 36, 36), cam44_path, cam44_hash,
                  "Source portrait used in UHG44 reconstruction control", "source_crop",
                  {"left": 0.19194, "top": 0.06977, "right": 0.19402, "bottom": 0.31619}),
            text("cam44-name", rect(79.45, 155.2109375, 125, 10.8), "Cam Cross", 9, NAVY, True),
            text("cam44-role", rect(79.45, 166.010940551758, 125, 10.8), "Executive Sponsor", 9, NAVY),
            text("cam67-label", rect(246, 128, 158, 12), "UHG67 • source bio portrait crop", 9, SLATE, True),
            image("cam67-photo", rect(246, 148, 126, 125.1703), cam67_path, cam67_hash,
                  "Source portrait crop used in UHG67 reconstruction control", "source_crop",
                  {"left": 0.17779, "top": 0.06677, "right": 0.17779, "bottom": 0.29327}),
            text("variants-label", rect(415, 128, 348, 12), "FIT BEHAVIOR COMPARISON • SAME SOURCE IMAGE", 9, SLATE, True),
            image("cam44-contain", rect(415, 148, 76, 52), cam44_path, cam44_hash,
                  "Cam Cross source portrait shown with contain fit in a 76 by 52 point frame for visual comparison", "contain"),
            image("cam44-cover", rect(509, 148, 76, 52), cam44_path, cam44_hash,
                  "Cam Cross source portrait shown with cover fit in a 76 by 52 point frame for visual comparison", "cover", focal=(0.5, 0.25)),
            text("contain-label", rect(415, 231, 76, 13), "Contain", 9, NAVY, True, "center"),
            text("cover-label", rect(509, 231, 76, 13), "Cover", 9, NAVY, True, "center"),
            {"id": "inheritance-note-bg", "kind": "surface", "bounds": rect(36, 312, 888, 143), "background": PALE,
             "allow_overlap": ["inheritance-note-title", "inheritance-note"]},
            text("inheritance-note-title", rect(52, 328, 850, 19), "Source evidence and open style questions", 14, NAVY, True),
            text("inheritance-note", rect(52, 354, 850, 85),
                 "The UHG44 card’s 43.2 pt left text inset reserves the portrait area. The name and role use separate 10.8 pt text boxes placed at native-read glyph starts (155.2109375 pt and 166.010940551758 pt) with top alignment. Source itself has one paragraph with a vertical-tab soft break; this preview splits it into boxes and does not claim identical run/paragraph structure. Native paragraph alignment, spacing, margins, vertical anchor and line spacing are documented above. The 76×52 pt contain/cover examples use different fit modes and the cover frame has focal point (0.5, 0.25); they are behavior comparisons, not source treatments. UHG67 crop values come from the layout-resolved portrait placeholder. The native source does not encode separate focal coordinates. Compare crops visually before promoting either fit treatment.",
                 11, NAVY),
        ] + footer(1),
    },
    {
        "id": "synthetic-profile",
        "title": "Illustrative synthetic profile — no real person depicted",
        "width_pt": 960, "height_pt": 540,
        "title_bounds": rect(36, 42, 888, 58), "title_font_face": "Arial", "title_font_size_pt": 20,
        "title_bold": True, "title_foreground": NAVY, "pods": [],
        "role": "ILLUSTRATIVE SYNTHETIC PROFILE",
        "takeaway": "This fictional profile describes proposed delivery responsibilities, not a real person’s credentials or results.",
        "notes": "All profile content is synthetic and illustrative; it is not a representation of any named individual, team member, client, prior engagement or achieved result. Neutral IA initials are native text, not a portrait. Three-column geometry follows the resolved UHG67 profile layout as a reconstruction reference, with changed content and unresolved inherited styles made explicit. Footer/logo and safe zone align to library/layout-components/controls.json.",
        "canvas": [
            text("synthetic-ribbon", rect(36, 105, 600, 13), "ILLUSTRATIVE • FICTIONAL CASE-FLOW MODERNIZATION SCENARIO", 9, SLATE, True),
            {"id": "initials-tile", "kind": "surface", "bounds": rect(36.25, 137, 126, 100), "background": LAVENDER,
             "allow_overlap": ["initials"]},
            text("initials", rect(36.25, 162, 126, 42), "IA", 30, BLUE, True, "center"),
            text("synthetic-name", rect(36.25, 247, 126, 40), "Illustrative\nPerson A", 12, NAVY, True),
            text("synthetic-role", rect(36.25, 293, 126, 46), "Proposed role:\nPlatform delivery lead", 10.5, NAVY),
        ],
        "layouts": [
            {
                "id": "synthetic-three-column-profile", "bounds": rect(178.588, 128, 729.84, 361),
                "columns": [{"fixed_pt": 443.412}, {"fixed_pt": 272.424}],
                "column_gaps_pt": [14], "rows": [{"fixed_pt": 176}, {"fixed_pt": 176}], "row_gap_pt": 9,
                "padding": {"top": 0, "right": 0, "bottom": 0, "left": 0},
                "cells": [
                    {"id": "bio-copy", "row": 0, "column": 0, "row_span": 2, "padding": {"top": 0, "right": 5, "bottom": 0, "left": 0},
                     "blocks": [
                         block("bio-heading", "", 13, NAVY, paragraphs=[para("role-heading", ("role-heading-run", "Role focus", 13, NAVY, True, False, True))], min_h=18),
                         block("bio-paragraph-one", "", 10.5, NAVY, paragraphs=[
                             para("bio-p1-a", ("bio-p1-r1", "The proposed role coordinates the discovery and design work for a fictional case-flow modernization. It maps current handoffs, documents decision points, and works with service and technology owners to agree which steps need clearer ownership.", 10.5, NAVY, False, False, False), after=4),
                             para("bio-p1-b", ("bio-p1-r2", "Early work would establish", 10.5, NAVY, False, False, False), ("bio-p1-r3", " observable baselines", 10.5, BLUE, True, True, False), ("bio-p1-r4", " for intake completeness, queue age and exception routing. Those are proposed measures to validate, not reported outcomes.", 10.5, NAVY, False, False, False), after=5),
                         ], min_h=56),
                         block("approach-heading", "", 13, NAVY, paragraphs=[para("approach-heading-p", ("approach-heading-r", "Working approach", 13, NAVY, True, False, True))], min_h=18, gap=2),
                         block("bio-paragraph-two", "", 10.5, NAVY, paragraphs=[
                             para("bio-p2-a", ("bio-p2-r1", "The role would facilitate working sessions, maintain a decision and dependency log, and translate agreed workflow rules into testable acceptance criteria. It would coordinate review with operations, architecture, security and product stakeholders.", 10.5, NAVY, False, False, False), after=4),
                             para("bio-p2-b", ("bio-p2-r2", "Any future plan remains subject to client discovery, data access, policy review and delivery capacity.", 10.5, NAVY, False, True, False)),
                         ], min_h=50),
                         block("bio-paragraph-three", "", 10.5, NAVY, paragraphs=[
                             para("bio-p3-a", ("bio-p3-r1", "In a proposed pilot, the role would help define a narrow workflow, identify the systems and owners involved, and document assumptions before solution choices are made. It would coordinate review of access, retention, control and exception-handling requirements with designated client stakeholders.", 10.5, NAVY, False, False, False), after=4),
                             para("bio-p3-b", ("bio-p3-r2", "The team would then prepare a reviewable set of process maps, decision records, interface questions and acceptance criteria. Scope and sequencing would be confirmed after discovery; this illustrative profile makes no claim about delivered work or business impact.", 10.5, NAVY, False, False, False)),
                         ], min_h=60),
                     ]},
                    {"id": "illustrative-work-areas", "row": 0, "column": 1, "background": PALE,
                     "padding": {"top": 12, "right": 12, "bottom": 10, "left": 12},
                     "blocks": [
                         block("industry-panel-heading", "", 12, NAVY, paragraphs=[para("areas-title", ("areas-title-run", "Illustrative work areas", 12, NAVY, True, False, True))], min_h=16),
                         block("area-intake", "Intake and validation", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("area-case", "Case preparation and evidence traceability", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("area-queue", "Decision queues and exception routing", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("area-integrations", "Managed interfaces and recovery paths", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("area-measures", "Baseline measures and ownership", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                     ]},
                    {"id": "proposed-responsibilities", "row": 1, "column": 1, "background": PALE,
                     "padding": {"top": 12, "right": 12, "bottom": 10, "left": 12},
                     "blocks": [
                         block("responsibility-heading", "", 12, NAVY, paragraphs=[para("responsibility-title", ("responsibility-title-run", "Proposed responsibilities", 12, NAVY, True, False, True))], min_h=16),
                         block("resp-discovery", "Facilitate discovery and map handoffs", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("resp-decisions", "Record decisions, risks and dependencies", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("resp-validation", "Coordinate review of proposed measures", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("resp-criteria", "Prepare testable acceptance criteria", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                         block("resp-followup", "Track open questions and review actions", 9.5, NAVY, marker="•", marker_width=14, min_h=23, left=14),
                     ]},
                ],
                "layer": 10
            }
        ],
    }
]

# Add shared footer as native canvas objects on slide two (the second slide has
# a layout grid, but the footer is below it and remains in the reserved band).
slides[1]["canvas"].extend(footer(2))

spec = {"schema": "pptxgengo.compose-spec.v1", "slides": slides}
OUT.parent.mkdir(parents=True, exist_ok=True)
OUT.write_text(json.dumps(spec, ensure_ascii=False, indent=2) + "\n")
print(OUT.relative_to(ROOT))
