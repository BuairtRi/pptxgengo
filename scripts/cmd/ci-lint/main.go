// ci-lint checks the repository's local GitLab CI graph without credentials.
// GitLab's server lint remains the authoritative rules/matrix simulation.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := validate("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("GitLab CI local configuration and dependency graph valid")
}

func validate(root string) error {
	for _, ext := range []string{"*.yml", "*.yaml"} {
		paths, err := filepath.Glob(filepath.Join(root, ".github", "workflows", ext))
		if err != nil {
			return err
		}
		if len(paths) != 0 {
			return fmt.Errorf("CI must run only in GitLab; GitHub workflow found: %s", paths[0])
		}
	}
	config := map[string]any{}
	seen := map[string]bool{}
	var read func(string) error
	read = func(relative string) error {
		clean := filepath.ToSlash(filepath.Clean(relative))
		if filepath.IsAbs(relative) || clean != relative || strings.HasPrefix(clean, "../") || clean == ".." {
			return fmt.Errorf("unsafe local CI include %q", relative)
		}
		if seen[relative] {
			return fmt.Errorf("duplicate/cyclic CI include %s", relative)
		}
		seen[relative] = true
		file := filepath.Join(root, filepath.FromSlash(relative))
		st, err := os.Stat(file)
		if err != nil {
			return err
		}
		if !st.Mode().IsRegular() || st.Size() > 1<<20 {
			return fmt.Errorf("CI file must be regular and at most 1MiB: %s", relative)
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var doc map[string]any
		if err = yaml.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("%s: %w", relative, err)
		}
		if includes, ok := doc["include"]; ok {
			entries, ok := includes.([]any)
			if !ok {
				return fmt.Errorf("%s: includes must be a list", relative)
			}
			for _, entry := range entries {
				m, ok := entry.(map[string]any)
				name, valid := m["local"].(string)
				if !ok || !valid || len(m) != 1 {
					return fmt.Errorf("only explicit local CI includes are permitted")
				}
				if err := read(name); err != nil {
					return err
				}
			}
		}
		for key, value := range doc {
			if key == "include" {
				continue
			}
			if _, exists := config[key]; exists {
				return fmt.Errorf("duplicate CI definition %s", key)
			}
			config[key] = value
		}
		return nil
	}
	if err := read(".gitlab-ci.yml"); err != nil {
		return err
	}
	stages := map[string]bool{".pre": true, ".post": true}
	stageList, ok := config["stages"].([]any)
	if !ok || len(stageList) == 0 {
		return fmt.Errorf("explicit CI stages required")
	}
	for _, stage := range stageList {
		s, ok := stage.(string)
		if !ok || s == "" || stages[s] {
			return fmt.Errorf("invalid/duplicate stage")
		}
		stages[s] = true
	}
	reserved := map[string]bool{"stages": true, "variables": true, "workflow": true, "default": true, "image": true, "services": true, "cache": true, "before_script": true, "after_script": true}
	jobs := map[string]map[string]any{}
	for name, value := range config {
		if reserved[name] {
			continue
		}
		job, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("job %s must be a mapping", name)
		}
		jobs[name] = job
	}
	stringsOf := func(v any) ([]string, error) {
		if v == nil {
			return nil, nil
		}
		if s, ok := v.(string); ok {
			return []string{s}, nil
		}
		values, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("expected string or list")
		}
		out := []string{}
		for _, value := range values {
			s, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("expected string list")
			}
			out = append(out, s)
		}
		return out, nil
	}
	var property func(string, string, map[string]bool) (any, error)
	property = func(name, key string, stack map[string]bool) (any, error) {
		job, ok := jobs[name]
		if !ok {
			return nil, fmt.Errorf("unknown extended job %s", name)
		}
		if stack[name] {
			return nil, fmt.Errorf("cyclic job inheritance %s", name)
		}
		stack[name] = true
		defer delete(stack, name)
		parents, err := stringsOf(job["extends"])
		if err != nil {
			return nil, err
		}
		var inherited any
		for _, parent := range parents {
			value, err := property(parent, key, stack)
			if err != nil {
				return nil, err
			}
			if value != nil {
				inherited = value
			}
		}
		if value, exists := job[key]; exists {
			return value, nil
		}
		return inherited, nil
	}
	graph := map[string][]string{}
	for name := range jobs {
		stage, err := property(name, "stage", map[string]bool{})
		if err != nil {
			return err
		}
		if stage == nil {
			stage = "test"
		}
		if s, ok := stage.(string); !ok || !stages[s] {
			return fmt.Errorf("job %s has invalid stage", name)
		}
		if strings.HasPrefix(name, ".") {
			continue
		}
		script, err := property(name, "script", map[string]bool{})
		if err != nil {
			return err
		}
		commands, err := stringsOf(script)
		if err != nil || len(commands) == 0 {
			return fmt.Errorf("job %s has no valid script", name)
		}
		needs, err := property(name, "needs", map[string]bool{})
		if err != nil {
			return err
		}
		if needs == nil {
			continue
		}
		entries, ok := needs.([]any)
		if !ok {
			return fmt.Errorf("job %s needs must be a list", name)
		}
		for _, entry := range entries {
			dependency, ok := entry.(string)
			if !ok {
				if m, valid := entry.(map[string]any); valid {
					dependency, ok = m["job"].(string)
				}
			}
			if !ok || jobs[dependency] == nil || strings.HasPrefix(dependency, ".") {
				return fmt.Errorf("job %s has unknown need %v", name, entry)
			}
			graph[name] = append(graph[name], dependency)
		}
	}
	visit := map[string]int{}
	var walk func(string) error
	walk = func(name string) error {
		if visit[name] == 1 {
			return fmt.Errorf("cyclic CI needs at %s", name)
		}
		if visit[name] == 2 {
			return nil
		}
		visit[name] = 1
		for _, dependency := range graph[name] {
			if err := walk(dependency); err != nil {
				return err
			}
		}
		visit[name] = 2
		return nil
	}
	for name := range graph {
		if err := walk(name); err != nil {
			return err
		}
	}
	return nil
}
