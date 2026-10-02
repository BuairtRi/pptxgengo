#!/usr/bin/env python3
"""Read-only package and PDF review for an already generated WMDS reference.

PDF selections prove visible content/line breaks, not native character bounds or
font file identity. Visual review is recorded separately by the reviewer.
"""
import argparse,hashlib,json,pathlib,re,xml.etree.ElementTree as ET,zipfile
p=argparse.ArgumentParser();p.add_argument('directory',type=pathlib.Path);p.add_argument('--stem',default='component-reference');p.add_argument('--out',required=True,type=pathlib.Path);a=p.parse_args()
if a.out.exists():raise SystemExit('Output must be new')
root=a.directory;report=json.loads((root/'layout-report.json').read_text());pdf=json.loads((root/'native-pdf-text.json').read_text());deck=root/(a.stem+'.pptx');pdfpath=root/(a.stem+'.pdf')
sha=lambda b:hashlib.sha256(b).hexdigest();ns={'p':'http://schemas.openxmlformats.org/presentationml/2006/main','a':'http://schemas.openxmlformats.org/drawingml/2006/main'}
assert sha(deck.read_bytes())==report['pptx_sha256'];assert sha(pdfpath.read_bytes())==pdf['pdf_sha256'];assert len(report['slides'])==pdf['page_count']
fold=lambda s:''.join(s.split())
def geometry(xfrm,b):
 off=xfrm.find('a:off',ns);ext=xfrm.find('a:ext',ns)
 actual=[int(off.attrib['x'])/12700,int(off.attrib['y'])/12700,int(ext.attrib['cx'])/12700,int(ext.attrib['cy'])/12700]
 expected=[b[k] for k in ['x_pt','y_pt','width_pt','height_pt']]
 assert all(abs(x-y)<=.01 for x,y in zip(actual,expected)),(actual,expected)

