#!/usr/bin/env python3
"""Compile fixed-source evidence and conservative semantic family recommendations."""
from __future__ import annotations
import argparse, json, re
from pathlib import Path

SCHEMA = 'pptxgengo.adaptive-capability-catalog.v1'
FAMILIES = {
 'roadmap': {
  'source_categories':['roadmap'], 'schema_path':'internal/adapt/roadmap.go',
  'proposed_controls':['periods (3–12, ordered ID/label pairs)','workstreams (1–8; label, group, status)','intervals (period-ID range, non-overlapping)','milestones (period ID, label, optional caption_periods 1–3; unique period per workstream)'],
  'input_contract':{'required':'periods[], workstreams[]','status_values':['planned','active','complete','at_risk','blocked'],'bounds_points':{'minimum_width':480,'minimum_height':230},'period_cell_min_width':40,'row_min_height':35,'limits':['does not infer dates','does not infer dependencies']},
 },
 'architecture': {
  'source_categories':['business_architecture','technical_architecture','layers_components'], 'schema_path':'internal/adapt/architecture.go',
  'proposed_controls':['layers (2–8; ordered heading/component rows)','components (1–8 per layer; required unique ID and label, optional detail, fill, width weight)','left_rail (0–5 cross-cutting items)','relations (explicit known component ID pairs; reporting, dependency, advisory, or annotation)'],
  'input_contract':{'required':'layers[]','component_fill_values':['accent','muted','surface','active'],'width_weight_range':[0,4],'limits':['horizontal components within each layer','vertical layer ordering','no inferred arbitrary graph topology','no nested diagrams']},
 },
 'process': {
  'source_categories':['approach_delivery'], 'schema_path':'internal/adapt/process.go',
  'proposed_controls':['steps (2–6; label, summary, activities, outputs, optional state)','activities (1–8 per step)','outputs (1–5 per step)','gap_pt (12–36; default 18)'],
  'input_contract':{'required':'steps[]','state_values':['planned','active','complete'],'minimum_step_width_points':115,'bounds_points':{'default':[48,125,864,360],'permitted_x':[40,920],'permitted_y':[115,485]},'limits':['horizontal sequence of panels','does not restructure source OOXML']},
 },
 'team': {
  'source_categories':['team_credentials'], 'schema_path':'internal/adapt/team.go',
  'proposed_controls':['pods (1–6, each with 1–8 roles)','roles (ID, label, optional staffing semantic token)','reporting_lines (0–12; pod relationships using reporting, dependency, advisory, or annotation)','responsibilities matrix (optional; right/default position allows 2–4 columns; bottom allows 2–6; 1–6 rows)'],
  'input_contract':{'required':'pods[]','staffing_values':['staffing.wm_full_time','staffing.wm_part_time','staffing.client_part_time'],'limits':['reporting lines connect distinct pod IDs and use a supported relationship type','responsibility rows supply one value per column','matrix_position is right (default) or bottom and requires a matrix','no independent role subfields','no free-form matrix topology']},
 },
 'comparison': {
  'source_categories':['comparison_evidence'], 'schema_path':'internal/adapt/comparison.go',
  'proposed_controls':['gauge_mode (linear or dial; default linear)','options (2–4 ID/label pairs)','criteria (1–7; label, gauge, one finding per option)','gauge (caller supplied value, scale min/max, target min/max, optional unit)','finding (option ID, text, status)'],
  'input_contract':{'required':'options[], criteria[]','gauge_mode_values':['linear','dial'],'gauge_mode_default':'linear','status_values':['strong','mixed','limited','unknown'],'bounds_points':{'minimum_width':500,'minimum_height':230},'option_cell_min_width':115,'row_height_min_points':45,'dial_row_height_min_points':58,'limits':['gauge uses caller-supplied numeric evidence','does not invent or score evidence']},
 },
}

