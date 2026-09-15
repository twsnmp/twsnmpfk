package main

import (
	"testing"
)

func TestAppModelAPIs(t *testing.T) {
	app := NewApp()

	presets := app.GetModelPresets()
	if len(presets) == 0 {
		t.Fatal("expected non-empty presets from app.GetModelPresets()")
	}

	hw := app.GetAIHardwareStatus()
	if hw.Acceleration == "" {
		t.Error("expected non-empty hardware acceleration string")
	}
	if hw.ModelDir == "" || hw.LibDir == "" {
		t.Error("expected non-empty ModelDir and LibDir")
	}

	models := app.GetLocalModels()
	t.Logf("Found %d local models; hardware: %s (%s)", len(models), hw.Acceleration, hw.Detail)

	// Ensure no crash on cancel when no download in progress
	res := app.CancelModelDownload()
	if res != "no download in progress" {
		t.Errorf("expected 'no download in progress', got %q", res)
	}
	resGPU := app.CancelGPUDownload()
	if resGPU != "no download in progress" {
		t.Errorf("expected 'no download in progress', got %q", resGPU)
	}

	// Verify that directories are set to ~/.twsnmpfk/
}
