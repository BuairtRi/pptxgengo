#!/usr/bin/env python3
"""Phrase-anchor controls and changed-content stress fixtures; source assets stay pinned."""
import copy,hashlib,json,zipfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'library/diagram-components'
ASSETS=ROOT/'samples/visual-wave3/assets'
def rect(x,y,w,h):return dict(x=x,y=y,width=w,height=h)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
 OUT.mkdir(parents=True,exist_ok=True);ASSETS.mkdir(parents=True,exist_ok=True)
 source=ROOT/'samples/UHG Fabric Platforming RFP Response - July 2026.pptx'
 with zipfile.ZipFile(source) as z:
  for name in ['image77.svg','image90.png']:
   p=ASSETS/name;b=z.read('ppt/media/'+name)
   if p.exists() and p.read_bytes()!=b:raise ValueError('asset conflict: '+str(p))
   p.write_bytes(b)
 base=json.loads((ROOT/'library/visual-components/followup-review.json').read_text())
 footer=[c for c in base['slides'][0]['canvas'] if c['id'] in ['footer-band','wm-logo','footer-copy','page']]
 under=ASSETS/'image77.svg';highlight=ASSETS/'image90.png'
 fallback=ROOT/'samples/reconstruction/work/underline-source.svg.png'
 alpha=json.loads((ROOT/'samples/reconstruction/work/underline-visible-bounds.json').read_text())
 vb=alpha['viewBox'];ink=alpha['visibleBounds'];alpha_rect=rect(ink['x']/vb[2],ink['y']/vb[3],ink['width']/vb[2],ink['height']/vb[3])
 def slide(identity,text,width=888,size=24,mode='underline',phrase='pivotal moment',**extra):
  cells=copy.deepcopy(footer)
  cells += [dict(id='claim',kind='text',bounds=rect(36,40,width,180),text=text,font_face='Arial',font_size_pt=size,bold=True,foreground='#070154',align='left',valign='top'),dict(id='explanation',kind='text',bounds=rect(36,285,860,90),text='SOURCE-STYLE CONTROL / ILLUSTRATIVE VARIANT: The selected phrase is measured from native character geometry. Artwork is anchored to the phrase, with explicit optical calibration and no edits to make the words fit the accent.',font_face='Arial',font_size_pt=14,foreground='#50658E',align='left',valign='top')]
  accent=dict(id='emphasis',target='claim',mode=mode,phrase=phrase,asset_path=str(under.relative_to(ROOT)),asset_sha256=sha(under),fallback_asset_path=str(fallback.relative_to(ROOT)),fallback_asset_sha256=sha(fallback),alpha_bounds=alpha_rect,padding_x_pt=-1.2317322059691946,offset_x_pt=.11945163576578466,offset_y_pt=-3.227455397478323,stroke_height_pt=2,container_aspect=1.64819110185159)
  if mode=='highlight':accent=dict(id='emphasis',target='claim',mode=mode,phrase=phrase,asset_path=str(highlight.relative_to(ROOT)),asset_sha256=sha(highlight),alpha_bounds=rect(0,0,1,1),padding_x_pt=5.058313518049971,padding_y_pt=5.220756405694651,offset_x_pt=1.6051474964713748,offset_y_pt=.18744265638906654,rotation_deg=1)
  accent.update(extra)
  return dict(id=identity,title='',role='Accent placement qualification',takeaway='Emphasis follows the intended phrase through layout changes.',notes='UHG5 underline / UHG6 highlight optical calibration. Source SHA256 '+sha(source)+'. This is a component control, not whole-slide identity.',width_pt=960,height_pt=540,title_bounds=rect(0,0,0,0),title_font_face='Arial',title_font_size_pt=24,title_foreground='#070154',canvas=cells,accents=[accent],pods=[])
 text='You’re at a pivotal moment in your journey to establish Fabric as a governed enterprise platform without creating another data silo.'
 slides=[slide('underline-source-copy',text),slide('underline-narrow-wrap',text,width=440),slide('underline-repeated','A pivotal moment requires evidence. The next pivotal moment requires an explicit decision.',width=670,occurrence=2,multiline='per_line'),slide('underline-per-line','Frame the pivotal moment clearly before choosing the next delivery decision.',width=230,multiline='per_line'),slide('highlight-source-copy','That’s where West Monroe comes in',mode='highlight',phrase='West Monroe'),slide('highlight-larger-wrap','That’s where West Monroe comes in',width=430,size=32,mode='highlight',phrase='West Monroe')]
 for i,s in enumerate(slides,1):
  next(c for c in s['canvas'] if c['id']=='page')['text']=str(i)
  if s['id']=='underline-per-line':s['accents'][0]['phrase']='pivotal moment clearly'
 manual=slide('ambiguous-manual-stage','A pivotal moment informs the next pivotal moment.',width=820,staging=dict(asset_bounds=rect(600,235,150,90),note_bounds=rect(600,350,280,100),note='Place this underline under the intended occurrence of “pivotal moment.” Two occurrences exist; choose the takeaway before positioning it.'))
 manual['canvas']=[c for c in manual['canvas'] if c['id']!='explanation']
 for name,items in [('accent-review',slides),('accent-manual',[manual]),('accent-qualification',slides+[manual])]:
  (OUT/(name+'.json')).write_text(json.dumps(dict(schema=base['schema'],slides=items),indent=2)+'\n')
 print('Wrote six regular accent fixtures plus one explicitly unfinished manual-placement fixture.')
if __name__=='__main__':main()
