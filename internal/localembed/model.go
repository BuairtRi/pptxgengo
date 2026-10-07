package localembed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/buairtri/pptxgengo/internal/modelpackage"
	"github.com/gomlx/compute"
	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx"
	"github.com/gomlx/onnx-gomlx/onnx/parser"
)

const Dimensions = modelpackage.Dimensions
const ModelID = modelpackage.ModelID
const ModelRevision = modelpackage.ModelRevision
const Runtime = modelpackage.Runtime
const Pooling = modelpackage.Pooling

type Artifact = modelpackage.Artifact
type Identity = modelpackage.Identity

func Artifacts() []Artifact    { return modelpackage.Artifacts() }
func PinnedIdentity() Identity { return modelpackage.PinnedIdentity() }
func digest(b []byte) string   { return fmt.Sprintf("%x", sha256.Sum256(b)) }

type Model struct {
	mu        sync.Mutex
	tokenizer *Tokenizer
	backend   compute.Backend
	store     *model.Store
	network   onnx.Model
	exec      *model.Exec
	closed    bool
}

func pinnedBytes(dir string, a Artifact) ([]byte, error) {
	p := filepath.Join(dir, a.File)
	info, e := os.Lstat(p)
	if e != nil {
		return nil, fmt.Errorf("embedding.model_missing: %s: %w", a.File, e)
	}
	if !info.Mode().IsRegular() || info.Size() != a.Bytes {
		return nil, fmt.Errorf("embedding.model_incompatible: %s size/type", a.File)
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, e
	}
	if int64(len(b)) != a.Bytes || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
		return nil, fmt.Errorf("embedding.model_hash_mismatch: %s", a.File)
	}
	return b, nil
}

// Load never uses the network, native libraries, Python, default backend plugins
// or unpinned models. Parse the exact bytes just hashed, avoiding a second read.
func Load(dir string) (m *Model, err error) {
	m = &Model{}
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("embedding.model_load: %v", p)
		}
		if err != nil {
			m.Close()
			m = nil
		}
	}()
	if dir == "" {
		return m, fmt.Errorf("embedding.model_missing: supply an offline --model-dir")
	}
	var weights, vocab []byte
	for _, a := range Artifacts() {
		b, e := pinnedBytes(dir, a)
		if e != nil {
			return m, e
		}
		if a.File == "model.onnx" {
			weights = b
		}
		if a.File == "vocab.txt" {
			vocab = b
		}
	}
	if m.tokenizer, err = NewTokenizer(vocab); err != nil {
		return m, err
	}
	if m.network, err = parser.Parse(weights); err != nil {
		return m, err
	}
	m.store = model.NewStore()
	if err = m.network.VariablesToScope(m.store.RootScope()); err != nil {
		return m, err
	}
	// Default portable parallelism is included in the runtime identity. Explicit Go backend
	// construction avoids CGO and any implicit accelerator/plugin download.
	if m.backend, err = gobackend.New(""); err != nil {
		return m, err
	}
	m.exec = model.MustNewExec(m.backend, m.store, func(s *model.Scope, ids, mask, types *graph.Node) *graph.Node {
		return m.network.CallGraph(s, ids.Graph(), map[string]*graph.Node{"input_ids": ids, "attention_mask": mask, "token_type_ids": types})[0]
	})
	// All input axes are dynamic: varying sentence lengths reuse one graph.
	m.exec.WithDynamicAxes([]string{"batch", "sequence"}, []string{"batch", "sequence"}, []string{"batch", "sequence"})
	return m, nil
}

func (m *Model) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.closed = true
	if m.exec != nil {
		m.exec.Finalize()
	}
	if m.store != nil {
		m.store.Finalize()
	}
	if m.network != nil {
		m.network.Close()
	}
	if m.backend != nil {
		m.backend.Finalize()
	}
}

func (m *Model) Embed(ctx context.Context, text string) (vector []float32, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("embedding.inference_failed: %v", p)
			vector = nil
		}
	}()
	if m.closed {
		return nil, fmt.Errorf("embedding.model_closed")
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	ids, e := m.tokenizer.Encode(text)
	if e != nil {
		return nil, e
	}
	mask := make([]int64, len(ids))
	types := make([]int64, len(ids))
	for i := range mask {
		mask[i] = 1
	}
	out := m.exec.MustCall1([][]int64{ids}, [][]int64{mask}, [][]int64{types})
	defer out.MustFinalizeAll()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	rows, ok := out.Value().([][][]float32)
	if !ok || len(rows) != 1 || len(rows[0]) != len(ids) {
		return nil, fmt.Errorf("embedding.output_shape_changed")
	}
	return Pool(rows[0], mask)
}

func Pool(rows [][]float32, mask []int64) ([]float32, error) {
	if len(rows) == 0 || len(rows) != len(mask) {
		return nil, fmt.Errorf("embedding.invalid_mask")
	}
	sums := make([]float64, Dimensions)
	count := 0
	for i, row := range rows {
		if len(row) != Dimensions || (mask[i] != 0 && mask[i] != 1) {
			return nil, fmt.Errorf("embedding.invalid_dimensions_or_mask")
		}
		for _, x := range row {
			if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
				return nil, fmt.Errorf("embedding.nonfinite_output")
			}
		}
		if mask[i] == 0 {
			continue
		}
		count++
		for j, x := range row {
			sums[j] += float64(x)
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("embedding.empty_mask")
	}
	norm := 0.0
	for i := range sums {
		sums[i] /= float64(count)
		norm += sums[i] * sums[i]
	}
	norm = math.Sqrt(norm)
	if norm == 0 || math.IsNaN(norm) || math.IsInf(norm, 0) {
		return nil, fmt.Errorf("embedding.zero_or_invalid_norm")
	}
	v := make([]float32, Dimensions)
	for i := range v {
		v[i] = float32(sums[i] / norm)
	}
	return v, nil
}
