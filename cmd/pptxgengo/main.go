// pptxgengo dispatches to the tools in a frozen local release.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var version = "dev"

var tools = map[string]string{
	"template": "pptxtemplate", "compose": "pptxcompose", "scene": "pptxscene",
	"component": "pptxcomponent", "lib": "pptxlib", "anchor": "pptxanchor",
	"diff": "pptxdiff", "adapt": "pptxadapt", "design": "pptxdesign",
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pptxgengo <template|compose|scene|component|lib|anchor|diff|adapt|design> <command> [flags]")
	fmt.Fprintln(os.Stderr, "       pptxgengo catalog [--templates|--components|--design-system] [--print|--open]")
	fmt.Fprintln(os.Stderr, "       pptxgengo paths")
	fmt.Fprintln(os.Stderr, "       pptxgengo --version")
}

func releaseRoot() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(resolved)), nil
}

func hasRoot(args []string) bool {
	return hasFlag(args, "--root")
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

func publishedBundle(root string) (string, error) {
	path := filepath.Join(root, "release", "default-bundle.txt")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "v3", nil
	}
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(b))
	if value != "v1" && value != "v2" && value != "v3" && value != "v4" && value != "v5" {
		return "", fmt.Errorf("invalid published bundle in %s: %q", path, value)
	}
	return value, nil
}

func designArgs(root string, input []string) ([]string, error) {
	args := append([]string{}, input...)
	if len(args) == 0 {
		return args, nil
	}
	command := args[0]
	if command != "project" && !hasFlag(args, "--bundle") {
		bundle, err := publishedBundle(root)
		if err != nil {
			return nil, err
		}
		args = append(args, "--bundle", filepath.Join(root, "library", "wm-design-system", bundle))
	}
	indexRead := command == "library-find" || command == "library-inspect" || command == "library-preview"
	if indexRead && !hasFlag(args, "--index") {
		args = append(args, "--index", filepath.Join(root, "catalog", "library.sqlite"))
		if !hasFlag(args, "--gallery") {
			args = append(args, "--gallery", filepath.Join(root, "catalog"))
		}
		if !hasFlag(args, "--legacy-index") {
			args = append(args, "--legacy-index", filepath.Join(root, "library", "catalog-library.sqlite"))
		}
		if !hasFlag(args, "--legacy-root") {
			args = append(args, "--legacy-root", root)
		}
	}
	engineAllowed := command != "library-index" && command != "library-inspect" && command != "library-preview" && command != "project"
	if engineAllowed && !hasFlag(args, "--engine") {
		args = append(args, "--engine", "wmds-go-foundation.v2")
	}
	return args, nil
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return fmt.Errorf("missing command")
	}
	name := os.Args[1]
	if name == "--version" || name == "version" {
		fmt.Println(version)
		return nil
	}
	if name == "--help" || name == "help" {
		usage()
		return nil
	}
	root, err := releaseRoot()
	if err != nil {
		return err
	}
	if name == "catalog" {
		page := "index.html"
		open := false
		seenPage := false
		seenAction := false
		for _, arg := range os.Args[2:] {
			switch arg {
			case "--templates", "--components", "--design-system":
				if seenPage {
					return fmt.Errorf("choose one catalog gallery")
				}
				seenPage = true
				if arg == "--templates" {
					page = "templates.html"
				} else if arg == "--components" {
					page = "components.html"
				} else {
					page = "design-system.html"
				}
			case "--open", "--print":
				if seenAction {
					return fmt.Errorf("choose --open or --print")
				}
				seenAction = true
				open = arg == "--open"
			default:
				return fmt.Errorf("usage: pptxgengo catalog [--templates|--components|--design-system] [--print|--open]")
			}
		}
		path := filepath.Join(root, "catalog", page)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("catalog unavailable: %w", err)
		}
		if !open {
			fmt.Println(path)
			return nil
		}
		cmd := exec.Command("open", path)
		return cmd.Run()
	}
	if name == "paths" {
		if len(os.Args) != 2 {
			return fmt.Errorf("usage: pptxgengo paths")
		}
		bundle, err := publishedBundle(root)
		if err != nil {
			return err
		}
		paths := map[string]string{
			"root":                  root,
			"library":               filepath.Join(root, "library"),
			"scripts":               filepath.Join(root, "scripts"),
			"catalog":               filepath.Join(root, "catalog", "index.html"),
			"catalog_templates":     filepath.Join(root, "catalog", "templates.html"),
			"catalog_components":    filepath.Join(root, "catalog", "components.html"),
			"catalog_design_system": filepath.Join(root, "catalog", "design-system.html"),
			"design_system_v2":      filepath.Join(root, "library", "wm-design-system", "v2"),
			"design_system_v3":      filepath.Join(root, "library", "wm-design-system", "v3"),
			"design_system_v4":      filepath.Join(root, "library", "wm-design-system", "v4"),
			"design_system_v5":      filepath.Join(root, "library", "wm-design-system", "v5"),
			"design_system_default": filepath.Join(root, "library", "wm-design-system", bundle),
			"design_index":          filepath.Join(root, "catalog", "library.sqlite"),
			"project_example":       filepath.Join(root, "examples", "deck-project"),
			"skill":                 filepath.Join(root, "skills", "west-monroe-presentations", "SKILL.md"),
		}
		return json.NewEncoder(os.Stdout).Encode(paths)
	}
	tool := tools[name]
	if tool == "" {
		for _, full := range tools {
			if name == full {
				tool = full
				break
			}
		}
	}
	if tool == "" {
		usage()
		return fmt.Errorf("unknown tool %q", name)
	}
	path := filepath.Join(root, "bin", tool)
	args := append([]string{}, os.Args[2:]...)
	if tool == "pptxdesign" && len(args) > 0 {
		args, err = designArgs(root, args)
		if err != nil {
			return err
		}
	}
	if (tool == "pptxtemplate" || tool == "pptxlib" || tool == "pptxadapt") && !hasRoot(args) && len(args) > 0 && !(tool == "pptxtemplate" && (args[0] == "apply-accent" || args[0] == "apply-gauge")) {
		args = append(args, "--root", root)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = append(os.Environ(), "PPTXGENGO_RELEASE_ROOT="+root)
	if err := cmd.Run(); err != nil {
		if status, ok := err.(*exec.ExitError); ok {
			os.Exit(status.ExitCode())
		}
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "pptxgengo:", err)
		os.Exit(1)
	}
}
