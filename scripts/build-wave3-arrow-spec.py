#!/usr/bin/env python3
"""Candidate native experiments for endpoint-constrained WM arrow artwork.

The fixture geometry deliberately controls endpoint spans, not picture widths.
No experiment envelope below is a production qualification claim.
"""
import copy
import json
import math
import hashlib
import subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'library/diagram-components'
CATALOG=OUT/'arrow-assets.json'
NAVY='#070154'

def ink_regions(asset):
    source=ROOT/asset['asset_path']
    out=ROOT/'samples/visual-wave3/arrow-ink'/asset['asset_sha256']
    out.mkdir(parents=True,exist_ok=True)
    raster=out/'source-720.png'
    if not raster.exists():
        subprocess.run(['sips','-s','format','png','-z','720','720',str(source),'--out',str(raster)],check=True,stdout=subprocess.DEVNULL)
    digest=hashlib.sha256(raster.read_bytes()).hexdigest()
    info=out/'ink-regions.json'
    if not info.exists():
        result=subprocess.run(['swift',str(ROOT/'scripts/artwork-ink-regions.swift'),str(raster)],check=True,capture_output=True,text=True)
        data=json.loads(result.stdout)
        data.update(source_svg_sha256=asset['asset_sha256'],raster_sha256=digest)
        info.write_text(json.dumps(data,indent=2)+'\n')
    data=json.loads(info.read_text())
    if data['source_svg_sha256'] != asset['asset_sha256'] or data['raster_sha256'] != digest:
        raise ValueError('stale ink geometry: '+str(info))
    return data['regions'],f"{data['method']}; original SVG {asset['asset_sha256']}; sips normalized 720x720 raster {digest}; native visual verification still required"

def derived_fallback(asset):
    # Stock PNG counterparts sometimes have a different canvas/crop. Render the
    # exact pinned SVG to its own aspect ratio instead of stretching those PNGs.
    path=ROOT/'samples/visual-wave3/arrow-ink'/asset['asset_sha256']/'fallback.png'
    if not path.exists():
        width=math.ceil(asset['intrinsic_width']*5)
        height=math.ceil(asset['intrinsic_height']*5)
        subprocess.run(['sips','-s','format','png','-z',str(height),str(width),str(ROOT/asset['asset_path']),'--out',str(path)],check=True,stdout=subprocess.DEVNULL)
    return str(path.relative_to(ROOT)),hashlib.sha256(path.read_bytes()).hexdigest()

def rect(x,y,w,h):return dict(x=x,y=y,width=w,height=h)
def text(id,copy,bounds,size=14):
    return dict(id=id,kind='text',text=copy,bounds=bounds,font_face='Arial',font_size_pt=size,foreground=NAVY,align='left',valign='top')
def rotate(p,a):
    c,s=math.cos(a),math.sin(a)
    return (c*p[0]-s*p[1],s*p[0]+c*p[1])
