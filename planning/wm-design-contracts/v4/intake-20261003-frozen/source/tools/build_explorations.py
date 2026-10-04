#!/usr/bin/env python3
"""Build the self-contained exploration boards in explorations/.

Inlines official brand assets (from the branding folder) and the template spec
so each page can be published as a single HTML file.

Usage: python3 tools/build_explorations.py [--branding ~/Documents/branding]
"""
import argparse
import base64
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent
EXP = ROOT / "explorations"

MARK_FILES = {
    "underscore": "graphics/underscore/wm_handrawn_underscore_rgb_240912.svg",
    "circle": "graphics/circle/wm_handrawn_circle_rgb_240912.svg",
    "arrow-right-angle": "graphics/arrows/wm_handrawn_right_angle_arrow_rgb_240912.svg",
    "arrow-dashed": "graphics/arrows/wm_handrawn_dashed_arrow_rgb_240912.svg",
    "arrow-connecting": "graphics/arrows/wm_handrawn_connecting_arrow_rgb_240912.svg",
    "arrow-double": "graphics/arrows/wm_handrawn_double_arrow_rgb_240912.svg",
    "spark": "graphics/spark/wm_handrawn_spark_rgb_240912.svg",
    # straight arrow (points right; flipX for left): a level arrow without the hand-drawn arrows' strong curl
    "arrow-straight": "graphics/arrows/left-facing-arrow-navy.svg",
}
# Photo name used in slide JSON ("photo": "<name>") -> file in the branding photo library.
PHOTOS = {
    "photo-technician": "West Monroe Photos/Industry/Practices/Consumer Industrial Products/industrial-technician-reviewing-equipment-data-laptop-1.jpeg",
    "photo-warehouse": "West Monroe Photos/Industry/Practices/Consumer Industrial Products/warehouse-operations-manager-inventory-tablet-1.jpeg",
    "photo-clinical-team": "West Monroe Photos/Industry/Practices/Healthcare/clinical-team-reviewing-care-data-tablet-1.jpeg",
    "photo-clinical-leaders": "West Monroe Photos/Industry/Practices/Healthcare/clinical-leaders-reviewing-tablet-balcony.jpg",
    "photo-clinician-data": "West Monroe Photos/Industry/Practices/Healthcare/clinician-reviewing-patient-data-dual-monitors-1.jpeg",
    "photo-corridor": "West Monroe Photos/Industry/Practices/Healthcare/clinical-staff-collaborating-hospital-corridor-1.jpg",
    "photo-team-meeting": "West Monroe Photos/Industry/Practices/Organization People Change/team-meeting-glass-conference-room-1.jpeg",
    "photo-working-session": "West Monroe Photos/Industry/Practices/Organization People Change/colleagues-in-focused-working-session-1.jpg",
    "photo-executive": "West Monroe Photos/Industry/Practices/Organization People Change/executive-team-reviewing-documents-glass-conference-room-1.jpeg",
    "photo-abstract-cubes": "West Monroe Photos/Abstract/Practices/Technology Experience/abstract-blue-cubic-architecture-1.jpeg",
    "photo-abstract-grid": "West Monroe Photos/Abstract/Practices/Technology Experience/abstract-illuminated-data-grid-1.jpeg",
    "photo-abstract-blocks": "West Monroe Photos/Abstract/Practices/Technology Experience/abstract-blue-digital-blocks-1.jpeg",
    "photo-abstract-led": "West Monroe Photos/Abstract/Trending Topics/Artificial Intelligence/abstract-blue-led-data-pattern-1.jpeg",
    # Owner-supplied headshot (absolute path, outside the branding folder): the sample portrait for bios.
    "photo-headshot": "~/Documents/career/profile_image_1.png",
    # Tight face crop of the same image for small portraits (person tiles, bio cards): (height, width, top, left) in pixels.
    "photo-headshot-face": ("~/Documents/career/profile_image_1.png", (720, 720, 150, 201)),
    "photo-stethoscope": "West Monroe Photos/Abstract/Practices/Healthcare/stethoscope-medical-care-copy-space.jpeg",
}


