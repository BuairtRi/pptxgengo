# Local search runtime evaluation — 2026-10-06

## State

BM25 and structural constraints are implemented separately. The managed
`pptx-local-vectors` slot now implements the optional offline query model,
source-bound vector snapshots, exact cosine scan and reciprocal rank fusion.
Metadata remains the default. The candidate is explicit and optional; hosted
CI and engineering relevance evidence are required before merge, with operator
and native review still required for content selection.

## Candidate and distribution decision

Use the pure Go GoMLX backend as the first runtime candidate, with
`github.com/gomlx/onnx-gomlx v0.5.13`, `gomlx v0.28.16`, and
`compute v0.1.14`. Import only the Go backend explicitly; avoid default XLA/ONNX
Runtime imports, automatic plugins and hosted embedding services. The actual
experiment compiled with `CGO_ENABLED=0`, executed on macOS ARM64, and
cross-built for Darwin/Linux/Windows on AMD64 and ARM64. Actual Windows execution
and supported-platform latency remain required before adoption.

Evaluate `sentence-transformers/all-MiniLM-L6-v2`, 384 dimensions, at immutable
Hugging Face revision `1110a243fdf4706b3f48f1d95db1a4f5529b4d41`.
Model and runtime are Apache 2.0 according to their primary repositories:
[model card](https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/tree/1110a243fdf4706b3f48f1d95db1a4f5529b4d41),
[ONNX-GoMLX](https://github.com/gomlx/onnx-gomlx),
[GoMLX](https://github.com/gomlx/gomlx).

Prefer a separate optional offline search package: model files exceed 90 MB and
are unnecessary for lexical discovery. Normal queries must not download files.
Publish any distributable package privately in GitLab with hashes and signed
release evidence. No model or executable experiment bytes are checked into Git.

## Exact experiment inputs

| Artifact at the pinned revision | Bytes | SHA256 |
| --- | ---: | --- |
| `onnx/model.onnx` | 90,405,214 | `6fd5d72fe4589f189f8ebc006442dbb529bb7ce38f8082112682524616046452` |
| `vocab.txt` | 231,508 | `07eced375cec144d27c900241f3e339478dec958f92fddbc551f295c992038a3` |
| `tokenizer.json` | 466,247 | `be50c3628f2bf5bb5e3a7f17b1f74611b2561a3a27eeab05e5aa30f411572037` |

These are actual downloaded hashes; TLS verification was retained. They should
be hard pins in any maintenance/package producer, with model identity, license
and source revision retained in the optional package manifest.

## Execution evidence and limits

On this Apple M5 Max, Go 1.27.1, using four backend workers, the full model loaded
in a fresh process in about 27–33 ms with cached filesystem pages. A fixed
seven-token input produced a finite `Float32[1,7,384]` output in about 30–34 ms
for its first graph call and 18–20 ms for repeated calls. Padding to 256 positions
produced `Float32[1,256,384]` in about 547–549 ms. The process that exercised both
shapes reached approximately 477 MB maximum resident memory (`time -l`).

This proves model parsing, graph execution, repeated execution and a no-CGO
six-target build matrix. It does not prove sentence pooling, tokenizer equivalence,
retrieval relevance, embedding snapshot freshness, target-system latency or
Windows native execution. Input tokens were hardcoded from the upstream example.
No Python or hosted service was involved in inference. The development download
used a local script and curl; production offline loading must be implemented in Go.

## Next implementation contract

The six items below describe the implemented contract; release qualification
and actual target-platform evidence must still be recorded before distribution.

1. Implement pinned uncased BERT WordPiece tokenization in Go, with golden token
   IDs covering punctuation, accents, Chinese segmentation, unknown words and
   maximum length. Mean-pool by attention mask and L2-normalize, rejecting invalid
   dimensions, nonfinite values and zero vectors.
2. Version semantic text preparation independently; prioritize names, purpose,
   relationships and aliases before truncation. Exclude synthetic example copy.
3. Persist vectors with model/revision/artifact hashes, dimensions, preparation
   version, complete source/entity projection and per-entity prepared-text hashes.
   Reject stale or incomplete snapshots; use a simple exact cosine scan initially.
4. Add explicit keyword, semantic and hybrid modes. Fuse separate lexical and
   vector ranks with a documented reciprocal rank policy, preserving explicit
   filters, source count/structure caveats and discovery-only fit status.
5. Report lexical fallback when an optional model is missing. A semantic request
   must never silently become metadata search. Compare the judged query set,
   exact IDs and synonyms, then collect real supported-platform latency/memory.
6. Update dependency/license notices and run the existing security and six-target
   build/signing gates before distributing the runtime or optional model package.

## Integrated source implementation evidence

The pinned Go tokenizer matched 15 independent `tokenizers 0.23.2` cases
(including accents, CJK, Unicode separators, special tokens, unknown words and
256-token truncation). Ten mean-pooled normalized vectors agreed with
`onnxruntime 1.30.0` within `2e-5` maximum absolute error and cosine at least
`0.99999`. The small checked-in oracle fixture contains token IDs/vectors only;
the development Python oracle is not a product dependency.

Explicit TLS maintenance download, offline package copy, non-replacing
publication, pinned load and complete V5 generation/query passed on macOS ARM64.
The V5 fixture contained 1,897 entities; the final V11 projection contained
1,969. V11 vector generation took 106.36 seconds and reached 1,245,364,224 bytes
maximum RSS while another full-model test and builds shared the machine. These
are development measurements, not an isolated latency target. Snapshot size was
9775374 bytes.

Engineering/source-purpose judgments, acceptable candidate position in first ten:

| Query | Keyword | Semantic | Hybrid |
| --- | ---: | ---: | ---: |
| cards/3 | 1 | miss | 1 |
| interview lists | 2 | miss | 9 |
| practices heat maps | 2 | 1 | 1 |
| modernization economics | 1 | 1 | 1 |
| roadmap | 2 | 1 | 1 |
| pillars | 1 | 2 | 1 |
| make or buy technology investment tradeoffs | 1 | 10 | 3 |
| stakeholder conversations summarized in bullet points | miss | miss | 8 |
| sequence of delivery stages over time | miss | 7 | miss |
| three independent messages side by side | miss | 6 | 4 |
| compare business capability strengths in a color coded matrix | miss | miss | miss |
| organize strategic priorities into pillars | 2 | 3 | 2 |

Baseline coverage was 6/6 keyword, 4/6 semantic and 6/6 hybrid; synonym
coverage was 2/6, 4/6 and 4/6 respectively. Pure semantic ranking does not
prioritize canonical keys; hybrid retains exact-identity precedence. Do not
claim universal improvement: hybrid missed the delivery-stage and capability
matrix judgments. This small set is engineering evidence, not operator approval,
visual review or measured content fit. No ranking weights were tuned on it.

Each query ran in a fresh process with cached filesystem pages, including full
index verification, snapshot verification and model load where applicable.
Median elapsed seconds were keyword 0.768, semantic 1.157, hybrid 1.161. These include startup, not just inference.

A separate fresh hybrid process took 1.42 seconds and reached 463,667,200 bytes
maximum RSS (`time -l`). The benchmark query was modernization economics with
templates only. Cache, hardware, corpus size and model load affect these figures.

All six OS/architecture distribution builds compiled with `CGO_ENABLED=0`.
The new dependency inventory was refreshed from linked checksum-verified module
license/notice texts; local reachable-symbol vulnerability findings were zero.
Linux/Mac/Windows CI now includes explicit pinned download, golden inference
and complete source-bound snapshot/query checks. Release resource/build jobs
require the Linux offline-model gate. Hosted qualification remains pending
until those new jobs pass; native PowerPoint review and signed distribution of
the optional model package remain outstanding.
