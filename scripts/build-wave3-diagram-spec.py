#!/usr/bin/env python3
"""Generate measured Wave 3 diagram review and negative qualification specs."""
from __future__ import annotations

import copy
import hashlib
import json
import subprocess
from pathlib import Path
from zipfile import ZipFile

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "library/diagram-components"
ASSETS = ROOT / "samples/visual-wave3/source-assets"
OUT.mkdir(parents=True, exist_ok=True)
ASSETS.mkdir(parents=True, exist_ok=True)
SOURCE = ROOT / "samples/UHG Fabric Platforming RFP Response - July 2026.pptx"
SOURCE_SHA = "b0f254ed7739768d0f345257689264393049d06cdd3849a177fc764f3d348d99"
assert hashlib.sha256(SOURCE.read_bytes()).hexdigest() == SOURCE_SHA
SCHEMA = "pptxgengo.compose-spec.v1"
NAVY, BLUE, MAGENTA, PALE, MID, WHITE = "#070154", "#0047FF", "#F900D3", "#CDD6E5", "#536A92", "#FFFFFF"

def rect(x,y,w,h): return dict(x=x,y=y,width=w,height=h)
def pad(t=0,r=0,b=0,l=0): return dict(top=t,right=r,bottom=b,left=l)
def txt(id,text,x,y,w,h,size=12,color=NAVY,bold=False,align="left",valign="top"):
    return dict(id=id,kind="text",bounds=rect(x,y,w,h),text=text,font_face="Arial",font_size_pt=size,bold=bold,foreground=color,align=align,valign=valign,inset_x=0,inset_y=0)
def surface(id,x,y,w,h,color,children=()):
    d=dict(id=id,kind="surface",bounds=rect(x,y,w,h),background=color)
    if children:d["allow_overlap"]=list(children)
    return d
def path(id,x,y,w,h,labels,node_bg,node_fg,gap=19.014,parent="",size=12):
    d=dict(id=id,bounds=rect(x,y,w,h),labels=labels,gap_pt=gap,node_padding=pad(3,6,3,6),node_background=node_bg,font_face="Arial",font_size_pt=size,foreground=node_fg,arrow_color=NAVY,arrow_height_pt=13.724,arrow_margin_pt=3.3)
    if parent:d["parent_id"]=parent
    return d
def slide(id,notes,page):
    return dict(id=id,title="",width_pt=960,height_pt=540,title_bounds=rect(0,0,0,0),title_font_face="Arial",title_font_size_pt=23,title_bold=False,title_foreground=NAVY,pods=[],canvas=footer(page),role="diagram",notes=notes)

def footer(page):
    base=json.loads((ROOT/"library/showcase/dense-deck.json").read_text())["slides"][0]["canvas"]
    items=[copy.deepcopy(c) for c in base if c["id"] in {"wm-logo","footer-copy","page"}]
    for c in items:
        if c["id"]=="page":c["text"]=str(page)
    return items

def save(name,slides):
    (OUT/name).write_text(json.dumps(dict(schema=SCHEMA,slides=slides),indent=2,ensure_ascii=False)+"\n")

