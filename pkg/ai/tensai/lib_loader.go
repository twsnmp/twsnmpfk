package tensai

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	tensaimod "github.com/mattn/tensai"
	"github.com/mattn/tensai/gpu"
	"golang.org/x/sys/cpu"
)

// AccelerationType represents the hardware backend in use
type AccelerationType string

const (
	AccelGPU  AccelerationType = "GPU"
	AccelSIMD AccelerationType = "SIMD (AVX2)"
	AccelCPU  AccelerationType = "CPU (Portable)"
)

var (
	initLibOnce sync.Once
	gpuDevice   *gpu.Device
	gpuInitErr  error
	gpuInitOnce sync.Once
)

// InitWGPULibrary searches for the wgpu-native library in standard locations
// and sets TENSAI_WGPU_LIB if found.
func InitWGPULibrary() {
	initLibOnce.Do(func() {
		if os.Getenv("TENSAI_WGPU_LIB") != "" {
			return
		}

		libName := "libwgpu_native.so"
		switch runtime.GOOS {
		case "darwin":
			libName = "libwgpu_native.dylib"
		case "windows":
			libName = "wgpu_native.dll"
		}

		candidates := make([]string, 0, 8)

		// 1. Next to executable
		if exe, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(exe), libName))
			if runtime.GOOS == "windows" {
				candidates = append(candidates, filepath.Join(filepath.Dir(exe), "libwgpu_native.dll"))
			}
		}

		// 2. Current working directory
		candidates = append(candidates, libName)
		if runtime.GOOS == "windows" {
			candidates = append(candidates, "libwgpu_native.dll")
		}

		// 3. ~/.twsnmpfk/lib/
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, ".twsnmpfk", "lib", libName))
			if runtime.GOOS == "windows" {
				candidates = append(candidates, filepath.Join(home, ".twsnmpfk", "lib", "libwgpu_native.dll"))
			}
		}

		// 4. macOS homebrew paths
		if runtime.GOOS == "darwin" {
			candidates = append(candidates, "/opt/homebrew/lib/"+libName, "/usr/local/lib/"+libName)
		}

		for _, path := range candidates {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				if abs, err := filepath.Abs(path); err == nil {
					_ = os.Setenv("TENSAI_WGPU_LIB", abs)
					break
				}
			}
		}
	})
}

// GetGPUDevice tries to open and initialize the WebGPU device (cached).
// It also registers the GPU device as a global tensai accelerator.
func GetGPUDevice() (*gpu.Device, error) {
	InitWGPULibrary()
	gpuInitOnce.Do(func() {
		dev, err := gpu.Open(gpu.HighPerformance)
		if err != nil {
			// Fallback: try Default power preference
			dev, err = gpu.Open(gpu.Default)
		}
		if err == nil && dev != nil {
			gpuDevice = dev
			// Enable global tensai accelerator for matrix multiplications
			tensaimod.UseAccelerator(dev)
		} else {
			gpuInitErr = err
		}
	})
	return gpuDevice, gpuInitErr
}

// HasAVX2 reports whether the current CPU supports AVX2 instructions
func HasAVX2() bool {
	return runtime.GOARCH == "amd64" && cpu.X86.HasAVX2
}

// DetectAcceleration returns the active acceleration type and details string.
func DetectAcceleration() (AccelerationType, string) {
	return DetectAccelerationWithOptions(false)
}

// DetectAccelerationWithOptions returns the active acceleration type with optional GPU bypass.
func DetectAccelerationWithOptions(noGPU bool) (AccelerationType, string) {
	if !noGPU {
		dev, err := GetGPUDevice()
		if err == nil && dev != nil {
			name := dev.Name()
			if name == "" {
				name = "WebGPU Adapter"
			}
			backend := gpu.Backend()
			if backend != "" {
				return AccelGPU, fmt.Sprintf("%s (%s)", name, backend)
			}
			return AccelGPU, name
		}
		if err != nil && strings.Contains(err.Error(), "built without wgpu support") {
			return AccelCPU, fmt.Sprintf("Pure Go (%s/%s) [binary built without -tags wgpu24]", runtime.GOOS, runtime.GOARCH)
		}
	}

	if HasAVX2() {
		return AccelSIMD, "AVX2 8-lane FMA Vectorized"
	}

	return AccelCPU, fmt.Sprintf("Pure Go (%s/%s)", runtime.GOOS, runtime.GOARCH)
}
