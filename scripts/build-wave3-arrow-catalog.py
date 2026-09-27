#!/usr/bin/env python3
"""Catalog pinned West Monroe hand-drawn SVG arrows with audited geometry.

The SVG files are copied byte-for-byte to the local fixture directory. Preview
overlays embed those bytes as a data URI; endpoint/tangent markers are additional
review annotations and are not written into the source artwork.
"""
from __future__ import annotations

import argparse
import base64
import hashlib
import json
import math
import re
import shutil
import struct
import subprocess
import tempfile
import urllib.request
import zipfile
import zlib
from pathlib import Path
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_BRAND_ROOT = Path('/Users/rscott/Documents/branding')
DEFAULT_INVENTORY = Path('/Users/rscott/.codex/skills/wm-brand-assets/references/asset-inventory.json')
DEFAULT_OUT = ROOT / 'samples/visual-wave3/arrow-assets'
DEFAULT_CATALOG = ROOT / 'library/diagram-components/arrow-assets.json'

# Control points below are taken from the named SVG artwork paths. They describe
# semantic visible tips/tails, not the rectangular image frame. Head tips may be
# at the contour extremum while shaft centerlines stop short of the head.
ASSETS = [
    {
        'id': 'wm-handdrawn-single-arrow',
        'filename': 'handdrawn-single-arrow.svg',
        'fallback_filename': 'handdrawn-single-arrow.png',
        'legacy_inventory_path': 'handdrawn-animations/svg/handdrawn-single-arrow.svg',
        'legacy_fallback_inventory_path': 'handdrawn-animations/png/handdrawn-single-arrow.png',
        'hosted_source': True,
        'direction': 'left-to-right, single tip',
        'bend': 'straight horizontal shaft with a hand-drawn filled head',
        'geometry': {
            'tail': {'point': [5.48, 20.37], 'tangent': [1.0, 0.0],
                     'evidence': 'first shaft path begins M24.39,22.46; its relative cubic ends at (5.48,20.37), the leftmost rounded shaft tail. The shaft axis enters the long segment to the right.'},
            'tip': {'point': [1189.37, 18.37], 'tangent': [1.0, 0.0],
                    'evidence': 'second filled head path ends its upper-right cubic at (1189.37,18.37), the rightmost head apex; the first filled shaft path terminates at x=1155.5 and leads rightward into this head.'},
        },
        'confidence': 'high geometric; low placement',
        'confidence_note': 'Visible shaft tail and filled head apex are explicit endpoints in the SVG path data and confirmed in the preview. Native optical attachment remains untested.',
        'svg_description': {'title':'Handdrawn single arrow','desc':'Handdrawn single arrow annotation for pointing to a specific item or direction.'},
    },
    {
        'id': 'wm-handdrawn-connecting-arrow',
        'filename': 'wm_handrawn_connecting_arrow_rgb_240912.svg',
        'fallback_filename': 'wm_handdrawn_connecting_arrow_rgb_240912.png',
        'inventory_path': 'assets/graphics/arrows/wm_handrawn_connecting_arrow_rgb_240912.svg',
        'legacy_inventory_path': 'handdrawn-animations/svg/handdrawn-connecting-arrow.svg',
        'legacy_fallback_inventory_path': 'handdrawn-animations/png/handdrawn-connecting-arrow.png',
        'direction': 'upper-right tail, looping down/left/up into a right-facing tip',
        'bend': 'large three-sided return loop with horizontal exit',
        'geometry': {
            'tail': {'point': [53.0, 22.1], 'tangent': [0.0, 1.0],
                     'evidence': 'artwork shaft path class st3 begins M53,22.1 v32.5; this open start is the upper end of the vertical return stroke, with initial direction downward into the loop.'},
            'tip': {'point': [48.3, 21.1], 'tangent': [1.0, 0.0],
                    'evidence': 'artwork head path class st3 starts M42.2,16.6 and its first cubic ends at (48.3,21.1), the right-facing arrowhead apex. Main shaft path class st3 ends at (48.3,22.4), approaching the head from left to right.'},
        },
        'confidence': 'medium-high geometric; low placement',
        'confidence_note': 'The tail is the actual open start of the main stroked path; the tip is the explicit right-facing head apex in a separate stroke path. Geometry is visible in the rendered preview. Native optical attachment remains untested.',
    },
    {
        'id': 'wm-handdrawn-dashed-arrow',
        'filename': 'wm_handrawn_dashed_arrow_rgb_240912.svg',
        'fallback_filename': 'wm_handdrawn_dashed_arrow_rgb_240912.png',
        'inventory_path': 'assets/graphics/arrows/wm_handrawn_dashed_arrow_rgb_240912.svg',
        'legacy_inventory_path': 'handdrawn-animations/svg/handdrawn-dashed-arrow.svg',
        'legacy_fallback_inventory_path': 'handdrawn-animations/png/handdrawn-dashed-arrow.png',
        'direction': 'curved upper-left to lower-right; single tip',
        'bend': 'long descending arc with shallow lower sweep',
        'geometry': {
            'tail': {'point': [3.0, 2.7], 'tangent': [-0.2873478856, 0.9578262852],
                     'evidence': 'artwork path class st6, first dash path begins M3,2.7 and first nonzero cubic control is (-0.1,0.3); tangent is that cubic start direction.'},
            'tip': {'point': [71.7, 61.6], 'tangent': [0.954415, -0.298486],
                    'evidence': 'artwork path class st0 head contour starts M71.7,61.6. Shaft path class st7 final relative cubic ends at (64.2,65.4), with end derivative 3*(25.2,-7.9), used as incoming tangent.'},
        },
        'confidence': 'medium',
        'confidence_note': 'The centerline and head contour are distinct SVG paths. Endpoint is the explicit head contour apex; tangent uses the final shaft cubic derivative. Native optical attachment remains untested.',
        'uhg_source': {'source': 'UHG Fabric Platforming RFP Response - July 2026.pptx',
                       'source_sha256': 'b0f254ed7739768d0f345257689264393049d06cdd3849a177fc764f3d348d99',
                       'media_part': 'ppt/media/image86.svg', 'media_sha256': 'a0902a239b555ed25b977ac65ba87749aad2cb3474450c1194a262f65673cea7',
                       'byte_identity_with_local_brand_asset': False,
                       'relationship': 'same hand-drawn dashed-arrow artwork/geometry family; local branding SVG has a different hash and must not be described as a byte-identical copy of the PPTX media.',
                       'comparison': {'method':'Mac sips rasterization, both SVGs rendered at 720x720 RGBA', 'exact_pixels_fraction':0.9999922839506172, 'max_channel_difference':3, 'mean_absolute_channel_difference':0.000004822530864197531, 'qualification':'visual raster similarity for this renderer and size only; not byte identity or native PowerPoint qualification'},
                       'observed_object': 'UHG5 Graphic 20; UHG14 Graphic 17',
                       'note': 'The UHG source media hash differs from the local branding SVG hash. UHG5 notes describe a navy dashed arrow curving down/right toward OUR UNDERSTANDING. Original source rotation 345.7749167 degrees is not generalized.'},
    },
    {
        'id': 'wm-handdrawn-right-angle-arrow',
        'filename': 'wm_handrawn_right_angle_arrow_rgb_240912.svg',
        'fallback_filename': 'wm_handrawn_right_angle_arrow_rgb_240912.png',
        'inventory_path': 'assets/graphics/arrows/wm_handrawn_right_angle_arrow_rgb_240912.svg',
        'legacy_inventory_path': 'handdrawn-animations/svg/handdrawn-right-angle-arrow.svg',
        'legacy_fallback_inventory_path': 'handdrawn-animations/png/handdrawn-right-angle-arrow.png',
        'direction': 'down then right; single tip',
        'bend': 'one near-right-angle elbow',
        'geometry': {
            'tail': {'point': [19.0, 14.8], 'tangent': [0.009132, 0.999958],
                     'evidence': 'artwork path class st3 starts M19,14.8 c0,0 0.2,21.9; first nonzero departure is toward control point (19.2,36.7), approximately downward.'},
            'tip': {'point': [53.7, 46.6], 'tangent': [0.953583, -0.301132],
                    'evidence': 'artwork filled path class st0 head contour starts M53.7,46.6. Shaft path class st3 reaches (51.8,47.2) immediately before the head; that last shaft-to-head segment defines incoming tangent.'},
        },
        'confidence': 'medium',
        'confidence_note': 'Tail follows the open shaft start, tip follows the filled head apex, and local direction follows actual shaft/head approach. Exact source geometry is clear; optical attachment remains untested.',
    },
]


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open('rb') as f:
        for block in iter(lambda: f.read(1 << 20), b''):
            h.update(block)
    return h.hexdigest()


