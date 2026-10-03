#!/usr/bin/env python3
"""Generate the @wmds/slides package sources from the reference board.

The board (explorations/components.src.html) holds the one renderer: everything between its
`renderer:start` and `renderer:end` markers. This script lifts that range into
packages/slides/src/generated/renderer.js as a factory, scopes the board's stylesheet under
`.wm-root` into styles.css, embeds the data the renderer needs (tokens, frames, ink rules,
template library, assets, icons, marks), and writes:

  generated/renderer.js     export function createRenderer(env)  (the board's renderer, verbatim)
  generated/data.js         tokens, frames, ink rules, templates, catalog
  generated/styles.css      board CSS scoped under .wm-root, with highlight images inlined
  generated/nodes.ts        node-type -> component-name map and the inferred props interfaces
  docs/<Name>.md            per-component docs for the design agent (summary, rules, fields)

Usage: python3 tools/gen_renderer.py [--branding ~/Documents/branding]
"""
import argparse
import glob
import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
PKG = ROOT / "packages/slides"
GEN = PKG / "src/generated"
sys.path.insert(0, str(ROOT / "tools"))
import build_explorations as bx  # noqa: E402

# node type -> exported component name, and the library group it is documented under
NODES = {
    "slide": ("Slide", "Slides"),
    "text": ("Text", "Text and lists"), "rule": ("Rule", "Text and lists"), "grouplabel": ("GroupLabel", "Headings"),
    "numhead": ("NumberedHeading", "Headings"), "colhead": ("ColumnHeading", "Headings"), "textblock": ("TextBlock", "Text and lists"),
    "bullets": ("Bullets", "Text and lists"), "ol": ("NumberedList", "Text and lists"), "strongnum": ("StrongNumberList", "Text and lists"),
    "schedule": ("Schedule", "Text and lists"), "list": ("IndexList", "Text and lists"),
    "card": ("Card", "Cards"), "cardrow": ("CardRow", "Cards"),
    "metric": ("Metric", "Data points"), "callout": ("Callout", "Data points"), "pullquote": ("PullQuote", "Data points"),
    "legend": ("Legend", "Data points"), "gauge": ("Gauge", "Data points"), "indicators": ("Indicators", "Primitives"),
    "icons": ("IconRow", "Primitives"), "person": ("Person", "People"), "role": ("RoleBox", "People"),
    "stepper": ("Stepper", "Sequence"), "vstepper": ("VerticalStepper", "Sequence"), "phasehead": ("PhaseHeading", "Sequence"),
    "timeaxis": ("TimeAxis", "Sequence"),
    "preset": ("Shape", "Shapes"), "block": ("Block", "Shapes"), "chevron": ("Chevron", "Shapes"), "textarrow": ("TextArrow", "Shapes"),
    "connector": ("Connector", "Shapes"), "node": ("DiagramNode", "Shapes"), "frame": ("FrameBox", "Shapes"),
    "container": ("Container", "Architecture"), "layer": ("Layer", "Architecture"), "layerrow": ("LayerRow", "Architecture"),
    "pod": ("Pod", "Architecture"), "cylinder": ("Cylinder", "Architecture"), "device": ("Device", "Architecture"),
    "screen": ("Screen", "Architecture"), "plane": ("Plane", "Architecture"),
    "table": ("Table", "Composites"), "chart": ("Chart", "Composites"), "gantt": ("Gantt", "Composites"),
    "horizons": ("Horizons", "Composites"), "phases": ("Phases", "Composites"), "orgchart": ("OrgChart", "Composites"),
    "governance": ("GovernanceStack", "Composites"), "feesummary": ("FeeSummary", "Composites"), "swimlane": ("Swimlane", "Composites"),
    "cycle": ("Cycle", "Composites"), "pyramid": ("Pyramid", "Composites"), "beforeafter": ("BeforeAfter", "Composites"),
    "annotation": ("Annotation", "Composites"), "matrix": ("Matrix", "Composites"),
    "thumbnail": ("Thumbnail", "Media"), "mark": ("Mark", "Marks"), "whiteboard": ("Whiteboard", "Marks"),
    "square": ("Square", "Media"), "imageframe": ("ImageFrame", "Media"), "logo": ("Logo", "Media"), "art": ("Art", "Media"),
    "reviewnote": ("ReviewNote", "Collaboration"),
}
# fields of nested objects that are themselves node lists
NODE_LISTS = {"body", "nodes"}


