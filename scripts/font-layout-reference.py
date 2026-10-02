#!/usr/bin/env python3
"""Build a reproducible three-family typography corpus and compare PDF references.

PowerPoint reference PDFs are exported from the exact probe PPTX using PowerPoint.
PDF selection bounds are a separate measurement contract from TextRange bounds.
No layout calibration is applied by this script.
"""
import argparse
import base64
import collections
import copy
import hashlib
import json
import math
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]
FAMILIES = ("IBM Plex Sans", "IBM Plex Mono", "Arial")
SIZES = (10, 12, 14, 16, 18, 24, 32, 40)
STYLES = (("regular", False, False), ("bold", True, False),
          ("italic", False, True), ("bold_italic", True, True))
COLOR = "#070154"
WRAP = ("Clear typography helps readers follow the argument. Each font needs its "
        "own measurements at the chosen size and box width.")
BOUNDARY = "A clear typography plan"


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path, value):
    with path.open("x") as f:
        json.dump(value, f, indent=2, ensure_ascii=False, allow_nan=False)
        f.write("\n")


def load(path):
    return json.loads(path.read_text())


def run(binary, *args):
    subprocess.run([str(binary), *map(str, args)], check=True, cwd=ROOT)


def run_spec(text, family, size, bold=False, italic=False, underline=False):
    return dict(id="run", text=text, font_face=family, font_size_pt=size,
                bold=bold, italic=italic, underline=underline, foreground=COLOR)


def case(cid, family, category, size=16, width=360, style="regular", text=WRAP,
         paragraphs=None, align="left", inset_x=0, inset_y=0, **extra):
    _, bold, italic = next(s for s in STYLES if s[0] == style)
    canvas = dict(id="sample", kind="text", bounds=dict(x=40, y=70, width=width + 2*inset_x,
                 height=400), align=align, valign="top", inset_x=inset_x, inset_y=inset_y)
    if paragraphs is not None or italic:
        canvas["paragraphs"] = paragraphs or [dict(id="paragraph", align=align,
            runs=[run_spec(text, family, size, bold, italic)])]
    else:
        canvas.update(text=text, font_face=family, font_size_pt=size, bold=bold, foreground=COLOR)
    return dict(id=cid, family=family, category=category, size_pt=size,
                width_pt=width, style=style, canvas=canvas, **extra)


def family_key(family):
    return {"IBM Plex Sans": "plex-sans", "IBM Plex Mono": "plex-mono", "Arial": "arial"}[family]


def corpus():
    out = []
    for fi, family in enumerate(FAMILIES):
        key = family_key(family)
        for size in SIZES:
            for style, _, _ in STYLES:
                out.append(case(f"{key}-size-{size}-{style}", family, "size_style", size, 800,
                                style, "Agility AV office 0123"))
        for size in (12.5, 16.5, 27.25):
            for style in ("regular", "bold"):
                out.append(case(f"{key}-fraction-{size}-{style}", family, "fractional_size",
                                size, 800, style, "Fractional font sizes stay explicit."))
        for size in (12, 16, 24):
            for width in (180, 360):
                for style in ("regular", "bold"):
                    out.append(case(f"{key}-wrap-{size}-{width}-{style}", family, "wrapping",
                                    size, width, style))
        for multiple in (1, 1.15, 1.5):
            paragraphs = [dict(id="first", align="left", space_before_pt=6, space_after_pt=8,
                line_spacing_multiple=multiple, runs=[run_spec("Paragraph spacing remains explicit.", family, 16)]),
                dict(id="second", align="left", space_before_pt=4, space_after_pt=6,
                line_spacing_multiple=multiple, runs=[run_spec("A second paragraph wraps in this narrow text box.", family, 16)])]
            out.append(case(f"{key}-spacing-{multiple}", family, "paragraph_spacing", width=360,
                            paragraphs=paragraphs, line_spacing_multiple=multiple))
        for n, text in enumerate(("A clear plan.\n\nA blank line separates the paragraphs.\nA final line.",
                                  "A clear plan.\nA trailing hard break follows.\n")):
            out.append(case(f"{key}-hard-break-{n}", family, "hard_break", width=420, text=text))
        for align in ("left", "center", "right"):
            out.append(case(f"{key}-alignment-{align}", family, "alignment", width=240,
                text="Aligned text wraps into several lines within this fixed width.", align=align))
        for bi, character in enumerate(("•", "–")):
            for width in (180, 360):
                paragraphs = [dict(id="bullet", align="left", space_after_pt=6,
                    bullet=dict(character=character, margin_left_pt=24, hanging_pt=16),
                    runs=[run_spec("A wrapped bullet keeps its continuation aligned with the text after the indent.", family, 16)])]
                out.append(case(f"{key}-bullet-{bi}-{width}", family, "bullet", width=width,
                                paragraphs=paragraphs, bullet=character))
        paragraphs = [dict(id="styles", align="left", runs=[
            dict(run_spec("Regular, ", family, 16), id="regular"),
            dict(run_spec("larger bold, ", family, 20, bold=True), id="bold"),
            dict(run_spec("smaller italic, ", family, 14, italic=True), id="italic"),
            dict(run_spec("bold italic and underlined text.", family, 16, bold=True, italic=True, underline=True), id="bold-italic")])]
        out.append(case(f"{key}-mixed-styles", family, "mixed_styles", width=420, paragraphs=paragraphs))
        paragraphs = [dict(id="families", align="left", runs=[
            dict(run_spec("Sans and mono: ", family, 16), id="primary"),
            dict(run_spec("a measured label ", FAMILIES[(fi+1)%3], 18, bold=True), id="second"),
            dict(run_spec("beside another family.", FAMILIES[(fi+2)%3], 16, italic=True), id="third")])]
        out.append(case(f"{key}-mixed-families", family, "mixed_families", width=420, paragraphs=paragraphs))
        out.append(case(f"{key}-insets", family, "insets", width=220, inset_x=8, inset_y=4,
                        text="Insets reduce the width available for the words inside this text box."))
        out.append(case(f"{key}-long-word", family, "long_word", width=120,
                        text="MeasuredTypographyWithoutSpaces_0123456789"))
        out.append(case(f"{key}-punctuation", family, "punctuation", width=420,
                        text="“Flexible layouts” – €123.45 / résumés / field_names."))
    return out


