package tensai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattn/tensai/encoding/gguf"
)

func TestTensaiLoadError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "twsnmpfk-tensai-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dummyModel := filepath.Join(tempDir, "dummy.gguf")
	if err := os.WriteFile(dummyModel, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = NewWithOptions(dummyModel, true)
	if err == nil {
		t.Fatalf("expected error creating TensaiLLM with invalid dummy model, got nil")
	}
}

func TestDetectAcceleration(t *testing.T) {
	accType, accDetail := DetectAccelerationWithOptions(true)
	if accType == "" {
		t.Error("expected acceleration type to be non-empty")
	}
	if accDetail == "" {
		t.Error("expected acceleration detail to be non-empty")
	}
}

func TestBuildGGUFTokenizer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "twsnmpfk-tokenizer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Write a mock GGUF file with tokens as string array (elem 8) and types as int32 array (elem 5)
	buf := new(strings.Builder)

	// Header: Magic "GGUF", version 3, nTensors 0, nKV 4
	buf.WriteString("GGUF")
	buf.Write([]byte{3, 0, 0, 0})             // version uint32 = 3
	buf.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0}) // nTensors uint64 = 0
	buf.Write([]byte{4, 0, 0, 0, 0, 0, 0, 0}) // nKV uint64 = 4

	writeStr := func(s string) {
		l := uint64(len(s))
		var b [8]byte
		for i := 0; i < 8; i++ {
			b[i] = byte(l >> (i * 8))
		}
		buf.Write(b[:])
		buf.WriteString(s)
	}

	// 1. tokenizer.ggml.pre = "qwen2" (type 8)
	writeStr("tokenizer.ggml.pre")
	buf.Write([]byte{8, 0, 0, 0}) // type = string (8)
	writeStr("qwen2")

	// 2. tokenizer.ggml.tokens = ["hello", "world", "<|im_end|>"] (type 9 array, elem 8 string)
	writeStr("tokenizer.ggml.tokens")
	buf.Write([]byte{9, 0, 0, 0})             // type = array (9)
	buf.Write([]byte{8, 0, 0, 0})             // elem = string (8)
	buf.Write([]byte{3, 0, 0, 0, 0, 0, 0, 0}) // n = 3
	writeStr("hello")
	writeStr("world")
	writeStr("<|im_end|>")

	// 3. tokenizer.ggml.merges = ["h e", "l l", "o w"] (type 9 array, elem 8 string)
	writeStr("tokenizer.ggml.merges")
	buf.Write([]byte{9, 0, 0, 0})             // type = array (9)
	buf.Write([]byte{8, 0, 0, 0})             // elem = string (8)
	buf.Write([]byte{3, 0, 0, 0, 0, 0, 0, 0}) // n = 3
	writeStr("h e")
	writeStr("l l")
	writeStr("o w")

	// 4. tokenizer.ggml.token_type = [1, 1, 3] (type 9 array, elem 5 int32)
	writeStr("tokenizer.ggml.token_type")
	buf.Write([]byte{9, 0, 0, 0})             // type = array (9)
	buf.Write([]byte{5, 0, 0, 0})             // elem = int32 (5)
	buf.Write([]byte{3, 0, 0, 0, 0, 0, 0, 0}) // n = 3
	buf.Write([]byte{1, 0, 0, 0})             // 1
	buf.Write([]byte{1, 0, 0, 0})             // 1
	buf.Write([]byte{3, 0, 0, 0})             // 3 (control token)

	ggufPath := filepath.Join(tempDir, "mock.gguf")
	if err := os.WriteFile(ggufPath, []byte(buf.String()), 0644); err != nil {
		t.Fatal(err)
	}

	g, err := gguf.Open(ggufPath)
	if err != nil {
		t.Fatalf("gguf.Open failed: %v", err)
	}
	defer g.Close()

	tok, err := buildGGUFTokenizer(g)
	if err != nil {
		t.Fatalf("buildGGUFTokenizer failed: %v", err)
	}
	if tok == nil {
		t.Fatal("expected tokenizer to be non-nil")
	}

	id, ok := tok.ID("<|im_end|>")
	if !ok || id != 2 {
		t.Errorf("expected <|im_end|> token id 2, got id=%d, ok=%v", id, ok)
	}
}
