package installstate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fixture(t *testing.T, version string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "package with spaces")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	source := filepath.Join(t.TempDir(), "fixture.go")
	code := "package main\nimport \"fmt\"\nvar version=\"dev\"\nfunc main(){fmt.Println(version)}\n"
	if e := os.WriteFile(source, []byte(code), 0600); e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(t.TempDir(), toolName("fixture"))
	cmd := exec.Command("go", "build", "-buildvcs=false", "-ldflags", "-X main.version="+version, "-o", binary, source)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("fixture build: %v: %s", e, out)
	}
	files := map[string]string{"VERSION": version + "\n", "release/VERSION": version + "\n", "skills/west-monroe-presentations/SKILL.md": "skill " + version, "wmds-docs/site/SOURCE.json": "{}", "library/wm-design-system/v11/bundle.json": "{}", "library/wm-design-system/v11/library.sqlite": "fixture", "library/wm-design-system/v11/catalog/design-system.html": "fixture", "library/wm-design-system/v11/fonts/example.ttf": "fixture"}
	for path, data := range files {
		p := filepath.Join(root, filepath.FromSlash(path))
		if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, []byte(data), 0600); e != nil {
			t.Fatal(e)
		}
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		p := filepath.Join(root, "bin", toolName(tool))
		os.MkdirAll(filepath.Dir(p), 0700)
		if e := copyRegular(binary, p, 0755); e != nil {
			t.Fatal(e)
		}
	}
	hashes, e := inventory(root)
	if e != nil {
		t.Fatal(e)
	}
	manifest := map[string]any{"schema": "pptxgengo.local-release-manifest.v1", "version": version, "target_os": runtime.GOOS, "target_arch": runtime.GOARCH, "selected_bundle": "v11", "files_sha256": hashes, "file_count": len(hashes)}
	if e = atomicJSON(filepath.Join(root, "release-manifest.json"), manifest); e != nil {
		t.Fatal(e)
	}
	return root
}
func config(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	c, e := (Config{Root: filepath.Join(dir, "installation with spaces"), BinDir: filepath.Join(dir, "user bin"), SkillDir: filepath.Join(dir, "user skills", "west-monroe-presentations"), Launcher: exe}).normalized()
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestVerifyRejectsPackageDriftAndExtraFiles(t *testing.T) {
	source := fixture(t, "v1.0.0")
	if _, e := Verify(source); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(source, "unexpected.txt")
	os.WriteFile(path, []byte("extra"), 0600)
	if _, e := Verify(source); e == nil {
		t.Fatal("unlisted file accepted")
	}
	os.Remove(path)
	os.WriteFile(filepath.Join(source, "VERSION"), []byte("v9.9.9\n"), 0600)
	if _, e := Verify(source); e == nil {
		t.Fatal("modified file accepted")
	}
}
func TestInstallUpgradeAndRollbackPreserveUserSkill(t *testing.T) {
	c := config(t)
	os.MkdirAll(c.SkillDir, 0700)
	os.WriteFile(filepath.Join(c.SkillDir, "SKILL.md"), []byte("user skill"), 0600)
	one := fixture(t, "v1.0.0")
	two := fixture(t, "v1.1.0")
	if _, e := c.Install(context.Background(), one, Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	if _, e := c.Install(context.Background(), two, Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	// Reinstalling the same version must retain the rollback target and avoid
	// creating another skill backup when the owned settings already match.
	priorBackups, _ := filepath.Glob(c.SkillDir + ".backup-*")
	if _, e := c.Install(context.Background(), two, Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	afterBackups, _ := filepath.Glob(c.SkillDir + ".backup-*")
	if len(priorBackups) != len(afterBackups) {
		t.Fatal("idempotent install created a backup")
	}
	selected, e := c.Rollback(context.Background(), Options{NoPath: true})
	if e != nil {
		t.Fatal(e)
	}
	if selected.Package.Version != "v1.0.0" {
		t.Fatal(selected)
	}
	if got := mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")); got != "skill v1.0.0" {
		t.Fatal(got)
	}
	backups, _ := filepath.Glob(c.SkillDir + ".backup-*")
	preserved := false
	for _, backup := range backups {
		if mustRead(t, filepath.Join(backup, "SKILL.md")) == "user skill" {
			preserved = true
		}
	}
	if !preserved {
		t.Fatal("original user skill backup lost")
	}
	if runtime.GOOS != "windows" {
		for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
			target, e := os.Readlink(filepath.Join(c.BinDir, tool))
			if e != nil || target != filepath.Join(c.Root, "bin", tool) {
				t.Fatal(tool, target, e)
			}
		}
	}
	if _, e = os.Stat(filepath.Join(c.Root, "pending.json")); !os.IsNotExist(e) {
		t.Fatal("pending journal remains", e)
	}
}
func interrupted(t *testing.T, c Config, source string) transaction {
	t.Helper()
	before, e := loadState(c.Root)
	if e != nil {
		t.Fatal(e)
	}
	selected, e := c.Install(context.Background(), source, Options{StageOnly: true})
	if e != nil {
		t.Fatal(e)
	}
	id := token()
	old, e := optionalSkillHash(c.SkillDir)
	if e != nil {
		t.Fatal(e)
	}
	tx := transaction{Schema: "pptxgengo.activation/v1", Before: before, After: State{Schema: before.Schema, Current: selected, Previous: before.Current}, SkillPath: c.SkillDir, SkillStage: c.SkillDir + ".stage-" + id, SkillBackup: c.SkillDir + ".backup-" + id, BeforeSkillHash: old, AfterSkillHash: selected.Package.SkillSHA256}
	if e = copyTree(filepath.Join(selected.Release, "skills", "west-monroe-presentations"), tx.SkillStage); e != nil {
		t.Fatal(e)
	}
	if e = atomicJSON(filepath.Join(c.Root, "pending.json"), tx); e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(tx.SkillPath, tx.SkillBackup); e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(tx.SkillStage, tx.SkillPath); e != nil {
		t.Fatal(e)
	}
	if e = atomicJSON(filepath.Join(c.Root, "active.json"), tx.After); e != nil {
		t.Fatal(e)
	}
	return tx
}
func TestInterruptedActivationRecoversPredecessor(t *testing.T) {
	c := config(t)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	interrupted(t, c, fixture(t, "v1.1.0"))
	if e := c.Recover(); e != nil {
		t.Fatal(e)
	}
	state, e := loadState(c.Root)
	if e != nil {
		t.Fatal(e)
	}
	if state.Current.Package.Version != "v1.0.0" {
		t.Fatal(state)
	}
	if got := mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")); got != "skill v1.0.0" {
		t.Fatal(got)
	}
	if e = c.Recover(); e != nil {
		t.Fatal("repeated recovery", e)
	}
}
func TestRecoveryPreservesPostInterruptionUserEdits(t *testing.T) {
	c := config(t)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	tx := interrupted(t, c, fixture(t, "v1.1.0"))
	path := filepath.Join(c.SkillDir, "SKILL.md")
	os.WriteFile(path, []byte("user edit after crash"), 0600)
	if e := c.Recover(); e == nil || !strings.Contains(e.Error(), "preserved") {
		t.Fatal("foreign edit not rejected", e)
	}
	if mustRead(t, path) != "user edit after crash" || mustRead(t, filepath.Join(tx.SkillBackup, "SKILL.md")) != "skill v1.0.0" {
		t.Fatal("recovery lost user bytes")
	}
}
func TestStageOnlyAndCorruptionDoNotActivate(t *testing.T) {
	c := config(t)
	source := fixture(t, "v1.0.0")
	if _, e := c.Install(context.Background(), source, Options{StageOnly: true}); e != nil {
		t.Fatal(e)
	}
	if state, e := loadState(c.Root); e != nil || state.Current != nil {
		t.Fatal(state, e)
	}
	if _, e := os.Stat(c.SkillDir); !os.IsNotExist(e) {
		t.Fatal("skill changed")
	}
	os.WriteFile(filepath.Join(source, "VERSION"), []byte("corrupted"), 0600)
	if _, e := c.Install(context.Background(), source, Options{}); e == nil {
		t.Fatal("corrupt package accepted")
	}
	if state, e := loadState(c.Root); e != nil || state.Current != nil {
		t.Fatal(state, e)
	}
}
func TestRecoveryRejectsForeignReceiptPaths(t *testing.T) {
	c := config(t)
	os.MkdirAll(c.Root, 0700)
	foreign := filepath.Join(t.TempDir(), "foreign")
	os.WriteFile(foreign, []byte("user-owned"), 0600)
	state := State{Schema: "pptxgengo.installation/v1"}
	tx := transaction{Schema: "pptxgengo.activation/v1", Before: state, After: state, Links: []LinkChange{{Path: foreign, After: filepath.Join(c.Root, "bin", "pptxgengo")}}}
	if e := atomicJSON(filepath.Join(c.Root, "pending.json"), tx); e != nil {
		t.Fatal(e)
	}
	if e := c.Recover(); e == nil {
		t.Fatal("foreign receipt accepted")
	}
	if mustRead(t, foreign) != "user-owned" {
		t.Fatal("foreign file changed")
	}
}
func TestLauncherSetupResumesAndDoesNotClaimExistingFiles(t *testing.T) {
	c := config(t)
	os.MkdirAll(c.Root, 0700)
	os.MkdirAll(filepath.Join(c.Root, "bin"), 0700)
	h, e := hashFile(c.Launcher)
	if e != nil {
		t.Fatal(e)
	}
	expected := map[string]string{}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		expected[toolName(tool)] = h
	}
	if e = atomicJSON(filepath.Join(c.Root, "launchers.json"), expected); e != nil {
		t.Fatal(e)
	}
	if e = c.ensureLaunchers(); e != nil {
		t.Fatal(e)
	}
	if e = c.ensureLaunchers(); e != nil {
		t.Fatal(e)
	}
	other := config(t)
	os.MkdirAll(filepath.Join(other.Root, "bin"), 0700)
	owned := filepath.Join(other.Root, "bin", toolName("pptxgengo"))
	os.WriteFile(owned, []byte("user-owned"), 0600)
	if e = other.ensureLaunchers(); e == nil {
		t.Fatal("existing file claimed")
	}
	if mustRead(t, owned) != "user-owned" {
		t.Fatal("existing file overwritten")
	}
}
func TestSkillSymlinkIsBackedUpWithoutChangingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows symlink creation needs separate runner policy")
	}
	c := config(t)
	target := t.TempDir()
	os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("external skill"), 0600)
	os.MkdirAll(filepath.Dir(c.SkillDir), 0700)
	os.Symlink(target, c.SkillDir)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	if mustRead(t, filepath.Join(target, "SKILL.md")) != "external skill" {
		t.Fatal("symlink target modified")
	}
	backups, _ := filepath.Glob(c.SkillDir + ".backup-*")
	if len(backups) != 1 {
		t.Fatal(backups)
	}
	if got, e := os.Readlink(backups[0]); e != nil || got != target {
		t.Fatal(got, e)
	}
}
func TestCLIExtraBytesAreBoundToRecordedContent(t *testing.T) {
	source := fixture(t, "v1.0.0")
	os.RemoveAll(filepath.Join(source, "library"))
	os.RemoveAll(filepath.Join(source, "release"))
	os.RemoveAll(filepath.Join(source, "skills"))
	os.RemoveAll(filepath.Join(source, "wmds-docs"))
	os.Remove(filepath.Join(source, "release-manifest.json"))
	hashes := map[string]string{}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		name := "bin/" + toolName(tool)
		h, e := hashFile(filepath.Join(source, filepath.FromSlash(name)))
		if e != nil {
			t.Fatal(e)
		}
		hashes[name] = h
	}
	evidence := map[string]any{"schema": "pptxgengo.build-evidence/v1", "target": runtime.GOOS + "-" + runtime.GOARCH, "version": "v1.0.0", "binaries_sha256": hashes}
	atomicJSON(filepath.Join(source, "build-evidence.json"), evidence)
	os.WriteFile(filepath.Join(source, "README.txt"), []byte("original"), 0600)
	c := config(t)
	if _, e := c.Install(context.Background(), source, Options{StageOnly: true}); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(source, "README.txt"), []byte("changed"), 0600)
	if _, e := c.Install(context.Background(), source, Options{StageOnly: true}); e == nil {
		t.Fatal("different bytes reused an existing version")
	}
}
func TestStateRejectsEscapingRelease(t *testing.T) {
	c := config(t)
	os.MkdirAll(c.Root, 0700)
	s := State{Schema: "pptxgengo.installation/v1", Current: &Selection{Release: t.TempDir(), Package: Package{Version: "v1.0.0"}}}
	b, _ := json.Marshal(s)
	os.WriteFile(filepath.Join(c.Root, "active.json"), b, 0600)
	if _, e := ActiveRelease(c.Root); e == nil {
		t.Fatal("out-of-root active selection accepted")
	}
}

