#!/usr/bin/env python3
"""Create a portable local QA viewer containing reference/candidate PNGs."""
import base64
import json
from pathlib import Path

root = Path(__file__).resolve().parent.parent
base = root / 'samples/reconstruction'


def image(path):
    return 'data:image/png;base64,' + base64.b64encode(path.read_bytes()).decode()


final = json.loads((base / 'qa/final/report.json').read_text())
control = {r['slide']: r for r in json.loads((base / 'qa/delivery-control/report.json').read_text())}
data = []
for r in final:
    n, page = r['slide'], r['output_page']
    data.append({'name': f'UHG {n}', 'reference': image(base / f'reference/png/slide-{n:03d}.png'),
                 'candidate': image(base / f'qa/final/png/slide-{page:03d}.png'),
                 'difference': image(base / f'qa/final/{n:03d}-diff/difference.png'),
                 'description': f"Original full-deck reference: {r['exact_mismatch_count']:,} differing pixels; mean RGB error {r['mean_absolute_rgb_channel_error']:.6f}/255. Same-deck source control: {control[n]['exact_mismatch_count']:,} differing pixels."})
for n in [67, 5]:
    data.append({'name': f'UHG {n} — original content', 'reference': image(base / f'reference/png/slide-{n:03d}.png'),
                 'candidate': image(base / f'qa/{n:03d}-bound/candidate/slide-001.png'),
                 'difference': image(base / f'qa/{n:03d}-bound/diff/difference.png'),
                 'description': 'Original wording preserved. Separate one-slide native reconstruction; its unchanged-source control has zero differing pixels.'})
data.append({'name': 'EXPERIMENT — changed title', 'reference': image(base / 'qa/005-reflow-before/png/slide-001.png'),
             'candidate': image(base / 'qa/005-reflow-after/png/slide-001.png'),
             'description': 'EXPERIMENT ONLY: the title wording was intentionally changed to move “pivotal moment” to line two. Before = stale underline. After = underline repositioned using native character bounds and visible SVG bounds. This is not the exact-recreation baseline.'})
data.append({'name': 'EXPERIMENT — manual fallback', 'reference': image(base / 'reference/png/slide-005.png'),
             'candidate': image(base / 'qa/005-manual/png/slide-001.png'),
             'description': 'Original wording. Underline and arrow are deliberately off-slide to the right, with native text labels and speaker notes giving placement instructions. The assets remain in the PowerPoint file; this PNG intentionally omits them.'})
if (base / 'qa/006-highlight-after/png/slide-001.png').exists():
    data.append({'name': 'UHG 6 — original highlight',
                 'reference': image(base / 'reference/png/slide-006.png'),
                 'candidate': image(base / 'qa/006-bound/candidate/slide-001.png'),
                 'difference': image(base / 'qa/006-bound/diff/difference.png'),
                 'description': 'Original wording and original 24-point title. Source-derived native baseline with unchanged highlight asset, rotation, size and placement. Same-deck source control: zero differing pixels.'})
    data.append({'name': 'EXPERIMENT — highlight at 32 pt',
                 'reference': image(base / 'qa/006-highlight-before/png/slide-001.png'),
                 'candidate': image(base / 'qa/006-highlight-after/png/slide-001.png'),
                 'description': 'Wording is unchanged. Only the title font was increased from 24 to 32 points. Before = old highlight position/size, intentionally wrong. After = calculated position and width/height, retaining the 1-degree rotation, transparent PNG and layer behind the text.'})
html = '''<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>UHG reconstruction review</title><style>
*{box-sizing:border-box}body{font:16px system-ui;margin:0;background:#f3f5fa;color:#070154}header,main{max-width:1500px;margin:auto;padding:20px}h1{margin:0 0 8px;font-size:26px}p{line-height:1.5}select,button{font:inherit;padding:10px;border:1px solid #9aa8c0;border-radius:5px;background:white;color:#070154}button:focus-visible,select:focus-visible,input:focus-visible{outline:3px solid #2444ff}.controls{display:flex;gap:12px;flex-wrap:wrap;align-items:center}.canvas{position:relative;aspect-ratio:16/9;background:white;box-shadow:0 4px 20px #07015422;overflow:hidden}.canvas img{position:absolute;inset:0;width:100%;height:100%;object-fit:contain}.caption{font-size:14px;min-height:45px}.tag{font-weight:bold;color:#2444ff}input{min-width:220px}footer{font-size:13px;padding:20px;color:#53678e}#difference{display:none}a{color:#2444ff}
</style><header><h1>UHG reconstruction review</h1><p>Ten requested slides, bio 67, and two clearly labeled placement experiments. Native editable objects; source-derived scene structure. All ten delivered slides match their same-deck source controls at 1920×1080. Small differences against the original full-deck export remain documented.</p><div class="controls"><label>Slide <select id="slides"></select></label><button id="source">Before / reference</button><button id="rebuilt">After / reconstruction</button><label>Blend <input id="blend" type="range" min="0" max="100" value="100" aria-label="Reconstruction opacity"></label><button id="diff">Difference ×4</button></div></header>
<main><p id="caption" class="caption"></p><div class="canvas"><img id="reference" alt="Reference slide"><img id="candidate" alt="Reconstructed slide"><img id="difference" alt="Four times amplified pixel differences"></div><p id="mode" class="tag">After / reconstruction</p></main><footer>Local self-contained viewer. Differences are literal RGBA comparisons. Same-deck controls use source slide bodies with only slide-number fields converted to source-number text. Exact pixel equality is scoped to this native PowerPoint/PDFKit environment.</footer>
<script>const data=DATA;const $=id=>document.getElementById(id);data.forEach((d,i)=>{const o=document.createElement('option');o.value=i;o.textContent=d.name;$('slides').append(o)});function show(){const d=data[+$('slides').value];$('reference').src=d.reference;$('candidate').src=d.candidate;$('caption').textContent=d.description;$('difference').src=d.difference||'';$('diff').disabled=!d.difference;blend(100)}function blend(n){$('blend').value=n;$('candidate').style.opacity=n/100;$('difference').style.display='none';$('mode').textContent=n===0?'Before / reference':n===100?'After / reconstruction':`Blend: ${n}% reconstruction`} $('slides').onchange=show;$('source').onclick=()=>blend(0);$('rebuilt').onclick=()=>blend(100);$('blend').oninput=e=>blend(+e.target.value);$('diff').onclick=()=>{$('difference').style.display='block';$('mode').textContent='Absolute RGB differences amplified ×4 (black = unchanged)'};show();</script>'''
html = html.replace('DATA', json.dumps(data).replace('</', '<\\/'))
path = base / 'output/qa-review.html'
path.write_text(html)
print(f'{path}: {path.stat().st_size:,} bytes')
