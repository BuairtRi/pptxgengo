# PowerPoint automation and recovery

Read this when native rendering, GUI editing, AppleScript or document access
fails, or when resuming an interrupted native qualification. Continue source
editing and inspection when those operations remain available; record native
qualification as pending until the actual saved/exported bytes are verified.

## Identify the layer before acting

| Evidence | Meaning | Next bounded action |
|---|---|---|
| Harness tool rejection | Agent execution policy | Inspect effective permissions/writable roots and retain the exact rejected operation. Do not infer a reset from saved defaults. |
| Native-pipe / ScreenCaptureKit `-3811` / ComputerUse `-10005` | GUI transport or capture | Reload the supported computer-use entry point once and fetch fresh state. Check another app separately; Apple-event health is a separate capability. |
| `noWindowsAvailable` after closing a document | Window binding may be stale | Select PowerPoint again and obtain fresh state. |
| Fresh app binding reports `cgWindowNotFound` | Corresponding capture window unavailable | Record the exact error and owned-document state; probe Apple events separately. A fresh rebind has failed in one observed case despite successful native render/open; do not infer Automation denial or loop retries. |
| Multiple matching accessibility elements after duplication | Duplicate object names | Refresh Selection Pane state; select the copied root unambiguously and rename that root uniquely. Avoid a stale element refetch. |
| REPL `ReferenceError` after reset | Missing JavaScript binding | Reinitialize through the documented API and declare a fresh binding. |
| Explicit Automation `-1743` | This caller's Apple events denied | Report the caller and exact denial. Resolve the observed Automation setting before retrying that path. |
| System Events `-1728` assistive-access denial | This caller's assistive access denied | Do not extrapolate to PowerPoint Apple events. Use a permitted path, or resolve that caller's denied setting if required. |
| `prepare_identity` / caller filesystem denial | Caller cannot access the staged file | Verify the same caller can read/write the explicit folder. PowerPoint folder access remains unverified. |
| Owned working copy opens read-only and its file mode is read-only | Immutable-build permissions were copied | Close/reopen only the saved owned copy after making that working file writable; leave the baseline unchanged. See [project structure](project-structure.md). |
| Observed Grant File Access prompt | PowerPoint requests access to the displayed folder | Resolve that prompt for that folder; reuse the same staging directory. |
| `-9074`, `-1712`, open/identity timeout | Open/identity failed or did not answer | Inspect for a modal prompt if GUI works, then send individual bounded health queries. These codes alone do not establish permission denial. |
| Explicit identity ambiguity/change | Owned document not reliably identified | Stop the mutation; retain the exact task files. Resolve ownership before any close/save. |
| Named presentation reference returns `-1728` after Save As | The saved presentation may have a new name; the old reference can be stale | Resolve a fresh reference to the exact saved path and verify identity before reading, exporting or closing it. Check the output independently; this error does not establish that Save As failed. |
| Export/close failure or worker deadline | Native operation failed/timed out | Read structured evidence, preserve the failed attempt and use exact-task cleanup only. |

## CLI rendering workflow

Use a local, project-owned staging folder under the signed-in user's Documents.
Keep PowerPoint-visible task files in that stable folder; placing each task in a
new authorization folder can cause repeated folder grants. Give each attempt a
new output directory and retain failed attempts.

Opening or saving a PPTX through the GUI does not establish access to a sibling
PDF created through Apple events. If export displays **Grant File Access**,
check that the requested folder is the exact owned staging/output folder before
granting it. Reuse that folder and verify the exported bytes and page count.
Do not widen access to all Documents or reset privacy settings to resolve one
task-folder prompt.

```sh
pptxgengo design render-doctor --staging-dir "$HOME/Documents/pptxgengo-native-staging" --json
pptxgengo design render --pptx ./client-deck/builds/<build-id>/deck.pptx \
  --out ./client-deck/reviews/native-001 \
  --staging-dir "$HOME/Documents/pptxgengo-native-staging" --pdf --png
```

