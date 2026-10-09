package nativeexport

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNativeFailurePhaseAndLayer(t *testing.T) {
	for _, item := range []struct {
		message, phase, layer, class string
		code                         int
	}{
		{"native_phase=open_document; native_error_code=-9074; Unknown error (-9074)", "open_document", "powerpoint_document", "document_open_failed", -9074},
		{"Unknown error (-9074)", "unknown", "unknown", "unclassified", -9074},
		{"native_phase=export_pdf; native_error_code=-9074; Unknown error (-9074)", "export_pdf", "powerpoint_document", "pdf_export_failed", -9074},
		{"native_phase=open_document; Not authorized (-1743)", "open_document", "apple_events", "automation_denied", -1743},
		{"native_phase=identify_document; open_identity_timeout", "identify_document", "powerpoint_document", "document_open_or_identity_timeout", 0},
		{"native_phase=close_document; identity_changed", "close_document", "document_identity", "identity_unconfirmed", 0},
		{"file_access_denied: Grant File Access", "unknown", "powerpoint_file_access", "file_access_blocked", 0},
		{"open /documents/file: permission denied", "unknown", "filesystem_access", "filesystem_permission_denied", 0},
		{"context deadline exceeded", "unknown", "unknown", "deadline_exceeded", 0},
	} {
		t.Run(item.class+item.phase, func(t *testing.T) {
			got := classifyNativeFailure(item.message, "task")
			if got.Phase != item.phase || got.Layer != item.layer || got.Classification != item.class || got.ErrorCode != item.code || got.CauseConfirmed || got.TaskID != "task" || got.Message != item.message {
				t.Fatalf("unexpected evidence: %+v", got)
			}
			if _, err := time.Parse(time.RFC3339Nano, got.RecordedAt); err != nil {
				t.Fatal(err)
			}
		})
	}
	message := exportFailure(errors.New("native_phase=open_document; native_error_code=-9074; Unknown error (-9074)"), "/staging/task.pptx").Error()
	if !strings.Contains(message, "during open_document") || !strings.Contains(message, "cause unconfirmed") || strings.Contains(message, "PDF export failed") || strings.Contains(message, "automation_denied") || strings.Contains(message, "file_access_denied") {
		t.Fatal(message)
	}
	message = exportFailure(errors.New("open /source.pptx: permission denied"), "/staging/task.pptx").Error()
	if strings.Contains(message, "grant PowerPoint") || !strings.Contains(message, "filesystem_permission_denied") {
		t.Fatal(message)
	}
}

func TestRenderFailureStructuredEvidence(t *testing.T) {
	out := t.TempDir()
	message := "native_phase=open_document; native_error_code=-9074; Unknown error (-9074)"
	if err := writeRenderFailure(renderFailure{Out: out, Message: message, TaskID: "task"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "render-error.json"))
	if err != nil {
		t.Fatal(err)
	}
	var evidence FailureEvidence
	if err = json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.SchemaVersion != 1 || evidence.Phase != "open_document" || evidence.ErrorCode != -9074 || evidence.CauseConfirmed || evidence.Message != message {
		t.Fatal(evidence)
	}
	text, err := os.ReadFile(filepath.Join(out, "render-error.txt"))
	if err != nil || string(text) != message+"\n" {
		t.Fatal(string(text), err)
	}
}

func TestRenderFailureRefusesStructuredEvidenceSymlink(t *testing.T) {
	out := t.TempDir()
	outside := filepath.Join(t.TempDir(), "preserve.json")
	if err := os.WriteFile(outside, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(out, "render-error.json")); err != nil {
		t.Skip(err)
	}
	if err := writeRenderFailure(renderFailure{Out: out, Message: "failed"}); err == nil || !strings.Contains(err.Error(), "render-error.json is not a regular file") {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "preserve" {
		t.Fatal(string(data), err)
	}
	if _, err = os.Stat(filepath.Join(out, "render-error.txt")); !os.IsNotExist(err) {
		t.Fatal("partial diagnostic written before unsafe target validation", err)
	}
}

func TestExportScriptPreservesPhaseAndErrorNumber(t *testing.T) {
	script := string(exportScript)
	for _, phase := range []string{"prepare_identity", "open_document", "identify_document", "export_pdf", "close_document"} {
		if !strings.Contains(script, `set nativePhase to "`+phase+`"`) {
			t.Fatal("missing phase", phase)
		}
	}
	if !strings.Contains(script, `error "native_phase=" & nativePhase & "; native_error_code=" & errorNumber & "; " & messageText number errorNumber`) || !strings.Contains(script, "set nativePhase to failedPhase") {
		t.Fatal("script loses original export phase/error number during cleanup")
	}
}
