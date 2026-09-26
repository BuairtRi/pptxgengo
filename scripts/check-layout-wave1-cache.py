#!/usr/bin/env python3
"""Exercise cache integrity against local native evidence; outputs must be new.

Usage: check-layout-wave1-cache.py /path/to/the-original-measured-cli
"""
from pathlib import Path
import hashlib,json,subprocess,shutil,sys
R=Path(__file__).resolve().parents[1];D=R/'samples/layout-wave1/qa';C=R/'samples/layout-wave1/cache';B=R/'samples/component-adaptation/dynamic-pods';CLI=sys.argv[1] if len(sys.argv)>1 else '/tmp/pptxcompose-wave1-release'
def read(p):return json.loads(p.read_text())
def write(p,x):p.write_text(json.dumps(x,indent=2)+'\n')
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
evidence=B/'wave1-release-evidence.json';env=read(evidence)['environment']
# Go JSON field order is significant in environment serialization. Resolve by
# matching the stored environment binding in a known native build manifest.
control=read(B/'wave1-controls-release/manifest.json');envkey=next(x.name for x in (C/'entries').iterdir() if any(read(p)['evidence']==str(evidence) for p in x.glob('*.json')))
results=[]
def prepare(name):
 cache=D/(name+'-cache');shutil.copytree(C/'entries'/envkey,cache/'entries'/envkey)
 entrypath=next(p for p in (cache/'entries'/envkey).glob('*.json') if read(p)['evidence']==str(evidence))
 return cache,entrypath,read(entrypath)
def run(name,cache,want):
 out=D/(name+'-output');p=subprocess.run([CLI,'build','--spec',str(R/'library/layout-components/controls.json'),'--cache',str(cache),'--out',str(out)],text=True,capture_output=True)
 row=dict(case=name,exit_code=p.returncode,error=p.stderr.strip(),no_output=not out.exists(),passed=p.returncode!=0 and want in p.stderr and not out.exists());results.append(row)
cache,p,e=prepare('contract-tamper');e['contract']['text']+=' changed';write(p,e);run('contract-tamper',cache,'cache contract mismatch')
cache,p,e=prepare('evidence-hash');copy=D/'evidence-whitespace.json';copy.write_bytes(evidence.read_bytes()+b'\n');e['evidence']=str(copy);write(p,e);run('evidence-hash',cache,'source hash/environment mismatch')
cache,p,e=prepare('legacy-environment');copy=D/'legacy-evidence.json';v=read(evidence);del v['environment'];write(copy,v);e['evidence']=str(copy);e['evidence_sha256']=sha(copy);write(p,e);run('legacy-environment',cache,'environment fingerprint')
cache,p,e=prepare('legacy-native-schema');copy=D/'legacy-native.json';v=read(evidence);v['native']['schema']='pptxgengo.compose-text-measurement.v4';write(copy,v);e['evidence']=str(copy);e['evidence_sha256']=sha(copy);write(p,e);run('legacy-native-schema',cache,'requires v5')
cache,p,e=prepare('request-binding');newbundle=D/'request-binding-bundle';shutil.copytree(Path(e['bundle']),newbundle);m=read(newbundle/'manifest.json');m['requests'][0]['text']+=' changed';write(newbundle/'manifest.json',m);v=read(evidence);v['manifest_sha256']=sha(newbundle/'manifest.json');copy=D/'request-binding-evidence.json';write(copy,v);e['bundle']=str(newbundle);e['evidence']=str(copy);e['evidence_sha256']=sha(copy);write(p,e);run('request-binding',cache,'requests do not describe probe deck')
cache,p,e=prepare('summary-ignored');copy=D/'summary-edited.json';v=read(evidence)
for item in v['measurements']['by_request_id'].values():item['rendered_height_pt']=1;item['rendered_width_pt']=1
write(copy,v);e['evidence']=str(copy);e['evidence_sha256']=sha(copy);write(p,e);out=D/'summary-ignored-output';proc=subprocess.run([CLI,'build','--spec',str(R/'library/layout-components/controls.json'),'--cache',str(cache),'--out',str(out)],text=True,capture_output=True)
passed=proc.returncode==0 and read(out/'manifest.json')['plan']==control['plan'];results.append(dict(case='summary-ignored',exit_code=proc.returncode,passed=passed,meaning='Editable summary changes did not alter the plan reconstructed from original native rows.'))
write(D/'cache-integrity-results.json',results);print(json.dumps(results,indent=2));assert all(x['passed'] for x in results)
