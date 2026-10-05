package nativeexport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stagingTestDirectory(t *testing.T) string {
	t.Helper()
	path, err := canonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDoctorAndRenderUseSameAuthorizationFolder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := stagingTestDirectory(t)
	root := filepath.Join(dir, "stable staging")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	grantedFolder := ""
	var doctorCopy, renderCopy string
	doctorRunner := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name != "/usr/bin/osascript" || args[5] != "export" {
			t.Fatal("unexpected doctor command", name, args)
		}
		if filepath.Dir(args[1]) != root || filepath.Dir(args[2]) != root || filepath.Dir(args[0]) == root {
			t.Fatal("probe authorization folder differs from actual output folder", args)
		}
		grantedFolder = filepath.Dir(args[1])
		doctorCopy = args[1]
		return nil, os.WriteFile(args[2], []byte("%PDF-1.7\ndiagnostic"), 0600)
	}
	check := doctorFileAccess(context.Background(), root, "", doctorRunner)
	if check.Status != "pass" || !strings.Contains(check.Detail, "same stable staging folder") {
		t.Fatal(check)
	}
	source := filepath.Join(dir, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	fake := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "/usr/bin/swift" {
			return []byte(`{"pages":1}`), nil
		}
		if filepath.Dir(args[1]) != grantedFolder || filepath.Dir(args[2]) != grantedFolder {
			return nil, errors.New("file_access_denied: new private folder is unqualified")
		}
		if args[5] != "export" || len(args) != 7 {
			t.Fatal("doctor/render export contract differs", args)
		}
		renderCopy = args[1]
		return nil, os.WriteFile(args[2], []byte("%PDF-1.7\nrender"), 0600)
	}
	receipt, err := render(context.Background(), Options{PPTX: source, Out: filepath.Join(dir, "out"), PDF: true, Slides: "1", IncludeHidden: true, Timeout: time.Second, StagingRoot: root}, fake, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if doctorCopy == renderCopy || filepath.Dir(renderCopy) != root || receipt.StagingPath != root {
		t.Fatal("task names not unique, or staging receipt mismatched", doctorCopy, renderCopy, receipt)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("confirmed tasks retained staging artifacts", entries, err)
	}
}

