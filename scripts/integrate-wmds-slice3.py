#!/usr/bin/env python3
"""Integrate the four independently generated added-family specimen packets."""
import argparse
import copy
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PACKET = ROOT / "samples/wmds-refresh-slice3-20261002"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--cli", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    inventory = json.loads((ROOT / "planning/wm-design-contracts/v2/source-update-2026-10-02.json").read_text())
    additions = [t for t in inventory["templates"] if t["change"] == "added"]
    by_key = {}
    base = None
    for group in ("approach-commercials", "proof-evidence", "team-solution", "argument-openers"):
        doc = json.loads((PACKET / "work" / group / "combined.foundation.json").read_text())
        if base is None:
            base = copy.deepcopy(doc)
        for slide in doc["slides"]:
            key = slide["template_binding"]["template"]
            by_key.setdefault(key, []).append(slide)
    expected = {t["key"] for t in additions}
    if set(by_key) != expected or any(len(pair) != 2 for pair in by_key.values()):
        raise ValueError("Expected exactly two specimens for every one of the 70 added templates")
    base["slides"] = []
    manifest = []
    for family in ("approach", "commercials", "proof", "evidence", "team", "solution", "argument", "openers"):
        for item in additions:
            if item["family"] != family:
                continue
            key = item["key"]
            pair = by_key[key]
            pages = []
            for index, raw in enumerate(pair):
                slide = copy.deepcopy(raw)
                kind = "source" if index == 0 else "alternate"
                slide["id"] = kind + "-" + key.replace("/", "-")
                slide["template_binding"]["slide_id"] = slide["id"]
                base["slides"].append(slide)
                pages.append(len(base["slides"]))
            manifest.append({"template": key, "family": family,
                             "source_page": pages[0], "alternate_page": pages[1],
                             "native_review": "pending"})
    PACKET.mkdir(parents=True, exist_ok=True)
    spec = PACKET / "reference.foundation.json"
    spec.write_text(json.dumps(base, indent=2) + "\n")
    (PACKET / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    subprocess.run([str(args.cli.resolve()), "build", "--bundle", "library/wm-design-system/v2",
                    "--engine", "wmds-go-foundation.v2", "--spec", str(spec),
                    "--out", str(args.out.resolve())], cwd=ROOT, check=True)


if __name__ == "__main__":
    main()