def data_uri(path: pathlib.Path, mime: str) -> str:
    return f"data:{mime};base64," + base64.b64encode(path.read_bytes()).decode()


def mark_paths(svg: pathlib.Path) -> list[str]:
    """Filled artwork paths from the official SVG's #artwork group."""
    text = svg.read_text()
    # The artwork group can contain nested groups; take everything up to the hidden stroke group.
    start = text.index('<g id="artwork">')
    end = text.find('<g id="stroke"', start)
    group = text[start:end if end > 0 else len(text)]
    return [re.sub(r"\s+", " ", d) for d in re.findall(r'<path[^>]*\sd="([^"]*)"', group)]


CACHE = ROOT / "tools/.cache"


def photo_uri(src: pathlib.Path, tmp: pathlib.Path, crop=None) -> str:
    """Resized JPEG data URI, cached by source path, modification time and crop."""
    CACHE.mkdir(exist_ok=True)
    tag = "-c" + "x".join(map(str, crop)) if crop else ""
    key = CACHE / (src.stem + "-" + str(int(src.stat().st_mtime)) + tag + ".b64")
    if key.exists():
        return key.read_text()
    if crop:
        cropped = tmp / (src.stem + tag + src.suffix)
        h, w, top, left = crop
        subprocess.run(["sips", "-c", str(h), str(w), "--cropOffset", str(top), str(left), str(src), "--out", str(cropped)],
                       check=True, capture_output=True)
        src = cropped
    out = tmp / (src.stem + ".jpg")
    subprocess.run(["sips", "-Z", "900", "-s", "format", "jpeg", "-s", "formatOptions", "72", str(src), "--out", str(out)],
                   check=True, capture_output=True)
    uri = data_uri(out, "image/jpeg")
    key.write_text(uri)
    return uri


def icon_library(brand: pathlib.Path) -> dict:
    """The brand icon library (branding/icons): one geometry per icon (with its own viewBox), colored at render time.

    The navy, magenta and white folders hold the same artwork with different fills, so only the navy
    geometry is inlined. Keys are the index names; a repeated name gets a numeric suffix (database-2).
    """
    import csv
    root = brand / "icons"
    if not (root / "index.csv").exists():
        return {}
    out, seen = {}, {}
    for row in csv.DictReader((root / "index.csv").open()):
        svg = (root / row["navy_file"]).read_text()
        vb = (re.search(r'viewBox="([^"]+)"', svg) or [None, "0 0 72 72"])[1]
        body = re.sub(r"<style>.*?</style>", "", svg[svg.index(">") + 1:svg.rindex("</svg>")], flags=re.S)
        body = re.sub(r'\s(class|fill)="[^"]*"', "", body)
        body = re.sub(r"\s+", " ", body).strip()
        name = row["name"]
        seen[name] = seen.get(name, 0) + 1
        key = name if seen[name] == 1 else f"{name}-{seen[name]}"
        out[key] = {"id": row["id"], "desc": row["description"], "svg": body, "vb": vb}
    return out


