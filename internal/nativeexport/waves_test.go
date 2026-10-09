package nativeexport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "render-native-worker" {
		if phase := os.Getenv("PPTXGENGO_TEST_WORKER_PHASE"); phase != "" {
			fmt.Fprintf(os.Stderr, "native_helper_entered=%s;\n", phase)
		}
		if os.Getenv("PPTXGENGO_TEST_BLOCK_WORKER") == "1" {
			child := exec.Command("/bin/sleep", "60")
			if runtime.GOOS == "windows" {
				child = exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "Start-Sleep -Seconds 60")
			}
			child.Stdout = os.Stdout
			child.Stderr = os.Stderr
			if err := child.Start(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if ready := os.Getenv("PPTXGENGO_TEST_WORKER_READY"); ready != "" {
				if err := os.WriteFile(ready, []byte(fmt.Sprintf("%d\n", child.Process.Pid)), 0600); err != nil {
					_ = child.Process.Kill()
					_ = child.Wait()
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			_ = child.Wait()
			os.Exit(0)
		}
		if err := RenderWorker(context.Background(), os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// Hermetic successful render tests must never create a real operator's key.
	home, err := os.MkdirTemp("", "pptxgengo-native-test-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if os.Getenv("PPTXGENGO_NATIVE_LIVE_OUT") == "" {
		_ = os.Setenv("HOME", home)
		_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
		if runtime.GOOS == "windows" {
			_ = os.Setenv("USERPROFILE", home)
			_ = os.Setenv("APPDATA", filepath.Join(home, "config"))
			_ = os.Setenv("LOCALAPPDATA", filepath.Join(home, "cache"))
		}
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

func TestParseSlideSelection(t *testing.T) {
	got, err := ParseSlides("5-7,3,5", 8)
	if err != nil || !reflect.DeepEqual(got, []int{3, 5, 6, 7}) {
		t.Fatal(got, err)
	}
	for _, selector := range []string{"0", "9", "2-1", "1-", "1,,2", "1-2-3", "-1", "1e2"} {
		if _, err := ParseSlides(selector, 8); err == nil {
			t.Errorf("accepted %q", selector)
		}
	}
}

func TestSelectedReviewHiddenIdentity(t *testing.T) {
	original := fixture(t)
	if _, _, _, _, err := selectedReview(original, "1", false); err == nil {
		t.Fatal("all-hidden selection accepted")
	}
	for _, include := range []bool{false, true} {
		review, total, visible, mapping, err := selectedReview(original, "2", include)
		if err != nil || total != 2 || len(mapping) != 1 || mapping[0].SourceSlide != 2 || mapping[0].SourcePart != "ppt/slides/slide2.xml" || mapping[0].SourceHidden || len(visible) != 0 {
			t.Fatal(total, visible, mapping, err)
		}
		if bytes.Equal(review, original) {
			t.Fatal("selection did not change review copy")
		}
		_, selectedCount, hidden, err := reviewCopy(review, false)
		if err != nil || selectedCount != 1 || len(hidden) != 0 {
			t.Fatal(selectedCount, hidden, err)
		}
		// Hidden orphan stays byte-for-byte unchanged; it was not selected.
		if !bytes.Equal(parts(t, original)["ppt/slides/slide1.xml"], parts(t, review)["ppt/slides/slide1.xml"]) {
			t.Fatal("unselected hidden slide changed")
		}
	}
	_, _, visible, mapping, err := selectedReview(original, "1", true)
	if err != nil || len(visible) != 1 || len(mapping) != 1 || !mapping[0].SourceHidden || mapping[0].SourceSlide != 1 {
		t.Fatal(visible, mapping, err)
	}
}

func TestCanonicalMissingOutputUnderSymlink(t *testing.T) {
	real := t.TempDir()
	parent := t.TempDir()
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	got, err := canonicalPath(filepath.Join(alias, "new", "out"))
	expected, _ := filepath.EvalSymlinks(real)
	if err != nil || got != filepath.Join(expected, "new", "out") {
		t.Fatal(got, err)
	}
}

func TestRenderSelectedContactSheetAndStaging(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.pptx")
	original := fixture(t)
	if err := os.WriteFile(source, original, 0600); err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(dir, "stage with spaces")
	canonicalStaging, err := canonicalPath(staging)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "/usr/bin/osascript" {
			if !strings.HasPrefix(args[1], canonicalStaging+string(filepath.Separator)) {
				t.Fatal("copy outside configured staging", args)
			}
			data, err := os.ReadFile(args[1])
			if err != nil {
				return nil, err
			}
			_, count, _, err := reviewCopy(data, false)
			if err != nil || count != 1 {
				t.Fatal(count, err)
			}
			return nil, os.WriteFile(args[2], []byte("%PDF-1.7\nfixture"), 0600)
		}
		if name != "/usr/bin/swift" || args[3] != "png" {
			t.Fatal(name, args)
		}
		if err := os.MkdirAll(args[2], 0700); err != nil {
			return nil, err
		}
		f, err := os.Create(filepath.Join(args[2], "slide-001.png"))
		if err != nil {
			return nil, err
		}
		err = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 32, 18)))
		f.Close()
		return []byte(`{"pages":1}`), err
	}
	receipt, err := render(context.Background(), Options{PPTX: source, Out: out, PNG: true, Slides: "2", ContactSheet: true, StagingRoot: staging, Timeout: time.Minute}, run, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Pages != 1 || receipt.Slides != 2 || receipt.PNGs[0].Path != "native-pages/slide-002.png" || receipt.PageMappings[0].SourceSlide != 2 || receipt.ContactSheet == nil || receipt.ContactSheet.Width != 420 {
		t.Fatal(receipt)
	}
	entries, err := os.ReadDir(staging)
	if err != nil || len(entries) != 0 {
		t.Fatal("task staging retained", entries, err)
	}
	after, _ := os.ReadFile(source)
	if !bytes.Equal(after, original) {
		t.Fatal("source changed")
	}
	if _, err := os.Stat(filepath.Join(out, "native-pages", "slide-001.png")); !os.IsNotExist(err) {
		t.Fatal("wrong source slide filename")
	}
}

func TestNativeErrorClassifications(t *testing.T) {
	for _, item := range []struct{ message, want string }{{"Not authorized (-1743)", "automation_denied"}, {"file_access_denied", "grant PowerPoint access"}, {"identity_ambiguous", "multiple presentations"}, {"open_identity_timeout", "unconfirmed"}, {"context deadline exceeded", "unconfirmed"}} {
		if got := exportFailure(errors.New(item.message), "/task.pptx").Error(); !strings.Contains(got, item.want) {
			t.Fatal(got)
		}
	}
	checks := doctor(context.Background(), DoctorOptions{}, nil, "linux")
	if len(checks) != 1 || checks[0].Status != "fail" {
		t.Fatal(checks)
	}
}

func TestDoctorUsesOperationalEvent(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "/usr/bin/osascript" {
			if !strings.Contains(strings.Join(args, " "), "count presentations") {
				t.Fatal("metadata event cannot qualify automation", args)
			}
			return nil, errors.New("An error of type -10827 has occurred")
		}
		return []byte("available"), nil
	}
	checks := doctor(context.Background(), DoctorOptions{StagingRoot: t.TempDir()}, run, "darwin")
	foundDispatch, foundUnknown := false, false
	for _, check := range checks {
		if check.Check == "powerpoint-dispatch" && check.Status == "fail" {
			foundDispatch = true
		}
		if check.Check == "automation" && check.Status == "unknown" {
			foundUnknown = true
		}
	}
	if !foundDispatch || !foundUnknown {
		t.Fatal(checks)
	}
}

func TestWorkerDeadlineDoesNotPublishReceipt(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("native worker deadline is macOS-only")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	_, err := Render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: time.Nanosecond, StagingRoot: filepath.Join(root, "stage")})
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "render-manifest.json")); !os.IsNotExist(err) {
		t.Fatal("deadline issued success receipt")
	}
	data, readErr := os.ReadFile(filepath.Join(out, "render-error.txt"))
	if readErr != nil {
		// Failure recording has its own bounded worker. Under startup or
		// filesystem contention it can expire as well; the returned error
		// must disclose that failure instead of claiming a recorded artifact.
		if !os.IsNotExist(readErr) || !strings.Contains(err.Error(), "render-error.txt could not be recorded:") {
			t.Fatal("deadline diagnostic neither recorded nor reported unavailable", string(data), readErr, err)
		}
	} else if !strings.Contains(string(data), "deadline") {
		t.Fatal("deadline diagnostic omitted deadline", string(data), err)
	}
}

