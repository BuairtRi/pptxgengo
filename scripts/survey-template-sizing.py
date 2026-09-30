#!/usr/bin/env python3
"""Source-grounded text/image geometry and heuristic authoring budget inventory."""
import collections, json, math, re, subprocess, sys, tempfile
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
ASSIGN=ROOT/'library/templates/rollout/assignments.json'
EMU_PER_IN=914400
EMU_PER_PT=12700
SLIDE_W=12192000
SLIDE_H=6858000

def read(p): return json.loads(Path(p).read_text())
def write(p,x):
    Path(p).parent.mkdir(parents=True,exist_ok=True)
    Path(p).write_text(json.dumps(x,indent=2,ensure_ascii=False)+'\n')
def prop(obj, prefix):
    return {x.get('property'):x.get('value') for x in obj.get('geometry',[]) if x.get('property','').startswith(prefix)}
def get_sizes(typography):
    vals=[]
    for x in typography or []:
        if x.get('property','').endswith('.rPr.sz') and str(x.get('value','')).isdigit():
            v=float(x['value'])
            vals.append(v/100 if v>100 else v)
    return vals
def attrs_for(obj, source_info):
    # catalog-geometry provides fully transformed frames and explicit text-style audit for 4 registered decks.
    g=source_info.get(str(obj.get('object_id')))
    if g: return g
    geo=obj.get('geometry',[])
    vals={x.get('property'):x.get('value') for x in geo}
    try:
        x,y,w,h=[float(vals[k]) for k in ('transform.off.x','transform.off.y','transform.ext.cx','transform.ext.cy')]
    except (KeyError,ValueError,TypeError):
        return None
    # Direct guide transforms are used only for top-level objects; nested-group transforms are unresolved.
    ancestor=(geo[0].get('ancestors',[]) if geo else [])
    if 'p:grpSp' in ancestor: return {'geometry_status':'unresolved_nested_group','shape_name':obj.get('name'),'shape_id':str(obj.get('object_id'))}
    return {'geometry_status':'direct_source_frame','shape_name':obj.get('name'),'shape_id':str(obj.get('object_id')),
            'bounds_inches':{'x':x/EMU_PER_IN,'y':y/EMU_PER_IN,'width':w/EMU_PER_IN,'height':h/EMU_PER_IN},
            'object_kind':'unknown','transform_provenance':{'kind':'binding_guide_direct_frame'},
            'explicit_text_style':None,'inherited_typography_status':'unresolved'}

