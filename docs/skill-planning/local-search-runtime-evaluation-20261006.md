# Local search runtime evaluation — 2026-10-06

## State

BM25 and structural constraints are implemented separately. Model-backed query
embedding, persisted corpus vectors and reciprocal rank fusion are not yet in
the product. This record captures a real isolated runtime experiment so the
next implementation can proceed without an operator model choice.

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