func TestUninstallRestoresUserSkillAndPreservesReleasesAndProjects(t *testing.T) {
	c := config(t)
	os.MkdirAll(c.SkillDir, 0700)
	os.WriteFile(filepath.Join(c.SkillDir, "SKILL.md"), []byte("original user skill"), 0600)
	project := filepath.Join(t.TempDir(), "authored-deck.yaml")
	os.WriteFile(project, []byte("user-authored"), 0600)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	report, e := c.Uninstall(false)
	if e != nil {
		t.Fatal(e)
	}
	if report.Skill != "original_user_skill_restored" {
		t.Fatal(report)
	}
	if mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")) != "original user skill" || mustRead(t, project) != "user-authored" {
		t.Fatal("user resources changed")
	}
	if state, e := loadState(c.Root); e != nil || state.Current != nil || state.Previous != nil {
		t.Fatal(state, e)
	}
	if _, e := Verify(filepath.Join(c.Root, "releases", "v1.0.0")); e != nil {
		t.Fatal("retained release changed", e)
	}
	if _, e = c.Uninstall(false); e != nil {
		t.Fatal("repeat uninstall", e)
	}
	// A later fresh installation must capture the user's current original skill.
	os.WriteFile(filepath.Join(c.SkillDir, "SKILL.md"), []byte("updated user original"), 0600)
	if _, e = c.Install(context.Background(), fixture(t, "v1.1.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	if _, e = c.Uninstall(false); e != nil {
		t.Fatal(e)
	}
	if mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")) != "updated user original" {
		t.Fatal("fresh install restored stale original")
	}
}
func TestUninstallPreservesEditedSkill(t *testing.T) {
	c := config(t)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(c.SkillDir, "SKILL.md"), []byte("user edits"), 0600)
	report, e := c.Uninstall(false)
	if e != nil {
		t.Fatal(e)
	}
	if report.Skill != "user_modifications_preserved" || mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")) != "user edits" {
		t.Fatal(report)
	}
}
func TestUninstallRestoresOriginalSkillSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows symlink creation needs runner policy")
	}
	c := config(t)
	original := t.TempDir()
	os.WriteFile(filepath.Join(original, "SKILL.md"), []byte("external original"), 0600)
	os.MkdirAll(filepath.Dir(c.SkillDir), 0700)
	os.Symlink(original, c.SkillDir)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	if _, e := c.Uninstall(false); e != nil {
		t.Fatal(e)
	}
	if target, e := os.Readlink(c.SkillDir); e != nil || target != original {
		t.Fatal(target, e)
	}
	if mustRead(t, filepath.Join(original, "SKILL.md")) != "external original" {
		t.Fatal("external original changed")
	}
}

