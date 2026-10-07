package installstate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Finding struct {
	Code   string `json:"code"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}
type Report struct {
	Schema          string    `json:"schema"`
	Root            string    `json:"root"`
	State           State     `json:"state"`
	Findings        []Finding `json:"findings"`
	PATHExecutables []string  `json:"path_executables"`
	Next            string    `json:"next"`
}

// Doctor is read-only. It diagnoses package/skill drift and executable selection
// independently; it does not imply that fonts or PowerPoint have been qualified.
func (c Config) Doctor() (Report, error) {
	c, e := c.normalized()
	r := Report{Schema: "pptxgengo.installation-doctor/v1", Root: c.Root, Findings: []Finding{}, PATHExecutables: []string{}, Next: "Run pptxgengo design render-doctor --json for native PowerPoint and font readiness."}
	if e != nil {
		return r, e
	}
	add := func(code, status, detail string) { r.Findings = append(r.Findings, Finding{code, status, detail}) }
	s, e := loadState(c.Root)
	if e != nil {
		add("state", "error", e.Error())
		return r, nil
	}
	r.State = s
	if _, e = os.Lstat(filepath.Join(c.Root, "pending.json")); e == nil {
		add("activation", "error", "unfinished activation; run installation recover before another install")
	} else if !os.IsNotExist(e) {
		add("activation", "error", e.Error())
	} else {
		add("activation", "ok", "no pending activation")
	}
	if s.Current == nil {
		add("selection", "missing", "no active release recorded")
	} else {
		p, e := Verify(s.Current.Release)
		if e != nil {
			add("package", "error", e.Error())
		} else if p.ManifestSHA256 != s.Current.Package.ManifestSHA256 || p.ContentSHA256 != s.Current.Package.ContentSHA256 {
			add("package", "error", "active manifest differs from recorded selection")
		} else {
			add("package", "ok", fmt.Sprintf("%s %s/%s; %s", p.Version, p.OS, p.Arch, p.Kind))
		}
		if s.Current.SkillDir != "" {
			h, e := treeHash(s.Current.SkillDir)
			if e != nil {
				add("skill", "error", e.Error())
			} else if h != s.Current.SkillSHA256 {
				add("skill", "changed", "installed skill differs; modifications are preserved")
			} else {
				add("skill", "ok", s.Current.SkillDir+" ("+h+")")
			}
		} else {
			add("skill", "not_selected", "CLI-only release or skill installation was skipped")
		}
		if s.Current.Package.Kind == "full" {
			fonts := filepath.Join(s.Current.Release, "library", "wm-design-system", s.Current.Package.Bundle, "fonts")
			matches, _ := filepath.Glob(filepath.Join(fonts, "*.ttf"))
			if len(matches) == 0 {
				add("fonts", "missing", "no packaged TTF fonts")
			} else {
				add("fonts", "not_qualified", fmt.Sprintf("%d packaged TTF fonts; registration and PowerPoint availability require render-doctor", len(matches)))
			}
		} else {
			add("fonts", "not_included", "CLI-only package")
		}
	}
	resolved, e := exec.LookPath(toolName("pptxgengo"))
	if e != nil {
		add("session_path", "missing", "this session cannot resolve pptxgengo")
	} else {
		absolute, _ := filepath.Abs(resolved)
		actual, _ := filepath.EvalSymlinks(absolute)
		expected := filepath.Join(c.Root, "bin", toolName("pptxgengo"))
		if actual == expected {
			add("session_path", "ok", absolute)
		} else {
			add("session_path", "conflict", absolute+"; expected "+expected+". Open a new terminal after activation.")
		}
	}
	seen := map[string]bool{}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, toolName("pptxgengo"))
		if info, e := os.Stat(path); e == nil && !info.IsDir() {
			absolute, _ := filepath.Abs(path)
			key := absolute
			if runtime.GOOS == "windows" {
				key = strings.ToLower(key)
			}
			if !seen[key] {
				r.PATHExecutables = append(r.PATHExecutables, absolute)
				seen[key] = true
			}
		}
	}
	if len(r.PATHExecutables) > 1 {
		add("path_conflicts", "warning", "multiple executable locations; inspect path_executables before removing old entries")
	}
	if runtime.GOOS == "windows" {
		path, e := userPath()
		if e != nil {
			add("user_path", "error", e.Error())
		} else {
			found := false
			for _, part := range strings.Split(path.Value, ";") {
				if strings.EqualFold(filepath.Clean(part), filepath.Join(c.Root, "bin")) {
					found = true
				}
			}
			if found {
				add("user_path", "ok", "stable user PATH entry recorded; already-open shells may be stale")
			} else {
				add("user_path", "missing", "stable bin directory is absent from user PATH")
			}
		}
	}
	return r, nil
}
