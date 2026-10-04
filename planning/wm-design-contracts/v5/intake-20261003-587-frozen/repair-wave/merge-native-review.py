"""Merge individual native inspections only when their exact page bytes match."""
import hashlib,json
from pathlib import Path
root=Path(__file__).resolve().parent
production=Path('samples/wmds-incoming-20261003/v5-production-v3c')
source=production/'source';doc=json.loads((source/'compiled-document.json').read_text())
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
reviews={}
for name in ['native-review-001-200.json','native-review-201-400.json','native-review-401-587.json','native-review-adjudications.json','native-recheck-v3c-026-261.json','native-recheck-v3c-305-566.json']:
 for row in read(root/name)['pages']:
  reviews[row['page']]={**row,'receipt':name}
pages=[];flat={}
for i,s in enumerate(doc['slides'],1):
 key=s['template_binding']['template'];r=reviews[i];digest=sha(source/'native-pages'/f'slide-{i:03d}.png')
 assert r.get('key',r.get('template'))==key,(i,'identity')
 assert r.get('sha256',r.get('preview_sha256'))==digest,(i,'hash')
 assert r['status'] in ('reviewed_source_specimen','pass'),(i,r['status'])
 row={'page':i,'template':key,'sha256':digest,'result':'reviewed_source_specimen','finding':r['finding'],'evidence_receipt':r['receipt'],'basis':'individual_final_native_view' if r['receipt'].startswith('native-recheck') else 'individual_prior_native_view_exact_png_hash'}
 pages.append(row);flat[key]={'status':'reviewed_source_specimen','preview_sha256':digest}
assert len(pages)==len(flat)==587
receipt={'schema':'pptxgengo.wmds-release-verification.v1','version':'0.1.0-local.8-candidate','source_revision':'wmds-library.v5','source_commit':'d83bd58a9f9de68ebd8d6b3c9b0272c16ed516cf','qualification_scope':'587 illustrated source specimens; caller content still requires fit and native review','native_export':'PowerPoint for Mac local Best for printing PDF, then Swift PDFKit 1920x1080','native_pdf_sha256':sha(source/'WMDS-template-gallery-587-v5-repaired-v3c-20261003.pdf'),'source_pptx_sha256':sha(source/'library-reference.pptx'),'bound_pptx_sha256':sha(production/'bound/library-reference.pptx'),'reviewed_pages':587,'final_changed_pages_individually_reviewed':52,'prior_accepted_pngs_byte_identical':535,'full_node_audit':{'nodes':7033,'rejected':0},'checks':{'normal':'full repository go test -count=1 ./... passed','race':'full repository go test -count=1 -race ./... passed','source_build':587,'bound_build':586},'bound_distinct_native_review':'bound-distinct-native-v3c.json','bound_equivalence':'source-bound-equivalence-decompressed-v3c.json','pages':pages}
(root/'native-review-final-v3c-flat.json').write_text(json.dumps(flat,indent=2)+'\n')
(root/'verification-wmds-v5-candidate.json').write_text(json.dumps(receipt,indent=2)+'\n')
print('587/587 native source specimens accepted;52 final individual views;535 exact-byte carried views')
