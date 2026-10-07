"""Hermetic tests: no network, credentials, deployments or pipeline creation."""
import base64
import copy
import hashlib
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import github_pr_poller as p

S = "a" * 40
REF = "slot/pptxgengo/a/b"
ROOT = b'pr-unit:\n PPTXGENGO_CI_TIER: pr\n PPTXGENGO_PR_HEAD: required\n PPTXGENGO_PR_PIPELINE_NAME: pr:$PPTXGENGO_PR_KEY\n script: make test-pr\n'
FILES = {name: ROOT if name == ".gitlab-ci.yml" else name.encode() for name in p.CI_FILES}
PROFILE = {name: hashlib.sha256(data).hexdigest() for name, data in FILES.items()}


def pr(number=7, sha=S):
    repo = {"id": p.REPO_ID, "full_name": p.REPO}
    return {"state": "open", "number": number, "head": {"repo": repo, "sha": sha, "ref": REF}, "base": {"repo": repo, "ref": "main"}}


class Store:
    def __init__(self):
        self.value = {"schema": 1, "records": {}}
        self.saves = []

    def load(self):
        return copy.deepcopy(self.value)

    def save(self, value):
        self.value = copy.deepcopy(value)
        self.saves.append(copy.deepcopy(value))


class GitHub:
    def __init__(self, prs=None):
        self.prs = [pr()] if prs is None else prs

    def request(self, *args, **kwargs):
        return copy.deepcopy(self.prs), {}


class GitLab:
    def __init__(self, store):
        self.store = store
        self.calls = []
        self.protected = False
        self.mirror_sha = S
        self.files = copy.deepcopy(FILES)
        self.pipelines = []
        self.lose_response = False
        self.result_sha = S
        self.result_source = "api"
        self.result_name = None
        self.cancel_status = None
        self.refused_create = False
        self.omit_create_name = False
        self.detail_error = False
        self.detail_id = 123

    def request(self, method, path, query=None, body=None):
        self.calls.append((method, path, copy.deepcopy(query), copy.deepcopy(body)))
        if method == "GET" and "/repository/branches/" in path:
            return {"name": REF, "protected": self.protected, "commit": {"id": self.mirror_sha}}, {}
        if method == "GET" and "/repository/files/" in path:
            name = p.urllib.parse.unquote(path.split("/repository/files/")[1])
            return {"file_path": name, "encoding": "base64", "content": base64.b64encode(self.files[name]).decode()}, {}
        if method == "GET" and path == "projects/17/pipelines":
            return copy.deepcopy(self.pipelines), {}
        if method == "GET" and path == "projects/17/pipelines/123":
            if self.detail_error:
                raise p.Uncertain("detail-transport-response-uncertain")
            detail = copy.deepcopy(self.pipelines[-1])
            detail["id"] = self.detail_id
            return detail, {}
        if method == "GET" and path.endswith("/variables"):
            raise AssertionError("private variables endpoint must never be requested")
        if method == "POST" and path.endswith("/cancel"):
            if self.cancel_status:
                raise p.HTTPFailure(self.cancel_status)
            return {}, {}
        if method == "POST" and path == "projects/17/pipeline":
            # Durable intent must already exist before mutating the API.
            assert any(r["state"] == "pending" for r in self.store.value["records"].values())
            if self.refused_create:
                raise p.HTTPFailure(400)
            values = {v["key"]: v["value"] for v in body["variables"]}
            name = self.result_name if self.result_name is not None else "pr:" + values["PPTXGENGO_PR_KEY"]
            result = {"id": 123, "sha": self.result_sha, "ref": REF, "source": self.result_source, "tag": False, "name": name}
            self.pipelines.append(result)
            if self.lose_response:
                raise p.Uncertain("creation-response-uncertain")
            response = dict(result)
            if self.omit_create_name:
                response.pop("name")
            return response, {}
        raise AssertionError((method, path))


