// Package localembed provides pinned, offline MiniLM sentence embeddings.
package localembed

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const MaxTokens = 256
const TokenizerVersion = "pptxgengo.bert-uncased-wordpiece.v1"

type Tokenizer struct{ vocab map[string]int64 }

func NewTokenizer(vocab []byte) (*Tokenizer, error) {
	words := strings.Split(strings.TrimSuffix(string(vocab), "\n"), "\n")
	t := &Tokenizer{vocab: make(map[string]int64, len(words))}
	for i, word := range words {
		word = strings.TrimSuffix(word, "\r")
		if word == "" {
			return nil, fmt.Errorf("embedding.empty_vocabulary_entry")
		}
		if _, exists := t.vocab[word]; exists {
			return nil, fmt.Errorf("embedding.duplicate_vocabulary_entry")
		}
		t.vocab[word] = int64(i)
	}
	for word, id := range map[string]int64{"[PAD]": 0, "[UNK]": 100, "[CLS]": 101, "[SEP]": 102, "[MASK]": 103} {
		if actual, ok := t.vocab[word]; !ok || actual != id {
			return nil, fmt.Errorf("embedding.incompatible_vocabulary: %s", word)
		}
	}
	return t, nil
}

func chinese(r rune) bool {
	for _, pair := range [][2]rune{{0x4e00, 0x9fff}, {0x3400, 0x4dbf}, {0x20000, 0x2a6df}, {0x2a700, 0x2b73f}, {0x2b740, 0x2b81f}, {0x2b820, 0x2ceaf}, {0xf900, 0xfaff}, {0x2f800, 0x2fa1f}} {
		if r >= pair[0] && r <= pair[1] {
			return true
		}
	}
	return false
}

func bertWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || unicode.Is(unicode.Zs, r)
}
func bertPunctuation(r rune) bool {
	return (r >= 33 && r <= 47) || (r >= 58 && r <= 64) || (r >= 91 && r <= 96) || (r >= 123 && r <= 126) || unicode.IsPunct(r)
}

func basicTokens(text string) []string {
	var clean strings.Builder
	for _, r := range text {
		if r == 0 || r == utf8.RuneError || (!bertWhitespace(r) && (unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r))) {
			continue
		}
		if bertWhitespace(r) {
			clean.WriteByte(' ')
			continue
		}
		if chinese(r) {
			clean.WriteByte(' ')
			clean.WriteRune(r)
			clean.WriteByte(' ')
		} else {
			clean.WriteRune(r)
		}
	}
	text = norm.NFD.String(strings.ToLower(clean.String()))
	var word strings.Builder
	parts := []string{}
	flush := func() {
		if word.Len() > 0 {
			parts = append(parts, word.String())
			word.Reset()
		}
	}
	for _, r := range text {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsSpace(r) {
			flush()
		} else if bertPunctuation(r) {
			flush()
			parts = append(parts, string(r))
		} else {
			word.WriteRune(r)
		}
	}
	flush()
	return parts
}

func (t *Tokenizer) wordPieces(word string) []int64 {
	runes := []rune(word)
	if len(runes) > 100 {
		return []int64{100}
	}
	ids := []int64{}
	for start := 0; start < len(runes); {
		end := len(runes)
		var id int64
		found := false
		for ; end > start; end-- {
			piece := string(runes[start:end])
			if start > 0 {
				piece = "##" + piece
			}
			if id, found = t.vocab[piece]; found {
				break
			}
		}
		if !found {
			return []int64{100}
		}
		ids = append(ids, id)
		start = end
	}
	return ids
}

// Encode preserves the pinned tokenizer's added special tokens, applies the
// uncased BERT normalization and longest-match WordPiece recipe, and truncates
// only sentence content. CLS and SEP always remain present.
func (t *Tokenizer) Encode(text string) ([]int64, error) {
	if len(text) > 64*1024 || !utf8.ValidString(text) {
		return nil, fmt.Errorf("embedding.invalid_text: require UTF-8 and at most 64 KiB")
	}
	ids := []int64{101}
	appendOrdinary := func(s string) {
		for _, word := range basicTokens(s) {
			for _, id := range t.wordPieces(word) {
				if len(ids) == MaxTokens-1 {
					return
				}
				ids = append(ids, id)
			}
		}
	}
	// Special tokens are recognized before normalization, as in tokenizer.json.
	for len(text) > 0 && len(ids) < MaxTokens-1 {
		position := -1
		special := ""
		for _, candidate := range []string{"[PAD]", "[UNK]", "[CLS]", "[SEP]", "[MASK]"} {
			if p := strings.Index(text, candidate); p >= 0 && (position < 0 || p < position) {
				position = p
				special = candidate
			}
		}
		if position < 0 {
			appendOrdinary(text)
			break
		}
		appendOrdinary(text[:position])
		if len(ids) < MaxTokens-1 {
			ids = append(ids, t.vocab[special])
		}
		text = text[position+len(special):]
	}
	return append(ids, 102), nil
}