On failure, read `render-error.json` and `render-error.txt`. Older releases may
only emit the text file; structured phases and `last_helper_phase` are newer
development additions. Inspect installed help and retain the available evidence
without inventing missing phase observations. `phase` is an
explicit returned failure phase or `unknown`. `last_helper_phase`, when present,
is only the last observed helper entry before cancellation; it does not prove
completion, successful event delivery or the cause. `cause_confirmed: false`
means the classification is an observation, not a privacy diagnosis. Preserve
task ID, input hash, binary version/hash and exact staging path with the error.

Resolve an observed problem before retrying once with a new output directory.
A doctor pass does not guarantee the next render succeeds. If the same failure
recurs, stop retries on that path, record native qualification as pending and
continue independent CLI work. Windows COM has a separate diagnostic surface;
do not expect macOS phase markers there.

## GUI and Apple-event recovery

1. Retain the first error and last successful action. Fetch fresh UI state;
   reuse neither old app bindings nor element IDs after a reset or document close.
2. If GUI is unavailable, inspect Apple events independently when authorized.
   Use separate bounded version, presentation-count and exact-path queries;
   a compound query timeout does not establish that every operation fails.
3. A successful health query establishes delivery only. Opening, editing and
   saving require exact owned-document identity plus independent output checks.
   After Save As, discard references bound to the previous presentation name.
   Re-resolve the saved document and verify its exact path/file identity. The
   accessibility window URL can lag the displayed new name; neither alone is an
   ownership check.
4. Closing and reopening only the exact saved test presentation recovered one
   observed capture failure. Try this only when ownership and saved bytes are
   confirmed and the document has no unsaved work. Close by exact file identity,
   rebind the app and reopen that path. If identity or save state is unclear,
   retain the document and report what is unknown.
5. Do not quit/kill PowerPoint, close colleagues' presentations, reset TCC or
   change privacy settings as speculative recovery. Apply only the observed
   caller/folder resolution needed for the authorized workflow.

For native geometry editing, Escape out of text editing and select the entire
component through the Selection Pane or its filled surface. Verify the saved
transform actually changed: a drag can merely select text. GUI duplication can
retain child and root names; give the new root a unique name and supply an
explicit reconciliation copy map. Renaming alone does not qualify the copy.

## Save, reconcile and qualify

Save As to a separate owned working PPTX. Retain the untouched build and source
receipt. Reconcile that file, review supported proposals and manual findings,
adopt selected changes, rebuild, then compare native PowerPoint exports of the
edited and rebuilt decks. A valid ZIP or successful Save As is insufficient.
Byte-identical PNGs establish visual equivalence for that specimen only.
Do not suppress package/format findings because geometry or visuals match;
general imported objects and unsupported format edits still require review.

For a native chart, **Edit Data** opens its embedded Excel workbook. Save that
owned workbook explicitly, close it, then Save As the PowerPoint working copy.
Reopen or export the saved deck and check the displayed value. Reconciliation
must also verify that workbook cells and chart caches agree; a stale cache is a
refusal, even when the workbook contains the requested value. Follow the
[quantitative chart runbook](quantitative-charts.md) for explicit numeric adoption.

### Local GUI PDF fallback

When a bounded Apple-event export fails but the GUI still works, use **Save As
→ PDF → Best for printing** on the exact owned working deck. That option writes
locally; the electronic-distribution option uses Microsoft's online service.
Choose a fresh PDF filename, verify it exists with the expected page count,
and render its pages for review. Repeat for the rebuilt deck before comparing.
Keep failed CLI evidence and unconfirmed cleanup files. This fallback does not
repair Apple-event automation or create a CLI native-render receipt.

## Checkpoint before compaction or leaving the session

Retain a private checkpoint with project/slot, CLI version/hash, PowerPoint
version, caller terminal/agent host, effective permissions, baseline/edited
paths and hashes, current owned open file, receipt, last successful action,
first failure with timestamp/domain/code, probes/recovery tried, and next step.
Record whether save and cleanup were confirmed. Keep credentials and full
unredacted session logs out of the checkpoint.

On resume, read the checkpoint and current source state, inspect effective
permissions, reload computer-use documentation through its supported entry point
and obtain fresh app/UI state. Compare recorded effective states to investigate
permission drift; do not assume compaction changed permissions. Repository
maintainer evidence and unresolved failure history are in
`docs/powerpoint-automation-diagnostics.md`.
