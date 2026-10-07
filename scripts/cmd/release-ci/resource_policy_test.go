package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopedBrowsingPolicyRefusesInvalidOrMissingResources(t *testing.T) {
	version, commit := "v4.2.0", strings.Repeat("a", 40)
	for _, tc := range []struct{ kind, policy string }{{"cli-only", "unknown"}, {"full", "deferred"}, {"unknown", "required"}} {
		t.Setenv("PPTXGENGO_PACKAGE_KIND", tc.kind)
		t.Setenv("PPTXGENGO_BROWSING_POLICY", tc.policy)
		if err := prepareResourcePolicy(filepath.Join(t.TempDir(), "resources"), version, commit); err == nil {
			t.Fatal("invalid policy accepted", tc)
		}
	}
	t.Setenv("PPTXGENGO_PACKAGE_KIND", "cli-only")
	t.Setenv("PPTXGENGO_BROWSING_POLICY", "")
	root := filepath.Join(t.TempDir(), "required")
	if err := prepareResourcePolicy(root, version, commit); err != nil {
		t.Fatal(err)
	}
	if err := addScopedBrowsingFiles(map[string]Input{}, root, version, commit); err == nil {
		t.Fatal("required decks silently omitted")
	}
}

func TestDeferredBrowsingHasNoArchiveResourcesAndRejectsStaleFiles(t *testing.T) {
	t.Setenv("PPTXGENGO_PACKAGE_KIND", "cli-only")
	t.Setenv("PPTXGENGO_BROWSING_POLICY", "deferred")
	version, commit := "v4.2.0", strings.Repeat("a", 40)
	root := filepath.Join(t.TempDir(), "resources")
	if err := prepareResourcePolicy(root, version, commit); err != nil {
		t.Fatal(err)
	}
	files := map[string]Input{}
	if err := addScopedBrowsingFiles(files, root, version, commit); err != nil || len(files) != 0 {
		t.Fatal("deferred policy added archive payload", err, files)
	}
	if err := addScopedBrowsingFiles(files, root, version, strings.Repeat("b", 40)); err == nil {
		t.Fatal("wrong commit accepted")
	}
	if err := os.Mkdir(filepath.Join(root, "browsing"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := addScopedBrowsingFiles(files, root, version, commit); err == nil {
		t.Fatal("stale browsing directory accepted")
	}
	if err := prepareResourcePolicy(root, version, commit); err == nil {
		t.Fatal("existing resources overwritten")
	}
}

func TestSignedManifestRecordsAndVerifiesDeferredScope(t *testing.T) {
	t.Setenv("PPTXGENGO_PACKAGE_KIND", "cli-only")
	t.Setenv("PPTXGENGO_BROWSING_POLICY", "deferred")
	t.Setenv("PPTXGENGO_OFFLINE_MODEL", "false")
	root, prior := manifestFixture(t)
	// manifestFixture includes a prior manifest; sealing a final directory must
	// receive only inputs, never a pre-existing manifest or checksum/signature.
	if err := os.Remove(filepath.Join(root, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	p := ResourcePolicy{Schema: "pptxgengo.release-resources-policy.v1", Version: prior.Version, Commit: prior.Commit, PackageKind: "cli-only", BrowsingPolicy: "deferred"}
	if err := writeJSON(filepath.Join(root, resourcePolicyName), p); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(root, "security-policy.json"), map[string]bool{"passed": true}); err != nil {
		t.Fatal(err)
	}
	if err := seal(root, prior.Version, prior.Commit); err != nil {
		t.Fatal(err)
	}
	var sealed Manifest
	if err := readJSON(filepath.Join(root, "manifest.json"), &sealed); err != nil {
		t.Fatal(err)
	}
	if sealed.BrowsingPolicy != "deferred" || sealed.Branding != "none-distribution-deferred" || sealed.Files[resourcePolicyName] == "" {
		t.Fatal("signed scope inaccurate", sealed)
	}
	if err := verify(root, prior.Version, prior.Commit); err != nil {
		t.Fatal(err)
	}
	sealed.BrowsingPolicy = ""
	if err := writeJSON(filepath.Join(root, "manifest.json"), sealed); err != nil {
		t.Fatal(err)
	}
	if err := verify(root, prior.Version, prior.Commit); err == nil {
		t.Fatal("deleted scope accepted for a policy-bearing manifest")
	}
	sealed.BrowsingPolicy = "required"
	if err := writeJSON(filepath.Join(root, "manifest.json"), sealed); err != nil {
		t.Fatal(err)
	}
	if err := verify(root, prior.Version, prior.Commit); err == nil {
		t.Fatal("scope tampering accepted")
	}
	if description := browsingContentDescription("cli-only", "deferred"); !strings.Contains(description, "deferred") || strings.Contains(description, "includes two") {
		t.Fatal(description)
	}
}
