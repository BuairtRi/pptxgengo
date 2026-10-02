#!/usr/bin/env python3
"""Merge feature audits against the exact canonical WMDS template inventory.

This is a planning ledger, not executable renderer capability discovery or a
native qualification check. Audit details remain attributed to their track.
"""

import argparse
import hashlib
import json
from pathlib import Path
import re


def unique_object(pairs):
    value = {}
    for key, item in pairs:
        if key in value:
            raise ValueError(f"duplicate JSON property: {key}")
        value[key] = item
    return value


def load(path):
    return json.loads(path.read_text(), object_pairs_hook=unique_object)


def build(repo, source, tracks, catalog_path=None, source_receipts=None, bound_receipts=None,
          native_review=None):
    inventory = {}
    source_files = []
    for path in sorted((source / "templates/library").glob("*.json")):
        if path.name.startswith("_"):
            continue
        catalog = load(path)
        if catalog["schema"] != "wmds.templates.v2":
            raise ValueError(f"unsupported source schema: {path}")
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        source_file = str(path.relative_to(source))
        source_files.append({"path": source_file, "sha256": digest})
        for entry in catalog["templates"]:
            key = entry["id"] + "/" + entry["variant"]
            if key in inventory:
                raise ValueError(f"duplicate source template: {key}")
            inventory[key] = {
                "key": key, "family": catalog["family"], "tier": entry["tier"],
                "source_file": source_file, "source_sha256": digest,
            }
    if len(inventory) != 97:
        raise ValueError("source inventory changed; migrate the completion baseline")

    bindings = (repo / "internal/wmdesign/template_bindings.go").read_text()
    declaration = re.search(r"var executableTemplateKeys = \[\]string\{([^}]+)\}", bindings)
    if not declaration:
        raise ValueError("cannot read executable adapter list; update ledger builder")
    bound = set(re.findall(r'"([^"]+)"', declaration[1]))
    legacy_bound = set(bound)
    catalog = {}
    if catalog_path:
        catalog = {entry["key"]: entry for entry in load(catalog_path)}
        if catalog.keys() != inventory.keys():
            raise ValueError("runtime catalog does not match canonical inventory")
        for key, entry in catalog.items():
            for field in ("source_file", "source_sha256"):
                if entry[field] != inventory[key][field]:
                    raise ValueError(f"catalog/source mismatch: {key}/{field}")
        bound = set(catalog)
    evidence = {}
    for kind, path in (("source_generation", source_receipts), ("bound_generation", bound_receipts)):
        if path:
            rows = load(path)
            if {r["template"] for r in rows} != inventory.keys() or len(rows) != 97:
                raise ValueError(f"incomplete or duplicate generation receipts: {path}")
            for row in rows:
                evidence.setdefault(row["template"], {})[kind] = {
                    "receipt": str(path.relative_to(repo)), **row}
    if native_review:
        review = load(native_review)
        rows = review["templates"]
        if len(rows) != 97 or {r["template"] for r in rows} != inventory.keys():
            raise ValueError("native review does not cover the canonical inventory")
        artifact = repo / review["artifact"]
        if hashlib.sha256(artifact.read_bytes()).hexdigest() != review["sha256"]:
            raise ValueError("native review artifact hash changed")
        if review["native_open"] != "accepted_without_repair":
            raise ValueError("native review does not record package acceptance")
        for row in rows:
            evidence.setdefault(row["template"], {})["exact_source_specimen_review"] = {
                "receipt": str(native_review.relative_to(repo)),
                "page": row["page"], "native_open": review["native_open"],
                "visual_review": review["visual_review"],
                "qualified_envelope": False,
            }
    if not bound.issubset(inventory):
        raise ValueError("executable adapter references unknown source template")

    audited = {}
    track_records = []
    required = {"key", "family", "tier", "source_file", "source_sha256",
                "node_types", "features", "dependencies", "binding_work",
                "source_resolutions", "status"}
    for path in tracks:
        track = load(path)
        entries = track["templates"]
        if track["template_count"] != len(entries):
            raise ValueError(f"track count mismatch: {path}")
        track_records.append({
            "path": str(path.relative_to(repo)),
            "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "families": track["families"],
            "template_count": len(entries),
            "dependencies_catalog": track.get("dependencies_catalog", {}),
        })
        for entry in entries:
            key = entry["key"]
            if key in audited or key not in inventory:
                raise ValueError(f"duplicate or unknown audit template: {key}")
            if not required.issubset(entry):
                raise ValueError(f"missing audit fields for {key}: {required - entry.keys()}")
            for field, expected in inventory[key].items():
                if entry[field] != expected:
                    raise ValueError(f"audit/source mismatch: {key}/{field}")
            if entry["family"] not in track["families"]:
                raise ValueError(f"audit family not owned by track: {key}")
            audited[key] = dict(entry)
            audited[key]["audit_status"] = entry["status"]
            audited[key]["status"] = "implemented_bound" if key in bound else "inventoried"
            audited[key]["qualified_envelope"] = False
            audited[key]["audit_track"] = str(path.relative_to(repo))
            if key in catalog:
                audited[key]["content_contract"] = catalog[key]["content_contract"]
                audited[key]["render_status"] = catalog[key]["render_status"]
            audited[key].update(evidence.get(key, {}))
            # Current exact specimen evidence is intentionally separate from reuse.
            if key in legacy_bound:
                audited[key]["existing_specimen_receipt"] = (
                    "samples/wmds-templates-20261002/final/native-review.json")
    if audited.keys() != inventory.keys():
        raise ValueError(f"incomplete coverage: {sorted(inventory.keys() - audited.keys())}")

    families = []
    for family in sorted({e["family"] for e in inventory.values()}):
        keys = {k for k, e in inventory.items() if e["family"] == family}
        families.append({"family": family, "total": len(keys),
                         "implemented_bound": len(keys & bound),
                         "remaining": len(keys - bound)})
    return {
        "schema": "pptxgengo.wmds-template-coverage.v1", "date": "2026-10-02",
        "purpose": "Feature audit and completion planning; not runtime capability detection.",
        "source_files": source_files, "tracks": track_records,
        "summary": {"total": len(inventory), "implemented_bound": len(bound),
                    "remaining": len(inventory) - len(bound),
                    "core": sum(e["tier"] == "core" for e in inventory.values()),
                    "working": sum(e["tier"] == "working" for e in inventory.values())},
        "families": families,
        "qualification_note": "Audit does not confer native verification, visual review or a qualified content envelope.",
        "templates": [audited[k] for k in sorted(audited)],
    }


def main():
    repo = Path(__file__).resolve().parents[1]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path,
                        default=repo / "library/wm-design-system/v1/source")
    parser.add_argument("--out", type=Path,
                        default=repo / "planning/wm-design-contracts/v1/template-coverage.json")
    parser.add_argument("--catalog", type=Path, help="pptxdesign library-catalog JSON")
    parser.add_argument("--source-receipts", type=Path)
    parser.add_argument("--bound-receipts", type=Path)
    parser.add_argument("--native-review", type=Path,
                        help="hashed native acceptance and exact source specimen review receipt")
    args = parser.parse_args()
    folder = repo / "planning/wm-design-contracts/v1"
    tracks = [folder / f"coverage-{name}.json"
              for name in ("primitives", "data", "diagrams")]
    result = build(repo, args.source.resolve(), tracks,
                   args.catalog.resolve() if args.catalog else None,
                   args.source_receipts.resolve() if args.source_receipts else None,
                   args.bound_receipts.resolve() if args.bound_receipts else None,
                   args.native_review.resolve() if args.native_review else None)
    args.out.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n")
    print(json.dumps({"output": str(args.out), **result["summary"]}))


if __name__ == "__main__":
    main()
