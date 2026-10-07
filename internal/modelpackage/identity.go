// Package modelpackage defines the pinned offline model without loading inference dependencies.
package modelpackage

const MaxTokens = 256
const TokenizerVersion = "pptxgengo.bert-uncased-wordpiece.v1"

const Dimensions = 384
const ModelID = "sentence-transformers/all-MiniLM-L6-v2"
const ModelRevision = "1110a243fdf4706b3f48f1d95db1a4f5529b4d41"
const Runtime = "onnx-gomlx/v0.5.13;gomlx/v0.28.16;compute/go/v0.1.14;parallelism=default"
const Pooling = "attention-mask-mean_then_l2.v1"

type Artifact struct {
	File   string `json:"file"`
	Source string `json:"source"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

func Artifacts() []Artifact {
	return []Artifact{
		{"model.onnx", "onnx/model.onnx", 90405214, "6fd5d72fe4589f189f8ebc006442dbb529bb7ce38f8082112682524616046452"},
		{"vocab.txt", "vocab.txt", 231508, "07eced375cec144d27c900241f3e339478dec958f92fddbc551f295c992038a3"},
		{"tokenizer.json", "tokenizer.json", 466247, "be50c3628f2bf5bb5e3a7f17b1f74611b2561a3a27eeab05e5aa30f411572037"},
	}
}

type Identity struct {
	Model      string     `json:"model"`
	Revision   string     `json:"revision"`
	Artifacts  []Artifact `json:"artifacts"`
	Dimensions int        `json:"dimensions"`
	Tokenizer  string     `json:"tokenizer"`
	MaxTokens  int        `json:"max_tokens"`
	Pooling    string     `json:"pooling"`
	Runtime    string     `json:"runtime"`
}

func PinnedIdentity() Identity {
	return Identity{ModelID, ModelRevision, Artifacts(), Dimensions, TokenizerVersion, MaxTokens, Pooling, Runtime}
}
