package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPresetModels(t *testing.T) {
	presets := GetPresetNames()
	if len(presets) == 0 {
		t.Fatal("expected at least one preset model")
	}

	for _, p := range presets {
		url, ok := PresetModels[p]
		if !ok || url == "" {
			t.Errorf("preset %s has invalid URL", p)
		}
	}
}

func TestListAndFindModel(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "twsnmpfk-model-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	mPath := filepath.Join(tempDir, "test-model.gguf")
	if err := os.WriteFile(mPath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	models, err := ListModels(tempDir)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "test-model.gguf" {
		t.Errorf("expected name 'test-model.gguf', got %s", models[0].Name)
	}

	found, err := FindModel(tempDir, "test-model")
	if err != nil {
		t.Fatalf("FindModel failed: %v", err)
	}
	if found != mPath {
		t.Errorf("expected %s, got %s", mPath, found)
	}
}