# ---------- type inference from every real usage ----------
class Shape:
    def __init__(self):
        self.kinds = set()      # "number" | "string" | "boolean" | "null"
        self.obj = None         # dict[str, Shape] with counts
        self.obj_n = 0
        self.req = {}
        self.arr = None         # Shape of elements
        self.literals = set()

    def add(self, v, depth=0):
        if isinstance(v, bool):
            self.kinds.add("boolean")
        elif isinstance(v, (int, float)):
            self.kinds.add("number")
        elif isinstance(v, str):
            self.kinds.add("string")
            if len(v) <= 24 and re.fullmatch(r"[a-z0-9.\-]+", v):
                self.literals.add(v)
        elif v is None:
            self.kinds.add("null")
        elif isinstance(v, list):
            self.arr = self.arr or Shape()
            for x in v:
                self.arr.add(x, depth + 1)
            if not v:
                self.arr.kinds.add("unknown")
        elif isinstance(v, dict):
            if self.obj is None:
                self.obj = {}
            self.obj_n += 1
            for k, x in v.items():
                self.obj.setdefault(k, Shape()).add(x, depth + 1)
                self.req[k] = self.req.get(k, 0) + 1

    def ts(self, indent="  ", key=None, enum_keys=(), ntype=None):
        parts = []
        canon = PER_TYPE.get((ntype, key)) or (CANON.get(key) if ntype is not None or key in CANON else None)
        if "string" in self.kinds:
            if canon:
                parts.append(" | ".join(x if x == OPEN else json.dumps(x) for x in canon))
            else:
                parts.append("string")
        for k in ("number", "boolean"):
            if k in self.kinds:
                parts.append(k)
        if self.arr is not None:
            inner = self.arr.ts(indent, key, enum_keys, ntype)
            parts.append(f"Array<{inner}>")
        if self.obj is not None:
            lines = []
            for k, s in sorted(self.obj.items()):
                nm = k if re.fullmatch(r"[A-Za-z_$][\w$]*", k) else json.dumps(k)
                lines.append(f"{indent}  {nm}?: {s.ts(indent + '  ', k, enum_keys, '')};")
            parts.append("{\n" + "\n".join(lines) + f"\n{indent}}}" if lines else "Record<string, unknown>")
        if "unknown" in self.kinds and not parts:
            parts.append("unknown")
        return " | ".join(parts) if parts else "unknown"


STYLES = ["display", "title", "heading", "subhead", "lead", "body", "small", "source", "stat", "stat-sm", "number", "eyebrow", "label"]
INKS = ["display", "primary", "secondary", "emphasis", "callout"]
SURFACES = ["light", "subtle", "strong", "inverse", "deep", "callout"]
OPEN = "(string & {})"  # keeps autocompletion for the named values while allowing palette references like "series.2"
# Canonical value sets from the spec (authoring reference), so types never narrow to whatever the samples happened to use.
CANON = {
    "ink": INKS + [OPEN], "numInk": INKS + [OPEN], "keyInk": INKS + [OPEN], "markInk": INKS + [OPEN], "titleInk": INKS + [OPEN],
    "surface": SURFACES + ["outline"], "railSurface": ["inverse", "deep", "subtle"], "on": SURFACES,
    "emphasis": ["highlight", "underscore", "circle", "spark"], "rail": ["none", "nav", "left", "right"], "footer": ["compact", "tall"],
    "density": ["standard", "appendix"], "fade": ["corner", "right", "left", "bottom", "none"], "head": ["end", "start", "both", "none"],
    "elbow": ["h", "v"], "labels": ["above", "left"], "valign": ["top", "middle", "bottom"], "titleStyle": STYLES, "bodySize": ["small", "body"],
    "stamp": ["Draft for discussion", "Confidential", "Preliminary", "For internal use", OPEN], "labelPos": ["above", "on"],
}
PER_TYPE = {
    ("text", "style"): STYLES, ("block", "style"): STYLES, ("chevron", "style"): STYLES, ("textarrow", "style"): STYLES, ("matrix", "style"): STYLES,
    ("connector", "style"): ["solid", "dashed", "dotted"], ("container", "style"): ["region", "boundary", "layer", "frame", "external"],
    ("frame", "style"): ["solid", "dashed", "filled"], ("bullets", "size"): ["small", "body"], ("ol", "size"): ["small", "body"],
    ("strongnum", "size"): ["small", "body"], ("chart", "kind"): ["column", "bar", "line", "pie", "doughnut", "scatter", "quadrant"],
    ("thumbnail", "kind"): ["text", "table", "chart", "diagram"], ("mark", "mark"): ["underscore", "circle", "spark", "arrow-right-angle", "arrow-double", "arrow-connecting", "arrow-dashed"],
    ("logo", "variant"): ["pos", "rev"], ("preset", "shape"): ["rect", "circle", "homeplate", "chevron", "arrow", "diamond"],
    ("card", "titleStyle"): STYLES, ("cardrow", "numbering"): ["inline", "band", "corner"], ("textarrow", "dir"): ["right", "left", "up", "down"],
}
ENUMS = set()