UNMAPPED = 'No taxonomy hint to one of the five in-progress semantic family builders; retain this item in fixed-source inventory. Lack of a hint does not rule out future redesign.'

def load(p): return json.loads(Path(p).read_text(encoding='utf-8'))
def rel_attrs(node):
    out={}
    for a in node.get('attributes',[]): out[a['name']]=a['value']
    return out
def walk(node):
    if isinstance(node,dict):
        yield node
        for child in node.get('children',[]): yield from walk(child)
def image_relationships(scene):
    rels=scene.get('relationships',{})
    picture_meta={}
    for node in walk(scene.get('scene',{})):
        if node.get('name') != 'p:pic': continue
        name=descr=''; rid=None
        for child in walk(node):
            if child.get('name')=='p:cNvPr':
                a=rel_attrs(child); name=a.get('name',''); descr=a.get('descr','')
            if child.get('name')=='a:blip':
                a=rel_attrs(child); rid=a.get('r:embed')
        if rid: picture_meta[rid]={'object_name':name,'source_alt_text':descr or None}

    rows=[]
    for item in rels.get('children',[]):
        a=rel_attrs(item)
        if a.get('Type','').endswith('/image'):
            rows.append({'relationship_id':a.get('Id'),'package_target':a.get('Target'),'source_picture_metadata':picture_meta.get(a.get('Id'),{}),'description_status':'source alt text preserved when present; otherwise no visual description inferred'})
    return rows

def family_choice(entry):
    cat=entry['category']
    for name, spec in FAMILIES.items():
        if cat in spec['source_categories']:
            return {'status':'category_candidate_requires_review','family':name,'rationale':f"Source taxonomy category `{cat}` is a discovery hint for the `{name}` family only. Structural suitability is unreviewed; inspect the source before choosing an adaptive builder.",'semantic_only':True,'source_fidelity_claimed':False,'requires_explicit_redesign_intent':True,'proposed_controls':spec['proposed_controls']}
    return {'status':'unmapped','family':None,'rationale':UNMAPPED,'semantic_only':True,'source_fidelity_claimed':False,'proposed_controls':[]}

def reviewed_family_examples(checkpoint, root):
    """Return only exact examples the checkpoint explicitly marks visually reviewed."""
    out = {}
    for family, record in checkpoint.get('families', {}).items():
        examples=[]
        for example in record.get('examples', []):
            if example.get('status') != 'visually_reviewed':
                continue
            examples.append({k:example[k] for k in ('id','family','variant','source_template_id','slide_index','status','controls','limitations','evidence','hashes','render_png','spec_path','values_path') if k in example})
        if examples:
            out[family] = examples
    return out

