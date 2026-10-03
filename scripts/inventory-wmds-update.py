#!/usr/bin/env python3
"""Compare a live WMDS library to the frozen bundle; inventory only, no qualification."""
import argparse
import collections
import hashlib
import json
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()


def templates(root):
    result = {}
    for path in sorted((root / "templates/library").glob("*.json")):
        if path.name.startswith("_"):
            continue
        family = read(path)
        for template in family["templates"]:
            key = template["id"] + "/" + template["variant"]
            if key in result:
                raise ValueError("Duplicate template: " + key)
            result[key] = (family["family"], template, path.relative_to(root).as_posix())
    return result


def nodes(value):
    """Walk render nodes, excluding table data and column type discriminators."""
    if isinstance(value, list):
        for child in value:
            yield from nodes(child)
    elif isinstance(value, dict):
        if "type" in value:
            yield value
        for field, child in value.items():
            if field not in {"rows", "cols"}:
                yield from nodes(child)


def features(template):
    slide = template["slide"]
    result = []
    if slide.get("split"):
        result.append("split-frame")
    if slide.get("rail") == "nav":
        result.append("source-nav-binding")
    if slide.get("titleLines", 1) == 3:
        result.append("three-line-split-title")
    for node in nodes(slide):
        if node.get("type") == "table" and any(c.get("type") == "checkbox" for c in node.get("cols", [])):
            result.append("table-checkbox")
        if node.get("type") == "chart" and node.get("key") == "markers":
            result.append("quadrant-markers")
        if node.get("type") == "annotation":
            result.append("named-target-annotation")
    return sorted(set(result))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=pathlib.Path, default=ROOT.parent / "wm-design-system")
    parser.add_argument("--baseline", type=pathlib.Path, default=ROOT / "library/wm-design-system/v1/source")
    parser.add_argument("--out", type=pathlib.Path, default=ROOT / "planning/wm-design-contracts/v2/source-update-2026-10-02.json")
    args = parser.parse_args()
    old, new = templates(args.baseline), templates(args.source)
    notes_path = args.source / "templates/change-notes.json"
    notes = read(notes_path) if notes_path.exists() else {}
    entries = []
    for key in sorted(set(old) | set(new)):
        before = old.get(key)
        after = new.get(key)
        template = (after or before)[1]
        if not before:
            change = "added"
        elif not after:
            change = "removed"
        elif template.get("status") == "deprecated" and before[1].get("status") != "deprecated":
            change = "deprecated"
        elif before[1] != after[1]:
            change = "revised"
        else:
            change = "unchanged"
        fields = sorted(k for k in set(before[1]) | set(after[1]) if before[1].get(k) != after[1].get(k)) if before and after else []
        entries.append(dict(key=key, family=(after or before)[0], change=change,
                            status=template.get("status", "active"), replaced_by=template.get("replacedBy"),
                            old_template_sha256=digest(before[1]) if before else None,
                            new_template_sha256=digest(after[1]) if after else None,
                            old_file_sha256=sha(args.baseline / before[2]) if before else None,
                            new_file_sha256=sha(args.source / after[2]) if after else None,
                            source_file=(after or before)[2], changed_fields=fields,
                            required_update_features=features(template) if after else [],
                            design_notes=notes.get(key),
                            implementation_state="existing_frozen_revision" if change == "unchanged" else "migration_pending",
                            native_review_state="new_revision_not_reviewed" if change != "unchanged" else "prior_revision_evidence_only"))
    source_paths = ["tokens/v0/tokens.json", "components/v0/components.json", "frames/v0/frames.json",
                    "templates/catalog.json", "templates/changes.json", "templates/change-notes.json",
                    "docs/authoring-reference.md", "explorations/components.src.html"]
    source_paths += sorted({v[2] for v in new.values()})
    files = [dict(path=p, old_sha256=sha(args.baseline / p) if (args.baseline / p).exists() else None,
                  new_sha256=sha(args.source / p)) for p in source_paths if (args.source / p).exists()]
    result = dict(schema="pptxgengo.wmds-update-inventory.v1", scope="Source and code inspection only; no implementation or native qualification inferred.",
                  source_root=str(args.source.resolve()), baseline_root=str(args.baseline.resolve()),
                  source_commit=subprocess.check_output(["git", "-C", str(args.source), "rev-parse", "HEAD"], text=True).strip(),
                  source_dirty=bool(subprocess.check_output(["git", "-C", str(args.source), "status", "--porcelain"], text=True).strip()),
                  old_count=len(old), new_count=len(new),
                  summary=dict(collections.Counter(e["change"] for e in entries)),
                  added_by_family=dict(sorted(collections.Counter(e["family"] for e in entries if e["change"] == "added").items())),
                  feature_dependents={f: [e["key"] for e in entries if f in e["required_update_features"]] for f in sorted({f for e in entries for f in e["required_update_features"]})},
                  files=files, templates=entries)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n")
    print(json.dumps({k: result[k] for k in ["source_commit", "source_dirty", "old_count", "new_count", "summary", "added_by_family"]}, indent=2))


if __name__ == "__main__":
    main()