def slide(asset,factor,degrees,n):
    # Fixed source-port location. Target follows the unchanged art endpoint vector.
    # All placement in the output spec is anchor-relative; no image coordinates.
    base=180 if 'dashed' in asset['id'] else 240
    scale=base*factor/asset['intrinsic_width']
    theta=math.radians(degrees)
    delta=((asset['tip']['x']-asset['tail']['x'])*asset['intrinsic_width']*scale,
           (asset['tip']['y']-asset['tail']['y'])*asset['intrinsic_height']*scale)
    dx,dy=rotate(delta,theta)
    tail=(350,240)
    tip=(tail[0]+dx,tail[1]+dy)
    start_tan=rotate(asset['tail_tangent'],theta)
    end_tan=rotate(asset['tip_tangent'],theta)
    # Choose natural edge normals from visible tangent directions, not the bbox.
    def port(v,incoming=False):
        x,y=v
        if incoming:x,y=-x,-y
        return ('right' if x>0 else 'left') if abs(x)>abs(y) else ('bottom' if y>0 else 'top')
    fp,tp=port(start_tan),port(end_tan,True)
    def label_bounds(point,edge):
        x,y=point;w,h,g=126,42,20
        if edge=='bottom':return rect(x-w/2,y-g-h,w,h)
        if edge=='top':return rect(x-w/2,y+g,w,h)
        if edge=='left':return rect(x+g,y-h/2,w,h)
        return rect(x-g-w,y-h/2,w,h)
    art={k:copy.deepcopy(asset[k]) for k in ('id','asset_path','asset_sha256','fallback_asset_path','fallback_asset_sha256','intrinsic_width','intrinsic_height','alpha_bounds','tail','tip')}
    for key in ('tail_tangent','tip_tangent'):
        art[key]=dict(zip(('x','y'),asset[key]))
    natural=math.hypot((asset['tip']['x']-asset['tail']['x'])*base,
                       (asset['tip']['y']-asset['tail']['y'])*base*asset['intrinsic_height']/asset['intrinsic_width'])
    art.update(min_span_pt=natural*.7,max_span_pt=natural*1.3,max_rotation_delta_deg=16,
               endpoint_provenance='Candidate geometric endpoints from arrow-assets.json; native optical qualification pending.')
    art['ink_regions'],art['ink_regions_provenance']=ink_regions(asset)
    art['fallback_asset_path'],art['fallback_asset_sha256']=derived_fallback(asset)
    id=asset['id'].replace('wm-handdrawn-','')+f'-{factor:g}-{degrees:+g}'
    # Reuse the existing original WM footer bytes and declaration.
    base_slide=json.loads((OUT/'accent-review.json').read_text())['slides'][0]
    canvas=copy.deepcopy(base_slide['canvas'][:4])
    canvas[3]['text']=str(n)
    canvas += [text('experiment-title',f"{asset['id'].replace('wm-handdrawn-','').replace('-',' ').title()} / {factor:g}× size / {degrees:+g}°",rect(36,40,888,42),24),
               text('description','Candidate placement uses visible tail and tip coordinates, uniform scale, and measured approach directions. Native visual review remains required.',rect(36,92,860,40),14),
               text('source','Decision ready',label_bounds(tail,fp),14),
               text('target','Review evidence\nand release',label_bounds(tip,tp),14),
               text('qualification','ILLUSTRATIVE GEOMETRY EXPERIMENT • Not a qualified reusable placement envelope',rect(36,456,850,28),11)]
    return dict(id=id,title='',role='Arrow placement qualification',takeaway='Artwork should meet the intended source and target without distortion or accidental overlap.',notes=f"Candidate source {asset['asset_sha256']}; proposed base frame {base}pt, uniform factor {factor}, rotation {degrees} degrees; actual endpoint span {natural*factor:.4f}pt. No source slide identity claimed.",width_pt=960,height_pt=540,title_bounds=rect(0,0,0,0),title_font_face='Arial',title_font_size_pt=24,title_foreground=NAVY,pods=[],canvas=canvas,
                artwork_arrows=[dict(id='relationship',**{'from':dict(target='source',port=fp,gap_pt=20),'to':dict(target='target',port=tp,gap_pt=20)},artwork=art,clearance_pt=2)])

def main():
    assets=json.loads(CATALOG.read_text())['assets']
    # Return-loop endpoints are too close for a general source-to-target contract.
    usable=[a for a in assets if 'connecting-arrow' not in a['id']]
    slides=[]
    for a in usable:
        for factor,angle in [(.75,0),(1.25,-15),(1,15)]:
            slides.append(slide(a,factor,angle,len(slides)+1))
    spec=dict(schema='pptxgengo.compose-spec.v1',slides=slides)
    (OUT/'arrow-review.json').write_text(json.dumps(spec,indent=2,ensure_ascii=False)+'\n')
    # Explicit failed direction stages the original chosen art with an actionable note.
    failed=copy.deepcopy(next(s for s in slides if s['id'].startswith('dashed-arrow')));failed['id']='arrow-direction-manual'
    arrow=failed['artwork_arrows'][0];arrow['from']['port']='top'
    arrow['staging']=dict(asset_bounds=rect(640,185,160,160),note_bounds=rect(640,365,284,80),note='Move this dashed arrow from “Decision ready” toward “Review evidence and release.” The chosen source port faces away from the artwork tangent; choose a compatible source edge before positioning it.')
    failed['canvas'][4]['text']='Incompatible direction requires explicit placement review'
    (OUT/'arrow-manual.json').write_text(json.dumps(dict(schema=spec['schema'],slides=[failed]),indent=2,ensure_ascii=False)+'\n')
    (OUT/'arrow-qualification.json').write_text(json.dumps(dict(schema=spec['schema'],slides=slides+[failed]),indent=2,ensure_ascii=False)+'\n')
    print('wrote',len(slides),'arrow placement candidates and one manual fallback')
if __name__=='__main__':main()
