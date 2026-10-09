# Diagnosing PowerPoint automation

PowerPoint rendering and actual native-editor qualification are separate from
headless PPTX generation, source editing, inspection and reconciliation. Keep
their evidence separate. A failed GUI connection does not establish that YAML
editing or Apple-event automation is unavailable.

For the agent's operational sequence, use the West Monroe skill's
[PowerPoint recovery reference](../skills/west-monroe-presentations/references/powerpoint-recovery.md).
This document owns the implementation details, evidence history and unresolved
reliability work. A workaround succeeding once is not a durable root-cause fix.

## Failure catalog and resolution status

| Observed failure | Recognize it | Recovery or implemented fix | Status and remaining limit |
|---|---|---|---|
| Agent policy rejects an operation | Exact tool rejection plus effective policy | Inspect effective permissions and writable roots; retain the rejection and selected launch/profile settings | No permission drift reproduced. Saved defaults differing from an injected policy are not evidence of a reset. |
| GUI transport cannot initialize | Native-pipe startup error | Reinitialize through the supported computer-use entry point once; obtain fresh state | Earlier occurrence; underlying transport cause unresolved. |
| GUI capture stops after successful use | ComputerUse `-10005`, ScreenCaptureKit `SCStreamErrorDomain -3811` | Record the failure; test other GUI surfaces and Apple events separately. An exact owned-document close/reopen recovered this specimen | Empirical recovery, not a framework fix or proof of a permission change. |
| Old app binding has no window | `noWindowsAvailable` after document close | Select the app again and acquire fresh state | Observed successful recovery; old element handles remain invalid. |
| GUI cannot find a corresponding captured window | Fresh ComputerUse `-10005: cgWindowNotFound` | Record this distinct window/capture failure and probe Apple events independently; retain exact owned-document state | New fresh bindings still failed while individual Apple events and an owned open succeeded. Root cause and durable GUI recovery remain unresolved. |
| GUI object lookup is ambiguous | Refetch reports multiple matching elements after duplication | Use fresh Selection Pane state, select the copied root unambiguously and give it a unique name | Repeatable operator pattern for tested copy; accessibility framework ambiguity itself is unchanged. |
| JavaScript handle is missing after REPL reset | `ReferenceError` or undeclared binding | Reinitialize documented API and declare the new app/tab binding | Agent execution-state error, distinct from app permissions. |
| Automation delivery denied | Explicit `-1743` / Automation denial for actual caller | Resolve the reported Automation permission for that caller; repeat the bounded probe | CLI reports this layer; permissions are owned by macOS/operator. |
| Assistive access denied | System Events says `osascript` not allowed assistive access (`-1728`) | Use a permitted path for the authorized action; resolve the actual caller's access only if that path is needed | Denied probe not resolved. It did not prove that Apple events or document access were unavailable. |
| Caller cannot stat/read staging file | `prepare_identity`, explicit filesystem failure | Check local read/write access using the same caller and stable staging folder | Fixed misleading PowerPoint-grant advice; underlying filesystem cause remains task-specific. |
| Visible PowerPoint folder prompt | The visibility monitor reports Grant File Access, or an observed prompt | Resolve only the displayed folder request, then rerun with that same explicit stable staging path and a new output directory | Folder-scoped staging implemented; an unobserved prompt must not be assumed. |
| Open/identity timeout or unexplained open failure | `open_document`, `identify_document`, `-1712`, `-9074`, `open_identity_timeout` | Split health queries; inspect for a modal prompt; retain the phase/error and exact owned paths | Basic events can succeed while open fails. `-9074` alone does not establish denial. |
| Worker deadline loses helper's returned error | `context deadline exceeded` / cancellation | Phase-entry observations stream to the parent before blocking calls; failure JSON retains `last_helper_phase` when observed | Implemented and subprocess-tested. Last entry is not the failing/completed phase; an unobserved phase stays unknown. |
| Export or close fails | `export_pdf` or `close_document` plus original error | Retain original phase/code; separately bounded cleanup targets only the exact owned task | Implemented identity guards; cleanup can remain unconfirmed and retain files. |
| Office Save As/copy normalization blocks reconciliation | Explicit style/structure/route refusal | Only the three qualified serialization equivalences are normalized; unsupported findings remain manual | Regression-tested and actual GUI copy/delete round trip passed. Arbitrary imports/general edits remain unqualified. |

## Implementation and acceptance boundaries

The CLI owns staging, identity checks, deadlines, structured evidence and narrow
reconciliation behavior. It does not fix ScreenCaptureKit, the computer-use
transport, macOS privacy settings, or Codex's effective execution policy.

The automated checks for this slice use fake helpers and owned subprocesses;
they do not manipulate PowerPoint or request privacy permissions. They verify
that fragmented phase markers are forwarded, unrelated stderr is not streamed,
successful worker JSON stays valid, cancellation preserves the last observed
entry, and unknown phases remain unknown. Actual PowerPoint lifecycle recovery
still needs repeated trials across document open/save/close, app rebind and
session resume. Windows remains a separate qualification lane.

## Classify the failing layer

