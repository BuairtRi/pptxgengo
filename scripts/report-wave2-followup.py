#!/usr/bin/env python3
"""Bind Wave 2 follow-up fixture acceptance to native, structural and reviewed artifacts."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LOCAL = ROOT / "samples/visual-wave2-followup"
LIBRARY = ROOT / "library/visual-components"


def read(path):
    return json.loads(path.read_text())


def record(path):
    data = path.read_bytes()
    return {"path": str(path.relative_to(ROOT)), "sha256": hashlib.sha256(data).hexdigest(),
            "size_bytes": len(data)}


def main():
    spec_path = LIBRARY / "followup-review.json"
    spec = read(spec_path)
    bundle = LOCAL / "followup-review"
    manifest_path = bundle / "manifest.json"
    manifest = read(manifest_path)
    deck = bundle / manifest["deck_file"]
    evidence_path = LOCAL / "verification.json"
    evidence = read(evidence_path)
    fit_path = LOCAL / "fit.json"
    fit = read(fit_path)
    negative_path = LOCAL / "negative-fit.json"
    negative = read(negative_path)
    reviews_path = LOCAL / "visual-review.json"
    reviews = read(reviews_path)
    assert manifest["spec_sha256"] == evidence["spec_sha256"] == record(spec_path)["sha256"]
    assert manifest["deck_sha256"] == evidence["deck_sha256"] == record(deck)["sha256"]
    assert evidence["manifest_sha256"] == record(manifest_path)["sha256"]
    assert evidence["native"]["schema"] == "pptxgengo.compose-text-measurement.v8"
    assert fit["planner_passed"] and fit["overflow_count"] == fit["layout_failure_count"] == 0
    assert fit["spec_sha256"] == record(spec_path)["sha256"]
    assert not negative["planner_passed"] and negative["overflow_count"] > 0
    assert negative["spec_sha256"] == record(LIBRARY / "followup-negative.json")["sha256"]
    assert [review["id"] for review in reviews] == [slide["id"] for slide in spec["slides"]]
    assert all(review["reviewed"] and review["accepted"] for review in reviews)
    for review in reviews:
        assert review["sha256"] == record(ROOT / review["artifact"])["sha256"]
    native_rows = evidence["native"]["measurements"]
    rows = {(row["slide_index"], row["shape_name"]): row for row in native_rows}
    assert len(rows) == len(native_rows)
    max_frame = max_overflow = 0.0
    text_count = image_count = shape_count = rich_count = 0
    svg_count = outline_count = bullet_count = 0
    picture_layer_checks = []
    for number, slide in enumerate(manifest["slides"], 1):
        for element_index, element in enumerate(slide["elements"]):
            native = rows[(number, element["name"])]
            frame, observed = element["frame"], native["shape_frame"]
            max_frame = max(max_frame, *(abs(frame[a] - observed[b]) for a, b in
                                        [("x", "left"), ("y", "top"), ("width", "width"), ("height", "height")]))
            image_count += element["kind"] == "image"
            svg_count += element["kind"] == "image" and element.get("asset_path", "").endswith(".svg")
            outline_count += element["kind"] == "image" and bool(element.get("outline_color"))
            bullet_count += sum(bool(p.get("bullet")) for p in element.get("paragraphs", []))
            shape_count += element["kind"] == "shape"
            if element["kind"] == "image":
                # Regression for QA83. This catches complete rectangular surface
                # occlusion, not every possible partial or irregular obstruction.
                for later in slide["elements"][element_index + 1:]:
                    if later["kind"] != "surface" or not later.get("background"):
                        continue
                    cover = later["frame"]
                    contains = (cover["x"] <= frame["x"] and cover["y"] <= frame["y"]
                                and cover["x"] + cover["width"] >= frame["x"] + frame["width"]
                                and cover["y"] + cover["height"] >= frame["y"] + frame["height"])
                    assert not contains, f"{element['name']} hidden behind {later['name']}"
                picture_layer_checks.append({"slide": number, "picture": element["name"],
                                             "not_fully_covered_by_later_surface": True})
            if element["kind"] == "text":
                text_count += 1
                rich_count += bool(element.get("paragraphs"))
                bounds = native["text_bounds"]
                ix, iy = element["inset_x"], element["inset_y"]
                max_overflow = max(max_overflow, frame["x"] + ix - bounds["left"],
                                   frame["y"] + iy - bounds["top"],
                                   bounds["left"] + bounds["width"] - frame["x"] - frame["width"] + ix,
                                   bounds["top"] + bounds["height"] - frame["y"] - frame["height"] + iy)
    assert max_frame <= .12 and max_overflow <= .15
    report = {
        "schema": "pptxgengo.visual-fixture-proof.v1",
        "status": "bounded_fixtures_verified_not_full_wave2_or_pixel_identity_approval",
        "slides": len(spec["slides"]), "objects": len(rows), "text_objects": text_count,
        "native_svg_pictures": svg_count, "native_picture_outlines": outline_count, "native_bullet_paragraphs": bullet_count,
        "rich_text_objects": rich_count, "pictures": image_count, "preset_shapes": shape_count,
        "max_frame_delta_pt": max_frame, "max_text_bound_excursion_pt": max_overflow,
        "environment": evidence["environment"], "visual_review": reviews,
        "picture_layer_checks": picture_layer_checks,
        "source_region_comparison": read(LOCAL / "source-region-comparison/comparison.json"),
        "adapter_regression": "Six-object historical smoke: all existing native fields exactly equal under v8",
        "code": [record(path) for path in sorted(
            list((ROOT / "cmd/pptxcompose").glob("*.go"))
            + list((ROOT / "internal/compose").glob("*.go"))
            + list((ROOT / "pptx").glob("*.go"))
            + [ROOT / "scripts/measure-compose-text.applescript",
               ROOT / "scripts/compose-environment.swift",
               ROOT / "scripts/render-svg-preview.swift",
               ROOT / "scripts/build-wave2-response-spec.py",
               ROOT / "scripts/build-wave2-followup-spec.py",
               ROOT / "scripts/build-wave2-people-spec.py"])],
        "negative_case": {"overflow_count": negative["overflow_count"],
                          "failures": [zone for zone in negative["zones"] if not zone["fits"]]},
        "limitations": ["Bullets limited to three glyphs; indentation structural, bullet glyph fit visually reviewed",
                        "Source-derived component controls are not whole-slide pixel reconstructions",
                        "Native groups and tables are separate editable shapes in these fixtures",
                        "Native SVG fallback fidelity depends on explicitly reviewed pinned PNG assets",
                        "SVG pictures retain original vector media, not editable individual paths",
                        "Rich-text recovery and automatic aesthetic layout selection remain unsupported"],
        "artifacts": [record(path) for path in [spec_path, deck, manifest_path, evidence_path,
                                                fit_path, negative_path, reviews_path,
                                                LOCAL / "source-region-comparison/comparison.json",
                                                ROOT / "samples/visual-wave2/response-assets.json",
                                                LOCAL / "followup-review.pdf"]],
    }
    (LIBRARY / "followup-proof.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps({key: report[key] for key in ["slides", "objects", "text_objects", "rich_text_objects",
                                                   "pictures", "preset_shapes", "max_frame_delta_pt",
                                                   "max_text_bound_excursion_pt"]}, indent=2))


if __name__ == "__main__":
    main()
