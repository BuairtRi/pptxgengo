#!/usr/bin/env python3
"""GitHub public PR discovery -> exact mirrored GitLab project 17 PR pipelines.

No GitHub writes. Tokens come only from mounted files, never CLI arguments.
The journal records creation intent before the API request; uncertain responses
are recovered by checking pipeline variables before another creation attempt.
"""
import argparse
import base64
import hashlib
import json
import re
import ssl
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

REPO = "BuairtRi/pptxgengo"
REPO_ID = 1306202368
PROJECT = 17
SHA = re.compile(r"^[0-9a-f]{40}$")
DIGEST = re.compile(r"^[0-9a-f]{64}$")
CI_FILES = {".gitlab-ci.yml", ".gitlab/ci/release.yml", ".gitlab/ci/native-cli.yml", "Makefile"}
MAX_BODY = 2 * 1024 * 1024
MAX_RECORDS = 1500
MAX_PRS = 300
MAX_PAGES = 3
MAX_ATTEMPTS = 3
EXPECTED_KEYS = {"PPTXGENGO_CI_TIER", "PPTXGENGO_PR_HEAD", "PPTXGENGO_PR_NUMBER", "PPTXGENGO_PR_KEY"}


class Refused(Exception):
    pass


class HTTPFailure(Refused):
    def __init__(self, status):
        self.status = status
        super().__init__("http-%d" % status)


class Uncertain(Refused):
    pass


class Backoff(Refused):
    def __init__(self, until):
        self.until = until
        super().__init__("rate-backoff")


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def bounded_json(data):
    if len(data) > MAX_BODY:
        raise Refused("response-too-large")
    try:
        return json.loads(data)
    except (ValueError, UnicodeError):
        raise Refused("invalid-json") from None


def mounted_secret(path):
    with open(path, "rb") as f:
        raw = f.read(4097)
    if len(raw) > 4096:
        raise Refused("secret-file-too-large")
    try:
        value = raw.decode("ascii").strip()
    except UnicodeError:
        raise Refused("invalid-secret-file") from None
    if not re.fullmatch(r"[A-Za-z0-9_.-]{16,4096}", value):
        raise Refused("invalid-secret-file")
    return value


def gitlab_tls_context(cafile):
    # Private GitLab roots supplement the system store: ingress can present a
    # publicly trusted certificate, and hostname/certificate checks stay enabled.
    context = ssl.create_default_context()
    if cafile:
        context.load_verify_locations(cafile=cafile)
    return context


class Budget:
    def __init__(self, seconds=110, requests=96):
        self.end = time.monotonic() + seconds
        self.remaining = requests

    def take(self):
        self.remaining -= 1
        left = self.end - time.monotonic()
        if self.remaining < 0 or left <= 0:
            raise Refused("request-budget-exhausted")
        return min(10, left)


