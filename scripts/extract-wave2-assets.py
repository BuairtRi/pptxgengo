#!/usr/bin/env python3
"""Extract only hash-pinned media enumerated in Wave 2 candidate inventory."""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import shutil
import sys
import zipfile

REPO = Path(__file__).resolve().parent.parent
DEFAULT_INVENTORY = REPO / "library/component-contracts/visual-wave2-candidates.json"
DEFAULT_OUTPUT = REPO / "samples/visual-wave2/assets"
MANIFEST_SCHEMA = "pptxgengo.visual-wave2-assets.v1"


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def fail(message: str) -> "NoReturn":
    raise ValueError(message)


def source_part(asset: dict, candidate_id: str, index: int) -> list[tuple[str, str, dict]]:
    """Expand only explicitly enumerated PPTX part/hash fields; never infer."""
    parts = asset.get("pptx_part")
    digest = asset.get("sha256")
    if isinstance(parts, str) and isinstance(digest, str):
        return [(parts, digest, asset)]
    parts = asset.get("pptx_parts")
    hashes = asset.get("sha256")
    if isinstance(parts, list) and isinstance(hashes, list) and len(parts) == len(hashes):
        if all(isinstance(p, str) and isinstance(h, str) for p, h in zip(parts, hashes)):
            return [(p, h, asset) for p, h in zip(parts, hashes)]
    fail(f"{candidate_id}.assets[{index}] lacks explicit matching pptx_part(s)/sha256 field(s)")