func TestStableStagingPreservesCollidingFiles(t *testing.T) {
	for _, kind := range []string{"pptx", "pdf", "pptx-symlink", "pdf-symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := stagingTestDirectory(t)
			root := filepath.Join(dir, "stage")
			if err := os.Mkdir(root, 0700); err != nil {
				t.Fatal(err)
			}
			id, err := newTaskID()
			if err != nil {
				t.Fatal(err)
			}
			input, pdf := taskPresentationPaths(root, id)
			collision := input
			if strings.HasPrefix(kind, "pdf") {
				collision = pdf
			}
			outside := filepath.Join(dir, "foreign-file")
			foreign := []byte("preserve existing bytes")
			if err = os.WriteFile(outside, foreign, 0600); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(kind, "symlink") {
				err = os.Symlink(outside, collision)
			} else {
				err = os.WriteFile(collision, foreign, 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(dir, "source.pptx")
			if err = os.WriteFile(source, fixture(t), 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			fake := func(context.Context, string, ...string) ([]byte, error) { calls++; return nil, nil }
			_, err = render(context.Background(), Options{PPTX: source, Out: filepath.Join(dir, "out"), PDF: true, StagingRoot: root, Timeout: time.Second, taskID: id}, fake, "darwin")
			if err == nil || calls != 0 {
				t.Fatal("collision reached PowerPoint", err, calls)
			}
			current, readErr := os.ReadFile(collision)
			if readErr != nil || string(current) != string(foreign) {
				t.Fatal("collision deleted or overwritten", readErr, string(current))
			}
			current, readErr = os.ReadFile(outside)
			if readErr != nil || string(current) != string(foreign) {
				t.Fatal("foreign symlink target changed", readErr)
			}
			if _, err = os.Stat(filepath.Join(root, id)); !os.IsNotExist(err) {
				t.Fatal("failed preflight retained private work", err)
			}
		})
	}
}

func TestExactCleanupStableTaskOwnership(t *testing.T) {
	for _, scenario := range []string{"success", "close-failure", "replaced-pptx", "symlink-pptx", "symlink-pdf"} {
		t.Run(scenario, func(t *testing.T) {
			root := stagingTestDirectory(t)
			id, err := newTaskID()
			if err != nil {
				t.Fatal(err)
			}
			work := filepath.Join(root, id)
			if err = os.Mkdir(work, 0700); err != nil {
				t.Fatal(err)
			}
			input, pdf, err := acquireTaskFiles(root, id, []byte("owned presentation"))
			if err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "foreign")
			if err = os.WriteFile(outside, []byte("foreign"), 0600); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "replaced-pptx":
				os.Remove(input)
				err = os.WriteFile(input, []byte("foreign"), 0600)
			case "symlink-pptx":
				os.Remove(input)
				err = os.Symlink(outside, input)
			case "symlink-pdf":
				os.Remove(pdf)
				err = os.Symlink(outside, pdf)
			}
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			fake := func(_ context.Context, _ string, args ...string) ([]byte, error) {
				calls++
				if args[1] != input || args[5] != "close" || args[6] != filepath.Join(work, taskIdentityFile) {
					t.Fatal("cleanup changed exact identity", args)
				}
				if scenario == "close-failure" {
					return nil, errors.New("close failed")
				}
				return nil, nil
			}
			err = cleanupExactTaskWithRunner(context.Background(), cleanupTask{StagingRoot: root, TaskID: id}, fake)
			if scenario == "success" {
				if err != nil || calls != 1 {
					t.Fatal(err, calls)
				}
				entries, _ := os.ReadDir(root)
				if len(entries) != 0 {
					t.Fatal("confirmed task not removed", entries)
				}
			} else {
				if err == nil {
					t.Fatal("unconfirmed cleanup reported success")
				}
				if scenario == "replaced-pptx" || scenario == "symlink-pptx" {
					if calls != 0 {
						t.Fatal("substituted presentation reached PowerPoint")
					}
				}
				if _, err = os.Lstat(input); err != nil {
					t.Fatal("unconfirmed task removed", err)
				}
			}
			foreign, _ := os.ReadFile(outside)
			if string(foreign) != "foreign" {
				t.Fatal("cleanup touched foreign file")
			}
		})
	}
}

func TestExactCleanupMissingMetadataKeepsStableFiles(t *testing.T) {
	for _, extension := range []string{"pptx", "pdf"} {
		t.Run(extension, func(t *testing.T) {
			root := stagingTestDirectory(t)
			id, err := newTaskID()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, id+"."+extension)
			if err = os.WriteFile(path, []byte("unconfirmed"), 0600); err != nil {
				t.Fatal(err)
			}
			fake := func(context.Context, string, ...string) ([]byte, error) {
				t.Fatal("unowned cleanup called PowerPoint")
				return nil, nil
			}
			if err = cleanupExactTaskWithRunner(context.Background(), cleanupTask{StagingRoot: root, TaskID: id}, fake); err == nil {
				t.Fatal("metadata missing but cleanup reported success")
			}
			if _, err = os.Stat(path); err != nil {
				t.Fatal("unowned task file deleted", err)
			}
		})
	}
}

func TestExactCleanupSupportsRetainedLegacyTask(t *testing.T) {
	root := stagingTestDirectory(t)
	id, err := newTaskID()
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, id)
	if err = os.Mkdir(work, 0700); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(work, id+".pptx")
	if err = os.WriteFile(input, []byte("legacy"), 0600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	fake := func(_ context.Context, _ string, args ...string) ([]byte, error) {
		calls++
		if args[1] != input || len(args) != 6 || args[5] != "close" {
			t.Fatal("legacy identity changed", args)
		}
		return nil, nil
	}
	if err = cleanupExactTaskWithRunner(context.Background(), cleanupTask{StagingRoot: root, TaskID: id}, fake); err != nil || calls != 1 {
		t.Fatal(err, calls)
	}
	if _, err = os.Stat(work); !os.IsNotExist(err) {
		t.Fatal("legacy confirmed task retained", err)
	}
}
