//go:build !windows
// +build !windows

package cmd

import (
	"os/exec"
)

func GetCmd(path string, params []string) *exec.Cmd {
	if params == nil {
		return exec.Command(path)
	}
	return exec.Command(path, params...)
}
