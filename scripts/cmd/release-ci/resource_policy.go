package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const resourcePolicyName = "release-resources-policy.json"

type ResourcePolicy struct {
	Schema         string `json:"schema"`
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	PackageKind    string `json:"package_kind"`
	BrowsingPolicy string `json:"browsing_policy"`
}

func browsingPolicy() string {
	p := os.Getenv("PPTXGENGO_BROWSING_POLICY")
	if p == "" {
		return "required"
	}
	return p
}

func validateResourcePolicy(p ResourcePolicy, version, commit string) error {
	decoded, decodeErr := hex.DecodeString(commit)
	if decodeErr != nil || len(decoded) != 20 {
		return fmt.Errorf("resource policy commit must be a SHA-1 object ID")
	}
	if p.Schema != "pptxgengo.release-resources-policy.v1" || p.Version != version || p.Commit != commit || !versionRE.MatchString(version) || len(commit) != 40 {
		return fmt.Errorf("resource policy identity mismatch")
	}
	if p.PackageKind != "cli-only" && p.PackageKind != "full" {
		return fmt.Errorf("unknown package kind")
	}
	if p.BrowsingPolicy != "required" && p.BrowsingPolicy != "deferred" && p.BrowsingPolicy != "templates-only" {
		return fmt.Errorf("unknown browsing policy")
	}
	if p.BrowsingPolicy == "deferred" && p.PackageKind != "cli-only" {
		return fmt.Errorf("deferred browsing is allowed only for CLI-only releases")
	}
	if p.BrowsingPolicy == "templates-only" && p.PackageKind != "cli-only" {
		return fmt.Errorf("template-only browsing is allowed only for CLI-only releases")
	}
	return nil
}

func readResourcePolicy(root, version, commit string) (ResourcePolicy, error) {
	var p ResourcePolicy
	i, err := os.Lstat(root)
	if err != nil || !i.IsDir() || i.Mode()&os.ModeSymlink != 0 {
		return p, fmt.Errorf("resource policy requires a real directory")
	}
	name := filepath.Join(root, resourcePolicyName)
	i, err = os.Lstat(name)
	if err != nil || !i.Mode().IsRegular() || i.Size() > 4096 {
		return p, fmt.Errorf("resource policy requires a regular file")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		return p, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&p); err != nil {
		return p, err
	}
	if err = decoder.Decode(new(any)); err != io.EOF {
		return p, fmt.Errorf("resource policy must contain one JSON object")
	}
	return p, validateResourcePolicy(p, version, commit)
}

func prepareResourcePolicy(root, version, commit string) error {
	p := ResourcePolicy{Schema: "pptxgengo.release-resources-policy.v1", Version: version, Commit: commit, PackageKind: packageKind(), BrowsingPolicy: browsingPolicy()}
	if err := validateResourcePolicy(p, version, commit); err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	i, err := os.Lstat(root)
	if err != nil || !i.IsDir() || i.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("resource policy requires a real directory")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if p.BrowsingPolicy == "deferred" && len(entries) != 0 {
		return fmt.Errorf("deferred resource directory must be empty; refusing stale presentation resources")
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(root, resourcePolicyName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(append(raw, '\n'))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func addScopedBrowsingFiles(files map[string]Input, root, version, commit string) error {
	p, err := readResourcePolicy(root, version, commit)
	if err != nil {
		return err
	}
	if p.PackageKind != packageKind() || p.BrowsingPolicy != browsingPolicy() {
		return fmt.Errorf("resource policy differs from current packaging intent")
	}
	if p.BrowsingPolicy == "required" {
		return addBrowsingFiles(files, root)
	}
	if p.BrowsingPolicy == "templates-only" {
		return addTemplateCatalogFiles(files, root, version, commit)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) != 1 || entries[0].Name() != resourcePolicyName || len(files) != 0 {
		return fmt.Errorf("deferred release refuses staged presentation resources")
	}
	return nil
}

func browsingContentDescription(kind, policy string) string {
	if policy == "templates-only" {
		return "Every installation archive includes the generated template browsing deck and a pinned source catalog, SQLite discovery index and fonts. Reusable-slide inventory and private branding/photo originals are not included. Template slides contain illustrative placeholders and are not native PowerPoint qualification."
	}
	if policy == "deferred" {
		return "Browsing decks, reusable-slide inventory and branding/graphics/photo distribution are deferred to the following release. CLI-only archives omit authoring libraries, fonts and skills; presentation builds require an existing pinned authoring bundle and its fonts."
	}
	if kind == "full" {
		return "Every installation archive includes two private generated browsing PowerPoint libraries with coverage manifests and verified private authoring resources and branding originals."
	}
	return "Every installation archive includes two private generated browsing PowerPoint libraries with coverage manifests. CLI-only archives omit branding originals and authoring resources."
}