| Layer | Evidence to retain | What it establishes |
|---|---|---|
| Agent execution policy | Effective sandbox/approval profile, writable roots, rejected tool call | Whether the harness permitted this action. A saved config is only a default. |
| GUI transport/capture | Exact computer-use error, underlying domain/code, timestamp | Whether the UI connection/capture worked. This does not establish Apple-event permission. |
| Apple events | Exact command and response, caller identity, error number | Operational delivery; `count presentations` does not establish document opening. |
| File staging | Exact folder and owned task file, local read/write checks | Caller file access; this does not establish PowerPoint file access. |
| Document opening/identity | Script phase, error number, exact owned file identity | Whether PowerPoint opened and identified this document. |
| Export/save | Script phase, output file/hash, native app observations | Whether the intended file was actually written. |
| Cleanup | Exact-task close result, retained task path | Whether only the owned task was closed; unknown cleanup retains evidence. |

The macOS export helper now prefixes errors with `native_phase` and the original
`native_error_code`. Phases are `prepare_identity`, `open_document`,
`identify_document`, `export_pdf` and `close_document`. Cleanup after an export
error preserves the original failing phase. The helper retains its exact-file
identity guards and never uses the active presentation as ownership evidence.

Failed render runs write `render-error.txt` and `render-error.json` in the selected
output directory. JSON includes the task, recording time, phase, layer, error
code, classification and full returned error. `cause_confirmed` remains false:
classification identifies the observed failure mechanism, not the underlying
cause. The timestamp is when evidence was recorded, not when a GUI event began.

Unknown failing phases stay `unknown`, including worker termination before a
helper returns. macOS helpers also emit `native_phase_entered` before blocking
calls; the worker forwards these observations on stderr independently of its
JSON stdout. Deadline errors retain them in `render-error.json` as
`last_helper_phase`. This is the last observed entry, not a completed phase or
proof that the subsequent Apple event was delivered. Cleanup calls do not stream
render progress; the parent snapshots primary-operation observations before its
separate cleanup worker, so cleanup errors cannot replace that entry evidence.
If cancellation precedes observation, the field is omitted. Windows helpers do
not yet emit these macOS stage markers. Doctor output remains independent:
an explicit denial can fail its access check, while an unexplained open failure
leaves permission status unknown.

Do not interpret `-9074` alone as permission denial. Do not interpret caller
filesystem `permission denied` as a PowerPoint folder grant problem. A visible
Grant File Access prompt or an explicit Automation denial is stronger evidence.
Resolve only an observed prompt for the actual caller/folder, then use a new
output directory and retain the failed attempt.

## Native qualification and safe recovery

Use a project-owned folder under Documents for manual fixtures. Pass the same
explicit `--staging-dir` to `render-doctor` and `render` when diagnosing exports.
Retain exact binaries, versions, commands, input hashes, receipts, original
saved files, and observations. A GUI Save As means saving a separate working
PPTX through PowerPoint, then independently verifying and reconciling those
bytes. XML-only edits do not qualify PowerPoint's save normalization.

After a failure, capture its layer before retrying. Inspect the UI for an actual
prompt if computer use works. Reconnect a failed GUI transport once and record
the result. Check Apple-event health separately when authorized; a successful
health response must not be described as successful document access. Do not
quit or kill PowerPoint, close unrelated presentations, reset privacy databases,
or change privacy settings as a speculative recovery step.

## Resume after compaction or a new session

Keep a private checkpoint containing:

- Current project/worktree, CLI binary hash, toolchain and PowerPoint version.
- Owned baseline/edited paths, build receipt and hashes, and active presentation.
- Last successful action and first failed action, timestamps and exact errors.
- Effective harness permission profile and caller terminal/agent host.
- Observed permission prompts and any user-authorized changes.
- GUI transport state and whether Apple-event health/document access were tested.
- Next intended operation and bounded recovery already attempted.

On resumption, reread this checkpoint, inspect current permissions and source
state, reload computer-use documentation through its supported entry point, and
obtain fresh UI state before using any old element IDs. A compaction summary
cannot establish that an old GUI handle is still valid or that permissions changed.

## Codex configuration evidence

On 2026-10-08, a selected-key local inspection found `codex-cli 0.161.0` and a
user config default of `sandbox_mode = "workspace-write"`; this session's
injected execution policy was `danger-full-access` with approval `never`.
No harness rejection was observed during this diagnostic slice. This difference
is evidence of different default and effective settings, not proof of a
mid-session reset. No auth files or full session transcripts were inspected.