def build(root: Path, checkpoint_path: Path | None = None):
    assignments=load(root/'library/templates/rollout/assignments.json')
    checkpoint=load(root/'library/templates/rollout/checkpoint.json')
    adaptive_checkpoint = load(checkpoint_path) if checkpoint_path and checkpoint_path.exists() else {}
    reviewed = reviewed_family_examples(adaptive_checkpoint, root)
    rows=[]
    for e in assignments['entries']:
        template_id=e['template_id']; d=root/e['implementation_directory']
        contract=load(d/'contract.json'); impl=load(d/'implementation.json'); vals=load(d/'example-values.json')
        check=next((x for x in checkpoint['templates'] if x['template_id']==template_id),{})
        source_scene_path=root/e['source_scene']
        source_scene=load(source_scene_path) if source_scene_path.exists() else {}
        slots=[]
        for name, slot in contract.get('slots',{}).items():
            bindings=slot.get('binding_ids',[])
            sample=(vals.get('slots',{}).get(name))
            slots.append({'name':name,'description':slot.get('description',''),'binding_count':len(bindings),'binding_ids':bindings,'example_value_count':len(sample) if isinstance(sample,list) else (1 if sample is not None else 0),'value_format':slot.get('value_format','segments'),'source_bound':True})
        # preserve contract roles/profiles exactly: these are declared source-bound controls.
        rows.append({
          'template_id':template_id,'family_id':e['id'],'name':e['name'],'source':e['representative'],
          'lane':e['lane'],'category':e['category'],'altitudes':e.get('altitudes',[]),'densities':e.get('densities',[]),
          'source_identity':{'package_sha256':contract.get('source_sha256'),'scene_sha256':contract.get('scene_sha256'),'source_slide':contract.get('slide'),'source_project':e['source_project'],'source_scene':e['source_scene'],'preview':impl.get('preview')},
          'semantic_summary':{'role':impl.get('role',e['name']),'takeaway':impl.get('takeaway',''),'visibility':'role and takeaway are placed in speaker notes; they are not visible slide text'},
          'fixed_source':{'adaptation_mode':'fixed_source_geometry','qualification':'source-bound editing contract; not arbitrary-content qualified','source_bound_slots':slots,'source_bound_color_roles':contract.get('roles',{}),'source_bound_style_profiles':contract.get('profiles',{}),'text_zones':contract.get('zones',{}),'contract_constraints':contract.get('constraints',[]),'fixed_areas':impl.get('opaque_areas',[]),'retained_source_content':impl.get('retained_source_content',[]),'fixed_arrangement_and_item_count':True},
          'preferred_authoring_route':'template_preserve_source',
          'source_gauge_controls':({'command':'template apply-gauge','cells':5,'rows':5,'highlight_and_pointer_independent':True,'pointer_default':'sole highlighted cell','colors':['gray','navy','blue','pink'],'reference':'library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/gauge-authoring.md'} if template_id=='t045-graphics-and-layouts-045' else None),
          'adaptive_family_recommendation':family_choice(e),
          'adaptive_family_qualification':{'state':'pending_native_review','implementation_state':'semantic family compiler implemented; native qualification pending','source_fidelity_claimed':False,'changed_content_capacity_qualified':False,'native_fit':'pending','visual_review':'pending','pptx_open_repair':'pending'},
          'current_bounded_engine_capabilities':['native editable text and shapes','explicit canvas text, surfaces, lines, images and bounded shape presets','measured numbered and metric cards','measured variable-role pods and bounded team/reporting compositions','measured named grids and panels for specific accepted patterns','PNG/JPEG images with pinned hashes','bounded phrase accents using measured native character bounds'],
          'engine_limits':['Family mapping is a semantic recommendation and does not import or preserve this source slide geometry.','template_id, when supplied to the adaptive adapter, is provenance/reference metadata only; it is echoed in the report and does not select, import, or preserve the source template.','No adaptive family support for arbitrary source reconstruction, unrestricted layout inference, open-ended item counts, arbitrary changed-content capacity, or automatic qualification is implied; use only cardinalities explicitly supported by the selected compose component.','Family-specific builder controls remain unqualified until native fit, changed-content, and visual review evidence is recorded.'],
          'source_assets':image_relationships(source_scene),
          'asset_description_metadata':{'status':'linked_where_present; no description inferred from a filename','source_scene':e['source_scene'],'retained_content_notes':impl.get('retained_source_content',[]),'catalog_reference':'library/showcase/assets.json','hosted_asset_inventory_note':'The WM hosted asset inventory is filename/path metadata, not visual descriptions; search with the packaged find_asset.py helper and inspect selected assets before assigning a semantic role.'},
          'fixed_source_example_review':{'status':check.get('status','unknown'),'render_method':check.get('render_method','direct_native_open_export'),'current_inputs_match_render':check.get('current_inputs_match_render'),'findings':check.get('findings',[]),'limitations':check.get('limitations',[])},
        })
    byfamily={f:sum(1 for r in rows if r['adaptive_family_recommendation']['family']==f) for f in FAMILIES}
    category_candidate_count=sum(1 for r in rows if r['adaptive_family_recommendation']['family'])
    reviewed_count=sum(len(v) for v in reviewed.values())
    return {'schema':SCHEMA,'version':1,'source_date':checkpoint.get('date'),'purpose':'Machine-readable capability inventory separating existing fixed-source editing contracts from proposed semantic family adaptation.','qualification_summary':{'fixed_source_contract_count':len(rows),'fixed_source_examples_visually_reviewed':checkpoint.get('review_counts',{}).get('example_visually_reviewed',0),'direct_native_open_export_examples':checkpoint.get('current_example_direct_native_open_exports'),'native_geometry_replay_examples':checkpoint.get('current_example_native_geometry_replays'),'generated_accent_pptx_reopen_pending':checkpoint.get('generated_accent_pptx_reopen_pending'),'arbitrary_content_qualified':checkpoint.get('arbitrary_content_qualified',0),'adaptive_family_native_qualified':0,'families_with_reviewed_examples':len(reviewed),'reviewed_adaptive_example_count':reviewed_count,'adaptive_checkpoint_scope':adaptive_checkpoint.get('scope'),'category_candidate_count':category_candidate_count,'unmapped_category_count':len(rows)-category_candidate_count},'family_builders':{name:{**spec,'state':'implemented_pending_native_review','native_qualification':'pending','source_fidelity':'not_claimed','category_candidate_contract_count':byfamily[name],'bounded_examples_visually_reviewed':bool(reviewed.get(name)),'reviewed_adaptive_example_count':len(reviewed.get(name,[])),'evidence_reference':'library/adaptive/checkpoint.json' if reviewed.get(name) else None,'reviewed_examples':reviewed.get(name,[])} for name,spec in FAMILIES.items()},'unmapped_contract_count':len(rows)-category_candidate_count,'adaptive_style_controls':{'profiles':['neutral','subtle','inverse'],'font_face':{'type':'string','default':'Arial','policy':'Explicit installed font family; requested styles must resolve without substitution during native measurement','examples':['Arial','IBM Plex Sans','IBM Plex Mono']},'title_font_pt':[18,36],'body_font_pt':[9,18],'label_font_pt':[9,18],'semantic_color_roles':['text.primary','text.secondary','surface','surface.muted','accent','state.active','state.inactive','border'],'contrast_rule':'text.primary must meet 4.5:1 against surface'},'current_bounded_engine_capabilities':rows[0]['current_bounded_engine_capabilities'] if rows else [],'asset_catalogs':{'wm_hosted_inventory':{'records':1331,'fields':['path','directory','filename','extension','size_bytes','modified_at'],'visual_description_field_present':False,'search_helper':'wm-brand-assets/scripts/find_asset.py','selection_rule':'Search by real catalog metadata and directory/filename; inspect selected visuals. Do not invent subject tags or descriptions.'},'packaged_showcase_assets':'library/showcase/assets.json; selected local assets include explicit source paths, SHA-256, descriptions/source descriptions and usage restrictions.'},'items':rows}

def main():
    ap=argparse.ArgumentParser(); ap.add_argument('--root',default=str(Path(__file__).resolve().parents[1])); ap.add_argument('--out',default='library/adaptive/capabilities.json'); ap.add_argument('--checkpoint',default='library/adaptive/checkpoint.json',help='Optional adaptive example-review checkpoint'); ap.add_argument('--no-checkpoint',action='store_true',help='Ignore the adaptive review checkpoint'); args=ap.parse_args(); root=Path(args.root).resolve(); out=Path(args.out); out=out if out.is_absolute() else root/out
    cp=None if args.no_checkpoint else Path(args.checkpoint); cp=(root/cp if cp and not cp.is_absolute() else cp)
    data=build(root,cp); out.parent.mkdir(parents=True,exist_ok=True); out.write_text(json.dumps(data,indent=2,ensure_ascii=False)+'\n',encoding='utf-8'); print(f"wrote {len(data['items'])} capability records to {out}")
if __name__=='__main__': main()
