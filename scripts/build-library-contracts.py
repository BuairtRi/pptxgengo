#!/usr/bin/env python3
"""Build candidate semantic contracts; native qualification is a separate step.

This script only binds existing templates. It never treats the historical fixture
proof as blanket approval for arbitrary replacements, colors, or cardinalities.
"""
import hashlib
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEST = ROOT / 'library/contracts'
EXAMPLES = ROOT / 'library/contract-examples'

PATTERNS = [
    ('dense-argument', 'library/proposal/argument-template.json', 'dense-argument',
     'Dense opening argument with phrase highlight', 'Connect the decision to operating constraints, first-release proof, evidence, and qualifications with targeted emphasis.',
     ['argument', 'decision', 'evidence', 'qualification', 'highlight'], None),
    ('five-phase', 'library/layout-components/review.json', 'five-phase',
     'Five-phase delivery approach', 'Explain an ordered delivery approach with activities, timing, and evidence for each phase.',
     ['approach', 'phases', 'activities', 'evidence'], 'library/layout-components/proof.json'),
    ('phase-detail', 'library/visual-components/review.json', 'phase-detail-visual',
     'Detailed phase with deliverable previews', 'Explain activities, outcomes, variable scope, and actual attributed deliverable examples.',
     ['phase-detail', 'activities', 'deliverables', 'scope'], 'library/visual-components/proof.json'),
    ('workflow-matrix', 'library/layout-components/review.json', 'workflow-matrix',
     'Workflow constraint-response matrix', 'Map a service journey to constraints, delivery responses, and measures without dropping detail.',
     ['workflow', 'constraint', 'response', 'measure'], 'library/layout-components/proof.json'),
    ('delivery-team', 'library/diagram-components/dense-review.json', 'dense-delivery-relationships',
     'Delivery pods and client decision rights', 'Explain delivery accountability through 21 roles, three pods, client counterparts, and typed relationships.',
     ['team', 'roles', 'pods', 'decision-rights', 'staffing'], None),
    ('release-roadmap', 'library/visual-components/review.json', 'roadmap-interval-variant',
     'Release roadmap with editable ribbons', 'Show delivery sequencing, milestones, dependencies, and conditional extension windows.',
     ['roadmap', 'timeline', 'milestones', 'dependencies'], 'library/visual-components/proof.json'),
    ('layered-architecture', 'library/diagram-components/dense-review.json', 'dense-layered-architecture',
     'Layered platform architecture', 'Separate experience, orchestration, and platform responsibilities under shared governance and security controls.',
     ['architecture', 'layers', 'governance', 'security', 'dependency'], None),
    ('parallel-process', 'library/diagram-components/review.json', 'process-changed-content',
     'Parallel governed process paths', 'Compare two sequenced workflows with explicit decision and ownership steps.',
     ['process', 'flow', 'parallel-paths', 'decision'], None),
    ('dense-roster', 'library/visual-components/review.json', 'illustrative-19-tile-roster',
     'Core team and specialist roster', 'Distinguish a core delivery team from specialist capacity using role tiles and responsibility labels.',
     ['roster', 'roles', 'specialists', 'capacity'], 'library/visual-components/proof.json'),
    ('biography', 'library/visual-components/followup-review.json', 'synthetic-profile',
     'Detailed leader biography', 'Present background, work areas, and proposed responsibilities; fixture identity and experience are explicitly fictional.',
     ['biography', 'credentials', 'work-areas', 'responsibilities'], 'library/visual-components/followup-proof.json'),
    ('needs-response', 'library/visual-components/review.json', 'enablecomp-response-fixture',
     'Client needs and proposed responses', 'Pair five client needs with detailed responses, keeping implications and supporting evidence visible.',
     ['needs', 'response', 'evidence', 'outcome'], 'library/visual-components/proof.json'),
    ('phrase-underline', 'library/diagram-components/accent-review.json', 'underline-source-copy',
     'Assertion with phrase underline', 'Emphasize the exact takeaway phrase using native glyph bounds and preserved WM underline artwork.',
     ['assertion', 'takeaway', 'emphasis', 'underline'], None),
    ('phrase-highlight', 'library/diagram-components/accent-review.json', 'highlight-source-copy',
     'Assertion with phrase highlight', 'Highlight a selected phrase behind native editable text, using optical calibration and measured glyph bounds.',
     ['assertion', 'takeaway', 'emphasis', 'highlight'], None),
]


