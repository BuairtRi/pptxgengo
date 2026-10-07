package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/buairtri/pptxgengo/internal/installstate"
)

func runInstallation(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pptxgengo installation install|rollback|recover|doctor|uninstall [flags]")
	}
	command := args[0]
	if command != "install" && command != "rollback" && command != "recover" && command != "doctor" && command != "uninstall" {
		return fmt.Errorf("unknown installation command %q", command)
	}
	c, e := installstate.Defaults()
	if e != nil {
		return e
	}
	f := flag.NewFlagSet("installation "+command, flag.ContinueOnError)
	root := f.String("root", c.Root, "user installation state and immutable release directories")
	bin := f.String("bin-dir", c.BinDir, "Unix user launcher directory; Windows uses ROOT/bin")
	skill := f.String("skill-dir", c.SkillDir, "user presentation skill directory (backups preserved)")
	from := f.String("from", "", "extracted, authenticated release package")
	stage := f.Bool("stage-only", false, "verify and copy without activating")
	skip := f.Bool("skip-skill", false, "leave the existing user skill untouched")
	keepSkill := f.Bool("keep-skill", false, "uninstall preserves every user skill")
	noPath := f.Bool("no-path", false, "leave PATH and external launchers untouched")
	if e = f.Parse(args[1:]); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	invalid := ""
	f.Visit(func(v *flag.Flag) {
		if v.Name == "keep-skill" && command != "uninstall" {
			invalid = v.Name
		}
		if command != "install" && (v.Name == "from" || v.Name == "stage-only" || v.Name == "skip-skill") || (command == "doctor" || command == "recover" || command == "uninstall") && v.Name == "no-path" {
			invalid = v.Name
		}
	})
	if invalid != "" {
		return fmt.Errorf("%s does not accept --%s", command, invalid)
	}
	c.Root, c.BinDir, c.SkillDir = *root, *bin, *skill
	o := installstate.Options{SkipSkill: *skip, NoPath: *noPath, StageOnly: *stage}
	switch command {
	case "install":
		if *from == "" {
			return fmt.Errorf("installation install requires --from DIR; authenticate the signed release manifest first")
		}
		selection, e := c.Install(context.Background(), *from, o)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"selection": selection, "staged_only": *stage, "activation": !(*stage), "fonts": "not_changed; use bundled installer or native font diagnostics", "next": "Open a new terminal and run pptxgengo installation doctor"})
	case "rollback":
		selection, e := c.Rollback(context.Background(), o)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"selection": selection, "fonts": "not_changed", "next": "Open a new terminal and run pptxgengo installation doctor"})
	case "recover":
		if e = c.Recover(); e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"recovery": "complete; previous selection and owned settings restored"})
	case "uninstall":
		report, e := c.Uninstall(*keepSkill)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	case "doctor":
		report, e := c.Doctor()
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	return nil
}

// The immutable dispatcher selects the active release at invocation time. Renamed
// copies provide all three tools on Windows without symlink privileges or EXE
// replacement during upgrades. Actual package binaries have no active.json.
func runManagedLauncher(args []string) (bool, error) {
	exe, e := os.Executable()
	if e != nil {
		return false, e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return false, e
	}
	root := filepath.Dir(filepath.Dir(exe))
	if _, e = os.Lstat(filepath.Join(root, "launchers.json")); os.IsNotExist(e) {
		return false, nil
	} else if e != nil {
		return true, e
	}
	release, e := installstate.ActiveRelease(root)
	if e != nil {
		return true, e
	}
	if release == "" {
		return true, fmt.Errorf("no active release; use installation install or recover")
	}
	tool := strings.TrimSuffix(filepath.Base(exe), ".exe")
	if tool != "pptxgengo" && tool != "pptxdesign" && tool != "wmdsdocs" {
		return true, fmt.Errorf("unknown managed launcher")
	}
	path := filepath.Join(release, "bin", toolFilename(tool, runtime.GOOS))
	command := exec.Command(path, args...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	command.Env = append(os.Environ(), "PPTXGENGO_INSTALLATION_ROOT="+root)
	if e = command.Run(); e != nil {
		if exit, ok := e.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		return true, e
	}
	return true, nil
}

func runManagedInstallation(args []string) (bool, error) {
	exe, e := os.Executable()
	if e != nil {
		return false, e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return false, e
	}
	root := filepath.Dir(filepath.Dir(exe))
	if _, e = os.Lstat(filepath.Join(root, "launchers.json")); e != nil {
		return false, nil
	}
	api, e := installstate.ActiveManagerAPI(root)
	// The bootstrap's own doctor/recovery remain available if active state is
	// corrupt. They report the error and preserve user settings.
	if e != nil || api == 0 {
		return false, nil
	}
	return runManagedLauncher(args)
}
