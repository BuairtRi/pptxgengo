package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryGitLabConfiguration(t *testing.T) {
	if err := validate("../../.."); err != nil {
		t.Fatal(err)
	}
}

func TestCIConfigurationRefusesInvalidGraphs(t *testing.T) {
	for _, tc := range []struct {
		name, body, include string
		github              bool
	}{
		{"valid", "stages: [test]\none:\n  script: [echo]\n", "", false},
		{"missing-need", "stages: [test]\none:\n  script: [echo]\n  needs: [missing]\n", "", false},
		{"cycle", "stages: [test]\none:\n  script: [echo]\n  needs: [two]\ntwo:\n  script: [echo]\n  needs: [one]\n", "", false},
		{"extends-cycle", "stages: [test]\none:\n  extends: .base\n  script: [echo]\n.base:\n  extends: one\n", "", false},
		{"bad-stage", "stages: [test]\none:\n  stage: publish\n  script: [echo]\n", "", false},
		{"duplicate-key", "stages: [test]\none:\n  script: [echo]\none:\n  script: [echo]\n", "", false},
		{"unsafe-include", "stages: [test]\ninclude: [{local: ../other.yml}]\none:\n  script: [echo]\n", "", false},
		{"remote-include", "stages: [test]\ninclude: [{remote: https://example.com/ci.yml}]\none:\n  script: [echo]\n", "", false},
		{"duplicate-include", "stages: [test]\ninclude: [{local: other.yml}, {local: other.yml}]\none:\n  script: [echo]\n", ".base:\n  script: [echo]\n", false},
		{"github-workflow", "stages: [test]\none:\n  script: [echo]\n", "", true},
		{"bad-script", "stages: [test]\none:\n  script: [42]\n", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ".gitlab-ci.yml"), []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			if tc.include != "" {
				if err := os.WriteFile(filepath.Join(root, "other.yml"), []byte(tc.include), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.github {
				path := filepath.Join(root, ".github/workflows")
				if err := os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "test.yaml"), []byte("name: unwanted"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := validate(root)
			if (err == nil) != (tc.name == "valid") {
				t.Fatalf("unexpected validation result: %v", err)
			}
		})
	}
}
