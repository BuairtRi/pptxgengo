package component

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

// ResourceColorTarget addresses one color attribute in a retained XML resource.
// A pinned part hash and node path rule out global theme or XML replacements.
type ResourceColorTarget struct {
	Part        string `json:"part"`
	SHA256      string `json:"sha256"`
	NodePath    []int  `json:"node_path"`
	SourceValue string `json:"source_value"`
}
type validatedResourceTarget struct {
	Role   string
	Target ResourceColorTarget
}

func localResourcePath(project, part string) (string, error) {
	if !filepath.IsLocal(part) || filepath.Clean(part) != part || !strings.HasPrefix(filepath.ToSlash(part), "resources/") {
		return "", fmt.Errorf("resource part must be local under resources/: %q", part)
	}
	root, e := filepath.EvalSymlinks(project)
	if e != nil {
		return "", e
	}
	root, e = filepath.Abs(root)
	if e != nil {
		return "", e
	}
	file, e := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(part)))
	if e != nil {
		return "", e
	}
	file, e = filepath.Abs(file)
	if e != nil {
		return "", e
	}
	if !strings.HasPrefix(file, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("resource escapes project: %q", part)
	}
	info, e := os.Stat(file)
	if e != nil {
		return "", e
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("resource is not regular: %q", part)
	}
	return file, nil
}
func validateResourceTargets(project string, roles map[string]Role, profiles map[string]map[string]string, hexColor, schemeColor *regexp.Regexp) ([]validatedResourceTarget, error) {
	seen := map[string]bool{}
	sha256Hex := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	var out []validatedResourceTarget
	for role, r := range roles {
		for _, t := range r.ResourceTargets {
			if t.Part == "" || len(t.NodePath) == 0 || !sha256Hex.MatchString(t.SHA256) {
				return nil, fmt.Errorf("role %s has invalid resource identity", role)
			}
			key := t.Part + ":" + pathKey(t.NodePath)
			if seen[key] {
				return nil, fmt.Errorf("duplicate resource color target %s", key)
			}
			seen[key] = true
			file, e := localResourcePath(project, t.Part)
			if e != nil {
				return nil, e
			}
			data, e := os.ReadFile(file)
			if e != nil {
				return nil, e
			}
			if hash(data) != strings.ToLower(t.SHA256) {
				return nil, fmt.Errorf("resource %s differs from pinned hash", t.Part)
			}
			root, e := nativepkg.Parse(data)
			if e != nil {
				return nil, e
			}
			node, e := nodeAt(root, t.NodePath)
			if e != nil {
				return nil, e
			}
			want := "a:srgbClr"
			valid := hexColor.MatchString(t.SourceValue)
			if r.ColorKind == "scheme" {
				want = "a:schemeClr"
				valid = schemeColor.MatchString(t.SourceValue)
			}
			if node.Name != want || !valid || node.Attr("val") != t.SourceValue {
				return nil, fmt.Errorf("role %s resource color source mismatch", role)
			}
			if source, ok := profiles["source"]; ok && source[role] != t.SourceValue {
				return nil, fmt.Errorf("role %s resource source profile mismatch", role)
			}
			out = append(out, validatedResourceTarget{Role: role, Target: t})
		}
	}
	return out, nil
}
func resourceProfileChanges(profile string, roles map[string]Role, profiles map[string]map[string]string, targets []validatedResourceTarget) []map[string]any {
	var out []map[string]any
	if profile == "" {
		return out
	}
	for _, v := range targets {
		value := profiles[profile][v.Role]
		if roles[v.Role].ColorKind != "scheme" {
			value = strings.ToUpper(value)
		}
		out = append(out, map[string]any{"role": v.Role, "part": v.Target.Part, "node_path": v.Target.NodePath, "before": v.Target.SourceValue, "after": value, "source_sha256": v.Target.SHA256})
	}
	return out
}
func applyResourceProfile(stage, profile string, roles map[string]Role, profiles map[string]map[string]string, targets []validatedResourceTarget) ([]map[string]any, error) {
	if profile == "" || len(targets) == 0 {
		return nil, nil
	}
	byPart := map[string][]validatedResourceTarget{}
	for _, v := range targets {
		byPart[v.Target.Part] = append(byPart[v.Target.Part], v)
	}
	var changes []map[string]any
	for part, group := range byPart {
		file, e := localResourcePath(stage, part)
		if e != nil {
			return nil, e
		}
		data, e := os.ReadFile(file)
		if e != nil {
			return nil, e
		}
		if hash(data) != strings.ToLower(group[0].Target.SHA256) {
			return nil, fmt.Errorf("resource %s changed during staging", part)
		}
		root, e := nativepkg.Parse(data)
		if e != nil {
			return nil, e
		}
		for _, v := range group {
			if hash(data) != strings.ToLower(v.Target.SHA256) {
				return nil, fmt.Errorf("resource %s has inconsistent target hashes", part)
			}
			n, e := nodeAt(root, v.Target.NodePath)
			if e != nil {
				return nil, e
			}
			if n.Attr("val") != v.Target.SourceValue {
				return nil, fmt.Errorf("resource %s target changed", part)
			}
			value := profiles[profile][v.Role]
			if roles[v.Role].ColorKind != "scheme" {
				value = strings.ToUpper(value)
			}
			n.SetAttr("val", value)
			changes = append(changes, map[string]any{"role": v.Role, "part": part, "node_path": v.Target.NodePath, "before": v.Target.SourceValue, "after": value, "source_sha256": v.Target.SHA256})
		}
		output := root.XML()
		if e := os.WriteFile(file, output, 0644); e != nil {
			return nil, e
		}
		for _, c := range changes {
			if c["part"] == part {
				c["output_sha256"] = hash(output)
			}
		}
	}
	return changes, nil
}