def sha(path):
    return hashlib.sha256((ROOT / path).read_bytes()).hexdigest()


def write(path, obj):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(obj, indent=2, ensure_ascii=False) + '\n')


def semantic_slots(slide):
    slots, values = [], {}
    # These are template-owned bindings, not caller-supplied coordinates.
    skip = {'footer-copy', 'page', 'page-number', 'source-artwork-note',
            'synthetic-ribbon', 'synthetic-note'}
    def add(name, pointer, value, role='content', **kw):
        if name in values:
            raise ValueError('duplicate semantic name: ' + name)
        slots.append(dict(name=name, role=role, pointer=pointer,
                          value_type='string', required=True,
                          max_chars=2048 if pointer in ('/role', '/takeaway') else max(128 if pointer == '/title' else 32, math.ceil(len(value)*1.5)), **kw))
        values[name] = value
    def walk(node, path='', context=''):
        if isinstance(node, dict):
            ident = str(node.get('id', ''))
            if ident in skip:
                return
            ctx = '/'.join(x for x in [context, ident] if x)
            for key, value in node.items():
                ptr = path + '/' + key
                if key in ('text', 'title', 'label') and isinstance(value, str) and value.strip():
                    # Static marker numbers remain a property of the layout.
                    if value.strip().isdigit():
                        continue
                    add((ctx or 'slide') + '/' + key, ptr, value,
                        'assertion_title' if ptr == '/title' else key)
                elif key == 'labels' and isinstance(value, list) and all(isinstance(v, str) for v in value):
                    name = (ctx or 'slide') + '/steps'
                    slots.append(dict(name=name, role='process_steps', pointer=ptr,
                                      value_type='string_array', required=True,
                                      min_items=4, max_items=6,
                                      max_chars=max(32, math.ceil(max(map(len,value))*1.5))))
                    values[name] = value
                elif key not in ('accents', 'artwork_arrows', 'notes', 'role', 'takeaway'):
                    walk(value, ptr, ctx)
        elif isinstance(node, list):
            for i, value in enumerate(node):
                # Paragraph/run and other anonymous children need stable semantic ordinals.
                ctx = context if isinstance(value, dict) and value.get('id') else f'{context}/item-{i+1}'
                walk(value, path + '/' + str(i), ctx)
    walk(slide)
    for key in ('role', 'takeaway'):
        if slide.get(key):
            add('slide/' + key, '/' + key, slide[key], key)
    for i, accent in enumerate(slide.get('accents', [])):
        # Exact phrase is an editorial slot, resolved to glyph geometry by the CLI.
        if accent.get('phrase'):
            add('emphasis/' + accent['id'] + '/phrase', f'/accents/{i}/phrase', accent['phrase'], 'emphasis_target')
    return slots, values


def assets(slide):
    out = {}
    def walk(node):
        if isinstance(node, dict):
            for prefix in ('', 'fallback_'):
                p, h = node.get(prefix+'asset_path'), node.get(prefix+'asset_sha256')
                if p and h:
                    if sha(p) != h:
                        raise ValueError('stale asset ' + p)
                    out[p] = dict(path=p, sha256=h, role='fallback' if prefix else 'artwork')
            for v in node.values(): walk(v)
        elif isinstance(node, list):
            for v in node: walk(v)
    walk(slide)
    return list(out.values())


