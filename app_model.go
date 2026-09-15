package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfk/pkg/ai/tensai"
	"github.com/twsnmp/twsnmpfk/pkg/model"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AIHardwareStatus contains hardware acceleration details
type AIHardwareStatus struct {
	Acceleration string `json:"acceleration"`
	Detail       string `json:"detail"`
	ModelDir     string `json:"model_dir"`
	LibDir       string `json:"lib_dir"`
	WGPULibPath  string `json:"wgpu_lib_path"`
	HasGPULib    bool   `json:"has_gpu_lib"`
}

// DownloadProgressEvent is emitted to frontend during downloads
type DownloadProgressEvent struct {
	Downloaded      int64  `json:"downloaded"`
	Total           int64  `json:"total"`
	Percent         int    `json:"percent"`
	DownloadedHuman string `json:"downloaded_human"`
	TotalHuman      string `json:"total_human"`
}

var (
	modelCancelMu  sync.Mutex
	modelCancelCtx context.CancelFunc

	gpuCancelMu  sync.Mutex
	gpuCancelCtx context.CancelFunc
)

// GetAIHardwareStatus returns hardware acceleration info and library paths
func (a *App) GetAIHardwareStatus() AIHardwareStatus {
	tensai.InitWGPULibrary()
	accType, accDetail := tensai.DetectAcceleration()

	wgpuLib := os.Getenv("TENSAI_WGPU_LIB")
	hasLib := wgpuLib != ""

	return AIHardwareStatus{
		Acceleration: string(accType),
		Detail:       accDetail,
		ModelDir:     model.DefaultModelDir(),
		LibDir:       model.DefaultLibDir(),
		WGPULibPath:  wgpuLib,
		HasGPULib:    hasLib,
	}
}

// GetLocalModels returns list of available local models
func (a *App) GetLocalModels() []model.ModelInfo {
	models, err := model.ListModels("")
	if err != nil {
		log.Printf("GetLocalModels err=%v", err)
		return []model.ModelInfo{}
	}
	return models
}

// GetModelPresets returns list of recommended preset models
func (a *App) GetModelPresets() []model.PresetModelInfo {
	return model.PresetModelMetadata
}

// DownloadModel downloads a model by preset name or URL
func (a *App) DownloadModel(target string) string {
	modelCancelMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	modelCancelCtx = cancel
	modelCancelMu.Unlock()

	defer func() {
		modelCancelMu.Lock()
		modelCancelCtx = nil
		modelCancelMu.Unlock()
	}()

	progress := func(downloaded, total int64) {
		pct := 0
		if total > 0 {
			pct = int(float64(downloaded) / float64(total) * 100)
		}
		if a.ctx != nil {
			wails.EventsEmit(a.ctx, "model_download_progress", DownloadProgressEvent{
				Downloaded:      downloaded,
				Total:           total,
				Percent:         pct,
				DownloadedHuman: humanize.Bytes(uint64(downloaded)),
				TotalHuman:      humanize.Bytes(uint64(total)),
			})
		}
	}

	path, err := model.DownloadModel(ctx, "", target, progress)
	if err != nil {
		log.Printf("DownloadModel err=%v", err)
		return err.Error()
	}
	log.Printf("DownloadModel success path=%s", path)
	return ""
}

// CancelModelDownload cancels the ongoing model download
func (a *App) CancelModelDownload() string {
	modelCancelMu.Lock()
	defer modelCancelMu.Unlock()
	if modelCancelCtx != nil {
		modelCancelCtx()
		modelCancelCtx = nil
		return ""
	}
	return "no download in progress"
}

// DeleteLocalModel deletes a local model
func (a *App) DeleteLocalModel(name string) string {
	if err := model.RemoveModel("", name); err != nil {
		log.Printf("DeleteLocalModel err=%v", err)
		return err.Error()
	}
	log.Printf("DeleteLocalModel success name=%s", name)
	return ""
}

// DownloadGPULibrary downloads and installs the wgpu-native library
func (a *App) DownloadGPULibrary() string {
	gpuCancelMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	gpuCancelCtx = cancel
	gpuCancelMu.Unlock()

	defer func() {
		gpuCancelMu.Lock()
		gpuCancelCtx = nil
		gpuCancelMu.Unlock()
	}()

	progress := func(downloaded, total int64) {
		pct := 0
		if total > 0 {
			pct = int(float64(downloaded) / float64(total) * 100)
		}
		if a.ctx != nil {
			wails.EventsEmit(a.ctx, "gpu_download_progress", DownloadProgressEvent{
				Downloaded:      downloaded,
				Total:           total,
				Percent:         pct,
				DownloadedHuman: humanize.Bytes(uint64(downloaded)),
				TotalHuman:      humanize.Bytes(uint64(total)),
			})
		}
	}

	path, err := model.DownloadWGPULibrary(ctx, "", progress)
	if err != nil {
		log.Printf("DownloadGPULibrary err=%v", err)
		return err.Error()
	}

	_ = os.Setenv("TENSAI_WGPU_LIB", path)
	log.Printf("DownloadGPULibrary installed to %s", path)
	return ""
}

// CancelGPUDownload cancels the ongoing GPU library download
func (a *App) CancelGPUDownload() string {
	gpuCancelMu.Lock()
	defer gpuCancelMu.Unlock()
	if gpuCancelCtx != nil {
		gpuCancelCtx()
		gpuCancelCtx = nil
		return ""
	}
	return "no download in progress"
}
