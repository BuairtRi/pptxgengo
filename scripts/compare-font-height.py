#!/usr/bin/env python3
"""Compare character-box heights and PDF baseline pitches as separate contracts.

Uses saved native evidence; never invokes PowerPoint or claims final native fit.
Output is new-only. Run after measuring candidates and extracting PDF baselines.
"""
import argparse
import base64
import hashlib
import json
from pathlib import Path
import statistics

ROOT = Path(__file__).resolve().parents[1]


def load(path):
    return json.loads(path.read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def case_id(request_id):
    key = request_id.removeprefix("measure:")
    return json.loads(base64.urlsafe_b64decode(key + "="*((-len(key)) % 4)))[0]


def stats(values):
    return dict(count=len(values), mean_absolute_delta_pt=statistics.mean(map(abs, values)),
                max_absolute_delta_pt=max(map(abs, values))) if values else dict(count=0)


def native_controls(data, candidate):
    evidence = load(data / "native-control-evidence.json")
    manifest = load(data / "native-control-manifest.json")
    previous = load(data / "native-control-original-go.json")
    require(evidence["manifest_sha256"] == digest(data / "native-control-manifest.json"), "Native manifest hash mismatch")
    require(evidence["deck_sha256"] == digest(data / "native-control-probes.pptx"), "Native probe hash mismatch")
    require(evidence["spec_sha256"] == digest(ROOT / "library/dynamic-components/fonts.json") == candidate["spec_sha256"] == previous["spec_sha256"], "Native control spec mismatch")
    require(candidate["fonts"] == previous["fonts"], "Native control Go font resolution changed")
    signature = lambda r: (r["family"].lower(), r["style"], r["sha256"])
    require({signature(r) for r in evidence["environment"]["fonts"]} == {signature(r) for r in candidate["fonts"]}, "Native/Go font family, style or file hash mismatch")
    require(not candidate["powerpoint_verified"], "Go results incorrectly claim native verification")
    native_rows = evidence["native"]["measurements"]
    require(len(native_rows) == len(manifest["requests"]) == len(candidate["requests"]), "Native control count mismatch")
    rows = []
    for slide, native in zip(manifest["slides"], native_rows):
        element = slide["elements"][0]
        require(element["name"] == native["shape_name"], "Native shape order mismatch")
        # These saved frames establish a 1:1 raw-unit-to-point mapping.
        for key, native_key in [("x", "left"), ("y", "top"), ("width", "width"), ("height", "height")]:
            require(abs(element["frame"][key]-native["shape_frame"][native_key]) < .0001, "Native raw coordinate scale differs from slide points")
        rid = element["measurement_id"]
        current = candidate["requests"][rid]
        before = previous["requests"][rid]
        require([l["text"] for l in current["lines"]] == [l["text"] for l in before["lines"]], "Native control wrapping changed")
        groups = {}
        for char in native["characters"]:
            if char["text"].strip():
                top = round(char["bounds"]["top"]-native["shape_frame"]["top"], 4)
                groups.setdefault(top, []).append(char["text"])
        native_lines = ["".join(groups[y]) for y in sorted(groups)]
        require(["".join(l["text"].split()) for l in current["lines"]] == native_lines, "Native control character topology mismatch")
        actual = evidence["measurements"]["by_request_id"][rid]
        now = candidate["measurements"]["by_request_id"][rid]
        old = previous["measurements"]["by_request_id"][rid]
        rows.append(dict(request_id=rid, text=native["text"], line_count=len(current["lines"]),
            native_height_pt=actual["rendered_height_pt"], go_height_pt=now["rendered_height_pt"],
            old_height_delta_pt=old["rendered_height_pt"]-actual["rendered_height_pt"],
            height_delta_pt=now["rendered_height_pt"]-actual["rendered_height_pt"],
            native_line_tops_pt=sorted(groups), go_line_tops_pt=[l["y_pt"] for l in current["lines"]]))
    return dict(contract="native non-whitespace character bounds in slide points",
        scope="Ten saved Arial/Plex Sans controls used for calibration, not independent validation; no Mono native character-bound capture",
        evidence_sha256=digest(data / "native-control-evidence.json"),
        font_file_hashes_match=True, raw_coordinate_scale=1,
        before=stats([r["old_height_delta_pt"] for r in rows]),
        after=stats([r["height_delta_pt"] for r in rows]), rows=rows)


def pdf_pitches(data, prefix, dataset, previous_path, candidate):
    cases = load(dataset / "cases.json")["cases"]
    previous = load(previous_path)
    selection = load(dataset / "native-measurements.json")
    baselines = load(data / (prefix+"-baselines.json"))
    old_comparison = load(previous_path.parent / "comparison-final.json")
    require(baselines["contract"] == "powerpoint-pdf-text-show-origins.v1", "Unexpected baseline contract")
    require(baselines["pdf_sha256"] == selection["pdf_sha256"] == old_comparison["native_pdf_sha256"], "PDF capture mismatch")
    pdf_path = dataset / ("native-reference.pdf" if prefix == "reference" else "heldout-native.pdf")
    require(digest(pdf_path) == baselines["pdf_sha256"], "Archived PDF hash mismatch")
    require(candidate["spec_sha256"] == previous["spec_sha256"] == digest(dataset / "corpus.json"), "Corpus mismatch")
    require(candidate["fonts"] == previous["fonts"], "Go font resolution changed")
    require(not candidate["powerpoint_verified"], "Go results incorrectly claim native verification")
    require(len(cases) == len(baselines["pages"]) == selection["page_count"] == len(candidate["requests"]), "PDF/corpus count mismatch")
    ids = {case_id(k): k for k in candidate["requests"]}
    comparisons = {r["case_id"]: r for r in old_comparison["rows"]}
    rows = []
    spans = []
    skipped = []
    line_matches = 0
    for i, case in enumerate(cases):
        cid = case["id"]
        rid = ids[cid]
        now = candidate["requests"][rid]
        old = previous["requests"][rid]
        # Widths, clusters, text and warning decisions must remain frozen.
        width_fields = ("text", "start_rune", "end_rune", "x_pt", "advance_pt")
        require([[l[k] for k in width_fields] for l in now["lines"]] == [[l[k] for k in width_fields] for l in old["lines"]], "Wrapping/advance changed: "+cid)
        require(now.get("boundary_warnings", []) == old.get("boundary_warnings", []), "Wrap warning changed: "+cid)
        page = baselines["pages"][i]
        require(page["page_number"] == i+1, "PDF page order mismatch")
        ys = []
        for origin in sorted(page["origins"], key=lambda r: r["baseline_pt"]):
            if not ys or origin["baseline_pt"]-ys[-1] > .02:
                ys.append(origin["baseline_pt"])
        require(len(ys) == len(selection["pages"][i]["lines"]), "PDF baseline/selection row disagreement: "+cid)
        if not comparisons[cid]["line_breaks_match"]:
            skipped.append(dict(case_id=cid, reason="Recorded wrap mismatch; no one-to-one baseline comparison"))
            continue
        line_matches += 1
        visible = [l for l in now["lines"] if l["text"].strip()]
        old_visible = [l for l in old["lines"] if l["text"].strip()]
        require(len(visible) == len(ys), "Candidate/PDF row disagreement: "+cid)
        if len(ys) > 1:
            native_span = ys[-1]-ys[0]
            spans.append(dict(case_id=cid,
                delta_pt=visible[-1]["baseline_pt"]-visible[0]["baseline_pt"]-native_span,
                old_delta_pt=old_visible[-1]["baseline_pt"]-old_visible[0]["baseline_pt"]-native_span))
        for j in range(1, len(ys)):
            native = ys[j]-ys[j-1]
            current = visible[j]["baseline_pt"]-visible[j-1]["baseline_pt"]
            old_pitch = old_visible[j]["baseline_pt"]-old_visible[j-1]["baseline_pt"]
            rows.append(dict(case_id=cid, family=case["family"], category=case["category"], line_index=j,
                native_pdf_pitch_pt=native, go_pitch_pt=current, old_go_pitch_pt=old_pitch,
                delta_pt=current-native, old_delta_pt=old_pitch-native))
    return dict(contract="PDF text-show baseline pitch; never selection height or TextRange bounds",
        pdf_sha256=baselines["pdf_sha256"], case_count=len(cases), line_break_match_count=line_matches,
        wrapping_unchanged=True, font_provenance_unchanged=True,
        before=stats([r["old_delta_pt"] for r in rows]), after=stats([r["delta_pt"] for r in rows]),
        first_to_last_baseline_span=dict(before=stats([r["old_delta_pt"] for r in spans]), after=stats([r["delta_pt"] for r in spans])),
        skipped=skipped, rows=rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--data", type=Path, default=ROOT / "library/dynamic-components/font-height-calibration")
    parser.add_argument("--control-go", type=Path)
    parser.add_argument("--reference-go", type=Path)
    parser.add_argument("--additional-go", type=Path)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    root = ROOT / "library/dynamic-components"
    control = load(args.control_go or args.data / "native-control-final-go.json")
    reference = load(args.reference_go or args.data / "reference-final-go.json")
    additional = load(args.additional_go or args.data / "additional-final-go.json")
    require(control["engine"] == reference["engine"] == additional["engine"], "Candidate engines differ")
    require(control["height_policy"] == reference["height_policy"] == additional["height_policy"], "Candidate height policies differ")
    report = dict(schema="pptxgengo.font-height-comparison.v1", engine=control["engine"],
        height_policy=control["height_policy"], powerpoint_verified=False,
        native_controls=native_controls(args.data, control),
        pdf_sets=[pdf_pitches(args.data, "reference", root / "font-reference", root / "font-wrap-calibration/calibrated-go.json", reference),
                  pdf_pitches(args.data, "additional", root / "font-wrap-calibration/heldout", root / "font-wrap-calibration/heldout/calibrated-go.json", additional)])
    with args.out.open("x") as stream:
        json.dump(report, stream, indent=2, ensure_ascii=False, allow_nan=False)
        stream.write("\n")
    print(args.out)


if __name__ == "__main__":
    main()
