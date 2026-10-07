package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestSearchBenchmarkOptionsAndQuantiles(t *testing.T) {
	base := options{index: "index", output: "new.json", mode: "keyword", query: "cards/3", warm: 3}
	if err := validateOptions(base); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"warm", "query", "utf8", "mode", "model", "commit"} {
		o := base
		switch bad {
		case "warm":
			o.warm = 11
		case "query":
			o.query = " "
		case "utf8":
			o.query = string([]byte{255})
		case "mode":
			o.mode = "metadata"
		case "model":
			o.model = "unneeded"
		case "commit":
			o.commit = "not-a-sha"
		}
		if validateOptions(o) == nil {
			t.Fatal("invalid options accepted", bad)
		}
	}
	for _, mode := range []string{"semantic", "hybrid"} {
		o := base
		o.mode = mode
		if validateOptions(o) == nil {
			t.Fatal("missing offline resources accepted")
		}
		o.model = "pinned"
		o.embeddings = "snapshot"
		if err := validateOptions(o); err != nil {
			t.Fatal(err)
		}
	}
	v := []float64{4, 1, 3, 2, 5}
	if quantile(v, .5) != 3 || quantile(v, .95) != 5 || v[0] != 4 {
		t.Fatal("quantiles changed observations")
	}
}

func TestSearchBenchmarkPeakMemoryUnits(t *testing.T) {
	// Touch actual pages so a mistaken Darwin/Linux unit conversion cannot pass
	// merely because a process has some resident memory.
	pages := make([]byte, 32<<20)
	for i := 0; i < len(pages); i += 4096 {
		pages[i] = 1
	}
	peak, method, err := peakResidentBytes()
	runtime.KeepAlive(pages)
	if err != nil || peak < uint64(len(pages)) || method == "" {
		t.Fatal(peak, method, err)
	}
}

func TestSearchBenchmarkKeywordProofAndNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	bundle, err := filepath.Abs("../../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(dir, "catalog.sqlite")
	report, err := wmdesign.BuildLibraryIndex(index, wmdesign.LibraryIndexOptions{Bundle: bundle})
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "measurement with spaces.json")
	args := []string{"--index", index, "--mode", "keyword", "--query", "cards/3", "--warm-runs", "1", "--out", out}
	if err = run(args); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var m measurement
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Schema != "pptxgengo.search-performance.v1" || m.Mode != "keyword" || m.Model != nil || m.Embeddings != nil || m.ProjectionSHA256 != report.ProjectionSHA256 || m.PeakResidentBytes == 0 || len(m.WarmFindMS) != 1 || len(m.FirstMatches) == 0 || m.FirstMatches[0] != "wmds/template/cards/3" || m.RelevanceAcceptance == "" || m.PerformanceTargets == "" {
		t.Fatal(m)
	}
	if err = run(args); err == nil {
		t.Fatal("measurement overwritten")
	}
	after, err := os.ReadFile(out)
	if err != nil || string(after) != string(raw) {
		t.Fatal("existing measurement changed", err)
	}
	args[len(args)-1] = filepath.Join(dir, "fallback.json")
	args = append(args, "--mode", "hybrid", "--model-dir", filepath.Join(dir, "missing-model"), "--embeddings", filepath.Join(dir, "missing-snapshot"))
	if err = run(args); err == nil {
		t.Fatal("hybrid fallback recorded as hybrid benchmark", err)
	}
	if _, err = os.Stat(filepath.Join(dir, "fallback.json")); !os.IsNotExist(err) {
		t.Fatal("failed measurement published")
	}
}

func TestSearchBenchmarkFileBoundsAndIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data")
	if err := os.WriteFile(path, []byte("bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := fileHash(path, 4); err == nil {
		t.Fatal("oversized file accepted")
	}
	value, err := fileHash(path, 5)
	if err != nil || value.Bytes != 5 || len(value.SHA256) != 64 {
		t.Fatal(value, err)
	}
	if _, err = fileHash(dir, 1024); err == nil {
		t.Fatal("directory accepted")
	}
}
