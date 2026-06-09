//go:build !windows
// +build !windows

package cmd

import (
	"os/exec"
)

func GetCmd(path string, params []string) *exec.Cmd {
	if params == nil {
		// #nosec G204
		return exec.Command(path)
	}
	// #nosec G204
	return exec.Command(path, params...)
}