def spec_for(cases):
    return dict(schema="pptxgengo.compose-spec.v1", slides=[dict(id=c["id"], width_pt=960,
        height_pt=540, title_font_face="Arial", title_font_size_pt=24, pods=[],
        canvas=[c["canvas"]], notes=f"Reference case {c['id']}. {c['category']}. Usable width {c['width_pt']}pt.")
        for c in cases])


def case_from_request(request_id):
    # Composer IDs encode [slide ID, object kind, canvas ID].
    value = request_id.removeprefix("measure:")
    return json.loads(base64.urlsafe_b64decode(value + "="*((-len(value))%4)))[0]


def indexed(report):
    return {case_from_request(k): dict(measurement=report["measurements"]["by_request_id"][k],
             detail=v, request_id=k) for k, v in report["requests"].items()}


def review_spec(cases, go, width, native=None):
    by_id = {c["id"]: c for c in cases}
    measures = indexed(go)
    slides = []
    for family in FAMILIES:
        key = family_key(family)
        sections = [
            ("Point sizes and styles", [f"{key}-size-{size}-{STYLES[i%4][0]}" for i,size in enumerate(SIZES)]),
            ("Wrapping boundaries", [f"{key}-wrap-16-180-regular", f"{key}-wrap-16-360-regular",
                                     f"{key}-boundary-16-regular-minus-0.15", f"{key}-boundary-16-regular-plus-0.15"]),
            ("Paragraph controls", [f"{key}-bullet-0-360", f"{key}-spacing-1.5",
                                    f"{key}-mixed-families", f"{key}-hard-break-0"])]
        for si, (section, ids) in enumerate(sections):
            def new_page():
                slide = dict(id=f"{key}-{si}-{len(slides)+1}", width_pt=width, height_pt=540,
                    title=f"{family}: {section}", title_bounds=dict(x=40,y=20,width=width-80,height=32),
                    title_font_face="Arial", title_font_size_pt=20, title_bold=True,
                    title_foreground=COLOR, pods=[], canvas=[], notes="Typography reference. Point units. Labels identify corpus cases. Frames include review clearance. PDF reference geometry is distinct from native TextRange verification.")
                slides.append(slide)
                return slide
            slide = new_page(); y=70
            for cid in ids:
                c = by_id[cid]; m = measures[cid]["measurement"]
                body = copy.deepcopy(c["canvas"])
                body_w = min(body["bounds"]["width"], width-80)
                # Size/style controls use a wide single-line frame in both formats.
                body_h = math.ceil(m["rendered_height_pt"] + max(3, c["size_pt"]*.15) + 2*body.get("inset_y",0))
                if native:
                    body_h = max(body_h, math.ceil(native[cid]["bounds"]["height"] + 8 + 2*body.get("inset_y",0)))
                if y+15+body_h > 526:
                    slide=new_page();y=70
                label = f"{c['category']}   {c['size_pt']} pt   {c['style']}   {body_w:g} pt frame"
                slide["canvas"].append(dict(id=cid+"-label",kind="text",bounds=dict(x=40,y=y,width=width-80,height=12),
                    text=label,font_face="Arial",font_size_pt=9,foreground=COLOR,align="left",valign="top"))
                body["id"]=cid;body["bounds"]=dict(x=40,y=y+15,width=body_w,height=body_h)
                slide["canvas"].append(body);slide["notes"] += "\nCorpus: "+cid
                y += 15+body_h+8
    return dict(schema="pptxgengo.compose-spec.v1",slides=slides)