source=slide("uhg36-source-control", "UHG36 source-control text and resolved node frames. Source deck SHA-256 "+SOURCE_SHA+"; slide XML ppt/slides/slide36.xml. Original grouped shapes were anisotropically transformed; this recreation uses resolved slide-space frames. Requires native visual comparison, not component approval.",36)
source["canvas"] += [
    txt("source-kicker","OUR APPROACH",36,27,170,16,9,NAVY,True),
    txt("source-phase","PHASE 03",36,53,215,21,16,MAGENTA,True),
    txt("source-title","General Availability\n& Optimization",36,75,264,50,24,NAVY,True),
    txt("source-weeks","Weeks 35 - 42",36,139,220,22,14,NAVY),
    txt("source-headline","We will harden both the platform and operating model by embedding\ngovernance throughout the rollout & optimization processes",324,50,605,38,16,NAVY,True),
    txt("source-pathway-heading","TWO CRITICAL PATHWAYS TO OPERATIONALIZE UHG’s PLATFORM GOVERNANCE",324,115,603,20,12,BLUE,True),
    txt("source-number-one","01",324,140,25,20,16,MAGENTA,True),
    txt("source-path-one-title","Request-to-provisioning intake: Team identified to approved pattern access",349,140,585,20,14,NAVY,True),
    txt("source-number-two","02",324,242,25,20,16,MAGENTA,True),
    txt("source-path-two-title","Use-case/Optimization intake: New idea to governed build, deployment, and monitoring",349,242,585,20,14,NAVY,True),
    surface("source-result-band",324.210,341,601.981,59.5,BLUE,["source-result-text"]),
    txt("source-result-text","Result: Controlled Fabric scale, approval-only tool access, governed\ndata patterns, and a repeatable model for future use cases.",350,356,551,35,14,WHITE,True,"center","middle"),
    txt("source-alignment","Designed to align with UHG’s existing governance structure, Data Enablement team,\nEnterprise L&D, security, and operational support functions.",324,416,600,38,15,NAVY),
    surface("source-takeaway",0,258.5,291.5,188.5,NAVY,["takeaway-head","takeaway-body","takeaway-list"]),
    txt("takeaway-head","Make governance executable",43,276,236,22,16,WHITE,True),
    txt("takeaway-body","Scale depends on clear decision rights,\nauditable controls, and operational handoffs -\nnot one-time enablement.",43,298,238,42,12,WHITE),
    txt("takeaway-list","•   Named approvers\n•   SLA tracking\n•   Access controls\n•   Measured adoption",43,350,230,78,13,WHITE,True),
]
source["paths"]=[
    path("source-request",324.210,169.039,601.981,60.996,["Team Identified","Use-case / Pattern\nmapping","Management +\nsecurity approval","Provisioning","SLA + usage\ntracking"],PALE,NAVY),
    path("source-optimization",324.210,272.268,601.981,60.996,["Team use-case\nproposal","Governance\nreview","Risk + control\nassessment","Approved build\n/ deploy","Monitoring +\nlifecycle"],MID,WHITE),
]
for c in source["canvas"]:
    if c["id"]=="source-result-text":c["contrast_background"]=BLUE
    if c["id"].startswith("takeaway-"):c["contrast_background"]=NAVY

synthetic=slide("process-changed-content", "Synthetic illustration using the UHG36 two-path geometry. Labels and outcome are illustrative; no UHG business result is asserted.",1)
synthetic["takeaway"]="Explicit intake and change-review steps make decisions visible and handoffs repeatable."
synthetic["canvas"] += [txt("synthetic-kicker","ILLUSTRATIVE PROCESS PATHS",36,42,700,22,15,BLUE,True),txt("synthetic-title","Two governed paths from request to review",36,70,870,38,23,NAVY,True),txt("synthetic-one","01  Access request path",324,140,600,22,14,NAVY,True),txt("synthetic-two","02  Change review path",324,242,600,22,14,NAVY,True),surface("illustrative-result",324,341,602,60,BLUE,["illustrative-result-text"]),txt("illustrative-result-text","Illustrative result: visible decisions and repeatable handoffs.",344,357,562,28,14,WHITE,True,"center","middle"),txt("synthetic-note","Synthetic content for component qualification; not an operating claim.",324,419,602,25,11,NAVY)]
synthetic["paths"]=[path("request",324.210,169.039,601.981,60.996,["Request logged","Owner assigned","Controls checked","Access granted","Usage reviewed"],PALE,NAVY),path("change",324.210,272.268,601.981,60.996,["Idea submitted","Scope reviewed","Risk assessed","Release approved","Outcome checked"],MID,WHITE)]
for c in synthetic["canvas"]:
    if c["id"]=="illustrative-result-text":c["contrast_background"]=BLUE

def stress(id,labels,x=175,y=210,w=610,parent_move=False):
    s=slide(id,"Synthetic cardinality and measured-fit stress. No source business claims.",2)
    s["canvas"] += [txt("stress-title",id.replace("-"," ").title(),50,5 if parent_move else 60,850,30 if parent_move else 42,23,NAVY,True)]
    if parent_move:
        s["layouts"]=[dict(id="translated-parent",bounds=rect(45,40,800,300),padding=pad(10,10,10,10),columns=[dict(fixed_pt=780)],rows=[dict(fixed_pt=280)],cells=[],surface="#F2F5FA")]
        s["paths"]=[path("stress",x,y,w,62,labels,PALE,NAVY,gap=19,parent="translated-parent")]
    else:s["paths"]=[path("stress",x,y,w,62,labels,PALE,NAVY,gap=19)]
    return s