class Client:
    def __init__(self, base, budget, token=None, header="PRIVATE-TOKEN", context=None):
        self.base = base.rstrip("/") + "/"
        self.budget = budget
        self.headers = {"Accept": "application/json", "User-Agent": "pptxgengo-pr-poller/1"}
        if token:
            self.headers[header] = ("Bearer " if header == "Authorization" else "") + token
        self.opener = urllib.request.build_opener(NoRedirect(), urllib.request.HTTPSHandler(context=context))

    def request(self, method, path, query=None, body=None):
        # Callers provide paths, never URLs; credentials cannot follow redirects.
        if path.startswith("/") or ":" in path or "?" in path or "#" in path:
            raise Refused("invalid-api-path")
        url = self.base + path
        if query:
            url += "?" + urllib.parse.urlencode(query)
        payload = None if body is None else json.dumps(body, separators=(",", ":")).encode()
        headers = dict(self.headers)
        if payload is not None:
            headers["Content-Type"] = "application/json"
        # Only safe GETs retry transport/5xx failures. Pipeline creation never retries here.
        attempts = 3 if method == "GET" else 1
        for attempt in range(attempts):
            try:
                req = urllib.request.Request(url, data=payload, headers=headers, method=method)
                with self.opener.open(req, timeout=self.budget.take()) as response:
                    raw = response.read(MAX_BODY + 1)
                    return bounded_json(raw), {k.lower(): v for k, v in response.headers.items()}
            except urllib.error.HTTPError as e:
                if e.code == 429 or (e.code == 403 and e.headers.get("X-RateLimit-Remaining") == "0"):
                    now = time.time()
                    try:
                        until = max(now + float(e.headers.get("Retry-After", "60")), float(e.headers.get("X-RateLimit-Reset", "0")))
                    except ValueError:
                        until = now + 60
                    raise Backoff(min(now + 21600, max(now + 10, until))) from None
                if method == "GET" and e.code >= 500 and attempt + 1 < attempts:
                    time.sleep(2 ** attempt)
                    continue
                if method != "GET" and e.code >= 500:
                    raise Uncertain("creation-response-uncertain") from None
                raise HTTPFailure(e.code) from None
            except (urllib.error.URLError, TimeoutError, OSError):
                if method == "GET" and attempt + 1 < attempts:
                    time.sleep(2 ** attempt)
                    continue
                raise Uncertain("transport-response-uncertain") from None
        raise Refused("request-failed")


def paged(client, path, query, max_pages=MAX_PAGES, max_items=MAX_PRS):
    items = []
    for page in range(1, max_pages + 1):
        values, headers = client.request("GET", path, dict(query, per_page=100, page=page))
        if not isinstance(values, list) or len(values) > 100:
            raise Refused("invalid-page")
        items.extend(values)
        if len(items) > max_items:
            raise Refused("inventory-bound-exceeded")
        more = bool(headers.get("x-next-page")) or 'rel="next"' in headers.get("link", "")
        if not more:
            return items
    raise Refused("pagination-bound-exceeded")


def load_capabilities(path):
    with open(path, "rb") as f:
        data = bounded_json(f.read(MAX_BODY + 1))
    profiles = data.get("profiles") if isinstance(data, dict) else None
    if not isinstance(profiles, list) or not 1 <= len(profiles) <= 8:
        raise Refused("ci-capability-not-configured")
    for profile in profiles:
        if not isinstance(profile, dict) or set(profile) != CI_FILES or any(not isinstance(x, str) or not DIGEST.fullmatch(x) for x in profile.values()):
            raise Refused("invalid-ci-capability")
    return profiles


class ConfigMapJournal:
    def __init__(self, kube, namespace, name):
        self.kube = kube
        self.path = "api/v1/namespaces/%s/configmaps/%s" % (urllib.parse.quote(namespace, safe=""), urllib.parse.quote(name, safe=""))
        self.document = None

    def load(self):
        self.document, _ = self.kube.request("GET", self.path)
        if not isinstance(self.document, dict) or not self.document.get("metadata", {}).get("resourceVersion"):
            raise Refused("invalid-journal-configmap")
        raw = self.document.get("data", {}).get("journal.json", "")
        return bounded_json(raw.encode())

    def save(self, journal):
        raw = json.dumps(journal, separators=(",", ":"), sort_keys=True)
        if len(raw.encode()) > 800000:
            raise Refused("journal-too-large")
        self.document.setdefault("data", {})["journal.json"] = raw
        # resourceVersion makes a competing journal writer fail closed (HTTP409).
        self.document, _ = self.kube.request("PUT", self.path, body=self.document)


