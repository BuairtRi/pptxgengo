// search-benchmark records production LibraryIndex API timings in a fresh
// process. It does not download models, mutate an index or claim content fit.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/buairtri/pptxgengo/internal/localembed"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type options struct {
	index, embeddings, model, output, mode, query, commit string
	warm                                                  int
}
type fileEvidence struct {
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}
type goMemory struct {
	HeapAlloc  uint64 `json:"heap_alloc_bytes"`
	HeapSys    uint64 `json:"heap_sys_bytes"`
	RuntimeSys uint64 `json:"runtime_sys_bytes"`
}
type measurement struct {
	Schema                 string               `json:"schema"`
	SourceCommit           string               `json:"source_commit,omitempty"`
	OS                     string               `json:"os"`
	Architecture           string               `json:"architecture"`
	GoVersion              string               `json:"go_version"`
	CPUCount               int                  `json:"cpu_count"`
	GOMAXPROCS             int                  `json:"gomaxprocs"`
	Created                string               `json:"created"`
	Scope                  string               `json:"scope"`
	FilesystemCache        string               `json:"filesystem_cache"`
	Mode                   string               `json:"mode"`
	Query                  string               `json:"query"`
	Kinds                  []string             `json:"kinds"`
	EntityCounts           map[string]int       `json:"entity_counts"`
	Index                  fileEvidence         `json:"index"`
	Embeddings             *fileEvidence        `json:"embeddings,omitempty"`
	ProjectionSHA256       string               `json:"projection_sha256"`
	RetrievalSHA256        string               `json:"retrieval_sha256"`
	SourceRevision         string               `json:"source_revision"`
	Model                  *localembed.Identity `json:"model,omitempty"`
	ModelArtifactBytes     int64                `json:"model_artifact_bytes,omitempty"`
	OpenMS                 float64              `json:"verified_index_open_ms"`
	FirstFindMS            float64              `json:"first_find_ms"`
	WarmFindMS             []float64            `json:"repeated_find_ms"`
	WarmMedianMS           float64              `json:"repeated_median_ms"`
	WarmP95MS              float64              `json:"repeated_p95_ms"`
	FirstMatches           []string             `json:"first_match_ids"`
	VectorCoverage         int                  `json:"vector_coverage"`
	GoMemoryBefore         goMemory             `json:"go_memory_before"`
	GoMemoryAfterFirst     goMemory             `json:"go_memory_after_first_find"`
	GoMemoryAfter          goMemory             `json:"go_memory_after"`
	FirstPeakResidentBytes uint64               `json:"process_peak_resident_after_first_find_bytes"`
	PeakResidentBytes      uint64               `json:"process_peak_resident_bytes"`
	PeakResidentMethod     string               `json:"process_peak_resident_method"`
	RelevanceAcceptance    string               `json:"relevance_acceptance"`
	PerformanceTargets     string               `json:"performance_targets"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string) error {
	var o options
	flags := flag.NewFlagSet("search-benchmark", flag.ContinueOnError)
	flags.StringVar(&o.index, "index", "", "verified library SQLite index")
	flags.StringVar(&o.embeddings, "embeddings", "", "pinned vector snapshot (semantic/hybrid)")
	flags.StringVar(&o.model, "model-dir", "", "existing pinned offline model (semantic/hybrid)")
	flags.StringVar(&o.output, "out", "", "NEW measurement JSON")
	flags.StringVar(&o.mode, "mode", "hybrid", "keyword, semantic or hybrid")
	flags.StringVar(&o.query, "query", "three parallel findings", "fixed template search query")
	flags.StringVar(&o.commit, "source-commit", "", "optional exact source SHA supplied by CI")
	flags.IntVar(&o.warm, "warm-runs", 3, "repeated same-process finds (1–10)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if err := validateOptions(o); err != nil {
		return err
	}
	if _, err := os.Lstat(o.output); !os.IsNotExist(err) {
		return fmt.Errorf("measurement destination exists or is unavailable: %s", o.output)
	}
	report, err := measure(o)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(o.output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("%s/%s %s: first %.3f ms, repeated median %.3f ms, peak resident %d bytes\n", report.OS, report.Architecture, report.Mode, report.FirstFindMS, report.WarmMedianMS, report.PeakResidentBytes)
	return nil
}

func validateOptions(o options) error {
	if o.index == "" || o.output == "" || !utf8.ValidString(o.query) || strings.TrimSpace(o.query) == "" || len(o.query) > 64<<10 || o.warm < 1 || o.warm > 10 {
		return fmt.Errorf("index, new output, bounded query and 1–10 repeated runs are required")
	}
	if o.mode != "keyword" && o.mode != "semantic" && o.mode != "hybrid" {
		return fmt.Errorf("unsupported measurement mode")
	}
	if o.mode == "keyword" && (o.model != "" || o.embeddings != "") {
		return fmt.Errorf("keyword measurements must omit model options")
	}
	if o.mode != "keyword" && (o.model == "" || o.embeddings == "") {
		return fmt.Errorf("semantic/hybrid measurements require existing offline model and snapshot")
	}
	if o.commit != "" {
		if len(o.commit) != 40 {
			return fmt.Errorf("source commit must be a full SHA")
		}
		for _, c := range o.commit {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
				return fmt.Errorf("source commit must be a lowercase SHA")
			}
		}
	}
	return nil
}

func heapMemory() goMemory {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return goMemory{s.HeapAlloc, s.HeapSys, s.Sys}
}
func milliseconds(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }
func quantile(values []float64, q float64) float64 {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	return v[int(math.Ceil(float64(len(v))*q))-1]
}

func measure(o options) (measurement, error) {
	m := measurement{Schema: "pptxgengo.search-performance.v1", SourceCommit: o.commit, OS: runtime.GOOS, Architecture: runtime.GOARCH, GoVersion: runtime.Version(), CPUCount: runtime.NumCPU(), GOMAXPROCS: runtime.GOMAXPROCS(0), Created: time.Now().UTC().Format(time.RFC3339Nano), Scope: "fresh benchmark process using production LibraryIndex API; index open and first find timed separately; repeated finds reuse the index, not a cached model; timings exclude CLI parsing/process launch, preparation, footprint hashing and output writing; peak resident covers this process through the last timed query including initial hashing", FilesystemCache: "not flushed or controlled; initial footprint hashing and prior fixture/model preparation may warm OS file cache", Mode: o.mode, Query: o.query, Kinds: []string{"template"}, WarmFindMS: []float64{}, FirstMatches: []string{}, RelevanceAcceptance: "not_recorded; retrieval does not measure content fit or native acceptance", PerformanceTargets: "not_agreed; observational measurements only"}
	var err error
	m.Index, err = fileHash(o.index, 512<<20)
	if err != nil {
		return m, err
	}
	if o.mode != "keyword" {
		observed, err := fileHash(o.embeddings, 128<<20)
		if err != nil {
			return m, err
		}
		m.Embeddings = &observed
		identity := localembed.PinnedIdentity()
		m.Model = &identity
		m.ModelArtifactBytes, err = modelFootprint(o.model)
		if err != nil {
			return m, err
		}
	}
	m.GoMemoryBefore = heapMemory()
	start := time.Now()
	index, err := wmdesign.OpenLibraryIndex(o.index, wmdesign.LibraryIndexOptions{})
	m.OpenMS = milliseconds(time.Since(start))
	if err != nil {
		return m, err
	}
	defer index.Close()
	m.EntityCounts = index.Report.Counts
	m.ProjectionSHA256 = index.Report.ProjectionSHA256
	m.RetrievalSHA256 = index.Report.RetrievalSHA256
	m.SourceRevision = index.Report.SourceRevision
	opts := wmdesign.LibraryIndexFindOptions{Retrieval: o.mode, Embeddings: o.embeddings, ModelDir: o.model, Kinds: m.Kinds, Shape: wmdesign.LibrarySearchOptions{Query: o.query, Limit: 10}}
	var first wmdesign.LibraryIndexFindResult
	for i := 0; i <= o.warm; i++ {
		start = time.Now()
		result, err := index.Find(opts)
		elapsed := milliseconds(time.Since(start))
		if err != nil {
			return m, err
		}
		if result.Retrieval.Actual != o.mode {
			return m, fmt.Errorf("measurement refused retrieval fallback: requested %s, got %s", o.mode, result.Retrieval.Actual)
		}
		for _, hit := range result.Matches {
			if hit.Entity.Kind != "template" || hit.FitStatus != "not_measured_for_query" {
				return m, fmt.Errorf("measurement encountered changed filter/fit semantics")
			}
		}
		if i == 0 {
			first = result
			m.FirstFindMS = elapsed
			m.GoMemoryAfterFirst = heapMemory()
			m.FirstPeakResidentBytes, _, err = peakResidentBytes()
			if err != nil || m.FirstPeakResidentBytes == 0 {
				return m, fmt.Errorf("cannot capture first-query peak resident memory: %v", err)
			}
			m.VectorCoverage = result.Retrieval.VectorCoverage
			for _, hit := range result.Matches {
				m.FirstMatches = append(m.FirstMatches, hit.Entity.ID)
			}
		} else {
			if !reflect.DeepEqual(first, result) {
				return m, fmt.Errorf("repeated retrieval changed results; refuse incomparable timings")
			}
			m.WarmFindMS = append(m.WarmFindMS, elapsed)
		}
	}
	m.WarmMedianMS = quantile(m.WarmFindMS, .5)
	m.WarmP95MS = quantile(m.WarmFindMS, .95)
	m.GoMemoryAfter = heapMemory()
	m.PeakResidentBytes, m.PeakResidentMethod, err = peakResidentBytes()
	if err != nil || m.PeakResidentBytes == 0 {
		return m, fmt.Errorf("cannot capture process peak resident memory: %v", err)
	}
	// Recheck exact physical inputs AFTER the timed calls. Never attach a hash
	// from a replaced index/snapshot to measurements from the earlier input.
	observed, err := fileHash(o.index, 512<<20)
	if err != nil || observed != m.Index {
		return m, fmt.Errorf("index changed during measurement: %v", err)
	}
	if o.mode != "keyword" {
		bytes, err := modelFootprint(o.model)
		if err != nil || bytes != m.ModelArtifactBytes {
			return m, fmt.Errorf("model changed during measurement: %v", err)
		}
		observed, err := fileHash(o.embeddings, 128<<20)
		if err != nil || observed != *m.Embeddings {
			return m, fmt.Errorf("snapshot changed during measurement: %v", err)
		}
	}
	return m, nil
}

func modelFootprint(root string) (int64, error) {
	var size int64
	for _, artifact := range localembed.Artifacts() {
		observed, err := fileHash(filepath.Join(root, artifact.File), artifact.Bytes)
		if err != nil || observed.Bytes != artifact.Bytes || observed.SHA256 != artifact.SHA256 {
			return 0, fmt.Errorf("invalid pinned model footprint: %s: %v", artifact.File, err)
		}
		size += observed.Bytes
	}
	return size, nil
}

func fileHash(path string, limit int64) (fileEvidence, error) {
	var result fileEvidence
	before, err := os.Lstat(path)
	if err != nil {
		return result, err
	}
	if !before.Mode().IsRegular() || before.Size() > limit {
		return result, fmt.Errorf("invalid or oversized measurement input")
	}
	f, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return result, err
	}
	if !os.SameFile(before, opened) {
		return result, fmt.Errorf("measurement input identity changed")
	}
	h := sha256.New()
	size, err := io.Copy(h, io.LimitReader(f, limit+1))
	if err != nil {
		return result, err
	}
	after, err := f.Stat()
	if err != nil {
		return result, err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return result, err
	}
	if size != before.Size() || size > limit || after.Size() != size || !os.SameFile(before, current) || !after.ModTime().Equal(before.ModTime()) {
		return result, fmt.Errorf("measurement input changed during hashing")
	}
	return fileEvidence{fmt.Sprintf("%x", h.Sum(nil)), size}, nil
}
