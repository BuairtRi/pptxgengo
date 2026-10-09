package nativeexport

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestPhaseForwarderStreamsOnlyCompleteKnownMarkers(t *testing.T) {
	var forwarded bytes.Buffer
	w := &phaseForwarder{target: &forwarded}
	chunks := []string{"private path\n", "native_phase_entered=op", "en_document;\n", "native_phase_entered=invalid;\n", "native_phase_entered=export_pdf;\n"}
	for _, chunk := range chunks {
		if n, err := w.Write([]byte(chunk)); err != nil || n != len(chunk) {
			t.Fatal(n, err)
		}
	}
	if got, want := forwarded.String(), "native_helper_entered=open_document;\nnative_helper_entered=export_pdf;\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if w.output.String() != "private path\nnative_phase_entered=open_document;\nnative_phase_entered=invalid;\nnative_phase_entered=export_pdf;\n" {
		t.Fatal("returned error output was altered")
	}
	e := classifyNativeFailure(helperObservationSummary(forwarded.String())+"context deadline exceeded", "task")
	if e.Phase != "unknown" || e.LastHelperPhase != "export_pdf" || e.CauseConfirmed || e.Classification != "deadline_exceeded" {
		t.Fatal(e)
	}
}

func TestCleanupCannotReplacePrimaryPhaseObservation(t *testing.T) {
	message := "native_last_helper_phase=export_pdf; context deadline exceeded; cleanup unconfirmed: native_last_helper_phase=close_document; native_phase_entered=close_document;"
	e := classifyNativeFailure(message, "task")
	if e.LastHelperPhase != "export_pdf" || e.Phase != "unknown" {
		t.Fatal(e)
	}
	if summary := helperObservationSummary("native_phase_entered=close_document;\n"); summary != "" {
		t.Fatal("helper error text was mistaken for streamed progress", summary)
	}
}

func TestWorkerSeparatesProgressFromSuccessfulJSON(t *testing.T) {
	// The test worker emits an entry marker to stderr before this app-free
	// operation. Successful JSON must contain only its stdout receipt.
	t.Setenv("PPTXGENGO_TEST_WORKER_PHASE", "prepare_identity")
	data, err := workerProcess(context.Background(), workerRequest{Failure: &renderFailure{Out: filepath.Join(t.TempDir(), "out"), Message: "native_phase_entered=open_document; test"}})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]bool
	if err = json.Unmarshal(data, &result); err != nil || !result["recorded"] {
		t.Fatal(string(data), err)
	}
}
