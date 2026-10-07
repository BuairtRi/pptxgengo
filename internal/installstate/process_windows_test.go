package installstate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/buairtri/pptxgengo/internal/powershellenv"
)

// Script execution uses full-format synthetic resources with real unsigned tools.
// LOCALAPPDATA is owned, while -NoPath/-SkipSkill/-SkipFonts preserve user settings.
func qualifyWindowsInstallerScript(t *testing.T, repo, work, binaries string) []string {
	t.Helper()
	beforePath, err := userPath()
	if err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(work, "synthetic presentation package")
	if err := copyTree(binaries, full); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(full, "build-evidence.json")); err != nil {
		t.Fatal(err)
	}
	resources := map[string]string{
		"release/VERSION": "v0.0.1-process-fixture\n",
		"skills/west-monroe-presentations/SKILL.md":               "synthetic qualification skill; never activated\n",
		"wmds-docs/site/SOURCE.json":                              "{}\n",
		"library/wm-design-system/v11/bundle.json":                "{}\n",
		"library/wm-design-system/v11/library.sqlite":             "synthetic resource; never used for rendering\n",
		"library/wm-design-system/v11/catalog/design-system.html": "<p>synthetic qualification resource</p>\n",
	}
	for name, data := range resources {
		path := filepath.Join(full, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := copyRegular(filepath.Join(repo, "internal", "releasepackage", "install-windows.ps1"), filepath.Join(full, "install-windows.ps1"), 0600); err != nil {
		t.Fatal(err)
	}
	hashes, err := inventory(full)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := Verify(binaries)
	if err != nil {
		t.Fatal(err)
	}
	if err := atomicJSON(filepath.Join(full, "release-manifest.json"), map[string]any{
		"schema": "pptxgengo.local-release-manifest.v1", "version": "v0.0.1-process-fixture",
		"target_os": "windows", "target_arch": verified.Arch, "selected_bundle": "v11",
		"files_sha256": hashes, "file_count": len(hashes),
	}); err != nil {
		t.Fatal(err)
	}
	appdata := filepath.Join(work, "isolated Windows app data")
	destination := filepath.Join(work, "checked script destination")
	if err := os.Mkdir(appdata, 0700); err != nil {
		t.Fatal(err)
	}
	run := func(stage bool) {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
		defer cancel()
		args := []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(full, "install-windows.ps1"), "-Destination", destination, "-NoPath", "-SkipSkill", "-SkipFonts"}
		if stage {
			args = append(args, "-StageOnly")
		}
		cmd := exec.CommandContext(ctx, "powershell.exe", args...)
		cmd.Dir = work
		cmd.Env = powershellenv.ForWindowsPowerShell(append(os.Environ(), "LOCALAPPDATA="+appdata, "PPTXGENGO_INSTALLATION_ROOT="))
		cmd.WaitDelay = 2 * time.Second
		if data, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("actual PowerShell installer stage=%v: %v: %s", stage, err, data)
		}
	}
	run(true)
	managed := filepath.Join(appdata, "pptxgengo")
	state, err := loadState(managed)
	if err != nil || state.Current != nil {
		t.Fatal("PowerShell stage-only activated", state, err)
	}
	if _, err := Verify(destination); err != nil {
		t.Fatal("PowerShell stage destination is not a verified package", err)
	}
	run(false)
	state, err = loadState(managed)
	if err != nil || state.Current == nil || state.Current.Package.Version != "v0.0.1-process-fixture" || state.Current.Package.Kind != "full" || state.Current.SkillDir != "" {
		t.Fatal("PowerShell activation did not honor fixture/skill opt-out", state, err)
	}
	prior := state
	run(false)
	state, err = loadState(managed)
	if err != nil || !reflect.DeepEqual(prior, state) {
		t.Fatal("repeated PowerShell activation changed selection", state, err)
	}
	afterPath, err := userPath()
	if err != nil || beforePath != afterPath {
		t.Fatal("PowerShell -NoPath changed user registry PATH", err)
	}
	if _, err := os.Lstat(filepath.Join(managed, "pending.json")); !os.IsNotExist(err) {
		t.Fatal("PowerShell fixture left pending activation", err)
	}
	return []string{"real-powershell-installer-stage-only", "real-powershell-no-path-skip-skill-skip-fonts-activation", "real-powershell-repeated-activation"}
}
