package model

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// ModelInfo represents information about a local model
type ModelInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	SizeHuman string    `json:"size_human"`
	ModTime   time.Time `json:"mod_time"`
	Type      string    `json:"type"` // gguf, directory
}

// DefaultModelDir returns the primary directory where models are stored (~/.twsnmpfk/models)
func DefaultModelDir() string {
	if dir, err := os.UserHomeDir(); err == nil && dir != "" {
		return filepath.Join(dir, ".twsnmpfk", "models")
	}
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		return filepath.Join(u.HomeDir, ".twsnmpfk", "models")
	}
	return "twsnmpfk-models"
}

// PresetModelInfo holds metadata for a preset model
type PresetModelInfo struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Params      string `json:"params"`
}

// PresetModels defines recommended lightweight models for twsnmpfk
var PresetModels = map[string]string{
	"qwen2.5-0.5b":       "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q8_0.gguf",
	"qwen2.5-1.5b":       "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf",
	"qwen2.5-coder-0.5b": "https://huggingface.co/Qwen/Qwen2.5-Coder-0.5B-Instruct-GGUF/resolve/main/qwen2.5-coder-0.5b-instruct-q8_0.gguf",
	"qwen2.5-coder-1.5b": "https://huggingface.co/Qwen/Qwen2.5-Coder-1.5B-Instruct-GGUF/resolve/main/qwen2.5-coder-1.5b-instruct-q4_k_m.gguf",
	"smollm2-360m":       "https://huggingface.co/HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/resolve/main/smollm2-360m-instruct-q8_0.gguf",
	"smollm2-1.7b":       "https://huggingface.co/HuggingFaceTB/SmolLM2-1.7B-Instruct-GGUF/resolve/main/smollm2-1.7b-instruct-q4_k_m.gguf",
	"llama-3.2-1b":       "https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_K_M.gguf",
	"deepseek-r1-1.5b":   "https://huggingface.co/unsloth/DeepSeek-R1-Distill-Qwen-1.5B-GGUF/resolve/main/DeepSeek-R1-Distill-Qwen-1.5B-Q4_K_M.gguf",
	"tinyllama":          "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf",
}

// PresetModelMetadata provides UI-friendly info for presets
var PresetModelMetadata = []PresetModelInfo{
	{Name: "qwen2.5-0.5b", URL: PresetModels["qwen2.5-0.5b"], Description: "Qwen 2.5 0.5B Instruct (Q8_0, Recommended Fast)", Size: "~500 MB", Params: "0.5B"},
	{Name: "qwen2.5-1.5b", URL: PresetModels["qwen2.5-1.5b"], Description: "Qwen 2.5 1.5B Instruct (Q4_K_M, Balanced)", Size: "~980 MB", Params: "1.5B"},
	{Name: "qwen2.5-coder-0.5b", URL: PresetModels["qwen2.5-coder-0.5b"], Description: "Qwen 2.5 Coder 0.5B Instruct (Q8_0)", Size: "~500 MB", Params: "0.5B"},
	{Name: "qwen2.5-coder-1.5b", URL: PresetModels["qwen2.5-coder-1.5b"], Description: "Qwen 2.5 Coder 1.5B Instruct (Q4_K_M)", Size: "~980 MB", Params: "1.5B"},
	{Name: "smollm2-360m", URL: PresetModels["smollm2-360m"], Description: "SmolLM2 360M Instruct (Q8_0, Ultra-lightweight)", Size: "~380 MB", Params: "360M"},
	{Name: "smollm2-1.7b", URL: PresetModels["smollm2-1.7b"], Description: "SmolLM2 1.7B Instruct (Q4_K_M)", Size: "~1.1 GB", Params: "1.7B"},
	{Name: "llama-3.2-1b", URL: PresetModels["llama-3.2-1b"], Description: "Llama 3.2 1B Instruct (Q4_K_M)", Size: "~750 MB", Params: "1B"},
	{Name: "deepseek-r1-1.5b", URL: PresetModels["deepseek-r1-1.5b"], Description: "DeepSeek R1 Distill Qwen 1.5B (Q4_K_M)", Size: "~1.1 GB", Params: "1.5B"},
	{Name: "tinyllama", URL: PresetModels["tinyllama"], Description: "TinyLlama 1.1B Chat (Q4_K_M)", Size: "~670 MB", Params: "1.1B"},
}

// ListModels returns a list of models in modelDir
func ListModels(modelDir string) ([]ModelInfo, error) {
	if modelDir == "" {
		modelDir = DefaultModelDir()
	}

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(modelDir)
	if err != nil {
		return nil, err
	}

	var list []ModelInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		name := entry.Name()
		fullPath := filepath.Join(modelDir, name)

		if entry.IsDir() {
			cfgPath := filepath.Join(fullPath, "config.json")
			if _, err := os.Stat(cfgPath); err == nil {
				var dirSize int64
				_ = filepath.Walk(fullPath, func(_ string, fi os.FileInfo, _ error) error {
					if fi != nil && !fi.IsDir() {
						dirSize += fi.Size()
					}
					return nil
				})
				list = append(list, ModelInfo{
					Name:      name,
					Path:      fullPath,
					Size:      dirSize,
					SizeHuman: humanize.Bytes(uint64(dirSize)),
					ModTime:   info.ModTime(),
					Type:      "directory",
				})
			}
		} else if strings.HasSuffix(strings.ToLower(name), ".gguf") {
			list = append(list, ModelInfo{
				Name:      name,
				Path:      fullPath,
				Size:      info.Size(),
				SizeHuman: humanize.Bytes(uint64(info.Size())),
				ModTime:   info.ModTime(),
				Type:      "gguf",
			})
		}
	}

	return list, nil
}