def template_entries() -> list:
    """Every template in templates/library, tagged with its family (the board groups by it). A file that does not parse is skipped."""
    out = []
    order = ["covers", "core", "argument", "evidence", "diagrams", "heatmaps", "status", "venn", "architecture", "solution", "lifecycle", "software", "modernization", "maturity", "decisions", "approach", "team-curves", "runbooks", "change", "commercials", "value", "team", "about", "proof"]
    files = sorted((ROOT / "templates/library").glob("[!_]*.json"), key=lambda f: order.index(f.stem) if f.stem in order else 99)
    skip = set(filter(None, os.environ.get("WMDS_SKIP_FAMILIES", "").split(",")))   # e.g. a family an agent is still writing
    for f in files:
        if f.stem in skip:
            continue
        try:
            data = json.loads(f.read_text())
        except json.JSONDecodeError as e:
            print(f"skipping {f.relative_to(ROOT)}: {e}")
            continue
        out += [dict(x, _batch=data.get("family", f.stem), _batchName=data.get("name", f.stem)) for x in data.get("templates", [])]
    return out


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--branding", default="~/Documents/branding")
    ap.add_argument("--out", help="write the built pages to this folder instead of explorations/")
    ap.add_argument("--only", help="build only this board, e.g. components")
    args = ap.parse_args()
    subprocess.run([sys.executable, str(ROOT / "tools/build_catalog.py")], check=True, capture_output=True)
    brand = pathlib.Path(args.branding).expanduser()
    assets = brand / "assets"

    subs = {
        "HL": data_uri(assets / "graphics/highlight/wm_highlight_1_rgb_240912.png", "image/png"),
        "HL2": data_uri(assets / "graphics/highlight/wm_highlight_2_rgb_240912.png", "image/png"),
        "HL3": data_uri(assets / "graphics/highlight/wm_highlight_3_rgb_240912.png", "image/png"),
        "HL4": data_uri(assets / "graphics/highlight/wm_highlight_4_rgb_240912.png", "image/png"),
        "LOGO": data_uri(assets / "logos/primary/wm_h_pos_clr_rgb_august2024.svg", "image/svg+xml"),
        "LOGO_POS": data_uri(assets / "logos/primary/wm_h_pos_clr_rgb_august2024.svg", "image/svg+xml"),
        "LOGO_WHITE": data_uri(assets / "logos/primary/wm_h_rev_wht_rgb_august2024.svg", "image/svg+xml"),
        "LOGO_REV": data_uri(assets / "logos/primary/wm_h_rev_wht_rgb_august2024.svg", "image/svg+xml"),
        "TAGLINE_REV": data_uri(assets / "logos/primary/tagline_neg_leftalign.svg", "image/svg+xml"),
        "MARKS": json.dumps({k: mark_paths(assets / v) for k, v in MARK_FILES.items()}),
        "TOKENS": json.dumps(json.loads((ROOT / "tokens/v0/tokens.json").read_text())).replace("</", "<\\/"),
        "FRAMES": json.dumps(json.loads((ROOT / "frames/v0/frames.json").read_text())).replace("</", "<\\/"),
        "CATALOG": json.dumps(json.loads((ROOT / "templates/catalog.json").read_text())).replace("</", "<\\/"),
        "TEMPLATES_V1": json.dumps(template_entries()).replace("</", "<\\/"),
        "ICONLIB": json.dumps(icon_library(brand)).replace("</", "<\\/"),
        "COMPONENTS": json.dumps(json.loads((ROOT / "components/v0/components.json").read_text())).replace("</", "<\\/"),
    }
    with tempfile.TemporaryDirectory() as tmp:
        def resolve(rel):
            return pathlib.Path(rel).expanduser() if rel.startswith("~") else brand / rel
        subs["PHOTOS"] = json.dumps({name: photo_uri(resolve(v[0]), pathlib.Path(tmp), v[1]) if isinstance(v, tuple) else photo_uri(resolve(v), pathlib.Path(tmp))
                                     for name, v in PHOTOS.items()})

    for src in sorted(EXP.glob((args.only or "*") + ".src.html")):
        html = src.read_text()
        for key, val in subs.items():
            html = html.replace("{{" + key + "}}", val)
        left = re.findall(r"\{\{[A-Z_]+\}\}", html)
        if left:
            raise SystemExit(f"{src.name}: unresolved placeholders {sorted(set(left))}")
        out_dir = pathlib.Path(args.out).expanduser() if args.out else EXP
        out_dir.mkdir(parents=True, exist_ok=True)
        out = out_dir / src.name.replace(".src.html", ".html")
        out.write_text(html)
        print(f"{out}  {len(html) / 1024:.0f} KB")


if __name__ == "__main__":
    main()