class PollerTests(unittest.TestCase):
    def setUp(self):
        self.store = Store()
        self.gh = GitHub()
        self.gl = GitLab(self.store)
        self.logs = []
        self.clock = 1000
        self.poller = p.Poller(self.gh, self.gl, self.store, [PROFILE], self.logs.append, lambda: self.clock)

    def creations(self):
        return [x for x in self.gl.calls if x[:2] == ("POST", "projects/17/pipeline")]

    def test_exact_head_encoded_ref_and_dedup(self):
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)
        self.assertEqual(self.store.value["records"][f"{p.REPO_ID}/7/{S}"]["pipeline"], 123)
        self.assertTrue(any("slot%2Fpptxgengo%2Fa%2Fb" in x[1] for x in self.gl.calls))
        self.assertEqual({v["key"]: v["value"] for v in self.creations()[0][3]["variables"]}, p.Poller.variables(7, S, f"{p.REPO_ID}/7/{S}"))
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)

    def test_real_creation_representation_omits_name_but_detail_confirms_it(self):
        self.gl.omit_create_name = True
        self.poller.run()
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "scheduled")
        self.assertTrue(any(c[:2] == ("GET", "projects/17/pipelines/123") for c in self.gl.calls))
        self.assertFalse(any(c[1].endswith("/cancel") for c in self.gl.calls))

    def test_lost_detail_response_recovers_without_recreating_pipeline(self):
        self.gl.omit_create_name = self.gl.detail_error = True
        self.poller.run()
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "pending")
        self.gl.detail_error = False
        self.clock += 180
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "scheduled")

    def test_wrong_detail_id_cancels_original_creation_only(self):
        self.gl.omit_create_name = True
        self.gl.detail_id = 999
        self.poller.run()
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "refused")
        self.assertTrue(any(c[:2] == ("POST", "projects/17/pipelines/123/cancel") for c in self.gl.calls))
        self.assertFalse(any(c[:2] == ("POST", "projects/17/pipelines/999/cancel") for c in self.gl.calls))

    def test_fork_closed_wrong_base_bad_sha_and_protected_are_refused(self):
        changes = [lambda x: x["head"].update(repo={"id": 1, "full_name": "other/fork"}), lambda x: x.update(state="closed"), lambda x: x["base"].update(ref="master"), lambda x: x["head"].update(sha="not-a-sha"), lambda x: x["head"].update(ref="main"), lambda x: x.update(head=[])]
        for change in changes:
            item = pr()
            change(item)
            with self.assertRaises(p.Refused):
                self.poller.identity(item)
        for ref in ("../main", "a//b", "a/../b", "x?token=y", "refs/a.lock", "x\\y"):
            item = pr()
            item["head"]["ref"] = ref
            with self.assertRaises(p.Refused):
                self.poller.identity(item)
        self.gl.protected = True
        self.poller.run()
        self.assertEqual(self.creations(), [])
        self.gl.protected = False
        self.gl.mirror_sha = "b" * 40
        self.poller.run()
        self.assertEqual(self.creations(), [])

    def test_old_heavy_or_changed_included_ci_is_never_triggered(self):
        for name in p.CI_FILES:
            with self.subTest(file=name):
                self.gl.files = dict(FILES)
                self.gl.files[name] += b"\nchanged rules"
                self.poller.run()
                self.assertEqual(self.creations(), [])
        self.assertTrue(any("unsupported-ci-config" in x for x in self.logs))

    def test_lost_creation_response_recovers_by_exact_public_name(self):
        self.gl.lose_response = True
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "pending")
        self.gl.lose_response = False
        self.clock += 180
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "scheduled")
        self.assertTrue(any('"event": "recovered"' in x for x in self.logs))
        lists = [call for call in self.gl.calls if call[:2] == ("GET", "projects/17/pipelines")]
        self.assertTrue(all(call[2]["name"] == f"pr:{p.REPO_ID}/7/{S}" for call in lists))
        self.assertFalse(any(call[1].endswith("/variables") for call in self.gl.calls))

    def test_unrelated_api_pipeline_is_not_recovery_evidence(self):
        self.gl.pipelines = [{"id": 9, "sha": S, "ref": REF, "source": "api", "name": "ARM qualification"}]
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)

    def test_ref_move_response_is_cancelled_and_never_accepted(self):
        self.gl.result_sha = "b" * 40
        self.poller.run()
        self.assertTrue(any(x[:2] == ("POST", "projects/17/pipelines/123/cancel") for x in self.gl.calls))
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "refused")

    def test_known_failure_creation_attempts_are_bounded(self):
        self.gl.refused_create = True
        for _ in range(5):
            self.poller.run()
            self.clock += 180
        self.assertEqual(len(self.creations()), 3)
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "refused")

    def test_pending_response_waits_before_retry(self):
        self.gl.lose_response = True
        self.poller.run()
        self.gl.pipelines = []
        self.clock += 30
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)

    def test_github_rate_backoff_is_durable(self):
        self.gh.request = lambda *a, **k: (_ for _ in ()).throw(p.Backoff(self.clock + 180))
        self.poller.run()
        self.assertEqual(self.store.value["not_before"], self.clock + 180)
        self.gh.request = lambda *a, **k: self.fail("request during backoff")
        self.poller.run()

    def test_request_budget_stops_and_discovery_cursor_advances(self):
        self.gh.prs = [pr(7), pr(8)]
        self.gl.request = lambda *a, **k: (_ for _ in ()).throw(p.Refused("request-budget-exhausted"))
        self.poller.run()
        self.assertEqual(self.store.value["cursor"], 7)
        self.poller.run()
        self.assertEqual(self.store.value["cursor"], 8)

    def test_invalid_journal_refuses_before_any_api_creation(self):
        self.store.value = {"schema": 1, "records": {"wrong": {"state": "pending"}}}
        with self.assertRaises(p.Refused):
            self.poller.run()
        self.assertEqual(self.creations(), [])

    def test_partial_wrong_pr_and_missing_name_do_not_recover(self):
        for name in (f"pr:{p.REPO_ID}/7/{S[:-1]}", f"pr:{p.REPO_ID}/8/{S}", None):
            with self.subTest(name=name):
                self.gl.pipelines = [{"id": 9, "sha": S, "ref": REF, "source": "api", "name": name}]
                self.assertIsNone(self.poller.recover(7, S, REF, f"{p.REPO_ID}/7/{S}"))

    def test_correct_name_wrong_sha_ref_source_is_refused(self):
        for field, value in (("sha", "b" * 40), ("ref", "different"), ("source", "web")):
            item = {"id": 9, "sha": S, "ref": REF, "source": "api", "name": f"pr:{p.REPO_ID}/7/{S}"}
            item[field] = value
            self.gl.pipelines = [item]
            with self.assertRaises(p.Refused):
                self.poller.recover(7, S, REF, f"{p.REPO_ID}/7/{S}")

    def test_mismatched_creation_name_is_cancelled(self):
        self.gl.result_name = f"pr:{p.REPO_ID}/8/{S}"
        self.poller.run()
        self.assertTrue(any(x[:2] == ("POST", "projects/17/pipelines/123/cancel") for x in self.gl.calls))
        self.assertEqual(next(iter(self.store.value["records"].values()))["state"], "refused")

    def test_cancel_permission_denial_keeps_mismatch_refused(self):
        self.gl.result_name = "unexpected-name"
        self.gl.cancel_status = 403
        with self.assertRaises(p.HTTPFailure):
            self.poller.run()
        record = next(iter(self.store.value["records"].values()))
        self.assertEqual(record["state"], "refused")
        self.assertEqual(record["pipeline"], 123)
        self.poller.run()
        self.assertEqual(len(self.creations()), 1)

    def test_ci_without_name_contract_is_refused_even_if_hash_is_approved(self):
        root = ROOT.replace(b" PPTXGENGO_PR_PIPELINE_NAME: pr:$PPTXGENGO_PR_KEY\n", b"")
        self.gl.files[".gitlab-ci.yml"] = root
        profile = dict(PROFILE, **{".gitlab-ci.yml": hashlib.sha256(root).hexdigest()})
        self.poller.profiles = [profile]
        self.poller.run()
        self.assertEqual(self.creations(), [])


