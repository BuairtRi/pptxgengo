# Native capture diagnosis — 2026-10-01

Fresh native capture was retried after the v3 height calibration. It still
fails from the agent's sandboxed shell. The failure is localized to access to
the host's application registry and live GUI processes.

## Observations

| Check | Result |
| --- | --- |
| Plain `osascript` returning a string | Succeeds |
| PowerPoint bundle metadata and scripting dictionary | Readable; version 16.113.3, build 16.113.26092714 |
| PowerPoint executable | Present, executable, signed, contains arm64/x86_64 architectures |
| Native UI connection | PowerPoint is running with an open presentation |
| AppleScript application-name lookup | `com.apple.hiservices-xpcservice` connection invalid, followed by dictionary-dependent compilation failure `-2741` |
| AppleScript bundle-ID lookup | Cannot get application ID, `-1728` |
| Explicit bundle path, reading open presentations | Execution error `-10827` |
| Explicit Finder path, reading startup disk | Same execution error `-10827` |
| `NSWorkspace` lookup for PowerPoint's bundle ID | `nil` |
| `NSWorkspace.runningApplications` | Zero visible applications |
| Read-only LaunchServices database dump | Zero application bundle records; database path unavailable |
| Reading sandbox logs / process inventory | `log: Cannot run while sandboxed`; `ps` reports operation not permitted |

The local SDK identifies `-10827` as `kLSNoExecutableErr`. PowerPoint's
executable is present and its GUI is running, so that error does not establish
that the installation lacks an executable. In this execution context, the
application registry cannot identify the live process correctly. Finder
exhibits the same behavior, providing a control outside PowerPoint.

An explicit-path `get version` returns the bundle version. That result is
insufficient to establish live Apple-event automation: the reads of
`presentations` still fail.

## Diagnosis and remaining uncertainty

The agent's sandboxed shell cannot see the host's LaunchServices application
registration or running GUI app inventory. The native adapter fails while
locating its target app, before it can read PowerPoint character bounds.
The UI connector has a separate access path and can see PowerPoint.

The specific denying sandbox rule cannot be inspected from this shell.
No Apple-event authorization denial (`-1743`) was observed, so changing macOS
Automation permissions alone has not been established as a remedy.
No system permissions, application registration, or user documents were
changed during diagnosis. No new character-bound capture succeeded.

## Confirmed boundary and next capture

The user subsequently confirmed that the presentation-name query returned
names in their normal Terminal session. This confirms the execution-context
boundary: native automation is reachable from that session.

The successful read-only check was:

```sh
osascript -e 'tell application "/Applications/Microsoft PowerPoint.app" to get name of every presentation'
```

The first local capture packet at `samples/font-height-native-capture-20261001`
reached PowerPoint and saved all 276 original reference rows. Validation then
failed on case 55. Replay of every saved row through the production validator
found exactly three text mismatches, on slides 55, 122 and 189; the other 273
passed.

Those three cases end in a newline. The composer let the upstream writer
store their line breaks inside an `<a:t>` text run. PowerPoint rendered the
breaks but omitted them from its native text content. The composer now emits
explicit paragraphs for plain-text hard breaks, including a trailing empty
paragraph. Source inspection confirms only those three reference text bodies
change, and Go predictions are unchanged.

The corrected packet is at `samples/font-height-native-capture-fix-20261001`.
Its `capture.sh` first captures the three corrected cases, then proceeds to the
276 original and 117 additional cases if strict validation passes. It includes
a rebuilt CLI and uniquely named probe decks. Fresh validated evidence remains
pending until that local script completes.

The prior failed payload lacks font-file/environment provenance because the
old CLI omitted that metadata on validation failure. It remains failed raw
evidence. The CLI now retains environment, font hashes, adapter, binary, spec,
manifest and timestamp provenance on such failures, without promoting failed
observations into accepted evidence.

Reference capture requires a process with access to the live GUI application.
The Go `--engine go` measurement and generation path remains independent of
this native reference-capture access.
