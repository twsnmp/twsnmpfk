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
	mergesAny, _ := g.KV("tokenizer.ggml.merges")
	typesAny, _ := g.KV("tokenizer.ggml.token_type")

	pre, _ := g.String("tokenizer.ggml.pre")
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

	tokens, ok := toksAny.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid token list format")
	}

	vocab := make(map[string]int, len(tokens))
	for id, t := range tokens {
		if s, okStr := t.(string); okStr {
			vocab[s] = id
		}
	}

	var merges []string
	if arr, ok := mergesAny.([]any); ok {
		merges = make([]string, len(arr))
		for i, m := range arr {
			merges[i], _ = m.(string)
		}
	}

	type added struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}
	var specials []added
	if arr, ok := typesAny.([]any); ok {
		for id, tp := range arr {
			if n, ok := tp.(int32); ok && (n == 3 || n == 4) && id < len(tokens) {
				if s, okStr := tokens[id].(string); okStr {
					specials = append(specials, added{ID: id, Content: s})
				}
			}
		}
	}

	spec := map[string]any{
		"pre_tokenizer": json.RawMessage(preJSON),
		"added_tokens":  specials,
		"model": map[string]any{
			"type":   "BPE",
			"vocab":  vocab,
			"merges": merges,
		},
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return tokenizer.Parse(raw)
}
