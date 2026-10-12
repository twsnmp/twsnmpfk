package tensai

import (
	"encoding/json"
	"fmt"

	"github.com/mattn/tensai/encoding/gguf"
	"github.com/mattn/tensai/tokenizer"
)

// buildGGUFTokenizer rebuilds tokenizer from GGUF metadata
func buildGGUFTokenizer(g *gguf.File) (*tokenizer.Tokenizer, error) {
	toksAny, ok := g.KV("tokenizer.ggml.tokens")
	if !ok {
		return nil, fmt.Errorf("gguf has no embedded tokenizer")
	}

	var tokens []string
	switch v := toksAny.(type) {
	case []string:
		tokens = v
	case []any:
		tokens = make([]string, len(v))
		for i, t := range v {
			if s, okStr := t.(string); okStr {
				tokens[i] = s
			}
		}
	default:
		return nil, fmt.Errorf("invalid token list format: expected []string or []any, got %T", toksAny)
	}

	mergesAny, _ := g.KV("tokenizer.ggml.merges")
	var merges []string
	switch v := mergesAny.(type) {
	case []string:
		merges = v
	case []any:
		merges = make([]string, len(v))
		for i, m := range v {
			if s, okStr := m.(string); okStr {
				merges[i] = s
			}
		}
	}

	typesAny, _ := g.KV("tokenizer.ggml.token_type")
	var types []int32
	switch v := typesAny.(type) {
	case []int32:
		types = v
	case []uint32:
		types = make([]int32, len(v))
		for i, t := range v {
			types[i] = int32(t)
		}
	case []any:
		types = make([]int32, len(v))
		for i, t := range v {
			if n, okN := t.(int32); okN {
				types[i] = n
			} else if n, okN := t.(int64); okN {
				types[i] = int32(n)
			}
		}
	}

	var scores []float32
	if scoresAny, okScores := g.KV("tokenizer.ggml.scores"); okScores {
		switch v := scoresAny.(type) {
		case []float32:
			scores = v
		case []float64:
			scores = make([]float32, len(v))
			for i, s := range v {
				scores[i] = float32(s)
			}
		case []any:
			scores = make([]float32, len(v))
			for i, s := range v {
				if f, okF := s.(float32); okF {
					scores[i] = f
				} else if f, okF := s.(float64); okF {
					scores[i] = float32(f)
				}
			}
		}
	}

	modelType, _ := g.String("tokenizer.ggml.model")

	// If no merges exist (SentencePiece SPM models like Gemma / Llama-SPM), try NewSPM
	if len(merges) == 0 && len(scores) == len(tokens) && len(types) == len(tokens) {
		addSpace := modelType != "gemma" && modelType != "gemma2"
		if spmTok, err := tokenizer.NewSPM(tokens, scores, types, addSpace); err == nil {
			return spmTok, nil
		}
	}

	// For models like Gemma 2/3/4 that use SentencePiece with BPE merges
	if (modelType == "gemma" || modelType == "gemma2" || modelType == "gemma4") && len(merges) > 0 && len(types) == len(tokens) {
		if spmTok, err := tokenizer.NewSPMBPE(tokens, merges, types, false); err == nil {
			return spmTok, nil
		}
	}

	pre, _ := g.String("tokenizer.ggml.pre")
	arch, _ := g.String("general.architecture")
	if pre == "" {
		switch arch {
		case "qwen2", "qwen":
			pre = "qwen2"
		case "llama":
			pre = "llama3"
		}
	}

	var preJSON string
	switch pre {
	case "smollm":
		preJSON = `{"type":"Sequence","pretokenizers":[{"type":"Digits","individual_digits":true},{"type":"ByteLevel","use_regex":true}]}`
	case "qwen2", "llama-bpe", "llama3", "smaug-bpe", "deepseek-r1-qwen":
		preJSON = `{"type":"Split","pattern":{"Regex":"(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\\r\\n\\p{L}\\p{N}]?\\p{L}+|\\p{N}{1,3}| ?[^\\s\\p{L}\\p{N}]+[\\r\\n]*|\\s*[\\r\\n]+|\\s+(?!\\S)|\\s+"}}`
	case "gpt-4o":
		preJSON = `{"type":"Split","pattern":{"Regex":"[^\\r\\n\\p{L}\\p{N}]?((?=[\\p{L}])([^a-z]))*((?=[\\p{L}])([^A-Z]))+(?:'[sS]|'[tT]|'[rR][eE]|'[vV][eE]|'[mM]|'[lL][lL]|'[dD])?|[^\\r\\n\\p{L}\\p{N}]?((?=[\\p{L}])([^a-z]))+((?=[\\p{L}])([^A-Z]))*(?:'[sS]|'[tT]|'[rR][eE]|'[vV][eE]|'[mM]|'[lL][lL]|'[dD])?|\\p{N}{1,3}| ?[^\\s\\p{L}\\p{N}]+[\\r\\n/]*|\\s*[\\r\\n]+|\\s+(?!\\S)|\\s+"}}`
	case "gpt-2", "olmo", "":
		preJSON = `{"type":"ByteLevel","use_regex":true}`
	default:
		preJSON = `{"type":"ByteLevel","use_regex":true}`
	}

	vocab := make(map[string]int, len(tokens))
	for id, t := range tokens {
		vocab[t] = id
	}

	var specials []tokenizer.AddedToken
	for id, tp := range types {
		if (tp == 3 || tp == 4) && id < len(tokens) {
			specials = append(specials, tokenizer.AddedToken{
				ID:      id,
				Content: tokens[id],
			})
		}
	}

	// Try fast constructor NewBPE first
	tok, err := tokenizer.NewBPE(vocab, merges, specials, json.RawMessage(preJSON))
	if err == nil {
		return tok, nil
	}

	// Fallback to ByteLevel if preJSON was rejected
	tok, err = tokenizer.NewBPE(vocab, merges, specials, json.RawMessage(`{"type":"ByteLevel","use_regex":true}`))
	if err == nil {
		return tok, nil
	}

	return nil, err
}
