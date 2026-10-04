# Local macOS agent-session recovery notes

These steps apply only to the Codex installation inspected on 2026-10-04. They
are observations from one machine, not general `pptxgengo` or PowerPoint
troubleshooting instructions.

## Observed local state

The agent's native events failed with −10827/−600. Read-only probes showed an
inherited task bootstrap port that was a dead Mach port. Its background daemon
started September 30 and ran Codex 0.159.3; the desktop app bundled 0.160.0. A
desktop restart previously left the background daemon running. The version
mismatch was observed, not established as the cause.

## Suggested local recovery

After all agent tasks are checkpointed, use a normal macOS Terminal to update
the daemon from the app's bundled CLI and restart it. These commands are
specific to the installed app path and interrupt agent sessions:

```bash
"/Applications/ChatGPT.app/Contents/Resources/codex-cli/CodexCLI.app/Contents/MacOS/codex" app-server daemon update --from-cli
"/Applications/ChatGPT.app/Contents/Resources/codex-cli/CodexCLI.app/Contents/MacOS/codex" app-server daemon restart
```

Reopen ChatGPT/Codex, resume the thread and run `render-doctor` in the fresh
agent session. Check for successful operational presentation count and a usable
bootstrap port before a one-slide export. This recovery procedure has not been
verified. An explicit Automation denial (−1743) is a separate permission
check; the observed dispatch errors do not establish such a denial. The
inspected official OpenAI documentation did not provide a specific fix for
this dead-port failure; these steps follow local probes and installed CLI help.
