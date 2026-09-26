#!/usr/bin/env python3
"""Create a task copy with hidden slides shown for unambiguous PDF page mapping.

Originals remain unchanged. Only each slide root's show attribute changes; a
sidecar records every source slide number and the exact original source hash.
"""
import argparse
import hashlib
import json
from pathlib import Path
import posixpath
import re
import xml.etree.ElementTree as ET
from zipfile import ZipFile, ZIP_DEFLATED

P='http://schemas.openxmlformats.org/presentationml/2006/main'
R='http://schemas.openxmlformats.org/officeDocument/2006/relationships'

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--source',type=Path,required=True)
    p.add_argument('--out',type=Path,required=True)
    a=p.parse_args()
    report=Path(str(a.out)+'.render-map.json')
    if a.out.exists() or report.exists():p.error('Output or render map exists')
    source_hash=hashlib.sha256(a.source.read_bytes()).hexdigest()
    changes={}; pages=[]
    with ZipFile(a.source) as z:
        pres=ET.fromstring(z.read('ppt/presentation.xml'))
        rels={r.get('Id'):posixpath.normpath(posixpath.join('ppt',r.get('Target'))).lstrip('/')
              for r in ET.fromstring(z.read('ppt/_rels/presentation.xml.rels')) if r.get('TargetMode')!='External'}
        for i,s in enumerate(pres.findall('{'+P+'}sldIdLst/{'+P+'}sldId'),1):
            part=rels[s.get('{'+R+'}id')];data=z.read(part)
            root=ET.fromstring(data);hidden=root.get('show') in ('0','false')
            if hidden:
                data,n=re.subn(rb'(<(?:[A-Za-z_][\w.-]*:)?sld\b[^>]*\s)show=["\'](?:0|false)["\']',rb'\1show="1"',data,count=1)
                if n!=1 or ET.fromstring(data).get('show')!='1':raise ValueError('Could not unhide '+part)
                changes[part]=data
            pages.append({'pdf_page':i,'source_slide_number':i,'part':part,'source_hidden':hidden})
        a.out.parent.mkdir(parents=True,exist_ok=True)
        created_output=False
        created_report=False
        try:
            with a.out.open('xb') as f:
                created_output=True
                with ZipFile(f,'w',ZIP_DEFLATED) as out:
                    for entry in z.infolist():out.writestr(entry,changes.get(entry.filename,z.read(entry.filename)))
            with report.open('x') as f:
                created_report=True
                json.dump({'source':str(a.source),'source_sha256':source_hash,'expected_pdf_pages':len(pages),
                           'render_only_changes':'Hidden slide root show attributes set to 1; no source edit',
                           'pages':pages},f,indent=2);f.write('\n')
        except Exception:
            if created_report:report.unlink(missing_ok=True)
            if created_output:a.out.unlink(missing_ok=True)
            raise
    print(json.dumps({'out':str(a.out),'expected_pdf_pages':len(pages),'unhidden_slides':len(changes)}))
if __name__=='__main__':main()