def safe_media_part(value: str, candidate_id: str) -> str:
    path = PurePosixPath(value)
    if path.is_absolute() or path.as_posix() != value or ".." in path.parts or path.parts[:1] != ("ppt",) or path.parts[1:2] != ("media",):
        fail(f"{candidate_id}: unsafe or non-media package path: {value!r}")
    return path.as_posix()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--inventory", type=Path, default=DEFAULT_INVENTORY)
    parser.add_argument("--out", type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()
    inventory = args.inventory.resolve()
    output = args.out.resolve()

    try:
        data = json.loads(inventory.read_text(encoding="utf-8"))
        if data.get("schema_version") != 1 or data.get("status") != "candidate_inventory_not_approved":
            fail("unexpected candidate inventory schema/status")
        sources = data.get("sources")
        candidates = data.get("candidates")
        if not isinstance(sources, dict) or not isinstance(candidates, list):
            fail("candidate inventory lacks sources/candidates")

        # Validate each declared source file once before reading any package part.
        source_paths: dict[str, Path] = {}
        source_hashes: dict[str, str] = {}
        for source_id, source in sources.items():
            if not isinstance(source, dict) or not isinstance(source.get("path"), str) or not isinstance(source.get("sha256"), str):
                fail(f"source {source_id!r} lacks an explicit path/hash")
            source_path = (REPO / source["path"]).resolve()
            try:
                source_path.relative_to(REPO)
            except ValueError:
                fail(f"source {source_id!r} escapes repository: {source_path}")
            if not source_path.is_file():
                fail(f"source deck missing: {source_path}")
            actual_hash = sha256(source_path.read_bytes())
            if actual_hash != source["sha256"]:
                fail(f"source deck hash mismatch for {source_id}: expected {source['sha256']}, got {actual_hash}")
            source_paths[source_id] = source_path
            source_hashes[source_id] = actual_hash

        # Read and hash-check all declared assets before creating the output directory.
        occurrences: list[dict] = []
        for candidate in candidates:
            candidate_id = candidate.get("id")
            if not isinstance(candidate_id, str) or not candidate_id:
                fail("candidate missing id")
            candidate_sources = candidate.get("source_slides")
            if not isinstance(candidate_sources, list):
                fail(f"{candidate_id}: source_slides must be explicitly listed")
            for source_slide in candidate_sources:
                source_id = source_slide.partition(":")[0] if isinstance(source_slide, str) else ""
                if source_id not in source_paths:
                    fail(f"{candidate_id}: source slide refers to unknown source {source_slide!r}")
            asset_rows = candidate.get("assets")
            if not isinstance(asset_rows, list):
                fail(f"{candidate_id}: assets must be an explicit list (use [] when there are no image assets)")
            for asset_index, asset in enumerate(asset_rows):
                if not isinstance(asset, dict):
                    fail(f"{candidate_id}.assets[{asset_index}] must be an object")
                for part, expected_hash, metadata in source_part(asset, candidate_id, asset_index):
                    part = safe_media_part(part, candidate_id)
                    if len(expected_hash) != 64 or any(ch not in "0123456789abcdef" for ch in expected_hash):
                        fail(f"{candidate_id}: invalid SHA-256 for {part}")
                    # An asset entry can override source by source slide, otherwise candidate source must be unambiguous.
                    hinted_slide = asset.get("slide")
                    if hinted_slide is not None:
                        matches = [s for s in candidate_sources if s.endswith(f":{hinted_slide}")]
                        if len(matches) != 1:
                            fail(f"{candidate_id}: asset slide {hinted_slide!r} does not resolve to one declared source")
                        source_slide = matches[0]
                    elif len(candidate_sources) == 1:
                        source_slide = candidate_sources[0]
                    else:
                        # For multi-slide candidates, every enumerated part needs a slide pin.
                        fail(f"{candidate_id}: asset {part} must carry an explicit slide field for multi-slide source association")
                    source_id = source_slide.partition(":")[0]
                    deck = source_paths[source_id]
                    try:
                        with zipfile.ZipFile(deck) as package:
                            payload = package.read(part)
                    except (KeyError, zipfile.BadZipFile) as exc:
                        fail(f"{candidate_id}: cannot read {part} from {deck}: {exc}")
                    actual_hash = sha256(payload)
                    if actual_hash != expected_hash:
                        fail(f"media hash mismatch for {source_slide} {part}: expected {expected_hash}, got {actual_hash}")
                    occurrences.append({
                        "candidate_id": candidate_id,
                        "family": candidate.get("family"),
                        "source_slide": source_slide,
                        "source_id": source_id,
                        "source_deck": str(deck.relative_to(REPO)),
                        "source_deck_sha256": source_hashes[source_id],
                        "source_part": part,
                        "source_part_sha256": actual_hash,
                        "source_asset_metadata": metadata,
                        "bytes": payload,
                    })

        if output.exists():
            fail(f"output must be a new directory; refusing to overwrite: {output}")
        output.parent.mkdir(parents=True, exist_ok=True)
        output.mkdir()
        unique: dict[str, dict] = {}
        try:
            for item in occurrences:
                digest = item["source_part_sha256"]
                previous = unique.get(digest)
                if previous is not None:
                    if previous["bytes"] != item["bytes"]:
                        fail(f"SHA-256 collision while deduplicating {item['source_part']}")
                    continue
                suffix = Path(item["source_part"]).suffix.lower()
                filename = f"{digest}{suffix}"
                target = output / filename
                with target.open("xb") as stream:
                    stream.write(item["bytes"])
                unique[digest] = {"file": filename, "bytes": item["bytes"]}

            aliases = []
            for item in occurrences:
                aliases.append({k: v for k, v in item.items() if k != "bytes"} | {
                    "extracted_file": unique[item["source_part_sha256"]]["file"]
                })
            manifest = {
                "schema": MANIFEST_SCHEMA,
                "status": "extracted_unapproved_source_assets",
                "inventory": str(inventory.relative_to(REPO)) if inventory.is_relative_to(REPO) else str(inventory),
                "source_decks": {
                    key: {"path": str(path.relative_to(REPO)), "sha256": source_hashes[key]}
                    for key, path in source_paths.items()
                },
                "unique_binary_count": len(unique),
                "source_asset_association_count": len(aliases),
                "deduplication": "Exact SHA-256 with byte equality confirmation; every source/family association remains in aliases.",
                "conversions": [],
                "assets": [
                    {"file": item["file"], "sha256": digest, "size_bytes": len(item["bytes"])}
                    for digest, item in sorted(unique.items())
                ],
                "aliases": aliases,
            }
            with (output / "manifest.json").open("x", encoding="utf-8") as stream:
                json.dump(manifest, stream, ensure_ascii=False, indent=2)
                stream.write("\n")
        except Exception:
            shutil.rmtree(output)
            raise
        print(json.dumps({"output": str(output), "unique_binary_count": len(unique), "source_asset_association_count": len(occurrences), "manifest": str(output / "manifest.json")}, indent=2))
        return 0
    except (OSError, ValueError, json.JSONDecodeError) as exc:
        print(f"extract-wave2-assets: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
