#!/usr/bin/env python3
"""Build explicit source-panel and changed-content fixtures for Wave 2."""
from pathlib import Path
import copy,hashlib,json,posixpath,zipfile,xml.etree.ElementTree as E
R=Path(__file__).resolve().parents[1];D=R/'library/visual-components';D.mkdir(exist_ok=True)
A=R/'samples/visual-wave2/assets';assets=json.loads((A/'manifest.json').read_text())
source=R/'samples/UHG Fabric Platforming RFP Response - July 2026.pptx'
assert hashlib.sha256(source.read_bytes()).hexdigest()==assets['source_decks']['uhg']['sha256']
NS={'a':'http://schemas.openxmlformats.org/drawingml/2006/main','p':'http://schemas.openxmlformats.org/presentationml/2006/main','r':'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
with zipfile.ZipFile(source) as z:
 tree=E.fromstring(z.read('ppt/slides/slide28.xml'))
 rels={x.get('Id'):posixpath.normpath(posixpath.join('ppt/slides',x.get('Target'))) for x in E.fromstring(z.read('ppt/slides/_rels/slide28.xml.rels'))}
objects={int(n.get('id')):o for o in tree.find('p:cSld/p:spTree',NS) for n in [o.find('.//p:cNvPr',NS)] if n is not None}
def frame(o):
 x=o.find('p:spPr/a:xfrm',NS);off=x.find('a:off',NS);ext=x.find('a:ext',NS)
 return dict(x=int(off.get('x'))/12700,y=int(off.get('y'))/12700,width=int(ext.get('cx'))/12700,height=int(ext.get('cy'))/12700)
def run(text,id='text',size=12,bold=False,italic=False,underline=False,color='#070154'):
 return dict(id=id,text=text,font_face='Arial',font_size_pt=size,bold=bold,italic=italic,underline=underline,foreground=color)
def rich(id,bounds,paragraphs,align='left'):
 return dict(id=id,kind='text',bounds=bounds,paragraphs=paragraphs,align=align,valign='top',inset_x=0,inset_y=0)
def para(runs,id='p1',align='left',after=0,before=0):return dict(id=id,align=align,space_before_pt=before,space_after_pt=after,runs=runs)
def text(id,content,x,y,w,h,size=12,bold=False,color='#070154'):
 return dict(id=id,kind='text',bounds=dict(x=x,y=y,width=w,height=h),text=content,font_face='Arial',font_size_pt=size,bold=bold,foreground=color,align='left',valign='top',inset_x=0,inset_y=0)
def panel(prefix,dx=0,dy=0,scale=1):
 ids=[8,11,15,19,16,18];out=[]
 for id in ids:
  o=objects[id];b=frame(o);b=dict(x=dx+b['x']*scale,y=dy+b['y']*scale,width=b['width']*scale,height=b['height']*scale)
  if id in [8,11,15,19]:
   part=rels[o.find('p:blipFill/a:blip',NS).get('{'+NS['r']+'}embed')]
   a=next(a for a in assets['aliases'] if a['source_slide']=='uhg:28' and a['source_part']==part)
   item=dict(id=f'{prefix}-pic-{id}',kind='image',bounds=b,asset_path=str((A/a['extracted_file']).relative_to(R)),asset_sha256=a['source_part_sha256'],alt_text=f'Source UHG28 deliverable thumbnail, picture {id}; illustrative reference artwork',image_fit='stretch',layer=10+ids.index(id)*10)
   # Original picture 19 is visibly stretched; preserve it only in these source-derived fixtures.
   if id in [11,19]:item['allow_overlap']=[f'{prefix}-pic-{8 if id==11 else 15}']
  else:
   words=''.join(n.text or '' for n in o.findall('.//a:t',NS))
   caption_paragraph=para([run(words,size=9*scale,italic=True)],align='center')
   caption_paragraph['line_spacing_multiple']=.9
   item=rich(f'{prefix}-caption-{id}',b,[caption_paragraph],align='center');item['layer']=60
  out.append(item)
  if id in [8,11,15,19]:
   # Original picture outline resolves to theme accent3 and 0.75pt native line weight (read from source PowerPoint).
   x,y,w,h=b['x'],b['y'],b['width'],b['height']
   pair=[8,11] if id in [8,11] else [15,19]
   peers=[f'{prefix}-pic-{n}' for n in pair]+[f'{prefix}-border-{n}-{edge}' for n in pair for edge in ['top','right','bottom','left']]
   item['allow_overlap']=[peer for peer in peers if peer!=item['id']]
   for edge,bb in zip(['top','right','bottom','left'],[(x,y,w,0),(x+w,y,0,h),(x,y+h,w,0),(x,y,0,h)]):
    borderid=f'{prefix}-border-{id}-{edge}'
    out.append(dict(id=borderid,kind='line',bounds=dict(zip(['x','y','width','height'],bb)),foreground='#CED7E6',line_width_pt=.75*scale,layer=item['layer']+1,allow_overlap=[peer for peer in peers if peer!=borderid]))
 return out
base=json.loads((R/'library/layout-components/controls.json').read_text())
control=copy.deepcopy(base['slides'][0]);control.update(id='source-deliverable-panel',title='Deliverable preview panel',layouts=[],role='source composition control',takeaway='Keep each caption paired with its source thumbnails and preserve their overlap order.',notes='SOURCE RECONSTRUCTION CONTROL: only UHG28 deliverable thumbnail/caption composition. Original source frames, copy, italic caption typography and picture stretch retained; picture19 is distorted in original. No claim of whole-slide identity.')
control['title_bounds']['width']=580
control['canvas']=[copy.deepcopy(x) for x in control['canvas'] if x['id'] in ['footer-band','wm-logo','footer-copy','page']]
control['canvas']+=panel('source')
control['canvas'] += [text('source-explanation','Source reference: UHG slide 28',36,155,520,30,18,True),rich('control-notes',dict(x=36,y=204,width=510,height=200),[
 para([run('Four pinned source thumbnails',id='lead',size=14,bold=True),run(' form two preview pairs. Each pair retains its editable italic caption.',id='body',size=14)],after=12),
 para([run('Original geometry',id='lead',size=14,bold=True),run(' and overlap order are preserved. The lower-right image is stretched in the source; this control retains that source behavior explicitly.',id='body',size=14)],id='p2',after=12),
 para([run('The following phase-detail page',id='lead',size=14,bold=True),run(' translates and uniformly scales the paired composition into a dense proposal layout.',id='body',size=14)],id='p3')])]
adapt=copy.deepcopy(base['slides'][1]);adapt['id']='phase-detail-visual';adapt['layouts']=[l for l in adapt['layouts'] if l['id']!='acceptance'];adapt['canvas']=[c for c in adapt['canvas'] if c['id']!='deliverable-caption']
next(c for c in adapt['canvas'] if c['id']=='deliverable-title')['text']='Deliverable examples'
# Rigid source-composition transform fits the bounded right-column slot.
scale=.84;adapt['canvas']+=panel('adapted',dx=716-frame(objects[8])['x']*scale,dy=71-frame(objects[8])['y']*scale,scale=scale)
adapt['canvas'].append(text('source-artwork-note','Illustrative source examples from UHG 28',716,249,208,12,7.5))
for l in adapt['layouts']:
 for c in l['cells']:
  for b in c['blocks']:
   if b['id']=='phase-dates':
    b['paragraphs']=[para([run(b['text'],size=b['font_size_pt'],bold=True,italic=True,color=b['foreground'])])]
    for k in ['text','font_face','font_size_pt','bold','foreground']:b.pop(k,None)
adapt['notes']='ILLUSTRATIVE PROPOSAL. UHG28 source artwork is shown as sample deliverable imagery, not as completed work for this illustrative engagement. Caption/image pairs are translated and uniformly scaled as one composition. Source picture19 stretch is intentionally retained.'
for i,s in enumerate([control,adapt],1):
 next(c for c in s['canvas'] if c['id']=='page')['text']=str(i)
review=dict(schema='pptxgengo.compose-spec.v1',slides=[control,adapt])
(D/'deliverable-review.json').write_text(json.dumps(review,indent=2)+'\n')
(D/'deliverable-control.json').write_text(json.dumps(dict(schema=review['schema'],slides=[control]),indent=2)+'\n')
print(D/'deliverable-review.json')