class BoundaryTests(unittest.TestCase):
    def test_private_gitlab_root_supplements_default_store(self):
        with patch.object(p.ssl, "create_default_context") as create:
            context = p.gitlab_tls_context("/mounted/private-root.pem")
            create.assert_called_once_with()
            context.load_verify_locations.assert_called_once_with(cafile="/mounted/private-root.pem")
        # Default capath stores can be loaded lazily (and cert enumeration can be
        # empty before a handshake). Check verification policy without requiring
        # this hermetic test to perform a network request or enumerate all roots.
        context = p.gitlab_tls_context(None)
        self.assertEqual(context.verify_mode, p.ssl.CERT_REQUIRED)
        self.assertTrue(context.check_hostname)

    def test_optional_gitlab_root_and_main_full_kubernetes_dns(self):
        with patch.object(p.ssl, "create_default_context") as create:
            context = p.gitlab_tls_context(None)
            create.assert_called_once_with()
            context.load_verify_locations.assert_not_called()
        with patch.object(p.sys, "argv", ["poller", "--gitlab-ca", "/mounted/root.pem"]), patch.object(p, "load_capabilities", return_value=[PROFILE]), patch.object(p, "mounted_secret", return_value="credential-must-not-appear"), patch.object(p, "gitlab_tls_context") as gitlab_context, patch.object(p.ssl, "create_default_context"), patch.object(p, "Client") as client, patch.object(p, "ConfigMapJournal"), patch.object(p, "Poller"):
            self.assertEqual(p.main(), 0)
        gitlab_context.assert_called_once_with("/mounted/root.pem")
        self.assertEqual(client.call_args_list[2].args[0], "https://kubernetes.default.svc.cluster.local")
        self.assertEqual(client.call_args_list[2].kwargs["header"], "Authorization")
        self.assertEqual(client.call_args_list[2].args[1].remaining, 32)

    def test_tls_setup_failure_does_not_log_certificate_error_or_secret(self):
        import io
        diagnostics = io.StringIO()
        with patch.object(p.sys, "argv", ["poller"]), patch.object(p, "load_capabilities", return_value=[PROFILE]), patch.object(p, "mounted_secret", return_value="credential-must-not-appear"), patch.object(p, "gitlab_tls_context", side_effect=OSError("credential-must-not-appear")), patch.object(p.sys, "stderr", diagnostics):
            self.assertEqual(p.main(), 1)
        self.assertEqual(diagnostics.getvalue(), '{"event":"poller-refused-or-unavailable"}\n')

    def test_pagination_follows_headers_without_following_untrusted_urls(self):
        class Pages:
            def request(self, method, path, query):
                return ([pr(query["page"])], {"link": '<https://evil.invalid>; rel="next"'} if query["page"] < 2 else {})
        self.assertEqual([x["number"] for x in p.paged(Pages(), "fixed", {})], [1, 2])
        with self.assertRaises(p.Refused):
            p.paged(Pages(), "fixed", {}, max_pages=1)

    def test_payload_and_secret_bounds(self):
        with self.assertRaises(p.Refused):
            p.bounded_json(b" " * (p.MAX_BODY + 1))
        with tempfile.TemporaryDirectory() as directory:
            file = Path(directory) / "secret"
            file.write_text("do-not-log-this-token-value")
            self.assertEqual(p.mounted_secret(file), "do-not-log-this-token-value")
            file.write_text("x" * 4097)
            with self.assertRaisesRegex(p.Refused, "secret-file-too-large"):
                p.mounted_secret(file)

    def test_empty_capability_is_suspended_fail_closed(self):
        with tempfile.TemporaryDirectory() as directory:
            file = Path(directory) / "capabilities.json"
            file.write_text('{"profiles":[]}')
            with self.assertRaises(p.Refused):
                p.load_capabilities(file)

    def test_redirects_and_url_paths_cannot_forward_token(self):
        self.assertIsNone(p.NoRedirect().redirect_request(None, None, 302, "", {}, "https://evil.invalid"))
        client = p.Client("https://gitlab.samcott.com/api/v4", p.Budget(), "secret-marker-value")
        for path in ("https://evil.invalid", "//evil.invalid", "pipeline?token=secret", "x#y"):
            with self.assertRaises(p.Refused):
                client.request("GET", path)

    def test_create_transport_is_not_retried_and_errors_are_sanitized(self):
        client = p.Client("https://gitlab.samcott.com/api/v4", p.Budget(), "credential-must-not-appear")
        with patch.object(client.opener, "open", side_effect=p.urllib.error.URLError("credential-must-not-appear")) as call:
            with self.assertRaises(p.Uncertain) as caught:
                client.request("POST", "projects/17/pipeline", body={"ref": REF})
        self.assertEqual(call.call_count, 1)
        self.assertNotIn("credential-must-not-appear", str(caught.exception))

    def test_get_retries_are_bounded(self):
        client = p.Client("https://api.github.com", p.Budget())
        with patch.object(client.opener, "open", side_effect=p.urllib.error.URLError("offline")) as call, patch.object(p.time, "sleep"):
            with self.assertRaises(p.Uncertain):
                client.request("GET", "repos/BuairtRi/pptxgengo/pulls")
        self.assertEqual(call.call_count, 3)

    def test_rate_limit_delay_is_bounded_without_logging_body_or_token(self):
        client = p.Client("https://api.github.com", p.Budget())
        error = p.urllib.error.HTTPError("https://api.github.com", 429, "do-not-log", {"Retry-After": "999999999"}, None)
        with patch.object(client.opener, "open", side_effect=error) as call:
            with self.assertRaises(p.Backoff) as caught:
                client.request("GET", "repos/BuairtRi/pptxgengo/pulls")
        self.assertEqual(call.call_count, 1)
        self.assertLessEqual(caught.exception.until, p.time.time() + 21600)
        self.assertEqual(str(caught.exception), "rate-backoff")

    def test_configmap_uses_resource_version_compare_and_swap(self):
        document = {"metadata": {"resourceVersion": "11"}, "data": {"journal.json": '{"schema":1,"records":{}}'}}
        class Kube:
            def request(self, method, path, query=None, body=None):
                if method == "GET":
                    return copy.deepcopy(document), {}
                assert body["metadata"]["resourceVersion"] == "11"
                raise p.HTTPFailure(409)
        store = p.ConfigMapJournal(Kube(), "gitlab-runner", "pptxgengo-pr-poller-journal")
        journal = store.load()
        with self.assertRaises(p.HTTPFailure):
            store.save(journal)


if __name__ == "__main__":
    unittest.main()
