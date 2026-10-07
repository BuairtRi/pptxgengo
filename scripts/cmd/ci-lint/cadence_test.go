package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// This is a structural policy check, not a substitute for GitLab's rule evaluator.
// Server lint and actual pipelines also exercise the selected job graphs.
func TestRepositoryQualificationCadenceAndReleaseGates(t *testing.T) {
	read := func(path string) map[string]any {
		t.Helper()
		raw, err := os.ReadFile("../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		var config map[string]any
		if err := yaml.Unmarshal(raw, &config); err != nil {
			t.Fatal(err)
		}
		return config
	}
	root, native, release := read(".gitlab-ci.yml"), read(".gitlab/ci/native-cli.yml"), read(".gitlab/ci/release.yml")
	job := func(config map[string]any, name string) map[string]any {
		t.Helper()
		value, ok := config[name].(map[string]any)
		if !ok {
			t.Fatalf("missing job %s", name)
		}
		return value
	}
	developer := job(root, "developer")
	if developer["parallel"] != nil || developer["rules"] != nil || !reflect.DeepEqual(developer["script"], []any{"make test"}) {
		t.Fatal("ordinary developer job must run only the short suite, once")
	}
	qualification := job(root, ".qualification-rules")["rules"].([]any)
	if len(qualification) != 4 {
		t.Fatal("unexpected full qualification rules")
	}
	first := qualification[0].(map[string]any)["if"].(string)
	for _, required := range []string{`$CI_PIPELINE_SOURCE == "schedule"`, `$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH`, `$CI_COMMIT_REF_PROTECTED == "true"`} {
		if !strings.Contains(first, required) {
			t.Fatalf("nightly must be protected default branch: %s", first)
		}
	}
	if !reflect.DeepEqual(qualification[1], job(release, ".release-rules")["rules"].([]any)[0]) {
		t.Fatal("full qualification must cover every supported protected release tag")
	}
	if qualification[2].(map[string]any)["if"] != `$CI_PIPELINE_SOURCE == "web" && $PPTXGENGO_QUALIFICATION == "true"` || !reflect.DeepEqual(qualification[3], map[string]any{"when": "never"}) {
		t.Fatal("full qualification needs explicit web opt-in and terminal exclusion")
	}
	for _, name := range []string{"installation-process", "offline-model", "security:offline-model"} {
		if job(root, name)["extends"] != ".qualification-rules" {
			t.Fatalf("%s must use full qualification rules", name)
		}
	}
	for name, parent := range map[string]string{"installation-process-arm64": "installation-process", "offline-model-arm64": "offline-model"} {
		j := job(root, name)
		if j["extends"] != parent || j["rules"] != nil {
			t.Fatalf("%s must retain its parent's qualification rules", name)
		}
	}
	race := job(root, "developer-race")
	if !reflect.DeepEqual(race["extends"], []any{"developer", ".qualification-rules"}) || !reflect.DeepEqual(race["script"], []any{"make test-race"}) || race["parallel"] != nil {
		t.Fatal("race suite must be one explicit full-qualification job")
	}
	if job(native, "cross-platform-build")["extends"] != ".qualification-rules" {
		t.Fatal("cross-builds must use full qualification rules")
	}
	macRules := job(native, ".native-macos")["rules"].([]any)
	if len(macRules) != 4 || !reflect.DeepEqual(macRules[3], map[string]any{"when": "never"}) {
		t.Fatal("unexpected native Mac rules")
	}
	for i, source := range []string{"schedule", "web"} {
		condition := macRules[i].(map[string]any)["if"].(string)
		for _, part := range []string{`$CI_COMMIT_REF_PROTECTED == "true"`, `$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH`, `$CI_PIPELINE_SOURCE == "` + source + `"`} {
			if !strings.Contains(condition, part) {
				t.Fatalf("Mac qualification must require protected main and explicit source: %s", condition)
			}
		}
		if source == "web" && !strings.Contains(condition, `$PPTXGENGO_QUALIFICATION == "true"`) {
			t.Fatal("Mac web qualification must be opt-in")
		}
	}
	macTag := macRules[2].(map[string]any)["if"].(string)
	if !strings.Contains(macTag, `$CI_COMMIT_REF_PROTECTED == "true"`) || !strings.Contains(macTag, `$CI_COMMIT_TAG =~ /^v`) {
		t.Fatal("Mac tag qualification must remain protected")
	}
	required := []string{"developer", "developer-race", "installation-process", "installation-process-arm64", "offline-model", "offline-model-arm64", "workflow-lint", "cross-platform-build", "macos-cli", "macos-model", "security:go", "security:secrets", "security:offline-model"}
	for _, name := range []string{"release:resources", "release:build"} {
		needs := job(release, name)["needs"].([]any)
		found := map[string]bool{}
		for _, value := range needs {
			need := value.(map[string]any)
			if need["optional"] == true {
				t.Fatalf("release %s cannot skip qualification", name)
			}
			found[need["job"].(string)] = true
		}
		for _, gate := range required {
			if !found[gate] {
				t.Fatalf("release %s missing mandatory %s", name, gate)
			}
		}
	}
}
