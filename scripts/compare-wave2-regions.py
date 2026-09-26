#!/usr/bin/env python3
"""Report source/control pixel differences; metrics are not an approval threshold."""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops, ImageStat

ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("render_dir", type=Path)
    parser.add_argument("out", type=Path)
    args = parser.parse_args()
    if args.out.exists():
        parser.error("output must be a new directory")
    spec = json.loads((ROOT / "library/visual-components/review.json").read_text())
    fixtures = {slide["id"]: (i + 1, slide) for i, slide in enumerate(spec["slides"])}
    page, control = fixtures["source-deliverable-panel"]
    source_path = ROOT / "samples/reconstruction/reference/png/slide-028.png"
    target_path = args.render_dir / f"slide-{page:03d}.png"
    source = Image.open(source_path).convert("RGB")
    target = Image.open(target_path).convert("RGB")
    if source.size != (1920, 1080) or target.size != source.size:
        raise ValueError("Expected matching 1920×1080 native renders; do not resample")
    args.out.mkdir(parents=True)
    rows = []
    for item in control["canvas"]:
        if not item["id"].startswith(("source-pic-", "source-caption-")):
            continue
        b = item["bounds"]
        # Rounded crop edges include no extra region outside the declared frame.
        box = tuple(round(v * 2) for v in
                    (b["x"], b["y"], b["x"] + b["width"], b["y"] + b["height"]))
        a, z = source.crop(box), target.crop(box)
        diff = ImageChops.difference(a, z)
        stats = ImageStat.Stat(diff)
        equal = sum(pixel == (0, 0, 0) for pixel in diff.get_flattened_data())
        within_one = sum(max(pixel) <= 1 for pixel in diff.get_flattened_data())
        count = a.width * a.height
        prefix = args.out / item["id"]
        a.save(str(prefix) + "-source.png")
        z.save(str(prefix) + "-control.png")
        diff.save(str(prefix) + "-difference.png")
        rows.append({"element": item["id"], "crop_pixels": box,
                     "pixel_count": count, "exact_pixel_fraction": equal / count,
                     "within_one_rgb_level_fraction": within_one / count,
                     "mean_absolute_rgb_error_0_255": sum(stats.mean) / 3,
                     "maximum_channel_error": max(high for _, high in stats.extrema)})
    report = {"schema": "pptxgengo.wave2-region-comparison.v1",
              "scope": "UHG28 original geometry control, overlapping artwork included; no global pixel-fidelity claim",
              "source_render": str(source_path.relative_to(ROOT)), "source_sha256": sha(source_path),
              "control_render": str(target_path), "control_sha256": sha(target_path),
              "regions": rows}
    (args.out / "comparison.json").write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps(rows, indent=2))


if __name__ == "__main__":
    main()
