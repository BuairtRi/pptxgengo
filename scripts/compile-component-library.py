#!/usr/bin/env python3
"""Merge reviewed component slices, deduplicate exact selections, retain family variants.

This compiles catalog metadata. It neither edits slides nor approves adaptation.
Input order sets the canonical identity of an exact duplicate (base seeds first).
"""
import argparse
import copy
import hashlib
import json
import os
import tempfile
from collections import Counter, defaultdict
from pathlib import Path


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--inputs', nargs='+', type=Path, required=True)
    p.add_argument('--taxonomy', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--families', type=Path, required=True)
    p.add_argument('--report', type=Path, required=True)
    a = p.parse_args()
    outputs = (a.out, a.families, a.report)
    if len(set(outputs)) != 3 or any(x.exists() for x in outputs):
        p.error('Choose three distinct new output paths')
    taxonomy = json.loads(a.taxonomy.read_text())
    definitions = {f['id']: f for f in taxonomy['families']}
    if len(definitions) != len(taxonomy['families']):
        raise ValueError('Duplicate family definitions')
    known_profiles = {s['id'] for s in taxonomy['style_profiles']}
    for f in definitions.values():
        if not set(f['proposed_style_options']).issubset(known_profiles):
            raise ValueError('Unknown style profile: '+f['id'])
    merged = []; seen_seeds = set(); seen_instance_ids = set(); exact = {}; retained_ids = {}; aliases = []; hashes = None
    counts = Counter({'input_instances': 0, 'exact_duplicate_instances': 0, 'reviewed_duplicate_instances': 0}); by_family = defaultdict(list)
    for path in a.inputs:
        doc = json.loads(path.read_text())
        if hashes is None:
            hashes = doc['source_hashes']
        if doc['source_hashes'] != hashes:
            raise ValueError('Mismatched source hashes: '+str(path))
        for original in doc['seeds']:
            seed = copy.deepcopy(original)
            if seed['id'] in seen_seeds:
                raise ValueError('Repeated seed ID: '+seed['id'])
            seen_seeds.add(seed['id'])
            family_id = taxonomy['seed_family_overrides'].get(seed['id'])
            if not family_id:
                candidate = seed.get('semantic_family')
                family_id = taxonomy['family_aliases'].get(candidate, candidate)
            if family_id not in definitions:
                raise ValueError(f"Unmapped family: {seed['id']} ({seed.get('semantic_family')})")
            seed['canonical_family_id'] = 'component-family:'+family_id
            seed['proposed_style_options'] = ['component-style:'+v for v in definitions[family_id]['proposed_style_options']]
            seed['style_application_status'] = 'contract_only_no_object_role_bindings'
            seed['input_slice'] = str(path)
            kept = []
            for index, inst in enumerate(seed['instances'], 1):
                counts['input_instances'] += 1
                for field in ('structural_variant', 'observed_style_description', 'evidence_status', 'evidence'):
                    if field not in inst and field in seed:
                        inst[field] = copy.deepcopy(seed[field])
                sid = inst['source_id']
                paths = inst['object_paths']
                if sid not in hashes or not paths or len(set(paths)) != len(paths):
                    raise ValueError('Invalid source/path list in '+seed['id'])
                if any(q.startswith(p+'/') for p in paths for q in paths if p != q):
                    raise ValueError('Ancestor and descendant selected in '+seed['id'])
                inst['instance_id'] = inst.get('instance_id', seed['id']+f':instance-{index:03d}')
                if not isinstance(inst['instance_id'], str) or not inst['instance_id'] or inst['instance_id'] in seen_instance_ids:
                    raise ValueError('Missing or duplicate instance ID: '+str(inst['instance_id']))
                seen_instance_ids.add(inst['instance_id'])
                key = (sid, hashes[sid], inst['slide_number'], tuple(sorted(paths)))
                if inst.get('duplicate_of'):
                    target = retained_ids.get(inst['duplicate_of'])
                    if not target or target['family_id'] != family_id or not inst.get('duplicate_review_reason'):
                        raise ValueError('Reviewed duplicate needs an earlier same-family target and rationale: '+inst['instance_id'])
                    aliases.append({'alias_id': inst['instance_id'], 'canonical_id': inst['duplicate_of'],
                                    'alias_seed_id': seed['id'], 'reason': inst['duplicate_review_reason'],
                                    'alternative_annotation': inst})
                    counts['reviewed_duplicate_instances'] += 1
                    continue
                if key in exact:
                    canonical = exact[key]
                    if retained_ids[canonical['instance_id']]['family_id'] != family_id:
                        raise ValueError('Exact selection has conflicting family assignments: '+inst['instance_id'])
                    aliases.append({'alias_id': inst['instance_id'], 'canonical_id': canonical['instance_id'],
                                    'alias_seed_id': seed['id'], 'reason': 'identical_source_slide_and_selected_roots',
                                    'alternative_annotation': inst})
                    counts['exact_duplicate_instances'] += 1
                    continue
                exact[key] = inst
                retained_ids[inst['instance_id']] = {'family_id': family_id, 'instance': inst}
                kept.append(inst)
                by_family[family_id].append({'id': inst['instance_id'], 'seed_id': seed['id'],
                    'source_id': sid, 'slide_number': inst['slide_number'],
                    'structural_variant': inst.get('structural_variant', seed['id']),
                    'observed_style_description': inst.get('observed_style_description'),
                    'evidence_status': inst.get('evidence_status', 'prior_candidate_review')})
            seed['instances'] = kept
            if kept:
                merged.append(seed)
    all_ids = [i['instance_id'] for s in merged for i in s['instances']]
    if len(all_ids) != len(set(all_ids)):
        raise ValueError('Duplicate instance IDs')
    counts.update({'source_pattern_seeds': len(merged), 'retained_source_compositions': len(all_ids),
        'canonical_semantic_families': len(by_family),
        'named_slot_candidates': sum(len(i.get('slots', [])) for s in merged for i in s['instances']),
        'source_slides_with_compositions': len({(i['source_id'],i['slide_number']) for s in merged for i in s['instances']}),
        'adaptation_approved_components': 0})
    lineage = [{'path': str(x), 'sha256': digest(x)} for x in a.inputs+[a.taxonomy]]
    seeds_doc = {'schema_version': 2, 'source_hashes': hashes,
        'readiness': 'proposed_semantic_boundaries_not_adaptation_tested',
        'inputs': lineage, 'summary': dict(counts), 'exact_duplicate_aliases': aliases, 'seeds': merged}
    family_rows = []
    for fid, instances in sorted(by_family.items()):
        family_rows.append({**definitions[fid], 'id': 'component-family:'+fid, 'kind': 'component_family',
            'proposed_style_options': ['component-style:'+v for v in definitions[fid]['proposed_style_options']],
            'source_hashes': hashes, 'readiness': 'semantic_candidate',
            'style_application_status': 'contract_only_no_object_role_bindings',
            'instances': instances, 'source_pattern_ids': sorted({i['seed_id'] for i in instances}),
            'simplification_rule': 'Shared retrieval family; structural variants are not interchangeable without adaptation proof.'})
    styles = [{**s, 'kind': 'style_profile', 'id': 'component-style:'+s['id'],
               'readiness': 'proposed_contract', 'guidance': taxonomy['guidance'],
               'application_rules': taxonomy['application_rules']} for s in taxonomy['style_profiles']]
    report = '# Expanded component inventory\n\n'
    report += '| Measure | Count |\n|---|---:|\n'+''.join(f'| {k.replace("_", " ")} | {v} |\n' for k,v in counts.items())
    report += '\n## Simplification\n\nExact selections are deduplicated before component enrichment. Alias records preserve alternative slot annotations and original IDs. Separate occurrences of a repeated card are source examples, not additional canonical components. Nested subcomponents remain available because they can serve a different reuse boundary.\n\nCanonical semantic families simplify retrieval across decks. Source patterns retain geometry, cardinality and original colors; family membership does not establish structural interchangeability. Automatic structural fingerprints remain review candidates.\n\n'
    report += '| Family | Source patterns | Source compositions |\n|---|---:|---:|\n'
    report += ''.join(f"| {r['name']} | {len(r['source_pattern_ids'])} | {len(r['instances'])} |\n" for r in family_rows)
    report += '\n## Style contract\n\nSemantic styles are proposed role-to-token contracts, not applied recolorings. Observed OOXML colors remain separate; inheritance and text-run targeting still need resolution. Data, actor, phase and intensity colors must retain their meaning. Yellow is restricted to headline highlighter use. Every adapted variant still needs native rendering and fit QA.\n\n## Inputs\n\n'
    report += ''.join(f"- `{x['path']}` — `{x['sha256']}`\n" for x in lineage)
    content = [(a.out, json.dumps(seeds_doc, indent=2)+'\n'),
               (a.families, ''.join(json.dumps(r)+'\n' for r in family_rows+styles)), (a.report, report)]
    staged = []; installed = []
    try:
        for dest, body in content:
            dest.parent.mkdir(parents=True, exist_ok=True)
            fd, name = tempfile.mkstemp(prefix='.'+dest.name+'.', dir=dest.parent)
            staged.append(Path(name))
            with os.fdopen(fd, 'w', encoding='utf-8') as stream:
                stream.write(body); stream.flush(); os.fsync(stream.fileno())
        for path, (dest, _) in zip(staged, content):
            os.link(path, dest); installed.append(dest)
    except Exception:
        for path in installed:
            path.unlink(missing_ok=True)
        raise
    finally:
        for path in staged:
            path.unlink(missing_ok=True)
    print(json.dumps(dict(counts)))


if __name__ == '__main__':
    main()