def inventory_asset(inventory: dict, pth: str) -> dict | None:
    return next((x for x in inventory.get('assets', []) if x.get('path') == pth), None)


def png_alpha_bounds(path: Path) -> list[float]:
    """Return tight alpha bounds in rendered pixels as [x,y,width,height]."""
    data = path.read_bytes()
    if data[:8] != b'\x89PNG\r\n\x1a\n':
        raise ValueError(f'not a PNG: {path}')
    pos = 8
    compressed = bytearray()
    while pos < len(data):
        n = struct.unpack('>I', data[pos:pos+4])[0]
        tag = data[pos+4:pos+8]
        chunk = data[pos+8:pos+8+n]
        pos += 12 + n
        if tag == b'IHDR':
            width, height, bit_depth, color_type, comp, filt, interlace = struct.unpack('>IIBBBBB', chunk)
        elif tag == b'IDAT':
            compressed.extend(chunk)
        elif tag == b'IEND':
            break
    if bit_depth != 8 or color_type != 6 or interlace != 0:
        raise ValueError(f'Expected noninterlaced RGBA8 PNG, got depth={bit_depth}, type={color_type}, interlace={interlace}')
    raw = zlib.decompress(compressed)
    stride, bpp = width * 4, 4
    prev = bytearray(stride)
    lo_x, lo_y, hi_x, hi_y = width, height, -1, -1
    at = 0
    for y in range(height):
        ft = raw[at]; at += 1
        row = bytearray(raw[at:at+stride]); at += stride
        for i in range(stride):
            a = row[i-bpp] if i >= bpp else 0
            b = prev[i]
            c = prev[i-bpp] if i >= bpp else 0
            if ft == 1: row[i] = (row[i] + a) & 255
            elif ft == 2: row[i] = (row[i] + b) & 255
            elif ft == 3: row[i] = (row[i] + ((a+b)//2)) & 255
            elif ft == 4:
                p = a + b - c
                pa, pb, pc = abs(p-a), abs(p-b), abs(p-c)
                pr = a if pa <= pb and pa <= pc else (b if pb <= pc else c)
                row[i] = (row[i] + pr) & 255
            elif ft != 0: raise ValueError(f'unsupported PNG filter {ft}')
        for x in range(width):
            if row[x*4+3] > 0:
                lo_x=min(lo_x,x); lo_y=min(lo_y,y); hi_x=max(hi_x,x); hi_y=max(hi_y,y)
        prev = row
    if hi_x < lo_x:
        return [0, 0, 0, 0]
    return [lo_x, lo_y, hi_x-lo_x+1, hi_y-lo_y+1]


def fmt(x: float) -> str:
    return f'{x:.6g}'


def preview_svg(asset: dict, original_bytes: bytes) -> str:
    data = base64.b64encode(original_bytes).decode('ascii')
    vb=asset['viewBox']; vx,vy,vw,vh=vb
    tall=vw/vh < 3.0
    out_h=97 if tall else 50
    scale=min(82/vw,68/vh)
    draw_w,draw_h=vw*scale,vh*scale
    ix=(96-draw_w)/2
    iy=12+(68-draw_h)/2 if tall else 12+(24-draw_h)/2
    footer_y=94.5 if tall else 47.0
    w,h=960,out_h*10
    elems = [f'''<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" viewBox="0 0 96 {out_h}">
<rect width="96" height="{out_h}" fill="#f4f5f7"/>
<text x="3" y="5" font-family="Arial,sans-serif" font-size="2.1" font-weight="700" fill="#111">{asset['id']}</text>
<text x="3" y="8" font-family="Arial,sans-serif" font-size="1.25" fill="#333">{asset['direction']} · {asset['bend']}</text>
<rect x="{fmt(ix)}" y="{fmt(iy)}" width="{fmt(draw_w)}" height="{fmt(draw_h)}" fill="white" stroke="#aab0b8" stroke-width="0.12"/>
<image x="{fmt(ix)}" y="{fmt(iy)}" width="{fmt(draw_w)}" height="{fmt(draw_h)}" href="data:image/svg+xml;base64,{data}"/>
''']
    colors = {'tail': '#00a7d6', 'tip': '#e7334c'}
    for name, endpoint in asset['geometry'].items():
        x, y = endpoint['point']; px=ix+(x-vx)*scale; py=iy+(y-vy)*scale
        tx, ty=endpoint['tangent']; ln=3.1
        elems.append(f'<line x1="{fmt(px-tx*ln/2)}" y1="{fmt(py-ty*ln/2)}" x2="{fmt(px+tx*ln/2)}" y2="{fmt(py+ty*ln/2)}" stroke="{colors[name]}" stroke-width="0.34"/>')
        elems.append(f'<circle cx="{fmt(px)}" cy="{fmt(py)}" r="0.62" fill="{colors[name]}" stroke="white" stroke-width="0.18"/>')
        if name == 'tip':
            lx,ly=px+2.0,py-1.5
        elif px < 16:
            lx,ly=px+2.0,py+3.0
        else:
            lx,ly=px-8.4,py+2.8
        elems.append(f'<path d="M{fmt(px)},{fmt(py)} L{fmt(lx)},{fmt(ly)}" stroke="{colors[name]}" stroke-width="0.17" fill="none"/>')
        elems.append(f'<text x="{fmt(lx)}" y="{fmt(ly-0.35)}" font-family="Arial,sans-serif" font-size="1.35" font-weight="700" fill="{colors[name]}">{name.upper()} {fmt(x)},{fmt(y)}</text>')
    if 'uhg_source' in asset:
        footer='UHG5/14 use related dashed-arrow artwork; source and local SVG bytes differ'
    else:
        footer='WM branding source SVG; no UHG source instance located in this audit'
    elems.append(f'<text x="3" y="{footer_y}" font-family="Arial,sans-serif" font-size="1.25" fill="#333">{footer}</text>')
    elems.append('</svg>')
    return '\n'.join(elems)


def main() -> None:
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--brand-root',type=Path,default=DEFAULT_BRAND_ROOT)
    ap.add_argument('--inventory',type=Path,default=DEFAULT_INVENTORY)
    ap.add_argument('--out',type=Path,default=DEFAULT_OUT)
    ap.add_argument('--catalog',type=Path,default=DEFAULT_CATALOG)
    args=ap.parse_args()
    inventory=json.loads(args.inventory.read_text())
    if args.out.exists():
        raise SystemExit(f'refusing to overwrite existing preview directory: {args.out}')
    args.out.mkdir(parents=True)
    (args.out/'source').mkdir()
    records=[]
    with tempfile.TemporaryDirectory(prefix='wave3-arrow-') as tmp:
        tmp=Path(tmp)
        for spec in ASSETS:
            src_path=(args.brand_root/spec['inventory_path']) if spec.get('inventory_path') else None
            inv=inventory_asset(inventory,spec['legacy_inventory_path'])
            if inv is None: raise ValueError(f"asset inventory has no exact path {spec['legacy_inventory_path']}")
            fallback_inv=inventory_asset(inventory,spec['legacy_fallback_inventory_path'])
            if fallback_inv is None: raise ValueError(f"asset inventory has no exact path {spec['legacy_fallback_inventory_path']}")
            copy=args.out/'source'/spec['filename']
            fallback_copy=args.out/'source'/spec['fallback_filename']
            if src_path is not None and src_path.is_file():
                content=src_path.read_bytes(); shutil.copyfile(src_path,copy)
                source_description_path=str(src_path)
            else:
                host=inventory.get('assetHostDefault','').rstrip('/')
                url=host+'/'+spec['legacy_inventory_path']
                content=urllib.request.urlopen(url).read(); copy.write_bytes(content)
                source_description_path=url
            fallback_src=(args.brand_root/'assets/graphics/arrows'/spec['fallback_filename']) if src_path is not None else None
            if fallback_src is not None and fallback_src.is_file():
                fallback_content=fallback_src.read_bytes(); shutil.copyfile(fallback_src,fallback_copy)
                fallback_description_path=str(fallback_src)
            else:
                url=inventory.get('assetHostDefault','').rstrip('/')+'/'+spec['legacy_fallback_inventory_path']
                fallback_content=urllib.request.urlopen(url).read(); fallback_copy.write_bytes(fallback_content)
                fallback_description_path=url
            # Locally branded copies can differ byte-for-byte from the older
            # hosted inventory variants. Enforce inventory lengths only when
            # bytes were fetched from that exact hosted path.
            if (src_path is None or not src_path.is_file()) and len(content)!=inv.get('sizeBytes'):
                raise ValueError(f"SVG byte size differs from hosted inventory for {spec['legacy_inventory_path']}")
            if (fallback_src is None or not fallback_src.is_file()) and len(fallback_content)!=fallback_inv.get('sizeBytes'):
                raise ValueError(f"PNG byte size differs from hosted inventory for {spec['legacy_fallback_inventory_path']}")
            source_hash=hashlib.sha256(content).hexdigest()
            if sha(copy)!=source_hash: raise ValueError(f'source copy hash mismatch: {copy}')
            if sha(fallback_copy)!=hashlib.sha256(fallback_content).hexdigest(): raise ValueError(f'fallback source copy hash mismatch: {fallback_copy}')
            root=ET.fromstring(content)
            try: vbox=[float(v) for v in root.attrib['viewBox'].replace(',',' ').split()]
            except (KeyError,ValueError): raise ValueError(f'source SVG missing a numeric viewBox: {copy}')
            if len(vbox)!=4 or vbox[2]<=0 or vbox[3]<=0: raise ValueError(f'invalid source viewBox: {copy}')
            vx,vy,vw,vh=vbox
            aspect=vw/vh
            raster=tmp/(spec['id']+'.png')
            subprocess.run(['sips','-s','format','png','-z','720','720',str(copy),'--out',str(raster)],check=True,stdout=subprocess.DEVNULL)
            alpha_pixels=png_alpha_bounds(raster)
            alpha=[vx+alpha_pixels[0]*vw/720,vy+alpha_pixels[1]*vh/720,alpha_pixels[2]*vw/720,alpha_pixels[3]*vh/720]
            desc=spec.copy()
            alpha_n={'x':alpha_pixels[0]/720,'y':alpha_pixels[1]/720,'width':alpha_pixels[2]/720,'height':alpha_pixels[3]/720}
            desc.update({
                'status':'manual-candidate',
                'asset_path':str(copy.relative_to(ROOT)),
                'asset_sha256':source_hash,
                'fallback_asset_path':str(fallback_copy.relative_to(ROOT)),
                'fallback_asset_sha256':sha(fallback_copy),
                'intrinsic_width':vw,
                'intrinsic_height':vh,
                'alpha_bounds':alpha_n,
                'tail':{'x':(spec['geometry']['tail']['point'][0]-vx)/vw,'y':(spec['geometry']['tail']['point'][1]-vy)/vh},
                'tip':{'x':(spec['geometry']['tip']['point'][0]-vx)/vw,'y':(spec['geometry']['tip']['point'][1]-vy)/vh},
                'tail_tangent':spec['geometry']['tail']['tangent'],
                'tip_tangent':spec['geometry']['tip']['tangent'],
                'min_span_pt':None,
                'max_span_pt':None,
                'span_status':'not calibrated; null bounds require manual placement',
                'max_rotation_delta_deg':None,
                'min_uniform_scale':None,
                'max_uniform_scale':None,
                'proposed_rotation_delta_deg':0.0,
                'proposed_uniform_scale':[0.9,1.1],
                'qualification':'unqualified',
                'proposed_experiment':{'uniform_scale_factors':[0.9,1.0,1.1],
                                       'rotation_deltas_deg':[0,-15,15],
                                       'span_envelope_pt':([70.29*0.9,70.29*1.1] if 'uhg_source' in spec else None),
                                       'span_envelope_basis':('0.9–1.1 of observed UHG5 image frame width 70.29pt; exploratory short/long render only, not calibrated visible endpoint span.' if 'uhg_source' in spec else 'No source PowerPoint frame-size measurement found for this local asset; test only as 0.9–1.1 of a declared fixture size, with fixture size recorded separately.'),
                                       'review_only':True,
                                       'note':'Proposed conservative baseline is uniform scale 0.9–1.1 at 0 degrees for fixture exploration only; ±15 degree cases are deliberate alternate-angle stress controls. These are not production limits. Production point-span values remain null without source placement calibration.'},
                'nonuniform_scale_allowed':False,
                'flip_allowed':False,
                'source':{'path':source_description_path,'inventory_relative_path':spec.get('inventory_path',spec['legacy_inventory_path']),
                          'legacy_hosted_inventory_path':spec['legacy_inventory_path'],
                          'inventory_record':inv,'sha256':source_hash,'bytes':len(content),
                          'preserved_copy':str(copy.relative_to(ROOT)),
                          'fallback_source_path':fallback_description_path,'fallback_inventory_path':spec['legacy_fallback_inventory_path'],
                          'fallback_inventory_record':fallback_inv,'fallback_sha256':sha(fallback_copy),
                          'preserved_fallback_copy':str(fallback_copy.relative_to(ROOT)),
                          'verbatim_inventory_description_sources':inv.get('description_sources',[]),
                          'inventory_category':inv.get('category'),'inventory_category_status':inv.get('category_status'),
                          'inventory_variant_status':inv.get('variant_status')},
                'inventory_visual_description':'No visual description is present in the exact hosted inventory record (description_sources is empty); direction/bend labels below are this audit\'s observed descriptions, not copied inventory claims.',
                'brand_guidance':{'source':'/Users/rscott/Documents/branding/guidance/Hand-drawn Graphics  West Monroe Brand Compass.md',
                                  'excerpt':'Hand-drawn graphics signify human interaction and add a personal, human touch to our visual identity. These graphics enhance the fluidity of our compositions, adding meaning and drawing attention to specific areas. Use hand-drawn elements as a supplemental tool to enrich layouts, making them more dynamic and engaging.'},
                'native_aspect_ratio':aspect,
                'viewBox':vbox,
                'alpha_bounds_viewbox':alpha,
                'transform_limits_status':'proposed conservative limits; not validated by native render',
                'source_geometry_notes':spec['geometry'],
                'placement_qualification':'not_native_or_optically_qualified',
            })
            psvg=args.out/(spec['id']+'-annotated.svg')
            psvg.write_text(preview_svg(desc,content))
            ppng=args.out/(spec['id']+'-annotated.png')
            subprocess.run(['sips','-s','format','png',str(psvg),'--out',str(ppng)],check=True,stdout=subprocess.DEVNULL)
            desc['annotated_preview']={'svg':str(psvg.relative_to(ROOT)),'svg_sha256':sha(psvg),
                                      'png':str(ppng.relative_to(ROOT)),'png_sha256':sha(ppng)}
            if 'uhg_source' in spec:
                deck=ROOT/'samples/UHG Fabric Platforming RFP Response - July 2026.pptx'
                if sha(deck)!=spec['uhg_source']['source_sha256']:
                    raise ValueError('UHG source deck hash differs from pinned provenance')
                with zipfile.ZipFile(deck) as archive:
                    uhg_bytes=archive.read(spec['uhg_source']['media_part'])
                uhg_hash=hashlib.sha256(uhg_bytes).hexdigest()
                if uhg_hash!=spec['uhg_source']['media_sha256']:
                    raise ValueError('UHG SVG media hash differs from pinned provenance')
                uhg_copy=args.out/'source/UHG-image86.svg'; uhg_copy.write_bytes(uhg_bytes)
                uhg_svg=args.out/(spec['id']+'-uhg-media-annotated.svg')
                uhg_svg.write_text(preview_svg(desc,uhg_bytes))
                uhg_png=args.out/(spec['id']+'-uhg-media-annotated.png')
                subprocess.run(['sips','-s','format','png',str(uhg_svg),'--out',str(uhg_png)],check=True,stdout=subprocess.DEVNULL)
                desc['uhg_source']['preserved_copy']=str(uhg_copy.relative_to(ROOT))
                desc['uhg_source']['preserved_copy_sha256']=sha(uhg_copy)
                desc['uhg_source']['annotated_preview']={'svg':str(uhg_svg.relative_to(ROOT)),'svg_sha256':sha(uhg_svg),
                                                        'png':str(uhg_png.relative_to(ROOT)),'png_sha256':sha(uhg_png)}
            records.append(desc)
    catalog={'schema':'pptxgengo.wave3-arrow-assets.v1','status':'curated_geometry_candidates_not_native_qualified',
             'description_provenance':'WM Brand Asset inventory paths are preserved verbatim; local branding-file paths and source bytes/hashes are pinned per entry.',
             'endpoint_convention':'Normalized viewBox coordinates use SVG y-down axes. Tangents are unit vectors in the same local coordinate system. These visible endpoints are geometric candidates observed from specific SVG path contours and rendered previews; they are not calibrated PowerPoint attachment points.',
             'assets':records}
    args.catalog.parent.mkdir(parents=True,exist_ok=True)
    if args.catalog.exists(): raise SystemExit(f'refusing to overwrite catalog: {args.catalog}')
    args.catalog.write_text(json.dumps(catalog,indent=2,ensure_ascii=False)+'\n')
    print(json.dumps({'catalog':str(args.catalog.relative_to(ROOT)),'assets':len(records),'preview_dir':str(args.out.relative_to(ROOT))},indent=2))

if __name__=='__main__': main()
