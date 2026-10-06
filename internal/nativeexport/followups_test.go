package nativeexport

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRenderFailureRecordPreflight(t *testing.T) {
	for _, item := range []struct {
		name, platform, source, slides string
		timeout                        time.Duration
	}{
		{"missing-source", "darwin", "", "", time.Second},
		{"unreadable-source", "darwin", "missing.pptx", "", time.Second},
		{"invalid-selection", "darwin", "source.pptx", "900", time.Second},
		{"invalid-timeout", "darwin", "source.pptx", "", 0},
		{"unsupported-platform", "linux", "source.pptx", "", time.Second},
	} {
		t.Run(item.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "source.pptx"), fixture(t), 0600); err != nil {
				t.Fatal(err)
			}
			source := ""
			if item.source != "" {
				source = filepath.Join(root, item.source)
			}
			out := filepath.Join(root, "out")
			_, err := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: item.timeout, Slides: item.slides}, nil, item.platform)
			if err == nil {
				t.Fatal("preflight succeeded")
			}
			recorded, e := os.ReadFile(filepath.Join(out, "render-error.txt"))
			if e != nil || !strings.Contains(string(recorded), err.Error()) {
				t.Fatal(string(recorded), e, err)
			}
			if _, e = os.Stat(filepath.Join(out, "render-manifest.json")); !os.IsNotExist(e) {
				t.Fatal("failure published receipt", e)
			}
		})
	}
}

func TestFailureRecordPreservesSuccessfulOutputAndRejectsSymlink(t *testing.T) {
	out := t.TempDir()
	path := filepath.Join(out, "render-manifest.json")
	if err := os.WriteFile(path, []byte("prior successful receipt"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeRenderFailure(renderFailure{Out: out, Message: "new failure"}); err == nil {
		t.Fatal("wrote into prior success")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(out, "render-error.txt")); err != nil {
		t.Fatal(err)
	}
	if err := writeRenderFailure(renderFailure{Out: out, Message: "overwrite"}); err == nil {
		t.Fatal("followed error-file symlink")
	}
	if data, _ := os.ReadFile(outside); string(data) != "original" {
		t.Fatal("changed outside file")
	}
}

func TestTimeoutRevokesOnlyExactTaskSignedReceipt(t *testing.T) {
	out := t.TempDir()
	taskID, err := newTaskID()
	if err != nil {
		t.Fatal(err)
	}
	receipt := Receipt{TaskID: taskID, Renderer: "Microsoft PowerPoint (local native PDF)"}
	if err = signReceipt(&receipt); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(receipt)
	path := filepath.Join(out, "render-manifest.json")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	otherID, _ := newTaskID()
	if err = writeRenderFailure(renderFailure{Out: out, TaskID: otherID, Message: "deadline"}); err == nil {
		t.Fatal("revoked another task receipt")
	}
	if err = writeRenderFailure(renderFailure{Out: out, TaskID: taskID, Message: "deadline"}); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("deadline left success receipt", err)
	}
	if raw, err = os.ReadFile(filepath.Join(out, "render-error.txt")); err != nil || string(raw) != "deadline\n" {
		t.Fatal(string(raw), err)
	}
}

func TestRenderGrantDialogAppearingDuringOpenFailsFast(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.pptx")
	if err := os.WriteFile(source, fixture(t), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(ctx context.Context, _ string, args ...string) ([]byte, error) {
		if args[len(args)-1] == "close" {
			return nil, nil
		}
		return monitoredCommand(ctx, func(child context.Context) ([]byte, error) { <-child.Done(); return nil, child.Err() }, func(child context.Context, observations chan<- dialogObservation) {
			select {
			case <-time.After(20 * time.Millisecond):
			case <-child.Done():
				return
			}
			select {
			case observations <- dialogObservation{Status: "blocked", Dialog: "Grant File Access", Detail: "Select folder"}:
			case <-child.Done():
			}
		})
	}
	started := time.Now()
	out := filepath.Join(root, "out")
	_, err := render(context.Background(), Options{PPTX: source, Out: out, PDF: true, Timeout: 30 * time.Second, StagingRoot: filepath.Join(root, "stage")}, run, "darwin")
	if err == nil || !strings.Contains(err.Error(), "file_access_denied") || !strings.Contains(err.Error(), "Grant File Access") {
		t.Fatal(err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("grant detection waited for export timeout")
	}
	if data, e := os.ReadFile(filepath.Join(out, "render-error.txt")); e != nil || !strings.Contains(string(data), "Grant File Access") {
		t.Fatal(string(data), e)
	}
}

func TestUnknownDialogVisibilityRemainsUnknown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := monitoredCommand(ctx, func(child context.Context) ([]byte, error) { <-child.Done(); return nil, child.Err() }, func(child context.Context, observations chan<- dialogObservation) {
		select {
		case observations <- dialogObservation{Status: "unknown", Detail: "no visibility"}:
		case <-child.Done():
		}
		<-child.Done()
	})
	if !errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "file_access_denied") {
		t.Fatal(err)
	}
}

func TestDoctorOperationalFileAccessProbe(t *testing.T) {
	for _, name := range []string{"pass", "grant", "unknown", "no-pdf", "invalid-pdf"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			calls := 0
			run := func(_ context.Context, _ string, args ...string) ([]byte, error) {
				calls++
				if _, e := os.Stat(args[1]); e != nil {
					t.Fatal("probe did not create exact task copy", e)
				}
				if args[5] == "close" {
					return nil, nil
				}
				if args[5] != "export" || len(args) != 7 || filepath.Dir(args[1]) != root || filepath.Dir(args[2]) != root {
					t.Fatal("doctor did not use the actual stable-folder export contract", args)
				}
				if name == "grant" {
					return nil, errors.New("file_access_denied: Grant File Access")
				}
				if name == "unknown" {
					return nil, context.DeadlineExceeded
				}
				if name == "no-pdf" {
					return nil, nil
				}
				content := []byte("%PDF-1.7\ndiagnostic")
				if name == "invalid-pdf" {
					content = []byte("not a PDF")
				}
				return nil, os.WriteFile(args[2], content, 0600)
			}
			check := doctorFileAccess(context.Background(), root, "", run)
			want := "pass"
			if name == "grant" {
				want = "fail"
			}
			if name == "unknown" || name == "no-pdf" || name == "invalid-pdf" {
				want = "unknown"
			}
			if check.Status != want || (name != "pass" && !strings.Contains(check.Fix, "Select/Grant")) {
				t.Fatal(check)
			}
			if calls != 1 && calls != 2 {
				t.Fatal(calls)
			}
			entries, e := os.ReadDir(root)
			if e != nil || len(entries) != 0 {
				t.Fatal("confirmed probe cleanup retained work", entries, e)
			}
		})
	}
}

