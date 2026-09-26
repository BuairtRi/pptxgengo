#!/usr/bin/env python3
"""Record the local, native-verified dense showcase fixture; requires ignored QA artifacts."""
from pathlib import Path
import json, hashlib, collections, re, zipfile, xml.etree.ElementTree as E
R=Path(__file__).resolve().parents[1]; B=R/'samples/component-adaptation/dynamic-pods'; NAME='proposal-complexity-v2'
def read(p):return json.loads(Path(p).read_text())
def record(p):
 p=Path(p);b=p.read_bytes();return dict(path=str(p.relative_to(R)),sha256=hashlib.sha256(b).hexdigest(),size_bytes=len(b))
def write(p,x):Path(p).write_text(json.dumps(x,indent=2)+'\n')
specp=R/'library/showcase/dense-deck.json';spec=read(specp);m=read(B/NAME/'manifest.json');v=read(B/(NAME+'-verification.json'))
assert v['spec_sha256']==m['spec_sha256']==record(specp)['sha256']
assert v['deck_sha256']==m['deck_sha256']==record(B/NAME/(NAME+'.pptx'))['sha256']
review=read(R/'samples/showcase/dense-visual-review.json')
assert len(review)==len(spec['slides']) and all(x['reviewed'] and x['accepted'] for x in review)
assert [x['id'] for x in review]==[x['id'] for x in spec['slides']]
rows={(r['slide_index'],r['shape_name']):r for r in v['native']['measurements']};maxdelta=overflow=0
for i,s in enumerate(m['slides'],1):
 for e in s['elements']:
  r=rows[i,e['name']];f=e['frame'];rf=r['shape_frame']
  maxdelta=max(maxdelta,*[abs(f[k]-rf[n]) for k,n in [('x','left'),('y','top'),('width','width'),('height','height')]])
  if e['kind']=='text':
   b=r['text_bounds'];ix=e['inset_x'];iy=e['inset_y']
   overflow=max(overflow,f['x']+ix-b['left'],f['y']+iy-b['top'],b['left']+b['width']-f['x']-f['width']+ix,b['top']+b['height']-f['y']-f['height']+iy)
assert maxdelta <= .12 and overflow <= .15, 'Final native frame or text containment failure'
fit=read(R/'samples/showcase/dense-fit-final-v2.json');assert fit['overflow_count']==0 and fit['planner_passed']
roundtrip=read(R/'samples/showcase/dense-roundtrip-audit.json');assert all(s['shape_tree_xml_exact_after_normalization'] for s in roundtrip['slides'])
recovery=read(R/'samples/showcase/recovered-dense-native-edit/recovery.json');assert len(recovery['changes'])==1
ns={'a':'http://schemas.openxmlformats.org/drawingml/2006/main','p':'http://schemas.openxmlformats.org/presentationml/2006/main'}
def xmlmetrics(file,number):
 with zipfile.ZipFile(file) as z:r=E.fromstring(z.read(f'ppt/slides/slide{number}.xml'))
 texts=[t.text or '' for t in r.findall('.//a:t',ns)]
 return dict(word_count=len(re.findall(r'\S+',' '.join(texts))),direct_shape_nodes=len(r.findall('.//p:sp',ns)),picture_nodes=len(r.findall('.//p:pic',ns)),group_nodes=len(r.findall('.//p:grpSp',ns)),table_nodes=len(r.findall('.//a:tbl',ns)))
