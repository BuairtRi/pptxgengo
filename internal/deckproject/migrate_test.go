package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestMigrateDryRunCommitAndRepeat(t *testing.T) {
	p := example(t)
	pin(t, p)
	_, old, err := ReadLock(p)
	if err != nil {
		t.Fatal(err)
	}
	target, err := filepath.Abs("../../library/wm-design-system/v11")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Migrate(p, target, wmdesign.CandidateEngine, true)
	if err != nil || r.Status != "ready_to_migrate" || r.Backup != "" {
		t.Fatalf("dry run: %+v %v", r, err)
	}
	_, unchanged, _ := ReadLock(p)
	if !bytes.Equal(old, unchanged) {
		t.Fatal("dry run changed lock")
	}
	r, err = Migrate(p, target, wmdesign.CandidateEngine, false)
	if err != nil || r.Status != "migrated" || r.From != "wmds-library.v5" || r.To != "wmds-library.v11" {
		t.Fatalf("migration: %+v %v", r, err)
	}
	backup, err := os.ReadFile(filepath.Join(p.Root, r.Backup))
	if err != nil || !bytes.Equal(backup, old) {
		t.Fatalf("backup changed: %v", err)
	}
	for relative, raw := range p.SourceFiles {
		got, err := os.ReadFile(filepath.Join(p.Root, relative))
		if err != nil || !bytes.Equal(raw, got) {
			t.Fatalf("source changed: %s %v", relative, err)
		}
	}
	if _, err := Check(p, target, wmdesign.CandidateEngine); err != nil {
		t.Fatal(err)
	}
	r, err = Migrate(p, target, wmdesign.CandidateEngine, false)
	if err != nil || r.Status != "already_current" || r.Backup != "" {
		t.Fatalf("repeat: %+v %v", r, err)
	}
}

func TestMigrateFailuresPreserveLock(t *testing.T) {
	for _, failure := range []string{"busy", "missing-target", "template", "fit"} {
		t.Run(failure, func(t *testing.T) {
			p := example(t)
			pin(t, p)
			_, old, _ := ReadLock(p)
			target, _ := filepath.Abs("../../library/wm-design-system/v11")
			switch failure {
			case "busy":
				if err := os.WriteFile(filepath.Join(p.Root, ".project-build.lock"), []byte("active"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing-target":
				target = filepath.Join(t.TempDir(), "absent")
			case "template":
				p.Document.Slides[0].Template.ID = "missing/template"
			case "fit":
				p.Document.Slides[0].Density = "unknown"
			}
			if _, err := Migrate(p, target, wmdesign.CandidateEngine, false); err == nil {
				t.Fatal("invalid migration succeeded")
			}
			_, unchanged, _ := ReadLock(p)
			if !bytes.Equal(old, unchanged) {
				t.Fatal("failed migration changed lock")
			}
			entries, _ := os.ReadDir(p.Root)
			for _, entry := range entries {
				if strings.Contains(entry.Name(), ".pre-migrate-") || entry.Name() == ".deck-source-mutation.lock" {
					t.Fatalf("failed migration left artifact: %s", entry.Name())
				}
			}
		})
	}
}
