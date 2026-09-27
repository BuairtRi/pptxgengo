#!/usr/bin/env python3
"""Recompile manifest-selected adaptive fixtures and compare semantic specs.

The check ignores only the visible text of the `chrome/page` number. It does
not perform native measurement or qualify a family or source template.
"""
from __future__ import annotations

import argparse
import copy
import hashlib
import json
import subprocess
import tempfile
from pathlib import Path

FAMILIES = {"roadmap", "architecture", "process", "team", "comparison"}
SCHEMA = "pptxgengo.adaptive-reproducibility.v1"


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read_json(path: Path):
    return json.loads(path.read_text(encoding="utf-8"))


def selected_slide(document: dict, one_based_index: int, label: str) -> dict:
    slides = document.get("slides")
    if not isinstance(slides, list) or one_based_index < 1 or one_based_index > len(slides):
        raise ValueError(f"{label}: slide index {one_based_index} is out of range")
    return slides[one_based_index - 1]


def without_page_number(slide: dict) -> dict:
    result = copy.deepcopy(slide)
    for item in result.get("canvas") or []:
        if item.get("id") == "chrome/page":
            item.pop("text", None)
    return result


def canonical_fixture(root: Path, ident: str):
    matches = []
    for path in sorted((root / "library/adaptive").glob("*.json")):
        try:
            document = read_json(path)
        except (OSError, json.JSONDecodeError):
            continue
        slides = document.get("slides")
        if isinstance(slides, list) and len(slides) == 1 and slides[0].get("id") == ident:
            matches.append((path, document))
    if len(matches) != 1:
        raise ValueError(f"{ident}: expected exactly one canonical single-slide fixture, found {len(matches)}")
    return matches[0]


def differences(expected, actual, path="$", limit=100):
    found = []
    def visit(a, b, p):
        if len(found) >= limit:
            return
        if type(a) is not type(b):
            found.append({"path": p, "expected_type": type(a).__name__, "actual_type": type(b).__name__})
        elif isinstance(a, dict):
            for key in sorted(set(a) | set(b)):
                if key not in a:
                    found.append({"path": f"{p}.{key}", "expected": "<missing>", "actual": b[key]})
                elif key not in b:
                    found.append({"path": f"{p}.{key}", "expected": a[key], "actual": "<missing>"})
                else:
                    visit(a[key], b[key], f"{p}.{key}")
        elif isinstance(a, list):
            if len(a) != len(b):
                found.append({"path": p, "expected_length": len(a), "actual_length": len(b)})
            for i, (x, y) in enumerate(zip(a, b)):
                visit(x, y, f"{p}[{i}]")
        elif a != b:
            found.append({"path": p, "expected": a, "actual": b})
    visit(expected, actual, path)
    if len(found) == limit:
        found.append({"path": "$", "note": f"difference list capped at {limit}"})
    return found


def run(root: Path, review_path: Path, out_path: Path, compiler: Path | None):
    root = root.resolve()
    review_path = review_path if review_path.is_absolute() else root / review_path
    out_path = out_path if out_path.is_absolute() else root / out_path
    review = read_json(review_path)
    rows = [r for r in review.get("slides", []) if r.get("family") in FAMILIES]
    if len(rows) != 32:
        raise ValueError(f"expected 32 canonical family examples, found {len(rows)}")

    with tempfile.TemporaryDirectory(prefix="pptxadapt-repro-") as td:
        temp = Path(td)
        if compiler is None:
            compiler_path = temp / "pptxadapt"
            subprocess.run(["go", "build", "-o", str(compiler_path), "./cmd/pptxadapt"], cwd=root, check=True)
        else:
            compiler_path = compiler.resolve()
        compiler_digest = sha256(compiler_path)
        entries = []
        flat_differences = []
        for i, row in enumerate(rows, 1):
            archived_raw = row.get("values_path")
            target_raw = row.get("spec_path")
            slide_index = row.get("slide_index")
            if not archived_raw or not target_raw or not isinstance(slide_index, int):
                raise ValueError(f"{row.get('id')}: requires values_path, spec_path, and integer slide_index")
            fixture_path, source_doc = canonical_fixture(root, row["id"])
            target_path = Path(target_raw)
            fixture_path = fixture_path.resolve()
            target_path = target_path if target_path.is_absolute() else root / target_path
            source_slide = selected_slide(source_doc, 1, fixture_path.as_posix())
            if source_slide.get("family") != row["family"]:
                raise ValueError(f"{row['id']}: selected input family does not match review family")
            selected_input = {"schema": source_doc["schema"], "slides": [source_slide]}
            one_spec_input = temp / f"input-{i:02d}.json"
            one_spec_input.write_text(json.dumps(selected_input, ensure_ascii=False) + "\n", encoding="utf-8")
            compiled_dir = temp / f"compiled-{i:02d}"
            subprocess.run([str(compiler_path), "compile", "--spec", str(one_spec_input), "--out", str(compiled_dir)], cwd=root, check=True, capture_output=True, text=True)
            compiled_path = compiled_dir / "spec.json"
            actual_slide = selected_slide(read_json(compiled_path), 1, compiled_path.as_posix())
            expected_slide = selected_slide(read_json(target_path), slide_index, target_path.as_posix())
            delta = differences(without_page_number(expected_slide), without_page_number(actual_slide))
            entry = {"id": row["id"], "family": row["family"], "variant": row.get("variant"), "slide_index": slide_index,
                     "input_path": fixture_path.relative_to(root).as_posix() if fixture_path.is_relative_to(root) else str(fixture_path),
                     "input_sha256": sha256(fixture_path),
                     "archived_review_input_path": Path(archived_raw).as_posix(),
                     "manifest_spec_path": target_path.relative_to(root).as_posix() if target_path.is_relative_to(root) else str(target_path),
                     "manifest_spec_sha256": sha256(target_path), "recompiled_spec_sha256": sha256(compiled_path),
                     "status": "match" if not delta else "mismatch", "differences": delta}
            entries.append(entry)
            if delta:
                flat_differences.append({"id": row["id"], "differences": delta})

        record = {"schema": SCHEMA,
                  "scope": "Recompile canonical single-slide fixtures selected by exact slide ID with the current adapter and compare to manifest-selected compose slides. Ignore only chrome/page.text. Archived review input paths are retained as provenance. This is reproducibility evidence, not native or visual qualification.",
                  "checked": len(entries), "matched": sum(e["status"] == "match" for e in entries),
                  "ignored_only": ["slides[].canvas[id=chrome/page].text"],
                  "compiler_sha256": compiler_digest, "review_manifest": review_path.relative_to(root).as_posix() if review_path.is_relative_to(root) else str(review_path),
                  "review_manifest_sha256": sha256(review_path), "differences": flat_differences, "entries": entries}
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(record, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return record


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--review", type=Path, default=Path("samples/adaptive/review.json"))
    parser.add_argument("--out", type=Path, default=Path("planning/adaptive/reproducibility.json"))
    parser.add_argument("--compiler", type=Path, help="Existing pptxadapt binary; default builds one temporarily from this checkout")
    args = parser.parse_args()
    record = run(args.root, args.review, args.out, args.compiler)
    print(f"checked {record['checked']} examples: {record['matched']} matched; {len(record['differences'])} differed")
    if record["differences"]:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
