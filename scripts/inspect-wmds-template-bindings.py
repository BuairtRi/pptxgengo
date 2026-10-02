#!/usr/bin/env python3
"""Read-only comparison of generated bound template artifacts to frozen sources."""
import argparse, hashlib, json, pathlib
p=argparse.ArgumentParser();p.add_argument('directory',type=pathlib.Path);p.add_argument('--bundle',type=pathlib.Path,default=pathlib.Path('library/wm-design-system/v1'));p.add_argument('--out',type=pathlib.Path,required=True);a=p.parse_args()
if a.out.exists():raise SystemExit('Output must be new')
read=lambda f:json.loads(f.read_text());sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
input_doc=read(a.directory/'template-content.json');compiled=read(a.directory/'compiled-document.json');bindings=read(a.directory/'binding-report.json');layout=read(a.directory/'layout-report.json')
assert len(input_doc['slides'])==len(compiled['slides'])==len(bindings['slides'])==len(layout['slides'])
records=[]
for supplied,slide,binding,reported in zip(input_doc['slides'],compiled['slides'],bindings['slides'],layout['slides']):
 assert supplied['id']==slide['id']==binding['slide_id']==reported['id']
 assert supplied['content_kind']==slide['content_kind']==binding['content_kind']
 assert slide['template_binding']==binding==reported['template_binding']
 assert not binding['native_qualified']
 source_path=a.bundle/'source'/binding['source_file'];assert sha(source_path)==binding['source_sha256']
 source=read(source_path);key=supplied['template'];source_slide=next(t['slide'] for t in source['templates'] if t['id']+'/'+t['variant']==key)
 assert binding['template']==key and len(source_slide['body'])==len(slide['nodes'])
 assert slide['frame']['rail']==source_slide['rail'] and slide['frame']['footer']==source_slide['footer']
 assert slide['frame']['surface']==source_slide.get('surface','light')
 assert slide['frame']['rail_surface']==source_slide.get('railSurface','inverse')
 values=supplied['values'];assert slide['title']==values['title'] and slide['eyebrow']==values['eyebrow']
 if 'source' in values:assert slide['source']==values['source'] and slide['frame']['source_lines']==2
 else:assert not slide.get('source') and slide['frame']['source_lines']==0
 ids={i['id']:i for i in binding['identities']};assert len(ids)==len(binding['identities'])
 node_map={n['id']:n for n in slide['nodes']}
 for n,sn in zip(slide['nodes'],source_slide['body']):
  assert ids[n['id']]['expected_type']==sn['type']
  rect=n['rect'];assert rect['x_pt']==sn['x'] and rect['y_pt']==sn['y']
  if sn['type']=='cardrow':
   row=n['card_row'];cards=values['cards'];assert len(row['items'])==len(cards)==len(sn['items'])
   assert rect['width_pt']==len(cards)*sn['w']+(len(cards)-1)*sn['gap'] and rect['height_pt']==sn['h']
   assert n['surface']==sn['card']['surface'] and row['numbering']==sn['numbering']
   assert [r['key'] for r in row['items']]==[c['key'] for c in cards]
   assert [r['key'] for r in binding['card_keys']]==[c['key'] for c in cards]
   for position,(r,c) in enumerate(zip(row['items'],cards),1):
    card=r['card'];assert card['padding_pt']==sn['card']['pad']
    assert card.get('band')==sn['card'].get('band') and card.get('number_ink','')==sn['card'].get('numInk','')
    assert card['title']==c['title'] and card['body']==[{'key':'copy','p':c['body']}]
    identity=binding['card_keys'][position-1];assert identity['ordinal']==position and identity['target_id']=='cards.items.'+c['key']
   continue
  assert rect['width_pt']==sn['w']
  if sn['type']=='metric':assert n['data_metric']==values[n['id']]
  elif sn['type']=='text':
   assert n['style']==sn['style'] and n['ink']==sn['ink'] and n['surface']==sn.get('on',source_slide.get('surface','light'))
   slot={'rail-eyebrow':'railEyebrow','rail-heading':'railHeading','rail-body':'railBody'}.get(n['id'],n['id']);value=values[slot]
   if isinstance(value,str):assert n['kind']=='text' and n['text']==value
   else:assert n['kind']=='richtext' and n['richtext']==value
  elif sn['type']=='rule':assert n['kind']=='rule' and rect['height_pt']==.75 and n['ink']=='line'
  else:raise AssertionError('Unsupported source shape')
 for assignment in binding['assignments']:
  assert assignment['target_id'] in ids and assignment['source_pointer']==ids[assignment['target_id']]['source_pointer']
  slot=assignment['slot']
  if slot.startswith('cards.'):
   _,key,field=slot.split('.');expected=next(c for c in values['cards'] if c['key']==key)[field]
  else:expected=values[slot]
  assert assignment['value']==expected
 records.append({'slide_id':slide['id'],'template':key,'source_sha256':binding['source_sha256'],'assignments':len(binding['assignments']),'source_geometry_styles_and_content_match':True})
# Paired reference instances prove content replacement retains fixed geometry.
pairs=[]
for i in range(0,len(compiled['slides'])-1,2):
 left,right=compiled['slides'][i:i+2]
 if left['template_binding']['template']!=right['template_binding']['template']:continue
 assert left['frame']==right['frame']
 assert [(n['id'],n['rect'],n.get('style'),n.get('ink'),n.get('surface'),n.get('scope')) for n in left['nodes']]==[(n['id'],n['rect'],n.get('style'),n.get('ink'),n.get('surface'),n.get('scope')) for n in right['nodes']]
 pairs.append({'slides':[i+1,i+2],'fixed_geometry_and_style_match':True})
result={'schema':'pptxgengo.wmds-binding-inspection.v1','input_sha256':sha(a.directory/'template-content.json'),'compiled_sha256':sha(a.directory/'compiled-document.json'),'binding_report_sha256':sha(a.directory/'binding-report.json'),'slides':records,'reference_pairs':pairs,'slot_assignments':sum(r['assignments'] for r in records),'passed':True,'general_envelope_qualified':False,'native_character_capture':False}
a.out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'slides':len(records),'assignments':result['slot_assignments'],'pairs':len(pairs),'passed':True}))
