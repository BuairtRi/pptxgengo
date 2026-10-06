import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from package_files import clone_or_copy, file_digest, validate_asset_gallery, validate_design_docs


class PackageFilesTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        self.source = self.root / "source.jpg"
        self.destination = self.root / "stage.jpg"
        self.payload = b"registered artwork\x00" * 100000
        self.source.write_bytes(self.payload)

    def test_streamed_hash_matches_full_payload(self):
        self.assertEqual(file_digest(self.source), hashlib.sha256(self.payload).hexdigest())

    def test_actual_copy_preserves_independent_source_and_stage(self):
        method = clone_or_copy(self.source, self.destination)
        self.assertIn(method, ("cloned", "copied"))
        self.assertEqual(file_digest(self.destination), file_digest(self.source))
        self.assertNotEqual(self.source.stat().st_ino, self.destination.stat().st_ino)
        self.source.write_bytes(b"later source edit")
        self.assertEqual(self.destination.read_bytes(), self.payload)
        self.destination.write_bytes(b"later stage edit")
        self.assertEqual(self.source.read_bytes(), b"later source edit")

    def test_failed_clone_cleans_partial_destination_before_copy(self):
        def failed_clone(*args, **kwargs):
            self.destination.write_bytes(b"partial")
            return subprocess.CompletedProcess(args[0], 1)

        with patch("package_files.sys.platform", "darwin"), patch("package_files.subprocess.run", side_effect=failed_clone):
            self.assertEqual(clone_or_copy(self.source, self.destination), "copied")
        self.assertEqual(self.destination.read_bytes(), self.payload)

    def test_existing_destination_is_preserved(self):
        self.destination.write_bytes(b"existing release")
        with self.assertRaises(FileExistsError):
            clone_or_copy(self.source, self.destination)
        self.assertEqual(self.destination.read_bytes(), b"existing release")

    def gallery_fixture(self):
        (self.root / "index.html").write_text("asset gallery")
        metadata = {"title": "Existing photo title", "keywords": ["workshop"]}
        registry = [{"key": "photo/library/test", "path": "West Monroe Photos/test.jpg", "sha256": file_digest(self.source)}]
        photos = {"schema": "pptxgengo.photo-registry.v1", "photos": [{"path": registry[0]["path"], "sha256": registry[0]["sha256"], "metadata": metadata}]}
        items = [{"id": registry[0]["key"], "source_metadata": metadata, "variants": [{"id": registry[0]["key"], "path": registry[0]["path"], "sha256": registry[0]["sha256"], "thumbnail_path": "source.jpg", "thumbnail_sha256": registry[0]["sha256"], "thumbnail_state": "derived_go_preview_from_verified_original"}]}]
        (self.root / "assets.json").write_text(json.dumps(items))
        return registry, photos, items

    def test_gallery_pins_metadata_and_thumbnail_hashes(self):
        registry, photos, _ = self.gallery_fixture()
        self.assertEqual(validate_asset_gallery(self.root, registry, photos), (1, 1))
        self.source.write_bytes(b"altered preview")
        with self.assertRaisesRegex(ValueError, "thumbnail drift"):
            validate_asset_gallery(self.root, registry, photos)

    def test_gallery_rejects_stale_sidecar_metadata(self):
        registry, photos, items = self.gallery_fixture()
        items[0]["source_metadata"] = {"title": "stale metadata"}
        (self.root / "assets.json").write_text(json.dumps(items))
        with self.assertRaisesRegex(ValueError, "photo metadata mismatch"):
            validate_asset_gallery(self.root, registry, photos)

    def test_gallery_rejects_missing_variant_and_escape(self):
        registry, photos, items = self.gallery_fixture()
        (self.root / "assets.json").write_text("[]")
        with self.assertRaisesRegex(ValueError, "incomplete"):
            validate_asset_gallery(self.root, registry, photos)
        items[0]["variants"][0]["thumbnail_path"] = "../escape.jpg"
        (self.root / "assets.json").write_text(json.dumps(items))
        with self.assertRaisesRegex(ValueError, "nonlocal"):
            validate_asset_gallery(self.root, registry, photos)

    def docs_fixture(self):
        site, bundle = self.root / "docs", self.root / "bundle"
        (site / "assets").mkdir(parents=True)
        catalog = {"templates": [{"key": "workshop/test"}], "families": ["workshops"]}
        objects = {
            "tokens": ("tokens/v0/tokens.json", {"styles": []}),
            "data": ("components/v0/components.json", {"components": [{"id": "test"}]}),
            "frames": ("frames/v0/frames.json", {"rails": []}),
            "catalog": ("templates/catalog.json", catalog),
        }
        html = ""
        for element, (relative, obj) in objects.items():
            path = bundle / "source" / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(json.dumps(obj))
            html += f'<script id="{element}" type="application/json">{json.dumps(obj)}</script>'
        payload = b"<svg>test</svg>"
        name = hashlib.sha256(payload).hexdigest()[:12] + ".svg"
        (site / "assets" / name).write_bytes(payload)
        (site / "index.html").write_text(html + f'<img src="assets/{name}">')
        (site / "catalog.json").write_text(json.dumps(catalog))
        for name in ("changes.json", "CHANGELOG.md", "web-CHANGELOG.md"):
            (site / name).write_text("{}")
        source = {"commit": "pinned", "dirty": False, "templates": 1,
                  "families": 1, "components": 1, "assets": [hashlib.sha256(payload).hexdigest()[:12] + ".svg"]}
        (site / "SOURCE.json").write_text(json.dumps(source))
        (bundle / "bundle.json").write_text(json.dumps({"source_commit": "pinned", "template_count": 1}))
        return site, bundle, source

    def test_docs_validate_exact_source_and_assets(self):
        site, bundle, source = self.docs_fixture()
        self.assertEqual(validate_design_docs(site, bundle), source)

    def test_docs_reject_newer_or_dirty_source(self):
        site, bundle, source = self.docs_fixture()
        for change in ({"commit": "newer"}, {"dirty": True}):
            (site / "SOURCE.json").write_text(json.dumps(dict(source, **change)))
            with self.assertRaisesRegex(ValueError, "source differs"):
                validate_design_docs(site, bundle)

    def test_docs_reject_embedded_source_and_catalog_drift(self):
        site, bundle, _ = self.docs_fixture()
        html = (site / "index.html").read_text()
        (site / "index.html").write_text(html.replace('"rails": []', '"rails": ["changed"]'))
        with self.assertRaisesRegex(ValueError, "embedded source drift: frames"):
            validate_design_docs(site, bundle)
        (site / "index.html").write_text(html)
        (site / "catalog.json").write_text("{}")
        with self.assertRaisesRegex(ValueError, "catalog drift"):
            validate_design_docs(site, bundle)

    def test_docs_reject_asset_drift_and_missing_changelog(self):
        site, bundle, source = self.docs_fixture()
        asset = site / "assets" / source["assets"][0]
        payload = asset.read_bytes()
        asset.write_bytes(b"drift")
        with self.assertRaisesRegex(ValueError, "asset drift"):
            validate_design_docs(site, bundle)
        asset.write_bytes(payload)
        (site / "CHANGELOG.md").unlink()
        with self.assertRaisesRegex(ValueError, "missing: CHANGELOG.md"):
            validate_design_docs(site, bundle)


if __name__ == "__main__":
    unittest.main()
