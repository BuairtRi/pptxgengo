# Install, upgrade, diagnose and recover

These commands are implemented in the source containing the installation manager.
The published `v4.1.0` CLI predates them; use a later package containing this code,
or a build from this source. Read its `installation` help before using the workflow.
Signing and Windows/PowerPoint qualification are separate from these commands.

## Verify and install an extracted package

Authenticate the GitLab release's signed manifest and check its archive SHA-256
as described in [release CI](../release/CI.md) before executing downloaded code.
The installer checks package consistency; it is not a replacement for that signature.

```sh
/path/to/extracted/bin/pptxgengo installation install --from /path/to/extracted
pptxgengo installation doctor
```

On Windows, use `.\bin\pptxgengo.exe` from PowerShell. The full presentation
package's `install-windows.ps1` also installs fonts unless `-SkipFonts` is set.
The Go command never changes fonts. macOS/Linux source installation uses
`scripts/install-local-release.sh`; it prepares resources and calls the same manager.

The installer validates the native architecture, executable build metadata,
version files, hashes and required resources. Full packages must contain a complete
manifest inventory with no unlisted files or symlinks. CLI-only packages use their
binary evidence and record every included file's digest. CLI-only packages do not
install the library, skill or fonts.

A new release is copied to a unique stage, rechecked, and each of the three tools
must start and report its version within 30 seconds. Only then is the directory
promoted. Releases are immutable. A repeated install of identical bytes repairs
missing owned settings or does nothing; different bytes cannot reuse a version.
Interrupted copying may leave an unactivated `.stage-*` folder. Do not remove a
stage while another installer is running.

## Active selection and upgrades

Use the same `installation install --from NEW_PACKAGE` command for an upgrade.
The manager records the previous selection before activating the new version.
All three stable launchers read one active selection. Windows uses one user PATH
entry, `ROOT\bin`, and copied dispatchers that need no symlink privilege. Unix
places owned symlinks in `~/.local/bin`; add that directory to your shell PATH if
it is absent. Unowned executables and launchers are never replaced automatically.

The default state root is `%LOCALAPPDATA%\pptxgengo` on Windows and
`${XDG_DATA_HOME:-~/.local/share}/pptxgengo` on macOS/Linux. The skill defaults to
`${CODEX_HOME:-~/.codex}/skills/west-monroe-presentations`. Existing skill folders
or symlinks are preserved in uniquely named backups. Linked targets are not changed.

Open a new terminal after activation, and restart the agent to load the skill.
Existing shells may retain old PATH resolution. `installation doctor` reports:

- Current and previous versions, package kind, native target and resource pins.
- Manifest/content consistency and pending activation receipts.
- Selected skill location, digest and modifications.
- The executable resolved by this session and competing PATH locations.
- Packaged fonts separately from actual registration and native readiness.

Use `pptxgengo design render-doctor --json` for actual PowerPoint/font readiness.
An installation result never establishes that new presentation content is visually
accepted or that Office has been qualified.

## Stage without activation

```sh
pptxgengo installation install --from PACKAGE --stage-only
```

`--skip-skill` preserves the existing user skill. `--no-path` preserves PATH and
external launcher bindings. These are explicit opt-outs, not proof that the
corresponding resources match the new release. Full Windows installer equivalents
are `-StageOnly`, `-SkipSkill`, `-NoPath` and the separate `-SkipFonts`.
A custom PowerShell `-Destination` is an additional checked staging copy;
activation uses the standard managed root. To relocate managed state and releases,
use the Go command with `--root` and retain those flags for later operations.

## Roll back or recover

```sh
pptxgengo installation rollback
pptxgengo installation recover
pptxgengo installation doctor
```

Rollback validates the recorded previous release and switches CLI and matching
skill together. It uses recorded history, never directory sort order. Font
registration is not rolled back. A repeated rollback switches between the last
recorded selections; use `install --from` to select another retained version.

Activation saves a durable receipt before changing skill, launchers or Windows
user PATH (including its registry value type). Normal failures restore the prior
selection. A killed process leaves the receipt; recovery restores the predecessor
before another installation proceeds. If a user edited a setting or skill after
interruption, recovery refuses to discard it and keeps the receipt and backups.
Inspect those paths and resolve the conflicting edit before retrying. Recovery
never silently breaks a lock or takes over a live installer.

Use the original `--root`, `--bin-dir` and `--skill-dir` when managing custom
locations. These flags are available on every installation command. Doctor is
read-only. Repair of missing owned launchers/skills is a verified reinstallation
of the same package; it will preserve modified skills in new backups.

## Uninstall without removing authored work

```sh
pptxgengo installation uninstall
# Preserve the current skill deliberately:
pptxgengo installation uninstall --keep-skill
```

Uninstall removes owned launcher/PATH bindings and clears the active selection.
An unchanged installed skill is removed or the original user skill is restored.
User modifications and unowned launchers remain in place. Releases and skill
backups are retained for recovery; fonts and authored decks/projects are never
deleted. Multiple custom skill locations require `--keep-skill` unless managing
the single recorded location explicitly. No purge operation is implemented.

## Evidence and remaining qualification

The opt-in `TestInstallationRealToolProcesses` harness builds and executes
the actual three source tools with synthetic CLI package versions. It exercises
stage-only behavior, install/upgrade, repeated installation, all three immutable
dispatchers, rollback, a missing-dispatcher repair, false-version startup refusal,
recovery and uninstall. Owned paths contain spaces. The source tools are unsigned;
fixture metadata is not a publisher signature.

Every activation uses `--no-path`, explicit managed/bin/skill directories and
a fresh controlled child-process PATH. The harness checks the user registry PATH
is unchanged, preserves an existing user skill, retains releases after uninstall
and records a JSON scenario list with actual OS/architecture and package hashes.
It does not qualify a newly opened desktop terminal's inherited registry PATH.

On Windows, the same opt-in also executes the actual
`install-windows.ps1` with `-StageOnly`, then repeated activation using
`-NoPath -SkipSkill -SkipFonts`. Its full-format resources are synthetic
placeholders in an isolated app-data directory. No fake font is registered,
private branding is included or PowerPoint opened. This covers the script and
opt-outs, not a usable full presentation package.

Run from the repository root with a new directory in an existing parent:

```sh
PPTXGENGO_INSTALL_PROCESS_OUT=/path/new-qualification-directory \
  go test -count=1 -timeout=8m -run '^TestInstallationRealToolProcesses$' -v ./internal/installstate
```

CI adds this on Kubernetes Linux, hosted macOS and hosted Windows. Only the
generic `qualification.json` is retained for 14 days; unsigned executable
fixtures are not uploaded. The Linux result also gates future protected release
build/resource jobs. Actual hosted results must pass before source integration.

The bounded regression lane covers hashes/inventory, staging, repeated install,
upgrade, rollback, interrupted activation, conflicting edits, skill symlinks,
uninstall preservation and launcher setup recovery. An isolated macOS arm64
execution check installed the actual signed `v4.1.0` and `v4.1.0-rc.2` CLI
packages, exercised all three dispatchers, rolled back and checked diagnostics.
No user's active installation was changed for that check.

Windows and Linux cross-compilation is available. Windows PATH registration,
PowerShell installer execution, full private resource packages and interactive
PowerPoint qualification still need their respective native evidence. Headless
Go checks and source scans do not replace desktop Office review.
