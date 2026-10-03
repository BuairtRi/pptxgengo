#!/usr/bin/env python3
"""Freeze the reviewed round-4 WMDS source; leaves the existing v1 bundle intact."""
import hashlib
import json
import pathlib
import shutil
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]
SOURCE = ROOT.parent / "wm-design-system"
TARGET = ROOT / "library/wm-design-system/v2"
COMMIT = "7bdcaee5030a12275a1f881a8542f4d302d207df"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    commit = subprocess.check_output(["git", "-C", str(SOURCE), "rev-parse", "HEAD"], text=True).strip()
    dirty = subprocess.check_output(["git", "-C", str(SOURCE), "status", "--porcelain"], text=True).strip()
    if commit != COMMIT or dirty:
        raise SystemExit("Expected clean round-4 source commit " + COMMIT)
    if TARGET.exists():
        raise SystemExit("Refusing to replace existing bundle: " + str(TARGET))
    baseline = ROOT / "library/wm-design-system/v1"
    old_inventory = json.loads((baseline / "inventory.json").read_text())
    paths = {f["path"] for f in old_inventory["sources"]}
    paths.update(["templates/changes.json", "templates/change-notes.json", "templates/CHANGELOG.md"])
    files = []
    for path in sorted(paths):
        destination = TARGET / "source" / path
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(SOURCE / path, destination)
        files.append(dict(path=path, sha256=sha(destination)))
    for directory in ["fonts", "assets"]:
        shutil.copytree(baseline / directory, TARGET / directory)
    inventory = dict(schema="pptxgengo.wmds-source-inventory.v1", source_revision="wmds-library.v2",
                     source_commit=COMMIT, review_date="2026-10-02", source_root=str(SOURCE),
                     status="source_frozen_implementation_in_progress", sources=files,
                     coverage_note="167 source definitions; inventory does not imply rendering support or native review.")
    (TARGET / "inventory.json").write_text(json.dumps(inventory, indent=2) + "\n")
    manifest = json.loads((baseline / "bundle.json").read_text())
    manifest["source_revision"] = "wmds-library.v2"
    manifest["source_commit"] = COMMIT
    (TARGET / "bundle.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps(dict(inventory_sha256=sha(TARGET / "inventory.json"), bundle_sha256=sha(TARGET / "bundle.json")), indent=2))


if __name__ == "__main__":
    main()
