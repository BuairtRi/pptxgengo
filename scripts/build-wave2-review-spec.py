#!/usr/bin/env python3
"""Rebuild Wave 2 review and measurement-only stress specifications."""
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DIRECTORY = ROOT / "library/visual-components"
BUILDERS = [
    ("build-visual-wave2-spec.py", "deliverable-review.json"),
    ("build-wave2-people-spec.py", None),
    ("build-wave2-roster-spec.py", "people-review-plus-roster.json"),
    ("build-wave2-roadmap-spec.py", "roadmap-review.json"),
    ("build-wave2-response-spec.py", "response-review.json"),
]


def main():
    slides = []
    for script, output in BUILDERS:
        subprocess.run([sys.executable, str(ROOT / "scripts" / script)], cwd=ROOT, check=True)
        if output is None:
            continue
        document = json.loads((DIRECTORY / output).read_text())
        if document["schema"] != "pptxgengo.compose-spec.v1":
            raise ValueError(f"Unsupported schema in {output}")
        slides.extend(document["slides"])
    if len({slide["id"] for slide in slides}) != len(slides):
        raise ValueError("Duplicate slide IDs in Wave 2 fixtures")
    for number, slide in enumerate(slides, 1):
        footers = [item for item in slide.get("canvas", [])
                   if item["id"] in {"page", "page-number", "footer-page"}]
        if len(footers) != 1:
            raise ValueError(f"Expected one page number on {slide['id']}")
        footers[0]["text"] = str(number)
    path = DIRECTORY / "review.json"
    path.write_text(json.dumps({"schema": "pptxgengo.compose-spec.v1", "slides": slides}, indent=2) + "\n")
    print(f"{path}: {len(slides)} slides")
    # Probe this superset once, then build only review.json from the same cache.
    # The deliberately overfull card must never become a deliverable review page.
    stress = json.loads((DIRECTORY / "people-role-stress.json").read_text())
    qualification = {"schema": "pptxgengo.compose-spec.v1", "slides": slides + stress["slides"]}
    ids = [slide["id"] for slide in qualification["slides"]]
    if len(set(ids)) != len(ids):
        raise ValueError("Duplicate qualification slide IDs")
    (DIRECTORY / "qualification.json").write_text(json.dumps(qualification, indent=2) + "\n")


if __name__ == "__main__":
    main()