four=stress("process-four-node",["Receive","Check","Approve","Record"])
six=stress("process-six-node",["Receive","Classify","Assign","Review","Approve","Record"])
moved=stress("process-parent-translation",["Receive","Check","Approve","Record"],x=50,y=140,w=610,parent_move=True)
negative=stress("process-long-label-overflow",["Receive","This intentionally long label exceeds its measured node width without shrinking the declared type","Approve","Record"])

def box(id,parent,x,y,w,h,label,bg=PALE):
    fg=WHITE if bg==MID else NAVY
    return dict(id=id,parent_id=parent,bounds=rect(x,y,w,h),padding=pad(0,0,0,0),columns=[dict(fixed_pt=w)],rows=[dict(fixed_pt=h)],surface=bg,cells=[dict(id="node",row=0,column=0,padding=pad(5,6,5,6),blocks=[dict(id="label",text=label,font_face="Arial",font_size_pt=11,foreground=fg,align="center",valign="middle")])])
arch=slide("editable-architecture", "UNQUALIFIED relation/layer smoke. Synthetic editable architecture; UHG14 architecture board is a raster source and is not used as semantic extraction. A denser architecture deliverable is still required.",3)
arch["canvas"] += [txt("arch-kicker","EDITABLE ARCHITECTURE STUDY",45,39,870,22,13,BLUE,True),txt("arch-title","Layered services with visible control paths",45,68,870,37,23,NAVY,True)]
arch["layouts"]=[dict(id="arch",bounds=rect(75,125,810,335),padding=pad(),columns=[dict(fixed_pt=810)],rows=[dict(fixed_pt=335)],surface="#F2F5FA",cells=[])]
rows=[("intake","Request intake","delivery","Delivery services",205,"reporting"),("data","Data products","experience","Consumer experience",255,"dependency"),("governance","Governance review","platform","Platform operations",305,"advisory")]
for left,ll,right,rl,y,rel in rows:
    arch["layouts"] += [box(left,"arch",55,y-125,260,43,ll),box(right,"arch",495,y-125,260,43,rl)]
    arch.setdefault("connections",[]).append(dict(id=rel+"-link",**{"from":left+"/node","to":right+"/node"},relationship=rel,from_anchor="right",to_anchor="left",color=NAVY,width_pt=1.3,clearance_pt=3,preferred_direction="horizontal"))
arch["layouts"] += [box("security","arch",55,265,700,42,"Cross-cutting identity, access and security controls",MID)]
arch["connections"].append(dict(id="security-annotation",**{"from":"security/node","to":"platform/node"},relationship="annotation",from_anchor="top",to_anchor="bottom",color=BLUE,width_pt=1.1,clearance_pt=3,preferred_direction="vertical"))
arch["paths"]=[path("service-flow",55,25,700,42,["Capture","Evaluate","Provision","Observe"],PALE,NAVY,gap=28,parent="arch",size=10)]

team=slide("delivery-organization", "UNQUALIFIED six-role relation smoke; a dense UHG43-based delivery organization is still required. Staffing labels are illustrative, not source staffing commitments.",4)
team["canvas"] += [txt("team-kicker","DELIVERY ORGANIZATION STUDY",45,39,870,21,13,BLUE,True),txt("team-title","Decision and advisory lines across delivery pods",45,68,870,34,23,NAVY,True)]
def role(id,label,x,y,w=160,bg="staffing.wm_full_time"):
    return dict(id=id,label=label,background=bg,foreground="auto",bounds=rect(x,y,w,42),font_face="Arial",font_size_pt=11,bold=True,horizontal_inset_pt=6,vertical_inset_pt=5)
team["roles"]=[role("program-lead","Program Lead",400,145),role("architect","Fabric Architect",85,145,160,"staffing.wm_part_time"),role("owner","Platform Owner",715,145,160,"staffing.client_part_time"),role("pod-one","Delivery Pod 1",130,310),role("pod-two","Delivery Pod 2",400,310),role("pod-three","Delivery Pod 3",670,310)]
team["legend"]=dict(bounds=rect(180,435,600,35),font_face="Arial",font_size_pt=10,bold=False,foreground=NAVY,swatch_size_pt=11,gap_pt=5,item_gap_pt=14)
team["connections"]=[dict(id="architect-advice",**{"from":"architect","to":"program-lead"},relationship="advisory",from_anchor="right",to_anchor="left",color=NAVY,width_pt=1.1,clearance_pt=3),dict(id="owner-dependency",**{"from":"program-lead","to":"owner"},relationship="dependency",from_anchor="right",to_anchor="left",color=NAVY,width_pt=1.1,clearance_pt=3)]
for id in ("pod-one","pod-two","pod-three"):
    team["connections"].append(dict(id="lead-"+id,**{"from":"program-lead","to":id},relationship="reporting",from_anchor="bottom",to_anchor="top",color=NAVY,width_pt=1.1,clearance_pt=3))

