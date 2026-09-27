#!/usr/bin/env python3
"""Export a frozen template bundle through PowerPoint in its approved folder.

This records render provenance, not visual acceptance or measured capacity.
Run sequentially: PowerPoint's document model is a shared resource.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("bundle", type=Path)
    parser.add_argument("--name", required=True, help="Unique native filename prefix")
    args = parser.parse_args()
    if not args.name or any(c not in "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_" for c in args.name):
        parser.error("name must contain only letters, numbers, hyphens and underscores")
    root = Path(__file__).resolve().parents[1]
    bundle = args.bundle.resolve()
    record = json.loads((bundle / "review-bundle.json").read_text())
    output = bundle / "native-render.json"
    if output.exists():
        parser.error("native-render.json already exists; use a new immutable bundle")
    native = root / "samples/visual-wave3"
    if not native.is_dir():
        parser.error("the established PowerPoint working directory is missing")
    jobs = []
    for deck in record["decks"]:
        source = (bundle / deck["path"]).resolve()
        if not source.is_relative_to(bundle):
            raise ValueError("Deck path escapes review bundle")
        if digest(source) != deck["sha256"]:
            raise ValueError(f"Deck changed: {source}")
        stem = args.name + "-" + source.stem.removesuffix("-review")
        pptx, pdf = native / (stem + ".pptx"), native / (stem + ".pdf")
        if pptx.exists() or pdf.exists():
            raise FileExistsError(stem)
        render = bundle / "render" / stem
        if render.exists():
            raise FileExistsError(render)
        jobs.append((deck, source, pptx, pdf, render))
    record["status"] = "native_render_in_progress"
    record["visual_review"] = "pending"
    def save():
        output.write_text(json.dumps(record, indent=2) + "\n")
    save()
    try:
        for deck, source, pptx, pdf, render in jobs:
            shutil.copy2(source, pptx)
            if digest(pptx) != deck["sha256"]:
                raise ValueError("Native copy hash mismatch")
            print(f"Exporting {pptx.name}", flush=True)
            subprocess.run(["osascript", str(root / "scripts/export-powerpoint.applescript"), str(pptx), str(pdf)], check=True, timeout=150)
            page_count = int(subprocess.check_output(["swift", "-e", 'import PDFKit; import Foundation; let d = PDFDocument(url: URL(fileURLWithPath: CommandLine.arguments[1]))!; print(d.pageCount)', str(pdf)], text=True).strip())
            if page_count != deck["expected_pdf_pages"]:
                raise ValueError(f"Expected {deck['expected_pdf_pages']} pages, got {page_count}")
            subprocess.run(["swift", str(root / "scripts/render-pdf.swift"), str(pdf), str(render), *map(str, range(1, page_count + 1))], check=True, timeout=150, stdout=subprocess.DEVNULL)
            deck.update(native_pptx=str(pptx.relative_to(root)), native_pdf=str(pdf.relative_to(root)), native_pdf_sha256=digest(pdf), render_directory=str(render.relative_to(root)), actual_pdf_pages=page_count)
            for page in deck["pages"]:
                png = render / f"slide-{page['pdf_page']:03d}.png"
                page.update(png=str(png.relative_to(root)), png_sha256=digest(png))
            save()
            print(f"Rendered {page_count} pages", flush=True)
    except Exception as error:
        record["status"] = "native_render_failed"
        record["failure"] = str(error)
        save()
        raise
    record["status"] = "native_rendered_pending_visual_review"
    save()
    print(output)


if __name__ == "__main__":
    main()
