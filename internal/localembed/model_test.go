package localembed

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testTokenizer(t *testing.T) *Tokenizer {
	t.Helper()
	words := make([]string, 104)
	for i := range words {
		words[i] = fmt.Sprintf("[unused%d]", i)
	}
	for word, id := range map[string]int{"[PAD]": 0, "[UNK]": 100, "[CLS]": 101, "[SEP]": 102, "[MASK]": 103} {
		words[id] = word
	}
	words = append(words, "hello", "world", "cafe", "!", "a", "##b", "中")
	tok, err := NewTokenizer([]byte(strings.Join(words, "\n") + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestTokenizerRecipeAndBounds(t *testing.T) {
	tok := testTokenizer(t)
	for text, want := range map[string]string{"Héllo WORLD!": "[101 104 105 107 102]", "Café": "[101 106 102]", "ab": "[101 108 109 102]", "abc": "[101 100 102]", "中": "[101 110 102]", "[MASK]hello[PAD]": "[101 103 104 0 102]", "": "[101 102]"} {
		got, err := tok.Encode(text)
		if err != nil || fmt.Sprint(got) != want {
			t.Fatalf("%q: %v %v; want %s", text, got, err, want)
		}
	}
	ids, err := tok.Encode(strings.Repeat("hello ", 300))
	if err != nil || len(ids) != 256 || ids[255] != 102 {
		t.Fatal(ids, err)
	}
	for _, text := range []string{string([]byte{0xff}), strings.Repeat("x", 65537)} {
		if _, err := tok.Encode(text); err == nil {
			t.Fatal("invalid/unbounded input accepted")
		}
	}
	if _, err := NewTokenizer([]byte("[PAD]\n[PAD]")); err == nil {
		t.Fatal("duplicate vocabulary accepted")
	}
}

func TestPoolMasksNormalizesAndRejectsInvalidOutput(t *testing.T) {
	row := make([]float32, Dimensions)
	row[0] = 3
	row[1] = 4
	ignored := make([]float32, Dimensions)
	ignored[2] = 100
	vector, err := Pool([][]float32{row, ignored}, []int64{1, 0})
	if err != nil || math.Abs(float64(vector[0])-.6) > 1e-6 || math.Abs(float64(vector[1])-.8) > 1e-6 || vector[2] != 0 {
		t.Fatal(vector, err)
	}
	bad := append([]float32(nil), row...)
	bad[0] = float32(math.NaN())
	for _, test := range []struct {
		rows [][]float32
		mask []int64
	}{{nil, nil}, {[][]float32{row}, nil}, {[][]float32{{1}}, []int64{1}}, {[][]float32{row}, []int64{2}}, {[][]float32{row}, []int64{0}}, {[][]float32{make([]float32, Dimensions)}, []int64{1}}, {[][]float32{bad}, []int64{1}}} {
		if _, err := Pool(test.rows, test.mask); err == nil {
			t.Fatal("invalid output accepted")
		}
	}
}

func TestLoadRejectsMissingAndChangedArtifacts(t *testing.T) {
	if _, err := Load(""); err == nil {
		t.Fatal("missing directory accepted")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "model.onnx"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("changed artifact accepted")
	}
}

// Opt-in pinned weights keep normal unit tests offline and lightweight. CI may
// supply this directory only after explicit, hash-checked maintenance download.
func TestPinnedModelGolden(t *testing.T) {
	dir := os.Getenv("PPTXGENGO_EMBED_MODEL_DIR")
	if dir == "" {
		t.Skip("set PPTXGENGO_EMBED_MODEL_DIR to the pinned offline package")
	}
	data, err := os.ReadFile("testdata/minilm-golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Model    string
		Revision string
		Cases    []struct {
			Text   string
			IDs    []int64
			Vector []float32
		}
	}
	if err = json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	if golden.Model != ModelID || golden.Revision != ModelRevision {
		t.Fatal("fixture model identity changed")
	}
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for i, c := range golden.Cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ids, err := m.tokenizer.Encode(c.Text)
			if err != nil || !reflect.DeepEqual(ids, c.IDs) {
				t.Fatalf("tokenizer %q: got %v; want %v: %v", c.Text, ids, c.IDs, err)
			}
			if len(c.Vector) == 0 {
				return
			}
			vector, err := m.Embed(context.Background(), c.Text)
			if err != nil {
				t.Fatal(err)
			}
			if len(vector) != len(c.Vector) {
				t.Fatal("dimensions changed")
			}
			maxError := 0.0
			dot, norm := 0.0, 0.0
			for j, x := range vector {
				maxError = math.Max(maxError, math.Abs(float64(x)-float64(c.Vector[j])))
				dot += float64(x) * float64(c.Vector[j])
				norm += float64(x) * float64(x)
			}
			if maxError > 2e-5 || dot < .99999 || math.Abs(norm-1) > 1e-5 {
				t.Fatalf("oracle disagreement: max abs %g, dot %g, squared norm %g", maxError, dot, norm)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = m.Embed(ctx, "cancelled"); err != context.Canceled {
		t.Fatal("cancellation ignored", err)
	}
	m.Close()
	if _, err = m.Embed(context.Background(), "closed"); err == nil {
		t.Fatal("closed model accepted")
	}
}
