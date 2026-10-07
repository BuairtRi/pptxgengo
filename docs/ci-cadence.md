# Private GitLab CI tiers

All CI and artifacts remain in private GitLab. GitHub is the source PR/merge
system used by slotctl; GitHub Actions is disabled.

| Trigger | Automatic checks | Deliberately excluded |
| --- | --- | --- |
| PR opened/updated | `pr-unit`, subsecond hermetic PR-relay tests, workflow validation, secret scan | Full unit suite, races, installers, models, native desktop jobs |
| Ordinary branch push without a PR | None | Every CI job |
| Successful merge/push to main | `developer` full hermetic unit suite, PR-relay tests, workflow validation, secret scan | Races, installers, models, desktop qualification |
| Protected release tag | Full hermetic units, actual Linux amd64/arm64 and Mac ARM64 installer qualification, pinned model closure/goldens, source security, six-target builds, generated decks/resources, binary signing/notarization, final scans, signature verification, attestation, publication | Races, performance diagnostics, duplicate cross-build job |
| Nightly protected main | Full unit suite, Linux/Mac races, installers, model/retrieval/performance on Linux amd64/arm64 and Mac, cross-builds, security | Actual Office/human acceptance unless separately provisioned/requested |

## Focused PR suite

`make test-pr` compiles every Go package without executing its tests, then runs
the short regression suites for the core PPTX package, installer state,
model packaging/runtime validation, reusable slide storage, PowerShell environment,
and CI policy. Actual installer processes, model weights, and Office actions are
opt-in and remain disabled. This is a deliberately focused suite; `make test`
runs the complete hermetic short suite after main merges and on tags.

The hermetic suite excludes tests explicitly requiring private branding,
catalog-wide integration, downloaded model weights, or an interactive Office
desktop. Those remain distinct qualification jobs, not silently claimed unit
coverage. Mac/Windows CI capacity is reserved for native execution and release
work rather than ordinary PR or main pipelines.

## PR discovery

GitLab CE does not provide the native GitHub external-PR integration. A
project-specific Kubernetes poller reads this public repository's open PRs every
three minutes without a GitHub credential. It accepts same-repository heads
mirrored into private GitLab only, refuses protected branch heads, verifies exact
head SHA before/after pipeline creation, and supplies `PPTXGENGO_PR_HEAD` so the
PR test job checks the expected source again before execution.

The relay deduplicates PR/head/pipeline records and checks trusted hashes for
CI configuration and the Makefile. CI-changing PRs require explicit operator
review and an on-demand PR pipeline until the trusted capability profile is
updated. Forks, missing mirrors, or mismatched heads are logged and not run.
The three-minute discovery delay and GitHub's anonymous shared-IP rate limit
apply. The relay does not publish GitHub checks/statuses; see the private GitLab
pipeline result. Credentials live only in its Kubernetes Secret.

## Individual checks on demand

Create a web/API pipeline on a mirrored branch with `PPTXGENGO_CI_JOB` set to an
exact job name. Examples:

```sh
glab ci run --branch '<mirrored-branch>' --variables PPTXGENGO_CI_JOB:installation-process-arm64
glab ci run --branch '<mirrored-branch>' --variables PPTXGENGO_CI_JOB:developer-race
```

Available selectors include `developer`, `pr-relay-unit`, `workflow-lint`, `installation-process`,
`installation-process-arm64`, `offline-model`, `offline-model-arm64`,
`search-performance`, `search-performance-arm64`, `cross-platform-build`,
`developer-race`, `security:go`, `security:secrets`, `security:offline-model`,
`macos-cli`, `macos-model`, and `macos-race`. Performance/model scans also select
the model job they need. Native Mac selectors require protected main.
`PPTXGENGO_CI_JOB=full` explicitly requests all supported qualification jobs;
use the individual selectors for ordinary subsystem changes.

An explicit `PPTXGENGO_CI_TIER=pr` API/web pipeline selects just the PR tier.
Agents should qualify changed installation/model/platform subsystems on demand
before merging; ordinary PR checks do not imply those paths were exercised.
Local slot submission retains its declared complete preflight, including selected
races. This CI change does not weaken that separate local development contract.

Private-asset exhaustive runs still require `PPTXGENGO_FULL_TESTS=true` and an
integration runner. Windows/Intel qualification stays explicitly opt-in until
those runners are provisioned; see [testing](testing.md).

## Linux ARM64 routing and cache ownership

All native Linux ARM64 jobs require tags `linux`, `arm64`, and `macmini-linux`.
Project-specific runner 24 is locked to project 17. Its manager and Kubernetes
build/helper pods are pinned to `macmini-builder` with a hard node selector and
matching affinity; there is no Pi fallback. Two jobs can run concurrently. The
node's `builder-validation=true:NoSchedule` taint is explicitly tolerated.
The reproducible token-free manifest is [here](../deploy/ci/macmini-linux-runner.json).
Runner authentication is stored separately in the Kubernetes Secret.

Short and race Go caches use separate keys. Installer/model/cross-build consumers
only pull the short cache. The ARM model job is the single native cache writer;
ARM installer/performance jobs only pull it. Cross-target artifacts do not bloat
the everyday cache. All keys include go.mod/go.sum; misses are valid.

In old pipeline 21288 the ARM installer assertions passed in 171.3 seconds, but
cache restore took 215.3 seconds and the 600-second overall job deadline expired
while saving the cache. The failure/trace remain retained; the exact-head retry
passed. Test assertions and package deadlines have not been increased.

## Nightly schedule and failure emails

Project 17 schedule 12 targets protected `refs/heads/main`: cron `0 1 * * *`,
timezone `America/Los_Angeles` (1 a.m. local time, following daylight saving).
The scheduler may enqueue slightly after the nominal minute. The schedule owner
must retain protected-main pipeline permission.

GitLab's native Pipeline status emails integration sends failures only to the
operator's authenticated GitLab account email. Its filter is the default branch,
so it covers failed main and nightly pipelines. Successful runs and slot-branch
failures do not generate these integration emails. Configuration readback is not
proof of delivery; confirm delivery from a real failure, without deliberately
breaking main to manufacture an alert.

## Validation and measured baseline

GitLab server lint must simulate PR, ordinary branch, main, protected stable/RC
and unprotected tags, nightly, and representative individual selectors. Assert
job names as well as YAML validity: a valid pipeline containing an accidental
race job is still a policy failure. CI unit regressions check gate/routing policy.

Old pipeline 21278 ran Linux races for 743.9 seconds and short units for 282.1 seconds;
main pipeline 21277 ran Mac CLI tests for 873.2 seconds. These are measured old job
durations, not promises for the new pipelines. Compare actual PR/main/on-demand
and scheduled runs after rollout. Release signing/notary latency remains necessary.
