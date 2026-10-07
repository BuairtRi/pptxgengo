package nativeexport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testRoundTripPlan() RoundTripPlan {
	p := RoundTripPlan{DeckToken: strings.Repeat("A", 64), BuildToken: strings.Repeat("B", 64), MoveSlideToken: strings.Repeat("C", 64)}
	for _, token := range []string{"D", "E", "F"} {
		p.Edits = append(p.Edits, RoundTripEdit{ShapeToken: strings.Repeat(token, 64), SlideToken: strings.Repeat("0", 64), Before: "Original", After: "Edited"})
	}
	return p
}

func TestWindowsRoundTripOwnedPathsAndJSON(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "fixture with spaces Ω & $(literal)")
	dir, canonicalErr := canonicalPath(dir)
	if canonicalErr != nil {
		t.Fatal(canonicalErr)
	}
	calls := 0
	run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls++
		if name != "powershell.exe" || args[len(args)-2] != "-Config" {
			t.Fatal(name, args)
		}
		raw, err := os.ReadFile(args[len(args)-1])
		if err != nil {
			t.Fatal(err)
		}
		var req roundTripRequest
		if err = json.Unmarshal(raw, &req); err != nil {
			t.Fatal(err)
		}
		if req.Action != "edit" || filepath.Dir(req.Input) != dir || req.SavedAs != filepath.Join(dir, "saved-as.pptx") || req.Edited != filepath.Join(dir, "edited.pptx") {
			t.Fatal(req)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("unbounded helper")
		}
		for _, path := range []string{req.SavedAs, req.Edited} {
			if err := os.WriteFile(path, []byte("simulated saved bytes"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		return json.Marshal(RoundTripExecution{Schema: "pptxgengo.windows-roundtrip-execution.v1", PowerPointVersion: "simulated; not native qualification", OSVersion: "simulated OS", SavedAs: req.SavedAs, Edited: req.Edited, Closed: true})
	}
	out, err := windowsRoundTrip(context.Background(), dir, []byte("synthetic helper input"), testRoundTripPlan(), run)
	if err != nil || calls != 1 || !out.Closed || out.Started == "" || out.Finished == "" || len(out.InputSHA256) != 64 || len(out.SavedAsSHA256) != 64 || len(out.EditedSHA256) != 64 {
		t.Fatal(out, err, calls)
	}
	if _, err = windowsRoundTrip(context.Background(), dir, []byte("new"), testRoundTripPlan(), run); err == nil || calls != 1 {
		t.Fatal("existing fixture reused")
	}
}

func TestWindowsRoundTripFailureCleanupRetainsFixture(t *testing.T) {
	for _, failure := range []string{"helper", "wrong-output", "trailing", "not-closed", "missing-save", "input-changed"} {
		t.Run(failure, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "fixture")
			calls := 0
			run := func(ctx context.Context, name string, args ...string) ([]byte, error) {
				calls++
				raw, err := os.ReadFile(args[len(args)-1])
				if err != nil {
					t.Fatal(err)
				}
				var req roundTripRequest
				if err = json.Unmarshal(raw, &req); err != nil {
					t.Fatal(err)
				}
				if calls == 2 {
					if req.Action != "close" || ctx.Err() != nil {
						t.Fatal("cleanup does not have its own live context")
					}
					if _, ok := ctx.Deadline(); !ok {
						t.Fatal("cleanup unbounded")
					}
					return nil, fmt.Errorf("simulated close failure")
				}
				if failure == "helper" {
					return nil, fmt.Errorf("simulated helper failure")
				}
				out := RoundTripExecution{Schema: "pptxgengo.windows-roundtrip-execution.v1", PowerPointVersion: "fake", OSVersion: "simulated OS", SavedAs: req.SavedAs, Edited: req.Edited, Closed: true}
				if failure == "input-changed" {
					if err := os.WriteFile(req.Input, []byte("changed input"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if failure == "wrong-output" {
					out.Edited = filepath.Join(t.TempDir(), "personal-deck.pptx")
				}
				if failure == "not-closed" {
					out.Closed = false
				}
				data, _ := json.Marshal(out)
				if failure == "trailing" {
					data = append(data, []byte(" {}")...)
				}
				return data, nil
			}
			_, err := windowsRoundTrip(context.Background(), dir, []byte("retained input"), testRoundTripPlan(), run)
			if err == nil || calls != 2 || !strings.Contains(err.Error(), "close unconfirmed") {
				t.Fatal(err, calls)
			}
			data, err := os.ReadFile(filepath.Join(dir, "input.pptx"))
			if err != nil || (failure != "input-changed" && string(data) != "retained input") {
				t.Fatal("fixture removed after failure", err)
			}
			if _, err = os.Stat(filepath.Join(dir, "execution.json")); !os.IsNotExist(err) {
				t.Fatal("success metadata emitted after failure")
			}
		})
	}
}

func TestWindowsRoundTripRejectsInvalidPlanBeforeCreation(t *testing.T) {
	for _, mutation := range []string{"token", "duplicate", "paragraph", "utf8", "no-op", "count"} {
		t.Run(mutation, func(t *testing.T) {
			plan := testRoundTripPlan()
			switch mutation {
			case "token":
				plan.BuildToken = "invalid"
			case "duplicate":
				plan.Edits[1].ShapeToken = plan.Edits[0].ShapeToken
			case "paragraph":
				plan.Edits[0].After = "two\nparagraphs"
			case "utf8":
				plan.Edits[0].After = string([]byte{255})
			case "no-op":
				plan.Edits[0].After = plan.Edits[0].Before
			case "count":
				plan.Edits = plan.Edits[:2]
			}
			dir := filepath.Join(t.TempDir(), "fixture")
			_, err := windowsRoundTrip(context.Background(), dir, []byte("fixture"), plan, func(context.Context, string, ...string) ([]byte, error) {
				t.Fatal("invalid plan executed")
				return nil, nil
			})
			if err == nil {
				t.Fatal("invalid plan accepted")
			}
			if _, err = os.Stat(dir); !os.IsNotExist(err) {
				t.Fatal("invalid plan created fixture")
			}
		})
	}
}
