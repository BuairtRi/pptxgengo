#!/usr/bin/env python3
"""Find review candidates across different group encodings using resolved frames.

This is retrieval only. It never merges layouts or declares pixel equivalence.
"""
import argparse
from collections import defaultdict, Counter
import hashlib
import json
import os
from pathlib import Path
import tempfile


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def features(row):
    b = row.get('bounds_emu'); size = row['slide_size_emu']
    if b is None or row['object_kind'] == 'grpSp':
        return None
    ph = row.get('placeholder') or {}
    if ph.get('type') in {'sldNum', 'ftr', 'dt'}:
        return None
    if 'think-cell data' in (row.get('shape_name') or '').lower():
        return None
    x,y,w,h = b['x']/size['width'], b['y']/size['height'], b['width']/size['width'], b['height']/size['height']
    if (w < .001 and h < .001) or x+w < 0 or y+h < 0 or x > 1 or y > 1:
        return None
    kind = row['object_kind']
    role = ('picture' if kind == 'pic' else 'graphic_frame' if kind == 'graphicFrame' else
            'connector' if kind == 'cxnSp' else 'text' if (row.get('text') or '').strip() else 'shape')
    return {'id':row['occurrence_id'],'path':row['object_path'],'role':role,
            'box':[x,y,w,h], 'preset':row.get('preset_geometry')}


def compare(left, right):
    if min(len(left),len(right)) < 4 or min(len(left),len(right))/max(len(left),len(right)) < .70:
        return None
    edges=[]
    for i,l in enumerate(left):
        for j,r in enumerate(right):
            if l['role'] != r['role']:
                continue
            errors=[abs(x-y) for x,y in zip(l['box'],r['box'])]
            if max(errors) <= .035:
                edges.append((sum(errors)/4,max(errors),i,j))
    used_l=set();used_r=set();matches=[]
    for mean,worst,i,j in sorted(edges):
        if i in used_l or j in used_r:
            continue
        used_l.add(i);used_r.add(j)
        matches.append({'left_path':left[i]['path'],'right_path':right[j]['path'],
                        'mean_frame_error':mean,'max_frame_error':worst,
                        'preset_changed':left[i]['preset'] != right[j]['preset']})
    ratio=len(matches)/max(len(left),len(right))
    if len(matches)<4 or ratio < .72:
        return None
    avg=sum(m['mean_frame_error'] for m in matches)/len(matches)
    return {'match_ratio':ratio,'mean_frame_error':avg,'score':round(ratio-avg*3,6),
            'left_feature_count':len(left),'right_feature_count':len(right),'matches':matches,
            'unmatched_left':[l['path'] for i,l in enumerate(left) if i not in used_l],
            'unmatched_right':[r['path'] for i,r in enumerate(right) if i not in used_r]}


def main():
    p=argparse.ArgumentParser(description=__doc__)
    for name in ('geometry','decisions','out','report'):
        p.add_argument('--'+name,type=Path,required=True)
    p.add_argument('--neighbors',type=int,default=3)
    a=p.parse_args()
    if a.out.exists() or a.report.exists():p.error('Output/report exists')
    if not 1<=a.neighbors<=10:p.error('Neighbors must be 1–10')
    decisions=json.loads(a.decisions.read_text()); groups={}; known=decisions['source_hashes']
    group_ids=set()
    for g in decisions['groups']:
        if g['id'] in group_ids:raise ValueError('Duplicate layout decision ID')
        group_ids.add(g['id'])
        for member in g['members']:
            if member in groups:raise ValueError('Overlapping layout decision members')
            groups[member]=g['id']
    separated=set()
    for pair in decisions.get('pair_decisions',[]):
        if pair['decision']=='keep_separate':
            separated.add(tuple(sorted(groups.get(pair[k],pair[k]) for k in ('left','right'))))
    slides=defaultdict(list);unresolved=Counter();source_hashes={};seen=set()
    for line in a.geometry.read_text().splitlines():
        if not line.strip():continue
        row=json.loads(line);ref=f"{row['source_id']}:{row['slide_number']:03d}"
        if row['occurrence_id'] in seen:raise ValueError('Duplicate geometry occurrence')
        seen.add(row['occurrence_id'])
        if known.get(row['source_id']) != row['source_sha256']:raise ValueError('Geometry/decision source mismatch')
        source_hashes[row['source_id']]=row['source_sha256']
        slides[ref]
        if row.get('bounds_emu') is None:unresolved[ref]+=1
        f=features(row)
        if f:slides[ref].append(f)
    if source_hashes!=known:raise ValueError('Source coverage mismatch')
    if set(groups)-set(slides):raise ValueError('Reviewed slide missing from geometry input')
    best={};refs=sorted(slides)
    for i,left in enumerate(refs):
        for right in refs[i+1:]:
            lg,rg=groups.get(left,left),groups.get(right,right)
            if lg==rg:continue
            if tuple(sorted((lg,rg))) in separated:continue
            result=compare(slides[left],slides[right])
            if result is None:continue
            key=tuple(sorted((lg,rg)))
            row={'left':left,'right':right,'left_work_unit':lg,'right_work_unit':rg,
                 'status':'review_required','auto_merge':False,**result,
                 'unresolved_frames':{'left':unresolved[left],'right':unresolved[right]},
                 'warnings':['Position-sorted greedy frame correspondence is retrieval evidence only',
                             'Styles, text roles, layer order, visible ink and content fit require review',
                             'Different cardinalities remain distinct unless explicitly reviewed']}
            if key not in best or row['score']>best[key]['score']:best[key]=row
    ranked=sorted(best.values(),key=lambda r:(-r['score'],r['left'],r['right']))
    queue=[];degree=Counter()
    for row in ranked:
        l,r=row['left_work_unit'],row['right_work_unit']
        if degree[l]>=a.neighbors or degree[r]>=a.neighbors:continue
        degree[l]+=1;degree[r]+=1;queue.append(row)
    out={'schema_version':1,'algorithm':'resolved-frame-retrieval-2','source_hashes':source_hashes,
         'inputs':{'geometry_sha256':sha(a.geometry),'decisions_sha256':sha(a.decisions)},
         'summary':{'source_slides':len(slides),'scored_work_unit_pairs':len(ranked),'queued_pairs':len(queue)},
         'review_queue':queue}
    report='# Broader layout retrieval\n\nThis queue compares flattened object frames, allowing different group encodings and unmatched objects. It is not a merge decision.\n\n'
    report+=f'{len(ranked)} scored work-unit pairs; {len(queue)} pairs queued with at most {a.neighbors} neighbors per work unit.\n\n'
    report+='| Left | Right | Matched fraction | Mean frame error |\n|---|---|---:|---:|\n'
    report+=''.join(f"| {r['left']} | {r['right']} | {r['match_ratio']:.3f} | {r['mean_frame_error']:.5f} |\n" for r in queue)
    staged=[];installed=[]
    try:
        for path,content in [(a.out,json.dumps(out,indent=2)+'\n'),(a.report,report)]:
            path.parent.mkdir(parents=True,exist_ok=True)
            fd,name=tempfile.mkstemp(prefix='.'+path.name+'.',dir=path.parent);staged.append(Path(name))
            with os.fdopen(fd,'w') as f:f.write(content);f.flush();os.fsync(f.fileno())
        for src,dst in zip(staged,(a.out,a.report)):os.link(src,dst);installed.append(dst)
    except Exception:
        for dst in installed:dst.unlink(missing_ok=True)
        raise
    finally:
        for src in staged:src.unlink(missing_ok=True)
    print(json.dumps(out['summary']))

if __name__=='__main__':main()