func TestWorkerDeadlineReportsDiagnosticRecordingFailure(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("native worker deadline is macOS-only")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	if err := os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "preserve.txt")
	const marker = "existing unrelated content"
	if err := os.WriteFile(outside, []byte(marker), 0600); err != nil {
		t.Fatal(err)
	}
	// A nonregular diagnostic target deterministically prevents recording,
	// independent of whether auxiliary-worker startup meets its budget.
	if err := os.Symlink(outside, filepath.Join(out, "render-error.txt")); err != nil {
		t.Fatal(err)
	}
	receipt, err := Render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: time.Nanosecond, StagingRoot: filepath.Join(root, "stage")})
	if receipt != nil || err == nil || !strings.Contains(err.Error(), "deadline") || !strings.Contains(err.Error(), "render-error.txt could not be recorded:") {
		t.Fatal(receipt, err)
	}
	if _, err := os.Stat(filepath.Join(out, "render-manifest.json")); !os.IsNotExist(err) {
		t.Fatal("failed diagnostic recording published success receipt", err)
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != marker {
		t.Fatal("diagnostic worker modified unrelated target", string(data), err)
	}
}

func TestWorkerDeadlineStopsHelperDescendants(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("process group termination is native macOS behavior")
	}
	t.Setenv("PPTXGENGO_TEST_BLOCK_WORKER", "1")
	t.Setenv("PPTXGENGO_TEST_WORKER_PHASE", "export_pdf")
	ready := filepath.Join(t.TempDir(), "worker-ready")
	t.Setenv("PPTXGENGO_TEST_WORKER_READY", ready)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := workerProcess(ctx, workerRequest{Doctor: &DoctorOptions{}})
		result <- err
	}()
	// Start cancellation only after the worker has started the pipe-owning
	// descendant. A short initial deadline could pass before child startup on
	// a cold or busy race runner and never exercise process-group termination.
	readyDeadline := time.NewTimer(3 * time.Second)
	defer readyDeadline.Stop()
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		select {
		case err := <-result:
			t.Fatalf("worker exited before helper readiness: %v", err)
		case <-readyDeadline.C:
			t.Fatal("worker did not confirm helper readiness within 3s")
		case <-poll.C:
		}
	}
	started := time.Now()
	cancel()
	err := <-result
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	evidence := classifyNativeFailure(err.Error(), "")
	if evidence.Phase != "unknown" || evidence.LastHelperPhase != "export_pdf" || evidence.CauseConfirmed {
		t.Fatalf("deadline lost entry observation or invented failing phase: %+v", evidence)
	}
	// A surviving helper would hold the output pipe open until WaitDelay (2s).
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("worker descendant retained its output pipe after cancellation: %s", elapsed)
	}
}

