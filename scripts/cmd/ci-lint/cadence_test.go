package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Structural policy regression checks complement GitLab's authoritative rules
// simulations and actual pipeline runs; they do not reimplement its evaluator.
func TestRepositoryQualificationCadenceAndReleaseGates(t *testing.T) {
	read := func(path string) map[string]any {
		t.Helper()
		raw, err := os.ReadFile("../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		var config map[string]any
		if err = yaml.Unmarshal(raw, &config); err != nil {
			t.Fatal(err)
		}
		return config
	}
	root, native, release := read(".gitlab-ci.yml"), read(".gitlab/ci/native-cli.yml"), read(".gitlab/ci/release.yml")
	job := func(config map[string]any, name string) map[string]any {
		t.Helper()
		j, ok := config[name].(map[string]any)
		if !ok {
			t.Fatalf("missing job %s", name)
		}
		return j
	}
	rules := func(j map[string]any) []any {
		t.Helper()
		r, ok := j["rules"].([]any)
		if !ok || len(r) == 0 {
			t.Fatal("explicit rules required")
		}
		if !reflect.DeepEqual(r[len(r)-1], map[string]any{"when": "never"}) {
			t.Fatal("rules must end with exclusion")
		}
		return r
	}
	has := func(j map[string]any, condition string) bool {
		for _, v := range rules(j) {
			if v.(map[string]any)["if"] == condition {
				return true
			}
		}
		return false
	}
	tag := rules(job(release, ".release-rules"))[0].(map[string]any)["if"].(string)
	main := `$CI_PIPELINE_SOURCE == "push" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH`
	night := `$CI_PIPELINE_SOURCE == "schedule" && $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH && $CI_COMMIT_REF_PROTECTED == "true"`
	pr := `$CI_PIPELINE_SOURCE == "merge_request_event" || ($CI_PIPELINE_SOURCE =~ /^(api|web)$/ && $PPTXGENGO_CI_TIER == "pr")`
	workflow := job(root, "workflow")
	for _, condition := range []string{main, night, tag, pr} {
		if !has(workflow, condition) {
			t.Fatalf("workflow missing %s", condition)
		}
	}
	for _, value := range rules(workflow) {
		m := value.(map[string]any)
		if m["if"] == `$CI_COMMIT_BRANCH` || m["if"] == `$CI_PIPELINE_SOURCE == "push"` {
			t.Fatal("ordinary branch pushes must not create pipelines")
		}
	}
	developer := job(root, "developer")
	if developer["parallel"] != nil || !reflect.DeepEqual(developer["script"], []any{"make test"}) || !has(developer, main) || !has(developer, tag) || has(developer, pr) {
		t.Fatal("main/tag require one full hermetic unit suite; PRs use their separate focused suite")
	}
	pj := job(root, "pr-unit")
	if !has(pj, pr) || has(pj, main) || !reflect.DeepEqual(pj["script"], []any{`test -z "${PPTXGENGO_PR_HEAD:-}" || test "$CI_COMMIT_SHA" = "$PPTXGENGO_PR_HEAD"`, "make test-pr"}) {
		t.Fatal("PR tier must guard exact head and run focused units")
	}
	for _, pair := range []struct {
		config map[string]any
		name   string
	}{{root, "developer-race"}, {root, "search-performance"}, {root, "search-performance-arm64"}, {root, "offline-model-arm64"}, {native, "cross-platform-build"}, {native, "macos-race"}, {native, "macos-model"}} {
		j := job(pair.config, pair.name)
		if !has(j, night) || has(j, tag) || has(j, main) || has(j, pr) {
			t.Fatalf("%s must be nightly/on-demand only", pair.name)
		}
		for _, value := range rules(j) {
			c, _ := value.(map[string]any)["if"].(string)
			if strings.Contains(c, "CI_COMMIT_TAG") || strings.Contains(c, `== "push"`) {
				t.Fatalf("%s contains a routine/tag rule", pair.name)
			}
		}
	}
	for _, name := range []string{"exhaustive", "windows-native"} {
		for _, value := range rules(job(root, name)) {
			condition, _ := value.(map[string]any)["if"].(string)
			if condition != "" && (!strings.Contains(condition, "PPTXGENGO_") || strings.Contains(condition, "CI_COMMIT_TAG") || strings.Contains(condition, `== "push"`)) {
				t.Fatalf("%s requires an explicit opt-in", name)
			}
		}
	}
	for _, name := range []string{"installation-process", "installation-process-arm64", "offline-model", "security:go", "security:offline-model"} {
		j := job(root, name)
		if !has(j, tag) || !has(j, night) || has(j, main) || has(j, pr) {
			t.Fatalf("%s belongs to tag/night/on-demand", name)
		}
	}
	for _, name := range []string{"installation-process-arm64", "offline-model-arm64", "search-performance-arm64"} {
		if !reflect.DeepEqual(job(root, name)["tags"], []any{"linux", "arm64", "macmini-linux"}) {
			t.Fatalf("%s must never use the Pi pool", name)
		}
	}
	for _, name := range []string{"installation-process", "installation-process-arm64", "offline-model", "search-performance-arm64"} {
		if job(root, name)["cache"].(map[string]any)["policy"] != "pull" {
			t.Fatalf("%s must not re-upload shared cache", name)
		}
	}
	if job(native, "cross-platform-build")["cache"].(map[string]any)["policy"] != "pull" {
		t.Fatal("cross-target cache must not bloat routine jobs")
	}
	prefix := func(j map[string]any) any { return j["cache"].(map[string]any)["key"].(map[string]any)["prefix"] }
	if prefix(developer) == prefix(job(root, "developer-race")) {
		t.Fatal("short/race caches must be separate")
	}
	if job(root, "offline-model-arm64")["cache"].(map[string]any)["policy"] != "pull-push" || prefix(job(root, "offline-model-arm64")) != prefix(job(root, "installation-process-arm64")) {
		t.Fatal("ARM cache writer/reader mismatch")
	}
	for _, line := range job(native, "macos-cli")["script"].([]any) {
		if strings.Contains(line.(string), "test-race") {
			t.Fatal("Mac release qualification cannot include races")
		}
	}
	required := []string{"developer", "pr-relay-unit", "installation-process", "installation-process-arm64", "offline-model", "workflow-lint", "macos-cli", "security:go", "security:secrets", "security:offline-model"}
	for _, name := range []string{"release:resources", "release:build"} {
		found := map[string]bool{}
		for _, v := range job(release, name)["needs"].([]any) {
			n := v.(map[string]any)
			if n["optional"] == true {
				t.Fatalf("%s cannot skip its release gates", name)
			}
			found[n["job"].(string)] = true
		}
		for _, gate := range required {
			if !found[gate] {
				t.Fatalf("%s missing mandatory %s", name, gate)
			}
		}
		for _, gate := range []string{"developer-race", "macos-race", "macos-model", "search-performance", "cross-platform-build", "offline-model-arm64"} {
			if found[gate] {
				t.Fatalf("%s waits for night-only %s", name, gate)
			}
		}
	}
}
