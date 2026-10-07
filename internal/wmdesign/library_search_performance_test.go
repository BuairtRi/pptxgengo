package wmdesign

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Opt-in diagnostics use separate fresh benchmark processes; model preparation
// in this parent never contributes to the child's peak resident memory.
func TestPinnedLibrarySearchPerformance(t *testing.T) {
	modelDir := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	out := os.Getenv("PPTXGENGO_SEARCH_BENCH_OUT")
	if testing.Short() || modelDir == "" || out == "" {
		t.Skip("explicit pinned offline model and NEW performance output directory required")
	}
	if err := os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	path, source := indexFixtureSubset(t, 4)
	if os.Getenv("PPTXGENGO_EMBED_FULL_LIBRARY") == "1" {
		path, source = indexFixture(t)
	}
	index, err := OpenLibraryIndex(path, LibraryIndexOptions{})
	if err != nil {
		t.Fatal(err)
	}
	snapshotPath := filepath.Join(t.TempDir(), "embeddings.json")
	snapshot, err := index.BuildEmbeddings(t.Context(), modelDir, snapshotPath)
	closeErr := index.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	bin := filepath.Join(t.TempDir(), "search-benchmark")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	buildCtx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-trimpath", "-o", bin, "./scripts/cmd/search-benchmark")
	build.Dir = root
	build.WaitDelay = 2 * time.Second
	if data, err := build.CombinedOutput(); err != nil {
		t.Fatalf("benchmark build: %v: %s", err, data)
	}
	for _, mode := range []string{"keyword", "semantic", "hybrid"} {
		output := filepath.Join(out, mode+".json")
		args := []string{"--index", path, "--mode", mode, "--query", "three parallel findings", "--warm-runs", "3", "--out", output}
		if commit := os.Getenv("PPTXGENGO_SEARCH_BENCH_COMMIT"); commit != "" {
			args = append(args, "--source-commit", commit)
		}
		if mode != "keyword" {
			args = append(args, "--embeddings", snapshotPath, "--model-dir", modelDir)
		}
		ctx, stop := context.WithTimeout(t.Context(), 90*time.Second)
		cmd := exec.CommandContext(ctx, bin, args...)
		cmd.WaitDelay = 2 * time.Second
		data, err := cmd.CombinedOutput()
		stop()
		if err != nil {
			t.Fatalf("%s benchmark: %v: %s", mode, err, data)
		}
		t.Logf("%s", data)
		raw, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var m struct {
			Schema, Mode, OS, Architecture string
			Projection                     string         `json:"projection_sha256"`
			Retrieval                      string         `json:"retrieval_sha256"`
			Counts                         map[string]int `json:"entity_counts"`
			Warm                           []float64      `json:"repeated_find_ms"`
			Peak                           uint64         `json:"process_peak_resident_bytes"`
			Coverage                       int            `json:"vector_coverage"`
		}
		if err = json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		if m.Schema != "pptxgengo.search-performance.v1" || m.Mode != mode || m.OS != runtime.GOOS || m.Architecture != runtime.GOARCH || m.Projection != source.ProjectionSHA256 || m.Retrieval != source.RetrievalSHA256 || len(m.Warm) != 3 || m.Peak == 0 || string(indexJSON(m.Counts)) != string(indexJSON(source.Counts)) {
			t.Fatal("incomplete or incomparable performance evidence", fmt.Sprint(m))
		}
		if mode != "keyword" && m.Coverage != len(snapshot.Rows) {
			t.Fatal("incomplete vector benchmark coverage")
		}
	}
}
