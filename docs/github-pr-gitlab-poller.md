# GitHub PR discovery for private GitLab CE CI

GitHub remains source/PR/merge authority. This small stdlib poller reads the
public `BuairtRi/pptxgengo` PR list and creates **only project 17 PR-tier API
pipelines** on existing, unprotected, exactly mirrored same-repository heads.
It creates no source refs, GitLab reviews/merges, GitHub statuses or artifacts.
Ordinary branch pushes do not run CLI checks; main/tag/nightly cadence is owned
by `.gitlab-ci.yml`. GitHub Actions remains disabled.

## Current deployment and activation contract

The project 17 poller is activated as a Kubernetes CronJob. It discovers same-repository
PRs every three minutes and creates only PR-tier pipelines. The live capability
ConfigMap holds complete exact-hash profiles of reviewed merged CI configurations.
After a CI change passes explicit qualification and merges, append its complete
profile from the exact mirrored main commit; preserve existing approved profiles
and the journal. A historical hash does not approve a newer CI graph.
The remaining steps below document deployment and recovery for maintainers; they
are not outstanding activation work.

When provisioning a replacement deployment or reactivating a suspended CronJob,
an operator:

1. Reviews the exact merged workflow/job graph. Confirms `source=api` and
   `PPTXGENGO_CI_TIER=pr` select only `pr-unit`, `pr-relay-unit`, `workflow-lint`, and secrets.
   `pr-unit` must verify `PPTXGENGO_PR_HEAD` against `CI_COMMIT_SHA` before tests.
   The workflow name is `$PPTXGENGO_PR_PIPELINE_NAME` (default empty); only its
   PR rule sets it to `pr:$PPTXGENGO_PR_KEY`. Both naming markers are required by
   the poller's trusted capability check. Existing four PR variables are retained.
2. Pins the reviewed Python image digest. Generates `pptxgengo-pr-poller-code`
   from `scripts/ci/github_pr_poller.py`; the CronJob never pulls source at runtime.
3. Provisions a dedicated **project 17 Developer/api** token in the externally
   managed Secret `pptxgengo-pr-poller-token`, key `token`. No token is committed
   here, passed as an argument, stored in CI variables or written to the journal.
4. Creates `deploy/ci/github-pr-poller-journal.yaml` **once** using `kubectl create`.
   Never reset/apply its initial data during upgrades. The deployment/RBAC
   manifest is separate so ordinary deployment updates preserve the journal.
5. Configures `pptxgengo-pr-poller-capabilities` with approved exact SHA256
   profiles for `.gitlab-ci.yml`, `.gitlab/ci/release.yml`,
   `.gitlab/ci/native-cli.yml`, and `Makefile`. Profiles have this shape:

   ```json
   {"profiles":[{".gitlab-ci.yml":"64hex",".gitlab/ci/release.yml":"64hex",
     ".gitlab/ci/native-cli.yml":"64hex","Makefile":"64hex"}]}
   ```

   Generate values from the reviewed merged checkout using `hashlib.sha256`.
   Empty profiles intentionally refuse discovery. Up to eight complete profiles
   may be approved; individual hashes cannot be mixed between profiles.
6. Applies the deployment suspended, checks configuration/secret mounts,
   namespace RBAC and diagnostics, then unsuspends after source/config review.

**CI-changing PRs are not triggered automatically** until the exact config is
approved. An operator can explicitly qualify such a PR after reviewing the rule
graph; capability profiles are then updated after merge. Old pre-cadence heads
are refused even when their source branch is mirrored. If additional CI includes
are introduced, update the poller's exact file allowlist and review the profile.

## Creation and recovery

Discovery uses repositoryID/PR number/headSHA, refuses forks, closed PRs, wrong
bases and protected/missing/mismatched GitLab branches. Slashes and other ref
characters are encoded in API path segments. Each accepted creation saves a
`pending` journal record before the POST, with PR number, source SHA/ref and
attempt timestamp. Scheduled records retain the returned pipeline ID.

