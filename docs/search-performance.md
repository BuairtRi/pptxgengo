# Search performance measurements

The developer command `scripts/cmd/search-benchmark` measures the production
`LibraryIndex` API for keyword, semantic or hybrid template retrieval. It runs
offline against an existing verified SQLite index and, when needed, the pinned
model and complete vector snapshot. This source tool is not part of the
released v4.1.0 CLI.

## Reproduce

From the repository root, choose a new report path in an existing directory:

```sh
go run ./scripts/cmd/search-benchmark \
  --index /path/catalog.sqlite --mode keyword \
  --query "modernization economics" --warm-runs 5 \
  --out /path/new-keyword-measurement.json

go run ./scripts/cmd/search-benchmark \
  --index /path/catalog.sqlite --mode hybrid \
  --embeddings /path/embeddings.json --model-dir /path/verified-model \
  --query "modernization economics" --warm-runs 5 \
  --out /path/new-hybrid-measurement.json
```

Use `--mode semantic` for vector-only ranking. Keyword measurements must omit
model options. There are 1–10 repeated queries, with three by default. All finds
use the template hard filter and limit ten. Downloads, snapshot creation and
native PowerPoint are excluded. Missing/stale vectors, invalid model pins,
retrieval fallback, changed results between queries, replaced physical inputs
and existing report destinations fail instead of producing comparable evidence.

For the standard CI corpus:

```sh
PPTXGENGO_EMBED_MODEL_DIR=/path/verified-model \
PPTXGENGO_SEARCH_BENCH_OUT=/path/new-measurement-directory \
  make test-search-performance
```

This explicitly prepares a 26-entity corpus from pinned V5 source: four each
of assets, components, composites, frames and primitives, plus six templates.
It builds complete embeddings and launches a separate freshly built benchmark
process for each mode. Preparation and parent-process memory are excluded.
`PPTXGENGO_EMBED_FULL_LIBRARY=1` selects the full V5 fixture instead; it does
not select the full V11 library. Use the standalone command with the maintained
V11 index/snapshot to measure V11. Ordinary headless Make targets clear the
benchmark output opt-in.

## What the report means

- Verified index open is timed separately from the first find. First find is
  the first API call in that benchmark process. Repeated finds reuse the open
  index; production retrieval currently loads the model on each call.
- Times exclude process launch, argument parsing, model/vector preparation,
  physical file hashing and JSON output. They are not end-to-end CLI timings.
  OS file cache is uncontrolled and is not flushed; initial footprint hashing
  and previous preparation can warm it. These are not disk-cold startup figures.
- Every repeated duration is retained. Median averages the two middle values
  for an even sample count. P95 uses nearest rank; with 1–10 samples it is the
  observed maximum, not a statistically established service percentile.
- Peak resident memory is the greatest observed OS high-water-mark sample after
  the first/last timed queries, including startup and initial hashing. Both raw
  samples are retained. Darwin uses `getrusage` bytes, Windows uses
  `K32GetProcessMemoryInfo.PeakWorkingSetSize`, and Linux uses
  `/proc/self/status`'s `VmHWM` KiB converted to bytes.
- Linux `getrusage` was deliberately replaced: it retains the pre-exec
  address-space peak, which can include the model-preparing Go parent's memory.
  A touched-memory/subprocess regression distinguishes this from the current
  executable's address space. Missing OS counters fail the measurement.
  Kernel proc RSS accounting is asynchronous/approximate: Linux raw VmHWM reads
  can decrease slightly. Their observed maximum preserves the measured samples
  without asserting exact page accounting or a strictly monotonic kernel counter.
- Go live heap, reserved heap and total runtime reservations are separate
  samples, not interchangeable with resident memory or cumulative allocations.
  Actual Go GC percentage and memory limit are recorded, with no forced GC.
- Reports bind exact physical index/snapshot byte counts and hashes, source
  revision, projection/retrieval hashes, entity counts, complete vector coverage,
  compiled model identity, OS/architecture, Go version and process CPU settings.
  An optional full `--source-commit` SHA is a caller declaration, not an
  authenticated attestation. Each platform's independently generated SQLite
  and snapshot bytes can differ; source/model pins establish comparability.

The pinned runtime's three inference artifacts total **91,102,969 bytes**;
the optional distributable ZIP also contains license/attribution/evidence.
No weights are added to CLI archives.

Reports contain the supplied query and matched template IDs. Keep measurements
using private/customer inputs within the private project. CI uses a generic
query and synthetic/pinned source fixtures; its diagnostic artifacts contain
neither customer decks nor model weights and expire after 14 days.

