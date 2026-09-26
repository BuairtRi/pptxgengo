#!/usr/bin/env python3
"""Build the EnableComp five-row response fixture from pinned source art."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ASSET_DIR = ROOT / "samples/visual-wave2/assets"
PREVIEW_DIR = ASSET_DIR / "response-preview"
OUT_SPEC = ROOT / "library/visual-components/response-review.json"
OUT_MANIFEST = ROOT / "samples/visual-wave2/response-assets.json"
SOURCE = ROOT / "samples/EnableComp_RFP Response_DRAFT_081126.pptx"
SOURCE_SHA = "22c39b27bd97fece99008fd456cdde377ec0750eb296100d8570ee83f767c77b"

ROWS = [
    ("Complex RCM Operating Model", "1. Complex RCM Operator Experience", "Experienced former operators and advisors in end-to-end healthcare RCM and complex claims, along with the challenges these present to modernizing technology and operations", "407394d1e53f5c9d006fbb80af48d5a194377621aeb47d90e47864ab516a9f14"),
    ("Drive Adoption", "2. Embedded execution, not just recommendations", "We work alongside operating teams to reinforce adoption, resolve friction, and help translate deployed capabilities into measurable performance", "db9a0035f6f7999f1af460e4e7555757913e75de64408f7ef587f8846166caba"),
    ("Measurement and Intervention Discipline", "3. Process and performance measurement", "We connect workflow evidence, adoption data, productivity measures, and value outcomes to identify where performance is moving and what should change next", "47ba711b3a17f89536e988d80a8a4ef915c3041b1c45c1f0e60ba5800c3ded96"),
    ("Practical Change Reinforcement", "4. Change leadership built into delivery", "Change practitioners help translate new capabilities into new behaviors, leader routines, role expectations, and sustained use", "a55e4e676b43addeeb1f825bd3fec34587ceef60791939549564face83975197"),
    ("A Repeatable Path to Scale", "5. One integrated team with targeted technical depth", "RCM, operations, AI, process, change, and data expertise work together, with architecture/integration support brought in as needed rather than treated as a standalone workstream", "613e43c7aa81fdbad682b16cec81b49707dd4686c76789e68acacb0bff2ddd5b"),
]
PARTS = ["image21.svg", "image20.svg", "image22.svg", "image24.svg", "image23.svg"]
PICTURE_IDS = [91, 90, 94, 97, 95]
RESPONSE_IDS = [26, 2, 6, 10, 7]
NEED_IDS = [9, 12, 15, 17, 39]
PREVIEW_HASHES = [
    "9e8408b8f85a393e8437f317de5e2e248434dc7888e5d67c3ae62d9d9b56cedb",
    "441e4fe4ab8c3e948a08473017dd15ab4e709aee87f0ec43313a7f26ddc648fb",
    "8cd7c636e3e2537c5dfae7ef097e11b5ebc2a346c6d9548479461384223ffad9",
    "f97dd27d8e0ebcb96ec7d733d73e87c57c10dfd82404af24b33bcc74c5877e9c",
    "1ce75551cbe43ad9e91829d524e69d5d7db42f975acd132d2a02194e1661415c",
]
NAVY, MAGENTA, PALE, WHITE = "#070154", "#C000A8", "#F2F2F2", "#FFFFFF"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rect(x, y, w, h):
    return {"x": x, "y": y, "width": w, "height": h}


def text(id, value, bounds, size, color=NAVY, bold=False, align="left", layer=20):
    return {"id": id, "kind": "text", "bounds": rect(*bounds), "text": value,
            "font_face": "Arial", "font_size_pt": size, "bold": bold,
            "foreground": color, "align": align, "valign": "middle",
            "inset_x": 0, "inset_y": 0, "layer": layer}


def main():
    if sha(SOURCE) != SOURCE_SHA:
        raise SystemExit("Source deck hash does not match pinned candidate registry")
    ids = []
    manifest_assets = []
    for i, ((_, _, _, source_hash), part, preview_hash_expected) in enumerate(zip(ROWS, PARTS, PREVIEW_HASHES), 1):
        svg = ASSET_DIR / f"{source_hash}.svg"
        if sha(svg) != source_hash:
            raise SystemExit(f"Pinned source SVG hash mismatch: {svg}")
        preview = PREVIEW_DIR / f"{preview_hash_expected}.png"
        if not preview.exists():
            raise SystemExit(f"Missing derived AppKit preview for row {i}; render the pinned SVG first")
        preview_hash = sha(preview)
        if preview_hash != preview_hash_expected:
            raise SystemExit(f"Derived preview hash mismatch for row {i}")
        hashed_preview = PREVIEW_DIR / f"{preview_hash}.png"
        if preview != hashed_preview:
            if hashed_preview.exists():
                if sha(hashed_preview) != preview_hash:
                    raise SystemExit(f"Conflicting derived preview at {hashed_preview}")
                preview.unlink()
            else:
                preview.rename(hashed_preview)
        ids.append((svg, hashed_preview, preview_hash))
        manifest_assets.append({
            "row": i, "source_deck": str(SOURCE.relative_to(ROOT)), "source_deck_sha256": SOURCE_SHA,
            "source_part": f"ppt/media/{part}", "source_sha256": source_hash,
            "source_file": str(svg.relative_to(ROOT)),
            "derived_file": str(hashed_preview.relative_to(ROOT)), "derived_sha256": preview_hash,
            "derived_type": "raster_preview_not_editable_vector",
            "renderer": "scripts/render-svg-preview.swift", "renderer_recipe": "appkit-nsimage-svg-8x-v1",
            "renderer_sha256": sha(ROOT / "scripts/render-svg-preview.swift"),
            "preview_dimensions_px": [608, 608],
            "association": f"EnableComp slide 5 picId={PICTURE_IDS[i-1]} aligned with spId={RESPONSE_IDS[i-1]}",
        })

    controls = json.loads((ROOT / "library/layout-components/controls.json").read_text())
    base = controls["slides"][0]
    page = {**base, "id": "enablecomp-response-fixture", "title": "Five connected needs and delivery responses",
            "layouts": [], "pods": [],
            "title_bounds": rect(36, 44, 888, 52), "title_font_size_pt": 23,
            "role": "ILLUSTRATIVE CHANGED-CONTENT FIXTURE",
            "takeaway": "Five paired needs and responses demonstrate icon-led rows, rich copy, separators and an editable directional shape.",
            "notes": "Illustrative changed-content fixture based on EnableComp source slide 5 (source deck SHA-256 " + SOURCE_SHA + "). Source visual uses a rotated triangle behind five need tiles; this fixture uses an editable rightArrow as a source-like directional treatment. Five source SVG icons are preserved as pinned originals and rendered to 8x AppKit PNG previews because the current compose image decoder accepts PNG/JPEG, not SVG. Derived PNGs are raster previews, not editable vectors; see samples/visual-wave2/response-assets.json for original and derived hashes. Need labels and response titles/body are adapted sample copy; source panel geometry is retained approximately. This is a nonnative probe fixture and not native PowerPoint rendering evidence.",
            "canvas": []}
    canvas = page["canvas"]
    # Keep branded footer and suppress layout-specific engagement strip.
    for item in base["canvas"]:
        if item["id"] in {"footer-band", "wm-logo", "footer-copy", "page"}:
            clone = dict(item)
            if clone["id"] == "page": clone["text"] = "1"
            canvas.append(clone)
    canvas += [
        text("section-needs", "ILLUSTRATIVE NEEDS", [72, 126, 230, 22], 18, MAGENTA, True, "center"),
        text("section-response", "WHY WEST MONROE", [500, 126, 285, 22], 18, MAGENTA, True, "center"),
        {"id": "direction-arrow", "kind": "shape", "bounds": rect(218, 193, 120, 247),
         "preset": "rightArrow", "background": "#BFBFBF", "layer": 1,
         "allow_overlap": [f"need-tile-{j}" for j in range(1, 6)]},
    ]
    pitch = 61.7
    left_x, left_w, left_h = 39.7, 226.8, 43.2
    row_x, row_w, row_h = 344.6, 578.6, 46.8
    top = 155.0
    for i, (need, lead, body, source_hash) in enumerate(ROWS):
        y = top + i * pitch
        # Source left tile: navy block and centered white need label.
        canvas.append({"id": f"need-tile-{i+1}", "kind": "surface", "bounds": rect(left_x, y+4, left_w, left_h),
                       "background": NAVY, "layer": 5, "allow_overlap": [f"need-label-{i+1}", "direction-arrow"]})
        need_text = text(f"need-label-{i+1}", need, [left_x+8, y+7, left_w-16, left_h-6], 14, WHITE, True, "center", 8)
        need_text["contrast_background"] = NAVY
        need_text["allow_overlap"] = [f"need-tile-{i+1}", "direction-arrow"]
        canvas.append(need_text)
        # Pale response row and original icon raster preview in source proportions.
        canvas.append({"id": f"response-row-{i+1}", "kind": "surface", "bounds": rect(row_x, y, row_w, row_h),
                       "background": PALE, "layer": 5,
                       "allow_overlap": [f"row-icon-{i+1}", f"row-divider-{i+1}", f"row-copy-{i+1}"]})
        png = ids[i][1]
        canvas.append({"id": f"row-icon-{i+1}", "kind": "image", "bounds": rect(row_x+13, y+7.2, 32.2, 32.2),
                       "asset_path": str(png.relative_to(ROOT)), "asset_sha256": ids[i][2],
                       "alt_text": f"Pinned EnableComp source icon for response row {i+1}; raster preview of original SVG", "image_fit": "contain", "layer": 10,
                       "allow_overlap": [f"response-row-{i+1}"]})
        canvas.append({"id": f"row-divider-{i+1}", "kind": "line", "bounds": rect(row_x+58, y+6.5, 0, 33.8),
                       "foreground": NAVY, "line_width_pt": 1.2, "layer": 10, "allow_overlap": [f"response-row-{i+1}"]})
        canvas.append({"id": f"row-copy-{i+1}", "kind": "text", "bounds": rect(row_x+70, y+3.8, row_w-79, row_h-6.5),
                       "align": "left", "valign": "middle", "inset_x": 0, "inset_y": 0, "layer": 12,
                       "allow_overlap": [f"response-row-{i+1}"], "contrast_background": PALE,
                       "paragraphs": [
                           {"id": f"p-title-{i+1}", "align": "left", "space_after_pt": 0,
                            "runs": [{"id": f"r-title-{i+1}", "text": lead, "font_face": "Arial", "font_size_pt": 13.5,
                                      "bold": True, "foreground": NAVY}]},
                           {"id": f"p-body-{i+1}", "align": "left", "space_before_pt": 0,
                            "runs": [{"id": f"r-body-{i+1}", "text": body, "font_face": "Arial", "font_size_pt": 9.5,
                                      "bold": False, "foreground": NAVY}]},
                       ]})
    document = {"schema": "pptxgengo.compose-spec.v1", "slides": [page]}
    OUT_SPEC.write_text(json.dumps(document, indent=2) + "\n")
    manifest = {"schema": "pptxgengo.wave2-response-assets.v1", "status": "source_pinned_preview_derived",
                "source_deck": str(SOURCE.relative_to(ROOT)), "source_deck_sha256": SOURCE_SHA,
                "source_has_png_fallbacks": False, "rendered_source_svg_count": len(manifest_assets),
                "items": manifest_assets}
    OUT_MANIFEST.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"{OUT_SPEC.relative_to(ROOT)}; {len(ROWS)} source-pinned SVGs rendered; manifest {OUT_MANIFEST.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