def collect(samples):
    def walk(o):
        if isinstance(o, dict):
            t = o.get("type")
            if isinstance(t, str) and t in NODES and ("x" in o or "y" in o or t in ("slide", "connector", "art")):
                samples.setdefault(t, []).append(o)
            for k, v in o.items():
                walk(v)
        elif isinstance(o, list):
            for v in o:
                walk(v)
    walk(json.loads((ROOT / "components/v0/components.json").read_text()))
    walk(json.loads((ROOT / "frames/v0/frames.json").read_text()))
    for f in sorted((ROOT / "templates/library").glob("[!_]*.json")):
        walk(json.loads(f.read_text()))


def interfaces(samples):
    out = []
    for t, (name, _) in NODES.items():
        s = Shape()
        for o in samples.get(t, []):
            s.add({k: v for k, v in o.items() if k not in ("type", "_h") and not (t == "slide" and k == "body")})
        fields = s.obj or {}
        lines = []
        for k, sh in sorted(fields.items()):
            nm = k if re.fullmatch(r"[A-Za-z_$][\w$]*", k) else json.dumps(k)
            lines.append(f"  {nm}?: {sh.ts('  ', k, ENUMS, t)};")
        if t == "slide":
            lines.append("  /** Body nodes as data (the JSON form). Children components are appended after these. */\n  body?: Array<Record<string, unknown>>;")
        out.append((name, t, lines, len(samples.get(t, []))))
    return out