def main():
    assignments=read(ASSIGN)['entries']
    # Reuse the existing transformed source-deck geometry inventory. AI source-scene guides
    # supplement it with direct transforms and are flagged when a group transform is needed.
    tmp=tempfile.TemporaryDirectory(prefix='pptx-sizing-')
    geom=Path(tmp.name)/'geometry.jsonl'; report=Path(tmp.name)/'geometry-report.json'
    subprocess.run([sys.executable,str(ROOT/'scripts/catalog-geometry.py'),'--registry',str(ROOT/'planning/source-registry.json'),
                    '--out',str(geom),'--report',str(report)],check=True,cwd=ROOT,stdout=subprocess.DEVNULL)
    geo_rows=[json.loads(x) for x in geom.read_text().splitlines() if x]
    byslide=collections.defaultdict(list)
    for r in geo_rows: byslide[(r['source_id'],int(r['slide_number']))].append(r)
    result=[]; categories=collections.Counter(); missing=[]
    for a in assignments:
        tid=a['template_id']; source,ns=a['representative'].rsplit(':',1); slide=int(ns)
        impl=ROOT/a['implementation_directory']; contract_path=impl/'contract.json'
        if not contract_path.exists(): missing.append(tid); contract={}
        else: contract=read(contract_path)
        project=ROOT/a['source_project']; guide_path=project/f'binding-guide-{slide:03}.json'
        guide=read(guide_path) if guide_path.exists() else None
        if guide is None:
            candidates=list(project.glob(f'**/binding-guide-{slide:03}.json')) if project.exists() else []
            guide=read(candidates[0]) if candidates else {'objects':[]}
        objects={str(o.get('object_id')):o for o in guide.get('objects',[])}
        slide_geo=byslide.get((source,slide),[])
        geom_byid=collections.defaultdict(list)
        for gr in slide_geo:
            if gr.get('shape_id') is not None: geom_byid[str(gr['shape_id'])].append(gr)
        source_info={oid:(rows[0] if len(rows)==1 else next((r for r in rows if not r.get('parent_object_path')),rows[0]))
                     for oid,rows in geom_byid.items()}
        # Build shape-zone to slot associations; rich text runs sharing the same source shape
        # are counted together exactly once, preserving spaces and source boundaries.
        slot_by_oid=collections.defaultdict(list)
        binding_to_slot={}
        for slot,spec in contract.get('slots',{}).items():
            for bid in spec.get('binding_ids',[]) if isinstance(spec,dict) else []:
                parts=bid.split('/')
                oid=parts[1] if len(parts)>2 and parts[0]=='b' else None
                if oid:
                    slot_by_oid[oid].append(slot); binding_to_slot[bid]=slot
        zones=[]
        for oid,slots in slot_by_oid.items():
            obj=objects.get(oid)
            if not obj: continue
            chars=''.join(x.get('value','') for x in obj.get('text',[]) if x.get('property')=='text')
            gr=attrs_for(obj,source_info)
            bounds=gr.get('bounds_inches') if gr else None
            typography=obj.get('typography',[])
            sizes=get_sizes(typography)
            unique=sorted(set(round(x,2) for x in sizes))
            run_count=len([x for x in obj.get('text',[]) if x.get('property')=='text'])
            font_entries=[x for x in typography if x.get('property','').endswith('.latin.typeface')]
            font_values=[x.get('value') for x in font_entries]
            fonts=sorted(set(x for x in font_values if x and x not in {'+mn-lt','+mj-lt','+mn-ea','+mj-ea'}))
            direct_sizes=len(sizes)==run_count and run_count>0
            uniform=(len(unique)==1 and direct_sizes and len(font_entries)==run_count and len(fonts)==1)
            is_table=any('a:tbl' in t.get('ancestors',[]) for t in obj.get('text',[]))
            style=gr.get('explicit_text_style') if gr else None
            body=(style.get('bodyPr') or [{}])[0] if style else {}
            guide_props={x.get('property','').split('.')[-1]:x.get('value') for x in typography if x.get('property','').startswith('style.bodyPr.')}
            body={**guide_props,**body}
            # Use observed source insets where both sides are known; otherwise do not fabricate them.
            insets={k:(float(body[k])/EMU_PER_PT if str(body.get(k,'')).isdigit() else None) for k in ('lIns','rIns','tIns','bIns')}
            all_insets=all(v is not None for v in insets.values())
            slot_text=' '.join(slots).lower(); object_name=(obj.get('name') or '')
            media_placeholder=bool(re.search(r'picture|photo|image|icon',object_name,re.I) or re.search(r'picture|photo|image|icon',slot_text))
            rotation=(gr.get('geometry_flags') or {}).get('rotation_degrees',0) if gr else 0
            group_container=bool(gr and gr.get('object_kind')=='grpSp')
            if uniform and bounds and gr and gr.get('geometry_status') in {'resolved','direct_source_frame'} and all_insets and not is_table and not media_placeholder and not rotation and not group_container:
                exp=unique[0]
                wpt=max(0,bounds['width']*72-insets['lIns']-insets['rIns'])
                hpt=max(0,bounds['height']*72-insets['tIns']-insets['bIns'])
                # Low/high estimates span glyph advances and leading, then preserve 25–45% headroom.
                low_lines=math.floor(hpt/(exp*1.35)); high_lines=math.floor(hpt/(exp*1.15))
                low_line_chars=math.floor(wpt/(exp*0.60)); high_line_chars=math.floor(wpt/(exp*0.45))
                low=max(0,math.floor(low_lines*low_line_chars*0.55)); high=max(low,math.floor(high_lines*high_line_chars*0.75))
                budget={'basis':'low-confidence geometric heuristic','characters':{'low':low,'high':high},
                        'assumptions':{'glyph_advance_em_range':[0.45,0.60],'line_height_em_range':[1.15,1.35],
                                       'headroom_factor_range':[0.55,0.75],'observed_insets_pt':insets},
                        'confidence':'low; text wrapping, paragraph spacing and rendering are not measured'}
                if low_lines<1 or high_lines<1: budget=None
            else:
                budget=None
            paragraph_count=len(obj.get('paragraphs',[]))
            zones.append({
                'object_id':oid,'object_name':obj.get('name',''),'slots':sorted(set(slots)),
                'observed_source_characters':len(chars),
                'source_text_run_count':run_count,'source_paragraph_count':paragraph_count,
                'frame_inches':bounds,'frame_points':({k:round(v*72,2) for k,v in bounds.items()} if bounds else None),
                'geometry_status':gr.get('geometry_status','unresolved') if gr else 'unresolved',
                'geometry_provenance':gr.get('transform_provenance') if gr else None,
                'explicit_font_sizes_pt':unique,'font_faces_observed':fonts,
                'typography_status':'uniform_explicit_run_size_candidate_low_confidence' if uniform else ('mixed_explicit_sizes' if len(unique)>1 else 'inherited_or_partial_unresolved'),
                'table_text_zone':is_table,
                'media_placeholder_zone':media_placeholder,
                'rotation_degrees':rotation,
                'insets_pt':insets if any(v is not None for v in insets.values()) else None,
                'authoring_budget_heuristic':budget,
                'source_text_exceeds_estimated_range':bool(budget and len(chars)>budget['characters']['high'])
            })
        # Pictures/icons are identified only from source object kind or explicit object naming.
        visuals=[]
        for oid,obj in objects.items():
            gr=attrs_for(obj,source_info)
            kind=gr.get('object_kind') if gr else None
            name=(obj.get('name') or '')
            if kind=='pic' or re.search(r'icon|logo|mark|illustration|image|picture|photo',name,re.I):
                b=gr.get('bounds_inches') if gr else None
                visuals.append({'object_id':oid,'object_name':name,'object_kind':kind or 'name_suggests_visual',
                                'semantic_identification':'name suggests icon/artwork' if re.search(r'icon|logo|mark',name,re.I) else ('source picture object' if kind=='pic' else 'visual name heuristic'),
                                'frame_inches':b,'frame_points':({k:round(v*72,2) for k,v in b.items()} if b else None),
                                'aspect_ratio':round(b['width']/b['height'],4) if b and b['height'] else None,
                                'geometry_status':gr.get('geometry_status','unresolved') if gr else 'unresolved',
                                'crop_or_fit':'not resolved by geometry export; source picture crop/fit not inferred'})
        # Semantic color controls are exact contract roles/profiles, separate from source artwork colors.
        color_roles={}
        for name,spec in contract.get('roles',{}).items():
            targets=spec.get('resource_targets',[])
            source_values=sorted(set(t.get('source_value') for t in targets if t.get('source_value')))
            color_roles[name]={'color_kind':spec.get('color_kind'),'binding_count':len(spec.get('binding_ids',[])),
                               'source_values':source_values,'description':spec.get('description'),
                               'target_resource_count':len(targets)}
        profiles=contract.get('profiles',{})
        row={'template_id':tid,'name':a.get('name'),'source':source,'source_slide':slide,'category':a.get('category'),
             'implementation_directory':a['implementation_directory'],'contract_path':str(contract_path.relative_to(ROOT)),
             'source_guide_path':str(guide_path.relative_to(ROOT)) if guide_path.exists() else None,
             'text_zone_count':len(zones),'observed_source_characters_total':sum(z['observed_source_characters'] for z in zones),
             'text_zones':zones,'semantic_color_roles':color_roles,'style_profiles':profiles,
             'semantic_coloring_status':'declared_roles_or_profiles' if color_roles or profiles else 'no_semantic_color_contract; retained source colors are not mapped to semantics',
             'visual_object_count':len(visuals),'visual_objects':visuals,
             'survey_limits':['Observed source character count is a description of current copy, not capacity.','Budget ranges are conservative planning heuristics only and do not replace PowerPoint fit measurement.','Inherited/mixed typography, unresolved nested transforms and source picture crop are flagged rather than estimated.']}
        result.append(row); categories[a.get('category')]+=1
    zone_sizes=collections.Counter()
    observed=0; budgets=0; resolved_visuals=0
    for r in result:
        for z in r['text_zones']:
            observed+=1
            if z['authoring_budget_heuristic']: budgets+=1
        for v in r['visual_objects']:
            if v['frame_inches']: resolved_visuals+=1
            if v['frame_inches'] and v['object_kind'] in {'pic','name_suggests_visual'}:
                w,h=v['frame_inches']['width'],v['frame_inches']['height']
                if w>0.03 and h>0.03: zone_sizes[(round(w,2),round(h,2))]+=1
    size_candidates=[{'width_in':w,'height_in':h,'count':n} for (w,h),n in zone_sizes.most_common(20) if n>=3]
    payload={'schema':'pptxgengo.component-sizing-survey.v1','survey_date':'2026-09-28',
             'source_assignment_count':len(result),'counts':{'text_zones':observed,'zones_with_conservative_budget_heuristic':budgets,
             'visual_objects':sum(x['visual_object_count'] for x in result),'visual_objects_with_resolved_bounds':resolved_visuals},
             'method':{'geometry':'Resolved transformed object bounds from scripts/catalog-geometry.py for registered source decks; binding-guide direct source frames only when top-level and explicitly present.',
             'text':'Binding IDs are joined to the source shape and source runs aggregated into one text zone per shape, avoiding run-level overcount.',
             'budget':'A low-confidence planning range is emitted only for non-table zones with uniform observed size/typeface candidates (run-count matching is not full font-inheritance resolution), geometry is resolved, and all four text insets are explicit. It spans 0.45–0.60em glyph advance and 1.15–1.35em leading with 25–45% headroom. It is not measured capacity.',
             'colors':'Semantic colors list only declared contract roles and style profiles; retained source artwork colors are not recast as semantic roles.',
             'visuals':'Picture object rectangles and names suggesting icon/logo/mark are inventoried; all other pictures remain unclassified artwork. Crop/fitting is not inferred.'},
             'canonical_visual_size_candidates':size_candidates,'templates':result,'missing_contracts':missing,
             'category_counts':dict(categories)}
    out=ROOT/'planning/component-survey/sizing.json'; write(out,payload)
    md=['# Component-family sizing survey','',
        f"Surveyed **{len(result)} of 101 assigned source templates**, covering {observed:,} mapped text zones and {payload['counts']['visual_objects']:,} picture/name-suggested visual objects.",'',
        '## Interpretation','',
        'Observed source character counts describe the existing copy only; they are not capacity limits. The low/high drafting ranges in JSON are low-confidence geometric hints, not fit guarantees. They appear only when the text has one explicit observed run size and face, a resolvable leaf frame, all four source insets, and is not table text, a media placeholder, a rotated zone, or an unresolved group. Text fit still needs to be checked in PowerPoint for changed copy.','',
        'The geometry helper resolves transformed frames for the four registered source decks. The 11 AI Accelerator scenes use direct binding-guide geometry where present; nested group transforms remain unresolved. Inherited theme/master typography is not inferred from unrelated text. Picture/crop transforms and semantic artwork meaning are not inferred. Pictures are not all called icons. Colors below refer to declared semantic contract roles/profiles, not an interpretation of every retained fill.','',
        'Heuristic ranges use a 0.45–0.60 em glyph-advance span, 1.15–1.35 em leading, actual explicit insets where available, and 25–45% headroom. There is no minimum-line assumption; too-short boxes receive no estimate. Table cells require per-cell layout analysis and are not budgeted as one large box.','',
        '## Coverage by assigned template','',
        '| Template | Source | Text zones | Existing characters | Zones with rough budget | Visual objects | Color roles / profiles |','|---|---|---:|---:|---:|---:|---:|']
    for r in result:
        count=sum(bool(z['authoring_budget_heuristic']) for z in r['text_zones'])
        md.append(f"| `{r['template_id']}` | {r['source']}:{r['source_slide']} | {r['text_zone_count']} | {r['observed_source_characters_total']} | {count} | {r['visual_object_count']} | {len(r['semantic_color_roles'])} / {len(r['style_profiles'])} |")
    md += ['', '## Recurring visual-object size candidates','',
           'Exact source-pixel dimensions are not semantic component standards. These rounded source-frame sizes occurred at least three times among pictures/name-suggested visuals and are useful only as starting candidates. Crop and fit are unresolved.','',
           '| Width (in) | Height (in) | Occurrences |','|---:|---:|---:|']
    for s in size_candidates: md.append(f"| {s['width_in']:.2f} | {s['height_in']:.2f} | {s['count']} |")
    md += ['', '## Representative text-zone examples','']
    for tid in ['t001-uhg-013','t045-graphics-and-layouts-045','t074-ai-accelerator-020']:
        r=next((x for x in result if x['template_id']==tid),None)
        if not r: continue
        md += [f"### {r['template_id']} — {r['name']}",'',f"Source `{r['source']}:{r['source_slide']}`. {r['text_zone_count']} contract-mapped shape zones, {r['observed_source_characters_total']} observed characters, {r['visual_object_count']} visual objects.",'',
               '| Object | Slot(s) | Frame (in) | Existing chars | Font evidence | Budget hint |','|---|---|---|---:|---|---|']
        for z in r['text_zones'][:5]:
            b=z['frame_inches']; frame=(f"{b['width']:.2f} × {b['height']:.2f}" if b else 'unresolved')
            budget=z['authoring_budget_heuristic']; budget=(f"{budget['characters']['low']}–{budget['characters']['high']} (rough)" if budget else 'none; inspect geometry/style')
            md.append(f"| {z['object_id']} · {z['object_name']} | {', '.join(z['slots'])} | {frame} | {z['observed_source_characters']} | {z['typography_status']} {z['explicit_font_sizes_pt']} pt | {budget} |")
        if r['source']=='graphics-and-layouts' and r['source_slide']==45:
            md.append('| Note | The source map is a native table; aggregate table text is deliberately not given a single text-capacity estimate. | | | | |')
        md.append('')
    md += ['## Files and scope','',
           '- `sizing.json` contains all 101 per-template text zones, source character counts, bounds in inches and points, explicit font evidence, semantic color roles/profiles, and image/icon-candidate bounds/aspect ratios.','- `scripts/survey-template-sizing.py` regenerates the report and calls the existing `scripts/catalog-geometry.py` for transformed geometry on registered source decks.','- No PowerPoint render, visual fit assertion, arbitrary-content qualification, image crop inference, or source artwork semantic classification is included.']
    (ROOT/'planning/component-survey/sizing.md').write_text('\n'.join(md)+'\n')
    print(json.dumps({'templates':len(result),'text_zones':observed,'heuristic_zones':budgets,'visual_objects':payload['counts']['visual_objects'],'visual_bounds':resolved_visuals,'missing_contracts':missing,'out':str(out)},indent=2))
    tmp.cleanup()

if __name__=='__main__': main()