def validate_journal(journal):
    if not isinstance(journal, dict) or journal.get("schema") != 1 or not isinstance(journal.get("records"), dict) or len(journal["records"]) > MAX_RECORDS:
        raise Refused("invalid-journal")
    not_before = journal.get("not_before", 0)
    if not isinstance(not_before, (float, int)) or not 0 <= not_before <= time.time() + 21601:
        raise Refused("invalid-journal-backoff")
    if not isinstance(journal.get("cursor", 0), int) or not 0 <= journal.get("cursor", 0) <= 2147483647:
        raise Refused("invalid-journal-cursor")
    for key, record in journal["records"].items():
        if not isinstance(record, dict) or record.get("state") not in ("pending", "scheduled", "refused") or record.get("attempts", 0) not in range(4):
            raise Refused("invalid-journal-record")
        number, sha = record.get("pr"), record.get("sha")
        if not isinstance(number, int) or isinstance(number, bool) or number <= 0 or not isinstance(sha, str) or not SHA.fullmatch(sha) or key != "%s/%s/%s" % (REPO_ID, number, sha):
            raise Refused("invalid-journal-identity")
        if not isinstance(record.get("ref"), str) or len(record["ref"]) > 255:
            raise Refused("invalid-journal-ref")
        if not isinstance(record.get("last_attempt", 0), (int, float)) or not 0 <= record.get("last_attempt", 0) <= time.time() + 60:
            raise Refused("invalid-journal-time")