refs=[('enablecomp',3),('uhg',28),('enablecomp',4),('uhg',43),('uhg',38)]
files={'uhg':R/'samples/UHG Fabric Platforming RFP Response - July 2026.pptx','enablecomp':R/'samples/EnableComp_RFP Response_DRAFT_081126.pptx'}
comparison=[dict(slide_id=s['id'],reference=f'{source}:{number}',source=xmlmetrics(files[source],number),generated=xmlmetrics(B/NAME/(NAME+'.pptx'),i+1)) for i,(s,(source,number)) in enumerate(zip(spec['slides'],refs))]
report=dict(schema='pptxgengo.showcase-proof.v1',date='2026-09-26',status='native_fixture_proof_not_general_catalog_approval',slides=len(m['slides']),final_native_shapes=len(rows),shape_kinds=dict(collections.Counter(e['kind'] for s in m['slides'] for e in s['elements'])),native_measurement_schema=v['native']['schema'],max_frame_delta_pt=maxdelta,max_text_overflow_pt=overflow,geometry_tolerance_pt=.12,fit_tolerance_pt=.15,visual_review=review,native_checks=['exact text and stable object coverage','actual frames and rotation','uniform character font name/size/weight/color','non-whitespace character-bound containment','text margins','fill RGB and opacity','stroke RGB/weight/opacity/arrowhead absence','image frames; verified embedded asset hashes'],fixed_text_zone_count=fit['fixed_zone_count'],fixed_text_overflow_count=0,initial_overflows=read(R/'samples/showcase/dense-fit-initial.json')['overflow_count'],reference_comparison=comparison,comparison_scope='Counts use direct slide OOXML including table text, excluding inherited text and embedded-picture text. They describe density, not quality or fidelity.',scene_roundtrip=dict(slides=len(roundtrip['slides']),normalized_shape_trees_match=True),text_recovery=recovery,compilation='go build ./... passed',unit_suite='not run',limitations=['New synthetic compositions inspired by references; not exact source rebuilds','Explicit geometry; no reusable measured grid/panel API yet','Native editable text/shapes; comparison/roadmap have no native table/chart data model','Source-shaped ribbons, deliverable montages, hatching, nested image-rich architecture and rich-text emphasis remain separate fidelity gates','Team counterparts are authored rules, reporting routes are editable segments; no glued connector semantics','No broad catalog adaptation approval; revised copy needs measurement and visual review','Known-generated-deck text recovery restores original styling; returned formatting/notes/layout changes are not imported'],assistance='Two GPT-6-Luna agents audited reference complexity, code/padding gaps, narrative consistency and preview visuals; root authored, integrated, measured and reviewed the dense sample.',code=[record(R/p) for p in ['cmd/pptxcompose/main.go','cmd/pptxcompose/render.go','cmd/pptxcompose/canvas_render.go','cmd/pptxcompose/recover.go','internal/compose/planner.go','internal/compose/canvas.go','internal/compose/cards.go','internal/compose/accents.go','internal/compose/fit_report.go','internal/compose/routes.go','scripts/measure-compose-text.applescript','scripts/build-dense-showcase-spec.py','scripts/catalog-index.py','scripts/report-dense-showcase.py']],inputs=[record(specp),record(R/'library/showcase/assets.json')],artifacts=[record(p) for p in [B/'dense-evidence-v1.json',B/'dense-evidence-v2.json',B/NAME/'manifest.json',B/NAME/(NAME+'.pptx'),B/(NAME+'-verification.json'),B/(NAME+'.pdf'),R/'samples/showcase/dense-fit-initial.json',R/'samples/showcase/dense-fit-final-v2.json',R/'samples/showcase/dense-roundtrip-audit.json',R/'samples/showcase/recovered-dense-native-edit/recovery.json',R/'samples/showcase/dense-visual-review.json',R/'samples/showcase/two-overflows-report.json']]+[record(B/(NAME+'-render')/f'slide-{i:03}.png') for i in range(1,6)])
checks=['samples/showcase/recovery-negative-fixtures/results.json','samples/showcase/fit-contract-results.json','samples/showcase/cli-negative-results-v3.json','samples/showcase/planner-negative-results.json']
report['guard_checks']=[]
for path in checks:
 result=read(R/path)
 assert isinstance(result,list) and all(row['passed'] for row in result), path
 report['guard_checks'].append(dict(artifact=record(R/path),cases=result))
proofp=R/'library/showcase/dense-proof.json';write(proofp,report)
recipes=[dict(id='dynamic-recipe:dense:'+s['id'],kind='dynamic_recipe',name=s['title'] or 'Build & prove the first slice',readiness='native_fixture_verified',adaptation_approved=False,role=s['role'],takeaway=s['takeaway'],spec_path=str(specp.relative_to(R)),spec_sha256=record(specp)['sha256'],slide_id=s['id'],proof_path=str(proofp.relative_to(R)),proof_sha256=record(proofp)['sha256'],requires='New native measurements for changed copy/width/style; final native verification and visual review') for s in spec['slides']]
(R/'library/showcase/recipes.jsonl').write_text(''.join(json.dumps(r)+'\n' for r in recipes))
print(json.dumps({k:report[k] for k in ['slides','final_native_shapes','shape_kinds','max_frame_delta_pt','max_text_overflow_pt','reference_comparison']},indent=2))
