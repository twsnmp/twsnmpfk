package tensai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTensaiFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "twsnmpfk-tensai-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dummyModel := filepath.Join(tempDir, "dummy.gguf")
	if err := os.WriteFile(dummyModel, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	llm, err := NewWithOptions(dummyModel, true)
	if err != nil {
		t.Fatalf("unexpected error creating TensaiLLM: %v", err)
	}

	ctx := context.Background()
	ans, err := llm.Call(ctx, "Analyze this network status")
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !strings.Contains(ans, "TWSNMP AI Analysis") {
		t.Errorf("expected fallback analysis text, got: %s", ans)
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