def main():
    for ident, specpath, slideid, name, purpose, roles, proofpath in PATTERNS:
        spec = json.loads((ROOT / specpath).read_text())
        index, slide = next((i, s) for i, s in enumerate(spec['slides']) if s['id'] == slideid)
        slots, values = semantic_slots(slide)
        previews, evidence = [], []
        if proofpath:
            proof = json.loads((ROOT / proofpath).read_text())
            for item in proof.get('visual_review', []):
                if item.get('id') == slideid and item.get('accepted') and item.get('artifact'):
                    if sha(item['artifact']) != item['sha256']:
                        raise ValueError('stale preview ' + item['artifact'])
                    previews.append(dict(path=item['artifact'], sha256=item['sha256'],
                                         caption='Reviewed historical fixture; changed copy requires new native fit and visual review.', variant='fixture'))
            evidence.append(dict(path=proofpath, sha256=sha(proofpath), role='historical_fixture'))
        contract = dict(
            schema='pptxgengo.library-component.v1', id='wm/'+ident, version='0.1.0', kind='layout',
            name=name, purpose=purpose, content_roles=roles,
            source=dict(source_id='authored-fixture:'+slideid, path=specpath,
                        source_sha256=sha(specpath), slide=index+1),
            composition=dict(spec_path=specpath, spec_sha256=sha(specpath), slide_id=slideid, slots=slots,
                             narrative_bindings=dict(assertion_title={
                                 'dense-argument': '/canvas/5/text',
                                 'phase-detail': '/layouts/0/cells/0/blocks/2/text',
                                 'parallel-process': '/canvas/4/text',
                                 'phrase-underline': '/canvas/4/text',
                                 'phrase-highlight': '/canvas/4/text',
                             }.get(ident, '/title'), role='/role', takeaway='/takeaway')),
            assets=assets(slide),
            cardinality={s['name']: dict(min=s['min_items'], max=s['max_items']) for s in slots if s['value_type']=='string_array'},
            fit_envelope=dict(measured=['Historical fixture copy only; see pinned proof.'] if proofpath else [],
                              unsupported=['Arbitrary geometry edits.', 'Automatic font shrinking or content truncation.',
                                           'Character caps are editorial input guardrails, not measured fit guarantees.',
                                           'New content is unqualified until native measurement, final verification, and visual review.'],
                              font_policy='no_silent_shrink'),
            transforms=dict(translation='unsupported', resize='unsupported', rotation='unsupported'),
            style_variants=[dict(id='source', semantic_profile='wm.fixture-original', tokens={}, qualified=bool(previews))],
            preference=dict(value='unreviewed', note='No user preference for this complete layout; do not infer preference from related components.'),
            qualification=dict(state='measured_fixture' if previews else 'reviewed', evidence=evidence,
                               failures=['Candidate semantic bindings have not completed changed-content end-to-end qualification.']),
            previews=previews,
            provenance=dict(created_from=[specpath, slide.get('notes','')], review_history=[
                'Candidate contract generated from existing authored fixture. Original sample provenance remains in fixture notes and referenced proof.',
                'This does not promote the source deck, unrelated layouts, or unrestricted slot values.']),
        )
        for i, element in enumerate(slide.get('canvas', [])):
            if element.get('id') in ('page', 'page-number') and element.get('kind') == 'text':
                contract['composition']['pagination_binding'] = f'/canvas/{i}/text'
                break
        write(DEST/(ident+'.json'), contract)
        write(EXAMPLES/(ident+'-values.json'), dict(slots=values, style_variant='source'))
        print(ident, len(slots), 'slots', len(previews), 'previews')
    write(EXAMPLES/'assembly.json', dict(
        schema='pptxgengo.library-assembly.v1',
        slides=[dict(contract_id='wm/'+p[0], values_path=p[0]+'-values.json', slide_id=p[0]) for p in PATTERNS]))

if __name__ == '__main__':
    main()
