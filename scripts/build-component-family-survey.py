#!/usr/bin/env python3
"""Build a discussion crosswalk. This does not migrate executable contracts."""
from pathlib import Path
from collections import Counter
import json, hashlib, re
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'planning/component-survey'
# Family boundaries are curated by role/topology, not by color or keyword matching.
DEFS=[
('heading','Heading block','Eyebrow, title, optional section number','Hierarchy, title width and line count; divider pages are recipes'),
('textpanel','Text/list/callout panel','Heading, paragraphs or ordered bullet groups','Border, surface, emphasis and density; retain paragraph/list structure'),
('icon-text','Icon with content','Icon, label and supporting text','Icon position and optical size; keep icon-to-copy association'),
('numbereditem','Numbered item/card','Number, heading and description','Horizontal/vertical arrangements and badge placement; keep numbering semantic'),
('metric','Metric display','Value, unit, label and optional source','Single/multiple metric arrangement; preserve value/label/source association'),
('quote','Quotation','Quotation, attribution and optional portrait','Quote length and attribution position; source claims remain explicit'),
('personprofile','Person profile','Portrait, name, role and optional biography','Compact identity versus detailed biography variants'),
('roletile','Role tile','Role label, owner and optional allocation','Staffing/ownership colors and explicit legend; distinct from a person biography'),
('podcontainer','Pod container','Pod title and ordered role tiles','Role count, columns and spacing; parent owns inter-pod relationships'),
('table-row','Aligned table/evidence row','Named columns, row heading and cell content','Column widths, row count and row/cell styling; headers remain aligned'),
('processstep','Process step','Step label, activities, outputs and connection ports','Step count/order belong to parent flow; local slot hierarchy stays explicit'),
('architecturetile','Architecture node/tile','Capability/product/system label and optional detail','Node role, width and ports; do not infer arbitrary graph topology'),
('architecturelayer','Architecture layer','Layer label plus ordered capability tiles','Label and tiles share row geometry; repeatable layer owns its segmentation'),
('crosscutrail','Cross-cutting rail','Shared-control labels and declared scope','Separate from layer labels; divider and all-layer scope must remain clear'),
('timeaxis','Time axis','Ordered period labels and ticks','Period count/width; shared axis coordinates used by all lanes'),
('ganttlane','Gantt workstream lane','Workstream label, activity intervals and status','Label/bar vertical alignment, period endpoints and explicit status colors'),
('milestone','Milestone/annotation','Marker, date/period and caption','Anchor and caption placement; belongs to a timeline or event sequence'),
('rating-indicator','Gauge or discrete rating','Selected cell/range, pointer and scale/legend','Renderer geometry stays locked; do not substitute progress bars for source gauges'),
('legend','Legend/key item','Swatch/mark and meaning label','Actor/status/series meaning stays stable across presentation profiles'),
('mediaframe','Media frame','Image/artifact plus optional caption','Separate photo, portrait, icon, logo and thumbnail aspect/crop policies'),
('accent','Anchored design accent','Target phrase or source/target objects and approved artwork','Underline, marker and arrow have different placement rules; no standalone slide template'),
]
D={x[0]:dict(id=x[0],name=x[1],content_zones=x[2],variant_boundary=x[3]) for x in DEFS}
# Exact curated group-name crosswalk. Repeated group names share their documented role.
M={}
def add(names,tags,kind='component_candidate',recipe=None):
 for name in names.split(','):
  M[name]=dict(family_ids=tags.split(','),granularity=kind,recipe=recipe)