## Observed full V11 sample — 2026-10-07

Apple M5 Max, macOS 27.0.1 arm64, Go 1.27.1, default GC 100 and no configured
Go memory limit. Separate processes used the same pre-existing 1,969-entity
V11 index (1,204 assets, 38 components, 29 composites, 36 frames, 13 primitives,
649 templates), query `modernization economics`, template filter, and five
repeats per mode. Filesystem cache/other workstation activity were uncontrolled.
These locally measured source changes preceded final submission; no source SHA
was falsely attached to the dirty checkout.

| Mode | Verified open ms | First find ms | Repeated median ms | First peak MiB | Final peak MiB |
| --- | ---: | ---: | ---: | ---: | ---: |
| Keyword | 487.7 | 104.4 | 105.1 | 185.0 | 187.4 |
| Semantic | 474.4 | 482.9 | 470.7 | 369.5 | 774.9 |
| Hybrid | 461.5 | 475.4 | 468.7 | 368.3 | 772.3 |

Index: 72,069,120 bytes,
`23a9cf2162c180ecba6f512276840fad67543d168a7a26dfe83bdfd99958670e`.
Vector snapshot: 9,775,374 bytes,
`f4609904b54060d1c6be6127773532830443bbdf677e4c8b89aeaf803d441f9e`.
This sample reuses the maintained authoring snapshot with full vector coverage.

The local 26-entity sample is a different workload: hybrid first 90.1 ms,
repeated median 82.1 ms, first peak 383.1 MiB and final peak 710.8 MiB.
Do not extrapolate its timings to the full catalog.

## Cross-platform CI and remaining qualification

GitLab Kubernetes Linux and GitHub macOS/Windows pinned model lanes retain
`keyword.json`, `semantic.json` and `hybrid.json`. They check actual mode,
platform, source fingerprints, counts, complete vector coverage, repeated
sample count and nonzero memory. The portable Windows lane also exercises the
native memory API and the no-model keyword/report regressions.

Preliminary head `54851875249b3757d650d0e2abcc38d9c7e10abd` passed pinned
model jobs on Linux amd64 (private GitLab pipeline 21215), macOS arm64 and
Windows amd64 (GitHub run 37602299427). Independently downloaded reports confirm
26 actual entities and the same runtime model identity. Hybrid first/repeated
median were 188.0/158.0 ms on Linux, 135.1/139.4 ms on hosted Mac and
233.1/207.5 ms on Windows. That preliminary Linux memory result used the
pre-exec-sensitive counter and is excluded from memory conclusions.
PR #15's corrected exact head `8cec1d7b` passed full slot preflight, all ten
GitHub checks and all nine GitLab pipeline 21225 jobs. It merged as `bd63c193`.
Nine independently downloaded reports verified counts/source SHA, full vectors,
actual GC settings and retained raw first/last memory samples with their observed
maximum. Both complete hosted race runs passed in 17m27s/16m24s under the
20-minute job ceiling; per-package ceilings/assertions were preserved.

For that 26-entity workload:

| Native platform | Hybrid first ms | Repeated median ms | Observed peak MiB |
| --- | ---: | ---: | ---: |
| Kubernetes Linux amd64 | 221.8 | 216.4 | 659.3 |
| Hosted macOS arm64 | 225.2 | 242.3 | 672.7 |
| Hosted Windows amd64 | 277.6 | 264.3 | 701.4 |

The installer qualification work also adds pinned model/performance lanes on
Intel Mac and Windows ARM64 hosted runners and the home lab's tagged ARM64
Kubernetes pool. Each requires the declared Go host architecture. Those new
native lanes must pass and their reports must be inspected before qualification;
the original source release's six-target cross-build is separate evidence.

Cross-compilation is not actual performance qualification of macOS amd64,
Windows arm64 or Linux arm64. Full-library Windows timings, controlled cache
experiments, measured process-launch costs, agreed performance/memory budgets,
operator relevance judgments and native content-fit/visual acceptance remain
open. Retrieval reports explicitly declare performance targets and relevance
acceptance unrecorded.

Counter references: [Linux kernel proc documentation](https://www.kernel.org/doc/html/latest/filesystems/proc.html),
[Linux exec source](https://github.com/torvalds/linux/blob/master/fs/exec.c),
[Windows memory API](https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getprocessmemoryinfo).
Current Darwin units follow the installed macOS `getrusage(2)` manual.
