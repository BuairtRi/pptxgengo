# Diagnosing PowerPoint automation

PowerPoint rendering and actual native-editor qualification are separate from
headless PPTX generation, source editing, inspection and reconciliation. Keep
their evidence separate. A failed GUI connection does not establish that YAML
editing or Apple-event automation is unavailable.

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

Unknown phases stay `unknown`, including worker termination before a helper
returns. The JSON does not invent a last completed phase. Windows helpers do
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