Official [configuration guidance](https://learn.chatgpt.com/docs/config-file/config-basic)
describes CLI overrides, trusted project configuration, selected profiles and
user defaults in descending precedence; managed requirements can constrain them.
The [CLI documentation](https://learn.chatgpt.com/docs/codex/cli) describes
`/permissions` for inspecting the active sandbox and writable roots. The fetched
documentation does not establish that compaction changes permission settings.

For a reproducible drift report, record `/status` and `/permissions` at startup,
after a user change, after compaction/resume and immediately after a rejection.
Retain the launch flags, selected profile and only relevant config keys. Compare
effective states and the exact rejected operation; do not upload credentials,
customer decks or full unredacted session logs. Keep full-access selection an
explicit operator choice rather than changing global defaults as a workaround.

## Observations on 2026-10-08

Prior native-export probes failed during Apple-event `open` with `-9074`, while
basic automation answered. Launch Services returned success without the expected
owned presentation appearing within ten seconds. Computer use previously failed
with native-pipe startup. The root cause of those observations remains unconfirmed.

In a fresh attempt, computer use connected, opened an owned catalog deck and
completed GUI Save As; accessibility state identified the saved file's URL.
The next UI action failed with ComputerUse `-10005` and underlying ScreenCaptureKit
`SCStreamErrorDomain -3811`, reporting an audio/video capture stream failure.
Subsequent accessibility-state retrieval also failed. This is a separately
observed GUI transport/capture failure following a successful save; it does not
prove document access or Apple-event permission was revoked. Keep the final
qualification results and later recovery attempts in their private evidence
folders rather than treating these observations as whole-library qualification.

In that attempt, Finder computer use still connected, PowerPoint's sampled main
thread was in its idle event loop, a subsequent bounded PowerPoint health event
timed out, and System Events introspection explicitly reported that `osascript`
was not allowed assistive access (`-1728`). The latter proves that particular
assistive-access call was unavailable; it does not explain the capture stream
failure or establish a PowerPoint hang. The saved PPTX remained a valid ZIP.
Reconciliation reported known geometry/text as unchanged but also package and
non-text differences. Office normalization needs independent qualification;
a successful Save As alone does not complete a moved/edited/rebuilt round trip.

The successful GUI actions and subsequent capture failure occurred in the same
turn with no intervening compaction. This occurrence therefore does not require
compaction to explain the loss of computer-use access. Later, a bounded
AppleScript `activate` succeeded, followed by separately bounded version,
presentation count, names and exact full-path queries. An owned-slide shape
inspection also returned the expected object names and positions. An earlier
combined full-name query had timed out. These observations show that some
Apple-event operations recovered while PowerPoint computer-use capture remained
unavailable; they do not establish a permanently hung application or resolved
privacy settings. Splitting diagnostic events isolates failures more clearly
than sending one compound query and treating its timeout as total app failure.

The later owned-document AppleScript editing sequence moved `node-06` from top
234 to 228 points and resized `node-08` from width 144 to 132 points, then saved
through PowerPoint successfully. Reconciliation identified three native-only
changes (`node-06`, `node-08`, and their attached `service-edge` connector), with
39 geometry no-ops. PowerPoint also adjusted that connector's elbow guide from
65000 to 50000 and its endpoint Y from 252 to 246 points. A fresh export using
the diagnostic development CLI subsequently produced native PDF and PNG output
without changing the original source hash.

After closing only the exact owned saved presentation, computer use recovered
through its home/reopen sequence without quitting PowerPoint or changing privacy
settings. A GUI two-key Left nudge of the selected block, followed by Save,
changed `node06` X from 4381500 to 4361404 EMU in the saved file. Three geometry
changes were adopted and rebuilt; actual PowerPoint PNG exports of the edited
and rebuilt decks were byte-identical (SHA-256
`1bc26f9ae5c4baf4470db57568130c8d74d8d6a0bd18fa8aaf1a5f54615df2d2`).
This is specimen-specific GUI nudge, API move/resize, connector and rebuild
evidence. The drag attempt selected text without changing geometry and is excluded.
All 63 package/non-text review findings remain retained, not blanket-qualified.

A later GUI duplicate retained the source object's names and caused ambiguous
accessibility refetch. Coordinate selection and a unique root rename recovered
selection. Copy reconciliation then exposed narrowly classifiable Office
serialization differences. A fresh exact-pinned baseline with the narrow fix
subsequently passed actual GUI Save As/copy/delete, adoption and native
edited/rebuilt PNG comparison, while retaining 25 manual findings; see
[Save As equivalence](native-save-as-equivalence.md) for its precise scope.
Private `attempt-01/qualification-summary.json` records these phases and limits.
The capture stream's underlying failure and Windows qualification remain open.

After a later exact-owned document close, the old GUI binding returned
`ComputerUse -10005: noWindowsAvailable`; selecting PowerPoint again immediately
returned its home view. Rebind the app after closing a document when the prior
binding reports no window. This observed window-lifecycle failure does not
establish permission denial or require changing privacy settings.

During the next synthetic team-composition qualification, two native render
attempts succeeded. A fresh computer-use PowerPoint selection then returned
`ComputerUse -10005: cgWindowNotFound`. Separately bounded Apple events returned
PowerPoint version `16.113.4` and presentation count zero; an exact owned working
copy subsequently opened through AppleScript. Another fresh computer-use
selection still returned `cgWindowNotFound`. This is an additional observed
window/capture failure, distinct from the earlier ScreenCaptureKit `-3811` stream
failure. It does not establish denied Automation or a failed document open.
Native exports and individual Apple events remained usable in this attempt;
GUI recovery and the underlying window/capture cause are unresolved. Private
live qualification records retain the owned paths and cleanup observations.