class Poller:
    def __init__(self, github, gitlab, store, profiles, emit=print, now=time.time):
        self.github, self.gitlab, self.store = github, gitlab, store
        self.profiles, self.emit, self.now = profiles, emit, now

    def log(self, event, **fields):
        self.emit(json.dumps(dict(event=event, **fields), sort_keys=True))

    def identity(self, pr):
        if not isinstance(pr, dict) or pr.get("state") != "open":
            raise Refused("not-open-main-pr")
        number = pr.get("number")
        head, base = pr.get("head", {}), pr.get("base", {})
        if not isinstance(head, dict) or not isinstance(base, dict):
            raise Refused("invalid-pr-shape")
        if base.get("ref") != "main":
            raise Refused("not-open-main-pr")
        if not isinstance(number, int) or isinstance(number, bool) or number <= 0:
            raise Refused("invalid-pr-number")
        for repo in (head.get("repo") or {}, base.get("repo") or {}):
            if not isinstance(repo, dict) or repo.get("id") != REPO_ID or repo.get("full_name") != REPO:
                raise Refused("fork-or-unexpected-repository")
        sha, ref = head.get("sha"), head.get("ref")
        if not isinstance(sha, str) or not SHA.fullmatch(sha):
            raise Refused("invalid-pr-sha")
        if not isinstance(ref, str) or not 1 <= len(ref) <= 255 or ref in ("main", "master") or re.search(r"[\x00-\x20\x7f~^:?*\[\]\\]", ref) or ".." in ref or "@{" in ref or ref.startswith("-") or ref.endswith(".") or any(not part or part.startswith(".") or part.endswith(".lock") for part in ref.split("/")):
            raise Refused("invalid-pr-ref")
        return number, sha, ref, "%s/%s/%s" % (REPO_ID, number, sha)

    def mirror(self, sha, ref):
        data, _ = self.gitlab.request("GET", "projects/17/repository/branches/" + urllib.parse.quote(ref, safe=""))
        if not isinstance(data, dict) or data.get("name") != ref or data.get("protected") is not False or not isinstance(data.get("commit"), dict) or data["commit"].get("id") != sha:
            raise Refused("mirror-sha-or-protection-mismatch")

    def capable(self, sha):
        actual, root = {}, ""
        remaining = self.profiles
        for path in sorted(CI_FILES):
            data, _ = self.gitlab.request("GET", "projects/17/repository/files/" + urllib.parse.quote(path, safe=""), {"ref": sha})
            if not isinstance(data, dict) or data.get("encoding") != "base64" or data.get("file_path") != path:
                raise Refused("invalid-ci-file")
            try:
                content = base64.b64decode(data.get("content", ""), validate=True)
            except (ValueError, TypeError):
                raise Refused("invalid-ci-content") from None
            if len(content) > 256 * 1024:
                raise Refused("ci-file-too-large")
            actual[path] = hashlib.sha256(content).hexdigest()
            remaining = [p for p in remaining if p[path] == actual[path]]
            if not remaining:
                raise Refused("unsupported-ci-config")
            if path == ".gitlab-ci.yml":
                root = content.decode("utf-8")
        if actual not in self.profiles or not all(x in root for x in ("pr-unit:", "PPTXGENGO_CI_TIER", "PPTXGENGO_PR_HEAD", "make test-pr")):
            raise Refused("unsupported-ci-config")

    def recover(self, number, sha, ref, key):
        pipelines = paged(self.gitlab, "projects/17/pipelines", {"sha": sha, "ref": ref, "source": "api", "order_by": "id", "sort": "desc"}, max_pages=2, max_items=200)
        if len(pipelines) > 20:
            raise Refused("recovery-pipeline-bound-exceeded")
        for pipeline in pipelines:
            if not isinstance(pipeline, dict) or pipeline.get("sha") != sha or pipeline.get("ref") != ref or pipeline.get("source") != "api" or not isinstance(pipeline.get("id"), int) or isinstance(pipeline.get("id"), bool) or pipeline["id"] <= 0:
                raise Refused("invalid-recovery-pipeline")
            variables, _ = self.gitlab.request("GET", "projects/17/pipelines/%s/variables" % pipeline["id"])
            if not isinstance(variables, list) or len(variables) > 256:
                raise Refused("invalid-pipeline-variables")
            selected = {}
            for item in variables:
                if not isinstance(item, dict):
                    raise Refused("invalid-pipeline-variable")
                k = item.get("key")
                if isinstance(k, str) and k in EXPECTED_KEYS:
                    if k in selected:
                        raise Refused("duplicate-pipeline-variable")
                    selected[k] = item.get("value")
            if selected == self.variables(number, sha, key):
                return pipeline["id"]
        return None

    @staticmethod
    def variables(number, sha, key):
        return {"PPTXGENGO_CI_TIER": "pr", "PPTXGENGO_PR_HEAD": sha, "PPTXGENGO_PR_NUMBER": str(number), "PPTXGENGO_PR_KEY": key}

    def process(self, pr, journal):
        number, sha, ref, key = self.identity(pr)
        records = journal["records"]
        record = records.get(key)
        if record and record["ref"] != ref:
            raise Refused("journal-ref-conflict")
        if record and record["state"] in ("scheduled", "refused"):
            return
        self.mirror(sha, ref)
        self.capable(sha)
        existing = self.recover(number, sha, ref, key)
        if existing:
            records[key] = dict(pr=number, sha=sha, ref=ref, state="scheduled", attempts=(record or {}).get("attempts", 0), pipeline=existing, last_attempt=(record or {}).get("last_attempt", 0))
            self.store.save(journal)
            self.log("recovered", pr=number, sha=sha, pipeline=existing)
            return
        if record and self.now() - record.get("last_attempt", 0) < 120:
            self.log("awaiting-uncertain-response", pr=number, sha=sha)
            return
        if record and record.get("attempts", 0) >= MAX_ATTEMPTS:
            record["state"] = "refused"
            self.store.save(journal)
            raise Refused("creation-attempt-bound-exceeded")
        if not record and len(records) >= MAX_RECORDS:
            raise Refused("journal-record-bound-exceeded")
        record = dict(pr=number, sha=sha, ref=ref, state="pending", attempts=(record or {}).get("attempts", 0) + 1, last_attempt=self.now())
        records[key] = record
        self.store.save(journal)
        # Repeat mirror guard immediately before creation. The API response and
        # pr-unit guard cover a ref moving after this read; no SHA-only API pin.
        self.mirror(sha, ref)
        values = self.variables(number, sha, key)
        body = {"ref": ref, "variables": [{"key": k, "value": v, "variable_type": "env_var"} for k, v in values.items()]}
        result, _ = self.gitlab.request("POST", "projects/17/pipeline", body=body)
        if not isinstance(result, dict) or not isinstance(result.get("id"), int) or isinstance(result.get("id"), bool) or result["id"] <= 0:
            raise Uncertain("invalid-creation-response")
        if result.get("sha") != sha or result.get("ref") != ref or result.get("source") != "api" or result.get("tag") is not False:
            self.gitlab.request("POST", "projects/17/pipelines/%s/cancel" % result["id"])
            record["state"] = "refused"
            record["pipeline"] = result["id"]
            self.store.save(journal)
            raise Refused("created-pipeline-identity-mismatch")
        record.update(state="scheduled", pipeline=result["id"])
        self.store.save(journal)
        self.log("scheduled", pr=number, sha=sha, pipeline=result["id"])

    def run(self):
        journal = self.store.load()
        validate_journal(journal)
        if journal.get("not_before", 0) > self.now():
            self.log("rate-backoff-active")
            return
        try:
            prs = paged(self.github, "repos/%s/pulls" % REPO, {"state": "open", "base": "main"})
        except Backoff as e:
            journal["not_before"] = e.until
            self.store.save(journal)
            self.log("rate-backoff-recorded")
            return
        # Rotate discovery so unsupported CI-changing PRs cannot starve other
        # heads when the per-run request/creation budget is reached.
        if any(not isinstance(p, dict) or not isinstance(p.get("number"), int) or isinstance(p.get("number"), bool) or not 0 < p["number"] <= 2147483647 for p in prs):
            raise Refused("invalid-pr-inventory")
        prs.sort(key=lambda p: (p["number"] <= journal.get("cursor", 0), p["number"]))
        changed = 0
        for pr in prs:
            journal["cursor"] = pr["number"]
            try:
                before = len(journal["records"])
                self.process(pr, journal)
                changed += len(journal["records"]) - before
                if changed >= 6:
                    self.log("creation-batch-bound-reached")
                    break
            except Backoff as e:
                journal["not_before"] = e.until
                self.store.save(journal)
                self.log("rate-backoff-recorded")
                break
            except HTTPFailure as e:
                if e.status in (401, 403, 409):
                    raise
                self.log("skipped", reason=str(e))
            except Uncertain:
                self.log("uncertain-request-recorded")
                break
            except Refused as e:
                self.log("skipped", reason=str(e))
                if str(e) == "request-budget-exhausted":
                    break
        self.store.save(journal)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--namespace", default="gitlab-runner")
    parser.add_argument("--journal", default="pptxgengo-pr-poller-journal")
    parser.add_argument("--capabilities", default="/config/capabilities.json")
    parser.add_argument("--gitlab-ca", help="mounted CA bundle for the private GitLab endpoint")
    args = parser.parse_args()
    try:
        profiles = load_capabilities(args.capabilities)
        budget = Budget()
        github = Client("https://api.github.com", budget)
        gitlab = Client("https://gitlab.samcott.com/api/v4", budget, mounted_secret("/var/run/secrets/pptxgengo-poller/token"), context=gitlab_tls_context(args.gitlab_ca))
        sa = "/var/run/secrets/kubernetes.io/serviceaccount/"
        # Reserve a separate bounded journal budget so discovery exhaustion can
        # still persist the cursor/backoff before this invocation exits.
        kube = Client("https://kubernetes.default.svc.cluster.local", Budget(requests=32), mounted_secret(sa + "token"), header="Authorization", context=ssl.create_default_context(cafile=sa + "ca.crt"))
        Poller(github, gitlab, ConfigMapJournal(kube, args.namespace, args.journal), profiles).run()
    except Backoff:
        print('{"event":"rate-backoff-before-discovery"}', file=sys.stderr)
        return 1
    except (Refused, OSError, ValueError, UnicodeError):
        # Never print exception bodies, payloads, headers, URLs or credentials.
        print('{"event":"poller-refused-or-unavailable"}', file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