// FindModel locates a model by name, filename, preset name, or returns the first available model if name is empty
func FindModel(modelDir, name string) (string, error) {
	primaryDir := modelDir
	if primaryDir == "" {
		primaryDir = DefaultModelDir()
	}

	if name != "" {
		if fi, err := os.Stat(name); err == nil {
			if fi.IsDir() {
				if _, err := os.Stat(filepath.Join(name, "config.json")); err == nil {
					return name, nil
				}
			} else if strings.HasSuffix(strings.ToLower(name), ".gguf") {
				return name, nil
			}
		}

		p := filepath.Join(primaryDir, name)
		if fi, err := os.Stat(p); err == nil {
			if !fi.IsDir() || fileExists(filepath.Join(p, "config.json")) {
				return p, nil
			}
		}
		if !strings.HasSuffix(strings.ToLower(p), ".gguf") {
			pGguf := p + ".gguf"
			if _, err := os.Stat(pGguf); err == nil {
				return pGguf, nil
			}
		}

		// Check if name is a known preset
		if presetURL, ok := PresetModels[strings.ToLower(name)]; ok {
			filename := filepath.Base(presetURL)
			if idx := strings.Index(filename, "?"); idx != -1 {
				filename = filename[:idx]
			}
			pPreset := filepath.Join(primaryDir, filename)
			if _, err := os.Stat(pPreset); err == nil {
				return pPreset, nil
			}
		}
	}

	models, err := ListModels(modelDir)
	if err != nil {
		return "", err
	}
	if len(models) == 0 {
		return "", fmt.Errorf("no models found in %s", primaryDir)
	}
	if name == "" {
		return models[0].Path, nil
	}

	// 1. Exact match (by filename or filename without extension)
	for _, m := range models {
		if strings.EqualFold(m.Name, name) || strings.EqualFold(strings.TrimSuffix(m.Name, ".gguf"), name) {
			return m.Path, nil
		}
	}

	// 2. Preset match against list of models
	if presetURL, ok := PresetModels[strings.ToLower(name)]; ok {
		presetFile := filepath.Base(presetURL)
		if idx := strings.Index(presetFile, "?"); idx != -1 {
			presetFile = presetFile[:idx]
		}
		presetBase := strings.TrimSuffix(presetFile, ".gguf")
		for _, m := range models {
			if strings.EqualFold(m.Name, presetFile) || strings.EqualFold(strings.TrimSuffix(m.Name, ".gguf"), presetBase) {
				return m.Path, nil
			}
		}
	}

	// 3. Prefix or contains match (e.g. "qwen2.5-coder" matches "qwen2.5-coder-0.5b-instruct-q8_0.gguf")
	nameLower := strings.ToLower(name)
	for _, m := range models {
		mBase := strings.ToLower(strings.TrimSuffix(m.Name, ".gguf"))
		if strings.HasPrefix(mBase, nameLower) || strings.Contains(mBase, nameLower) {
			return m.Path, nil
		}
	}

	return "", fmt.Errorf("model %q not found in %s", name, primaryDir)
}

// DownloadProgress callback for download monitoring
type DownloadProgress func(downloaded, total int64)

// DownloadModel downloads a model from a preset name or URL to the model directory
func DownloadModel(ctx context.Context, modelDir, target string, progress DownloadProgress) (string, error) {
	if modelDir == "" {
		modelDir = DefaultModelDir()
	}
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return "", err
	}

	url := target
	var filename string

	if presetURL, ok := PresetModels[strings.ToLower(target)]; ok {
		url = presetURL
		filename = filepath.Base(presetURL)
	} else if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		url = target
		filename = filepath.Base(target)
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
	} else if strings.Contains(target, "/") {
		parts := strings.Split(target, "/")
		if len(parts) >= 3 && strings.HasSuffix(parts[len(parts)-1], ".gguf") {
			repo := strings.Join(parts[:len(parts)-1], "/")
			file := parts[len(parts)-1]
			url = fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repo, file)
			filename = file
		} else {
			repo := target
			file := strings.ToLower(parts[len(parts)-1]) + "-q8_0.gguf"
			url = fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repo, file)
			filename = file
		}
	} else {
		return "", fmt.Errorf("unknown model preset or invalid URL: %s. Available presets: %s",
			target, strings.Join(GetPresetNames(), ", "))
	}

	if filename == "" {
		filename = "model.gguf"
	}
	destPath := filepath.Join(modelDir, filename)
	tmpPath := destPath + ".tmp"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download model: HTTP %s (%s)", resp.Status, url)
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	defer func() {
		out.Close()
		_ = os.Remove(tmpPath)
	}()

	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 64*1024)
	lastUpdate := time.Now()
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return "", writeErr
			}
			downloaded += int64(n)
			if progress != nil && time.Since(lastUpdate) > 100*time.Millisecond {
				progress(downloaded, totalSize)
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", readErr
		}
	}

	if progress != nil {
		progress(downloaded, totalSize)
	}

	out.Close()
	if err := os.Rename(tmpPath, destPath); err != nil {
		return "", err
	}

	return destPath, nil
}

// RemoveModel deletes a model from the model directory
func RemoveModel(modelDir, name string) error {
	p, err := FindModel(modelDir, name)
	if err != nil {
		if _, statErr := os.Stat(name); statErr == nil {
			return os.RemoveAll(name)
		}
		return err
	}
	return os.RemoveAll(p)
}

// GetPresetNames returns the available preset model names
func GetPresetNames() []string {
	var names []string
	for k := range PresetModels {
		names = append(names, k)
	}
	return names
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
