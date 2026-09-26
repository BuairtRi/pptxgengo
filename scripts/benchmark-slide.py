#!/usr/bin/env python3
"""Run ONE selected scene through native export and pixel QA; inspect before next.

Requires PowerPoint macOS, Swift, and Go. Source PNGs must already exist.
Reports literal full-deck raster differences; no acceptance threshold is hidden.
"""
import argparse
import json
from pathlib import Path
import subprocess
import zipfile


def run(args):
    subprocess.run([str(a) for a in args], check=True)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("slide", type=int)
    p.add_argument("--project", type=Path, default=Path("samples/reconstruction/native-scene"))
    p.add_argument("--label", default="native")
    p.add_argument("--control", action="store_true", help="also render untouched source slide in identical deck context")
    p.add_argument("--control-only", action="store_true", help="create control for a previously rendered candidate")
    args = p.parse_args()
    root = Path(__file__).resolve().parent.parent
    base = root / "samples/reconstruction"
    number = args.slide
    output = base / f"work/uhg-{number:03d}-{args.label}.pptx"
    pdf = base / f"reference/uhg-{number:03d}-{args.label}.pdf"
    qa = base / f"qa/{number:03d}-{args.label}"
    if not args.control_only and (output.exists() or pdf.exists() or qa.exists()):
        p.error("Output exists; choose a new --label to retain prior evidence")
    if not args.control_only:
        run(["go", "run", "./cmd/pptxscene", "build", "--project", args.project,
             "--slides", number, "--freeze-slide-numbers", "--out", output])
        run(["osascript", root / "scripts/export-powerpoint.applescript", output, pdf])
        if not pdf.is_file():
            raise RuntimeError("PowerPoint reported success but did not create a PDF")
        run(["swift", root / "scripts/render-pdf.swift", pdf, qa / "candidate", 1])
        run(["go", "run", "./cmd/pptxdiff", "--reference", base / f"reference/png/slide-{number:03d}.png",
             "--candidate", qa / "candidate/slide-001.png", "--out", qa / "diff"])
    if args.control or args.control_only:
        control = base / f"work/uhg-{number:03d}-{args.label}-control.pptx"
        control_pdf = base / f"reference/uhg-{number:03d}-{args.label}-control.pdf"
        if control.exists() or control_pdf.exists():
            raise RuntimeError("Control already exists")
        source = root / "samples/UHG Fabric Platforming RFP Response - July 2026.pptx"
        project = json.loads((args.project / f"slides/uhg-{number:03d}.json").read_text())
        part = project["part"]
        rel = str(Path(part).parent / "_rels" / (Path(part).name + ".rels"))
        with zipfile.ZipFile(output) as rebuilt, zipfile.ZipFile(source) as original, zipfile.ZipFile(control, "w", zipfile.ZIP_DEFLATED) as dest:
            for item in rebuilt.infolist():
                data = rebuilt.read(item.filename)
                if item.filename in (part, rel):
                    data = original.read(item.filename)
                if item.filename == "ppt/presentation.xml":
                    # This control retains the untouched source slide-number field.
                    # Set its starting page to the original source slide number.
                    import re
                    text = data.decode()
                    text = re.sub(r' firstSlideNum="[^"]*"', '', text)
                    text = text.replace('<p:presentation ', f'<p:presentation firstSlideNum="{number}" ', 1)
                    data = text.encode()
                dest.writestr(item.filename, data)
        run(["osascript", root / "scripts/export-powerpoint.applescript", control, control_pdf])
        run(["swift", root / "scripts/render-pdf.swift", control_pdf, qa / "control", 1])
        run(["go", "run", "./cmd/pptxdiff", "--reference", qa / "control/slide-001.png",
             "--candidate", qa / "candidate/slide-001.png", "--out", qa / "control-diff"])
    report = json.loads((qa / "diff/report.json").read_text())
    print(json.dumps({"slide": number, "qa": str(qa),
                      "mismatch_pixels": report["exact_mismatch_count"],
                      "mean_rgb_error": report["mean_absolute_rgb_channel_error"],
                      "max_rgb_delta": report["max_delta_rgb"]}))


if __name__ == "__main__":
    main()
