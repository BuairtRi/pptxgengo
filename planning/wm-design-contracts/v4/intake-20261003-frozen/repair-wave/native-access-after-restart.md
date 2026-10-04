# Native access after user restart — 2026-10-03

Computer use now works: `cua.getApp("Microsoft PowerPoint")` returned the
accessibility tree for the saved `WMDS-intake-review.pptx`, currently on slide
27 of 50. The prior computer-use startup failure is resolved in this session.

Shell automation remains blocked. The read-only diagnostic still reports
`MACH_PORT_DEAD` as the task bootstrap endpoint, Launch Services invalid
destination port, and PowerPoint count failure −10827. Tool subprocesses still
descend from Codex PIDs 43692 and 13364, as before the restart. The restart did
not replace this shell execution ancestry.

See `native-access-after-restart.json` for the new diagnostic. Native visual
review can now use computer access. The shell capture workflow is not yet
restored. No application edits, exports, process termination, or permission
changes were performed in this check.

Remaining slide feedback is recorded in `maturity-feedback-handoff.md` and
`venn-feedback-handoff.md`. The current decks predate the small Venn badge fix;
maturity centering and label clearance still require implementation and review.

## In-session recovery attempt

The user asked whether the remaining shell connection could be repaired from
this session. Existing evidence already includes unsuccessful `launchctl asuser`
and `bsexec` probes. A supported computer-use Terminal bridge was attempted:
`cua.getApp("Terminal")` was rejected with "Computer Use is not allowed to use
the app 'com.apple.Terminal' for safety reasons." No command was typed and no
alternative terminal app was used to bypass that restriction.

No supported in-place repair of the inherited shell bootstrap endpoint has
been established. PowerPoint GUI review remains available. Replacing the
background command runtime would interrupt this session and needs an external
restart; success is not guaranteed. No daemon was killed or configuration
changed during this attempt.

## Desktop update timing

At 13:31 PDT on October 3, process start times confirmed the user's report:
the desktop app PID 87219 started October 3 at 11:27 PDT, approximately two
hours earlier. The daemon PID 13364 started September 30 at 15:18 PDT; this
session's app-server PID 43692 started September 30 at 16:23 PDT. Both use
the `0.159.3` daemon release. The desktop app is version 26.930.31730, build
12947. Thus restarting/updating the desktop did not replace the background
runtime serving these commands. This supports the persistence diagnosis but
does not establish that the update caused the invalid bootstrap endpoint.
The updated desktop bundles codex-cli 0.160.0, while the surviving command
runtime is 0.159.3, confirmed by each executable's `--version` output.
