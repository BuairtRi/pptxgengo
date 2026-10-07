# Private GitLab CI cadence

## Everyday pushes and merges

Six jobs run on ordinary branch and main pushes:

- `developer`: the complete hermetic short Go suite (`make test`).
- `workflow-lint`: local CI graph validation, including refusal of GitHub workflows.
- `security:go`: vulnerability analysis for Linux, macOS, and Windows (three jobs).
- `security:secrets`: secret scanning.

These checks do not download model weights, invoke Office, or qualify installers.
All CI runs and artifacts remain in private GitLab; GitHub Actions is disabled.

## Full qualification

Full qualification adds Linux race checks, actual installer process tests on
amd64 and arm64, model closure/goldens/retrieval/performance on both architectures,
model archive SBOM/vulnerability scanning, six-target compilation, and Mac ARM64
short/race/installer/model execution. It runs:

- Nightly against protected `main`, cron `0 1 * * *`, timezone
  `America/Los_Angeles` (1 a.m. local time, following daylight saving time).
- On protected semantic version tags, including the supported `-rc.N` tags.
- On explicitly requested web pipelines with `PPTXGENGO_QUALIFICATION=true`.
  Slot branches get Linux qualification; Mac jobs require protected main.

Every full job remains a mandatory release build prerequisite. Moving checks to
nightly does not let a tag reuse an earlier green pipeline or skip qualification
of the tagged commit. Signing, notarization, final archive scanning, signature
verification, manifest attestation, and publication remain tag-only.

Private branding integration and exhaustive races remain separately opt-in with
`PPTXGENGO_FULL_TESTS=true` and an integration runner. Windows/Intel native jobs
remain opt-in until their runners exist. The nightly suite is headless; it does
not assert Office or human acceptance. Local slot submission retains its complete
five-stage preflight, including selected races.

## Duplicate coverage

The previous developer matrix ran `make test` and `make test-race` on every push;
protected main also repeated both on Mac. The suites now have explicit
`developer` and `developer-race` job names. Only the short Linux suite belongs to
everyday CI. Nightly/release Mac repetition deliberately checks a different OS.

Model jobs previously reran ordinary unit tests for three entire packages.
They now select the four tests that need the actual pinned model weights:
model goldens, offline package copy, closed model package, and reproducible
relocatable release packaging. Ordinary unit cases stay in `developer`.
Installer and retrieval process tests are distinct from their hermetic unit cases;
separate architecture execution is retained in full qualification.

Multiple `go test -race` invocations within the Make target select different test
families. Splitting them preserves independent 150-second package ceilings. These
are not repeated full-package race passes.

## Schedule and failure emails

GitLab project 17 schedule 12 targets `refs/heads/main`. Its owner must retain
permission to run protected-main pipelines. Check the schedule's next-run time
and latest pipeline after changes; GitLab's scheduler can enqueue slightly after
the nominal cron minute.

The native **Pipeline status emails** integration sends failures only to the
owner's authenticated GitLab account email. Its branch scope is `default`.
GitLab filters this integration by branch, not pipeline source: it also reports
failed ordinary main pipelines, not just nightly failures. Successful runs and
slot-branch failures do not produce these integration emails. No additional
SMTP credentials or notification service are stored in this repository.

API configuration readback verifies settings, not actual email delivery. Delivery
must be confirmed from a real failed default-branch pipeline; do not intentionally
break main or manufacture a failure just to send an email.

## Baseline timings

Before this change, exact portable-project commit `814042d4` passed pipeline
21278: Linux races 743.9 seconds, short tests 282.1 seconds, six-target builds
205.9 seconds, model jobs about 190 seconds each. Protected-main pipeline 21277
ran Mac CLI checks for 873.2 seconds. These are measured job durations, not
promised future pipeline budgets. Compare actual fast and scheduled pipelines
once the policy is merged and exercised.