def prepare(binary, out):
    out.mkdir(parents=True, exist_ok=False)
    seeds=[]
    for family in FAMILIES:
        for size in (16,16.5,24):
            for style in (("regular","bold") if size!=16.5 else ("regular",)):
                seeds.append(case(f"{family_key(family)}-seed-{size}-{style}",family,"boundary_seed",size,800,style,BOUNDARY))
    write_json(out/"boundary-seeds.json",spec_for(seeds))
    run(binary,"measure","--engine","go","--spec",out/"boundary-seeds.json","--out",out/"boundary-seed-measurements.json")
    seed_go=indexed(load(out/"boundary-seed-measurements.json"))
    cases=corpus()
    for seed in seeds:
        advance=seed_go[seed["id"]]["detail"]["lines"][0]["advance_pt"]
        for delta in (-1,-.15,0,.15,1):
            tag=("minus-"+str(-delta)) if delta<0 else ("plus-"+str(delta))
            cid=f"{family_key(seed['family'])}-boundary-{seed['size_pt']}-{seed['style']}-{tag}"
            cases.append(case(cid,seed["family"],"boundary",seed["size_pt"],advance+delta,
                              seed["style"],BOUNDARY, threshold_advance_pt=advance, threshold_delta_pt=delta))
    write_json(out/"corpus.json",spec_for(cases))
    write_json(out/"cases.json",dict(schema="pptxgengo.font-reference-cases.v1",families=FAMILIES,
        sizes_pt=SIZES,styles=[s[0] for s in STYLES],cases=cases))
    run(binary,"measure","--engine","go","--spec",out/"corpus.json","--out",out/"go-measurements.json")
    run(binary,"probe","--spec",out/"corpus.json","--out",out/"font-reference-probes")
    go=load(out/"go-measurements.json")
    decks=[]
    for name,width in (("widescreen",960),("standard",720)):
        review=review_spec(cases,go,width)
        write_json(out/(name+".json"),review)
        run(binary,"build","--engine","go","--spec",out/(name+".json"),"--out",out/name)
        run(binary,"fit-report","--engine","go","--spec",out/(name+".json"),"--out",out/(name+"-fit.json"))
        decks.append(dict(path=f"{name}/{name}.pptx",sha256=digest(out/name/(name+".pptx")),
                          slide_count=len(review["slides"]),width_pt=width,height_pt=540))
    write_json(out/"reference-set.json",dict(schema="pptxgengo.font-reference-set.v1",status="go_measured_native_pending",
        case_count=len(cases),family_counts=dict(collections.Counter(c["family"] for c in cases)),
        categories=dict(collections.Counter(c["category"] for c in cases)),
        corpus_sha256=digest(out/"corpus.json"),go_measurements_sha256=digest(out/"go-measurements.json"),
        binary_sha256=digest(binary),generator_sha256=digest(Path(__file__)),
        probe_sha256=digest(out/"font-reference-probes/font-reference-probes.pptx"),decks=decks,
        native_reference_contract="PowerPoint-exported PDF, extracted line selections and font attributes; distinct from TextRange bounds"))
    print(f"Prepared {len(cases)} cases and {sum(d['slide_count'] for d in decks)} review slides",flush=True)


def canonical(text):
    import unicodedata
    return unicodedata.normalize("NFKC",text).strip()