func TestDirectoryPromotionRefusesExistingEmptyDestination(t *testing.T) {
	root := t.TempDir()
	stage := filepath.Join(root, "stage")
	dest := filepath.Join(root, "destination")
	os.Mkdir(stage, 0700)
	os.Mkdir(dest, 0700)
	os.WriteFile(filepath.Join(stage, "owned.txt"), []byte("stage"), 0600)
	if e := publishDirectory(stage, dest); e == nil {
		t.Fatal("existing empty destination was replaced")
	}
	if _, e := os.Stat(filepath.Join(stage, "owned.txt")); e != nil {
		t.Fatal("stage lost", e)
	}
	if entries, e := os.ReadDir(dest); e != nil || len(entries) != 0 {
		t.Fatal("destination changed", e)
	}
}

func TestConfigurationRejectsOverlappingUserAndManagedResources(t *testing.T) {
	c := config(t)
	c.SkillDir = filepath.Join(c.Root, "bin")
	if _, e := c.normalized(); e == nil {
		t.Fatal("managed bin directory could be replaced as a skill")
	}
	c = config(t)
	c.SkillDir = c.BinDir
	if _, e := c.normalized(); e == nil {
		t.Fatal("user launcher directory could be replaced as a skill")
	}
}
func TestFailedExecutableProbePreservesPreviousSelection(t *testing.T) {
	c := config(t)
	if _, e := c.Install(context.Background(), fixture(t, "v1.0.0"), Options{NoPath: true}); e != nil {
		t.Fatal(e)
	}
	source := fixture(t, "v1.1.0")
	old := filepath.Join(c.Root, "releases", "v1.0.0", "bin", toolName("pptxgengo"))
	replacement := filepath.Join(source, "bin", toolName("pptxgengo"))
	os.Remove(replacement)
	if e := copyRegular(old, replacement, 0755); e != nil {
		t.Fatal(e)
	}
	var manifest map[string]any
	if e := readJSON(filepath.Join(source, "release-manifest.json"), &manifest); e != nil {
		t.Fatal(e)
	}
	hash, e := hashFile(replacement)
	if e != nil {
		t.Fatal(e)
	}
	manifest["files_sha256"].(map[string]any)["bin/"+toolName("pptxgengo")] = hash
	atomicJSON(filepath.Join(source, "release-manifest.json"), manifest)
	if _, e = c.Install(context.Background(), source, Options{NoPath: true}); e == nil || !strings.Contains(e.Error(), "unexpected version") {
		t.Fatal("startup mismatch accepted", e)
	}
	state, e := loadState(c.Root)
	if e != nil || state.Current.Package.Version != "v1.0.0" {
		t.Fatal(state, e)
	}
	if mustRead(t, filepath.Join(c.SkillDir, "SKILL.md")) != "skill v1.0.0" {
		t.Fatal("prior skill lost")
	}
	stages, _ := filepath.Glob(filepath.Join(c.Root, "releases", ".stage-*"))
	if len(stages) != 0 {
		t.Fatal("failed stage remains", stages)
	}
}