func TestWorkerRejectsUnsafeTaskIdentity(t *testing.T) {
	input := strings.NewReader(`{"cleanup":{"StagingRoot":"/tmp","TaskID":"../../other"}}`)
	var output bytes.Buffer
	if err := RenderWorker(context.Background(), input, &output); err == nil {
		t.Fatal("unsafe cleanup identity accepted")
	}
}

func TestCleanupRejectsTaskDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	taskID, err := newTaskID()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(root, taskID)); err != nil {
		t.Fatal(err)
	}
	if err = cleanupExactTask(context.Background(), cleanupTask{StagingRoot: root, TaskID: taskID}); err == nil {
		t.Fatal("cleanup followed task directory symlink")
	}
	if _, err = os.Stat(outside); err != nil {
		t.Fatal("cleanup changed outside directory", err)
	}
}

func TestRenderRetainsIdentityWhenCloseUnconfirmed(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(root, "stage")
	taskPath := ""
	calls := 0
	run := func(_ context.Context, _ string, args ...string) ([]byte, error) {
		calls++
		taskPath = args[1]
		return nil, errors.New("PowerPoint unavailable")
	}
	_, err := render(context.Background(), Options{PPTX: source, Out: filepath.Join(root, "out"), PDF: true, Timeout: time.Second, StagingRoot: staging}, run, "darwin")
	if err == nil || !strings.Contains(err.Error(), "task copy retained") || calls != 2 {
		t.Fatal(err, calls)
	}
	if _, err = os.Stat(taskPath); err != nil {
		t.Fatal("exact task identity removed before confirmed close", err)
	}
}

// Opt-in integration check; ordinary go test never opens a GUI application.
func TestNativeLiveSmoke(t *testing.T) {
	root := os.Getenv("PPTXGENGO_NATIVE_LIVE_OUT")
	if root == "" {
		t.Skip("set PPTXGENGO_NATIVE_LIVE_OUT to a new output location for a real PowerPoint check")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "hidden-source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	receipt, err := Render(context.Background(), Options{PPTX: source, Out: filepath.Join(root, "render"), PDF: true, PNG: true, IncludeHidden: true, Slides: "1-2", ContactSheet: true, StagingRoot: os.Getenv("PPTXGENGO_NATIVE_LIVE_STAGING"), Timeout: 45 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Pages != 2 || len(receipt.PageMappings) != 2 || !receipt.PageMappings[0].SourceHidden || receipt.PageMappings[1].SourceHidden {
		t.Fatal(receipt)
	}
	t.Logf("Native artifacts: %s", filepath.Join(root, "render"))
}
