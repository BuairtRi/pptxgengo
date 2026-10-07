package installstate

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// This opt-in fixture executes the actual three source tools. It is unsigned,
// uses CLI-only packages and never changes user PATH, fonts or existing skills.
func TestInstallationRealToolProcesses(t *testing.T) {
	out := os.Getenv("PPTXGENGO_INSTALL_PROCESS_OUT")
	if testing.Short() || out == "" {
		t.Skip("explicit NEW process-qualification output directory required")
	}
	if err := os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(out, "owned workspace with spaces")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(dir, "managed state")
	bin := filepath.Join(dir, "external bin")
	skill := filepath.Join(dir, "user skills", "west-monroe-presentations")
	if err := os.MkdirAll(skill, 0700); err != nil {
		t.Fatal(err)
	}
	userSkill := filepath.Join(skill, "SKILL.md")
	if err := os.WriteFile(userSkill, []byte("original user skill\n"), 0600); err != nil {
		t.Fatal(err)
	}
	beforePath, err := userPath()
	if err != nil {
		t.Fatal(err)
	}
	packages := map[string]Package{}
	build := func(version string) string {
		t.Helper()
		pkg := filepath.Join(dir, "package "+version)
		if err := os.MkdirAll(filepath.Join(pkg, "bin"), 0700); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-ldflags", "-X main.version="+version, "-o", filepath.Join(pkg, "bin")+string(filepath.Separator), "./cmd/pptxgengo", "./cmd/pptxdesign", "./cmd/wmdsdocs")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
		cmd.WaitDelay = 2 * time.Second
		if data, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build actual tools: %v: %s", err, data)
		}
		if err := os.WriteFile(filepath.Join(pkg, "VERSION"), []byte(version+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		hashes := map[string]string{}
		for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
			name := "bin/" + toolName(tool)
			h, err := hashFile(filepath.Join(pkg, filepath.FromSlash(name)))
			if err != nil {
				t.Fatal(err)
			}
			hashes[name] = h
		}
		evidence := map[string]any{"schema": "pptxgengo.build-evidence/v1", "version": version, "target": runtime.GOOS + "-" + runtime.GOARCH, "binaries_sha256": hashes}
		if err := atomicJSON(filepath.Join(pkg, "build-evidence.json"), evidence); err != nil {
			t.Fatal(err)
		}
		p, err := Verify(pkg)
		if err != nil || p.ManagerAPI != 1 || p.Kind != "cli-only" {
			t.Fatal("actual CLI fixture verification", p, err)
		}
		packages[version] = p
		return pkg
	}
	one := build("v0.0.1-process-fixture")
	two := build("v0.0.2-process-fixture")
	bootstrap := filepath.Join(one, "bin", toolName("pptxgengo"))
	launcher := filepath.Join(managed, "bin", toolName("pptxgengo"))
	paths := []string{"--root", managed, "--bin-dir", bin, "--skill-dir", skill}
	steps := []string{}
	run := func(executable string, args ...string) ([]byte, error) {
		t.Helper()
		ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, executable, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "PPTXGENGO_INSTALLATION_ROOT="+managed, "PATH="+filepath.Join(managed, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		cmd.WaitDelay = 2 * time.Second
		return cmd.CombinedOutput()
	}
	command := func(executable, action string, extra ...string) []byte {
		t.Helper()
		args := append([]string{"installation", action}, paths...)
		args = append(args, extra...)
		data, err := run(executable, args...)
		if err != nil {
			t.Fatalf("%s: %v: %s", action, err, data)
		}
		return data
	}
	checkDispatch := func(want string) {
		t.Helper()
		for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
			data, err := run(filepath.Join(managed, "bin", toolName(tool)), "--version")
			if err != nil || strings.TrimSpace(string(data)) != want {
				t.Fatalf("actual %s dispatcher: %v: %s", tool, err, data)
			}
		}
	}
	checkState := func(want string) Report {
		t.Helper()
		var report Report
		if err := json.Unmarshal(command(launcher, "doctor"), &report); err != nil {
			t.Fatal(err)
		}
		if report.State.Current == nil || report.State.Current.Package.Version != want || report.State.Current.Package.Kind != "cli-only" || report.State.Current.SkillDir != "" {
			t.Fatal("wrong actual active state", report)
		}
		pathOK := false
		for _, f := range report.Findings {
			if (f.Code == "package" || f.Code == "activation") && f.Status != "ok" {
				t.Fatal("actual process diagnosed package/activation failure", f)
			}
			if f.Code == "session_path" && f.Status == "ok" {
				pathOK = true
			}
		}
		if !pathOK {
			t.Fatal("fresh controlled process PATH did not resolve managed launcher", report)
		}
		if mustRead(t, userSkill) != "original user skill\n" {
			t.Fatal("CLI-only package changed user skill")
		}
		return report
	}
	command(bootstrap, "install", "--from", one, "--stage-only", "--no-path")
	state, err := loadState(managed)
	if err != nil || state.Current != nil {
		t.Fatal("stage-only activated", state, err)
	}
	if _, err := os.Lstat(filepath.Join(managed, "launchers.json")); !os.IsNotExist(err) {
		t.Fatal("stage-only installed dispatchers", err)
	}
	steps = append(steps, "stage-only-no-activation")
	command(bootstrap, "install", "--from", one, "--no-path")
	checkDispatch("v0.0.1-process-fixture")
	checkState("v0.0.1-process-fixture")
	steps = append(steps, "install-three-real-tool-dispatchers")
	command(launcher, "install", "--from", two, "--no-path")
	checkDispatch("v0.0.2-process-fixture")
	prior := checkState("v0.0.2-process-fixture")
	if prior.State.Previous == nil || prior.State.Previous.Package.Version != "v0.0.1-process-fixture" {
		t.Fatal("upgrade lost predecessor", prior)
	}
	command(launcher, "install", "--from", two, "--no-path")
	repeat := checkState("v0.0.2-process-fixture")
	if !reflect.DeepEqual(prior.State, repeat.State) {
		t.Fatal("repeat install changed immutable selection/history", prior.State, repeat.State)
	}
	steps = append(steps, "upgrade-repeat-install-retains-predecessor")
	command(launcher, "rollback", "--no-path")
	checkDispatch("v0.0.1-process-fixture")
	checkState("v0.0.1-process-fixture")
	steps = append(steps, "rollback-three-real-tool-dispatchers")
	// Repair missing dispatchers must use the original bootstrap bytes. Keep
	// this executable available; never remove or overwrite a running process.
	if err := os.Remove(filepath.Join(managed, "bin", toolName("wmdsdocs"))); err != nil {
		t.Fatal(err)
	}
	command(bootstrap, "install", "--from", one, "--no-path")
	checkDispatch("v0.0.1-process-fixture")
	steps = append(steps, "repair-missing-owned-dispatcher")
	// Give a package consistent hashes/manifest but a false reported version.
	// The real startup probe must reject it without changing the active release.
	bad := filepath.Join(dir, "wrong version package")
	if err := copyTree(two, bad); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bad, "VERSION"), []byte("v0.0.3-process-fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var ev map[string]any
	if err := readJSON(filepath.Join(bad, "build-evidence.json"), &ev); err != nil {
		t.Fatal(err)
	}
	ev["version"] = "v0.0.3-process-fixture"
	if err := atomicJSON(filepath.Join(bad, "build-evidence.json"), ev); err != nil {
		t.Fatal(err)
	}
	args := append([]string{"installation", "install"}, paths...)
	args = append(args, "--from", bad, "--no-path")
	if data, err := run(launcher, args...); err == nil || !strings.Contains(string(data), "unexpected version") {
		t.Fatalf("actual version probe did not refuse false metadata: %v: %s", err, data)
	}
	checkDispatch("v0.0.1-process-fixture")
	checkState("v0.0.1-process-fixture")
	steps = append(steps, "failed-real-startup-preserves-active-release")
	command(launcher, "recover")
	checkDispatch("v0.0.1-process-fixture")
	steps = append(steps, "idempotent-recovery")
	command(launcher, "uninstall")
	state, err = loadState(managed)
	if err != nil || state.Current != nil || mustRead(t, userSkill) != "original user skill\n" {
		t.Fatal("uninstall changed unrelated user skill or retained selection", state, err)
	}
	if _, err := os.Stat(prior.State.Current.Release); err != nil {
		t.Fatal("uninstall deleted retained release", err)
	}
	if data, err := run(launcher, "--version"); err == nil || !strings.Contains(string(data), "no active release") {
		t.Fatalf("inactive launcher unexpectedly executed a release: %v: %s", err, data)
	}
	steps = append(steps, "uninstall-retains-releases-and-original-skill")
	afterPath, err := userPath()
	if err != nil || beforePath != afterPath {
		t.Fatal("qualification changed user PATH", err)
	}
	if err := atomicJSON(filepath.Join(out, "qualification.json"), map[string]any{
		"schema": "pptxgengo.installation-process-qualification/v1",
		"os":     runtime.GOOS, "architecture": runtime.GOARCH, "go_version": runtime.Version(),
		"source_commit": os.Getenv("PPTXGENGO_INSTALL_PROCESS_COMMIT"),
		"created":       time.Now().UTC().Format(time.RFC3339Nano), "steps": steps,
		"packages": packages, "user_path": "unchanged; --no-path", "fonts": "not_changed",
		"scope":               "actual unsigned source tools; synthetic CLI-only metadata; owned paths only",
		"publisher_signature": "not_qualified", "full_presentation_package": "not_qualified",
		"native_office": "not_exercised",
	}); err != nil {
		t.Fatal(err)
	}
}