add('executive_message,lead_and_proof_heading,location_heading,company_intro','heading', 'composite_candidate')
add('divider_panel,section_panel','heading','recipe','section-divider')
add('understanding_panel','heading,icon-text,textpanel','recipe','icon-content-panel')
add('partner_takeaway,closing_alignment,value_strip,readiness_statement,center_value,summary,goal_band,outcome_strip','textpanel')
add('comparison_panels','textpanel,table-row','recipe','comparison')
add('proof_cards','icon-text','repeatable_composition')
add('case_study_one,case_study_two','heading,textpanel,metric','recipe','case-study')
add('relationship_metrics,impact_metrics,interest_metrics','metric','repeatable_composition')
add('client_quotes,testimonial','quote','repeatable_composition')
add('five_capability_columns,operating_drivers','numbereditem,textpanel','repeatable_composition')
add('risk_band','textpanel','repeatable_composition')
add('experience_groups,service_categories,governance_domains,network_and_assurance,placement_criteria,leadership_responsibility_panel,services_and_industries,presence,awards_and_sources,workshop_application,facilitation_sidebar,participant_expectations','textpanel','composite_candidate')
add('architecture_inset,highlighted_architecture_map','architecturetile,architecturelayer,crosscutrail','recipe','architecture-map')
add('platform_thumbnail','mediaframe')
add('six_readiness_gates','textpanel','repeatable_composition')
add('phase_1,phase_2,phase_3','processstep','component_candidate')
add('phase_rail','heading,textpanel','recipe','phase-detail')
add('workstream_matrix,ea_selection_scorecard,role_by_phase_matrix','table-row','recipe','table-matrix')
add('right_support_panels','mediaframe,textpanel','recipe','phase-detail')
add('three_pattern_rows','table-row,architecturetile,textpanel','recipe','architecture-comparison')
add('phase_indicator','heading,timeaxis','composite_candidate')
add('crosscut_controls','crosscutrail')
add('source_systems,consumer_platforms','architecturetile','repeatable_composition')
add('enterprise_platform_layers','architecturelayer,architecturetile','recipe','architecture-map')
add('three_early_adopter_cards,early_adopter_cards','heading,textpanel','repeatable_composition')
add('six_accelerator_cards','heading,textpanel,mediaframe','recipe','artifact-gallery')
add('eight_benefit_spokes','icon-text,textpanel','recipe','radial-benefits')
add('six_workflow_stages,five_stage_flow,pathway_one,pathway_two,three_step_loop','processstep','recipe','process-flow')
add('shared_services_ribbon,variable_delivery_footer','textpanel','repeatable_composition')
add('five_exit_gate_columns','numbereditem,textpanel','repeatable_composition')
add('enablement_narrative','heading,textpanel,mediaframe','recipe','evidence-narrative')
add('governance_callout','heading,textpanel','composite_candidate')
add('two_party_timeline','timeaxis,processstep,textpanel','recipe','timeline')
add('transition_annotation','textpanel,accent','composite_candidate')
add('timeline_scale','timeaxis')
add('workstream_bars','ganttlane','repeatable_composition')
add('milestones','milestone','repeatable_composition')
add('fixed_scope,variable_scope','textpanel','recipe','scope-options')
add('optional_service_annotation','textpanel,accent','composite_candidate')
add('three_service_rows','table-row','repeatable_composition')
add('three_location_cards','textpanel,mediaframe','repeatable_composition')
add('hierarchy_chart','podcontainer,roletile','recipe','organization-chart')
add('staffing_legend','legend','repeatable_composition')
add('core_roster,specialist_roster','personprofile','recipe','team-roster')
add('responsibility_groups','textpanel','repeatable_composition')
add('outcome_panel','numbereditem,textpanel','repeatable_composition')
add('feedback_loop','accent')
add('feedback_explanations','textpanel','repeatable_composition')
add('shared_team_band','roletile','repeatable_composition')
add('column_headers','table-row','composite_candidate')
add('between_step_arrows','accent','repeatable_composition')
H={
 'callout':['textpanel'],'case-study-card':['textpanel','metric'],'comparison-card':['textpanel'],'comparison-row':['table-row'],
 'diagram-node':['architecturetile'],'gantt-row':['ganttlane'],'headline-emphasis':['accent'],'hierarchy-tier':['roletile','podcontainer'],
 'icon-panel':['icon-text'],'image-caption':['mediaframe'],'image-comparison':['mediaframe'],'legend-item':['legend'],'list-panel':['textpanel'],
 'map-label':['textpanel'],'metric-panel':['metric'],'numbered-card':['numbereditem'],'numbered-item':['numbereditem'],
 'person-card':['personprofile'],'phase-column':['processstep','textpanel'],'pricing-option':['textpanel','metric'],'process-ribbon':['processstep'],
 'process-step':['processstep'],'prompt-panel':['textpanel'],'quote':['quote'],'rating-scale':['rating-indicator'],'section-header':['heading'],
 'service-panel':['textpanel'],'status-band':['textpanel'],'team-pod':['podcontainer','roletile'],'text-card':['textpanel'],'timeline-event':['milestone'],
}
EXTRA={'architecturelayer':['t020-graphics-and-layouts-062'],'crosscutrail':['t060-uhg-026'],'rating-indicator':['t045-graphics-and-layouts-045'],'ganttlane':['t015-graphics-and-layouts-083'],'timeaxis':['t015-graphics-and-layouts-083']}
def main():
 inputs=['library/component-families.jsonl','library/component-seeds-expanded.json','library/templates/requested-components.json','library/templates/rollout/assignments.json']
 old=[v for x in (ROOT/inputs[0]).read_text().splitlines() if (v:=json.loads(x)).get('kind')=='component_family'];new=json.loads((ROOT/inputs[2]).read_text())['components']
 rows=[]
 for c in new:
  name=c['id'].split(':',1)[1]
  if name in M:m=dict(M[name])
  elif re.fullmatch(r'(profile|facilitator)_\d+',name):m=dict(family_ids=['personprofile'],granularity='component_candidate',recipe=None)
  elif re.fullmatch(r'(stage|step)_\d+',name):m=dict(family_ids=['processstep'],granularity='component_candidate',recipe=None)
  elif re.fullmatch(r'(framework|agenda)_row_\d+',name):m=dict(family_ids=['table-row'],granularity='component_candidate',recipe='workshop')
  else:raise ValueError('Unreviewed group '+c['id'])
  rows.append(dict(source_id=c['id'],template_id=c['template_id'],source=c['source'],purpose=c['purpose'],object_ids=c['object_ids'],slots=c['slots'],executable=bool(c.get('contract')),confidence='inventory_semantics_proposal_not_geometry_equivalence',**m))
 historical=[]
 for f in old:
  key=f['id'].split(':')[-1];historical.append(dict(historical_family_id=f['id'],name=f['name'],family_ids=H[key],historical_compositions=len(f['instances']),migration='Retain retrieval label and source variants; shared implementation needs geometric qualification.'))
 families=[]
 for ident,base in D.items():
  r=[x for x in rows if ident in x['family_ids']];h=[x for x in historical if ident in x['family_ids']]
  families.append(dict(base,requested_memberships=len(r),historical_memberships=sum(x['historical_compositions'] for x in h),representative_template_ids=list(dict.fromkeys(EXTRA.get(ident,[])+[x['template_id'] for x in r]))[:6],status='proposed_reuse_boundary'))
 recipes=sorted({x['recipe'] for x in rows if x['recipe']})
 payload=dict(schema='pptxgengo.component-family-survey.v2',status='discussion_proposal_no_contract_migration',sources=[dict(path=p,sha256=hashlib.sha256((ROOT/p).read_bytes()).hexdigest()) for p in inputs],counts=dict(proposed_building_block_families=len(families),historical_retrieval_families=len(old),historical_compositions=244,requested_occurrences=len(rows),requested_executable=sum(x['executable'] for x in rows),requested_visual_only=sum(not x['executable'] for x in rows),counting_note='244 and132 overlap; memberships are multi-label and cannot be summed into unique components.'),families=families,recipes=recipes,historical_mapping=historical,requested_mapping=rows)
 OUT.mkdir(parents=True,exist_ok=True);(OUT/'families.json').write_text(json.dumps(payload,indent=2)+'\n')
 text=['# Component families: consolidation proposal','',f'Propose **{len(families)} building-block families**, preserving the **31 historical retrieval labels** and source designs. Larger arrangements remain composition recipes. These boundaries are inferred from curated source purpose, slots and current authoring behavior; they do not establish geometric interchangeability or portable implementation readiness.','', 'The 244 historical compositions and 132 newer occurrences overlap and use different boundaries. Do not add their counts. A source group can contain several families; table counts below are memberships.','', '| Family | Content zones | Variant / consolidation boundary | Requested memberships | Examples |','|---|---|---|---:|---|']
 for f in families:text.append('| '+f['name']+' | '+f['content_zones']+' | '+f['variant_boundary']+' | '+str(f['requested_memberships'])+' | '+', '.join(f['representative_template_ids'][:2])+' |')
 text+=['','## Recipes and boundaries','',', '.join(recipes)+'.','', 'Keep person profiles, role tiles and pod containers distinct. Keep architecture tiles, layer rows and cross-cutting rails distinct. A takeaway band is a text panel with emphasis styling; it is not a hand-drawn accent. A governance callout is not a gauge. Gauges retain their selected renderer and separate selection/pointer semantics.','', 'Dividers use a heading and a slide-level background arrangement. Comparison/pricing/case-study/phase-detail pages assemble shared components while retaining their narrative and topology. A content-panel classification alone does not prove that a complex source group can be transplanted or resized.','', '## Exact crosswalk','', 'Every historical family and all 132 requested occurrences are addressable in [families.json](families.json), with source/template IDs, object IDs, slot names, family memberships, composition granularity and executable/reference status. Source aliases should survive any eventual consolidation.','', '## What consolidation changes','', 'Unify content schemas, semantic style roles, sizing metadata and compatible layout behavior. Retain separate source geometry variants when font hierarchy, slot structure, reading order, dependency topology or asset ownership differs. Color differences become style options only when those structural conditions match. No library contracts are migrated by this survey.','']
 (OUT/'families.md').write_text('\n'.join(text))
 print('Survey:',len(families),'proposed families;',len(rows),'mapped occurrences;',len(old),'historical labels')
if __name__=='__main__':main()
