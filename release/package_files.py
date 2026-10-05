"""Bounded-memory hashes and independent artwork copies for local releases."""

import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys


def file_digest(path):
    digest = hashlib.sha256()
    with Path(path).open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def clone_or_copy(source, destination):
    """Use APFS copy-on-write cloning when available; preserve separate files."""
    source, destination = Path(source), Path(destination)
    if destination.exists() or destination.is_symlink():
        raise FileExistsError(destination)
    if sys.platform == "darwin":
        try:
            result = subprocess.run(
                ["/bin/cp", "-c", str(source), str(destination)],
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                check=False,
            )
            if result.returncode == 0:
                return "cloned"
        except OSError:
            pass
        # A failed clone may leave a partially copied destination behind.
        destination.unlink(missing_ok=True)
    shutil.copy2(source, destination)
    return "copied"


def validate_asset_gallery(directory, registry, photo_snapshot):
    """Verify retained gallery variants, sidecar metadata and thumbnail bytes."""
    directory = Path(directory)
    if not (directory / "index.html").is_file():
        raise ValueError("asset gallery HTML missing")
    items = json.loads((directory / "assets.json").read_text())
    expected = {item["key"]: (item["path"], item["sha256"]) for item in registry}
    if len(expected) != len(registry):
        raise ValueError("duplicate registered asset ID")
    if photo_snapshot.get("schema") != "pptxgengo.photo-registry.v1":
        raise ValueError("unsupported photo metadata snapshot")
    photos = {photo["path"]: photo for photo in photo_snapshot["photos"]}
    seen, concepts = {}, set()
    for item in items:
        if item["id"] in concepts:
            raise ValueError("duplicate asset gallery concept")
        concepts.add(item["id"])
        for variant in item["variants"]:
            key = variant["id"]
            pin = (variant["path"], variant["sha256"])
            if key in seen or expected.get(key) != pin:
                raise ValueError(f"asset gallery registry mismatch: {key}")
            seen[key] = pin
            photo = photos.get(variant["path"])
            if photo and (photo["sha256"] != variant["sha256"] or item.get("source_metadata") != photo["metadata"]):
                raise ValueError(f"asset gallery photo metadata mismatch: {key}")
            relative = Path(variant["thumbnail_path"])
            thumbnail = directory / relative
            if relative.is_absolute() or ".." in relative.parts or not thumbnail.resolve().is_relative_to(directory.resolve()):
                raise ValueError(f"nonlocal asset gallery thumbnail: {key}")
            if variant["thumbnail_state"] not in ("derived_go_preview_from_verified_original", "verified_registered_original"):
                raise ValueError(f"asset gallery thumbnail unverified: {key}")
            if file_digest(thumbnail) != variant["thumbnail_sha256"]:
                raise ValueError(f"asset gallery thumbnail drift: {key}")
    if seen != expected:
        raise ValueError("asset gallery is incomplete")
    return len(items), len(seen)
