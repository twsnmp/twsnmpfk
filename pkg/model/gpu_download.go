package model

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const WGPUVersion = "v24.0.0.1"

// DefaultLibDir returns the primary directory where native libraries are stored (~/.twsnmpfk/lib)
func DefaultLibDir() string {
	if dir, err := os.UserHomeDir(); err == nil && dir != "" {
		return filepath.Join(dir, ".twsnmpfk", "lib")
	}
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		return filepath.Join(u.HomeDir, ".twsnmpfk", "lib")
	}
	return "twsnmpfk-lib"
}

// GetWGPUAssetInfo returns the download URL and expected library filename for the current OS/Arch
func GetWGPUAssetInfo() (string, string, error) {
	baseURL := fmt.Sprintf("https://github.com/gfx-rs/wgpu-native/releases/download/%s", WGPUVersion)

	var zipName string
	var libFileName string

	switch runtime.GOOS {
	case "darwin":
		libFileName = "libwgpu_native.dylib"
		switch runtime.GOARCH {
		case "arm64":
			zipName = "wgpu-macos-aarch64-release.zip"
		case "amd64":
			zipName = "wgpu-macos-x86_64-release.zip"
		default:
			return "", "", fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
	case "linux":
		libFileName = "libwgpu_native.so"
		switch runtime.GOARCH {
		case "amd64":
			zipName = "wgpu-linux-x86_64-release.zip"
		case "arm64":
			zipName = "wgpu-linux-aarch64-release.zip"
		default:
			return "", "", fmt.Errorf("unsupported Linux architecture: %s", runtime.GOARCH)
		}
	case "windows":
		libFileName = "wgpu_native.dll"
		switch runtime.GOARCH {
		case "amd64":
			zipName = "wgpu-windows-x86_64-msvc-release.zip"
		default:
			return "", "", fmt.Errorf("unsupported Windows architecture: %s", runtime.GOARCH)
		}
	default:
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	url := fmt.Sprintf("%s/%s", baseURL, zipName)
	return url, libFileName, nil
}

// DownloadWGPULibrary downloads and installs the wgpu-native library for the current platform
func DownloadWGPULibrary(ctx context.Context, libDir string, progress func(downloaded, total int64)) (string, error) {
	if libDir == "" {
		libDir = DefaultLibDir()
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create library directory: %w", err)
	}

	downloadURL, libFileName, err := GetWGPUAssetInfo()
	if err != nil {
		return "", err
	}

	destPath := filepath.Join(libDir, libFileName)

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "twsnmpfk-downloader/1.0")

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with HTTP status: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	var downloaded int64
	var zipBuf bytes.Buffer

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			zipBuf.Write(buf[:n])
			downloaded += int64(n)
			if progress != nil {
				progress(downloaded, totalSize)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error reading response body: %w", err)
		}
	}

	// Extract the target dynamic library from zip
	reader, err := zip.NewReader(bytes.NewReader(zipBuf.Bytes()), int64(zipBuf.Len()))
	if err != nil {
		return "", fmt.Errorf("failed to read zip archive: %w", err)
	}

	var foundFile *zip.File
	for _, f := range reader.File {
		base := filepath.Base(f.Name)
		if strings.EqualFold(base, libFileName) {
			foundFile = f
			break
		}
	}

	if foundFile == nil {
		return "", fmt.Errorf("could not find %s inside the downloaded archive", libFileName)
	}

	rc, err := foundFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file inside zip: %w", err)
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		return "", fmt.Errorf("failed to write library file: %w", err)
	}

	return destPath, nil
}