If a POST response or subsequent journal save is lost, the next invocation
searches exact `name/sha/ref/source=api` pipelines. The name must equal
`pr:<repositoryID>/<PRnumber>/<40hexHeadSHA>`; returned names are checked again,
so partial names or unrelated qualification pipelines are insufficient evidence.
Recovery never reads private pipeline variables, which can require Maintainer
access even when listing pipelines is permitted for the Developer poller token.
The full pipeline representation must have the same name before it is accepted.
GitLab omits the name in its POST response, so the poller reads the full GET
representation first. A lost/unavailable detail response retains the pending
intent for exact-name recovery. Creation
has no immediate transport retry;
an unresolved intent waits at least 120 seconds before a later retry, and has at
most three attempts. Failed scheduled tests are not automatically re-created.
Manual reruns remain GitLab operations, not journal deletion.

GitLab's documented pipeline API accepts a branch/tag ref, not an atomic source
SHA pin. The poller checks branch protection/SHA twice before POST, verifies the
returned name/SHA/ref/source and cancels mismatches, and the PR job checks the expected
SHA. A concurrent ref move can briefly create a wrong-head pipeline before
cancellation; these controls do not claim a transaction across GitHub/GitLab.
Mismatches are recorded as refused with their pipeline ID before the documented
pipeline-cancel API is called. Cancellation denial/failure cannot be accepted
as evidence or cause automatic re-creation; an operator must inspect that ID.
Developer cancellation is the GitLab default; this deployment does not widen
the token role or change project visibility/cancellation settings.
Creation and journal writes also have an unavoidable crash boundary: recovery
reduces duplicate runs but does not claim server-side exactly-once creation.

## Bounds and operations

- Three-minute discovery; one anonymous GitHub list request normally. The public
  API allowance is 60 requests/hour per shared outbound IP. Pagination is bounded
  to three 100-item pages; partial/oversized inventories are refused.
- At most 96 discovery/API HTTP attempts plus 32 journal HTTP attempts per run;
  total process budget 110 seconds and Kubernetes deadline 120 seconds. GETs have
  at most three transport/5xx attempts; writes have one. Rate limits persist a
  bounded backoff instead of blocking the worker. HTTP bodies are at most 2 MiB;
  each CI file at most 256 KiB; journal at most 800,000 bytes/1,500 records.
- New creations/recoveries are limited to six records per invocation. A journal
  cursor rotates inspection so unsupported heads cannot starve other PRs.
- Redirects are refused; exceptions/body/headers are never logged. JSON logs
  contain only static reasons and PR/SHA/pipeline IDs. Inspect CronJob failures;
  expired tokens, invalid capabilities, exhausted journal/recovery bounds and
  concurrent journal updates fail closed. Review archived completed records
  before explicitly compacting the bounded journal; do not delete open intents.
- `concurrencyPolicy: Forbid` plus ConfigMap resourceVersion protects the journal.
  RBAC grants get/update on its exact ConfigMap only. No Secret-read API rights
  are granted; the dedicated token is mounted by the pod specification.
- The Developer/api token has other Developer actions within project 17; CE has
  no pipeline-only API scope. It cannot access Maintainer-only protected refs;
  project bot users may also read Internal projects. Keep signing variables and
  native/signing runners protected. No GitHub credential means no GitHub checks
  publishing; exact GitLab pipeline evidence remains the merge gate.
- Pause by suspending the CronJob. Preserve its journal and revoke the dedicated
  token when retiring it. Fork/unmirrored PRs need an explicit separate workflow.

## Hermetic tests

```sh
python3 -B -m unittest discover -s scripts/ci -p 'test_github_pr_poller.py'
```

These tests use fake APIs/journals only and never provision infrastructure or
create pipelines. They cover pins/ref encoding/protection/forks, old CI refusal,
creation intent, lost-response recovery, exact public correlation names, cancellation,
deduplication, request/pagination/secret bounds, rate backoff and journal conflicts.