def compare(out, native_path, name, go_path=None, report_dir=None):
    record=load(out/"reference-set.json");native=load(native_path)
    cases=load(out/"cases.json")["cases"];go=load(go_path or out/"go-measurements.json");by_id=indexed(go)
    manifest_path=out/"font-reference-probes/manifest.json"
    if not manifest_path.exists():manifest_path=out/"probe-manifest.json"
    if not manifest_path.exists():manifest_path=out/"manifest.json"
    manifest=load(manifest_path)
    baseline_go=load(out/"go-measurements.json")
    report_dir=report_dir or out
    if record["corpus_sha256"]!=digest(out/"corpus.json") or go["spec_sha256"]!=record["corpus_sha256"]:
        raise ValueError("Spec/Go measurement hashes disagree")
    if native["page_count"]!=len(cases) or len(native["pages"])!=len(cases):
        raise ValueError("Native PDF page count does not match the corpus")
    if len(manifest["requests"])!=len(cases):raise ValueError("Probe request count mismatch")
    probe_path=out/"font-reference-probes/font-reference-probes.pptx"
    if not probe_path.exists():probe_path=out/"font-reference-probes.pptx"
    if digest(probe_path)!=record["probe_sha256"]:raise ValueError("Probe source hash mismatch")
    pdf_path=Path(native["pdf_file"])
    if not pdf_path.is_absolute():pdf_path=native_path.parent/pdf_path
    if native["pdf_sha256"]!=digest(pdf_path):
        raise ValueError("Native PDF changed after extraction")
    rows=[]
    for c,p,q in zip(cases,native["pages"],manifest["requests"]):
        if case_from_request(q["id"])!=c["id"]:raise ValueError("Probe request order mismatch")
        d=by_id[c["id"]];m=d["measurement"]
        glines=[canonical(l["text"]) for l in d["detail"]["lines"] if canonical(l["text"])]
        bullet_chars={para["bullet"]["character"] for para in q.get("paragraphs",[]) if para.get("bullet")}
        nlines=[]
        for line in p["lines"]:
            text=canonical(line["text"])
            if text and text[0] in bullet_chars: text=text[1:].strip()
            if text:nlines.append(text)
        expected_fonts=sorted({r["font_face"] for para in q.get("paragraphs",[]) for r in para["runs"]} or {q["font_face"]})
        # PDF fonts can have subset prefixes and PostScript style suffixes.
        fonts=p["font_names"]
        unexpected=[f for f in fonts if not any(family.replace(" ","").lower() in f.replace(" ","").lower() for family in expected_fonts)]
        missing=[family for family in expected_fonts if not any(family.replace(" ","").lower() in f.replace(" ","").lower() for f in fonts)]
        if missing:raise ValueError(f"Missing PDF font resources for {c['id']}: {missing}")
        bounds=p["bounds"]
        rows.append(dict(case_id=c["id"],family=c["family"],category=c["category"],size_pt=c["size_pt"],
            style=c["style"],width_pt=c["width_pt"],threshold_delta_pt=c.get("threshold_delta_pt"),
            line_breaks_match=glines==nlines,go_lines=glines,native_lines=nlines,
            source_text_preserved="".join("".join(nlines).split())=="".join((q["text"] or "\n".join("".join(run["text"] for run in para["runs"]) for para in q["paragraphs"])).split()),
            go_visible_line_count=len(glines),native_visible_line_count=len(nlines),
            go_width_pt=m["rendered_width_pt"],native_pdf_width_pt=bounds["width"],
            width_delta_pt=m["rendered_width_pt"]-bounds["width"],
            go_height_pt=m["rendered_height_pt"],native_pdf_height_pt=bounds["height"],
            height_delta_pt=m["rendered_height_pt"]-bounds["height"],
            native_fonts=fonts,unexpected_native_fonts=unexpected,extractor_selection_fonts=p.get("selection_font_names",[]),
            wrap_boundary_warnings=d["detail"].get("boundary_warnings",[]),
            go_predicted_height_contains_pdf=bounds["height"]<=m["rendered_height_pt"]+.15))
    groups=[]
    for family in FAMILIES:
        subset=[r for r in rows if r["family"]==family]
        groups.append(dict(family=family,case_count=len(subset),line_break_match_count=sum(r["line_breaks_match"] for r in subset),
            unexpected_font_cases=sum(bool(r["unexpected_native_fonts"]) for r in subset),
            max_absolute_width_delta_pt=max(abs(r["width_delta_pt"]) for r in subset),
            max_absolute_pdf_height_delta_pt=max(abs(r["height_delta_pt"]) for r in subset),
            go_height_underprediction_count=sum(not r["go_predicted_height_contains_pdf"] for r in subset)))
    result=dict(schema="pptxgengo.font-reference-comparison.v1",contract=native["contract"],
        scope="Actual PowerPoint PDF layout versus Go predictions. PDF glyph selections do not expose TextRange height or baseline. This is a calibration reference, not a native TextRange verification certificate.",
        go_engine=go["engine"],go_advance_policy=go.get("advance_policy"),
        go_measurements_sha256=digest(go_path or out/"go-measurements.json"),
        font_provenance_matches_baseline=go["fonts"]==baseline_go["fonts"],
        corpus_sha256=record["corpus_sha256"],native_pdf_sha256=native["pdf_sha256"],case_count=len(rows),groups=groups,rows=rows)
    write_json(report_dir/(name+".json"),result)
    report=["# Font layout reference set", "",f"{len(rows)} cases. PowerPoint PDF references captured for all three families.","",
        f"Go engine: `{go['engine']}`. Font provenance matches the dataset's baseline: {result['font_provenance_matches_baseline']}.","",
        "The PDF selections measure exported glyph geometry. Their height differs in meaning from PowerPoint TextRange bounds and Go line boxes.","",
        "| Family | Cases | Matching line breaks | Font warnings | Largest width delta | Largest PDF height delta |",
        "| --- | ---: | ---: | ---: | ---: | ---: |"]
    report.extend(f"| {g['family']} | {g['case_count']} | {g['line_break_match_count']} | {g['unexpected_font_cases']} | {g['max_absolute_width_delta_pt']:.3f} pt | {g['max_absolute_pdf_height_delta_pt']:.3f} pt |" for g in groups)
    report += ["","## Cases needing calibration",""]
    for r in rows:
        if not r["line_breaks_match"] or r["unexpected_native_fonts"]:
            report.append(f"- `{r['case_id']}`: Go {r['go_visible_line_count']} visible lines, PowerPoint {r['native_visible_line_count']}. Font warnings: {', '.join(r['unexpected_native_fonts']) or 'none'}.")
    report += ["","Width differences include cases with different wrapping; they are not single-line advance errors. Blank lines are excluded from the visible-line comparison, but remain in the corpus and geometry. Bullet markers are excluded from line text comparisons. Font names come from PDF BaseFont resources; PDFKit selection font attributes can report fallback names in this process.","",f"Full measurements and line text are in `{name}.json`. Exact source contracts are in the dataset's `corpus.json` and `cases.json`. The comparator does not modify measurements or calibrate the engine.",""]
    with (report_dir/(name+".md")).open("x") as f:f.write("\n".join(report))
    print(json.dumps(groups,indent=2))