# ---------- css scoping ----------
def scope_css(css: str, prefix: str = ".wm-root") -> str:
    out, i = [], 0
    css = re.sub(r"/\*.*?\*/", "", css, flags=re.S)
    while i < len(css):
        j = css.find("{", i)
        if j < 0:
            break
        head = css[i:j].strip()
        if head.startswith("@media") or head.startswith("@supports"):
            depth, k = 1, j + 1
            while depth and k < len(css):
                depth += {"{": 1, "}": -1}.get(css[k], 0)
                k += 1
            out.append(head + " {" + scope_css(css[j + 1:k - 1], prefix) + "}")
            i = k
            continue
        k = css.find("}", j)
        body = css[j + 1:k]
        if head.startswith("@"):
            out.append(head + " {" + body + "}")
        else:
            sels = []
            for sel in head.split(","):
                sel = sel.strip()
                if not sel or sel in ("html", "body") or sel.startswith("body ") or sel.startswith("html "):
                    continue
                if sel == ":root":
                    sels.append(prefix)
                elif sel == "*":
                    sels.append(f"{prefix}, {prefix} *")
                else:
                    sels.append(f"{prefix} {sel}")
            if sels:
                out.append(", ".join(sels) + " {" + body + "}")
        i = k + 1
    return "\n".join(out)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--branding", default="~/Documents/branding")
    args = ap.parse_args()
    brand = pathlib.Path(args.branding).expanduser()
    a = brand / "assets"
    src = (ROOT / "explorations/components.src.html").read_text()
    m = re.search(r"// ===== renderer:start =====.*?\n(.*)\n\s*// ===== renderer:end =====", src, flags=re.S)
    core = m.group(1)
    import tempfile
    with tempfile.TemporaryDirectory() as tmp:
        def resolve(rel):
            return pathlib.Path(rel).expanduser() if rel.startswith("~") else brand / rel
        photos = {name: bx.photo_uri(resolve(v[0]), pathlib.Path(tmp), v[1]) if isinstance(v, tuple) else bx.photo_uri(resolve(v), pathlib.Path(tmp))
                  for name, v in bx.PHOTOS.items()}
    subs = {
        "MARKS": json.dumps({k: bx.mark_paths(a / v) for k, v in bx.MARK_FILES.items()}),
        "ICONLIB": "__ENV__.icons",
        "PHOTOS": "__ENV__.photos",
        "LOGO_POS": bx.data_uri(a / "logos/primary/wm_h_pos_clr_rgb_august2024.svg", "image/svg+xml"),
        "LOGO_REV": bx.data_uri(a / "logos/primary/wm_h_rev_wht_rgb_august2024.svg", "image/svg+xml"),
        "TAGLINE_REV": bx.data_uri(a / "logos/primary/tagline_neg_leftalign.svg", "image/svg+xml"),
    }
    for k, v in subs.items():
        core = core.replace("{{" + k + "}}", v)
    left = re.findall(r"\{\{[A-Z_0-9]+\}\}", core)
    if left:
        sys.exit(f"renderer: unresolved placeholders {sorted(set(left))}")
    renderer = (
        "/* Generated by tools/gen_renderer.py from explorations/components.src.html (renderer:start..end). Do not edit. */\n"
        "/* eslint-disable */\n"
        "export function createRenderer(__ENV__) {\n"
        "  var T = __ENV__.tokens, FR = __ENV__.frames, D = { inkRules: __ENV__.inkRules };\n"
        + core +
        "\n  return { stage: stage, node: node, attachWarnings: attachWarnings, icon: icon, ROLES: ROLES, SURF: SURF };\n}\n"
    )
    GEN.mkdir(parents=True, exist_ok=True)
    (GEN / "renderer.js").write_text(renderer)
    (GEN / "renderer.d.ts").write_text("export declare function createRenderer(env: Record<string, unknown>): any;\n")
    (GEN / "data.d.ts").write_text("export declare const DATA: { tokens: any; frames: any; inkRules: any };\n"
                                   "/** Library templates: { id, variant, name, tier, purpose, uses, legacy, budget, slots, slide }. */\n"
                                   "export declare const TEMPLATES: Array<{ id: string; variant: string; name: string; tier: string; purpose: string; slide: Record<string, unknown>; [k: string]: unknown }>;\n"
                                   "export declare const CATALOG: { families: Array<{ id: string; name: string; summary: string }>; templates: Array<Record<string, unknown>> };\n")
    (GEN / "assets.d.ts").write_text("export declare const PHOTOS: Record<string, string>;\nexport declare const ICONS: Record<string, { id: string; desc: string; svg: string }>;\n")

    comps = json.loads((ROOT / "components/v0/components.json").read_text())
    templates = bx.template_entries()
    catalog = json.loads((ROOT / "templates/catalog.json").read_text())
    data = {"tokens": json.loads((ROOT / "tokens/v0/tokens.json").read_text()), "frames": json.loads((ROOT / "frames/v0/frames.json").read_text()),
            "inkRules": comps["inkRules"]}
    (GEN / "data.js").write_text(
        "/* Generated by tools/gen_renderer.py. Do not edit. */\n"
        f"export const DATA = {json.dumps(data)};\n"
        f"export const TEMPLATES = {json.dumps([{k: v for k, v in t.items() if k != '_batchName'} for t in templates])};\n"
        f"export const CATALOG = {json.dumps(catalog)};\n")
    (GEN / "assets.js").write_text(
        "/* Generated by tools/gen_renderer.py. Do not edit. Brand photos and the icon library. */\n"
        f"export const PHOTOS = {json.dumps(photos)};\n"
        f"export const ICONS = {json.dumps(bx.icon_library(brand))};\n")

    css = "\n".join(re.findall(r"<style[^>]*>(.*?)</style>", src.split("</head>", 1)[0], flags=re.S))
    for key, rel in (("HL", "wm_highlight_1"), ("HL2", "wm_highlight_2"), ("HL3", "wm_highlight_3"), ("HL4", "wm_highlight_4")):
        css = css.replace("{{" + key + "}}", bx.data_uri(a / f"graphics/highlight/{rel}_rgb_240912.png", "image/png"))
    left = re.findall(r"\{\{[A-Z_0-9]+\}\}", css)
    if left:
        sys.exit(f"css: unresolved placeholders {sorted(set(left))}")
    scoped = scope_css(css)
    scoped += ("\n/* package additions */\n.wm-root { display: block; width: 100%; font-family: var(--sans); color: var(--grounded); }"
               "\n.wm-root .stage { outline: none; }"
               "\n.wm-root:not(.wm-warnings) .warn-box { outline: none; }\n.wm-root:not(.wm-warnings) .warn { display: none; }\n")
    (GEN / "styles.css").write_text("/* Generated by tools/gen_renderer.py: the board's stylesheet scoped under .wm-root. Do not edit. */\n"
                                    "@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap');\n"
                                    + scoped)

    samples = {}
    collect(samples)
    ifs = interfaces(samples)
    lines = ["/* Generated by tools/gen_renderer.py: props inferred from every use in the component library, frames and templates. Do not edit. */",
             "/** Every node takes slide coordinates in points (960 x 540 slide; x on 3 + 18k, y on 18k). */"]
    for name, t, fl, n in ifs:
        lines.append(f"/** Props for `{name}` (node type `{t}`; {n} uses in the library). */")
        lines.append(f"export interface {name}Props {{\n" + "\n".join(fl) + "\n}")
    lines.append("export const NODE_TYPES = " + json.dumps({NODES[t][0]: t for t in NODES}, indent=2) + " as const;")
    (GEN / "nodes.ts").write_text("\n".join(lines) + "\n")

    # per-component docs: summary, variants and rules from the library components that use this node type
    ref = (ROOT / "docs/authoring-reference.md").read_text()
    docs = PKG / "docs"
    docs.mkdir(exist_ok=True)
    by_type = {}
    for c in comps["components"]:
        for ex in c.get("examples", []):
            types = [n.get("type") for n in ex["nodes"]]
            if types:
                main = max(set(types), key=types.count)
                by_type.setdefault(main, [])
                if c not in by_type[main]:
                    by_type[main].append(c)
    for t, (name, group) in NODES.items():
        refline = next((ln.strip() for ln in ref.splitlines() if ln.strip().startswith(f"- `{t}` ")), "")
        out = [f"---\ncategory: {group}\n---\n", f"# {name}\n"]
        if t == "slide":
            out.append((PKG / "docs-src/Slide.md").read_text())
        else:
            cs = by_type.get(t, [])
            lead = cs[0]["summary"] if cs else f"The `{t}` node."
            out.append(f"{lead}\n\nUse inside `<Slide>` (positions are slide points) or `<Canvas>`. Props mirror the `{t}` node in the JSON spec pptxgengo reads, so whatever you compose maps 1:1 onto native PowerPoint.\n")
            if refline:
                out.append(f"Fields: {refline[2:]}\n")
            for c in cs:
                out.append(f"## {c['name']} (`{c['id']}`)\n\n{c['summary']}\n")
                if c.get("variants"):
                    out.append("Variants: " + "; ".join(c["variants"]) + "\n")
                for r in c.get("rules", []):
                    out.append(f"- {r}")
                if c.get("sizing"):
                    out.append(f"\nSizing: {c['sizing']}\n")
        (docs / f"{name}.md").write_text("\n".join(out) + "\n")
    for extra in ("Canvas", "Template", "Icon"):
        (docs / f"{extra}.md").write_text((PKG / f"docs-src/{extra}.md").read_text())
    print(f"generated renderer ({len(core) // 1024} KB), {len(ifs)} components, {len(templates)} templates")


if __name__ == "__main__":
    main()