records=[];totals={'components':0,'text_records':0,'predicted_visible_lines':0,'native_groups':0,'preserved_paragraphs':0,'card_rows':0,'rich_text_records':0,'mixed_runs':0,'unanchored_rich_runs':0,'primitive_shapes':0,'bound_templates':0};failures=[]
with zipfile.ZipFile(deck)as z:
 assert z.testzip() is None
 for i,(s,page)in enumerate(zip(report['slides'],pdf['pages']),1):
  assert page['width_pt']==960 and page['height_pt']==540
  sp=ET.fromstring(z.read(f'ppt/slides/slide{i}.xml'));groups=sp.findall('.//p:grpSp',ns)
  components=s.get('components',[]);rows=s.get('card_rows',[])
  assert len(groups)==len(components)+len(rows);totals['native_groups']+=len(groups)
  ids=[x.attrib['id']for x in sp.findall('.//p:cNvPr',ns)];assert len(ids)==len(set(ids))
  byname={g.find('p:nvGrpSpPr/p:cNvPr',ns).attrib['name']:g for g in groups};assert len(byname)==len(groups)
  for c in components:
   g=byname[c['id']]
   xf=g.find('p:grpSpPr/a:xfrm',ns);geometry(xf,c['rect'])
   assert xf.find('a:off',ns).attrib==xf.find('a:chOff',ns).attrib
   assert xf.find('a:ext',ns).attrib==xf.find('a:chExt',ns).attrib
   assert {x.attrib['name']for x in g.findall('p:sp/p:nvSpPr/p:cNvPr',ns)}==set(c['parts'])
   totals['components']+=1
  for row in rows:
   g=byname[row['id']];xf=g.find('p:grpSpPr/a:xfrm',ns);geometry(xf,row['rect'])
   assert xf.find('a:off',ns).attrib==xf.find('a:chOff',ns).attrib
   assert xf.find('a:ext',ns).attrib==xf.find('a:chExt',ns).attrib
   assert [child.find('p:nvGrpSpPr/p:cNvPr',ns).attrib['name']for child in g.findall('p:grpSp',ns)]==row['parts']
   assert len(row['keys'])==len(row['parts']) and len(row['keys'])==len(set(row['keys']))
   children=[next(c for c in components if c['id']==part)for part in row['parts']]
   assert all(abs(c['rect']['height_pt']-row['rect']['height_pt'])<=.01 and abs(c['rect']['y_pt']-row['rect']['y_pt'])<=.01 for c in children)
   assert all(abs(c['rect']['width_pt']-children[0]['rect']['width_pt'])<=.01 for c in children)
   assert all(abs(right['rect']['x_pt']-left['rect']['x_pt']-left['rect']['width_pt']-row['gap_pt'])<=.01 for left,right in zip(children,children[1:]))
   totals['card_rows']+=1
  rel=ET.fromstring(z.read(f'ppt/slides/_rels/slide{i}.xml.rels'))
  target=next(x.attrib['Target']for x in rel if x.attrib['Type'].endswith('/slideLayout'))
  import posixpath
  layout=ET.fromstring(z.read(posixpath.normpath(posixpath.join('ppt/slides',target))))
  native={x.find('p:nvSpPr/p:cNvPr',ns).attrib['name']:x for r in [layout,sp]for x in r.findall('.//p:sp',ns)}
  if s.get('template_binding'):totals['bound_templates']+=1
  totals['primitive_shapes']+=len(s.get('shapes',[]))
  for c in s.get('components',[])+[{'shapes':s.get('shapes',[])}]:
   for sh in c.get('shapes',[]):
    obj=native[sh['id']];geometry(obj.find('p:spPr/a:xfrm',ns),sh['rect'])
    assert obj.find('p:spPr/a:solidFill/a:srgbClr',ns).attrib['val']==sh['color']
    assert obj.find('p:spPr/a:prstGeom',ns).attrib['prst']==sh.get('geometry','rect')
    assert abs(int(obj.find('p:spPr/a:xfrm',ns).attrib.get('rot','0'))/60000-sh.get('rotation_deg',0))<=.01
  for tr in s['texts']:
   totals['text_records']+=1;shape=native[tr['id']];geometry(shape.find('p:spPr/a:xfrm',ns),tr['rect']);paras=shape.findall('p:txBody/a:p',ns)
   bp=shape.find('p:txBody/a:bodyPr',ns);assert all(bp.attrib[k]=='0'for k in ['lIns','tIns','rIns','bIns'])
   assert bp.find('a:spAutoFit',ns)is None and bp.find('a:normAutofit',ns)is None
   font=tr['layout']['font'];st=tr['layout']['style']
   def properties(rp,font,st,color):
    assert rp.find('a:latin',ns).attrib['typeface']==font['pptx_typeface']
    assert int(rp.attrib['sz'])==round(st['size']*100) and int(rp.attrib.get('spc','0'))==round(st['tracking_pt']*100)
    assert rp.attrib['b']==str(int(font['pptx_bold'])) and rp.attrib['i']==str(int(font['pptx_italic'])) and rp.attrib['kern']=='0'
    assert rp.find('a:solidFill/a:srgbClr',ns).attrib['val']==color
   rich=tr.get('rich')
   if rich:
    assert len(paras)==len(rich['paragraphs']);totals['rich_text_records']+=1
   for pi,para in enumerate(paras):
    assert int(para.find('a:pPr/a:lnSpc/a:spcPts',ns).attrib['val'])==round(st['leading']*100)
    properties(para.find('a:endParaRPr',ns),font,st,tr['color'])
    if rich:
     authored=rich['paragraphs'][pi]['runs'];native_runs=para.findall('a:r',ns);assert len(authored)==len(native_runs)
     for run,native_run in zip(authored,native_runs):
      assert native_run.find('a:t',ns).text==run['displayed']
      properties(native_run.find('a:rPr',ns),run['font'],run['style'],run['color'])
      totals['mixed_runs']+=1;totals['unanchored_rich_runs']+=int(not run['vertical_anchored'])
    else:
     for rp in para.findall('a:r/a:rPr',ns)+para.findall('a:fld/a:rPr',ns):properties(rp,font,st,tr['color'])
   actual='\n'.join(''.join(t.text or ''for t in para.findall('.//a:t',ns))for para in paras)
   assert actual==tr['layout']['displayed'];totals['preserved_paragraphs']+=len(paras)
   assert all(p.find('a:endParaRPr/a:latin',ns)is not None for p in paras)
   expected=[l['text']for l in tr['layout']['lines']if l['text'].strip()]
   b=tr['rect'];selected=[]
   for line in page['lines']:
    lb=line['bounds'];cy=lb['y']+lb['height']/2
    # PDFKit can combine the inline number and title into one physical line.
    # Select lines intersecting the text region, then compare each predicted line.
    if b['y_pt']-1<=cy<=b['y_pt']+b['height_pt']+1 and lb['x']<b['x_pt']+b['width_pt']+.5 and lb['x']+lb['width']>b['x_pt']-.5:
     selected.append(line['text'])
   matches=[any(fold(t)in fold(x)for x in selected)for t in expected]
   totals['predicted_visible_lines']+=len(expected)
   if not all(matches):failures.append({'slide':i,'id':tr['id'],'expected':expected,'observed':selected,'matches':matches})
   records.append({'slide':i,'id':tr['id'],'predicted_lines':len(expected),'matched_lines':sum(matches),'saved_text_and_paragraphs_match':True})
  # Every filled native shape emitted by the WMDS v2 helper is explicitly line-free.
  for tree in [layout,sp]:
   for sh in tree.findall('.//p:sp',ns):
    if sh.find('p:txBody',ns)is None:
     line=sh.find('p:spPr/a:ln',ns);assert line is not None and line.find('a:noFill',ns)is not None
  assert all('IBM' in n for n in page['font_names']),page['font_names']
result={'schema':'pptxgengo.wmds-component-inspection.v1','pptx_sha256':sha(deck.read_bytes()),'pdf_sha256':sha(pdfpath.read_bytes()),'totals':totals,'pdf_line_checks':records,'failures':failures,'passed':not failures,'native_character_capture':False,'general_envelope_qualified':False,'exact_native_font_file_identity':False,'xml_geometry_and_styles_match':True,'pdf_selection_bounds_are_native_character_bounds':False}
a.out.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'totals':totals,'failures':failures,'passed':not failures},indent=2))
raise SystemExit(1 if failures else 0)