def reviewed(binary, out, native_path):
    cases=load(out/"cases.json")["cases"]
    native=load(native_path)
    if len(cases)!=len(native["pages"]):raise ValueError("Native page count mismatch")
    pages={c["id"]:p for c,p in zip(cases,native["pages"])}
    go=load(out/"go-measurements.json")
    for name,width in (("widescreen-reviewed",960),("standard-reviewed",720)):
        spec=review_spec(cases,go,width,pages)
        write_json(out/(name+".json"),spec)
        run(binary,"build","--engine","go","--spec",out/(name+".json"),"--out",out/name)
        run(binary,"fit-report","--engine","go","--spec",out/(name+".json"),"--out",out/(name+"-fit.json"))
        print(f"{name}: {len(spec['slides'])} slides")


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    sub=parser.add_subparsers(dest="command",required=True)
    p=sub.add_parser("prepare");p.add_argument("--binary",type=Path,required=True);p.add_argument("--out",type=Path,required=True)
    p=sub.add_parser("compare");p.add_argument("--out",type=Path,required=True);p.add_argument("--native",type=Path,required=True);p.add_argument("--name",default="comparison");p.add_argument("--go",type=Path);p.add_argument("--reports",type=Path)
    p=sub.add_parser("review");p.add_argument("--out",type=Path,required=True);p.add_argument("--native",type=Path,required=True);p.add_argument("--binary",type=Path,required=True)
    args=parser.parse_args()
    if args.command=="prepare":prepare(args.binary.resolve(),args.out.resolve())
    elif args.command=="compare":compare(args.out.resolve(),args.native.resolve(),args.name,args.go.resolve() if args.go else None,args.reports.resolve() if args.reports else None)
    else:reviewed(args.binary.resolve(),args.out.resolve(),args.native.resolve())


if __name__=="__main__":main()
