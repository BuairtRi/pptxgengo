# Native access diagnosis — 2026-10-03

Status: **the current agent tool process has no usable inherited macOS bootstrap endpoint**. Native PowerPoint review remains pending. The exact runtime component that supplied that endpoint has not been identified.

## Evidence

The read-only comparison script is `native-access-diagnose.sh`; this tool's output is `native-access-diagnosis.json`.

| Observation | Current tool session |
| --- | --- |
| Effective UID | 502, matching PowerPoint and console UID |
| PowerPoint | PID 53736, running executable at `/Applications/Microsoft PowerPoint.app/Contents/MacOS/Microsoft PowerPoint` |
| Tool ancestry | shell → Codex app-server PID 43692 → Codex daemon PID 13364, whose parent is PID 1 |
| Daemon binary | `~/.codex/packages/app-server-daemon/releases/0.159.3-aarch64-apple-darwin/bin/codex` |
| `osascript -e '1 + 1'` | Returns 2 |
| Explicit-path PowerPoint count | Fails with -10827 |
| Explicit-path Finder window count | Fails with -10827 |
| PowerPoint bundle version lookup | Returns 16.113.3; this alone does not establish live Apple Event access |
| `launchctl manageruid` / `managername` | Cannot get manager UID/name |
| `launchctl print gui/502`, `user/502`, `pid/self` | `141: Reentrancy avoided` |
| `launchctl asuser 502` count query | Cannot get user context, error 141 |
| `launchctl bsexec 53736` count query | Still fails with -10827 |
| Global libSystem `bootstrap_port` | 0 |
| `task_get_special_port(self, TASK_BOOTSTRAP_PORT)` | Succeeds, returns **4294967295 / 0xffffffff / MACH_PORT_DEAD** |
| Direct Launch Services lookup through task bootstrap port | `268435459: (ipc/send) invalid destination port` |

The constants were checked in the installed macOS SDK: `TASK_BOOTSTRAP_PORT` is 4 (`mach/task_special_ports.h:78`), and `MACH_PORT_DEAD` is the all-bits-one port name (`mach/port.h:173`). The direct Mach query distinguishes a missing service from an unusable bootstrap destination.

The user separately reported the same basic PowerPoint presentation-count operation returning **4** in Terminal. That user-reported success was not executed by this tool. A same-script Terminal comparison is the remaining diagnostic needed to pin the difference precisely.

Parent-agent observations also include direct PID-addressed Apple Event failure -600, Launch Services `open -a` failure -10827, CUA native-app startup failure in the Sky service, and no document-connector sessions. No ordinary Automation denial (-1743) was observed. Secondary permissions may need checking after native transport is restored; the present evidence does not justify changing TCC or rebuilding Launch Services registration.

## What this establishes

Filesystem visibility, numeric UID and a process listed by `ps` are insufficient for this process to reach the GUI application's native services. The inherited task bootstrap endpoint is dead, so Launch Services lookup through that endpoint fails before an ordinary application operation can succeed. This fits the cross-application failures and explains why retrying deck or presentation-name logic cannot repair this tool session.

The detached daemon ancestry is an observation, not proof that daemon mode caused the dead endpoint. It remains unclear whether the launcher, app-server worker or another runtime boundary supplied it. This session already reports unrestricted filesystem/network access.

## Narrow startup-log review

Three most recent desktop startup logs from `~/Library/Logs/com.openai.codex/2026/10/03` were searched for native/bootstrap service errors. The latest main log reports:

- 18:27:17.693 UTC: managed computer-use service available, PID 87787.
- 18:27:22.444 UTC: computer-use MCP tools empty, `authStatus=unsupported`, `runtimeStatus=null`.

No exact Sky-service startup error was found in those three files. Their scope does not establish that the native-app bridge is healthy or rule out errors in other logs. Unrelated logs and private session content were not inspected. No raw logs are copied into this packet.

## Read-only Terminal comparison

Run this in the Terminal session where the count succeeded:

```bash
bash /Users/rscott/Projects/pptxgengo/planning/wm-design-contracts/v4/intake-20261003-frozen/repair-wave/native-access-diagnose.sh
```

It prints manager UID/name, narrow process ancestry, Mach bootstrap state and a presentation count. It only queries PowerPoint if its executable is already running for that UID. Each command has a short timeout. It does not start/close applications or change registry, TCC, launchd configuration or security settings.

Expected useful comparison: Terminal has a usable bootstrap port and resolves its manager, while this receipt shows `task_bootstrap_dead=true`. If Terminal also has a dead port, retain that result rather than assuming the mechanism explains every path.

## Recovery and immediate workaround

1. Preserve the current project context and this receipt.
2. Through the desktop client's supported UI, reconnect local/native computer access if that control is available. If it remains unavailable, normally restart the desktop client and open a fresh task session after saving context. These are recovery suggestions; no restart or reconnect was performed here, and success is not guaranteed by the evidence.
3. Run the same diagnostic in the fresh tool session. Require a usable bootstrap endpoint and a successful presentation-count query before resuming automation.
4. If the tool still fails while Terminal succeeds, run the prepared native capture/export script from Terminal and hand its PDF/diagnostic files back to the agent for inspection. This preserves progress while the native bridge issue is investigated.

The existing deck/native capture script is the supported immediate workflow already demonstrated by the user's Terminal count. PowerPoint and macOS permissions were left untouched during this investigation.

[Official OpenAI sandbox documentation](https://learn.chatgpt.com/docs/sandboxing) distinguishes execution boundaries from approval settings and documents full access as removing filesystem/network restrictions. It does not document or diagnose this specific dead-bootstrap-port failure; the diagnosis above is based on the local read-only probes.