with ZipFile(SOURCE) as z:
    art={}
    for name in ("image123.png","image124.png","image86.svg"):
        data=z.read("ppt/media/"+name)
        target=ASSETS/name
        target.write_bytes(data)
        art[name]=dict(path=str(target.relative_to(ROOT)),sha256=hashlib.sha256(data).hexdigest())
fallback=ASSETS/"image86-fallback.png"
subprocess.run(["sips","-s","format","png",str(ASSETS/"image86.svg"),"--out",str(fallback)],check=True,capture_output=True)
art["image86-fallback.png"]=dict(path=str(fallback.relative_to(ROOT)),sha256=hashlib.sha256(fallback.read_bytes()).hexdigest())
artslide=slide("uhg14-art-placement", "UHG14 positioning control: architecture board image124.png and preview image123.png are raster; image86.svg is pinned vector artwork. Their internal labels/arrows are not editable or semantically anchored. Whole-image placement only.",14)
artslide["canvas"] += [txt("art-kicker","UHG14 PINNED ART POSITIONING CONTROL",35,28,420,25,13,BLUE,True)]
artslide["canvas"] += [dict(id="uhg14-board",kind="image",bounds=rect(471.29,58.34,379.34,404.35),asset_path=art["image124.png"]["path"],asset_sha256=art["image124.png"]["sha256"],alt_text="UHG14 source architecture board raster; internals are not editable",image_fit="stretch",allow_overlap=["uhg14-preview","uhg14-arrow-art"]),dict(id="uhg14-preview",kind="image",bounds=rect(819.07,22.78,113.06,120.04),asset_path=art["image123.png"]["path"],asset_sha256=art["image123.png"]["sha256"],alt_text="UHG14 source preview raster",image_fit="stretch",allow_overlap=["uhg14-board","uhg14-arrow-art"]),dict(id="uhg14-arrow-art",kind="image",bounds=rect(852.18,132.80,70.29,70.29),asset_path=art["image86.svg"]["path"],asset_sha256=art["image86.svg"]["sha256"],fallback_asset_path=art["image86-fallback.png"]["path"],fallback_asset_sha256=art["image86-fallback.png"]["sha256"],alt_text="UHG14 pinned SVG artwork; visible endpoints are not semantic links",image_fit="stretch",allow_overlap=["uhg14-board","uhg14-preview"])]

review=[source,synthetic,four,six,moved,arch,team,artslide]
save("review.json",review)
save("overflow.json",[negative])
arttranslated=copy.deepcopy(artslide)
arttranslated["id"]="uhg14-art-translation"
arttranslated["notes"]="Whole pinned-art assembly translated +8 pt x and +6 pt y; each image hash and relative offset remain unchanged. No internal node or semantic-arrow editability is implied."
for item in arttranslated["canvas"]:
    if item["id"] in {"uhg14-board","uhg14-preview","uhg14-arrow-art"}:
        item["bounds"]["x"]+=8
        item["bounds"]["y"]+=6
save("art-translation.json",[artslide,arttranslated])
(OUT/"source-evidence.json").write_text(json.dumps(dict(source_file=str(SOURCE.relative_to(ROOT)),source_sha256=SOURCE_SHA,uhg36=dict(slide_xml="ppt/slides/slide36.xml",reference_png="samples/reconstruction/reference/png/slide-036.png",node_frame_pt=rect(324.210,169.039,105.185,60.996),node_left_edges_pt=[324.210,448.409,572.607,696.806,821.004],row_tops_pt=[169.039,272.268],decorative_number_color="#F900D3",decorative_number_font_pt=16,decorative_number_bold=True,source_object_provenance="two grouped five-node pathways with four native rightArrow shapes per row; group anisotropic scaling resolved to slide coordinates"),uhg14=dict(slide_xml="ppt/slides/slide14.xml",reference_png="samples/reconstruction/reference/png/slide-014.png",art=art,art_translation_pt=dict(x=8,y=6),semantic_internals_editable=False)),indent=2)+"\n")
print("wrote Wave 3 review, overflow, art translation and source evidence")
