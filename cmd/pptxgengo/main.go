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
	"diff": "pptxdiff",
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pptxgengo <template|compose|scene|component|lib|anchor|diff> <command> [flags]")
	fmt.Fprintln(os.Stderr, "       pptxgengo catalog [--print|--open]")
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
	for _, arg := range args {
		if arg == "--root" || strings.HasPrefix(arg, "--root=") {
			return true
		}
	}
	return false
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
		path := filepath.Join(root, "catalog", "index.html")
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("catalog unavailable: %w", err)
		}
		if len(os.Args) == 2 || (len(os.Args) == 3 && os.Args[2] == "--print") {
			fmt.Println(path)
			return nil
		}
		if len(os.Args) == 3 && os.Args[2] == "--open" {
			cmd := exec.Command("open", path)
			return cmd.Run()
		}
		return fmt.Errorf("usage: pptxgengo catalog [--print|--open]")
	}
	if name == "paths" {
		if len(os.Args) != 2 {
			return fmt.Errorf("usage: pptxgengo paths")
		}
		paths := map[string]string{
			"root":    root,
			"library": filepath.Join(root, "library"),
			"scripts": filepath.Join(root, "scripts"),
			"catalog": filepath.Join(root, "catalog", "index.html"),
			"skill":   filepath.Join(root, "skills", "west-monroe-presentations", "SKILL.md"),
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
	if (tool == "pptxtemplate" || tool == "pptxlib") && !hasRoot(args) && len(args) > 0 && !(tool == "pptxtemplate" && args[0] == "apply-accent") {
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