func TestSignedNativeReceiptRejectsHandmadeAndTamperedEvidence(t *testing.T) {
	receipt := Receipt{Renderer: "Microsoft PowerPoint (local native PDF); macOS PDFKit PNG", Source: Artifact{SHA256: "source"}, Pages: 1}
	unsigned, _ := json.Marshal(receipt)
	if _, err := VerifyReceipt(unsigned); err == nil {
		t.Fatal("handmade unsigned renderer accepted")
	}
	if err := signReceipt(&receipt); err != nil {
		t.Fatal(err)
	}
	trusted, _ := json.Marshal(receipt)
	if _, err := VerifyReceipt(trusted); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Receipt){
		func(r *Receipt) { r.Source.SHA256 = "other" },
		func(r *Receipt) { r.Pages++ },
		func(r *Receipt) { r.PageMappings = []PageMapping{{Page: 1, SourceSlide: 3}} },
		func(r *Receipt) { r.PDF = &Artifact{Path: "deck.pdf", SHA256: "other"} },
		func(r *Receipt) { r.Provenance.KeyID = "untrusted" },
		func(r *Receipt) { r.Provenance.IssuedAt = "2026-01-01T00:00:00Z" },
	} {
		var modified Receipt
		if err := json.Unmarshal(trusted, &modified); err != nil {
			t.Fatal(err)
		}
		change(&modified)
		raw, _ := json.Marshal(modified)
		if _, err := VerifyReceipt(raw); err == nil {
			t.Fatal("tampered signed receipt accepted", string(raw))
		}
	}
	_, other, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Provenance.KeyID = hash(other.Public().(ed25519.PublicKey))
	payload, _ := receiptSigningBytes(receipt)
	receipt.Provenance.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(other, payload))
	raw, _ := json.Marshal(receipt)
	if _, err := VerifyReceipt(raw); err == nil {
		t.Fatal("self-nominated issuer accepted")
	}
	if _, err := VerifyReceipt(append(trusted[:len(trusted)-1], []byte(`,"unknown":true}`)...)); err == nil {
		t.Fatal("unknown manifest field accepted")
	}
}

func TestNativeIssuerConcurrentCreation(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	var wg sync.WaitGroup
	keys := make(chan string, 12)
	errors := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := localIssuerKey(true)
			if err != nil {
				errors <- err
				return
			}
			keys <- hash(key.Public().(ed25519.PublicKey))
		}()
	}
	wg.Wait()
	close(keys)
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	first := ""
	for key := range keys {
		if first != "" && key != first {
			t.Fatal("concurrent issuers used different keys")
		}
		first = key
	}
	if first == "" {
		t.Fatal("no issuer created")
	}
	path, err := issuerKeyPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(info, err)
	}
	if err := validateTrustPermissions(path, info); err != nil {
		t.Fatal(err)
	}
}
