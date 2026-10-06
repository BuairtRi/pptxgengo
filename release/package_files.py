"""Bounded-memory hashes and independent artwork copies for local releases."""

import hashlib
import json
import re
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


def validate_design_docs(directory, bundle_directory):
    """Keep the documentation board and native library pinned to the same source."""
    directory, bundle_directory = Path(directory), Path(bundle_directory)
    source = json.loads((directory / "SOURCE.json").read_text())
    bundle = json.loads((bundle_directory / "bundle.json").read_text())
    if source.get("dirty") is not False or source.get("commit") != bundle["source_commit"]:
        raise ValueError("design documentation source differs from the qualified native library")
    html = (directory / "index.html").read_text()
    for element, path in (("tokens", "tokens/v0/tokens.json"),
                          ("data", "components/v0/components.json"),
                          ("frames", "frames/v0/frames.json"),
                          ("catalog", "templates/catalog.json")):
        match = re.search(r'<script id="' + element + r'" type="application/json">(.*?)</script>', html, re.S)
        expected = json.loads((bundle_directory / "source" / path).read_text())
        if match is None or json.loads(match.group(1)) != expected:
            raise ValueError(f"design documentation embedded source drift: {element}")
    catalog = json.loads((directory / "catalog.json").read_text())
    if catalog != json.loads((bundle_directory / "source/templates/catalog.json").read_text()):
        raise ValueError("design documentation catalog drift")
    if source.get("templates") != bundle["template_count"] or len(catalog["templates"]) != source["templates"]:
        raise ValueError("design documentation template count mismatch")
    components = json.loads((bundle_directory / "source/components/v0/components.json").read_text())
    if source.get("families") != len(catalog["families"]) or source.get("components") != len(components["components"]):
        raise ValueError("design documentation inventory count mismatch")
    for path in ("changes.json", "CHANGELOG.md", "web-CHANGELOG.md"):
        if not (directory / path).is_file():
            raise ValueError(f"design documentation missing: {path}")
    assets = source.get("assets", [])
    if len(assets) != len(set(assets)):
        raise ValueError("duplicate design documentation asset")
    for name in assets:
        if not re.fullmatch(r"[0-9a-f]{12}\.(jpg|png|svg|webp|woff2)", name):
            raise ValueError(f"nonlocal design documentation asset: {name}")
        if not file_digest(directory / "assets" / name).startswith(name.split(".")[0]):
            raise ValueError(f"design documentation asset drift: {name}")
    references = set(re.findall(r"assets/([0-9a-f]{12}\.[a-z0-9]+)", html))
    if references != set(assets):
        raise ValueError("design documentation asset references differ from its manifest")
    return source
